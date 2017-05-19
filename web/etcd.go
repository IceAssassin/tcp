package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/pkg/errors"
	"im/pkg/util"
	"log"
	"os"
	"sort"
	"sync"
	"time"
)

type Server struct {
	Info util.ServerInfo // 服务器信息
}

// all server
type server_pool struct {
	services    map[string]*Server
	server_list []*Server
	client      *clientv3.Client
	mu          sync.RWMutex
}

var (
	Default_pool server_pool
	once         sync.Once
)

func EtcdInit(conf *Config) {
	once.Do(func() { Default_pool.init(conf) })
}

type ServerList []*Server

func (li ServerList) Less(i, j int) bool {
	return li[i].Info.ConnNum < li[j].Info.ConnNum
}

func (li ServerList) Len() int {
	return len(li)
}

func (li ServerList) Swap(i, j int) {
	li[i], li[j] = li[j], li[i]
}

func SortServer() {
	// 定时根据 Node 负载来进行排序
	for {
		Default_pool.mu.Lock()

		server_list := make([]*Server, 0)
		for _, s := range Default_pool.services {
			server_list = append(server_list, s)
		}

		sort.Sort(ServerList(server_list))
		Default_pool.server_list = server_list

		Default_pool.mu.Unlock()

		time.Sleep(10 * time.Second)
	}
}
func (p *server_pool) init(conf *Config) {
	// init etcd client
	cfg := clientv3.Config{
		Endpoints: []string{conf.EtcdAddr.String(), "127.0.0.1:2379"}, // c.StringSlice("etcd-hosts"),
	}
	cli, err := clientv3.New(cfg)
	if err != nil {
		log.Panic(err)
		os.Exit(-1)
	}
	// TODO 判断连接成功与否，与连接的释放
	defer cli.Close()
	go SortServer()

	p.client = cli
	// init
	p.services = make(map[string]*Server)
	fmt.Println("watching service: ／node")
	rch := cli.Watch(context.Background(), "node/", clientv3.WithPrefix())
	fmt.Println(000)

	for {
		select {
		case xx := <-rch:
			for _, ev := range xx.Events { // 只处理最新的请求？
				switch ev.Type {
				case mvccpb.PUT:
					p.add_server(string(ev.Kv.Key), string(ev.Kv.Value))
				case mvccpb.DELETE:
					p.remove_server(string(ev.Kv.Key))
				}
			}
		}

	}
}

// 添加服务器
func (p *server_pool) add_server(key string, value string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	ser := util.ServerInfo{}
	if e := json.Unmarshal([]byte(value), &ser); e != nil {
		return fmt.Errorf("unmashal json error. key %s, value %s", key, value)
	}

	s, exist := p.services[key]
	if !exist {
		fmt.Println("new node ", key)
		server := Server{
			Info: ser,
		}
		p.services[key] = &server
	} else {
		s.Info = ser
		fmt.Println("update server", key, s.Info)
	}

	return nil
}

// 删除删除服务器
func (p *server_pool) remove_server(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Println("remove server", key)
	delete(p.services, key)
}

// get an available node server
func (p *server_pool) GetServer() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(Default_pool.server_list) > 0 {
		return Default_pool.server_list[0].Info.PublicIP, nil
	}

	return "", errors.New("no available server")
}
