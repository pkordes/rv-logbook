package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
)

// TagRepo defines the persistence operations for Tags and the stop_tags join table.
type TagRepo interface {
	// Upsert inserts a tag by slug, or returns the existing tag if the slug
	// already exists. The name of the first creator is preserved on conflict.
	Upsert(ctx context.Context, name, slug string) (domain.Tag, error)

	// List returns all tags whose slug starts with prefix, ordered by slug.
	// If prefix is empty, all tags are returned.
	List(ctx context.Context, prefix string) ([]domain.Tag, error)

	// AddToStop links a tag to a stop. Idempotent — no error if already linked.
	AddToStop(ctx context.Context, stopID, tagID uuid.UUID) error

	// RemoveFromStop unlinks a tag from a stop by slug.
	// Returns domain.ErrNotFound if the tag is not linked to the stop.
	RemoveFromStop(ctx context.Context, stopID uuid.UUID, slug string) error

	// ListByStop returns all tags linked to a stop, ordered by slug.
	ListByStop(ctx context.Context, stopID uuid.UUID) ([]domain.Tag, error)
}

// pgTagRepo is the Postgres implementation of TagRepo.
type pgTagRepo struct {
	db db
}

// NewTagRepo constructs a TagRepo backed by the provided db connection.
func NewTagRepo(db db) TagRepo {
	return &pgTagRepo{db: db}
}

func (r *pgTagRepo) Upsert(ctx context.Context, name, slug string) (domain.Tag, error) {
	return domain.Tag{}, fmt.Errorf("not implemented")
}

func (r *pgTagRepo) List(ctx context.Context, prefix string) ([]domain.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *pgTagRepo) AddToStop(ctx context.Context, stopID, tagID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *pgTagRepo) RemoveFromStop(ctx context.Context, stopID uuid.UUID, slug string) error {
	return fmt.Errorf("not implemented")
}

func (r *pgTagRepo) ListByStop(ctx context.Context, stopID uuid.UUID) ([]domain.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
