package interfaces

import (
	"kisanlink-erp/internal/database/models"

	"gorm.io/gorm"
)

// PurchaseOrderInterface defines the contract for purchase order repository operations
type PurchaseOrderInterface interface {
	// WithTransaction executes a function within a transaction
	WithTransaction(fn func(*gorm.DB) error) error

	// Create creates a new purchase order
	Create(po *models.PurchaseOrder) error

	// CreateWithTx creates a new purchase order within a transaction
	CreateWithTx(tx *gorm.DB, po *models.PurchaseOrder) error

	// CreateItem creates a purchase order item
	CreateItem(item *models.PurchaseOrderItem) error

	// CreateItemWithTx creates a purchase order item within a transaction
	CreateItemWithTx(tx *gorm.DB, item *models.PurchaseOrderItem) error

	// GetByID retrieves a purchase order by ID
	GetByID(id string) (*models.PurchaseOrder, error)

	// GetByIDWithItems retrieves a purchase order by ID with items preloaded
	GetByIDWithItems(id string) (*models.PurchaseOrder, error)

	// GetByPONumber retrieves a purchase order by PO number
	GetByPONumber(poNumber string) (*models.PurchaseOrder, error)

	// GetAll retrieves all purchase orders
	GetAll() ([]models.PurchaseOrder, error)

	// GetByCollaborator retrieves purchase orders by collaborator ID
	GetByCollaborator(collaboratorID string) ([]models.PurchaseOrder, error)

	// GetByStatus retrieves purchase orders by status
	GetByStatus(status string) ([]models.PurchaseOrder, error)

	// GetPendingDeliveries retrieves purchase orders pending delivery
	GetPendingDeliveries() ([]models.PurchaseOrder, error)

	// Update updates an existing purchase order
	Update(po *models.PurchaseOrder) error

	// UpdateWithTx updates an existing purchase order within a transaction
	UpdateWithTx(tx *gorm.DB, po *models.PurchaseOrder) error

	// UpdateStatus updates the status of a purchase order
	UpdateStatus(poID, status string) error

	// GetItemByID retrieves a purchase order item by ID
	GetItemByID(itemID string) (*models.PurchaseOrderItem, error)

	// UpdateItemReceivedQuantity updates the received quantity of a purchase order item
	UpdateItemReceivedQuantity(itemID string, receivedQty int64) error

	// UpdateItemReceivedQuantityWithTx updates the received quantity within a transaction
	UpdateItemReceivedQuantityWithTx(tx *gorm.DB, itemID string, receivedQty int64) error

	// PONumberExists checks if a PO number already exists
	PONumberExists(poNumber string) (bool, error)

	// FindByExternalOrderID finds a purchase order by external order ID (for webhook integration)
	FindByExternalOrderID(externalOrderID string) (*models.PurchaseOrder, error)
}
