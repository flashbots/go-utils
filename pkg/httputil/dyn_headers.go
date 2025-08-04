package httputil

import (
	"context"
	"net/http"
)

type dynamicHeadersCtxKey struct{}

func CtxWithHeaders(ctx context.Context, headers http.Header) context.Context {
	ctx = context.WithValue(ctx, dynamicHeadersCtxKey{}, headers)
	return ctx
}

func DynamicHeadersFromCtx(ctx context.Context) http.Header {
	val, ok := ctx.Value(dynamicHeadersCtxKey{}).(http.Header)
	if !ok {
		return nil
	}

	return val
}
