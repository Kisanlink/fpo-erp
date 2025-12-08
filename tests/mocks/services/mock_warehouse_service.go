package services

import (
	"context"

	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockWarehouseService is a mock implementation of WarehouseServiceInterface
type MockWarehouseService struct {
	mock.Mock
}

func (m *MockWarehouseService) CreateWarehouse(ctx context.Context, request *models.CreateWarehouseRequest, userID string, jwtToken string) (*models.WarehouseResponse, error) {
	args := m.Called(ctx, request, userID, jwtToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WarehouseResponse), args.Error(1)
}

func (m *MockWarehouseService) GetWarehouse(ctx context.Context, id string, jwtToken string) (*models.WarehouseResponse, error) {
	args := m.Called(ctx, id, jwtToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WarehouseResponse), args.Error(1)
}

func (m *MockWarehouseService) GetAllWarehouses(ctx context.Context, limit, offset int, jwtToken string) ([]models.WarehouseResponse, int64, error) {
	args := m.Called(ctx, limit, offset, jwtToken)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.WarehouseResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockWarehouseService) UpdateWarehouse(ctx context.Context, id string, request *models.UpdateWarehouseRequest, jwtToken string) (*models.WarehouseResponse, error) {
	args := m.Called(ctx, id, request, jwtToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WarehouseResponse), args.Error(1)
}

func (m *MockWarehouseService) DeleteWarehouse(ctx context.Context, id string, jwtToken string) error {
	args := m.Called(ctx, id, jwtToken)
	return args.Error(0)
}

func (m *MockWarehouseService) SearchWarehouses(ctx context.Context, query string, jwtToken string) ([]models.WarehouseResponse, error) {
	args := m.Called(ctx, query, jwtToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.WarehouseResponse), args.Error(1)
}
