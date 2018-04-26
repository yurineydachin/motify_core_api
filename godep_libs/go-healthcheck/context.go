package healthcheck

import "context"

type ctxKey int8

const contextKey ctxKey = 1

// NewContext creates new context based on provided one and adds healthcheck instance into it
func NewContext(ctx context.Context, hc *HealthCheck) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, contextKey, hc)
}

// FromContext returns healthcheck instance from context (can be nil)
func FromContext(ctx context.Context) (*HealthCheck, bool) {
	if hc, ok := ctx.Value(contextKey).(*HealthCheck); ok {
		return hc, ok
	}

	return nil, false
}
