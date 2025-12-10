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
	GetBatchesByWarehouse(warehouseID string, limit, offset int) ([]models.InventoryBatchResponse, int64, error)
	GetBatchesByVariant(variantID string, limit, offset int) ([]models.InventoryBatchResponse, int64, error)
	CreateTransaction(batchID string, request *models.CreateInventoryTransactionRequest) (*models.InventoryTransactionResponse, error)
	GetTransactionsByBatch(batchID string, limit, offset int) ([]models.InventoryTransactionResponse, int64, error)
	GetExpiringBatches(days int, limit, offset int) ([]models.InventoryBatchResponse, int64, error)
	GetLowStockBatches(threshold int64, limit, offset int) ([]models.InventoryBatchResponse, int64, error)
	// GetAllProductsAvailability returns grouped availability by SKU with per-warehouse breakdown
	// Only includes non-expired stock in availability counts, but shows expired stock separately
	GetAllProductsAvailability(ctx context.Context, jwtToken string, limit, offset int) ([]models.ProductAvailabilityGroupedResponse, int64, error)
}
