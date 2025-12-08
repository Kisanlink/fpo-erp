package repositories

import (
	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockWarehouseRepository is a mock implementation of WarehouseInterface
type MockWarehouseRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockWarehouseRepository) Create(warehouse *models.Warehouse) error {
	args := m.Called(warehouse)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *MockWarehouseRepository) GetByID(id string) (*models.Warehouse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Warehouse), args.Error(1)
}

// GetAll mocks the GetAll method
func (m *MockWarehouseRepository) GetAll(limit, offset int) ([]models.Warehouse, int64, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Warehouse), args.Get(1).(int64), args.Error(2)
}

// Update mocks the Update method
func (m *MockWarehouseRepository) Update(warehouse *models.Warehouse) error {
	args := m.Called(warehouse)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockWarehouseRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// Exists mocks the Exists method
func (m *MockWarehouseRepository) Exists(id string) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

// GetByName mocks the GetByName method
func (m *MockWarehouseRepository) GetByName(name string) ([]models.Warehouse, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Warehouse), args.Error(1)
}

// GetByLocation mocks the GetByLocation method
func (m *MockWarehouseRepository) GetByLocation(location string) ([]models.Warehouse, error) {
	args := m.Called(location)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Warehouse), args.Error(1)
}
