package ipip

import (
	"encoding/json"
	"fmt"
	"ip_detect/utils/request"
)

type Region struct {
	Region    string `json:"regin"` //新增自定义字段,非接口返回
	Accuracy  string `json:"accuracy"`
	Adcode    string `json:"adcode"`
	Areacode  string `json:"areacode"`
	Asnumber  string `json:"asnumber"`
	City      string `json:"city"`
	Continent string `json:"continent"`
	Country   string `json:"country"`
	District  string `json:"district"`
	Isp       string `json:"isp"`
	Latwgs    string `json:"latwgs"`
	Lngwgs    string `json:"lngwgs"`
	Owner     string `json:"owner"`
	Province  string `json:"province"`
	Radius    string `json:"radius"`
	Source    string `json:"source"`
	Timezone  string `json:"timezone"`
	Zipcode   string `json:"zipcode"`
}

type Response struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Ip   string `json:"ip"`
	Data Region `json:"data"`
}

func Query(ip string) (region Region, err error) {
	var header = make(map[string]string)
	header["token"] = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MTAwMCwicm9sZXMiOlsiZGV2b3BzIl0sImFjY291bnQiOiJpcHNlYXJjaCIsImV4cCI6MTY4NjY2ODkwMCwiaXNzIjoicHAuaW8iLCJuYmYiOjE2NTUxMzE5MDB9.dXnhT4nLu3xmaz9uwCrVBqG58TklGclfmxB0Vw7J7_c"
	body, httpcode, err := request.Get("https://ipaas.paigod.work/paiip/api/v1/address/ppio?ip="+ip, header)
	if err != nil || httpcode != 200 {
		return region, fmt.Errorf("get region by ip: %v httpcode: %v %v", ip, httpcode, err)
	}

	var resp Response
	if err = json.Unmarshal(body, &resp); err != nil {
		return region, err
	}

	return resp.Data, nil
}
