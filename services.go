package v2scar

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	v2stats "v2ray.com/core/app/stats/command"
)

// GetAndResetUserTraffic 统计所有user的上行下行流量
// V2ray的stats的统计模块设计的非常奇怪，具体规则如下
// 上传流量："user>>>" + user.Email + ">>>traffic>>>uplink"
// 下载流量："user>>>" + user.Email + ">>>traffic>>>downlink"
func GetAndResetUserTraffic(ctx context.Context, conn *grpc.ClientConn, up *UserPool) {
	client := v2stats.NewStatsServiceClient(conn)
	req := &v2stats.QueryStatsRequest{
		Pattern: "user>>>",
		Reset_:  true,
	}
	resp, _ := client.QueryStats(ctx, req)
	for _, stat := range resp.Stat {
		email, trafficType := getEmailAndTrafficType(stat.Name)
		user := up.GetOrCreateUser(email)
		switch trafficType {
		case "uplink":
			user.setUploadTraffic(stat.Value)
		case "downlink":
			user.setDownloadTraffic(stat.Value)
		}
	}
}

func getEmailAndTrafficType(input string) (string, string) {
	s := strings.Split(input, ">>>")
	return s[1], s[len(s)-1]
}
