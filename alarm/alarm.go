package alarm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"ip_detect/request"
	"ip_detect/utils/log"
	"net/http"
	"strings"
	"time"
)

//const ALERT_API = "http://47.97.252.215/alert/report"
const ALERT_API = "https://api.external.paigod.work/alert/report"

type Data struct {
	Device   string   `json:"device"`
	Business string   `json:"business"`
	Bd       string   `json:"bd"`
	Bid      string   `json:"bid"`
	Region   string   `json:"region"`
	Dst      string   `json:"dst"`
	Count    string   `json:"count"`
	Pingr    []string `json:"pingr"`
}

func (d *Data) AlarmString() (msg string) {
	msg = fmt.Sprintf("%v\n业务名: %v\n业务ID: %v\n业务BD: %v\n触发策略: %v\n监控目标: 阿里 -> %v(%v)\n\n历史结果: \n%v",
		time.Now().Format("2006-01-02 15:04:05"), d.Business, d.Bid, d.Bd, "过去10分钟, 丢包率>5%大于2次", d.Dst, d.Region, strings.Join(d.Pingr, "\n"))
	return
}

type CkResponse struct {
	Data []Data `json:"data"`
	Rows int    `json:"rows"`
}

type AlertBody struct {
	Business string `json:"business"`
	Data     string `json:"data"`
}

type ResBody struct {
	Code int `json:"code"`
}

func Alarm(msg string) error {
	alertBody := AlertBody{Business: "zjzxdb", Data: msg}
	msgByte, err := json.Marshal(alertBody)
	if err != nil {
		return fmt.Errorf("send alarm %v", err.Error())
	}

	response, httpcode, err := request.Post(ALERT_API, msgByte)
	var res ResBody
	json.Unmarshal(response, &res)
	if res.Code != 0 || httpcode != 200 {
		return fmt.Errorf("send alert resCode:%v, httpcode: %v", res.Code, httpcode)
	}
	return nil
}

func chQuery() (Data []Data, err error) {
	url := "https://ipaas.paigod.work/api/v1/ckquery"
	method := "POST"

	payload := strings.NewReader(`select device,business,bd,bid,region,dst,count(dst) c, groupArray(concat(toString(t),', ','avg:',toString(round(avg,2)),', ','max:',toString(round(max,2)),', ','min:',toString(round(min,2)),', ','loss:',toString(round(loss_rate,2)))) pingr from (select t,device,business,bd,bid,region,dst,avg,max,min,loss_rate from ipaas.ip_detect_all  where t > toStartOfMinute(toDateTime(now())-900) and loss_rate > 5 order by t)  group by device,business,bd,bid,region,dst having c > 2 FORMAT JSON`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "text/plain")
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	var ckResponse CkResponse
	if err = json.Unmarshal(body, &ckResponse); err != nil {
		return
	}
	return ckResponse.Data, nil
}

func Check() {
	for {
		time.Sleep(10 * time.Minute)
		lossHosts, err := chQuery()
		if err != nil {
			log.GlobalLog.Error(err.Error())
			continue
		}

		for _, loss := range lossHosts {
			if err = Alarm(loss.AlarmString()); err != nil {
				log.GlobalLog.Error(err)
				continue
			}
		}
	}
}
