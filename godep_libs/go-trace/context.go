package gotrace

import (
	"context"
	"errors"

	"github.com/opentracing/opentracing-go"
)

var (
	ErrNoSpanInContext = errors.New("go-trace: No `opentracing.Span` in context")
)

const (
	// IsAccessLogSpanTag is a special tag for Span. If it sets Span will be written in special format:
	// `[access log] {transaction_name} - {request_data}` (according to agreement)
	// https://confluence.lzd.co/display/DEV/Microservice+Architecture+%28SOA%29+Conventions#MicroserviceArchitecture(SOA)Conventions-Accesslogs
	IsAccessLogSpanTag = "is.access.log"
	// RequestDataSpanTag is a tag for setting {request_data} part of access log message
	RequestDataSpanTag = "request.data"
	// LogLevelSpanTag sets log level for Span (default: DEBUG)
	LogLevelSpanTag    = "log.level"
)

// StartSpanFromRequest() starts new `opentracing.Span` with `operationName`,
// using SpanContext extracted from `request` (can be `http.Header` or protobuf `metadata.MD`)
// as a ChildOfRef. If no such parent could be found, StartSpanFromRequest() creates a root (parentless) Span.
// It returns `opentracing.Span` and new `context.Context` with started `opentracing.Span`
// N.B. This function automatically sets IsAccessLogSpanTag to true
//
// Before call you should init tracer like this:
//
//     opentracing.SetGlobalTracer(gotrace.NewTracer(...))
//
// Example usage (start span from http.Header):
//
//     span, ctx := gotrace.StartSpanFromRequest(ctx, r.URL.Path, opentracing.HTTPHeadersCarrier(r.Header),
//         gotrace.RequestDataSpanOption(r.URL.RequestURI())))
//     defer span.Finish()
//
// Example usage (start span from protobuf metadata.MD):
//
//     md, _ := metadata.FromContext(ctx)
//     span, ctx := gotrace.StartSpanFromRequest(ctx, "method", gotrace.ProtoMetadataCarrier(md))
//     defer span.Finish()
//
func StartSpanFromRequest(ctx context.Context, operationName string, request Carrier,
	opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {

	if parentSpanContext, err := opentracing.GlobalTracer().Extract(opentracing.TextMap,
		RequestCarrier{Carrier: request}); err == nil {

		opts = append(opts, opentracing.ChildOf(parentSpanContext))
	}
	span, ctx := startSpan(ctx, operationName, opts...)
	span.SetTag(IsAccessLogSpanTag, true)
	return span, ctx
}

// StartAccessLogSpan() wraps StartSpanFromRequest with span option RequestDataSpanOption
// Deprecated since 5.*. use StartSpanFromRequest and RequestDataSpanOption instead:
//
// Example usage:
//
//     // NOT recommended style
//     span, ctx := gotrace.StartAccessLogSpan(ctx, r.URL.Path, r.URL.RequestURI(),
//         opentracing.HTTPHeadersCarrier(r.Header),
//
//     // recommended style
//     span, ctx := gotrace.StartSpanFromRequest(ctx, r.URL.Path, opentracing.HTTPHeadersCarrier(r.Header),
//         gotrace.RequestDataSpanOption(r.URL.RequestURI())))
//
//     // or if you don't want to set RequestDataSpanOption just:
//     span, ctx := gotrace.StartSpanFromRequest(ctx, r.URL.Path, opentracing.HTTPHeadersCarrier(r.Header))
//
//     // also you can attach RequestDataSpanTag for existing span, if you don't know requestData when you start Span:
//     span.SetTag(RequestDataSpanTag, requestData)
//
func StartAccessLogSpan(ctx context.Context, operationName string, requestData string, request Carrier,
	opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {

	span, ctx := StartSpanFromRequest(ctx, operationName, request, opts...)

	span.SetTag(IsAccessLogSpanTag, true)
	span.SetTag(RequestDataSpanTag, requestData)

	return span, ctx
}

// StartSpanFromContext() starts new `opentracing.Span` with `operationName`,
// using any Span found within `ctx` as a ChildOfRef. If no such parent could be found,
// StartSpanFromContext() creates a root (parentless) Span.
// It returns `opentracing.Span` and new `context.Context` with started `opentracing.Span`
//
// Before call you should init tracer like this:
//
//     opentracing.SetGlobalTracer(gotrace.NewTracer(...))
//
// Example usage:
//
//     span, ctx := gotrace.StartSpanFromContext(ctx, "operationName")
//     defer span.Finish()
//
func StartSpanFromContext(ctx context.Context, operationName string,
	opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {

	if sc := spanContext(ctx); sc != nil {
		opts = append(opts, opentracing.ChildOf(sc))
	}
	return startSpan(ctx, operationName, opts...)
}

// InjectSpanToResponseFromContext injects `opentracing.Span` from `context.Context` into response
//
// Example usage (inject into http.Header):
//
//     gotrace.InjectSpanToResponseFromContext(ctx, opentracing.HTTPHeadersCarrier(r.Header))
//
// Example usage (inject into protobuf metadata.MD):
//
//     gotrace.InjectSpanToResponseFromContext(ctx, gotrace.ProtoMetadataCarrier(md))
//
func InjectSpanToResponseFromContext(ctx context.Context, carrier Carrier) error {
	if sc := spanContext(ctx); sc != nil {
		return InjectSpanToResponse(sc, carrier)
	}
	return ErrNoSpanInContext
}

// InjectSpanToRequestFromContext injects `opentracing.Span` from `context.Context` into request
//
// Example usage (inject into http.Header):
//
//     gotrace.InjectSpanToRequestFromContext(ctx, opentracing.HTTPHeadersCarrier(r.Header))
//
// Example usage (inject into protobuf metadata.MD):
//
//     gotrace.InjectSpanToRequestFromContext(ctx, gotrace.ProtoMetadataCarrier(md))
//
func InjectSpanToRequestFromContext(ctx context.Context, carrier Carrier) error {
	if sc := spanContext(ctx); sc != nil {
		return InjectSpanToRequest(sc, carrier)
	}
	return ErrNoSpanInContext
}

// InjectSpanToResponse injects `opentracing.Span` into response
//
// Example usage (inject into http.Header):
//
//     gotrace.InjectSpanToResponse(span.Context(), opentracing.HTTPHeadersCarrier(r.Header))
//
// Example usage (inject into protobuf metadata.MD):
//
//     gotrace.InjectSpanToResponse(span.Context(), gotrace.ProtoMetadataCarrier(md))
//
func InjectSpanToResponse(sc opentracing.SpanContext, carrier Carrier) error {
	return opentracing.GlobalTracer().Inject(sc, opentracing.TextMap,
		ResponseCarrier{Carrier: carrier})
}

// InjectSpanToRequest injects `opentracing.Span` into request
//
// Example usage (inject into http.Header):
//
//     gotrace.InjectSpanToRequest(span.Context(), opentracing.HTTPHeadersCarrier(r.Header))
//
// Example usage (inject into protobuf metadata.MD):
//
//     gotrace.InjectSpanToRequest(span.Context(), gotrace.ProtoMetadataCarrier(md))
//
func InjectSpanToRequest(sc opentracing.SpanContext, carrier Carrier) error {
	return opentracing.GlobalTracer().Inject(sc, opentracing.TextMap,
		RequestCarrier{Carrier: carrier})
}

func spanContext(ctx context.Context) opentracing.SpanContext {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		return span.Context()
	}
	return nil
}

func startSpan(ctx context.Context, name string,
	opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {

	span := opentracing.StartSpan(name, opts...)
	return span, opentracing.ContextWithSpan(ctx, span)
}
