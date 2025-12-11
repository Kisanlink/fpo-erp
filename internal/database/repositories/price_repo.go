package repositories

import (
	"fmt"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"

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
		return errors.NewInternalServerError(fmt.Sprintf("Failed to create product price: %v", err))
	}
	return nil
}

// CreateWithTx creates a new product price within a transaction
func (r *ProductPriceRepository) CreateWithTx(tx *gorm.DB, price *models.ProductPrice) error {
	if err := tx.Create(price).Error; err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("Failed to create product price: %v", err))
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

// GetByVariantID retrieves all prices for a variant
func (r *ProductPriceRepository) GetByVariantID(variantID string) ([]models.ProductPrice, error) {
	var prices []models.ProductPrice
	if err := r.db.Where("variant_id = ?", variantID).Find(&prices).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve variant prices")
	}
	return prices, nil
}

// GetActiveByVariantID retrieves active prices for a variant
func (r *ProductPriceRepository) GetActiveByVariantID(variantID string) ([]models.ProductPrice, error) {
	var prices []models.ProductPrice
	now := time.Now()

	if err := r.db.Where("variant_id = ? AND is_active = ? AND (effective_to IS NULL OR effective_to > ?)",
		variantID, true, now).Find(&prices).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve active variant prices")
	}
	return prices, nil
}

// GetByVariantIDAndType retrieves prices for a variant by type
func (r *ProductPriceRepository) GetByVariantIDAndType(variantID, priceType string) ([]models.ProductPrice, error) {
	var prices []models.ProductPrice
	if err := r.db.Where("variant_id = ? AND price_type = ?", variantID, priceType).Find(&prices).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve variant prices by type")
	}
	return prices, nil
}

// GetCurrentPrice retrieves the current active price for a variant and type
func (r *ProductPriceRepository) GetCurrentPrice(variantID, priceType string) (*models.ProductPrice, error) {
	var price models.ProductPrice
	now := time.Now()

	if err := r.db.Where("variant_id = ? AND price_type = ? AND is_active = ? AND effective_from <= ? AND (effective_to IS NULL OR effective_to > ?)",
		variantID, priceType, true, now, now).Order("effective_from DESC").First(&price).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("ProductPrice")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve current variant price")
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

// GetByVariantIDPaginated retrieves prices for a variant with pagination
func (r *ProductPriceRepository) GetByVariantIDPaginated(variantID string, limit, offset int) ([]models.ProductPrice, int64, error) {
	var prices []models.ProductPrice
	var total int64

	query := r.db.Model(&models.ProductPrice{}).Where("variant_id = ?", variantID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to count variant prices")
	}

	if err := query.Limit(limit).Offset(offset).Order("effective_from DESC").Find(&prices).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve variant prices")
	}

	return prices, total, nil
}

// GetExpiredPricesPaginated retrieves expired prices with pagination
func (r *ProductPriceRepository) GetExpiredPricesPaginated(limit, offset int) ([]models.ProductPrice, int64, error) {
	var prices []models.ProductPrice
	var total int64
	now := time.Now()

	query := r.db.Model(&models.ProductPrice{}).Where("effective_to IS NOT NULL AND effective_to <= ?", now)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to count expired product prices")
	}

	if err := query.Limit(limit).Offset(offset).Order("effective_to DESC").Find(&prices).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve expired product prices")
	}

	return prices, total, nil
}
