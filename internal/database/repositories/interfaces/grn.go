package interfaces

import (
	"kisanlink-erp/internal/database/models"

	"gorm.io/gorm"
)

// GRNInterface defines the contract for GRN repository operations
type GRNInterface interface {
	// WithTransaction executes a function within a transaction
	WithTransaction(fn func(*gorm.DB) error) error

	// Create creates a new GRN
	Create(grn *models.GRN) error

	// CreateWithTx creates a new GRN within a transaction
	CreateWithTx(tx *gorm.DB, grn *models.GRN) error

	// CreateItem creates a GRN item
	CreateItem(item *models.GRNItem) error

	// CreateItemWithTx creates a GRN item within a transaction
	CreateItemWithTx(tx *gorm.DB, item *models.GRNItem) error

	// UpdateItemBatchIDWithTx updates the inventory batch ID for a GRN item within a transaction
	UpdateItemBatchIDWithTx(tx *gorm.DB, itemID, batchID string) error

	// Update updates a GRN
	Update(id string, updates map[string]interface{}) error

	// GetByID retrieves a GRN by ID
	GetByID(id string) (*models.GRN, error)

	// GetByIDWithItems retrieves a GRN by ID with items preloaded
	GetByIDWithItems(id string) (*models.GRN, error)

	// GetByGRNNumber retrieves a GRN by GRN number
	GetByGRNNumber(grnNumber string) (*models.GRN, error)

	// GetByPurchaseOrder retrieves a GRN by purchase order ID
	GetByPurchaseOrder(poID string) (*models.GRN, error)

	// GetAll retrieves all GRNs
	GetAll() ([]models.GRN, error)

	// GetByWarehouse retrieves GRNs by warehouse ID
	GetByWarehouse(warehouseID string) ([]models.GRN, error)

	// GRNNumberExists checks if a GRN number already exists
	GRNNumberExists(grnNumber string) (bool, error)

	// GRNExistsForPO checks if a GRN already exists for a purchase order
	GRNExistsForPO(poID string) (bool, error)
}
