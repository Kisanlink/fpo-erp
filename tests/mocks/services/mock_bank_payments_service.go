package services

import (
	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockBankPaymentsService is a mock implementation of BankPaymentsServiceInterface
type MockBankPaymentsService struct {
	mock.Mock
}

func (m *MockBankPaymentsService) CreateBankPayment(req *models.CreateBankPaymentRequest) (*models.BankPaymentResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BankPaymentResponse), args.Error(1)
}

func (m *MockBankPaymentsService) GetBankPayment(id string) (*models.BankPaymentResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BankPaymentResponse), args.Error(1)
}

func (m *MockBankPaymentsService) GetAllBankPayments(limit, offset int) ([]models.BankPaymentResponse, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.BankPaymentResponse), args.Error(1)
}

func (m *MockBankPaymentsService) GetBankPaymentsBySaleID(saleID string) ([]models.BankPaymentResponse, error) {
	args := m.Called(saleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.BankPaymentResponse), args.Error(1)
}

func (m *MockBankPaymentsService) GetBankPaymentsByReturnID(returnID string) ([]models.BankPaymentResponse, error) {
	args := m.Called(returnID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.BankPaymentResponse), args.Error(1)
}

func (m *MockBankPaymentsService) UpdateBankPayment(id string, req *models.UpdateBankPaymentRequest) (*models.BankPaymentResponse, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BankPaymentResponse), args.Error(1)
}

func (m *MockBankPaymentsService) DeleteBankPayment(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
