package main

import (
	"im/pkg/yaml"
)

type Config struct {
	AdminAddr  []yaml.Address "admin_addr"
	PublicAddr []yaml.Address "public_addr"
	PprofAddr  []yaml.Address "pprof_addr"
	Log        struct {
		Dir     string "dir"
		Level   string "level"
		BufSize int32  "buf_size"
	} "log"

	EtcdAddr    yaml.Address "etcd_addr"
	HttpTimeout int32        "http_timeout"
	MaxProc     int32        "max_proc"
	PidFile     string       "pid_file"
}

func (c *Config) Load(path string) error {
	return yaml.Load(c, path)
}
