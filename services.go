package v2scar

import (
	"context"
	"log"
	"strings"

	v2proxyman "v2ray.com/core/app/proxyman/command"
	v2stats "v2ray.com/core/app/stats/command"
	"v2ray.com/core/common/protocol"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/proxy/vmess"
)

// GetAndResetUserTraffic 统计所有user的上行下行流量
// V2ray的stats的统计模块设计的非常奇怪，具体规则如下
// 上传流量："user>>>" + user.Email + ">>>traffic>>>uplink"
// 下载流量："user>>>" + user.Email + ">>>traffic>>>downlink"
func GetAndResetUserTraffic(c v2stats.StatsServiceClient, up *UserPool) {
	req := &v2stats.QueryStatsRequest{
		Pattern: "user>>>",
		Reset_:  true,
	}
	resp, err := c.QueryStats(context.Background(), req)
	if err != nil {
		log.Println("[ERROR]:", err)
	} else {
		for _, stat := range resp.Stat {
			email, trafficType := getEmailAndTrafficType(stat.Name)
			user, err := up.GetUserByEmail(email)
			if err != nil {
				log.Println(err)
			} else {
				switch trafficType {
				case "uplink":
					user.setUploadTraffic(stat.Value)
				case "downlink":
					user.setDownloadTraffic(stat.Value)
				}
			}
		}
	}
}

func getEmailAndTrafficType(input string) (string, string) {
	s := strings.Split(input, ">>>")
	return s[1], s[len(s)-1]
}

// AddInboundUser add user to inbound by tag
func AddInboundUser(c v2proxyman.HandlerServiceClient, tag string, user *User) {
	_, err := c.AlterInbound(context.Background(), &v2proxyman.AlterInboundRequest{
		Tag: tag,
		Operation: serial.ToTypedMessage(&v2proxyman.AddUserOperation{
			User: &protocol.User{
				Level: user.Level,
				Email: user.Email,
				Account: serial.ToTypedMessage(&vmess.Account{
					Id:               user.UUID,
					AlterId:          user.AlterId,
					SecuritySettings: &protocol.SecurityConfig{Type: protocol.SecurityType_AUTO},
				}),
			},
		}),
	})
	if err != nil {
		log.Println("[ERROR]:", err)
		if strings.Contains(err.Error(), "already exists.") {
			// TODO 优化这里的逻辑 这里针对side car重启而v2ray没重启的状态
			user.setRunning(true)
		}
	} else {
		log.Printf("[INFO] User: %v Add To V2ray Server Tag: %v", user.Email, tag)
		user.setRunning(true)
	}
}

//RemoveInboundUser remove user from inbound by tag
func RemoveInboundUser(c v2proxyman.HandlerServiceClient, tag string, user *User) {
	_, err := c.AlterInbound(context.Background(), &v2proxyman.AlterInboundRequest{
		Tag: tag,
		Operation: serial.ToTypedMessage(&v2proxyman.RemoveUserOperation{
			Email: user.Email,
		}),
	})
	if err != nil {
		log.Println("[ERROR]:", err)
	} else {
		log.Printf("[INFO] User: %v Removed From V2ray Server Tag: %v", user.Email, tag)
		user.setRunning(false)
	}
}
