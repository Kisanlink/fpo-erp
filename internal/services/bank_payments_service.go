package services

import (
	"strconv"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
)

type BankPaymentsService struct {
	bankPaymentsRepo *repositories.BankPaymentsRepository
	salesRepo        *repositories.SalesRepository
	returnsRepo      *repositories.ReturnsRepository
}

func NewBankPaymentsService(bankPaymentsRepo *repositories.BankPaymentsRepository, salesRepo *repositories.SalesRepository, returnsRepo *repositories.ReturnsRepository) *BankPaymentsService {
	return &BankPaymentsService{
		bankPaymentsRepo: bankPaymentsRepo,
		salesRepo:        salesRepo,
		returnsRepo:      returnsRepo,
	}
}

// CreateBankPayment creates a new bank payment
func (s *BankPaymentsService) CreateBankPayment(req *models.CreateBankPaymentRequest) (*models.BankPaymentResponse, error) {
	// Validate request
	if err := s.validateBankPaymentRequest(req); err != nil {
		return nil, err
	}

	// Generate transaction reference
	transactionRef := "TXN" + strconv.FormatInt(time.Now().Unix(), 10)

	// Create bank payment
	payment := &models.BankPayment{
		SaleID:         req.SaleID,
		ReturnID:       req.ReturnID,
		PaymentMethod:  req.PaymentMethod,
		TransactionRef: transactionRef,
		Amount:         req.Amount,
		PaidAt:         time.Now(),
	}

	if err := s.bankPaymentsRepo.CreateBankPayment(payment); err != nil {
		return nil, err
	}

	return s.mapBankPaymentToResponse(payment), nil
}

// GetBankPayment retrieves a bank payment by ID
func (s *BankPaymentsService) GetBankPayment(id string) (*models.BankPaymentResponse, error) {
	payment, err := s.bankPaymentsRepo.GetBankPaymentByID(id)
	if err != nil {
		return nil, errors.NewNotFoundError("Bank payment not found")
	}
	return s.mapBankPaymentToResponse(payment), nil
}

// GetAllBankPayments retrieves all bank payments with pagination
func (s *BankPaymentsService) GetAllBankPayments(limit, offset int) ([]models.BankPaymentResponse, error) {
	payments, err := s.bankPaymentsRepo.GetAllBankPayments(limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []models.BankPaymentResponse
	for _, payment := range payments {
		responses = append(responses, *s.mapBankPaymentToResponse(&payment))
	}

	return responses, nil
}

// GetBankPaymentsBySaleID retrieves bank payments for a specific sale
func (s *BankPaymentsService) GetBankPaymentsBySaleID(saleID string) ([]models.BankPaymentResponse, error) {
	payments, err := s.bankPaymentsRepo.GetBankPaymentsBySaleID(saleID)
	if err != nil {
		return nil, err
	}

	var responses []models.BankPaymentResponse
	for _, payment := range payments {
		responses = append(responses, *s.mapBankPaymentToResponse(&payment))
	}

	return responses, nil
}

// GetBankPaymentsByReturnID retrieves bank payments for a specific return
func (s *BankPaymentsService) GetBankPaymentsByReturnID(returnID string) ([]models.BankPaymentResponse, error) {
	payments, err := s.bankPaymentsRepo.GetBankPaymentsByReturnID(returnID)
	if err != nil {
		return nil, err
	}

	var responses []models.BankPaymentResponse
	for _, payment := range payments {
		responses = append(responses, *s.mapBankPaymentToResponse(&payment))
	}

	return responses, nil
}

// UpdateBankPayment updates a bank payment
func (s *BankPaymentsService) UpdateBankPayment(id string, req *models.UpdateBankPaymentRequest) (*models.BankPaymentResponse, error) {
	payment, err := s.bankPaymentsRepo.GetBankPaymentByID(id)
	if err != nil {
		return nil, errors.NewNotFoundError("Bank payment not found")
	}

	// Update fields
	if req.PaymentMethod != nil {
		payment.PaymentMethod = *req.PaymentMethod
	}
	if req.Amount != nil {
		payment.Amount = *req.Amount
	}

	if err := s.bankPaymentsRepo.UpdateBankPayment(payment); err != nil {
		return nil, err
	}

	return s.mapBankPaymentToResponse(payment), nil
}

// DeleteBankPayment deletes a bank payment
func (s *BankPaymentsService) DeleteBankPayment(id string) error {
	return s.bankPaymentsRepo.DeleteBankPayment(id)
}

// Helper methods
func (s *BankPaymentsService) validateBankPaymentRequest(req *models.CreateBankPaymentRequest) error {
	if req.Amount <= 0 {
		return errors.NewNotFoundError("amount must be greater than 0")
	}
	if req.PaymentMethod == "" {
		return errors.NewNotFoundError("payment method is required")
	}
	if req.SaleID == nil && req.ReturnID == nil {
		return errors.NewNotFoundError("either sale ID or return ID is required")
	}
	if req.SaleID != nil && req.ReturnID != nil {
		return errors.NewNotFoundError("cannot have both sale ID and return ID")
	}

	// Validate sale exists if provided
	if req.SaleID != nil {
		_, err := s.salesRepo.GetSaleByID(*req.SaleID)
		if err != nil {
			return errors.NewNotFoundError("sale not found")
		}
	}

	// Validate return exists if provided
	if req.ReturnID != nil {
		_, err := s.returnsRepo.GetReturnByID(*req.ReturnID)
		if err != nil {
			return errors.NewNotFoundError("return not found")
		}
	}

	return nil
}

func (s *BankPaymentsService) mapBankPaymentToResponse(payment *models.BankPayment) *models.BankPaymentResponse {
	return &models.BankPaymentResponse{
		ID:             payment.ID,
		SaleID:         payment.SaleID,
		ReturnID:       payment.ReturnID,
		PaymentMethod:  payment.PaymentMethod,
		TransactionRef: payment.TransactionRef,
		Amount:         payment.Amount,
		PaidAt:         payment.PaidAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:      payment.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      payment.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
