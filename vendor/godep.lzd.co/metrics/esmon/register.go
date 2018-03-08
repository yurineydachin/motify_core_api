// ElasticSearch (https://www.elastic.co) metrics for Prometheus.
package esmon

import (
	"godep.lzd.co/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// The registry. Naming rules described in wiki https://confluence.lazada.com/x/OihVAQ
var (
	ResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metrics.NS,
			Name:      "es_query_duration_milliseconds",
			Help:      "Histogram of RT for the request to ElasticSearch (ms).",
			Buckets:   []float64{5, 10, 15, 20, 30, 50, 70, 100, 150, 200, 300, 400, 500, 750, 1000, 1400, 2000, 3000, 5000},
		},
		[]string{"code", "endpoint", "is_error"},
	)
	FacetNumber = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: metrics.NS,
			Name:      "es_search_facets_total",
			Help:      "Number of facets used in ES queries.",
			Buckets:   []float64{1, 5, 10, 20, 30, 50, 75, 100, 125, 150, 175, 200, 225, 250, 275, 300, 325, 350, 375, 400, 425, 450, 475, 500, 550, 600, 650, 700, 800, 900, 1000},
		},
	)
)

// Don't forget register here the metric that was declared in var clause above!
func init() {
	prometheus.MustRegister(ResponseTime)
	prometheus.MustRegister(FacetNumber)
}
