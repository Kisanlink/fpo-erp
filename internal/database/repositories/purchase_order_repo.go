package repositories

import (
	"fmt"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"

	"gorm.io/gorm"
)

type PurchaseOrderRepository struct {
	db *gorm.DB
}

func NewPurchaseOrderRepository(db *gorm.DB) *PurchaseOrderRepository {
	return &PurchaseOrderRepository{db: db}
}

// WithTransaction executes a function within a transaction
func (r *PurchaseOrderRepository) WithTransaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// Create creates a new purchase order
func (r *PurchaseOrderRepository) Create(po *models.PurchaseOrder) error {
	if err := r.db.Create(po).Error; err != nil {
		return errors.NewInternalServerError("Failed to create purchase order")
	}
	return nil
}

// CreateWithTx creates a new purchase order within a transaction
func (r *PurchaseOrderRepository) CreateWithTx(tx *gorm.DB, po *models.PurchaseOrder) error {
	if err := tx.Create(po).Error; err != nil {
		return errors.NewInternalServerError("Failed to create purchase order")
	}
	return nil
}

// CreateItem creates a purchase order item
func (r *PurchaseOrderRepository) CreateItem(item *models.PurchaseOrderItem) error {
	if err := r.db.Create(item).Error; err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("Failed to create purchase order item: %v", err))
	}
	return nil
}

// CreateItemWithTx creates a purchase order item within a transaction
func (r *PurchaseOrderRepository) CreateItemWithTx(tx *gorm.DB, item *models.PurchaseOrderItem) error {
	if err := tx.Create(item).Error; err != nil {
		return errors.NewInternalServerError("Failed to create purchase order item")
	}
	return nil
}

// GetByID retrieves a purchase order by ID
func (r *PurchaseOrderRepository) GetByID(id string) (*models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	if err := r.db.Where("id = ?", id).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("PurchaseOrder")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve purchase order")
	}
	return &po, nil
}

// GetByIDWithItems retrieves a purchase order by ID with items preloaded
func (r *PurchaseOrderRepository) GetByIDWithItems(id string) (*models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	if err := r.db.Preload("Items.Variant").
		Preload("Collaborator").
		Preload("Warehouse").
		Where("id = ?", id).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("PurchaseOrder")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve purchase order with items")
	}
	return &po, nil
}

// GetByPONumber retrieves a purchase order by PO number
func (r *PurchaseOrderRepository) GetByPONumber(poNumber string) (*models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	if err := r.db.Preload("Items").Where("po_number = ?", poNumber).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("PurchaseOrder")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve purchase order by PO number")
	}
	return &po, nil
}

// GetAll retrieves all purchase orders with pagination
func (r *PurchaseOrderRepository) GetAll(limit, offset int) ([]models.PurchaseOrder, int64, error) {
	var pos []models.PurchaseOrder
	var total int64

	// Get total count
	if err := r.db.Model(&models.PurchaseOrder{}).Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to count purchase orders")
	}

	// Get paginated records
	if err := r.db.Preload("Collaborator").Preload("Warehouse").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&pos).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve purchase orders")
	}
	return pos, total, nil
}

// GetByCollaborator retrieves purchase orders by collaborator ID with pagination
func (r *PurchaseOrderRepository) GetByCollaborator(collaboratorID string, limit, offset int) ([]models.PurchaseOrder, int64, error) {
	var pos []models.PurchaseOrder
	var total int64

	query := r.db.Model(&models.PurchaseOrder{}).Where("collaborator_id = ?", collaboratorID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to count purchase orders by collaborator")
	}

	// Get paginated records
	if err := r.db.Preload("Collaborator").
		Preload("Warehouse").
		Where("collaborator_id = ?", collaboratorID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&pos).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve purchase orders by collaborator")
	}
	return pos, total, nil
}

// GetByStatus retrieves purchase orders by status with pagination
func (r *PurchaseOrderRepository) GetByStatus(status string, limit, offset int) ([]models.PurchaseOrder, int64, error) {
	var pos []models.PurchaseOrder
	var total int64

	query := r.db.Model(&models.PurchaseOrder{}).Where("status = ?", status)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to count purchase orders by status")
	}

	// Get paginated records
	if err := r.db.Preload("Collaborator").
		Preload("Warehouse").
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&pos).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve purchase orders by status")
	}
	return pos, total, nil
}

// GetPendingDeliveries retrieves purchase orders with status not delivered or paid with pagination
func (r *PurchaseOrderRepository) GetPendingDeliveries(limit, offset int) ([]models.PurchaseOrder, int64, error) {
	var pos []models.PurchaseOrder
	var total int64

	query := r.db.Model(&models.PurchaseOrder{}).Where("status NOT IN ?", []string{"delivered", "paid"})

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError(fmt.Sprintf("Failed to count pending deliveries: %v", err))
	}

	// Get paginated records
	if err := r.db.Preload("Collaborator").
		Preload("Warehouse").
		Where("status NOT IN ?", []string{"delivered", "paid"}).
		Order("expected_delivery_date ASC").
		Limit(limit).Offset(offset).
		Find(&pos).Error; err != nil {
		return nil, 0, errors.NewInternalServerError(fmt.Sprintf("Failed to retrieve pending deliveries: %v", err))
	}
	return pos, total, nil
}

// Update updates an existing purchase order
func (r *PurchaseOrderRepository) Update(po *models.PurchaseOrder) error {
	if err := r.db.Save(po).Error; err != nil {
		return errors.NewInternalServerError("Failed to update purchase order")
	}
	return nil
}

// UpdateWithTx updates a purchase order within a transaction
func (r *PurchaseOrderRepository) UpdateWithTx(tx *gorm.DB, po *models.PurchaseOrder) error {
	if err := tx.Save(po).Error; err != nil {
		return errors.NewInternalServerError("Failed to update purchase order")
	}
	return nil
}

// UpdateStatus updates only the status field
func (r *PurchaseOrderRepository) UpdateStatus(poID, status string) error {
	if err := r.db.Model(&models.PurchaseOrder{}).Where("id = ?", poID).Update("status", status).Error; err != nil {
		return errors.NewInternalServerError("Failed to update purchase order status")
	}
	return nil
}

// GetItemByID retrieves a purchase order item by ID
func (r *PurchaseOrderRepository) GetItemByID(itemID string) (*models.PurchaseOrderItem, error) {
	var item models.PurchaseOrderItem
	if err := r.db.Preload("Variant").Where("id = ?", itemID).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("PurchaseOrderItem")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve purchase order item")
	}
	return &item, nil
}

// UpdateItemReceivedQuantity updates the received quantity for a PO item
func (r *PurchaseOrderRepository) UpdateItemReceivedQuantity(itemID string, receivedQty int64) error {
	if err := r.db.Model(&models.PurchaseOrderItem{}).
		Where("id = ?", itemID).
		Update("received_quantity", receivedQty).Error; err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("Failed to update received quantity: %v", err))
	}
	return nil
}

// UpdateItemReceivedQuantityWithTx updates the received quantity for a PO item within a transaction
func (r *PurchaseOrderRepository) UpdateItemReceivedQuantityWithTx(tx *gorm.DB, itemID string, receivedQty int64) error {
	if err := tx.Model(&models.PurchaseOrderItem{}).
		Where("id = ?", itemID).
		Update("received_quantity", receivedQty).Error; err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("Failed to update received quantity: %v", err))
	}
	return nil
}

// PONumberExists checks if a PO number already exists
func (r *PurchaseOrderRepository) PONumberExists(poNumber string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.PurchaseOrder{}).Where("po_number = ?", poNumber).Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check PO number existence")
	}
	return count > 0, nil
}

// FindByExternalOrderID finds a purchase order by external_order_id (for webhook integration)
func (r *PurchaseOrderRepository) FindByExternalOrderID(externalOrderID string) (*models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	if err := r.db.Preload("Items").Where("external_order_id = ?", externalOrderID).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Purchase Order")
		}
		return nil, errors.NewInternalServerError("Failed to find purchase order by external_order_id")
	}
	return &po, nil
}

// CountByCollaborator counts purchase orders for a collaborator
func (r *PurchaseOrderRepository) CountByCollaborator(collaboratorID string) (int64, error) {
	var count int64
	if err := r.db.Model(&models.PurchaseOrder{}).Where("collaborator_id = ?", collaboratorID).Count(&count).Error; err != nil {
		return 0, errors.NewInternalServerError("Failed to count purchase orders by collaborator")
	}
	return count, nil
}

// SumAmountByCollaborator calculates total amount for a collaborator
func (r *PurchaseOrderRepository) SumAmountByCollaborator(collaboratorID string) (float64, error) {
	var total float64
	if err := r.db.Model(&models.PurchaseOrder{}).
		Where("collaborator_id = ?", collaboratorID).
		Select("COALESCE(SUM(total_amount), 0)").
		Row().Scan(&total); err != nil {
		return 0, errors.NewInternalServerError("Failed to sum amounts by collaborator")
	}
	return total, nil
}

// CountActiveByCollaborator counts active purchase orders for a collaborator (not paid or cancelled)
func (r *PurchaseOrderRepository) CountActiveByCollaborator(collaboratorID string) (int64, error) {
	var count int64
	if err := r.db.Model(&models.PurchaseOrder{}).
		Where("collaborator_id = ? AND status NOT IN ?", collaboratorID, []string{"paid", "cancelled"}).
		Count(&count).Error; err != nil {
		return 0, errors.NewInternalServerError("Failed to count active purchase orders by collaborator")
	}
	return count, nil
}

// GetLatestPODate retrieves the date of the most recent purchase order for a collaborator
func (r *PurchaseOrderRepository) GetLatestPODate(collaboratorID string) (*string, error) {
	var dates []time.Time

	err := r.db.Model(&models.PurchaseOrder{}).
		Where("collaborator_id = ?", collaboratorID).
		Order("created_at DESC").
		Limit(1).
		Pluck("created_at", &dates).Error

	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get latest purchase order date")
	}

	// If no results, return nil (valid state - no purchase orders yet)
	if len(dates) == 0 {
		return nil, nil
	}

	// Format the time to RFC3339 string (ISO 8601 format)
	dateStr := dates[0].Format(time.RFC3339)
	return &dateStr, nil
}

// GetAllCollaboratorsStats retrieves PO counts for all collaborators efficiently
func (r *PurchaseOrderRepository) GetAllCollaboratorsStats() (map[string]int64, error) {
	type statsResult struct {
		CollaboratorID string
		POCount        int64
	}

	var results []statsResult
	err := r.db.Model(&models.PurchaseOrder{}).
		Select("collaborator_id, COUNT(*) as po_count").
		Group("collaborator_id").
		Scan(&results).Error
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get collaborators stats")
	}

	statsMap := make(map[string]int64, len(results))
	for _, result := range results {
		statsMap[result.CollaboratorID] = result.POCount
	}

	return statsMap, nil
}

// GetTotalPOCount returns the total count of all purchase orders
func (r *PurchaseOrderRepository) GetTotalPOCount() (int64, error) {
	var total int64
	if err := r.db.Model(&models.PurchaseOrder{}).Count(&total).Error; err != nil {
		return 0, errors.NewInternalServerError("Failed to get total purchase order count")
	}
	return total, nil
}
