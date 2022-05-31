package main

import (
	"ip_detect/app"
	"ip_detect/utils/log"
)

func init() {
	log.NewLogger(3)
}

func main() {
	app.Run()
}
