// Bytecache metrics for Prometheus.
// These metrics should be used for caches that realize caching of
// serialized data ([]byte etc).
//
// These metrics not included in Lazada standard yet.
package bytecachemon

import (
	"motify_core_api/godep_libs/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// The registry. Naming rules described in wiki https://confluence.lazada.com/x/OihVAQ
var (
	ItemNumber = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.NS,
			Name:      "byte_cache_items_total",
			Help:      "Total items in byte cache.",
		},
		[]string{"set"},
	)
	HitCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.NS,
			Name:      "byte_cache_hit_total",
			Help:      "Counter of misses to byte cache.",
		},
		[]string{"set"},
	)
	MissCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.NS,
			Name:      "byte_cache_miss_total",
			Help:      "Counter of misses to byte cache.",
		},
		[]string{"set"},
	)
)

// Don't forget register here the metric that was declared in var clause above!
func init() {
	prometheus.MustRegister(ItemNumber)
	prometheus.MustRegister(HitCount)
	prometheus.MustRegister(MissCount)
}
