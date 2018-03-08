package gotrace

import (
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/metadata"
)

const (
	// TraceIDHeader header name containing TraceID
	TraceIDHeader = "x-lzd-traceid"

	// SpanIDHeader header name containing SpanID
	SpanIDHeader = "x-lzd-spanid"

	// ParentSpanIDHeader parentSpanIDHeader
	ParentSpanIDHeader = "x-lzd-parentspanid"

	// ForwardedAppsHeader list
	ForwardedAppsHeader = "x-lzd-forwarded-apps"

	// AppNameH§eader appNameHeader
	AppNameHeader = "x-lzd-app-name"

	// AppVersionHeader appVersionHeader
	AppVersionHeader = "x-lzd-app-version"

	// NodeHeader nodeHeader
	NodeHeader = "x-lzd-node"

	// SegregationIDHeader keeps value ["1":"1000"] from segregation cookie for progressive service rollout
	SegregationIDHeader = "x-lzd-segregationid"

	// ForceDebugLogHeader [0,1] for enabling the most verbose level of logs for the request
	ForceDebugLogHeader = "x-lzd-force-debug-log"
)

const (
	baggagePrefixHeader    = "x-lzd-baggage-"
	forwardedAppsSeparator = ","
)

type ProtoMetadataCarrier metadata.MD

// Set conforms to the TextMapWriter interface.
func (c ProtoMetadataCarrier) Set(key, val string) {
	c[key] = metadata.Pairs(key, val)[key]
}

// ForeachKey conforms to the TextMapReader interface.
func (c ProtoMetadataCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range c {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

type Carrier interface {
	opentracing.TextMapReader
	opentracing.TextMapWriter
}

// RequestCarrier knows how to cook trace info for request from current SpanContext
type RequestCarrier struct {
	Carrier
}

// ResponseCarrier knows how to cook trace info for response from current SpanContext
type ResponseCarrier struct {
	Carrier
}

func (c RequestCarrier) Inject(spanContext Span) error {
	return injectRequest(spanContext, c.Carrier)
}

func (c RequestCarrier) Extract() (Span, error) {
	return extractRequest(c.Carrier)
}

func (c ResponseCarrier) Inject(spanContext Span) error {
	return injectResponse(spanContext, c.Carrier)
}

func (c ResponseCarrier) Extract() (Span, error) {
	return extractResponse(c.Carrier)
}

var _ opentracing.Tracer = &Tracer{}

type Tracer struct {
	options *tracerOptions
}

// NewTracer create tracer. It implements `opentracing.Tracer` interface
//
// After creation it should be inited, for example:
//
//    logger = log.NewLogger(...)
//    goLogCollector := golog.NewSpanCollector(logger)
//    goTraceSpanRecorder := gotrace.NewRecorder(
//        gotrace.WithLogCollector(goLogCollector)
//    )
//
//    tracer := gotrace.NewTracer(
//        gotrace.WithAppEnv("service", "1.0.0", "localhost"),
//        gotrace.WithSpanRecorder(goTraceSpanRecorder),
//        gotrace.WithSpanRecorder(/* your span recorder */), // you can use as many SpanRecorder as you want
//    )
//
//    opentracing.SetGlobalTracer(tracer)
//
func NewTracer(options ...TracerOption) *Tracer {
	opts := &tracerOptions{}
	for _, o := range options {
		o(opts)
	}
	return &Tracer{options: opts}
}

// StartSpan implements the `opentracing.SpanContext` interface.
func (t *Tracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	sso := opentracing.StartSpanOptions{}
	for _, o := range opts {
		o.Apply(&sso)
	}
	return t.StartSpanWithOptions(operationName, sso)
}

func (t *Tracer) StartSpanWithOptions(operationName string, opts opentracing.StartSpanOptions) opentracing.Span {
	startTime := opts.StartTime
	if startTime.IsZero() {
		startTime = time.Now()
	}

	sp := &spanImpl{
		tracer: t,
		raw: RawSpan{
			Start:     startTime,
			Operation: operationName,
			Tags:      opts.Tags,
		},
	}

loop:
	for _, ref := range opts.References {
		switch ref.Type {
		case opentracing.ChildOfRef, opentracing.FollowsFromRef:
			if refCtx, ok := ref.ReferencedContext.(Span); ok {
				sp.raw.Context = refCtx
				sp.raw.Context.ParentSpanID = refCtx.SpanID

				if l := len(refCtx.Baggage); l > 0 {
					sp.raw.Context.Baggage = make(map[string]string, l)
					for k, v := range refCtx.Baggage {
						sp.raw.Context.Baggage[k] = v
					}
				}
			}
			break loop
		}
	}

	sp.raw.Context.CurrentAppInfo.AppName = t.options.appName
	sp.raw.Context.CurrentAppInfo.AppVersion = t.options.appVersion
	sp.raw.Context.CurrentAppInfo.Node = t.options.appNode

	sp.raw.Context.CurrentAppInfo.ForwardedApps = sp.raw.Context.ParentAppInfo.ForwardedApps
	if sp.raw.Context.CurrentAppInfo.ForwardedApps != "" {
		sp.raw.Context.CurrentAppInfo.ForwardedApps += forwardedAppsSeparator
	}
	sp.raw.Context.CurrentAppInfo.ForwardedApps += t.options.appName

	sp.raw.Context.SpanID = generator.Generate()
	if sp.raw.Context.TraceID == "" {
		sp.raw.Context.TraceID = sp.raw.Context.SpanID
	}

	return sp
}

type injector interface {
	Inject(spanContext Span) error
}

func (t *Tracer) Inject(spanContext opentracing.SpanContext, format interface{}, carrier interface{}) error {
	sc, ok := spanContext.(Span)
	if !ok {
		return opentracing.ErrInvalidSpanContext
	}
	i, ok := carrier.(injector)
	if !ok {
		if c, ok := carrier.(Carrier); ok {
			i = RequestCarrier{Carrier: c}
		} else {
			return opentracing.ErrInvalidCarrier
		}
	}
	return i.Inject(sc)
}

type extractor interface {
	Extract() (Span, error)
}

func (t *Tracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	e, ok := carrier.(extractor)
	if !ok {
		if c, ok := carrier.(Carrier); ok {
			e = RequestCarrier{Carrier: c}
		} else {
			return nil, opentracing.ErrInvalidCarrier
		}
	}
	return e.Extract()
}

func injectRequest(spanContext Span, carrier opentracing.TextMapWriter) error {
	carrier.Set(TraceIDHeader, spanContext.TraceID)
	carrier.Set(ParentSpanIDHeader, spanContext.SpanID)
	carrier.Set(ForwardedAppsHeader, spanContext.CurrentAppInfo.ForwardedApps)
	carrier.Set(AppNameHeader, spanContext.CurrentAppInfo.AppName)
	carrier.Set(AppVersionHeader, spanContext.CurrentAppInfo.AppVersion)
	carrier.Set(NodeHeader, spanContext.CurrentAppInfo.Node)
	carrier.Set(SegregationIDHeader, spanContext.SegregationID)
	carrier.Set(ForceDebugLogHeader, spanContext.ForceDebugLog)

	// Baggage:
	for baggageKey, baggageVal := range spanContext.Baggage {
		safeVal := baggageVal
		carrier.Set(baggagePrefixHeader+baggageKey, safeVal)
	}
	return nil
}

func extractRequest(carrier opentracing.TextMapReader) (Span, error) {
	spanContext := Span{}
	empty := true
	err := carrier.ForeachKey(func(key, val string) error {
		lowerKey := strings.ToLower(key)
		switch lowerKey {
		case TraceIDHeader:
			spanContext.TraceID, empty = val, false
		case ParentSpanIDHeader:
			spanContext.SpanID, empty = val, false
		case ForwardedAppsHeader:
			spanContext.ParentAppInfo.ForwardedApps, empty = val, false
		case AppNameHeader:
			spanContext.ParentAppInfo.AppName, empty = val, false
		case AppVersionHeader:
			spanContext.ParentAppInfo.AppVersion, empty = val, false
		case NodeHeader:
			spanContext.ParentAppInfo.Node, empty = val, false
		case SegregationIDHeader:
			spanContext.SegregationID, empty = val, false
		case ForceDebugLogHeader:
			spanContext.ForceDebugLog, empty = val, false
		}
		if strings.HasPrefix(lowerKey, baggagePrefixHeader) {
			// Baggage:
			if spanContext.Baggage == nil {
				spanContext.Baggage = make(map[string]string)
			}
			spanContext.Baggage[lowerKey[len(baggagePrefixHeader):]], empty = val, false
		}
		return nil
	})
	if empty {
		return spanContext, opentracing.ErrSpanContextNotFound
	}
	return spanContext, err
}

func injectResponse(spanContext Span, carrier opentracing.TextMapWriter) error {
	carrier.Set(TraceIDHeader, spanContext.TraceID)
	carrier.Set(SpanIDHeader, spanContext.SpanID)
	carrier.Set(ParentSpanIDHeader, spanContext.ParentSpanID)
	carrier.Set(AppNameHeader, spanContext.CurrentAppInfo.AppName)
	carrier.Set(AppVersionHeader, spanContext.CurrentAppInfo.AppVersion)
	carrier.Set(NodeHeader, spanContext.CurrentAppInfo.Node)
	carrier.Set(SegregationIDHeader, spanContext.SegregationID)
	carrier.Set(ForceDebugLogHeader, spanContext.ForceDebugLog)

	return nil
}

func extractResponse(carrier opentracing.TextMapReader) (Span, error) {
	spanContext := Span{}
	empty := true
	err := carrier.ForeachKey(func(key, val string) error {
		switch strings.ToLower(key) {
		case TraceIDHeader:
			spanContext.TraceID, empty = val, false
		case SpanIDHeader:
			spanContext.SpanID, empty = val, false
		case ParentSpanIDHeader:
			spanContext.ParentSpanID, empty = val, false
		case AppNameHeader:
			spanContext.CurrentAppInfo.AppName, empty = val, false
		case AppVersionHeader:
			spanContext.CurrentAppInfo.AppVersion, empty = val, false
		case NodeHeader:
			spanContext.CurrentAppInfo.Node, empty = val, false
		case SegregationIDHeader:
			spanContext.SegregationID, empty = val, false
		case ForceDebugLogHeader:
			spanContext.ForceDebugLog, empty = val, false
		}
		return nil
	})
	if empty {
		return spanContext, opentracing.ErrSpanContextNotFound
	}
	return spanContext, err
}
