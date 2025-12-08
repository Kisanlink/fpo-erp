package interfaces

import (
	"context"

	"kisanlink-erp/internal/database/models"
)

type EcommerceWebhookServiceInterface interface {
	ProcessOrderCreated(ctx context.Context, webhook *models.OrderCreatedWebhook) (string, error)
	ProcessOrderConfirmed(ctx context.Context, webhook *models.OrderConfirmedWebhook) error
	ProcessOrderShipped(ctx context.Context, webhook *models.OrderShippedWebhook) error
	ProcessOrderDelivered(ctx context.Context, webhook *models.OrderDeliveredWebhook) error
	ProcessOrderPayment(ctx context.Context, webhook *models.OrderPaymentWebhook) error
}
