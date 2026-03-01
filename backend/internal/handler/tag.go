package handler

import (
	"context"
	"fmt"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/pkordes/rv-logbook/backend/internal/domain"
	"github.com/pkordes/rv-logbook/backend/internal/handler/gen"
)

// ListTags handles GET /tags.
// The optional ?q= query parameter filters tags by slug prefix.
func (s *Server) ListTags(ctx context.Context, req gen.ListTagsRequestObject) (gen.ListTagsResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// ListTagsByStop handles GET /trips/{tripId}/stops/{stopId}/tags.
func (s *Server) ListTagsByStop(ctx context.Context, req gen.ListTagsByStopRequestObject) (gen.ListTagsByStopResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// AddTagToStop handles POST /trips/{tripId}/stops/{stopId}/tags.
func (s *Server) AddTagToStop(ctx context.Context, req gen.AddTagToStopRequestObject) (gen.AddTagToStopResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// RemoveTagFromStop handles DELETE /trips/{tripId}/stops/{stopId}/tags/{slug}.
func (s *Server) RemoveTagFromStop(ctx context.Context, req gen.RemoveTagFromStopRequestObject) (gen.RemoveTagFromStopResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
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
