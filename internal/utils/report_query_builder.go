package utils

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// ReportQueryBuilder provides a fluent interface for building report queries
type ReportQueryBuilder struct {
	query *gorm.DB
}

// NewReportQueryBuilder creates a new query builder instance
func NewReportQueryBuilder(db *gorm.DB) *ReportQueryBuilder {
	return &ReportQueryBuilder{query: db}
}

// ApplyDateFilter applies date range filtering to a query
func (b *ReportQueryBuilder) ApplyDateFilter(field string, start, end *time.Time) *ReportQueryBuilder {
	if start != nil {
		b.query = b.query.Where(field+" >= ?", *start)
	}
	if end != nil {
		// Add one day to include end date fully
		endPlusOne := end.AddDate(0, 0, 1)
		b.query = b.query.Where(field+" < ?", endPlusOne)
	}
	return b
}

// ApplyStatusFilter applies status filtering (handles multiple statuses)
func (b *ReportQueryBuilder) ApplyStatusFilter(field string, statuses []string) *ReportQueryBuilder {
	if len(statuses) > 0 {
		b.query = b.query.Where(field+" IN ?", statuses)
	}
	return b
}

// ApplySearch applies full-text search across multiple fields
func (b *ReportQueryBuilder) ApplySearch(fields []string, term string) *ReportQueryBuilder {
	if term != "" {
		var conditions []string
		var values []interface{}
		searchTerm := "%" + term + "%"
		for _, field := range fields {
			conditions = append(conditions, field+" ILIKE ?")
			values = append(values, searchTerm)
		}
		b.query = b.query.Where(strings.Join(conditions, " OR "), values...)
	}
	return b
}

// ApplyPagination applies limit and offset for pagination
func (b *ReportQueryBuilder) ApplyPagination(limit, offset int) *ReportQueryBuilder {
	if limit > 0 {
		b.query = b.query.Limit(limit)
	}
	if offset > 0 {
		b.query = b.query.Offset(offset)
	}
	return b
}

// ApplySorting applies sorting to the query
func (b *ReportQueryBuilder) ApplySorting(sortBy, sortOrder string, defaultSort string) *ReportQueryBuilder {
	if sortBy != "" {
		order := "DESC"
		if sortOrder == "asc" {
			order = "ASC"
		}
		b.query = b.query.Order(sortBy + " " + order)
	} else if defaultSort != "" {
		b.query = b.query.Order(defaultSort)
	}
	return b
}

// ApplyRangeFilter applies numeric range filtering
func (b *ReportQueryBuilder) ApplyRangeFilter(field string, min, max *float64) *ReportQueryBuilder {
	if min != nil {
		b.query = b.query.Where(field+" >= ?", *min)
	}
	if max != nil {
		b.query = b.query.Where(field+" <= ?", *max)
	}
	return b
}

// ApplyIntRangeFilter applies integer range filtering
func (b *ReportQueryBuilder) ApplyIntRangeFilter(field string, min, max *int64) *ReportQueryBuilder {
	if min != nil {
		b.query = b.query.Where(field+" >= ?", *min)
	}
	if max != nil {
		b.query = b.query.Where(field+" <= ?", *max)
	}
	return b
}

// ApplyBoolFilter applies boolean filtering if value is not nil
func (b *ReportQueryBuilder) ApplyBoolFilter(field string, value *bool) *ReportQueryBuilder {
	if value != nil {
		b.query = b.query.Where(field+" = ?", *value)
	}
	return b
}

// ApplyStringFilter applies exact string match filtering
func (b *ReportQueryBuilder) ApplyStringFilter(field string, value string) *ReportQueryBuilder {
	if value != "" {
		b.query = b.query.Where(field+" = ?", value)
	}
	return b
}

// Build returns the final constructed query
func (b *ReportQueryBuilder) Build() *gorm.DB {
	return b.query
}
