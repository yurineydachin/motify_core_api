package main

import (
"os"
    "time"

	"godep.lzd.co/service"
	"godep.lzd.co/service/config"
	"godep.lzd.co/service/handlersmanager"
	"godep.lzd.co/service/logger"

	//resourceSearchEngine "motify_core_api/resources/searchengine"
	"motify_core_api/resources/database"

	"motify_core_api/token"
	"motify_core_api/srv/user"

	"motify_core_api/handlers/payslip/set"
	"motify_core_api/handlers/user/login"
	"motify_core_api/handlers/user/create"
	//handlerHelloWorld "motify_core_api/handlers/hello/world"
	//handlerSearchGoogle "motify_core_api/handlers/search/google"
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

	config.RegisterString("config-shop-timezone", "Config shop timezone", "Local")
	config.RegisterString("mysql-db-read-nodes", "DB read nodes", "root:123456@tcp(localhost:3306)/motify_core_api")
	config.RegisterString("mysql-db-write-nodes", "DB write nodes", "root:123456@tcp(localhost:3306)/motify_core_api")

	config.RegisterString("token-triple-des-key", "24-bit key for token DES encryption", "123456789012345678901234")
	config.RegisterString("token-salt", "8-bit salt for token DES encryption", "12345678")
}

func main() {
    /*
	if err := config.ParseAll(); err != nil {
		logger.Critical(nil, err.Error())
		os.Exit(1)
	}
    */

	if err := initToken(); err != nil {
		logger.Critical(nil, "failed to init token encryption: %v", err)
		os.Exit(1)
	}

	dbReadNodes, _ := config.GetStringSlice("mysql-db-read-nodes")
	dbWriteNodes, _ := config.GetStringSlice("mysql-db-write-nodes")
    config.Dump()
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
	}

	srvc := service.New(serviceName, "motify_core_api/handlers")
	//se := &resourceSearchEngine.SearchEngine{}
	//srvc.RegisterResource(se)

    userService := user_service.NewUserService(db)

	srvc.SetOptions(service.Options{HM: handlersmanager.New("motify_core_api/handlers")})
	srvc.MustRegisterHandlers(
    /*
- login/ singup/ restore pass/ set new pass/ social logins
- get payslips (одним наверно запросом все данные можно получать). тут надо подумать про апдейт, когда надо получить только новые данные и про пагинацию 
- enter magic code (enroll new enployer)
- get employers, employer details
- и возможно всякие системные/служебные хендлеры для включения и выключения нотификаций, данные для аккаунта и прочее
    */
		payslip_set.New(),
		user_login.New(userService),
		user_create.New(userService),
		//handlerHelloWorld.New(),
		//handlerSearchGoogle.New(se),
	)

	logger.Error(nil, "dbNodes: %#v, DB adapter %#v", dbNodes, db)

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
	return token.InitTokenV1([]byte(key), []byte(salt), "", "")
}
