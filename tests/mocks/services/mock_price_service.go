package services

import (
	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockProductPriceService is a mock implementation of ProductPriceServiceInterface
type MockProductPriceService struct {
	mock.Mock
}

func (m *MockProductPriceService) CreateProductPrice(request *models.CreateProductPriceRequest) (*models.ProductPriceResponse, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductPriceResponse), args.Error(1)
}

func (m *MockProductPriceService) GetProductPrice(id string) (*models.ProductPriceResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductPriceResponse), args.Error(1)
}

func (m *MockProductPriceService) GetVariantPrices(variantID string) ([]models.ProductPriceResponse, error) {
	args := m.Called(variantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ProductPriceResponse), args.Error(1)
}

func (m *MockProductPriceService) GetCurrentPrice(variantID, priceType string) (*models.ProductPriceResponse, error) {
	args := m.Called(variantID, priceType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductPriceResponse), args.Error(1)
}

func (m *MockProductPriceService) UpdateProductPrice(id string, request *models.UpdateProductPriceRequest) (*models.ProductPriceResponse, error) {
	args := m.Called(id, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductPriceResponse), args.Error(1)
}

func (m *MockProductPriceService) DeleteProductPrice(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductPriceService) GetExpiredPrices() ([]models.ProductPriceResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ProductPriceResponse), args.Error(1)
}
