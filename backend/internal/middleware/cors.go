package middleware

import "net/http"

// NewCORSHandler returns a middleware that sets CORS headers allowing the
// given origins to make requests to this server.
// An empty origins slice disables CORS (development only).
func NewCORSHandler(allowedOrigins []string) func(http.Handler) http.Handler {
	// Stub: passes the request through without setting any CORS headers.
	// Tests will fail against this until the real implementation is added.
	return func(next http.Handler) http.Handler {
		return next
	}
}
