package main

import (
	"flag"

	"godep.lzd.co/discovery/_example/common"
)

type Config struct {
	ServiceName     string
	RolloutType     string
	Host            string
	Port            int
	AdminPort       int
	EtcdEndpoints   common.StringList
	ServiceToListen string
}

func NewConfig() *Config {
	c := &Config{}
	c.initFlags()
	flag.Parse()

	return c
}

func (c *Config) initFlags() {
	flag.StringVar(&c.ServiceName, "service_name", "", "Service name")
	flag.StringVar(&c.RolloutType, "rollout_type", "stable", "Rollout Type - (stable|unstable1|...|unstable20)")
	flag.StringVar(&c.Host, "host", "0.0.0.0", "Web server host")
	flag.IntVar(&c.Port, "port", 8010, "Web server port number")
	flag.IntVar(&c.AdminPort, "admin_port", 8011, "Web server admin port number")

	// Discovery
	flag.Var(&c.EtcdEndpoints, "etcd_endpoints", "Discovery etcdV2 endpoint urls separated by comma")
	flag.StringVar(&c.ServiceToListen, "try_service", "", "Service to include in balancing")
}
