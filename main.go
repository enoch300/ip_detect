package main

import (
	"github.com/enoch300/nt/mtr"
	"ip_detect/api"
	"ip_detect/utils/log"
	"time"
)

func init() {
	log.NewLogger(3)
}

func detect(t api.Targets) {
	t.T = time.Now().Format("2006-01-02 15:04:05")
	_, hops, err := mtr.Mtr("0.0.0.0", t.Ip, 32, 3, 800)
	if err != nil {
		log.GlobalLog.Errorf("mtr %v", err.Error())
		return
	}
	if hops[len(hops)-1].Addr == "???" || hops[len(hops)-1].Loss < 5 {
		api.Alert(t, hops)
	}
}

func main() {
	i := 0
	for _, target := range api.MonitorTargets() {
		go detect(target)
		i++
		if i > 2 {
			break
		}
	}
	select {}
}
