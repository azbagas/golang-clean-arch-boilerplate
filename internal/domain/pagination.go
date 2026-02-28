package domain

import "math"

// PaginationParams holds the pagination query parameters.
type PaginationParams struct {
	Page    int
	PerPage int
}

// GetOffset calculates the SQL offset from the page and page size.
func (p PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PerPage
}

// PaginationMeta holds the pagination metadata.
type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// PaginatedResult holds paginated data with metadata.
type PaginatedResult struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// NewPaginatedResult creates a PaginatedResult with calculated metadata.
func NewPaginatedResult(data interface{}, totalItems int64, params PaginationParams) *PaginatedResult {
	totalPages := int(math.Ceil(float64(totalItems) / float64(params.PerPage)))
	return &PaginatedResult{
		Data: data,
		Pagination: PaginationMeta{
			Page:       params.Page,
			PerPage:    params.PerPage,
			TotalItems: totalItems,
			TotalPages: totalPages,
			HasNext:    params.Page < totalPages,
			HasPrev:    params.Page > 1,
		},
	}
}
