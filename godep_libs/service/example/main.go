package main

import (
	"motify_core_api/godep_libs/service"
	"motify_core_api/godep_libs/service/config"
	"motify_core_api/godep_libs/service/dconfig"
	"motify_core_api/godep_libs/service/logger"

	resourceSearchEngine "motify_core_api/godep_libs/service/example/resources/searchengine"

	handlerHelloWorld "motify_core_api/godep_libs/service/example/handlers/hello/world"
	handlerSearchGoogle "motify_core_api/godep_libs/service/example/handlers/search/google"
)

const serviceName = "Example"

var (
	AppVersion  string // this value set by compiler
	GoVersion   string // this value set by compiler
	BuildDate   string // this value set by compiler
	GitRev      string // this value set by compiler
	GitHash     string // this value set by compiler
	GitDescribe string // this value set by compiler
)

func init() {
	service.AppVersion = AppVersion
	service.GoVersion = GoVersion
	service.BuildDate = BuildDate
	service.GitRev = GitRev
	service.GitHash = GitHash
	service.GitDescribe = GitDescribe

	config.RegisterBool("something-enabled", "Turn off/on something", false)
}

func main() {
	/* If you want custom transport cache
	if err := config.ParseAll(); err != nil {
		logger.Critical(nil, err.Error())
		os.Exit(1)
	}
	logger.Notice(nil, "app configuration: %s", config.String())

	proxy := http.DefaultClient

	monitor := monitoring.NewFromFlags(serviceName, proxy)

	cache, _ := aerocache.NewCacheResource(monitor, "")

	opts := service.ServiceOpts{
		Monitoring:     monitor,
		TransportCache: cache,
		ProxyClient:    proxy,
	}
	srvc := service.NewWithOpts(serviceName, "motify_core_api/godep_libs/service/example/handlers", opts)

	srvc.RegisterResource(cache)
	*/

	srvc := service.New(serviceName, "motify_core_api/godep_libs/service/example/handlers")

	se := &resourceSearchEngine.SearchEngine{}
	srvc.RegisterResource(se)

	srvc.MustRegisterHandlers(
		handlerHelloWorld.New(),
		handlerSearchGoogle.New(se),
	)

	dconfig.RegisterInt("exampleInt", "Example value to show how to use dconfig", 0, func(val int) {
		logger.Info(nil, "[dconfig] 'exampleInt' was changed to %d\n", val)
	})

	err := srvc.Run()
	if err != nil {
		logger.Critical(nil, "Server stopped with error: %v", err)
	} else {
		logger.Info(nil, "Server stopped")
	}
}
