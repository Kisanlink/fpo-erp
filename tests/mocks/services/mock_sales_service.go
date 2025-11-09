package services

import (
	"time"

	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockSalesService is a mock implementation of SalesServiceInterface
type MockSalesService struct {
	mock.Mock
}

func (m *MockSalesService) CreateSale(req *models.CreateSaleRequest) (*models.SaleResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SaleResponse), args.Error(1)
}

func (m *MockSalesService) GetSale(id string) (*models.SaleResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SaleResponse), args.Error(1)
}

func (m *MockSalesService) GetAllSales(limit, offset int) ([]models.SaleResponse, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SaleResponse), args.Error(1)
}

func (m *MockSalesService) UpdateSale(id string, req *models.UpdateSaleRequest) (*models.SaleResponse, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SaleResponse), args.Error(1)
}

func (m *MockSalesService) DeleteSale(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockSalesService) GetSalesByDateRange(startDate, endDate time.Time) ([]models.SaleResponse, error) {
	args := m.Called(startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SaleResponse), args.Error(1)
}

func (m *MockSalesService) GetSalesByStatus(status string) ([]models.SaleResponse, error) {
	args := m.Called(status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SaleResponse), args.Error(1)
}

func (m *MockSalesService) GetTotalSalesAmount(startDate, endDate time.Time) (float64, error) {
	args := m.Called(startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockSalesService) GetTopSellingProducts(limit int) ([]models.TopSellingProductResponse, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TopSellingProductResponse), args.Error(1)
}
