//go:build integration

package apitest_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/pressly/goose/v3"

	"github.com/pkordes/rv-logbook/backend/migrations"
	"github.com/pkordes/rv-logbook/backend/testutil"
)

// TestMain applies all pending migrations once before any test in this package
// runs. This is identical to the pattern in internal/repo — each package that
// runs against a real DB manages its own migration setup.
//
// Unlike the repo tests, apitest tests cannot use per-test transactions for
// isolation (an HTTP handler opens its own DB connection). Each test creates
// its own data and removes it via t.Cleanup.
func TestMain(m *testing.M) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		os.Exit(m.Run())
	}

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
