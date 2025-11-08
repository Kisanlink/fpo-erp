# Test Utilities Reference

**Complete reference for all test utility functions in the `tests/testutils` package**

---

## Overview

The `testutils` package provides:
- Database setup and teardown
- Test data fixtures (456 lines)
- Custom assertion helpers
- Mock infrastructure (AAA client, S3 service)
- Repository mocks for unit testing
- Context and time helpers

---

## Table of Contents

1. [Database Setup](#database-setup)
2. [Fixtures](#fixtures)
3. [Assertion Helpers](#assertion-helpers)
4. [Context Helpers](#context-helpers)
5. [Time Helpers](#time-helpers)
6. [Database Helpers](#database-helpers)
7. [Mock Infrastructure](#mock-infrastructure)
8. [Repository Mocks](#repository-mocks)

---

## Database Setup

### SetupTestDB

**Creates an in-memory SQLite database for testing**

```go
func SetupTestDB(t *testing.T) *gorm.DB
```

**What It Does:**
1. Suppresses log output
2. Registers JSON serializer
3. Creates silent GORM logger
4. Opens `:memory:` SQLite database
5. Creates 27 SQLite-compatible tables
6. Returns database connection

**Usage:**
```go
db := testutils.SetupTestDB(t)
defer testutils.CleanupTestDB(db)
```

### CleanupTestDB

**Closes database and restores log output**

```go
func CleanupTestDB(db *gorm.DB)
```

**Usage:**
```go
defer testutils.CleanupTestDB(db)
```

### RegisterSQLiteJSONSerializer

**Registers custom JSON serializer for []string fields** (called automatically by SetupTestDB)

```go
func RegisterSQLiteJSONSerializer(db *gorm.DB)
```

---

## Fixtures

All fixtures return pointers to models ready to use or save to database.

### Product Fixtures

```go
// Basic product with auto-generated ID
func FixtureProduct(name string) *models.Product

// Product with custom ID
func FixtureProductWithID(id, name string) *models.Product

// Inactive product
func FixtureInactiveProduct(name string) *models.Product

// Product with description
func FixtureProductWithDescription(name, description string) *models.Product
```

**Example:**
```go
product := testutils.FixtureProduct("Tomato")
db.Create(product)
// product.ID is auto-generated like "PROD00000001"
```

### Product Variant Fixtures

```go
// Basic variant
func FixtureProductVariant(productID, sku, size string) *models.ProductVariant

// Variant with custom ID
func FixtureProductVariantWithID(id, productID, sku, size string) *models.ProductVariant

// Inactive variant
func FixtureInactiveProductVariant(productID, sku, size string) *models.ProductVariant
```

**Example:**
```go
variant := testutils.FixtureProductVariant(product.ID, "TOM-500G", "500g")
db.Create(variant)
```

### Product Price Fixtures

```go
// Basic price
func FixtureProductPrice(variantID, priceType string, price float64) *models.ProductPrice

// Price with custom ID
func FixtureProductPriceWithID(id, variantID, priceType string, price float64) *models.ProductPrice

// Price with date range
func FixtureProductPriceWithDates(variantID, priceType string, price float64, effectiveFrom time.Time, effectiveTo *time.Time) *models.ProductPrice

// Expired price
func FixtureExpiredProductPrice(variantID, priceType string, price float64) *models.ProductPrice

// Inactive price
func FixtureInactiveProductPrice(variantID, priceType string, price float64) *models.ProductPrice
```

**Example:**
```go
price := testutils.FixtureProductPrice(variant.ID, "retail", 100.50)
db.Create(price)

// Price with date range
from := time.Now().UTC()
to := time.Now().UTC().Add(30 * 24 * time.Hour)
price := testutils.FixtureProductPriceWithDates(variant.ID, "retail", 100.50, from, &to)
```

### Warehouse Fixtures

```go
// Basic warehouse
func FixtureWarehouse(name string) *models.Warehouse

// Warehouse with custom ID
func FixtureWarehouseWithID(id, name string) *models.Warehouse

// Warehouse with address
func FixtureWarehouseWithAddress(name, addressID string) *models.Warehouse

// Inactive warehouse
func FixtureInactiveWarehouse(name string) *models.Warehouse

// Warehouse with capacity
func FixtureWarehouseWithCapacity(name string, capacity int64) *models.Warehouse
```

**Example:**
```go
warehouse := testutils.FixtureWarehouse("Main Warehouse")
db.Create(warehouse)
```

### Inventory Batch Fixtures

```go
// Basic batch
func FixtureInventoryBatch(warehouseID, variantID string, quantity int64, costPrice float64) *models.InventoryBatch

// Batch with custom ID
func FixtureInventoryBatchWithID(id, warehouseID, variantID string, quantity int64, costPrice float64) *models.InventoryBatch

// Batch expiring in N days
func FixtureInventoryBatchExpiring(warehouse *models.Warehouse, variant *models.ProductVariant, quantity int64, expiringInDays int) *models.InventoryBatch

// Batch with specific dates
func FixtureInventoryBatchWithDates(warehouse *models.Warehouse, variant *models.ProductVariant, quantity int64, receivedDate, expiryDate time.Time) *models.InventoryBatch

// Zero quantity batch
func FixtureInventoryBatchZeroQuantity(warehouse *models.Warehouse, variant *models.ProductVariant) *models.InventoryBatch
```

**Example:**
```go
// Batch expiring in 30 days
batch := testutils.FixtureInventoryBatchExpiring(warehouse, variant, 100, 30)
db.Create(batch)

// Custom dates
received := time.Now().UTC()
expiry := time.Now().UTC().Add(90 * 24 * time.Hour)
batch := testutils.FixtureInventoryBatchWithDates(warehouse, variant, 100, received, expiry)
```

### Collaborator Fixtures

```go
// Basic collaborator
func FixtureCollaborator(companyName string) *models.Collaborator

// Collaborator with custom ID
func FixtureCollaboratorWithID(id, companyName string) *models.Collaborator

// Inactive collaborator
func FixtureInactiveCollaborator(companyName string) *models.Collaborator
```

**Example:**
```go
collaborator := testutils.FixtureCollaborator("ABC Suppliers Ltd")
db.Create(collaborator)
```

### Purchase Order Fixtures

```go
// Basic purchase order
func FixturePurchaseOrder(collaboratorID, warehouseID string) *models.PurchaseOrder

// PO with custom ID
func FixturePurchaseOrderWithID(id, collaboratorID, warehouseID string) *models.PurchaseOrder

// PO with specific status
func FixturePurchaseOrderWithStatus(collaboratorID, warehouseID, status string) *models.PurchaseOrder

// PO with delivery dates
func FixturePurchaseOrderWithDates(collaboratorID, warehouseID string, expected, actual *time.Time) *models.PurchaseOrder

// Completed PO
func FixtureCompletedPurchaseOrder(collaboratorID, warehouseID string) *models.PurchaseOrder
```

**Example:**
```go
po := testutils.FixturePurchaseOrder(collaborator.ID, warehouse.ID)
db.Create(po)

// PO with status
po := testutils.FixturePurchaseOrderWithStatus(collaborator.ID, warehouse.ID, "delivered")
```

### Purchase Order Item Fixtures

```go
// Basic PO item
func FixturePurchaseOrderItem(poID, variantID string, quantity int64, unitPrice float64) *models.PurchaseOrderItem

// PO item with custom ID
func FixturePurchaseOrderItemWithID(id, poID, variantID string, quantity int64, unitPrice float64) *models.PurchaseOrderItem
```

**Example:**
```go
item := testutils.FixturePurchaseOrderItem(po.ID, variant.ID, 100, 50.00)
db.Create(item)
```

### GRN Fixtures

```go
// Basic GRN
func FixtureGRN(poID, warehouseID string) *models.GRN

// GRN with custom ID and number
func FixtureGRNWithID(id, grnNumber, poID, warehouseID string) *models.GRN

// GRN with quality status
func FixtureGRNWithQualityStatus(poID, warehouseID, qualityStatus string) *models.GRN

// Completed GRN
func FixtureCompletedGRN(poID, warehouseID string) *models.GRN
```

**Example:**
```go
grn := testutils.FixtureGRN(po.ID, warehouse.ID)
db.Create(grn)
```

### GRN Item Fixtures

```go
// Basic GRN item
func FixtureGRNItem(grnID, poItemID string, received, accepted int64, expiryDate time.Time) *models.GRNItem

// GRN item with custom ID
func FixtureGRNItemWithID(id, grnID, poItemID string, received, accepted int64, expiryDate time.Time) *models.GRNItem
```

**Example:**
```go
expiry := time.Now().UTC().Add(90 * 24 * time.Hour)
item := testutils.FixtureGRNItem(grn.ID, poItem.ID, 100, 95, expiry)
db.Create(item)
```

### Sale Fixtures

```go
// Basic sale
func FixtureSale(warehouseID string, totalAmount float64) *models.Sale

// Sale with custom ID
func FixtureSaleWithID(id, warehouseID string, totalAmount float64) *models.Sale

// Sale with status
func FixtureSaleWithStatus(warehouseID string, totalAmount float64, status string) *models.Sale

// Completed sale
func FixtureCompletedSale(warehouseID string, totalAmount float64) *models.Sale
```

**Example:**
```go
sale := testutils.FixtureSale(warehouse.ID, 500.00)
db.Create(sale)
```

### Discount Fixtures

```go
// Basic discount
func FixtureDiscount(name, discountType string, value float64) *models.Discount

// Discount with custom ID
func FixtureDiscountWithID(id, name, discountType string, value float64) *models.Discount

// Discount with code
func FixtureDiscountWithCode(code, discountType string, value float64) *models.Discount

// Expired discount
func FixtureExpiredDiscount(name, discountType string, value float64) *models.Discount

// Inactive discount
func FixtureInactiveDiscount(name, discountType string, value float64) *models.Discount

// Discount with constraints
func FixtureDiscountWithConstraints(name, discountType string, value, minOrder, maxOrder float64) *models.Discount

// Percentage discount
func FixturePercentageDiscount(name string, percentage float64) *models.Discount

// Flat discount
func FixtureFlatDiscount(name string, amount float64) *models.Discount
```

**Example:**
```go
discount := testutils.FixturePercentageDiscount("Summer Sale", 20.0)
db.Create(discount)

// Discount with code
discount := testutils.FixtureDiscountWithCode("SAVE20", "flat", 50.0)
```

### Tax Fixtures

```go
// Basic tax
func FixtureTax(name, taxType string, rate float64) *models.Tax

// Tax with custom ID
func FixtureTaxWithID(id, name, taxType string, rate float64) *models.Tax

// GST tax
func FixtureGSTTax(name string, rate float64) *models.Tax

// Inactive tax
func FixtureInactiveTax(name string, rate float64) *models.Tax
```

**Example:**
```go
tax := testutils.FixtureGSTTax("CGST", 9.0)
db.Create(tax)
```

---

## Assertion Helpers

All assertion helpers use `t.Helper()` to report errors at the correct line.

### Error Assertions

```go
// Assert no error occurred
func AssertNoError(t *testing.T, err error, message string)

// Assert error occurred
func AssertError(t *testing.T, err error, message string)
```

**Example:**
```go
result, err := service.CreateProduct(request)
testutils.AssertNoError(t, err, "CreateProduct should succeed")

_, err = service.GetProduct("invalid-id")
testutils.AssertError(t, err, "Should fail with invalid ID")
```

### Equality Assertions

```go
// Assert values are equal
func AssertEqual(t *testing.T, got, want interface{}, message string)

// Assert values are not equal
func AssertNotEqual(t *testing.T, got, notWant interface{}, message string)
```

**Example:**
```go
testutils.AssertEqual(t, result.Name, "Tomato", "Name should match")
testutils.AssertNotEqual(t, result.ID, "", "ID should be generated")
```

### Nil Assertions

```go
// Assert value is nil
func AssertNil(t *testing.T, value interface{}, message string)

// Assert value is not nil
func AssertNotNil(t *testing.T, value interface{}, message string)
```

**Example:**
```go
testutils.AssertNotNil(t, result, "Result should not be nil")
testutils.AssertNil(t, result.DeletedAt, "Should not be deleted")
```

### Boolean Assertions

```go
// Assert condition is true
func AssertTrue(t *testing.T, condition bool, message string)

// Assert condition is false
func AssertFalse(t *testing.T, condition bool, message string)
```

**Example:**
```go
testutils.AssertTrue(t, result.IsActive, "Should be active")
testutils.AssertFalse(t, result.IsDeleted, "Should not be deleted")
```

### Numeric Assertions

```go
// Assert value is greater than threshold
func AssertGreaterThan(t *testing.T, value, threshold int64, message string)

// Assert value is less than threshold
func AssertLessThan(t *testing.T, value, threshold int64, message string)
```

**Example:**
```go
testutils.AssertGreaterThan(t, result.Quantity, 0, "Quantity should be positive")
testutils.AssertLessThan(t, result.Price, 1000.0, "Price too high")
```

### String Assertions

```go
// Assert string contains substring
func AssertContains(t *testing.T, str, substr string, message string)
```

**Example:**
```go
testutils.AssertContains(t, err.Error(), "not found", "Error should mention not found")
```

---

## Context Helpers

### CreateTestContext

**Creates a basic context.Context for testing**

```go
func CreateTestContext() context.Context
```

**Usage:**
```go
ctx := testutils.CreateTestContext()
result, err := service.GetProduct(ctx, "PROD001")
```

### CreateTestContextWithUserID

**Creates context with user ID for audit trail testing**

```go
func CreateTestContextWithUserID(userID string) context.Context
```

**Usage:**
```go
ctx := testutils.CreateTestContextWithUserID("USER_12345")
result, err := service.CreateProduct(ctx, request)
// Service can extract user_id from context for created_by field
```

---

## Time Helpers

### FutureDate

**Returns a date N days in the future**

```go
func FutureDate(days int) time.Time
```

**Example:**
```go
expiry := testutils.FutureDate(30) // 30 days from now
```

### PastDate

**Returns a date N days in the past**

```go
func PastDate(days int) time.Time
```

**Example:**
```go
received := testutils.PastDate(7) // 7 days ago
```

### TodayDate

**Returns today's date at midnight UTC**

```go
func TodayDate() time.Time
```

**Example:**
```go
today := testutils.TodayDate()
```

### RandomString

**Generates a random string of specified length**

```go
func RandomString(length int) string
```

**Example:**
```go
uniqueName := "Product-" + testutils.RandomString(8)
```

---

## Database Helpers

### CreateTestProduct

**Creates and saves a product to database**

```go
func CreateTestProduct(t *testing.T, db *gorm.DB, id, name string) *models.Product
```

**Example:**
```go
product := testutils.CreateTestProduct(t, db, "PROD001", "Tomato")
// Product is already saved to database
```

### CreateTestWarehouse

**Creates and saves a warehouse to database**

```go
func CreateTestWarehouse(t *testing.T, db *gorm.DB, id, name string) *models.Warehouse
```

**Example:**
```go
warehouse := testutils.CreateTestWarehouse(t, db, "WH001", "Main Warehouse")
```

### CreateTestProductVariant

**Creates and saves a product variant to database**

```go
func CreateTestProductVariant(t *testing.T, db *gorm.DB, productID string) *models.ProductVariant
```

**Example:**
```go
variant := testutils.CreateTestProductVariant(t, db, product.ID)
```

---

## Mock Infrastructure

### MockAAAClient

**Mock for external AAA service (address operations)**

```go
type MockAAAClient struct {
    shouldFail      bool
    failureMessage  string
    addresses       map[string]*AAA_AddressResponse
}

func NewMockAAAClient() *MockAAAClient
func (m *MockAAAClient) CreateAddress(ctx context.Context, req *AAA_AddressRequest, jwtToken string) (*AAA_AddressResponse, error)
func (m *MockAAAClient) GetAddress(ctx context.Context, addressID, jwtToken string) (*AAA_AddressResponse, error)
func (m *MockAAAClient) UpdateAddress(ctx context.Context, addressID string, req *AAA_AddressRequest, jwtToken string) (*AAA_AddressResponse, error)
func (m *MockAAAClient) DeleteAddress(ctx context.Context, addressID, jwtToken string) error
func (m *MockAAAClient) SearchAddresses(ctx context.Context, filters map[string]interface{}, jwtToken string) ([]*AAA_AddressResponse, error)
func (m *MockAAAClient) SetShouldFail(shouldFail bool, message string)
```

**Example:**
```go
mockAAA := testutils.NewMockAAAClient()

// Create address
address, err := mockAAA.CreateAddress(ctx, &AAA_AddressRequest{
    Type:   "business",
    Street: "123 Main St",
    City:   "Mumbai",
    State:  "Maharashtra",
}, "jwt-token")
testutils.AssertNoError(t, err, "Should create address")

// Configure failure
mockAAA.SetShouldFail(true, "Network error")
_, err = mockAAA.CreateAddress(ctx, req, "token")
testutils.AssertError(t, err, "Should fail when configured")
```

### MockS3Service

**Mock for AWS S3 file operations**

```go
type MockS3Service struct {
    shouldFail     bool
    failureMessage string
    uploadedFiles  map[string]string
}

func NewMockS3Service() *MockS3Service
func (m *MockS3Service) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, entityType, entityID string) (string, error)
func (m *MockS3Service) GeneratePresignedURL(ctx context.Context, s3URL string, expiration time.Duration) (string, error)
func (m *MockS3Service) DeleteFile(ctx context.Context, s3URL string) error
func (m *MockS3Service) SetShouldFail(shouldFail bool, message string)
```

**Example:**
```go
mockS3 := testutils.NewMockS3Service()

// Upload file
key, err := mockS3.UploadFile(ctx, fileHeader, "grn", "GRN_123")
testutils.AssertNoError(t, err, "Should upload file")

// Generate presigned URL
url, err := mockS3.GeneratePresignedURL(ctx, key, 15*time.Minute)
testutils.AssertNoError(t, err, "Should generate URL")

// Configure failure
mockS3.SetShouldFail(true, "S3 service unavailable")
_, err = mockS3.UploadFile(ctx, fileHeader, "grn", "GRN_123")
testutils.AssertError(t, err, "Should fail when configured")
```

---

## Repository Mocks

**726 lines of testify/mock-based repository mocks for unit testing**

Available mock repositories:
- `MockInventoryRepository` (22 methods)
- `MockPurchaseOrderRepository` (18 methods)
- `MockGRNRepository` (13 methods)
- `MockSalesRepository` (13 methods)
- `MockDiscountsRepository` (12 methods)
- `MockWarehouseRepository` (6 methods)
- `MockProductRepository` (6 methods)
- `MockProductVariantRepository` (5 methods)
- `MockCollaboratorRepository` (5 methods)

### Example: MockInventoryRepository

```go
import "github.com/stretchr/testify/mock"

mockRepo := &testutils.MockInventoryRepository{}

// Setup expectation
mockRepo.On("GetBatchByID", "BATCH-001").Return(batch, nil)
mockRepo.On("CreateBatch", mock.Anything).Return(nil)

// Use in service
service := services.NewInventoryService(mockRepo)
result, err := service.GetBatch("BATCH-001")

// Verify expectations were met
mockRepo.AssertExpectations(t)
```

**Note:** Current tests primarily use **real database integration** instead of mocks.

---

## Usage Patterns

### Simple Test (Single Repository)

```go
func TestProductService_GetProduct(t *testing.T) {
    db := testutils.SetupTestDB(t)
    defer testutils.CleanupTestDB(db)

    repo := repositories.NewProductRepository(db)
    service := services.NewProductService(repo)

    product := testutils.FixtureProduct("Tomato")
    db.Create(product)

    result, err := service.GetProduct(product.ID)

    testutils.AssertNoError(t, err, "GetProduct should succeed")
    testutils.AssertEqual(t, result.ID, product.ID, "ID should match")
}
```

### Complex Test (Multiple Repositories with Setup Helper)

```go
func setupPurchaseOrderService(t *testing.T) (*services.PurchaseOrderService, *gorm.DB, func()) {
    t.Helper()

    db := testutils.SetupTestDB(t)

    poRepo := repositories.NewPurchaseOrderRepository(db)
    collaboratorRepo := repositories.NewCollaboratorRepository(db)
    warehouseRepo := repositories.NewWarehouseRepository(db)
    productRepo := repositories.NewProductRepository(db)
    variantRepo := repositories.NewProductVariantRepository(db)
    grnRepo := repositories.NewGRNRepository(db)
    inventoryRepo := repositories.NewInventoryRepository(db)

    service := services.NewPurchaseOrderService(
        poRepo, collaboratorRepo, warehouseRepo,
        productRepo, variantRepo, grnRepo, inventoryRepo,
    )

    cleanup := func() { testutils.CleanupTestDB(db) }
    return service, db, cleanup
}

func TestPurchaseOrder_Something(t *testing.T) {
    service, db, cleanup := setupPurchaseOrderService(t)
    defer cleanup()

    // Test code...
}
```

---

## Additional Notes

### Fixture vs Database Helper

**Fixture (In-Memory Only):**
```go
product := testutils.FixtureProduct("Tomato")
// product exists in memory only
db.Create(product) // Must manually save
```

**Database Helper (Auto-Saved):**
```go
product := testutils.CreateTestProduct(t, db, "PROD001", "Tomato")
// product is already saved to database
```

### When to Use Mocks vs Real DB

**Use Real Database (Preferred):**
- Integration tests
- Testing business logic with database interactions
- FEFO allocation, auto-GRN, complex queries
- Most service-level tests

**Use Mocks:**
- Pure unit tests (testing logic without DB)
- External service integration (AAA, S3)
- Testing error handling in isolation

---

## Questions?

- See `TESTING_GUIDE.md` for comprehensive testing guide
- Look at `tests/services/price_service_test.go` for comprehensive examples
- Check individual files for specific implementations
