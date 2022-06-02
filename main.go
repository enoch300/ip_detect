package main

import (
	"fmt"
	"ip_detect/app"
	"ip_detect/db/redis"
	"ip_detect/utils/log"
	"os"
)

func init() {
	log.NewLogger(3)
	if err := redis.Connect(); err != nil {
		fmt.Printf("redis connect %v\n", err.Error())
		log.GlobalLog.Errorf("redis connect %v", err.Error())
		os.Exit(1)
	}
}

func main() {
	app.Run()
}
