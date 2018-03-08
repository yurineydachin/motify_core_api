// HTTP metrics for Prometheus.
package httpmon

import (
	"godep.lzd.co/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// The registry. Naming rules described in wiki https://confluence.lazada.com/x/OihVAQ
// Don't set in the label "client_name" full version of UserAgent,
// set only short version with limited count of unique values.
var (
	buckets      = []float64{1, 3, 5, 10, 15, 20, 30, 50, 70, 100, 150, 200, 300, 400, 500, 750, 1000, 1400, 2000, 3000, 5000}
	ResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metrics.NS,
			Name:      "http_response_time_milliseconds",
			Help:      "Histogram of RT for HTTP requests (ms).",
			Buckets:   buckets,
		},
		[]string{"code", "handler", "client_name"},
	)
	ResponseTimeSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: metrics.NS,
			Name:      "http_summary_response_time_milliseconds",
			Help:      "Summary of RT for HTTP requests (ms).",
		},
		[]string{"code", "handler", "client_name"},
	)
	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.NS,
			Name:      "http_requests_total",
			Help:      "Counter of HTTP requests to the service.",
		},
		[]string{"handler", "client_name"},
	)
)

// Don't forget register here the metric that was declared in var clause above!
func init() {
	prometheus.MustRegister(ResponseTime)
	prometheus.MustRegister(ResponseTimeSummary)
	prometheus.MustRegister(RequestCount)
}
