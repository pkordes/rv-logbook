package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/handler"
	"github.com/pkordes/rv-logbook/backend/internal/handler/gen"
)

// mockTripServicer is a test double for handler.TripServicer.
// Set only the method fields your test needs.
type mockTripServicer struct {
	create  func(ctx context.Context, trip domain.Trip) (domain.Trip, error)
	getByID func(ctx context.Context, id uuid.UUID) (domain.Trip, error)
	list    func(ctx context.Context) ([]domain.Trip, error)
	update  func(ctx context.Context, trip domain.Trip) (domain.Trip, error)
	delete  func(ctx context.Context, id uuid.UUID) error
}

func (m *mockTripServicer) Create(ctx context.Context, t domain.Trip) (domain.Trip, error) {
	return m.create(ctx, t)
}
func (m *mockTripServicer) GetByID(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	return m.getByID(ctx, id)
}
func (m *mockTripServicer) List(ctx context.Context) ([]domain.Trip, error) {
	return m.list(ctx)
}
func (m *mockTripServicer) Update(ctx context.Context, t domain.Trip) (domain.Trip, error) {
	return m.update(ctx, t)
}
func (m *mockTripServicer) Delete(ctx context.Context, id uuid.UUID) error {
	return m.delete(ctx, id)
}

// compile-time check: mockTripServicer must satisfy handler.TripServicer.
var _ handler.TripServicer = (*mockTripServicer)(nil)

// ---- helpers ---------------------------------------------------------------

// newHTTPHandler wires a Server with the given mock into the generated chi router.
// This mirrors exactly how main.go wires it in production.
func newHTTPHandler(svc handler.TripServicer) http.Handler {
	srv := handler.NewServer(svc, nil, nil, nil)
	return gen.Handler(gen.NewStrictHandler(srv, nil))
}

func tripFixture() domain.Trip {
	start := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	return domain.Trip{
		ID:        uuid.New(),
		Name:      "Summer Tour",
		StartDate: start,
		EndDate:   &end,
		Notes:     "test notes",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

func dateStr(t time.Time) string {
	return t.Format("2006-01-02")
}

// ---- POST /trips -----------------------------------------------------------

func TestCreateTrip_201(t *testing.T) {
	fixture := tripFixture()
	svc := &mockTripServicer{
		create: func(_ context.Context, _ domain.Trip) (domain.Trip, error) {
			return fixture, nil
		},
	}

	body := jsonBody(t, map[string]any{
		"name":       "Summer Tour",
		"start_date": dateStr(fixture.StartDate),
		"end_date":   dateStr(*fixture.EndDate),
	})

	req := httptest.NewRequest(http.MethodPost, "/trips", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	newHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp gen.Trip
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, fixture.Name, resp.Name)
	assert.Equal(t, fixture.ID, resp.Id)
}

func TestCreateTrip_422_ValidationError(t *testing.T) {
	svc := &mockTripServicer{
		create: func(_ context.Context, _ domain.Trip) (domain.Trip, error) {
			return domain.Trip{}, fmt.Errorf("%w: name is required", domain.ErrValidation)
		},
	}

	body := jsonBody(t, map[string]any{
		"name":       "",
		"start_date": "2025-06-01",
	})

	req := httptest.NewRequest(http.MethodPost, "/trips", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	newHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var resp gen.ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.NotEmpty(t, resp.Error)
}

// ---- GET /trips ------------------------------------------------------------

func TestListTrips_200(t *testing.T) {
	trips := []domain.Trip{tripFixture(), tripFixture()}
	svc := &mockTripServicer{
		list: func(_ context.Context) ([]domain.Trip, error) { return trips, nil },
	}

	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	rec := httptest.NewRecorder()

	newHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp []gen.Trip
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Len(t, resp, 2)
}

func TestListTrips_200_Empty(t *testing.T) {
	svc := &mockTripServicer{
		list: func(_ context.Context) ([]domain.Trip, error) { return []domain.Trip{}, nil },
	}

	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	rec := httptest.NewRecorder()

	newHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Must be a JSON array, not null.
	assert.Contains(t, rec.Body.String(), "[")
}

// ---- GET /trips/{id} -------------------------------------------------------

func TestGetTrip_200(t *testing.T) {
	fixture := tripFixture()
	svc := &mockTripServicer{
		getByID: func(_ context.Context, id uuid.UUID) (domain.Trip, error) {
			return fixture, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/trips/"+fixture.ID.String(), nil)
	rec := httptest.NewRecorder()

	newHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp gen.Trip
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, fixture.ID, resp.Id)
}

func TestGetTrip_404(t *testing.T) {
	svc := &mockTripServicer{
		getByID: func(_ context.Context, _ uuid.UUID) (domain.Trip, error) {
			return domain.Trip{}, domain.ErrNotFound
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/trips/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()

	newHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ---- PUT /trips/{id} -------------------------------------------------------

func TestUpdateTrip_200(t *testing.T) {
	fixture := tripFixture()
	fixture.Name = "Updated Name"
	svc := &mockTripServicer{
		update: func(_ context.Context, t domain.Trip) (domain.Trip, error) {
			return fixture, nil
		},
	}

	body := jsonBody(t, map[string]any{
		"name":       "Updated Name",
		"start_date": dateStr(fixture.StartDate),
	})

	req := httptest.NewRequest(http.MethodPut, "/trips/"+fixture.ID.String(), body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	newHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp gen.Trip
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "Updated Name", resp.Name)
}

func TestUpdateTrip_404(t *testing.T) {
	svc := &mockTripServicer{
		update: func(_ context.Context, _ domain.Trip) (domain.Trip, error) {
			return domain.Trip{}, domain.ErrNotFound
		},
	}

	body := jsonBody(t, map[string]any{
		"name":       "X",
		"start_date": "2025-06-01",
	})

	req := httptest.NewRequest(http.MethodPut, "/trips/"+uuid.New().String(), body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	newHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ---- DELETE /trips/{id} ----------------------------------------------------

func TestDeleteTrip_204(t *testing.T) {
	svc := &mockTripServicer{
		delete: func(_ context.Context, _ uuid.UUID) error { return nil },
	}

	req := httptest.NewRequest(http.MethodDelete, "/trips/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()

	newHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteTrip_404(t *testing.T) {
	svc := &mockTripServicer{
		delete: func(_ context.Context, _ uuid.UUID) error { return domain.ErrNotFound },
	}

	req := httptest.NewRequest(http.MethodDelete, "/trips/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()

	newHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
