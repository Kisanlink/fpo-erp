package models

import (
	"time"

	"kisanlink-erp/internal/constants"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// PurchaseOrder represents a purchase order from a collaborator
type PurchaseOrder struct {
	base.BaseModel

	// PO Identification
	PONumber        string  `gorm:"type:varchar(50);unique;not null" json:"po_number"`       // Auto-generated: PO-2025-0001
	ExternalOrderID *string `gorm:"type:varchar(100);unique;index" json:"external_order_id"` // E-commerce order ID for webhook mapping

	// Relationships
	CollaboratorID string `gorm:"type:varchar(100);not null;index" json:"collaborator_id"`
	WarehouseID    string `gorm:"type:varchar(100);not null;index" json:"warehouse_id"` // Destination warehouse

	// Dates
	OrderDate        time.Time  `gorm:"type:date;not null" json:"order_date"`
	ExpectedDelivery time.Time  `gorm:"type:date;not null" json:"expected_delivery_date"`
	ActualDelivery   *time.Time `gorm:"type:date" json:"actual_delivery_date"`

	// Status workflow
	Status string `gorm:"type:varchar(30);not null;index" json:"status"`
	// Values: "placed", "confirmed", "out_for_delivery", "delivered", "verified", "paid"

	// Financial (ALL-IN pricing - includes everything)
	TotalAmount float64 `gorm:"type:numeric(14,4);not null" json:"total_amount"` // Grand total

	// Payment tracking
	PaymentStatus string `gorm:"type:varchar(20);not null" json:"payment_status"`
	// Values: "unpaid", "partial", "paid"
	PaidAmount float64 `gorm:"type:numeric(14,4);default:0" json:"paid_amount"`

	// Inter-state flag (determined at PO creation by comparing collaborator vs warehouse state)
	IsInterState *bool `gorm:"type:boolean" json:"is_inter_state"` // nil = unknown, true = inter-state (IGST), false = intra-state (CGST+SGST)

	// Associations
	Collaborator Collaborator        `gorm:"foreignKey:CollaboratorID" json:"collaborator,omitempty"`
	Warehouse    Warehouse           `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Items        []PurchaseOrderItem `gorm:"foreignKey:POID" json:"items,omitempty"`
	GRN          *GRN                `gorm:"foreignKey:POID" json:"grn,omitempty"`
}

// NewPurchaseOrder creates a new PurchaseOrder with initialized fields
func NewPurchaseOrder(poNumber, collaboratorID, warehouseID string, orderDate, expectedDelivery time.Time) *PurchaseOrder {
	baseModel := base.NewBaseModel(constants.TablePurchaseOrder, hash.Medium)
	return &PurchaseOrder{
		BaseModel:        *baseModel,
		PONumber:         poNumber,
		CollaboratorID:   collaboratorID,
		WarehouseID:      warehouseID,
		OrderDate:        orderDate,
		ExpectedDelivery: expectedDelivery,
		Status:           "placed",
		PaymentStatus:    "unpaid",
		TotalAmount:      0,
		PaidAmount:       0,
	}
}

func (PurchaseOrder) TableName() string {
	return "purchase_orders"
}

// PurchaseOrderItem represents a line item in a purchase order
type PurchaseOrderItem struct {
	base.BaseModel

	POID      string `gorm:"type:varchar(100);not null;index" json:"po_id"`
	VariantID string `gorm:"type:varchar(100);not null;index" json:"variant_id"`

	// Quantity & Pricing (ALL-IN price)
	Quantity  int64   `gorm:"type:bigint;not null;check:quantity > 0" json:"quantity"`
	UnitPrice float64 `gorm:"type:numeric(12,4);not null" json:"unit_price"` // ALL-IN price per unit
	LineTotal float64 `gorm:"type:numeric(14,4);not null" json:"line_total"` // Quantity × UnitPrice

	// Product Details (snapshot for reference)
	ProductName *string `gorm:"type:varchar(150)" json:"product_name"` // Snapshot at PO time
	ProductSKU  *string `gorm:"type:varchar(50)" json:"product_sku"`   // Snapshot at PO time

	// Delivery tracking
	ReceivedQuantity *int64 `gorm:"type:bigint" json:"received_quantity"` // Actual received (set during GRN)

	// GST Breakdown (reverse-calculated from ALL-IN unit price)
	BasePrice  float64 `gorm:"type:numeric(14,4)" json:"base_price"`  // Price before GST
	GSTRate    float64 `gorm:"type:numeric(5,2)" json:"gst_rate"`     // GST rate used (from variant)
	GSTAmount  float64 `gorm:"type:numeric(14,4)" json:"gst_amount"`  // Total GST per unit
	CGSTRate   float64 `gorm:"type:numeric(5,2)" json:"cgst_rate"`    // 0 if inter-state
	CGSTAmount float64 `gorm:"type:numeric(14,4)" json:"cgst_amount"` // 0 if inter-state
	SGSTRate   float64 `gorm:"type:numeric(5,2)" json:"sgst_rate"`    // 0 if inter-state
	SGSTAmount float64 `gorm:"type:numeric(14,4)" json:"sgst_amount"` // 0 if inter-state
	IGSTRate   float64 `gorm:"type:numeric(5,2)" json:"igst_rate"`    // 0 if intra-state
	IGSTAmount float64 `gorm:"type:numeric(14,4)" json:"igst_amount"` // 0 if intra-state

	// Associations
	PurchaseOrder PurchaseOrder  `gorm:"foreignKey:POID" json:"purchase_order,omitempty"`
	Variant       ProductVariant `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
}

// NewPurchaseOrderItem creates a new PurchaseOrderItem with initialized fields
func NewPurchaseOrderItem(poID, variantID string, quantity int64, unitPrice float64) *PurchaseOrderItem {
	baseModel := base.NewBaseModel(constants.TablePurchaseOrderItem, hash.Medium)
	lineTotal := float64(quantity) * unitPrice
	return &PurchaseOrderItem{
		BaseModel: *baseModel,
		POID:      poID,
		VariantID: variantID,
		Quantity:  quantity,
		UnitPrice: unitPrice,
		LineTotal: lineTotal,
	}
}

func (PurchaseOrderItem) TableName() string {
	return "purchase_order_items"
}

// PurchaseOrderResponse represents the API response for purchase order
type PurchaseOrderResponse struct {
	ID                  string                      `json:"id"`
	PONumber            string                      `json:"po_number"`
	CollaboratorID      string                      `json:"collaborator_id"`
	CollaboratorName    string                      `json:"collaborator_name"`
	WarehouseID         string                      `json:"warehouse_id"`
	WarehouseName       string                      `json:"warehouse_name"`
	OrderDate           string                      `json:"order_date"`
	ExpectedDelivery    string                      `json:"expected_delivery_date"`
	ActualDelivery      *string                     `json:"actual_delivery_date"`
	Status              string                      `json:"status"`
	TotalAmount         float64                     `json:"total_amount"`
	TotalRejectedAmount float64                     `json:"total_rejected_amount"` // Total value of rejected items from GRN
	AmountOwed          float64                     `json:"amount_owed"`           // TotalAmount - TotalRejectedAmount
	PaymentStatus       string                      `json:"payment_status"`
	PaidAmount          float64                     `json:"paid_amount"`
	IsInterState        *bool                       `json:"is_inter_state"` // nil = unknown, true = inter-state, false = intra-state
	Items               []PurchaseOrderItemResponse `json:"items,omitempty"`
	CreatedAt           string                      `json:"created_at"`
	UpdatedAt           string                      `json:"updated_at"`
}

// PurchaseOrderItemResponse represents the API response for purchase order item
type PurchaseOrderItemResponse struct {
	ID               string  `json:"id"`
	POID             string  `json:"po_id"`
	VariantID        string  `json:"variant_id"`
	ProductName      *string `json:"product_name"`
	ProductSKU       *string `json:"product_sku"`
	Quantity         int64   `json:"quantity"`
	UnitPrice        float64 `json:"unit_price"`
	LineTotal        float64 `json:"line_total"`
	ReceivedQuantity *int64  `json:"received_quantity"`
	// GST Breakdown
	BasePrice  float64 `json:"base_price"`
	GSTRate    float64 `json:"gst_rate"`
	GSTAmount  float64 `json:"gst_amount"`
	CGSTRate   float64 `json:"cgst_rate,omitempty"`
	CGSTAmount float64 `json:"cgst_amount,omitempty"`
	SGSTRate   float64 `json:"sgst_rate,omitempty"`
	SGSTAmount float64 `json:"sgst_amount,omitempty"`
	IGSTRate   float64 `json:"igst_rate,omitempty"`
	IGSTAmount float64 `json:"igst_amount,omitempty"`
	CreatedAt  string  `json:"created_at"`
}

// CreatePurchaseOrderRequest represents the request to create a purchase order
type CreatePurchaseOrderRequest struct {
	CollaboratorID   string                           `json:"collaborator_id" binding:"required"`
	WarehouseID      string                           `json:"warehouse_id" binding:"required"`
	OrderDate        *string                          `json:"order_date"` // Optional, defaults to now
	ExpectedDelivery string                           `json:"expected_delivery_date" binding:"required"`
	Items            []CreatePurchaseOrderItemRequest `json:"items" binding:"required,min=1"`
}

// CreatePurchaseOrderItemRequest represents the request to create a purchase order item
type CreatePurchaseOrderItemRequest struct {
	VariantID string  `json:"variant_id" binding:"required"`
	Quantity  int64   `json:"quantity" binding:"required,gt=0"`
	UnitPrice float64 `json:"unit_price" binding:"required,gt=0"` // ALL-IN price
}

// UpdatePOStatusRequest represents the request to update purchase order status
type UpdatePOStatusRequest struct {
	Status         string     `json:"status" binding:"required"` // placed, confirmed, out_for_delivery, delivered, verified, paid
	ActualDelivery *time.Time `json:"actual_delivery_date"`      // Set when status = delivered

	// Pattern 1: Accept All (simplest - for quick processing)
	AcceptAll         *bool   `json:"accept_all"`          // If true, accept all items with default expiry
	DefaultExpiryDate *string `json:"default_expiry_date"` // Applied to all items when accept_all is true

	// Pattern 2 & 3: Per-item details
	Items []DeliveryItemRequest `json:"items"` // Optional: for auto-GRN creation
}

// DeliveryItemRequest represents item details for delivery processing
type DeliveryItemRequest struct {
	POItemID string `json:"po_item_id" binding:"required"`

	// Pattern 2: Simple Accept/Reject (for yes/no buttons)
	Accept          *bool   `json:"accept"`           // true = accept, false = reject
	RejectionReason *string `json:"rejection_reason"` // Optional reason for rejection

	// Pattern 3: Detailed Quantities (for partial acceptances)
	ReceivedQuantity *int64 `json:"received_quantity"` // Actual quantity received
	AcceptedQuantity *int64 `json:"accepted_quantity"` // Quantity accepted after inspection

	// Common fields (required for accepted items)
	ExpiryDate  string  `json:"expiry_date" binding:"required"` // Format: YYYY-MM-DD
	BatchNumber *string `json:"batch_number"`                   // Optional vendor batch number
}

// UpdatePOPaymentRequest represents the request to update payment information
type UpdatePOPaymentRequest struct {
	PaidAmount    float64 `json:"paid_amount" binding:"required,gt=0"`
	PaymentStatus string  `json:"payment_status" binding:"required"` // unpaid, partial, paid
}
