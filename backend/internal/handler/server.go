// Package handler implements the HTTP handlers for the RV Logbook API.
// All handlers are methods on Server, which implements gen.StrictServerInterface.
// Methods are split into domain-specific files (health.go, trip.go, etc.) but
// all share the same Server struct so they can access its dependencies.
package handler

import (
	"context"

	"github.com/google/uuid"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
)

// TripServicer defines the business operations the trip handler depends on.
// Defining the interface here (in the consumer package) follows the Go
// convention: "accept interfaces, return concrete types". It lets handler
// tests inject a mock without touching the database or service layer.
type TripServicer interface {
	Create(ctx context.Context, trip domain.Trip) (domain.Trip, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Trip, error)
	List(ctx context.Context) ([]domain.Trip, error)
	Update(ctx context.Context, trip domain.Trip) (domain.Trip, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// Server implements gen.StrictServerInterface for all API endpoints.
// Wire it in main.go via gen.NewStrictHandler(server, nil).
// Methods are in domain-specific files but all operate on this struct.
type Server struct {
	trips TripServicer
}

// NewServer constructs the Server with all its dependencies.
func NewServer(trips TripServicer) *Server {
	return &Server{trips: trips}
}

// NewHealthHandler returns a Server for health-check-only use.
// Keeps existing handler tests compiling without modification.
func NewHealthHandler() *Server {
	return NewServer(nil)
}
