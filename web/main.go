package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

var conf *Config = nil

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("invalid args : %s [config]\n", os.Args[0])
		return
	}

	conf = new(Config)
	if e := conf.Load(os.Args[1]); e != nil {
		fmt.Printf("conf load %v, error %v\n", os.Args[1], e)
		return
	}

	rand.Seed(time.Now().UnixNano())

	// etcd init
	EtcdInit(conf)

	// web init
	StartHTTP(conf)

	// create pid file

	// init signals, block wait signals
	signal := InitSignal()
	HandleSignal(signal)
	fmt.Printf("web stop")
}
