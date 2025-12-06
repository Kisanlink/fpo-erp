package interfaces

import (
	"time"

	"kisanlink-erp/internal/database/models"

	"gorm.io/gorm"
)

// InventoryInterface defines the contract for inventory repository operations
type InventoryInterface interface {
	// CreateBatchWithTransaction creates a batch and transaction in a single database transaction
	CreateBatchWithTransaction(batch *models.InventoryBatch, transaction *models.InventoryTransaction) error

	// CreateBatch creates a new inventory batch
	CreateBatch(batch *models.InventoryBatch) error

	// CreateBatchWithTx creates a new inventory batch within a transaction
	CreateBatchWithTx(tx *gorm.DB, batch *models.InventoryBatch) error

	// GetBatchByID retrieves an inventory batch by ID
	GetBatchByID(id string) (*models.InventoryBatch, error)

	// GetBatchesByWarehouse retrieves all batches for a warehouse (paginated)
	GetBatchesByWarehouse(warehouseID string, limit, offset int) ([]models.InventoryBatch, int64, error)

	// GetBatchesByVariant retrieves all batches for a product variant (paginated)
	GetBatchesByVariant(variantID string, limit, offset int) ([]models.InventoryBatch, int64, error)

	// GetBatchesByVariantOrderedByExpiry retrieves batches for a variant ordered by expiry (FEFO)
	GetBatchesByVariantOrderedByExpiry(variantID string) ([]models.InventoryBatch, error)

	// GetBatchesByVariantAndWarehouseOrderedByExpiry retrieves batches for FEFO allocation (critical for sales)
	GetBatchesByVariantAndWarehouseOrderedByExpiry(variantID, warehouseID string) ([]models.InventoryBatch, error)

	// GetAllBatches retrieves all inventory batches without pagination (for legacy/internal use)
	GetAllBatches() ([]models.InventoryBatch, error)

	// UpdateBatch updates an existing inventory batch
	UpdateBatch(batch *models.InventoryBatch) error

	// DeleteBatch soft-deletes an inventory batch
	DeleteBatch(id string) error

	// GetExpiringBatches retrieves batches expiring within specified days (paginated)
	GetExpiringBatches(days int, limit, offset int) ([]models.InventoryBatch, int64, error)

	// GetLowStockBatches retrieves batches below threshold quantity (paginated)
	GetLowStockBatches(threshold int64, limit, offset int) ([]models.InventoryBatch, int64, error)

	// CreateTransaction creates a new inventory transaction
	CreateTransaction(transaction *models.InventoryTransaction) error

	// CreateTransactionWithTx creates a new inventory transaction within a transaction
	CreateTransactionWithTx(tx *gorm.DB, transaction *models.InventoryTransaction) error

	// GetTransactionsByBatch retrieves all transactions for a batch (paginated)
	GetTransactionsByBatch(batchID string, limit, offset int) ([]models.InventoryTransaction, int64, error)

	// GetTransactionsByType retrieves transactions by type
	GetTransactionsByType(transactionType string) ([]models.InventoryTransaction, error)

	// GetTransactionsByDateRange retrieves transactions within date range
	GetTransactionsByDateRange(startDate, endDate time.Time) ([]models.InventoryTransaction, error)

	// GetBatchStockLevel retrieves current stock level for a batch
	GetBatchStockLevel(batchID string) (int64, error)

	// UpdateBatchStock updates batch stock level
	UpdateBatchStock(batchID string, quantityChange int64) error

	// UpdateBatchStockWithTx updates batch stock level within a transaction
	UpdateBatchStockWithTx(tx *gorm.DB, batchID string, quantityChange int64) error
}
