// The syntetic test with samples of metric usage.
// For benchmarking etc.

package main

import (
	"net/http"

	"./package1"
	"./package2"
	"./package3"

	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	go package1.DoSomething()
	go package2.DoSomething()
	go package3.DoSomething()

	http.ListenAndServe("127.0.1.2:8080", prometheus.Handler())
	select {}
}
