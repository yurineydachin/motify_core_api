package response

import "context"

type Response struct {
	StatusCode int
}

type key int

var responseKey key = 0

func NewContext(ctx context.Context, r *Response) context.Context {
	return context.WithValue(ctx, responseKey, r)
}

func FromContext(ctx context.Context) (*Response, bool) {
	u, ok := ctx.Value(responseKey).(*Response)
	return u, ok
}
