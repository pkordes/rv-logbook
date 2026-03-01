package domain

// PaginationParams carries page/limit values from the HTTP layer to the repo layer.
// Page is 1-indexed. Limit is capped at 100 by NewPaginationParams.
type PaginationParams struct {
	// Page is the current page number, starting at 1.
	Page int
	// Limit is the maximum number of items to return.
	Limit int
}

// NewPaginationParams builds a PaginationParams from optional HTTP query params.
// Nil pointers fall back to sane defaults (page=1, limit=20).
// The limit is capped at 100 to prevent runaway queries.
func NewPaginationParams(page, limit *int) PaginationParams {
	p := PaginationParams{Page: 1, Limit: 20}
	if page != nil && *page >= 1 {
		p.Page = *page
	}
	if limit != nil && *limit >= 1 {
		p.Limit = *limit
		if p.Limit > 100 {
			p.Limit = 100
		}
	}
	return p
}

// Offset returns the zero-based row offset for a SQL OFFSET clause.
func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}
