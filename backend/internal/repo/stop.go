package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
)

// StopRepo defines the persistence operations for Stops.
// All write and single-read operations are scoped by tripID to enforce ownership:
// a caller cannot read or mutate a stop that does not belong to the given trip.
type StopRepo interface {
	// Create inserts a new stop and returns the persisted record.
	Create(ctx context.Context, stop domain.Stop) (domain.Stop, error)

	// GetByID retrieves a single stop by its UUID, scoped to the given tripID.
	// Returns domain.ErrNotFound if no stop with that ID exists under that trip.
	GetByID(ctx context.Context, tripID, stopID uuid.UUID) (domain.Stop, error)

	// ListByTripID returns all stops for a trip ordered by arrived_at ascending.
	ListByTripID(ctx context.Context, tripID uuid.UUID) ([]domain.Stop, error)

	// ListByTripIDPaged returns one page of stops for a trip and the total count across all pages.
	// Results are ordered by arrived_at ascending.
	ListByTripIDPaged(ctx context.Context, tripID uuid.UUID, p domain.PaginationParams) ([]domain.Stop, int64, error)

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

// Create inserts a new stop row and returns the full persisted record.
func (r *pgStopRepo) Create(ctx context.Context, stop domain.Stop) (domain.Stop, error) {
	const q = `
		INSERT INTO stops (trip_id, name, location, arrived_at, departed_at, notes)
		VALUES (@trip_id, @name, @location, @arrived_at, @departed_at, @notes)
		RETURNING id, trip_id, name, location, arrived_at, departed_at, notes, created_at, updated_at`

	args := pgx.NamedArgs{
		"trip_id":     stop.TripID,
		"name":        stop.Name,
		"location":    nullableString(stop.Location),
		"arrived_at":  stop.ArrivedAt,
		"departed_at": stop.DepartedAt, // nil becomes NULL
		"notes":       nullableString(stop.Notes),
	}

	row := r.db.QueryRow(ctx, q, args)
	result, err := scanStop(row)
	if err != nil {
		return domain.Stop{}, fmt.Errorf("repo.StopRepo.Create: %w", err)
	}
	return result, nil
}

// GetByID retrieves a stop by primary key, scoped to the given tripID.
func (r *pgStopRepo) GetByID(ctx context.Context, tripID, stopID uuid.UUID) (domain.Stop, error) {
	const q = `
		SELECT id, trip_id, name, location, arrived_at, departed_at, notes, created_at, updated_at
		FROM stops
		WHERE id = @id AND trip_id = @trip_id`

	row := r.db.QueryRow(ctx, q, pgx.NamedArgs{"id": stopID, "trip_id": tripID})
	result, err := scanStop(row)
	if err != nil {
		return domain.Stop{}, fmt.Errorf("repo.StopRepo.GetByID: %w", err)
	}
	return result, nil
}

// ListByTripID returns all stops for a trip, ordered by arrival time.
func (r *pgStopRepo) ListByTripID(ctx context.Context, tripID uuid.UUID) ([]domain.Stop, error) {
	const q = `
		SELECT id, trip_id, name, location, arrived_at, departed_at, notes, created_at, updated_at
		FROM stops
		WHERE trip_id = @trip_id
		ORDER BY arrived_at ASC`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{"trip_id": tripID})
	if err != nil {
		return nil, fmt.Errorf("repo.StopRepo.ListByTripID: %w", err)
	}
	defer rows.Close()

	stops := []domain.Stop{} // initialise as empty slice so JSON serialises as [] not null
	for rows.Next() {
		s, err := scanStop(rows)
		if err != nil {
			return nil, fmt.Errorf("repo.StopRepo.ListByTripID: scan: %w", err)
		}
		stops = append(stops, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.StopRepo.ListByTripID: rows: %w", err)
	}

	return stops, nil
}

// ListByTripIDPaged returns one page of stops for a trip ordered by arrived_at ascending.
// Stub: returns empty results. Real implementation added in step 7.4 (Green).
func (r *pgStopRepo) ListByTripIDPaged(ctx context.Context, tripID uuid.UUID, p domain.PaginationParams) ([]domain.Stop, int64, error) {
	_, _ = tripID, p
	return nil, 0, nil
}

// Update overwrites the mutable fields of a stop and returns the updated record.
func (r *pgStopRepo) Update(ctx context.Context, stop domain.Stop) (domain.Stop, error) {
	const q = `
		UPDATE stops
		SET name        = @name,
		    location    = @location,
		    arrived_at  = @arrived_at,
		    departed_at = @departed_at,
		    notes       = @notes,
		    updated_at  = now()
		WHERE id = @id AND trip_id = @trip_id
		RETURNING id, trip_id, name, location, arrived_at, departed_at, notes, created_at, updated_at`

	args := pgx.NamedArgs{
		"id":          stop.ID,
		"trip_id":     stop.TripID,
		"name":        stop.Name,
		"location":    nullableString(stop.Location),
		"arrived_at":  stop.ArrivedAt,
		"departed_at": stop.DepartedAt,
		"notes":       nullableString(stop.Notes),
	}

	row := r.db.QueryRow(ctx, q, args)
	result, err := scanStop(row)
	if err != nil {
		return domain.Stop{}, fmt.Errorf("repo.StopRepo.Update: %w", err)
	}
	return result, nil
}

// Delete removes a stop by primary key, scoped to the given tripID.
func (r *pgStopRepo) Delete(ctx context.Context, tripID, stopID uuid.UUID) error {
	const q = `DELETE FROM stops WHERE id = @id AND trip_id = @trip_id`

	tag, err := r.db.Exec(ctx, q, pgx.NamedArgs{"id": stopID, "trip_id": tripID})
	if err != nil {
		return fmt.Errorf("repo.StopRepo.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("repo.StopRepo.Delete: %w", domain.ErrNotFound)
	}
	return nil
}

// scanStop maps a single database row into a domain.Stop.
// It handles UUID conversions and nullable location, departed_at, and notes columns.
func scanStop(s scanner) (domain.Stop, error) {
	var (
		t          domain.Stop
		id         pgtype.UUID
		tripID     pgtype.UUID
		location   *string
		departedAt *time.Time
		notes      *string
	)

	err := s.Scan(&id, &tripID, &t.Name, &location, &t.ArrivedAt, &departedAt, &notes, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Stop{}, domain.ErrNotFound
		}
		return domain.Stop{}, err
	}

	t.ID = uuid.UUID(id.Bytes)
	t.TripID = uuid.UUID(tripID.Bytes)
	if location != nil {
		t.Location = *location
	}
	t.DepartedAt = departedAt
	if notes != nil {
		t.Notes = *notes
	}

	return t, nil
}

// nullableString converts an empty Go string to nil so it is stored as SQL NULL.
// This keeps the domain model clean (no *string fields) while storing NULL in the DB
// for columns that are logically optional.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
