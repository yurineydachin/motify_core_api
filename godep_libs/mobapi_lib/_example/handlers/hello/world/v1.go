package world

import (
	"context"
	"net/http"

	ctxMngr "motify_core_api/godep_libs/mobapi_lib/context"
	"motify_core_api/godep_libs/mobapi_lib/logger"

	"github.com/sergei-svistunov/gorpc/transport/cache"
)

// v1Args contains a request arguments
type v1Args struct{}

// Implementation
func (*Handler) V1(ctx context.Context, opts *v1Args) (string, error) {
	logger.Debug(ctx, "Hello, world")
	cache.EnableTransportCache(ctx)

	cookie1 := http.Cookie{
		Name:     "test1",
		Value:    "true",
		HttpOnly: true,
		// set more optional parameters if you'd like
	}
	cookie2 := http.Cookie{
		Name:  "test2",
		Value: "hello",
	}

	if err := ctxMngr.AddResponseCookies(ctx, &cookie1, &cookie2); err != nil {
		// handle error
	}

	return "Hello world", nil
}
