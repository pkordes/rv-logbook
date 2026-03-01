package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/repo"
)

// StopService implements business logic for Stop operations.
// It holds trips, stops, and tags repos because creating a stop requires
// verifying the parent trip exists, and tag operations are scoped to a stop.
type StopService struct {
	trips repo.TripRepo
	stops repo.StopRepo
	tags  repo.TagRepo
}

// NewStopService constructs a StopService backed by the provided repos.
func NewStopService(trips repo.TripRepo, stops repo.StopRepo, tags repo.TagRepo) *StopService {
	return &StopService{trips: trips, stops: stops, tags: tags}
}

// Create validates the stop, verifies the parent trip exists, then persists.
// Returns domain.ErrValidation if input violates business rules.
// Returns domain.ErrNotFound if the parent trip does not exist.
func (s *StopService) Create(ctx context.Context, stop domain.Stop) (domain.Stop, error) {
	if _, err := s.trips.GetByID(ctx, stop.TripID); err != nil {
		return domain.Stop{}, fmt.Errorf("service.StopService.Create: %w", err)
	}
	if err := validateStop(stop); err != nil {
		return domain.Stop{}, err
	}
	result, err := s.stops.Create(ctx, stop)
	if err != nil {
		return domain.Stop{}, fmt.Errorf("service.StopService.Create: %w", err)
	}
	return result, nil
}

// GetByID returns a single stop by ID, scoped to the given tripID.
// Returns domain.ErrNotFound if no stop with that ID exists under that trip.
func (s *StopService) GetByID(ctx context.Context, tripID, stopID uuid.UUID) (domain.Stop, error) {
	result, err := s.stops.GetByID(ctx, tripID, stopID)
	if err != nil {
		return domain.Stop{}, fmt.Errorf("service.StopService.GetByID: %w", err)
	}
	return result, nil
}

// ListByTripID returns all stops for a trip ordered by arrived_at ascending.
// Always returns a non-nil slice so callers can safely range over it.
func (s *StopService) ListByTripID(ctx context.Context, tripID uuid.UUID) ([]domain.Stop, error) {
	stops, err := s.stops.ListByTripID(ctx, tripID)
	if err != nil {
		return nil, fmt.Errorf("service.StopService.ListByTripID: %w", err)
	}
	if stops == nil {
		return []domain.Stop{}, nil
	}
	return stops, nil
}

// ListByTripIDPaged returns one page of stops for a trip and the total count.
// The caller controls page and limit via domain.PaginationParams.
func (s *StopService) ListByTripIDPaged(ctx context.Context, tripID uuid.UUID, p domain.PaginationParams) ([]domain.Stop, int64, error) {
	stops, total, err := s.stops.ListByTripIDPaged(ctx, tripID, p)
	if err != nil {
		return nil, 0, fmt.Errorf("service.StopService.ListByTripIDPaged: %w", err)
	}
	if stops == nil {
		stops = []domain.Stop{}
	}
	return stops, total, nil
}

// Update validates and persists changes to an existing stop.
// Returns domain.ErrValidation for invalid input, domain.ErrNotFound if the
// stop does not exist under the given trip.
func (s *StopService) Update(ctx context.Context, stop domain.Stop) (domain.Stop, error) {
	if err := validateStop(stop); err != nil {
		return domain.Stop{}, err
	}
	result, err := s.stops.Update(ctx, stop)
	if err != nil {
		return domain.Stop{}, fmt.Errorf("service.StopService.Update: %w", err)
	}
	return result, nil
}

// Delete removes a stop by ID, scoped to the given tripID.
// Returns domain.ErrNotFound if the stop does not exist under the given trip.
func (s *StopService) Delete(ctx context.Context, tripID, stopID uuid.UUID) error {
	if err := s.stops.Delete(ctx, tripID, stopID); err != nil {
		return fmt.Errorf("service.StopService.Delete: %w", err)
	}
	return nil
}

// AddTag upserts a tag by name and links it to the given stop.
// The name is normalized to a slug using the same rules as TagService.
// Returns domain.ErrValidation if tagName is empty or normalizes to empty.
func (s *StopService) AddTag(ctx context.Context, stopID uuid.UUID, tagName string) (domain.Tag, error) {
	tagName = strings.TrimSpace(tagName)
	if tagName == "" {
		return domain.Tag{}, fmt.Errorf("%w: tag name is required", domain.ErrValidation)
	}
	slug := toSlug(tagName)
	if slug == "" {
		return domain.Tag{}, fmt.Errorf("%w: tag name contains no usable characters", domain.ErrValidation)
	}

	tag, err := s.tags.Upsert(ctx, tagName, slug)
	if err != nil {
		return domain.Tag{}, fmt.Errorf("service.StopService.AddTag: %w", err)
	}
	if err := s.tags.AddToStop(ctx, stopID, tag.ID); err != nil {
		return domain.Tag{}, fmt.Errorf("service.StopService.AddTag: %w", err)
	}
	return tag, nil
}

// RemoveTagFromStop unlinks a tag from a stop by slug.
// Returns domain.ErrNotFound if the tag is not linked to the stop.
func (s *StopService) RemoveTagFromStop(ctx context.Context, stopID uuid.UUID, slug string) error {
	if err := s.tags.RemoveFromStop(ctx, stopID, slug); err != nil {
		return fmt.Errorf("service.StopService.RemoveTagFromStop: %w", err)
	}
	return nil
}

// ListTagsByStop returns all tags linked to a stop, ordered by slug.
// Always returns a non-nil slice so callers can safely range over it.
func (s *StopService) ListTagsByStop(ctx context.Context, stopID uuid.UUID) ([]domain.Tag, error) {
	tags, err := s.tags.ListByStop(ctx, stopID)
	if err != nil {
		return nil, fmt.Errorf("service.StopService.ListTagsByStop: %w", err)
	}
	if tags == nil {
		return []domain.Tag{}, nil
	}
	return tags, nil
}

// validateStop enforces business rules common to both Create and Update.
//   - Name must be non-empty (whitespace-only names are rejected).
//   - DepartedAt, if set, must not be before ArrivedAt.
func validateStop(stop domain.Stop) error {
	if strings.TrimSpace(stop.Name) == "" {
		return fmt.Errorf("%w: name is required", domain.ErrValidation)
	}
	if stop.DepartedAt != nil && stop.DepartedAt.Before(stop.ArrivedAt) {
		return fmt.Errorf("%w: departed_at must not be before arrived_at", domain.ErrValidation)
	}
	return nil
}
