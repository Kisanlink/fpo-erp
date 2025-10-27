package repositories

import (
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"

	"gorm.io/gorm"
)

type ProductVariantRepository struct {
	db *gorm.DB
}

func NewProductVariantRepository(db *gorm.DB) *ProductVariantRepository {
	return &ProductVariantRepository{db: db}
}

// Create creates a new product variant
func (r *ProductVariantRepository) Create(variant *models.ProductVariant) error {
	if err := r.db.Create(variant).Error; err != nil {
		return errors.NewInternalServerError("Failed to create product variant")
	}
	return nil
}

// GetByID retrieves a product variant by ID
func (r *ProductVariantRepository) GetByID(id string) (*models.ProductVariant, error) {
	var variant models.ProductVariant
	if err := r.db.Preload("Product").Where("id = ?", id).First(&variant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("ProductVariant")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve product variant")
	}
	return &variant, nil
}

// GetByProductID retrieves all variants for a product
func (r *ProductVariantRepository) GetByProductID(productID string) ([]models.ProductVariant, error) {
	var variants []models.ProductVariant
	if err := r.db.Where("product_id = ? AND is_active = ?", productID, true).Find(&variants).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve variants by product")
	}
	return variants, nil
}

// GetBySKU retrieves a product variant by SKU
func (r *ProductVariantRepository) GetBySKU(sku string) (*models.ProductVariant, error) {
	var variant models.ProductVariant
	if err := r.db.Preload("Product").Where("sku = ?", sku).First(&variant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("ProductVariant")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve product variant by SKU")
	}
	return &variant, nil
}

// GetByBarcode retrieves a product variant by barcode
func (r *ProductVariantRepository) GetByBarcode(barcode string) (*models.ProductVariant, error) {
	var variant models.ProductVariant
	if err := r.db.Preload("Product").Where("barcode = ?", barcode).First(&variant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("ProductVariant")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve product variant by barcode")
	}
	return &variant, nil
}

// Update updates an existing product variant
func (r *ProductVariantRepository) Update(variant *models.ProductVariant) error {
	if err := r.db.Save(variant).Error; err != nil {
		return errors.NewInternalServerError("Failed to update product variant")
	}
	return nil
}

// Delete deletes a product variant (soft delete by setting is_active = false)
func (r *ProductVariantRepository) Delete(id string) error {
	if err := r.db.Model(&models.ProductVariant{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return errors.NewInternalServerError("Failed to delete product variant")
	}
	return nil
}

// SKUExists checks if a variant SKU already exists
func (r *ProductVariantRepository) SKUExists(sku string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.ProductVariant{}).Where("sku = ?", sku).Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check variant SKU existence")
	}
	return count > 0, nil
}

// BarcodeExists checks if a barcode already exists
func (r *ProductVariantRepository) BarcodeExists(barcode string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.ProductVariant{}).Where("barcode = ?", barcode).Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check barcode existence")
	}
	return count > 0, nil
}
