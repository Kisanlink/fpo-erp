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

func (m *MockInventoryService) GetBatchesByWarehouse(warehouseID string) ([]models.InventoryBatchResponse, error) {
	args := m.Called(warehouseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryBatchResponse), args.Error(1)
}

func (m *MockInventoryService) GetBatchesByVariant(variantID string) ([]models.InventoryBatchResponse, error) {
	args := m.Called(variantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryBatchResponse), args.Error(1)
}

func (m *MockInventoryService) CreateTransaction(batchID string, request *models.CreateInventoryTransactionRequest) (*models.InventoryTransactionResponse, error) {
	args := m.Called(batchID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InventoryTransactionResponse), args.Error(1)
}

func (m *MockInventoryService) GetTransactionsByBatch(batchID string) ([]models.InventoryTransactionResponse, error) {
	args := m.Called(batchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryTransactionResponse), args.Error(1)
}

func (m *MockInventoryService) GetExpiringBatches(days int) ([]models.InventoryBatchResponse, error) {
	args := m.Called(days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryBatchResponse), args.Error(1)
}

func (m *MockInventoryService) GetLowStockBatches(threshold int64) ([]models.InventoryBatchResponse, error) {
	args := m.Called(threshold)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryBatchResponse), args.Error(1)
}

func (m *MockInventoryService) GetAllProductsAvailability(ctx context.Context, jwtToken string) ([]models.ProductAvailabilityResponse, error) {
	args := m.Called(ctx, jwtToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ProductAvailabilityResponse), args.Error(1)
}
