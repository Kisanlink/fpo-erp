package models

import (
	"kisanlink-erp/internal/constants"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// GST-Only Tax System
// Tax rates are now stored on ProductVariant (GSTRate, HSNCode)
// This file contains only structures needed for tax calculation and summary storage

// TaxSummary represents a summary of GST taxes for a sale or return
// Simplified to support only GST (CGST+SGST or IGST)
type TaxSummary struct {
	base.BaseModel
	SaleID   *string `json:"sale_id"`
	ReturnID *string `json:"return_id"`

	// GST Tax breakdown
	CGSTAmount     float64 `json:"cgst_amount" gorm:"default:0"`     // Central GST (intra-state)
	SGSTAmount     float64 `json:"sgst_amount" gorm:"default:0"`     // State GST (intra-state)
	IGSTAmount     float64 `json:"igst_amount" gorm:"default:0"`     // Integrated GST (inter-state)
	TotalTaxAmount float64 `json:"total_tax_amount" gorm:"not null"` // CGST + SGST or IGST

	// Totals
	SubTotal   float64 `json:"sub_total" gorm:"not null"`   // Amount before tax
	GrandTotal float64 `json:"grand_total" gorm:"not null"` // Amount after tax

	// Inter-state flag
	IsInterState bool `json:"is_inter_state" gorm:"default:false"` // true = IGST, false = CGST+SGST
}

func (TaxSummary) TableName() string {
	return "tax_summaries"
}

// NewTaxSummary creates a new TaxSummary with initialized fields
func NewTaxSummary() *TaxSummary {
	baseModel := base.NewBaseModel(constants.TableTaxSummary, hash.Medium)
	return &TaxSummary{
		BaseModel: *baseModel,
	}
}

// GSTCalculation represents the result of a GST calculation for a line item
// Used internally by TaxService to return calculation results
type GSTCalculation struct {
	CGSTAmount     float64 `json:"cgst_amount"`      // Central GST amount
	SGSTAmount     float64 `json:"sgst_amount"`      // State GST amount
	IGSTAmount     float64 `json:"igst_amount"`      // Integrated GST amount
	TotalTaxAmount float64 `json:"total_tax_amount"` // Total GST amount
	IsInterState   bool    `json:"is_inter_state"`   // true if IGST, false if CGST+SGST
}

// TaxSummaryResponse represents the API response for tax summary
type TaxSummaryResponse struct {
	ID             string  `json:"id"`
	SaleID         *string `json:"sale_id,omitempty"`
	ReturnID       *string `json:"return_id,omitempty"`
	CGSTAmount     float64 `json:"cgst_amount"`
	SGSTAmount     float64 `json:"sgst_amount"`
	IGSTAmount     float64 `json:"igst_amount"`
	TotalTaxAmount float64 `json:"total_tax_amount"`
	SubTotal       float64 `json:"sub_total"`
	GrandTotal     float64 `json:"grand_total"`
	IsInterState   bool    `json:"is_inter_state"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// ===== ARCHIVED MODELS (kept for reference and migration compatibility) =====
// The following models are no longer used but kept for database migration safety.
// They can be fully removed after migration scripts archive the data.

// TaxType represents different types of taxes (ARCHIVED - now GST only)
type TaxType string

const (
	TaxTypeCGST TaxType = "cgst" // Central GST
	TaxTypeSGST TaxType = "sgst" // State GST
	TaxTypeIGST TaxType = "igst" // Integrated GST
)

// Tax represents a tax configuration (ARCHIVED - tax rates now on ProductVariant)
// Kept for migration reference only
type Tax struct {
	base.BaseModel
	Code            string  `json:"code" gorm:"uniqueIndex;not null"`
	Name            string  `json:"name" gorm:"not null"`
	Description     string  `json:"description"`
	TaxType         TaxType `json:"tax_type" gorm:"not null"`
	Rate            float64 `json:"rate" gorm:"not null"`
	IsActive        bool    `json:"is_active" gorm:"default:true"`
}

func (Tax) TableName() string {
	return "taxes"
}

// TaxTier represents tiered tax calculation (ARCHIVED - no longer used)
type TaxTier struct {
	base.BaseModel
	TaxID     string   `json:"tax_id" gorm:"not null"`
	MinAmount float64  `json:"min_amount" gorm:"not null"`
	MaxAmount *float64 `json:"max_amount"`
	Rate      float64  `json:"rate" gorm:"not null"`
}

func (TaxTier) TableName() string {
	return "tax_tiers"
}

// TaxApplication represents tax applied to a sale (ARCHIVED - now calculated on-the-fly)
type TaxApplication struct {
	base.BaseModel
	TaxID      string  `json:"tax_id" gorm:"not null"`
	SaleID     *string `json:"sale_id"`
	ReturnID   *string `json:"return_id"`
	ItemID     *string `json:"item_id"`
	BaseAmount float64 `json:"base_amount" gorm:"not null"`
	TaxRate    float64 `json:"tax_rate" gorm:"not null"`
	TaxAmount  float64 `json:"tax_amount" gorm:"not null"`
	TaxType    TaxType `json:"tax_type" gorm:"not null"`
}

func (TaxApplication) TableName() string {
	return "tax_applications"
}
