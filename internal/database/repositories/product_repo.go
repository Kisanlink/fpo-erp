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

// GetBySKU retrieves a product by SKU
func (r *ProductRepository) GetBySKU(sku string) (*models.Product, error) {
	var product models.Product
	if err := r.db.Where("sku = ?", sku).First(&product).Error; err != nil {
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

// SKUExists checks if a SKU already exists
func (r *ProductRepository) SKUExists(sku string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Product{}).Where("sku = ?", sku).Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check SKU existence")
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
