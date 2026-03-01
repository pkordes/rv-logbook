package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

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

	// ListPaged returns one page of tags matching the slug prefix and the total count.
	// If prefix is empty, all tags are included in the result set.
	ListPaged(ctx context.Context, prefix string, p domain.PaginationParams) ([]domain.Tag, int64, error)

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

// Upsert inserts a tag or returns the existing row on slug conflict.
// The DO UPDATE SET trick forces the RETURNING clause to fire even when
// the conflict handler skips the insert — without it, RETURNING returns
// nothing on DO NOTHING conflicts.
func (r *pgTagRepo) Upsert(ctx context.Context, name, slug string) (domain.Tag, error) {
	const q = `
		INSERT INTO tags (name, slug)
		VALUES (@name, @slug)
		ON CONFLICT (slug) DO UPDATE SET slug = EXCLUDED.slug
		RETURNING id, name, slug, created_at`

	row := r.db.QueryRow(ctx, q, pgx.NamedArgs{"name": name, "slug": slug})
	result, err := scanTag(row)
	if err != nil {
		return domain.Tag{}, fmt.Errorf("repo.TagRepo.Upsert: %w", err)
	}
	return result, nil
}

// List returns all tags whose slug starts with prefix, ordered by slug.
// Pass prefix="" to return all tags.
func (r *pgTagRepo) List(ctx context.Context, prefix string) ([]domain.Tag, error) {
	const q = `
		SELECT id, name, slug, created_at
		FROM tags
		WHERE slug LIKE @prefix || '%'
		ORDER BY slug`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{"prefix": prefix})
	if err != nil {
		return nil, fmt.Errorf("repo.TagRepo.List: %w", err)
	}
	defer rows.Close()

	tags := []domain.Tag{}
	for rows.Next() {
		tag, err := scanTag(rows)
		if err != nil {
			return nil, fmt.Errorf("repo.TagRepo.List: scan: %w", err)
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.TagRepo.List: rows: %w", err)
	}
	return tags, nil
}

// ListPaged returns one page of tags matching prefix ordered by slug.
// Stub: returns empty results. Real implementation added in step 7.4 (Green).
func (r *pgTagRepo) ListPaged(ctx context.Context, prefix string, p domain.PaginationParams) ([]domain.Tag, int64, error) {
	_, _ = prefix, p
	return nil, 0, nil
}

// AddToStop links a tag to a stop. Idempotent via ON CONFLICT DO NOTHING.
func (r *pgTagRepo) AddToStop(ctx context.Context, stopID, tagID uuid.UUID) error {
	const q = `
		INSERT INTO stop_tags (stop_id, tag_id)
		VALUES (@stop_id, @tag_id)
		ON CONFLICT (stop_id, tag_id) DO NOTHING`

	_, err := r.db.Exec(ctx, q, pgx.NamedArgs{"stop_id": stopID, "tag_id": tagID})
	if err != nil {
		return fmt.Errorf("repo.TagRepo.AddToStop: %w", err)
	}
	return nil
}

// RemoveFromStop unlinks a tag from a stop using a slug-based subquery lookup.
func (r *pgTagRepo) RemoveFromStop(ctx context.Context, stopID uuid.UUID, slug string) error {
	const q = `
		DELETE FROM stop_tags
		WHERE stop_id = @stop_id
		  AND tag_id = (SELECT id FROM tags WHERE slug = @slug)`

	tag, err := r.db.Exec(ctx, q, pgx.NamedArgs{"stop_id": stopID, "slug": slug})
	if err != nil {
		return fmt.Errorf("repo.TagRepo.RemoveFromStop: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("repo.TagRepo.RemoveFromStop: %w", domain.ErrNotFound)
	}
	return nil
}

// ListByStop returns all tags linked to a stop, ordered by slug.
func (r *pgTagRepo) ListByStop(ctx context.Context, stopID uuid.UUID) ([]domain.Tag, error) {
	const q = `
		SELECT t.id, t.name, t.slug, t.created_at
		FROM tags t
		JOIN stop_tags st ON st.tag_id = t.id
		WHERE st.stop_id = @stop_id
		ORDER BY t.slug`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{"stop_id": stopID})
	if err != nil {
		return nil, fmt.Errorf("repo.TagRepo.ListByStop: %w", err)
	}
	defer rows.Close()

	tags := []domain.Tag{}
	for rows.Next() {
		tag, err := scanTag(rows)
		if err != nil {
			return nil, fmt.Errorf("repo.TagRepo.ListByStop: scan: %w", err)
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.TagRepo.ListByStop: rows: %w", err)
	}
	return tags, nil
}

// scanTag maps a single database row into a domain.Tag.
func scanTag(s scanner) (domain.Tag, error) {
	var (
		t  domain.Tag
		id pgtype.UUID
	)
	err := s.Scan(&id, &t.Name, &t.Slug, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Tag{}, domain.ErrNotFound
		}
		return domain.Tag{}, err
	}
	t.ID = uuid.UUID(id.Bytes)
	return t, nil
}
