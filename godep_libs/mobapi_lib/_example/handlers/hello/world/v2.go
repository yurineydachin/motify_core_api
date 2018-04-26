package world

import (
	"context"
	"net/http"

	ctxMngr "motify_core_api/godep_libs/mobapi_lib/context"
	"motify_core_api/godep_libs/mobapi_lib/logger"

	"github.com/sergei-svistunov/gorpc/transport/cache"
)

// v2Args contains a request arguments
type v2Args struct{}

type v2Response string

// V2 is a version 2 implementation of the handler
func (*Handler) V2(ctx context.Context, opts *v2Args) (v2Response, error) {
	logger.Debug(ctx, "Hello, world")
	cache.EnableTransportCache(ctx)

	cookie1 := http.Cookie{
		Name:     "test1",
		Value:    "true",
		HttpOnly: true,
		// set more optional parameters if you'd like
	}
	if err := ctxMngr.AddResponseCookies(ctx, &cookie1); err != nil {
		// handle error
	}

	cookie2 := http.Cookie{
		Name:  "test2",
		Value: "hello",
	}

	if err := ctxMngr.AddResponseCookies(ctx, &cookie2); err != nil {
		// handle error
	}

	return v2Response("Hello world"), nil
}
