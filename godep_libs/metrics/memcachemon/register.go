// Memcache (https://memcached.org/) metrics for Prometheus.
package memcachemon

import (
	"godep.lzd.co/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// The registry. Naming rules described in wiki https://confluence.lazada.com/x/OihVAQ
var (
	buckets      = []float64{1, 2, 5, 8, 11, 15, 20, 25, 30, 38, 50, 75, 100, 150, 200, 300, 500, 700, 1100, 2000}
	ResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metrics.NS,
			Name:      "memcache_request_duration_milliseconds",
			Help:      "Histogram of RT for the request to Memcache (ms).",
			Buckets:   buckets,
		},
		[]string{"host", "operation", "is_error"},
	)
	HitCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.NS,
			Name:      "memcache_hit_total",
			Help:      "Counter of hits to Memcache.",
		},
		[]string{"host"},
	)
	MissCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.NS,
			Name:      "memcache_miss_total",
			Help:      "Counter of misses to Memcache.",
		},
		[]string{"host"},
	)
)

// Don't forget register here the metric that was declared in var clause above!
func init() {
	prometheus.MustRegister(ResponseTime)
	prometheus.MustRegister(HitCount)
	prometheus.MustRegister(MissCount)
}
