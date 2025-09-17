package repositories

import (
	"kisanlink-erp/internal/database/models"
	"gorm.io/gorm"
)

type BankPaymentsRepository struct {
	db *gorm.DB
}

func NewBankPaymentsRepository(db *gorm.DB) *BankPaymentsRepository {
	return &BankPaymentsRepository{db: db}
}

// CreateBankPayment creates a new bank payment
func (r *BankPaymentsRepository) CreateBankPayment(payment *models.BankPayment) error {
	return r.db.Create(payment).Error
}

// GetBankPaymentByID retrieves a bank payment by ID
func (r *BankPaymentsRepository) GetBankPaymentByID(id string) (*models.BankPayment, error) {
	var payment models.BankPayment
	err := r.db.First(&payment, "id = ?", id).Error
	return &payment, err
}

// GetAllBankPayments retrieves all bank payments with pagination
func (r *BankPaymentsRepository) GetAllBankPayments(limit, offset int) ([]models.BankPayment, error) {
	var payments []models.BankPayment
	err := r.db.Limit(limit).Offset(offset).Find(&payments).Error
	return payments, err
}

// GetBankPaymentsBySaleID retrieves bank payments for a specific sale
func (r *BankPaymentsRepository) GetBankPaymentsBySaleID(saleID string) ([]models.BankPayment, error) {
	var payments []models.BankPayment
	err := r.db.Where("sale_id = ?", saleID).Find(&payments).Error
	return payments, err
}

// GetBankPaymentsByReturnID retrieves bank payments for a specific return
func (r *BankPaymentsRepository) GetBankPaymentsByReturnID(returnID string) ([]models.BankPayment, error) {
	var payments []models.BankPayment
	err := r.db.Where("return_id = ?", returnID).Find(&payments).Error
	return payments, err
}

// UpdateBankPayment updates a bank payment
func (r *BankPaymentsRepository) UpdateBankPayment(payment *models.BankPayment) error {
	return r.db.Save(payment).Error
}

// DeleteBankPayment deletes a bank payment
func (r *BankPaymentsRepository) DeleteBankPayment(id string) error {
	return r.db.Delete(&models.BankPayment{}, "id = ?", id).Error
}
