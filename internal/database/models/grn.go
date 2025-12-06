package models

import (
	"kisanlink-erp/internal/constants"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// GRN represents a Goods Receipt Note for incoming inventory from purchase orders
type GRN struct {
	base.BaseModel

	// GRN Identification
	GRNNumber string `gorm:"type:varchar(50);unique;not null" json:"grn_number"` // User-provided from vendor PDF

	// Document Attachment
	GRNDocument *string `gorm:"type:varchar(100)" json:"grn_document"` // Attachment ID (ATT_xxxxxxxx) for vendor's GRN PDF

	// Relationships
	POID        string `gorm:"type:varchar(100);not null;index" json:"po_id"`
	WarehouseID string `gorm:"type:varchar(100);not null;index" json:"warehouse_id"`

	// Receipt Details
	ReceivedDate time.Time `gorm:"type:timestamptz;not null" json:"received_date"`
	ReceivedBy   string    `gorm:"type:varchar(100);not null" json:"received_by"` // User ID from AAA

	// Quality Inspection
	QualityStatus string `gorm:"type:varchar(20);not null" json:"quality_status"`
	// Values: "accepted", "rejected", "partial"
	Remarks *string `gorm:"type:text" json:"remarks"`

	// Associations
	PurchaseOrder PurchaseOrder `gorm:"foreignKey:POID" json:"purchase_order,omitempty"`
	Warehouse     Warehouse     `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Items         []GRNItem     `gorm:"foreignKey:GRNID" json:"items,omitempty"`
}

// NewGRN creates a new GRN with initialized fields
func NewGRN(grnNumber, poID, warehouseID, receivedBy string, receivedDate time.Time, qualityStatus string) *GRN {
	baseModel := base.NewBaseModel(constants.TableGRN, hash.Medium)
	return &GRN{
		BaseModel:     *baseModel,
		GRNNumber:     grnNumber,
		POID:          poID,
		WarehouseID:   warehouseID,
		ReceivedDate:  receivedDate,
		ReceivedBy:    receivedBy,
		QualityStatus: qualityStatus,
	}
}

func (GRN) TableName() string {
	return "goods_receipt_notes"
}

// GRNItem represents a line item in a goods receipt note
type GRNItem struct {
	base.BaseModel

	GRNID     string `gorm:"type:varchar(100);not null;index" json:"grn_id"`
	POItemID  string `gorm:"type:varchar(100);not null;index" json:"po_item_id"`
	VariantID string `gorm:"type:varchar(100);not null;index" json:"variant_id"`

	// Quantity tracking
	OrderedQuantity  int64 `gorm:"type:bigint;not null" json:"ordered_quantity"`
	ReceivedQuantity int64 `gorm:"type:bigint;not null" json:"received_quantity"`
	AcceptedQuantity int64 `gorm:"type:bigint;not null" json:"accepted_quantity"`
	RejectedQuantity int64 `gorm:"type:bigint;default:0" json:"rejected_quantity"`

	// Return tracking (for rejected items)
	ReturnStatus       *string    `gorm:"type:varchar(30)" json:"return_status"`        // pending, sent, received_by_vendor, closed
	ReturnSentDate     *time.Time `gorm:"type:timestamptz" json:"return_sent_date"`     // When shipped to vendor
	ReturnReceivedDate *time.Time `gorm:"type:timestamptz" json:"return_received_date"` // When vendor confirmed receipt
	ReturnClosedDate   *time.Time `gorm:"type:timestamptz" json:"return_closed_date"`   // When return process closed
	ReturnRemarks      *string    `gorm:"type:text" json:"return_remarks"`              // Notes about return

	// Batch tracking
	ExpiryDate  time.Time `gorm:"type:date;not null" json:"expiry_date"`
	BatchNumber *string   `gorm:"type:varchar(50)" json:"batch_number"` // Vendor's batch number

	// Link to created inventory
	InventoryBatchID *string `gorm:"type:varchar(100)" json:"inventory_batch_id"` // Created batch ID

	// Associations
	GRN               GRN               `gorm:"foreignKey:GRNID" json:"grn,omitempty"`
	PurchaseOrderItem PurchaseOrderItem `gorm:"foreignKey:POItemID" json:"po_item,omitempty"`
	Variant           ProductVariant    `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
	InventoryBatch    *InventoryBatch   `gorm:"foreignKey:InventoryBatchID" json:"inventory_batch,omitempty"`
}

// NewGRNItem creates a new GRNItem with initialized fields
func NewGRNItem(grnID, poItemID, variantID string, orderedQty, receivedQty, acceptedQty int64, expiryDate time.Time) *GRNItem {
	baseModel := base.NewBaseModel(constants.TableGRNItem, hash.Medium)
	rejectedQty := receivedQty - acceptedQty
	if rejectedQty < 0 {
		rejectedQty = 0
	}
	return &GRNItem{
		BaseModel:        *baseModel,
		GRNID:            grnID,
		POItemID:         poItemID,
		VariantID:        variantID,
		OrderedQuantity:  orderedQty,
		ReceivedQuantity: receivedQty,
		AcceptedQuantity: acceptedQty,
		RejectedQuantity: rejectedQty,
		ExpiryDate:       expiryDate,
	}
}

func (GRNItem) TableName() string {
	return "grn_items"
}

// GRNResponse represents the API response for goods receipt note
type GRNResponse struct {
	ID            string            `json:"id"`
	GRNNumber     string            `json:"grn_number"`
	GRNDocument   *string           `json:"grn_document"` // Attachment ID for vendor's GRN PDF
	POID          string            `json:"po_id"`
	PONumber      string            `json:"po_number"`
	WarehouseID   string            `json:"warehouse_id"`
	WarehouseName string            `json:"warehouse_name"`
	ReceivedDate  string            `json:"received_date"`
	ReceivedBy    string            `json:"received_by"`
	QualityStatus string            `json:"quality_status"`
	Remarks       *string           `json:"remarks"`
	Items         []GRNItemResponse `json:"items,omitempty"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     string            `json:"updated_at"`
}

// GRNItemResponse represents the API response for grn item
type GRNItemResponse struct {
	ID               string  `json:"id"`
	GRNID            string  `json:"grn_id"`
	POItemID         string  `json:"po_item_id"`
	VariantID        string  `json:"variant_id"`
	ProductName      string  `json:"product_name"`
	ProductSKU       string  `json:"product_sku"`
	OrderedQuantity  int64   `json:"ordered_quantity"`
	ReceivedQuantity int64   `json:"received_quantity"`
	AcceptedQuantity int64   `json:"accepted_quantity"`
	RejectedQuantity int64   `json:"rejected_quantity"`
	ExpiryDate       string  `json:"expiry_date"`
	BatchNumber      *string `json:"batch_number"`
	InventoryBatchID *string `json:"inventory_batch_id"`
	// Return tracking fields
	ReturnStatus       *string `json:"return_status,omitempty"`
	ReturnSentDate     *string `json:"return_sent_date,omitempty"`
	ReturnReceivedDate *string `json:"return_received_date,omitempty"`
	ReturnClosedDate   *string `json:"return_closed_date,omitempty"`
	ReturnRemarks      *string `json:"return_remarks,omitempty"`
	CreatedAt          string  `json:"created_at"`
}

// CreateGRNRequest represents the request to create a goods receipt note
type CreateGRNRequest struct {
	GRNNumber     string                 `json:"grn_number" binding:"required"` // User-provided from vendor PDF
	POID          string                 `json:"po_id" binding:"required"`
	ReceivedDate  *string                `json:"received_date"` // Optional, defaults to now
	ReceivedBy    string                 `json:"received_by" binding:"required"`
	QualityStatus string                 `json:"quality_status" binding:"required"` // accepted, rejected, partial
	Remarks       *string                `json:"remarks"`
	Items         []CreateGRNItemRequest `json:"items" binding:"required,min=1"`
}

// CreateGRNItemRequest represents the request to create a grn item
type CreateGRNItemRequest struct {
	POItemID         string  `json:"po_item_id" binding:"required"`
	ReceivedQuantity int64   `json:"received_quantity" binding:"required,gt=0"`
	AcceptedQuantity int64   `json:"accepted_quantity" binding:"required,gte=0"`
	RejectedQuantity int64   `json:"rejected_quantity" binding:"gte=0"`
	ExpiryDate       string  `json:"expiry_date" binding:"required"` // Format: YYYY-MM-DD
	BatchNumber      *string `json:"batch_number"`
}

// UpdateGRNRequest represents the request to update a GRN
type UpdateGRNRequest struct {
	GRNDocument   *string `json:"grn_document,omitempty"` // Attachment ID for vendor's GRN PDF
	Remarks       *string `json:"remarks,omitempty"`
	QualityStatus *string `json:"quality_status,omitempty"` // accepted, rejected, partial
}

// UpdateItemReturnStatusRequest represents the request to update return status of a rejected item
type UpdateItemReturnStatusRequest struct {
	ReturnStatus  string  `json:"return_status" binding:"required"` // pending, sent, received_by_vendor, closed
	ReturnRemarks *string `json:"return_remarks"`                   // Optional notes
}

// RejectedItemsResponse represents the response for rejected items endpoint
type RejectedItemsResponse struct {
	GRNID                string                    `json:"grn_id"`
	GRNNumber            string                    `json:"grn_number"`
	POID                 string                    `json:"po_id"`
	PONumber             string                    `json:"po_number"`
	RejectedItems        []RejectedItemDetail      `json:"rejected_items"`
	TotalRejectedValue   float64                   `json:"total_rejected_value"`
	ReturnStatusBreakdown map[string]int            `json:"return_status_breakdown"`
}

// RejectedItemDetail represents a single rejected item with return tracking
type RejectedItemDetail struct {
	ID                 string  `json:"id"`
	VariantID          string  `json:"variant_id"`
	ProductName        string  `json:"product_name"`
	ProductSKU         string  `json:"product_sku"`
	RejectedQuantity   int64   `json:"rejected_quantity"`
	UnitPrice          float64 `json:"unit_price"`
	TotalValue         float64 `json:"total_value"`
	ReturnStatus       *string `json:"return_status"`
	ReturnSentDate     *string `json:"return_sent_date,omitempty"`
	ReturnReceivedDate *string `json:"return_received_date,omitempty"`
	ReturnClosedDate   *string `json:"return_closed_date,omitempty"`
	ReturnRemarks      *string `json:"return_remarks,omitempty"`
}
