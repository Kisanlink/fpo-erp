package models

import (
	"kisanlink-erp/internal/constants"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// CollaboratorProduct represents the junction table linking collaborators to products
// with additional metadata from the collaborator's perspective
type CollaboratorProduct struct {
	base.BaseModel

	CollaboratorID string `gorm:"type:varchar(100);not null;index:idx_collab_product" json:"collaborator_id"`
	ProductID      string `gorm:"type:varchar(100);not null;index:idx_collab_product" json:"product_id"`

	// Product details from collaborator (BRD requirements)
	BrandName string  `gorm:"type:varchar(100);not null" json:"brand_name"` // Collaborator's brand
	HSNCode   string  `gorm:"type:varchar(8);not null" json:"hsn_code"`     // For GST classification
	GSTRate   float64 `gorm:"type:numeric(5,2);not null" json:"gst_rate"`   // e.g., 5.00, 12.00, 18.00, 28.00

	// Rich product info
	Images             *string `gorm:"type:json" json:"images"`              // JSON array of S3 paths
	DosageInstructions *string `gorm:"type:text" json:"dosage_instructions"` // Usage instructions
	UsageDetails       *string `gorm:"type:text" json:"usage_details"`       // Detailed usage

	IsActive bool `gorm:"default:true" json:"is_active"`

	// Associations
	Collaborator Collaborator `gorm:"foreignKey:CollaboratorID" json:"collaborator,omitempty"`
	Product      Product      `gorm:"foreignKey:ProductID;references:ID;tableName:sku" json:"product,omitempty"`
}

// NewCollaboratorProduct creates a new CollaboratorProduct with initialized fields
func NewCollaboratorProduct(collaboratorID, productID, brandName, hsnCode string, gstRate float64) *CollaboratorProduct {
	baseModel := base.NewBaseModel(constants.TableCollaboratorProduct, hash.Medium)
	return &CollaboratorProduct{
		BaseModel:      *baseModel,
		CollaboratorID: collaboratorID,
		ProductID:      productID,
		BrandName:      brandName,
		HSNCode:        hsnCode,
		GSTRate:        gstRate,
		IsActive:       true,
	}
}

func (CollaboratorProduct) TableName() string {
	return "collaborator_products"
}

// CollaboratorProductResponse represents the API response for collaborator product
type CollaboratorProductResponse struct {
	ID                 string          `json:"id"`
	CollaboratorID     string          `json:"collaborator_id"`
	CollaboratorName   string          `json:"collaborator_name"`
	ProductID          string          `json:"product_id"`
	ProductName        string          `json:"product_name"`
	ProductSKU         string          `json:"product_sku"`
	BrandName          string          `json:"brand_name"`
	HSNCode            string          `json:"hsn_code"`
	GSTRate            float64         `json:"gst_rate"`
	Images             []string        `json:"images,omitempty"` // Parsed from JSON
	DosageInstructions *string         `json:"dosage_instructions"`
	UsageDetails       *string         `json:"usage_details"`
	IsActive           bool            `json:"is_active"`
	Product            *ProductSummary `json:"product,omitempty"` // Embedded product info
	CreatedAt          string          `json:"created_at"`
	UpdatedAt          string          `json:"updated_at"`
}

// ProductSummary is a simplified product representation for nested responses
type ProductSummary struct {
	ID          string  `json:"id"`
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// CreateCollaboratorProductRequest represents the request to add a product to collaborator
type CreateCollaboratorProductRequest struct {
	ProductID          string   `json:"product_id" binding:"required"`
	BrandName          string   `json:"brand_name" binding:"required"`
	HSNCode            string   `json:"hsn_code" binding:"required,len=8"` // HSN codes are 8 digits
	GSTRate            float64  `json:"gst_rate" binding:"required,min=0,max=100"`
	Images             []string `json:"images"` // Array of S3 paths
	DosageInstructions *string  `json:"dosage_instructions"`
	UsageDetails       *string  `json:"usage_details"`
}

// UpdateCollaboratorProductRequest represents the request to update collaborator product metadata
type UpdateCollaboratorProductRequest struct {
	BrandName          *string   `json:"brand_name,omitempty"`
	HSNCode            *string   `json:"hsn_code,omitempty"`
	GSTRate            *float64  `json:"gst_rate,omitempty"`
	Images             *[]string `json:"images,omitempty"`
	DosageInstructions *string   `json:"dosage_instructions,omitempty"`
	UsageDetails       *string   `json:"usage_details,omitempty"`
	IsActive           *bool     `json:"is_active,omitempty"`
}
