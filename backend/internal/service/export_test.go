package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/repo"
	"github.com/pkordes/rv-logbook/backend/internal/service"
)

// ---- mock repos for ExportService -----------------------------------------
// TripRepo and TagRepo mocks are already declared in trip_test.go and
// tag_test.go (same package). Only StopRepo is reused from stop_test.go.
// ExportService receives all three, so we wire them up below.

// newExportService is a convenience constructor for tests.
func newExportService(trips repo.TripRepo, stops repo.StopRepo, tags repo.TagRepo) *service.ExportService {
	return service.NewExportService(trips, stops, tags)
}

// ---- helpers ---------------------------------------------------------------

func tripFixtureExport(name string, start time.Time) domain.Trip {
	return domain.Trip{
		ID:        uuid.New(),
		Name:      name,
		StartDate: start,
	}
}

func stopFixtureExport(tripID uuid.UUID, name string, arrived time.Time) domain.Stop {
	return domain.Stop{
		ID:        uuid.New(),
		TripID:    tripID,
		Name:      name,
		ArrivedAt: arrived,
	}
}

// ---- Export ----------------------------------------------------------------

func TestExportService_Export_OneTrip_OneStop_NoTags(t *testing.T) {
	trip := tripFixtureExport("Summer Tour", time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	stop := stopFixtureExport(trip.ID, "Yellowstone Camp", time.Date(2025, 6, 2, 10, 0, 0, 0, time.UTC))

	svc := newExportService(
		&mockTripRepo{
			list: func(_ context.Context) ([]domain.Trip, error) {
				return []domain.Trip{trip}, nil
			},
		},
		&mockStopRepo{
			listByTripID: func(_ context.Context, _ uuid.UUID) ([]domain.Stop, error) {
				return []domain.Stop{stop}, nil
			},
		},
		&mockTagRepo{
			listByStop: func(_ context.Context, _ uuid.UUID) ([]domain.Tag, error) {
				return []domain.Tag{}, nil
			},
		},
	)

	rows, err := svc.Export(context.Background())

	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "Summer Tour", rows[0].TripName)
	assert.Equal(t, "2025-06-01", rows[0].TripStartDate)
	assert.Equal(t, "Yellowstone Camp", rows[0].StopName)
	require.NotNil(t, rows[0].ArrivedAt)
	assert.Equal(t, stop.ArrivedAt, *rows[0].ArrivedAt)
	assert.Empty(t, rows[0].Tags)
}

func TestExportService_Export_StopWithTags(t *testing.T) {
	trip := tripFixtureExport("Tour", time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	stop := stopFixtureExport(trip.ID, "Yosemite", time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC))
	tags := []domain.Tag{
		{ID: uuid.New(), Slug: "camping"},
		{ID: uuid.New(), Slug: "national-park"},
	}

	svc := newExportService(
		&mockTripRepo{
			list: func(_ context.Context) ([]domain.Trip, error) { return []domain.Trip{trip}, nil },
		},
		&mockStopRepo{
			listByTripID: func(_ context.Context, _ uuid.UUID) ([]domain.Stop, error) {
				return []domain.Stop{stop}, nil
			},
		},
		&mockTagRepo{
			listByStop: func(_ context.Context, _ uuid.UUID) ([]domain.Tag, error) {
				return tags, nil
			},
		},
	)

	rows, err := svc.Export(context.Background())

	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, []string{"camping", "national-park"}, rows[0].Tags)
}

func TestExportService_Export_TripWithNoStops(t *testing.T) {
	trip := tripFixtureExport("Empty Trip", time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))

	svc := newExportService(
		&mockTripRepo{
			list: func(_ context.Context) ([]domain.Trip, error) { return []domain.Trip{trip}, nil },
		},
		&mockStopRepo{
			listByTripID: func(_ context.Context, _ uuid.UUID) ([]domain.Stop, error) {
				return []domain.Stop{}, nil
			},
		},
		&mockTagRepo{},
	)

	rows, err := svc.Export(context.Background())

	require.NoError(t, err)
	require.Len(t, rows, 1, "trips with no stops should still produce one row")
	assert.Equal(t, "Empty Trip", rows[0].TripName)
	assert.Empty(t, rows[0].StopName)
	assert.Nil(t, rows[0].ArrivedAt)
	assert.Empty(t, rows[0].Tags)
}

func TestExportService_Export_MultipleTripsMultipleStops(t *testing.T) {
	trip1 := tripFixtureExport("Trip A", time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	trip2 := tripFixtureExport("Trip B", time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))

	stopsByTrip := map[uuid.UUID][]domain.Stop{
		trip1.ID: {
			stopFixtureExport(trip1.ID, "Stop A1", time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC)),
			stopFixtureExport(trip1.ID, "Stop A2", time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC)),
		},
		trip2.ID: {
			stopFixtureExport(trip2.ID, "Stop B1", time.Date(2025, 7, 3, 0, 0, 0, 0, time.UTC)),
		},
	}

	svc := newExportService(
		&mockTripRepo{
			list: func(_ context.Context) ([]domain.Trip, error) {
				return []domain.Trip{trip1, trip2}, nil
			},
		},
		&mockStopRepo{
			listByTripID: func(_ context.Context, tripID uuid.UUID) ([]domain.Stop, error) {
				return stopsByTrip[tripID], nil
			},
		},
		&mockTagRepo{
			listByStop: func(_ context.Context, _ uuid.UUID) ([]domain.Tag, error) {
				return []domain.Tag{}, nil
			},
		},
	)

	rows, err := svc.Export(context.Background())

	require.NoError(t, err)
	assert.Len(t, rows, 3) // 2 stops for trip1, 1 stop for trip2
	// First two rows belong to trip1
	assert.Equal(t, "Trip A", rows[0].TripName)
	assert.Equal(t, "Trip A", rows[1].TripName)
	assert.Equal(t, "Stop A1", rows[0].StopName)
	assert.Equal(t, "Stop A2", rows[1].StopName)
	// Third row belongs to trip2
	assert.Equal(t, "Trip B", rows[2].TripName)
	assert.Equal(t, "Stop B1", rows[2].StopName)
}

func TestExportService_Export_NoTrips(t *testing.T) {
	svc := newExportService(
		&mockTripRepo{
			list: func(_ context.Context) ([]domain.Trip, error) { return []domain.Trip{}, nil },
		},
		&mockStopRepo{},
		&mockTagRepo{},
	)

	rows, err := svc.Export(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, rows)
	assert.Empty(t, rows)
}

func TestExportService_Export_TripRepoError(t *testing.T) {
	svc := newExportService(
		&mockTripRepo{
			list: func(_ context.Context) ([]domain.Trip, error) {
				return nil, domain.ErrNotFound
			},
		},
		&mockStopRepo{},
		&mockTagRepo{},
	)

	_, err := svc.Export(context.Background())

	assert.Error(t, err)
}

func TestExportService_Export_TripEndDateIncluded(t *testing.T) {
	endDate := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	trip := domain.Trip{
		ID:        uuid.New(),
		Name:      "Dated Trip",
		StartDate: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   &endDate,
	}

	svc := newExportService(
		&mockTripRepo{
			list: func(_ context.Context) ([]domain.Trip, error) { return []domain.Trip{trip}, nil },
		},
		&mockStopRepo{
			listByTripID: func(_ context.Context, _ uuid.UUID) ([]domain.Stop, error) {
				return []domain.Stop{}, nil
			},
		},
		&mockTagRepo{},
	)

	rows, err := svc.Export(context.Background())

	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "2025-06-01", rows[0].TripStartDate)
	assert.Equal(t, "2025-06-15", rows[0].TripEndDate)
}
