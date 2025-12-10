package models

import (
	"kisanlink-erp/internal/constants"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Subcategory represents a product subcategory in the system
// Note: Subcategory names are unique WITHIN a category, not globally.
// Example: "OTHER" can exist in both "BIO_PRODUCTS" and "IRRIGATION" categories.
// Name is enumeration identifier (ALL_CAPS_SNAKE_CASE)
type Subcategory struct {
	base.BaseModel

	Name        string  `gorm:"type:varchar(100);not null;uniqueIndex:idx_subcategory_name_category,priority:1" json:"name"`
	Description *string `gorm:"type:text" json:"description"`
	CategoryID  string  `gorm:"type:varchar(50);uniqueIndex:idx_subcategory_name_category,priority:2;index:idx_subcategories_category_id" json:"category_id"`
	// NOTE: Category association removed to reduce database load
	// Use CategoryRepository.GetByID() to fetch category if needed
}

// NewSubcategory creates a new Subcategory with initialized fields
func NewSubcategory(name string, categoryID string, description *string) *Subcategory {
	baseModel := base.NewBaseModel(constants.TableSubcategory, hash.Medium)
	return &Subcategory{
		BaseModel:   *baseModel,
		Name:        name,
		CategoryID:  categoryID,
		Description: description,
	}
}

func (Subcategory) TableName() string {
	return "subcategories"
}

// SubcategoryResponse represents the API response for subcategory
type SubcategoryResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	CategoryID  string  `json:"category_id"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// CreateSubcategoryRequest represents the request to create a subcategory
type CreateSubcategoryRequest struct {
	Name        string  `json:"name" binding:"required"`
	CategoryID  string  `json:"category_id" binding:"required"`
	Description *string `json:"description"`
}

// UpdateSubcategoryRequest represents the request to update a subcategory
type UpdateSubcategoryRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}
