package repositories

import (
	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockProductRepository is a mock implementation of ProductInterface
type MockProductRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockProductRepository) Create(product *models.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *MockProductRepository) GetByID(id string) (*models.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

// GetBySKU mocks the GetBySKU method
func (m *MockProductRepository) GetBySKU(sku string) (*models.Product, error) {
	args := m.Called(sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

// GetAll mocks the GetAll method
func (m *MockProductRepository) GetAll() ([]models.Product, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Product), args.Error(1)
}

// Update mocks the Update method
func (m *MockProductRepository) Update(product *models.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockProductRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// Exists mocks the Exists method
func (m *MockProductRepository) Exists(id string) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

// SKUExists mocks the SKUExists method
func (m *MockProductRepository) SKUExists(sku string) (bool, error) {
	args := m.Called(sku)
	return args.Bool(0), args.Error(1)
}

// GetByName mocks the GetByName method
func (m *MockProductRepository) GetByName(name string) ([]models.Product, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Product), args.Error(1)
}

// FindByExternalID mocks the FindByExternalID method
func (m *MockProductRepository) FindByExternalID(externalID string) (*models.Product, error) {
	args := m.Called(externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}
