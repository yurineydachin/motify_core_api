package gotrace

import (
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
)

// RequestDataSpanOption implements `opentracing.StartSpanOption`. It sets `RequestDataSpanTag`
// and `IsAccessLogSpanTag`; the last tag ensures that current span's info will be sent to
// access log.
type RequestDataSpanOption string

// Apply implements `opentracing.StartSpanOption`.
func (o RequestDataSpanOption) Apply(opts *opentracing.StartSpanOptions) {
	if opts.Tags == nil {
		opts.Tags = make(map[string]interface{})
	}
	opts.Tags[RequestDataSpanTag] = string(o)
	opts.Tags[IsAccessLogSpanTag] = true
}

// LogLevelSpanOption sets log.level tag for span. It parses by golog.ParseSeverity
// This option can be used if you use gotrace.Recorder and WithLogCollector option
// Span with this option will be recorder only if log.level > golog.Level()
type LogLevelSpanOption string

func (o LogLevelSpanOption) Apply(opts *opentracing.StartSpanOptions) {
	if opts.Tags == nil {
		opts.Tags = make(map[string]interface{})
	}
	opts.Tags[LogLevelSpanTag] = string(o)
}

// MetricSpanOption implements `opentracing.StartSpanOption` interface. It can also be
// used as a special tag value when calling `gotrace.spanImpl.SetTag`.
type MetricSpanOption struct {
	precision time.Duration
	name      string
	metric    prometheus.Collector
	labels    prometheus.Labels
}

// NewMetricSpanOption is a constructor for `MetricSpanOption`. `metric` should be any of the following
// `prometheus` metrics:
//		- `prometheus.CounterVec`
//		- `prometheus.Counter`
//		- `prometheus.HistogramVec`
// 		- `prometheus.Histogram`.
// `labelsCb` is a callback that should return a slice of labels for the `metric` (if
// it uses any labels) or a `nil` value if labels are not used.
func NewMetricSpanOption(metric prometheus.Collector, labels prometheus.Labels) *MetricSpanOption {
	return &MetricSpanOption{
		precision: time.Microsecond,
		name:      time.Now().String(),
		metric:    metric,
		labels:    labels,
	}
}

func NewMetricSpanOptionPrecise(metric prometheus.Collector, labels prometheus.Labels, precision time.Duration) *MetricSpanOption {
	out := NewMetricSpanOption(metric, labels)
	out.SetPrecision(precision)
	return out
}

// Apply implements `opentracing.StartSpanOption`.
func (t *MetricSpanOption) Apply(opts *opentracing.StartSpanOptions) {
	if opts.Tags == nil {
		opts.Tags = make(map[string]interface{})
	}
	opts.Tags[t.name] = t
}

// Update updates the `metric` according to its type. It `metric` is from the counters family,
// `val` argument will be ignored.
func (t *MetricSpanOption) Update(val time.Duration) {
	val64 := float64(val) / float64(t.precision)
	switch metric := t.metric.(type) {
	case *prometheus.CounterVec:
		metric.With(t.labels).Inc()
	case prometheus.Counter:
		metric.Inc()
	case *prometheus.HistogramVec:
		metric.With(t.labels).Observe(val64)
	case prometheus.Histogram:
		metric.Observe(val64)
	}
}

// SetPrecision defines how we measure span duration (for `prometheus.Histogram` metrics). I.e., we might want
// to observe in /nano/mirco/etc.-seconds.
func (t *MetricSpanOption) SetPrecision(res time.Duration) {
	t.precision = res
}
