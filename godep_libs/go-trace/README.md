# go-trace
`go-trace` lib for tracing requests in SOA. It supports
the OpenTracing API. Please see the
http://opentracing.io/documentation/pages/quick-start for more info.

- [Installation](#Installation)
- [Usage](#Usage)
- [Documentation](#Documentation)

## Installation
Add dependency in your `glide.yaml` and `glide up`:
```
- package: godep.lzd.co/go-trace
```

## Usage
Initialize Tracer at the beginning!
```go
    import "github.com/opentracing/opentracing-go"
    import "godep.lzd.co/go-trace"
    import "godep.lzd.co/go-log"

    func main() {
        logger = log.NewLogger(...)
        goLogCollector := golog.NewSpanCollector(logger)
        goTraceSpanRecorder := gotrace.NewRecorder(
            gotrace.WithLogCollector(goLogCollector)
        )

        tracer := gotrace.NewTracer(
            gotrace.WithAppEnv("service", "1.0.0", "localhost"),
            gotrace.WithSpanRecorder(goTraceSpanRecorder),
            gotrace.WithSpanRecorder(/* your span recorder */), // you can use as many SpanRecorder as you want
        )

        opentracing.SetGlobalTracer(tracer)
        ...
    }
```

Note that you can use as many recorders in `Tracer` constructor as you want, each of them will be used for each span.

Then just use it in accordance with
[opentracing](https://github.com/opentracing/opentracing-go) interface!

Example starting Span:
```go
    func xyz() {
        ...
        sp := opentracing.StartSpan("operation_name")
        defer sp.Finish()
        ...
    }
```

Be informed that when you call `Finish()` your registered `Tracer`
will record `Span`, so if you will not call it, your `Span`
will not be recorded!

`gotrace.Tracker` uses `go-log` as recorder. But you can easily
change it, without modifying any other code.
For more information see [sample application](./example).
It uses different tracer `appdash` or `go-trace`.

Example starting access log Span

With `gotrace` wrapper:
```go
    func foo() {
        ...
        span, ctx := gotrace.StartSpanFromRequest(ctx, "handlerName", opentracing.HTTPHeadersCarrier(r.Header))
        defer span.Finish()
        ...
    }
```

**N.B.:** Pass a `gotrace.RequestDataSpanOption` to `gotrace.StartSpanFromRequest()` trailing arguments if you want to provide request data (the old way to do that, `gotrace.StartAccessLogSpan()`, is now deprecated).

Directly via `opentracing`
```go
    ...
    var opts []opentracing.StartSpanOption
    if parentSpanContext, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, opentracing.HTTPHeadersCarrier(request.Header)); err == nil {
        opts = append(opts, opentracing.ChildOf(parentSpanContext))
    }

    span := opentracing.StartSpan("handlerName", opts...)
    defer span.Finish()

    span.SetTag(gotrace.IsAccessLogSpanTag, true)  // key = "is.access.log"
    span.SetTag(gotrace.RequestDataSpanTag, "requestData")  // key = "request.data"
    ...
```

Where `requestData`: Request body (URL if it's REST GET request). If request body is too big you can skip it or log here only some important information
Read more: https://confluence.lzd.co/display/DEV/Microservice+Architecture+%28SOA%29+Conventions#MicroserviceArchitecture(SOA)Conventions-Accesslogs

#### How to add Prometheus metrics

You can add Prometheus metrics to span constructors. Consider an example of starting a span that counts HTTP requests:

```go
package main

import (
	"net/http"
	
	"godep.lzd.co/go-trace"
	"godep.lzd.co/metrics"
	"github.com/opentracing/opentracing-go"
)

// Create the counter metric.
var (
	registry        = metrics.DefaultRegistry()
	request_counter = registry.NewCounterVec("request_counter", "Counts requests", "status")
)

func HttpTracer(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var status = "200"
		span, ctx := gotrace.StartSpanFromRequest(
			r.Context(), r.URL.Path, opentracing.HTTPHeadersCarrier(r.Header),
			gotrace.NewMetricSpanOption(request_counter, func() []string { // Now `status` is captured.
				return []string{status}
			}))
		defer span.Finish()

		// Call handler.
		h(w, r.WithContext(ctx))
        
		// Set the status code according to app logic.
		status = "300"
	}
}

```

Here we call `gotrace.NewMetricSpanOption()` with a `prometheus.Labels` object. Thus you can easily update labels value labels after option was created.

The following Prometheus metrics are supported:

* `prometheus.CounterVec`
* `prometheus.Counter`
* `prometheus.HistogramVec`
* `prometheus.Histogram`

Histograms will `Observe()` current span execution duration, and counters will be `Inc()`'ed when the span is finished.

N.B.: default histogram precision is `time.Microsecond`. You can change precision by calling `MetricSpanOption.SetResolution` with an appropriate duration. There's also a special constructor `NewMetricsSpanOptionPrecise()` which returns an option with specified precision.

## Documentation
- https://confluence.lzd.co/display/DEV/Microservice+Architecture+%28SOA%29+Conventions#MicroserviceArchitecture(SOA)Conventions-xlzdHowtouseX-LZD-\*headers
- [moving from `4.3.1` to `4.4.0`](#from_4_3_1_to_4_4_0)
- [moving from `4.*` to `5.*`](#from_4_to_5)

## <a name="from_4_3_1_to_4_4_0"></a> Migration guide: moving from `4.3.1` to `4.4.0`
Version `4.4.0` adds OpenTracing Tracker, but it maintains
backward compatibility. So you can use all previous API as usual,
moreover you can even use both style together. But you should
remember that all previous API is deprecated now. And all deprecated
functions will be removed in next major version.

- Register trace information in `context.Context` from `http.Request`
    - Old style
        ```go
            ctx = gotrace.SetTraceHeadersToContext(ctx, r.Header, service, version, hostname)
        ```
    - New style
        ```go
            span, ctx := gotrace.StartSpanFromRequest(ctx, "handlerName", opentracing.HTTPHeadersCarrier(r.Header))
            // finish must be called at the end of processing request
            defer span.Finish()
        ```

- Inject trace information in response
    - Old style
        ```go
            gotrace.SetHeaderResponseFromContext(ctx, w.Header())
        ```

    - New style
        ```go
            gotrace.InjectSpanToResponseFromContext(ctx, opentracing.HTTPHeadersCarrier(w.Header()))
        ```

- Inject trace information in external request
    - Old style
        ```go
            // create ext request
            r, _ := http.NewRequest("GET", "http://server:port", nil)

            // inject the trace information into the HTTP Headers in request
            gotrace.SetHeaderRequestFromContext(ctx, r.Header)

            // do request
            _, _ = http.DefaultClient.Do(r)
        ```
    - New style
        ```go
            // create ext request
            r, _ := http.NewRequest("GET", "http://server:port", nil)

            // inject the trace information into the HTTP Headers in request
            _ = gotrace.InjectSpanToRequestFromContext(ctx, opentracing.HTTPHeadersCarrier(r.Header))

            // do request
            _, _ = http.DefaultClient.Do(r)
        ```

- Using both style together (it's not recommended)
    ```go
        // start span old style
        ctx = gotrace.SetTraceHeadersToContext(ctx, r.Header, service, version, hostname)

        // create ext request new style
        r, _ := http.NewRequest("GET", "http://server:port", nil)
        _ = gotrace.InjectSpanToRequestFromContext(ctx, opentracing.HTTPHeadersCarrier(r.Header))
        _, _ = http.DefaultClient.Do(r)

        // start child span from old one (new style)
        span, ctx := gotrace.StartSpanFromContext(ctx, "someOperation", ctx)
        defer span.Finish()
    ```

    You should remember that if you start Span using old style this Span
    can't be recorded because only new Span can be `Finish()`'ed.
   
## <a name="from_4_to_5"></a> Migration guide: moving from `4.*` to `5.*`
First of all read [moving from `4.3.1` to `4.4.0`](#from_4_3_1_to_4_4_0) then follow the replacemnt guid 
- Instead of `GetGotraceSpan()` use `SpanContext()`
- Instead of `gotrace.FromContext(ctx)` use `gotrace.SpanContext(opentracing.SpanFromContext(ctx))`
- Instead of
    ```go
      logger := golog.NewLogger(...)
      gotrace.NewTracer(golog.NewCollector(logger, hostname, appVersion))
    ```
    use
    ```go
      gotrace.NewTracer(gotrace.WithAppEnv(serviceName, hostname, appVersion))
    ```
    also, if you want record Span use `WithSpanRecorder` option in `NewTracer`
    ```go
      gotrace.WithSpanRecorder(recorder)
    ```
    to create recorder which follow the Lazada agreement, use the following code:
    ```go
      gotrace.NewRecorder(
          gotrace.WithLogCollector(
              golog.NewSpanCollector(logger))))
    ```
    This recorder writes the root Span as access log (https://confluence.lzd.co/display/DEV/Microservice+Architecture+%28SOA%29+Conventions#MicroserviceArchitecture(SOA)Conventions-Logsformat).
    Other spans will be written in log if `LogLevelSpanOption` more than logger.Level()