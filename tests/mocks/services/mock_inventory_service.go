package services

import (
	"context"
	"time"

	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockInventoryService is a mock implementation of InventoryServiceInterface
type MockInventoryService struct {
	mock.Mock
}

func (m *MockInventoryService) CreateBatch(warehouseID, variantID string, costPrice float64, expiryDate time.Time, quantity int64) (*models.InventoryBatchResponse, error) {
	args := m.Called(warehouseID, variantID, costPrice, expiryDate, quantity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InventoryBatchResponse), args.Error(1)
}

func (m *MockInventoryService) GetBatch(id string) (*models.InventoryBatchResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InventoryBatchResponse), args.Error(1)
}

func (m *MockInventoryService) GetBatchesByWarehouse(warehouseID string, limit, offset int) ([]models.InventoryBatchResponse, int64, error) {
	args := m.Called(warehouseID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.InventoryBatchResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockInventoryService) GetBatchesByVariant(variantID string, limit, offset int) ([]models.InventoryBatchResponse, int64, error) {
	args := m.Called(variantID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.InventoryBatchResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockInventoryService) CreateTransaction(batchID string, request *models.CreateInventoryTransactionRequest) (*models.InventoryTransactionResponse, error) {
	args := m.Called(batchID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InventoryTransactionResponse), args.Error(1)
}

func (m *MockInventoryService) GetTransactionsByBatch(batchID string, limit, offset int) ([]models.InventoryTransactionResponse, int64, error) {
	args := m.Called(batchID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.InventoryTransactionResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockInventoryService) GetExpiringBatches(days int, limit, offset int) ([]models.InventoryBatchResponse, int64, error) {
	args := m.Called(days, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.InventoryBatchResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockInventoryService) GetLowStockBatches(threshold int64, limit, offset int) ([]models.InventoryBatchResponse, int64, error) {
	args := m.Called(threshold, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.InventoryBatchResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockInventoryService) GetAllProductsAvailability(ctx context.Context, jwtToken string, limit, offset int) ([]models.ProductAvailabilityGroupedResponse, int64, error) {
	args := m.Called(ctx, jwtToken, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.ProductAvailabilityGroupedResponse), args.Get(1).(int64), args.Error(2)
}
