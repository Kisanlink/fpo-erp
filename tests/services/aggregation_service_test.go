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

// =============================================================================
// Test Setup & Fixtures
// =============================================================================

// setupAggregationService creates an AggregationService with in-memory database
func setupAggregationService(t *testing.T) (*services.AggregationService, *gorm.DB, func()) {
	t.Helper()

	db := testutils.SetupTestDB(t)

	// Create all required repositories
	productRepo := repositories.NewProductRepository(db)
	variantRepo := repositories.NewProductVariantRepository(db)
	inventoryRepo := repositories.NewInventoryRepository(db)
	warehouseRepo := repositories.NewWarehouseRepository(db)
	collaboratorRepo := repositories.NewCollaboratorRepository(db)
	discountsRepo := repositories.NewDiscountsRepository(db)
	taxRepo := repositories.NewTaxRepository(db)
	refundPoliciesRepo := repositories.NewRefundPoliciesRepository(db)
	purchaseOrderRepo := repositories.NewPurchaseOrderRepository(db)
	grnRepo := repositories.NewGRNRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	priceRepo := repositories.NewProductPriceRepository(db)
	service := services.NewAggregationService(
		productRepo,
		variantRepo,
		priceRepo,
		inventoryRepo,
		warehouseRepo,
		collaboratorRepo,
		discountsRepo,
		taxRepo,
		refundPoliciesRepo,
		purchaseOrderRepo,
		grnRepo,
		mockLogger,
	)

	cleanup := func() {
		testutils.CleanupTestDB(db)
	}

	return service, db, cleanup
}

// =============================================================================
// ParseIncludeOptions Tests
// =============================================================================

func TestParseIncludeOptions_All(t *testing.T) {
	// Test "all" keyword
	options := services.ParseIncludeOptions("all")
	testutils.AssertTrue(t, options.Variants, "Variants should be true for 'all'")
	testutils.AssertTrue(t, options.Prices, "Prices should be true for 'all'")
	testutils.AssertTrue(t, options.Inventory, "Inventory should be true for 'all'")
	testutils.AssertTrue(t, options.Collaborators, "Collaborators should be true for 'all'")
	testutils.AssertTrue(t, options.Taxes, "Taxes should be true for 'all'")
}

func TestParseIncludeOptions_Empty(t *testing.T) {
	// Test empty string (should default to all)
	options := services.ParseIncludeOptions("")
	testutils.AssertTrue(t, options.Variants, "Variants should be true for empty string")
	testutils.AssertTrue(t, options.Prices, "Prices should be true for empty string")
}

func TestParseIncludeOptions_Wildcard(t *testing.T) {
	// Test "*" wildcard
	options := services.ParseIncludeOptions("*")
	testutils.AssertTrue(t, options.Variants, "Variants should be true for '*'")
	testutils.AssertTrue(t, options.Prices, "Prices should be true for '*'")
}

func TestParseIncludeOptions_None(t *testing.T) {
	// Test "none" keyword
	options := services.ParseIncludeOptions("none")
	testutils.AssertFalse(t, options.Variants, "Variants should be false for 'none'")
	testutils.AssertFalse(t, options.Prices, "Prices should be false for 'none'")
	testutils.AssertFalse(t, options.Inventory, "Inventory should be false for 'none'")
}

func TestParseIncludeOptions_Specific(t *testing.T) {
	// Test specific includes
	options := services.ParseIncludeOptions("variants,prices")
	testutils.AssertTrue(t, options.Variants, "Variants should be true")
	testutils.AssertTrue(t, options.Prices, "Prices should be true")
	testutils.AssertFalse(t, options.Inventory, "Inventory should be false")
	testutils.AssertFalse(t, options.Collaborators, "Collaborators should be false")
}

func TestParseIncludeOptions_CaseInsensitive(t *testing.T) {
	// Test case insensitivity
	options := services.ParseIncludeOptions("VARIANTS,Prices,INVENTORY")
	testutils.AssertTrue(t, options.Variants, "Variants should be true (case insensitive)")
	testutils.AssertTrue(t, options.Prices, "Prices should be true (case insensitive)")
	testutils.AssertTrue(t, options.Inventory, "Inventory should be true (case insensitive)")
}

// =============================================================================
// ParsePOIncludeOptions Tests
// =============================================================================

func TestParsePOIncludeOptions_All(t *testing.T) {
	options := services.ParsePOIncludeOptions("all")
	testutils.AssertTrue(t, options["collaborator"], "collaborator should be true for 'all'")
	testutils.AssertTrue(t, options["warehouse"], "warehouse should be true for 'all'")
	testutils.AssertTrue(t, options["items"], "items should be true for 'all'")
	testutils.AssertTrue(t, options["grns"], "grns should be true for 'all'")
	testutils.AssertTrue(t, options["inventory"], "inventory should be true for 'all'")
	testutils.AssertTrue(t, options["payments"], "payments should be true for 'all'")
	testutils.AssertTrue(t, options["timeline"], "timeline should be true for 'all'")
}

func TestParsePOIncludeOptions_None(t *testing.T) {
	options := services.ParsePOIncludeOptions("none")
	testutils.AssertFalse(t, options["collaborator"], "collaborator should be false for 'none'")
	testutils.AssertFalse(t, options["items"], "items should be false for 'none'")
}

func TestParsePOIncludeOptions_Specific(t *testing.T) {
	options := services.ParsePOIncludeOptions("collaborator,items,grns")
	testutils.AssertTrue(t, options["collaborator"], "collaborator should be true")
	testutils.AssertTrue(t, options["items"], "items should be true")
	testutils.AssertTrue(t, options["grns"], "grns should be true")
	testutils.AssertFalse(t, options["warehouse"], "warehouse should be false")
	testutils.AssertFalse(t, options["timeline"], "timeline should be false")
}

// =============================================================================
// ParseInventoryIncludeOptions Tests
// =============================================================================

func TestParseInventoryIncludeOptions_All(t *testing.T) {
	options := services.ParseInventoryIncludeOptions("all")
	testutils.AssertTrue(t, options["variant"], "variant should be true for 'all'")
	testutils.AssertTrue(t, options["product"], "product should be true for 'all'")
	testutils.AssertTrue(t, options["warehouse"], "warehouse should be true for 'all'")
	testutils.AssertTrue(t, options["prices"], "prices should be true for 'all'")
	testutils.AssertTrue(t, options["taxes"], "taxes should be true for 'all'")
}

func TestParseInventoryIncludeOptions_Specific(t *testing.T) {
	options := services.ParseInventoryIncludeOptions("variant,product")
	testutils.AssertTrue(t, options["variant"], "variant should be true")
	testutils.AssertTrue(t, options["product"], "product should be true")
	testutils.AssertFalse(t, options["warehouse"], "warehouse should be false")
	testutils.AssertFalse(t, options["prices"], "prices should be false")
}

// =============================================================================
// GetProductDetail Tests
// =============================================================================

func TestAggregationService_GetProductDetail_Success(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create test product
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")

	// Create variant for the product
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	// Execute
	response, err := service.GetProductDetail(product.ID, &models.ProductDetailRequest{
		Include: "all",
	})

	// Assert
	testutils.AssertNoError(t, err, "GetProductDetail should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.Product.ID, product.ID, "Product ID mismatch")
	testutils.AssertEqual(t, response.Product.Name, "Test Product", "Product name mismatch")
	testutils.AssertEqual(t, len(response.Variants), 1, "Should have 1 variant")
	testutils.AssertEqual(t, response.Variants[0].ID, variant.ID, "Variant ID mismatch")
}

func TestAggregationService_GetProductDetail_NotFound(t *testing.T) {
	service, _, cleanup := setupAggregationService(t)
	defer cleanup()

	// Execute
	_, err := service.GetProductDetail("INVALID-PRODUCT-ID", &models.ProductDetailRequest{
		Include: "all",
	})

	// Assert
	testutils.AssertError(t, err, "Should return error for non-existent product")
}

func TestAggregationService_GetProductDetail_WithVariantsOnly(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create test product with multiple variants
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "500g")
	testutils.CreateTestVariant(t, db, "VAR-002", product.ID, "VAR-SKU-002", "1kg")

	// Execute with variants only
	response, err := service.GetProductDetail(product.ID, &models.ProductDetailRequest{
		Include: "variants",
	})

	// Assert
	testutils.AssertNoError(t, err, "GetProductDetail should succeed")
	testutils.AssertEqual(t, len(response.Variants), 2, "Should have 2 variants")
}

func TestAggregationService_GetProductDetail_WithInventory(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create test product and variant
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	// Create warehouse and inventory
	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")
	expiryDate := time.Now().Add(30 * 24 * time.Hour)
	batch := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 100, expiryDate)
	db.Create(batch)

	// Execute with inventory
	response, err := service.GetProductDetail(product.ID, &models.ProductDetailRequest{
		Include: "variants,inventory",
	})

	// Assert
	testutils.AssertNoError(t, err, "GetProductDetail should succeed")
	testutils.AssertEqual(t, len(response.Variants), 1, "Should have 1 variant")
	testutils.AssertNotNil(t, response.Variants[0].StockSummary, "StockSummary should not be nil")
	testutils.AssertEqual(t, response.Variants[0].StockSummary.TotalQuantity, int64(100), "Total quantity mismatch")
	testutils.AssertTrue(t, response.Variants[0].StockSummary.InStock, "Should be in stock")
}

func TestAggregationService_GetProductDetail_ActiveOnly(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create test product
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")

	// Create active and inactive variants
	activeVariant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")
	inactiveVariant := testutils.CreateTestVariant(t, db, "VAR-002", product.ID, "VAR-SKU-002", "2kg")
	inactiveVariant.IsActive = false
	db.Save(inactiveVariant)

	// Execute with active_only filter
	response, err := service.GetProductDetail(product.ID, &models.ProductDetailRequest{
		Include:    "variants",
		ActiveOnly: true,
	})

	// Assert
	testutils.AssertNoError(t, err, "GetProductDetail should succeed")
	testutils.AssertEqual(t, len(response.Variants), 1, "Should have 1 active variant")
	testutils.AssertEqual(t, response.Variants[0].ID, activeVariant.ID, "Should be the active variant")
}

func TestAggregationService_GetProductDetail_ByWarehouse(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create test product and variant
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	// Create two warehouses with inventory
	warehouse1 := testutils.CreateTestWarehouse(t, db, "WH-001")
	warehouse2 := testutils.CreateTestWarehouse(t, db, "WH-002")

	expiryDate := time.Now().Add(30 * 24 * time.Hour)
	batch1 := testutils.FixtureInventoryBatchWithExpiry(warehouse1.ID, variant.ID, 100, expiryDate)
	batch2 := testutils.FixtureInventoryBatchWithExpiry(warehouse2.ID, variant.ID, 50, expiryDate)
	db.Create(batch1)
	db.Create(batch2)

	// Execute with warehouse filter
	response, err := service.GetProductDetail(product.ID, &models.ProductDetailRequest{
		Include:     "variants,inventory",
		WarehouseID: warehouse1.ID,
	})

	// Assert
	testutils.AssertNoError(t, err, "GetProductDetail should succeed")
	testutils.AssertEqual(t, len(response.Variants), 1, "Should have 1 variant")
	testutils.AssertNotNil(t, response.Variants[0].StockSummary, "StockSummary should not be nil")
	testutils.AssertEqual(t, response.Variants[0].StockSummary.TotalQuantity, int64(100), "Should only count warehouse1 stock")
}

func TestAggregationService_GetProductDetail_Metadata(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create test product with variants
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")
	testutils.CreateTestVariant(t, db, "VAR-002", product.ID, "VAR-SKU-002", "2kg")

	// Execute
	response, err := service.GetProductDetail(product.ID, &models.ProductDetailRequest{
		Include: "all",
	})

	// Assert
	testutils.AssertNoError(t, err, "GetProductDetail should succeed")
	testutils.AssertEqual(t, response.Metadata.TotalVariants, 2, "Total variants should be 2")
	testutils.AssertTrue(t, response.Metadata.ReadTimestamp != "", "ReadTimestamp should be set")
	testutils.AssertTrue(t, response.Metadata.ConsistencyToken != "", "ConsistencyToken should be set")
}

// =============================================================================
// GetVariantDetail Tests
// =============================================================================

func TestAggregationService_GetVariantDetail_Success(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create test product and variant
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	// Execute
	response, err := service.GetVariantDetail(variant.ID, "all", "")

	// Assert
	testutils.AssertNoError(t, err, "GetVariantDetail should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.Variant.ID, variant.ID, "Variant ID mismatch")
	testutils.AssertEqual(t, response.Variant.ProductID, product.ID, "ProductID mismatch")
}

func TestAggregationService_GetVariantDetail_NotFound(t *testing.T) {
	service, _, cleanup := setupAggregationService(t)
	defer cleanup()

	// Execute
	_, err := service.GetVariantDetail("INVALID-VARIANT-ID", "all", "")

	// Assert
	testutils.AssertError(t, err, "Should return error for non-existent variant")
}

func TestAggregationService_GetVariantDetail_WithProduct(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create test product and variant
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	// Execute with variants include (which triggers product fetch)
	response, err := service.GetVariantDetail(variant.ID, "variants", "")

	// Assert
	testutils.AssertNoError(t, err, "GetVariantDetail should succeed")
	testutils.AssertNotNil(t, response.Variant.Product, "Product should be included")
	testutils.AssertEqual(t, response.Variant.Product.ID, product.ID, "Product ID mismatch")
	testutils.AssertEqual(t, response.Variant.Product.Name, "Test Product", "Product name mismatch")
}

func TestAggregationService_GetVariantDetail_WithInventory(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create test product and variant
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	// Create warehouse and inventory
	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")
	expiryDate := time.Now().Add(30 * 24 * time.Hour)
	batch := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 100, expiryDate)
	db.Create(batch)

	// Execute with inventory
	response, err := service.GetVariantDetail(variant.ID, "inventory", "")

	// Assert
	testutils.AssertNoError(t, err, "GetVariantDetail should succeed")
	testutils.AssertNotNil(t, response.Variant.StockSummary, "StockSummary should not be nil")
	testutils.AssertEqual(t, response.Variant.StockSummary.TotalQuantity, int64(100), "Total quantity mismatch")
}

// =============================================================================
// GetSalesContext Tests
// =============================================================================

func TestAggregationService_GetSalesContext_Success(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create warehouse
	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")

	// Create product, variant, and inventory
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	expiryDate := time.Now().Add(30 * 24 * time.Hour)
	batch := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 100, expiryDate)
	db.Create(batch)

	// Execute
	response, err := service.GetSalesContext(&models.SalesContextRequest{
		WarehouseID: warehouse.ID,
		PriceType:   "retail",
	})

	// Assert
	testutils.AssertNoError(t, err, "GetSalesContext should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.Warehouse.ID, warehouse.ID, "Warehouse ID mismatch")
	testutils.AssertTrue(t, len(response.AvailableInventory) > 0, "Should have available inventory")
	testutils.AssertEqual(t, len(response.PaymentMethods), 3, "Should have 3 payment methods")
}

func TestAggregationService_GetSalesContext_InvalidWarehouse(t *testing.T) {
	service, _, cleanup := setupAggregationService(t)
	defer cleanup()

	// Execute
	_, err := service.GetSalesContext(&models.SalesContextRequest{
		WarehouseID: "INVALID-WAREHOUSE-ID",
	})

	// Assert
	testutils.AssertError(t, err, "Should return error for invalid warehouse")
}

func TestAggregationService_GetSalesContext_IncludeZeroStock(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create warehouse
	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")

	// Create product, variant with zero stock
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	expiryDate := time.Now().Add(30 * 24 * time.Hour)
	batch := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 0, expiryDate)
	db.Create(batch)

	// Execute without zero stock
	responseWithoutZero, err := service.GetSalesContext(&models.SalesContextRequest{
		WarehouseID:      warehouse.ID,
		IncludeZeroStock: false,
	})
	testutils.AssertNoError(t, err, "GetSalesContext should succeed")
	testutils.AssertEqual(t, len(responseWithoutZero.AvailableInventory), 0, "Should not include zero stock")

	// Execute with zero stock
	responseWithZero, err := service.GetSalesContext(&models.SalesContextRequest{
		WarehouseID:      warehouse.ID,
		IncludeZeroStock: true,
	})
	testutils.AssertNoError(t, err, "GetSalesContext should succeed")
	testutils.AssertEqual(t, len(responseWithZero.AvailableInventory), 1, "Should include zero stock")
}

func TestAggregationService_GetSalesContext_Metadata(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create warehouse with inventory
	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	expiryDate := time.Now().Add(30 * 24 * time.Hour)
	batch := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 100, expiryDate)
	db.Create(batch)

	// Execute
	response, err := service.GetSalesContext(&models.SalesContextRequest{
		WarehouseID: warehouse.ID,
	})

	// Assert
	testutils.AssertNoError(t, err, "GetSalesContext should succeed")
	testutils.AssertTrue(t, response.Metadata.TotalProducts > 0, "Should have products")
	testutils.AssertTrue(t, response.Metadata.TotalVariants > 0, "Should have variants")
	testutils.AssertTrue(t, response.Metadata.TotalBatches > 0, "Should have batches")
	testutils.AssertTrue(t, response.Metadata.ReadTimestamp != "", "ReadTimestamp should be set")
}

// =============================================================================
// GetPODetail Tests
// =============================================================================

func TestAggregationService_GetPODetail_Success(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create test data
	collaborator := testutils.FixtureCollaboratorWithID("COLLAB-001", "Test Vendor")
	db.Create(collaborator)

	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	// Create purchase order with items
	po := testutils.FixturePurchaseOrderWithID("PO-001", "PO-2025-0001", collaborator.ID, warehouse.ID, 1000.00)
	db.Create(po)

	poItem := testutils.FixturePurchaseOrderItemWithID("POITEM-001", po.ID, variant.ID, 100, 10.00)
	db.Create(poItem)

	// Execute
	response, err := service.GetPODetail(po.ID, &models.PODetailRequest{
		Include: "all",
	})

	// Assert
	testutils.AssertNoError(t, err, "GetPODetail should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.PurchaseOrder.ID, po.ID, "PO ID mismatch")
	testutils.AssertEqual(t, response.PurchaseOrder.PONumber, "PO-2025-0001", "PO number mismatch")
	testutils.AssertNotNil(t, response.Collaborator, "Collaborator should be included")
	testutils.AssertEqual(t, response.Collaborator.ID, collaborator.ID, "Collaborator ID mismatch")
	testutils.AssertNotNil(t, response.Warehouse, "Warehouse should be included")
	testutils.AssertEqual(t, len(response.Items), 1, "Should have 1 item")
}

func TestAggregationService_GetPODetail_NotFound(t *testing.T) {
	service, _, cleanup := setupAggregationService(t)
	defer cleanup()

	// Execute
	_, err := service.GetPODetail("INVALID-PO-ID", &models.PODetailRequest{
		Include: "all",
	})

	// Assert
	testutils.AssertError(t, err, "Should return error for non-existent PO")
}

func TestAggregationService_GetPODetail_IncludeNone(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create minimal PO
	collaborator := testutils.FixtureCollaboratorWithID("COLLAB-001", "Test Vendor")
	db.Create(collaborator)

	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")
	po := testutils.FixturePurchaseOrderWithID("PO-001", "PO-2025-0001", collaborator.ID, warehouse.ID, 1000.00)
	db.Create(po)

	// Execute with include=none
	response, err := service.GetPODetail(po.ID, &models.PODetailRequest{
		Include: "none",
	})

	// Assert
	testutils.AssertNoError(t, err, "GetPODetail should succeed")
	testutils.AssertNil(t, response.Collaborator, "Collaborator should be nil")
	testutils.AssertNil(t, response.Warehouse, "Warehouse should be nil")
	testutils.AssertEqual(t, len(response.Items), 0, "Items should be empty")
	testutils.AssertEqual(t, len(response.Timeline), 0, "Timeline should be empty")
}

func TestAggregationService_GetPODetail_Summary(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create PO with items
	collaborator := testutils.FixtureCollaboratorWithID("COLLAB-001", "Test Vendor")
	db.Create(collaborator)

	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	po := testutils.FixturePurchaseOrderWithID("PO-001", "PO-2025-0001", collaborator.ID, warehouse.ID, 1000.00)
	db.Create(po)

	poItem := testutils.FixturePurchaseOrderItemWithID("POITEM-001", po.ID, variant.ID, 100, 10.00)
	receivedQty := int64(50)
	poItem.ReceivedQuantity = &receivedQty
	db.Create(poItem)

	// Execute
	response, err := service.GetPODetail(po.ID, &models.PODetailRequest{
		Include: "items",
	})

	// Assert
	testutils.AssertNoError(t, err, "GetPODetail should succeed")
	testutils.AssertEqual(t, response.Summary.TotalItemsOrdered, int64(100), "Total ordered mismatch")
	testutils.AssertEqual(t, response.Summary.TotalItemsReceived, int64(50), "Total received mismatch")
	testutils.AssertEqual(t, response.Summary.TotalItemsPending, int64(50), "Total pending mismatch")
	testutils.AssertEqual(t, response.Summary.FulfillmentStatus, "partially_received", "Fulfillment status mismatch")
	testutils.AssertEqual(t, response.Summary.CompletionPercentage, 50.0, "Completion percentage mismatch")
}

func TestAggregationService_GetPODetail_Timeline(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create PO
	collaborator := testutils.FixtureCollaboratorWithID("COLLAB-001", "Test Vendor")
	db.Create(collaborator)

	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")
	po := testutils.FixturePurchaseOrderWithID("PO-001", "PO-2025-0001", collaborator.ID, warehouse.ID, 1000.00)
	po.Status = "delivered"
	db.Create(po)

	// Execute with timeline
	response, err := service.GetPODetail(po.ID, &models.PODetailRequest{
		Include: "timeline",
	})

	// Assert
	testutils.AssertNoError(t, err, "GetPODetail should succeed")
	testutils.AssertTrue(t, len(response.Timeline) > 0, "Timeline should have events")

	// Check for PO created event
	hasCreatedEvent := false
	for _, event := range response.Timeline {
		if event.Event == "purchase_order_created" {
			hasCreatedEvent = true
			break
		}
	}
	testutils.AssertTrue(t, hasCreatedEvent, "Timeline should have PO created event")
}

// =============================================================================
// GetInventoryList Tests
// =============================================================================

func TestAggregationService_GetInventoryList_Success(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create test data
	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	expiryDate := time.Now().Add(30 * 24 * time.Hour)
	batch := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 100, expiryDate)
	db.Create(batch)

	// Execute
	response, err := service.GetInventoryList(&models.InventoryListRequest{
		Include:     "all",
		InStockOnly: true,
		Limit:       50,
		Offset:      0,
	})

	// Assert
	testutils.AssertNoError(t, err, "GetInventoryList should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, len(response.Batches), 1, "Should have 1 batch")
	testutils.AssertEqual(t, response.Batches[0].ID, batch.ID, "Batch ID mismatch")
	testutils.AssertEqual(t, response.Summary.TotalBatches, 1, "Total batches should be 1")
}

func TestAggregationService_GetInventoryList_FilterByWarehouse(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create two warehouses with inventory
	warehouse1 := testutils.CreateTestWarehouse(t, db, "WH-001")
	warehouse2 := testutils.CreateTestWarehouse(t, db, "WH-002")

	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	expiryDate := time.Now().Add(30 * 24 * time.Hour)
	batch1 := testutils.FixtureInventoryBatchWithExpiry(warehouse1.ID, variant.ID, 100, expiryDate)
	batch2 := testutils.FixtureInventoryBatchWithExpiry(warehouse2.ID, variant.ID, 50, expiryDate)
	db.Create(batch1)
	db.Create(batch2)

	// Execute with warehouse filter
	response, err := service.GetInventoryList(&models.InventoryListRequest{
		WarehouseID: warehouse1.ID,
		Include:     "all",
	})

	// Assert
	testutils.AssertNoError(t, err, "GetInventoryList should succeed")
	testutils.AssertEqual(t, len(response.Batches), 1, "Should have 1 batch from warehouse1")
	testutils.AssertEqual(t, response.Batches[0].ID, batch1.ID, "Should be batch from warehouse1")
}

func TestAggregationService_GetInventoryList_FilterByVariant(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create warehouse with multiple variants
	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant1 := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")
	variant2 := testutils.CreateTestVariant(t, db, "VAR-002", product.ID, "VAR-SKU-002", "2kg")

	expiryDate := time.Now().Add(30 * 24 * time.Hour)
	batch1 := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant1.ID, 100, expiryDate)
	batch2 := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant2.ID, 50, expiryDate)
	db.Create(batch1)
	db.Create(batch2)

	// Execute with variant filter
	response, err := service.GetInventoryList(&models.InventoryListRequest{
		VariantID: variant1.ID,
		Include:   "all",
	})

	// Assert
	testutils.AssertNoError(t, err, "GetInventoryList should succeed")
	testutils.AssertEqual(t, len(response.Batches), 1, "Should have 1 batch for variant1")
}

func TestAggregationService_GetInventoryList_InStockOnly(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create warehouse with some zero stock
	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	expiryDate := time.Now().Add(30 * 24 * time.Hour)
	batchWithStock := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 100, expiryDate)
	batchZeroStock := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 0, expiryDate)
	db.Create(batchWithStock)
	db.Create(batchZeroStock)

	// Execute with in_stock_only
	response, err := service.GetInventoryList(&models.InventoryListRequest{
		InStockOnly: true,
		Include:     "all",
	})

	// Assert
	testutils.AssertNoError(t, err, "GetInventoryList should succeed")
	testutils.AssertEqual(t, len(response.Batches), 1, "Should only have 1 batch with stock")
}

func TestAggregationService_GetInventoryList_ExpiringSoon(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create warehouse with expiring and non-expiring stock
	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	expiringIn15Days := time.Now().Add(15 * 24 * time.Hour)
	expiringIn90Days := time.Now().Add(90 * 24 * time.Hour)

	batchExpiring := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 100, expiringIn15Days)
	batchNotExpiring := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 50, expiringIn90Days)
	db.Create(batchExpiring)
	db.Create(batchNotExpiring)

	// Execute with expiring_soon filter
	response, err := service.GetInventoryList(&models.InventoryListRequest{
		ExpiringSoon: true,
		Include:      "all",
	})

	// Assert
	testutils.AssertNoError(t, err, "GetInventoryList should succeed")
	testutils.AssertEqual(t, len(response.Batches), 1, "Should only have 1 expiring batch")
	testutils.AssertEqual(t, response.Batches[0].ID, batchExpiring.ID, "Should be the expiring batch")
}

func TestAggregationService_GetInventoryList_Pagination(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create multiple batches
	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	for i := 0; i < 5; i++ {
		expiryDate := time.Now().Add(time.Duration(30+i) * 24 * time.Hour)
		batch := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, int64(100+i*10), expiryDate)
		db.Create(batch)
	}

	// Execute first page
	page1, err := service.GetInventoryList(&models.InventoryListRequest{
		Limit:  2,
		Offset: 0,
	})
	testutils.AssertNoError(t, err, "GetInventoryList should succeed")
	testutils.AssertEqual(t, len(page1.Batches), 2, "Should have 2 batches")
	testutils.AssertEqual(t, page1.Pagination.Total, 5, "Total should be 5")
	testutils.AssertTrue(t, page1.Pagination.HasMore, "Should have more pages")

	// Execute second page
	page2, err := service.GetInventoryList(&models.InventoryListRequest{
		Limit:  2,
		Offset: 2,
	})
	testutils.AssertNoError(t, err, "GetInventoryList should succeed")
	testutils.AssertEqual(t, len(page2.Batches), 2, "Should have 2 batches")
	testutils.AssertTrue(t, page2.Pagination.HasMore, "Should have more pages")

	// Execute third page
	page3, err := service.GetInventoryList(&models.InventoryListRequest{
		Limit:  2,
		Offset: 4,
	})
	testutils.AssertNoError(t, err, "GetInventoryList should succeed")
	testutils.AssertEqual(t, len(page3.Batches), 1, "Should have 1 batch")
	testutils.AssertFalse(t, page3.Pagination.HasMore, "Should not have more pages")
}

func TestAggregationService_GetInventoryList_SortByExpiry(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create batches with different expiry dates
	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	expiry30Days := time.Now().Add(30 * 24 * time.Hour)
	expiry60Days := time.Now().Add(60 * 24 * time.Hour)
	expiry10Days := time.Now().Add(10 * 24 * time.Hour)

	batch1 := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 100, expiry30Days)
	batch2 := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 50, expiry60Days)
	batch3 := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 75, expiry10Days)
	db.Create(batch1)
	db.Create(batch2)
	db.Create(batch3)

	// Execute with sort by expiry ascending
	response, err := service.GetInventoryList(&models.InventoryListRequest{
		SortBy:    "expiry_date",
		SortOrder: "asc",
		Include:   "all",
	})

	// Assert
	testutils.AssertNoError(t, err, "GetInventoryList should succeed")
	testutils.AssertEqual(t, len(response.Batches), 3, "Should have 3 batches")
	// First should be the one expiring soonest (10 days)
	testutils.AssertEqual(t, response.Batches[0].ID, batch3.ID, "First should be batch expiring in 10 days")
}

func TestAggregationService_GetInventoryList_Summary(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create batches with various states
	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	expiryNormal := time.Now().Add(60 * 24 * time.Hour)
	expirySoon := time.Now().Add(15 * 24 * time.Hour)

	batch1 := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 100, expiryNormal)
	batch2 := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 50, expirySoon)
	batch3 := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 0, expiryNormal) // Zero stock
	db.Create(batch1)
	db.Create(batch2)
	db.Create(batch3)

	// Execute
	lowStock := int64(25)
	response, err := service.GetInventoryList(&models.InventoryListRequest{
		Include:           "all",
		LowStockThreshold: &lowStock,
	})

	// Assert
	testutils.AssertNoError(t, err, "GetInventoryList should succeed")
	testutils.AssertEqual(t, response.Summary.TotalBatches, 3, "Total batches should be 3")
	testutils.AssertEqual(t, response.Summary.TotalStockQuantity, int64(150), "Total stock should be 150")
	testutils.AssertEqual(t, response.Summary.ExpiringSoonCount, 1, "Expiring soon should be 1")
	testutils.AssertEqual(t, response.Summary.ZeroStockCount, 1, "Zero stock should be 1")
}

func TestAggregationService_GetInventoryList_ExpiryStatus(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create batches with different expiry states
	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	expiryGood := time.Now().Add(60 * 24 * time.Hour)    // Good (>30 days)
	expiryWarning := time.Now().Add(20 * 24 * time.Hour) // Warning (7-30 days)
	expiryCritical := time.Now().Add(5 * 24 * time.Hour) // Critical (<=7 days)
	expiryExpired := time.Now().Add(-5 * 24 * time.Hour) // Expired (past)

	batchGood := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 100, expiryGood)
	batchWarning := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 50, expiryWarning)
	batchCritical := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 25, expiryCritical)
	batchExpired := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 10, expiryExpired)
	db.Create(batchGood)
	db.Create(batchWarning)
	db.Create(batchCritical)
	db.Create(batchExpired)

	// Execute
	response, err := service.GetInventoryList(&models.InventoryListRequest{
		Include: "all",
	})

	// Assert
	testutils.AssertNoError(t, err, "GetInventoryList should succeed")
	testutils.AssertEqual(t, len(response.Batches), 4, "Should have 4 batches")

	// Find batches by ID and check their expiry status
	for _, b := range response.Batches {
		switch b.ID {
		case batchGood.ID:
			testutils.AssertEqual(t, b.BatchInfo.ExpiryStatus, "good", "Good batch should have 'good' status")
		case batchWarning.ID:
			testutils.AssertEqual(t, b.BatchInfo.ExpiryStatus, "warning", "Warning batch should have 'warning' status")
		case batchCritical.ID:
			testutils.AssertEqual(t, b.BatchInfo.ExpiryStatus, "critical", "Critical batch should have 'critical' status")
		case batchExpired.ID:
			testutils.AssertEqual(t, b.BatchInfo.ExpiryStatus, "expired", "Expired batch should have 'expired' status")
		}
	}
}

func TestAggregationService_GetInventoryList_WithContext(t *testing.T) {
	service, db, cleanup := setupAggregationService(t)
	defer cleanup()

	// Create full context
	warehouse := testutils.CreateTestWarehouse(t, db, "WH-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Test Product")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "VAR-SKU-001", "1kg")

	expiryDate := time.Now().Add(30 * 24 * time.Hour)
	batch := testutils.FixtureInventoryBatchWithExpiry(warehouse.ID, variant.ID, 100, expiryDate)
	db.Create(batch)

	// Execute with all includes
	response, err := service.GetInventoryList(&models.InventoryListRequest{
		Include: "variant,product,warehouse,taxes",
	})

	// Assert
	testutils.AssertNoError(t, err, "GetInventoryList should succeed")
	testutils.AssertEqual(t, len(response.Batches), 1, "Should have 1 batch")

	batchResp := response.Batches[0]
	testutils.AssertNotNil(t, batchResp.Variant, "Variant should be included")
	testutils.AssertNotNil(t, batchResp.Product, "Product should be included")
	testutils.AssertNotNil(t, batchResp.Warehouse, "Warehouse should be included")
	testutils.AssertNotNil(t, batchResp.TaxConfig, "TaxConfig should be included")

	testutils.AssertEqual(t, batchResp.Variant.ID, variant.ID, "Variant ID mismatch")
	testutils.AssertEqual(t, batchResp.Product.ID, product.ID, "Product ID mismatch")
	testutils.AssertEqual(t, batchResp.Warehouse.ID, warehouse.ID, "Warehouse ID mismatch")
}
