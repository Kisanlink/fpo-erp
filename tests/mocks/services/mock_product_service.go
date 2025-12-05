package services

import (
	"context"

	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockProductService is a mock implementation of ProductServiceInterface
type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) CreateProduct(request *models.CreateProductRequest) (*models.ProductResponse, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductResponse), args.Error(1)
}

func (m *MockProductService) GetProduct(ctx context.Context, id string) (*models.ProductResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductResponse), args.Error(1)
}

func (m *MockProductService) GetAllProducts(ctx context.Context) ([]models.ProductResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ProductResponse), args.Error(1)
}

func (m *MockProductService) UpdateProduct(id string, request *models.UpdateProductRequest) (*models.ProductResponse, error) {
	args := m.Called(id, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductResponse), args.Error(1)
}

func (m *MockProductService) DeleteProduct(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductService) SearchProducts(query string) ([]models.ProductResponse, error) {
	args := m.Called(query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ProductResponse), args.Error(1)
}

func (m *MockProductService) GetProductWithPrices(id string) (*models.ProductWithPricesResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductWithPricesResponse), args.Error(1)
}
