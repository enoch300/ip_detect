package app

import (
	"ip_detect/app/sub"
)

func Run() {
	sub.SubMessage("agent2")
	//go alarm.Check()
	//for {
	//	for _, target := range api.Targets() {
	//		if target.Biz == "kuaishou" {
	//			if net.ParseIP(target.OuterIp).To4() != nil {
	//				task := NewTask(target, false, false, true)
	//				go detect(task)
	//			}
	//		} else {
	//			task := NewTask(target, true, false, true)
	//			go detect(task)
	//		}
	//	}
	//	time.Sleep(5 * time.Minute)
	//}
}
