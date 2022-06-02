package main

import (
	"ip_detect/app"
	"ip_detect/db/redis"
	"ip_detect/utils/log"
	"os"
)

func init() {
	log.NewLogger(3)
	if err := redis.Connect(); err != nil {
		log.GlobalLog.Errorf("redis connect %v", err.Error())
		os.Exit(1)
	}
}

func main() {
	app.Run()
}
