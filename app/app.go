package app

import (
	"fmt"
	"ip_detect/app/sub"
	"os"
)

func Run() {
	hostname, _ := os.Hostname()
	sub.SubMessage(fmt.Sprintf("ip_detect_%s", hostname))
}
