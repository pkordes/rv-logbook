package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkordes/rv-logbook/backend/internal/middleware"
)

// trivialHandler is a minimal http.Handler that always returns 200.
var trivialHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

// TestCORSHandler_GET_AllowedOrigin verifies that a GET from an allowed origin
// receives the Access-Control-Allow-Origin header in the response.
func TestCORSHandler_GET_AllowedOrigin(t *testing.T) {
	h := middleware.NewCORSHandler([]string{"http://localhost:5173"})(trivialHandler)

	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "http://localhost:5173", rec.Header().Get("Access-Control-Allow-Origin"))
}

// TestCORSHandler_OPTIONS_Preflight verifies that an OPTIONS preflight request
// returns a 2xx status and the necessary CORS headers.
// Browsers send this before any cross-origin request with custom headers or
// non-simple methods (e.g. PUT, DELETE, or Content-Type: application/json).
func TestCORSHandler_OPTIONS_Preflight(t *testing.T) {
	h := middleware.NewCORSHandler([]string{"http://localhost:5173"})(trivialHandler)

	req := httptest.NewRequest(http.MethodOptions, "/trips", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	// The Fetch specification requires browsers to send Access-Control-Request-Headers
	// values in lowercase. rs/cors normalises its allowed-headers list to lowercase and
	// compares verbatim, so the test must match that convention.
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	// rs/cors returns 204 for OPTIONS preflights.
	assert.True(t, rec.Code == http.StatusNoContent || rec.Code == http.StatusOK,
		"expected 2xx for OPTIONS preflight, got %d", rec.Code)
	assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Methods"))
}

// TestCORSHandler_GET_DisallowedOrigin verifies that a request from a
// disallowed origin does NOT receive the Access-Control-Allow-Origin header.
// The browser will then block the response — the response itself can still be 200,
// but the CORS header must be absent.
func TestCORSHandler_GET_DisallowedOrigin(t *testing.T) {
	h := middleware.NewCORSHandler([]string{"http://localhost:5173"})(trivialHandler)

	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	req.Header.Set("Origin", "http://evil.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}
