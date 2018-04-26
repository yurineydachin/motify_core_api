# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)


## [4.5.0] - [GOLIBS-1582](https://jira.lzd.co/browse/GOLIBS-1582) - 2017-09-04

### Added
- discovery of additional endpoint of system services


## [4.4.2] - [GOLIBS-1542](https://jira.lzd.co/browse/GOLIBS-1542) - 2017-08-25

### Fixed
- exit stuck goroutines, creating watch channel on etcd error


## [4.4.0] - [GOLIBS-1397](https://jira.lzd.co/browse/GOLIBS-1397) - 2017-08-04

### Upgraded
- `github.com/coreos/etcd` dependency from version 3.1.0 to **3.2.4**
- `google.golang.org/grpc` dependency from version 1.0.5 to **1.5.0**
- `github.com/golang/protobuf` dependency enforced to `748d386b5c1ea99658fd69fe9f03991ce86a90c1` to prevent
  `github.com/coreos/etcd` v3.2.4 from fetching `protobuf` version `4bd1920723d7b7c925de087aa32e2187708897f7`,
  which is incompatible with `google.golang.org/grpc` v1.5.0+

### Fixed
- tests failing after upgrade


## [4.3.3] - [GOLIBS-1454](https://jira.lzd.co/browse/GOLIBS-1454) - 2017-08-08

### Added
- CreatedNotify for etcd Watch request. In case of network problems, warnings in logs will appear
every 30 seconds on each Watch attempt after failure.

## [4.3.2] - [GOLIBS-1380](https://jira.lzd.co/browse/GOLIBS-1380) - 2017-08-08

### Added
- ProgressNotify for etcd Watch request. Now we can find network problems at least every
20 minutes.


## [4.3.1] - [GOLIBS-1344](https://jira.lzd.co/browse/GOLIBS-1344) - 2017-07-19

### Fixed
- Watch freeze if connected to etcd node, which lost leader in past


## [4.3.0] - [GOLIBS-1354](https://jira.lzd.co/browse/GOLIBS-1354) - 2017-07-13

### Fixed
- panic in grpc.Dial, because it closed Balancer after error


## [4.2.0] - [GOLIBS-1211](https://jira.lzd.co/browse/GOLIBS-1211) - 2017-07-07

### Added
- **resource** registration parameters contain optional fields for setting admin or metrics values


## [4.1.0] - [GOLIBS-1259](https://jira.lzd.co/browse/GOLIBS-1259) - 2017-07-06

### Added
- ExportedEntities []string into application registration parameters. It's a
special optional field for DataSync API integration.
- discovery provider can register any key using `KV.RawKey` field. But still, there's a
whitelist of namespaces for watching the key changes - we want to validate the data somehow.
Without this whitelist we will never be able to separate **invalid** data from **custom** one.
This behavior may be changed in future.


## 4.0.3 - [GOLIBS-1253](https://jira.lzd.co/browse/GOLIBS-1253) - 2017-06-23

### Fixed
- registrator.NewAdminInfoFromString() - fixed data unmarshaling


## 4.0.2 - [GOLIBS-1133](https://jira.lzd.co/browse/GOLIBS-1133) - 2017-05-24

### Fixed
- decrease log level for rollout balancer warning


## 4.0.1 - [GOLIBS-1073](https://jira.lzd.co/browse/GOLIBS-1073) - 2017-05-19

### Fixed
- service name length is limited with 30 chars


## 4.0.0 - [GOLIBS-1065](https://jira.lzd.co/browse/GOLIBS-1065) - 2017-05-12

Please, refer to [UPGRADE.md](./UPGRADE.md) for migration guide.

### Added
- `balancer.NewFallbackBalancerEtcd2` helper for creating fallback etcd3-etcd2 balancer
- comments and other codestyle stuff to pass gohint (finally!)

### Changed
- etcd2 registration becomes private
- etcd2 locator moved to "locator" subpackage
- ILogger interface and constants are moved to the lib core directory
- discovery provider related code moved to "provider" subpackage
- `registrator.NewV2V3` takes etcd.Client instance instead IServiceRegistrator2
- `balancer.LoadBalancer` interface is renamed to ILoadBalancer
- `balancer.NewRoundRobin` and `balancer.NewWeightedRoundRobin` signatures changed to work with IProvider
- `balancer.NewRoundRobinBalancerFabric` and `balancer.NewRoundRobinBalancerRegestry` changed to work with fallback balancer

### Removed
- discovery/discovery subpackage
- `LoadBalancer.NextN()` method
- deprecated code


## 3.5.3 - [GOLIBS-1253](https://jira.lzd.co/browse/GOLIBS-1253) - 2017-06-23

### Fixed
- registrator.NewAdminInfoFromString() - fixed data unmarshaling (4.0.3 backport)


## 3.5.2 - [GOLIBS-1133](https://jira.lzd.co/browse/GOLIBS-1133) - 2017-05-24

### Fixed
- decrease log level for rollout balancer warning (4.0.2 backport)


## 3.5.1 - [GOLIBS-1040](https://jira.lzd.co/browse/GOLIBS-1040) - 2017-05-03

### Fixed
- unstable service won't register in etcd2 anymore


## 3.5.0 - [GOLIBS-1043](https://jira.lzd.co/browse/GOLIBS-1043) - 2017-05-02

### Added
- Stats method for `grpc.Balancer`


## 3.4.3 - [GOLIBS-985](https://jira.lzd.co/browse/GOLIBS-985) - 2017-04-21

### Added
- tests for `grpc.Balancer`


## 3.4.2 - [GOLIBS-982](https://jira.lzd.co/browse/GOLIBS-982) - 2017-04-17

### Fixed
- decrease log level for RolloutBalancer message for wrong segregation ID header


## 3.4.1 - [GOLIBS-954](https://jira.lzd.co/browse/GOLIBS-954) - 2017-04-13

### Fixed
- WeightedRoundRobin for using in grpc.Balancer


## 3.4.0 - [GOLIBS-961](https://jira.lzd.co/browse/GOLIBS-961) - 2017-04-12

### Added
- Extracted rollout watcher to standalone type (for performance optimization)


## 3.3.1 - [GOLIBS-959](https://jira.lzd.co/browse/GOLIBS-959) - 2017-04-11

### Fixed
- check balancer.initTimeout only once (in RR and WRR balancers)


## 3.3.0 - [GOLIBS-932](https://jira.lzd.co/browse/GOLIBS-932) - 2017-04-03

### Added
- LoadBalancer2 interface supporting service name getter


## 3.2.0 - [GOLIBS-720](https://jira.lzd.co/browse/GOLIBS-720) - 2017-03-30

### Added
- `balancer.IRolloutBalancer` interface for rollout-based service balancing


## 3.1.0 - [GOLIBS-667](https://jira.lzd.co/browse/GOLIBS-667) - 2017-03-16

### Added
- native gRPC balancer on top of balancer.LoadBalancer interface


## 3.0.2 - [GOLIBS-863](https://jira.lzd.co/browse/GOLIBS-863) - 2017-03-13

### Changed
- bump etcd3 version


## 3.0.1 - [GOLIBS-861](https://jira.lzd.co/browse/GOLIBS-861) - 2017-03-13

### Fixed
- added workaround for etcd3.Leaser buggy KeepAlive method


## 3.0.0 - [GOLIBS-523](https://jira.lzd.co/browse/GOLIBS-523) - 2017-02-16

### Added
- resource registration support along with applications
- unit test
- example app

### Changed
- etcd3 key scheme changed
- discovery value scheme changed


## 2.2.x - 2017-01-12

### Added
- etcd3 registration and service location interfaces:
    - IProvider
    - IRegistrator
    - ILocator
- etcd3 locator support for balancers

### Changed
- rewritten Balancers internals - no more etcd2 dependency
