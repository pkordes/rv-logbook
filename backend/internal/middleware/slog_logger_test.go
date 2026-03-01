package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/require"

	"github.com/pkordes/rv-logbook/backend/internal/middleware"
)

// TestSlogLogger_logsRequestFields verifies that the SlogLogger middleware
// writes a structured JSON log line containing method, path, status, duration,
// and the request ID placed in context by chi's RequestID middleware.
func TestSlogLogger_logsRequestFields(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Wrap the SlogLogger around a trivial 200 handler.
	h := middleware.NewSlogLogger(logger)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	// Simulate what chimiddleware.RequestID does: inject a known ID into context.
	// This keeps the test focused on our middleware's logging behaviour only.
	ctx := context.WithValue(req.Context(), chimiddleware.RequestIDKey, "test-req-id")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	// Parse the single JSON log line written by the middleware.
	var logEntry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	require.Equal(t, "GET", logEntry["method"])
	require.Equal(t, "/healthz", logEntry["path"])
	require.EqualValues(t, http.StatusOK, logEntry["status"])
	require.Equal(t, "test-req-id", logEntry["request_id"])
	require.NotNil(t, logEntry["duration_ms"])
}
