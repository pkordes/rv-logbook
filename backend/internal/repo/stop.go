package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
)

// StopRepo defines the persistence operations for Stops.
// All write and single-read operations are scoped by tripID to enforce ownership.
type StopRepo interface {
	// Create inserts a new stop and returns the persisted record.
	Create(ctx context.Context, stop domain.Stop) (domain.Stop, error)

	// GetByID retrieves a single stop by its UUID, scoped to the given tripID.
	// Returns domain.ErrNotFound if no stop with that ID exists under that trip.
	GetByID(ctx context.Context, tripID, stopID uuid.UUID) (domain.Stop, error)

	// ListByTripID returns all stops for a trip ordered by arrived_at ascending.
	ListByTripID(ctx context.Context, tripID uuid.UUID) ([]domain.Stop, error)

	// Update overwrites the mutable fields of a stop, scoped to the given tripID.
	// Returns domain.ErrNotFound if no stop with that ID exists under that trip.
	Update(ctx context.Context, stop domain.Stop) (domain.Stop, error)

	// Delete removes a stop by ID, scoped to the given tripID.
	// Returns domain.ErrNotFound if no stop with that ID exists under that trip.
	Delete(ctx context.Context, tripID, stopID uuid.UUID) error
}

// pgStopRepo is the Postgres implementation of StopRepo.
type pgStopRepo struct {
	db db
}

// NewStopRepo constructs a StopRepo backed by the provided db connection.
// In production pass *pgxpool.Pool; in tests pass a pgx.Tx for rollback isolation.
func NewStopRepo(db db) StopRepo {
	return &pgStopRepo{db: db}
}

func (r *pgStopRepo) Create(ctx context.Context, stop domain.Stop) (domain.Stop, error) {
	return domain.Stop{}, fmt.Errorf("not implemented")
}

func (r *pgStopRepo) GetByID(ctx context.Context, tripID, stopID uuid.UUID) (domain.Stop, error) {
	return domain.Stop{}, fmt.Errorf("not implemented")
}

func (r *pgStopRepo) ListByTripID(ctx context.Context, tripID uuid.UUID) ([]domain.Stop, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *pgStopRepo) Update(ctx context.Context, stop domain.Stop) (domain.Stop, error) {
	return domain.Stop{}, fmt.Errorf("not implemented")
}

func (r *pgStopRepo) Delete(ctx context.Context, tripID, stopID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
