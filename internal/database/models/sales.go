package models

import (
	"fmt"
	"kisanlink-erp/internal/constants"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Sale status constants
const (
	SaleStatusPending   = "pending"
	SaleStatusCompleted = "completed"
	SaleStatusCancelled = "cancelled"
)

// IsReservationStatus returns true if the status means inventory is reserved (not yet deducted)
// Pending sales have inventory reserved; completed/cancelled sales have inventory deducted or released
func IsReservationStatus(status string) bool {
	return status == SaleStatusPending
}

// ValidStatusTransitions defines allowed status transitions for sales
var ValidStatusTransitions = map[string][]string{
	SaleStatusPending:   {SaleStatusCompleted, SaleStatusCancelled},
	SaleStatusCompleted: {SaleStatusCancelled},
	SaleStatusCancelled: {}, // Terminal state - no transitions allowed
}

// ValidateStatusTransition checks if a status transition is allowed
func ValidateStatusTransition(fromStatus, toStatus string) error {
	allowedTargets, exists := ValidStatusTransitions[fromStatus]
	if !exists {
		return fmt.Errorf("unknown current status: %s", fromStatus)
	}

	for _, allowed := range allowedTargets {
		if toStatus == allowed {
			return nil
		}
	}

	return fmt.Errorf("invalid status transition from '%s' to '%s'", fromStatus, toStatus)
}

// Sale represents a sale transaction
type Sale struct {
	base.BaseModel
	WarehouseID   string    `gorm:"type:varchar(100);not null" json:"warehouse_id"`
	InvoiceNumber string    `gorm:"type:varchar(10);uniqueIndex" json:"invoice_number"` // Format: MMYYNNNN (e.g., 12250001)
	SaleDate      time.Time `gorm:"type:timestamptz;not null;default:now()" json:"sale_date"`
	TotalAmount   float64   `gorm:"type:numeric(14,4);not null" json:"total_amount"`
	Status        string    `gorm:"type:varchar(20);not null" json:"status"`

	// BRD Requirements - Customer tracking
	CustomerPhone *string `gorm:"type:varchar(20)" json:"customer_phone"`                   // Customer phone number
	CustomerName  *string `gorm:"type:varchar(255)" json:"customer_name"`                   // Customer name
	IsOrgMember   bool    `gorm:"type:boolean;not null;default:false" json:"is_org_member"` // Whether customer belongs to the FPO/organization
	PaymentMode   string  `gorm:"type:varchar(20);not null" json:"payment_mode"`            // cash, upi, online
	SaleType      string  `gorm:"type:varchar(20);not null" json:"sale_type"`               // in_store, delivery
	ApplyTaxes    bool    `gorm:"type:boolean;not null;default:false" json:"apply_taxes"`   // Controls tax calculation for this sale

	// Cancellation fields
	CancelledAt        *time.Time `gorm:"type:timestamptz" json:"cancelled_at,omitempty"`
	CancellationReason *string    `gorm:"type:varchar(50)" json:"cancellation_reason,omitempty"`

	// Associations
	Warehouse Warehouse  `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Items     []SaleItem `gorm:"foreignKey:SaleID" json:"items,omitempty"`
}

func (Sale) TableName() string {
	return "sales"
}

// SaleItem represents an item in a sale
type SaleItem struct {
	base.BaseModel
	SaleID       string  `gorm:"type:varchar(100);not null" json:"sale_id"`
	BatchID      string  `gorm:"type:varchar(100);not null" json:"batch_id"`
	Quantity     int64   `gorm:"type:bigint;not null;check:quantity > 0" json:"quantity"`
	SellingPrice float64 `gorm:"type:numeric(12,4);not null" json:"selling_price"`
	LineTotal    float64 `gorm:"type:numeric(14,4);not null" json:"line_total"`

	// BRD Requirements - Cost and Margin tracking
	CostPrice float64 `gorm:"type:numeric(12,4);not null" json:"cost_price"` // Purchase price from batch
	Margin    float64 `gorm:"type:numeric(12,4);not null" json:"margin"`     // SellingPrice - CostPrice

	// Tax amounts (calculated from variant GST rate)
	CGSTAmount     float64 `gorm:"type:numeric(12,4);default:0" json:"cgst_amount"`
	SGSTAmount     float64 `gorm:"type:numeric(12,4);default:0" json:"sgst_amount"`
	IGSTAmount     float64 `gorm:"type:numeric(12,4);default:0" json:"igst_amount"` // For inter-state sales
	TotalTaxAmount float64 `gorm:"type:numeric(12,4);default:0" json:"total_tax_amount"`

	// Associations
	Sale  Sale           `gorm:"foreignKey:SaleID" json:"sale,omitempty"`
	Batch InventoryBatch `gorm:"foreignKey:BatchID" json:"batch,omitempty"`
}

func (SaleItem) TableName() string {
	return "sale_items"
}

// SaleSummary represents a denormalized summary of sales
type SaleSummary struct {
	base.BaseModel
	SummaryDate time.Time `gorm:"type:date;not null" json:"summary_date"`
	WarehouseID string    `gorm:"type:varchar(100);not null" json:"warehouse_id"`
	TotalSales  float64   `gorm:"type:numeric(14,4);not null" json:"total_sales"`
	TotalItems  int64     `gorm:"type:bigint;not null" json:"total_items"`

	// Associations
	Warehouse Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
}

func (SaleSummary) TableName() string {
	return "sale_summaries"
}

// NewSale creates a new Sale with initialized fields
func NewSale(warehouseID, invoiceNumber string, saleDate time.Time, totalAmount float64, status string, customerPhone, customerName *string, isOrgMember bool, paymentMode, saleType string, applyTaxes bool) *Sale {
	baseModel := base.NewBaseModel(constants.TableSale, hash.Medium)
	return &Sale{
		BaseModel:     *baseModel,
		WarehouseID:   warehouseID,
		InvoiceNumber: invoiceNumber,
		SaleDate:      saleDate,
		TotalAmount:   totalAmount,
		Status:        status,
		CustomerPhone: customerPhone,
		CustomerName:  customerName,
		IsOrgMember:   isOrgMember,
		PaymentMode:   paymentMode,
		SaleType:      saleType,
		ApplyTaxes:    applyTaxes,
	}
}

// NewSaleItem creates a new SaleItem with initialized fields
func NewSaleItem(saleID, batchID string, quantity int64, sellingPrice, costPrice, lineTotal float64) *SaleItem {
	baseModel := base.NewBaseModel(constants.TableSaleItem, hash.Medium)
	margin := sellingPrice - costPrice
	return &SaleItem{
		BaseModel:      *baseModel,
		SaleID:         saleID,
		BatchID:        batchID,
		Quantity:       quantity,
		SellingPrice:   sellingPrice,
		CostPrice:      costPrice,
		Margin:         margin,
		LineTotal:      lineTotal,
		CGSTAmount:     0,
		SGSTAmount:     0,
		IGSTAmount:     0,
		TotalTaxAmount: 0,
	}
}

// NewSaleItemWithTax creates a new SaleItem with GST tax amounts
// For intra-state: cgstAmount and sgstAmount are set, igstAmount is 0
// For inter-state: igstAmount is set, cgstAmount and sgstAmount are 0
func NewSaleItemWithTax(saleID, batchID string, quantity int64, sellingPrice, costPrice, lineTotal, cgstAmount, sgstAmount, igstAmount float64) *SaleItem {
	baseModel := base.NewBaseModel(constants.TableSaleItem, hash.Medium)
	margin := sellingPrice - costPrice
	totalTaxAmount := cgstAmount + sgstAmount + igstAmount
	return &SaleItem{
		BaseModel:      *baseModel,
		SaleID:         saleID,
		BatchID:        batchID,
		Quantity:       quantity,
		SellingPrice:   sellingPrice,
		CostPrice:      costPrice,
		Margin:         margin,
		LineTotal:      lineTotal,
		CGSTAmount:     cgstAmount,
		SGSTAmount:     sgstAmount,
		IGSTAmount:     igstAmount,
		TotalTaxAmount: totalTaxAmount,
	}
}

// NewSaleSummary creates a new SaleSummary with initialized fields
func NewSaleSummary(summaryDate time.Time, warehouseID string, totalSales float64, totalItems int64) *SaleSummary {
	baseModel := base.NewBaseModel(constants.TableSaleSummary, hash.Medium)
	return &SaleSummary{
		BaseModel:   *baseModel,
		SummaryDate: summaryDate,
		WarehouseID: warehouseID,
		TotalSales:  totalSales,
		TotalItems:  totalItems,
	}
}

// SaleResponse represents the API response for sale
type SaleResponse struct {
	ID            string  `json:"id"`
	InvoiceNumber string  `json:"invoice_number"` // Format: MMYYNNNN (e.g., 12250001)
	WarehouseID   string  `json:"warehouse_id"`
	SaleDate      string  `json:"sale_date"`
	TotalAmount   float64 `json:"total_amount"`
	Status        string  `json:"status"`

	// BRD Requirements - Customer tracking
	CustomerPhone *string `json:"customer_phone,omitempty"`
	CustomerName  *string `json:"customer_name,omitempty"`
	IsOrgMember   bool    `json:"is_org_member"`
	PaymentMode   string  `json:"payment_mode"`
	SaleType      string  `json:"sale_type"`
	ApplyTaxes    bool    `json:"apply_taxes"`

	// Cancellation fields
	CancelledAt        *string `json:"cancelled_at,omitempty"`
	CancellationReason *string `json:"cancellation_reason,omitempty"`

	Items     []SaleItemResponse `json:"items,omitempty"`
	Breakdown *SaleBreakdown     `json:"breakdown,omitempty"` // Detailed breakdown of amounts
	CreatedAt string             `json:"created_at"`
	UpdatedAt string             `json:"updated_at"`
}

// SaleListResponse represents the API response for sale list (without items for performance)
// Use GET /api/v1/sales/{id} to get full details with items
type SaleListResponse struct {
	ID            string  `json:"id"`
	InvoiceNumber string  `json:"invoice_number"`
	WarehouseID   string  `json:"warehouse_id"`
	SaleDate      string  `json:"sale_date"`
	TotalAmount   float64 `json:"total_amount"`
	Status        string  `json:"status"`

	// BRD Requirements - Customer tracking
	CustomerPhone *string `json:"customer_phone,omitempty"`
	CustomerName  *string `json:"customer_name,omitempty"`
	IsOrgMember   bool    `json:"is_org_member"`
	PaymentMode   string  `json:"payment_mode"`
	SaleType      string  `json:"sale_type"`
	ApplyTaxes    bool    `json:"apply_taxes"`

	// Cancellation fields
	CancelledAt        *string `json:"cancelled_at,omitempty"`
	CancellationReason *string `json:"cancellation_reason,omitempty"`

	// Note: Items and Breakdown are omitted for performance
	// Use GET /api/v1/sales/{id} for full details

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// SaleBreakdown represents detailed breakdown of sale calculations
type SaleBreakdown struct {
	BaseAmount       float64               `json:"base_amount"`       // Total before discounts and taxes
	AppliedDiscounts []DiscountApplication `json:"applied_discounts"` // All discounts applied
	DiscountAmount   float64               `json:"discount_amount"`   // Total discount amount
	TaxBreakdown     *TaxSummaryBreakdown  `json:"tax_breakdown"`     // Tax details
	TaxAmount        float64               `json:"tax_amount"`        // Total tax amount
	TotalSavings     float64               `json:"total_savings"`     // Total discount savings
	FinalAmount      float64               `json:"final_amount"`      // Final amount after discounts and taxes
}

// DiscountApplication represents an applied discount in the breakdown
type DiscountApplication struct {
	DiscountID   string  `json:"discount_id"`
	DiscountCode string  `json:"discount_code"`
	DiscountName string  `json:"discount_name"`
	DiscountType string  `json:"discount_type"`
	Amount       float64 `json:"amount"`
	AppliedBy    string  `json:"applied_by"` // "manual", "coupon", "auto"
}

// TaxSummaryBreakdown represents GST tax breakdown details
// For intra-state: CGSTAmount and SGSTAmount are set, IGSTAmount is 0
// For inter-state: IGSTAmount is set, CGSTAmount and SGSTAmount are 0
type TaxSummaryBreakdown struct {
	CGSTAmount     float64 `json:"cgst_amount"`
	SGSTAmount     float64 `json:"sgst_amount"`
	IGSTAmount     float64 `json:"igst_amount"`
	TotalTaxAmount float64 `json:"total_tax_amount"`
	IsInterState   bool    `json:"is_inter_state"` // true if IGST applies, false if CGST+SGST
}

// SaleItemResponse represents the API response for sale item
type SaleItemResponse struct {
	ID           string  `json:"id"`
	SaleID       string  `json:"sale_id"`
	BatchID      string  `json:"batch_id"`
	Quantity     int64   `json:"quantity"`
	SellingPrice float64 `json:"selling_price"`
	LineTotal    float64 `json:"line_total"`

	// BRD Requirements - Cost and Margin
	CostPrice float64 `json:"cost_price"` // Purchase price
	Margin    float64 `json:"margin"`     // Profit margin

	// GST Tax amounts
	CGSTAmount     float64 `json:"cgst_amount"`
	SGSTAmount     float64 `json:"sgst_amount"`
	IGSTAmount     float64 `json:"igst_amount"` // For inter-state sales
	TotalTaxAmount float64 `json:"total_tax_amount"`

	CreatedAt string `json:"created_at"`
}

// CreateSaleRequest represents the request to create a sale
type CreateSaleRequest struct {
	WarehouseID string  `json:"warehouse_id" binding:"required"`
	SaleDate    *string `json:"sale_date"`

	// BRD Requirements - Customer tracking
	CustomerPhone *string `json:"customer_phone"`                  // Customer phone number
	CustomerName  *string `json:"customer_name"`                   // Customer name
	IsOrgMember   bool    `json:"is_org_member"`                   // Whether customer belongs to the FPO/organization
	PaymentMode   string  `json:"payment_mode" binding:"required"` // cash, upi, online
	SaleType      string  `json:"sale_type" binding:"required"`    // in_store, delivery
	ApplyTaxes    *bool   `json:"apply_taxes"`                     // Controls tax calculation (default: false)

	// GST Inter-state detection
	DeliveryState *string `json:"delivery_state"` // Optional: state code for delivery (for inter-state IGST calculation)

	DiscountID         *string                 `json:"discount_id"`          // Manual discount by ID (highest priority)
	CouponCode         *string                 `json:"coupon_code"`          // Manual discount by code (second priority)
	AutoApplyDiscounts *bool                   `json:"auto_apply_discounts"` // Enable automatic discount discovery (default: true)
	Items              []CreateSaleItemRequest `json:"items" binding:"required"`
}

// CreateSaleItemRequest represents the request to create a sale item
type CreateSaleItemRequest struct {
	VariantID string `json:"variant_id" binding:"required"` // Changed from product_id to variant_id
	Quantity  int64  `json:"quantity" binding:"required,gt=0"`
	// Batch will be automatically selected based on FEFO (First Expired, First Out)
	// SellingPrice is calculated automatically from product_prices table (by variant_id)
}

// UpdateSaleRequest represents the request to update a sale
type UpdateSaleRequest struct {
	Status      *string `json:"status,omitempty"`
	PerformedBy string  `json:"-"` // Set by handler from JWT context, not from request body
}

// UpdateSaleStatusRequest represents the request to update sale status
type UpdateSaleStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// TopSellingProductResponse represents the response for top selling products
type TopSellingProductResponse struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	TotalSold   int     `json:"total_sold"`
	TotalAmount float64 `json:"total_amount"`
}
