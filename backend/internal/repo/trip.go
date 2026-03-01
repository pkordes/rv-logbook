// Package repo contains all database access logic for the RV Logbook API.
// Each resource has its own file with an interface and a Postgres implementation.
// No business logic lives here — only SQL and type mapping.
package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
)

// db is the minimal interface satisfied by *pgxpool.Pool, pgx.Conn, and pgx.Tx.
// Accepting this interface instead of *pgxpool.Pool directly allows integration
// tests to pass a transaction that is rolled back after each test, giving free
// per-test isolation without any manual cleanup.
type db interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// TripRepo defines the persistence operations for Trips.
// The service layer depends on this interface, not the concrete Postgres implementation,
// which allows the service to be unit-tested with a mock.
type TripRepo interface {
	// Create inserts a new trip and returns the persisted record (with DB-generated
	// id, created_at, and updated_at populated).
	Create(ctx context.Context, trip domain.Trip) (domain.Trip, error)

	// GetByID retrieves a single trip by its UUID primary key.
	// Returns domain.ErrNotFound if no trip with that ID exists.
	GetByID(ctx context.Context, id uuid.UUID) (domain.Trip, error)

	// List returns all trips ordered by start_date descending.
	List(ctx context.Context) ([]domain.Trip, error)

	// Update overwrites the mutable fields of an existing trip and returns the
	// updated record. Returns domain.ErrNotFound if no trip with that ID exists.
	Update(ctx context.Context, trip domain.Trip) (domain.Trip, error)

	// Delete removes a trip by ID. Returns domain.ErrNotFound if it does not exist.
	Delete(ctx context.Context, id uuid.UUID) error
}

// pgTripRepo is the Postgres implementation of TripRepo.
type pgTripRepo struct {
	db db
}

// NewTripRepo constructs a TripRepo backed by the provided db connection.
// In production pass *pgxpool.Pool; in tests pass a pgx.Tx for rollback isolation.
func NewTripRepo(db db) TripRepo {
	return &pgTripRepo{db: db}
}

// Create inserts a new trip row and returns the full persisted record.
func (r *pgTripRepo) Create(ctx context.Context, trip domain.Trip) (domain.Trip, error) {
	const q = `
		INSERT INTO trips (name, start_date, end_date, notes)
		VALUES (@name, @start_date, @end_date, @notes)
		RETURNING id, name, start_date, end_date, notes, created_at, updated_at`

	args := pgx.NamedArgs{
		"name":       trip.Name,
		"start_date": trip.StartDate,
		"end_date":   trip.EndDate, // nil becomes NULL
		"notes":      trip.Notes,
	}

	row := r.db.QueryRow(ctx, q, args)
	result, err := scanTrip(row)
	if err != nil {
		return domain.Trip{}, fmt.Errorf("repo.TripRepo.Create: %w", err)
	}
	return result, nil
}

// GetByID retrieves a trip by primary key.
func (r *pgTripRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	const q = `
		SELECT id, name, start_date, end_date, notes, created_at, updated_at
		FROM trips
		WHERE id = @id`

	row := r.db.QueryRow(ctx, q, pgx.NamedArgs{"id": id})
	result, err := scanTrip(row)
	if err != nil {
		return domain.Trip{}, fmt.Errorf("repo.TripRepo.GetByID: %w", err)
	}
	return result, nil
}

// List returns all trips ordered by start_date descending (most recent first).
func (r *pgTripRepo) List(ctx context.Context) ([]domain.Trip, error) {
	const q = `
		SELECT id, name, start_date, end_date, notes, created_at, updated_at
		FROM trips
		ORDER BY start_date DESC`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("repo.TripRepo.List: %w", err)
	}
	defer rows.Close()

	var trips []domain.Trip
	for rows.Next() {
		t, err := scanTrip(rows)
		if err != nil {
			return nil, fmt.Errorf("repo.TripRepo.List: scan: %w", err)
		}
		trips = append(trips, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.TripRepo.List: rows: %w", err)
	}

	return trips, nil
}

// Update overwrites the mutable fields of a trip and returns the updated record.
func (r *pgTripRepo) Update(ctx context.Context, trip domain.Trip) (domain.Trip, error) {
	const q = `
		UPDATE trips
		SET name       = @name,
		    start_date = @start_date,
		    end_date   = @end_date,
		    notes      = @notes,
		    updated_at = now()
		WHERE id = @id
		RETURNING id, name, start_date, end_date, notes, created_at, updated_at`

	args := pgx.NamedArgs{
		"id":         trip.ID,
		"name":       trip.Name,
		"start_date": trip.StartDate,
		"end_date":   trip.EndDate,
		"notes":      trip.Notes,
	}

	row := r.db.QueryRow(ctx, q, args)
	result, err := scanTrip(row)
	if err != nil {
		return domain.Trip{}, fmt.Errorf("repo.TripRepo.Update: %w", err)
	}
	return result, nil
}

// Delete removes a trip by primary key.
func (r *pgTripRepo) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM trips WHERE id = @id`

	tag, err := r.db.Exec(ctx, q, pgx.NamedArgs{"id": id})
	if err != nil {
		return fmt.Errorf("repo.TripRepo.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("repo.TripRepo.Delete: %w", domain.ErrNotFound)
	}
	return nil
}

// scanner is satisfied by both pgx.Row and pgx.Rows, allowing scanTrip to be
// reused for both QueryRow and Query calls.
type scanner interface {
	Scan(dest ...any) error
}

// scanTrip maps a single database row into a domain.Trip.
// It handles the UUID and nullable end_date conversions.
func scanTrip(s scanner) (domain.Trip, error) {
	var (
		t       domain.Trip
		id      pgtype.UUID
		endDate pgtype.Date
		sdRaw   pgtype.Date
	)

	err := s.Scan(&id, &t.Name, &sdRaw, &endDate, &t.Notes, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Trip{}, domain.ErrNotFound
		}
		return domain.Trip{}, err
	}

	t.ID = uuid.UUID(id.Bytes)
	t.StartDate = sdRaw.Time
	if endDate.Valid {
		ed := endDate.Time
		t.EndDate = &ed
	}

	return t, nil
}
