package models

import (
	"kisanlink-erp/internal/constants"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Subcategory represents a product subcategory in the system
// Note: Subcategory names are unique WITHIN a category, not globally.
// Example: "Others" can exist in both "Bio Products" and "Irrigation" categories.
type Subcategory struct {
	base.BaseModel

	Name         string  `gorm:"type:varchar(100);not null;uniqueIndex:idx_subcategory_name_category,priority:1" json:"name"`
	Description  *string `gorm:"type:text" json:"description"`
	CategoryName string  `gorm:"type:varchar(100);not null;uniqueIndex:idx_subcategory_name_category,priority:2" json:"category_name"`

	// Associations
	Category Category `gorm:"foreignKey:CategoryName;references:Name" json:"category,omitempty"`
}

// NewSubcategory creates a new Subcategory with initialized fields
func NewSubcategory(name string, categoryName string, description *string) *Subcategory {
	baseModel := base.NewBaseModel(constants.TableSubcategory, hash.Medium)
	return &Subcategory{
		BaseModel:    *baseModel,
		Name:         name,
		CategoryName: categoryName,
		Description:  description,
	}
}

func (Subcategory) TableName() string {
	return "subcategories"
}

// SubcategoryResponse represents the API response for subcategory
type SubcategoryResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	CategoryName string  `json:"category_name"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// CreateSubcategoryRequest represents the request to create a subcategory
type CreateSubcategoryRequest struct {
	Name         string  `json:"name" binding:"required"`
	CategoryName string  `json:"category_name" binding:"required"`
	Description  *string `json:"description"`
}

// UpdateSubcategoryRequest represents the request to update a subcategory
type UpdateSubcategoryRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}
