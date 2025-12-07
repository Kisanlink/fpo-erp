package interfaces

import (
	"context"
	"time"

	"kisanlink-erp/internal/database/models"
)

// InventoryServiceInterface defines the contract for inventory service operations
// Simplified for GST-only tax system - tax rates are on ProductVariant, not on batches
type InventoryServiceInterface interface {
	CreateBatch(warehouseID, variantID string, costPrice float64, expiryDate time.Time, quantity int64) (*models.InventoryBatchResponse, error)
	GetBatch(id string) (*models.InventoryBatchResponse, error)
	GetBatchesByWarehouse(warehouseID string) ([]models.InventoryBatchResponse, error)
	GetBatchesByVariant(variantID string) ([]models.InventoryBatchResponse, error)
	CreateTransaction(batchID string, request *models.CreateInventoryTransactionRequest) (*models.InventoryTransactionResponse, error)
	GetTransactionsByBatch(batchID string) ([]models.InventoryTransactionResponse, error)
	GetExpiringBatches(days int) ([]models.InventoryBatchResponse, error)
	GetLowStockBatches(threshold int64) ([]models.InventoryBatchResponse, error)
	GetAllProductsAvailability(ctx context.Context, jwtToken string) ([]models.ProductAvailabilityResponse, error)
}
