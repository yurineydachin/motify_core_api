package main

import (
	"context"
)

const (
	verboseCtxKey = "verbose"
)

func newContext(verbose bool) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, verboseCtxKey, verbose)
	return ctx
}

func getVerbose(ctx context.Context) bool {
	verbose := ctx.Value(verboseCtxKey)
	if verbose == nil {
		return false
	}
	return verbose.(bool)
}
