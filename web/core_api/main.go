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

	"motify_core_api/resources/database"

	"motify_core_api/srv/agent"
	"motify_core_api/srv/email"
	"motify_core_api/srv/integration"
	"motify_core_api/srv/payslip"
	"motify_core_api/srv/user"

	"motify_core_api/handlers/core_api/agent/create"
	"motify_core_api/handlers/core_api/agent/list"
	"motify_core_api/handlers/core_api/agent/update"
	"motify_core_api/handlers/core_api/employee/create"
	"motify_core_api/handlers/core_api/employee/details"
	"motify_core_api/handlers/core_api/employee/invite"
	"motify_core_api/handlers/core_api/employee/update"
	"motify_core_api/handlers/core_api/integration/check"
	"motify_core_api/handlers/core_api/integration/create"
	"motify_core_api/handlers/core_api/payslip/create"
	"motify_core_api/handlers/core_api/payslip/details"
	"motify_core_api/handlers/core_api/payslip/list"
	"motify_core_api/handlers/core_api/setting/create"
	"motify_core_api/handlers/core_api/setting/update"
	"motify_core_api/handlers/core_api/user/create"
	"motify_core_api/handlers/core_api/user/login"
	"motify_core_api/handlers/core_api/user/update"
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

	config.RegisterString("config-shop-timezone", "Config shop timezone", "Local")
	config.RegisterString("mysql-db-read-nodes", "DB read nodes", "")
	config.RegisterString("mysql-db-write-nodes", "DB write nodes", "")

	config.RegisterString("mail-host", "Email smtp host", "")
	config.RegisterUint("mail-port", "Email smtp port", 587)
	config.RegisterString("mail-user", "Email user", "")
	config.RegisterString("mail-password", "Email password", "")
	config.RegisterString("mail-employee-invite-from", "Email user who sends invite", "")

	config.RegisterString("token-triple-des-key", "24-bit key for token DES encryption", "")
	config.RegisterString("token-salt", "8-bit salt for token DES encryption", "")
}

func main() {
	srvc := service.New(serviceName, "motify_core_api/handlers/core_api")
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

	host, _ := config.GetString("mail-host")
	port, _ := config.GetUint("mail-port")
	userEmail, _ := config.GetString("mail-user")
	userPassword, _ := config.GetString("mail-password")
	userInvite, _ := config.GetString("mail-employee-invite-from")

	dbReadNodes, _ := config.GetStringSlice("mysql-db-read-nodes")
	dbWriteNodes, _ := config.GetStringSlice("mysql-db-write-nodes")
	if len(dbReadNodes) == 0 || len(dbWriteNodes) == 0 {
		logger.Critical(nil, "No DB nodes in config: %v, %v", dbReadNodes, dbWriteNodes)
		os.Exit(1)
	}
	dbNodes := append(dbWriteNodes, dbReadNodes...)

	location, _ := config.GetString("config-shop-timezone")
	tz, err := time.LoadLocation(location)
	time.Local = tz
	if err != nil {
		logger.Critical(nil, "config-shop-timezone has wrong value %s: set a valid timezone in a config", location)
		os.Exit(1)
	}
	location = tz.String()

	db, err := database.NewDbAdapter(dbNodes, location, false)
	if err != nil {
		logger.Critical(nil, "DB adapter init error: %v", err)
		os.Exit(1)
	}

	srvc.RegisterResource(db)

	agentService := agent_service.NewAgentService(db)
	userService := user_service.NewUserService(db)
	payslipService := payslip_service.NewPayslipService(db)
	integrationService := integration_service.NewIntegrationService(db)
	emailService := email_service.NewService(host, port, userEmail, userPassword)

	dconfm := dconfig.NewManager(serviceName, mobLogger.GetLoggerInstance())
	sessionLogger, err := sessionlogger.NewSessionLoggerFromFlags(dconfm)
	if err != nil {
		logger.Critical(nil, err.Error())
		os.Exit(1)
	}
	srvc.SetOptions(
		service.Options{
			HM:                  handlersmanager.New("motify_core_api/handlers/core_api"),
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
		payslip_details.New(agentService, payslipService),
		payslip_create.New(agentService, payslipService),
		payslip_list.New(payslipService),
		agent_create.New(agentService),
		agent_update.New(agentService),
		agent_list.New(agentService),
		employee_create.New(agentService, userService),
		employee_details.New(agentService, payslipService),
		employee_invite.New(agentService, emailService, userInvite),
		employee_update.New(agentService, userService),
		user_login.New(userService),
		user_create.New(userService),
		user_update.New(userService),
		setting_create.New(agentService, userService),
		setting_update.New(agentService, userService),
		integration_create.New(integrationService),
		integration_check.New(integrationService),
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
