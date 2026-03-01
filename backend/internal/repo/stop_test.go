package repo_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/repo"
	"github.com/pkordes/rv-logbook/backend/testutil"
)

// newTestStopRepos opens a single transaction and returns both a TripRepo and
// a StopRepo backed by it. Tests can create a parent trip and child stops within
// the same transaction, which is rolled back automatically when the test finishes.
func newTestStopRepos(t *testing.T) (repo.TripRepo, repo.StopRepo) {
	t.Helper()
	pool := testutil.NewPool(t)

	tx, err := pool.Begin(context.Background())
	require.NoError(t, err, "begin transaction")

	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})

	return repo.NewTripRepo(tx), repo.NewStopRepo(tx)
}

// mustCreateTrip is a test helper that inserts a parent trip and fails the test
// if the insert does not succeed.
func mustCreateTrip(t *testing.T, r repo.TripRepo) domain.Trip {
	t.Helper()
	start := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	trip, err := r.Create(context.Background(), domain.Trip{
		Name:      "Test Trip",
		StartDate: start,
	})
	require.NoError(t, err, "create parent trip")
	return trip
}

// stopFixture returns a Stop ready for insertion against the given tripID.
func stopFixture(tripID uuid.UUID) domain.Stop {
	return domain.Stop{
		TripID:    tripID,
		Name:      "Camp Grounds A",
		Location:  "Yellowstone, WY",
		ArrivedAt: time.Date(2025, 6, 2, 10, 0, 0, 0, time.UTC),
		Notes:     "Great spot",
	}
}

func TestStopRepo_Create(t *testing.T) {
	tripRepo, stopRepo := newTestStopRepos(t)
	ctx := context.Background()

	parent := mustCreateTrip(t, tripRepo)
	input := stopFixture(parent.ID)

	got, err := stopRepo.Create(ctx, input)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.UUID{}, got.ID, "ID should be DB-generated UUID")
	assert.Equal(t, parent.ID, got.TripID)
	assert.Equal(t, input.Name, got.Name)
	assert.Equal(t, input.Location, got.Location)
	assert.True(t, got.ArrivedAt.Equal(input.ArrivedAt), "ArrivedAt mismatch")
	assert.Nil(t, got.DepartedAt, "DepartedAt should be nil when not provided")
	assert.Equal(t, input.Notes, got.Notes)
	assert.False(t, got.CreatedAt.IsZero(), "CreatedAt should be set by DB")
	assert.False(t, got.UpdatedAt.IsZero(), "UpdatedAt should be set by DB")
}

func TestStopRepo_Create_WithDepartedAt(t *testing.T) {
	tripRepo, stopRepo := newTestStopRepos(t)
	ctx := context.Background()

	parent := mustCreateTrip(t, tripRepo)
	input := stopFixture(parent.ID)
	departed := time.Date(2025, 6, 4, 9, 0, 0, 0, time.UTC)
	input.DepartedAt = &departed

	got, err := stopRepo.Create(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, got.DepartedAt, "DepartedAt should be set")
	assert.True(t, got.DepartedAt.Equal(departed), "DepartedAt mismatch")
}

func TestStopRepo_GetByID(t *testing.T) {
	tripRepo, stopRepo := newTestStopRepos(t)
	ctx := context.Background()

	parent := mustCreateTrip(t, tripRepo)
	created, err := stopRepo.Create(ctx, stopFixture(parent.ID))
	require.NoError(t, err)

	got, err := stopRepo.GetByID(ctx, parent.ID, created.ID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, created.Name, got.Name)
}

func TestStopRepo_GetByID_WrongTrip(t *testing.T) {
	tripRepo, stopRepo := newTestStopRepos(t)
	ctx := context.Background()

	parent := mustCreateTrip(t, tripRepo)
	created, err := stopRepo.Create(ctx, stopFixture(parent.ID))
	require.NoError(t, err)

	// Use a different (random) tripID — should not find the stop.
	_, err = stopRepo.GetByID(ctx, uuid.New(), created.ID)

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestStopRepo_ListByTripID(t *testing.T) {
	tripRepo, stopRepo := newTestStopRepos(t)
	ctx := context.Background()

	parent := mustCreateTrip(t, tripRepo)
	other := mustCreateTrip(t, tripRepo)

	// Create two stops for parent, one for other.
	_, err := stopRepo.Create(ctx, stopFixture(parent.ID))
	require.NoError(t, err)
	s2 := stopFixture(parent.ID)
	s2.Name = "Stop 2"
	_, err = stopRepo.Create(ctx, s2)
	require.NoError(t, err)
	_, err = stopRepo.Create(ctx, stopFixture(other.ID))
	require.NoError(t, err)

	got, err := stopRepo.ListByTripID(ctx, parent.ID)

	require.NoError(t, err)
	assert.Len(t, got, 2, "should return only stops for the given trip")
	for _, s := range got {
		assert.Equal(t, parent.ID, s.TripID)
	}
}

func TestStopRepo_ListByTripID_Empty(t *testing.T) {
	tripRepo, stopRepo := newTestStopRepos(t)
	ctx := context.Background()

	parent := mustCreateTrip(t, tripRepo)

	got, err := stopRepo.ListByTripID(ctx, parent.ID)

	require.NoError(t, err)
	assert.NotNil(t, got, "should return empty slice, not nil")
	assert.Len(t, got, 0)
}

func TestStopRepo_Update(t *testing.T) {
	tripRepo, stopRepo := newTestStopRepos(t)
	ctx := context.Background()

	parent := mustCreateTrip(t, tripRepo)
	created, err := stopRepo.Create(ctx, stopFixture(parent.ID))
	require.NoError(t, err)

	created.Name = "Updated Name"
	created.Location = "Grand Canyon, AZ"
	departed := time.Date(2025, 6, 5, 8, 0, 0, 0, time.UTC)
	created.DepartedAt = &departed

	updated, err := stopRepo.Update(ctx, created)

	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Grand Canyon, AZ", updated.Location)
	require.NotNil(t, updated.DepartedAt)
	assert.True(t, updated.DepartedAt.Equal(departed))
}

func TestStopRepo_Update_WrongTrip(t *testing.T) {
	tripRepo, stopRepo := newTestStopRepos(t)
	ctx := context.Background()

	parent := mustCreateTrip(t, tripRepo)
	created, err := stopRepo.Create(ctx, stopFixture(parent.ID))
	require.NoError(t, err)

	// Swap the TripID to a random UUID — should not find the stop.
	created.TripID = uuid.New()
	_, err = stopRepo.Update(ctx, created)

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestStopRepo_Delete(t *testing.T) {
	tripRepo, stopRepo := newTestStopRepos(t)
	ctx := context.Background()

	parent := mustCreateTrip(t, tripRepo)
	created, err := stopRepo.Create(ctx, stopFixture(parent.ID))
	require.NoError(t, err)

	err = stopRepo.Delete(ctx, parent.ID, created.ID)
	require.NoError(t, err)

	_, err = stopRepo.GetByID(ctx, parent.ID, created.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestStopRepo_Delete_WrongTrip(t *testing.T) {
	tripRepo, stopRepo := newTestStopRepos(t)
	ctx := context.Background()

	parent := mustCreateTrip(t, tripRepo)
	created, err := stopRepo.Create(ctx, stopFixture(parent.ID))
	require.NoError(t, err)

	err = stopRepo.Delete(ctx, uuid.New(), created.ID)

	assert.ErrorIs(t, err, domain.ErrNotFound)
}
