// pkg/pagination/pagination.go
package pagination

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

// Params holds pagination parameters
type Params struct {
	Page    int
	PerPage int
	Offset  int
}

// Meta contains pagination metadata for responses
type Meta struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// NewParams extracts pagination parameters from the request
func NewParams(c echo.Context) Params {
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(c.QueryParam("per_page"))
	if err != nil || perPage < 1 || perPage > 100 {
		perPage = 20 // Default limit
	}

	offset := (page - 1) * perPage

	return Params{
		Page:    page,
		PerPage: perPage,
		Offset:  offset,
	}
}

// NewMeta creates pagination metadata based on results
func NewMeta(params Params, total int) Meta {
	totalPages := (total + params.PerPage - 1) / params.PerPage

	return Meta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    params.Page < totalPages,
		HasPrev:    params.Page > 1,
	}
}

// Response creates a standardized paginated response
func Response(data interface{}, meta Meta) map[string]interface{} {
	return map[string]interface{}{
		"data": data,
		"meta": meta,
	}
}
