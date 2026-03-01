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

// mockTripRepo is a hand-written test double for repo.TripRepo.
// Each method is a function field — set only the ones your test needs.
// This is idiomatic Go: no mock generation library required for simple cases.
// Think of it like passing a lambda per method in Java or a MagicMock in Python.
type mockTripRepo struct {
	create  func(ctx context.Context, trip domain.Trip) (domain.Trip, error)
	getByID func(ctx context.Context, id uuid.UUID) (domain.Trip, error)
	list    func(ctx context.Context) ([]domain.Trip, error)
	update  func(ctx context.Context, trip domain.Trip) (domain.Trip, error)
	delete  func(ctx context.Context, id uuid.UUID) error
}

func (m *mockTripRepo) Create(ctx context.Context, trip domain.Trip) (domain.Trip, error) {
	return m.create(ctx, trip)
}
func (m *mockTripRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	return m.getByID(ctx, id)
}
func (m *mockTripRepo) List(ctx context.Context) ([]domain.Trip, error) {
	return m.list(ctx)
}
func (m *mockTripRepo) Update(ctx context.Context, trip domain.Trip) (domain.Trip, error) {
	return m.update(ctx, trip)
}
func (m *mockTripRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.delete(ctx, id)
}

// compile-time check: mockTripRepo must satisfy repo.TripRepo.
var _ repo.TripRepo = (*mockTripRepo)(nil)

// ---- helpers ---------------------------------------------------------------

func validTrip() domain.Trip {
	start := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	return domain.Trip{
		Name:      "Summer Tour",
		StartDate: start,
		EndDate:   &end,
	}
}

func echoRepo() *mockTripRepo {
	// A repo that echoes whatever it receives back — useful for Create/Update tests
	// that only care about validation logic, not what the DB returns.
	return &mockTripRepo{
		create: func(_ context.Context, t domain.Trip) (domain.Trip, error) { return t, nil },
		update: func(_ context.Context, t domain.Trip) (domain.Trip, error) { return t, nil },
	}
}

// ---- Create tests ----------------------------------------------------------

func TestTripService_Create_Valid(t *testing.T) {
	svc := service.NewTripService(echoRepo())

	got, err := svc.Create(context.Background(), validTrip())

	require.NoError(t, err)
	assert.Equal(t, "Summer Tour", got.Name)
}

func TestTripService_Create_MissingName(t *testing.T) {
	svc := service.NewTripService(echoRepo())

	trip := validTrip()
	trip.Name = "   " // whitespace-only should be treated as empty

	_, err := svc.Create(context.Background(), trip)

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestTripService_Create_EndDateBeforeStartDate(t *testing.T) {
	svc := service.NewTripService(echoRepo())

	trip := validTrip()
	bad := trip.StartDate.AddDate(0, 0, -1) // one day before start
	trip.EndDate = &bad

	_, err := svc.Create(context.Background(), trip)

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestTripService_Create_EndDateEqualToStartDate(t *testing.T) {
	svc := service.NewTripService(echoRepo())

	trip := validTrip()
	same := trip.StartDate // same day — a one-day trip is valid
	trip.EndDate = &same

	_, err := svc.Create(context.Background(), trip)

	// Same-day trips should be allowed (arrived and departed the same day).
	assert.NoError(t, err)
}

func TestTripService_Create_NilEndDate(t *testing.T) {
	svc := service.NewTripService(echoRepo())

	trip := validTrip()
	trip.EndDate = nil // trip still in progress — valid

	_, err := svc.Create(context.Background(), trip)

	assert.NoError(t, err)
}

func TestTripService_Create_RepoError(t *testing.T) {
	repoErr := errors.New("db exploded")
	r := &mockTripRepo{
		create: func(_ context.Context, _ domain.Trip) (domain.Trip, error) {
			return domain.Trip{}, repoErr
		},
	}
	svc := service.NewTripService(r)

	_, err := svc.Create(context.Background(), validTrip())

	// The service should propagate repo errors unchanged.
	assert.ErrorIs(t, err, repoErr)
}

// ---- GetByID tests ---------------------------------------------------------

func TestTripService_GetByID_Found(t *testing.T) {
	want := validTrip()
	want.ID = uuid.New()

	r := &mockTripRepo{
		getByID: func(_ context.Context, id uuid.UUID) (domain.Trip, error) {
			return want, nil
		},
	}
	svc := service.NewTripService(r)

	got, err := svc.GetByID(context.Background(), want.ID)

	require.NoError(t, err)
	assert.Equal(t, want.ID, got.ID)
}

func TestTripService_GetByID_NotFound(t *testing.T) {
	r := &mockTripRepo{
		getByID: func(_ context.Context, _ uuid.UUID) (domain.Trip, error) {
			return domain.Trip{}, domain.ErrNotFound
		},
	}
	svc := service.NewTripService(r)

	_, err := svc.GetByID(context.Background(), uuid.New())

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// ---- List tests ------------------------------------------------------------

func TestTripService_List(t *testing.T) {
	trips := []domain.Trip{validTrip(), validTrip()}
	r := &mockTripRepo{
		list: func(_ context.Context) ([]domain.Trip, error) { return trips, nil },
	}
	svc := service.NewTripService(r)

	got, err := svc.List(context.Background())

	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestTripService_List_Empty(t *testing.T) {
	r := &mockTripRepo{
		list: func(_ context.Context) ([]domain.Trip, error) { return nil, nil },
	}
	svc := service.NewTripService(r)

	got, err := svc.List(context.Background())

	require.NoError(t, err)
	// Should return an empty slice, not nil — callers can safely range over it.
	assert.NotNil(t, got)
	assert.Empty(t, got)
}

// ---- Update tests ----------------------------------------------------------

func TestTripService_Update_Valid(t *testing.T) {
	svc := service.NewTripService(echoRepo())

	trip := validTrip()
	trip.ID = uuid.New()
	trip.Name = "Renamed Trip"

	got, err := svc.Update(context.Background(), trip)

	require.NoError(t, err)
	assert.Equal(t, "Renamed Trip", got.Name)
}

func TestTripService_Update_MissingName(t *testing.T) {
	svc := service.NewTripService(echoRepo())

	trip := validTrip()
	trip.Name = ""

	_, err := svc.Update(context.Background(), trip)

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestTripService_Update_EndDateBeforeStartDate(t *testing.T) {
	svc := service.NewTripService(echoRepo())

	trip := validTrip()
	bad := trip.StartDate.AddDate(0, 0, -1)
	trip.EndDate = &bad

	_, err := svc.Update(context.Background(), trip)

	assert.ErrorIs(t, err, domain.ErrValidation)
}

// ---- Delete tests ----------------------------------------------------------

func TestTripService_Delete_OK(t *testing.T) {
	r := &mockTripRepo{
		delete: func(_ context.Context, _ uuid.UUID) error { return nil },
	}
	svc := service.NewTripService(r)

	err := svc.Delete(context.Background(), uuid.New())

	assert.NoError(t, err)
}

func TestTripService_Delete_NotFound(t *testing.T) {
	r := &mockTripRepo{
		delete: func(_ context.Context, _ uuid.UUID) error { return domain.ErrNotFound },
	}
	svc := service.NewTripService(r)

	err := svc.Delete(context.Background(), uuid.New())

	assert.ErrorIs(t, err, domain.ErrNotFound)
}
