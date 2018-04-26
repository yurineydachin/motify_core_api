package log

import (
	"regexp"

	"godep.lzd.co/go-log/format"
)

// NewSpanCollector creates collector that writes opentracing.Span to log
func NewSpanCollector(l *Logger) *spanCollector {
	return &spanCollector{l}
}

type spanCollector struct {
	l *Logger
}

var pwdFilter = regexp.MustCompile(`password=[^&\s]+`)

// Collect records Span data to log
// if isAccess is true, span writes according to access log agreement
// https://confluence.lzd.co/display/DEV/Microservice+Architecture+%28SOA%29+Conventions#MicroserviceArchitecture(SOA)Conventions-Accesslogs
// message part of log string has the following format: `[access log] {transaction_name} - {request_data}`
// if isAccess is false, message part of log is `operationName`,
// or `operationName - requestData` (if requestData is not empty)
func (r *spanCollector) Collect(
	level, traceID, parentSpanID, spanID, operationName, requestData string,
	additionalData map[string]interface{}, isAccess bool,
) {
	severity := DEBUG
	message := ``
	if isAccess {
		// access log writes always with INFO level
		severity = INFO
		message += `[access log] `
		requestData = pwdFilter.ReplaceAllString(requestData, "password=<HIDDEN_BY_SECURITY_FILTER>")
	} else {
		if level != "" {
			if s, ok := ParseSeverity(level); ok {
				severity = s
			}
		}
		if severity > r.l.Level() {
			return
		}
	}
	message += operationName
	if requestData != "" {
		message += ` - ` + requestData
	}
	f := &format.Std{
		CallStackSkip: NO_STACK_TRACE_INFO,
		TraceId:       traceID,
		SpanId:        spanID,
		ParentSpanId:  parentSpanID,
		Message:       message,
		Data:          additionalData,
	}
	f.EnableSyslogHeader(r.l.syslog)
	f.SetService(r.l.key.service)
	f.SetBacktraceSkips(r.l.opts.backtraceSkips)
	f.SetLevel(severity)
	r.l.writer.WriteFrom(f)
}

// collector implements godep.lzd.co/go-trace.collector
// Deprecated
type collector struct {
	l       *Logger
	node    string
	version string
}

// NewCollector returns collector, it implements godep.lzd.co/go-trace.collector
// Deprecated use NewSpanCollector instead
func NewCollector(l *Logger, node, version string) collector {
	return collector{
		l:       l,
		node:    node,
		version: version,
	}
}

// Collect writes trace information in log
// Deprecated
func (c collector) Collect(traceID, parentSpanID, spanID, message string, data map[string]interface{}) {
	f := &format.Std{
		CallStackSkip: NO_STACK_TRACE_INFO,
		TraceId:       traceID,
		SpanId:        spanID,
		ParentSpanId:  parentSpanID,
		Message:       message,
		Data:          data,
	}
	f.EnableSyslogHeader(c.l.syslog)
	f.SetService(c.l.key.service)
	f.SetLevel(INFO)
	c.l.writer.WriteFrom(f)
}

// Service returns name of service
// Deprecated
func (c collector) Service() string {
	return c.l.key.service
}

// Version returns version of service
// Deprecated
func (c collector) Version() string {
	return c.version
}

// Node returns hostname of service
// Deprecated
func (c collector) Node() string {
	return c.node
}
