// Package handler — export.go implements GET /export.
// Returns all trips, stops, and tags as a flat table.
// Supports content negotiation via ?format=csv (CSV) or default (JSON).
package handler

import (
	"bytes"
	"context"
	"encoding/csv"
	"strings"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/handler/gen"
)

// csvHeaders defines the column names written as the first row of any CSV export.
var csvHeaders = []string{
	"trip_id", "trip_name", "trip_start_date", "trip_end_date",
	"stop_name", "stop_location", "arrived_at", "departed_at",
	"stop_notes", "tags",
}

// GetExport implements GET /export.
// It returns a flat table of every trip, stop, and tag combination.
// Use ?format=csv to receive CSV; default is JSON.
func (s *Server) GetExport(ctx context.Context, req gen.GetExportRequestObject) (gen.GetExportResponseObject, error) {
	rows, err := s.export.Export(ctx)
	if err != nil {
		return nil, err
	}

	wantCSV := req.Params.Format != nil && *req.Params.Format == gen.Csv
	if wantCSV {
		return buildCSVResponse(rows), nil
	}
	return buildJSONResponse(rows), nil
}

// buildJSONResponse converts domain rows to the typed JSON response.
func buildJSONResponse(rows []domain.ExportRow) gen.GetExport200JSONResponse {
	out := make(gen.GetExport200JSONResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, domainRowToGenRow(r))
	}
	return out
}

// buildCSVResponse encodes domain rows as CSV and wraps in the streaming response type.
// Tags within a row are pipe-separated ("|") to keep each stop on a single CSV line.
func buildCSVResponse(rows []domain.ExportRow) gen.GetExport200TextcsvResponse {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	//nolint:errcheck — bytes.Buffer.Write never returns an error.
	w.Write(csvHeaders)
	for _, r := range rows {
		//nolint:errcheck
		w.Write(domainRowToCSVRecord(r))
	}
	w.Flush()

	return gen.GetExport200TextcsvResponse{
		Body:          &buf,
		ContentLength: int64(buf.Len()),
	}
}

// domainRowToGenRow maps a domain.ExportRow to the generated gen.ExportRow type.
// Fields that are empty strings become nil pointers (omitempty in JSON).
func domainRowToGenRow(r domain.ExportRow) gen.ExportRow {
	tripID, _ := uuid.Parse(r.TripID)

	startDate := mustParseDate(r.TripStartDate)
	row := gen.ExportRow{
		TripId:        tripID,
		TripName:      r.TripName,
		TripStartDate: startDate,
		ArrivedAt:     r.ArrivedAt,
		DepartedAt:    r.DepartedAt,
		Tags:          r.Tags,
	}

	if r.TripEndDate != "" {
		d := mustParseDate(r.TripEndDate)
		row.TripEndDate = &d
	}
	if r.StopName != "" {
		row.StopName = &r.StopName
	}
	if r.StopLocation != "" {
		row.StopLocation = &r.StopLocation
	}
	if r.StopNotes != "" {
		row.StopNotes = &r.StopNotes
	}
	return row
}

// domainRowToCSVRecord encodes a domain.ExportRow as a flat string slice.
// Nil time pointers are encoded as empty strings.
// Tags are joined with "|".
func domainRowToCSVRecord(r domain.ExportRow) []string {
	arrivedAt := formatOptionalTime(r.ArrivedAt)
	departedAt := formatOptionalTime(r.DepartedAt)
	return []string{
		r.TripID,
		r.TripName,
		r.TripStartDate,
		r.TripEndDate,
		r.StopName,
		r.StopLocation,
		arrivedAt,
		departedAt,
		r.StopNotes,
		strings.Join(r.Tags, "|"),
	}
}

// mustParseDate parses an "2006-01-02" string into an openapi_types.Date.
// Panics on malformed input; callers are expected to pass service-generated dates.
func mustParseDate(s string) openapi_types.Date {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic("handler: malformed date from service: " + s)
	}
	return openapi_types.Date{Time: t}
}

// formatOptionalTime returns the RFC3339 representation of t, or "" if t is nil.
func formatOptionalTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
