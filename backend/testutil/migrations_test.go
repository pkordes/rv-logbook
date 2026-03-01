package testutil_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkordes/rv-logbook/backend/migrations"
	"github.com/pkordes/rv-logbook/backend/testutil"
)

// TestMigrations is an integration test that verifies the full migration
// round-trip against a real Postgres database:
//
//  1. Apply all migrations (goose up).
//  2. Assert every expected table exists.
//  3. Roll back all migrations (goose reset).
//  4. Assert every table has been removed.
//
// The test is skipped automatically when TEST_DATABASE_URL is not set.
func TestMigrations(t *testing.T) {
	db := testutil.NewSQLDB(t)

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		migrations.FS,
	)
	require.NoError(t, err, "create goose provider")

	ctx := context.Background()

	// --- Ensure a clean baseline before testing ---
	// Another package's TestMain may have already applied migrations against this
	// shared test DB. Reset to version 0 first so this test is self-contained and
	// order-independent, whether run alone or as part of the full suite.
	if _, err := provider.DownTo(ctx, 0); err != nil {
		t.Fatalf("TestMigrations: initial reset: %v", err)
	}

	// --- Red → Green: apply all migrations ---
	results, err := provider.Up(ctx)
	require.NoError(t, err, "goose up")
	assert.NotEmpty(t, results, "expected at least one migration to be applied")

	// Verify all expected tables exist after applying migrations.
	for _, table := range []string{"trips", "stops", "tags", "stop_tags"} {
		assertTableExists(t, db, table)
	}

	// --- Roll back all migrations ---
	_, err = provider.DownTo(ctx, 0)
	require.NoError(t, err, "goose down-to 0")

	// Verify all tables have been removed after rolling back.
	for _, table := range []string{"trips", "stops", "tags", "stop_tags"} {
		assertTableNotExists(t, db, table)
	}
}

// assertTableExists fails the test if the named table does not exist in the
// public schema of the connected database.
func assertTableExists(t *testing.T, db *sql.DB, table string) {
	t.Helper()
	assertTablePresence(t, db, table, true)
}

// assertTableNotExists fails the test if the named table exists in the
// public schema of the connected database.
func assertTableNotExists(t *testing.T, db *sql.DB, table string) {
	t.Helper()
	assertTablePresence(t, db, table, false)
}

func assertTablePresence(t *testing.T, db *sql.DB, table string, shouldExist bool) {
	t.Helper()

	// Use the information_schema to check table existence in a portable way.
	const q = `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public'
			AND   table_name   = $1
		)`
	var exists bool
	err := db.QueryRowContext(context.Background(), q, table).Scan(&exists)
	require.NoError(t, err, "check table existence for %q", table)

	if shouldExist {
		assert.True(t, exists, "expected table %q to exist", table)
	} else {
		assert.False(t, exists, "expected table %q to not exist", table)
	}
}
