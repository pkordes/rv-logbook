package repo_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/pressly/goose/v3"

	"github.com/pkordes/rv-logbook/backend/migrations"
	"github.com/pkordes/rv-logbook/backend/testutil"
)

// TestMain runs before any test in the repo_test package.
// It applies all pending migrations to the test database so individual tests
// never need to think about schema state.
//
// This is the Go equivalent of a JUnit @BeforeAll — it runs once for the
// entire test binary, not once per test function.
func TestMain(m *testing.M) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		// No test DB configured — skip all tests in this package cleanly.
		os.Exit(m.Run())
	}

	// Use a plain *sql.DB for goose (it needs database/sql, not pgx pool).
	// We construct it manually here rather than through testutil.NewPool
	// because TestMain doesn't have a *testing.T to pass.
	db := testutil.MustOpenSQLDB(os.Getenv("TEST_DATABASE_URL"))
	defer db.Close()

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
	if err != nil {
		log.Fatalf("TestMain: create goose provider: %v", err)
	}

	if _, err := provider.Up(context.Background()); err != nil {
		log.Fatalf("TestMain: run migrations: %v", err)
	}

	os.Exit(m.Run())
}
