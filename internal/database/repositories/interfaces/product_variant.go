package interfaces

import (
	"kisanlink-erp/internal/database/models"
)

// ProductVariantInterface defines the contract for product variant repository operations
type ProductVariantInterface interface {
	// Create creates a new product variant
	Create(variant *models.ProductVariant) error

	// GetByID retrieves a product variant by ID
	GetByID(id string) (*models.ProductVariant, error)

	// GetByProductID retrieves all variants for a product
	GetByProductID(productID string) ([]models.ProductVariant, error)

	// GetBySKU retrieves a product variant by SKU
	GetBySKU(sku string) (*models.ProductVariant, error)

	// GetByBarcode retrieves a product variant by barcode
	GetByBarcode(barcode string) (*models.ProductVariant, error)

	// Update updates an existing product variant
	Update(variant *models.ProductVariant) error

	// Delete deletes a product variant (soft delete)
	Delete(id string) error

	// SKUExists checks if a SKU already exists
	SKUExists(sku string) (bool, error)

	// BarcodeExists checks if a barcode already exists
	BarcodeExists(barcode string) (bool, error)

	// FindByExternalID finds a variant by external_id (for webhook integration)
	FindByExternalID(externalID string) (*models.ProductVariant, error)

	// FindBySKU finds a variant by SKU (alias for GetBySKU for webhook compatibility)
	FindBySKU(sku string) (*models.ProductVariant, error)
}
