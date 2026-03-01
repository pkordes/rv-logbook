package domain

import "time"

// ExportRow is a single row in the full-data export.
// It is a flat, denormalized view: one row per stop, with trip fields repeated
// for every stop on that trip. Trips with no stops yield one row with zero
// values for all stop fields.
//
// Tags is a slice of slugs for the stop, ordered alphabetically.
// Callers that need a joined string (e.g. CSV) should join with ",".
type ExportRow struct {
	// Trip fields — repeated for every stop on the trip.
	TripID        string
	TripName      string
	TripStartDate string // "2006-01-02" formatted date
	TripEndDate   string // empty string when nil

	// Stop fields — zero values when the trip has no stops.
	StopName     string
	StopLocation string
	ArrivedAt    *time.Time
	DepartedAt   *time.Time
	StopNotes    string

	// Tags — slugs of all tags attached to this stop.
	Tags []string
}
