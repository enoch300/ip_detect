package sys

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"ip_detect/utils/logger"
	"ip_detect/utils/request"
	"net"
	"regexp"
	"strings"
	"time"
)

var Nic string

type MachineInfo struct {
	Nic  string `json:"nic"`
	Code int    `json:"code"`
}

type Ifi struct {
	Name string
	Ip   string
}

func MachineId() (id string, err error) {
	machineId, err := ioutil.ReadFile("/etc/machine-id")
	id = strings.TrimSpace(string(machineId))
	if err != nil {
		return "", err
	}

	return
}

func UpdateLocalEth() {
	err := LocalEth()
	if err != nil {
		logger.Global.Errorf(err.Error())
	}
	logger.Global.Infof("LocalEth: %s", Nic)

	for {
		time.Sleep(10 * time.Minute)
		err = LocalEth()
		if err != nil {
			continue
		}
		logger.Global.Infof("LocalEth: %s", Nic)
	}
}

func LocalEth() error {
	mid, err := MachineId()
	if err != nil {
		return fmt.Errorf("update config %s", err.Error())
	}

	header := make(map[string]string)
	header["Authorization"] = "Bearer eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjg3MiwidXNlcm5hbWUiOiJvcHMiLCJyb2xlIjoxNn0.E7-04n_bdl6_LtxoR0qecFEAxdkLG6ZaOR0n4DbwnOqe4SRgTOoyLcOiz6ZxRbjSO9PjyGvBxJ3tQCnO29dUiVn_HMaJvFLe0v-wuQrbjFaARjCxFaGqB93ViDwCNRcHINv4H7GX2PkMKYOfwFZ6033BOMMzbHIdYSrSwcVORpvfYVDcIBZHI7-zcf7qgkCyGJLF7X1z6NLKwlwuPyvgNyssJF_GZne0w1-nYYNgSIqlmcv4smEESz15ng9aQ5SdCaqlI4c7BvmSjb1OuzzGKDRGsu5TGYLV8U51KF3qNHyTfzZ_us0J0FY3QMmsTH6PNs09SVUHcp1Wt7EARttBQQ"
	body, httpcode, err := request.Get("https://internal.api.paigod.work/internal/device_info?deviceUUID="+mid, header)

	if err != nil {
		return fmt.Errorf("update config: %s", err.Error())
	}

	if httpcode != 200 {
		return fmt.Errorf("update config body: %s, httpcode: %d", strings.TrimSpace(string(body)), httpcode)
	}

	var m MachineInfo
	if err := json.Unmarshal(body, &m); err != nil {
		return fmt.Errorf("update config err: %s", err.Error())
	}

	if m.Code != 0 {
		return fmt.Errorf("update config m.Code != 0, code: %v", m.Code)
	}

	Nic = m.Nic
	return nil
}

// LocalIpv4 本机所有外网IP
func LocalIpv4() (ips []Ifi, err error) {
	reg := regexp.MustCompile(Nic)
	if reg == nil {
		return ips, fmt.Errorf("LocalIpv4 err: reg is nil")
	}

	inters, err := net.Interfaces()
	if err != nil {
		return ips, fmt.Errorf("LocalIpv4 err: %v", err.Error())
	}

	for _, inter := range inters {
		result := reg.FindAllStringSubmatch(inter.Name, -1)
		if len(result) > 0 && inter.Flags&net.FlagUp != 0 {
			// 获取网卡下所有的地址
			addrs, err := inter.Addrs()
			if err != nil || len(addrs) == 0 {
				continue
			}

			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsLinkLocalUnicast() {
					if ipnet.IP.To4() != nil {
						ips = append(ips, Ifi{Name: inter.Name, Ip: ipnet.IP.To4().String()})
					} else {
						ipv6 := fmt.Sprintf("%s", ipnet.IP.To16().String())
						ips = append(ips, Ifi{Name: inter.Name, Ip: ipv6})
					}
				}
			}
		}
	}

	return
}
