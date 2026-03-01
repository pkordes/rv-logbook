package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/repo"
	"github.com/pkordes/rv-logbook/backend/internal/service"
)

// ---- mock TagRepo ----------------------------------------------------------

type mockTagRepo struct {
	upsert         func(ctx context.Context, name, slug string) (domain.Tag, error)
	list           func(ctx context.Context, prefix string) ([]domain.Tag, error)
	addToStop      func(ctx context.Context, stopID, tagID uuid.UUID) error
	removeFromStop func(ctx context.Context, stopID uuid.UUID, slug string) error
	listByStop     func(ctx context.Context, stopID uuid.UUID) ([]domain.Tag, error)
}

func (m *mockTagRepo) Upsert(ctx context.Context, name, slug string) (domain.Tag, error) {
	return m.upsert(ctx, name, slug)
}
func (m *mockTagRepo) List(ctx context.Context, prefix string) ([]domain.Tag, error) {
	return m.list(ctx, prefix)
}
func (m *mockTagRepo) AddToStop(ctx context.Context, stopID, tagID uuid.UUID) error {
	return m.addToStop(ctx, stopID, tagID)
}
func (m *mockTagRepo) RemoveFromStop(ctx context.Context, stopID uuid.UUID, slug string) error {
	return m.removeFromStop(ctx, stopID, slug)
}
func (m *mockTagRepo) ListByStop(ctx context.Context, stopID uuid.UUID) ([]domain.Tag, error) {
	return m.listByStop(ctx, stopID)
}

// compile-time check
var _ repo.TagRepo = (*mockTagRepo)(nil)

// ---- UpsertByName ----------------------------------------------------------

func TestTagService_UpsertByName_OK(t *testing.T) {
	var capturedSlug string
	svc := service.NewTagService(&mockTagRepo{
		upsert: func(_ context.Context, name, slug string) (domain.Tag, error) {
			capturedSlug = slug
			return domain.Tag{ID: uuid.New(), Name: name, Slug: slug}, nil
		},
	})

	got, err := svc.UpsertByName(context.Background(), "Rocky Mountains")

	require.NoError(t, err)
	assert.Equal(t, "rocky-mountains", capturedSlug)
	assert.Equal(t, "rocky-mountains", got.Slug)
}

func TestTagService_UpsertByName_NormalizesCase(t *testing.T) {
	var capturedSlug string
	svc := service.NewTagService(&mockTagRepo{
		upsert: func(_ context.Context, _, slug string) (domain.Tag, error) {
			capturedSlug = slug
			return domain.Tag{Slug: slug}, nil
		},
	})

	_, err := svc.UpsertByName(context.Background(), "WALMART")
	require.NoError(t, err)
	assert.Equal(t, "walmart", capturedSlug)
}

func TestTagService_UpsertByName_CollapsesPunctuation(t *testing.T) {
	var capturedSlug string
	svc := service.NewTagService(&mockTagRepo{
		upsert: func(_ context.Context, _, slug string) (domain.Tag, error) {
			capturedSlug = slug
			return domain.Tag{Slug: slug}, nil
		},
	})

	_, err := svc.UpsertByName(context.Background(), "Rocky  Mountains!")
	require.NoError(t, err)
	assert.Equal(t, "rocky-mountains", capturedSlug)
}

func TestTagService_UpsertByName_EmptyName(t *testing.T) {
	svc := service.NewTagService(&mockTagRepo{})

	_, err := svc.UpsertByName(context.Background(), "   ")

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestTagService_UpsertByName_EmptyAfterNormalization(t *testing.T) {
	svc := service.NewTagService(&mockTagRepo{})

	// Input that normalizes to empty (only special chars)
	_, err := svc.UpsertByName(context.Background(), "!!! ---")

	assert.ErrorIs(t, err, domain.ErrValidation)
}

// ---- List ------------------------------------------------------------------

func TestTagService_List_All(t *testing.T) {
	tags := []domain.Tag{
		{ID: uuid.New(), Name: "Mountains", Slug: "mountains"},
		{ID: uuid.New(), Name: "Desert", Slug: "desert"},
	}
	svc := service.NewTagService(&mockTagRepo{
		list: func(_ context.Context, prefix string) ([]domain.Tag, error) {
			assert.Equal(t, "", prefix)
			return tags, nil
		},
	})

	got, err := svc.List(context.Background(), "")

	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestTagService_List_PrefixNormalized(t *testing.T) {
	var capturedPrefix string
	svc := service.NewTagService(&mockTagRepo{
		list: func(_ context.Context, prefix string) ([]domain.Tag, error) {
			capturedPrefix = prefix
			return []domain.Tag{}, nil
		},
	})

	_, err := svc.List(context.Background(), "Mount")

	require.NoError(t, err)
	assert.Equal(t, "mount", capturedPrefix, "prefix should be lowercased")
}

func TestTagService_List_ReturnsEmptySlice(t *testing.T) {
	svc := service.NewTagService(&mockTagRepo{
		list: func(_ context.Context, _ string) ([]domain.Tag, error) {
			return nil, nil
		},
	})

	got, err := svc.List(context.Background(), "")

	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got)
}
