// Package service contains the business logic for the RV Logbook API.
// Services validate inputs, enforce business rules, and orchestrate repo calls.
// No SQL lives here — services depend on repo interfaces, not implementations.
package service

import (
	"context"
	"fmt"
	"strings"

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
// Returns domain.ErrValidation if the input violates business rules.
func (s *TripService) Create(ctx context.Context, trip domain.Trip) (domain.Trip, error) {
	if err := validateTrip(trip); err != nil {
		return domain.Trip{}, err
	}
	result, err := s.repo.Create(ctx, trip)
	if err != nil {
		return domain.Trip{}, fmt.Errorf("service.TripService.Create: %w", err)
	}
	return result, nil
}

// GetByID returns a single trip by ID.
// Returns domain.ErrNotFound if no trip with that ID exists.
func (s *TripService) GetByID(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	result, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Trip{}, fmt.Errorf("service.TripService.GetByID: %w", err)
	}
	return result, nil
}

// List returns all trips ordered by start_date descending.
// Always returns a non-nil slice so callers can safely range over it.
func (s *TripService) List(ctx context.Context) ([]domain.Trip, error) {
	trips, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("service.TripService.List: %w", err)
	}
	if trips == nil {
		return []domain.Trip{}, nil
	}
	return trips, nil
}

// ListPaged returns one page of trips and the total count across all pages.
// The caller controls page and limit via domain.PaginationParams.
func (s *TripService) ListPaged(ctx context.Context, p domain.PaginationParams) ([]domain.Trip, int64, error) {
	trips, total, err := s.repo.ListPaged(ctx, p)
	if err != nil {
		return nil, 0, fmt.Errorf("service.TripService.ListPaged: %w", err)
	}
	if trips == nil {
		trips = []domain.Trip{}
	}
	return trips, total, nil
}

// Update validates and persists changes to an existing trip.
// Returns domain.ErrValidation for invalid input, domain.ErrNotFound if the
// trip does not exist.
func (s *TripService) Update(ctx context.Context, trip domain.Trip) (domain.Trip, error) {
	if err := validateTrip(trip); err != nil {
		return domain.Trip{}, err
	}
	result, err := s.repo.Update(ctx, trip)
	if err != nil {
		return domain.Trip{}, fmt.Errorf("service.TripService.Update: %w", err)
	}
	return result, nil
}

// Delete removes a trip by ID.
// Returns domain.ErrNotFound if no trip with that ID exists.
func (s *TripService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("service.TripService.Delete: %w", err)
	}
	return nil
}

// validateTrip enforces business rules common to both Create and Update.
//   - Name must be non-empty (whitespace-only names are rejected).
//   - EndDate, if set, must not be before StartDate.
func validateTrip(trip domain.Trip) error {
	if strings.TrimSpace(trip.Name) == "" {
		return fmt.Errorf("%w: name is required", domain.ErrValidation)
	}
	if trip.EndDate != nil && trip.EndDate.Before(trip.StartDate) {
		return fmt.Errorf("%w: end_date must not be before start_date", domain.ErrValidation)
	}
	return nil
}
