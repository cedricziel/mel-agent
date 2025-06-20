package client

import (
	"context"
	"fmt"
	"net/http"
)

// WithBearerToken returns a RequestEditorFn that adds a Bearer token to the Authorization header
func WithBearerToken(token string) RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		return nil
	}
}

// WithHeader returns a RequestEditorFn that adds a custom header
func WithHeader(key, value string) RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		req.Header.Set(key, value)
		return nil
	}
}
