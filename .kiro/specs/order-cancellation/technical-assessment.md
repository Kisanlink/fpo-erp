# Order Cancellation Inventory Return - Technical Assessment

**Date**: 2025-11-25
**Status**: Assessment Phase
**Priority**: High
**Complexity**: Medium

---

## Executive Summary

This document provides a comprehensive technical assessment for implementing order cancellation with automatic inventory return functionality in the FPO ERP system. The assessment covers existing patterns, required changes, technical debt, and implementation strategy.

**Key Findings**:
- ✅ Strong foundation: Existing purchase order workflow already implements similar patterns
- ✅ Transaction-based operations: GORM transaction support is well-established
- ✅ Inventory tracking: Comprehensive transaction logging and FEFO allocation
- ⚠️ Status workflow: Need to add cancellation states to sales flow
- ⚠️ Reverse transaction: New transaction type needed for inventory return
- 📋 Estimated effort: 3-5 days for complete implementation with tests

---

## 1. Current System Analysis

### 1.1 Existing Order/Sale Model

**Location**: `/Users/kaushik/fpo-erp/internal/database/models/sales.go`

**Current Structure**:
```go
type Sale struct {
    base.BaseModel
    WarehouseID string
    SaleDate    time.Time
    TotalAmount float64
    Status      string        // CURRENT: Only "pending", "completed"
    FarmerID    *string
    PaymentMode string
    SaleType    string
    ApplyTaxes  bool
    Items       []SaleItem
}

type SaleItem struct {
    base.BaseModel
    SaleID       string
    BatchID      string
    Quantity     int64
    SellingPrice float64
    LineTotal    float64
    CostPrice    float64
    Margin       float64
    // Tax fields...
}
```

**Current Status Values**:
- `"pending"` - Initial state
- `"completed"` - Finalized

**Gap**: No cancellation status or cancelled state

### 1.2 Inventory Transaction System

**Location**: `/Users/kaushik/fpo-erp/internal/database/models/inventory.go`

**Current Structure**:
```go
type InventoryBatch struct {
    base.BaseModel
    WarehouseID   string
    VariantID     string
    CostPrice     float64
    ExpiryDate    time.Time
    TotalQuantity int64  // Updated on every transaction
    // Tax config...
}

type InventoryTransaction struct {
    base.BaseModel
    BatchID         string
    TransactionType string  // CURRENT TYPES: see below
    QuantityChange  int64   // Negative for deductions, positive for additions
    RelatedEntityID *string // Links to sale/purchase/return ID
    PerformedBy     *string
    Note            *string
    OccurredAt      time.Time
}
```

**Current Transaction Types** (from `/Users/kaushik/fpo-erp/internal/services/inventory_service.go:269`):
```go
validTypes := []string{
    "import",          // Initial inventory import
    "manual_add",      // Manual stock addition
    "adjustment",      // Stock adjustment (audit)
    "sale_deduction",  // CURRENT: Used for sale inventory deductions  <-- RENAME CANDIDATE
    "return_add",      // Return to inventory
    "transfer_in",     // Inter-warehouse transfer in
    "transfer_out",    // Inter-warehouse transfer out
}
```

**Gap**: Need new transaction type for cancellation return: `"cancellation_return"` or rename `"sale_deduction"` to `"sale"` and add `"sale_cancellation"`

### 1.3 Existing Sale Creation Flow

**Location**: `/Users/kaushik/fpo-erp/internal/services/sales_service.go:47-300`

**Transaction Pattern** (lines 144-350):
```go
err := s.salesRepo.WithTransaction(func(tx *gorm.DB) error {
    // 1. Create sale record
    sale := models.NewSale(...)
    s.salesRepo.CreateSaleWithTx(tx, sale)

    // 2. For each item, using FEFO allocation:
    for _, itemReq := range req.Items {
        // 2a. Allocate from batches (oldest expiry first)
        for _, batch := range batches {
            // 2b. Create sale item
            saleItem := models.NewSaleItemWithTax(...)
            s.salesRepo.CreateSaleItemWithTx(tx, saleItem)

            // 2c. Create inventory transaction (DEDUCTION)
            transaction := models.NewInventoryTransaction(
                batch.ID,
                "sale",              // Transaction type
                -quantityFromBatch,  // NEGATIVE for deduction
                &sale.ID,
                nil,
                stringPtr("Sale transaction"),
                time.Now(),
            )
            s.inventoryRepo.CreateTransactionWithTx(tx, transaction)

            // 2d. Update batch total quantity
            s.inventoryRepo.UpdateBatchStockWithTx(tx, batch.ID, -quantityFromBatch)
        }
    }

    // 3. Update sale total amount
    sale.TotalAmount = totalAmount
    s.salesRepo.UpdateSaleWithTx(tx, sale)

    return nil
})
```

**Key Observations**:
- ✅ Atomic transaction handling via GORM
- ✅ FEFO (First-Expired-First-Out) allocation tracked via batch IDs in sale_items
- ✅ Audit trail maintained through inventory_transactions
- ✅ Row-level locking prevents race conditions
- ⚠️ Sale items store batch_id - enables precise reversal
- ⚠️ No status transition validation (pending → completed)

### 1.4 Existing Purchase Order Cancellation Pattern

**Location**: `/Users/kaushik/fpo-erp/internal/services/purchase_order_service.go:315-418`

**Status Transition Logic** (lines 342-344):
```go
// Validate status transition
if !isValidPOStatusTransition(po.Status, request.Status) {
    return nil, errors.NewBadRequestError(...)
}

// Valid transitions defined (lines 876-896)
transitions := map[string][]string{
    "placed":           {"confirmed"},
    "confirmed":        {"out_for_delivery"},
    "out_for_delivery": {"delivered"},
    "delivered":        {"paid"},
}
```

**Key Pattern**: State machine validation prevents invalid transitions

### 1.5 Existing GRN → Inventory Flow (Reference Pattern)

**Location**: `/Users/kaushik/fpo-erp/internal/services/purchase_order_service.go:599-653`

**Inventory Addition Pattern** (lines 599-653):
```go
// When GRN accepted, create inventory batch
if acceptedQty > 0 {
    batch := models.NewInventoryBatch(
        po.WarehouseID,
        poItem.VariantID,
        poItem.UnitPrice,  // Cost price
        expiryDate,
        acceptedQty,
        // Tax config...
    )
    s.inventoryRepo.CreateBatchWithTx(tx, batch)

    // Create inventory transaction for audit
    transaction := models.NewInventoryTransaction(
        batch.ID,
        "purchase",
        acceptedQty,       // POSITIVE for addition
        &grn.ID,
        &userID,
        &note,
        actualDelivery,
    )
    s.inventoryRepo.CreateTransactionWithTx(tx, transaction)
}
```

**Key Observations**:
- ✅ Creates NEW batch on purchase (doesn't update existing)
- ✅ Positive quantity change for additions
- ✅ Links transaction to source entity (GRN ID)
- ⚠️ Different from cancellation: we need to UPDATE existing batch, not create new

---

## 2. Files Requiring Changes

### 2.1 Models Layer

#### `/Users/kaushik/fpo-erp/internal/database/models/sales.go`

**Changes Required**:
1. Add cancellation states to `Sale.Status`:
   ```go
   // Values: "pending", "completed", "cancelled", "refunded"
   ```

2. Add cancellation metadata:
   ```go
   type Sale struct {
       // ... existing fields
       Status           string
       CancellationDate *time.Time `gorm:"type:timestamptz" json:"cancellation_date,omitempty"`
       CancellationNote *string    `gorm:"type:text" json:"cancellation_note,omitempty"`
       CancelledBy      *string    `gorm:"type:varchar(100)" json:"cancelled_by,omitempty"` // User ID
   }
   ```

3. Update request/response models:
   ```go
   type CancelSaleRequest struct {
       Reason      string  `json:"reason" binding:"required,max=500"`
       CancelledBy string  `json:"cancelled_by" binding:"required"` // From JWT token
   }

   type SaleResponse struct {
       // ... existing fields
       CancellationDate *string `json:"cancellation_date,omitempty"`
       CancellationNote *string `json:"cancellation_note,omitempty"`
       CancelledBy      *string `json:"cancelled_by,omitempty"`
   }
   ```

**Risk**: Database migration required (backward compatible)

#### `/Users/kaushik/fpo-erp/internal/constants/table_ids.go`

**Changes Required**: None (already has `TableSale = "SALE"`)

### 2.2 Repository Layer

#### `/Users/kaushik/fpo-erp/internal/database/repositories/sales_repo.go`

**New Methods Required**:
1. `GetSaleWithItemsAndBatches(saleID string) (*models.Sale, error)`
   - Preload: `Items.Batch` for inventory return
   - Required for reversal logic

2. `UpdateSaleStatusWithTx(tx *gorm.DB, sale *models.Sale) error`
   - Update sale status + cancellation metadata
   - Within transaction

**Existing Methods to Use**:
- `WithTransaction(fn func(*gorm.DB) error)` - Already exists
- `CreateSaleItemWithTx` - Reusable
- No changes needed to core CRUD

#### `/Users/kaushik/fpo-erp/internal/database/repositories/inventory_repo.go`

**Existing Methods Sufficient**:
- ✅ `CreateTransactionWithTx(tx *gorm.DB, transaction *models.InventoryTransaction) error` (line 155)
- ✅ `UpdateBatchStockWithTx(tx *gorm.DB, batchID string, quantityChange int64) error` (line 168)
- ✅ `GetBatchByID(id string) (*models.InventoryBatch, error)` (line 63)

**No changes required** - existing methods handle inventory return

### 2.3 Service Layer

#### `/Users/kaushik/fpo-erp/internal/services/sales_service.go`

**New Method Required**:
```go
// CancelSale cancels a sale and returns inventory to batches
func (s *SalesService) CancelSale(saleID string, request *models.CancelSaleRequest) (*models.SaleResponse, error)
```

**Implementation Pattern** (based on purchase_order_service.go:505-684):
```go
err := s.salesRepo.WithTransaction(func(tx *gorm.DB) error {
    // 1. Validate sale exists and is cancellable
    sale, err := s.salesRepo.GetSaleWithItemsAndBatches(saleID)
    if sale.Status != "pending" {
        return errors.NewBadRequestError("Only pending sales can be cancelled")
    }

    // 2. For each sale item, reverse inventory deduction
    for _, saleItem := range sale.Items {
        // 2a. Create reverse inventory transaction
        transaction := models.NewInventoryTransaction(
            saleItem.BatchID,
            "cancellation_return",  // NEW transaction type
            saleItem.Quantity,      // POSITIVE (return to inventory)
            &sale.ID,
            &request.CancelledBy,
            &request.Reason,
            time.Now(),
        )
        s.inventoryRepo.CreateTransactionWithTx(tx, transaction)

        // 2b. Update batch quantity (add back)
        s.inventoryRepo.UpdateBatchStockWithTx(tx, saleItem.BatchID, saleItem.Quantity)
    }

    // 3. Update sale status
    now := time.Now()
    sale.Status = "cancelled"
    sale.CancellationDate = &now
    sale.CancellationNote = &request.Reason
    sale.CancelledBy = &request.CancelledBy

    return s.salesRepo.UpdateSaleStatusWithTx(tx, sale)
})
```

**Validation Logic Required**:
```go
func (s *SalesService) validateSaleCancellation(sale *models.Sale) error {
    // Business rules
    if sale.Status == "cancelled" {
        return errors.NewBadRequestError("Sale already cancelled")
    }
    if sale.Status == "refunded" {
        return errors.NewBadRequestError("Cannot cancel refunded sale")
    }
    if sale.Status != "pending" {
        return errors.NewBadRequestError("Only pending sales can be cancelled")
    }

    // Time-based validation (optional)
    // if time.Since(sale.SaleDate) > 24*time.Hour {
    //     return errors.NewBadRequestError("Sale too old to cancel")
    // }

    return nil
}
```

**Status Transition Validator** (similar to PO service line 875):
```go
func isValidSaleStatusTransition(from, to string) bool {
    transitions := map[string][]string{
        "pending":   {"completed", "cancelled"},
        "completed": {"refunded"},  // Future: refund flow
        // "cancelled": {},  // Terminal state
        // "refunded": {},   // Terminal state
    }

    validNextStatuses, ok := transitions[from]
    if !ok {
        return false
    }

    for _, validStatus := range validNextStatuses {
        if validStatus == to {
            return true
        }
    }
    return false
}
```

#### `/Users/kaushik/fpo-erp/internal/services/inventory_service.go`

**Changes Required**:
1. Update valid transaction types (line 269):
   ```go
   validTypes := []string{
       "import",
       "manual_add",
       "adjustment",
       "sale",                    // RENAMED from "sale_deduction"
       "sale_cancellation",       // NEW
       "return_add",
       "transfer_in",
       "transfer_out",
   }
   ```

2. Update `CreateSale` to use `"sale"` instead of `"sale_deduction"` (line 274)

**Risk**: Breaking change if transaction type is used elsewhere - need grep search

### 2.4 Handler Layer

#### `/Users/kaushik/fpo-erp/internal/api/handlers/sales_handler.go`

**New Handler Required**:
```go
// CancelSale handles POST /api/v1/sales/:id/cancel
func (h *SalesHandler) CancelSale(c *gin.Context) {
    saleID := c.Param("id")

    var req models.CancelSaleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Extract user ID from JWT context
    userID := c.GetString("user_id")
    req.CancelledBy = userID

    response, err := h.salesService.CancelSale(saleID, &req)
    if err != nil {
        // Error handling...
        return
    }

    c.JSON(200, response)
}
```

**Permission Check** (following AAA pattern):
```go
// In route registration
salesRoutes.POST("/:id/cancel",
    middleware.RequireOrgPermission("sales", "cancel"),  // NEW permission
    salesHandler.CancelSale,
)
```

#### `/Users/kaushik/fpo-erp/internal/api/routes/routes.go`

**Route Addition**:
```go
// In setupSalesRoutes or equivalent
salesGroup.POST("/:id/cancel", salesHandler.CancelSale)
```

### 2.5 Database Migration

#### `/Users/kaushik/fpo-erp/migrations/`

**New Migration File Required**: `YYYYMMDD_add_sale_cancellation_fields.sql`

```sql
-- Up Migration
ALTER TABLE sales ADD COLUMN IF NOT EXISTS cancellation_date TIMESTAMPTZ;
ALTER TABLE sales ADD COLUMN IF NOT EXISTS cancellation_note TEXT;
ALTER TABLE sales ADD COLUMN IF NOT EXISTS cancelled_by VARCHAR(100);

-- Create index for cancelled sales queries
CREATE INDEX IF NOT EXISTS idx_sales_status_cancellation
ON sales(status, cancellation_date)
WHERE status IN ('cancelled', 'refunded');

-- Add check constraint for cancellation metadata
ALTER TABLE sales ADD CONSTRAINT chk_cancellation_metadata
CHECK (
    (status = 'cancelled' AND cancellation_date IS NOT NULL AND cancelled_by IS NOT NULL) OR
    (status != 'cancelled')
);

-- Down Migration (in separate file or same file with comments)
-- ALTER TABLE sales DROP CONSTRAINT IF EXISTS chk_cancellation_metadata;
-- DROP INDEX IF EXISTS idx_sales_status_cancellation;
-- ALTER TABLE sales DROP COLUMN IF EXISTS cancelled_by;
-- ALTER TABLE sales DROP COLUMN IF EXISTS cancellation_note;
-- ALTER TABLE sales DROP COLUMN IF EXISTS cancellation_date;
```

---

## 3. Technical Debt & Issues Found

### 3.1 Current Issues

1. **Sale Status Validation Missing** (sales_service.go)
   - No validation for status transitions
   - Status field accepts any string value
   - **Impact**: Can transition from any status to any status
   - **Fix**: Add state machine validator (similar to PO service)

2. **Transaction Type Naming Inconsistency** (inventory_service.go:269)
   - Currently uses `"sale_deduction"` but PO uses `"purchase"` (not `"purchase_addition"`)
   - **Impact**: Confusion in transaction logs
   - **Fix**: Rename to `"sale"` for consistency

3. **No Sale Item-Batch Relationship Preloading** (sales_repo.go)
   - Current repo doesn't preload `Items.Batch` relationship
   - **Impact**: N+1 query problem for cancellation
   - **Fix**: Add preload method

4. **Weak Status Field Definition** (models/sales.go:17)
   - No enum validation at model level
   - Relies on service layer validation only
   - **Impact**: Data integrity risk
   - **Fix**: Consider adding GORM enum type or DB constraint

### 3.2 Missing Patterns

1. **No Soft Delete Support**
   - BaseModel uses hard delete
   - **Impact**: Cannot recover accidentally cancelled sales
   - **Recommendation**: Consider soft delete with `deleted_at` field

2. **No Audit Log for Status Changes**
   - Only final cancellation captured
   - No history of who changed status and when
   - **Impact**: Limited auditability
   - **Recommendation**: Consider separate `sale_status_history` table

3. **No Idempotency Key**
   - Duplicate cancel requests can cause issues
   - **Impact**: Risk of duplicate inventory returns
   - **Recommendation**: Add idempotency key to prevent duplicate operations

---

## 4. Dependencies & Related Systems

### 4.1 AAA Service Integration

**Location**: `/Users/kaushik/fpo-erp/internal/aaa/`

**Requirements**:
- ✅ JWT token validation already implemented
- ✅ User ID extraction from context (middleware.go)
- ⚠️ New permission needed: `"sales:cancel"`
- ⚠️ Organization-scoped permission check required

**Permission Matrix Update Required**:
| Role | Cancel Sale Permission |
|------|----------------------|
| CEO | Yes |
| Store Manager | Yes |
| Store Staff | No (or conditional) |
| Auditor | No |
| Accountant | Conditional |

### 4.2 Tax System Integration

**Location**: `/Users/kaushik/fpo-erp/internal/services/tax_service.go`

**Considerations**:
- ✅ No tax reversal needed (sale was pending, taxes not filed)
- ⚠️ If sale status is "completed", tax reversal logic may be needed
- **Decision**: Limit cancellation to "pending" sales to avoid tax complications

### 4.3 Discount System Integration

**Location**: `/Users/kaushik/fpo-erp/internal/services/discounts_service.go`

**Considerations**:
- Sale items may have discounts applied
- Discount usage tracking in `discount_uses` table
- **Impact**: Need to release discount usage count on cancellation
- **Fix Required**: Add discount usage reversal logic

### 4.4 Payment System Integration

**Location**: `/Users/kaushik/fpo-erp/internal/database/models/bank_payments.go`

**Current State**:
- Sales have `payment_mode` field (cash, upi, online)
- No direct bank_payments table link to sales (based on model inspection)
- **Impact**: If payments are recorded, need refund workflow
- **Decision**: Limit cancellation to unpaid/pending sales initially

---

## 5. Testing Requirements

### 5.1 Unit Tests

**Files to Create/Update**:
1. `/Users/kaushik/fpo-erp/tests/services/sales_service_test.go`
   - `TestCancelSale_Success`
   - `TestCancelSale_InvalidStatus`
   - `TestCancelSale_NotFound`
   - `TestCancelSale_AlreadyCancelled`
   - `TestCancelSale_InventoryReturned`

2. `/Users/kaushik/fpo-erp/tests/handlers/sales_handler_test.go`
   - `TestCancelSaleHandler_Success`
   - `TestCancelSaleHandler_Unauthorized`
   - `TestCancelSaleHandler_ValidationError`

**Test Patterns** (following existing style from purchase_order_service_test.go):
```go
func TestCancelSale_Success(t *testing.T) {
    // Setup
    mockRepo := new(mocks.MockSalesRepository)
    mockInventoryRepo := new(mocks.MockInventoryRepository)
    service := NewSalesService(mockRepo, ..., mockInventoryRepo, ...)

    // Create test sale with items
    sale := fixtures.FixtureSaleWithItems(...)

    // Mock expectations
    mockRepo.On("GetSaleWithItemsAndBatches", sale.ID).Return(sale, nil)
    mockRepo.On("WithTransaction", mock.Anything).Return(nil)

    // Execute
    result, err := service.CancelSale(sale.ID, &models.CancelSaleRequest{
        Reason: "Customer request",
        CancelledBy: "USER_12345",
    })

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "cancelled", result.Status)
    mockRepo.AssertExpectations(t)
    mockInventoryRepo.AssertExpectations(t)
}
```

### 5.2 Integration Tests

**Test Scenarios**:
1. End-to-end: Create sale → Cancel sale → Verify inventory restored
2. Concurrent cancellation attempts (pessimistic locking test)
3. FEFO verification: Multiple batch allocations restored correctly
4. Transaction rollback: Partial failure rolls back all changes

**Test Data Requirements**:
- Multiple inventory batches with different expiry dates
- Sales with multiple items across different batches
- Sales in various statuses

### 5.3 Performance Tests

**Benchmarks Required**:
```go
func BenchmarkCancelSale_SingleItem(b *testing.B)
func BenchmarkCancelSale_MultipleItems(b *testing.B)
func BenchmarkCancelSale_ManyBatchAllocations(b *testing.B)
```

**Performance Targets**:
- Single item cancellation: < 50ms (P95)
- 10 items cancellation: < 200ms (P95)
- 100 concurrent cancellations: No deadlocks

---

## 6. Security Considerations

### 6.1 Authorization

**Checks Required**:
1. ✅ User has `sales:cancel` permission
2. ✅ User belongs to same organization as sale (organization_id scope)
3. ⚠️ Sale belongs to warehouse user has access to
4. ⚠️ Time-based restriction (e.g., only within 24 hours)

### 6.2 Audit Trail

**Logging Requirements**:
```go
s.logger.Info("Sale cancellation initiated",
    zap.String("sale_id", saleID),
    zap.String("cancelled_by", request.CancelledBy),
    zap.String("reason", request.Reason),
    zap.String("original_status", sale.Status),
    zap.Float64("total_amount", sale.TotalAmount))

s.logger.Info("Inventory restored from cancelled sale",
    zap.String("sale_id", saleID),
    zap.Int("items_count", len(sale.Items)),
    zap.Int64("total_quantity_returned", totalQty))
```

### 6.3 Input Validation

**Validations Required**:
```go
// In models/sales.go
type CancelSaleRequest struct {
    Reason      string `json:"reason" binding:"required,min=10,max=500"`
    CancelledBy string `json:"cancelled_by" binding:"required,uuid"`
}
```

**SQL Injection Prevention**: ✅ Already handled by GORM

---

## 7. Observability & Monitoring

### 7.1 Metrics

**RED Metrics**:
- `sale_cancellation_requests_total` (counter)
- `sale_cancellation_errors_total` (counter, labeled by error_type)
- `sale_cancellation_duration_seconds` (histogram)

**Business Metrics**:
- `cancelled_sales_total` (counter, labeled by warehouse, reason_type)
- `cancelled_sales_amount_total` (counter)
- `inventory_returned_quantity_total` (counter)

### 7.2 Alerts

**Critical Alerts**:
- High cancellation rate (> 10% of sales in 1 hour)
- Transaction failures (> 5 in 5 minutes)
- Inventory discrepancies after cancellation

### 7.3 Distributed Tracing

**Trace Spans Required**:
```
sale_cancellation
├── validate_sale
├── load_sale_items
├── database_transaction
│   ├── create_inventory_transactions
│   ├── update_batch_stocks
│   └── update_sale_status
└── build_response
```

---

## 8. Implementation Strategy

### 8.1 Phased Approach

**Phase 1: Foundation (Day 1-2)**
- ✅ Database migration
- ✅ Model updates (Sale, SaleItem)
- ✅ Repository method additions
- ✅ Unit tests for models

**Phase 2: Core Logic (Day 2-3)**
- ✅ Service layer implementation
- ✅ Inventory return logic
- ✅ Status transition validation
- ✅ Unit tests for service

**Phase 3: API Layer (Day 3-4)**
- ✅ Handler implementation
- ✅ Route registration
- ✅ Permission integration
- ✅ Integration tests

**Phase 4: Polish & Deploy (Day 4-5)**
- ✅ API documentation (Swagger/OpenAPI)
- ✅ Performance testing
- ✅ Security review
- ✅ Runbook creation

### 8.2 Rollout Strategy

**Step 1: Feature Flag**
```go
// In config
EnableSaleCancellation bool `env:"ENABLE_SALE_CANCELLATION" default:"false"`

// In handler
if !h.config.EnableSaleCancellation {
    c.JSON(501, gin.H{"error": "Feature not enabled"})
    return
}
```

**Step 2: Canary Deployment**
- Deploy to staging
- Test with real-world data
- Monitor metrics for 24 hours

**Step 3: Gradual Rollout**
- Enable for 10% of organizations
- Monitor for 1 week
- Expand to 100%

### 8.3 Rollback Plan

**Rollback Triggers**:
- Inventory discrepancies detected
- Transaction failure rate > 5%
- Database deadlocks

**Rollback Steps**:
1. Disable feature flag
2. Investigate failed transactions
3. Manually correct inventory if needed
4. Fix bugs and redeploy

---

## 9. API Contract

### 9.1 Endpoint Specification

**Request**:
```http
POST /api/v1/sales/{sale_id}/cancel
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "reason": "Customer requested cancellation due to wrong order"
}
```

**Response (Success - 200)**:
```json
{
  "id": "SALE_abc123",
  "warehouse_id": "WHSE_xyz789",
  "sale_date": "2025-11-25T10:30:00Z",
  "total_amount": 1250.00,
  "status": "cancelled",
  "cancellation_date": "2025-11-25T11:00:00Z",
  "cancellation_note": "Customer requested cancellation due to wrong order",
  "cancelled_by": "USER_def456",
  "payment_mode": "cash",
  "sale_type": "in_store",
  "items": [
    {
      "id": "SITM_item1",
      "batch_id": "BATC_batch1",
      "quantity": 10,
      "selling_price": 125.00,
      "line_total": 1250.00
    }
  ],
  "created_at": "2025-11-25T10:30:00Z",
  "updated_at": "2025-11-25T11:00:00Z"
}
```

**Error Responses**:

```json
// 400 Bad Request - Invalid Status
{
  "error": "Only pending sales can be cancelled",
  "code": "INVALID_SALE_STATUS"
}

// 404 Not Found
{
  "error": "Sale not found",
  "code": "SALE_NOT_FOUND"
}

// 403 Forbidden
{
  "error": "Insufficient permissions to cancel sale",
  "code": "PERMISSION_DENIED"
}

// 409 Conflict
{
  "error": "Sale already cancelled",
  "code": "SALE_ALREADY_CANCELLED"
}
```

---

## 10. Risks & Mitigation

### 10.1 High Risks

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| Race condition: Concurrent cancel + update | Medium | High | Pessimistic locking (SELECT FOR UPDATE) |
| Inventory mismatch after cancellation | Low | Critical | Transaction rollback + reconciliation job |
| Partial cancellation due to DB failure | Low | High | ACID transaction + idempotency key |
| Performance degradation on large sales | Medium | Medium | Batch processing + async option |

### 10.2 Medium Risks

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| Breaking change to transaction types | Low | Medium | Feature flag + gradual rollout |
| Missing discount usage reversal | High | Low | Add to Phase 2 implementation |
| Tax calculation issues | Low | Medium | Limit to pending sales only |

### 10.3 Low Risks

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| API rate limiting issues | Low | Low | Standard rate limiting applies |
| Logging volume increase | Medium | Low | Structured logging with levels |

---

## 11. Open Questions for Design Document

1. **Business Rules**:
   - Should we allow cancellation of "completed" sales?
   - Time limit for cancellation (e.g., 24 hours)?
   - Should cancellation require approval workflow?
   - Partial cancellation support (cancel some items only)?

2. **Financial Implications**:
   - What if payment was already processed?
   - Refund workflow integration?
   - Accounting journal entry needed?

3. **Inventory Rules**:
   - Return to original batches or create new batch?
   - Update batch expiry date if goods are inspected again?
   - Quality check required before returning to inventory?

4. **Discount Handling**:
   - Release discount usage count?
   - Allow discount reuse by same customer?
   - Blacklist discount code if abuse detected?

5. **Notification**:
   - Notify customer of cancellation?
   - Email/SMS integration?
   - Webhook for external systems?

---

## 12. Next Steps

### Immediate Actions (Before Implementation):
1. ✅ Review this technical assessment with tech lead
2. ⏳ Create design document answering open questions
3. ⏳ Get business stakeholder approval for rules
4. ⏳ Create JIRA/Linear tickets for each phase
5. ⏳ Schedule tech review session

### Implementation Checklist:
- [ ] Database migration script created and tested
- [ ] Model changes implemented
- [ ] Repository methods added
- [ ] Service layer implemented with full transaction logic
- [ ] Handler and routes configured
- [ ] Unit tests written (coverage > 80%)
- [ ] Integration tests written
- [ ] API documentation updated (OpenAPI spec)
- [ ] Permission matrix updated
- [ ] Runbook created for operations team
- [ ] Feature flag added
- [ ] Metrics and alerts configured
- [ ] Security review completed
- [ ] Performance tests passed
- [ ] Staging deployment successful
- [ ] Production deployment planned

---

## 13. References

**Codebase Files Analyzed**:
- `/Users/kaushik/fpo-erp/internal/database/models/sales.go`
- `/Users/kaushik/fpo-erp/internal/database/models/inventory.go`
- `/Users/kaushik/fpo-erp/internal/database/models/purchase_order.go`
- `/Users/kaushik/fpo-erp/internal/services/sales_service.go`
- `/Users/kaushik/fpo-erp/internal/services/inventory_service.go`
- `/Users/kaushik/fpo-erp/internal/services/purchase_order_service.go`
- `/Users/kaushik/fpo-erp/internal/database/repositories/inventory_repo.go`
- `/Users/kaushik/fpo-erp/internal/constants/table_ids.go`
- `/Users/kaushik/fpo-erp/README.md`
- `/Users/kaushik/fpo-erp/.kiro/specs/IMPLEMENTATION-SUMMARY.md`

**Existing Patterns Referenced**:
- Purchase Order status transitions (lines 875-896 in purchase_order_service.go)
- GRN inventory creation (lines 599-653 in purchase_order_service.go)
- Sale creation with FEFO allocation (lines 207-296 in sales_service.go)
- Transaction-based operations (lines 144-350 in sales_service.go)

**Tech Stack**:
- Go 1.21+
- GORM (ORM)
- PostgreSQL (Database)
- Gin (HTTP Framework)
- Zap (Logging)
- JWT (Authentication)

---

**Document Version**: 1.0
**Last Updated**: 2025-11-25
**Author**: SDE-2 Backend Engineer
**Review Status**: Ready for Tech Lead Review
