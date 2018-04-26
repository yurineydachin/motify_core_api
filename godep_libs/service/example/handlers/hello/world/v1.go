package world

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/service/logger"
)

// v1Args contains a request arguments
type v1Args struct{}

// Implementation
func (*Handler) V1(ctx context.Context, opts *v1Args) (string, error) {
	logger.Debug(ctx, "Hello, world")
	cache.EnableTransportCache(ctx)
	return "Hello world", nil
}
