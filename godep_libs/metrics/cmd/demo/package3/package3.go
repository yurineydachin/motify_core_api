package package3

import (
	"time"

	"motify_core_api/godep_libs/metrics"
	"motify_core_api/godep_libs/metrics/bytecachemon"
	"motify_core_api/godep_libs/metrics/httpmon"
)

func DoSomething() {
	for {
		time.Sleep(30 * time.Millisecond)
		bytecachemon.ItemNumber.WithLabelValues("sample-set").Inc()
		bytecachemon.HitCount.WithLabelValues("sample-set").Inc()
		bytecachemon.MissCount.WithLabelValues("sample-set").Inc()
		httpmon.ResponseTime.With(map[string]string{
			"code":        "300",
			"handler":     "example",
			"client_name": "client",
		}).Observe(metrics.Ms(300 * time.Millisecond))
	}
}
