package interfaces

import (
	"kisanlink-erp/internal/database/models"
)

// ProductInterface defines the contract for product repository operations
type ProductInterface interface {
	// Create creates a new product
	Create(product *models.Product) error

	// GetByID retrieves a product by ID
	GetByID(id string) (*models.Product, error)

	// GetAll retrieves all products
	GetAll() ([]models.Product, error)

	// Update updates an existing product
	Update(product *models.Product) error

	// Delete deletes a product (soft delete)
	Delete(id string) error

	// Exists checks if a product exists by ID
	Exists(id string) (bool, error)

	// GetByName searches products by name
	GetByName(name string) ([]models.Product, error)

	// FindByExternalID finds a product by external_id (for webhook integration)
	FindByExternalID(externalID string) (*models.Product, error)
}
