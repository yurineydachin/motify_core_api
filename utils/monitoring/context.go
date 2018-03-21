package monitoring

import (
	"context"
	"godep.lzd.co/metrics"
	"godep.lzd.co/metrics/extmon"
	"regexp"
	"time"
)

type dataKeyType int

var dataKey dataKeyType

type MonitorData struct {
	Service   string
	Path      string
	TimeStart time.Time
}

func SetMonitorData(ctx context.Context, service, path string) context.Context {
	return ToContext(ctx, MonitorData{Service: service, Path: path})
}

func ToContext(ctx context.Context, data MonitorData) context.Context {
	if ctx == nil {
		return nil
	}
	return context.WithValue(ctx, dataKey, &data)
}

func FromContext(ctx context.Context) *MonitorData {
	if ctx == nil {
		return &MonitorData{}
	}
	result, ok := ctx.Value(dataKey).(*MonitorData)
	if !ok {
		return &MonitorData{}
	}
	return result
}

func MonitorTimeResponse(ctx context.Context, code int) {
	data := FromContext(ctx)
	handler := clearPath(data.Path)
	extmon.ResponseTime.WithLabelValues(metrics.Status(code), data.Service, handler).
		Observe(metrics.SinceMs(data.TimeStart))
}

func MonitorRTAndStatus(ctx context.Context, status string) {
	data := FromContext(ctx)
	handler := clearPath(data.Path)
	extmon.ResponseTime.WithLabelValues(status, data.Service, handler).
		Observe(metrics.SinceMs(data.TimeStart))
}

var clearPathRe = regexp.MustCompile("([^v])[0-9]+")

func clearPath(path string) string {
	result := clearPathRe.ReplaceAllString(path, "$1%d")
	return result
}
