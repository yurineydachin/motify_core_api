# Health Check Library
This library provides a tool for easy health check handler implementation. It fits common Lazada's [health check convention](https://confluence.lzd.co/display/DEV/Microservice+Architecture+%28SOA%29+Conventions#MicroserviceArchitecture(SOA)Conventions-GET%3Cservice%3E%3A%3Cport%3E/health_check).

## Usage
0. Import healthcheck library
```
    import "motify_core_api/godep_libs/go-healthcheck"
```

1. Initialize base health check
```
    var buildDateTimestamp = 1497432488
    var serviceID = "service_name_as_in_etcd"
    var goVersion = "1.9"
    var serviceVersion = "branch_name@hash"   // branch_name = `git rev-parse --abbrev-ref HEAD` ; hash = `git rev-parse --short HEAD`
    var gitDescribe = "git-describe value"    // (optional) `git describe --long --tags`

    // isEtcdAppConfigRequired must be TRUE if your service uses configuration from ETCD instead of *.ini file (see https://bitbucket.lzd.co/projects/GOLIBS/repos/etcdconfig for details)
    var isEtcdAppConfigRequired = true

    // venture can be empty on start and should be set after reading config from ETCD using SetVenture() method
    var venture = "venture"

    hc, err := healthcheck.New(serviceID, serviceVersion, gitDescribe, goVersion, venture, buildDateTimestamp, isEtcdAppConfigRequired)
    if err != nil {
        panic(err)
    }
```

2. Set checkers for external resources that your service uses (optional)
```
    // Set MySQL checker func. And this resource is critical (service can't work if resource is down)
    if err := hc.SetResourceChecker(healthcheck.ResourceTypeMySQL, true, mySQLChecker, time.Second*3); err != nil {
        panic(err)
    }
    // Set Aerospike checker func. And this resource is not critical (service can work without it)
    if err := hc.SetResourceChecker(healthcheck.ResourceTypeAerospike, false, aerospikeChecker, time.Second*1); err != nil {
        panic(err)
    }

    // init RabbitMQ client and use it for resource health check
    // import "motify_core_api/godep_libs/rabbit"
    rabbitClient, _ = rabbit.New(rabbit.WithEtcdEndpoints("127.0.0.1:2379"))

    // rabbitClient has method "HealthCheck(context.Context)error" so it's ok to pass "rabbitClient" as checker
    // instead of passing "rabbitClient.HealthCheck"
    if err := hc.SetResourceChecker(healthcheck.ResourceTypeRabbitMQ, false, rabbitClient); err != nil {
       panic(err)
    }

```

3. If your service already can work with configuration stored in ETCD you need to set current configuration path in ETCD and flag if ETCD config is enabled.
And you can safely change (re-set) this values while service is running (optional).
```
    hc.SetAppConfigInfo("some/path/in/etcd", true)
    // read venture value from etcd and set in healthcheck
    hc.SetVenture("vn")
```

4. Set custom key/value in response (optional)
```
    hc.SetCustomValue("custom_key_in_healtch_check", func(ctx context.Context) interface{} {
        // do any logic here or just return data
        // for example, you can get data from context (if you set some custom data in context in middleware)
        // check TestHealthCheck_CustomValuesInContext
        return ctx.Value("some_key")
    })
```

5. Start to serve it
```
    http.Handle(healthcheck.DefaultHealthCheckPath, hc.Handler())
    // OR
    // http.HandleFunc(healthcheck.DefaultHealthCheckPath, hc.HandlerFunc())

    http.ListenAndServe(":8080", nil)
```

In result (if you open http://localhost:127.0.0.1/health_check) you can see a JSON:
```JSON
    {
      "aerospike_status": "Ok",
      "app_config_path": "some/path/in/etcd",
      "app_config_ready": true,
      "build_date": "build-date-is-here",
      "git_describe": "gi-describe value",
      "go_version": "1.9",
      "mysql_status": "Ok",
      "service": "service_name_as_in_etcd",
      "status": "Ok",
      "version": "version",
      "venture": "vn",
      "custom_key_in_healtch_check": "some_value",
    }
```

You can check out [the example](https://bitbucket.lzd.co/projects/GOLIBS/repos/go-healthcheck/browse/_example/example.go) .

-----

### Important notes
#### Difference between setting resource checker and custom value
When you set resource checker using `SetResourceChecker()` method it affects `/health_check` response HTTP status code and `status` field (in case of `isCritical == true`).
And in this case you can't manage value in response (library sets it using internal logic based on error that checker returns and `isCritical` parameter).

When you set custom value using `SetCustomValue()` method it does not affect `status` field in response and HTTP status code.

#### Difference in HealthCheck behaviour if `isEtcdAppConfigRequired` parameter is TRUE/FALSE
If `isEtcdAppConfigRequired == true` and you still did not call `SetAppConfigInfo()` and set valid `appConfigPath` and `appConfigReady` values then HealthCheck will return `500` status code, because your service depends on ETCD configuration, but it's still not ready to use.

If `isEtcdAppConfigRequired == false` then HealthCheck result does not depend on `appConfigPath`/`appConfigReady` parameters values and returns `200` status code if they are empty.

#### Can I set additional checkers and/or custom values after I did `http.Handle(healthcheck.DefaultHealthCheckPath, hc.Handler())`?
Yes, you can set additional checker and custom values any time during your service lifetime.

#### Do common resource checkers exist?
Yes. You can find them in repos:
- RabbitMQ: https://bitbucket.lzd.co/projects/GOLIBS/repos/rabbit/browse/healthcheck.go
- DB Manager v8: https://bitbucket.lzd.co/projects/GOLIBS/repos/go-db-manager/browse/dbcore/healthcheck.go
- DB Manager v9: https://bitbucket.lzd.co/projects/GOLIBS/repos/go-db-manager/browse/healthcheck.go?at=refs%2Fheads%2Fv9
- Aerospike: https://bitbucket.lzd.co/projects/GOLIBS/repos/go-healthcheck-aerospike/browse