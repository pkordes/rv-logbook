package handler

import (
	"github.com/pkordes/rv-logbook/backend/internal/handler/gen"
)

// notFoundBody returns an ErrorResponse for a missing resource.
// The caller supplies the human-readable message (e.g. "trip not found")
// because the handler is the layer that knows what was being looked up.
func notFoundBody(message string) gen.ErrorResponse {
	return gen.ErrorResponse{Error: gen.ErrorDetail{Code: "not_found", Message: message}}
}

// validationBody returns an ErrorResponse for a domain validation failure.
// The message is extracted from the wrapped domain.ErrValidation error.
func validationBody(err error) gen.ErrorResponse {
	return gen.ErrorResponse{Error: gen.ErrorDetail{Code: "validation_error", Message: unwrapMessage(err)}}
}

// requestBody returns an ErrorResponse for a bad request rejected before
// reaching the service layer (e.g. missing or malformed body).
func requestBody(message string) gen.ErrorResponse {
	return gen.ErrorResponse{Error: gen.ErrorDetail{Code: "validation_error", Message: message}}
}

// unwrapMessage extracts the human-readable part from a wrapped sentinel error.
// e.g. "service.TripService.Create: validation error: name is required" → "name is required"
func unwrapMessage(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	for _, prefix := range []string{
		"service.TripService.Create: validation error: ",
		"service.TripService.Update: validation error: ",
		"service.StopService.Create: validation error: ",
		"service.StopService.Update: validation error: ",
		"service.TagService.AddTag: validation error: ",
		"validation error: ",
	} {
		if len(msg) > len(prefix) && msg[:len(prefix)] == prefix {
			return msg[len(prefix):]
		}
	}
	return msg
}
