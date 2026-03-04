//go:build integration

package testutil_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkordes/rv-logbook/backend/migrations"
	"github.com/pkordes/rv-logbook/backend/testutil"
)

// TestMigration006_FixStopDatesMidnightToNoonEST verifies that migration 006
// correctly shifts only the stop dates that were stored as midnight UTC
// (the old frontend default) to 17:00 UTC (noon EST), while leaving rows
// whose time components are not exactly midnight untouched.
//
// This is an example of a data migration test — it exercises the Up and Down
// paths in isolation by controlling exactly which version the database is at
// before and after the migration runs.
func TestMigration006_FixStopDatesMidnightToNoonEST(t *testing.T) {
	db := testutil.NewSQLDB(t)
	ctx := context.Background()

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
	require.NoError(t, err)

	// Start from a known-clean baseline regardless of shared test DB state.
	_, err = provider.DownTo(ctx, 0)
	require.NoError(t, err, "reset to version 0")
	t.Cleanup(func() {
		// Always roll back to 0 so we don't leave state for other tests.
		if _, err := provider.DownTo(ctx, 0); err != nil {
			t.Logf("cleanup: down-to-0 failed: %v", err)
		}
	})

	// Apply schema migrations 001–005 to get the stops table in place.
	_, err = provider.UpTo(ctx, 5)
	require.NoError(t, err, "apply schema migrations 001-005")

	// Seed: insert a trip and two stops with different time components.
	tripID, stopMidnightID, stopNoonID, stopDepartedID := seedMigration006Data(t, db)
	_ = tripID

	// --- Apply migration 006 (Up) ---
	_, err = provider.UpTo(ctx, 6)
	require.NoError(t, err, "apply migration 006 up")

	// Midnight UTC row → should be shifted to 17:00 UTC.
	arrivedAt := queryArrivedAt(t, db, stopMidnightID)
	assert.Equal(t, 17, arrivedAt.UTC().Hour(),
		"midnight UTC arrived_at should be shifted to 17:00 UTC after Up")
	assert.Equal(t, 0, arrivedAt.UTC().Minute())
	assert.Equal(t, 0, arrivedAt.UTC().Second())

	// Non-midnight row → must be left untouched.
	arrivedAtNoon := queryArrivedAt(t, db, stopNoonID)
	assert.Equal(t, 14, arrivedAtNoon.UTC().Hour(),
		"non-midnight row should not be changed by migration 006 Up")

	// departed_at midnight → should also be shifted.
	departedAt := queryDepartedAt(t, db, stopDepartedID)
	require.NotNil(t, departedAt, "departed_at should not be nil")
	assert.Equal(t, 17, departedAt.UTC().Hour(),
		"midnight UTC departed_at should be shifted to 17:00 UTC after Up")

	// --- Roll back migration 006 (Down) ---
	_, err = provider.DownTo(ctx, 5)
	require.NoError(t, err, "roll back migration 006 down")

	// 17:00 UTC row → should be reverted to midnight UTC.
	arrivedAtReverted := queryArrivedAt(t, db, stopMidnightID)
	assert.Equal(t, 0, arrivedAtReverted.UTC().Hour(),
		"17:00 UTC arrived_at should revert to midnight UTC after Down")

	// Non-midnight row → still untouched by Down.
	arrivedAtNoonReverted := queryArrivedAt(t, db, stopNoonID)
	assert.Equal(t, 14, arrivedAtNoonReverted.UTC().Hour(),
		"non-midnight row should not be changed by migration 006 Down")
}

// seedMigration006Data inserts a trip and three stops with distinct
// arrived_at / departed_at values for the migration test to verify.
// Returns the IDs of the inserted rows.
func seedMigration006Data(t *testing.T, db *sql.DB) (tripID, stopMidnightID, stopNoonID, stopDepartedID string) {
	t.Helper()
	ctx := context.Background()

	err := db.QueryRowContext(ctx,
		`INSERT INTO trips (name, start_date) VALUES ('Test Trip', '2025-06-01') RETURNING id`,
	).Scan(&tripID)
	require.NoError(t, err, "insert trip")

	// Stop 1: arrived_at is midnight UTC — the old frontend default.
	midnight := time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC)
	err = db.QueryRowContext(ctx,
		`INSERT INTO stops (trip_id, name, arrived_at) VALUES ($1, 'Midnight Stop', $2) RETURNING id`,
		tripID, midnight,
	).Scan(&stopMidnightID)
	require.NoError(t, err, "insert midnight stop")

	// Stop 2: arrived_at is 14:00 UTC — has a non-zero time, must not be touched.
	noon := time.Date(2025, 6, 12, 14, 0, 0, 0, time.UTC)
	err = db.QueryRowContext(ctx,
		`INSERT INTO stops (trip_id, name, arrived_at) VALUES ($1, 'Noon Stop', $2) RETURNING id`,
		tripID, noon,
	).Scan(&stopNoonID)
	require.NoError(t, err, "insert non-midnight stop")

	// Stop 3: both arrived_at and departed_at are midnight UTC.
	err = db.QueryRowContext(ctx,
		`INSERT INTO stops (trip_id, name, arrived_at, departed_at) VALUES ($1, 'Departed Stop', $2, $3) RETURNING id`,
		tripID, midnight, time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
	).Scan(&stopDepartedID)
	require.NoError(t, err, "insert departed stop")

	return tripID, stopMidnightID, stopNoonID, stopDepartedID
}

func queryArrivedAt(t *testing.T, db *sql.DB, stopID string) time.Time {
	t.Helper()
	var ts time.Time
	err := db.QueryRowContext(context.Background(),
		`SELECT arrived_at FROM stops WHERE id = $1`, stopID,
	).Scan(&ts)
	require.NoError(t, err, "query arrived_at for stop %s", stopID)
	return ts
}

func queryDepartedAt(t *testing.T, db *sql.DB, stopID string) *time.Time {
	t.Helper()
	var ts *time.Time
	err := db.QueryRowContext(context.Background(),
		`SELECT departed_at FROM stops WHERE id = $1`, stopID,
	).Scan(&ts)
	require.NoError(t, err, "query departed_at for stop %s", stopID)
	return ts
}
