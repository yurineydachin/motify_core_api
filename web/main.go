package main

import (
	"godep.lzd.co/service"
	"godep.lzd.co/service/config"
	"godep.lzd.co/service/logger"
	"godep.lzd.co/service/handlersmanager"

	resourceSearchEngine "motify_core_api/resources/searchengine"

	handlerHelloWorld "motify_core_api/handlers/hello/world"
	handlerSearchGoogle "motify_core_api/handlers/search/google"
)

const serviceName = "MotifyCoreAPI"

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
	srvc := service.New(serviceName, "motify_core_api/handlers")

	se := &resourceSearchEngine.SearchEngine{}
	srvc.RegisterResource(se)

	srvc.SetOptions(service.Options{HM: handlersmanager.New("motify_core_api/handlers")})
	srvc.MustRegisterHandlers(
		handlerHelloWorld.New(),
		handlerSearchGoogle.New(se),
	)

	err := srvc.Run()
	if err != nil {
		logger.Critical(nil, "Server stopped with error: %v", err)
	} else {
		logger.Info(nil, "Server stopped")
	}
}
