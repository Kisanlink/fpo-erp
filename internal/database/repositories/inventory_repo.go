package repositories

import (
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

// Batch operations

// CreateBatch creates a new inventory batch
func (r *InventoryRepository) CreateBatch(batch *models.InventoryBatch) error {
	if err := r.db.Create(batch).Error; err != nil {
		return errors.NewInternalServerError("Failed to create inventory batch")
	}
	return nil
}

// GetBatchByID retrieves an inventory batch by ID
func (r *InventoryRepository) GetBatchByID(id string) (*models.InventoryBatch, error) {
	var batch models.InventoryBatch
	if err := r.db.Preload("Warehouse").Preload("Product").Where("id = ?", id).First(&batch).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Inventory batch")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve inventory batch")
	}
	return &batch, nil
}

// GetBatchesByWarehouse retrieves all batches for a warehouse
func (r *InventoryRepository) GetBatchesByWarehouse(warehouseID string) ([]models.InventoryBatch, error) {
	var batches []models.InventoryBatch
	if err := r.db.Preload("Product").Where("warehouse_id = ?", warehouseID).Find(&batches).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve warehouse batches")
	}
	return batches, nil
}

// GetBatchesByProduct retrieves all batches for a product
func (r *InventoryRepository) GetBatchesByProduct(productID string) ([]models.InventoryBatch, error) {
	var batches []models.InventoryBatch
	if err := r.db.Preload("Warehouse").Where("product_id = ?", productID).Find(&batches).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve product batches")
	}
	return batches, nil
}

// GetBatchesByProductOrderedByExpiry retrieves batches for a product ordered by expiry date (FEFO)
func (r *InventoryRepository) GetBatchesByProductOrderedByExpiry(productID string) ([]models.InventoryBatch, error) {
	var batches []models.InventoryBatch
	if err := r.db.Preload("Warehouse").Where("product_id = ? AND total_quantity > 0", productID).Order("expiry_date ASC").Find(&batches).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve product batches ordered by expiry")
	}
	return batches, nil
}

// GetBatchesByProductAndWarehouseOrderedByExpiry retrieves batches for a product in a specific warehouse ordered by expiry date (FEFO)
func (r *InventoryRepository) GetBatchesByProductAndWarehouseOrderedByExpiry(productID, warehouseID string) ([]models.InventoryBatch, error) {
	var batches []models.InventoryBatch
	if err := r.db.Preload("Warehouse").Where("product_id = ? AND warehouse_id = ? AND total_quantity > 0", productID, warehouseID).Order("expiry_date ASC").Find(&batches).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve product batches in warehouse ordered by expiry")
	}
	return batches, nil
}

// GetAllBatches retrieves all inventory batches with warehouse and product details
func (r *InventoryRepository) GetAllBatches() ([]models.InventoryBatch, error) {
	var batches []models.InventoryBatch
	if err := r.db.Preload("Warehouse").Preload("Product").Find(&batches).Error; err != nil {
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

// GetExpiringBatches retrieves batches that expire within a given timeframe
func (r *InventoryRepository) GetExpiringBatches(days int) ([]models.InventoryBatch, error) {
	var batches []models.InventoryBatch
	expiryDate := time.Now().AddDate(0, 0, days)
	if err := r.db.Preload("Warehouse").Preload("Product").Where("expiry_date <= ?", expiryDate).Find(&batches).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve expiring batches")
	}
	return batches, nil
}

// GetLowStockBatches retrieves batches with low stock (below threshold)
func (r *InventoryRepository) GetLowStockBatches(threshold int64) ([]models.InventoryBatch, error) {
	var batches []models.InventoryBatch
	if err := r.db.Preload("Warehouse").Preload("Product").Where("total_quantity <= ?", threshold).Find(&batches).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve low stock batches")
	}
	return batches, nil
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

// GetTransactionsByBatch retrieves all transactions for a batch
func (r *InventoryRepository) GetTransactionsByBatch(batchID string) ([]models.InventoryTransaction, error) {
	var transactions []models.InventoryTransaction
	if err := r.db.Preload("Batch").Where("batch_id = ?", batchID).Order("occurred_at DESC").Find(&transactions).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve batch transactions")
	}
	return transactions, nil
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
