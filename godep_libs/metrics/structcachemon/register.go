// Struct cache metrics for Prometheus.
// These metrics should be used for caches that realize caching of structures
// contrary to caching of serialized byte arrays.
package structcachemon

import (
	"motify_core_api/godep_libs/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// The registry. Naming rules described in wiki https://confluence.lazada.com/x/OihVAQ
var (
	buckets      = []float64{1, 2, 3, 5, 10, 50, 100, 200, 500, 1000, 2000}
	ResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metrics.NS,
			Name:      "struct_cache_request_duration_milliseconds",
			Help:      "Histogram of RT for the request to struct cache (ms).",
			Buckets:   buckets,
		},
		[]string{"operation", "set"},
	)
	ItemNumber = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.NS,
			Name:      "struct_cache_items_total",
			Help:      "Total items in struct cache.",
		},
		[]string{"set"},
	)
	HitCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.NS,
			Name:      "struct_cache_hit_total",
			Help:      "Counter of misses to struct cache.",
		},
		[]string{"set"},
	)
	MissCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.NS,
			Name:      "struct_cache_miss_total",
			Help:      "Counter of misses to struct cache.",
		},
		[]string{"set"},
	)
)

// Don't forget register here the metric that was declared in var clause above!
func init() {
	prometheus.MustRegister(ResponseTime)
	prometheus.MustRegister(ItemNumber)
	prometheus.MustRegister(HitCount)
	prometheus.MustRegister(MissCount)
}
