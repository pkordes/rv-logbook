package service

import (
	"context"
	"fmt"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/repo"
)

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

func (s *TagService) UpsertByName(ctx context.Context, name string) (domain.Tag, error) {
	return domain.Tag{}, fmt.Errorf("not implemented")
}

func (s *TagService) List(ctx context.Context, prefix string) ([]domain.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
