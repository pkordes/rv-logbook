// Package domain contains the core data types for the RV Logbook application.
// This package has zero external dependencies and is imported by every other
// internal package (repo, service, handler).
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Trip represents a single RV trip from start to finish.
// A trip is the top-level aggregate; stops belong to a trip.
type Trip struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	StartDate time.Time  `json:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty"` // nil when trip is still in progress
	Notes     string     `json:"notes,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
