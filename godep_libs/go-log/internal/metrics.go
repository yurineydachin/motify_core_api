package internal

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_model/go"
	"motify_core_api/godep_libs/metrics"
)

type LogMetrics struct {
	bufferMsgCount  prometheus.Gauge
	bufferSize      prometheus.Gauge
	msgSize         prometheus.Histogram
	lostMsgs        *prometheus.CounterVec
	socketWriteTime prometheus.Histogram
}

var GlobalLoggerMetrics LogMetrics

func init() {
	GlobalLoggerMetrics = LogMetrics{
		bufferMsgCount:  metrics.DefaultRegistry().NewGauge("logger_buffer_messages_total", "count of messages in buffer"),
		bufferSize:      metrics.DefaultRegistry().NewGauge("logger_buffer_size_bytes", "total size in bytes of messages in buffer"),
		msgSize:         metrics.DefaultRegistry().NewHistogram("logger_message_size_bytes", "message size", []float64{64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536, 131072}),
		lostMsgs:        metrics.DefaultRegistry().NewCounterVec("logger_lost_messages_total", "lost messages counter groupped by error type ex. go_log_lost_messages{error=}", "error"),
		socketWriteTime: metrics.DefaultRegistry().NewSummary("logger_write_time_seconds", "write time to socket"),
	}
}

func (m *LogMetrics) BufferMessagesCount() prometheus.Gauge {
	if m == nil {
		return nilMetric{}
	}
	return m.bufferMsgCount
}
func (m *LogMetrics) BufferSize() prometheus.Gauge {
	if m == nil {
		return nilMetric{}
	}
	return m.bufferSize
}
func (m *LogMetrics) MsgSize() prometheus.Histogram {
	if m == nil {
		return nilMetric{}
	}
	return m.msgSize

}
func (m *LogMetrics) LostMsgs(err error) prometheus.Counter {
	if m == nil {
		return nilMetric{}
	}

	errStr := err.Error()
	errKey := "unknown"
	if strings.Contains(errStr, "unix") {
		errKey = "connection"
	} else if strings.Contains(errStr, "Memory") {
		errKey = "memory"
	} else if strings.Contains(errStr, "capacity") {
		errKey = "capacity"
	}

	return m.lostMsgs.With(map[string]string{"error": errKey})

}
func (m *LogMetrics) SocketWriteTime() prometheus.Histogram {
	if m == nil {
		return nilMetric{}
	}
	return m.socketWriteTime
}

type nilMetric struct{}

func (nilMetric) Desc() *prometheus.Desc                   { return nil }
func (nilMetric) Write(*io_prometheus_client.Metric) error { return nil }
func (nilMetric) Describe(chan<- *prometheus.Desc)         {}
func (nilMetric) Collect(chan<- prometheus.Metric)         {}
func (nilMetric) Set(float64)                              {}
func (nilMetric) Inc()                                     {}
func (nilMetric) Dec()                                     {}
func (nilMetric) Add(float64)                              {}
func (nilMetric) Sub(float64)                              {}
func (nilMetric) SetToCurrentTime()                        {}
func (nilMetric) Observe(float64)                          {}
