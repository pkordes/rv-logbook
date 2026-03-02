//go:build integration

// Package apitest provides a helper for wiring the full dependency chain
// (handler → service → repo → Postgres) into an httptest.Server.
//
// This is the "real stack" test layer — HTTP requests flow through every layer
// with no mocking. Use it to verify that the wiring in main.go is correct and
// that the layers interact as expected.
//
// Contrast with:
//   - internal/handler tests: handler only, service is a mock
//   - internal/repo tests:    SQL only, no HTTP
//   - apitest (this package): full stack, real HTTP + real DB
package apitest

import (
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pkordes/rv-logbook/backend/internal/handler"
	"github.com/pkordes/rv-logbook/backend/internal/handler/gen"
	"github.com/pkordes/rv-logbook/backend/internal/repo"
	"github.com/pkordes/rv-logbook/backend/internal/service"
)

// NewServer wires the real repo → service → handler stack against the supplied
// pool and returns a running *httptest.Server. The server is automatically
// closed when the test and all its subtests finish.
//
// The router contains only the Recoverer middleware — logging is omitted to
// keep test output clean. All other production middleware (CORS, body size
// limit) is also omitted because it is tested independently.
func NewServer(t *testing.T, pool *pgxpool.Pool) *httptest.Server {
	t.Helper()

	tripRepo := repo.NewTripRepo(pool)
	stopRepo := repo.NewStopRepo(pool)
	tagRepo := repo.NewTagRepo(pool)

	tripService := service.NewTripService(tripRepo)
	stopService := service.NewStopService(tripRepo, stopRepo, tagRepo)
	tagService := service.NewTagService(tagRepo)
	exportService := service.NewExportService(tripRepo, stopRepo, tagRepo)

	srv := handler.NewServer(tripService, stopService, tagService, exportService)

	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Mount("/", gen.Handler(gen.NewStrictHandler(srv, nil)))

	ts := httptest.NewServer(r)
	t.Cleanup(ts.Close)
	return ts
}
