package main

import (
	"motify_core_api/godep_libs/go-config"
	"motify_core_api/godep_libs/go-dconfig"
	"motify_core_api/godep_libs/mobapi_lib"
	"motify_core_api/godep_libs/mobapi_lib/logger"

	resourceSearchEngine "motify_core_api/godep_libs/mobapi_lib/_example/resources/searchengine"

	handlerHelloWorld "motify_core_api/godep_libs/mobapi_lib/_example/handlers/hello/world"
	handlerSearchGoogle "motify_core_api/godep_libs/mobapi_lib/_example/handlers/search/google"
)

const serviceName = "Example"

var (
	AppVersion  string // this value set by compiler
	GoVersion   string // this value set by compiler
	BuildDate   string // this value set by compiler
	GitDescribe string // this value set by compiler
)

func init() {
	service.AppVersion = AppVersion
	service.GoVersion = GoVersion
	service.BuildDate = BuildDate
	service.GitDescribe = GitDescribe

	config.RegisterBool("something-enabled", "Turn off/on something", false)
}

func main() {
	srvc := service.New(serviceName, "motify_core_api/godep_libs/mobapi_lib/_example/handlers")
	srvc.Init()

	se := &resourceSearchEngine.SearchEngine{}
	srvc.RegisterResources(se)

	srvc.MustRegisterHandlers(
		handlerHelloWorld.New(),
		handlerSearchGoogle.New(se),
	)

	dconfig.RegisterInt("exampleInt", "Example value to show how to use dconfig", 0, func(val int) {
		logger.Info(nil, "[dconfig] 'exampleInt' was changed to %d\n", val)
	})

	srvc.Run()
}
