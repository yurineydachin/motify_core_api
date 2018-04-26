// Metrics for Prometheus monitoring (http://prometheus.io).
// We use it with Grafana (http://grafana.org).
// Most of dashboards live at http://grafana.lzd.co.
package metrics

import (
	"strings"
	"time"
)

// NS is a common namespace for the all metrics specific to Lazada services.
const NS = "lzd"

// Ms calculates the duration in milliseconds instead on nanoseconds that default for time.Duration.
// Because Prometheus prefers milliseconds.
func Ms(duration time.Duration) float64 {
	return float64(duration / time.Millisecond)
}

// SinceMs just wraps time.Since() with converting result to milliseconds.
// Because Prometheus prefers milliseconds.
func SinceMs(started time.Time) float64 {
	return float64(time.Since(started) / time.Millisecond)
}

// IsError is a trivial helper for passing appropriate numbers to metrics.
func IsError(err error) string {
	if err != nil {
		return "1"
	}
	return "0"
}

// BucketsHTTP returns set of buckets for the histograms used in HTTP metrics
// as it declared by Lazada monitoring standard.
// See wiki https://confluence.lazada.com/x/OihVAQ
func BucketsHTTP() []float64 {
	return []float64{1, 3, 5, 10, 15, 20, 30, 50, 70, 100, 150, 200, 300, 400, 500, 750, 1000, 1400, 2000, 3000, 5000}
}

// MakePathTag makes the tag for a metric accordingly with recommendations of Prometheus and for compatibility with Grafana.
// It replaces slashes with underscores and replaces empty path (usual for GET /) with "root" word.
func MakePathTag(path string) string {
	if path == "/" {
		return "root"
	}
	return strings.ToLower(strings.Replace(strings.Trim(path, "/"), "/", "_", -1))
}

// Response status possible label values.
const (
	StatusOk          = "ok"
	StatusClientError = "client_error"
	StatusError       = "error"
)

// Status convert HTTP response code to the status string accordingly with Lazada metrics standard.
// Use it instead of returning HTTP codes!
// See wiki https://confluence.lazada.com/x/OihVAQ
func Status(httpCode int) string {
	switch {
	case httpCode >= 500:
		return StatusError
	case httpCode >= 400:
		return StatusClientError
	case httpCode >= 200:
		return StatusOk
	}
	return StatusOk
}
