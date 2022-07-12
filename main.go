package main

import (
	"ip_detect/app"
	"ip_detect/db/redis"
	"ip_detect/utils/logger"
	"os"
)

func init() {
	logger.InitLogger()
	if err := redis.Connect(); err != nil {
		logger.Global.Errorf("redis connect %v", err.Error())
		os.Exit(1)
	}
}

func main() {
	app.Run()
}
