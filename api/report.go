package api

import (
	"encoding/json"
	"github.com/enoch300/nt/mtr"
	"github.com/enoch300/nt/ping"
	"ip_detect/request"
	. "ip_detect/utils/log"
	"strconv"
)

type RequestIpaas struct {
	Database string          `json:"database"`
	Table    string          `json:"table"`
	Columns  []string        `json:"columns"`
	Values   [][]interface{} `json:"values"`
}

type ResponseIpaas struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type Data struct {
	T        string  `json:"t"`
	Device   string  `json:"device"`
	Business string  `json:"business"`
	Bd       string  `json:"bd"`
	Bid      string  `json:"bid"`
	Region   string  `json:"region"`
	Src      string  `json:"src"`
	Dst      string  `json:"dst"`
	Dport    string  `json:"dport"`
	Avg      float64 `json:"avg"`
	Max      float64 `json:"max"`
	Min      float64 `json:"min"`
	LossRate float64 `json:"lossRate"`
	Mtr      string  `json:"mtr"`
}

type MTR struct {
	T    string  `json:"t"`
	Id   string  `json:"id"`
	No   string  `json:"no"`
	Host string  `json:"host"`
	Loss float64 `json:"loss"`
	Snt  int     `json:"snt"`
	Avg  float64 `json:"avg"`
	Best float64 `json:"best"`
	Wrst float64 `json:"wrst"`
}

func Report(t *Targets, hops []mtr.Hop, pingReturn ping.PingReturn, opId string) {
	var values [][]interface{}
	value := []interface{}{t.T, opId, t.Dev, t.Biz, t.BD, t.BId, t.Region, "ali", t.Ip, t.Port, pingReturn.AvgTime.Seconds() * 1000, pingReturn.WrstTime.Seconds() * 1000, pingReturn.BestTime.Seconds() * 1000, pingReturn.DropRate}
	values = append(values, value)
	PushToIpaas("ipaas", "ip_detect", []string{"t", "id", "device", "business", "bd", "bid", "region", "src", "dst", "dport", "avg", "max", "min", "loss_rate"}, values)

	if len(hops) == 0 {
		return
	}

	var mtrValues [][]interface{}
	for _, h := range hops {
		value = []interface{}{t.T, opId, strconv.Itoa(h.RouteNo), h.Addr, float64(h.Loss), strconv.Itoa(h.Snt), float64(h.Avg), float64(h.Best), float64(h.Wrst)}
		mtrValues = append(mtrValues, value)
	}

	PushToIpaas("ipaas", "ip_detect_mtr", []string{"t", "id", "no", "host", "loss", "snt", "avg", "best", "wrst"}, mtrValues)
}

func PushToIpaas(db string, table string, columns []string, values [][]interface{}) {
	reportData := RequestIpaas{
		Database: db,
		Table:    table,
		Columns:  columns,
		Values:   values,
	}

	requestByte, err := json.Marshal(reportData)
	if err != nil {
		GlobalLog.Errorf("json.Marshal %v", err)
		return
	}

	respBody, httpCode, err := request.Post("https://ipaas.paigod.work/api/v1/ck", requestByte)
	if err != nil {
		GlobalLog.Errorf("ReportToIpaas %v", err)
		return
	}

	if httpCode != 200 {
		GlobalLog.Errorf("ReportToIpaas %v, httpCode: %v", respBody, httpCode)
		return
	}

	var responseIpaas ResponseIpaas
	if err = json.Unmarshal(respBody, &responseIpaas); err != nil {
		GlobalLog.Errorf("ReportToIpaas json.Unmarshal %v", err)
		return
	}

	if responseIpaas.Code != 0 {
		GlobalLog.Errorf("ReportToIpaas %v", responseIpaas)
		return
	}
	GlobalLog.Info("ReportToIpaas success tasks")
	//GlobalLog.Infof("ReportToIpaas success tasks >>> t: %v, src: %v, dst: %v, dport: %v", d.T, d.Src, d.Dst, d.Dport)
}
