# Upgrading mobapi_lib from v1 to v2
(if you need to upgrade up to v3 please check [how to upgrade from v2 to v3](https://bitbucket.lzd.co/projects/GOLIBS/repos/mobapi_lib/browse/UPGRADEv3.md) )

## Breaking changes
- new major discovery library [`motify_core_api/godep_libs/discovery`](https://bitbucket.lzd.co/projects/GOLIBS/repos/discovery/browse) version `v4`
- [`mobapi_lib/context`](https://bitbucket.lzd.co/projects/GOLIBS/repos/mobapi_lib/browse/context/manager.go) methods `NewContext` and `FromContext` requires pointer type *Context instead of Context
- new ldflag required on build: `'-X main.GitDescribe==$(shell git describe --tags --long'`

-----

## Changes in service initialization
You have to call `srv.RegisterHandlers()` method only after `srv.Init()` otherwise you will get error on application start, because client libraries initialization requires pre-initialized discovery service.

## Changes in `main.go`
You need to define additional variable:
```
var (
    // here you already have a few variables like GitRev, GitHash, etc.
    GitDescribe    string // this value set by compiler
)
```
Add one row in `init()` method:
```
func init() {
    // here's some exist code
    service.GitDescribe = GitDescribe
}
```

## Makefile changes that you should do
`make build` target should be a bit extended with additional ldflag.
For example, before changes:
```
VER:=$(shell git branch|grep '*'| cut -f2 -d' ')
GITLOG:=$(shell git rev-parse --abbrev-ref HEAD)
GITHASH:=$(shell git rev-parse HEAD)
GITHASH_SHORT:=$(shell git rev-parse --short HEAD)
DIRTY:=$(shell [ -n "$(shell git status --porcelain)" ] && echo ~dirty)
LDFLAGS=-X 'main.AppVersion=$(VER)($(GITHASH_SHORT))$(DIRTY)' -X 'main.GoVersion=$(GOVER)' -X 'main.BuildDate=$(DATE)' -X 'main.GitRev=$(GITLOG)' -X 'main.GitHash=$(GITHASH)'
```
After changes:
```
VER:=$(shell git branch|grep '*'| cut -f2 -d' ')
GITLOG:=$(shell git rev-parse --abbrev-ref HEAD)
GITHASH:=$(shell git rev-parse HEAD)
GITHASH_SHORT:=$(shell git rev-parse --short HEAD)
DIRTY:=$(shell [ -n "$(shell git status --porcelain)" ] && echo ~dirty)
GITDESCRIBE := $(shell git describe --tags --long)
LDFLAGS=-X 'main.AppVersion=$(VER)($(GITHASH_SHORT))$(DIRTY)' -X 'main.GoVersion=$(GOVER)' -X 'main.BuildDate=$(DATE)' -X 'main.GitRev=$(GITLOG)' -X 'main.GitHash=$(GITHASH)' -X 'main.GitDescribe=$(GITDESCRIBE)'
```

## FAQ
### I've set last mobapi_lib version in `glide.yaml` and performed `glide up`, but it still returns error.
```
[ERROR]    Error scanning motify_core_api/godep_libs/discovery/discovery: open /Users/minh.ton/.glide/cache/src/https-motify_core_api/godep_libs-discovery/discovery: no such file or directory
[ERROR]    This error means the referenced package was not found.
[ERROR]    Missing file or directory errors usually occur when multiple packages
[ERROR]    share a common dependency and the first reference encountered by the scanner
[ERROR]    sets the version to one that does not contain a subpackage needed required
[ERROR]    by another package that uses the shared dependency. Try setting a
[ERROR]    version in your glide.yaml that works for all packages that share this
[ERROR]    dependency.
```
It's very common case. As you can see the problem occured because in your code you still have imports like `import "motify_core_api/godep_libs-discovery/discovery"`, but this sub-package does not exist anymore in discovery library.

In most cases you just need to set proper version of client library in glide.yaml to particular external resource.

| Client library                                                                                           | Version that compatible with discovery v4                           |
|----------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------|
| [catalog_api_client_go](https://bitbucket.lzd.co/plugins/servlet/network/GOLIBS/catalog_api_client_go)   | discovery_4.0.0, based on v.12.0.0 (still does not have semver tag) |
| [go-client-transport](https://bitbucket.lzd.co/plugins/servlet/network/GOLIBS/go-client-transport)       | >=4.0.0                                                             |
| [customer_api_client_go](https://bitbucket.lzd.co/plugins/servlet/network/GOLIBS/customer_api_client_go) | >=3.5.0                                                             |
| [go-null-types](https://bitbucket.lzd.co/plugins/servlet/network/GOLIBS/go-null-types)                   | >=4.0.0                                                             |
| [leadtime_api_client_go](https://bitbucket.lzd.co/plugins/servlet/network/GOLIBS/leadtime_api_client_go) | discovery_4.0.0, based on v.6.0.0 (still does not have semver tag)  |
| [bundle_api_client_go](https://bitbucket.lzd.co/plugins/servlet/network/GOLIBS/bundle_api_client_go)     | >=1.0.12                                                            |
| [go-limiter](https://bitbucket.lzd.co/plugins/servlet/network/GOLIBS/go-limiter)                         | >=0.0.8                                                             |
| [bob_api_client_go](https://bitbucket.lzd.co/plugins/servlet/network/GOLIBS/bob_api_client_go)           | >=2.1.0                                                             |
| [geo_api_client_go](https://bitbucket.lzd.co/plugins/servlet/network/GOLIBS/geo_api_client_go)           | >=3.0.0                                                             |
| [config_api_client_go](https://bitbucket.lzd.co/plugins/servlet/network/GOLIBS/config_api_client_go)     | >=5.0.0                                                             |
| [stock_api_client_go](https://bitbucket.lzd.co/plugins/servlet/network/GOLIBS/stock_api_client_go)       | >=11.0.0                                                            |

### I've updated all libs to proper versions but still getting same `glide up` error
So, it seems some dependency still uses old discovery library version. How to find it?
1. Make `glide up` and get the error
2. Find wrong import in glide cache
```
grep -irn "motify_core_api/godep_libs/discovery/discovery" ~/.glide/cache
```
3. Define which component tries to make wrong import and find its latest version that already compatible with discovery v4.
4. Share information about this problem with @yuriy.savalin in Slack (to make this guide more relevant)
