# Code Locations Reference

This document provides exact file paths and line numbers for key patterns and implementations to reference during development.

---

## Existing Patterns to Follow

### 1. Status Transition Validation (Purchase Order)

**File**: `/Users/kaushik/fpo-erp/internal/services/purchase_order_service.go`

**State Machine Validator** (Lines 875-896):
```go
func isValidPOStatusTransition(from, to string) bool {
    transitions := map[string][]string{
        "placed":           {"confirmed"},
        "confirmed":        {"out_for_delivery"},
        "out_for_delivery": {"delivered"},
        "delivered":        {"paid"},
    }
    // ... validation logic
}
```

**Usage in Service** (Lines 342-344):
```go
if !isValidPOStatusTransition(po.Status, request.Status) {
    return nil, errors.NewBadRequestError(...)
}
```

**Adapt for Sales**: Replace with sale statuses (pending → completed/cancelled)

---

### 2. Transaction-Based Operation (Purchase Order Status Update)

**File**: `/Users/kaushik/fpo-erp/internal/services/purchase_order_service.go`

**Main Method** (Lines 315-418):
```go
func (s *PurchaseOrderService) UpdatePurchaseOrderStatus(ctx context.Context, id string, request *models.UpdatePOStatusRequest, userID string) (*models.PurchaseOrderResponse, error)
```

**Transaction Pattern** (Lines 533-683):
```go
return s.poRepo.WithTransaction(func(tx *gorm.DB) error {
    // Create GRN
    // Create GRN items
    // Create inventory batches
    // Create inventory transactions
    // Update PO status
    return nil
})
```

**Adapt for Cancellation**: Similar structure but reverse inventory instead of create

---

### 3. Inventory Addition (GRN Accepted Items)

**File**: `/Users/kaushik/fpo-erp/internal/services/purchase_order_service.go`

**Inventory Batch Creation** (Lines 606-617):
```go
batch := models.NewInventoryBatch(
    po.WarehouseID,
    poItem.VariantID,
    poItem.UnitPrice,
    expiryDate,
    acceptedQty,
    0, 0, []string{}, false,
)
s.inventoryRepo.CreateBatchWithTx(tx, batch)
```

**Inventory Transaction Creation** (Lines 634-642):
```go
transaction := models.NewInventoryTransaction(
    batch.ID,
    "purchase",
    acceptedQty,  // POSITIVE for addition
    &grn.ID,
    &userID,
    &note,
    actualDelivery,
)
s.inventoryRepo.CreateTransactionWithTx(tx, transaction)
```

**Difference for Cancellation**: Use existing batch, not create new

---

### 4. Sale Creation with FEFO (Multi-Batch Allocation)

**File**: `/Users/kaushik/fpo-erp/internal/services/sales_service.go`

**FEFO Allocation Loop** (Lines 207-296):
```go
for _, batch := range batches {
    if remainingQuantity <= 0 {
        break
    }

    quantityFromBatch := remainingQuantity
    if batch.TotalQuantity < remainingQuantity {
        quantityFromBatch = batch.TotalQuantity
    }

    // Create sale item
    saleItem := models.NewSaleItemWithTax(...)
    s.salesRepo.CreateSaleItemWithTx(tx, saleItem)

    // Create inventory transaction
    transaction := models.NewInventoryTransaction(
        batch.ID,
        "sale",
        -quantityFromBatch,  // NEGATIVE for deduction
        &sale.ID,
        nil,
        stringPtr("Sale transaction"),
        time.Now(),
    )
    s.inventoryRepo.CreateTransactionWithTx(tx, transaction)

    // Update batch stock
    s.inventoryRepo.UpdateBatchStockWithTx(tx, batch.ID, -quantityFromBatch)

    remainingQuantity -= quantityFromBatch
}
```

**Reverse for Cancellation**: Read existing sale_items, add back to batches

---

### 5. Transaction Types Validation

**File**: `/Users/kaushik/fpo-erp/internal/services/inventory_service.go`

**Valid Transaction Types** (Lines 269-276):
```go
validTypes := []string{
    "import",
    "manual_add",
    "adjustment",
    "sale_deduction",  // CURRENT NAME - consider renaming to "sale"
    "return_add",
    "transfer_in",
    "transfer_out",
}
```

**Add**: `"sale_cancellation"` to this list

---

## Model Definitions

### 1. Sale Model

**File**: `/Users/kaushik/fpo-erp/internal/database/models/sales.go`

**Current Structure** (Lines 11-28):
```go
type Sale struct {
    base.BaseModel
    WarehouseID string
    SaleDate    time.Time
    TotalAmount float64
    Status      string  // NEED TO ADD: cancellation fields
    FarmerID    *string
    PaymentMode string
    SaleType    string
    ApplyTaxes  bool
    Warehouse Warehouse
    Items     []SaleItem
}
```

**Fields to Add**:
- `CancellationDate *time.Time`
- `CancellationNote *string`
- `CancelledBy *string`

**Constructor** (Lines 79-92):
```go
func NewSale(warehouseID string, saleDate time.Time, totalAmount float64, status string, farmerID *string, paymentMode, saleType string, applyTaxes bool) *Sale
```

---

### 2. Sale Item Model

**File**: `/Users/kaushik/fpo-erp/internal/database/models/sales.go`

**Structure** (Lines 34-60):
```go
type SaleItem struct {
    base.BaseModel
    SaleID       string
    BatchID      string  // KEY: Used to reverse allocation
    Quantity     int64
    SellingPrice float64
    LineTotal    float64
    CostPrice    float64
    Margin       float64
    // Tax fields...
    Sale  Sale
    Batch InventoryBatch
}
```

**Important**: `BatchID` links to the exact batch that was deducted

---

### 3. Inventory Batch Model

**File**: `/Users/kaushik/fpo-erp/internal/database/models/inventory.go`

**Structure** (Lines 11-29):
```go
type InventoryBatch struct {
    base.BaseModel
    WarehouseID   string
    VariantID     string
    CostPrice     float64
    ExpiryDate    time.Time
    TotalQuantity int64  // Updated on every transaction
    // Tax config...
    Warehouse Warehouse
    Variant   ProductVariant
}
```

---

### 4. Inventory Transaction Model

**File**: `/Users/kaushik/fpo-erp/internal/database/models/inventory.go`

**Structure** (Lines 36-48):
```go
type InventoryTransaction struct {
    base.BaseModel
    BatchID         string
    TransactionType string  // "sale", "sale_cancellation", etc.
    QuantityChange  int64   // Negative = deduction, Positive = addition
    RelatedEntityID *string // Links to sale ID
    PerformedBy     *string // User ID
    Note            *string
    OccurredAt      time.Time
    Batch InventoryBatch
}
```

**Constructor** (Lines 153-165):
```go
func NewInventoryTransaction(batchID, transactionType string, quantityChange int64, relatedEntityID *string, performedBy *string, note *string, occurredAt time.Time) *InventoryTransaction
```

---

## Repository Methods

### 1. Sales Repository (Existing)

**File**: `/Users/kaushik/fpo-erp/internal/database/repositories/sales_repo.go`

**Transaction Helper** (check implementation):
```go
func (r *SalesRepository) WithTransaction(fn func(*gorm.DB) error) error
```

**Methods to Add**:
- `GetSaleWithItemsAndBatches(saleID string) (*models.Sale, error)`
- `UpdateSaleStatusWithTx(tx *gorm.DB, sale *models.Sale) error`

---

### 2. Inventory Repository (Existing)

**File**: `/Users/kaushik/fpo-erp/internal/database/repositories/inventory_repo.go`

**Transaction Methods** (Lines 24-42, 54-60, 155-169):
```go
// Create batch within transaction
func (r *InventoryRepository) CreateBatchWithTx(tx *gorm.DB, batch *models.InventoryBatch) error

// Create transaction within transaction
func (r *InventoryRepository) CreateTransactionWithTx(tx *gorm.DB, transaction *models.InventoryTransaction) error

// Update batch stock within transaction
func (r *InventoryRepository) UpdateBatchStockWithTx(tx *gorm.DB, batchID string, quantityChange int64) error
```

**All required methods already exist** - no changes needed

---

## Handler Patterns

### 1. Sales Handler (Existing)

**File**: `/Users/kaushik/fpo-erp/internal/api/handlers/sales_handler.go`

**Typical Handler Structure** (reference for new CancelSale handler):
```go
func (h *SalesHandler) SomeMethod(c *gin.Context) {
    // Extract parameters
    id := c.Param("id")

    // Parse request body
    var req models.SomeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Extract user from context (set by AAA middleware)
    userID := c.GetString("user_id")

    // Call service
    response, err := h.salesService.SomeMethod(id, &req)
    if err != nil {
        // Error handling based on type
        return
    }

    c.JSON(200, response)
}
```

**New Handler to Add**: `CancelSale(c *gin.Context)`

---

### 2. Purchase Order Handler (Reference)

**File**: `/Users/kaushik/fpo-erp/internal/api/handlers/purchase_order_handler.go`

**Status Update Handler** (reference pattern):
```go
func (h *PurchaseOrderHandler) UpdatePOStatus(c *gin.Context) {
    poID := c.Param("id")

    var req models.UpdatePOStatusRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    userID := c.GetString("user_id")

    response, err := h.purchaseOrderService.UpdatePurchaseOrderStatus(c.Request.Context(), poID, &req, userID)
    // Error handling...

    c.JSON(http.StatusOK, response)
}
```

---

## Routes Configuration

### 1. Route Registration

**File**: `/Users/kaushik/fpo-erp/internal/api/routes/routes.go`

**Pattern for Protected Route**:
```go
salesGroup := v1.Group("/sales")
{
    salesGroup.GET("",
        middleware.RequireOrgPermission("sales", "read"),
        salesHandler.GetAllSales,
    )
    salesGroup.POST("",
        middleware.RequireOrgPermission("sales", "create"),
        salesHandler.CreateSale,
    )
    // ADD NEW ROUTE HERE:
    salesGroup.POST("/:id/cancel",
        middleware.RequireOrgPermission("sales", "cancel"),
        salesHandler.CancelSale,
    )
}
```

---

## AAA Integration

### 1. Permission Middleware

**File**: `/Users/kaushik/fpo-erp/internal/aaa/middleware.go`

**Organization-Scoped Permission** (reference):
```go
middleware.RequireOrgPermission("sales", "cancel")
```

This checks if user has permission: `sales:cancel` for their organization

**User ID Extraction** (in handler):
```go
userID := c.GetString("user_id")  // Set by AAA middleware
```

---

## Constants

### 1. Table Identifiers

**File**: `/Users/kaushik/fpo-erp/internal/constants/table_ids.go`

**Existing Constants** (Lines 11, 16, 24-27):
```go
const (
    TableSale        = "SALE"
    TableBatch       = "BATC"
    TableTransaction = "TRAN"
    TableTax         = "TAXX"
    // ... others
)
```

**No changes required** - all needed constants exist

---

## Error Handling

### 1. Custom Errors

**File**: `/Users/kaushik/fpo-erp/internal/errors/errors.go`

**Error Constructors** (use these):
```go
errors.NewBadRequestError("message")        // 400
errors.NewNotFoundError("entity")           // 404
errors.NewConflictError("message")          // 409
errors.NewInternalServerError("message")    // 500
errors.NewValidationError("message")        // 400 with validation context
```

---

## Logging

### 1. Structured Logging Pattern

**File**: `/Users/kaushik/fpo-erp/internal/services/sales_service.go` (reference)

**Service Logging Pattern** (Lines 48-57):
```go
s.logger.Info("Creating sale",
    zap.String("warehouse_id", req.WarehouseID),
    zap.Int("item_count", len(req.Items)))

s.logger.Error("Sale validation failed",
    zap.Error(err),
    zap.String("warehouse_id", req.WarehouseID))

s.logger.Debug("Sale validation passed")
```

**Use for Cancellation**:
```go
s.logger.Info("Sale cancellation initiated",
    zap.String("sale_id", saleID),
    zap.String("cancelled_by", request.CancelledBy),
    zap.String("reason", request.Reason))
```

---

## Testing Infrastructure

### 1. Test Fixtures

**File**: `/Users/kaushik/fpo-erp/tests/testutils/fixtures.go`

**Existing Fixtures** (reference):
```go
func FixtureSale(...) *models.Sale
func FixtureSaleItem(...) *models.SaleItem
func FixtureInventoryBatch(...) *models.InventoryBatch
func FixtureInventoryTransaction(...) *models.InventoryTransaction
```

**May Need to Add**:
- `FixtureSaleWithItems()` - sale with multiple items
- `FixtureSaleWithCancellation()` - cancelled sale

---

### 2. Mock Repositories

**File**: `/Users/kaushik/fpo-erp/tests/mocks/repositories/`

**Pattern for Mocking** (reference):
```go
mockRepo := new(mocks.MockSalesRepository)
mockRepo.On("GetSaleByID", saleID).Return(sale, nil)
mockRepo.On("WithTransaction", mock.Anything).Return(nil)
```

---

### 3. Service Tests

**File**: `/Users/kaushik/fpo-erp/tests/services/sales_service_test.go`

**Test Pattern** (reference existing tests):
```go
func TestCancelSale_Success(t *testing.T) {
    // Setup
    mockRepo := new(mocks.MockSalesRepository)
    service := NewSalesService(mockRepo, ...)

    // Test data
    sale := fixtures.FixtureSaleWithItems(...)

    // Mocking
    mockRepo.On("GetSaleWithItemsAndBatches", sale.ID).Return(sale, nil)
    mockRepo.On("WithTransaction", mock.Anything).Return(nil)

    // Execute
    result, err := service.CancelSale(sale.ID, &models.CancelSaleRequest{...})

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "cancelled", result.Status)
    mockRepo.AssertExpectations(t)
}
```

---

## Database Migrations

### 1. Migration Location

**Directory**: `/Users/kaushik/fpo-erp/migrations/`

**Naming Convention**: `YYYYMMDD_description.sql`

**Example**: `20251125_add_sale_cancellation_fields.sql`

**Pattern** (reference existing migrations):
```sql
-- Up migration
ALTER TABLE sales ADD COLUMN IF NOT EXISTS cancellation_date TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_sales_status_cancellation ...;

-- Down migration (commented)
-- ALTER TABLE sales DROP COLUMN IF EXISTS cancellation_date;
```

---

## Summary of File Changes

### Files to Modify (7 files + 1 migration):

1. ✏️ `/Users/kaushik/fpo-erp/internal/database/models/sales.go`
   - Add cancellation fields to Sale struct
   - Add CancelSaleRequest model
   - Update SaleResponse model

2. ✏️ `/Users/kaushik/fpo-erp/internal/database/repositories/sales_repo.go`
   - Add GetSaleWithItemsAndBatches()
   - Add UpdateSaleStatusWithTx()

3. ✏️ `/Users/kaushik/fpo-erp/internal/services/sales_service.go`
   - Add CancelSale() method
   - Add validateSaleCancellation()
   - Add isValidSaleStatusTransition()

4. ✏️ `/Users/kaushik/fpo-erp/internal/services/inventory_service.go`
   - Update validTypes array (line 269)

5. ✏️ `/Users/kaushik/fpo-erp/internal/api/handlers/sales_handler.go`
   - Add CancelSale() handler

6. ✏️ `/Users/kaushik/fpo-erp/internal/api/routes/routes.go`
   - Add cancel route registration

7. ✏️ `/Users/kaushik/fpo-erp/tests/services/sales_service_test.go`
   - Add test cases for cancellation

8. ➕ `/Users/kaushik/fpo-erp/migrations/20251125_add_sale_cancellation_fields.sql`
   - New migration file

---

## Quick Navigation

**Start Here**:
1. Read: `/Users/kaushik/fpo-erp/.kiro/specs/order-cancellation/technical-assessment.md`
2. Reference: `/Users/kaushik/fpo-erp/.kiro/specs/order-cancellation/quick-reference.md`
3. This file: For exact line numbers and code locations

**Key Reference Patterns**:
- Status transitions: `purchase_order_service.go:875-896`
- Transaction pattern: `purchase_order_service.go:533-683`
- FEFO allocation: `sales_service.go:207-296`
- Inventory addition: `purchase_order_service.go:606-642`

**Models**: All in `/Users/kaushik/fpo-erp/internal/database/models/`
**Services**: All in `/Users/kaushik/fpo-erp/internal/services/`
**Handlers**: All in `/Users/kaushik/fpo-erp/internal/api/handlers/`

---

**Last Updated**: 2025-11-25
**Maintained By**: Backend Engineering Team
