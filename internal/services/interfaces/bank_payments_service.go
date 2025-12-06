package interfaces

import (
	"kisanlink-erp/internal/database/models"
)

type BankPaymentsServiceInterface interface {
	CreateBankPayment(req *models.CreateBankPaymentRequest) (*models.BankPaymentResponse, error)
	GetBankPayment(id string) (*models.BankPaymentResponse, error)
	GetAllBankPayments(limit, offset int) ([]models.BankPaymentResponse, int64, error)
	GetBankPaymentsBySaleID(saleID string) ([]models.BankPaymentResponse, error)
	GetBankPaymentsByReturnID(returnID string) ([]models.BankPaymentResponse, error)
	UpdateBankPayment(id string, req *models.UpdateBankPaymentRequest) (*models.BankPaymentResponse, error)
	DeleteBankPayment(id string) error
}
