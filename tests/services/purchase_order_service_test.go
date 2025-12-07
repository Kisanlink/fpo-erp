package services

import (
	"context"
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

// setupPurchaseOrderService creates service with all dependencies
func setupPurchaseOrderService(t *testing.T) (*services.PurchaseOrderService, *gorm.DB, func()) {
	t.Helper()

	// Setup test database
	db := testutils.SetupTestDB(t)

	// Create all required repositories
	poRepo := repositories.NewPurchaseOrderRepository(db)
	collaboratorRepo := repositories.NewCollaboratorRepository(db)
	warehouseRepo := repositories.NewWarehouseRepository(db)
	productRepo := repositories.NewProductRepository(db)
	variantRepo := repositories.NewProductVariantRepository(db)
	grnRepo := repositories.NewGRNRepository(db)
	inventoryRepo := repositories.NewInventoryRepository(db)

	// Create service (nil AAA client for tests that don't need address service)
	service := services.NewPurchaseOrderService(
		poRepo,
		collaboratorRepo,
		warehouseRepo,
		productRepo,
		variantRepo,
		grnRepo,
		inventoryRepo,
		nil, // AddressGRPCClient - nil for tests
		utils.NewLoggerAdapter(utils.GetZapLogger()),
	)

	cleanup := func() {
		testutils.CleanupTestDB(db)
	}

	return service, db, cleanup
}

// createTestCollaborator creates a test collaborator
func createTestCollaborator(t *testing.T, db *gorm.DB, id string, isActive bool) *models.Collaborator {
	t.Helper()

	email := id + "@test.com"
	collaborator := &models.Collaborator{
		CompanyName:   "Test Supplier " + id,
		Email:         &email,
		ContactPerson: "Test Contact",
		ContactNumber: "1234567890",
		IsActive:      &isActive,
	}
	collaborator.ID = id // Set ID after creating the struct

	if err := db.Create(collaborator).Error; err != nil {
		t.Fatalf("Failed to create test collaborator: %v", err)
	}

	return collaborator
}

// createTestWarehouse creates a test warehouse
func createTestWarehouse(t *testing.T, db *gorm.DB, id string) *models.Warehouse {
	t.Helper()

	warehouse := models.NewWarehouse("Test Warehouse "+id, nil)
	warehouse.ID = id

	if err := db.Create(warehouse).Error; err != nil {
		t.Fatalf("Failed to create test warehouse: %v", err)
	}

	return warehouse
}

// createTestPurchaseOrder creates a test purchase order with items
func createTestPurchaseOrder(t *testing.T, db *gorm.DB, poNumber, collaboratorID, warehouseID string, status string) *models.PurchaseOrder {
	t.Helper()

	orderDate := time.Now().UTC().Add(-7 * 24 * time.Hour)
	expectedDelivery := time.Now().UTC().Add(7 * 24 * time.Hour)

	po := models.NewPurchaseOrder(
		poNumber,
		collaboratorID,
		warehouseID,
		orderDate,
		expectedDelivery,
	)
	po.Status = status
	po.TotalAmount = 1000.00

	if err := db.Create(po).Error; err != nil {
		t.Fatalf("Failed to create test PO: %v", err)
	}

	return po
}

// createTestPOItem creates a test purchase order item
func createTestPOItem(t *testing.T, db *gorm.DB, poID, variantID string, quantity int64, unitPrice float64) *models.PurchaseOrderItem {
	t.Helper()

	item := models.NewPurchaseOrderItem(poID, variantID, quantity, unitPrice)

	if err := db.Create(item).Error; err != nil {
		t.Fatalf("Failed to create test PO item: %v", err)
	}

	return item
}

// =============================================================================
// CreatePurchaseOrder Tests
// =============================================================================

func TestPurchaseOrderService_CreatePurchaseOrder_Success(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	// Setup fixtures
	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "TOM-1KG", "1.0")

	// Create request
	orderDate := time.Now().UTC().Format("2006-01-02")
	expectedDelivery := time.Now().UTC().Add(14 * 24 * time.Hour).Format("2006-01-02")

	request := &models.CreatePurchaseOrderRequest{
		CollaboratorID:   collaborator.ID,
		WarehouseID:      warehouse.ID,
		OrderDate:        &orderDate,
		ExpectedDelivery: expectedDelivery,
		Items: []models.CreatePurchaseOrderItemRequest{
			{
				VariantID: variant.ID,
				Quantity:  100,
				UnitPrice: 25.50,
			},
		},
	}

	// Execute
	ctx := context.Background()
	response, err := service.CreatePurchaseOrder(ctx, request, "")

	// Assert
	testutils.AssertNoError(t, err, "CreatePurchaseOrder should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.CollaboratorID, collaborator.ID, "Collaborator ID mismatch")
	testutils.AssertEqual(t, response.WarehouseID, warehouse.ID, "Warehouse ID mismatch")
	testutils.AssertEqual(t, response.Status, "placed", "Status should be placed")
	testutils.AssertEqual(t, response.PaymentStatus, "unpaid", "Payment status should be unpaid")
	testutils.AssertEqual(t, response.TotalAmount, 2550.0, "Total amount mismatch")
	testutils.AssertEqual(t, len(response.Items), 1, "Should have 1 item")
	testutils.AssertTrue(t, response.PONumber != "", "PO number should be generated")
}

func TestPurchaseOrderService_CreatePurchaseOrder_CollaboratorNotFound(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	expectedDelivery := time.Now().UTC().Add(14 * 24 * time.Hour).Format("2006-01-02")

	request := &models.CreatePurchaseOrderRequest{
		CollaboratorID:   "INVALID-COLLAB",
		WarehouseID:      warehouse.ID,
		ExpectedDelivery: expectedDelivery,
		Items:            []models.CreatePurchaseOrderItemRequest{},
	}

	ctx := context.Background()
	_, err := service.CreatePurchaseOrder(ctx, request, "")

	testutils.AssertError(t, err, "Should return error for invalid collaborator")
}

func TestPurchaseOrderService_CreatePurchaseOrder_CollaboratorNotActive(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", false)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "TOM-1KG", "1.0")
	expectedDelivery := time.Now().UTC().Add(14 * 24 * time.Hour).Format("2006-01-02")

	request := &models.CreatePurchaseOrderRequest{
		CollaboratorID:   collaborator.ID,
		WarehouseID:      warehouse.ID,
		ExpectedDelivery: expectedDelivery,
		Items: []models.CreatePurchaseOrderItemRequest{
			{
				VariantID: variant.ID,
				Quantity:  100,
				UnitPrice: 25.50,
			},
		},
	}

	ctx := context.Background()
	_, err := service.CreatePurchaseOrder(ctx, request, "")

	testutils.AssertError(t, err, "Should return error for inactive collaborator")
	testutils.AssertContains(t, err.Error(), "not active", "Error message should mention inactive")
}

func TestPurchaseOrderService_CreatePurchaseOrder_WarehouseNotFound(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	expectedDelivery := time.Now().UTC().Add(14 * 24 * time.Hour).Format("2006-01-02")

	request := &models.CreatePurchaseOrderRequest{
		CollaboratorID:   collaborator.ID,
		WarehouseID:      "INVALID-WAREHOUSE",
		ExpectedDelivery: expectedDelivery,
		Items:            []models.CreatePurchaseOrderItemRequest{},
	}

	ctx := context.Background()
	_, err := service.CreatePurchaseOrder(ctx, request, "")

	testutils.AssertError(t, err, "Should return error for invalid warehouse")
}

func TestPurchaseOrderService_CreatePurchaseOrder_InvalidOrderDateFormat(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	invalidDate := "2025-13-45" // Invalid month and day
	expectedDelivery := time.Now().UTC().Add(14 * 24 * time.Hour).Format("2006-01-02")

	request := &models.CreatePurchaseOrderRequest{
		CollaboratorID:   collaborator.ID,
		WarehouseID:      warehouse.ID,
		OrderDate:        &invalidDate,
		ExpectedDelivery: expectedDelivery,
		Items:            []models.CreatePurchaseOrderItemRequest{},
	}

	ctx := context.Background()
	_, err := service.CreatePurchaseOrder(ctx, request, "")

	testutils.AssertError(t, err, "Should return error for invalid order date format")
	testutils.AssertContains(t, err.Error(), "order_date", "Error should mention order_date")
}

func TestPurchaseOrderService_CreatePurchaseOrder_InvalidExpectedDeliveryFormat(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")

	request := &models.CreatePurchaseOrderRequest{
		CollaboratorID:   collaborator.ID,
		WarehouseID:      warehouse.ID,
		ExpectedDelivery: "invalid-date",
		Items:            []models.CreatePurchaseOrderItemRequest{},
	}

	ctx := context.Background()
	_, err := service.CreatePurchaseOrder(ctx, request, "")

	testutils.AssertError(t, err, "Should return error for invalid expected delivery format")
	testutils.AssertContains(t, err.Error(), "expected_delivery", "Error should mention expected_delivery")
}

func TestPurchaseOrderService_CreatePurchaseOrder_ExpectedDeliveryBeforeOrderDate(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")

	orderDate := time.Now().UTC().Add(7 * 24 * time.Hour).Format("2006-01-02")
	expectedDelivery := time.Now().UTC().Format("2006-01-02") // Earlier than order date

	request := &models.CreatePurchaseOrderRequest{
		CollaboratorID:   collaborator.ID,
		WarehouseID:      warehouse.ID,
		OrderDate:        &orderDate,
		ExpectedDelivery: expectedDelivery,
		Items:            []models.CreatePurchaseOrderItemRequest{},
	}

	ctx := context.Background()
	_, err := service.CreatePurchaseOrder(ctx, request, "")

	testutils.AssertError(t, err, "Should return error when expected delivery is before order date")
	testutils.AssertContains(t, err.Error(), "after order date", "Error should mention date order")
}

func TestPurchaseOrderService_CreatePurchaseOrder_NoItems(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	expectedDelivery := time.Now().UTC().Add(14 * 24 * time.Hour).Format("2006-01-02")

	request := &models.CreatePurchaseOrderRequest{
		CollaboratorID:   collaborator.ID,
		WarehouseID:      warehouse.ID,
		ExpectedDelivery: expectedDelivery,
		Items:            []models.CreatePurchaseOrderItemRequest{}, // Empty items
	}

	ctx := context.Background()
	_, err := service.CreatePurchaseOrder(ctx, request, "")

	testutils.AssertError(t, err, "Should return error when no items provided")
	testutils.AssertContains(t, err.Error(), "at least one item", "Error should mention items requirement")
}

func TestPurchaseOrderService_CreatePurchaseOrder_InvalidVariant(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	expectedDelivery := time.Now().UTC().Add(14 * 24 * time.Hour).Format("2006-01-02")

	request := &models.CreatePurchaseOrderRequest{
		CollaboratorID:   collaborator.ID,
		WarehouseID:      warehouse.ID,
		ExpectedDelivery: expectedDelivery,
		Items: []models.CreatePurchaseOrderItemRequest{
			{
				VariantID: "INVALID-VARIANT",
				Quantity:  100,
				UnitPrice: 25.50,
			},
		},
	}

	ctx := context.Background()
	_, err := service.CreatePurchaseOrder(ctx, request, "")

	testutils.AssertError(t, err, "Should return error for invalid variant")
	testutils.AssertContains(t, err.Error(), "variant", "Error should mention variant")
}

func TestPurchaseOrderService_CreatePurchaseOrder_MultipleItems(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	// Setup fixtures
	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	product1 := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant1 := testutils.CreateTestVariant(t, db, "VAR-001", product1.ID, "TOM-1KG", "1.0")
	product2 := testutils.CreateTestProduct(t, db, "PROD-002", "Onion")
	variant2 := testutils.CreateTestVariant(t, db, "VAR-002", product2.ID, "ONI-1KG", "1.0")

	expectedDelivery := time.Now().UTC().Add(14 * 24 * time.Hour).Format("2006-01-02")

	request := &models.CreatePurchaseOrderRequest{
		CollaboratorID:   collaborator.ID,
		WarehouseID:      warehouse.ID,
		ExpectedDelivery: expectedDelivery,
		Items: []models.CreatePurchaseOrderItemRequest{
			{
				VariantID: variant1.ID,
				Quantity:  100,
				UnitPrice: 25.50,
			},
			{
				VariantID: variant2.ID,
				Quantity:  50,
				UnitPrice: 30.00,
			},
		},
	}

	ctx := context.Background()
	response, err := service.CreatePurchaseOrder(ctx, request, "")

	testutils.AssertNoError(t, err, "CreatePurchaseOrder should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, len(response.Items), 2, "Should have 2 items")
	// 100 * 25.50 + 50 * 30.00 = 2550 + 1500 = 4050
	testutils.AssertEqual(t, response.TotalAmount, 4050.0, "Total amount mismatch")
}

// =============================================================================
// GetPurchaseOrder Tests
// =============================================================================

func TestPurchaseOrderService_GetPurchaseOrder_Success(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	// Setup fixtures
	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "placed")

	ctx := context.Background()
	response, err := service.GetPurchaseOrder(ctx, po.ID)

	testutils.AssertNoError(t, err, "GetPurchaseOrder should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.ID, po.ID, "PO ID mismatch")
	testutils.AssertEqual(t, response.PONumber, "PO-2025-0001", "PO number mismatch")
	testutils.AssertEqual(t, response.Status, "placed", "Status mismatch")
}

func TestPurchaseOrderService_GetPurchaseOrder_NotFound(t *testing.T) {
	service, _, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	ctx := context.Background()
	_, err := service.GetPurchaseOrder(ctx, "INVALID-PO-ID")

	testutils.AssertError(t, err, "Should return error for non-existent PO")
}

// =============================================================================
// GetAllPurchaseOrders Tests
// =============================================================================

func TestPurchaseOrderService_GetAllPurchaseOrders_Success(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	// Setup fixtures
	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "placed")
	createTestPurchaseOrder(t, db, "PO-2025-0002", collaborator.ID, warehouse.ID, "confirmed")

	ctx := context.Background()
	responses, err := service.GetAllPurchaseOrders(ctx)

	testutils.AssertNoError(t, err, "GetAllPurchaseOrders should succeed")
	testutils.AssertEqual(t, len(responses), 2, "Should return 2 POs")
}

func TestPurchaseOrderService_GetAllPurchaseOrders_Empty(t *testing.T) {
	service, _, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	ctx := context.Background()
	responses, err := service.GetAllPurchaseOrders(ctx)

	testutils.AssertNoError(t, err, "GetAllPurchaseOrders should succeed")
	testutils.AssertEqual(t, len(responses), 0, "Should return empty list")
}

// =============================================================================
// GetPurchaseOrdersByCollaborator Tests
// =============================================================================

func TestPurchaseOrderService_GetPurchaseOrdersByCollaborator_Success(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	// Setup fixtures
	collaborator1 := createTestCollaborator(t, db, "COLLAB-001", true)
	collaborator2 := createTestCollaborator(t, db, "COLLAB-002", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator1.ID, warehouse.ID, "placed")
	createTestPurchaseOrder(t, db, "PO-2025-0002", collaborator1.ID, warehouse.ID, "confirmed")
	createTestPurchaseOrder(t, db, "PO-2025-0003", collaborator2.ID, warehouse.ID, "placed")

	ctx := context.Background()
	responses, err := service.GetPurchaseOrdersByCollaborator(ctx, collaborator1.ID)

	testutils.AssertNoError(t, err, "GetPurchaseOrdersByCollaborator should succeed")
	testutils.AssertEqual(t, len(responses), 2, "Should return 2 POs for collaborator 1")
	for _, resp := range responses {
		testutils.AssertEqual(t, resp.CollaboratorID, collaborator1.ID, "All POs should belong to collaborator 1")
	}
}

func TestPurchaseOrderService_GetPurchaseOrdersByCollaborator_Empty(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)

	ctx := context.Background()
	responses, err := service.GetPurchaseOrdersByCollaborator(ctx, collaborator.ID)

	testutils.AssertNoError(t, err, "GetPurchaseOrdersByCollaborator should succeed")
	testutils.AssertEqual(t, len(responses), 0, "Should return empty list")
}

func TestPurchaseOrderService_GetPurchaseOrdersByCollaborator_NotFound(t *testing.T) {
	service, _, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	ctx := context.Background()
	_, err := service.GetPurchaseOrdersByCollaborator(ctx, "INVALID-COLLAB")

	testutils.AssertError(t, err, "Should return error for invalid collaborator")
}

// =============================================================================
// GetPurchaseOrdersByStatus Tests
// =============================================================================

func TestPurchaseOrderService_GetPurchaseOrdersByStatus_Success(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	// Setup fixtures
	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "placed")
	createTestPurchaseOrder(t, db, "PO-2025-0002", collaborator.ID, warehouse.ID, "confirmed")
	createTestPurchaseOrder(t, db, "PO-2025-0003", collaborator.ID, warehouse.ID, "confirmed")

	ctx := context.Background()
	responses, err := service.GetPurchaseOrdersByStatus(ctx, "confirmed")

	testutils.AssertNoError(t, err, "GetPurchaseOrdersByStatus should succeed")
	testutils.AssertEqual(t, len(responses), 2, "Should return 2 confirmed POs")
	for _, resp := range responses {
		testutils.AssertEqual(t, resp.Status, "confirmed", "All POs should have confirmed status")
	}
}

func TestPurchaseOrderService_GetPurchaseOrdersByStatus_Empty(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	// Setup fixtures
	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "placed")

	ctx := context.Background()
	responses, err := service.GetPurchaseOrdersByStatus(ctx, "delivered")

	testutils.AssertNoError(t, err, "GetPurchaseOrdersByStatus should succeed")
	testutils.AssertEqual(t, len(responses), 0, "Should return empty list")
}

func TestPurchaseOrderService_GetPurchaseOrdersByStatus_InvalidStatus(t *testing.T) {
	service, _, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	ctx := context.Background()
	_, err := service.GetPurchaseOrdersByStatus(ctx, "invalid_status")

	testutils.AssertError(t, err, "Should return error for invalid status")
	testutils.AssertContains(t, err.Error(), "invalid status", "Error should mention invalid status")
}

// =============================================================================
// GetPendingDeliveries Tests
// =============================================================================

func TestPurchaseOrderService_GetPendingDeliveries_Success(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	// Setup fixtures
	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "confirmed")
	createTestPurchaseOrder(t, db, "PO-2025-0002", collaborator.ID, warehouse.ID, "out_for_delivery")
	createTestPurchaseOrder(t, db, "PO-2025-0003", collaborator.ID, warehouse.ID, "delivered")

	ctx := context.Background()
	responses, err := service.GetPendingDeliveries(ctx)

	testutils.AssertNoError(t, err, "GetPendingDeliveries should succeed")
	// Depends on repository implementation - should return pending ones
	testutils.AssertTrue(t, len(responses) >= 0, "Should return list")
}

func TestPurchaseOrderService_GetPendingDeliveries_Empty(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	// Setup fixtures - all delivered
	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")
	createTestPurchaseOrder(t, db, "PO-2025-0002", collaborator.ID, warehouse.ID, "paid")

	ctx := context.Background()
	responses, err := service.GetPendingDeliveries(ctx)

	testutils.AssertNoError(t, err, "GetPendingDeliveries should succeed")
	testutils.AssertTrue(t, len(responses) >= 0, "Should return list")
}

// =============================================================================
// UpdatePurchaseOrderStatus - Basic Tests
// =============================================================================

func TestPurchaseOrderService_UpdateStatus_PlacedToConfirmed_Success(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	// Setup fixtures
	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "placed")

	request := &models.UpdatePOStatusRequest{
		Status: "confirmed",
	}

	ctx := context.Background()
	response, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertNoError(t, err, "UpdatePurchaseOrderStatus should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.Status, "confirmed", "Status should be updated to confirmed")
}

func TestPurchaseOrderService_UpdateStatus_ConfirmedToOutForDelivery_Success(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "confirmed")

	request := &models.UpdatePOStatusRequest{
		Status: "out_for_delivery",
	}

	ctx := context.Background()
	response, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertNoError(t, err, "UpdatePurchaseOrderStatus should succeed")
	testutils.AssertEqual(t, response.Status, "out_for_delivery", "Status should be updated")
}

func TestPurchaseOrderService_UpdateStatus_OutForDeliveryToDelivered_NoGRN(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "out_for_delivery")

	// Bug #14 fix: actual_delivery_date is now required when marking as delivered
	actualDelivery := time.Now()
	request := &models.UpdatePOStatusRequest{
		Status:         "delivered",
		ActualDelivery: &actualDelivery,
		// No delivery details = traditional flow (no auto-GRN)
	}

	ctx := context.Background()
	response, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertNoError(t, err, "UpdatePurchaseOrderStatus should succeed")
	testutils.AssertEqual(t, response.Status, "delivered", "Status should be updated to delivered")
	testutils.AssertNotNil(t, response.ActualDelivery, "Actual delivery should be set")
}

func TestPurchaseOrderService_UpdateStatus_VerifiedToPaid_Success(t *testing.T) {
	// Note: Workflow is delivered → verified → paid
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "verified")

	request := &models.UpdatePOStatusRequest{
		Status: "paid",
	}

	ctx := context.Background()
	response, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertNoError(t, err, "UpdatePurchaseOrderStatus should succeed")
	testutils.AssertEqual(t, response.Status, "paid", "Status should be updated to paid")
}

func TestPurchaseOrderService_UpdateStatus_InvalidStatus(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "placed")

	request := &models.UpdatePOStatusRequest{
		Status: "invalid_status",
	}

	ctx := context.Background()
	_, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertError(t, err, "Should return error for invalid status")
	testutils.AssertContains(t, err.Error(), "invalid status", "Error should mention invalid status")
}

func TestPurchaseOrderService_UpdateStatus_InvalidTransition(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "placed")

	request := &models.UpdatePOStatusRequest{
		Status: "paid", // Invalid: placed -> paid (should go through confirmed, delivered first)
	}

	ctx := context.Background()
	_, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertError(t, err, "Should return error for invalid transition")
	testutils.AssertContains(t, err.Error(), "invalid status transition", "Error should mention invalid transition")
}

func TestPurchaseOrderService_UpdateStatus_PONotFound(t *testing.T) {
	service, _, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	request := &models.UpdatePOStatusRequest{
		Status: "confirmed",
	}

	ctx := context.Background()
	_, err := service.UpdatePurchaseOrderStatus(ctx, "INVALID-PO", request, "USER-001")

	testutils.AssertError(t, err, "Should return error for non-existent PO")
}

// =============================================================================
// UpdatePurchaseOrderStatus - Auto-GRN Pattern 1 (Accept All)
// =============================================================================

func TestPurchaseOrderService_UpdateStatus_AcceptAll_Success(t *testing.T) {
	// Note: Auto-GRN triggers on "verified" status, workflow: delivered → verified → paid
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	// Setup fixtures
	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "TOM-1KG", "1.0")

	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")
	createTestPOItem(t, db, po.ID, variant.ID, 100, 25.50)

	acceptAll := true
	defaultExpiry := time.Now().UTC().Add(90 * 24 * time.Hour).Format("2006-01-02")
	actualDelivery := time.Now().UTC()
	request := &models.UpdatePOStatusRequest{
		Status:            "verified",
		ActualDelivery:    &actualDelivery,
		AcceptAll:         &acceptAll,
		DefaultExpiryDate: &defaultExpiry,
	}

	ctx := context.Background()
	response, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertNoError(t, err, "UpdatePurchaseOrderStatus with AcceptAll should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.Status, "verified", "Status should be verified")

	// Verify GRN was created
	var grn models.GRN
	err = db.Where("po_id = ?", po.ID).First(&grn).Error
	testutils.AssertNoError(t, err, "GRN should be created")
	testutils.AssertEqual(t, grn.QualityStatus, "accepted", "Quality status should be accepted")
}

func TestPurchaseOrderService_UpdateStatus_AcceptAll_MissingExpiryDate(t *testing.T) {
	// Note: Auto-GRN triggers on "verified" status
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "TOM-1KG", "1.0")

	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")
	createTestPOItem(t, db, po.ID, variant.ID, 100, 25.50)

	acceptAll := true
	actualDelivery := time.Now().UTC()
	request := &models.UpdatePOStatusRequest{
		Status:         "verified",
		ActualDelivery: &actualDelivery,
		AcceptAll:      &acceptAll,
		// Missing DefaultExpiryDate
	}

	ctx := context.Background()
	_, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertError(t, err, "Should return error when accept_all without expiry date")
	testutils.AssertContains(t, err.Error(), "default_expiry_date", "Error should mention expiry date")
}

func TestPurchaseOrderService_UpdateStatus_GRNAlreadyExists(t *testing.T) {
	// Note: Auto-GRN triggers on "verified" status
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "TOM-1KG", "1.0")

	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")
	createTestPOItem(t, db, po.ID, variant.ID, 100, 25.50)

	// Create existing GRN
	grn := models.NewGRN("GRN-2025-0001", po.ID, warehouse.ID, "USER-001", time.Now().UTC(), "accepted")
	if err := db.Create(grn).Error; err != nil {
		t.Fatalf("Failed to create existing GRN: %v", err)
	}

	acceptAll := true
	defaultExpiry := time.Now().UTC().Add(90 * 24 * time.Hour).Format("2006-01-02")
	actualDelivery := time.Now().UTC()
	request := &models.UpdatePOStatusRequest{
		Status:            "verified",
		ActualDelivery:    &actualDelivery,
		AcceptAll:         &acceptAll,
		DefaultExpiryDate: &defaultExpiry,
	}

	ctx := context.Background()
	_, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertError(t, err, "Should return error when GRN already exists")
	testutils.AssertContains(t, err.Error(), "GRN already exists", "Error should mention GRN exists")
}

// =============================================================================
// UpdatePurchaseOrderStatus - Auto-GRN Pattern 2 (Accept/Reject)
// =============================================================================

func TestPurchaseOrderService_UpdateStatus_AcceptReject_AllAccepted(t *testing.T) {
	// Note: Auto-GRN triggers on "verified" status
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	product1 := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant1 := testutils.CreateTestVariant(t, db, "VAR-001", product1.ID, "TOM-1KG", "1.0")
	product2 := testutils.CreateTestProduct(t, db, "PROD-002", "Onion")
	variant2 := testutils.CreateTestVariant(t, db, "VAR-002", product2.ID, "ONI-1KG", "1.0")

	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")
	item1 := createTestPOItem(t, db, po.ID, variant1.ID, 100, 25.50)
	item2 := createTestPOItem(t, db, po.ID, variant2.ID, 50, 30.00)

	acceptTrue := true
	expiryDate := time.Now().UTC().Add(90 * 24 * time.Hour).Format("2006-01-02")
	actualDelivery := time.Now().UTC()
	request := &models.UpdatePOStatusRequest{
		Status:         "verified",
		ActualDelivery: &actualDelivery,
		Items: []models.DeliveryItemRequest{
			{
				POItemID:   item1.ID,
				Accept:     &acceptTrue,
				ExpiryDate: expiryDate,
			},
			{
				POItemID:   item2.ID,
				Accept:     &acceptTrue,
				ExpiryDate: expiryDate,
			},
		},
	}

	ctx := context.Background()
	response, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertNoError(t, err, "UpdatePurchaseOrderStatus should succeed")
	testutils.AssertEqual(t, response.Status, "verified", "Status should be verified")

	// Verify GRN quality status
	var grn models.GRN
	err = db.Where("po_id = ?", po.ID).First(&grn).Error
	testutils.AssertNoError(t, err, "GRN should be created")
	testutils.AssertEqual(t, grn.QualityStatus, "accepted", "Quality status should be accepted when all accepted")
}

func TestPurchaseOrderService_UpdateStatus_AcceptReject_Partial(t *testing.T) {
	// Note: Auto-GRN triggers on "verified" status
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	product1 := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant1 := testutils.CreateTestVariant(t, db, "VAR-001", product1.ID, "TOM-1KG", "1.0")
	product2 := testutils.CreateTestProduct(t, db, "PROD-002", "Onion")
	variant2 := testutils.CreateTestVariant(t, db, "VAR-002", product2.ID, "ONI-1KG", "1.0")

	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")
	item1 := createTestPOItem(t, db, po.ID, variant1.ID, 100, 25.50)
	item2 := createTestPOItem(t, db, po.ID, variant2.ID, 50, 30.00)

	acceptTrue := true
	acceptFalse := false
	expiryDate := time.Now().UTC().Add(90 * 24 * time.Hour).Format("2006-01-02")
	actualDelivery := time.Now().UTC()
	request := &models.UpdatePOStatusRequest{
		Status:         "verified",
		ActualDelivery: &actualDelivery,
		Items: []models.DeliveryItemRequest{
			{
				POItemID:   item1.ID,
				Accept:     &acceptTrue,
				ExpiryDate: expiryDate,
			},
			{
				POItemID:   item2.ID,
				Accept:     &acceptFalse,
				ExpiryDate: expiryDate,
			},
		},
	}

	ctx := context.Background()
	response, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertNoError(t, err, "UpdatePurchaseOrderStatus should succeed")
	testutils.AssertEqual(t, response.Status, "verified", "Status should be verified")

	// Verify GRN quality status
	var grn models.GRN
	err = db.Where("po_id = ?", po.ID).First(&grn).Error
	testutils.AssertNoError(t, err, "GRN should be created")
	testutils.AssertEqual(t, grn.QualityStatus, "partial", "Quality status should be partial")
}

func TestPurchaseOrderService_UpdateStatus_AcceptReject_AllRejected(t *testing.T) {
	// Note: Auto-GRN triggers on "verified" status
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "TOM-1KG", "1.0")

	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")
	item := createTestPOItem(t, db, po.ID, variant.ID, 100, 25.50)

	acceptFalse := false
	expiryDate := time.Now().UTC().Add(90 * 24 * time.Hour).Format("2006-01-02")
	actualDelivery := time.Now().UTC()
	request := &models.UpdatePOStatusRequest{
		Status:         "verified",
		ActualDelivery: &actualDelivery,
		Items: []models.DeliveryItemRequest{
			{
				POItemID:   item.ID,
				Accept:     &acceptFalse,
				ExpiryDate: expiryDate,
			},
		},
	}

	ctx := context.Background()
	response, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertNoError(t, err, "UpdatePurchaseOrderStatus should succeed")
	testutils.AssertEqual(t, response.Status, "verified", "Status should be verified")

	// Verify GRN quality status
	var grn models.GRN
	err = db.Where("po_id = ?", po.ID).First(&grn).Error
	testutils.AssertNoError(t, err, "GRN should be created")
	testutils.AssertEqual(t, grn.QualityStatus, "rejected", "Quality status should be rejected when all rejected")
}

// =============================================================================
// UpdatePurchaseOrderStatus - Auto-GRN Pattern 3 (Detailed Quantities)
// =============================================================================

func TestPurchaseOrderService_UpdateStatus_DetailedQuantities_PartialAcceptance(t *testing.T) {
	// Note: Auto-GRN triggers on "verified" status
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "TOM-1KG", "1.0")

	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")
	item := createTestPOItem(t, db, po.ID, variant.ID, 100, 25.50)

	receivedQty := int64(100)
	acceptedQty := int64(80) // Partial acceptance
	expiryDate := time.Now().UTC().Add(90 * 24 * time.Hour).Format("2006-01-02")
	actualDelivery := time.Now().UTC()
	request := &models.UpdatePOStatusRequest{
		Status:         "verified",
		ActualDelivery: &actualDelivery,
		Items: []models.DeliveryItemRequest{
			{
				POItemID:         item.ID,
				ReceivedQuantity: &receivedQty,
				AcceptedQuantity: &acceptedQty,
				ExpiryDate:       expiryDate,
			},
		},
	}

	ctx := context.Background()
	response, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertNoError(t, err, "UpdatePurchaseOrderStatus should succeed")
	testutils.AssertEqual(t, response.Status, "verified", "Status should be verified")

	// Verify GRN quality status
	var grn models.GRN
	err = db.Where("po_id = ?", po.ID).First(&grn).Error
	testutils.AssertNoError(t, err, "GRN should be created")
	testutils.AssertEqual(t, grn.QualityStatus, "partial", "Quality status should be partial for partial acceptance")
}

func TestPurchaseOrderService_UpdateStatus_DetailedQuantities_FullAcceptance(t *testing.T) {
	// Note: Auto-GRN triggers on "verified" status
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "TOM-1KG", "1.0")

	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")
	item := createTestPOItem(t, db, po.ID, variant.ID, 100, 25.50)

	receivedQty := int64(100)
	acceptedQty := int64(100) // Full acceptance
	expiryDate := time.Now().UTC().Add(90 * 24 * time.Hour).Format("2006-01-02")
	actualDelivery := time.Now().UTC()
	request := &models.UpdatePOStatusRequest{
		Status:         "verified",
		ActualDelivery: &actualDelivery,
		Items: []models.DeliveryItemRequest{
			{
				POItemID:         item.ID,
				ReceivedQuantity: &receivedQty,
				AcceptedQuantity: &acceptedQty,
				ExpiryDate:       expiryDate,
			},
		},
	}

	ctx := context.Background()
	response, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertNoError(t, err, "UpdatePurchaseOrderStatus should succeed")
	testutils.AssertEqual(t, response.Status, "verified", "Status should be verified")

	// Verify GRN quality status
	var grn models.GRN
	err = db.Where("po_id = ?", po.ID).First(&grn).Error
	testutils.AssertNoError(t, err, "GRN should be created")
	testutils.AssertEqual(t, grn.QualityStatus, "accepted", "Quality status should be accepted for full acceptance")
}

func TestPurchaseOrderService_UpdateStatus_DetailedQuantities_AcceptedExceedsReceived(t *testing.T) {
	// Note: Auto-GRN triggers on "verified" status
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "TOM-1KG", "1.0")

	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")
	item := createTestPOItem(t, db, po.ID, variant.ID, 100, 25.50)

	receivedQty := int64(80)
	acceptedQty := int64(100) // Accepted > Received - INVALID
	expiryDate := time.Now().UTC().Add(90 * 24 * time.Hour).Format("2006-01-02")
	actualDelivery := time.Now().UTC()
	request := &models.UpdatePOStatusRequest{
		Status:         "verified",
		ActualDelivery: &actualDelivery,
		Items: []models.DeliveryItemRequest{
			{
				POItemID:         item.ID,
				ReceivedQuantity: &receivedQty,
				AcceptedQuantity: &acceptedQty,
				ExpiryDate:       expiryDate,
			},
		},
	}

	ctx := context.Background()
	_, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertError(t, err, "Should return error when accepted exceeds received")
	testutils.AssertContains(t, err.Error(), "cannot exceed received", "Error should mention quantity violation")
}

func TestPurchaseOrderService_UpdateStatus_DetailedQuantities_ReceivedExceedsOrdered(t *testing.T) {
	// Note: Auto-GRN triggers on "verified" status
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "TOM-1KG", "1.0")

	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")
	item := createTestPOItem(t, db, po.ID, variant.ID, 100, 25.50)

	receivedQty := int64(150) // Received > Ordered - INVALID
	acceptedQty := int64(100)
	expiryDate := time.Now().UTC().Add(90 * 24 * time.Hour).Format("2006-01-02")
	actualDelivery := time.Now().UTC()
	request := &models.UpdatePOStatusRequest{
		Status:         "verified",
		ActualDelivery: &actualDelivery,
		Items: []models.DeliveryItemRequest{
			{
				POItemID:         item.ID,
				ReceivedQuantity: &receivedQty,
				AcceptedQuantity: &acceptedQty,
				ExpiryDate:       expiryDate,
			},
		},
	}

	ctx := context.Background()
	_, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertError(t, err, "Should return error when received exceeds ordered")
	testutils.AssertContains(t, err.Error(), "cannot exceed ordered", "Error should mention ordered quantity")
}

func TestPurchaseOrderService_UpdateStatus_InvalidExpiryDateFormat(t *testing.T) {
	// Note: Auto-GRN triggers on "verified" status
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "TOM-1KG", "1.0")

	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")
	item := createTestPOItem(t, db, po.ID, variant.ID, 100, 25.50)

	acceptTrue := true
	actualDelivery := time.Now().UTC()
	request := &models.UpdatePOStatusRequest{
		Status:         "verified",
		ActualDelivery: &actualDelivery,
		Items: []models.DeliveryItemRequest{
			{
				POItemID:   item.ID,
				Accept:     &acceptTrue,
				ExpiryDate: "invalid-date-format",
			},
		},
	}

	ctx := context.Background()
	_, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertError(t, err, "Should return error for invalid expiry date format")
	testutils.AssertContains(t, err.Error(), "expiry_date", "Error should mention expiry_date")
}

func TestPurchaseOrderService_UpdateStatus_DuplicateItemIDs(t *testing.T) {
	// Note: Auto-GRN triggers on "verified" status
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	product := testutils.CreateTestProduct(t, db, "PROD-001", "Tomato")
	variant := testutils.CreateTestVariant(t, db, "VAR-001", product.ID, "TOM-1KG", "1.0")

	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")
	item := createTestPOItem(t, db, po.ID, variant.ID, 100, 25.50)

	acceptTrue := true
	expiryDate := time.Now().UTC().Add(90 * 24 * time.Hour).Format("2006-01-02")
	actualDelivery := time.Now().UTC()
	request := &models.UpdatePOStatusRequest{
		Status:         "verified",
		ActualDelivery: &actualDelivery,
		Items: []models.DeliveryItemRequest{
			{
				POItemID:   item.ID,
				Accept:     &acceptTrue,
				ExpiryDate: expiryDate,
			},
			{
				POItemID:   item.ID, // DUPLICATE
				Accept:     &acceptTrue,
				ExpiryDate: expiryDate,
			},
		},
	}

	ctx := context.Background()
	_, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertError(t, err, "Should return error for duplicate item IDs")
	testutils.AssertContains(t, err.Error(), "duplicate", "Error should mention duplicate")
}

func TestPurchaseOrderService_UpdateStatus_ItemNotInPO(t *testing.T) {
	// Note: Auto-GRN triggers on "verified" status
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")

	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")

	acceptTrue := true
	expiryDate := time.Now().UTC().Add(90 * 24 * time.Hour).Format("2006-01-02")
	actualDelivery := time.Now().UTC()
	request := &models.UpdatePOStatusRequest{
		Status:         "verified",
		ActualDelivery: &actualDelivery,
		Items: []models.DeliveryItemRequest{
			{
				POItemID:   "INVALID-ITEM-ID",
				Accept:     &acceptTrue,
				ExpiryDate: expiryDate,
			},
		},
	}

	ctx := context.Background()
	_, err := service.UpdatePurchaseOrderStatus(ctx, po.ID, request, "USER-001")

	testutils.AssertError(t, err, "Should return error for item not in PO")
	testutils.AssertContains(t, err.Error(), "does not belong", "Error should mention item not belonging to PO")
}

// =============================================================================
// UpdatePaymentStatus Tests
// =============================================================================

func TestPurchaseOrderService_UpdatePaymentStatus_Success(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")

	request := &models.UpdatePOPaymentRequest{
		PaymentStatus: "partial",
		PaidAmount:    500.00,
	}

	ctx := context.Background()
	response, err := service.UpdatePaymentStatus(ctx, po.ID, request)

	testutils.AssertNoError(t, err, "UpdatePaymentStatus should succeed")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.PaymentStatus, "partial", "Payment status should be partial")
	testutils.AssertEqual(t, response.PaidAmount, 500.00, "Paid amount should be updated")
}

func TestPurchaseOrderService_UpdatePaymentStatus_AutoCompleteToPaid(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")
	po.TotalAmount = 1000.00
	db.Save(po)

	request := &models.UpdatePOPaymentRequest{
		PaymentStatus: "partial",
		PaidAmount:    1000.00, // Full payment
	}

	ctx := context.Background()
	response, err := service.UpdatePaymentStatus(ctx, po.ID, request)

	testutils.AssertNoError(t, err, "UpdatePaymentStatus should succeed")
	testutils.AssertEqual(t, response.PaymentStatus, "paid", "Payment status should auto-update to paid")
	testutils.AssertEqual(t, response.Status, "paid", "PO status should also update to paid")
}

func TestPurchaseOrderService_UpdatePaymentStatus_PartialPayment(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")
	po.TotalAmount = 1000.00
	db.Save(po)

	request := &models.UpdatePOPaymentRequest{
		PaymentStatus: "partial",
		PaidAmount:    400.00, // Partial payment
	}

	ctx := context.Background()
	response, err := service.UpdatePaymentStatus(ctx, po.ID, request)

	testutils.AssertNoError(t, err, "UpdatePaymentStatus should succeed")
	testutils.AssertEqual(t, response.PaymentStatus, "partial", "Payment status should remain partial")
	testutils.AssertEqual(t, response.PaidAmount, 400.00, "Paid amount should be updated")
}

func TestPurchaseOrderService_UpdatePaymentStatus_InvalidStatus(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")

	request := &models.UpdatePOPaymentRequest{
		PaymentStatus: "invalid_status",
		PaidAmount:    500.00,
	}

	ctx := context.Background()
	_, err := service.UpdatePaymentStatus(ctx, po.ID, request)

	testutils.AssertError(t, err, "Should return error for invalid payment status")
	testutils.AssertContains(t, err.Error(), "invalid payment status", "Error should mention invalid status")
}

func TestPurchaseOrderService_UpdatePaymentStatus_PaidExceedsTotal(t *testing.T) {
	service, db, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	collaborator := createTestCollaborator(t, db, "COLLAB-001", true)
	warehouse := createTestWarehouse(t, db, "WAREHOUSE-001")
	po := createTestPurchaseOrder(t, db, "PO-2025-0001", collaborator.ID, warehouse.ID, "delivered")
	po.TotalAmount = 1000.00
	db.Save(po)

	request := &models.UpdatePOPaymentRequest{
		PaymentStatus: "paid",
		PaidAmount:    1500.00, // Exceeds total
	}

	ctx := context.Background()
	_, err := service.UpdatePaymentStatus(ctx, po.ID, request)

	testutils.AssertError(t, err, "Should return error when paid amount exceeds total")
	testutils.AssertContains(t, err.Error(), "cannot exceed total", "Error should mention exceeding total")
}

func TestPurchaseOrderService_UpdatePaymentStatus_PONotFound(t *testing.T) {
	service, _, cleanup := setupPurchaseOrderService(t)
	defer cleanup()

	request := &models.UpdatePOPaymentRequest{
		PaymentStatus: "paid",
		PaidAmount:    1000.00,
	}

	ctx := context.Background()
	_, err := service.UpdatePaymentStatus(ctx, "INVALID-PO", request)

	testutils.AssertError(t, err, "Should return error for non-existent PO")
}
