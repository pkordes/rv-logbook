// Package service contains the business logic for the RV Logbook API.
// Services validate inputs, enforce business rules, and orchestrate repo calls.
// No SQL lives here — services depend on repo interfaces, not implementations.
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/repo"
)

// TripService implements business logic for Trip operations.
type TripService struct {
	repo repo.TripRepo
}

// NewTripService constructs a TripService backed by the provided TripRepo.
func NewTripService(r repo.TripRepo) *TripService {
	return &TripService{repo: r}
}

// Create validates and persists a new trip.
func (s *TripService) Create(ctx context.Context, trip domain.Trip) (domain.Trip, error) {
	return domain.Trip{}, fmt.Errorf("not implemented")
}

// GetByID returns a single trip by ID.
func (s *TripService) GetByID(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	return domain.Trip{}, fmt.Errorf("not implemented")
}

// List returns all trips.
func (s *TripService) List(ctx context.Context) ([]domain.Trip, error) {
	return nil, fmt.Errorf("not implemented")
}

// Update validates and updates an existing trip.
func (s *TripService) Update(ctx context.Context, trip domain.Trip) (domain.Trip, error) {
	return domain.Trip{}, fmt.Errorf("not implemented")
}

// Delete removes a trip by ID.
func (s *TripService) Delete(ctx context.Context, id uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
