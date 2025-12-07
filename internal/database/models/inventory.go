package models

import (
	"kisanlink-erp/internal/constants"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Transaction type constants for inventory movements
const (
	TransactionTypeReservation        = "reservation"         // Stock reserved for pending sale
	TransactionTypeReservationRelease = "reservation_release" // Pending sale cancelled, reservation released
	TransactionTypeSale               = "sale"                // Sale completed, stock deducted
	TransactionTypeCancellationReturn = "cancellation_return" // Completed sale cancelled, stock restored
	TransactionTypePurchase           = "purchase"            // GRN received, stock added
	TransactionTypeImport             = "import"              // Manual batch creation
)

// InventoryBatch represents a batch of inventory with specific cost and expiry
// Note: Tax rates are now on ProductVariant, not on batch level
type InventoryBatch struct {
	base.BaseModel
	WarehouseID      string    `gorm:"type:varchar(100);not null" json:"warehouse_id"`
	VariantID        string    `gorm:"type:varchar(100);not null" json:"variant_id"`
	CostPrice        float64   `gorm:"type:numeric(12,4);not null" json:"cost_price"`
	ExpiryDate       time.Time `gorm:"type:date;not null" json:"expiry_date"`
	TotalQuantity    int64     `gorm:"type:bigint;not null;check:total_quantity >= 0" json:"total_quantity"`
	ReservedQuantity int64     `gorm:"type:bigint;not null;default:0;check:reserved_quantity >= 0 AND reserved_quantity <= total_quantity" json:"reserved_quantity"`

	// Associations
	Warehouse Warehouse      `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Variant   ProductVariant `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
}

func (InventoryBatch) TableName() string {
	return "inventory_batches"
}

// AvailableQuantity returns the actual sellable quantity (total - reserved)
func (b *InventoryBatch) AvailableQuantity() int64 {
	return b.TotalQuantity - b.ReservedQuantity
}

// InventoryTransaction represents stock movements
type InventoryTransaction struct {
	base.BaseModel
	BatchID         string    `gorm:"type:varchar(100);not null" json:"batch_id"`
	TransactionType string    `gorm:"type:varchar(30);not null" json:"transaction_type"`
	QuantityChange  int64     `gorm:"type:bigint;not null" json:"quantity_change"`
	RelatedEntityID *string   `gorm:"type:varchar(100)" json:"related_entity_id"`
	PerformedBy     *string   `gorm:"type:varchar(100)" json:"performed_by"`
	Note            *string   `gorm:"type:text" json:"note"`
	OccurredAt      time.Time `gorm:"type:timestamptz;not null;default:now()" json:"occurred_at"`

	// Associations
	Batch InventoryBatch `gorm:"foreignKey:BatchID" json:"batch,omitempty"`
}

func (InventoryTransaction) TableName() string {
	return "inventory_transactions"
}

// InventoryBatchResponse represents the API response for inventory batch
// Note: Tax rates are retrieved from the associated ProductVariant
type InventoryBatchResponse struct {
	ID                string  `json:"id"`
	WarehouseID       string  `json:"warehouse_id"`
	VariantID         string  `json:"variant_id"`
	CostPrice         float64 `json:"cost_price"`
	ExpiryDate        string  `json:"expiry_date"`
	TotalQuantity     int64   `json:"total_quantity"`
	ReservedQuantity  int64   `json:"reserved_quantity"`
	AvailableQuantity int64   `json:"available_quantity"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ProductAvailabilityResponse represents the API response for product availability across warehouses
// Note: Tax rates (HSNCode, GSTRate) are retrieved from the associated ProductVariant
type ProductAvailabilityResponse struct {
	ID                 string       `json:"id"`
	WarehouseID        string       `json:"warehouse_id"`
	WarehouseName      string       `json:"warehouse_name"`
	Address            *AddressInfo `json:"address,omitempty"` // Embedded address info from AAA
	VariantID          string       `json:"variant_id"`
	ProductSKU         string       `json:"product_sku"`
	ProductName        string       `json:"product_name"`
	ProductDescription *string      `json:"product_description"`
	CostPrice          float64      `json:"cost_price"`
	ExpiryDate         string       `json:"expiry_date"`
	TotalQuantity      int64        `json:"total_quantity"`
	ReservedQuantity   int64        `json:"reserved_quantity"`
	AvailableQuantity  int64        `json:"available_quantity"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// InventoryTransactionResponse represents the API response for inventory transaction
type InventoryTransactionResponse struct {
	ID              string  `json:"id"`
	BatchID         string  `json:"batch_id"`
	TransactionType string  `json:"transaction_type"`
	QuantityChange  int64   `json:"quantity_change"`
	RelatedEntityID *string `json:"related_entity_id"`
	PerformedBy     *string `json:"performed_by"`
	Note            *string `json:"note"`
	OccurredAt      string  `json:"occurred_at"`
}

// CreateInventoryBatchRequest represents the request to create an inventory batch
// Note: Tax rates are now on ProductVariant (GSTRate, HSNCode), not on batch level
type CreateInventoryBatchRequest struct {
	WarehouseID string  `json:"warehouse_id" binding:"required"`
	VariantID   string  `json:"variant_id" binding:"required"`
	CostPrice   float64 `json:"cost_price" binding:"required,gt=0"`
	ExpiryDate  string  `json:"expiry_date" binding:"required"`
	Quantity    int64   `json:"quantity" binding:"required,gt=0"`
}

// CreateInventoryTransactionRequest represents the request to create an inventory transaction
type CreateInventoryTransactionRequest struct {
	TransactionType string  `json:"transaction_type" binding:"required"`
	QuantityChange  int64   `json:"quantity_change" binding:"required"`
	RelatedEntityID *string `json:"related_entity_id"`
	Note            *string `json:"note"`
}

// NewInventoryBatch creates a new InventoryBatch with initialized fields
// Note: Tax configuration is now on ProductVariant (GSTRate, HSNCode), not on batch level
func NewInventoryBatch(warehouseID, variantID string, costPrice float64, expiryDate time.Time, totalQuantity int64) *InventoryBatch {
	baseModel := base.NewBaseModel(constants.TableBatch, hash.Medium)
	return &InventoryBatch{
		BaseModel:        *baseModel,
		WarehouseID:      warehouseID,
		VariantID:        variantID,
		CostPrice:        costPrice,
		ExpiryDate:       expiryDate,
		TotalQuantity:    totalQuantity,
		ReservedQuantity: 0, // New batches have no reservations
	}
}

// NewInventoryTransaction creates a new InventoryTransaction with initialized fields
func NewInventoryTransaction(batchID, transactionType string, quantityChange int64, relatedEntityID *string, performedBy *string, note *string, occurredAt time.Time) *InventoryTransaction {
	baseModel := base.NewBaseModel(constants.TableTransaction, hash.Medium)
	return &InventoryTransaction{
		BaseModel:       *baseModel,
		BatchID:         batchID,
		TransactionType: transactionType,
		QuantityChange:  quantityChange,
		RelatedEntityID: relatedEntityID,
		PerformedBy:     performedBy,
		Note:            note,
		OccurredAt:      occurredAt,
	}
}
