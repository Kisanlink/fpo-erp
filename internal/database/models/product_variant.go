package models

import (
	"kisanlink-erp/internal/constants"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// ProductVariant represents different packaging/size variants of a product
type ProductVariant struct {
	base.BaseModel

	ProductID string `gorm:"type:varchar(100);not null;index" json:"product_id"`

	// Variant details
	Quantity string `gorm:"type:varchar(50);not null" json:"quantity"` // e.g., "500g", "1kg", "5kg", "1L"
	PackSize string `gorm:"type:varchar(100);not null" json:"pack_size"` // e.g., "Small Pack", "Medium Pack", "Bulk"

	// Pricing (optional - variant-specific pricing)
	Price *float64 `gorm:"type:numeric(12,4)" json:"price"`

	// Inventory tracking
	SKU     *string `gorm:"type:varchar(50);unique" json:"sku"`     // Variant-specific SKU (optional)
	Barcode *string `gorm:"type:varchar(50)" json:"barcode"` // For scanning

	IsActive bool `gorm:"default:true" json:"is_active"`

	// Association
	Product Product `gorm:"foreignKey:ProductID;references:ID;tableName:sku" json:"product,omitempty"`
}

// NewProductVariant creates a new ProductVariant with initialized fields
func NewProductVariant(productID, quantity, packSize string) *ProductVariant {
	baseModel := base.NewBaseModel(constants.TableProductVariant, hash.Medium)
	return &ProductVariant{
		BaseModel: *baseModel,
		ProductID: productID,
		Quantity:  quantity,
		PackSize:  packSize,
		IsActive:  true,
	}
}

func (ProductVariant) TableName() string {
	return "product_variants"
}

// ProductVariantResponse represents the API response for product variant
type ProductVariantResponse struct {
	ID        string   `json:"id"`
	ProductID string   `json:"product_id"`
	Quantity  string   `json:"quantity"`
	PackSize  string   `json:"pack_size"`
	Price     *float64 `json:"price"`
	SKU       *string  `json:"sku"`
	Barcode   *string  `json:"barcode"`
	IsActive  bool     `json:"is_active"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// CreateProductVariantRequest represents the request to create a product variant
type CreateProductVariantRequest struct {
	Quantity string   `json:"quantity" binding:"required"`
	PackSize string   `json:"pack_size" binding:"required"`
	Price    *float64 `json:"price" binding:"omitempty,gt=0"`
	SKU      *string  `json:"sku"`
	Barcode  *string  `json:"barcode"`
}

// UpdateProductVariantRequest represents the request to update a product variant
type UpdateProductVariantRequest struct {
	Quantity *string  `json:"quantity,omitempty"`
	PackSize *string  `json:"pack_size,omitempty"`
	Price    *float64 `json:"price,omitempty"`
	SKU      *string  `json:"sku,omitempty"`
	Barcode  *string  `json:"barcode,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}
