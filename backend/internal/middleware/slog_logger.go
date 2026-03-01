// Package middleware provides HTTP middleware for the RV Logbook API server.
package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// NewSlogLogger returns a middleware that logs each request as a structured
// JSON line via the provided slog.Logger. It captures method, path, HTTP
// status, duration, and the request ID set by chi's RequestID middleware.
//
// Wire it after chimiddleware.RequestID so the request ID is available.
func NewSlogLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// WrapResponseWriter intercepts WriteHeader so we can read the
			// status code after the downstream handler has run.
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			log.InfoContext(r.Context(), "request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", chimiddleware.GetReqID(r.Context()),
			)
		})
	}
}
