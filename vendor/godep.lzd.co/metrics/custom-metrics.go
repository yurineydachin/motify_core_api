// Metrics for Prometheus monitoring (http://prometheus.io).
// We use it with Grafana (http://grafana.org).
// Most of dashboards live at http://grafana.lzd.co.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Registry keeps custom registry if you don't want use default one.
// It implements prometheus.Registerer interface.
type Registry struct {
	registerer prometheus.Registerer
}

// Register a new metric in the registry.
func (r *Registry) Register(collector prometheus.Collector) error {
	return r.registerer.Register(collector)
}

// Register a new metric in the registry. If the metric fail to register it will panic.
func (r *Registry) MustRegister(collectors ...prometheus.Collector) {
	r.registerer.MustRegister(collectors...)
}

// Unregister metric from the registry.
func (r *Registry) Unregister(collector prometheus.Collector) bool {
	return r.registerer.Unregister(collector)
}

// DefaultRegistry declares that you want use default registry from the Prometheus.
// It is the most common way to use the client.
func DefaultRegistry() *Registry {
	return &Registry{prometheus.DefaultRegisterer}
}

// NewRegistry declares that you mant use a new custom registry but not default one.
// The custom registry will use registry type from Prometheus.
func NewRegistry() *Registry {
	return &Registry{prometheus.NewRegistry()}
}

// NewCounter declares a new Counter in the default Prometheus registry.
func (r *Registry) NewCounter(name, desc string) prometheus.Counter {
	counter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: NS,
			Name:      name,
			Help:      desc,
		},
	)
	r.MustRegister(counter)
	return counter
}

// NewGauge declares a new Gauge in the default Prometheus registry.
func (r *Registry) NewGauge(name, desc string) prometheus.Gauge {
	gauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: NS,
			Name:      name,
			Help:      desc,
		},
	)
	r.MustRegister(gauge)
	return gauge
}

// NewHistogram declares a new Histogram in the default Prometheus registry.
func (r *Registry) NewHistogram(name, desc string, buckets []float64) prometheus.Histogram {
	histogram := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: NS,
			Name:      name,
			Help:      desc,
			Buckets:   buckets,
		},
	)
	r.MustRegister(histogram)
	return histogram
}

// NewSummary declares a new Summary in the default Prometheus registry.
func (r *Registry) NewSummary(name, desc string) prometheus.Summary {
	summary := prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: NS,
			Name:      name,
			Help:      desc,
		},
	)
	r.MustRegister(summary)
	return summary
}

// NewCounterVec declares a new CounterVec in the default Prometheus registry.
func (r *Registry) NewCounterVec(name, desc string, tags ...string) *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: NS,
			Name:      name,
			Help:      desc,
		},
		tags,
	)
	r.MustRegister(counter)
	return counter
}

// NewGaugeVec declares a new GaugeVec in the default Prometheus registry.
func (r *Registry) NewGaugeVec(name, desc string, tags ...string) *prometheus.GaugeVec {
	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: NS,
			Name:      name,
			Help:      desc,
		},
		tags,
	)
	r.MustRegister(gauge)
	return gauge
}

// NewHistogramVec declares a new HistogramVec in the default Prometheus registry.
func (r *Registry) NewHistogramVec(name, desc string, buckets []float64, tags ...string) *prometheus.HistogramVec {
	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: NS,
			Name:      name,
			Help:      desc,
			Buckets:   buckets,
		},
		tags,
	)
	r.MustRegister(histogram)
	return histogram
}

// NewSummaryVec declares a new SummaryVec in the default Prometheus registry.
func (r *Registry) NewSummaryVec(name, desc string, tags ...string) *prometheus.SummaryVec {
	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: NS,
			Name:      name,
			Help:      desc,
		},
		tags,
	)
	r.MustRegister(summary)
	return summary
}
