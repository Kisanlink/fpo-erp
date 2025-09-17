package models

import (
	"kisanlink-erp/internal/constants"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// ProductPrice represents a price for a product
type ProductPrice struct {
	base.BaseModel
	ProductID     string     `gorm:"type:varchar(100);not null" json:"product_id"`
	PriceType     string     `gorm:"type:varchar(50);not null" json:"price_type"` // e.g., "retail", "wholesale", "bulk"
	Price         float64    `gorm:"type:numeric(12,4);not null" json:"price"`
	Currency      string     `gorm:"type:varchar(3);not null;default:'USD'" json:"currency"`
	EffectiveFrom time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"effective_from"`
	EffectiveTo   *time.Time `gorm:"type:timestamptz" json:"effective_to"`
	IsActive      bool       `gorm:"type:boolean;not null;default:true" json:"is_active"`

	// Associations
	Product Product `gorm:"foreignKey:ProductID;references:ID;tableName:sku" json:"product,omitempty"`
}

// NewProductPrice creates a new ProductPrice with initialized fields
func NewProductPrice(productID, priceType string, price float64, currency string, effectiveFrom time.Time, effectiveTo *time.Time, isActive bool) *ProductPrice {
	baseModel := base.NewBaseModel(constants.TablePrice, hash.Medium)
	return &ProductPrice{
		BaseModel:     *baseModel,
		ProductID:     productID,
		PriceType:     priceType,
		Price:         price,
		Currency:      currency,
		EffectiveFrom: effectiveFrom,
		EffectiveTo:   effectiveTo,
		IsActive:      isActive,
	}
}

func (ProductPrice) TableName() string {
	return "product_prices"
}

// ProductPriceResponse represents the API response for product price
type ProductPriceResponse struct {
	ID            string  `json:"id"`
	ProductID     string  `json:"product_id"`
	PriceType     string  `json:"price_type"`
	Price         float64 `json:"price"`
	Currency      string  `json:"currency"`
	EffectiveFrom string  `json:"effective_from"`
	EffectiveTo   *string `json:"effective_to,omitempty"`
	IsActive      bool    `json:"is_active"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// CreateProductPriceRequest represents the request to create a product price
type CreateProductPriceRequest struct {
	ProductID     string  `json:"product_id" binding:"required"`
	PriceType     string  `json:"price_type" binding:"required"`
	Price         float64 `json:"price" binding:"required,gt=0"`
	Currency      string  `json:"currency"`
	EffectiveFrom *string `json:"effective_from"`
	EffectiveTo   *string `json:"effective_to"`
	IsActive      *bool   `json:"is_active"`
}

// UpdateProductPriceRequest represents the request to update a product price
type UpdateProductPriceRequest struct {
	PriceType     *string  `json:"price_type,omitempty"`
	Price         *float64 `json:"price,omitempty"`
	Currency      *string  `json:"currency,omitempty"`
	EffectiveFrom *string  `json:"effective_from,omitempty"`
	EffectiveTo   *string  `json:"effective_to,omitempty"`
	IsActive      *bool    `json:"is_active,omitempty"`
}

// ProductWithPricesResponse represents a product with its prices
type ProductWithPricesResponse struct {
	ID          string                 `json:"id"`
	SKU         string                 `json:"sku"`
	Name        string                 `json:"name"`
	Description *string                `json:"description"`
	Prices      []ProductPriceResponse `json:"prices,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}
