package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pkordes/rv-logbook/backend/internal/middleware"
)

// noopHandler is a minimal handler that always returns 200. Used to verify
// that SecurityHeaders passes through to the next handler in the chain.
var noopHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

// TestSecurityHeaders_SetsExpectedHeaders verifies that every required defensive
// header is present on the response regardless of the route or method.
func TestSecurityHeaders_SetsExpectedHeaders(t *testing.T) {
	h := middleware.NewSecurityHeadersHandler()(noopHandler)

	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"),
		"X-Content-Type-Options: nosniff must be set to prevent MIME sniffing")
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"),
		"X-Frame-Options: DENY must be set to prevent clickjacking")
	assert.Equal(t, "no-referrer", rec.Header().Get("Referrer-Policy"),
		"Referrer-Policy: no-referrer must be set to prevent URL leakage")
}

// TestSecurityHeaders_PassesThroughStatus verifies the middleware does not
// alter the status code returned by the next handler.
func TestSecurityHeaders_PassesThroughStatus(t *testing.T) {
	// A handler that returns 201
	createdHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	h := middleware.NewSecurityHeadersHandler()(createdHandler)

	req := httptest.NewRequest(http.MethodPost, "/trips", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code,
		"security headers middleware must not modify the downstream status code")
}
