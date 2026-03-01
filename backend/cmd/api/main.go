// Package main is the entry point for the RV Logbook API server.
// Its sole responsibility is wiring dependencies together and starting the server.
// No business logic belongs here.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pkordes/rv-logbook/backend/internal/config"
	"github.com/pkordes/rv-logbook/backend/internal/handler"
	"github.com/pkordes/rv-logbook/backend/internal/handler/gen"
	"github.com/pkordes/rv-logbook/backend/internal/middleware"
	"github.com/pkordes/rv-logbook/backend/internal/repo"
	"github.com/pkordes/rv-logbook/backend/internal/service"
)

func main() {
	// --- Config -----------------------------------------------------------
	cfg, err := config.Load()
	if err != nil {
		// Use plain stderr before the logger is configured.
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}

	// --- Logger -----------------------------------------------------------
	// log/slog is the stdlib structured logger introduced in Go 1.21.
	// JSON handler writes machine-readable output suitable for log aggregators.
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	// --- Database ---------------------------------------------------------
	// pgxpool manages a pool of Postgres connections.
	// New() does not open connections immediately — the first query does.
	// Use per-operation timeouts so startup fails fast when the DB is unreachable.
	poolCtx, poolCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer poolCancel()

	pool, err := pgxpool.New(poolCtx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Verify the DB is reachable before accepting traffic.
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer pingCancel()

	if err := pool.Ping(pingCtx); err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	slog.Info("database connection established")

	// --- Router -----------------------------------------------------------
	// Middleware is applied in order: RequestID → RealIP → Logger → Recoverer.
	// RequestID generates a unique trace ID per request.
	// RealIP sets r.RemoteAddr from X-Forwarded-For / X-Real-IP (safe behind a proxy).
	// SlogLogger writes one structured JSON log line per request.
	// Recoverer catches panics and returns HTTP 500 instead of crashing.
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.NewSlogLogger(logger))
	r.Use(chimiddleware.Recoverer)

	// Wire the dependency chain: pool → repo → service → handler.
	tripRepo := repo.NewTripRepo(pool)
	tripService := service.NewTripService(tripRepo)
	server := handler.NewServer(tripService)
	r.Mount("/", gen.Handler(gen.NewStrictHandler(server, nil)))

	// --- HTTP Server ------------------------------------------------------
	// Explicit timeouts prevent slowloris and resource exhaustion attacks.
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown: wait for OS signal or server error, then give in-flight
	// requests up to 15 seconds to complete before forcefully closing.
	stop := make(chan os.Signal, 1)
	serverErr := make(chan error, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			serverErr <- err
		}
	}()

	select {
	case <-stop:
		slog.Info("shutting down server")
	case err := <-serverErr:
		slog.Error("shutting down due to server error", "error", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
