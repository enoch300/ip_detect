package main

import (
	"github.com/enoch300/nt/mtr"
	"ip_detect/api"
	"ip_detect/utils/log"
)

func init() {
	log.NewLogger(3)
}

func detect(t api.Targets) {
	_, hops, err := mtr.Mtr("0.0.0.0", t.Ip, 32, 15, 800)
	if err != nil {
		log.GlobalLog.Errorf("mtr %v", err.Error())
		return
	}

	if hops[len(hops)-1].Addr == "???" {
		api.Alert(t, hops[len(hops)-1])
	}
}

func main() {
	for _, target := range api.MonitorTargets() {
		go detect(target)
	}
	select {}
}
