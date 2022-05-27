package api

import (
	"encoding/json"
	"fmt"
	"github.com/enoch300/nt/mtr"
	"ip_detect/request"
	"ip_detect/utils/log"
)

const ALERT_API = "http://47.97.252.215/alert/report"

type AlertBody struct {
	Business string `json:"business"`
	Data     string `json:"data"`
}

type ResBody struct {
	Code int `json:"code"`
}

func Alert(t Targets, hop mtr.Hop) {
	msg := fmt.Sprintf("监测时间:%v\n业务名:%v\n业务ID:%v\n业务BD:%v\n触发策略:%v\n监控目标:阿里 -> %v\n结果指标: 平均延时:%.2f, 最大延时:%.2f, 最小延时:%.2f, 丢包率:%v%%\n",
		t.T, t.Business, t.BusinessID, t.BusinessOwner, "丢包率>5%", t.Ip, hop.Avg, hop.Wrst, hop.Best, hop.Loss)
	alertBody := AlertBody{Business: "zjzxdb", Data: msg}
	msgByte, err := json.Marshal(alertBody)
	if err != nil {
		log.GlobalLog.Errorf("send alert %v", err.Error())
		return
	}
	response, httpcode, err := request.Post(ALERT_API, msgByte)
	var res ResBody
	json.Unmarshal(response, &res)
	if res.Code != 0 || httpcode != 200 {
		log.GlobalLog.Errorf("send alert resCode:%v, httpcode: %v", res.Code, httpcode)
		return
	}
	log.GlobalLog.Infof("send alert success")
}
