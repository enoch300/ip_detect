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
	//"kuaishou": "http://ippool.paigod.work/peers?machineid=8e6a2db3d4e46c0c02ba30322fc994a0&appid=kuaishou&network=tcp",
}

type ResponseIpass struct {
	Code int `json:"code"`
	Data []struct {
		Target string `json:"target"`
	}
	Msg string `json:"msg"`
}

type ResponseIppool struct {
	Success bool         `json:"success"`
	Code    int          `json:"code"`
	Msg     string       `json:"msg"`
	Data    []IppoolData `json:"data"`
}

type IppoolData struct {
	Mid       string `json:"mid"`
	Appid     string `json:"appid"`
	InnerIP   string `json:"inner_ip"`
	InnerPort string `json:"inner_port"`
	OuterIP   string `json:"outer_ip"`
	OuterPort string `json:"outer_port"`
	Province  string `json:"province"`
	Isp       string `json:"isp"`
}

type Target struct {
	Dev         string `json:"dev"`
	Mid         string `json:"mid"`
	Biz         string `json:"biz"`
	BId         string `json:"bid"`
	BD          string `json:"bd"`
	Region      string `json:"region"`
	OuterIp     string `json:"outer_ip"`
	OuterPort   string `json:"outer_port"`
	DoPing      bool   `json:"do_ping"`
	DoMtr       bool   `json:"do_mtr"`
	DoCheckPort bool   `json:"do_check_port"`
}

func Targets() (targets []*Target) {
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

		if bn == "kuaishou" {
			var resIppool ResponseIppool
			if err := json.Unmarshal(body, &resIppool); err != nil {
				log.GlobalLog.Errorf("json.Unmarshal %v", httpcode)
				return
			}

			if resIppool.Code != 0 {
				log.GlobalLog.Errorf("get monitor targets code %v", httpcode)
				return
			}

			for _, t := range resIppool.Data {
				targets = append(targets, &Target{
					Dev:       "server",
					Mid:       t.Mid,
					Biz:       "kuaishou",
					BId:       "",
					Region:    "",
					OuterIp:   t.OuterIP,
					OuterPort: t.OuterPort,
					BD:        "",
				})
			}
		} else {
			var resIpass ResponseIpass
			if err := json.Unmarshal(body, &resIpass); err != nil {
				log.GlobalLog.Errorf("json.Unmarshal %v", httpcode)
				return
			}

			if resIpass.Code != 0 {
				log.GlobalLog.Errorf("get monitor targets code %v", httpcode)
				return
			}

			for _, t := range resIpass.Data {
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

					targets = append(targets, &Target{
						Dev:       device,
						Biz:       bn,
						BId:       strs[0],
						Region:    strs[1],
						OuterIp:   ip,
						OuterPort: port,
						BD:        strs[3],
					})
				}
			}
		}
	}

	return
}
