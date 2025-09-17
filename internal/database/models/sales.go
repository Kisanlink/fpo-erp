package models

import (
	"kisanlink-erp/internal/constants"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Sale represents a sale transaction
type Sale struct {
	base.BaseModel
	WarehouseID string    `gorm:"type:varchar(100);not null" json:"warehouse_id"`
	CustomerID  *string   `gorm:"type:varchar(100)" json:"customer_id"`
	SaleDate    time.Time `gorm:"type:timestamptz;not null;default:now()" json:"sale_date"`
	TotalAmount float64   `gorm:"type:numeric(14,4);not null" json:"total_amount"`
	Status      string    `gorm:"type:varchar(20);not null" json:"status"`

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
func NewSale(warehouseID string, customerID *string, saleDate time.Time, totalAmount float64, status string) *Sale {
	baseModel := base.NewBaseModel(constants.TableSale, hash.Medium)
	return &Sale{
		BaseModel:   *baseModel,
		WarehouseID: warehouseID,
		CustomerID:  customerID,
		SaleDate:    saleDate,
		TotalAmount: totalAmount,
		Status:      status,
	}
}

// NewSaleItem creates a new SaleItem with initialized fields
func NewSaleItem(saleID, batchID string, quantity int64, sellingPrice, lineTotal float64) *SaleItem {
	baseModel := base.NewBaseModel(constants.TableSaleItem, hash.Medium)
	return &SaleItem{
		BaseModel:    *baseModel,
		SaleID:       saleID,
		BatchID:      batchID,
		Quantity:     quantity,
		SellingPrice: sellingPrice,
		LineTotal:    lineTotal,
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
	ID          string             `json:"id"`
	WarehouseID string             `json:"warehouse_id"`
	CustomerID  *string            `json:"customer_id"`
	SaleDate    string             `json:"sale_date"`
	TotalAmount float64            `json:"total_amount"`
	Status      string             `json:"status"`
	Items       []SaleItemResponse `json:"items,omitempty"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
}

// SaleItemResponse represents the API response for sale item
type SaleItemResponse struct {
	ID           string  `json:"id"`
	SaleID       string  `json:"sale_id"`
	BatchID      string  `json:"batch_id"`
	Quantity     int64   `json:"quantity"`
	SellingPrice float64 `json:"selling_price"`
	LineTotal    float64 `json:"line_total"`
	CreatedAt    string  `json:"created_at"`
}

// CreateSaleRequest represents the request to create a sale
type CreateSaleRequest struct {
	WarehouseID    string                  `json:"warehouse_id" binding:"required"`
	CustomerID     *string                 `json:"customer_id"`
	SaleDate       *string                 `json:"sale_date"`
	DiscountID     *string                 `json:"discount_id"`     // Optional discount to apply
	CustomerState  *string                 `json:"customer_state"`  // Customer's state for tax calculation
	WarehouseState *string                 `json:"warehouse_state"` // Warehouse's state for tax calculation
	CustomerGSTIN  *string                 `json:"customer_gstin"`  // Customer's GSTIN for tax calculation
	CustomerPAN    *string                 `json:"customer_pan"`    // Customer's PAN for tax calculation
	IsInterState   *bool                   `json:"is_inter_state"`  // Whether it's an inter-state transaction
	Items          []CreateSaleItemRequest `json:"items" binding:"required"`
}

// CreateSaleItemRequest represents the request to create a sale item
type CreateSaleItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int64  `json:"quantity" binding:"required,gt=0"`
	// Batch will be automatically selected based on FEFO (First Expired, First Out)
	// SellingPrice is calculated automatically from product_prices table
}

// UpdateSaleRequest represents the request to update a sale
type UpdateSaleRequest struct {
	Status *string `json:"status,omitempty"`
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
