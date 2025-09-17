package models

import (
	"kisanlink-erp/internal/constants"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Return represents a return transaction
type Return struct {
	base.BaseModel
	SaleID      string    `gorm:"type:varchar(100);not null" json:"sale_id"`
	ReturnDate  time.Time `gorm:"type:timestamptz;not null;default:now()" json:"return_date"`
	TotalRefund float64   `gorm:"type:numeric(14,4);not null" json:"total_refund"`
	Status      string    `gorm:"type:varchar(20);not null" json:"status"`

	// Associations
	Sale  Sale         `gorm:"foreignKey:SaleID" json:"sale,omitempty"`
	Items []ReturnItem `gorm:"foreignKey:ReturnID" json:"items,omitempty"`
}

func (Return) TableName() string {
	return "returns"
}

// ReturnItem represents an item in a return
type ReturnItem struct {
	base.BaseModel
	ReturnID     string  `gorm:"type:varchar(100);not null" json:"return_id"`
	BatchID      string  `gorm:"type:varchar(100);not null" json:"batch_id"`
	Quantity     int64   `gorm:"type:bigint;not null;check:quantity > 0" json:"quantity"`
	RefundAmount float64 `gorm:"type:numeric(12,4);not null" json:"refund_amount"`

	// Associations
	Return Return         `gorm:"foreignKey:ReturnID" json:"return,omitempty"`
	Batch  InventoryBatch `gorm:"foreignKey:BatchID" json:"batch,omitempty"`
}

func (ReturnItem) TableName() string {
	return "return_items"
}

// ReturnSummary represents a denormalized summary of returns
type ReturnSummary struct {
	base.BaseModel
	SummaryDate  time.Time `gorm:"type:date;not null" json:"summary_date"`
	WarehouseID  string    `gorm:"type:varchar(100);not null" json:"warehouse_id"`
	TotalReturns float64   `gorm:"type:numeric(14,4);not null" json:"total_returns"`
	TotalItems   int64     `gorm:"type:bigint;not null" json:"total_items"`

	// Associations
	Warehouse Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
}

func (ReturnSummary) TableName() string {
	return "return_summaries"
}

// NewReturn creates a new Return with initialized fields
func NewReturn(saleID string, returnDate time.Time, totalRefund float64, status string) *Return {
	baseModel := base.NewBaseModel(constants.TableReturn, hash.Medium)
	return &Return{
		BaseModel:   *baseModel,
		SaleID:      saleID,
		ReturnDate:  returnDate,
		TotalRefund: totalRefund,
		Status:      status,
	}
}

// NewReturnItem creates a new ReturnItem with initialized fields
func NewReturnItem(returnID, batchID string, quantity int64, refundAmount float64) *ReturnItem {
	baseModel := base.NewBaseModel(constants.TableReturnItem, hash.Medium)
	return &ReturnItem{
		BaseModel:    *baseModel,
		ReturnID:     returnID,
		BatchID:      batchID,
		Quantity:     quantity,
		RefundAmount: refundAmount,
	}
}

// NewReturnSummary creates a new ReturnSummary with initialized fields
func NewReturnSummary(summaryDate time.Time, warehouseID string, totalReturns float64, totalItems int64) *ReturnSummary {
	baseModel := base.NewBaseModel(constants.TableReturnSummary, hash.Medium)
	return &ReturnSummary{
		BaseModel:    *baseModel,
		SummaryDate:  summaryDate,
		WarehouseID:  warehouseID,
		TotalReturns: totalReturns,
		TotalItems:   totalItems,
	}
}

// ReturnResponse represents the API response for return
type ReturnResponse struct {
	ID          string               `json:"id"`
	SaleID      string               `json:"sale_id"`
	ReturnDate  string               `json:"return_date"`
	TotalRefund float64              `json:"total_refund"`
	Status      string               `json:"status"`
	Items       []ReturnItemResponse `json:"items,omitempty"`
	CreatedAt   string               `json:"created_at"`
	UpdatedAt   string               `json:"updated_at"`
}

// ReturnItemResponse represents the API response for return item
type ReturnItemResponse struct {
	ID           string  `json:"id"`
	ReturnID     string  `json:"return_id"`
	BatchID      string  `json:"batch_id"`
	Quantity     int64   `json:"quantity"`
	RefundAmount float64 `json:"refund_amount"`
	CreatedAt    string  `json:"created_at"`
}

// CreateReturnRequest represents the request to create a return
type CreateReturnRequest struct {
	SaleID     string                    `json:"sale_id" binding:"required"`
	ReturnDate *string                   `json:"return_date"`
	Items      []CreateReturnItemRequest `json:"items" binding:"required"`
}

// CreateReturnItemRequest represents the request to create a return item
type CreateReturnItemRequest struct {
	BatchID      string  `json:"batch_id" binding:"required"`
	Quantity     int64   `json:"quantity" binding:"required,gt=0"`
	RefundAmount float64 `json:"refund_amount" binding:"required"`
}

// UpdateReturnRequest represents the request to update a return
type UpdateReturnRequest struct {
	Status *string `json:"status,omitempty"`
}

// UpdateReturnStatusRequest represents the request to update return status
type UpdateReturnStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// MostReturnedProductResponse represents the response for most returned products
type MostReturnedProductResponse struct {
	ProductID     string  `json:"product_id"`
	ProductName   string  `json:"product_name"`
	TotalReturned int     `json:"total_returned"`
	ReturnAmount  float64 `json:"return_amount"`
}

// RefundPolicy represents a refund policy
type RefundPolicy struct {
	base.BaseModel
	PolicyName    string  `gorm:"type:varchar(100);unique;not null" json:"policy_name"`
	Description   *string `gorm:"type:text" json:"description"`
	MaxDays       int     `gorm:"type:int;not null" json:"max_days"`
	RestockingFee float64 `gorm:"type:numeric(5,2);not null;default:0.00" json:"restocking_fee"`
}

func (RefundPolicy) TableName() string {
	return "refund_policy"
}

// NewRefundPolicy creates a new RefundPolicy with initialized fields
func NewRefundPolicy(policyName string, description *string, maxDays int, restockingFee float64) *RefundPolicy {
	baseModel := base.NewBaseModel(constants.TableRefundPolicy, hash.Medium)
	return &RefundPolicy{
		BaseModel:     *baseModel,
		PolicyName:    policyName,
		Description:   description,
		MaxDays:       maxDays,
		RestockingFee: restockingFee,
	}
}

// RefundPolicyResponse represents the API response for refund policy
type RefundPolicyResponse struct {
	ID            string  `json:"id"`
	PolicyName    string  `json:"policy_name"`
	Description   *string `json:"description"`
	MaxDays       int     `json:"max_days"`
	RestockingFee float64 `json:"restocking_fee"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// CreateRefundPolicyRequest represents the request to create a refund policy
type CreateRefundPolicyRequest struct {
	PolicyName    string  `json:"policy_name" binding:"required"`
	Description   *string `json:"description"`
	MaxDays       int     `json:"max_days" binding:"required"`
	RestockingFee float64 `json:"restocking_fee"`
}

// UpdateRefundPolicyRequest represents the request to update a refund policy
type UpdateRefundPolicyRequest struct {
	Description   *string  `json:"description,omitempty"`
	MaxDays       *int     `json:"max_days,omitempty"`
	RestockingFee *float64 `json:"restocking_fee,omitempty"`
}
