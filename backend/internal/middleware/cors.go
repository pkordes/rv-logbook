// Package middleware provides reusable HTTP middleware for the RV Logbook API.
package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

// NewCORSHandler returns a middleware that applies CORS headers based on allowedOrigins.
// Each entry in allowedOrigins must be a full origin (scheme + host, no trailing slash).
// Allowed methods and headers cover the full REST surface of the API.
func NewCORSHandler(allowedOrigins []string) func(http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})
	return func(next http.Handler) http.Handler {
		return c.Handler(next)
	}
}
