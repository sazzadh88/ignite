package routing

import (
	"context"
	"net/http"
)

type paramKey struct{ name string }

func contextWithParam(ctx context.Context, key, value string) context.Context {
	return context.WithValue(ctx, paramKey{key}, value)
}

// Param extracts a route parameter from the request context.
func Param(r *http.Request, key string) string {
	val, _ := r.Context().Value(paramKey{key}).(string)
	return val
}
