// External services calls metrics for Prometheus.
package extmon

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
			Name:      "external_service_response_time_milliseconds",
			Help:      "Histogram of RT for unmarshal the request to the external service (ms).",
			Buckets:   buckets,
		},
		[]string{"status", "external_service", "method"},
	)
	UnmarshalResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metrics.NS,
			Name:      "external_service_response_unmarshal_time_milliseconds",
			Help:      "Histogram of RT for unmarshal the request to the external service (ms).",
			Buckets:   buckets,
		},
		[]string{"status", "external_service", "method"},
	)
)

// Don't forget register here the metric that was declared in var clause above!
func init() {
	prometheus.MustRegister(ResponseTime)
	prometheus.MustRegister(UnmarshalResponseTime)
}
