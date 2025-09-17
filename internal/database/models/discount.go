package models

import (
	"kisanlink-erp/internal/constants"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// DiscountType represents the type of discount
type DiscountType string

const (
	DiscountTypeFlat       DiscountType = "flat"       // Fixed amount discount
	DiscountTypePercentage DiscountType = "percentage" // Percentage discount
	DiscountTypeBuyXGetY   DiscountType = "buy_x_get_y"
	DiscountTypeFirstOrder DiscountType = "first_order"
	DiscountTypeLoyalty    DiscountType = "loyalty"
	DiscountTypeSeasonal   DiscountType = "seasonal"
	DiscountTypeBulk       DiscountType = "bulk"
	DiscountTypeReferral   DiscountType = "referral"
)

// Discount represents a discount offer
type Discount struct {
	base.BaseModel
	Code                 string       `gorm:"type:varchar(50);unique;not null" json:"code"`
	Name                 string       `gorm:"type:varchar(100);not null" json:"name"`
	Description          *string      `gorm:"type:text" json:"description"`
	DiscountType         DiscountType `gorm:"type:varchar(20);not null" json:"discount_type"`
	Value                float64      `gorm:"type:numeric(10,4);not null" json:"value"`      // Amount or percentage
	MaxDiscountAmount    *float64     `gorm:"type:numeric(10,4)" json:"max_discount_amount"` // For percentage discounts
	MinOrderValue        *float64     `gorm:"type:numeric(10,4)" json:"min_order_value"`     // Minimum order value required
	MaxOrderValue        *float64     `gorm:"type:numeric(10,4)" json:"max_order_value"`     // Maximum order value for discount
	ApplicableProducts   *string      `gorm:"type:text" json:"applicable_products"`          // JSON array of product IDs
	ExcludedProducts     *string      `gorm:"type:text" json:"excluded_products"`            // JSON array of excluded product IDs
	ApplicableCategories *string      `gorm:"type:text" json:"applicable_categories"`        // JSON array of category IDs
	ExcludedCategories   *string      `gorm:"type:text" json:"excluded_categories"`          // JSON array of excluded category IDs
	ApplicableWarehouses *string      `gorm:"type:text" json:"applicable_warehouses"`        // JSON array of warehouse IDs
	CustomerGroups       *string      `gorm:"type:text" json:"customer_groups"`              // JSON array of customer group IDs
	UsageLimit           *int         `gorm:"type:int" json:"usage_limit"`                   // Total usage limit
	UsagePerCustomer     *int         `gorm:"type:int;default:1" json:"usage_per_customer"`  // Usage limit per customer
	CurrentUsage         int          `gorm:"type:int;default:0" json:"current_usage"`       // Current usage count
	ValidFrom            time.Time    `gorm:"type:timestamptz;not null" json:"valid_from"`
	ValidUntil           time.Time    `gorm:"type:timestamptz;not null" json:"valid_until"`
	IsActive             bool         `gorm:"type:boolean;default:true" json:"is_active"`
	IsStackable          bool         `gorm:"type:boolean;default:false" json:"is_stackable"` // Can be combined with other discounts
	Priority             int          `gorm:"type:int;default:0" json:"priority"`             // Higher priority discounts applied first
	Terms                *string      `gorm:"type:text" json:"terms"`                         // Terms and conditions
}

func (Discount) TableName() string {
	return "discounts"
}

// DiscountUsage tracks discount usage by customers
type DiscountUsage struct {
	base.BaseModel
	DiscountID string    `gorm:"type:varchar(100);not null" json:"discount_id"`
	CustomerID string    `gorm:"type:varchar(100);not null" json:"customer_id"`
	SaleID     string    `gorm:"type:varchar(100);not null" json:"sale_id"`
	UsedAt     time.Time `gorm:"type:timestamptz;not null;default:now()" json:"used_at"`
	Amount     float64   `gorm:"type:numeric(10,4);not null" json:"amount"` // Actual discount amount applied

	// Associations
	Discount Discount `gorm:"foreignKey:DiscountID" json:"discount,omitempty"`
	Sale     Sale     `gorm:"foreignKey:SaleID" json:"sale,omitempty"`
}

func (DiscountUsage) TableName() string {
	return "discount_usages"
}

// NewDiscount creates a new Discount with initialized fields
func NewDiscount(code, name string, description *string, discountType DiscountType, value float64, validFrom, validUntil time.Time) *Discount {
	baseModel := base.NewBaseModel(constants.TableDiscount, hash.Medium)
	return &Discount{
		BaseModel:    *baseModel,
		Code:         code,
		Name:         name,
		Description:  description,
		DiscountType: discountType,
		Value:        value,
		ValidFrom:    validFrom,
		ValidUntil:   validUntil,
		IsActive:     true,
		IsStackable:  false,
		Priority:     0,
	}
}

// NewDiscountUsage creates a new DiscountUsage with initialized fields
func NewDiscountUsage(discountID, customerID, saleID string, amount float64) *DiscountUsage {
	baseModel := base.NewBaseModel(constants.TableDiscountUse, hash.Medium)
	return &DiscountUsage{
		BaseModel:  *baseModel,
		DiscountID: discountID,
		CustomerID: customerID,
		SaleID:     saleID,
		UsedAt:     time.Now(),
		Amount:     amount,
	}
}

// DiscountResponse represents the API response for discount
type DiscountResponse struct {
	ID                   string       `json:"id"`
	Code                 string       `json:"code"`
	Name                 string       `json:"name"`
	Description          *string      `json:"description"`
	DiscountType         DiscountType `json:"discount_type"`
	Value                float64      `json:"value"`
	MaxDiscountAmount    *float64     `json:"max_discount_amount"`
	MinOrderValue        *float64     `json:"min_order_value"`
	MaxOrderValue        *float64     `json:"max_order_value"`
	ApplicableProducts   *string      `json:"applicable_products"`
	ExcludedProducts     *string      `json:"excluded_products"`
	ApplicableCategories *string      `json:"applicable_categories"`
	ExcludedCategories   *string      `json:"excluded_categories"`
	ApplicableWarehouses *string      `json:"applicable_warehouses"`
	CustomerGroups       *string      `json:"customer_groups"`
	UsageLimit           *int         `json:"usage_limit"`
	UsagePerCustomer     *int         `json:"usage_per_customer"`
	CurrentUsage         int          `json:"current_usage"`
	ValidFrom            string       `json:"valid_from"`
	ValidUntil           string       `json:"valid_until"`
	IsActive             bool         `json:"is_active"`
	IsStackable          bool         `json:"is_stackable"`
	Priority             int          `json:"priority"`
	Terms                *string      `json:"terms"`
	Status               string       `json:"status"` // "active", "expired", "inactive", "usage_limit_reached", "scheduled"
	CreatedAt            string       `json:"created_at"`
	UpdatedAt            string       `json:"updated_at"`
}

// DiscountUsageResponse represents the API response for discount usage
type DiscountUsageResponse struct {
	ID         string  `json:"id"`
	DiscountID string  `json:"discount_id"`
	CustomerID string  `json:"customer_id"`
	SaleID     string  `json:"sale_id"`
	UsedAt     string  `json:"used_at"`
	Amount     float64 `json:"amount"`
	CreatedAt  string  `json:"created_at"`
}

// CreateDiscountRequest represents the request to create a discount
type CreateDiscountRequest struct {
	Code                 string       `json:"code" binding:"required"`
	Name                 string       `json:"name" binding:"required"`
	Description          *string      `json:"description"`
	DiscountType         DiscountType `json:"discount_type" binding:"required"`
	Value                float64      `json:"value" binding:"required,gt=0"`
	MaxDiscountAmount    *float64     `json:"max_discount_amount"`
	MinOrderValue        *float64     `json:"min_order_value"`
	MaxOrderValue        *float64     `json:"max_order_value"`
	ApplicableProducts   *string      `json:"applicable_products"`
	ExcludedProducts     *string      `json:"excluded_products"`
	ApplicableCategories *string      `json:"applicable_categories"`
	ExcludedCategories   *string      `json:"excluded_categories"`
	ApplicableWarehouses *string      `json:"applicable_warehouses"`
	CustomerGroups       *string      `json:"customer_groups"`
	UsageLimit           *int         `json:"usage_limit"`
	UsagePerCustomer     *int         `json:"usage_per_customer"`
	ValidFrom            string       `json:"valid_from" binding:"required"`
	ValidUntil           string       `json:"valid_until" binding:"required"`
	IsActive             *bool        `json:"is_active"`
	IsStackable          *bool        `json:"is_stackable"`
	Priority             *int         `json:"priority"`
	Terms                *string      `json:"terms"`
}

// UpdateDiscountRequest represents the request to update a discount
type UpdateDiscountRequest struct {
	Name                 *string  `json:"name,omitempty"`
	Description          *string  `json:"description,omitempty"`
	Value                *float64 `json:"value,omitempty"`
	MaxDiscountAmount    *float64 `json:"max_discount_amount,omitempty"`
	MinOrderValue        *float64 `json:"min_order_value,omitempty"`
	MaxOrderValue        *float64 `json:"max_order_value,omitempty"`
	ApplicableProducts   *string  `json:"applicable_products,omitempty"`
	ExcludedProducts     *string  `json:"excluded_products,omitempty"`
	ApplicableCategories *string  `json:"applicable_categories,omitempty"`
	ExcludedCategories   *string  `json:"excluded_categories,omitempty"`
	ApplicableWarehouses *string  `json:"applicable_warehouses,omitempty"`
	CustomerGroups       *string  `json:"customer_groups,omitempty"`
	UsageLimit           *int     `json:"usage_limit,omitempty"`
	UsagePerCustomer     *int     `json:"usage_per_customer,omitempty"`
	ValidFrom            *string  `json:"valid_from,omitempty"`
	ValidUntil           *string  `json:"valid_until,omitempty"`
	IsActive             *bool    `json:"is_active,omitempty"`
	IsStackable          *bool    `json:"is_stackable,omitempty"`
	Priority             *int     `json:"priority,omitempty"`
	Terms                *string  `json:"terms,omitempty"`
}

// ValidateDiscountRequest represents the request to validate a discount
type ValidateDiscountRequest struct {
	DiscountCode string   `json:"discount_code" binding:"required"`
	CustomerID   *string  `json:"customer_id"`
	OrderValue   float64  `json:"order_value" binding:"required,gt=0"`
	ProductIDs   []string `json:"product_ids"`
	WarehouseID  string   `json:"warehouse_id" binding:"required"`
}

// DiscountValidationResponse represents the response for discount validation
type DiscountValidationResponse struct {
	IsValid            bool     `json:"is_valid"`
	DiscountID         string   `json:"discount_id,omitempty"`
	DiscountCode       string   `json:"discount_code,omitempty"`
	DiscountName       string   `json:"discount_name,omitempty"`
	DiscountType       string   `json:"discount_type,omitempty"`
	Value              float64  `json:"value,omitempty"`
	MaxDiscountAmount  *float64 `json:"max_discount_amount,omitempty"`
	CalculatedDiscount float64  `json:"calculated_discount,omitempty"`
	Message            string   `json:"message"`
}

// Helpers

func (d *Discount) statusAt(t time.Time) string {
	if !d.IsActive {
		return "inactive"
	}
	if d.UsageLimit != nil && d.CurrentUsage >= *d.UsageLimit {
		return "usage_limit_reached"
	}
	if t.Before(d.ValidFrom) {
		return "scheduled"
	}
	if t.After(d.ValidUntil) {
		return "expired"
	}
	return "active"
}

func (d *Discount) ToResponse() *DiscountResponse {
	const tf = "2006-01-02T15:04:05Z07:00"
	return &DiscountResponse{
		ID:                   d.ID,
		Code:                 d.Code,
		Name:                 d.Name,
		Description:          d.Description,
		DiscountType:         d.DiscountType,
		Value:                d.Value,
		MaxDiscountAmount:    d.MaxDiscountAmount,
		MinOrderValue:        d.MinOrderValue,
		MaxOrderValue:        d.MaxOrderValue,
		ApplicableProducts:   d.ApplicableProducts,
		ExcludedProducts:     d.ExcludedProducts,
		ApplicableCategories: d.ApplicableCategories,
		ExcludedCategories:   d.ExcludedCategories,
		ApplicableWarehouses: d.ApplicableWarehouses,
		CustomerGroups:       d.CustomerGroups,
		UsageLimit:           d.UsageLimit,
		UsagePerCustomer:     d.UsagePerCustomer,
		CurrentUsage:         d.CurrentUsage,
		ValidFrom:            d.ValidFrom.Format(tf),
		ValidUntil:           d.ValidUntil.Format(tf),
		IsActive:             d.IsActive,
		IsStackable:          d.IsStackable,
		Priority:             d.Priority,
		Terms:                d.Terms,
		Status:               d.statusAt(time.Now()),
		CreatedAt:            d.CreatedAt.Format(tf),
		UpdatedAt:            d.UpdatedAt.Format(tf),
	}
}

func (du *DiscountUsage) ToResponse() *DiscountUsageResponse {
	const tf = "2006-01-02T15:04:05Z07:00"
	return &DiscountUsageResponse{
		ID:         du.ID,
		DiscountID: du.DiscountID,
		CustomerID: du.CustomerID,
		SaleID:     du.SaleID,
		UsedAt:     du.UsedAt.Format(tf),
		Amount:     du.Amount,
		CreatedAt:  du.CreatedAt.Format(tf),
	}
}
