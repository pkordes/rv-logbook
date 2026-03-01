package handler

import (
	"context"
	"errors"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/handler/gen"
)

// CreateStop handles POST /trips/{tripId}/stops.
func (s *Server) CreateStop(ctx context.Context, req gen.CreateStopRequestObject) (gen.CreateStopResponseObject, error) {
	stop := domain.Stop{
		TripID:     req.TripId,
		Name:       req.Body.Name,
		Location:   derefString(req.Body.Location),
		ArrivedAt:  req.Body.ArrivedAt,
		DepartedAt: req.Body.DepartedAt,
		Notes:      derefString(req.Body.Notes),
	}

	created, err := s.stops.Create(ctx, stop)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.CreateStop404JSONResponse(notFoundBody("trip not found")), nil
		}
		if errors.Is(err, domain.ErrValidation) {
			return gen.CreateStop422JSONResponse(validationBody(err)), nil
		}
		return nil, err
	}

	return gen.CreateStop201JSONResponse(stopToResponse(created)), nil
}

// ListStops handles GET /trips/{tripId}/stops.
// Supports ?page= and ?limit= query parameters (defaults: page=1, limit=20, max=100).
func (s *Server) ListStops(ctx context.Context, req gen.ListStopsRequestObject) (gen.ListStopsResponseObject, error) {
	params := domain.NewPaginationParams(req.Params.Page, req.Params.Limit)
	stops, total, err := s.stops.ListByTripIDPaged(ctx, req.TripId, params)
	if err != nil {
		return nil, err
	}

	data := make([]gen.Stop, len(stops))
	for i, st := range stops {
		data[i] = stopToResponse(st)
	}
	return gen.ListStops200JSONResponse{
		Data: data,
		Pagination: gen.Pagination{
			Page:  params.Page,
			Limit: params.Limit,
			Total: int(total),
		},
	}, nil
}

// GetStop handles GET /trips/{tripId}/stops/{stopId}.
func (s *Server) GetStop(ctx context.Context, req gen.GetStopRequestObject) (gen.GetStopResponseObject, error) {
	stop, err := s.stops.GetByID(ctx, req.TripId, req.StopId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.GetStop404JSONResponse(notFoundBody("stop not found")), nil
		}
		return nil, err
	}

	return gen.GetStop200JSONResponse(stopToResponse(stop)), nil
}

// UpdateStop handles PUT /trips/{tripId}/stops/{stopId}.
func (s *Server) UpdateStop(ctx context.Context, req gen.UpdateStopRequestObject) (gen.UpdateStopResponseObject, error) {
	stop := domain.Stop{
		ID:         req.StopId,
		TripID:     req.TripId,
		Name:       req.Body.Name,
		Location:   derefString(req.Body.Location),
		ArrivedAt:  req.Body.ArrivedAt,
		DepartedAt: req.Body.DepartedAt,
		Notes:      derefString(req.Body.Notes),
	}

	updated, err := s.stops.Update(ctx, stop)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.UpdateStop404JSONResponse(notFoundBody("stop not found")), nil
		}
		if errors.Is(err, domain.ErrValidation) {
			return gen.UpdateStop422JSONResponse(validationBody(err)), nil
		}
		return nil, err
	}

	return gen.UpdateStop200JSONResponse(stopToResponse(updated)), nil
}

// DeleteStop handles DELETE /trips/{tripId}/stops/{stopId}.
func (s *Server) DeleteStop(ctx context.Context, req gen.DeleteStopRequestObject) (gen.DeleteStopResponseObject, error) {
	err := s.stops.Delete(ctx, req.TripId, req.StopId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.DeleteStop404JSONResponse(notFoundBody("stop not found")), nil
		}
		return nil, err
	}

	return gen.DeleteStop204Response{}, nil
}

// stopToResponse converts a domain.Stop to the generated API response type.
// Empty strings become nil pointers for optional JSON fields (location, notes)
// so they are omitted from the response rather than sent as empty strings.
func stopToResponse(s domain.Stop) gen.Stop {
	return gen.Stop{
		Id:         openapi_types.UUID(s.ID),
		TripId:     openapi_types.UUID(s.TripID),
		Name:       s.Name,
		Location:   nilIfEmpty(s.Location),
		ArrivedAt:  s.ArrivedAt,
		DepartedAt: s.DepartedAt,
		Notes:      nilIfEmpty(s.Notes),
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

// derefString safely dereferences a *string, returning "" when nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// nilIfEmpty converts an empty string to a nil pointer.
// Used when mapping domain strings to optional API response fields.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
