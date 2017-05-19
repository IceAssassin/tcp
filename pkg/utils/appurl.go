package utils

import (
	"encoding/json"
	"net/url"
)

const (
	CMD_JUMP_TO_GIFT  = "jump_to_gift"
	CMD_OPEN_GAME_WEB = "open_game_web"
)

func GetAppUrl(cmd string, data map[string]interface{}) (rurl string) {
	v := url.Values{}
	v.Set("cmd", cmd)
	if data != nil {
		sdata, e := json.Marshal(data)
		if e != nil {
			return
		}
		svalue := string(sdata)
		v.Set("data", svalue)
	}
	rurl = "/customAction?" + v.Encode()
	return
}
