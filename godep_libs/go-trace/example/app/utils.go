package app

import (
	"github.com/opentracing/basictracer-go"
	"motify_core_api/godep_libs/go-trace"
	"strconv"
	"github.com/openzipkin/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go-opentracing/types"
)

// BasictracerRecorder is a wrapper for the Appdash recorder. Appdash uses a basic
// opentracing tracer (which can be easily replaced by our own tracer), so the only
// thing we need is a mechanism to translate *our* spans to those used by Appdash.
type BasictracerRecorder struct {
	basictracer.SpanRecorder
}

// NewBasictracerRecorder is a constructor for BasictracerRecorder.
func NewBasictracerRecorder(spanRecorder basictracer.SpanRecorder) *BasictracerRecorder {
	return &BasictracerRecorder{
		SpanRecorder: spanRecorder,
	}
}

// RecordSpan translates `gotrace` spans to `Appdash` spans and call parent's `RecordSpan()`
// method.
func (r *BasictracerRecorder) RecordSpan(span gotrace.RawSpan) {
	basicSpanContext := basictracer.SpanContext{
		TraceID: SpanIDToUint64(span.Context.TraceID),
		SpanID:  SpanIDToUint64(span.Context.SpanID),
		Sampled: true,
		Baggage: span.Context.Baggage,
	}
	basicRawSpan := basictracer.RawSpan{
		Context:      basicSpanContext,
		ParentSpanID: SpanIDToUint64(span.Context.ParentSpanID),
		Operation:    span.Operation,
		Start:        span.Start,
		Duration:     span.Duration,
		Tags:         span.Tags,
		Logs:         span.Logs,
	}
	r.SpanRecorder.RecordSpan(basicRawSpan)
}

func SpanIDToUint64(spanID string) (out uint64) {
	out, _ = strconv.ParseUint(spanID, 16, 64)
	return
}

type ZipkinRecorder struct {
	zipkintracer.SpanRecorder
}

func NewZipkinRecorder(spanRecorder zipkintracer.SpanRecorder) *ZipkinRecorder {
	return &ZipkinRecorder{
		SpanRecorder: spanRecorder,
	}
}

func (r *ZipkinRecorder) RecordSpan(span gotrace.RawSpan) {
	spanContext := zipkintracer.SpanContext{
		TraceID: types.TraceID{
			Low: SpanIDToUint64(span.Context.TraceID),
		},
		SpanID:  SpanIDToUint64(span.Context.SpanID),
		Sampled: true,
		Baggage: span.Context.Baggage,
		Owner: true,
	}
	if id := SpanIDToUint64(span.Context.ParentSpanID); id > 0 {
		spanContext.ParentSpanID = &id
	}
	rawSpan := zipkintracer.RawSpan{
		Context:      spanContext,
		Operation:    span.Operation,
		Start:        span.Start,
		Duration:     span.Duration,
		Tags:         span.Tags,
		Logs:         span.Logs,
	}
	r.SpanRecorder.RecordSpan(rawSpan)
}