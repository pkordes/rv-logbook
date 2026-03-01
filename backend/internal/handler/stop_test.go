package handler_test

import (
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

// mockStopServicer is a test double for handler.StopServicer.
// Set only the method fields your test needs.
type mockStopServicer struct {
	create            func(ctx context.Context, stop domain.Stop) (domain.Stop, error)
	getByID           func(ctx context.Context, tripID, stopID uuid.UUID) (domain.Stop, error)
	listByTripID      func(ctx context.Context, tripID uuid.UUID) ([]domain.Stop, error)
	listByTripIDPaged func(ctx context.Context, tripID uuid.UUID, p domain.PaginationParams) ([]domain.Stop, int64, error)
	update            func(ctx context.Context, stop domain.Stop) (domain.Stop, error)
	delete            func(ctx context.Context, tripID, stopID uuid.UUID) error
	addTag            func(ctx context.Context, stopID uuid.UUID, tagName string) (domain.Tag, error)
	removeTagFrom     func(ctx context.Context, stopID uuid.UUID, slug string) error
	listTagsByStop    func(ctx context.Context, stopID uuid.UUID) ([]domain.Tag, error)
}

func (m *mockStopServicer) Create(ctx context.Context, s domain.Stop) (domain.Stop, error) {
	return m.create(ctx, s)
}
func (m *mockStopServicer) GetByID(ctx context.Context, tripID, stopID uuid.UUID) (domain.Stop, error) {
	return m.getByID(ctx, tripID, stopID)
}
func (m *mockStopServicer) ListByTripID(ctx context.Context, tripID uuid.UUID) ([]domain.Stop, error) {
	return m.listByTripID(ctx, tripID)
}
func (m *mockStopServicer) ListByTripIDPaged(ctx context.Context, tripID uuid.UUID, p domain.PaginationParams) ([]domain.Stop, int64, error) {
	return m.listByTripIDPaged(ctx, tripID, p)
}
func (m *mockStopServicer) Update(ctx context.Context, s domain.Stop) (domain.Stop, error) {
	return m.update(ctx, s)
}
func (m *mockStopServicer) Delete(ctx context.Context, tripID, stopID uuid.UUID) error {
	return m.delete(ctx, tripID, stopID)
}
func (m *mockStopServicer) AddTag(ctx context.Context, stopID uuid.UUID, tagName string) (domain.Tag, error) {
	return m.addTag(ctx, stopID, tagName)
}
func (m *mockStopServicer) RemoveTagFromStop(ctx context.Context, stopID uuid.UUID, slug string) error {
	return m.removeTagFrom(ctx, stopID, slug)
}
func (m *mockStopServicer) ListTagsByStop(ctx context.Context, stopID uuid.UUID) ([]domain.Tag, error) {
	return m.listTagsByStop(ctx, stopID)
}

// compile-time check: mockStopServicer must satisfy handler.StopServicer.
var _ handler.StopServicer = (*mockStopServicer)(nil)

// newStopHTTPHandler wires a Server with the given stop mock (no trip service needed).
func newStopHTTPHandler(svc handler.StopServicer) http.Handler {
	srv := handler.NewServer(nil, svc, nil, nil)
	return gen.Handler(gen.NewStrictHandler(srv, nil))
}

func stopFixture(tripID uuid.UUID) domain.Stop {
	return domain.Stop{
		ID:        uuid.New(),
		TripID:    tripID,
		Name:      "Yellowstone Camp",
		Location:  "Yellowstone, WY",
		ArrivedAt: time.Date(2025, 6, 2, 10, 0, 0, 0, time.UTC),
		Notes:     "Great spot",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

// ---- POST /trips/{tripId}/stops -------------------------------------------

func TestCreateStop_201(t *testing.T) {
	tripID := uuid.New()
	fixture := stopFixture(tripID)
	svc := &mockStopServicer{
		create: func(_ context.Context, _ domain.Stop) (domain.Stop, error) {
			return fixture, nil
		},
	}

	body := jsonBody(t, map[string]any{
		"name":       fixture.Name,
		"arrived_at": fixture.ArrivedAt.Format(time.RFC3339),
	})
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/trips/%s/stops", tripID), body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	newStopHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCreateStop_404_TripNotFound(t *testing.T) {
	tripID := uuid.New()
	svc := &mockStopServicer{
		create: func(_ context.Context, _ domain.Stop) (domain.Stop, error) {
			return domain.Stop{}, domain.ErrNotFound
		},
	}

	body := jsonBody(t, map[string]any{
		"name":       "Camp A",
		"arrived_at": time.Now().UTC().Format(time.RFC3339),
	})
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/trips/%s/stops", tripID), body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	newStopHTTPHandler(svc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)

	var errResp gen.ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp))
	assert.Equal(t, "not_found", errResp.Error.Code)
}

func TestCreateStop_422_Validation(t *testing.T) {
	tripID := uuid.New()
	svc := &mockStopServicer{
		create: func(_ context.Context, _ domain.Stop) (domain.Stop, error) {
			return domain.Stop{}, fmt.Errorf("%w: name is required", domain.ErrValidation)
		},
	}

	body := jsonBody(t, map[string]any{
		"name":       "   ",
		"arrived_at": time.Now().UTC().Format(time.RFC3339),
	})
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/trips/%s/stops", tripID), body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	newStopHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var errResp gen.ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp))
	assert.Equal(t, "validation_error", errResp.Error.Code)
}

// ---- GET /trips/{tripId}/stops --------------------------------------------

func TestListStops_200(t *testing.T) {
	tripID := uuid.New()
	stops := []domain.Stop{stopFixture(tripID), stopFixture(tripID)}
	svc := &mockStopServicer{
		listByTripIDPaged: func(_ context.Context, _ uuid.UUID, _ domain.PaginationParams) ([]domain.Stop, int64, error) {
			return stops, int64(len(stops)), nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/trips/%s/stops", tripID), nil)
	rec := httptest.NewRecorder()

	newStopHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp gen.StopList
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Len(t, resp.Data, 2)
	// Fails in Red because stub handler leaves Pagination.Total at 0.
	assert.Equal(t, 2, resp.Pagination.Total)
}

func TestListStops_200_Empty(t *testing.T) {
	tripID := uuid.New()
	svc := &mockStopServicer{
		listByTripIDPaged: func(_ context.Context, _ uuid.UUID, _ domain.PaginationParams) ([]domain.Stop, int64, error) {
			return []domain.Stop{}, 0, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/trips/%s/stops", tripID), nil)
	rec := httptest.NewRecorder()

	newStopHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp gen.StopList
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Len(t, resp.Data, 0)
	assert.Equal(t, 0, resp.Pagination.Total)
}

// ---- GET /trips/{tripId}/stops/{stopId} -----------------------------------

func TestGetStop_200(t *testing.T) {
	tripID := uuid.New()
	fixture := stopFixture(tripID)
	svc := &mockStopServicer{
		getByID: func(_ context.Context, _, _ uuid.UUID) (domain.Stop, error) {
			return fixture, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/trips/%s/stops/%s", tripID, fixture.ID), nil)
	rec := httptest.NewRecorder()

	newStopHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetStop_404(t *testing.T) {
	tripID := uuid.New()
	svc := &mockStopServicer{
		getByID: func(_ context.Context, _, _ uuid.UUID) (domain.Stop, error) {
			return domain.Stop{}, domain.ErrNotFound
		},
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/trips/%s/stops/%s", tripID, uuid.New()), nil)
	rec := httptest.NewRecorder()

	newStopHTTPHandler(svc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)

	var errResp gen.ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp))
	assert.Equal(t, "not_found", errResp.Error.Code)
}

// ---- PUT /trips/{tripId}/stops/{stopId} -----------------------------------

func TestUpdateStop_200(t *testing.T) {
	tripID := uuid.New()
	fixture := stopFixture(tripID)
	fixture.Name = "Updated Camp"
	svc := &mockStopServicer{
		update: func(_ context.Context, s domain.Stop) (domain.Stop, error) {
			return fixture, nil
		},
	}

	body := jsonBody(t, map[string]any{
		"name":       "Updated Camp",
		"arrived_at": fixture.ArrivedAt.Format(time.RFC3339),
	})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/trips/%s/stops/%s", tripID, fixture.ID), body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	newStopHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpdateStop_404(t *testing.T) {
	tripID := uuid.New()
	svc := &mockStopServicer{
		update: func(_ context.Context, _ domain.Stop) (domain.Stop, error) {
			return domain.Stop{}, domain.ErrNotFound
		},
	}

	body := jsonBody(t, map[string]any{
		"name":       "Updated",
		"arrived_at": time.Now().UTC().Format(time.RFC3339),
	})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/trips/%s/stops/%s", tripID, uuid.New()), body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	newStopHTTPHandler(svc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)

	var errResp gen.ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp))
	assert.Equal(t, "not_found", errResp.Error.Code)
}

// ---- DELETE /trips/{tripId}/stops/{stopId} --------------------------------

func TestDeleteStop_204(t *testing.T) {
	tripID := uuid.New()
	stopID := uuid.New()
	svc := &mockStopServicer{
		delete: func(_ context.Context, _, _ uuid.UUID) error {
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/trips/%s/stops/%s", tripID, stopID), nil)
	rec := httptest.NewRecorder()

	newStopHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteStop_404(t *testing.T) {
	tripID := uuid.New()
	stopID := uuid.New()
	svc := &mockStopServicer{
		delete: func(_ context.Context, _, _ uuid.UUID) error {
			return domain.ErrNotFound
		},
	}

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/trips/%s/stops/%s", tripID, stopID), nil)
	rec := httptest.NewRecorder()

	newStopHTTPHandler(svc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)

	var errResp gen.ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp))
	assert.Equal(t, "not_found", errResp.Error.Code)
}
