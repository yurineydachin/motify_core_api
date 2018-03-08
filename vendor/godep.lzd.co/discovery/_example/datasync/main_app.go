package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	etcdClientV3 "github.com/coreos/etcd/clientv3"
	"godep.lzd.co/discovery"
	"godep.lzd.co/discovery/_example/common"
	"godep.lzd.co/discovery/provider"
	"godep.lzd.co/discovery/provider/etcdV3"
	"godep.lzd.co/discovery/registrator"
	gologger "godep.lzd.co/go-logger"
	"godep.lzd.co/loggo"
)

var logger = initLogger()
var app *Application

type Config struct {
	Host             string
	Port             int
	EtcdEndpoints    common.StringList
	ExportedEntities common.StringList
}

type Application struct {
	logger discovery.ILogger
	config *Config

	etcdClientV3 *etcdClientV3.Client
	registrator  registrator.IRegistrator
	provider     provider.IProvider
}

func NewConfig() *Config {
	c := &Config{}
	c.initFlags()
	flag.Parse()

	return c
}

func NewApplication(config *Config, logger discovery.ILogger) (*Application, error) {
	app := &Application{
		config: config,
		logger: logger,
	}
	if err := app.init(); err != nil {
		return nil, err
	}

	return app, nil
}

func (app *Application) Shutdown() {
	app.logger.Warningf("Shutdown application")
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
	return nil
}

func (app *Application) Serve() error {
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

func (app *Application) init() error {
	if err := app.initDiscoveryClientV3(); err != nil {
		return fmt.Errorf("discovery client V3 init failed: %s", err)
	}
	if err := app.initDiscoveryRegistrator(); err != nil {
		return fmt.Errorf("discovery registrator init failed: %s", err)
	}
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
	app.logger.Infof("Registering entities: %s", app.config.ExportedEntities)
	app.provider = etcdV3.NewProvider(app.etcdClientV3, app.logger)
	p := registrator.AppRegistrationParams{
		// Service discovery info
		ServiceName: "example",
		RolloutType: "stable",
		Host:        app.config.Host,
		GRPCPort:    app.config.Port,
		// Admin info
		AdminPort: 666,
		Version: registrator.VersionInfo{
			AppVersion: "example version",
		},
		// Monitoring info
		MonitoringPort: 666,
		Venture:        "test_venture",
		Environment:    "test_env",

		ExportedEntities: app.config.ExportedEntities,
	}
	info, err := registrator.NewAppRegistrationInfo(p)
	if err != nil {
		return err
	}

	app.registrator = registrator.New(app.provider, info, app.logger)
	return nil
}

func initLogger() gologger.ILogger {
	logLevel := gologger.LevelDebug
	loggerFormatter := loggo.NewTextFormatter(":time: | :level: | :_package: | :_file: | :message: ")
	loggerHandler := loggo.NewStreamHandler(logLevel, loggerFormatter, os.Stdout)

	l := loggo.New("example", loggerHandler)
	l.AddProcessor(loggo.NewCalleeProcessor(0))
	return l
}

func handleStopSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	s := <-ch
	signal.Stop(ch)
	logger.Infof("%s recieved", s)

	switch s {
	case syscall.SIGINT, syscall.SIGTERM:
		logger.Info("stopping app")
		app.Shutdown()
	}
	os.Exit(0)
}

func (c *Config) initFlags() {
	flag.StringVar(&c.Host, "host", "0.0.0.0", "Web server host")
	flag.IntVar(&c.Port, "port", 8010, "Web server port number")
	flag.Var(&c.EtcdEndpoints, "etcd_endpoints", "Discovery etcdV2 endpoint urls separated by comma")
	flag.Var(&c.ExportedEntities, "exported_entities", "List of exported entities separated by comma")
}

func main() {
	config := NewConfig()
	var err error
	app, err = NewApplication(config, logger)
	if err != nil {
		logger.Critical(err)
		os.Exit(1)
	}

	go handleStopSignals()
	if err := app.Start(); err != nil {
		logger.Criticalf("Failed to start application: %s", err)
		os.Exit(1)
	}
	if err := app.Serve(); err != nil {
		logger.Critical(err)
		os.Exit(1)
	}
}
