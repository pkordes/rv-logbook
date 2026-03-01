package repo_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/repo"
	"github.com/pkordes/rv-logbook/backend/testutil"
)

// newTestRepo opens a transaction against the test database and returns a
// TripRepo backed by that transaction. The transaction is automatically rolled
// back when the test finishes, giving free per-test isolation.
//
// Requires TEST_DATABASE_URL to be set and all migrations to be applied
// (run `make db/migrate` against the test database before running these tests).
func newTestRepo(t *testing.T) repo.TripRepo {
	t.Helper()
	pool := testutil.NewPool(t)

	tx, err := pool.Begin(context.Background())
	require.NoError(t, err, "begin transaction")

	t.Cleanup(func() {
		// Rollback discards all changes made during the test — no cleanup SQL needed.
		_ = tx.Rollback(context.Background())
	})

	return repo.NewTripRepo(tx)
}

// tripFixture returns a domain.Trip with sensible defaults for use in tests.
// Callers can override individual fields after calling this function.
func tripFixture() domain.Trip {
	start := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	return domain.Trip{
		Name:      "Summer Tour",
		StartDate: start,
		EndDate:   &end,
		Notes:     "Test notes",
	}
}

func TestTripRepo_Create(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()

	input := tripFixture()
	got, err := r.Create(ctx, input)

	require.NoError(t, err)
	assert.NotEqual(t, [16]byte{}, got.ID, "ID should be DB-generated UUID")
	assert.Equal(t, input.Name, got.Name)
	assert.True(t, got.StartDate.Equal(input.StartDate), "StartDate mismatch")
	require.NotNil(t, got.EndDate, "EndDate should not be nil")
	assert.True(t, got.EndDate.Equal(*input.EndDate), "EndDate mismatch")
	assert.Equal(t, input.Notes, got.Notes)
	assert.False(t, got.CreatedAt.IsZero(), "CreatedAt should be set by DB")
	assert.False(t, got.UpdatedAt.IsZero(), "UpdatedAt should be set by DB")
}

func TestTripRepo_Create_NilEndDate(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()

	input := tripFixture()
	input.EndDate = nil // trip still in progress

	got, err := r.Create(ctx, input)

	require.NoError(t, err)
	assert.Nil(t, got.EndDate, "EndDate should be nil when not provided")
}

func TestTripRepo_GetByID(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()

	created, err := r.Create(ctx, tripFixture())
	require.NoError(t, err)

	got, err := r.GetByID(ctx, created.ID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, created.Name, got.Name)
}

func TestTripRepo_GetByID_NotFound(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()

	// Use a random UUID that was never inserted.
	id := [16]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	_, err := r.GetByID(ctx, id)

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestTripRepo_List(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()

	// Create two trips.
	t1 := tripFixture()
	t1.Name = "First Trip"

	t2 := tripFixture()
	t2.Name = "Second Trip"
	t2.StartDate = t1.StartDate.AddDate(0, 1, 0) // one month later

	_, err := r.Create(ctx, t1)
	require.NoError(t, err)
	_, err = r.Create(ctx, t2)
	require.NoError(t, err)

	trips, err := r.List(ctx)

	require.NoError(t, err)
	require.GreaterOrEqual(t, len(trips), 2, "should return at least the two created trips")

	// List is ordered by start_date DESC — t2 (later start) should come first.
	var names []string
	for _, tr := range trips {
		names = append(names, tr.Name)
	}
	assert.Contains(t, names, "First Trip")
	assert.Contains(t, names, "Second Trip")
}

func TestTripRepo_Update(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()

	created, err := r.Create(ctx, tripFixture())
	require.NoError(t, err)

	created.Name = "Updated Name"
	created.Notes = "Updated notes"
	created.EndDate = nil // clear end date

	updated, err := r.Update(ctx, created)

	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated notes", updated.Notes)
	assert.Nil(t, updated.EndDate)
	// updated_at should be refreshed — may be equal to created_at in fast tests,
	// but must not be zero.
	assert.False(t, updated.UpdatedAt.IsZero())
}

func TestTripRepo_Update_NotFound(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()

	ghost := tripFixture()
	ghost.ID = [16]byte{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef,
		0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef}

	_, err := r.Update(ctx, ghost)

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestTripRepo_Delete(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()

	created, err := r.Create(ctx, tripFixture())
	require.NoError(t, err)

	err = r.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = r.GetByID(ctx, created.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound, "trip should be gone after delete")
}

func TestTripRepo_Delete_NotFound(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()

	id := [16]byte{0xca, 0xfe, 0xba, 0xbe, 0xca, 0xfe, 0xba, 0xbe,
		0xca, 0xfe, 0xba, 0xbe, 0xca, 0xfe, 0xba, 0xbe}

	err := r.Delete(ctx, id)

	assert.ErrorIs(t, err, domain.ErrNotFound)
}
