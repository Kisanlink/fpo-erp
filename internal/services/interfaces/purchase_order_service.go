package interfaces

import (
	"context"

	"kisanlink-erp/internal/database/models"
)

// PurchaseOrderServiceInterface defines the contract for purchase order service operations
type PurchaseOrderServiceInterface interface {
	CreatePurchaseOrder(ctx context.Context, request *models.CreatePurchaseOrderRequest) (*models.PurchaseOrderResponse, error)
	GetPurchaseOrder(ctx context.Context, id string) (*models.PurchaseOrderResponse, error)
	GetAllPurchaseOrders(ctx context.Context) ([]models.PurchaseOrderResponse, error)
	GetPurchaseOrdersByCollaborator(ctx context.Context, collaboratorID string) ([]models.PurchaseOrderResponse, error)
	GetPurchaseOrdersByStatus(ctx context.Context, status string) ([]models.PurchaseOrderResponse, error)
	GetPendingDeliveries(ctx context.Context) ([]models.PurchaseOrderResponse, error)
	UpdatePurchaseOrderStatus(ctx context.Context, id string, request *models.UpdatePOStatusRequest, userID string) (*models.PurchaseOrderResponse, error)
	UpdatePaymentStatus(ctx context.Context, id string, request *models.UpdatePOPaymentRequest) (*models.PurchaseOrderResponse, error)
}
