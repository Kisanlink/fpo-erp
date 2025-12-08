package repositories

import (
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"

	"gorm.io/gorm"
)

// SubcategoryRepository handles subcategory data access
type SubcategoryRepository struct {
	db *gorm.DB
}

// NewSubcategoryRepository creates a new subcategory repository
func NewSubcategoryRepository(db *gorm.DB) *SubcategoryRepository {
	return &SubcategoryRepository{db: db}
}

// Create creates a new subcategory
func (r *SubcategoryRepository) Create(subcategory *models.Subcategory) error {
	if err := r.db.Create(subcategory).Error; err != nil {
		return errors.NewInternalServerError("Failed to create subcategory")
	}
	return nil
}

// GetByID retrieves a subcategory by ID
func (r *SubcategoryRepository) GetByID(id string) (*models.Subcategory, error) {
	var subcategory models.Subcategory
	if err := r.db.Where("id = ?", id).First(&subcategory).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Subcategory")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve subcategory")
	}
	return &subcategory, nil
}

// GetByName retrieves a subcategory by name (exact match, first match)
// Note: Name is not globally unique - same name can exist in different categories.
// Use GetByNameAndCategoryID for precise lookup.
func (r *SubcategoryRepository) GetByName(name string) (*models.Subcategory, error) {
	var subcategory models.Subcategory
	if err := r.db.Where("name = ?", name).First(&subcategory).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found, but not an error for idempotent operations
		}
		return nil, errors.NewInternalServerError("Failed to retrieve subcategory by name")
	}
	return &subcategory, nil
}

// GetByNameAndCategoryID retrieves a subcategory by name and category ID (case-insensitive name)
// This is the correct way to check for existing subcategories since name is unique per category
func (r *SubcategoryRepository) GetByNameAndCategoryID(name string, categoryID string) (*models.Subcategory, error) {
	var subcategory models.Subcategory
	if err := r.db.Where("LOWER(name) = LOWER(?) AND category_id = ?", name, categoryID).First(&subcategory).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found, but not an error for idempotent operations
		}
		return nil, errors.NewInternalServerError("Failed to retrieve subcategory by name and category")
	}
	return &subcategory, nil
}

// GetByCategoryID retrieves all subcategories for a category by ID
func (r *SubcategoryRepository) GetByCategoryID(categoryID string) ([]models.Subcategory, error) {
	var subcategories []models.Subcategory
	if err := r.db.Where("category_id = ?", categoryID).Find(&subcategories).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve subcategories by category")
	}
	return subcategories, nil
}

// GetAll retrieves all subcategories
func (r *SubcategoryRepository) GetAll() ([]models.Subcategory, error) {
	var subcategories []models.Subcategory
	if err := r.db.Find(&subcategories).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve subcategories")
	}
	return subcategories, nil
}

// Update updates a subcategory
func (r *SubcategoryRepository) Update(subcategory *models.Subcategory) error {
	if err := r.db.Save(subcategory).Error; err != nil {
		return errors.NewInternalServerError("Failed to update subcategory")
	}
	return nil
}

// Delete deletes a subcategory
func (r *SubcategoryRepository) Delete(id string) error {
	if err := r.db.Where("id = ?", id).Delete(&models.Subcategory{}).Error; err != nil {
		return errors.NewInternalServerError("Failed to delete subcategory")
	}
	return nil
}

// Exists checks if a subcategory exists by ID
func (r *SubcategoryRepository) Exists(id string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Subcategory{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check subcategory existence")
	}
	return count > 0, nil
}

// ExistsByName checks if a subcategory exists by name
func (r *SubcategoryRepository) ExistsByName(name string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Subcategory{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check subcategory existence by name")
	}
	return count > 0, nil
}

// FirstOrCreate finds or creates a subcategory by name (for idempotent seeding)
func (r *SubcategoryRepository) FirstOrCreate(subcategory *models.Subcategory) error {
	if err := r.db.Where("name = ?", subcategory.Name).FirstOrCreate(subcategory).Error; err != nil {
		return errors.NewInternalServerError("Failed to find or create subcategory")
	}
	return nil
}
