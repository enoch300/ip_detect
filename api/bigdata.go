/*
* @Author: wangqilong
* @Description:
* @File: dataCenter
* @Date: 2021/6/22 3:25 下午
 */

package api

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"ip_detect/utils/logger"
	"ip_detect/utils/request"
	"ip_detect/utils/sys"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var client *http.Client

var manageURLMap = map[bool]string{
	//false: "https://api.paigod.work",
	false: "https://internal.api.paigod.work",
	true:  "http://api.test.paigod.work",
}

var bigDataURLMap = map[bool]string{
	false: "https://datachannel.painet.work",
	true:  "http://datachannel.test.painet.work",
}

var baseURLMap = map[string]interface{}{
	"bd":     bigDataURLMap, // 大数据平台
	"manage": manageURLMap,  // 管理平台
}

var tokenMap = map[bool]string{
	true:  "Bearer eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjMxMCwidXNlcm5hbWUiOiLmtL7kupHnrqHnkIYiLCJyb2xlIjoxfQ.llISTPP_23krINPF35VXQutU3eLH_m4K_XSqRECNRNLVZI8WyLyicloyOazM8Ojf4JpUL7yXvDxX8YBQygXRL7nLHgEykmb0l93MabFQvUtny0nMSBBdAdpCaGce_MUT_yuilLHaClK2m2hAjYsUyS3tQ-rKmgVCYeJi_XchLOws6ZGyR89HGFt3IyW7d_z5lRPSbcvH6iYtMr3aPEB9VltmBBX5apZNHAHPbxK5Bc_zq6t5diLHpE1S43avUX4knGWbJUjUeuzEDvFFcXFUgQ1aCJ72PJvHfpX4hTM_hvVBlwGvPPCaqMGWtvK0pMnUHKpVMIJDbHOndE35Zhh0Zw",
	false: "Bearer eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjg3MiwidXNlcm5hbWUiOiJvcHMiLCJyb2xlIjoxNn0.E7-04n_bdl6_LtxoR0qecFEAxdkLG6ZaOR0n4DbwnOqe4SRgTOoyLcOiz6ZxRbjSO9PjyGvBxJ3tQCnO29dUiVn_HMaJvFLe0v-wuQrbjFaARjCxFaGqB93ViDwCNRcHINv4H7GX2PkMKYOfwFZ6033BOMMzbHIdYSrSwcVORpvfYVDcIBZHI7-zcf7qgkCyGJLF7X1z6NLKwlwuPyvgNyssJF_GZne0w1-nYYNgSIqlmcv4smEESz15ng9aQ5SdCaqlI4c7BvmSjb1OuzzGKDRGsu5TGYLV8U51KF3qNHyTfzZ_us0J0FY3QMmsTH6PNs09SVUHcp1Wt7EARttBQQ",
}

var URLMap = map[string]string{
	"aliServer":                "/internal/ali_server",
	"reportIP":                 "/internal/report/ip",
	"fetchIPS":                 "/internal/ip_pool",
	"fetchAllIP":               "/internal/ip_pool_all",
	"bandwidthReport":          "/internal/report/bandwidth_test_result",
	"unionBandwidthReport":     "/internal/report/union_bandwidth_test_result",
	"iopsReport":               "/internal/report/iops",
	"diskTrackDetectionReport": "/internal/report/disk_track_detection",
	"cpuIndicatorReport":       "/internal/report/cpu",
	"memoryInfoReport":         "/internal/report/memory",
	"multiInfoReport":          "/internal/report/evaluation_result",
	"deviceInfo":               "/internal/device_info",
	"fetchQosIP":               "/internal/ip_pool_network_qos",
	"reportQosData":            "/v4/machine_bw_detection",
}

func jointUrl(path, category string, debug bool) string {
	rootURL := make(map[bool]string)
	switch category {
	case "bd":
		rootURL = baseURLMap["bd"].(map[bool]string)
	case "manage":
		rootURL = baseURLMap["manage"].(map[bool]string)
	default:
		rootURL = baseURLMap["manage"].(map[bool]string)
	}

	baseURL := ""
	if debug {
		baseURL = rootURL[true] + path
	} else {
		baseURL = rootURL[false] + path
	}
	logger.Global.Infof("HTTP category %s, base url %s", category, baseURL)
	return baseURL
}

func configQuery(path, category string, query map[string]string, debug bool) (string, error) {
	baseURL := jointUrl(path, category, debug)
	parsedUrl, err := url.ParseRequestURI(baseURL)
	if err != nil {
		logger.Global.Errorf("Can not parse url(%s), error: %v", baseURL, err)
		return "", err
	}

	params := url.Values{}
	for i, v := range query {
		params.Set(i, v)
	}
	parsedUrl.RawQuery = params.Encode()
	baseURL = parsedUrl.String()
	logger.Global.Infof("HTTP full url: %s", baseURL)
	return baseURL, nil
}

func configHeader(request *http.Request, debug bool) {
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Content-Encoding", "gzip")
	if debug {
		request.Header.Add("Authorization", tokenMap[true])
	} else {
		request.Header.Add("Authorization", tokenMap[false])
	}
}

func HttpRequest(method string, getUrl string, body io.Reader, debug bool) (*http.Response, error) {
	r, err := http.NewRequest(method, getUrl, body)
	transport := &http.Transport{DisableKeepAlives: true, MaxConnsPerHost: 10}
	client = &http.Client{Transport: transport, Timeout: time.Second * 120}
	if err != nil {
		logger.Global.Errorf("HTTP new request failed, url: %s, error: %v", getUrl, err)
		return nil, err
	}

	configHeader(r, debug)

	response, err := client.Do(r)

	if err != nil {
		logger.Global.Errorf("HTTP %s request failed, request: %v, err: %v", method, r, err)
		return nil, err
	}
	return response, nil
}

func fetchData(path, category string, query map[string]string, debug bool) ([]byte, error) {
	fullUrl, err := configQuery(path, category, query, debug)
	if err != nil {
		return nil, err
	}

	response, err := HttpRequest(http.MethodGet, fullUrl, nil, debug)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	logger.Global.Debugf("HTTP response status %v", response.Status)

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP response error: %v", response.Status)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger.Global.Errorf("HTTP GET response error: %v", err)
	}

	return body, nil
}

type ResInfo struct {
	Data  map[string][]string `json:"data"`
	Total float64             `json:"total"`
	Code  int64               `json:"code"`
}

func RequestIPPool() (ipPool map[string][]string, err error) {
	var resInfo ResInfo
	urlPath := URLMap["fetchQosIP"]
	id, _ := sys.MachineId()
	query := map[string]string{"deviceUUID": id}
	response, err := fetchData(urlPath, "manage", query, false)
	if err != nil {
		return ipPool, err
	}

	_ = json.Unmarshal(response, &resInfo)
	return resInfo.Data, err
}

func SignV2(machineId string, timestamp int64, data []byte) string {
	buf := make([]byte, 0)
	buffer := bytes.NewBuffer(buf)
	buffer.WriteString(machineId)
	buffer.WriteString(strconv.FormatInt(timestamp, 10))
	size := len(data)
	if size > 16 {
		size = 16
	}
	buffer.Write(data[:size])
	md5Str := md5.Sum(buffer.Bytes())
	return fmt.Sprintf("%x", md5Str)
}

func PushToBigData(reportData []byte) {
	machineId, _ := sys.MachineId()
	timestamp := time.Now().Unix()
	sign := SignV2(machineId, timestamp, reportData)
	urlPath := fmt.Sprintf("%s?machine_id=%s&t=%v&sign=%s", "http://datachannel.test.painet.work/iaas", machineId, timestamp, sign)
	respData, httpCode, err := request.Post(urlPath, reportData)
	if err != nil {
		logger.Global.Errorf("PushToBigData fail, url: %s, error: %v", urlPath, err)
		return
	}

	switch httpCode {
	case 500:
		for i := 1; i <= 3; i++ {
			body, code, _ := request.Post(urlPath, reportData)
			if code == 202 {
				logger.Global.Infof("PushToBigData success, url: %v, code: %v, respBody: %v, reportData: %v", urlPath, code, string(body), string(reportData))
				break
			}

			if i == 3 {
				logger.Global.Errorf("PushToBigData and retry error, url: %s, code: %v", urlPath, httpCode)
			}
		}
	case 400:
		logger.Global.Errorf("PushToBigData params error, url: %s, code: %v, respBody: %v, reportData: %v", urlPath, httpCode, string(respData), string(reportData))
	case 202:
		logger.Global.Infof("PushToBigData success, url: %s, code: %v, respBody: %v", urlPath, httpCode, string(respData))
	default:
		logger.Global.Errorf("PushToBigData fail, url: %s, code: %v, respBody: %v, reportData: %v", urlPath, httpCode, string(respData), string(reportData))
	}
	return
}

func ReportAli(reportData []byte) {
	machineId, _ := sys.MachineId()
	timestamp := time.Now().Unix()
	sign := SignV2(machineId, timestamp, reportData)
	urlPath := fmt.Sprintf("%s?machine_id=%s&t=%v&sign=%s", "http://datachannel.test.painet.work/v4/line_net_check", machineId, timestamp, sign)
	respData, httpCode, err := request.Post(urlPath, reportData)
	if err != nil {
		logger.Global.Errorf("post fail, url: %s, error: %v", urlPath, err)
		return
	}

	switch httpCode {
	case 500:
		for i := 1; i <= 3; i++ {
			resp, code, _ := request.Post(urlPath, reportData)
			if code == 202 {
				logger.Global.Infof("post success, url: %v, code: %v, resp: %v", urlPath, code, resp)
				break
			}

			if i == 3 {
				logger.Global.Errorf("post and retry error, url: %s, code: %v", urlPath, httpCode)
			}
		}
	case 400:
		logger.Global.Errorf("post params error, url: %s, code: %v", urlPath, httpCode)
	case 202:
		logger.Global.Infof("post success, url: %s, code: %v, resp: %v", urlPath, httpCode, respData)
	default:
		logger.Global.Errorf("post fail, url: %s, code: %v", urlPath, httpCode)
	}

	return
}

func FetchAliServer(url string) (ip string, err error) {
	type RespBody struct {
		Address string `json:"address"`
	}
	var respBody RespBody

	client = &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Global.Errorf("http.NewRequest: %v", err.Error())
		return ip, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjY2MCwidXNlcm5hbWUiOiJpbnRlcm5hbCIsInJvbGUiOjh9.X3Fm9pTbxrA2gtbe-xXFGbd02BnHLaxXaqUWd3UzSCLX_jFMdt7f1nuvT_IumFIhncYB0y9ZQLEjrvt2MlIWsIbd6lDUlxqHPUBJ_3e9fYzc3MkoYJ3orsh1X90oDjF38FBV1Q7urtv2k0SFLlyAsj1nv45fAPSxBlrEOce3kYw2uc_oiIyBYf4vvb8t96itoeFKhqwSFwM5WtVfLBgHXgVA_z7vvDPMFLag274-ncSWRK9AYboYd-mesHV82bvVRZpp_iL1wN8jG34S8VUJH7g4CBarVf3pDVpLEBsjUMkA5T8aPAd51djPklBt8Y1vPXROUdR7-LTtwa4tnbNzZA")
	resp, err := client.Do(req)
	if err != nil {
		logger.Global.Errorf("client.Do: %v", err.Error())
		return ip, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ip, fmt.Errorf("status code: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(data, &respBody)
	return respBody.Address, nil
}
