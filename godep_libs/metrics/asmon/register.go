// Aerospike (http://aerospike.com) metrics for Prometheus.
package asmon

import (
	"motify_core_api/godep_libs/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// The registry. Naming rules described in wiki https://confluence.lazada.com/x/OihVAQ
var (
	buckets      = []float64{1, 2, 3, 5, 8, 11, 15, 20, 25, 30, 38, 50, 65, 87, 100}
	ResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metrics.NS,
			Name:      "aerospike_request_duration_milliseconds",
			Help:      "Histogram of RT for the request to Aerospike (ms).",
			Buckets:   buckets,
		},
		[]string{"host", "namespace", "set", "operation", "is_error"},
	)
	HitCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.NS,
			Name:      "aerospike_hit_total",
			Help:      "Counter of hits to Aerospike cache.",
		},
		[]string{"host", "namespace", "set"},
	)
	MissCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.NS,
			Name:      "aerospike_miss_total",
			Help:      "Counter of misses to Aerospike cache.",
		},
		[]string{"host", "namespace", "set"},
	)
	ConnectionNumber = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.NS,
			Name:      "aerospike_connections_total",
			Help:      "Total number of connections opened to Aerospike.",
		},
		[]string{"host"},
	)
)

// Don't forget register here the metric that was declared in var clause above!
func init() {
	prometheus.MustRegister(ResponseTime)
	prometheus.MustRegister(HitCount)
	prometheus.MustRegister(MissCount)
	prometheus.MustRegister(ConnectionNumber)
}
