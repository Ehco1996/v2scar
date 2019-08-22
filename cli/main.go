package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	v2scar "github.com/Ehco1996/v2scar"
)

func main() {

	up := v2scar.NewUserPool()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, _ := grpc.DialContext(ctx, "127.0.0.1:8080", grpc.WithInsecure(), grpc.WithBlock())
	defer conn.Close()

	v2scar.GetAndResetUserTraffic(ctx, conn, up)

	for _, user := range up.GetAllUsers() {
		fmt.Println(user)

	}
}
