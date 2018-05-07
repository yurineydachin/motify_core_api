package log

import (
	"fmt"
	"sync/atomic"

	logfmt "motify_core_api/godep_libs/go-log/format"
)

// Span request information
type Span struct {
	TraceID      string
	ParentSpanID string
	SpanID       string
	RolloutType  string
}

// Log writes log in Std format
func (l *Logger) Log(callStackSkip int, span Span, level Severity, message string, data map[string]interface{}) {
	f := &logfmt.Std{
		CallStackSkip: callStackSkip,
		TraceId:       span.TraceID,
		SpanId:        span.SpanID,
		ParentSpanId:  span.ParentSpanID,
		RolloutType:   span.RolloutType,
		Message:       message,
		Data:          data,
	}
	f.IncExtCallStackSkip(1)
	f.SetBacktraceSkips(l.opts.backtraceSkips)
	l.Record(level, f)
}

// Logf writes log in Std format
func (l *Logger) Logf(callStackSkip int, span Span, level Severity, format string,
	data map[string]interface{}, a ...interface{},
) {
	if level > Severity(atomic.LoadInt64(&l.opts.level)) {
		return
	}

	f := &logfmt.Std{
		CallStackSkip: callStackSkip,
		TraceId:       span.TraceID,
		SpanId:        span.SpanID,
		ParentSpanId:  span.ParentSpanID,
		RolloutType:   span.RolloutType,
		Message:       fmt.Sprintf(format, a...),
		Data:          data,
	}
	f.IncExtCallStackSkip(1)
	f.SetBacktraceSkips(l.opts.backtraceSkips)
	l.Record(level, f)
}
