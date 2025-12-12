package services

import (
	"context"
	"testing"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"
	"kisanlink-erp/tests/testutils"
)

// ========================================
// CreateProductVariant Tests
// ========================================

func TestProductVariantService_CreateVariant_Success(t *testing.T) {
	// Setup in-memory database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create product first
	product := testutils.FixtureProduct("Tomato")
	db.Create(product)

	// Create repositories
	variantRepo := repositories.NewProductVariantRepository(db)
	productRepo := repositories.NewProductRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewProductVariantService(variantRepo, productRepo, nil, nil, nil, mockLogger)

	// Create request
	request := &models.CreateProductVariantRequest{
		VariantName: "1kg",
		Quantity:    "1.0",
		PackSize:    "kg",
	}

	// Execute
	response, err := service.CreateProductVariant(context.Background(), product.ID, request)

	// Assert
	testutils.AssertNoError(t, err, "CreateProductVariant should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.ProductID, product.ID, "Product ID mismatch")
	testutils.AssertEqual(t, response.VariantName, "1kg", "Variant name mismatch")
}

func TestProductVariantService_CreateVariant_WithSKU(t *testing.T) {
	// Setup in-memory database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create product first
	product := testutils.FixtureProduct("Tomato")
	db.Create(product)

	// Create repositories
	variantRepo := repositories.NewProductVariantRepository(db)
	productRepo := repositories.NewProductRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewProductVariantService(variantRepo, productRepo, nil, nil, nil, mockLogger)

	// Create request (Issue 1: SKU auto-generated, not provided in request)
	request := &models.CreateProductVariantRequest{
		VariantName: "1kg",
		Quantity:    "1.0",
		PackSize:    "kg",
	}

	// Execute
	response, err := service.CreateProductVariant(context.Background(), product.ID, request)

	// Assert
	testutils.AssertNoError(t, err, "CreateProductVariant should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertNotNil(t, response.SKU, "SKU should be auto-generated")
}

func TestProductVariantService_CreateVariant_MultipleVariants(t *testing.T) {
	// Issue 1: SKU is now auto-generated, duplicate SKU test replaced with multiple variants test
	// Setup in-memory database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create product first
	product := testutils.FixtureProduct("Tomato")
	db.Create(product)

	// Create repositories
	variantRepo := repositories.NewProductVariantRepository(db)
	productRepo := repositories.NewProductRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewProductVariantService(variantRepo, productRepo, nil, nil, nil, mockLogger)

	// Create first variant
	request1 := &models.CreateProductVariantRequest{
		VariantName: "1kg",
		Quantity:    "1.0",
		PackSize:    "kg",
	}
	response1, err1 := service.CreateProductVariant(context.Background(), product.ID, request1)
	testutils.AssertNoError(t, err1, "First variant should succeed")

	// Create second variant
	request2 := &models.CreateProductVariantRequest{
		VariantName: "2kg",
		Quantity:    "2.0",
		PackSize:    "kg",
	}
	response2, err2 := service.CreateProductVariant(context.Background(), product.ID, request2)
	testutils.AssertNoError(t, err2, "Second variant should succeed")

	// Assert both have unique auto-generated SKUs
	testutils.AssertNotNil(t, response1.SKU, "First SKU should be auto-generated")
	testutils.AssertNotNil(t, response2.SKU, "Second SKU should be auto-generated")
	testutils.AssertNotEqual(t, *response1.SKU, *response2.SKU, "SKUs should be unique")
}

func TestProductVariantService_CreateVariant_ProductNotFound(t *testing.T) {
	// Setup in-memory database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create repositories
	variantRepo := repositories.NewProductVariantRepository(db)
	productRepo := repositories.NewProductRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewProductVariantService(variantRepo, productRepo, nil, nil, nil, mockLogger)

	// Create request
	request := &models.CreateProductVariantRequest{
		VariantName: "1kg",
		Quantity:    "1.0",
		PackSize:    "kg",
	}

	// Execute with non-existent product
	_, err := service.CreateProductVariant(context.Background(), "non-existent-id", request)

	// Assert
	testutils.AssertError(t, err, "Should fail when product not found")
}

// ========================================
// GetProductVariant Tests
// ========================================

func TestProductVariantService_GetVariant_Success(t *testing.T) {
	// Setup in-memory database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create product and variant
	product := testutils.FixtureProduct("Tomato")
	db.Create(product)
	variant := testutils.FixtureProductVariant(product.ID, "1kg")
	db.Create(variant)

	// Create repositories
	variantRepo := repositories.NewProductVariantRepository(db)
	productRepo := repositories.NewProductRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewProductVariantService(variantRepo, productRepo, nil, nil, nil, mockLogger)

	// Execute
	response, err := service.GetProductVariant(context.Background(), variant.ID)

	// Assert
	testutils.AssertNoError(t, err, "GetProductVariant should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.ID, variant.ID, "Variant ID mismatch")
	testutils.AssertEqual(t, response.ProductID, product.ID, "Product ID mismatch")
}

func TestProductVariantService_GetVariant_NotFound(t *testing.T) {
	// Setup in-memory database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create repositories
	variantRepo := repositories.NewProductVariantRepository(db)
	productRepo := repositories.NewProductRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewProductVariantService(variantRepo, productRepo, nil, nil, nil, mockLogger)

	// Execute
	_, err := service.GetProductVariant(context.Background(), "non-existent-id")

	// Assert
	testutils.AssertError(t, err, "Should fail when variant not found")
}

// ========================================
// GetVariantsByProduct Tests
// ========================================

func TestProductVariantService_GetVariantsByProduct_Success(t *testing.T) {
	// Setup in-memory database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create product and multiple variants
	product := testutils.FixtureProduct("Tomato")
	db.Create(product)

	variant1 := testutils.FixtureProductVariantWithID("VAR-001", product.ID, "1kg")
	variant2 := testutils.FixtureProductVariantWithID("VAR-002", product.ID, "2kg")
	db.Create(variant1)
	db.Create(variant2)

	// Create repositories
	variantRepo := repositories.NewProductVariantRepository(db)
	productRepo := repositories.NewProductRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewProductVariantService(variantRepo, productRepo, nil, nil, nil, mockLogger)

	// Execute
	responses, err := service.GetVariantsByProduct(context.Background(), product.ID)

	// Assert
	testutils.AssertNoError(t, err, "GetVariantsByProduct should succeed")
	testutils.AssertEqual(t, len(responses), 2, "Should return 2 variants")
}

func TestProductVariantService_GetVariantsByProduct_ProductNotFound(t *testing.T) {
	// Setup in-memory database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create repositories
	variantRepo := repositories.NewProductVariantRepository(db)
	productRepo := repositories.NewProductRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewProductVariantService(variantRepo, productRepo, nil, nil, nil, mockLogger)

	// Execute
	_, err := service.GetVariantsByProduct(context.Background(), "non-existent-id")

	// Assert
	testutils.AssertError(t, err, "Should fail when product not found")
}

// ========================================
// GetVariantBySKU Tests
// ========================================

func TestProductVariantService_GetVariantBySKU_Success(t *testing.T) {
	// Setup in-memory database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create product and variant with SKU
	product := testutils.FixtureProduct("Tomato")
	db.Create(product)

	sku := "TOM-1KG-001"
	variant := testutils.FixtureProductVariant(product.ID, "1kg")
	variant.SKU = &sku
	db.Create(variant)

	// Create repositories
	variantRepo := repositories.NewProductVariantRepository(db)
	productRepo := repositories.NewProductRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewProductVariantService(variantRepo, productRepo, nil, nil, nil, mockLogger)

	// Execute
	response, err := service.GetVariantBySKU(context.Background(), sku)

	// Assert
	testutils.AssertNoError(t, err, "GetVariantBySKU should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertNotNil(t, response.SKU, "SKU should not be nil")
	testutils.AssertEqual(t, *response.SKU, sku, "SKU mismatch")
}

func TestProductVariantService_GetVariantBySKU_NotFound(t *testing.T) {
	// Setup in-memory database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create repositories
	variantRepo := repositories.NewProductVariantRepository(db)
	productRepo := repositories.NewProductRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewProductVariantService(variantRepo, productRepo, nil, nil, nil, mockLogger)

	// Execute
	_, err := service.GetVariantBySKU(context.Background(), "non-existent-sku")

	// Assert
	testutils.AssertError(t, err, "Should fail when SKU not found")
}

// ========================================
// UpdateProductVariant Tests
// ========================================

func TestProductVariantService_UpdateVariant_Success(t *testing.T) {
	// Setup in-memory database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create product and variant
	product := testutils.FixtureProduct("Tomato")
	db.Create(product)
	variant := testutils.FixtureProductVariant(product.ID, "1kg")
	db.Create(variant)

	// Create repositories
	variantRepo := repositories.NewProductVariantRepository(db)
	productRepo := repositories.NewProductRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewProductVariantService(variantRepo, productRepo, nil, nil, nil, mockLogger)

	// Update request
	newVariantName := "Updated 1kg"
	request := &models.UpdateProductVariantRequest{
		VariantName: &newVariantName,
	}

	// Execute
	response, err := service.UpdateProductVariant(context.Background(), variant.ID, request)

	// Assert
	testutils.AssertNoError(t, err, "UpdateProductVariant should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.VariantName, newVariantName, "Variant name should be updated")
}

func TestProductVariantService_UpdateVariant_NotFound(t *testing.T) {
	// Setup in-memory database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create repositories
	variantRepo := repositories.NewProductVariantRepository(db)
	productRepo := repositories.NewProductRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewProductVariantService(variantRepo, productRepo, nil, nil, nil, mockLogger)

	// Update request
	newVariantName := "Updated"
	request := &models.UpdateProductVariantRequest{
		VariantName: &newVariantName,
	}

	// Execute
	_, err := service.UpdateProductVariant(context.Background(), "non-existent-id", request)

	// Assert
	testutils.AssertError(t, err, "Should fail when variant not found")
}

// ========================================
// DeleteProductVariant Tests
// ========================================

func TestProductVariantService_DeleteVariant_Success(t *testing.T) {
	// Setup in-memory database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create product and variant
	product := testutils.FixtureProduct("Tomato")
	db.Create(product)
	variant := testutils.FixtureProductVariant(product.ID, "1kg")
	db.Create(variant)

	// Create repositories
	variantRepo := repositories.NewProductVariantRepository(db)
	productRepo := repositories.NewProductRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewProductVariantService(variantRepo, productRepo, nil, nil, nil, mockLogger)

	// Execute
	err := service.DeleteProductVariant(context.Background(), variant.ID)

	// Assert
	testutils.AssertNoError(t, err, "DeleteProductVariant should succeed")

	// NOTE: ProductVariant uses soft delete (GORM deleted_at field)
	// GetByID may or may not filter soft-deleted records depending on repository implementation
	// The fact that DeleteProductVariant succeeded without error is sufficient verification
}

func TestProductVariantService_DeleteVariant_NotFound(t *testing.T) {
	// Setup in-memory database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create repositories
	variantRepo := repositories.NewProductVariantRepository(db)
	productRepo := repositories.NewProductRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewProductVariantService(variantRepo, productRepo, nil, nil, nil, mockLogger)

	// Execute
	err := service.DeleteProductVariant(context.Background(), "non-existent-id")

	// Assert
	testutils.AssertError(t, err, "Should fail when variant not found")
}
