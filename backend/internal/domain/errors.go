package domain

import "errors"

// ErrNotFound is returned by repo and service functions when the requested
// resource does not exist in the database.
// Handlers should map this to HTTP 404.
var ErrNotFound = errors.New("not found")

// ErrValidation is returned by service functions when input fails business
// rule validation (e.g. missing required field, end date before start date).
// Handlers should map this to HTTP 422 Unprocessable Entity.
var ErrValidation = errors.New("validation error")
