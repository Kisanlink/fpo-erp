package repositories

import (
	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockPurchaseOrderRepository is a mock implementation of PurchaseOrderInterface
type MockPurchaseOrderRepository struct {
	mock.Mock
}

// WithTransaction mocks the WithTransaction method
func (m *MockPurchaseOrderRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

// Create mocks the Create method
func (m *MockPurchaseOrderRepository) Create(po *models.PurchaseOrder) error {
	args := m.Called(po)
	return args.Error(0)
}

// CreateWithTx mocks the CreateWithTx method
func (m *MockPurchaseOrderRepository) CreateWithTx(tx *gorm.DB, po *models.PurchaseOrder) error {
	args := m.Called(tx, po)
	return args.Error(0)
}

// CreateItem mocks the CreateItem method
func (m *MockPurchaseOrderRepository) CreateItem(item *models.PurchaseOrderItem) error {
	args := m.Called(item)
	return args.Error(0)
}

// CreateItemWithTx mocks the CreateItemWithTx method
func (m *MockPurchaseOrderRepository) CreateItemWithTx(tx *gorm.DB, item *models.PurchaseOrderItem) error {
	args := m.Called(tx, item)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *MockPurchaseOrderRepository) GetByID(id string) (*models.PurchaseOrder, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchaseOrder), args.Error(1)
}

// GetByIDWithItems mocks the GetByIDWithItems method
func (m *MockPurchaseOrderRepository) GetByIDWithItems(id string) (*models.PurchaseOrder, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchaseOrder), args.Error(1)
}

// GetByPONumber mocks the GetByPONumber method
func (m *MockPurchaseOrderRepository) GetByPONumber(poNumber string) (*models.PurchaseOrder, error) {
	args := m.Called(poNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchaseOrder), args.Error(1)
}

// GetAll mocks the GetAll method
func (m *MockPurchaseOrderRepository) GetAll(limit, offset int) ([]models.PurchaseOrder, int64, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.PurchaseOrder), args.Get(1).(int64), args.Error(2)
}

// GetByCollaborator mocks the GetByCollaborator method
func (m *MockPurchaseOrderRepository) GetByCollaborator(collaboratorID string, limit, offset int) ([]models.PurchaseOrder, int64, error) {
	args := m.Called(collaboratorID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.PurchaseOrder), args.Get(1).(int64), args.Error(2)
}

// GetByStatus mocks the GetByStatus method
func (m *MockPurchaseOrderRepository) GetByStatus(status string, limit, offset int) ([]models.PurchaseOrder, int64, error) {
	args := m.Called(status, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.PurchaseOrder), args.Get(1).(int64), args.Error(2)
}

// GetPendingDeliveries mocks the GetPendingDeliveries method
func (m *MockPurchaseOrderRepository) GetPendingDeliveries(limit, offset int) ([]models.PurchaseOrder, int64, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.PurchaseOrder), args.Get(1).(int64), args.Error(2)
}

// Update mocks the Update method
func (m *MockPurchaseOrderRepository) Update(po *models.PurchaseOrder) error {
	args := m.Called(po)
	return args.Error(0)
}

// UpdateWithTx mocks the UpdateWithTx method
func (m *MockPurchaseOrderRepository) UpdateWithTx(tx *gorm.DB, po *models.PurchaseOrder) error {
	args := m.Called(tx, po)
	return args.Error(0)
}

// UpdateStatus mocks the UpdateStatus method
func (m *MockPurchaseOrderRepository) UpdateStatus(poID, status string) error {
	args := m.Called(poID, status)
	return args.Error(0)
}

// GetItemByID mocks the GetItemByID method
func (m *MockPurchaseOrderRepository) GetItemByID(itemID string) (*models.PurchaseOrderItem, error) {
	args := m.Called(itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchaseOrderItem), args.Error(1)
}

// UpdateItemReceivedQuantity mocks the UpdateItemReceivedQuantity method
func (m *MockPurchaseOrderRepository) UpdateItemReceivedQuantity(itemID string, receivedQty int64) error {
	args := m.Called(itemID, receivedQty)
	return args.Error(0)
}

// UpdateItemReceivedQuantityWithTx mocks the UpdateItemReceivedQuantityWithTx method
func (m *MockPurchaseOrderRepository) UpdateItemReceivedQuantityWithTx(tx *gorm.DB, itemID string, receivedQty int64) error {
	args := m.Called(tx, itemID, receivedQty)
	return args.Error(0)
}

// PONumberExists mocks the PONumberExists method
func (m *MockPurchaseOrderRepository) PONumberExists(poNumber string) (bool, error) {
	args := m.Called(poNumber)
	return args.Bool(0), args.Error(1)
}

// FindByExternalOrderID mocks the FindByExternalOrderID method
func (m *MockPurchaseOrderRepository) FindByExternalOrderID(externalOrderID string) (*models.PurchaseOrder, error) {
	args := m.Called(externalOrderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchaseOrder), args.Error(1)
}
