package services

import (
	"context"

	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockProductVariantService is a mock implementation of ProductVariantServiceInterface
type MockProductVariantService struct {
	mock.Mock
}

func (m *MockProductVariantService) CreateProductVariant(ctx context.Context, productID string, request *models.CreateProductVariantRequest) (*models.ProductVariantResponse, error) {
	args := m.Called(ctx, productID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariantResponse), args.Error(1)
}

func (m *MockProductVariantService) GetProductVariant(ctx context.Context, id string) (*models.ProductVariantResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariantResponse), args.Error(1)
}

func (m *MockProductVariantService) GetVariantsByProduct(ctx context.Context, productID string) ([]models.ProductVariantResponse, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ProductVariantResponse), args.Error(1)
}

func (m *MockProductVariantService) GetVariantBySKU(ctx context.Context, sku string) (*models.ProductVariantResponse, error) {
	args := m.Called(ctx, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariantResponse), args.Error(1)
}

func (m *MockProductVariantService) GetVariantByBarcode(ctx context.Context, barcode string) (*models.ProductVariantResponse, error) {
	args := m.Called(ctx, barcode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariantResponse), args.Error(1)
}

func (m *MockProductVariantService) UpdateProductVariant(ctx context.Context, id string, request *models.UpdateProductVariantRequest) (*models.ProductVariantResponse, error) {
	args := m.Called(ctx, id, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariantResponse), args.Error(1)
}

func (m *MockProductVariantService) DeleteProductVariant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
