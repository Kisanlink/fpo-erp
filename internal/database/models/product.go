package models

import (
	"kisanlink-erp/internal/constants"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Product represents a generic product category in the system
// SKU is now at the variant level for actual inventory tracking
type Product struct {
	base.BaseModel

	// E-commerce Integration
	ExternalID *string `gorm:"type:varchar(100);unique;index" json:"external_id"` // E-commerce product ID for webhook matching

	Name        string  `gorm:"type:varchar(150);not null" json:"name"`
	Description *string `gorm:"type:text" json:"description"`

	// Category fields - ID-based references (optional)
	CategoryID    *string `gorm:"type:varchar(50);index" json:"category_id"`
	SubcategoryID *string `gorm:"type:varchar(50);index" json:"subcategory_id"`

	// Category associations
	Category    *Category    `gorm:"foreignKey:CategoryID;references:ID" json:"category,omitempty"`
	Subcategory *Subcategory `gorm:"foreignKey:SubcategoryID;references:ID" json:"subcategory,omitempty"`

	// Associations
	Variants []ProductVariant `gorm:"foreignKey:ProductID" json:"variants,omitempty"`
}

// NewProduct creates a new Product with initialized fields
func NewProduct(name string, description *string) *Product {
	baseModel := base.NewBaseModel(constants.TableProduct, hash.Medium)
	return &Product{
		BaseModel:   *baseModel,
		Name:        name,
		Description: description,
	}
}

func (Product) TableName() string {
	return "products"
}

// ProductResponse represents the API response for product
type ProductResponse struct {
	ID            string                   `json:"id"`
	Name          string                   `json:"name"`
	Description   *string                  `json:"description"`
	CategoryID    *string                  `json:"category_id,omitempty"`
	SubcategoryID *string                  `json:"subcategory_id,omitempty"`
	Variants      []ProductVariantResponse `json:"variants,omitempty"` // Preloaded variants with image URLs
	CreatedAt     string                   `json:"created_at"`
	UpdatedAt     string                   `json:"updated_at"`
}

// CreateProductRequest represents the request to create a product
type CreateProductRequest struct {
	Name          string  `json:"name" binding:"required"`
	Description   *string `json:"description"`
	CategoryID    *string `json:"category_id"`    // Optional
	SubcategoryID *string `json:"subcategory_id"` // Optional
}

// UpdateProductRequest represents the request to update a product
type UpdateProductRequest struct {
	Name          *string `json:"name,omitempty"`
	Description   *string `json:"description,omitempty"`
	CategoryID    *string `json:"category_id,omitempty"`
	SubcategoryID *string `json:"subcategory_id,omitempty"`
}
