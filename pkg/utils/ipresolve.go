package utils

import (
	"encoding/json"
	"yf_pkg/net/http"
)

type TaobaoIp struct {
	Code int        `json:"code"`
	Data TaobaoData `json:"data"`
}

type TaobaoData struct {
	City      string `json:"city_id"`
	Province  string `json:"region_id"`
	SCity     string `json:"city"`
	SProvince string `json:"region"`
}

func IpResolve(ip string) (data TaobaoData, e error) {
	params := make(map[string]string)
	params["ip"] = ip

	b, e := http.HttpGet("ip.taobao.com", "/service/getIpInfo.php", params, 10)
	if e != nil {
		return
	}
	var m TaobaoIp
	if err := json.Unmarshal(b, &m); err != nil {
		e = err
		return
	}
	data = TaobaoData{}
	if m.Code == 0 {
		data = m.Data
	}
	return
}
