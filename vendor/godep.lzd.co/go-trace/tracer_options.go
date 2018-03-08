package gotrace

type TracerOption func(opts *tracerOptions)

type tracerOptions struct {
	appName    string
	appVersion string
	appNode    string
	recorders  []SpanRecorder
}

// WithAppEnv sets Application Environments
func WithAppEnv(appName, appVersion, appNode string) TracerOption {
	return func(opts *tracerOptions) {
		opts.appName = appName
		opts.appVersion = appVersion
		opts.appNode = appNode
	}
}

// WithSpanRecorder adds new SpanRecorder for Tracer
// You can use as many SpanRecorder as you want
func WithSpanRecorder(r SpanRecorder) TracerOption {
	return func(opts *tracerOptions) {
		opts.recorders = append(opts.recorders, r)
	}
}
