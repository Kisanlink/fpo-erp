package models

import (
	"kisanlink-erp/internal/constants"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// TaxType represents different types of taxes
type TaxType string

const (
	TaxTypeCGST         TaxType = "cgst"          // Central Goods and Services Tax
	TaxTypeSGST         TaxType = "sgst"          // State Goods and Services Tax
	TaxTypeIGST         TaxType = "igst"          // Integrated Goods and Services Tax
	TaxTypeVAT          TaxType = "vat"           // Value Added Tax
	TaxTypeSTT          TaxType = "stt"           // Securities Transaction Tax
	TaxTypeTDS          TaxType = "tds"           // Tax Deducted at Source
	TaxTypeTCS          TaxType = "tcs"           // Tax Collected at Source
	TaxTypeExcise       TaxType = "excise"        // Excise Duty
	TaxTypeCustoms      TaxType = "customs"       // Customs Duty
	TaxTypeItemSpecific TaxType = "item_specific" // Item-specific tax
	TaxTypeCategory     TaxType = "category"      // Category-based tax
	TaxTypeFlat         TaxType = "flat"          // Flat tax amount
)

// TaxCalculationType represents how tax is calculated
type TaxCalculationType string

const (
	TaxCalculationPercentage TaxCalculationType = "percentage" // Percentage of base amount
	TaxCalculationFixed      TaxCalculationType = "fixed"      // Fixed amount per unit
	TaxCalculationTiered     TaxCalculationType = "tiered"     // Tiered calculation
)

// TaxStatus represents the status of a tax
type TaxStatus string

const (
	TaxStatusActive   TaxStatus = "active"
	TaxStatusInactive TaxStatus = "inactive"
	TaxStatusExpired  TaxStatus = "expired"
)

// Tax represents a tax configuration
type Tax struct {
	base.BaseModel
	Code            string             `json:"code" gorm:"uniqueIndex;not null"`
	Name            string             `json:"name" gorm:"not null"`
	Description     string             `json:"description"`
	TaxType         TaxType            `json:"tax_type" gorm:"not null"`
	CalculationType TaxCalculationType `json:"calculation_type" gorm:"not null"`
	Rate            float64            `json:"rate" gorm:"not null"` // Percentage or fixed amount
	MinAmount       *float64           `json:"min_amount"`           // Minimum tax amount
	MaxAmount       *float64           `json:"max_amount"`           // Maximum tax amount
	MinOrderValue   *float64           `json:"min_order_value"`      // Minimum order value for tax to apply
	MaxOrderValue   *float64           `json:"max_order_value"`      // Maximum order value for tax to apply

	// Applicability
	ApplicableProducts   []string `json:"applicable_products" gorm:"type:json"`   // Product IDs where tax applies
	ExcludedProducts     []string `json:"excluded_products" gorm:"type:json"`     // Product IDs where tax doesn't apply
	ApplicableCategories []string `json:"applicable_categories" gorm:"type:json"` // Category IDs where tax applies
	ExcludedCategories   []string `json:"excluded_categories" gorm:"type:json"`   // Category IDs where tax doesn't apply
	ApplicableWarehouses []string `json:"applicable_warehouses" gorm:"type:json"` // Warehouse IDs where tax applies
	ExcludedWarehouses   []string `json:"excluded_warehouses" gorm:"type:json"`   // Warehouse IDs where tax doesn't apply
	ApplicableStates     []string `json:"applicable_states" gorm:"type:json"`     // State codes where tax applies
	ExcludedStates       []string `json:"excluded_states" gorm:"type:json"`       // State codes where tax doesn't apply

	// Customer groups
	ApplicableCustomerGroups []string `json:"applicable_customer_groups" gorm:"type:json"`
	ExcludedCustomerGroups   []string `json:"excluded_customer_groups" gorm:"type:json"`

	// Validity
	ValidFrom  time.Time  `json:"valid_from" gorm:"not null"`
	ValidUntil *time.Time `json:"valid_until"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`

	// Priority and stacking
	Priority      int  `json:"priority" gorm:"default:0"` // Higher priority taxes are applied first
	IsStackable   bool `json:"is_stackable" gorm:"default:true"`
	StackingOrder int  `json:"stacking_order" gorm:"default:0"`

	// Special conditions
	RequiresGSTIN bool `json:"requires_gstin" gorm:"default:false"` // Whether GSTIN is required
	RequiresPAN   bool `json:"requires_pan" gorm:"default:false"`   // Whether PAN is required
	IsInterState  bool `json:"is_inter_state" gorm:"default:false"` // Whether it's inter-state transaction

	// Reporting
	HSNCode     *string `json:"hsn_code"`     // HSN/SAC code for GST
	SACCode     *string `json:"sac_code"`     // SAC code for services
	TaxCategory *string `json:"tax_category"` // Tax category for reporting
}

func (Tax) TableName() string {
	return "taxes"
}

// TaxTier represents tiered tax calculation
type TaxTier struct {
	base.BaseModel
	TaxID       string   `json:"tax_id" gorm:"not null"`
	MinAmount   float64  `json:"min_amount" gorm:"not null"`
	MaxAmount   *float64 `json:"max_amount"`
	Rate        float64  `json:"rate" gorm:"not null"`
	FixedAmount *float64 `json:"fixed_amount"`
}

func (TaxTier) TableName() string {
	return "tax_tiers"
}

// TaxApplication represents tax applied to a sale or return
type TaxApplication struct {
	base.BaseModel
	TaxID    string  `json:"tax_id" gorm:"not null"`
	SaleID   *string `json:"sale_id"`
	ReturnID *string `json:"return_id"`
	ItemID   *string `json:"item_id"` // SaleItem or ReturnItem ID

	// Tax calculation details
	BaseAmount float64 `json:"base_amount" gorm:"not null"` // Amount on which tax is calculated
	TaxRate    float64 `json:"tax_rate" gorm:"not null"`    // Applied tax rate
	TaxAmount  float64 `json:"tax_amount" gorm:"not null"`  // Calculated tax amount
	TaxType    TaxType `json:"tax_type" gorm:"not null"`

	// Additional details
	HSNCode     *string `json:"hsn_code"`
	SACCode     *string `json:"sac_code"`
	TaxCategory *string `json:"tax_category"`
}

// TaxSummary represents a summary of taxes for a sale or return
type TaxSummary struct {
	base.BaseModel
	SaleID   *string `json:"sale_id"`
	ReturnID *string `json:"return_id"`

	// Tax breakdown
	CGSTAmount     float64 `json:"cgst_amount" gorm:"default:0"`
	SGSTAmount     float64 `json:"sgst_amount" gorm:"default:0"`
	IGSTAmount     float64 `json:"igst_amount" gorm:"default:0"`
	VATAmount      float64 `json:"vat_amount" gorm:"default:0"`
	STTAmount      float64 `json:"stt_amount" gorm:"default:0"`
	TDSAmount      float64 `json:"tds_amount" gorm:"default:0"`
	TCSAmount      float64 `json:"tcs_amount" gorm:"default:0"`
	ExciseAmount   float64 `json:"excise_amount" gorm:"default:0"`
	CustomsAmount  float64 `json:"customs_amount" gorm:"default:0"`
	OtherTaxAmount float64 `json:"other_tax_amount" gorm:"default:0"`

	// Totals
	TotalTaxAmount float64 `json:"total_tax_amount" gorm:"not null"`
	SubTotal       float64 `json:"sub_total" gorm:"not null"`   // Amount before tax
	GrandTotal     float64 `json:"grand_total" gorm:"not null"` // Amount after tax
}

func (TaxApplication) TableName() string {
	return "tax_applications"
}

func (TaxSummary) TableName() string {
	return "tax_summaries"
}

// Request/Response models

type TaxResponse struct {
	ID                       string             `json:"id"`
	Code                     string             `json:"code"`
	Name                     string             `json:"name"`
	Description              string             `json:"description"`
	TaxType                  TaxType            `json:"tax_type"`
	CalculationType          TaxCalculationType `json:"calculation_type"`
	Rate                     float64            `json:"rate"`
	MinAmount                *float64           `json:"min_amount"`
	MaxAmount                *float64           `json:"max_amount"`
	MinOrderValue            *float64           `json:"min_order_value"`
	MaxOrderValue            *float64           `json:"max_order_value"`
	ApplicableProducts       []string           `json:"applicable_products"`
	ExcludedProducts         []string           `json:"excluded_products"`
	ApplicableCategories     []string           `json:"applicable_categories"`
	ExcludedCategories       []string           `json:"excluded_categories"`
	ApplicableWarehouses     []string           `json:"applicable_warehouses"`
	ExcludedWarehouses       []string           `json:"excluded_warehouses"`
	ApplicableStates         []string           `json:"applicable_states"`
	ExcludedStates           []string           `json:"excluded_states"`
	ApplicableCustomerGroups []string           `json:"applicable_customer_groups"`
	ExcludedCustomerGroups   []string           `json:"excluded_customer_groups"`
	ValidFrom                time.Time          `json:"valid_from"`
	ValidUntil               *time.Time         `json:"valid_until"`
	IsActive                 bool               `json:"is_active"`
	Priority                 int                `json:"priority"`
	IsStackable              bool               `json:"is_stackable"`
	StackingOrder            int                `json:"stacking_order"`
	RequiresGSTIN            bool               `json:"requires_gstin"`
	RequiresPAN              bool               `json:"requires_pan"`
	IsInterState             bool               `json:"is_inter_state"`
	HSNCode                  *string            `json:"hsn_code"`
	SACCode                  *string            `json:"sac_code"`
	TaxCategory              *string            `json:"tax_category"`
	Status                   string             `json:"status"`
	CreatedBy                string             `json:"created_by"`
	UpdatedBy                string             `json:"updated_by"`
	CreatedAt                time.Time          `json:"created_at"`
	UpdatedAt                time.Time          `json:"updated_at"`
}

type CreateTaxRequest struct {
	Code                     string             `json:"code" binding:"required"`
	Name                     string             `json:"name" binding:"required"`
	Description              string             `json:"description"`
	TaxType                  TaxType            `json:"tax_type" binding:"required"`
	CalculationType          TaxCalculationType `json:"calculation_type" binding:"required"`
	Rate                     float64            `json:"rate" binding:"required,min=0"`
	MinAmount                *float64           `json:"min_amount"`
	MaxAmount                *float64           `json:"max_amount"`
	MinOrderValue            *float64           `json:"min_order_value"`
	MaxOrderValue            *float64           `json:"max_order_value"`
	ApplicableProducts       []string           `json:"applicable_products"`
	ExcludedProducts         []string           `json:"excluded_products"`
	ApplicableCategories     []string           `json:"applicable_categories"`
	ExcludedCategories       []string           `json:"excluded_categories"`
	ApplicableWarehouses     []string           `json:"applicable_warehouses"`
	ExcludedWarehouses       []string           `json:"excluded_warehouses"`
	ApplicableStates         []string           `json:"applicable_states"`
	ExcludedStates           []string           `json:"excluded_states"`
	ApplicableCustomerGroups []string           `json:"applicable_customer_groups"`
	ExcludedCustomerGroups   []string           `json:"excluded_customer_groups"`
	ValidFrom                time.Time          `json:"valid_from" binding:"required"`
	ValidUntil               *time.Time         `json:"valid_until"`
	IsActive                 bool               `json:"is_active"`
	Priority                 int                `json:"priority"`
	IsStackable              bool               `json:"is_stackable"`
	StackingOrder            int                `json:"stacking_order"`
	RequiresGSTIN            bool               `json:"requires_gstin"`
	RequiresPAN              bool               `json:"requires_pan"`
	IsInterState             bool               `json:"is_inter_state"`
	HSNCode                  *string            `json:"hsn_code"`
	SACCode                  *string            `json:"sac_code"`
	TaxCategory              *string            `json:"tax_category"`
}

type UpdateTaxRequest struct {
	Name                     *string             `json:"name"`
	Description              *string             `json:"description"`
	CalculationType          *TaxCalculationType `json:"calculation_type"`
	Rate                     *float64            `json:"rate"`
	MinAmount                *float64            `json:"min_amount"`
	MaxAmount                *float64            `json:"max_amount"`
	MinOrderValue            *float64            `json:"min_order_value"`
	MaxOrderValue            *float64            `json:"max_order_value"`
	ApplicableProducts       []string            `json:"applicable_products"`
	ExcludedProducts         []string            `json:"excluded_products"`
	ApplicableCategories     []string            `json:"applicable_categories"`
	ExcludedCategories       []string            `json:"excluded_categories"`
	ApplicableWarehouses     []string            `json:"applicable_warehouses"`
	ExcludedWarehouses       []string            `json:"excluded_warehouses"`
	ApplicableStates         []string            `json:"applicable_states"`
	ExcludedStates           []string            `json:"excluded_states"`
	ApplicableCustomerGroups []string            `json:"applicable_customer_groups"`
	ExcludedCustomerGroups   []string            `json:"excluded_customer_groups"`
	ValidFrom                *time.Time          `json:"valid_from"`
	ValidUntil               *time.Time          `json:"valid_until"`
	IsActive                 *bool               `json:"is_active"`
	Priority                 *int                `json:"priority"`
	IsStackable              *bool               `json:"is_stackable"`
	StackingOrder            *int                `json:"stacking_order"`
	RequiresGSTIN            *bool               `json:"requires_gstin"`
	RequiresPAN              *bool               `json:"requires_pan"`
	IsInterState             *bool               `json:"is_inter_state"`
	HSNCode                  *string             `json:"hsn_code"`
	SACCode                  *string             `json:"sac_code"`
	TaxCategory              *string             `json:"tax_category"`
}

type TaxCalculationRequest struct {
	CustomerID     *string              `json:"customer_id"`         // Optional - not required when no customer management
	CustomerState  *string              `json:"customer_state"`      // Optional - not required when no customer management
	CustomerGSTIN  *string              `json:"customer_gstin"`
	CustomerPAN    *string              `json:"customer_pan"`
	WarehouseID    string               `json:"warehouse_id" binding:"required"`
	WarehouseState string               `json:"warehouse_state" binding:"required"`
	Items          []TaxCalculationItem `json:"items" binding:"required,min=1"`
	IsInterState   bool                 `json:"is_inter_state"`
}

type TaxCalculationItem struct {
	ProductID  string  `json:"product_id" binding:"required"`
	CategoryID *string `json:"category_id"`
	Quantity   int     `json:"quantity" binding:"required,min=1"`
	UnitPrice  float64 `json:"unit_price" binding:"required,min=0"`
	LineTotal  float64 `json:"line_total" binding:"required,min=0"`
}

type TaxCalculationResponse struct {
	SubTotal       float64        `json:"sub_total"`
	TaxBreakdown   []TaxBreakdown `json:"tax_breakdown"`
	TotalTaxAmount float64        `json:"total_tax_amount"`
	GrandTotal     float64        `json:"grand_total"`
	AppliedTaxes   []AppliedTax   `json:"applied_taxes"`
}

type TaxBreakdown struct {
	TaxType TaxType `json:"tax_type"`
	TaxName string  `json:"tax_name"`
	TaxCode string  `json:"tax_code"`
	Rate    float64 `json:"rate"`
	Amount  float64 `json:"amount"`
	HSNCode *string `json:"hsn_code"`
	SACCode *string `json:"sac_code"`
}

type AppliedTax struct {
	TaxID      string  `json:"tax_id"`
	TaxCode    string  `json:"tax_code"`
	TaxName    string  `json:"tax_name"`
	TaxType    TaxType `json:"tax_type"`
	Rate       float64 `json:"rate"`
	Amount     float64 `json:"amount"`
	BaseAmount float64 `json:"base_amount"`
}

// Constructor functions
func NewTax() *Tax {
	baseModel := base.NewBaseModel(constants.TableTax, hash.Medium)
	return &Tax{
		BaseModel:     *baseModel,
		IsActive:      true,
		IsStackable:   true,
		Priority:      0,
		StackingOrder: 0,
	}
}

func NewTaxTier() *TaxTier {
	baseModel := base.NewBaseModel(constants.TableTaxTier, hash.Medium)
	return &TaxTier{
		BaseModel: *baseModel,
	}
}

func NewTaxApplication() *TaxApplication {
	baseModel := base.NewBaseModel(constants.TableTaxApp, hash.Medium)
	return &TaxApplication{
		BaseModel: *baseModel,
	}
}

// TaxApplicationResponse represents the API response for tax application
type TaxApplicationResponse struct {
	ID                string  `json:"id"`
	SaleID            string  `json:"sale_id"`
	TaxID             string  `json:"tax_id"`
	BaseAmount        float64 `json:"base_amount"`
	TaxableAmount     float64 `json:"taxable_amount"`
	TaxRate           float64 `json:"tax_rate"`
	TaxAmount         float64 `json:"tax_amount"`
	CGSTAmount        float64 `json:"cgst_amount"`
	SGSTAmount        float64 `json:"sgst_amount"`
	IGSTAmount        float64 `json:"igst_amount"`
	VATAmount         float64 `json:"vat_amount"`
	STTAmount         float64 `json:"stt_amount"`
	TDSAmount         float64 `json:"tds_amount"`
	TCSAmount         float64 `json:"tcs_amount"`
	ExciseAmount      float64 `json:"excise_amount"`
	CustomsAmount     float64 `json:"customs_amount"`
	OtherTaxAmount    float64 `json:"other_tax_amount"`
	TotalTaxAmount    float64 `json:"total_tax_amount"`
	SubTotal          float64 `json:"sub_total"`
	GrandTotal        float64 `json:"grand_total"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

// TaxSummaryResponse represents the API response for tax summary
type TaxSummaryResponse struct {
	ID              string  `json:"id"`
	SummaryDate     string  `json:"summary_date"`
	WarehouseID     string  `json:"warehouse_id"`
	TotalTax        float64 `json:"total_tax"`
	TotalCGST       float64 `json:"total_cgst"`
	TotalSGST       float64 `json:"total_sgst"`
	TotalIGST       float64 `json:"total_igst"`
	TotalVAT        float64 `json:"total_vat"`
	TotalSTT        float64 `json:"total_stt"`
	TotalTDS        float64 `json:"total_tds"`
	TotalTCS        float64 `json:"total_tcs"`
	TotalExcise     float64 `json:"total_excise"`
	TotalCustoms    float64 `json:"total_customs"`
	TotalOtherTax   float64 `json:"total_other_tax"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

func NewTaxSummary() *TaxSummary {
	baseModel := base.NewBaseModel(constants.TableTaxSummary, hash.Medium)
	return &TaxSummary{
		BaseModel: *baseModel,
	}
}

// Helper methods
func (t *Tax) statusAt(checkTime time.Time) string {
	if !t.IsActive {
		return "inactive"
	}

	if checkTime.Before(t.ValidFrom) {
		return "scheduled"
	}

	if t.ValidUntil != nil && checkTime.After(*t.ValidUntil) {
		return "expired"
	}

	return "active"
}

func (t *Tax) ToResponse() *TaxResponse {
	return &TaxResponse{
		ID:                       t.ID,
		Code:                     t.Code,
		Name:                     t.Name,
		Description:              t.Description,
		TaxType:                  t.TaxType,
		CalculationType:          t.CalculationType,
		Rate:                     t.Rate,
		MinAmount:                t.MinAmount,
		MaxAmount:                t.MaxAmount,
		MinOrderValue:            t.MinOrderValue,
		MaxOrderValue:            t.MaxOrderValue,
		ApplicableProducts:       t.ApplicableProducts,
		ExcludedProducts:         t.ExcludedProducts,
		ApplicableCategories:     t.ApplicableCategories,
		ExcludedCategories:       t.ExcludedCategories,
		ApplicableWarehouses:     t.ApplicableWarehouses,
		ExcludedWarehouses:       t.ExcludedWarehouses,
		ApplicableStates:         t.ApplicableStates,
		ExcludedStates:           t.ExcludedStates,
		ApplicableCustomerGroups: t.ApplicableCustomerGroups,
		ExcludedCustomerGroups:   t.ExcludedCustomerGroups,
		ValidFrom:                t.ValidFrom,
		ValidUntil:               t.ValidUntil,
		IsActive:                 t.IsActive,
		Priority:                 t.Priority,
		IsStackable:              t.IsStackable,
		StackingOrder:            t.StackingOrder,
		RequiresGSTIN:            t.RequiresGSTIN,
		RequiresPAN:              t.RequiresPAN,
		IsInterState:             t.IsInterState,
		HSNCode:                  t.HSNCode,
		SACCode:                  t.SACCode,
		TaxCategory:              t.TaxCategory,
		Status:                   t.statusAt(time.Now()),
		CreatedBy:                t.CreatedBy,
		UpdatedBy:                t.UpdatedBy,
		CreatedAt:                t.CreatedAt,
		UpdatedAt:                t.UpdatedAt,
	}
}
