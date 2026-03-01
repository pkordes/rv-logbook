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
// Rows are ordered by trip (in list order) then by stop (in arrived_at order).
func (s *ExportService) Export(ctx context.Context) ([]domain.ExportRow, error) {
	trips, err := s.trips.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("service.ExportService.Export: %w", err)
	}

	rows := make([]domain.ExportRow, 0, len(trips))

	for _, trip := range trips {
		endDate := ""
		if trip.EndDate != nil {
			endDate = trip.EndDate.Format("2006-01-02")
		}

		tripBase := domain.ExportRow{
			TripID:        trip.ID.String(),
			TripName:      trip.Name,
			TripStartDate: trip.StartDate.Format("2006-01-02"),
			TripEndDate:   endDate,
		}

		stops, err := s.stops.ListByTripID(ctx, trip.ID)
		if err != nil {
			return nil, fmt.Errorf("service.ExportService.Export: %w", err)
		}

		if len(stops) == 0 {
			// Trip exists but has no stops — emit one row with empty stop fields.
			rows = append(rows, tripBase)
			continue
		}

		for _, stop := range stops {
			tags, err := s.tags.ListByStop(ctx, stop.ID)
			if err != nil {
				return nil, fmt.Errorf("service.ExportService.Export: %w", err)
			}

			slugs := make([]string, len(tags))
			for i, t := range tags {
				slugs[i] = t.Slug
			}

			row := tripBase // copy trip fields
			row.StopName = stop.Name
			row.StopLocation = stop.Location
			arrivedAt := stop.ArrivedAt // copy to local so the pointer stays valid after the iteration
			row.ArrivedAt = &arrivedAt
			row.DepartedAt = stop.DepartedAt
			row.StopNotes = stop.Notes
			row.Tags = slugs

			rows = append(rows, row)
		}
	}

	return rows, nil
}
