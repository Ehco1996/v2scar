package main

import (
	"time"

	"github.com/Ehco1996/v2scar"
)

func main() {
	up := v2scar.NewUserPool()
	tick := time.Tick(60 * time.Second)
	for {
		go v2scar.SyncJob(up)
		<-tick
	}
}
