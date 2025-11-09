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

func (m *MockGRNService) GetAllGRNs(ctx context.Context) ([]models.GRNResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GRNResponse), args.Error(1)
}

func (m *MockGRNService) GetGRNsByWarehouse(ctx context.Context, warehouseID string) ([]models.GRNResponse, error) {
	args := m.Called(ctx, warehouseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GRNResponse), args.Error(1)
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
