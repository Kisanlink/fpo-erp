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

func (m *MockPurchaseOrderService) CreatePurchaseOrder(ctx context.Context, request *models.CreatePurchaseOrderRequest) (*models.PurchaseOrderResponse, error) {
	args := m.Called(ctx, request)
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

func (m *MockPurchaseOrderService) GetAllPurchaseOrders(ctx context.Context) ([]models.PurchaseOrderResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PurchaseOrderResponse), args.Error(1)
}

func (m *MockPurchaseOrderService) GetPurchaseOrdersByCollaborator(ctx context.Context, collaboratorID string) ([]models.PurchaseOrderResponse, error) {
	args := m.Called(ctx, collaboratorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PurchaseOrderResponse), args.Error(1)
}

func (m *MockPurchaseOrderService) GetPurchaseOrdersByStatus(ctx context.Context, status string) ([]models.PurchaseOrderResponse, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PurchaseOrderResponse), args.Error(1)
}

func (m *MockPurchaseOrderService) GetPendingDeliveries(ctx context.Context) ([]models.PurchaseOrderResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PurchaseOrderResponse), args.Error(1)
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
