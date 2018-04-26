package gotrace

type RecorderOption func(opts *recorderOptions)

type logCollector interface {
	Collect(level, traceID, parentSpanID, spanID, operationName, requestData string,
		additionalData map[string]interface{}, isAccess bool)
}

type recorderOptions struct {
	logCollector logCollector
}

// WithLogCollector makes Tracer able to write Span to go-log
// it takes golog.NewSpanCollector(...)
func WithLogCollector(collector logCollector) RecorderOption {
	return func(opts *recorderOptions) {
		opts.logCollector = collector
	}
}
