package repo

import (
	"context"
	"encoding/json"
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
	result.Tags = []domain.Tag{}
	return result, nil
}

// GetByID retrieves a stop by primary key, scoped to the given tripID.
func (r *pgStopRepo) GetByID(ctx context.Context, tripID, stopID uuid.UUID) (domain.Stop, error) {
	const q = `
		SELECT s.id, s.trip_id, s.name, s.location, s.arrived_at, s.departed_at, s.notes, s.created_at, s.updated_at,
		       COALESCE(
		           json_agg(
		               json_build_object('id', t.id, 'name', t.name, 'slug', t.slug, 'created_at', t.created_at)
		               ORDER BY t.slug
		           ) FILTER (WHERE t.id IS NOT NULL),
		           '[]'::json
		       ) AS tags
		FROM stops s
		LEFT JOIN stop_tags st ON st.stop_id = s.id
		LEFT JOIN tags t ON t.id = st.tag_id
		WHERE s.id = @id AND s.trip_id = @trip_id
		GROUP BY s.id`

	row := r.db.QueryRow(ctx, q, pgx.NamedArgs{"id": stopID, "trip_id": tripID})
	result, err := scanStopFull(row)
	if err != nil {
		return domain.Stop{}, fmt.Errorf("repo.StopRepo.GetByID: %w", err)
	}
	return result, nil
}

// ListByTripID returns all stops for a trip, ordered by arrival time.
func (r *pgStopRepo) ListByTripID(ctx context.Context, tripID uuid.UUID) ([]domain.Stop, error) {
	const q = `
		SELECT s.id, s.trip_id, s.name, s.location, s.arrived_at, s.departed_at, s.notes, s.created_at, s.updated_at,
		       COALESCE(
		           json_agg(
		               json_build_object('id', t.id, 'name', t.name, 'slug', t.slug, 'created_at', t.created_at)
		               ORDER BY t.slug
		           ) FILTER (WHERE t.id IS NOT NULL),
		           '[]'::json
		       ) AS tags
		FROM stops s
		LEFT JOIN stop_tags st ON st.stop_id = s.id
		LEFT JOIN tags t ON t.id = st.tag_id
		WHERE s.trip_id = @trip_id
		GROUP BY s.id
		ORDER BY s.arrived_at ASC`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{"trip_id": tripID})
	if err != nil {
		return nil, fmt.Errorf("repo.StopRepo.ListByTripID: %w", err)
	}
	defer rows.Close()

	stops := []domain.Stop{} // initialise as empty slice so JSON serialises as [] not null
	for rows.Next() {
		s, err := scanStopFull(rows)
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

// ListByTripIDPaged returns one page of stops for a trip ordered by arrived_at ascending,
// together with the total number of stops for that trip across all pages.
// Each stop includes its linked tags, aggregated in a single query.
func (r *pgStopRepo) ListByTripIDPaged(ctx context.Context, tripID uuid.UUID, p domain.PaginationParams) ([]domain.Stop, int64, error) {
	const countQ = `SELECT COUNT(*) FROM stops WHERE trip_id = @trip_id`

	var total int64
	if err := r.db.QueryRow(ctx, countQ, pgx.NamedArgs{"trip_id": tripID}).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("repo.StopRepo.ListByTripIDPaged: count: %w", err)
	}

	const q = `
		SELECT s.id, s.trip_id, s.name, s.location, s.arrived_at, s.departed_at, s.notes, s.created_at, s.updated_at,
		       COALESCE(
		           json_agg(
		               json_build_object('id', t.id, 'name', t.name, 'slug', t.slug, 'created_at', t.created_at)
		               ORDER BY t.slug
		           ) FILTER (WHERE t.id IS NOT NULL),
		           '[]'::json
		       ) AS tags
		FROM stops s
		LEFT JOIN stop_tags st ON st.stop_id = s.id
		LEFT JOIN tags t ON t.id = st.tag_id
		WHERE s.trip_id = @trip_id
		GROUP BY s.id
		ORDER BY s.arrived_at ASC
		LIMIT @limit OFFSET @offset`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{
		"trip_id": tripID,
		"limit":   p.Limit,
		"offset":  p.Offset(),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("repo.StopRepo.ListByTripIDPaged: query: %w", err)
	}
	defer rows.Close()

	stops := []domain.Stop{}
	for rows.Next() {
		s, err := scanStopFull(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("repo.StopRepo.ListByTripIDPaged: scan: %w", err)
		}
		stops = append(stops, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("repo.StopRepo.ListByTripIDPaged: rows: %w", err)
	}

	return stops, total, nil
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
	result.Tags = []domain.Tag{}
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
// Use this for write operations (Create, Update) whose RETURNING clause does not
// include the tag aggregation column.
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

// tagJSON is the intermediate type used to unmarshal the json_agg result from
// Postgres. UUIDs come back as strings (Postgres casts them automatically inside
// json_build_object), and created_at is an ISO 8601 timestamp.
type tagJSON struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

// scanStopFull maps a database row that includes the json_agg tags column into a
// domain.Stop with Tags populated. Use this for all read operations (GetByID,
// List*) whose SELECT includes the COALESCE/json_agg expression.
func scanStopFull(s scanner) (domain.Stop, error) {
	var (
		t          domain.Stop
		id         pgtype.UUID
		tripID     pgtype.UUID
		location   *string
		departedAt *time.Time
		notes      *string
		tagsJSON   []byte
	)

	err := s.Scan(&id, &tripID, &t.Name, &location, &t.ArrivedAt, &departedAt, &notes, &t.CreatedAt, &t.UpdatedAt, &tagsJSON)
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

	// Parse the JSON-aggregated tags. The COALESCE guarantees at least '[]',
	// so tagsJSON is never nil or empty.
	var rows []tagJSON
	if err := json.Unmarshal(tagsJSON, &rows); err != nil {
		return domain.Stop{}, fmt.Errorf("scanStopFull: unmarshal tags: %w", err)
	}

	t.Tags = make([]domain.Tag, len(rows))
	for i, r := range rows {
		id, err := uuid.Parse(r.ID)
		if err != nil {
			return domain.Stop{}, fmt.Errorf("scanStopFull: parse tag id: %w", err)
		}
		t.Tags[i] = domain.Tag{ID: id, Name: r.Name, Slug: r.Slug, CreatedAt: r.CreatedAt}
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
