package v2scar

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	v2proxyman "v2ray.com/core/app/proxyman/command"
	v2stats "v2ray.com/core/app/stats/command"
)

var API_ENDPOINT = os.Getenv("V2SCAR_API_ENDPOINT")
var GRPC_ENDPOINT = os.Getenv("V2SCAR_GRPC_ENDPOINT")

type UserConfig struct {
	Email   string `json:"email"`
	UUID    string `json:"uuid"`
	AlterId uint32 `json:"alter_id"`
	Level   uint32 `json:"level"`
	Enable  bool   `json:"enable"`
}

type syncResp struct {
	Configs []*UserConfig
	Tag     string `json:"tag"`
}

func SyncJob(up *UserPool) {

	// Connect to v2ray rpc
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, GRPC_ENDPOINT, grpc.WithInsecure(), grpc.WithBlock())
	defer conn.Close()
	if err != nil {
		log.Fatalf("GRPC连接失败,请检查V2ray是否运行并开放对应grpc端口 当前GRPC地址: %v", GRPC_ENDPOINT)
	}

	// Init Client
	proxymanClient := v2proxyman.NewHandlerServiceClient(conn)
	statClient := v2stats.NewStatsServiceClient(conn)
	httpClient := &http.Client{Timeout: 3 * time.Second}

	resp := syncResp{}
	err = getJson(httpClient, API_ENDPOINT, &resp)
	if err != nil {
		log.Fatalf("APi连接失败,请检查API地址 当前地址: %v", API_ENDPOINT)
	}
	initOrUpdateUser(up, proxymanClient, &resp)
	syncUserTrafficToServer(up, statClient)
}

func initOrUpdateUser(up *UserPool, c v2proxyman.HandlerServiceClient, sr *syncResp) {
	log.Println("[INFO] Call initOrUpdateUser")
	for _, cfg := range sr.Configs {
		user, _ := up.GetUserByEmail(cfg.Email)
		if user == nil {
			// New User
			newUser, err := up.CreateUser(cfg.Email, cfg.UUID, cfg.Level, cfg.AlterId, cfg.Enable)
			if err != nil {
				log.Fatalln(err)
			}
			AddInboundUser(c, sr.Tag, newUser)
		} else {
			// Old User
			if user.Enable != cfg.Enable {
				// update enable status
				user.setEnable(cfg.Enable)
			}
			if user.Enable && !user.running {
				// Start not Running user
				AddInboundUser(c, sr.Tag, user)
			}
			if !user.Enable && user.running {
				// Close Not Enable user
				RemoveInboundUser(c, sr.Tag, user)
			}
		}
	}
}

func syncUserTrafficToServer(up *UserPool, c v2stats.StatsServiceClient) {
	// TODO sync
	log.Println("[INFO] Call syncUserTrafficToServer")
	GetAndResetUserTraffic(c, up)
	for _, user := range up.GetAllUsers() {
		tf := user.DownloadTraffic + user.UploadTraffic
		if tf > 0 {
			log.Printf("[INFO] User: %v Now Used Total Traffic: %v", user.Email)
		}
	}

}
