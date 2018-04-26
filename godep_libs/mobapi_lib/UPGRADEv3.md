# Upgrading mobapi_lib from v2 to v3
(if you still use v1 please check [how to upgrade from v1 to v2](https://bitbucket.lzd.co/projects/GOLIBS/repos/mobapi_lib/browse/UPGRADE.md) first)

## Breaking changes
- new major health_check library [`godep.lzd.co/go-healthcheck`](https://bitbucket.lzd.co/projects/GOLIBS/repos/go-healthcheck/browse) version `v2`
- `service.Run()` method signature changed (it does not return error anymore)
- `service.Healthcheck()` method added. It returns pre-initialized healthcheck instance or nil if service is not initialized yet
- `service.SetResourceHealthChecker()` removed. Please use `service.Healthcheck().SetResourceChecker()` instead.
- `service.GitHash` and `service.GitRev` variables are removed
- if your UT does not set service's `AppVersion`, `BuildDate`, `GoVersion` or `GitDescribe` it will lead to failed UT. Please fill that variables with some data, for example

```
+	service.AppVersion = "test_version"
+	service.BuildDate = "2017-07-20 09:56:27"
+	service.GoVersion = "1.8.3"
+	service.GitDescribe = "mobile_api.20170607-14-ge85b8b9"
```

