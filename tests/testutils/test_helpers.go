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
func CreateTestVariant(t *testing.T, db *gorm.DB, id, productID, sku, quantity string) *models.ProductVariant {
	t.Helper()

	variant := models.NewProductVariant(productID, sku, quantity, "kg")
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
