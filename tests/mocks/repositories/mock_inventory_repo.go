package repositories

import (
	"time"

	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockInventoryRepository is a mock implementation of InventoryInterface
type MockInventoryRepository struct {
	mock.Mock
}

// CreateBatchWithTransaction mocks the CreateBatchWithTransaction method
func (m *MockInventoryRepository) CreateBatchWithTransaction(batch *models.InventoryBatch, transaction *models.InventoryTransaction) error {
	args := m.Called(batch, transaction)
	return args.Error(0)
}

// CreateBatch mocks the CreateBatch method
func (m *MockInventoryRepository) CreateBatch(batch *models.InventoryBatch) error {
	args := m.Called(batch)
	return args.Error(0)
}

// CreateBatchWithTx mocks the CreateBatchWithTx method
func (m *MockInventoryRepository) CreateBatchWithTx(tx *gorm.DB, batch *models.InventoryBatch) error {
	args := m.Called(tx, batch)
	return args.Error(0)
}

// GetBatchByID mocks the GetBatchByID method
func (m *MockInventoryRepository) GetBatchByID(id string) (*models.InventoryBatch, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InventoryBatch), args.Error(1)
}

// GetBatchesByWarehouse mocks the GetBatchesByWarehouse method
func (m *MockInventoryRepository) GetBatchesByWarehouse(warehouseID string, limit, offset int) ([]models.InventoryBatch, int64, error) {
	args := m.Called(warehouseID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.InventoryBatch), args.Get(1).(int64), args.Error(2)
}

// GetBatchesByVariant mocks the GetBatchesByVariant method
func (m *MockInventoryRepository) GetBatchesByVariant(variantID string, limit, offset int) ([]models.InventoryBatch, int64, error) {
	args := m.Called(variantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.InventoryBatch), args.Get(1).(int64), args.Error(2)
}

// GetBatchesByVariantOrderedByExpiry mocks the GetBatchesByVariantOrderedByExpiry method
func (m *MockInventoryRepository) GetBatchesByVariantOrderedByExpiry(variantID string) ([]models.InventoryBatch, error) {
	args := m.Called(variantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryBatch), args.Error(1)
}

// GetBatchesByVariantAndWarehouseOrderedByExpiry mocks the GetBatchesByVariantAndWarehouseOrderedByExpiry method
func (m *MockInventoryRepository) GetBatchesByVariantAndWarehouseOrderedByExpiry(variantID, warehouseID string) ([]models.InventoryBatch, error) {
	args := m.Called(variantID, warehouseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryBatch), args.Error(1)
}

// GetAllBatches mocks the GetAllBatches method
func (m *MockInventoryRepository) GetAllBatches() ([]models.InventoryBatch, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryBatch), args.Error(1)
}

// UpdateBatch mocks the UpdateBatch method
func (m *MockInventoryRepository) UpdateBatch(batch *models.InventoryBatch) error {
	args := m.Called(batch)
	return args.Error(0)
}

// DeleteBatch mocks the DeleteBatch method
func (m *MockInventoryRepository) DeleteBatch(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// GetExpiringBatches mocks the GetExpiringBatches method
func (m *MockInventoryRepository) GetExpiringBatches(days int, limit, offset int) ([]models.InventoryBatch, int64, error) {
	args := m.Called(days, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.InventoryBatch), args.Get(1).(int64), args.Error(2)
}

// GetLowStockBatches mocks the GetLowStockBatches method
func (m *MockInventoryRepository) GetLowStockBatches(threshold int64, limit, offset int) ([]models.InventoryBatch, int64, error) {
	args := m.Called(threshold, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.InventoryBatch), args.Get(1).(int64), args.Error(2)
}

// CreateTransaction mocks the CreateTransaction method
func (m *MockInventoryRepository) CreateTransaction(transaction *models.InventoryTransaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

// CreateTransactionWithTx mocks the CreateTransactionWithTx method
func (m *MockInventoryRepository) CreateTransactionWithTx(tx *gorm.DB, transaction *models.InventoryTransaction) error {
	args := m.Called(tx, transaction)
	return args.Error(0)
}

// GetTransactionsByBatch mocks the GetTransactionsByBatch method
func (m *MockInventoryRepository) GetTransactionsByBatch(batchID string, limit, offset int) ([]models.InventoryTransaction, int64, error) {
	args := m.Called(batchID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.InventoryTransaction), args.Get(1).(int64), args.Error(2)
}

// GetTransactionsByType mocks the GetTransactionsByType method
func (m *MockInventoryRepository) GetTransactionsByType(transactionType string) ([]models.InventoryTransaction, error) {
	args := m.Called(transactionType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryTransaction), args.Error(1)
}

// GetTransactionsByDateRange mocks the GetTransactionsByDateRange method
func (m *MockInventoryRepository) GetTransactionsByDateRange(startDate, endDate time.Time) ([]models.InventoryTransaction, error) {
	args := m.Called(startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryTransaction), args.Error(1)
}

// GetBatchStockLevel mocks the GetBatchStockLevel method
func (m *MockInventoryRepository) GetBatchStockLevel(batchID string) (int64, error) {
	args := m.Called(batchID)
	return args.Get(0).(int64), args.Error(1)
}

// UpdateBatchStock mocks the UpdateBatchStock method
func (m *MockInventoryRepository) UpdateBatchStock(batchID string, quantityChange int64) error {
	args := m.Called(batchID, quantityChange)
	return args.Error(0)
}

// UpdateBatchStockWithTx mocks the UpdateBatchStockWithTx method
func (m *MockInventoryRepository) UpdateBatchStockWithTx(tx *gorm.DB, batchID string, quantityChange int64) error {
	args := m.Called(tx, batchID, quantityChange)
	return args.Error(0)
}
