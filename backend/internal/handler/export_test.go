package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/handler"
	"github.com/pkordes/rv-logbook/backend/internal/handler/gen"
)

// ---- mock ExportServicer ---------------------------------------------------

type mockExportServicer struct {
	export func(ctx context.Context) ([]domain.ExportRow, error)
}

func (m *mockExportServicer) Export(ctx context.Context) ([]domain.ExportRow, error) {
	return m.export(ctx)
}

// compile-time check: mockExportServicer must satisfy handler.ExportServicer.
var _ handler.ExportServicer = (*mockExportServicer)(nil)

// ---- helpers ---------------------------------------------------------------

// newExportHTTPHandler wires a Server with only the export service mock.
func newExportHTTPHandler(exportSvc handler.ExportServicer) http.Handler {
	srv := handler.NewServer(nil, nil, nil, exportSvc)
	return gen.Handler(gen.NewStrictHandler(srv, nil))
}

// exportRowFixture returns a fully-populated domain.ExportRow for testing.
func exportRowFixture() domain.ExportRow {
	arrivedAt := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	departedAt := time.Date(2024, 6, 18, 10, 0, 0, 0, time.UTC)

	return domain.ExportRow{
		TripID:        uuid.New().String(),
		TripName:      "Pacific Coast Tour",
		TripStartDate: "2024-06-15",
		TripEndDate:   "2024-06-30",
		StopName:      "Big Sur Campground",
		StopLocation:  "Big Sur, CA",
		ArrivedAt:     &arrivedAt,
		DepartedAt:    &departedAt,
		StopNotes:     "Great weather",
		Tags:          []string{"camping", "ocean"},
	}
}

// ---- GET /export — JSON ----------------------------------------------------

func TestGetExport_DefaultJSON_EmptyResult(t *testing.T) {
	svc := &mockExportServicer{
		export: func(_ context.Context) ([]domain.ExportRow, error) {
			return []domain.ExportRow{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/export", nil)
	rec := httptest.NewRecorder()
	newExportHTTPHandler(svc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

	var rows []gen.ExportRow
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&rows))
	assert.Empty(t, rows)
}

func TestGetExport_FormatJSON_ExplicitParam(t *testing.T) {
	row := exportRowFixture()
	svc := &mockExportServicer{
		export: func(_ context.Context) ([]domain.ExportRow, error) {
			return []domain.ExportRow{row}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/export?format=json", nil)
	rec := httptest.NewRecorder()
	newExportHTTPHandler(svc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

	var rows []gen.ExportRow
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&rows))
	require.Len(t, rows, 1)
	assert.Equal(t, row.TripName, rows[0].TripName)
}

func TestGetExport_JSON_TripWithNoStops_EmptyStopFields(t *testing.T) {
	row := domain.ExportRow{
		TripID:        uuid.New().String(),
		TripName:      "No Stop Trip",
		TripStartDate: "2024-07-01",
		TripEndDate:   "",
		Tags:          []string{},
	}
	svc := &mockExportServicer{
		export: func(_ context.Context) ([]domain.ExportRow, error) {
			return []domain.ExportRow{row}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/export", nil)
	rec := httptest.NewRecorder()
	newExportHTTPHandler(svc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var rows []gen.ExportRow
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&rows))
	require.Len(t, rows, 1)
	assert.Equal(t, row.TripName, rows[0].TripName)
	assert.Nil(t, rows[0].StopName)
	assert.Nil(t, rows[0].TripEndDate)
	assert.Nil(t, rows[0].ArrivedAt)
}

// ---- GET /export — CSV -----------------------------------------------------

func TestGetExport_CSV_FormatParam_ContentType(t *testing.T) {
	svc := &mockExportServicer{
		export: func(_ context.Context) ([]domain.ExportRow, error) {
			return []domain.ExportRow{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/export?format=csv", nil)
	rec := httptest.NewRecorder()
	newExportHTTPHandler(svc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/csv")
}

func TestGetExport_CSV_EmptyResult_HasHeaderRow(t *testing.T) {
	svc := &mockExportServicer{
		export: func(_ context.Context) ([]domain.ExportRow, error) {
			return []domain.ExportRow{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/export?format=csv", nil)
	rec := httptest.NewRecorder()
	newExportHTTPHandler(svc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.True(t, strings.HasPrefix(body, "trip_id,"), "CSV should start with header row, got: %q", body)
}

func TestGetExport_CSV_OneRow_HasHeaderAndDataRow(t *testing.T) {
	row := exportRowFixture()
	svc := &mockExportServicer{
		export: func(_ context.Context) ([]domain.ExportRow, error) {
			return []domain.ExportRow{row}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/export?format=csv", nil)
	rec := httptest.NewRecorder()
	newExportHTTPHandler(svc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	lines := strings.Split(strings.TrimSpace(rec.Body.String()), "\n")
	// Header + 1 data row.
	require.Len(t, lines, 2)
	assert.Contains(t, lines[0], "trip_id")
	assert.Contains(t, lines[1], row.TripName)
}

func TestGetExport_CSV_TagsJoinedWithPipe(t *testing.T) {
	row := exportRowFixture()
	row.Tags = []string{"beach", "hiking"}
	svc := &mockExportServicer{
		export: func(_ context.Context) ([]domain.ExportRow, error) {
			return []domain.ExportRow{row}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/export?format=csv", nil)
	rec := httptest.NewRecorder()
	newExportHTTPHandler(svc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "beach|hiking")
}

// ---- error handling --------------------------------------------------------

func TestGetExport_ServiceError_Returns500(t *testing.T) {
	svc := &mockExportServicer{
		export: func(_ context.Context) ([]domain.ExportRow, error) {
			return nil, fmt.Errorf("database unavailable")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/export", nil)
	rec := httptest.NewRecorder()
	newExportHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
