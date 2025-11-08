package repositories

import (
	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockGRNRepository is a mock implementation of GRNInterface
type MockGRNRepository struct {
	mock.Mock
}

// WithTransaction mocks the WithTransaction method
func (m *MockGRNRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

// Create mocks the Create method
func (m *MockGRNRepository) Create(grn *models.GRN) error {
	args := m.Called(grn)
	return args.Error(0)
}

// CreateWithTx mocks the CreateWithTx method
func (m *MockGRNRepository) CreateWithTx(tx *gorm.DB, grn *models.GRN) error {
	args := m.Called(tx, grn)
	return args.Error(0)
}

// CreateItem mocks the CreateItem method
func (m *MockGRNRepository) CreateItem(item *models.GRNItem) error {
	args := m.Called(item)
	return args.Error(0)
}

// CreateItemWithTx mocks the CreateItemWithTx method
func (m *MockGRNRepository) CreateItemWithTx(tx *gorm.DB, item *models.GRNItem) error {
	args := m.Called(tx, item)
	return args.Error(0)
}

// UpdateItemBatchIDWithTx mocks the UpdateItemBatchIDWithTx method
func (m *MockGRNRepository) UpdateItemBatchIDWithTx(tx *gorm.DB, itemID, batchID string) error {
	args := m.Called(tx, itemID, batchID)
	return args.Error(0)
}

// Update mocks the Update method
func (m *MockGRNRepository) Update(id string, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *MockGRNRepository) GetByID(id string) (*models.GRN, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GRN), args.Error(1)
}

// GetByIDWithItems mocks the GetByIDWithItems method
func (m *MockGRNRepository) GetByIDWithItems(id string) (*models.GRN, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GRN), args.Error(1)
}

// GetByGRNNumber mocks the GetByGRNNumber method
func (m *MockGRNRepository) GetByGRNNumber(grnNumber string) (*models.GRN, error) {
	args := m.Called(grnNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GRN), args.Error(1)
}

// GetByPurchaseOrder mocks the GetByPurchaseOrder method
func (m *MockGRNRepository) GetByPurchaseOrder(poID string) (*models.GRN, error) {
	args := m.Called(poID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GRN), args.Error(1)
}

// GetAll mocks the GetAll method
func (m *MockGRNRepository) GetAll() ([]models.GRN, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GRN), args.Error(1)
}

// GetByWarehouse mocks the GetByWarehouse method
func (m *MockGRNRepository) GetByWarehouse(warehouseID string) ([]models.GRN, error) {
	args := m.Called(warehouseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GRN), args.Error(1)
}

// GRNNumberExists mocks the GRNNumberExists method
func (m *MockGRNRepository) GRNNumberExists(grnNumber string) (bool, error) {
	args := m.Called(grnNumber)
	return args.Bool(0), args.Error(1)
}

// GRNExistsForPO mocks the GRNExistsForPO method
func (m *MockGRNRepository) GRNExistsForPO(poID string) (bool, error) {
	args := m.Called(poID)
	return args.Bool(0), args.Error(1)
}
