package package1

import (
	"time"

	"motify_core_api/godep_libs/metrics"
	"motify_core_api/godep_libs/metrics/httpmon"
	"motify_core_api/godep_libs/metrics/structcachemon"
)

func DoSomething() {
	for {
		time.Sleep(10 * time.Millisecond)
		structcachemon.ItemNumber.WithLabelValues("sample-set").Inc()
		structcachemon.HitCount.WithLabelValues("sample-set").Inc()
		structcachemon.MissCount.WithLabelValues("sample-set").Inc()
		httpmon.ResponseTime.With(map[string]string{
			"code":        "100",
			"handler":     "example",
			"client_name": "client",
		}).Observe(metrics.Ms(100 * time.Millisecond))
	}
}
