package models

import (
	"kisanlink-erp/internal/constants"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Price type constants
const (
	PriceTypeMRP = "MRP" // Maximum Retail Price
	PriceTypeMSP = "MSP" // Minimum Selling Price
)

// VariantPrice represents a price point for a variant
type VariantPrice struct {
	PriceType string  `json:"price_type"` // "MRP" or "MSP"
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"` // "INR", "USD", etc.
}

// ProductVariant represents different packaging/size variants of a product
// Can also represent collaborator-specific variants (supplier-specific products)
type ProductVariant struct {
	base.BaseModel

	ProductID string `gorm:"type:varchar(100);not null;index:idx_product_collaborator" json:"product_id"`

	// E-commerce Integration
	ExternalID *string `gorm:"type:varchar(100);unique;index" json:"external_id"` // E-commerce variant ID for webhook matching

	// Variant identification
	VariantName string  `gorm:"type:varchar(150);not null" json:"variant_name"` // e.g., "1kg Premium Pack"
	Description *string `gorm:"type:text" json:"description"`                   // Detailed variant description

	// Variant details
	Quantity string `gorm:"type:varchar(50);not null" json:"quantity"`   // e.g., "500g", "1kg", "5kg", "1L"
	PackSize string `gorm:"type:varchar(100);not null" json:"pack_size"` // e.g., "Small Pack", "Medium Pack", "Bulk"

	// Inventory tracking
	SKU     *string `gorm:"type:varchar(50);unique" json:"sku"` // Variant-specific SKU (optional)
	Barcode *string `gorm:"type:varchar(50)" json:"barcode"`    // For scanning

	// Collaborator-specific fields (optional - for supplier-specific variants)
	CollaboratorIDs    []string `gorm:"type:json;serializer:json" json:"collaborator_ids"` // Multiple collaborators can supply same variant
	BrandName          *string  `gorm:"type:varchar(100)" json:"brand_name"`                // Collaborator's brand
	HSNCode            *string  `gorm:"type:varchar(8)" json:"hsn_code"`                    // For GST classification
	GSTRate            *float64 `gorm:"type:numeric(5,2)" json:"gst_rate"`                  // e.g., 5.00, 12.00, 18.00, 28.00
	Images             *string  `gorm:"type:json" json:"images"`                            // JSON array of S3 paths
	DosageInstructions *string  `gorm:"type:text" json:"dosage_instructions"`               // Usage instructions
	UsageDetails       *string  `gorm:"type:text" json:"usage_details"`                     // Detailed usage

	// Note: Prices are stored in product_prices table, not embedded in variant

	IsActive bool `gorm:"default:true" json:"is_active"`

	// Associations
	Product Product `gorm:"foreignKey:ProductID;references:ID" json:"product,omitempty"`
	// Note: Collaborator association removed since this is now many-to-many via JSON array
}

// NewProductVariant creates a new ProductVariant with initialized fields
func NewProductVariant(productID, variantName, quantity, packSize string) *ProductVariant {
	baseModel := base.NewBaseModel(constants.TableProductVariant, hash.Medium)
	return &ProductVariant{
		BaseModel:   *baseModel,
		ProductID:   productID,
		VariantName: variantName,
		Quantity:    quantity,
		PackSize:    packSize,
		IsActive:    true,
	}
}

// NewCollaboratorVariant creates a new ProductVariant for specific collaborator(s)/supplier(s)
func NewCollaboratorVariant(productID string, collaboratorIDs []string, variantName, quantity, packSize, brandName, hsnCode string, gstRate float64) *ProductVariant {
	baseModel := base.NewBaseModel(constants.TableProductVariant, hash.Medium)
	return &ProductVariant{
		BaseModel:       *baseModel,
		ProductID:       productID,
		CollaboratorIDs: collaboratorIDs,
		VariantName:     variantName,
		Quantity:        quantity,
		PackSize:        packSize,
		BrandName:       &brandName,
		HSNCode:         &hsnCode,
		GSTRate:         &gstRate,
		IsActive:        true,
	}
}

func (ProductVariant) TableName() string {
	return "product_variants"
}

// ProductVariantResponse represents the API response for product variant
type ProductVariantResponse struct {
	ID                 string   `json:"id"`
	ProductID          string   `json:"product_id"`
	VariantName        string   `json:"variant_name"`
	Description        *string  `json:"description"`
	Quantity           string   `json:"quantity"`
	PackSize           string   `json:"pack_size"`
	SKU                *string  `json:"sku"`
	Barcode            *string  `json:"barcode"`
	CollaboratorIDs    []string `json:"collaborator_ids,omitempty"` // Multiple collaborators
	BrandName          *string  `json:"brand_name,omitempty"`
	HSNCode            *string  `json:"hsn_code,omitempty"`
	GSTRate            *float64       `json:"gst_rate,omitempty"`
	Images             []string       `json:"images,omitempty"`     // S3 paths (for reference)
	ImageURLs          []string       `json:"image_urls,omitempty"` // Presigned URLs (valid for 1 hour)
	DosageInstructions *string        `json:"dosage_instructions,omitempty"`
	UsageDetails       *string        `json:"usage_details,omitempty"`
	Prices             []ProductPriceResponse `json:"prices"` // Fetched from product_prices table
	IsActive           bool           `json:"is_active"`
	CreatedAt          string         `json:"created_at"`
	UpdatedAt          string         `json:"updated_at"`
}

// CreateProductVariantRequest represents the request to create a product variant
type CreateProductVariantRequest struct {
	VariantName        string   `json:"variant_name" binding:"required"`
	Description        *string  `json:"description"`
	Quantity           string   `json:"quantity" binding:"required"`
	PackSize           string   `json:"pack_size" binding:"required"`
	SKU                *string  `json:"sku"`
	Barcode            *string  `json:"barcode"`
	CollaboratorIDs    []string `json:"collaborator_ids"`                           // Optional: multiple collaborators can supply same variant
	BrandName          *string  `json:"brand_name"`                                 // Required if collaborator_ids provided
	HSNCode            *string  `json:"hsn_code"`                                   // Required if collaborator_ids provided
	GSTRate            *float64 `json:"gst_rate" binding:"omitempty,min=0,max=100"` // Required if collaborator_ids provided
	Images             []string       `json:"images"`
	DosageInstructions *string        `json:"dosage_instructions"`
	UsageDetails       *string        `json:"usage_details"`
	Prices             []VariantPrice `json:"prices"` // Optional: can create variant without prices
}

// UpdateProductVariantRequest represents the request to update a product variant
type UpdateProductVariantRequest struct {
	VariantName        *string   `json:"variant_name,omitempty"`
	Description        *string   `json:"description,omitempty"`
	Quantity           *string   `json:"quantity,omitempty"`
	PackSize           *string   `json:"pack_size,omitempty"`
	SKU                *string   `json:"sku,omitempty"`
	Barcode            *string   `json:"barcode,omitempty"`
	CollaboratorIDs    *[]string `json:"collaborator_ids,omitempty"` // Optional: update collaborator associations
	BrandName          *string   `json:"brand_name,omitempty"`
	HSNCode            *string   `json:"hsn_code,omitempty"`
	GSTRate            *float64  `json:"gst_rate,omitempty"`
	Images             *[]string       `json:"images,omitempty"`
	DosageInstructions *string         `json:"dosage_instructions,omitempty"`
	UsageDetails       *string         `json:"usage_details,omitempty"`
	Prices             *[]VariantPrice `json:"prices,omitempty"` // Optional: update prices
	IsActive           *bool           `json:"is_active,omitempty"`
}
