package services

import (
	"strconv"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"

	"go.uber.org/zap"
)

type BankPaymentsService struct {
	bankPaymentsRepo *repositories.BankPaymentsRepository
	salesRepo        *repositories.SalesRepository
	returnsRepo      *repositories.ReturnsRepository
	logger           interfaces.Logger
}

func NewBankPaymentsService(bankPaymentsRepo *repositories.BankPaymentsRepository, salesRepo *repositories.SalesRepository, returnsRepo *repositories.ReturnsRepository, logger interfaces.Logger) *BankPaymentsService {
	return &BankPaymentsService{
		bankPaymentsRepo: bankPaymentsRepo,
		salesRepo:        salesRepo,
		returnsRepo:      returnsRepo,
		logger:           logger,
	}
}

// CreateBankPayment creates a new bank payment
func (s *BankPaymentsService) CreateBankPayment(req *models.CreateBankPaymentRequest) (*models.BankPaymentResponse, error) {
	s.logger.Info("Creating bank payment",
		zap.Stringp("sale_id", req.SaleID),
		zap.Stringp("return_id", req.ReturnID),
		zap.String("payment_method", req.PaymentMethod),
		zap.Float64("amount", req.Amount))

	// Validate request
	if err := s.validateBankPaymentRequest(req); err != nil {
		s.logger.Warn("Bank payment validation failed",
			zap.Error(err))
		return nil, err
	}

	// Generate transaction reference
	transactionRef := "TXN" + strconv.FormatInt(time.Now().Unix(), 10)

	s.logger.Debug("Generated transaction reference",
		zap.String("transaction_ref", transactionRef))

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
		s.logger.Error("Failed to create bank payment",
			zap.Error(err),
			zap.String("transaction_ref", transactionRef))
		return nil, err
	}

	s.logger.Info("Bank payment created successfully",
		zap.String("payment_id", payment.ID),
		zap.String("transaction_ref", transactionRef))

	return s.mapBankPaymentToResponse(payment), nil
}

// GetBankPayment retrieves a bank payment by ID
func (s *BankPaymentsService) GetBankPayment(id string) (*models.BankPaymentResponse, error) {
	payment, err := s.bankPaymentsRepo.GetBankPaymentByID(id)
	if err != nil {
		return nil, errors.NewNotFoundError("Bank payment")
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
	s.logger.Info("Updating bank payment",
		zap.String("payment_id", id),
		zap.Bool("has_payment_method", req.PaymentMethod != nil),
		zap.Bool("has_amount", req.Amount != nil))

	payment, err := s.bankPaymentsRepo.GetBankPaymentByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve bank payment for update",
			zap.Error(err),
			zap.String("payment_id", id))
		return nil, errors.NewNotFoundError("Bank payment")
	}

	// Update fields
	if req.PaymentMethod != nil {
		payment.PaymentMethod = *req.PaymentMethod
	}
	if req.Amount != nil {
		payment.Amount = *req.Amount
	}

	if err := s.bankPaymentsRepo.UpdateBankPayment(payment); err != nil {
		s.logger.Error("Failed to update bank payment",
			zap.Error(err),
			zap.String("payment_id", id))
		return nil, err
	}

	s.logger.Info("Bank payment updated successfully",
		zap.String("payment_id", id))

	return s.mapBankPaymentToResponse(payment), nil
}

// DeleteBankPayment deletes a bank payment
func (s *BankPaymentsService) DeleteBankPayment(id string) error {
	s.logger.Info("Deleting bank payment",
		zap.String("payment_id", id))

	if err := s.bankPaymentsRepo.DeleteBankPayment(id); err != nil {
		s.logger.Error("Failed to delete bank payment",
			zap.Error(err),
			zap.String("payment_id", id))
		return err
	}

	s.logger.Info("Bank payment deleted successfully",
		zap.String("payment_id", id))

	return nil
}

// Helper methods
func (s *BankPaymentsService) validateBankPaymentRequest(req *models.CreateBankPaymentRequest) error {
	if req.Amount <= 0 {
		return errors.NewValidationError("Amount must be greater than 0")
	}
	if req.PaymentMethod == "" {
		return errors.NewValidationError("Payment method is required")
	}
	if req.SaleID == nil && req.ReturnID == nil {
		return errors.NewValidationError("Either sale ID or return ID is required")
	}
	if req.SaleID != nil && req.ReturnID != nil {
		return errors.NewValidationError("Cannot have both sale ID and return ID")
	}

	// Validate sale exists if provided
	if req.SaleID != nil {
		_, err := s.salesRepo.GetSaleByID(*req.SaleID)
		if err != nil {
			return errors.NewNotFoundError("Sale")
		}
	}

	// Validate return exists if provided
	if req.ReturnID != nil {
		_, err := s.returnsRepo.GetReturnByID(*req.ReturnID)
		if err != nil {
			return errors.NewNotFoundError("Return")
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
