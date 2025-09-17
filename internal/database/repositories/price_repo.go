package repositories

import (
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"
	"time"

	"gorm.io/gorm"
)

// ProductPriceRepository handles product price data access
type ProductPriceRepository struct {
	db *gorm.DB
}

// NewProductPriceRepository creates a new product price repository
func NewProductPriceRepository(db *gorm.DB) *ProductPriceRepository {
	return &ProductPriceRepository{db: db}
}

// Create creates a new product price
func (r *ProductPriceRepository) Create(price *models.ProductPrice) error {
	if err := r.db.Create(price).Error; err != nil {
		return errors.NewInternalServerError("Failed to create product price")
	}
	return nil
}

// GetByID retrieves a product price by ID
func (r *ProductPriceRepository) GetByID(id string) (*models.ProductPrice, error) {
	var price models.ProductPrice
	if err := r.db.Where("id = ?", id).First(&price).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("ProductPrice")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve product price")
	}
	return &price, nil
}

// GetByProductID retrieves all prices for a product
func (r *ProductPriceRepository) GetByProductID(productID string) ([]models.ProductPrice, error) {
	var prices []models.ProductPrice
	if err := r.db.Where("product_id = ?", productID).Find(&prices).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve product prices")
	}
	return prices, nil
}

// GetActiveByProductID retrieves active prices for a product
func (r *ProductPriceRepository) GetActiveByProductID(productID string) ([]models.ProductPrice, error) {
	var prices []models.ProductPrice
	now := time.Now()

	if err := r.db.Where("product_id = ? AND is_active = ? AND (effective_to IS NULL OR effective_to > ?)",
		productID, true, now).Find(&prices).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve active product prices")
	}
	return prices, nil
}

// GetByProductIDAndType retrieves prices for a product by type
func (r *ProductPriceRepository) GetByProductIDAndType(productID, priceType string) ([]models.ProductPrice, error) {
	var prices []models.ProductPrice
	if err := r.db.Where("product_id = ? AND price_type = ?", productID, priceType).Find(&prices).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve product prices by type")
	}
	return prices, nil
}

// GetCurrentPrice retrieves the current active price for a product and type
func (r *ProductPriceRepository) GetCurrentPrice(productID, priceType string) (*models.ProductPrice, error) {
	var price models.ProductPrice
	now := time.Now()

	if err := r.db.Where("product_id = ? AND price_type = ? AND is_active = ? AND (effective_to IS NULL OR effective_to > ?)",
		productID, priceType, true, now).Order("effective_from DESC").First(&price).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("ProductPrice")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve current product price")
	}
	return &price, nil
}

// GetAll retrieves all product prices
func (r *ProductPriceRepository) GetAll() ([]models.ProductPrice, error) {
	var prices []models.ProductPrice
	if err := r.db.Find(&prices).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve product prices")
	}
	return prices, nil
}

// Update updates a product price
func (r *ProductPriceRepository) Update(price *models.ProductPrice) error {
	if err := r.db.Save(price).Error; err != nil {
		return errors.NewInternalServerError("Failed to update product price")
	}
	return nil
}

// Delete deletes a product price
func (r *ProductPriceRepository) Delete(id string) error {
	if err := r.db.Where("id = ?", id).Delete(&models.ProductPrice{}).Error; err != nil {
		return errors.NewInternalServerError("Failed to delete product price")
	}
	return nil
}

// Exists checks if a product price exists by ID
func (r *ProductPriceRepository) Exists(id string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.ProductPrice{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check product price existence")
	}
	return count > 0, nil
}

// GetExpiredPrices retrieves prices that have expired
func (r *ProductPriceRepository) GetExpiredPrices() ([]models.ProductPrice, error) {
	var prices []models.ProductPrice
	now := time.Now()

	if err := r.db.Where("effective_to IS NOT NULL AND effective_to <= ?", now).Find(&prices).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve expired product prices")
	}
	return prices, nil
}

// GetPricesByType retrieves all prices of a specific type
func (r *ProductPriceRepository) GetPricesByType(priceType string) ([]models.ProductPrice, error) {
	var prices []models.ProductPrice
	if err := r.db.Where("price_type = ?", priceType).Find(&prices).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve product prices by type")
	}
	return prices, nil
}
