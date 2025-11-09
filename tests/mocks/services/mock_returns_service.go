package services

import (
	"time"

	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockReturnsService is a mock implementation of ReturnsServiceInterface
type MockReturnsService struct {
	mock.Mock
}

func (m *MockReturnsService) CreateReturn(req *models.CreateReturnRequest) (*models.ReturnResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ReturnResponse), args.Error(1)
}

func (m *MockReturnsService) GetReturn(id string) (*models.ReturnResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ReturnResponse), args.Error(1)
}

func (m *MockReturnsService) GetAllReturns(limit, offset int) ([]models.ReturnResponse, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ReturnResponse), args.Error(1)
}

func (m *MockReturnsService) UpdateReturn(id string, req *models.UpdateReturnRequest) (*models.ReturnResponse, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ReturnResponse), args.Error(1)
}

func (m *MockReturnsService) DeleteReturn(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockReturnsService) GetReturnsByCustomer(customerID string) ([]models.ReturnResponse, error) {
	args := m.Called(customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ReturnResponse), args.Error(1)
}

func (m *MockReturnsService) GetReturnsBySaleID(saleID string) ([]models.ReturnResponse, error) {
	args := m.Called(saleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ReturnResponse), args.Error(1)
}

func (m *MockReturnsService) GetReturnsByDateRange(startDate, endDate time.Time) ([]models.ReturnResponse, error) {
	args := m.Called(startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ReturnResponse), args.Error(1)
}

func (m *MockReturnsService) GetReturnsByStatus(status string) ([]models.ReturnResponse, error) {
	args := m.Called(status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ReturnResponse), args.Error(1)
}

func (m *MockReturnsService) GetTotalReturnsAmount(startDate, endDate time.Time) (float64, error) {
	args := m.Called(startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockReturnsService) GetReturnRateByProduct(productID string, startDate, endDate time.Time) (float64, error) {
	args := m.Called(productID, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockReturnsService) GetMostReturnedProducts(limit int) ([]models.MostReturnedProductResponse, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.MostReturnedProductResponse), args.Error(1)
}
