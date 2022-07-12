package api

import (
	"encoding/json"
	"ip_detect/utils/logger"
	"ip_detect/utils/request"
)

type ResponseIpaas struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func PushToIpaas(data []byte) {
	respBody, httpCode, err := request.Post("https://ipaas.paigod.work/api/v1/ck/ipdetect", data)
	if err != nil {
		logger.Global.Errorf("PushToIpaas %v", err)
		return
	}

	if httpCode != 200 {
		logger.Global.Errorf("PushToIpaas %v, httpCode: %v", string(respBody), httpCode)
		return
	}

	var responseIpaas ResponseIpaas
	if err = json.Unmarshal(respBody, &responseIpaas); err != nil {
		logger.Global.Errorf("PushToIpaas json.Unmarshal %v", err)
		return
	}

	if responseIpaas.Code != 0 {
		logger.Global.Errorf("PushToIpaas %v", string(respBody))
		return
	}

	logger.Global.Infof("PushToIpaas success tasks")
}
