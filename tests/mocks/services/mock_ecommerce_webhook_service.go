package services

import (
	"context"

	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockEcommerceWebhookService is a mock implementation of EcommerceWebhookServiceInterface
type MockEcommerceWebhookService struct {
	mock.Mock
}

func (m *MockEcommerceWebhookService) ProcessOrderCreated(ctx context.Context, webhook *models.OrderCreatedWebhook) (string, error) {
	args := m.Called(ctx, webhook)
	return args.String(0), args.Error(1)
}

func (m *MockEcommerceWebhookService) ProcessOrderConfirmed(ctx context.Context, webhook *models.OrderConfirmedWebhook) error {
	args := m.Called(ctx, webhook)
	return args.Error(0)
}

func (m *MockEcommerceWebhookService) ProcessOrderShipped(ctx context.Context, webhook *models.OrderShippedWebhook) error {
	args := m.Called(ctx, webhook)
	return args.Error(0)
}

func (m *MockEcommerceWebhookService) ProcessOrderDelivered(ctx context.Context, webhook *models.OrderDeliveredWebhook) error {
	args := m.Called(ctx, webhook)
	return args.Error(0)
}

func (m *MockEcommerceWebhookService) ProcessOrderPayment(ctx context.Context, webhook *models.OrderPaymentWebhook) error {
	args := m.Called(ctx, webhook)
	return args.Error(0)
}
