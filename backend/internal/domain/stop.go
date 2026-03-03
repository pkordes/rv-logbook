package domain

import (
	"time"

	"github.com/google/uuid"
)

// Stop represents a single location visited during a trip.
// DepartedAt is nil when the traveller is still at this stop.
// Tags is populated when the stop is fetched from the repository;
// it is always an initialised (non-nil) slice.
type Stop struct {
	ID         uuid.UUID
	TripID     uuid.UUID
	Name       string
	Location   string
	ArrivedAt  time.Time
	DepartedAt *time.Time
	Notes      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Tags       []Tag
}
