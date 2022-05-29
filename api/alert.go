package api

import (
	"encoding/json"
	"fmt"
	"github.com/enoch300/nt/mtr"
	"github.com/enoch300/nt/ping"
	"ip_detect/request"
)

const ALERT_API = "http://47.97.252.215/alert/report"

type AlertBody struct {
	Business string `json:"business"`
	Data     string `json:"data"`
}

type ResBody struct {
	Code int `json:"code"`
}

func Alert(t *Targets, hops []mtr.Hop, pingReturn ping.PingReturn) error {
	msg := fmt.Sprintf("监测时间: %v\n业务名: %v\n业务ID: %v\n业务BD: %v\n触发策略: %v\n监控目标: 阿里 -> %v(%v)\n结果指标: 平均延时:%.2fms, 最大延时:%.2fms, 最小延时:%.2fms, 丢包率:%.2f%%",
		t.T, t.Biz, t.BId, t.BD, "丢包率>5%", t.Ip, t.Region, pingReturn.AvgTime.Seconds()*1000, pingReturn.WrstTime.Seconds()*1000, pingReturn.BestTime.Seconds()*1000, pingReturn.DropRate)

	if len(hops) > 0 {
		msg += fmt.Sprintf("%-5s%-5s%24s%10s%12s%12s%10s\n", "No", "Host", "Loss", "Snt", "Avg", "Best", "Wrst")
		for _, hop := range hops {
			if hop.Addr == "???" {
				msg += fmt.Sprintf("%-5v\t%-20v\t%-8.1f\t%-8d\t%-8.1f\t%-5.1f%8.f\n", hop.RouteNo, hop.Addr, hop.Loss, hop.Snt, hop.Avg, hop.Best, hop.Wrst)
				continue
			}
			msg += fmt.Sprintf("%-5v\t%-15v\t%-8.1f\t%-8d\t%-8.1f\t%-5.1f%8.f\n", hop.RouteNo, hop.Addr, hop.Loss, hop.Snt, hop.Avg, hop.Best, hop.Wrst)
		}
	}

	alertBody := AlertBody{Business: "zjzxdb", Data: msg}
	msgByte, err := json.Marshal(alertBody)
	if err != nil {
		return fmt.Errorf("send alert %v", err.Error())
	}

	response, httpcode, err := request.Post(ALERT_API, msgByte)
	var res ResBody
	json.Unmarshal(response, &res)
	if res.Code != 0 || httpcode != 200 {
		return fmt.Errorf("send alert resCode:%v, httpcode: %v", res.Code, httpcode)
	}

	return nil
}
