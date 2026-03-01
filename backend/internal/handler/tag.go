package handler

import (
	"context"
	"errors"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/handler/gen"
)

// ListTags handles GET /tags.
// The optional ?q= query parameter filters tags by slug prefix.
// Stub: returns tags but leaves Pagination metadata at zero-values.
// Real pagination wired in the Green commit of step 7.4.
func (s *Server) ListTags(ctx context.Context, req gen.ListTagsRequestObject) (gen.ListTagsResponseObject, error) {
	prefix := derefString(req.Params.Q)
	params := domain.NewPaginationParams(req.Params.Page, req.Params.Limit)

	tags, _, err := s.tags.ListPaged(ctx, prefix, params) // total ignored in stub
	if err != nil {
		return nil, err
	}

	data := make([]gen.Tag, len(tags))
	for i, t := range tags {
		data[i] = tagToResponse(t)
	}
	// Stub: Pagination not populated — tests will fail on Pagination.Total.
	return gen.ListTags200JSONResponse{Data: data}, nil
}

// ListTagsByStop handles GET /trips/{tripId}/stops/{stopId}/tags.
func (s *Server) ListTagsByStop(ctx context.Context, req gen.ListTagsByStopRequestObject) (gen.ListTagsByStopResponseObject, error) {
	tags, err := s.stops.ListTagsByStop(ctx, req.StopId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.ListTagsByStop404JSONResponse(notFoundBody("stop not found")), nil
		}
		return nil, err
	}

	resp := make([]gen.Tag, len(tags))
	for i, t := range tags {
		resp[i] = tagToResponse(t)
	}
	return gen.ListTagsByStop200JSONResponse(resp), nil
}

// AddTagToStop handles POST /trips/{tripId}/stops/{stopId}/tags.
func (s *Server) AddTagToStop(ctx context.Context, req gen.AddTagToStopRequestObject) (gen.AddTagToStopResponseObject, error) {
	tag, err := s.stops.AddTag(ctx, req.StopId, req.Body.Name)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.AddTagToStop404JSONResponse(notFoundBody("stop not found")), nil
		}
		if errors.Is(err, domain.ErrValidation) {
			return gen.AddTagToStop422JSONResponse(validationBody(err)), nil
		}
		return nil, err
	}

	return gen.AddTagToStop201JSONResponse(tagToResponse(tag)), nil
}

// RemoveTagFromStop handles DELETE /trips/{tripId}/stops/{stopId}/tags/{slug}.
func (s *Server) RemoveTagFromStop(ctx context.Context, req gen.RemoveTagFromStopRequestObject) (gen.RemoveTagFromStopResponseObject, error) {
	err := s.stops.RemoveTagFromStop(ctx, req.StopId, req.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.RemoveTagFromStop404JSONResponse(notFoundBody("tag not linked to stop")), nil
		}
		return nil, err
	}

	return gen.RemoveTagFromStop204Response{}, nil
}

// tagToResponse converts a domain.Tag to the generated API response type.
func tagToResponse(t domain.Tag) gen.Tag {
	return gen.Tag{
		Id:        openapi_types.UUID(t.ID),
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt,
	}
}
