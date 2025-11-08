package repositories

import (
	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockProductVariantRepository is a mock implementation of ProductVariantInterface
type MockProductVariantRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockProductVariantRepository) Create(variant *models.ProductVariant) error {
	args := m.Called(variant)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *MockProductVariantRepository) GetByID(id string) (*models.ProductVariant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariant), args.Error(1)
}

// GetByProductID mocks the GetByProductID method
func (m *MockProductVariantRepository) GetByProductID(productID string) ([]models.ProductVariant, error) {
	args := m.Called(productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ProductVariant), args.Error(1)
}

// GetBySKU mocks the GetBySKU method
func (m *MockProductVariantRepository) GetBySKU(sku string) (*models.ProductVariant, error) {
	args := m.Called(sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariant), args.Error(1)
}

// GetByBarcode mocks the GetByBarcode method
func (m *MockProductVariantRepository) GetByBarcode(barcode string) (*models.ProductVariant, error) {
	args := m.Called(barcode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariant), args.Error(1)
}

// Update mocks the Update method
func (m *MockProductVariantRepository) Update(variant *models.ProductVariant) error {
	args := m.Called(variant)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockProductVariantRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// SKUExists mocks the SKUExists method
func (m *MockProductVariantRepository) SKUExists(sku string) (bool, error) {
	args := m.Called(sku)
	return args.Bool(0), args.Error(1)
}

// BarcodeExists mocks the BarcodeExists method
func (m *MockProductVariantRepository) BarcodeExists(barcode string) (bool, error) {
	args := m.Called(barcode)
	return args.Bool(0), args.Error(1)
}

// FindByExternalID mocks the FindByExternalID method
func (m *MockProductVariantRepository) FindByExternalID(externalID string) (*models.ProductVariant, error) {
	args := m.Called(externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariant), args.Error(1)
}

// FindBySKU mocks the FindBySKU method
func (m *MockProductVariantRepository) FindBySKU(sku string) (*models.ProductVariant, error) {
	args := m.Called(sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariant), args.Error(1)
}
