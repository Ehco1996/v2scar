package main

import (
	"log"
	"os"
	"time"

	"github.com/urfave/cli"

	"github.com/Ehco1996/v2scar"
)

var SYNC_TIME int

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	app := cli.NewApp()
	app.Name = "v2scar"
	app.Usage = "sidecar for V2ray"
	app.Version = "0.0.11"
	app.Author = "Ehco1996"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "grpc-endpoint, gp",
			Value:       "127.0.0.1:8080",
			Usage:       "V2ray开放的GRPC地址",
			EnvVar:      "V2SCAR_GRPC_ENDPOINT",
			Destination: &v2scar.GRPC_ENDPOINT,
		},
		cli.StringFlag{
			Name:        "api-endpoint, ap",
			Value:       "http://fun.com/api",
			Usage:       "django-sspanel开放的Vemss Node Api地址",
			EnvVar:      "V2SCAR_API_ENDPOINT",
			Destination: &v2scar.API_ENDPOINT,
		},
		cli.IntFlag{
			Name:        "sync-time, st",
			Value:       60,
			Usage:       "与django-sspanel同步的时间间隔",
			EnvVar:      "V2SCAR_SYNC_TIME",
			Destination: &SYNC_TIME,
		},
	}

	app.Action = func(c *cli.Context) error {
		up := v2scar.NewUserPool()
		log.Println("Waitting v2ray start...")
		time.Sleep(time.Second * 3)
		tick := time.Tick(time.Duration(SYNC_TIME) * time.Second)
		for {
			go v2scar.SyncTask(up)
			<-tick
		}
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
