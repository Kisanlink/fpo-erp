package services

import (
	"context"

	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockPurchaseOrderService is a mock implementation of PurchaseOrderServiceInterface
type MockPurchaseOrderService struct {
	mock.Mock
}

func (m *MockPurchaseOrderService) CreatePurchaseOrder(ctx context.Context, request *models.CreatePurchaseOrderRequest, jwtToken string) (*models.PurchaseOrderResponse, error) {
	args := m.Called(ctx, request, jwtToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchaseOrderResponse), args.Error(1)
}

func (m *MockPurchaseOrderService) GetPurchaseOrder(ctx context.Context, id string) (*models.PurchaseOrderResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchaseOrderResponse), args.Error(1)
}

func (m *MockPurchaseOrderService) GetAllPurchaseOrders(ctx context.Context, limit, offset int) ([]models.PurchaseOrderResponse, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.PurchaseOrderResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockPurchaseOrderService) GetPurchaseOrdersByCollaborator(ctx context.Context, collaboratorID string, limit, offset int) ([]models.PurchaseOrderResponse, int64, error) {
	args := m.Called(ctx, collaboratorID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.PurchaseOrderResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockPurchaseOrderService) GetPurchaseOrdersByStatus(ctx context.Context, status string, limit, offset int) ([]models.PurchaseOrderResponse, int64, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.PurchaseOrderResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockPurchaseOrderService) GetPendingDeliveries(ctx context.Context, limit, offset int) ([]models.PurchaseOrderResponse, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.PurchaseOrderResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockPurchaseOrderService) UpdatePurchaseOrderStatus(ctx context.Context, id string, request *models.UpdatePOStatusRequest, userID string) (*models.PurchaseOrderResponse, error) {
	args := m.Called(ctx, id, request, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchaseOrderResponse), args.Error(1)
}

func (m *MockPurchaseOrderService) UpdatePaymentStatus(ctx context.Context, id string, request *models.UpdatePOPaymentRequest) (*models.PurchaseOrderResponse, error) {
	args := m.Called(ctx, id, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchaseOrderResponse), args.Error(1)
}
