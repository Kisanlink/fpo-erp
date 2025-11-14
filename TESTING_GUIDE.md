# Kisanlink ERP Testing Guide

**Complete guide for writing tests in the Kisanlink ERP system**

Last Updated: January 2025

---

## Table of Contents

1. [Overview](#overview)
2. [Test Infrastructure](#test-infrastructure)
3. [Quick Start](#quick-start)
4. [Writing Tests Step-by-Step](#writing-tests-step-by-step)
5. [Test Utilities](#test-utilities)
6. [Common Patterns](#common-patterns)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)
9. [Examples](#examples)

---

## Overview

The Kisanlink ERP project uses **in-memory SQLite databases** for fast, isolated integration tests. Tests interact with real database instances (not mocks) to ensure production compatibility while maintaining test speed.

### Key Characteristics

- **Real Database Integration**: Tests use actual SQLite databases, not mocks
- **Fixture-Based**: Pre-defined test data factories for quick setup
- **Custom Assertions**: Descriptive helpers instead of external libraries
- **Isolated**: Each test gets a fresh `:memory:` database
- **Fast**: In-memory databases with no I/O overhead

### Test Strategy

```
Unit Tests (service logic) → Integration Tests (real DB) → End-to-End Tests (future)
                              ↑ Current Focus
```

Currently, we focus on **integration tests** that verify service logic against real database behavior.

---

## Test Infrastructure

### Architecture Overview

```
SetupTestDB(t)
    ↓
Register JSON Serializer (SQLite compatibility)
    ↓
Create :memory: Database (GORM + custom naming strategy)
    ↓
Create SQLite-Compatible Tables (27 tables via raw SQL)
    ↓
Return DB Connection
```

### Critical Components

#### 1. Database Setup (`tests/testutils/database.go`)

**Main Function:**
```go
func SetupTestDB(t *testing.T) *gorm.DB
```

**What It Does:**
1. Suppresses log output (clean test output)
2. Registers custom JSON serializer for `[]string` fields
3. Creates silent GORM logger
4. Opens `:memory:` SQLite database
5. Creates 27 SQLite-compatible tables
6. Returns ready-to-use database connection

**Usage:**
```go
db := testutils.SetupTestDB(t)
defer testutils.CleanupTestDB(db)
```

#### 2. Table Creation (`tests/testutils/sqlite_migrations.go`)

**Why Manual Creation?**
- SQLite type compatibility (TEXT vs timestamptz, REAL vs numeric)
- Explicit field inclusion (like `deleted_by`)
- Custom constraints and indexes
- Production schema parity

**Tables Created** (27 total):
- Core: Product, ProductVariant, ProductPrice, Warehouse
- Procurement: Collaborator, CollaboratorProduct, PurchaseOrder, PurchaseOrderItem, GRN, GRNItem
- Inventory: InventoryBatch, InventoryTransaction
- Sales: Sale, SaleItem
- Pricing: Discount, DiscountUsage, Tax
- Support: Attachment, WebhookEvent

#### 3. Custom Naming Strategy

Maps Go struct fields to correct SQL column names:

```go
type SQLiteJSONNamingStrategy struct {
    schema.NamingStrategy
}

// Example: ExpectedDelivery → expected_delivery_date
```

**Why?** Production PostgreSQL uses `expected_delivery_date` while Go struct has `ExpectedDelivery`.

#### 4. JSON Serialization (`tests/testutils/sqlite_json_serializer.go`)

**Problem:** PostgreSQL has native JSON types, SQLite doesn't
**Solution:** Custom serializer converts `[]string` ↔ JSON TEXT

```go
// Registered before database connection
RegisterSQLiteJSONSerializer(nil)
```

---

## Quick Start

### Minimal Test Example

```go
package services

import (
    "testing"
    "kisanlink-erp/tests/testutils"
)

func TestMyService_MyMethod_Success(t *testing.T) {
    // Setup database
    db := testutils.SetupTestDB(t)
    defer testutils.CleanupTestDB(db)

    // Create repository and service
    repo := repositories.NewMyRepository(db)
    service := services.NewMyService(repo)

    // Create test data
    entity := testutils.FixtureEntity("Test Name")
    db.Create(entity)

    // Execute
    result, err := service.MyMethod(entity.ID)

    // Assert
    testutils.AssertNoError(t, err, "MyMethod should succeed")
    testutils.AssertNotNil(t, result, "Result should not be nil")
}
```

### Running Tests

```bash
# Run all tests
go test ./tests/...

# Run specific test
go test ./tests/services -run TestMyService_MyMethod_Success

# Run with verbose output
go test -v ./tests/services

# Run with coverage
go test -cover ./tests/...
```

---

## Writing Tests Step-by-Step

### Step 1: Create Test File

**Naming Convention:**
- File: `{entity}_service_test.go`
- Package: `package services`
- Location: `tests/services/`

**Template:**
```go
package services

import (
    "testing"
    "kisanlink-erp/internal/database/models"
    "kisanlink-erp/internal/database/repositories"
    "kisanlink-erp/internal/services"
    "kisanlink-erp/tests/testutils"
)
```

### Step 2: Create Setup Helper (if needed)

**For Complex Services:**
```go
func setupMyService(t *testing.T) (*services.MyService, *gorm.DB, func()) {
    t.Helper() // Important: marks this as helper for error reporting

    db := testutils.SetupTestDB(t)

    // Create all required repositories
    repo1 := repositories.NewRepo1(db)
    repo2 := repositories.NewRepo2(db)
    repo3 := repositories.NewRepo3(db)

    // Create service with dependencies
    service := services.NewMyService(repo1, repo2, repo3)

    // Return cleanup function
    cleanup := func() {
        testutils.CleanupTestDB(db)
    }

    return service, db, cleanup
}
```

**Usage:**
```go
func TestMyService_Something(t *testing.T) {
    service, db, cleanup := setupMyService(t)
    defer cleanup()
    // ... test code
}
```

### Step 3: Write Test Function

**Naming Convention:**
```
TestServiceName_MethodName_Scenario
```

**Examples:**
- `TestProductService_CreateProduct_Success`
- `TestPriceService_GetCurrentPrice_NotFound`
- `TestInventoryService_AllocateBatches_InsufficientStock`

**Standard Structure:**
```go
func TestMyService_MyMethod_Success(t *testing.T) {
    // 1. ARRANGE: Setup
    service, db, cleanup := setupMyService(t)
    defer cleanup()

    // Create test data
    entity := testutils.FixtureEntity("Test")
    db.Create(entity)

    // 2. ACT: Execute
    result, err := service.MyMethod(entity.ID)

    // 3. ASSERT: Verify
    testutils.AssertNoError(t, err, "Method should succeed")
    testutils.AssertNotNil(t, result, "Result should exist")
    testutils.AssertEqual(t, result.Name, "Test", "Name should match")
}
```

### Step 4: Add Negative Tests

**Test Failures Too:**
```go
func TestMyService_MyMethod_NotFound(t *testing.T) {
    service, _, cleanup := setupMyService(t)
    defer cleanup()

    // Don't create any test data

    _, err := service.MyMethod("non-existent-id")

    testutils.AssertError(t, err, "Should fail when entity not found")
}

func TestMyService_MyMethod_ValidationFails(t *testing.T) {
    service, _, cleanup := setupMyService(t)
    defer cleanup()

    _, err := service.MyMethod("") // Empty ID

    testutils.AssertError(t, err, "Should fail with empty ID")
}
```

### Step 5: Test Edge Cases

**Important Scenarios:**
- Nil values
- Empty strings
- Boundary conditions
- Concurrent operations
- Transaction rollbacks
- Soft deletes vs hard deletes

---

## Test Utilities

### Fixtures (`tests/testutils/fixtures.go`)

**456 lines of pre-built test data factories**

#### Product Fixtures
```go
// Basic product with ID
testutils.FixtureProduct("Tomato")

// Product with custom ID
testutils.FixtureProductWithID("PROD001", "Tomato")

// Inactive product
testutils.FixtureInactiveProduct("Tomato")
```

#### Warehouse Fixtures
```go
testutils.FixtureWarehouse("Main Warehouse")
testutils.FixtureWarehouseWithID("WH001", "Main Warehouse")
```

#### Inventory Fixtures
```go
// Batch expiring in N days
testutils.FixtureInventoryBatchExpiring(warehouse, variant, quantity, 30)

// Batch with custom dates
testutils.FixtureInventoryBatchWithDates(warehouse, variant, quantity, expiryDate)

// Zero quantity batch
testutils.FixtureInventoryBatchZeroQuantity(warehouse, variant)
```

#### Price Fixtures
```go
testutils.FixtureProductPrice(variantID, "retail", 100.50)
testutils.FixtureProductPriceWithDates(variantID, "retail", 100.50, from, to)
testutils.FixtureExpiredProductPrice(variantID, "retail", 90.00)
```

#### Purchase Order Fixtures
```go
testutils.FixturePurchaseOrder(collaboratorID, warehouseID)
testutils.FixturePurchaseOrderWithStatus(collaboratorID, warehouseID, "delivered")
testutils.FixturePurchaseOrderItem(poID, variantID, 100, 50.00)
```

#### Collaborator Fixtures
```go
testutils.FixtureCollaborator("Vendor ABC")
testutils.FixtureCollaboratorWithID("CLAB001", "Vendor ABC")
```

#### Discount Fixtures
```go
testutils.FixtureDiscount("Summer Sale", "percentage", 20.0)
testutils.FixtureDiscountWithCode("SAVE20", "flat", 50.0)
testutils.FixtureExpiredDiscount("Old Sale", "percentage", 15.0)
```

**Pattern:** All fixtures return pointers to models, ready to use or save to DB

### Assertion Helpers (`tests/testutils/helpers.go`)

**Custom assertions with descriptive messages**

#### Error Assertions
```go
testutils.AssertNoError(t, err, "CreateProduct should succeed")
testutils.AssertError(t, err, "Should fail when product not found")
```

#### Equality
```go
testutils.AssertEqual(t, got, want, "Product name mismatch")
testutils.AssertNotEqual(t, got, notWant, "IDs should be different")
```

#### Nil Checks
```go
testutils.AssertNil(t, value, "Value should be nil")
testutils.AssertNotNil(t, response, "Response should not be nil")
```

#### Boolean
```go
testutils.AssertTrue(t, condition, "Should be active")
testutils.AssertFalse(t, condition, "Should be inactive")
```

#### Numeric Comparisons
```go
testutils.AssertGreaterThan(t, actual, expected, "Quantity too low")
testutils.AssertLessThan(t, actual, maximum, "Exceeded limit")
```

#### String Operations
```go
testutils.AssertContains(t, str, substr, "Error should mention 'duplicate'")
```

**Why Custom Assertions?**
- Uses `t.Helper()` for correct line numbers in failures
- Consistent error messages across all tests
- No external dependencies (like testify)
- Better control over failure output

### Context Helpers

```go
// Basic context
ctx := testutils.CreateTestContext()

// Context with user ID (for audit trails)
ctx := testutils.CreateTestContextWithUserID("USER_12345")
```

### Time Helpers

```go
testutils.FutureDate(30)        // 30 days from now
testutils.PastDate(7)           // 7 days ago
testutils.TodayDate()           // Today at midnight
testutils.RandomString(10)      // Random 10-character string
```

### Database Helpers

```go
// Create test entities directly
func CreateTestProduct(t *testing.T, db *gorm.DB, id, name string) *models.Product

func CreateTestWarehouse(t *testing.T, db *gorm.DB, id, name string) *models.Warehouse

func CreateTestProductVariant(t *testing.T, db *gorm.DB, productID string) *models.ProductVariant
```

---

## Common Patterns

### Testing CRUD Operations

#### Create
```go
func TestCreate_Success(t *testing.T) {
    service, db, cleanup := setupService(t)
    defer cleanup()

    request := &CreateRequest{Name: "Test", Description: "Desc"}

    result, err := service.Create(request)

    testutils.AssertNoError(t, err, "Create should succeed")
    testutils.AssertNotNil(t, result, "Result should exist")
    testutils.AssertEqual(t, result.Name, "Test", "Name should match")

    // Verify in database
    var entity models.Entity
    db.First(&entity, "id = ?", result.ID)
    testutils.AssertEqual(t, entity.Name, "Test", "Should be saved")
}

func TestCreate_DuplicateKey(t *testing.T) {
    service, db, cleanup := setupService(t)
    defer cleanup()

    // Create first entity
    existing := testutils.FixtureEntity("Test")
    db.Create(existing)

    // Try to create duplicate
    request := &CreateRequest{Name: "Test"}
    _, err := service.Create(request)

    testutils.AssertError(t, err, "Should fail on duplicate")
}
```

#### Read
```go
func TestGet_Success(t *testing.T) {
    service, db, cleanup := setupService(t)
    defer cleanup()

    entity := testutils.FixtureEntity("Test")
    db.Create(entity)

    result, err := service.Get(entity.ID)

    testutils.AssertNoError(t, err, "Get should succeed")
    testutils.AssertEqual(t, result.ID, entity.ID, "ID should match")
}

func TestGet_NotFound(t *testing.T) {
    service, _, cleanup := setupService(t)
    defer cleanup()

    _, err := service.Get("non-existent-id")

    testutils.AssertError(t, err, "Should fail when not found")
}

func TestGetAll_Empty(t *testing.T) {
    service, _, cleanup := setupService(t)
    defer cleanup()

    results, err := service.GetAll()

    testutils.AssertNoError(t, err, "GetAll should succeed")
    testutils.AssertEqual(t, len(results), 0, "Should be empty")
}

func TestGetAll_Multiple(t *testing.T) {
    service, db, cleanup := setupService(t)
    defer cleanup()

    db.Create(testutils.FixtureEntity("One"))
    db.Create(testutils.FixtureEntity("Two"))
    db.Create(testutils.FixtureEntity("Three"))

    results, err := service.GetAll()

    testutils.AssertNoError(t, err, "GetAll should succeed")
    testutils.AssertEqual(t, len(results), 3, "Should have 3 entities")
}
```

#### Update
```go
func TestUpdate_Success(t *testing.T) {
    service, db, cleanup := setupService(t)
    defer cleanup()

    entity := testutils.FixtureEntity("Old Name")
    db.Create(entity)

    request := &UpdateRequest{Name: "New Name"}
    result, err := service.Update(entity.ID, request)

    testutils.AssertNoError(t, err, "Update should succeed")
    testutils.AssertEqual(t, result.Name, "New Name", "Name should be updated")
}

func TestUpdate_NotFound(t *testing.T) {
    service, _, cleanup := setupService(t)
    defer cleanup()

    request := &UpdateRequest{Name: "New"}
    _, err := service.Update("non-existent", request)

    testutils.AssertError(t, err, "Should fail when not found")
}
```

#### Delete
```go
func TestDelete_Success(t *testing.T) {
    service, db, cleanup := setupService(t)
    defer cleanup()

    entity := testutils.FixtureEntity("Test")
    db.Create(entity)

    err := service.Delete(entity.ID)

    testutils.AssertNoError(t, err, "Delete should succeed")

    // Verify deletion (soft delete check)
    var deleted models.Entity
    result := db.Unscoped().First(&deleted, "id = ?", entity.ID)
    testutils.AssertNoError(t, result.Error, "Should find soft-deleted record")
    testutils.AssertNotNil(t, deleted.DeletedAt, "Should be soft-deleted")
}
```

### Testing Business Logic

#### FEFO Allocation
```go
func TestAllocateBatches_FEFO(t *testing.T) {
    service, db, cleanup := setupInventoryService(t)
    defer cleanup()

    warehouse := testutils.FixtureWarehouse("WH1")
    variant := testutils.FixtureProductVariant("VAR1")
    db.Create(warehouse)
    db.Create(variant)

    // Create batches with different expiry dates
    batch1 := testutils.FixtureInventoryBatchExpiring(warehouse, variant, 50, 10)  // Expires in 10 days
    batch2 := testutils.FixtureInventoryBatchExpiring(warehouse, variant, 100, 30) // Expires in 30 days
    db.Create(batch1)
    db.Create(batch2)

    // Order 75 units - should take 50 from batch1 (oldest), 25 from batch2
    allocations, err := service.AllocateBatches(variant.ID, 75)

    testutils.AssertNoError(t, err, "Allocation should succeed")
    testutils.AssertEqual(t, len(allocations), 2, "Should allocate from 2 batches")
    testutils.AssertEqual(t, allocations[0].BatchID, batch1.ID, "First should be oldest batch")
    testutils.AssertEqual(t, allocations[0].Quantity, int64(50), "Should take all from first batch")
    testutils.AssertEqual(t, allocations[1].BatchID, batch2.ID, "Second should be newer batch")
    testutils.AssertEqual(t, allocations[1].Quantity, int64(25), "Should take remainder")
}

func TestAllocateBatches_InsufficientStock(t *testing.T) {
    service, db, cleanup := setupInventoryService(t)
    defer cleanup()

    warehouse := testutils.FixtureWarehouse("WH1")
    variant := testutils.FixtureProductVariant("VAR1")
    db.Create(warehouse)
    db.Create(variant)

    // Only 50 units available
    batch := testutils.FixtureInventoryBatchExpiring(warehouse, variant, 50, 30)
    db.Create(batch)

    // Try to order 100 units
    _, err := service.AllocateBatches(variant.ID, 100)

    testutils.AssertError(t, err, "Should fail when insufficient stock")
}
```

#### Date/Time Logic
```go
func TestGetCurrentPrice_WithinDateRange(t *testing.T) {
    service, db, cleanup := setupPriceService(t)
    defer cleanup()

    variant := testutils.FixtureProductVariant("VAR1")
    db.Create(variant)

    // Price effective for 7 days before and after today
    effectiveFrom := time.Now().UTC().Add(-7 * 24 * time.Hour)
    effectiveTo := time.Now().UTC().Add(7 * 24 * time.Hour)
    price := testutils.FixtureProductPriceWithDates(variant.ID, "retail", 100.50, effectiveFrom, &effectiveTo)
    db.Create(price)

    result, err := service.GetCurrentPrice(variant.ID, "retail")

    testutils.AssertNoError(t, err, "Should find current price")
    testutils.AssertEqual(t, result.ID, price.ID, "Should return correct price")
}

func TestGetCurrentPrice_Expired(t *testing.T) {
    service, db, cleanup := setupPriceService(t)
    defer cleanup()

    variant := testutils.FixtureProductVariant("VAR1")
    db.Create(variant)

    // Expired price
    price := testutils.FixtureExpiredProductPrice(variant.ID, "retail", 90.00)
    db.Create(price)

    _, err := service.GetCurrentPrice(variant.ID, "retail")

    testutils.AssertError(t, err, "Should not find expired price")
}
```

#### Discount Priority Logic
```go
func TestApplyDiscount_ManualByIDTakesPriority(t *testing.T) {
    service, db, cleanup := setupSalesService(t)
    defer cleanup()

    // Create auto-applied discount (20%)
    autoDiscount := testutils.FixtureDiscount("Auto 20%", "percentage", 20.0)
    db.Create(autoDiscount)

    // Create manual discount (10%)
    manualDiscount := testutils.FixtureDiscount("Manual 10%", "percentage", 10.0)
    db.Create(manualDiscount)

    // Apply manual discount by ID (should ignore auto discount)
    request := &CreateSaleRequest{
        DiscountID: &manualDiscount.ID,
        Items:      []SaleItemRequest{{VariantID: "VAR1", Quantity: 1, Price: 100}},
    }

    result, err := service.CreateSale(request)

    testutils.AssertNoError(t, err, "Sale should succeed")
    testutils.AssertEqual(t, *result.DiscountID, manualDiscount.ID, "Should use manual discount")
    testutils.AssertEqual(t, result.DiscountAmount, 10.0, "Should apply 10% discount")
}
```

---

## Best Practices

### DO ✅

1. **Use SetupTestDB for Database**
   ```go
   db := testutils.SetupTestDB(t)
   defer testutils.CleanupTestDB(db)
   ```

2. **Use Fixtures for Test Data**
   ```go
   product := testutils.FixtureProduct("Test Product")
   db.Create(product)
   ```

3. **Use Assertion Helpers**
   ```go
   testutils.AssertNoError(t, err, "CreateProduct should succeed")
   testutils.AssertEqual(t, result.Name, "Test", "Name should match")
   ```

4. **Write Descriptive Test Names**
   ```go
   func TestProductService_CreateProduct_Success(t *testing.T)
   func TestPriceService_GetCurrentPrice_NotFound(t *testing.T)
   ```

5. **Add Descriptive Assertion Messages**
   ```go
   testutils.AssertNoError(t, err, "Allocation should succeed with sufficient stock")
   ```

6. **Test Both Success and Failure Cases**
   ```go
   func TestMethod_Success(t *testing.T) { /* ... */ }
   func TestMethod_NotFound(t *testing.T) { /* ... */ }
   func TestMethod_ValidationFails(t *testing.T) { /* ... */ }
   ```

7. **Use t.Helper() in Helper Functions**
   ```go
   func setupService(t *testing.T) (*Service, func()) {
       t.Helper() // Makes error messages point to actual test, not helper
       // ...
   }
   ```

8. **Clean Up with defer**
   ```go
   service, db, cleanup := setupService(t)
   defer cleanup()
   ```

9. **Create Only Test Data You Need**
   ```go
   // Good: Only create what's needed for this test
   product := testutils.FixtureProduct("Test")
   db.Create(product)

   // Bad: Creating unused data
   for i := 0; i < 100; i++ {
       db.Create(testutils.FixtureProduct(fmt.Sprintf("Product %d", i)))
   }
   ```

10. **Verify Database State After Operations**
    ```go
    err := service.Delete(id)
    testutils.AssertNoError(t, err, "Delete should succeed")

    // Verify it's actually deleted
    var entity models.Entity
    result := db.First(&entity, "id = ?", id)
    testutils.AssertError(t, result.Error, "Should not find deleted entity")
    ```

### DON'T ❌

1. **Don't Use AutoMigrate in Tests**
   ```go
   // ❌ WRONG
   db.AutoMigrate(&models.Product{}, &models.Warehouse{})

   // ✅ CORRECT - Tables are already created by SetupTestDB
   db := testutils.SetupTestDB(t) // Creates all tables
   ```

2. **Don't Configure Connection Pooling**
   ```go
   // ❌ WRONG - Will break :memory: databases
   sqlDB, _ := db.DB()
   sqlDB.SetMaxIdleConns(0)
   sqlDB.SetMaxOpenConns(1)

   // ✅ CORRECT - Let GORM handle it automatically
   db := testutils.SetupTestDB(t)
   ```

3. **Don't Use sqlDB.Exec() for Table Creation**
   ```go
   // ❌ WRONG - Bypasses GORM transaction management
   sqlDB, _ := db.DB()
   sqlDB.Exec("CREATE TABLE...")

   // ✅ CORRECT - Stay in GORM context
   db.Exec("CREATE TABLE...").Error
   ```

4. **Don't Use testify/assert**
   ```go
   // ❌ WRONG
   assert.NoError(t, err)
   assert.Equal(t, got, want)

   // ✅ CORRECT - Use testutils helpers
   testutils.AssertNoError(t, err, "Description")
   testutils.AssertEqual(t, got, want, "Description")
   ```

5. **Don't Share Test Data Between Tests**
   ```go
   // ❌ WRONG - Tests will interfere with each other
   var sharedProduct *models.Product

   func TestOne(t *testing.T) {
       sharedProduct = testutils.FixtureProduct("Shared")
       db.Create(sharedProduct)
   }

   func TestTwo(t *testing.T) {
       // Uses sharedProduct - test order dependency!
   }

   // ✅ CORRECT - Each test creates its own data
   func TestOne(t *testing.T) {
       db := testutils.SetupTestDB(t)
       defer testutils.CleanupTestDB(db)
       product := testutils.FixtureProduct("Test")
       db.Create(product)
   }
   ```

6. **Don't Skip Error Messages in Assertions**
   ```go
   // ❌ WRONG - Unclear what failed
   testutils.AssertNoError(t, err, "")

   // ✅ CORRECT - Descriptive message
   testutils.AssertNoError(t, err, "CreateProduct should succeed with valid input")
   ```

7. **Don't Use ILIKE Operator**
   ```go
   // ❌ WRONG - SQLite doesn't support ILIKE
   db.Where("name ILIKE ?", "%tomato%")

   // ✅ CORRECT - Use LIKE or skip SQLite incompatible tests
   if err != nil {
       t.Skip("Skipping test due to SQLite incompatibility with ILIKE")
   }
   ```

8. **Don't Assume Auto-Increment IDs**
   ```go
   // ❌ WRONG - Uses kisanlink-db hash IDs, not integers
   product := testutils.FixtureProduct("Test")
   testutils.AssertEqual(t, product.ID, 1, "ID should be 1")

   // ✅ CORRECT - IDs are strings like "PROD00000001"
   testutils.AssertNotEqual(t, product.ID, "", "ID should be generated")
   ```

### WATCH OUT FOR ⚠️

1. **SQLite vs PostgreSQL Differences**
   - ILIKE operator (PostgreSQL-specific)
   - NUMERIC precision (PostgreSQL exact, SQLite floating-point)
   - timestamptz (PostgreSQL) vs DATETIME (SQLite)
   - JSON types (PostgreSQL native, SQLite needs serializer)

2. **Custom Naming Strategy**
   - `ExpectedDelivery` maps to `expected_delivery_date`
   - `ActualDelivery` maps to `actual_delivery_date`
   - Check `SQLiteJSONNamingStrategy` for other mappings

3. **JSON Fields**
   - `[]string` fields need custom serialization
   - Serializer registered automatically by SetupTestDB
   - Don't manually handle JSON serialization

4. **Soft Deletes**
   - GORM uses `deleted_at` for soft deletes
   - Use `db.Unscoped()` to query soft-deleted records
   - Check `deleted_at IS NOT NULL` to verify soft delete

5. **Time Zones**
   - Always use UTC: `time.Now().UTC()`
   - SQLite stores as DATETIME (no timezone)
   - Tests may fail if using local time

6. **Transaction Boundaries**
   - FEFO allocation spans multiple batches
   - Auto-GRN creates batches + transactions atomically
   - Be aware of transaction rollback behavior

---

## Troubleshooting

### Common Issues

#### 1. "no such table: main.{table_name}"

**Cause:** Table creation failed or using wrong database connection

**Solution:**
```go
// Ensure SetupTestDB is called
db := testutils.SetupTestDB(t)
defer testutils.CleanupTestDB(db)

// Don't get raw sqlDB
// ❌ sqlDB, _ := db.DB()
// ✅ Just use db
```

#### 2. "Failed to create inventory batch: " (empty error)

**Cause:** SQLite driver returns empty error messages

**Solution:** Add debug logging to see actual SQL errors:
```go
db.Exec("...").Error // Will show actual error
```

#### 3. Tests Hang or Timeout

**Cause:** Likely infinite loop or deadlock in service logic

**Solution:**
```bash
# Run with timeout
go test -timeout 10s ./tests/services
```

#### 4. "connection pool exhausted"

**Cause:** Manually configured connection pool

**Solution:** Remove connection pool configuration:
```go
// ❌ WRONG
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(0)

// ✅ CORRECT - Let GORM handle it
db := testutils.SetupTestDB(t)
```

#### 5. "record not found" in Tests

**Cause:** Test data not created or using wrong ID

**Solution:**
```go
// Verify data was created
entity := testutils.FixtureEntity("Test")
result := db.Create(entity)
testutils.AssertNoError(t, result.Error, "Should create test data")

// Check ID is set
testutils.AssertNotEqual(t, entity.ID, "", "ID should be generated")
```

#### 6. SQLite Compatibility Errors (ILIKE, etc.)

**Cause:** Using PostgreSQL-specific features

**Solution:**
```go
responses, err := service.SearchProducts("Tomato")
if err != nil {
    t.Logf("NOTE: SearchProducts failed (likely SQLite incompatibility with ILIKE): %v", err)
    t.Skip("Skipping test due to SQLite incompatibility")
    return
}
```

#### 7. Fixture Returns Nil

**Cause:** Fixture might have validation errors

**Solution:**
```go
entity := testutils.FixtureEntity("Test")
testutils.AssertNotNil(t, entity, "Fixture should not return nil")

// Create and check for errors
result := db.Create(entity)
testutils.AssertNoError(t, result.Error, "Should save fixture to database")
```

---

## Examples

### Example 1: Product Service Test

**File:** `tests/services/product_service_test.go`

```go
package services

import (
    "testing"
    "kisanlink-erp/internal/database/models"
    "kisanlink-erp/internal/database/repositories"
    "kisanlink-erp/internal/services"
    "kisanlink-erp/tests/testutils"
)

// Setup helper for product service
func setupProductService(t *testing.T) (*services.ProductService, *gorm.DB, func()) {
    t.Helper()

    db := testutils.SetupTestDB(t)
    repo := repositories.NewProductRepository(db)
    service := services.NewProductService(repo)

    cleanup := func() {
        testutils.CleanupTestDB(db)
    }

    return service, db, cleanup
}

// Success case
func TestProductService_CreateProduct_Success(t *testing.T) {
    service, _, cleanup := setupProductService(t)
    defer cleanup()

    desc := "Fresh tomatoes"
    request := &models.CreateProductRequest{
        Name:        "Tomato",
        Description: &desc,
    }

    result, err := service.CreateProduct(request)

    testutils.AssertNoError(t, err, "CreateProduct should succeed")
    testutils.AssertNotNil(t, result, "Result should not be nil")
    testutils.AssertEqual(t, result.Name, "Tomato", "Name should match")
    testutils.AssertNotEqual(t, result.ID, "", "ID should be generated")
}

// Validation failure case
func TestProductService_CreateProduct_EmptyName(t *testing.T) {
    service, _, cleanup := setupProductService(t)
    defer cleanup()

    request := &models.CreateProductRequest{
        Name: "", // Invalid: empty name
    }

    _, err := service.CreateProduct(request)

    testutils.AssertError(t, err, "Should fail with empty name")
}

// Not found case
func TestProductService_GetProduct_NotFound(t *testing.T) {
    service, _, cleanup := setupProductService(t)
    defer cleanup()

    _, err := service.GetProduct("non-existent-id")

    testutils.AssertError(t, err, "Should fail when product not found")
}
```

### Example 2: Inventory FEFO Test

```go
func TestInventoryService_AllocateBatches_FEFO_MultipleWarehouse(t *testing.T) {
    service, db, cleanup := setupInventoryService(t)
    defer cleanup()

    // Create test data
    wh1 := testutils.FixtureWarehouse("WH1")
    wh2 := testutils.FixtureWarehouse("WH2")
    variant := testutils.FixtureProductVariant("VAR1")
    db.Create(wh1)
    db.Create(wh2)
    db.Create(variant)

    // Create batches in different warehouses with different expiry
    batch1 := testutils.FixtureInventoryBatchExpiring(wh1, variant, 100, 10) // WH1, 10 days
    batch2 := testutils.FixtureInventoryBatchExpiring(wh2, variant, 50, 5)   // WH2, 5 days (oldest)
    batch3 := testutils.FixtureInventoryBatchExpiring(wh1, variant, 75, 20)  // WH1, 20 days
    db.Create(batch1)
    db.Create(batch2)
    db.Create(batch3)

    // Allocate 120 units from WH1 only
    allocations, err := service.AllocateBatchesFromWarehouse(variant.ID, wh1.ID, 120)

    testutils.AssertNoError(t, err, "Allocation should succeed")
    testutils.AssertEqual(t, len(allocations), 2, "Should use 2 batches from WH1")
    testutils.AssertEqual(t, allocations[0].BatchID, batch1.ID, "Should use batch1 first (earliest expiry in WH1)")
    testutils.AssertEqual(t, allocations[0].Quantity, int64(100), "Should take all from batch1")
    testutils.AssertEqual(t, allocations[1].BatchID, batch3.ID, "Should use batch3 second")
    testutils.AssertEqual(t, allocations[1].Quantity, int64(20), "Should take 20 from batch3")
}
```

### Example 3: Price Date Range Test

```go
func TestPriceService_GetCurrentPrice_DateRange(t *testing.T) {
    service, db, cleanup := setupPriceService(t)
    defer cleanup()

    variant := testutils.FixtureProductVariant("VAR1")
    db.Create(variant)

    now := time.Now().UTC()

    // Old price (expired)
    oldFrom := now.Add(-30 * 24 * time.Hour)
    oldTo := now.Add(-7 * 24 * time.Hour)
    oldPrice := testutils.FixtureProductPriceWithDates(variant.ID, "retail", 80.00, oldFrom, &oldTo)
    db.Create(oldPrice)

    // Current price
    currentFrom := now.Add(-7 * 24 * time.Hour)
    currentTo := now.Add(7 * 24 * time.Hour)
    currentPrice := testutils.FixtureProductPriceWithDates(variant.ID, "retail", 100.00, currentFrom, &currentTo)
    db.Create(currentPrice)

    // Future price
    futureFrom := now.Add(7 * 24 * time.Hour)
    futurePrice := testutils.FixtureProductPriceWithDates(variant.ID, "retail", 120.00, futureFrom, nil)
    db.Create(futurePrice)

    // Should return current price only
    result, err := service.GetCurrentPrice(variant.ID, "retail")

    testutils.AssertNoError(t, err, "Should find current price")
    testutils.AssertEqual(t, result.ID, currentPrice.ID, "Should return current price, not old or future")
    testutils.AssertEqual(t, result.Price, 100.00, "Price should be 100.00")
}
```

---

## Additional Resources

- **Fixture Reference**: See `tests/testutils/fixtures.go` for all available fixtures
- **Assertion Reference**: See `tests/testutils/helpers.go` for all assertion helpers
- **Mock Infrastructure**: See `tests/testutils/mocks.go` for AAA and S3 mocks
- **Example Tests**: Look at `price_service_test.go` as the gold standard (707 lines, comprehensive)

---

## Contributing

When adding new tests:

1. Follow existing patterns (look at `price_service_test.go`)
2. Use fixtures instead of manual creation
3. Add both success and failure tests
4. Write descriptive test names and assertion messages
5. Create setup helpers for complex services
6. Test edge cases and boundary conditions
7. Verify database state after operations

**Questions?** Check existing tests or ask the team!
