package models

import (
	"time"

	"kisanlink-erp/internal/constants"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// BankPayment represents a bank payment
type BankPayment struct {
	base.BaseModel
	SaleID         *string   `gorm:"type:varchar(100)" json:"sale_id"`
	ReturnID       *string   `gorm:"type:varchar(100)" json:"return_id"`
	PaymentMethod  string    `gorm:"type:varchar(50);not null" json:"payment_method"`
	TransactionRef string    `gorm:"type:varchar(100);unique;not null" json:"transaction_ref"`
	Amount         float64   `gorm:"type:numeric(14,4);not null" json:"amount"`
	PaidAt         time.Time `gorm:"type:timestamptz;not null" json:"paid_at"`

	// Associations
	Sale   *Sale   `gorm:"foreignKey:SaleID" json:"sale,omitempty"`
	Return *Return `gorm:"foreignKey:ReturnID" json:"return,omitempty"`
}

func (BankPayment) TableName() string {
	return "bank_payments"
}

// NewBankPayment creates a new BankPayment with initialized fields
func NewBankPayment(saleID, returnID *string, paymentMethod, transactionRef string, amount float64, paidAt time.Time) *BankPayment {
	baseModel := base.NewBaseModel(constants.TableBankPayment, hash.Medium)
	return &BankPayment{
		BaseModel:      *baseModel,
		SaleID:         saleID,
		ReturnID:       returnID,
		PaymentMethod:  paymentMethod,
		TransactionRef: transactionRef,
		Amount:         amount,
		PaidAt:         paidAt,
	}
}

// BankPaymentResponse represents the API response for bank payment
type BankPaymentResponse struct {
	ID             string  `json:"id"`
	SaleID         *string `json:"sale_id"`
	ReturnID       *string `json:"return_id"`
	PaymentMethod  string  `json:"payment_method"`
	TransactionRef string  `json:"transaction_ref"`
	Amount         float64 `json:"amount"`
	PaidAt         string  `json:"paid_at"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// CreateBankPaymentRequest represents the request to create a bank payment
type CreateBankPaymentRequest struct {
	SaleID        *string `json:"sale_id"`
	ReturnID      *string `json:"return_id"`
	PaymentMethod string  `json:"payment_method" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
}

// UpdateBankPaymentRequest represents the request to update a bank payment
type UpdateBankPaymentRequest struct {
	PaymentMethod *string  `json:"payment_method,omitempty"`
	Amount        *float64 `json:"amount,omitempty"`
}
