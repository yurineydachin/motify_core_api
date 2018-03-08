package main

import (
	"context"
	"net/http"

	"godep.lzd.co/go-healthcheck"
	"godep.lzd.co/rabbit"
)

// FIND MORE RESOURCE CHECKERS
// DB Manager v8: https://bitbucket.lzd.co/projects/GOLIBS/repos/go-db-manager/browse/dbcore/healthcheck.go
// DB Manager v9: https://bitbucket.lzd.co/projects/GOLIBS/repos/go-db-manager/browse/healthcheck.go?at=refs%2Fheads%2Fv9
// Aerospike: https://bitbucket.lzd.co/projects/GOLIBS/repos/go-healthcheck-aerospike/browse

var rabbitClient rabbit.Provider

func init() {
	var err error
	rabbitClient, err = rabbit.New(rabbit.WithEtcdEndpoints("127.0.0.1:2379"))

	if err != nil {
		panic(err)
	}
}

func main() {
	var buildDateTimestamp = int64(1497432488)
	var serviceID = "service_name_as_in_etcd"
	var serviceVersion = "branch_name@hash" // branch_name = `git rev-parse --abbrev-ref HEAD` ; hash = `git rev-parse --short HEAD`
	var gitDescribe = "git-describe value"  // (optional) `git describe --long --tags`
	var goVersion = "1.9"
	var isEtcdAppConfigRequired = true // must be TRUE if your service uses configuration from ETCD instead of *.ini file
	var venture = "venture"            // can be empty on start and should be set after reading config from ETCD using SetVenture() method

	// Init base health check
	hc, err := healthcheck.New(serviceID, serviceVersion, gitDescribe, goVersion, venture, buildDateTimestamp, isEtcdAppConfigRequired)
	if err != nil {
		panic(err)
	}

	// Set checkers for external resources that your service uses.
	// Set MySQL checker func. And this resource is critical (service can't work if resource is down)
	if err := hc.SetResourceChecker(healthcheck.ResourceTypeMySQL, true, mySQLChecker); err != nil {
		panic(err)
	}
	// Set Aerospike checker func. And this resource is not critical (service can work without it)
	if err := hc.SetResourceChecker(healthcheck.ResourceTypeAerospike, false, aerospikeChecker); err != nil {
		panic(err)
	}
	// Set RabbitMQ checker.This resource is not critical (service can work without it) and checker is struct that implements interface with method HealthCheck()
	if err := hc.SetResourceChecker(healthcheck.ResourceTypeRabbitMQ, false, rabbitClient); err != nil {
		panic(err)
	}

	// If your service already can work with configuration stored in ETCD
	// you need to set current configuration path in ETCD and flag if ETCD config is enabled
	hc.SetAppConfigInfo("some/path/in/etcd", true)

	// If your service gets venture from ETCD and you don't know what venture used on service start
	// you can set venture value later
	hc.SetVenture("vn")

	// If you wish to set custom key/value in health check response
	hc.SetCustomValue("custom_key_in_health_check", func(ctx context.Context) interface{} {
		// do any logic here or just return data
		// for example, you can get data from context (if you set some custom data in context in middleware)
		// check TestHealthCheck_CustomValuesInContext
		return ctx.Value("some_key")
	})

	// start to serve it
	http.Handle(healthcheck.DefaultHealthCheckPath, hc.Handler())
	// OR
	//http.HandleFunc(healthcheck.DefaultHealthCheckPath, hc.HandlerFunc())

	http.ListenAndServe(":8080", nil)

	// If you run this example and open http://127.0.0.1:8080/health_check
	// it returns JSON
	/*
		{
		  "aerospike_status": "Ok",
		  "app_config_path": "some/path/in/etcd",
		  "app_config_ready": true,
		  "build_date": "build-date-is-here",
		  "git_describe": "gi-describe value",
		  "go_version": "1.9",
		  "mysql_status": "Ok",
		  "rabbit_mq_status": "Ok",
		  "service": "service_name_as_in_etcd",
		  "status": "Ok",
		  "venture": "venture",
		  "version": "vn",
		  "custom_key_in_health_check": "some_value",
		}
	*/
}

func mySQLChecker(context.Context) error {
	// do some request to MySQL to check if it's still alive
	// return not nil error in case of some problems

	return nil
}

func aerospikeChecker(context.Context) error {
	// do some request to Aerospike to check if it's still alive
	// return not nil error in case of some problems

	return nil
}
