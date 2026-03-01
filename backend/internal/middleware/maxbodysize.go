package middleware

import "net/http"

// NewMaxBodySizeHandler returns a middleware that limits incoming request body
// sizes to limit bytes. Requests exceeding the limit are rejected with 413
// Request Entity Too Large before reaching the next handler.
//
// Stub: passes all requests through without size enforcement.
// Tests will fail against this until the real implementation is added.
func NewMaxBodySizeHandler(limit int64) func(http.Handler) http.Handler {
	_ = limit
	return func(next http.Handler) http.Handler {
		return next
	}
}
