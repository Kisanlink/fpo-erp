package utils

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Default pagination values
const (
	DefaultLimit = 50
	MaxLimit     = 200
	MinLimit     = 1
	MinOffset    = 0
)

// PaginationParams represents pagination query parameters
type PaginationParams struct {
	Limit  int `form:"limit" binding:"omitempty,min=1,max=200"`
	Offset int `form:"offset" binding:"omitempty,min=0"`
}

// PaginationMeta represents pagination metadata in response
type PaginationMeta struct {
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
	Pages  int   `json:"pages"`
	Total  int64 `json:"total"`
}

// SetDefaults sets default values for pagination if not provided
func (p *PaginationParams) SetDefaults() {
	if p.Limit == 0 {
		p.Limit = DefaultLimit
	}
	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}
	if p.Limit < MinLimit {
		p.Limit = MinLimit
	}
	if p.Offset < MinOffset {
		p.Offset = MinOffset
	}
}

// Validate validates the pagination parameters
func (p *PaginationParams) Validate() error {
	if p.Limit < MinLimit || p.Limit > MaxLimit {
		p.Limit = DefaultLimit
	}
	if p.Offset < MinOffset {
		p.Offset = MinOffset
	}
	return nil
}

// NewPaginationMeta creates pagination metadata from params and total count
func NewPaginationMeta(limit, offset int, total int64) PaginationMeta {
	pages := 0
	if total > 0 && limit > 0 {
		pages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return PaginationMeta{
		Limit:  limit,
		Offset: offset,
		Pages:  pages,
		Total:  total,
	}
}

// GetPaginationParams extracts and validates pagination parameters from gin context
func GetPaginationParams(c *gin.Context) PaginationParams {
	limitStr := c.DefaultQuery("limit", strconv.Itoa(DefaultLimit))
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < MinLimit {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < MinOffset {
		offset = MinOffset
	}

	return PaginationParams{
		Limit:  limit,
		Offset: offset,
	}
}

// BindPaginationParams binds and validates pagination from query params
func BindPaginationParams(c *gin.Context) (PaginationParams, error) {
	var params PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		// If binding fails, use defaults
		params = PaginationParams{
			Limit:  DefaultLimit,
			Offset: MinOffset,
		}
	}
	params.SetDefaults()
	return params, nil
}
