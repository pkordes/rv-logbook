package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pkordes/rv-logbook/backend/internal/handler"
	"github.com/pkordes/rv-logbook/backend/internal/handler/gen"
)

// TestGetHealth_returns200WithOKStatus verifies that GET /healthz returns
// HTTP 200 and a JSON body of {"status":"ok"}.
func TestGetHealth_returns200WithOKStatus(t *testing.T) {
	// Arrange: wire the strict handler through the generated chi router.
	// gen.NewStrictHandler wraps our StrictServerInterface implementation and
	// adapts it to the lower-level ServerInterface that HandlerFromMux expects.
	h := handler.NewHealthHandler()
	httpHandler := gen.Handler(gen.NewStrictHandler(h, nil))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	// Act
	httpHandler.ServeHTTP(rec, req)

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)

	var body gen.HealthResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Equal(t, "ok", body.Status)
}
