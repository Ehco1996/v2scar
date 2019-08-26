package v2scar

import (
	"context"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"
	v2proxyman "v2ray.com/core/app/proxyman/command"
	v2stats "v2ray.com/core/app/stats/command"
)

var API_ENDPOINT = "127.0.0.1:8080"
var GRPC_ENDPOINT = "http://fun.com/api"

type UserConfig struct {
	UserId  int    `json:"user_id"`
	Email   string `json:"email"`
	UUID    string `json:"uuid"`
	AlterId uint32 `json:"alter_id"`
	Level   uint32 `json:"level"`
	Enable  bool   `json:"enable"`
}
type UserTraffic struct {
	UserId          int   `json:"user_id"`
	DownloadTraffic int64 `json:"dt"`
	UploadTraffic   int64 `json:"ut"`
}

type syncReq struct {
	UserTraffics []*UserTraffic `json:"user_traffics"`
}

type syncResp struct {
	Configs []*UserConfig
	Tag     string `json:"tag"`
}

func SyncTask(up *UserPool) {

	// Connect to v2ray rpc
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, GRPC_ENDPOINT, grpc.WithInsecure(), grpc.WithBlock())
	defer conn.Close()
	if err != nil {
		log.Printf("[WARNING]: GRPC连接失败,请检查V2ray是否运行并开放对应grpc端口 当前GRPC地址: %v", GRPC_ENDPOINT)
		return
	}

	// Init Client
	proxymanClient := v2proxyman.NewHandlerServiceClient(conn)
	statClient := v2stats.NewStatsServiceClient(conn)
	httpClient := &http.Client{Timeout: 3 * time.Second}

	resp := syncResp{}
	err = getJson(httpClient, API_ENDPOINT, &resp)
	if err != nil {
		log.Printf("[WARNING]: API连接失败,请检查API地址 当前地址: %v", API_ENDPOINT)
		return
	}

	// get user config
	initOrUpdateUser(up, proxymanClient, &resp)

	// sync user traffic
	syncUserTrafficToServer(up, statClient, httpClient)

}

func initOrUpdateUser(up *UserPool, c v2proxyman.HandlerServiceClient, sr *syncResp) {
	log.Println("[INFO] Call initOrUpdateUser")
	for _, cfg := range sr.Configs {
		user, _ := up.GetUserByEmail(cfg.Email)
		if user == nil {
			// New User
			newUser, err := up.CreateUser(cfg.UserId, cfg.Email, cfg.UUID, cfg.Level, cfg.AlterId, cfg.Enable)
			if err != nil {
				log.Fatalln(err)
			}
			if newUser.Enable {
				AddInboundUser(c, sr.Tag, newUser)
			}
		} else {
			// Old User
			if user.Enable != cfg.Enable {
				// update enable status
				user.setEnable(cfg.Enable)
			}
			if user.Enable && !user.running {
				// Start Not Running user
				AddInboundUser(c, sr.Tag, user)
			}
			if !user.Enable && user.running {
				// Close Not Enable user
				RemoveInboundUser(c, sr.Tag, user)
			}
		}
	}
}

func syncUserTrafficToServer(up *UserPool, c v2stats.StatsServiceClient, hc *http.Client) {
	GetAndResetUserTraffic(c, up)

	tfs := make([]*UserTraffic, 0, up.GetUsersNum())
	for _, user := range up.GetAllUsers() {
		tf := user.DownloadTraffic + user.UploadTraffic
		if tf > 0 {
			log.Printf("[INFO] User: %v Now Used Total Traffic: %v", user.Email, tf)
			tfs = append(tfs, &UserTraffic{
				UserId:          user.UserId,
				DownloadTraffic: user.DownloadTraffic,
				UploadTraffic:   user.UploadTraffic,
			})
			user.resetTraffic()
		}
	}
	postJson(hc, API_ENDPOINT, &syncReq{UserTraffics: tfs})
	log.Printf("[INFO] Call syncUserTrafficToServer ONLINE USER COUNT: %d", len(tfs))
}
