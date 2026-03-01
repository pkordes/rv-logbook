package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/repo"
	"github.com/pkordes/rv-logbook/backend/internal/service"
)

// ---- mock repos ------------------------------------------------------------

// mockStopRepo is a hand-written test double for repo.StopRepo.
type mockStopRepo struct {
	create              func(ctx context.Context, stop domain.Stop) (domain.Stop, error)
	getByID             func(ctx context.Context, tripID, stopID uuid.UUID) (domain.Stop, error)
	listByTripID        func(ctx context.Context, tripID uuid.UUID) ([]domain.Stop, error)
	listByTripIDPaged   func(ctx context.Context, tripID uuid.UUID, p domain.PaginationParams) ([]domain.Stop, int64, error)
	update              func(ctx context.Context, stop domain.Stop) (domain.Stop, error)
	delete              func(ctx context.Context, tripID, stopID uuid.UUID) error
}

func (m *mockStopRepo) Create(ctx context.Context, stop domain.Stop) (domain.Stop, error) {
	return m.create(ctx, stop)
}
func (m *mockStopRepo) GetByID(ctx context.Context, tripID, stopID uuid.UUID) (domain.Stop, error) {
	return m.getByID(ctx, tripID, stopID)
}
func (m *mockStopRepo) ListByTripID(ctx context.Context, tripID uuid.UUID) ([]domain.Stop, error) {
	return m.listByTripID(ctx, tripID)
}
func (m *mockStopRepo) ListByTripIDPaged(ctx context.Context, tripID uuid.UUID, p domain.PaginationParams) ([]domain.Stop, int64, error) {
	if m.listByTripIDPaged != nil {
		return m.listByTripIDPaged(ctx, tripID, p)
	}
	return nil, 0, nil
}
func (m *mockStopRepo) Update(ctx context.Context, stop domain.Stop) (domain.Stop, error) {
	return m.update(ctx, stop)
}
func (m *mockStopRepo) Delete(ctx context.Context, tripID, stopID uuid.UUID) error {
	return m.delete(ctx, tripID, stopID)
}

// compile-time check: mockStopRepo must satisfy repo.StopRepo.
var _ repo.StopRepo = (*mockStopRepo)(nil)

// ---- helpers ---------------------------------------------------------------

func validStop(tripID uuid.UUID) domain.Stop {
	return domain.Stop{
		TripID:    tripID,
		Name:      "Camp Grounds A",
		Location:  "Yellowstone, WY",
		ArrivedAt: time.Date(2025, 6, 2, 10, 0, 0, 0, time.UTC),
	}
}

// newStopService constructs a StopService wired to the given mocks.
// Pass nil for tagRepo when the test does not exercise tag operations.
func newStopService(tripRepo repo.TripRepo, stopRepo repo.StopRepo) *service.StopService {
	return service.NewStopService(tripRepo, stopRepo, nil)
}

// ---- Create ----------------------------------------------------------------

func TestStopService_Create_OK(t *testing.T) {
	tripID := uuid.New()
	input := validStop(tripID)
	stored := input
	stored.ID = uuid.New()

	svc := newStopService(
		&mockTripRepo{
			getByID: func(_ context.Context, id uuid.UUID) (domain.Trip, error) {
				return domain.Trip{ID: id}, nil
			},
		},
		&mockStopRepo{
			create: func(_ context.Context, s domain.Stop) (domain.Stop, error) {
				return stored, nil
			},
		},
	)

	got, err := svc.Create(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, stored.ID, got.ID)
}

func TestStopService_Create_TripNotFound(t *testing.T) {
	svc := newStopService(
		&mockTripRepo{
			getByID: func(_ context.Context, _ uuid.UUID) (domain.Trip, error) {
				return domain.Trip{}, domain.ErrNotFound
			},
		},
		&mockStopRepo{},
	)

	_, err := svc.Create(context.Background(), validStop(uuid.New()))

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestStopService_Create_NameRequired(t *testing.T) {
	tripID := uuid.New()
	svc := newStopService(
		&mockTripRepo{
			getByID: func(_ context.Context, id uuid.UUID) (domain.Trip, error) {
				return domain.Trip{ID: id}, nil
			},
		},
		&mockStopRepo{},
	)

	input := validStop(tripID)
	input.Name = "   "

	_, err := svc.Create(context.Background(), input)

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestStopService_Create_DepartedBeforeArrived(t *testing.T) {
	tripID := uuid.New()
	svc := newStopService(
		&mockTripRepo{
			getByID: func(_ context.Context, id uuid.UUID) (domain.Trip, error) {
				return domain.Trip{ID: id}, nil
			},
		},
		&mockStopRepo{},
	)

	input := validStop(tripID)
	departed := input.ArrivedAt.Add(-1 * time.Hour) // depart before arriving — invalid
	input.DepartedAt = &departed

	_, err := svc.Create(context.Background(), input)

	assert.ErrorIs(t, err, domain.ErrValidation)
}

// ---- GetByID ---------------------------------------------------------------

func TestStopService_GetByID_OK(t *testing.T) {
	tripID, stopID := uuid.New(), uuid.New()
	expected := domain.Stop{ID: stopID, TripID: tripID, Name: "Stop A"}

	svc := newStopService(
		&mockTripRepo{},
		&mockStopRepo{
			getByID: func(_ context.Context, tID, sID uuid.UUID) (domain.Stop, error) {
				return expected, nil
			},
		},
	)

	got, err := svc.GetByID(context.Background(), tripID, stopID)

	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestStopService_GetByID_NotFound(t *testing.T) {
	svc := newStopService(
		&mockTripRepo{},
		&mockStopRepo{
			getByID: func(_ context.Context, _, _ uuid.UUID) (domain.Stop, error) {
				return domain.Stop{}, domain.ErrNotFound
			},
		},
	)

	_, err := svc.GetByID(context.Background(), uuid.New(), uuid.New())

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// ---- ListByTripID ----------------------------------------------------------

func TestStopService_ListByTripID_OK(t *testing.T) {
	tripID := uuid.New()
	stops := []domain.Stop{
		{ID: uuid.New(), TripID: tripID},
		{ID: uuid.New(), TripID: tripID},
	}

	svc := newStopService(
		&mockTripRepo{},
		&mockStopRepo{
			listByTripID: func(_ context.Context, _ uuid.UUID) ([]domain.Stop, error) {
				return stops, nil
			},
		},
	)

	got, err := svc.ListByTripID(context.Background(), tripID)

	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestStopService_ListByTripID_ReturnsEmptySlice(t *testing.T) {
	svc := newStopService(
		&mockTripRepo{},
		&mockStopRepo{
			listByTripID: func(_ context.Context, _ uuid.UUID) ([]domain.Stop, error) {
				return nil, nil
			},
		},
	)

	got, err := svc.ListByTripID(context.Background(), uuid.New())

	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got)
}

// ---- Update ----------------------------------------------------------------

func TestStopService_Update_OK(t *testing.T) {
	tripID := uuid.New()
	input := validStop(tripID)
	input.ID = uuid.New()
	input.Name = "Updated Name"

	svc := newStopService(
		&mockTripRepo{},
		&mockStopRepo{
			update: func(_ context.Context, s domain.Stop) (domain.Stop, error) {
				return s, nil
			},
		},
	)

	got, err := svc.Update(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, "Updated Name", got.Name)
}

func TestStopService_Update_ValidationFails(t *testing.T) {
	input := validStop(uuid.New())
	input.ID = uuid.New()
	input.Name = ""

	svc := newStopService(&mockTripRepo{}, &mockStopRepo{})

	_, err := svc.Update(context.Background(), input)

	assert.ErrorIs(t, err, domain.ErrValidation)
}

// ---- Delete ----------------------------------------------------------------

func TestStopService_Delete_OK(t *testing.T) {
	svc := newStopService(
		&mockTripRepo{},
		&mockStopRepo{
			delete: func(_ context.Context, _, _ uuid.UUID) error {
				return nil
			},
		},
	)

	err := svc.Delete(context.Background(), uuid.New(), uuid.New())

	require.NoError(t, err)
}

func TestStopService_Delete_NotFound(t *testing.T) {
	svc := newStopService(
		&mockTripRepo{},
		&mockStopRepo{
			delete: func(_ context.Context, _, _ uuid.UUID) error {
				return domain.ErrNotFound
			},
		},
	)

	err := svc.Delete(context.Background(), uuid.New(), uuid.New())

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// ---- error propagation helper check ----------------------------------------

func TestStopService_Create_RepoError(t *testing.T) {
	repoErr := errors.New("db exploded")
	tripID := uuid.New()

	svc := newStopService(
		&mockTripRepo{
			getByID: func(_ context.Context, id uuid.UUID) (domain.Trip, error) {
				return domain.Trip{ID: id}, nil
			},
		},
		&mockStopRepo{
			create: func(_ context.Context, _ domain.Stop) (domain.Stop, error) {
				return domain.Stop{}, repoErr
			},
		},
	)

	_, err := svc.Create(context.Background(), validStop(tripID))

	assert.ErrorIs(t, err, repoErr)
}

// ---- AddTag ----------------------------------------------------------------

func TestStopService_AddTag_OK(t *testing.T) {
	stopID := uuid.New()
	tagID := uuid.New()

	svc := service.NewStopService(
		&mockTripRepo{},
		&mockStopRepo{},
		&mockTagRepo{
			upsert: func(_ context.Context, name, slug string) (domain.Tag, error) {
				return domain.Tag{ID: tagID, Name: name, Slug: slug}, nil
			},
			addToStop: func(_ context.Context, sID, tID uuid.UUID) error {
				assert.Equal(t, stopID, sID)
				assert.Equal(t, tagID, tID)
				return nil
			},
		},
	)

	got, err := svc.AddTag(context.Background(), stopID, "Rocky Mountains")

	require.NoError(t, err)
	assert.Equal(t, "rocky-mountains", got.Slug)
	assert.Equal(t, tagID, got.ID)
}

func TestStopService_AddTag_NormalizesName(t *testing.T) {
	var capturedSlug string
	svc := service.NewStopService(
		&mockTripRepo{},
		&mockStopRepo{},
		&mockTagRepo{
			upsert: func(_ context.Context, _, slug string) (domain.Tag, error) {
				capturedSlug = slug
				return domain.Tag{ID: uuid.New(), Slug: slug}, nil
			},
			addToStop: func(_ context.Context, _, _ uuid.UUID) error { return nil },
		},
	)

	_, err := svc.AddTag(context.Background(), uuid.New(), "WALMART")

	require.NoError(t, err)
	assert.Equal(t, "walmart", capturedSlug)
}

func TestStopService_AddTag_EmptyName(t *testing.T) {
	svc := service.NewStopService(&mockTripRepo{}, &mockStopRepo{}, &mockTagRepo{})

	_, err := svc.AddTag(context.Background(), uuid.New(), "   ")

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestStopService_AddTag_UpsertError(t *testing.T) {
	repoErr := errors.New("upsert failed")
	svc := service.NewStopService(
		&mockTripRepo{},
		&mockStopRepo{},
		&mockTagRepo{
			upsert: func(_ context.Context, _, _ string) (domain.Tag, error) {
				return domain.Tag{}, repoErr
			},
		},
	)

	_, err := svc.AddTag(context.Background(), uuid.New(), "camping")

	assert.ErrorIs(t, err, repoErr)
}

func TestStopService_AddTag_AddToStopError(t *testing.T) {
	repoErr := errors.New("link failed")
	svc := service.NewStopService(
		&mockTripRepo{},
		&mockStopRepo{},
		&mockTagRepo{
			upsert: func(_ context.Context, name, slug string) (domain.Tag, error) {
				return domain.Tag{ID: uuid.New(), Slug: slug}, nil
			},
			addToStop: func(_ context.Context, _, _ uuid.UUID) error { return repoErr },
		},
	)

	_, err := svc.AddTag(context.Background(), uuid.New(), "camping")

	assert.ErrorIs(t, err, repoErr)
}

// ---- RemoveTagFromStop -----------------------------------------------------

func TestStopService_RemoveTagFromStop_OK(t *testing.T) {
	svc := service.NewStopService(
		&mockTripRepo{},
		&mockStopRepo{},
		&mockTagRepo{
			removeFromStop: func(_ context.Context, _ uuid.UUID, slug string) error {
				assert.Equal(t, "camping", slug)
				return nil
			},
		},
	)

	err := svc.RemoveTagFromStop(context.Background(), uuid.New(), "camping")

	require.NoError(t, err)
}

func TestStopService_RemoveTagFromStop_NotFound(t *testing.T) {
	svc := service.NewStopService(
		&mockTripRepo{},
		&mockStopRepo{},
		&mockTagRepo{
			removeFromStop: func(_ context.Context, _ uuid.UUID, _ string) error {
				return domain.ErrNotFound
			},
		},
	)

	err := svc.RemoveTagFromStop(context.Background(), uuid.New(), "camping")

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// ---- ListTagsByStop --------------------------------------------------------

func TestStopService_ListTagsByStop_OK(t *testing.T) {
	stopID := uuid.New()
	expected := []domain.Tag{
		{ID: uuid.New(), Slug: "camping"},
		{ID: uuid.New(), Slug: "national-park"},
	}

	svc := service.NewStopService(
		&mockTripRepo{},
		&mockStopRepo{},
		&mockTagRepo{
			listByStop: func(_ context.Context, sID uuid.UUID) ([]domain.Tag, error) {
				assert.Equal(t, stopID, sID)
				return expected, nil
			},
		},
	)

	got, err := svc.ListTagsByStop(context.Background(), stopID)

	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestStopService_ListTagsByStop_ReturnsEmptySlice(t *testing.T) {
	svc := service.NewStopService(
		&mockTripRepo{},
		&mockStopRepo{},
		&mockTagRepo{
			listByStop: func(_ context.Context, _ uuid.UUID) ([]domain.Tag, error) {
				return nil, nil
			},
		},
	)

	got, err := svc.ListTagsByStop(context.Background(), uuid.New())

	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got)
}
