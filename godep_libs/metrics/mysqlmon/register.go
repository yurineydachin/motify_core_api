// MySQL (http://mysql.org) metrics for Prometheus.
package mysqlmon

import (
	"motify_core_api/godep_libs/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// The registry. Naming rules described in wiki https://confluence.lazada.com/x/OihVAQ
var (
	buckets      = []float64{1, 3, 5, 10, 15, 20, 30, 50, 70, 100, 150, 200, 300, 400, 500, 750, 1000, 1400, 2000, 5000}
	ResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metrics.NS,
			Name:      "mysql_query_duration_milliseconds",
			Help:      "Histogram of RT for MySQL queries (ms).",
			Buckets:   buckets,
		},
		[]string{"host", "db", "is_error", "query"},
	)
	ConnectionNumber = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.NS,
			Name:      "mysql_connections_total",
			Help:      "Counter of currently used connections to MySQL.",
		},
		[]string{"host", "db"},
	)
)

// Don't forget register here the metric that was declared in var clause above!
func init() {
	prometheus.MustRegister(ResponseTime)
	prometheus.MustRegister(ConnectionNumber)
}
