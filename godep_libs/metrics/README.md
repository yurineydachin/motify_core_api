# Metrics for Prometheus #

This library allows you using Prometheus monitoring in the simplest way.
Per se it is just only common place for registering metrics for Prometheus 
Go client. With the names that described in Lazada monitoring standard.
Frankly using Prometheus with Go apps doesn't require any additional
libs beside the standard client library. So there is one reason for this 
package: declare the metrics in a single place for minimizing typing errors 
with metric names and tags in different services. It just reduces maintainance 
cost for the support of the same metrics in a bunch of microservices.

Also the library offers minimal set of helpers for more convenient work with
functions and structs of Prometheus client.

The library conforms [Lazada Prometheus Metrics Naming Standard (v2.0)](https://confluence.lazada.com/x/OihVAQ).

## This library doesn't offer ##

* Instrumenting your code with the metrics
* New wrapping structs around the standard client
* Interfaces that you must to implement 
* Registering your service in etcd

Because there is no reason here to convert simple things to complex ones.

## How to use ##

You just use methods from standard client but firstly import this library
instead of standard client. You import only subpackages for systems that
really need monitoring in your service.

```
// For example your service offers HTTP handlers and
// uses MySQL. So you import only subpackages for HTTP and MySQL metrics.
import (
	"godep.lzd.co/metrics"
	"godep.lzd.co/metrics/httpmon"
	"godep.lzd.co/metrics/mysqlmon"
)

// Example for middleware how to insert RT metric for HTTP.
func exampleHTTPMiddleWare() {
  started := time.Now()

  // ... your smart business logic here

  httpmon.ResponseTime.WithLabelValues(statusCode, handlerTag, req.Header.Get("X-LZD-App-Name")).Observe(metrics.SinceMs(started))
}

// Example for MySQL layer.
func exampleDatabaseManager() {
  // ...
  
  mysqlmon.OpenConnectionsTotal.WithLabelValues(dbHost, dbName).Inc()
  started := time.Now()
  
  // ... your smart business logic here

  mysqlmon.ResponseTime.WithLabelValues(dbHost, dbName, isError, query).Observe(metrics.SinceMs(started))
}

```

See more docs at https://godoc.org/github.com/prometheus/client_golang/prometheus


### How to define a new custom metrics? ###

Lazada Monitoring Standards cover the most of the metrics that you need in SOA service
but of course not cover all possible cases. So then you need more you can declare them
with help of `metrics` library:

```
// It is not requirement but just a recommendation for code clarity
// to move all declarations of custom metrics to a separate package.
package mymetrics

import (
  "godep.lzd.co/metrics"
)

var (
  registry = metrics.DefaultRegistry()
  SomethingInterestingTotal = registry.NewGauge("nosql_suspicious_connections_total", "The metric that measure something.")

  SomethingCompletelyDifferentMilliseconds = registry.NewHistogramVec("ourhandler_lost_time", "The metrics that measure something completely different (in ms).", "host", "i_like_this_tag")
)

// Well in rare case you may want separate metrics with the same names for another Prometheus, so you will need custom registry.
// In the most of use cases just forget about custom registry, use default one.
var 
  registry = metrics.NewRegistry()
  MySomethingInterestingTotal = registry.NewGauge("nosql_suspicious_connections_total", "The metric that measure something.")

  MySomethingCompletelyDifferentMilliseconds = registry.NewHistogramVec("ourhandler_lost_time", "The metrics that measure something completely different (in ms).", "host", "i_like_this_tag")  
)
```

Then in other packages of your service you will need import declarations above:

```
import (
  "godep.lzd.co/myservice/mymetrics"
)

func () {
  mymetrics.SomethingCompletelyDifferentMilliseconds.WithLabelValues(hostTag, anotherTag).Observe(metrics.Ms(theMeasuredTime))

  // ...

  // The metric for the default registry.
  mymetrics.SomethingInterestingTotal.Inc()
  // The same metric but for your custom registry.
  mymetrics.MySomethingInterestingTotal.Inc()
}
```

Alternatively you can just use functions from original Prometheus client for declarations
in the default registry.

## Support in other libraries ##

There are the libraries in Lazada bitbucket instrumented with `metrics`:

* [aerocache](https://bitbucket.lzd.co/projects/GOLIBS/repos/aerocache) — lite wrapper for Aerospike client
* [bytecache](https://bitbucket.lzd.co/projects/GOLIBS/repos/bytecache) — the fast realization of bytecache (cache for serialized data)
* [structcache](https://bitbucket.lzd.co/projects/GOLIBS/repos/structcache) — the realization of structcache with LRU

Just include them into your project and build with tags like `go build -tags=metrics`. 
So they will be included with Prometheus integration.
