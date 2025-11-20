package interfaces

import (
	"context"

	"kisanlink-erp/internal/database/models"
)

// GRNServiceInterface defines the contract for GRN service operations
type GRNServiceInterface interface {
	CreateGRN(ctx context.Context, request *models.CreateGRNRequest) (*models.GRNResponse, error)
	GetGRN(ctx context.Context, id string) (*models.GRNResponse, error)
	GetAllGRNs(ctx context.Context) ([]models.GRNResponse, error)
	GetGRNsByWarehouse(ctx context.Context, warehouseID string) ([]models.GRNResponse, error)
	GetGRNByPurchaseOrder(ctx context.Context, poID string) (*models.GRNResponse, error)
	UpdateGRN(ctx context.Context, id string, request *models.UpdateGRNRequest) (*models.GRNResponse, error)
	GetRejectedItems(ctx context.Context, grnID string) (*models.RejectedItemsResponse, error)
	UpdateItemReturnStatus(ctx context.Context, itemID string, request *models.UpdateItemReturnStatusRequest) (*models.GRNItemResponse, error)
}
