package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkordes/rv-logbook/backend/internal/middleware"
)

// bodyReadingHandler is a test http.Handler that reads the full request body.
// It returns 413 if the body read fails (as MaxBytesReader causes), otherwise 200.
// This simulates what a real JSON-decoding handler does on each request.
var bodyReadingHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	_, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}
	w.WriteHeader(http.StatusOK)
})

// TestMaxBodySizeHandler_SmallBody_PassesThrough verifies that a request whose body
// is within the limit is forwarded to the next handler unchanged.
func TestMaxBodySizeHandler_SmallBody_PassesThrough(t *testing.T) {
	const limit = 100
	h := middleware.NewMaxBodySizeHandler(limit)(bodyReadingHandler)

	body := strings.NewReader(strings.Repeat("x", 50))
	req := httptest.NewRequest(http.MethodPost, "/trips", body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

// TestMaxBodySizeHandler_ContentLengthExceedsLimit_Returns413 verifies that a
// request advertising a Content-Length larger than the limit is rejected before
// the handler runs — no body bytes are read.
func TestMaxBodySizeHandler_ContentLengthExceedsLimit_Returns413(t *testing.T) {
	const limit = 100
	h := middleware.NewMaxBodySizeHandler(limit)(bodyReadingHandler)

	// We set the Content-Length header manually so the middleware can reject early.
	body := strings.NewReader(strings.Repeat("x", 200))
	req := httptest.NewRequest(http.MethodPost, "/trips", body)
	req.ContentLength = 200
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
}

// TestMaxBodySizeHandler_StreamingBodyExceedsLimit_Returns413 verifies that when no
// Content-Length header is set, the http.MaxBytesReader wrapping causes the body
// read inside the handler to fail once the limit is exceeded.
func TestMaxBodySizeHandler_StreamingBodyExceedsLimit_Returns413(t *testing.T) {
	const limit = 100
	h := middleware.NewMaxBodySizeHandler(limit)(bodyReadingHandler)

	body := strings.NewReader(strings.Repeat("x", 200))
	req := httptest.NewRequest(http.MethodPost, "/trips", body)
	req.ContentLength = -1 // unknown — no Content-Length header
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
}
