## [3.2.0] - 2017-10-09
### Added
- [`godep.lzd.co/etcd`](https://bitbucket.lzd.co/projects/GOLIBS/repos/etcd) `^1.1.1` integrated. It provides an ability to use SSL connection between service and ETCD.

### Changed
- [`github.com/coreos/etcd`](https://github.com/coreos/etcd) updated: `^3.2.4`
- [`godep.lzd.co/go-log`](https://bitbucket.lzd.co/projects/GOLIBS/repos/go-log) updated: `^3.8.0`. Use logger.Flush() method instead of timer.

## [3.1.1] - 2017-09-01
### Changed
- [`github.com/sergei-svistunov/gorpc`](https://github.com/sergei-svistunov/gorpc) dependency changed from custom branch `issue_82` to semver `^3.5.0`

## [3.1.0] - 2017-08-28
### Added
- `context.SetErrorTranslatorDictionary(d dict)` method that allows you to set dictionary for automatic handler error messages translation.

`dict` interface description:
```
type dict interface {
	Translate(in string, argsH map[string]interface{}, argsA []interface{}) (out string, ok bool)
}
```
Usage example:
```
func (handler *Bar) V1(ctx context.Context, opts *V1Opts) (*V1Ret, error) {
    if ctxData, err := ctxManager.FromContext(ctx); err == nil {
        ctxData.SetErrorTranslatorDictionary(handler.i18nManager.GetDictionary("ID"))
    }
    return nil, v1Errors.MISSED_REQUIRED_FIELDS
}
```

## [3.0.0] - 2017-07-28
### Changed
- [`godep.lzd.co/go-healthcheck`](https://bitbucket.lzd.co/projects/GOLIBS/repos/go-healthcheck) updated up to new major version 2
- Some methods and variables have removed or signature changed (check [upgrade guide](https://bitbucket.lzd.co/projects/GOLIBS/repos/mobapi_lib/browse/UPGRADEv3.md) for details)
- `AppVersion`, `BuildDate`, `GoVersion` and `GitDescribe` are required. It could fail your UT.

## [2.2.0] - 2017-07-10
### Added
- ability to set cookies in handler response. See example for details
- ability to set custom resource health checker. Service got additional method `SetResourceHealthChecker(resourceName string, isCritical bool, checker func(context.Context) error, updatePeriod time.Duration) error`
### Changed
- need to adjust Makefile to pass additional ldflag `'-X main.GitDescribe==$(shell git describe --tags --long'` and define this variable in main.go

## [2.1.0] - 2017-06-06
### Fixed
- File descriptors leak in [go-log](https://bitbucket.lzd.co/projects/GOLIBS/repos/go-log/commits/00267ec061aa7d77b7adcfc5313a4b90f8e27c0c) library fixed
### Changed
-  [`godep.lzd.co/go-log`](https://bitbucket.lzd.co/projects/GOLIBS/repos/go-log/browse) dependency updated up to `^3.0.1`

## [2.0.1] - 2017-06-01
### Fixed
- Wrong app name in metrics (it used mobile client names like iOS/Android instead of mobapi service name)
- README.md content updated


## [2.0.0] - 2017-05-26
### Changed
- [`godep.lzd.co/discovery`](https://bitbucket.lzd.co/projects/GOLIBS/repos/discovery/browse) dependency updated up to new major version `^4.0.0`
- More information about migration to new discovery: https://bitbucket.lzd.co/projects/GOLIBS/repos/discovery/browse/UPGRADE.md?at=refs%2Fheads%2FGOLIBS-1080

## [1.9.0] - 2017-05-02
### Added
- Ability to add gRPC external service (AddExternalGRPCService() method)
### Changed
-  [`godep.lzd.co/discovery`](https://bitbucket.lzd.co/projects/GOLIBS/repos/discovery/browse) dependency updated up to `^3.5.0`


## [1.8.0] - 2017-04-24
### Changed
-  [`godep.lzd.co/go-log`](https://bitbucket.lzd.co/projects/GOLIBS/repos/go-log/browse) dependency updated up to new major version `^3.0.0`
-  [`godep.lzd.co/go-trace`](https://bitbucket.lzd.co/projects/GOLIBS/repos/go-trace/browse) dependency updated up to new minor version `^4.5.0`


## [1.7.0] - 2017-04-19
### Added
-  OpenTracing support (http://opentracing.io)


## [1.6.1] - 2017-04-14
### Added
-  `ReqMobAppVersion` field in context that stores mobile app version string


## [1.6.0] - 2017-04-11
### Added
- Service registers itself in ETCD v3 by default

### Changed
- Dependencies version changed: [`godep.lzd.co/discovery`](https://bitbucket.lzd.co/projects/GOLIBS/repos/discovery/browse) `~3.3.0`, [`github.com/coreos/etcd`](https://github.com/coreos/etcd) `^v3.1.0`


## [1.5.3] - 2017-04-10
### Added
- `--advertised-hostname` parameter that could be used for service hostname definition (used for ETCD registration and logging)

### Changed
- `os.Hostname()` used as fallback, `--advertised-hostname` parameter is primary


## [1.5.2] - 2017-03-15
### Changed
- [go-log](https://bitbucket.lzd.co/projects/GOLIBS/repos/go-log/browse) library is updated up to `~2.0.5`

### Fixed
- confusing error messages about session logger are removed, because session logger is disabled by default

## [1.5.1] - 2017-03-14
### Added
- Service writes access log for each request if ANY log level is set


## [1.5.0] - 2017-03-13
### Changed
- gorpc dependency is updates up to `3.1.0` (here same structs usage in different handler versions is restricted)


## [1.4.2] - 2017-03-13
### Changed
- [go-log](https://bitbucket.lzd.co/projects/GOLIBS/repos/go-log/browse) library is updated up to `2.0.4`


## [1.4.1] - 2017-03-03
### Fixed
- [go-log](https://bitbucket.lzd.co/projects/GOLIBS/repos/go-log/browse) data race

## [1.4.0] - 2017-02-28
### Removed
- `/docs` page with SwaggerUI documentation, because it's legacy and you should use SOA Manager swagger feature


## [1.3.0] - 2017-02-28
### Added
- `Logf()` method in logger instance

### Fixed
- Failover logger message format, time format and timezone (UTC instead of local)
- Failover logger messages have "FAILOVER_LOGGER" component name instead of confusing "AL" name
- "File name and line" logger feature is fixed (write real data instead of go-log library filename and its line)

### Removed
- `logger.Stdout` instance and its usage, because it duplicates logger interface and leads to confusion


## [1.2.2] - 2017-02-22
### Fixed
- Flag parsing bug if its name contains underscores

### Deprecated
- `logger.Stdout` will be removed in next versions because it duplicates logger interface. Please initialize logger with params for StdOUT usage and use it instead of `logger.Stdout`