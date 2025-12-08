package interfaces

import (
	"context"

	"kisanlink-erp/internal/database/models"
)

// PurchaseOrderServiceInterface defines the contract for purchase order service operations
type PurchaseOrderServiceInterface interface {
	CreatePurchaseOrder(ctx context.Context, request *models.CreatePurchaseOrderRequest, jwtToken string) (*models.PurchaseOrderResponse, error)
	GetPurchaseOrder(ctx context.Context, id string) (*models.PurchaseOrderResponse, error)
	GetAllPurchaseOrders(ctx context.Context, limit, offset int) ([]models.PurchaseOrderResponse, int64, error)
	GetPurchaseOrdersByCollaborator(ctx context.Context, collaboratorID string, limit, offset int) ([]models.PurchaseOrderResponse, int64, error)
	GetPurchaseOrdersByStatus(ctx context.Context, status string, limit, offset int) ([]models.PurchaseOrderResponse, int64, error)
	GetPendingDeliveries(ctx context.Context, limit, offset int) ([]models.PurchaseOrderResponse, int64, error)
	UpdatePurchaseOrderStatus(ctx context.Context, id string, request *models.UpdatePOStatusRequest, userID string) (*models.PurchaseOrderResponse, error)
	UpdatePaymentStatus(ctx context.Context, id string, request *models.UpdatePOPaymentRequest) (*models.PurchaseOrderResponse, error)
}
