package main

import (
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
	}
}
