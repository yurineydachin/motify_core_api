package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"time"

	etcdClient "github.com/coreos/etcd/client"
	etcdClientV3 "github.com/coreos/etcd/clientv3"
	"godep.lzd.co/discovery"
	"godep.lzd.co/discovery/balancer"
	"godep.lzd.co/discovery/provider"
	"godep.lzd.co/discovery/provider/etcdV3"
	"godep.lzd.co/discovery/registrator"
)

type Application struct {
	logger discovery.ILogger
	config *Config

	etcdClientV2 etcdClient.Client
	etcdClientV3 *etcdClientV3.Client
	registrator  registrator.IRegistrator
	provider     provider.IProvider
	balancer     balancer.IRolloutBalancer

	done chan bool
}

func NewApplication(config *Config, logger discovery.ILogger) (*Application, error) {
	app := &Application{
		config: config,
		logger: logger,
		done:   make(chan bool),
	}
	if err := app.init(); err != nil {
		return nil, err
	}

	return app, nil
}

func (app *Application) Shutdown() {
	app.logger.Warningf("Shutdown application")
	close(app.done)

	if app.balancer != nil {
		app.balancer.Stop()
	}
	app.registrator.Unregister()
}

func (app *Application) Start() error {
	if err := app.registrator.Register(); err != nil {
		return err
	}

	if err := app.registrator.EnableDiscovery(); err != nil {
		return err
	}

	app.logger.Infof("application registered")
	if app.config.ServiceToListen != "" {
		go app.Balancing()
	}
	return nil
}

func (app *Application) Serve() error {
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		n, err := w.Write(body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		app.logger.Infof("request served, %d bytes written", n)
	})
	http.HandleFunc("/disable", func(w http.ResponseWriter, r *http.Request) {
		app.registrator.DisableDiscovery()
	})
	http.HandleFunc("/enable", func(w http.ResponseWriter, r *http.Request) {
		if err := app.registrator.EnableDiscovery(); err != nil {
			app.logger.Errorf("EnableDiscovery: %s", err)
		}
	})

	addr := net.JoinHostPort(app.config.Host, fmt.Sprintf("%d", app.config.Port))
	app.logger.Infof("listening at %s", addr)
	return http.ListenAndServe(addr, nil)
}

func (app *Application) Balancing() {
	app.logger.Infof("balancing started")
	requestNumber := 0
	for {
		select {
		case <-app.done:
			app.logger.Infof("exit balancing")
			return
		case <-time.After(10 * time.Second):
			segregationID := fmt.Sprintf("%d", rand.Intn(1001))
			app.logger.Debugf("try calling service, segregationID: %q", segregationID)

			opts := balancer.NextOptions{
				SegregationID: segregationID,
			}
			addr, err := app.balancer.Next(opts)
			if err != nil {
				app.logger.Errorf("error getting next instance: %s", err)
				continue
			}

			app.logger.Infof("calling %s", addr)
			if err := app.doRequest(addr, requestNumber); err != nil {
				app.logger.Errorf("doRequest error: %s", err)
			}

			requestNumber++
		}
	}
}

func (app *Application) doRequest(addr string, n int) error {
	b := bytes.NewBufferString(fmt.Sprintf("echo %d", n))
	resp, err := http.Post(addr+"/echo", "text/plain", b)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("body read error: %s", err)
	}
	defer resp.Body.Close()

	app.logger.Infof("Got: %s", string(body))
	return nil
}

func (app *Application) init() error {
	if err := app.initDiscoveryClientV2(); err != nil {
		return fmt.Errorf("discovery client V2 init failed: %s", err)
	}

	if err := app.initDiscoveryClientV3(); err != nil {
		return fmt.Errorf("discovery client V3 init failed: %s", err)
	}

	if err := app.initDiscoveryRegistrator(); err != nil {
		return fmt.Errorf("discovery registrator init failed: %s", err)
	}

	app.initBalancer()
	return nil
}

func (app *Application) initDiscoveryClientV2() error {
	if len(app.config.EtcdEndpoints) == 0 {
		return fmt.Errorf("Wrong discovery endpoints list")
	}
	config := etcdClient.Config{
		Endpoints:               app.config.EtcdEndpoints,
		HeaderTimeoutPerRequest: 5 * time.Second,
	}

	c, err := etcdClient.New(config)
	if err != nil {
		return err
	}
	app.etcdClientV2 = c
	return nil
}

func (app *Application) initDiscoveryClientV3() error {
	config := etcdClientV3.Config{
		Endpoints:   app.config.EtcdEndpoints,
		DialTimeout: 5 * time.Second,
	}
	c, err := etcdClientV3.New(config)
	if err != nil {
		return err
	}
	app.etcdClientV3 = c
	return nil
}

func (app *Application) initDiscoveryRegistrator() error {
	info := registrator.AppRegistrationParams{
		// Service discovery info
		ServiceName: app.config.ServiceName,
		RolloutType: app.config.RolloutType,
		Host:        app.config.Host,
		HTTPPort:    app.config.Port,
		// Admin info
		AdminPort: app.config.AdminPort,
		Version: registrator.VersionInfo{
			AppVersion: "example version",
		},
		// Monitoring info
		MonitoringPort: app.config.AdminPort,
		Venture:        "test_venture",
		Environment:    "test_env",
	}
	app.provider = etcdV3.NewProvider(app.etcdClientV3, app.logger)

	r, err := registrator.NewV2V3(info, app.provider, app.etcdClientV2, app.logger)
	if err != nil {
		return err
	}
	app.registrator = r
	return nil
}

func (app *Application) initBalancer() {
	if app.config.ServiceToListen == "" {
		return
	}

	stableOpts := balancer.FallbackBalancerEtcd2Options{
		ServiceName: app.config.ServiceToListen,
		Venture:     "test_venture",
		Environment: "test_env",
	}
	stableBalancer := balancer.NewFallbackBalancerEtcd2(app.provider, app.etcdClientV2, app.logger, stableOpts)

	rolloutOpts := balancer.RolloutBalancerOptions{
		LoadBalancerOptions: balancer.LoadBalancerOptions{
			ServiceName: app.config.ServiceToListen,
		},
		FallbackBalancer: stableBalancer,
	}
	// RolloutWatcher should be initialized and saved on Application level
	// when you use cache with sharding by rollout type
	w := balancer.NewRolloutWatcher(app.provider, app.logger)
	app.balancer = balancer.NewRolloutBalancer(app.provider, w, app.logger, rolloutOpts)
}
