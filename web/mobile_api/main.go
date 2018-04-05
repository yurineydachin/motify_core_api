package main

import (
	"os"
	"time"

	"godep.lzd.co/service"
	"godep.lzd.co/service/config"
	//"godep.lzd.co/service/dconfig"
	"godep.lzd.co/go-dconfig"
	//"godep.lzd.co/service/handlersmanager"
	mobConfig "godep.lzd.co/go-config"
	"godep.lzd.co/mobapi_lib/handler"
	"godep.lzd.co/mobapi_lib/handlersmanager"
	mobLogger "godep.lzd.co/mobapi_lib/logger"
	"godep.lzd.co/mobapi_lib/sessionlogger"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"

	wrapToken "motify_core_api/utils/token"

	"motify_core_api/handlers/mobile_api/employer/adduser"
	"motify_core_api/handlers/mobile_api/employer/details"
	"motify_core_api/handlers/mobile_api/employer/list"
	"motify_core_api/handlers/mobile_api/payslip/details"
	"motify_core_api/handlers/mobile_api/payslip/list"
	"motify_core_api/handlers/mobile_api/user/login"
	"motify_core_api/handlers/mobile_api/user/signup"
	"motify_core_api/handlers/mobile_api/user/update"
)

const serviceName = "MotifyMobileAPI"

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
	srvc := service.New(serviceName, "motify_core_api/handlers/mobile_api")
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
			HM:                  handlersmanager.New("motify_core_api/handlers/mobile_api", wrapToken.ModelMobileUser),
			APIHandlerCallbacks: handler.NewHTTPHandlerCallbacks(serviceName, service.AppVersion, "localhost", sessionLogger),
		},
	)

	srvc.MustRegisterHandlers(
		/*
			- login/ singup/ restore pass/ set new pass/ social logins
			- get payslips (одним наверно запросом все данные можно получать). тут надо подумать про апдейт, когда надо получить только новые данные и про пагинацию
			- enter magic code (enroll new enployer)
			- get employers, employer details
			- и возможно всякие системные/служебные хендлеры для включения и выключения нотификаций, данные для аккаунта и прочее
		*/
		user_login.New(coreApi),
		user_signup.New(coreApi),
		user_update.New(coreApi),
		employer_adduser.New(coreApi),
		employer_details.New(coreApi),
		employer_list.New(coreApi),
		payslip_list.New(coreApi),
		payslip_details.New(coreApi),
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
