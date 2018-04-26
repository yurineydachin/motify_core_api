package package3

import (
	"time"

	"godep.lzd.co/metrics"
	"godep.lzd.co/metrics/bytecachemon"
	"godep.lzd.co/metrics/httpmon"
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
