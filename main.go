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

const alertApi = "http://47.97.252.215/alert/report"
const ipaasApi = "https://ipaas.paigod.work/api/v1/alxh"

type Addr struct {
	IP   string
	PORT string
}

func (a *Addr) String() string {
	return fmt.Sprintf("%s:%s", a.IP, a.PORT)
}

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

type DateItem struct {
	Target string `json:"target"`
}

type Response struct {
	Code int        `json:"code"`
	Data []DateItem `json:"data"`
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

func detect(dstAddr Addr, mid string, wg *sync.WaitGroup) {
	defer wg.Done()
	p := ping.NewPing("0.0.0.0", dstAddr.IP, 3)
	if err := p.SendICMP(); err != nil {
		log.GlobalLog.Errorf("Ping >>> srcIp: %v, dstIp: %v, %v", "0.0.0.0", dstAddr, err.Error())
		return
	}

	conn, err := net.DialTimeout("tcp", dstAddr.String(), time.Duration(3)*time.Second)
	if err != nil {
		sendAlarm(mid, fmt.Sprintf("[%s 连接失败] 平均延时: %vms, 最大延时: %vms, 最小延时: %vms, 丢包率: %v%%", dstAddr.String(), p.AvgDelay, p.MaxDelay, p.MinDelay, p.LossRate))
	} else {
		conn.Close()
		log.GlobalLog.Infof("%s 连接成功", dstAddr.String())
		//msgBuf = append(msgBuf, fmt.Sprintf("%s 连接成功", dstAddr.String()))
	}
}

func getIps() (ips []string, err error) {
	body, httpcode, err := request.Get(ipaasApi)
	if err != nil {
		return ips, err
	}

	if httpcode != 200 {
		return ips, fmt.Errorf("get ips fail, httpcode:%d, body: %s, err: %s", httpcode, string(body), err)
	}
	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		return ips, err
	}
	ips = strings.Split(r.Data[0].Target, "\n")
	return ips, nil
}

func init() {
	log.NewLogger(3)
}

func main() {
	for {
		time.Sleep(time.Minute)

		mid, err := utils.GetMachineId()
		if err != nil {
			log.GlobalLog.Errorf("get machineid err %s", err)
			continue
		}

		iplist, err := getIps()
		if err != nil {
			log.GlobalLog.Error(err)
			continue
		}

		wgMain := &sync.WaitGroup{}
		for _, dstAddr := range iplist {
			wgMain.Add(1)
			ip := strings.Split(dstAddr, ":")[0]
			port := strings.Split(dstAddr, ":")[1]
			go detect(Addr{IP: ip, PORT: port}, mid, wgMain)
		}
		wgMain.Wait()
	}
}
