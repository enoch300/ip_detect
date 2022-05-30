package main

import (
<<<<<<< HEAD
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
=======
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
	"xh_detect/ping"
	"xh_detect/utils"
	"xh_detect/utils/log"
	"xh_detect/utils/request"
)

const alertApi = "https://api.external.paigod.work/alert/report"
const ipaasApi = "https://ipaas.paigod.work/api/v1/alxh"

type AlertMsg struct {
	DeviceID string `json:"deviceID"`
	Msg      string `json:"msg"`
}

type Alert struct {
	Business string   `json:"business"`
	Data     AlertMsg `json:"data"`
}

func (a *Alert) Marshal() []byte {
	data, _ := json.Marshal(a)
	return data
}

type Item struct {
	Target string `json:"target"`
}

type Response struct {
	Code int    `json:"code"`
	Data []Item `json:"data"`
}

type Target struct {
	Id     string
	Region string
	Ip     string
	Port   string
}

func (t *Target) String() string {
	return fmt.Sprintf("%s:%s", t.Ip, t.Port)
}

func sendAlarm(mid string, msg string) {
	alert := Alert{
		Business: "aly",
		Data: AlertMsg{
			DeviceID: mid,
			Msg:      msg,
		},
	}
	response, httpcode, err := request.Post(alertApi, alert.Marshal())
	if err != nil {
		log.GlobalLog.Errorf("alert fail: %s, httpcode: %d, response: %s,", err, httpcode, response)
		return
	}
	log.GlobalLog.Errorf("alert success: %s, httpcode: %d, response: %s,", err, httpcode, response)
}

func detect(t Target, mid string, wg *sync.WaitGroup) {
	defer wg.Done()
	p := ping.NewPing("0.0.0.0", t.Ip, 3)
	if err := p.SendICMP(); err != nil {
		//fmt.Printf("Ping >>> srcIp: %v, dstIp: %v, %v\n", "0.0.0.0", dstAddr, err.Error())
		log.GlobalLog.Errorf("Ping >>> srcIp: %v, dstIp: %v, %v", "0.0.0.0", t, err.Error())
		return
	}

	if p.LossRate > 5 {
		log.GlobalLog.Errorf("[%s 丢包率超过5%%] 平均延时: %vms, 最大延时: %vms, 最小延时: %vms, 丢包率: %v%%", t.String(), p.AvgDelay, p.MaxDelay, p.MinDelay, p.LossRate)
		sendAlarm(mid, fmt.Sprintf("[%s 丢包率超过5%%] 平均延时: %vms, 最大延时: %vms, 最小延时: %vms, 丢包率: %v%%", t.String(), p.AvgDelay, p.MaxDelay, p.MinDelay, p.LossRate))
	}

	conn, err := net.DialTimeout("tcp", t.String(), time.Duration(3)*time.Second)
	if err != nil {
		//fmt.Printf("[%s 连接失败] 平均延时: %vms, 最大延时: %vms, 最小延时: %vms, 丢包率: %v%%\n", dstAddr.String(), p.AvgDelay, p.MaxDelay, p.MinDelay, p.LossRate)
		sendAlarm(mid, fmt.Sprintf("[%s 连接失败] 平均延时: %vms, 最大延时: %vms, 最小延时: %vms, 丢包率: %v%%", t.String(), p.AvgDelay, p.MaxDelay, p.MinDelay, p.LossRate))
	} else {
		conn.Close()
		//fmt.Printf("%s 连接成功\n", dstAddr.String())
		log.GlobalLog.Infof("%s 连接成功", t.String())
		//msgBuf = append(msgBuf, fmt.Sprintf("%s 连接成功", dstAddr.String()))
	}
}

func getTargets() (targets []Target, err error) {
	body, httpcode, err := request.Get(ipaasApi)
	if err != nil {
		return targets, err
	}

	if httpcode != 200 {
		return targets, fmt.Errorf("get ips fail, httpcode:%d, body: %s, err: %s", httpcode, string(body), err)
	}
	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		return targets, err
	}

	ips := strings.Split(r.Data[0].Target, "\n")
	for _, ipi := range ips {
		info := strings.Split(ipi, "|")
		if len(info) < 3 {
			return targets, fmt.Errorf("ip info unmarshal error, length < 3")
		}

		addr := strings.Split(strings.TrimSpace(info[2]), ":")
		if len(addr) < 2 {
			return targets, fmt.Errorf("ip address unmarshal error, length < 2")
		}

		targets = append(targets, Target{
			Id:     strings.TrimSpace(info[0]),
			Region: strings.TrimSpace(info[1]),
			Ip:     addr[0],
			Port:   addr[1],
		})
	}
	return targets, nil
}

func init() {
	log.NewLogger(3)
}

func main() {
	for {
		time.Sleep(time.Duration(5) * time.Minute)

		mid, err := utils.GetMachineId()
		if err != nil {
			log.GlobalLog.Errorf("get machineid err %s", err)
			continue
		}

		targets, err := getTargets()
		if err != nil {
			log.GlobalLog.Error(err)
			continue
		}

		wgMain := &sync.WaitGroup{}
		for _, target := range targets {
			wgMain.Add(1)
			go detect(target, mid, wgMain)
		}
		wgMain.Wait()
>>>>>>> 52c810ff9b823c0b781aeb745e80364c1b3b7a8b
	}
}
