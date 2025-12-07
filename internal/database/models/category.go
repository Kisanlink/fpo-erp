package models

import (
	"kisanlink-erp/internal/constants"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Category represents a product category in the system
type Category struct {
	base.BaseModel

	Name        string  `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Description *string `gorm:"type:text" json:"description"`

	// Associations
	Subcategories []Subcategory `gorm:"foreignKey:CategoryName;references:Name" json:"subcategories,omitempty"`
}

// NewCategory creates a new Category with initialized fields
func NewCategory(name string, description *string) *Category {
	baseModel := base.NewBaseModel(constants.TableCategory, hash.Medium)
	return &Category{
		BaseModel:   *baseModel,
		Name:        name,
		Description: description,
	}
}

func (Category) TableName() string {
	return "categories"
}

// CategoryResponse represents the API response for category
type CategoryResponse struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   *string                `json:"description"`
	Subcategories []SubcategoryResponse  `json:"subcategories,omitempty"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

// CreateCategoryRequest represents the request to create a category
type CreateCategoryRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

// UpdateCategoryRequest represents the request to update a category
type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// SeedCategoriesResponse represents the response from seeding categories
type SeedCategoriesResponse struct {
	CategoriesCreated    int `json:"categories_created"`
	SubcategoriesCreated int `json:"subcategories_created"`
	Message              string `json:"message"`
}
