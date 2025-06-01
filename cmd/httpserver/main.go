package main

import (
	"log"

	"github.krafton.com/ops2022/noti-aws-update/internal"
)

func main() {
	cfg := internal.LoadConfig()

	pool, err := internal.NewDBPool(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	internal.StartHTTPServer(pool, cfg.AppPort)
}
