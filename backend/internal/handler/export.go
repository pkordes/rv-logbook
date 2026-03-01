// Package handler — export.go implements GET /export.
// Returns all trips, stops, and tags as a flat table.
// Supports content negotiation via ?format=csv (CSV) or default (JSON).
package handler

import (
	"context"
	"fmt"

	"github.com/pkordes/rv-logbook/backend/internal/handler/gen"
)

// GetExport implements GET /export.
// It returns a flat table of every trip, stop, and tag combination.
// Use ?format=csv to receive CSV; default is JSON.
func (s *Server) GetExport(_ context.Context, _ gen.GetExportRequestObject) (gen.GetExportResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}
