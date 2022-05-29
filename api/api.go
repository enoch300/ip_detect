package api

import (
	"encoding/json"
	"ip_detect/request"
	"ip_detect/utils/log"
	"net"
	"strings"
)

var API = map[string]string{
	"阿里星火": "https://ipaas.paigod.work/api/v1/alxh",
	"字节跳动": "https://ipaas.paigod.work/api/v1/zjtd",
}

type Response struct {
	Code int `json:"code"`
	Data []struct {
		Target string `json:"target"`
	}
	Msg string `json:"msg"`
}

type Targets struct {
	T      string
	Dev    string
	Biz    string
	BId    string
	BD     string
	Region string
	Ip     string
	Port   string
}

func MonitorTargets() (targets []*Targets) {
	for bn, url := range API {
		body, httpcode, err := request.Get(url, nil)
		if err != nil {
			log.GlobalLog.Error(err.Error())
			return
		}

		if httpcode != 200 {
			log.GlobalLog.Errorf("get monitor targets httpcode %v", httpcode)
			return
		}

		var res Response
		if err := json.Unmarshal(body, &res); err != nil {
			log.GlobalLog.Errorf("json.Unmarshal %v", httpcode)
			return
		}

		if res.Code != 0 {
			log.GlobalLog.Errorf("get monitor targets code %v", httpcode)
			return
		}

		for _, t := range res.Data {
			lines := strings.Split(t.Target, "\n")
			for _, line := range lines {
				strs := strings.Split(line, "|")
				if len(strs) < 4 {
					continue
				}

				ip, port, err := net.SplitHostPort(strs[2])
				if err != nil {
					log.GlobalLog.Errorf("ip info err: %v", httpcode)
					continue
				}

				var device string
				if strings.Contains(strs[1], "交换机") {
					device = "switch"
				} else {
					device = "server"
				}

				targets = append(targets, &Targets{
					Dev:    device,
					Biz:    bn,
					BId:    strs[0],
					Region: strs[1],
					Ip:     ip,
					Port:   port,
					BD:     strs[3],
				})
			}
		}
	}
	return
}
