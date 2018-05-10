package main

import (
	"os"
	"time"

	"motify_core_api/godep_libs/service"
	"motify_core_api/godep_libs/service/config"
	//"motify_core_api/godep_libs/service/dconfig"
	"motify_core_api/godep_libs/go-dconfig"
	//"motify_core_api/godep_libs/service/handlersmanager"
	mobConfig "motify_core_api/godep_libs/go-config"
	"motify_core_api/godep_libs/mobapi_lib/admin"
	"motify_core_api/godep_libs/mobapi_lib/handler"
	"motify_core_api/godep_libs/mobapi_lib/handlersmanager"
	mobLogger "motify_core_api/godep_libs/mobapi_lib/logger"
	"motify_core_api/godep_libs/mobapi_lib/sessionlogger"
	"motify_core_api/godep_libs/mobapi_lib/token"
	"motify_core_api/godep_libs/service/logger"

	wrapToken "motify_core_api/utils/token"

	coreApiAdapter "motify_core_api/resources/motify_core_api"

	"motify_core_api/handlers/integration_api/agent/details"
	"motify_core_api/handlers/integration_api/agent/list"
	"motify_core_api/handlers/integration_api/agent/sync"
	"motify_core_api/handlers/integration_api/agent/update"
	"motify_core_api/handlers/integration_api/employee/details"
	"motify_core_api/handlers/integration_api/employee/invite"
	"motify_core_api/handlers/integration_api/employee/list"
	"motify_core_api/handlers/integration_api/employee/sync"
	"motify_core_api/handlers/integration_api/employee/update"
	"motify_core_api/handlers/integration_api/employer/create"
	"motify_core_api/handlers/integration_api/employer/details"
	"motify_core_api/handlers/integration_api/employer/update"
	"motify_core_api/handlers/integration_api/payslip/create"
	"motify_core_api/handlers/integration_api/payslip/list"
	"motify_core_api/handlers/integration_api/user/approve/send"
	"motify_core_api/handlers/integration_api/user/approve/update"
	"motify_core_api/handlers/integration_api/user/login"
	"motify_core_api/handlers/integration_api/user/remind/reset"
	"motify_core_api/handlers/integration_api/user/remind/send"
	"motify_core_api/handlers/integration_api/user/signup"
	"motify_core_api/handlers/integration_api/user/update"
)

const serviceName = "MotifyIntegrationAPI"

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

	config.RegisterString("token-triple-des-key", "24-bit key for token DES encryption", "")
	config.RegisterString("token-salt", "8-bit salt for token DES encryption", "")

	config.RegisterUint("motify_core_api-timeout", "MotifyCoreAPI timeout, sec", 10)
}

func main() {
	srvc := service.New(serviceName, "motify_core_api/handlers/integration_api")
	if err := srvc.Init(); err != nil {
		logger.Critical(nil, "failed to init service: %v", err)
		os.Exit(1)
	}
	if err := mobConfig.ParseAll(); err != nil {
		logger.Error(nil, err.Error())
	}

	if err := initToken(); err != nil {
		logger.Critical(nil, "failed to init token encryption: %v", err)
		os.Exit(1)
	}

	venture, _ := config.GetString("venture")

	coreApiTimeout, _ := config.GetUint("motify_core_api-timeout")
	coreApi := coreApiAdapter.NewMotifyCoreAPIClient(srvc, time.Duration(coreApiTimeout)*time.Second)

	dconfm := dconfig.NewManager(serviceName, mobLogger.GetLoggerInstance())
	sessionLogger, err := sessionlogger.NewSessionLoggerFromFlags(dconfm)
	if err != nil {
		logger.Critical(nil, err.Error())
		os.Exit(1)
	}
	srvc.SetOptions(
		service.Options{
			HM:                   handlersmanager.New("motify_core_api/handlers/integration_api", wrapToken.ModelAgentUser),
			APIHandlerCallbacks:  handler.NewHTTPHandlerCallbacks(serviceName, service.AppVersion, "localhost", sessionLogger),
			SwaggerJSONCallbacks: admin.NewSwaggerJSONCallbacks(serviceName, venture),
		},
	)

	srvc.MustRegisterHandlers(
		user_login.New(coreApi),
		user_signup.New(coreApi),
		user_update.New(coreApi),
		user_remind_send.New(coreApi),
		user_remind_reset.New(coreApi),
		user_approve_send.New(coreApi),
		user_approve_update.New(coreApi),
		employee_details.New(coreApi),
		employee_list.New(coreApi),
		employee_invite.New(coreApi),
		employee_sync.New(coreApi),
		employee_update.New(coreApi),
		employer_create.New(coreApi),
		employer_details.New(coreApi),
		employer_update.New(coreApi),
		agent_list.New(coreApi),
		agent_sync.New(coreApi),
		agent_details.New(coreApi),
		agent_update.New(coreApi),
		payslip_create.New(coreApi),
		payslip_list.New(coreApi),
	)

	err = srvc.Run()
	if err != nil {
		logger.Critical(nil, "Server stopped with error: %v", err)
	} else {
		logger.Info(nil, "Server stopped")
	}
}

func initToken() error {
	key, _ := config.GetString("token-triple-des-key")
	salt, _ := config.GetString("token-salt")
	return token.InitTokenV1([]byte(key), []byte(salt))
}
