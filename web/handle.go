package main

import (
	"net/http"
	"time"
)

/*
获取可用的node节点
*/
func GetNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	ser, err := Default_pool.GetServer()
	if err != nil {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	res := map[string]interface{}{"ret": "OK", "msg": "ok", "node": ser}
	retWrite(w, r, res, time.Now())
}

/*
推送消息测试
*/
func PushPrivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	res := map[string]interface{}{"ret": "OK", "msg": "ok", "node": "127.0.0.1:8080"}
	retWrite(w, r, res, time.Now())
}
