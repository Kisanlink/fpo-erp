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

	// GetAll retrieves all purchase orders with pagination
	GetAll(limit, offset int) ([]models.PurchaseOrder, int64, error)

	// GetByCollaborator retrieves purchase orders by collaborator ID with pagination
	GetByCollaborator(collaboratorID string, limit, offset int) ([]models.PurchaseOrder, int64, error)

	// GetByStatus retrieves purchase orders by status with pagination
	GetByStatus(status string, limit, offset int) ([]models.PurchaseOrder, int64, error)

	// GetPendingDeliveries retrieves purchase orders pending delivery with pagination
	GetPendingDeliveries(limit, offset int) ([]models.PurchaseOrder, int64, error)

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

	// CountByCollaborator counts purchase orders for a collaborator
	CountByCollaborator(collaboratorID string) (int64, error)

	// SumAmountByCollaborator calculates total amount for a collaborator
	SumAmountByCollaborator(collaboratorID string) (float64, error)

	// CountActiveByCollaborator counts active purchase orders for a collaborator
	CountActiveByCollaborator(collaboratorID string) (int64, error)

	// GetLatestPODate retrieves the date of the most recent purchase order for a collaborator
	GetLatestPODate(collaboratorID string) (*string, error)

	// GetAllCollaboratorsStats retrieves PO counts for all collaborators
	GetAllCollaboratorsStats() (map[string]int64, error)

	// GetTotalPOCount returns the total count of all purchase orders
	GetTotalPOCount() (int64, error)
}
