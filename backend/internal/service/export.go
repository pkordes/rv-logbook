package service

import (
	"context"
	"fmt"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/repo"
)

// ExportService assembles a full flat export of all trips, stops, and tags.
type ExportService struct {
	trips repo.TripRepo
	stops repo.StopRepo
	tags  repo.TagRepo
}

// NewExportService constructs an ExportService backed by the provided repos.
func NewExportService(trips repo.TripRepo, stops repo.StopRepo, tags repo.TagRepo) *ExportService {
	return &ExportService{trips: trips, stops: stops, tags: tags}
}

// Export returns one ExportRow per stop across all trips.
// Trips with no stops contribute one row with empty stop fields.
func (s *ExportService) Export(ctx context.Context) ([]domain.ExportRow, error) {
	return nil, fmt.Errorf("not implemented")
}
