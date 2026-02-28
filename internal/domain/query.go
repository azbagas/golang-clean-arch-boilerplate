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

// NewPaginationParams creates a PaginationParams with validated/clamped values.
func NewPaginationParams(page, perPage int) PaginationParams {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}
	return PaginationParams{Page: page, PerPage: perPage}
}

// SortParams holds the sorting query parameters.
type SortParams struct {
	SortBy    string
	SortOrder string // "asc" or "desc"
}

// NewSortParams creates a SortParams with validated values.
// It returns ErrInvalidSortField if sortBy is not in the allowedFields.
func NewSortParams(sortBy, sortOrder string, allowedFields map[string]bool) (SortParams, error) {
	if sortBy != "" && !allowedFields[sortBy] {
		return SortParams{}, ErrInvalidSortField
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}
	return SortParams{SortBy: sortBy, SortOrder: sortOrder}, nil
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
