package world

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/service/logger"
)

// v2Args contains a request arguments
type v2Args struct{}

type v2Response string

// V2 is a version 2 implementation of the handler
func (*Handler) V2(ctx context.Context, opts *v2Args) (v2Response, error) {
	logger.Debug(ctx, "Hello, world")
	cache.EnableTransportCache(ctx)
	return v2Response("Hello world"), nil
}
