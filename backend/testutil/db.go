// Package testutil provides shared helpers for integration tests.
// Helpers in this package skip automatically when required environment
// variables are not set, so unit tests can run without a running database.
package testutil

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // registers "pgx" driver for database/sql
)

// NewPool opens a *pgxpool.Pool connected to the database specified by the
// TEST_DATABASE_URL environment variable.
//
// The test is skipped automatically if TEST_DATABASE_URL is not set, so
// integration tests are opt-in and never break CI environments that lack a DB.
// The pool is closed automatically when the test (and all its subtests) finish.
func NewPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := requireDSN(t)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("testutil.NewPool: open pool: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Fatalf("testutil.NewPool: ping: %v", err)
	}

	t.Cleanup(pool.Close)
	return pool
}

// NewSQLDB opens a *sql.DB connected to the database specified by the
// TEST_DATABASE_URL environment variable using the pgx database/sql driver.
//
// Use this when you need a *sql.DB rather than a *pgxpool.Pool — for example,
// when driving goose migrations in integration tests.
// The connection is closed automatically when the test finishes.
func NewSQLDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := requireDSN(t)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("testutil.NewSQLDB: open: %v", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		t.Fatalf("testutil.NewSQLDB: ping: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

// MustOpenSQLDB opens a *sql.DB for the given DSN and panics on any error.
// Use this in TestMain functions where no *testing.T is available.
// Callers are responsible for closing the returned *sql.DB.
func MustOpenSQLDB(dsn string) *sql.DB {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic("testutil.MustOpenSQLDB: open: " + err.Error())
	}
	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		panic("testutil.MustOpenSQLDB: ping: " + err.Error())
	}
	return db
}

// requireDSN returns the TEST_DATABASE_URL environment variable value,
// skipping the test if it is not set.
func requireDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	return dsn
}
