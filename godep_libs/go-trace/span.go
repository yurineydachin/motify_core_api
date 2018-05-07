package gotrace

import (
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

// Span request information
type Span struct {
	TraceID        string
	SpanID         string
	ParentSpanID   string
	CurrentAppInfo AppInfo
	ParentAppInfo  AppInfo
	SegregationID  string
	ForceDebugLog  string
	Baggage        map[string]string
}

type AppInfo struct {
	ForwardedApps string
	AppName       string
	AppVersion    string
	Node          string
}

// ForeachBaggageItem belongs to the opentracing.SpanContext interface
func (c Span) ForeachBaggageItem(handler func(k, v string) bool) {
	for k, v := range c.Baggage {
		if !handler(k, v) {
			break
		}
	}
}

// WithBaggageItem returns an entirely new SpanContext with the
// given key:value baggage pair set.
func (c Span) WithBaggageItem(key, val string) Span {
	var newBaggage map[string]string
	if c.Baggage == nil {
		newBaggage = map[string]string{key: val}
	} else {
		newBaggage = make(map[string]string, len(c.Baggage)+1)
		for k, v := range c.Baggage {
			newBaggage[k] = v
		}
		newBaggage[key] = val
	}
	return Span{
		c.TraceID,
		c.SpanID,
		c.ParentSpanID,
		c.CurrentAppInfo,
		c.ParentAppInfo,
		c.SegregationID,
		c.ForceDebugLog,
		newBaggage,
	}
}

type RawSpan struct {
	Context   Span
	Operation string
	Start     time.Time
	Duration  time.Duration
	Tags      opentracing.Tags
	Logs      []opentracing.LogRecord
}

// spanImpl implements opentracing.Span
type spanImpl struct {
	sync.Mutex
	raw    RawSpan
	tracer *Tracer
}

func (s *spanImpl) SetOperationName(operationName string) opentracing.Span {
	s.Lock()
	defer s.Unlock()
	s.raw.Operation = operationName
	return s
}

func (s *spanImpl) SetTag(key string, value interface{}) opentracing.Span {
	s.Lock()
	defer s.Unlock()

	if s.raw.Tags == nil {
		s.raw.Tags = opentracing.Tags{}
	}

	if t, ok := value.(*MetricSpanOption); ok {
		key = t.name
	}

	s.raw.Tags[key] = value
	return s
}

func (s *spanImpl) LogKV(keyValues ...interface{}) {
	fields, err := log.InterleavedKVToFields(keyValues...)
	if err != nil {
		s.LogFields(log.Error(err), log.String("function", "LogKV"))
		return
	}
	s.LogFields(fields...)
}

func (s *spanImpl) appendLog(lr opentracing.LogRecord) {
	s.raw.Logs = append(s.raw.Logs, lr)
}

func (s *spanImpl) LogFields(fields ...log.Field) {
	lr := opentracing.LogRecord{
		Fields: fields,
	}
	s.Lock()
	defer s.Unlock()
	if lr.Timestamp.IsZero() {
		lr.Timestamp = time.Now()
	}
	s.appendLog(lr)
}

func (s *spanImpl) Finish() {
	s.FinishWithOptions(opentracing.FinishOptions{})
}

func (s *spanImpl) FinishWithOptions(opts opentracing.FinishOptions) {
	finishTime := opts.FinishTime
	if finishTime.IsZero() {
		finishTime = time.Now()
	}
	duration := finishTime.Sub(s.raw.Start)

	s.Lock()
	defer s.Unlock()

	for _, lr := range opts.LogRecords {
		s.appendLog(lr)
	}

	s.raw.Duration = duration

	for _, r := range s.tracer.options.recorders {
		r.RecordSpan(s.raw)
	}
}

func (s *spanImpl) Tracer() opentracing.Tracer {
	return s.tracer
}

func (s *spanImpl) Context() opentracing.SpanContext {
	return s.raw.Context
}

func (s *spanImpl) SetBaggageItem(key, val string) opentracing.Span {
	s.Lock()
	defer s.Unlock()
	s.raw.Context = s.raw.Context.WithBaggageItem(key, val)
	return s
}

func (s *spanImpl) BaggageItem(key string) string {
	s.Lock()
	defer s.Unlock()
	return s.raw.Context.Baggage[key]
}

// SpanContext returns Context from opentracing.Span
func SpanContext(span opentracing.Span) (Span, error) {
	gotraceSpan := Span{}

	if span == nil {
		return gotraceSpan, nil
	}

	if sc, ok := span.Context().(Span); ok {
		return sc, nil
	} else {
		return gotraceSpan, opentracing.ErrInvalidSpanContext
	}

	return gotraceSpan, nil
}

// Deprecated
func (s *spanImpl) LogEvent(event string) {
}

// Deprecated
func (s *spanImpl) LogEventWithPayload(event string, payload interface{}) {
}

// Deprecated
func (s *spanImpl) Log(ld opentracing.LogData) {
}
