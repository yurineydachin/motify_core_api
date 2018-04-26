# mobapi_lib
MobAPI framework

- [Installation](#Installation)
- [Usage](#Usage)
- [Documentation](#Documentation)
- [Changelog](#Changelog)

## Installation
Initially you can put mobapi_lib in your %GOPATH% using `go get motify_core_api/godep_libs/mobapi_lib`.

If you use vendoring with some dependency manager ([glide](https://github.com/Masterminds/glide) for example) you can simply add mobapi_lib dependency to glide.yaml and perform `glide up`.


## Usage
All you need is to create new service, initialize it and run.
```golang
// Create new service
// You need to define new service name (it will be used for ETCD registration and metrics collection)
// and path to directory that contains handlers sub-packages ()
srv := service.New("servive_name", "new_api_repo/api/handler")

// Initialize service internals and register handlers
if err := srv.Init(); err != nil {
    panic(err)
}

// Create handlers that implement IHandler interface
var handlers []gorpc.IHandler = NewHandlers()

// Register handlers
if err := app.srv.RegisterHandlers(api.NewHandlers(app.srv, getVirtualFS())...); err != nil {
    panic(err)
}

// Run it (service registers itself in discovery service and start to listen for incoming connections)
if err := app.srv.Run(); err != nil {
    panic(err)
}

```


## Documentation
There are a presentation about mobile horizontal and vertical architecture and some live examples available in [sharepoint](https://lazadagroup-my.sharepoint.com/personal/svistunov_sergey_lazada_com/_layouts/15/guestaccess.aspx?guestaccesstoken=Rm9N104MvvBfhylTuhMANEtQEvHKo5f8YWPn5JmX2Pw%3d&docid=2_0db521f4b9e6b42a3a83e9c7334042e3e&rev=1) (CAS authentication required).

Also you can check simple service [example](_example/main.go) in this repo.


## Changelog
All mobapi_lib update details could be found in [CHANGELOG.md](CHANGELOG.md)