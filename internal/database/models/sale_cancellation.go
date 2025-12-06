package models

import (
	"kisanlink-erp/internal/constants"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// CancellationType enum
const (
	CancellationTypeFull    = "full"
	CancellationTypePartial = "partial"
)

// CancellationReason enum
const (
	ReasonCustomerRequest = "customer_request"
	ReasonPaymentFailed   = "payment_failed"
	ReasonOutOfStock      = "out_of_stock"
	ReasonPricingError    = "pricing_error"
	ReasonDuplicateOrder  = "duplicate_order"
	ReasonFraudSuspected  = "fraud_suspected"
	ReasonSystemError     = "system_error"
	ReasonOther           = "other"
)

// SaleCancellation tracks cancellation events for sales
type SaleCancellation struct {
	base.BaseModel
	SaleID           string    `gorm:"type:varchar(100);not null;index" json:"sale_id"`
	CancellationType string    `gorm:"type:varchar(20);not null" json:"cancellation_type"` // full, partial
	CancelledBy      *string   `gorm:"type:varchar(100)" json:"cancelled_by"`
	Reason           string    `gorm:"type:varchar(50);not null" json:"reason"`
	ReasonDetails    *string   `gorm:"type:text" json:"reason_details"`
	CancelledAt      time.Time `gorm:"type:timestamptz;not null;default:now()" json:"cancelled_at"`

	// Financial impact
	// OriginalAmount: Sale total BEFORE this cancellation (not the initial sale total)
	// For multi-cancellation scenarios, this tracks the state at each cancellation
	OriginalAmount   float64 `gorm:"type:numeric(14,4);not null" json:"original_amount"`
	CancelledAmount  float64 `gorm:"type:numeric(14,4);not null" json:"cancelled_amount"`
	DiscountReversed float64 `gorm:"type:numeric(14,4);default:0" json:"discount_reversed"`
	TaxReversed      float64 `gorm:"type:numeric(14,4);default:0" json:"tax_reversed"`

	// Associations
	Sale  Sale                   `gorm:"foreignKey:SaleID" json:"sale,omitempty"`
	Items []SaleCancellationItem `gorm:"foreignKey:CancellationID" json:"items,omitempty"`
}

func (SaleCancellation) TableName() string {
	return "sale_cancellations"
}

// SaleCancellationItem tracks individual items in a cancellation
type SaleCancellationItem struct {
	base.BaseModel
	CancellationID    string  `gorm:"type:varchar(100);not null;index" json:"cancellation_id"`
	SaleItemID        string  `gorm:"type:varchar(100);not null" json:"sale_item_id"`
	BatchID           string  `gorm:"type:varchar(100);not null" json:"batch_id"`
	QuantityCancelled int64   `gorm:"type:bigint;not null;check:quantity_cancelled > 0" json:"quantity_cancelled"`
	RefundAmount      float64 `gorm:"type:numeric(14,4);not null" json:"refund_amount"`
	InventoryRestored bool    `gorm:"type:boolean;not null;default:false" json:"inventory_restored"`
	TransactionID     *string `gorm:"type:varchar(100)" json:"transaction_id"` // Inventory transaction ID

	// Associations
	Cancellation SaleCancellation `gorm:"foreignKey:CancellationID" json:"cancellation,omitempty"`
	SaleItem     SaleItem         `gorm:"foreignKey:SaleItemID" json:"sale_item,omitempty"`
	Batch        InventoryBatch   `gorm:"foreignKey:BatchID" json:"batch,omitempty"`
}

func (SaleCancellationItem) TableName() string {
	return "sale_cancellation_items"
}

// NewSaleCancellation creates a new SaleCancellation with initialized fields
func NewSaleCancellation(saleID, cancellationType, reason string, cancelledBy *string, reasonDetails *string, originalAmount, cancelledAmount float64) *SaleCancellation {
	baseModel := base.NewBaseModel(constants.TableSaleCancellation, hash.Medium)
	return &SaleCancellation{
		BaseModel:        *baseModel,
		SaleID:           saleID,
		CancellationType: cancellationType,
		CancelledBy:      cancelledBy,
		Reason:           reason,
		ReasonDetails:    reasonDetails,
		CancelledAt:      time.Now(),
		OriginalAmount:   originalAmount,
		CancelledAmount:  cancelledAmount,
		DiscountReversed: 0,
		TaxReversed:      0,
	}
}

// NewSaleCancellationItem creates a new SaleCancellationItem with initialized fields
// inventoryRestored indicates whether inventory was actually restored to the batch
func NewSaleCancellationItem(cancellationID, saleItemID, batchID string, quantityCancelled int64, refundAmount float64, transactionID *string, inventoryRestored bool) *SaleCancellationItem {
	baseModel := base.NewBaseModel(constants.TableSaleCancellationItem, hash.Medium)
	return &SaleCancellationItem{
		BaseModel:         *baseModel,
		CancellationID:    cancellationID,
		SaleItemID:        saleItemID,
		BatchID:           batchID,
		QuantityCancelled: quantityCancelled,
		RefundAmount:      refundAmount,
		InventoryRestored: inventoryRestored,
		TransactionID:     transactionID,
	}
}

// SaleCancellationResponse represents the API response for sale cancellation
type SaleCancellationResponse struct {
	ID               string                         `json:"id"`
	SaleID           string                         `json:"sale_id"`
	CancellationType string                         `json:"cancellation_type"`
	CancelledBy      *string                        `json:"cancelled_by,omitempty"`
	Reason           string                         `json:"reason"`
	ReasonDetails    *string                        `json:"reason_details,omitempty"`
	CancelledAt      string                         `json:"cancelled_at"`
	OriginalAmount   float64                        `json:"original_amount"`
	CancelledAmount  float64                        `json:"cancelled_amount"`
	DiscountReversed float64                        `json:"discount_reversed"`
	TaxReversed      float64                        `json:"tax_reversed"`
	Items            []SaleCancellationItemResponse `json:"items,omitempty"`
	CreatedAt        string                         `json:"created_at"`
	UpdatedAt        string                         `json:"updated_at"`
}

// SaleCancellationItemResponse represents the API response for sale cancellation item
type SaleCancellationItemResponse struct {
	ID                string  `json:"id"`
	CancellationID    string  `json:"cancellation_id"`
	SaleItemID        string  `json:"sale_item_id"`
	BatchID           string  `json:"batch_id"`
	QuantityCancelled int64   `json:"quantity_cancelled"`
	RefundAmount      float64 `json:"refund_amount"`
	InventoryRestored bool    `json:"inventory_restored"`
	TransactionID     *string `json:"transaction_id,omitempty"`
	CreatedAt         string  `json:"created_at"`
}

// CancelSaleRequest represents the request to cancel a full sale
type CancelSaleRequest struct {
	Reason              string  `json:"reason" binding:"required,oneof=customer_request payment_failed out_of_stock pricing_error duplicate_order fraud_suspected system_error other"`
	ReasonDetails       *string `json:"reason_details" binding:"omitempty,max=1000"`
	SkipInventoryReturn bool    `json:"skip_inventory_return"`
	PerformedBy         string  `json:"performed_by" binding:"required"`
}

// CancelSaleResponse represents the response after cancelling a sale
type CancelSaleResponse struct {
	Sale                 SaleResponse                  `json:"sale"`
	InventoryRestored    []InventoryRestoredItem       `json:"inventory_restored"`
	FinancialAdjustments *FinancialAdjustmentsResponse `json:"financial_adjustments,omitempty"`
	CancellationID       string                        `json:"cancellation_id"`
}

// InventoryRestoredItem represents an inventory item that was restored
type InventoryRestoredItem struct {
	BatchID          string `json:"batch_id"`
	VariantID        string `json:"variant_id"`
	QuantityRestored int64  `json:"quantity_restored"`
	TransactionID    string `json:"transaction_id"`
}

// FinancialAdjustmentsResponse represents financial adjustments made during cancellation
type FinancialAdjustmentsResponse struct {
	DiscountReversed *DiscountReversedInfo `json:"discount_reversed,omitempty"`
	TaxVoided        *TaxVoidedInfo        `json:"tax_voided,omitempty"`
}

// DiscountReversedInfo represents discount reversal information
type DiscountReversedInfo struct {
	DiscountID       string  `json:"discount_id"`
	AmountReversed   float64 `json:"amount_reversed"`
	UsageDecremented bool    `json:"usage_decremented"`
}

// TaxVoidedInfo represents tax voiding information
type TaxVoidedInfo struct {
	TaxSummaryID string  `json:"tax_summary_id"`
	AmountVoided float64 `json:"amount_voided"`
}

// CancelItemsRequest represents the request to cancel specific items in a sale
type CancelItemsRequest struct {
	Reason        string             `json:"reason" binding:"required,oneof=customer_request payment_failed out_of_stock pricing_error duplicate_order fraud_suspected system_error other"`
	ReasonDetails *string            `json:"reason_details" binding:"omitempty,max=1000"`
	PerformedBy   string             `json:"performed_by" binding:"required"`
	Items         []CancelItemDetail `json:"items" binding:"required,min=1,dive"`
}

// CancelItemDetail represents details for cancelling a specific item
type CancelItemDetail struct {
	SaleItemID string `json:"sale_item_id" binding:"required"`
	Quantity   int64  `json:"quantity" binding:"required,min=1"`
}

// CancelItemsResponse represents the response after cancelling specific items
type CancelItemsResponse struct {
	Sale                 SaleResponse                  `json:"sale"`
	ItemsCancelled       []CancelledItemInfo           `json:"items_cancelled"`
	InventoryRestored    []InventoryRestoredItem       `json:"inventory_restored"`
	FinancialAdjustments *FinancialAdjustmentsResponse `json:"financial_adjustments,omitempty"`
	CancellationID       string                        `json:"cancellation_id"`
	NewSaleTotal         float64                       `json:"new_sale_total"`
}

// CancelledItemInfo represents information about a cancelled item
type CancelledItemInfo struct {
	SaleItemID        string  `json:"sale_item_id"`
	QuantityCancelled int64   `json:"quantity_cancelled"`
	AmountRefunded    float64 `json:"amount_refunded"`
}

// GetCancellationsResponse represents the response for getting cancellation history
type GetCancellationsResponse struct {
	SaleID        string                     `json:"sale_id"`
	Cancellations []SaleCancellationResponse `json:"cancellations"`
	TotalCount    int                        `json:"total_count"`
}
