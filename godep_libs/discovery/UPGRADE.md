# etcd2 -> etcd3 migration guide

## Notes

Migration from **etcd2** to **etcd3** store is in progress.

It's done mainly to reduce boilerplate traffic, caused by etcd2 TTL keep-alive feature.
In order to keep some key alive you have to **put the same data** in etcd2 every `TTL/2` seconds (this what etcd2 Registrator does undercover).
This leads to growth both of outgoing traffic (service -> etcd) and incoming traffic to services subscribed for the key updates.

The migration is done in 2 steps:
- (**current**) supporting both etcd3 and etcd2 discovery for backwards compatibility for existing production services;
- drop etcd2 support, clean-up codebase and use etcd3-only discovery.

To make it happen we introduce both **new types and interfaces** for etcd3 registration and some **helper types** for migration period.

Many deprecated code was purged in 4.0.0 release.
Old etcd2 discovery code became private and is used only in migration helpers undercover.

It's done because too many people just updated the discovery lib version and thought they are using new features :)
So the decision was to break compatibility and force people to use the intended stuff.

Check [example application](./_example) with the full set of registration and balancing features.


## Provider

Provider is a wrapper around any key-value storage we want to use for service discovery system.
It encapsulates and hides any specific implementation details under [IProvider](https://bitbucket.lzd.co/projects/GOLIBS/repos/discovery/browse/provider/provider.go?at=refs%2Ftags%2F4.0.0#15) interface.
IProvider is used across the whole discovery lib for service registration and location.

etcd3 provider implementation can be found [here](https://bitbucket.lzd.co/projects/GOLIBS/repos/discovery/browse/provider/etcdV3/provider.go?at=refs%2Ftags%2F4.0.0).

To install etcd3 provider you simply pass [`etcd3.Client`](https://godoc.org/github.com/coreos/etcd/clientv3#Client) instance and [`discovery.ILogger`](https://bitbucket.lzd.co/projects/GOLIBS/repos/discovery/browse/logger.go?at=refs%2Ftags%2F4.0.0#4) instance. Logger can be nil, but it's **highly not recommended**.

```golang

import "motify_core_api/godep_libs/discovery/provider/etcdV3"

...

app.provider = etcdV3.NewProvider(app.etcdClientV3, app.logger)
```

Service application can safely create a single instance of IProvider for any further usage in balancing and registration.


## Registration

Old etcd2 registration contained too much boilerplate code:

```golang
wg.Add(1)
go func() {
    defer wg.Done()

    registrator.Register(
        ctx,
        discovery.RegistrationInfo{
            Namespace:   "lazada_api",
            Venture:     app.config.Venture,
            Environment: app.config.Environment,
            ServiceName: serviceName,
            Value:       fmt.Sprintf("http://%s:%d", app.config.Host, app.config.Host),
            Key:         fmt.Sprintf("%s:%d", app.config.Host, app.config.Host),
            Property:    discovery.NodesProperty,
        },
    )
}()

wg.Add(1)
go func() {
    defer wg.Done()

    registrator.Register(
        ctx,
        discovery.RegistrationInfo{
            Namespace:   "lazada_api",
            ...
            Value:       app.Version,
            Property:    discovery.VersionsProperty,
        }
    )
}()

wg.Add(1)
go func() {
    defer wg.Done()

    registrator.Register(
        ctx,
        discovery.RegistrationInfo{
            Namespace:   "lazada_api",
            ...
            Value:       monitoringPrometheus.GetEtcdValue(
                app.config.Host,
                ProfilingPort,
                app.config.Venture,
                app.config.Environment,
                serviceName,
            ),
            Property:    discovery.MetricsProperty,
        }
    )
}()
```

And unregistration looked like canceling the context and waiting on WaitGroup

```golang
registerCancel()
wg.Wait()
```

New etcd3 registration is done in terms of [Reliability and Resilience Convention](https://confluence.lzd.co/display/DEV/Microservice+Architecture+(SOA)+Conventions#MicroserviceArchitecture(SOA)Conventions-soa-reliability-and-resilienceReliabilityandResilience).

[IRegistrator](https://bitbucket.lzd.co/projects/GOLIBS/repos/discovery/browse/registrator/registrator.go?at=refs%2Ftags%2F4.0.0#12) interface implements this convention.

Service must register in discovery in 2 steps:
- call `registrator.Register()` on startup;
- call `registrator.EnableDiscovery()` when it's ready to serve incoming requests.

Unregistration is also 2-step:
- call `registrator.DisableDiscovery()` if the service lost connection with DB. Enable it back when the connection is re-established;
- call `registrator.Unregister()` on app shutdown.

There is special registrator implementation all services **MUST** use, enabling **both** etcd2 and etcd3 registration.

```golang
import (
    "motify_core_api/godep_libs/discovery/registrator"

)

...

func (app *Application) initDiscoveryRegistrator() error {
    info := registrator.AppRegistrationParams{
        // Service discovery info
        ServiceName: app.config.ServiceName,
        RolloutType: app.config.RolloutType,
        Host:        app.config.Host,
        HTTPPort:    app.config.Port,
        // Admin info
        AdminPort: app.config.AdminPort,
        Version: registrator.VersionInfo{
            AppVersion: app.Version,
            // please take a look at registrator.VersionInfo - you can set other fields you want here
        },
        // Monitoring info
        MonitoringPort: app.config.AdminPort,
        Venture:        app.config.Venture,
        Environment:    app.config.Env,
    }
    app.provider = etcdV3.NewProvider(app.etcdClientV3, app.logger)

    r, err := registrator.NewV2V3(info, app.provider, app.etcdClientV2, app.logger)
    if err != nil {
        return err
    }
    app.registrator = r
    return nil
}

func (app *Application) init(){
    ...

    if err := app.initDiscoveryRegistrator(); err != nil {
        // here are only registration info validation errors, which can't be fixed in runtime.
        // if recieved - you should fix the data you provided!
        // so you can easily panic when calling initDiscoveryRegistrator().
        panic(err)
    }

    // Register app
    app.registrator.Register()

    ...
    // prepare app - establish connections with DBs, etc...
    app.registrator.EnableDiscovery()
}

```


## Balancing

Old `LoadBalancer` interface is renamed to `ILoadBalancer`, to match our codestyle convention.

`NewRoundRobin` and `NewWeightedRoundRobin` signatures are changed to be used with new `locator.ILocator` interface.
You shouldn't think about them for now, because special FallbackBalancer is prepared for you and should be used instead.

### Fallback balancer usage

[FallbackBalancer](https://bitbucket.lzd.co/projects/GOLIBS/repos/discovery/browse/balancer/fallback_balancer.go?at=refs%2Ftags%2F4.0.0#8)
is a load balancer, which should be used for migration period between different discovery protocols and interfaces.

The main reason is to have some simple solution of watching both etcd3 and etcd2 discovery data changes.
It has several load balancer undercover set with some priority. In our case we should try to locate service **from etcd3 first** and etcd2 as a **fallback**.

There is a helper function to install FallbackBalancer for our case of etcd3 -> etcd2 migration - [NewFallbackBalancerEtcd2](https://bitbucket.lzd.co/projects/GOLIBS/repos/discovery/browse/balancer/etcd2_migration.go?at=4.0.0#88)

Example usage:

```golang
func (app *Application) createBalancer(serviceName string) balancer.ILoadBalancer {
    opts := balancer.FallbackBalancerEtcd2Options{
        ServiceName: serviceName,
        Venture:     app.config.Venture,
        Environment: app.config.Env,
    }
    return balancer.NewFallbackBalancerEtcd2(app.provider, app.etcdClientV2, app.logger, opts)
}
```

If you want to use `WeightedRoundRobin` - there is `BalancerType` option to choose underlying balancer type.

### Factory

There is some helper code like [IBalancerFabric and IBalancerRegestry](https://bitbucket.lzd.co/projects/GOLIBS/repos/discovery/browse/balancer/balancer_fabric.go?at=4.0.0)

I decided to keep it alive and **changed** `New...()` constructor signatures to correspond with balancing in both etcd3 and etcd2.
Fallback balancer is used undercover.

Please, tell if it's really useful, otherwise it will be **removed** in next major release with dropping etcd2 support.
