// Package handler implements the HTTP handlers for the RV Logbook API.
// Each handler implements gen.StrictServerInterface, which is generated from openapi.yaml.
package handler

import (
	"context"

	"github.com/pkordes/rv-logbook/backend/internal/handler/gen"
)

// HealthHandler implements gen.StrictServerInterface for the /healthz endpoint.
type HealthHandler struct{}

// NewHealthHandler returns a new HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// GetHealth handles GET /healthz.
// It returns HTTP 200 with {"status":"ok"} when the server is running.
func (h *HealthHandler) GetHealth(ctx context.Context, _ gen.GetHealthRequestObject) (gen.GetHealthResponseObject, error) {
	return gen.GetHealth200JSONResponse{Status: "ok"}, nil
}
