package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/repo"
)

// nonAlphanumeric matches any run of characters that are not a-z or 0-9.
// Used to replace punctuation, spaces, and special characters with a hyphen.
var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// TagService implements business logic for Tag operations.
// Its primary responsibility is slug normalization: all tag identity is
// determined by slug, which is always lowercase and hyphenated.
type TagService struct {
	tags repo.TagRepo
}

// NewTagService constructs a TagService backed by the provided TagRepo.
func NewTagService(tags repo.TagRepo) *TagService {
	return &TagService{tags: tags}
}

// UpsertByName normalizes name to a slug and upserts the tag.
// Returns domain.ErrValidation if the name is empty or normalizes to empty.
func (s *TagService) UpsertByName(ctx context.Context, name string) (domain.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Tag{}, fmt.Errorf("%w: tag name is required", domain.ErrValidation)
	}

	slug := toSlug(name)
	if slug == "" {
		return domain.Tag{}, fmt.Errorf("%w: tag name contains no usable characters", domain.ErrValidation)
	}

	result, err := s.tags.Upsert(ctx, name, slug)
	if err != nil {
		return domain.Tag{}, fmt.Errorf("service.TagService.UpsertByName: %w", err)
	}
	return result, nil
}

// List returns all tags whose slug starts with prefix.
// The prefix is normalized to lowercase before querying.
// Pass prefix="" to return all tags.
func (s *TagService) List(ctx context.Context, prefix string) ([]domain.Tag, error) {
	// Normalize prefix so "Mount" and "mount" return the same results.
	prefix = strings.ToLower(strings.TrimSpace(prefix))

	tags, err := s.tags.List(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("service.TagService.List: %w", err)
	}
	if tags == nil {
		return []domain.Tag{}, nil
	}
	return tags, nil
}

// ListPaged returns one page of tags whose slug starts with prefix and the total count.
// The prefix is normalized to lowercase before querying, just like List.
func (s *TagService) ListPaged(ctx context.Context, prefix string, p domain.PaginationParams) ([]domain.Tag, int64, error) {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	tags, total, err := s.tags.ListPaged(ctx, prefix, p)
	if err != nil {
		return nil, 0, fmt.Errorf("service.TagService.ListPaged: %w", err)
	}
	if tags == nil {
		tags = []domain.Tag{}
	}
	return tags, total, nil
}

// toSlug converts a display name to a URL-safe, lowercase, hyphenated slug.
// Examples:
//
//	"Rocky Mountains"  → "rocky-mountains"
//	"WALMART"          → "walmart"
//	"Rocky  Mountains!" → "rocky-mountains"
func toSlug(name string) string {
	lower := strings.ToLower(name)
	slug := nonAlphanumeric.ReplaceAllString(lower, "-")
	slug = strings.Trim(slug, "-")
	return slug
}
