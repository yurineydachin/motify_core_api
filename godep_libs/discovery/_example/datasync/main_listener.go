package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	etcdClientV3 "github.com/coreos/etcd/clientv3"
	"motify_core_api/godep_libs/discovery/_example/common"
	"motify_core_api/godep_libs/discovery/provider"
	"motify_core_api/godep_libs/discovery/provider/etcdV3"
	"motify_core_api/godep_libs/discovery/registrator"
	gologger "motify_core_api/godep_libs/go-logger"
	"motify_core_api/godep_libs/loggo"
)

var logger = initLogger()
var etcdEndpoints common.StringList
var port int

func initLogger() gologger.ILogger {
	logLevel := gologger.LevelDebug
	loggerFormatter := loggo.NewTextFormatter(":time: | :level: | :_package: | :_file: | :message: ")
	loggerHandler := loggo.NewStreamHandler(logLevel, loggerFormatter, os.Stdout)

	l := loggo.New("example", loggerHandler)
	l.AddProcessor(loggo.NewCalleeProcessor(0))
	return l
}

func handleStopSignals(cancel context.CancelFunc) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	for s := range ch {
		logger.Infof("%s recieved", s)
		switch s {
		case syscall.SIGINT, syscall.SIGTERM:
			signal.Stop(ch)
			logger.Info("stopping app")
			cancel()
			os.Exit(0)
		}
	}
}

func serveHTTP() {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	logger.Infof("Listening at: %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Errorf("failed to run profiling tools on '%v': %v", port, err)
	}
}

func main() {
	flag.Var(&etcdEndpoints, "etcd_endpoints", "Discovery etcdV2 endpoint urls separated by comma")
	flag.IntVar(&port, "port", 8888, "Profiling port")
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	go handleStopSignals(cancel)
	go serveHTTP()

	c, err := etcdClientV3.New(etcdClientV3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logger.Critical(err)
		os.Exit(1)
	}

	p := etcdV3.NewProvider(c, logger)

	// recieve all notifications about entities endpoints
	filter := provider.KeyFilter{
		Namespace: provider.NamespaceExportedEntities,
	}
	ch := p.Watch(ctx, filter)
	for event := range ch {
		logger.Infof("Got event %s", event.Type)
		for _, kv := range event.KVs {
			entity, err := registrator.NewExportedEntityFromKey(kv.RawKey)
			if err != nil {
				logger.Errorf("%s", err)
				continue
			}
			logger.Infof("ExportedEntity: %+v", entity)
		}
	}
}
