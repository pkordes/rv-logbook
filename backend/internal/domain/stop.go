package domain

import (
	"time"

	"github.com/google/uuid"
)

// Stop represents a single location visited during a trip.
// DepartedAt is nil when the traveller is still at this stop.
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
}
