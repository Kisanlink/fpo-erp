package repositories

import (
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"

	"gorm.io/gorm"
)

// CategoryRepository handles category data access
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create creates a new category
func (r *CategoryRepository) Create(category *models.Category) error {
	if err := r.db.Create(category).Error; err != nil {
		return errors.NewInternalServerError("Failed to create category")
	}
	return nil
}

// GetByID retrieves a category by ID
func (r *CategoryRepository) GetByID(id string) (*models.Category, error) {
	var category models.Category
	if err := r.db.Where("id = ?", id).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Category")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve category")
	}
	return &category, nil
}

// GetByName retrieves a category by name
func (r *CategoryRepository) GetByName(name string) (*models.Category, error) {
	var category models.Category
	if err := r.db.Where("name = ?", name).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found, but not an error for idempotent operations
		}
		return nil, errors.NewInternalServerError("Failed to retrieve category by name")
	}
	return &category, nil
}

// GetAll retrieves all categories
func (r *CategoryRepository) GetAll() ([]models.Category, error) {
	var categories []models.Category
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve categories")
	}
	return categories, nil
}

// GetAllWithSubcategories retrieves all categories with their subcategories
func (r *CategoryRepository) GetAllWithSubcategories() ([]models.Category, error) {
	var categories []models.Category
	if err := r.db.Preload("Subcategories").Find(&categories).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve categories with subcategories")
	}
	return categories, nil
}

// Update updates a category
func (r *CategoryRepository) Update(category *models.Category) error {
	if err := r.db.Save(category).Error; err != nil {
		return errors.NewInternalServerError("Failed to update category")
	}
	return nil
}

// Delete deletes a category
func (r *CategoryRepository) Delete(id string) error {
	if err := r.db.Where("id = ?", id).Delete(&models.Category{}).Error; err != nil {
		return errors.NewInternalServerError("Failed to delete category")
	}
	return nil
}

// Exists checks if a category exists by ID
func (r *CategoryRepository) Exists(id string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Category{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check category existence")
	}
	return count > 0, nil
}

// ExistsByName checks if a category exists by name
func (r *CategoryRepository) ExistsByName(name string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Category{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check category existence by name")
	}
	return count > 0, nil
}

// FirstOrCreate finds or creates a category by name (for idempotent seeding)
func (r *CategoryRepository) FirstOrCreate(category *models.Category) error {
	if err := r.db.Where("name = ?", category.Name).FirstOrCreate(category).Error; err != nil {
		return errors.NewInternalServerError("Failed to find or create category")
	}
	return nil
}
