package repositories

import (
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"

	"gorm.io/gorm"
)

type GRNRepository struct {
	db *gorm.DB
}

func NewGRNRepository(db *gorm.DB) *GRNRepository {
	return &GRNRepository{db: db}
}

// WithTransaction executes a function within a transaction
func (r *GRNRepository) WithTransaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// Create creates a new GRN
func (r *GRNRepository) Create(grn *models.GRN) error {
	if err := r.db.Create(grn).Error; err != nil {
		return errors.NewInternalServerError("Failed to create GRN")
	}
	return nil
}

// CreateWithTx creates a new GRN within a transaction
func (r *GRNRepository) CreateWithTx(tx *gorm.DB, grn *models.GRN) error {
	if err := tx.Create(grn).Error; err != nil {
		return errors.NewInternalServerError("Failed to create GRN")
	}
	return nil
}

// CreateItem creates a GRN item
func (r *GRNRepository) CreateItem(item *models.GRNItem) error {
	if err := r.db.Create(item).Error; err != nil {
		return errors.NewInternalServerError("Failed to create GRN item")
	}
	return nil
}

// CreateItemWithTx creates a GRN item within a transaction
func (r *GRNRepository) CreateItemWithTx(tx *gorm.DB, item *models.GRNItem) error {
	if err := tx.Create(item).Error; err != nil {
		return errors.NewInternalServerError("Failed to create GRN item")
	}
	return nil
}

// UpdateItemBatchIDWithTx updates the inventory batch ID for a GRN item within a transaction
func (r *GRNRepository) UpdateItemBatchIDWithTx(tx *gorm.DB, itemID, batchID string) error {
	if err := tx.Model(&models.GRNItem{}).
		Where("id = ?", itemID).
		Update("inventory_batch_id", batchID).Error; err != nil {
		return errors.NewInternalServerError("Failed to update GRN item batch ID")
	}
	return nil
}

// Update updates a GRN
func (r *GRNRepository) Update(id string, updates map[string]interface{}) error {
	if err := r.db.Model(&models.GRN{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return errors.NewInternalServerError("Failed to update GRN")
	}
	return nil
}

// GetByID retrieves a GRN by ID
func (r *GRNRepository) GetByID(id string) (*models.GRN, error) {
	var grn models.GRN
	if err := r.db.Where("id = ?", id).First(&grn).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("GRN")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve GRN")
	}
	return &grn, nil
}

// GetByIDWithItems retrieves a GRN by ID with items preloaded
func (r *GRNRepository) GetByIDWithItems(id string) (*models.GRN, error) {
	var grn models.GRN
	if err := r.db.Preload("Items.Variant").
		Preload("Items.PurchaseOrderItem").
		Preload("PurchaseOrder").
		Preload("Warehouse").
		Where("id = ?", id).First(&grn).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("GRN")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve GRN with items")
	}
	return &grn, nil
}

// GetByGRNNumber retrieves a GRN by GRN number
func (r *GRNRepository) GetByGRNNumber(grnNumber string) (*models.GRN, error) {
	var grn models.GRN
	if err := r.db.Preload("Items").Where("grn_number = ?", grnNumber).First(&grn).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("GRN")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve GRN by GRN number")
	}
	return &grn, nil
}

// GetByPurchaseOrder retrieves a GRN by purchase order ID
func (r *GRNRepository) GetByPurchaseOrder(poID string) (*models.GRN, error) {
	var grn models.GRN
	if err := r.db.Preload("Items").Where("po_id = ?", poID).First(&grn).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("GRN")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve GRN by purchase order")
	}
	return &grn, nil
}

// GetAll retrieves all GRNs
func (r *GRNRepository) GetAll() ([]models.GRN, error) {
	var grns []models.GRN
	if err := r.db.Preload("PurchaseOrder").
		Preload("Warehouse").
		Order("received_date DESC").
		Find(&grns).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve GRNs")
	}
	return grns, nil
}

// GetByWarehouse retrieves GRNs by warehouse ID
func (r *GRNRepository) GetByWarehouse(warehouseID string) ([]models.GRN, error) {
	var grns []models.GRN
	if err := r.db.Preload("PurchaseOrder").
		Where("warehouse_id = ?", warehouseID).
		Order("received_date DESC").
		Find(&grns).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve GRNs by warehouse")
	}
	return grns, nil
}

// GRNNumberExists checks if a GRN number already exists
func (r *GRNRepository) GRNNumberExists(grnNumber string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.GRN{}).Where("grn_number = ?", grnNumber).Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check GRN number existence")
	}
	return count > 0, nil
}

// GRNExistsForPO checks if a GRN already exists for a purchase order
func (r *GRNRepository) GRNExistsForPO(poID string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.GRN{}).Where("po_id = ?", poID).Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check GRN existence for PO")
	}
	return count > 0, nil
}

// GetRejectedItemsByGRN retrieves all rejected items for a GRN
func (r *GRNRepository) GetRejectedItemsByGRN(grnID string) ([]models.GRNItem, error) {
	var items []models.GRNItem
	if err := r.db.Preload("Variant").
		Preload("PurchaseOrderItem").
		Where("grn_id = ? AND rejected_quantity > 0", grnID).
		Find(&items).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve rejected items")
	}
	return items, nil
}

// GetItemByID retrieves a single GRN item by ID
func (r *GRNRepository) GetItemByID(itemID string) (*models.GRNItem, error) {
	var item models.GRNItem
	if err := r.db.Preload("Variant").
		Preload("PurchaseOrderItem").
		Where("id = ?", itemID).
		First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("GRN item")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve GRN item")
	}
	return &item, nil
}

// UpdateItemReturnStatus updates the return status of a GRN item
func (r *GRNRepository) UpdateItemReturnStatus(itemID string, updates map[string]interface{}) error {
	if err := r.db.Model(&models.GRNItem{}).
		Where("id = ?", itemID).
		Updates(updates).Error; err != nil {
		return errors.NewInternalServerError("Failed to update item return status")
	}
	return nil
}

// GetTotalRejectedAmountByPO calculates total value of rejected items for a purchase order
func (r *GRNRepository) GetTotalRejectedAmountByPO(poID string) (float64, error) {
	var total float64

	// Join GRN → GRN items → PO items to get unit prices
	err := r.db.Table("grn_items").
		Select("COALESCE(SUM(grn_items.rejected_quantity * purchase_order_items.unit_price), 0) as total").
		Joins("JOIN goods_receipt_notes ON grn_items.grn_id = goods_receipt_notes.id").
		Joins("JOIN purchase_order_items ON grn_items.po_item_id = purchase_order_items.id").
		Where("goods_receipt_notes.po_id = ?", poID).
		Row().Scan(&total)

	if err != nil {
		return 0, errors.NewInternalServerError("Failed to calculate total rejected amount")
	}

	return total, nil
}
