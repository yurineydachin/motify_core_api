package context

import (
	"context"
	"net/http"
)

func AddResponseCookies(ctx context.Context, cookies ...*http.Cookie) error {
	cont, err := FromContext(ctx)
	if err == nil {
		cont.RespCookies = append(cont.RespCookies, cookies...)
	}
	return err
}

func GetResponseCookies(ctx context.Context) ([]*http.Cookie, bool) {
	if cont, err := FromContext(ctx); err == nil {
		return cont.RespCookies, true
	}
	return nil, false
}
