package main

import (
	"github.com/enoch300/nt/mtr"
	"github.com/enoch300/nt/ping"
	"github.com/pochard/commons/randstr"
	"ip_detect/alarm"
	"ip_detect/api"
	"ip_detect/utils/log"
	"math/rand"
	"time"
)

func init() {
	log.NewLogger(3)
}

func detect(t *api.Targets) {
	//将任务打散发起请求
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(300)
	time.Sleep(time.Second * time.Duration(n))

	t.T = time.Now().Format("2006-01-02 15:04:05")
	opId := randstr.RandomAlphanumeric(17)
	_, pingReturn, err := ping.Ping("0.0.0.0", t.Ip, 32, 1000, 1000)
	if err != nil {
		log.GlobalLog.Errorf("ping %v", err.Error())
		return
	}
	log.GlobalLog.Infof("监测时间: %s, 业务名: %v, 业务ID: %v, 业务BD:%v, 触发策略: %v, 监控目标: %v(%s), 平均延时: %.2f, 最大延时: %.2f, 最小延时: %.2f, 丢包率: %.2f",
		t.T, t.Biz, t.BId, t.BD, "丢包率>5%", t.Ip, t.Region, pingReturn.AvgTime.Seconds()*1000, pingReturn.WrstTime.Seconds()*1000, pingReturn.BestTime.Seconds()*1000, pingReturn.DropRate)

	var hops []mtr.Hop
	//_, hops, err := mtr.Mtr("0.0.0.0", t.Ip, 32, 2, 800)
	//if err != nil {
	//	log.GlobalLog.Errorf("mtr %v", err.Error())
	//	return
	//}
	//
	//lastHop := hops[len(hops)-1]
	//log.GlobalLog.Infof("监测时间: %s, 业务名: %v, 业务ID: %v, 业务BD:%v, 触发策略: %v, 监控目标: %v(%s), 平均延时: %.2f, 最大延时: %.2f, 最小延时: %.2f, 丢包率: %.2f",
	//	t.T, t.Biz, t.BId, t.BD, "丢包率>5%", t.Ip, t.Region, lastHop.Avg, lastHop.Wrst, lastHop.Best, lastHop.Loss)
	//var mtrStr string
	//for _, hop := range hops {
	//	if hop.Addr == "???" {
	//		mtrStr += fmt.Sprintf("%-5v\t%-20v\t%-8.1f\t%-8d\t%-8.1f\t%-5.1f%8.f\n", hop.RouteNo, hop.Addr, hop.Loss, hop.Snt, hop.Avg, hop.Best, hop.Wrst)
	//		continue
	//	}
	//	mtrStr += fmt.Sprintf("%-5v\t%-15v\t%-8.1f\t%-8d\t%-8.1f\t%-5.1f%8.f\n", hop.RouteNo, hop.Addr, hop.Loss, hop.Snt, hop.Avg, hop.Best, hop.Wrst)
	//}

	//reportData := &api.Data{
	//	T:        t.T,
	//	Device:   t.Dev,
	//	Business: t.Biz,
	//	Bd:       t.BD,
	//	Bid:      t.BId,
	//	Region:   t.Region,
	//	Src:      "ali",
	//	Dst:      t.Ip,
	//	Dport:    t.Port,
	//	Avg:      float64(lastHop.Avg),
	//	Max:      float64(lastHop.Wrst),
	//	Min:      float64(lastHop.Best),
	//	LossRate: float64(lastHop.Loss),
	//}
	//
	api.Report(t, hops, pingReturn, opId)
	//if lastHop.Addr == "???" || lastHop.Loss > 5 {
	//	if err = api.Alert(t, hops); err != nil {
	//		log.GlobalLog.Errorf(err.Error())
	//		return
	//	}
	//}

	//if pingReturn.DropRate > 5 {
	//	if err = api.Alert(t, hops, pingReturn); err != nil {
	//		log.GlobalLog.Errorf(err.Error())
	//		return
	//	}
	//}
}

func main() {
	go alarm.Check()
	for {
		time.Sleep(5 * time.Minute)
		for _, target := range api.MonitorTargets() {
			go detect(target)
		}
	}
}
