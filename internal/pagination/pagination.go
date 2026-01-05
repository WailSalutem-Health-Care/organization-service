package pagination

import (
	"net/http"
	"strconv"
)

// Default pagination values
const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// Params represents pagination query parameters
type Params struct {
	Page  int `json:"page"`  // Current page number (1-based)
	Limit int `json:"limit"` // Number of items per page
}

// Meta contains pagination metadata for responses
type Meta struct {
	CurrentPage  int  `json:"current_page"`
	PerPage      int  `json:"per_page"`
	TotalPages   int  `json:"total_pages"`
	TotalRecords int  `json:"total_records"`
	HasNext      bool `json:"has_next"`
	HasPrevious  bool `json:"has_previous"`
}

// ParseParams extracts and validates pagination parameters from HTTP request
func ParseParams(r *http.Request) Params {
	page := DefaultPage
	limit := DefaultLimit

	// Parse page parameter
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Parse limit parameter
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			// Enforce maximum limit
			if limit > MaxLimit {
				limit = MaxLimit
			}
		}
	}

	return Params{
		Page:  page,
		Limit: limit,
	}
}

// Validate ensures pagination parameters are valid and sets defaults if needed
func (p *Params) Validate() {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.Limit < 1 {
		p.Limit = DefaultLimit
	}
	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}
}

// CalculateOffset returns the SQL OFFSET value based on page and limit
func (p *Params) CalculateOffset() int {
	return (p.Page - 1) * p.Limit
}

// CalculateMeta creates pagination metadata based on total records
func (p *Params) CalculateMeta(totalRecords int) Meta {
	totalPages := (totalRecords + p.Limit - 1) / p.Limit // Ceiling division
	if totalPages < 1 {
		totalPages = 1
	}

	return Meta{
		CurrentPage:  p.Page,
		PerPage:      p.Limit,
		TotalPages:   totalPages,
		TotalRecords: totalRecords,
		HasNext:      p.Page < totalPages,
		HasPrevious:  p.Page > 1,
	}
}

