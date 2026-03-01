package handler_test

import (
	"context"
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

// ---- mock TagServicer -------------------------------------------------------

type mockTagServicer struct {
	list func(ctx context.Context, prefix string) ([]domain.Tag, error)
}

func (m *mockTagServicer) List(ctx context.Context, prefix string) ([]domain.Tag, error) {
	return m.list(ctx, prefix)
}

// compile-time check: mockTagServicer must satisfy handler.TagServicer.
var _ handler.TagServicer = (*mockTagServicer)(nil)

// ---- helpers ---------------------------------------------------------------

// newTagHTTPHandler wires a Server with tag and stop service mocks.
// Pass nil for mocks that the test does not use.
func newTagHTTPHandler(tagSvc handler.TagServicer, stopSvc handler.StopServicer) http.Handler {
	srv := handler.NewServer(nil, stopSvc, tagSvc, nil)
	return gen.Handler(gen.NewStrictHandler(srv, nil))
}

func tagFixture() domain.Tag {
	return domain.Tag{
		ID:        uuid.New(),
		Name:      "National Park",
		Slug:      "national-park",
		CreatedAt: time.Now().UTC(),
	}
}

// ---- GET /tags -------------------------------------------------------------

func TestListTags_200(t *testing.T) {
	tags := []domain.Tag{tagFixture(), tagFixture()}
	svc := &mockTagServicer{
		list: func(_ context.Context, prefix string) ([]domain.Tag, error) {
			assert.Equal(t, "", prefix)
			return tags, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/tags", nil)
	rec := httptest.NewRecorder()
	newTagHTTPHandler(svc, nil).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestListTags_200_WithPrefix(t *testing.T) {
	var capturedPrefix string
	svc := &mockTagServicer{
		list: func(_ context.Context, prefix string) ([]domain.Tag, error) {
			capturedPrefix = prefix
			return []domain.Tag{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/tags?q=cam", nil)
	rec := httptest.NewRecorder()
	newTagHTTPHandler(svc, nil).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "cam", capturedPrefix)
}

// ---- GET /trips/{tripId}/stops/{stopId}/tags --------------------------------

func TestListTagsByStop_200(t *testing.T) {
	tripID, stopID := uuid.New(), uuid.New()
	tags := []domain.Tag{tagFixture()}

	svc := &mockStopServicer{
		listTagsByStop: func(_ context.Context, sID uuid.UUID) ([]domain.Tag, error) {
			assert.Equal(t, stopID, sID)
			return tags, nil
		},
	}

	url := fmt.Sprintf("/trips/%s/stops/%s/tags", tripID, stopID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	newTagHTTPHandler(nil, svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// ---- POST /trips/{tripId}/stops/{stopId}/tags --------------------------------

func TestAddTagToStop_201(t *testing.T) {
	tripID, stopID := uuid.New(), uuid.New()
	tag := tagFixture()

	svc := &mockStopServicer{
		addTag: func(_ context.Context, sID uuid.UUID, name string) (domain.Tag, error) {
			assert.Equal(t, stopID, sID)
			assert.Equal(t, "National Park", name)
			return tag, nil
		},
	}

	body := jsonBody(t, map[string]any{"name": "National Park"})
	url := fmt.Sprintf("/trips/%s/stops/%s/tags", tripID, stopID)
	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	newTagHTTPHandler(nil, svc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
}

func TestAddTagToStop_422_ValidationError(t *testing.T) {
	tripID, stopID := uuid.New(), uuid.New()

	svc := &mockStopServicer{
		addTag: func(_ context.Context, _ uuid.UUID, _ string) (domain.Tag, error) {
			return domain.Tag{}, fmt.Errorf("%w: tag name is required", domain.ErrValidation)
		},
	}

	body := jsonBody(t, map[string]any{"name": "   "})
	url := fmt.Sprintf("/trips/%s/stops/%s/tags", tripID, stopID)
	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	newTagHTTPHandler(nil, svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAddTagToStop_404_StopNotFound(t *testing.T) {
	tripID, stopID := uuid.New(), uuid.New()

	svc := &mockStopServicer{
		addTag: func(_ context.Context, _ uuid.UUID, _ string) (domain.Tag, error) {
			return domain.Tag{}, domain.ErrNotFound
		},
	}

	body := jsonBody(t, map[string]any{"name": "camping"})
	url := fmt.Sprintf("/trips/%s/stops/%s/tags", tripID, stopID)
	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	newTagHTTPHandler(nil, svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ---- DELETE /trips/{tripId}/stops/{stopId}/tags/{slug} ----------------------

func TestRemoveTagFromStop_204(t *testing.T) {
	tripID, stopID := uuid.New(), uuid.New()

	svc := &mockStopServicer{
		removeTagFrom: func(_ context.Context, sID uuid.UUID, slug string) error {
			assert.Equal(t, stopID, sID)
			assert.Equal(t, "camping", slug)
			return nil
		},
	}

	url := fmt.Sprintf("/trips/%s/stops/%s/tags/camping", tripID, stopID)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	rec := httptest.NewRecorder()
	newTagHTTPHandler(nil, svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestRemoveTagFromStop_404_NotLinked(t *testing.T) {
	tripID, stopID := uuid.New(), uuid.New()

	svc := &mockStopServicer{
		removeTagFrom: func(_ context.Context, _ uuid.UUID, _ string) error {
			return domain.ErrNotFound
		},
	}

	url := fmt.Sprintf("/trips/%s/stops/%s/tags/camping", tripID, stopID)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	rec := httptest.NewRecorder()
	newTagHTTPHandler(nil, svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
