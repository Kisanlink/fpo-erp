package repositories

import (
	"fmt"
	"log"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"

	"gorm.io/gorm"
)

// InventoryRepository handles inventory data access
type InventoryRepository struct {
	db *gorm.DB
}

// NewInventoryRepository creates a new inventory repository
func NewInventoryRepository(db *gorm.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// CreateBatchWithTransaction creates a batch and initial transaction atomically
func (r *InventoryRepository) CreateBatchWithTransaction(batch *models.InventoryBatch, transaction *models.InventoryTransaction) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create the batch first
		if err := tx.Create(batch).Error; err != nil {
			return errors.NewInternalServerError(fmt.Sprintf("Failed to create inventory batch: %v", err))
		}

		// Update the transaction with the created batch ID
		transaction.BatchID = batch.ID

		// Create the initial transaction
		if err := tx.Create(transaction).Error; err != nil {
			return errors.NewInternalServerError(fmt.Sprintf("Failed to create initial inventory transaction: %v", err))
		}

		return nil
	})
}

// Batch operations

// CreateBatch creates a new inventory batch
func (r *InventoryRepository) CreateBatch(batch *models.InventoryBatch) error {
	if err := r.db.Create(batch).Error; err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("Failed to create inventory batch: %v", err))
	}
	return nil
}

// CreateBatchWithTx creates a new inventory batch within a transaction
func (r *InventoryRepository) CreateBatchWithTx(tx *gorm.DB, batch *models.InventoryBatch) error {
	if err := tx.Create(batch).Error; err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("Failed to create inventory batch: %v", err))
	}
	return nil
}

// GetBatchByID retrieves an inventory batch by ID
func (r *InventoryRepository) GetBatchByID(id string) (*models.InventoryBatch, error) {
	var batch models.InventoryBatch
	if err := r.db.Preload("Warehouse").Preload("Variant").Where("id = ?", id).First(&batch).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Inventory batch")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve inventory batch")
	}
	return &batch, nil
}

// GetBatchesByWarehouse retrieves all batches for a warehouse (paginated)
func (r *InventoryRepository) GetBatchesByWarehouse(warehouseID string, limit, offset int) ([]models.InventoryBatch, int64, error) {
	var batches []models.InventoryBatch
	var total int64

	// Get total count
	if err := r.db.Model(&models.InventoryBatch{}).Where("warehouse_id = ?", warehouseID).Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to count warehouse batches")
	}

	// Get paginated records
	if err := r.db.Preload("Variant").Where("warehouse_id = ?", warehouseID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&batches).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve warehouse batches")
	}
	return batches, total, nil
}

// GetBatchesByVariant retrieves all batches for a product variant (paginated)
func (r *InventoryRepository) GetBatchesByVariant(variantID string, limit, offset int) ([]models.InventoryBatch, int64, error) {
	var batches []models.InventoryBatch
	var total int64

	// Get total count
	if err := r.db.Model(&models.InventoryBatch{}).Where("variant_id = ?", variantID).Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to count variant batches")
	}

	// Get paginated records
	if err := r.db.Preload("Warehouse").Where("variant_id = ?", variantID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&batches).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve variant batches")
	}
	return batches, total, nil
}

// GetBatchesByVariantOrderedByExpiry retrieves batches for a variant ordered by expiry date (FEFO)
func (r *InventoryRepository) GetBatchesByVariantOrderedByExpiry(variantID string) ([]models.InventoryBatch, error) {
	var batches []models.InventoryBatch
	if err := r.db.Preload("Warehouse").Where("variant_id = ? AND total_quantity > 0", variantID).Order("expiry_date ASC").Find(&batches).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve variant batches ordered by expiry")
	}
	return batches, nil
}

// GetBatchesByVariantAndWarehouseOrderedByExpiry retrieves batches for a variant in a specific warehouse ordered by expiry date (FEFO)
func (r *InventoryRepository) GetBatchesByVariantAndWarehouseOrderedByExpiry(variantID, warehouseID string) ([]models.InventoryBatch, error) {
	var batches []models.InventoryBatch
	if err := r.db.Preload("Warehouse").Where("variant_id = ? AND warehouse_id = ? AND total_quantity > 0", variantID, warehouseID).Order("expiry_date ASC").Find(&batches).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve variant batches in warehouse ordered by expiry")
	}
	return batches, nil
}

// GetAllBatchesPaginated retrieves all inventory batches with warehouse and variant details (paginated)
func (r *InventoryRepository) GetAllBatchesPaginated(limit, offset int) ([]models.InventoryBatch, int64, error) {
	var batches []models.InventoryBatch
	var total int64

	// Get total count
	if err := r.db.Model(&models.InventoryBatch{}).Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to count batches")
	}

	// Get paginated records
	if err := r.db.Preload("Warehouse").Preload("Variant").Order("created_at DESC").Limit(limit).Offset(offset).Find(&batches).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve all batches")
	}
	return batches, total, nil
}

// GetAllBatches retrieves all inventory batches without pagination (for legacy/internal use)
func (r *InventoryRepository) GetAllBatches() ([]models.InventoryBatch, error) {
	var batches []models.InventoryBatch
	if err := r.db.Preload("Warehouse").Preload("Variant").Order("created_at DESC").Find(&batches).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve all batches")
	}
	return batches, nil
}

// UpdateBatch updates an inventory batch
func (r *InventoryRepository) UpdateBatch(batch *models.InventoryBatch) error {
	if err := r.db.Save(batch).Error; err != nil {
		return errors.NewInternalServerError("Failed to update inventory batch")
	}
	return nil
}

// DeleteBatch deletes an inventory batch
func (r *InventoryRepository) DeleteBatch(id string) error {
	if err := r.db.Where("id = ?", id).Delete(&models.InventoryBatch{}).Error; err != nil {
		return errors.NewInternalServerError("Failed to delete inventory batch")
	}
	return nil
}

// GetExpiringBatches retrieves batches that expire within a given timeframe (paginated)
func (r *InventoryRepository) GetExpiringBatches(days int, limit, offset int) ([]models.InventoryBatch, int64, error) {
	var batches []models.InventoryBatch
	var total int64
	expiryDate := time.Now().AddDate(0, 0, days)

	// Get total count
	if err := r.db.Model(&models.InventoryBatch{}).Where("expiry_date <= ?", expiryDate).Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to count expiring batches")
	}

	// Get paginated records
	if err := r.db.Preload("Warehouse").Preload("Variant").Where("expiry_date <= ?", expiryDate).Order("created_at DESC").Limit(limit).Offset(offset).Find(&batches).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve expiring batches")
	}
	return batches, total, nil
}

// GetLowStockBatches retrieves batches with low stock (below threshold) (paginated)
func (r *InventoryRepository) GetLowStockBatches(threshold int64, limit, offset int) ([]models.InventoryBatch, int64, error) {
	var batches []models.InventoryBatch
	var total int64

	// Get total count
	if err := r.db.Model(&models.InventoryBatch{}).Where("total_quantity <= ?", threshold).Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to count low stock batches")
	}

	// Get paginated records
	if err := r.db.Preload("Warehouse").Preload("Variant").Where("total_quantity <= ?", threshold).Order("created_at DESC").Limit(limit).Offset(offset).Find(&batches).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve low stock batches")
	}
	return batches, total, nil
}

// Transaction operations

// CreateTransaction creates a new inventory transaction
func (r *InventoryRepository) CreateTransaction(transaction *models.InventoryTransaction) error {
	if err := r.db.Create(transaction).Error; err != nil {
		log.Printf("[ERROR] Database error creating inventory transaction: %v", err)
		return errors.NewInternalServerError("Failed to create inventory transaction")
	}
	return nil
}

// GetTransactionsByBatch retrieves all transactions for a batch (paginated)
func (r *InventoryRepository) GetTransactionsByBatch(batchID string, limit, offset int) ([]models.InventoryTransaction, int64, error) {
	var transactions []models.InventoryTransaction
	var total int64

	// Get total count
	if err := r.db.Model(&models.InventoryTransaction{}).Where("batch_id = ?", batchID).Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to count batch transactions")
	}

	// Get paginated records
	if err := r.db.Preload("Batch").Where("batch_id = ?", batchID).Order("occurred_at DESC").Limit(limit).Offset(offset).Find(&transactions).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve batch transactions")
	}
	return transactions, total, nil
}

// GetTransactionsByType retrieves transactions by type
func (r *InventoryRepository) GetTransactionsByType(transactionType string) ([]models.InventoryTransaction, error) {
	var transactions []models.InventoryTransaction
	if err := r.db.Preload("Batch").Where("transaction_type = ?", transactionType).Order("occurred_at DESC").Find(&transactions).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve transactions by type")
	}
	return transactions, nil
}

// GetTransactionsByDateRange retrieves transactions within a date range
func (r *InventoryRepository) GetTransactionsByDateRange(startDate, endDate time.Time) ([]models.InventoryTransaction, error) {
	var transactions []models.InventoryTransaction
	if err := r.db.Preload("Batch").Where("occurred_at BETWEEN ? AND ?", startDate, endDate).Order("occurred_at DESC").Find(&transactions).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve transactions by date range")
	}
	return transactions, nil
}

// GetBatchStockLevel gets the current stock level for a batch
func (r *InventoryRepository) GetBatchStockLevel(batchID string) (int64, error) {
	var batch models.InventoryBatch
	if err := r.db.Select("total_quantity").Where("id = ?", batchID).First(&batch).Error; err != nil {
		return 0, errors.NewInternalServerError("Failed to retrieve batch stock level")
	}
	return batch.TotalQuantity, nil
}

// UpdateBatchStock updates the stock level for a batch
func (r *InventoryRepository) UpdateBatchStock(batchID string, quantityChange int64) error {
	if err := r.db.Model(&models.InventoryBatch{}).Where("id = ?", batchID).Update("total_quantity", gorm.Expr("total_quantity + ?", quantityChange)).Error; err != nil {
		log.Printf("[ERROR] Database error updating batch stock: %v", err)
		return errors.NewInternalServerError("Failed to update batch stock")
	}
	return nil
}

// UpdateBatchStockWithTx updates the stock level for a batch within a transaction with row lock
func (r *InventoryRepository) UpdateBatchStockWithTx(tx *gorm.DB, batchID string, quantityChange int64) error {
	// Use FOR UPDATE to lock the row and prevent race conditions
	var batch models.InventoryBatch
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", batchID).First(&batch).Error; err != nil {
		log.Printf("[ERROR] Failed to lock batch for update: %v", err)
		return errors.NewInternalServerError("Failed to lock batch for stock update")
	}

	// Check for sufficient stock before update
	if batch.TotalQuantity+quantityChange < 0 {
		log.Printf("[ERROR] Insufficient stock: current=%d, requested=%d", batch.TotalQuantity, -quantityChange)
		return errors.NewBadRequestError("Insufficient stock available")
	}

	if err := tx.Model(&models.InventoryBatch{}).Where("id = ?", batchID).Update("total_quantity", gorm.Expr("total_quantity + ?", quantityChange)).Error; err != nil {
		log.Printf("[ERROR] Database error updating batch stock: %v", err)
		return errors.NewInternalServerError("Failed to update batch stock")
	}
	return nil
}

// CreateTransactionWithTx creates an inventory transaction within a transaction
func (r *InventoryRepository) CreateTransactionWithTx(tx *gorm.DB, transaction *models.InventoryTransaction) error {
	if err := tx.Create(transaction).Error; err != nil {
		log.Printf("[ERROR] Database error creating inventory transaction: %v", err)
		return errors.NewInternalServerError("Failed to create inventory transaction")
	}
	return nil
}
