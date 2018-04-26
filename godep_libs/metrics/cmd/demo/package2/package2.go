package package2

import (
	"time"

	"motify_core_api/godep_libs/metrics"
	"motify_core_api/godep_libs/metrics/asmon"
	"motify_core_api/godep_libs/metrics/httpmon"
)

func DoSomething() {
	for {
		time.Sleep(20 * time.Millisecond)
		asmon.ConnectionNumber.WithLabelValues("sample-host").Set(1234)
		asmon.HitCount.WithLabelValues("host", "ns", "sample-set").Inc()
		asmon.MissCount.WithLabelValues("host", "ns", "sample-set").Inc()
		httpmon.ResponseTime.With(map[string]string{
			"code":        "200",
			"handler":     "example",
			"client_name": "client",
		}).Observe(metrics.Ms(200 * time.Millisecond))
	}
}
