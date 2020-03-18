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

var API_ENDPOINT string
var GRPC_ENDPOINT string

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

	if err != nil {
		log.Printf("[WARNING]: GRPC连接失败,请检查V2ray是否运行并开放对应grpc端口 当前GRPC地址: %v 错误信息: %v", GRPC_ENDPOINT, err.Error())
		return
	} else {
		defer conn.Close()
	}

	// Init Client
	proxymanClient := v2proxyman.NewHandlerServiceClient(conn)
	statClient := v2stats.NewStatsServiceClient(conn)
	httpClient := &http.Client{Timeout: 3 * time.Second}

	resp := syncResp{}
	err = getJson(httpClient, API_ENDPOINT, &resp)
	if err != nil {
		log.Printf("[WARNING]: API连接失败,请检查API地址 当前地址: %v 错误信息:%v", API_ENDPOINT, err.Error())
		return
	}

	// init or update user config
	initOrUpdateUser(up, proxymanClient, &resp)

	// sync user traffic
	syncUserTrafficToServer(up, statClient, httpClient)
}

func initOrUpdateUser(up *UserPool, c v2proxyman.HandlerServiceClient, sr *syncResp) {
	// Enable line numbers in logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("[INFO] Call initOrUpdateUser")

	syncUserMap := make(map[string]bool)

	for _, cfg := range sr.Configs {
		syncUserMap[cfg.Email] = true
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

			// change user uuid
			if user.UUID != cfg.UUID {
				log.Printf("[INFO] user: %s 更换了uuid old: %s new: %s", user.Email, user.UUID, cfg.UUID)
				RemoveInboundUser(c, sr.Tag, user)
				user.setUUID(cfg.UUID)
				AddInboundUser(c, sr.Tag, user)
			}

			// remove not enable user
			if !user.Enable && user.running {
				// Close Not Enable user
				RemoveInboundUser(c, sr.Tag, user)
			}

			// start not runing user
			if user.Enable && !user.running {
				// Start Not Running user
				AddInboundUser(c, sr.Tag, user)
			}

		}
	}

	// remote user not in server
	for _, user := range up.GetAllUsers() {
		if _, ok := syncUserMap[user.Email]; !ok {
			RemoveInboundUser(c, sr.Tag, user)
			up.RemoveUserByEmail(user.Email)
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
