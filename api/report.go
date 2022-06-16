package api

import (
	"encoding/json"
	. "ip_detect/utils/log"
	"ip_detect/utils/request"
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
	GlobalLog.Infof("ReportToIpaas success tasks")
}
