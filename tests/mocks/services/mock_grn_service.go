package services

import (
	"context"

	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockGRNService is a mock implementation of GRNServiceInterface
type MockGRNService struct {
	mock.Mock
}

func (m *MockGRNService) CreateGRN(ctx context.Context, request *models.CreateGRNRequest) (*models.GRNResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GRNResponse), args.Error(1)
}

func (m *MockGRNService) GetGRN(ctx context.Context, id string) (*models.GRNResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GRNResponse), args.Error(1)
}

func (m *MockGRNService) GetAllGRNs(ctx context.Context, limit, offset int) ([]models.GRNResponse, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.GRNResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockGRNService) GetGRNsByWarehouse(ctx context.Context, warehouseID string, limit, offset int) ([]models.GRNResponse, int64, error) {
	args := m.Called(ctx, warehouseID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.GRNResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockGRNService) GetGRNByPurchaseOrder(ctx context.Context, poID string) (*models.GRNResponse, error) {
	args := m.Called(ctx, poID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GRNResponse), args.Error(1)
}

func (m *MockGRNService) UpdateGRN(ctx context.Context, id string, request *models.UpdateGRNRequest) (*models.GRNResponse, error) {
	args := m.Called(ctx, id, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GRNResponse), args.Error(1)
}

func (m *MockGRNService) GetRejectedItems(ctx context.Context, grnID string) (*models.RejectedItemsResponse, error) {
	args := m.Called(ctx, grnID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RejectedItemsResponse), args.Error(1)
}

func (m *MockGRNService) UpdateItemReturnStatus(ctx context.Context, itemID string, request *models.UpdateItemReturnStatusRequest) (*models.GRNItemResponse, error) {
	args := m.Called(ctx, itemID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GRNItemResponse), args.Error(1)
}
