package handler

import (
	"context"
	"errors"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/handler/gen"
)

// CreateTrip handles POST /trips.
func (s *Server) CreateTrip(ctx context.Context, req gen.CreateTripRequestObject) (gen.CreateTripResponseObject, error) {
	trip, err := requestToTrip(req.Body)
	if err != nil {
		return gen.CreateTrip422JSONResponse(requestBody(err.Error())), nil
	}

	created, err := s.trips.Create(ctx, trip)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			return gen.CreateTrip422JSONResponse(validationBody(err)), nil
		}
		return nil, err
	}

	return gen.CreateTrip201JSONResponse(tripToResponse(created)), nil
}

// ListTrips handles GET /trips.
// Supports ?page= and ?limit= query parameters (defaults: page=1, limit=20, max=100).
func (s *Server) ListTrips(ctx context.Context, req gen.ListTripsRequestObject) (gen.ListTripsResponseObject, error) {
	params := domain.NewPaginationParams(req.Params.Page, req.Params.Limit)
	trips, total, err := s.trips.ListPaged(ctx, params)
	if err != nil {
		return nil, err
	}

	data := make([]gen.Trip, len(trips))
	for i, t := range trips {
		data[i] = tripToResponse(t)
	}
	return gen.ListTrips200JSONResponse{
		Data: data,
		Pagination: gen.Pagination{
			Page:  params.Page,
			Limit: params.Limit,
			Total: int(total),
		},
	}, nil
}

// GetTrip handles GET /trips/{id}.
func (s *Server) GetTrip(ctx context.Context, req gen.GetTripRequestObject) (gen.GetTripResponseObject, error) {
	trip, err := s.trips.GetByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.GetTrip404JSONResponse(notFoundBody("trip not found")), nil
		}
		return nil, err
	}

	return gen.GetTrip200JSONResponse(tripToResponse(trip)), nil
}

// UpdateTrip handles PUT /trips/{id}.
func (s *Server) UpdateTrip(ctx context.Context, req gen.UpdateTripRequestObject) (gen.UpdateTripResponseObject, error) {
	trip, err := requestToTripUpdate(req.Id, req.Body)
	if err != nil {
		return gen.UpdateTrip422JSONResponse(requestBody(err.Error())), nil
	}

	updated, err := s.trips.Update(ctx, trip)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.UpdateTrip404JSONResponse(notFoundBody("trip not found")), nil
		}
		if errors.Is(err, domain.ErrValidation) {
			return gen.UpdateTrip422JSONResponse(validationBody(err)), nil
		}
		return nil, err
	}

	return gen.UpdateTrip200JSONResponse(tripToResponse(updated)), nil
}

// DeleteTrip handles DELETE /trips/{id}.
func (s *Server) DeleteTrip(ctx context.Context, req gen.DeleteTripRequestObject) (gen.DeleteTripResponseObject, error) {
	err := s.trips.Delete(ctx, req.Id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.DeleteTrip404JSONResponse(notFoundBody("trip not found")), nil
		}
		return nil, err
	}

	return gen.DeleteTrip204Response{}, nil
}

// --- mapping helpers --------------------------------------------------------

// requestToTrip converts a CreateTripRequest body into a domain.Trip.
// Returns an error if required fields are missing.
func requestToTrip(body *gen.CreateTripRequest) (domain.Trip, error) {
	if body == nil {
		return domain.Trip{}, errors.New("request body is required")
	}
	t := domain.Trip{
		Name:      body.Name,
		StartDate: body.StartDate.Time,
	}
	if body.EndDate != nil {
		ed := body.EndDate.Time
		t.EndDate = &ed
	}
	if body.Notes != nil {
		t.Notes = *body.Notes
	}
	return t, nil
}

// requestToTripUpdate builds a domain.Trip for an update, preserving the path ID.
func requestToTripUpdate(id openapi_types.UUID, body *gen.UpdateTripRequest) (domain.Trip, error) {
	if body == nil {
		return domain.Trip{}, errors.New("request body is required")
	}
	t := domain.Trip{
		ID:        id,
		Name:      body.Name,
		StartDate: body.StartDate.Time,
	}
	if body.EndDate != nil {
		ed := body.EndDate.Time
		t.EndDate = &ed
	}
	if body.Notes != nil {
		t.Notes = *body.Notes
	}
	return t, nil
}

// tripToResponse converts a domain.Trip into the generated gen.Trip type.
func tripToResponse(t domain.Trip) gen.Trip {
	resp := gen.Trip{
		Id:        t.ID,
		Name:      t.Name,
		StartDate: openapi_types.Date{Time: t.StartDate},
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
	if t.Notes != "" {
		resp.Notes = &t.Notes
	}
	if t.EndDate != nil {
		ed := openapi_types.Date{Time: *t.EndDate}
		resp.EndDate = &ed
	}
	return resp
}
