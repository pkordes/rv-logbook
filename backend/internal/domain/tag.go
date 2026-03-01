package domain

import (
	"time"

	"github.com/google/uuid"
)

// Tag represents a user-defined label that can be applied to stops.
// Tags are global — not owned by any trip or stop.
// Identity is determined by Slug, which is always lowercase and hyphenated.
// Name preserves the original casing supplied by the first user to create the tag.
type Tag struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	CreatedAt time.Time
}
