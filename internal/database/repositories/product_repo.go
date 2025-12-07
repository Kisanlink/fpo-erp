package repositories

import (
	"fmt"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"

	"gorm.io/gorm"
)

// ProductRepository handles product data access
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create creates a new product
func (r *ProductRepository) Create(product *models.Product) error {
	if err := r.db.Create(product).Error; err != nil {
		fmt.Printf("DEBUG: Database error in ProductRepository.Create: %v\n", err)
		return errors.NewInternalServerError("Failed to create product")
	}
	return nil
}

// GetByID retrieves a product by ID
func (r *ProductRepository) GetByID(id string) (*models.Product, error) {
	var product models.Product
	if err := r.db.Where("id = ?", id).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Product")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve product")
	}
	return &product, nil
}

// GetAll retrieves all products
func (r *ProductRepository) GetAll() ([]models.Product, error) {
	var products []models.Product
	if err := r.db.Find(&products).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve products")
	}
	return products, nil
}

// Update updates a product
func (r *ProductRepository) Update(product *models.Product) error {
	if err := r.db.Save(product).Error; err != nil {
		return errors.NewInternalServerError("Failed to update product")
	}
	return nil
}

// Delete deletes a product
func (r *ProductRepository) Delete(id string) error {
	if err := r.db.Where("id = ?", id).Delete(&models.Product{}).Error; err != nil {
		return errors.NewInternalServerError("Failed to delete product")
	}
	return nil
}

// Exists checks if a product exists by ID
func (r *ProductRepository) Exists(id string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Product{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check product existence")
	}
	return count > 0, nil
}

// GetByName retrieves products by name (for search functionality)
func (r *ProductRepository) GetByName(name string) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.Where("name ILIKE ?", "%"+name+"%").Find(&products).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to search products")
	}
	return products, nil
}

// FindByExternalID finds a product by external_id (for webhook integration)
func (r *ProductRepository) FindByExternalID(externalID string) (*models.Product, error) {
	var product models.Product
	if err := r.db.Where("external_id = ?", externalID).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found, but not an error for smart matching
		}
		return nil, errors.NewInternalServerError("Failed to find product by external_id")
	}
	return &product, nil
}

// GetByCategory retrieves all products in a category by category ID
func (r *ProductRepository) GetByCategory(categoryID string) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.Where("category_id = ?", categoryID).Find(&products).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve products by category")
	}
	return products, nil
}

// GetBySubcategory retrieves all products in a subcategory by subcategory ID
func (r *ProductRepository) GetBySubcategory(subcategoryID string) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.Where("subcategory_id = ?", subcategoryID).Find(&products).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve products by subcategory")
	}
	return products, nil
}

// GetByCategoryAndSubcategory retrieves all products in a specific category and subcategory by IDs
func (r *ProductRepository) GetByCategoryAndSubcategory(categoryID string, subcategoryID *string) ([]models.Product, error) {
	var products []models.Product
	query := r.db.Where("category_id = ?", categoryID)
	if subcategoryID != nil {
		query = query.Where("subcategory_id = ?", *subcategoryID)
	}
	if err := query.Find(&products).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve products by category and subcategory")
	}
	return products, nil
}

// GetWithFilters retrieves products with optional category and subcategory ID filters
func (r *ProductRepository) GetWithFilters(categoryID *string, subcategoryID *string) ([]models.Product, error) {
	var products []models.Product
	query := r.db.Model(&models.Product{})

	if categoryID != nil && *categoryID != "" {
		query = query.Where("category_id = ?", *categoryID)
	}
	if subcategoryID != nil && *subcategoryID != "" {
		query = query.Where("subcategory_id = ?", *subcategoryID)
	}

	if err := query.Find(&products).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve products with filters")
	}
	return products, nil
}
