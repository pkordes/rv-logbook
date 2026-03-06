package middleware

import "net/http"

// NewSecurityHeadersHandler returns a middleware that sets defensive HTTP
// response headers on every response.
//
// Headers applied:
//   - X-Content-Type-Options: nosniff — prevents browsers from MIME-sniffing a
//     response away from the declared content-type, blocking some XSS vectors.
//   - X-Frame-Options: DENY — prevents all framing of this site, blocking
//     classic clickjacking attacks.
//   - Referrer-Policy: no-referrer — ensures the browser sends no Referer header
//     when navigating away, preventing URL leakage to third-party origins.
//
// These three headers are the minimum baseline recommended by OWASP's
// Secure Headers Project for any HTTP API.
func NewSecurityHeadersHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "no-referrer")
			next.ServeHTTP(w, r)
		})
	}
}
