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

func (m *MockSalesService) GetAllSales(limit, offset int) ([]models.SaleListResponse, int64, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.SaleListResponse), args.Get(1).(int64), args.Error(2)
}

// GetSalesByCustomerPhone retrieves sales filtered by customer phone number (Issue 7)
func (m *MockSalesService) GetSalesByCustomerPhone(phone string, limit, offset int) ([]models.SaleListResponse, int64, error) {
	args := m.Called(phone, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.SaleListResponse), args.Get(1).(int64), args.Error(2)
}

// PatchSale partially updates a sale (Issue 9)
func (m *MockSalesService) PatchSale(id string, req *models.PatchSaleRequest) (*models.SaleResponse, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SaleResponse), args.Error(1)
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

func (m *MockSalesService) GetSalesByDateRange(startDate, endDate time.Time, limit, offset int) ([]models.SaleResponse, int64, error) {
	args := m.Called(startDate, endDate, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.SaleResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockSalesService) GetSalesByStatus(status string, limit, offset int) ([]models.SaleResponse, int64, error) {
	args := m.Called(status, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.SaleResponse), args.Get(1).(int64), args.Error(2)
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

//nolint:typecheck // mock.Mock.Called is available at runtime
func (m *MockSalesService) CancelSale(saleID string, req *models.CancelSaleRequest) (*models.CancelSaleResponse, error) {
	args := m.Called(saleID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CancelSaleResponse), args.Error(1)
}

func (m *MockSalesService) CancelItems(saleID string, req *models.CancelItemsRequest) (*models.CancelItemsResponse, error) {
	args := m.Called(saleID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CancelItemsResponse), args.Error(1)
}

func (m *MockSalesService) GetCancellations(saleID string) (*models.GetCancellationsResponse, error) {
	args := m.Called(saleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GetCancellationsResponse), args.Error(1)
}

func (m *MockSalesService) CompleteSale(saleID string, performedBy string) (*models.SaleResponse, error) {
	args := m.Called(saleID, performedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SaleResponse), args.Error(1)
}
