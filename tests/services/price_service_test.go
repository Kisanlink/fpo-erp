package services

import (
	"testing"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"
	"kisanlink-erp/tests/testutils"

	"gorm.io/gorm"
)

// ========================================
// Setup and Helper Functions
// ========================================

// setupPriceService creates a ProductPriceService with in-memory database
func setupPriceService(t *testing.T) (*services.ProductPriceService, *gorm.DB, func()) {
	t.Helper()

	db := testutils.SetupTestDB(t)

	// Create repositories
	priceRepo := repositories.NewProductPriceRepository(db)
	productRepo := repositories.NewProductRepository(db)
	variantRepo := repositories.NewProductVariantRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewProductPriceService(priceRepo, productRepo, variantRepo, mockLogger)

	cleanup := func() {
		testutils.CleanupTestDB(db)
	}

	return service, db, cleanup
}

// ========================================
// CreateProductPrice Tests
// ========================================

func TestPriceService_CreateProductPrice_Success(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	request := &models.CreateProductPriceRequest{
		VariantID: variant.ID,
		PriceType: "retail",
		Price:     100.50,
		Currency:  "INR",
	}

	response, err := service.CreateProductPrice(request)

	testutils.AssertNoError(t, err, "CreateProductPrice should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.VariantID, variant.ID, "Variant ID mismatch")
	testutils.AssertEqual(t, response.PriceType, "retail", "PriceType mismatch")
	testutils.AssertEqual(t, response.Price, 100.50, "Price mismatch")
	testutils.AssertEqual(t, response.Currency, "INR", "Currency mismatch")
	testutils.AssertEqual(t, response.IsActive, true, "IsActive should be true by default")
}

func TestPriceService_CreateProductPrice_WithDefaults(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	request := &models.CreateProductPriceRequest{
		VariantID: variant.ID,
		PriceType: "retail",
		Price:     100.50,
		// Currency not specified - should default to INR
		// IsActive not specified - should default to true
	}

	response, err := service.CreateProductPrice(request)

	testutils.AssertNoError(t, err, "CreateProductPrice should succeed")
	testutils.AssertEqual(t, response.Currency, "INR", "Currency should default to INR")
	testutils.AssertEqual(t, response.IsActive, true, "IsActive should default to true")
}

func TestPriceService_CreateProductPrice_WithOptionalFields(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	effectiveFrom := time.Now().UTC().Format(time.RFC3339)
	effectiveTo := time.Now().UTC().Add(30 * 24 * time.Hour).Format(time.RFC3339)
	isActive := false

	request := &models.CreateProductPriceRequest{
		VariantID:     variant.ID,
		PriceType:     "retail",
		Price:         100.50,
		EffectiveFrom: &effectiveFrom,
		EffectiveTo:   &effectiveTo,
		IsActive:      &isActive,
	}

	response, err := service.CreateProductPrice(request)

	testutils.AssertNoError(t, err, "CreateProductPrice should succeed")
	testutils.AssertNotNil(t, response.EffectiveTo, "EffectiveTo should be set")
	testutils.AssertEqual(t, response.IsActive, false, "IsActive should be false as specified")
}

func TestPriceService_CreateProductPrice_WithoutEffectiveTo(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	request := &models.CreateProductPriceRequest{
		VariantID: variant.ID,
		PriceType: "retail",
		Price:     100.50,
		// No EffectiveTo - price is indefinite
	}

	response, err := service.CreateProductPrice(request)

	testutils.AssertNoError(t, err, "CreateProductPrice should succeed")
	testutils.AssertNil(t, response.EffectiveTo, "EffectiveTo should be nil for indefinite prices")
}

func TestPriceService_CreateProductPrice_InvalidVariant(t *testing.T) {
	service, _, cleanup := setupPriceService(t)
	defer cleanup()

	request := &models.CreateProductPriceRequest{
		VariantID: "INVALID-VARIANT",
		PriceType: "retail",
		Price:     100.50,
	}

	_, err := service.CreateProductPrice(request)

	testutils.AssertError(t, err, "Should fail for invalid variant")
}

func TestPriceService_CreateProductPrice_InvalidEffectiveFromFormat(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	invalidDate := "2024-13-45" // Invalid date
	request := &models.CreateProductPriceRequest{
		VariantID:     variant.ID,
		PriceType:     "retail",
		Price:         100.50,
		EffectiveFrom: &invalidDate,
	}

	// Service handles invalid dates by using current time - no error
	response, err := service.CreateProductPrice(request)

	testutils.AssertNoError(t, err, "CreateProductPrice should succeed even with invalid date format")
	testutils.AssertNotNil(t, response, "Response should not be nil")
}

func TestPriceService_CreateProductPrice_InvalidEffectiveToFormat(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	invalidDate := "invalid-date-format"
	request := &models.CreateProductPriceRequest{
		VariantID:   variant.ID,
		PriceType:   "retail",
		Price:       100.50,
		EffectiveTo: &invalidDate,
	}

	// Service handles invalid dates gracefully - no error
	response, err := service.CreateProductPrice(request)

	testutils.AssertNoError(t, err, "CreateProductPrice should succeed even with invalid date format")
	testutils.AssertNil(t, response.EffectiveTo, "EffectiveTo should be nil when parse fails")
}

func TestPriceService_CreateProductPrice_MultiplePriceTypes(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	// Create retail price
	retailRequest := &models.CreateProductPriceRequest{
		VariantID: variant.ID,
		PriceType: "retail",
		Price:     100.50,
	}

	retailResponse, err := service.CreateProductPrice(retailRequest)
	testutils.AssertNoError(t, err, "Creating retail price should succeed")

	// Create wholesale price
	wholesaleRequest := &models.CreateProductPriceRequest{
		VariantID: variant.ID,
		PriceType: "wholesale",
		Price:     80.00,
	}

	wholesaleResponse, err := service.CreateProductPrice(wholesaleRequest)
	testutils.AssertNoError(t, err, "Creating wholesale price should succeed")

	testutils.AssertNotEqual(t, retailResponse.ID, wholesaleResponse.ID, "Prices should have different IDs")
	testutils.AssertEqual(t, retailResponse.Price, 100.50, "Retail price mismatch")
	testutils.AssertEqual(t, wholesaleResponse.Price, 80.00, "Wholesale price mismatch")
}

// ========================================
// GetProductPrice Tests
// ========================================

func TestPriceService_GetProductPrice_Success(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	// Create a price
	price := testutils.FixtureProductPrice(variant.ID, "retail", 100.50)
	db.Create(price)

	response, err := service.GetProductPrice(price.ID)

	testutils.AssertNoError(t, err, "GetProductPrice should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.ID, price.ID, "Price ID mismatch")
	testutils.AssertEqual(t, response.VariantID, variant.ID, "Variant ID mismatch")
	testutils.AssertEqual(t, response.Price, 100.50, "Price mismatch")
}

func TestPriceService_GetProductPrice_NotFound(t *testing.T) {
	service, _, cleanup := setupPriceService(t)
	defer cleanup()

	_, err := service.GetProductPrice("non-existent-id")

	testutils.AssertError(t, err, "Should fail when price not found")
}

// ========================================
// GetVariantPrices Tests
// ========================================

func TestPriceService_GetVariantPrices_MultiplePrices(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	// Create multiple prices for the variant
	price1 := testutils.FixtureProductPrice(variant.ID, "retail", 100.50)
	price2 := testutils.FixtureProductPrice(variant.ID, "wholesale", 80.00)
	price3 := testutils.FixtureProductPrice(variant.ID, "bulk", 70.00)
	db.Create(price1)
	db.Create(price2)
	db.Create(price3)

	responses, err := service.GetVariantPrices(variant.ID)

	testutils.AssertNoError(t, err, "GetVariantPrices should succeed")
	testutils.AssertEqual(t, len(responses), 3, "Should return all 3 prices")
}

func TestPriceService_GetVariantPrices_SinglePrice(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	price := testutils.FixtureProductPrice(variant.ID, "retail", 100.50)
	db.Create(price)

	responses, err := service.GetVariantPrices(variant.ID)

	testutils.AssertNoError(t, err, "GetVariantPrices should succeed")
	testutils.AssertEqual(t, len(responses), 1, "Should return 1 price")
}

func TestPriceService_GetVariantPrices_Empty(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	responses, err := service.GetVariantPrices(variant.ID)

	testutils.AssertNoError(t, err, "GetVariantPrices should succeed even with no prices")
	testutils.AssertEqual(t, len(responses), 0, "Should return empty array")
}

// ========================================
// GetCurrentPrice Tests
// ========================================

func TestPriceService_GetCurrentPrice_ActiveWithinDateRange(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	// Create an active price within date range
	effectiveFrom := time.Now().UTC().Add(-7 * 24 * time.Hour) // 7 days ago
	effectiveTo := time.Now().UTC().Add(7 * 24 * time.Hour)    // 7 days from now
	price := testutils.FixtureProductPriceWithDates(variant.ID, "retail", 100.50, effectiveFrom, &effectiveTo)
	db.Create(price)

	response, err := service.GetCurrentPrice(variant.ID, "retail")

	testutils.AssertNoError(t, err, "GetCurrentPrice should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.ID, price.ID, "Price ID mismatch")
}

func TestPriceService_GetCurrentPrice_IndefiniteEndDate(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	// Create a price without end date (indefinite)
	price := testutils.FixtureProductPrice(variant.ID, "retail", 100.50)
	db.Create(price)

	response, err := service.GetCurrentPrice(variant.ID, "retail")

	testutils.AssertNoError(t, err, "GetCurrentPrice should succeed for indefinite prices")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.ID, price.ID, "Price ID mismatch")
}

func TestPriceService_GetCurrentPrice_ExpiredPrice(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	// Create an expired price
	price := testutils.FixtureProductPriceExpired(variant.ID, "retail", 100.50)
	db.Create(price)

	_, err := service.GetCurrentPrice(variant.ID, "retail")

	testutils.AssertError(t, err, "Should fail for expired prices")
}

func TestPriceService_GetCurrentPrice_InactivePrice(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	// Create an inactive price
	price := testutils.FixtureProductPriceInactive(variant.ID, "retail", 100.50)
	db.Create(price)

	_, err := service.GetCurrentPrice(variant.ID, "retail")

	testutils.AssertError(t, err, "Should fail for inactive prices")
}

func TestPriceService_GetCurrentPrice_FutureEffectiveDate(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	// Create a price that starts in the future
	effectiveFrom := time.Now().UTC().Add(7 * 24 * time.Hour) // 7 days from now
	price := testutils.FixtureProductPriceWithDates(variant.ID, "retail", 100.50, effectiveFrom, nil)
	db.Create(price)

	_, err := service.GetCurrentPrice(variant.ID, "retail")

	testutils.AssertError(t, err, "Should fail for future prices")
}

func TestPriceService_GetCurrentPrice_MostRecentWhenMultiple(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	// Create older price
	olderEffectiveFrom := time.Now().UTC().Add(-30 * 24 * time.Hour)
	olderPrice := testutils.FixtureProductPriceWithDates(variant.ID, "retail", 90.00, olderEffectiveFrom, nil)
	db.Create(olderPrice)

	// Create newer price
	newerEffectiveFrom := time.Now().UTC().Add(-7 * 24 * time.Hour)
	newerPrice := testutils.FixtureProductPriceWithDates(variant.ID, "retail", 100.50, newerEffectiveFrom, nil)
	db.Create(newerPrice)

	response, err := service.GetCurrentPrice(variant.ID, "retail")

	testutils.AssertNoError(t, err, "GetCurrentPrice should succeed")
	testutils.AssertEqual(t, response.ID, newerPrice.ID, "Should return the most recent price")
	testutils.AssertEqual(t, response.Price, 100.50, "Should return the newer price amount")
}

// ========================================
// UpdateProductPrice Tests
// ========================================

func TestPriceService_UpdateProductPrice_UpdatePrice(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	price := testutils.FixtureProductPrice(variant.ID, "retail", 100.50)
	db.Create(price)

	newPrice := 120.00
	request := &models.UpdateProductPriceRequest{
		Price: &newPrice,
	}

	response, err := service.UpdateProductPrice(price.ID, request)

	testutils.AssertNoError(t, err, "UpdateProductPrice should succeed")
	testutils.AssertEqual(t, response.Price, 120.00, "Price should be updated")
}

func TestPriceService_UpdateProductPrice_UpdatePriceType(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	price := testutils.FixtureProductPrice(variant.ID, "retail", 100.50)
	db.Create(price)

	newPriceType := "wholesale"
	request := &models.UpdateProductPriceRequest{
		PriceType: &newPriceType,
	}

	response, err := service.UpdateProductPrice(price.ID, request)

	testutils.AssertNoError(t, err, "UpdateProductPrice should succeed")
	testutils.AssertEqual(t, response.PriceType, "wholesale", "PriceType should be updated")
}

func TestPriceService_UpdateProductPrice_UpdateCurrency(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	price := testutils.FixtureProductPrice(variant.ID, "retail", 100.50)
	db.Create(price)

	newCurrency := "USD"
	request := &models.UpdateProductPriceRequest{
		Currency: &newCurrency,
	}

	response, err := service.UpdateProductPrice(price.ID, request)

	testutils.AssertNoError(t, err, "UpdateProductPrice should succeed")
	testutils.AssertEqual(t, response.Currency, "USD", "Currency should be updated")
}

func TestPriceService_UpdateProductPrice_UpdateDates(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	price := testutils.FixtureProductPrice(variant.ID, "retail", 100.50)
	db.Create(price)

	newEffectiveFrom := time.Now().UTC().Add(-7 * 24 * time.Hour).Format(time.RFC3339)
	newEffectiveTo := time.Now().UTC().Add(30 * 24 * time.Hour).Format(time.RFC3339)

	request := &models.UpdateProductPriceRequest{
		EffectiveFrom: &newEffectiveFrom,
		EffectiveTo:   &newEffectiveTo,
	}

	response, err := service.UpdateProductPrice(price.ID, request)

	testutils.AssertNoError(t, err, "UpdateProductPrice should succeed")
	testutils.AssertNotNil(t, response.EffectiveTo, "EffectiveTo should be updated")
}

func TestPriceService_UpdateProductPrice_UpdateIsActive(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	price := testutils.FixtureProductPrice(variant.ID, "retail", 100.50)
	db.Create(price)

	isActive := false
	request := &models.UpdateProductPriceRequest{
		IsActive: &isActive,
	}

	response, err := service.UpdateProductPrice(price.ID, request)

	testutils.AssertNoError(t, err, "UpdateProductPrice should succeed")
	testutils.AssertEqual(t, response.IsActive, false, "IsActive should be updated")
}

func TestPriceService_UpdateProductPrice_PartialUpdate(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	price := testutils.FixtureProductPrice(variant.ID, "retail", 100.50)
	price.Currency = "INR"
	db.Create(price)

	// Only update price, leave other fields unchanged
	newPrice := 120.00
	request := &models.UpdateProductPriceRequest{
		Price: &newPrice,
	}

	response, err := service.UpdateProductPrice(price.ID, request)

	testutils.AssertNoError(t, err, "UpdateProductPrice should succeed")
	testutils.AssertEqual(t, response.Price, 120.00, "Price should be updated")
	testutils.AssertEqual(t, response.PriceType, "retail", "PriceType should remain unchanged")
	testutils.AssertEqual(t, response.Currency, "INR", "Currency should remain unchanged")
}

func TestPriceService_UpdateProductPrice_NotFound(t *testing.T) {
	service, _, cleanup := setupPriceService(t)
	defer cleanup()

	newPrice := 120.00
	request := &models.UpdateProductPriceRequest{
		Price: &newPrice,
	}

	_, err := service.UpdateProductPrice("non-existent-id", request)

	testutils.AssertError(t, err, "Should fail when price not found")
}

// ========================================
// DeleteProductPrice Tests
// ========================================

func TestPriceService_DeleteProductPrice_Success(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	price := testutils.FixtureProductPrice(variant.ID, "retail", 100.50)
	db.Create(price)

	err := service.DeleteProductPrice(price.ID)

	testutils.AssertNoError(t, err, "DeleteProductPrice should succeed")
}

func TestPriceService_DeleteProductPrice_NotFound(t *testing.T) {
	service, _, cleanup := setupPriceService(t)
	defer cleanup()

	err := service.DeleteProductPrice("non-existent-id")

	testutils.AssertError(t, err, "Should fail when price not found")
}

func TestPriceService_DeleteProductPrice_VerifyDeletion(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	price := testutils.FixtureProductPrice(variant.ID, "retail", 100.50)
	db.Create(price)

	// Delete the price
	err := service.DeleteProductPrice(price.ID)
	testutils.AssertNoError(t, err, "DeleteProductPrice should succeed")

	// Verify it's deleted
	_, err = service.GetProductPrice(price.ID)
	testutils.AssertError(t, err, "Price should no longer exist")
}

// ========================================
// GetExpiredPrices Tests
// ========================================

func TestPriceService_GetExpiredPrices_FindExpired(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	// Create expired price
	expiredPrice := testutils.FixtureProductPriceExpired(variant.ID, "retail", 100.50)
	db.Create(expiredPrice)

	// Create current price (should not be included)
	currentPrice := testutils.FixtureProductPrice(variant.ID, "wholesale", 80.00)
	db.Create(currentPrice)

	responses, err := service.GetExpiredPrices()

	testutils.AssertNoError(t, err, "GetExpiredPrices should succeed")
	testutils.AssertEqual(t, len(responses), 1, "Should find 1 expired price")
	testutils.AssertEqual(t, responses[0].ID, expiredPrice.ID, "Should return the expired price")
}

func TestPriceService_GetExpiredPrices_ExcludeIndefinite(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	// Create price without EffectiveTo (indefinite)
	indefinitePrice := testutils.FixtureProductPrice(variant.ID, "retail", 100.50)
	db.Create(indefinitePrice)

	responses, err := service.GetExpiredPrices()

	testutils.AssertNoError(t, err, "GetExpiredPrices should succeed")
	testutils.AssertEqual(t, len(responses), 0, "Should not include indefinite prices")
}

func TestPriceService_GetExpiredPrices_ExcludeFuture(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	// Create price that expires in the future
	effectiveFrom := time.Now().UTC()
	effectiveTo := time.Now().UTC().Add(30 * 24 * time.Hour) // 30 days from now
	futurePrice := testutils.FixtureProductPriceWithDates(variant.ID, "retail", 100.50, effectiveFrom, &effectiveTo)
	db.Create(futurePrice)

	responses, err := service.GetExpiredPrices()

	testutils.AssertNoError(t, err, "GetExpiredPrices should succeed")
	testutils.AssertEqual(t, len(responses), 0, "Should not include future expiring prices")
}

func TestPriceService_GetExpiredPrices_Empty(t *testing.T) {
	service, db, cleanup := setupPriceService(t)
	defer cleanup()

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariantSimple(t, db, "VAR-001", product.ID, "TOM-1KG")

	// Create only current prices
	currentPrice := testutils.FixtureProductPrice(variant.ID, "retail", 100.50)
	db.Create(currentPrice)

	responses, err := service.GetExpiredPrices()

	testutils.AssertNoError(t, err, "GetExpiredPrices should succeed")
	testutils.AssertEqual(t, len(responses), 0, "Should return empty array when no expired prices")
}
