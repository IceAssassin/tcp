package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// StartHTTP start listen http.
func StartHTTP(conf *Config) {
	// external
	httpServeMux := http.NewServeMux()
	httpServeMux.HandleFunc("/node/get", GetNode)

	// internal
	httpAdminServeMux := http.NewServeMux()
	httpAdminServeMux.HandleFunc("/admin/push", PushPrivate)

	for _, bind := range conf.PublicAddr {
		fmt.Printf("start http listen addr:%s", bind)
		go httpListen(httpServeMux, bind.String(), conf.HttpTimeout)
	}
	for _, bind := range conf.AdminAddr {
		fmt.Printf("start admin http listen addr:%s", bind)
		go httpListen(httpAdminServeMux, bind.String(), conf.HttpTimeout)
	}
}

func httpListen(mux *http.ServeMux, bind string, timeout int32) {
	server := &http.Server{Handler: mux, ReadTimeout: time.Duration(timeout), WriteTimeout: time.Duration(timeout)}
	server.SetKeepAlivesEnabled(false)
	l, err := net.Listen("tcp", bind)
	if err != nil {
		fmt.Printf("net.Listen(tcp, %s) error(%v)", bind, err)
		panic(err)
	}
	if err := server.Serve(l); err != nil {
		fmt.Printf("server.Serve() error(%v)", err)
		panic(err)
	}
}

// retWrite marshal the result and write to client(get).
func retWrite(w http.ResponseWriter, r *http.Request, res map[string]interface{}, start time.Time) {
	data, e := json.Marshal(res)
	if e != nil {
		fmt.Printf("json.Marshal(%v) error(%v)", res, e)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	if n, e := w.Write([]byte(data)); e != nil {
		fmt.Printf("w.Write(%s) error(%v)", data, e)
	} else {
		fmt.Printf("w.Write(%s) write %d bytes", data, n)
	}
	fmt.Printf("req: %s, res: %s, ip: %s, time: %fs", r.URL.String(), data, r.RemoteAddr, time.Now().Sub(start).Seconds())
}

// retPWrite marshal the result and write to client(post).
func retPWrite(w http.ResponseWriter, r *http.Request, res map[string]interface{}, body *string, start time.Time) {
	data, e := json.Marshal(res)
	if e != nil {
		fmt.Printf("json.Marshal(%v) error(%v)", res, e)
		return
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	dataStr := string(data)
	if n, e := w.Write([]byte(dataStr)); e != nil {
		fmt.Printf("w.Write(%s) error(%v)", dataStr, e)
	} else {
		fmt.Printf("w.Write(%s) write %d bytes", dataStr, n)
	}
	fmt.Printf("req: %s, post: %s, res: %s, ip: %s, time:%fs", r.URL.String(), *body, dataStr, r.RemoteAddr, time.Now().Sub(start).Seconds())
}
