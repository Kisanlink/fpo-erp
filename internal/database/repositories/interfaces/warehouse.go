package interfaces

import (
	"kisanlink-erp/internal/database/models"
)

// WarehouseInterface defines the contract for warehouse repository operations
type WarehouseInterface interface {
	// Create creates a new warehouse
	Create(warehouse *models.Warehouse) error

	// GetByID retrieves a warehouse by ID
	GetByID(id string) (*models.Warehouse, error)

	// GetAll retrieves all warehouses with pagination
	GetAll(limit, offset int) ([]models.Warehouse, int64, error)

	// Update updates an existing warehouse
	Update(warehouse *models.Warehouse) error

	// Delete deletes a warehouse (soft delete)
	Delete(id string) error

	// Exists checks if a warehouse exists by ID
	Exists(id string) (bool, error)

	// GetByName searches warehouses by name
	GetByName(name string) ([]models.Warehouse, error)

	// GetByLocation searches warehouses by location
	GetByLocation(location string) ([]models.Warehouse, error)
}
