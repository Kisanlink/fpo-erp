package repositories

import (
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"

	"gorm.io/gorm"
)

// WarehouseRepository handles warehouse data access
type WarehouseRepository struct {
	db *gorm.DB
}

// NewWarehouseRepository creates a new warehouse repository
func NewWarehouseRepository(db *gorm.DB) *WarehouseRepository {
	return &WarehouseRepository{db: db}
}

// Create creates a new warehouse
func (r *WarehouseRepository) Create(warehouse *models.Warehouse) error {
	if err := r.db.Create(warehouse).Error; err != nil {
		return errors.NewInternalServerError("Failed to create warehouse")
	}
	return nil
}

// GetByID retrieves a warehouse by ID
func (r *WarehouseRepository) GetByID(id string) (*models.Warehouse, error) {
	var warehouse models.Warehouse
	if err := r.db.Where("id = ?", id).First(&warehouse).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Warehouse")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve warehouse")
	}
	return &warehouse, nil
}

// GetAll retrieves all warehouses with pagination
func (r *WarehouseRepository) GetAll(limit, offset int) ([]models.Warehouse, int64, error) {
	var warehouses []models.Warehouse
	var total int64

	// Get total count
	if err := r.db.Model(&models.Warehouse{}).Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to count warehouses")
	}

	// Get paginated records
	if err := r.db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&warehouses).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve warehouses")
	}
	return warehouses, total, nil
}

// Update updates a warehouse
func (r *WarehouseRepository) Update(warehouse *models.Warehouse) error {
	if err := r.db.Save(warehouse).Error; err != nil {
		return errors.NewInternalServerError("Failed to update warehouse")
	}
	return nil
}

// Delete deletes a warehouse
func (r *WarehouseRepository) Delete(id string) error {
	if err := r.db.Where("id = ?", id).Delete(&models.Warehouse{}).Error; err != nil {
		return errors.NewInternalServerError("Failed to delete warehouse")
	}
	return nil
}

// Exists checks if a warehouse exists by ID
func (r *WarehouseRepository) Exists(id string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Warehouse{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check warehouse existence")
	}
	return count > 0, nil
}

// GetByName retrieves warehouses by name (for search functionality)
func (r *WarehouseRepository) GetByName(name string) ([]models.Warehouse, error) {
	var warehouses []models.Warehouse
	if err := r.db.Where("name ILIKE ?", "%"+name+"%").Find(&warehouses).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to search warehouses")
	}
	return warehouses, nil
}

// GetByLocation retrieves warehouses by location
func (r *WarehouseRepository) GetByLocation(location string) ([]models.Warehouse, error) {
	var warehouses []models.Warehouse
	if err := r.db.Where("location ILIKE ?", "%"+location+"%").Find(&warehouses).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to search warehouses by location")
	}
	return warehouses, nil
}
