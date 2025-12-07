package testutils

import (
	"testing"

	"kisanlink-erp/internal/database/models"

	"gorm.io/gorm"
)

// CreateTestProduct creates a test product and saves it to the database
func CreateTestProduct(t *testing.T, db *gorm.DB, id, name string) *models.Product {
	t.Helper()

	desc := "Test product " + name
	product := models.NewProduct(name, &desc)
	product.ID = id

	if err := db.Create(product).Error; err != nil {
		t.Fatalf("failed to create test product: %v", err)
	}

	return product
}

// CreateTestVariant creates a test product variant and saves it to the database
// GST-only tax system: HSNCode and GSTRate are required on variants
func CreateTestVariant(t *testing.T, db *gorm.DB, id, productID, sku, quantity string) *models.ProductVariant {
	t.Helper()

	// Default HSNCode and GSTRate for tests (GST-only tax system)
	variant := models.NewProductVariant(productID, sku, quantity, "kg", "12345678", 18.0)
	variant.ID = id
	variant.SKU = &sku

	if err := db.Create(variant).Error; err != nil {
		t.Fatalf("failed to create test variant: %v", err)
	}

	return variant
}

// CreateTestVariantSimple creates a test product variant with default quantity and saves it to the database
func CreateTestVariantSimple(t *testing.T, db *gorm.DB, id, productID, sku string) *models.ProductVariant {
	return CreateTestVariant(t, db, id, productID, sku, "1.0")
}

// CreateTestVariantWithGST creates a test product variant with custom GST settings
func CreateTestVariantWithGST(t *testing.T, db *gorm.DB, id, productID, sku, quantity, hsnCode string, gstRate float64) *models.ProductVariant {
	t.Helper()

	variant := models.NewProductVariant(productID, sku, quantity, "kg", hsnCode, gstRate)
	variant.ID = id
	variant.SKU = &sku

	if err := db.Create(variant).Error; err != nil {
		t.Fatalf("failed to create test variant: %v", err)
	}

	return variant
}

// CreateTestWarehouse creates a test warehouse and saves it to the database
func CreateTestWarehouse(t *testing.T, db *gorm.DB, id string) *models.Warehouse {
	t.Helper()

	warehouse := models.NewWarehouse("Test Warehouse "+id, nil)
	warehouse.ID = id

	if err := db.Create(warehouse).Error; err != nil {
		t.Fatalf("failed to create test warehouse: %v", err)
	}

	return warehouse
}
