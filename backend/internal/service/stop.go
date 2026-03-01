package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/repo"
)

// StopService implements business logic for Stop operations.
// It holds both repos because creating a stop requires verifying the parent
// trip exists before inserting.
type StopService struct {
	trips repo.TripRepo
	stops repo.StopRepo
}

// NewStopService constructs a StopService backed by the provided repos.
func NewStopService(trips repo.TripRepo, stops repo.StopRepo) *StopService {
	return &StopService{trips: trips, stops: stops}
}

func (s *StopService) Create(ctx context.Context, stop domain.Stop) (domain.Stop, error) {
	return domain.Stop{}, fmt.Errorf("not implemented")
}

func (s *StopService) GetByID(ctx context.Context, tripID, stopID uuid.UUID) (domain.Stop, error) {
	return domain.Stop{}, fmt.Errorf("not implemented")
}

func (s *StopService) ListByTripID(ctx context.Context, tripID uuid.UUID) ([]domain.Stop, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *StopService) Update(ctx context.Context, stop domain.Stop) (domain.Stop, error) {
	return domain.Stop{}, fmt.Errorf("not implemented")
}

func (s *StopService) Delete(ctx context.Context, tripID, stopID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
