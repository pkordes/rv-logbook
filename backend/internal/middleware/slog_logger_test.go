package middleware_test

import (
	"bytes"
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
// writes a structured log line containing method, path, status, and request ID.
func TestSlogLogger_logsRequestFields(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Build handler chain: RequestID → SlogLogger → simple 200 handler.
	// chimiddleware.RequestID attaches a unique ID to the context and sets
	// the X-Request-Id response header.
	h := chimiddleware.RequestID(
		middleware.NewSlogLogger(logger)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.NotEmpty(t, rec.Header().Get("X-Request-Id"))

	// Parse the single JSON log line written by the middleware.
	var logEntry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	require.Equal(t, "GET", logEntry["method"])
	require.Equal(t, "/healthz", logEntry["path"])
	require.EqualValues(t, http.StatusOK, logEntry["status"])
	require.NotEmpty(t, logEntry["request_id"])
	require.NotEmpty(t, logEntry["duration_ms"])
}
