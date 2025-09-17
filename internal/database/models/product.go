package models

import (
	"kisanlink-erp/internal/constants"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Product represents a product in the system
type Product struct {
	base.BaseModel
	SKU         string  `gorm:"type:varchar(50);unique;not null" json:"sku"`
	Name        string  `gorm:"type:varchar(150);not null" json:"name"`
	Description *string `gorm:"type:text" json:"description"`

	// Associations
	Prices []ProductPrice `gorm:"foreignKey:ProductID;references:ID;tableName:product_prices" json:"prices,omitempty"`
}

// NewProduct creates a new Product with initialized fields
func NewProduct(sku, name string, description *string) *Product {
	baseModel := base.NewBaseModel(constants.TableProduct, hash.Medium)
	return &Product{
		BaseModel:   *baseModel,
		SKU:         sku,
		Name:        name,
		Description: description,
	}
}

func (Product) TableName() string {
	return "sku"
}

// ProductResponse represents the API response for product
type ProductResponse struct {
	ID          string  `json:"id"`
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// CreateProductRequest represents the request to create a product
type CreateProductRequest struct {
	SKU         string  `json:"sku" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

// UpdateProductRequest represents the request to update a product
type UpdateProductRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}
