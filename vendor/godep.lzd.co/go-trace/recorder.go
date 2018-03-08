package gotrace

import (
	"time"
)

type SpanRecorder interface {
	RecordSpan(span RawSpan)
}

type Recorder struct {
	opts *recorderOptions
}

func NewRecorder(options ...RecorderOption) *Recorder {
	opts := &recorderOptions{}
	for _, o := range options {
		o(opts)
	}
	return &Recorder{opts: opts}
}

func (r *Recorder) RecordSpan(span RawSpan) {
	data := map[string]interface{}{"duration": span.Duration.String()}
	// Process span tags: update metrics if necessary and copy tags data.
	for key, val := range span.Tags {
		if t, ok := val.(*MetricSpanOption); ok {
			t.Update(span.Duration)
		} else {
			data[key] = val
		}
	}
	// If logCollector is not defined, just return
	if r.opts.logCollector == nil {
		return
	}
	_, accessLog := span.Tags[IsAccessLogSpanTag]
	if accessLog {
		delete(data, IsAccessLogSpanTag)
	}

	var requestData string
	if v, ok := data[RequestDataSpanTag]; ok {
		requestData = v.(string)
		delete(data, RequestDataSpanTag)
	}
	if len(span.Logs) > 0 {
		logs := []map[string]interface{}{}
		for _, rec := range span.Logs {
			for _, f := range rec.Fields {
				logs = append(logs, map[string]interface{}{
					"key": f.Key(),
					"value": f.Value(),
					"ts": rec.Timestamp.UTC().Format(time.RFC3339Nano),
				})
			}
		}
		data[`logs`] = logs
	}
	level := ""
	if v, ok := data[LogLevelSpanTag]; ok {
		level = v.(string)
		delete(data, LogLevelSpanTag)
	}
	r.opts.logCollector.Collect(
		level,
		span.Context.TraceID,
		span.Context.ParentSpanID,
		span.Context.SpanID,
		span.Operation,
		requestData,
		data,
		accessLog,
	)
}
