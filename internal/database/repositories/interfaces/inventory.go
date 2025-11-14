package interfaces

import (
	"kisanlink-erp/internal/database/models"
	"time"

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

	// GetBatchesByWarehouse retrieves all batches for a warehouse
	GetBatchesByWarehouse(warehouseID string) ([]models.InventoryBatch, error)

	// GetBatchesByVariant retrieves all batches for a product variant
	GetBatchesByVariant(variantID string) ([]models.InventoryBatch, error)

	// GetBatchesByVariantOrderedByExpiry retrieves batches for a variant ordered by expiry (FEFO)
	GetBatchesByVariantOrderedByExpiry(variantID string) ([]models.InventoryBatch, error)

	// GetBatchesByVariantAndWarehouseOrderedByExpiry retrieves batches for FEFO allocation (critical for sales)
	GetBatchesByVariantAndWarehouseOrderedByExpiry(variantID, warehouseID string) ([]models.InventoryBatch, error)

	// GetAllBatches retrieves all inventory batches
	GetAllBatches() ([]models.InventoryBatch, error)

	// UpdateBatch updates an existing inventory batch
	UpdateBatch(batch *models.InventoryBatch) error

	// DeleteBatch soft-deletes an inventory batch
	DeleteBatch(id string) error

	// GetExpiringBatches retrieves batches expiring within specified days
	GetExpiringBatches(days int) ([]models.InventoryBatch, error)

	// GetLowStockBatches retrieves batches below threshold quantity
	GetLowStockBatches(threshold int64) ([]models.InventoryBatch, error)

	// CreateTransaction creates a new inventory transaction
	CreateTransaction(transaction *models.InventoryTransaction) error

	// CreateTransactionWithTx creates a new inventory transaction within a transaction
	CreateTransactionWithTx(tx *gorm.DB, transaction *models.InventoryTransaction) error

	// GetTransactionsByBatch retrieves all transactions for a batch
	GetTransactionsByBatch(batchID string) ([]models.InventoryTransaction, error)

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
