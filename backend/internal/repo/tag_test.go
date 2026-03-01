package repo_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/repo"
	"github.com/pkordes/rv-logbook/backend/testutil"
)

// newTestTagRepos opens a single transaction and returns TripRepo, StopRepo,
// and TagRepo all backed by the same tx — so tests can create full hierarchies
// (trip → stop → tag) within one rolled-back transaction.
func newTestTagRepos(t *testing.T) (repo.TripRepo, repo.StopRepo, repo.TagRepo) {
	t.Helper()
	pool := testutil.NewPool(t)

	tx, err := pool.Begin(context.Background())
	require.NoError(t, err, "begin transaction")

	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})

	return repo.NewTripRepo(tx), repo.NewStopRepo(tx), repo.NewTagRepo(tx)
}

// ---- Upsert ----------------------------------------------------------------

func TestTagRepo_Upsert_Create(t *testing.T) {
	_, _, tagRepo := newTestTagRepos(t)
	ctx := context.Background()

	got, err := tagRepo.Upsert(ctx, "Rocky Mountains", "rocky-mountains")

	require.NoError(t, err)
	assert.NotEmpty(t, got.ID)
	assert.Equal(t, "Rocky Mountains", got.Name)
	assert.Equal(t, "rocky-mountains", got.Slug)
	assert.False(t, got.CreatedAt.IsZero())
}

func TestTagRepo_Upsert_IdempotentBySlug(t *testing.T) {
	_, _, tagRepo := newTestTagRepos(t)
	ctx := context.Background()

	first, err := tagRepo.Upsert(ctx, "walmart", "walmart")
	require.NoError(t, err)

	// Different display name, same slug — must return the original row.
	second, err := tagRepo.Upsert(ctx, "Walmart", "walmart")
	require.NoError(t, err)

	assert.Equal(t, first.ID, second.ID, "same slug must return same tag")
	assert.Equal(t, "walmart", second.Name, "name should be the original, not the new casing")
}

// ---- List ------------------------------------------------------------------

func TestTagRepo_List_All(t *testing.T) {
	_, _, tagRepo := newTestTagRepos(t)
	ctx := context.Background()

	_, err := tagRepo.Upsert(ctx, "Mountains", "mountains")
	require.NoError(t, err)
	_, err = tagRepo.Upsert(ctx, "Desert", "desert")
	require.NoError(t, err)

	got, err := tagRepo.List(ctx, "")

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(got), 2, "should return at least the two inserted tags")
}

func TestTagRepo_List_Prefix(t *testing.T) {
	_, _, tagRepo := newTestTagRepos(t)
	ctx := context.Background()

	_, err := tagRepo.Upsert(ctx, "Mountains", "mountains")
	require.NoError(t, err)
	_, err = tagRepo.Upsert(ctx, "Mountain Lake", "mountain-lake")
	require.NoError(t, err)
	_, err = tagRepo.Upsert(ctx, "Desert", "desert")
	require.NoError(t, err)

	got, err := tagRepo.List(ctx, "mount")

	require.NoError(t, err)
	assert.Len(t, got, 2, "prefix 'mount' should match mountains and mountain-lake only")
	for _, tag := range got {
		assert.Contains(t, tag.Slug, "mount")
	}
}

func TestTagRepo_List_Empty(t *testing.T) {
	_, _, tagRepo := newTestTagRepos(t)

	got, err := tagRepo.List(context.Background(), "zzz-no-match")

	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got)
}

// ---- AddToStop / RemoveFromStop / ListByStop -------------------------------

func TestTagRepo_AddToStop(t *testing.T) {
	tripRepo, stopRepo, tagRepo := newTestTagRepos(t)
	ctx := context.Background()

	parent := mustCreateTrip(t, tripRepo)
	stop, err := stopRepo.Create(ctx, stopFixture(parent.ID))
	require.NoError(t, err)
	tag, err := tagRepo.Upsert(ctx, "Mountains", "mountains")
	require.NoError(t, err)

	err = tagRepo.AddToStop(ctx, stop.ID, tag.ID)

	require.NoError(t, err)
}

func TestTagRepo_AddToStop_Idempotent(t *testing.T) {
	tripRepo, stopRepo, tagRepo := newTestTagRepos(t)
	ctx := context.Background()

	parent := mustCreateTrip(t, tripRepo)
	stop, err := stopRepo.Create(ctx, stopFixture(parent.ID))
	require.NoError(t, err)
	tag, err := tagRepo.Upsert(ctx, "Mountains", "mountains")
	require.NoError(t, err)

	require.NoError(t, tagRepo.AddToStop(ctx, stop.ID, tag.ID))
	// Adding the same tag twice must not error.
	err = tagRepo.AddToStop(ctx, stop.ID, tag.ID)
	require.NoError(t, err)
}

func TestTagRepo_ListByStop(t *testing.T) {
	tripRepo, stopRepo, tagRepo := newTestTagRepos(t)
	ctx := context.Background()

	parent := mustCreateTrip(t, tripRepo)
	stop, err := stopRepo.Create(ctx, stopFixture(parent.ID))
	require.NoError(t, err)

	tag1, err := tagRepo.Upsert(ctx, "Mountains", "mountains")
	require.NoError(t, err)
	tag2, err := tagRepo.Upsert(ctx, "Desert", "desert")
	require.NoError(t, err)
	require.NoError(t, tagRepo.AddToStop(ctx, stop.ID, tag1.ID))
	require.NoError(t, tagRepo.AddToStop(ctx, stop.ID, tag2.ID))

	got, err := tagRepo.ListByStop(ctx, stop.ID)

	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestTagRepo_ListByStop_Empty(t *testing.T) {
	tripRepo, stopRepo, tagRepo := newTestTagRepos(t)
	ctx := context.Background()

	parent := mustCreateTrip(t, tripRepo)
	stop, err := stopRepo.Create(ctx, stopFixture(parent.ID))
	require.NoError(t, err)

	got, err := tagRepo.ListByStop(ctx, stop.ID)

	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got)
}

func TestTagRepo_RemoveFromStop(t *testing.T) {
	tripRepo, stopRepo, tagRepo := newTestTagRepos(t)
	ctx := context.Background()

	parent := mustCreateTrip(t, tripRepo)
	stop, err := stopRepo.Create(ctx, stopFixture(parent.ID))
	require.NoError(t, err)
	tag, err := tagRepo.Upsert(ctx, "Mountains", "mountains")
	require.NoError(t, err)
	require.NoError(t, tagRepo.AddToStop(ctx, stop.ID, tag.ID))

	err = tagRepo.RemoveFromStop(ctx, stop.ID, "mountains")

	require.NoError(t, err)
	remaining, err := tagRepo.ListByStop(ctx, stop.ID)
	require.NoError(t, err)
	assert.Empty(t, remaining)
}

func TestTagRepo_RemoveFromStop_NotFound(t *testing.T) {
	tripRepo, stopRepo, tagRepo := newTestTagRepos(t)
	ctx := context.Background()

	parent := mustCreateTrip(t, tripRepo)
	stop, err := stopRepo.Create(ctx, stopFixture(parent.ID))
	require.NoError(t, err)

	err = tagRepo.RemoveFromStop(ctx, stop.ID, "nonexistent-slug")

	assert.ErrorIs(t, err, domain.ErrNotFound)
}
