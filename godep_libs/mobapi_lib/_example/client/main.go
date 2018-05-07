package main

import (
	"fmt"
	"net/http"
	"time"

	"context"
	gotrace "motify_core_api/godep_libs/go-trace"
	"motify_core_api/godep_libs/mobapi_lib/example/client/adapter"
	"motify_core_api/godep_libs/mobapi_lib/logger"
)

// TODO: go generate dumps it to stdout not to file //go:generate go run ../main.go -gen-client-lib > adapter/client.go

func PrintSearch(query string) {
	service := adapter.NewExample(nil, StaticBalancer{"http://localhost:8080/"}, DiscardMonitoring{})

	headers := make(http.Header, 1)
	headers.Set(gotrace.TraceIDHeader, "1")

	ctx := gotrace.SetTraceHeadersToContext(context.Background(), headers, "client", "1", "localhost")
	result, err := service.SearchGoogleV1(ctx, adapter.SearchGoogleV1Args{
		Query: query,
	})
	if err != nil {
		if apiErr, ok := err.(*adapter.ServiceError); ok {
			logger.Debug(nil, "service error: %d %s", apiErr.Code, apiErr.Message)
			return
		}
		panic(err)
	}
	for _, link := range result {
		fmt.Println(link.Title, link.URL)
	}
}

func main() {
	PrintSearch("cookie monster")
	PrintSearch("")
}

type StaticBalancer struct {
	URL string
}

func (b StaticBalancer) Next() (string, error) {
	return b.URL, nil
}

type DiscardMonitoring struct{}

func (m DiscardMonitoring) GetSource() string {
	return ""
}

func (m DiscardMonitoring) ForExternalService(name string) interfaces.IMonitoring {
	return m
}

func (m DiscardMonitoring) GetMetricsRegistry() metrics.Registry {
	return nil
}

func (DiscardMonitoring) MonitorPath(string) {}

func (DiscardMonitoring) IncreaseCounter(ctx context.Context, name int, requestPath string) {}

func (DiscardMonitoring) UpdateMetrics(ctx context.Context, requestPath string, startTime time.Time) {}

func (DiscardMonitoring) RegisterAndUpdateHitMiss(ctx context.Context, key string, appearance int64, miss bool) {
}
