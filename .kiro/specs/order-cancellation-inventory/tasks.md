# Order Cancellation Implementation Tasks

## Task Tracking Document

**Last Updated**: 2024-11-25
**Status**: Ready for Implementation

---

## Overview

This document tracks the implementation tasks for the Order Cancellation Inventory Return System.
Each task is designed to be a meaningful, committable unit of work.

---

## Phase 1: Database & Models

### CANC-001: Create Database Migration

**Priority**: P0 (Blocker)
**Estimate**: 2 hours
**Dependencies**: None

**Description**:
Create the database migration file for new cancellation tables and sale updates.

**Files to Create/Modify**:
- `migrations/sql/20241125_001_create_sale_cancellation_tables.sql`

**Acceptance Criteria**:
- [ ] Migration creates `sale_cancellations` table
- [ ] Migration creates `sale_cancellation_items` table
- [ ] Migration adds `cancelled_at` and `cancellation_reason` columns to `sales`
- [ ] All indexes created
- [ ] Migration is reversible
- [ ] Migration passes in test environment

**Implementation Notes**:
```sql
-- See design.md Section 6.4 for full DDL
```

---

### CANC-002: Implement SaleCancellation Model

**Priority**: P0 (Blocker)
**Estimate**: 2 hours
**Dependencies**: CANC-001

**Description**:
Create the Go model for SaleCancellation with all DTOs and constructors.

**Files to Create/Modify**:
- `internal/database/models/sale_cancellation.go`
- `internal/constants/table_ids.go` (add new table ID constants)

**Acceptance Criteria**:
- [ ] Model struct defined with GORM tags
- [ ] TableName() method implemented
- [ ] NewSaleCancellation() constructor created
- [ ] Response DTOs defined
- [ ] Request DTOs defined with validation tags
- [ ] Constants for cancellation types added
- [ ] Constants for cancellation reasons added

**Implementation Example**:
```go
// CancellationType enum
const (
    CancellationTypeFull    = "full"
    CancellationTypePartial = "partial"
)

// CancellationReason enum
const (
    ReasonCustomerRequest = "customer_request"
    ReasonPaymentFailed   = "payment_failed"
    ReasonOutOfStock      = "out_of_stock"
    ReasonPricingError    = "pricing_error"
    ReasonDuplicateOrder  = "duplicate_order"
    ReasonFraudSuspected  = "fraud_suspected"
    ReasonSystemError     = "system_error"
    ReasonOther           = "other"
)
```

---

### CANC-003: Implement SaleCancellationItem Model

**Priority**: P0 (Blocker)
**Estimate**: 2 hours
**Dependencies**: CANC-002

**Description**:
Create the Go model for SaleCancellationItem with associations.

**Files to Create/Modify**:
- `internal/database/models/sale_cancellation.go` (extend)

**Acceptance Criteria**:
- [ ] Model struct defined with GORM tags
- [ ] TableName() method implemented
- [ ] NewSaleCancellationItem() constructor created
- [ ] Proper foreign key associations defined
- [ ] Response DTO defined

---

### CANC-004: Add Cancellation Repository Methods

**Priority**: P0 (Blocker)
**Estimate**: 4 hours
**Dependencies**: CANC-003

**Description**:
Create repository layer for cancellation operations.

**Files to Create/Modify**:
- `internal/database/repositories/sale_cancellation_repo.go`
- `internal/database/repositories/interfaces/sale_cancellation.go`

**Acceptance Criteria**:
- [ ] CreateCancellation() method
- [ ] CreateCancellationWithTx() method (transactional)
- [ ] CreateCancellationItem() method
- [ ] CreateCancellationItemWithTx() method
- [ ] GetCancellationByID() method
- [ ] GetCancellationsBySaleID() method
- [ ] Interface defined for dependency injection
- [ ] All methods use proper error handling

**Implementation Notes**:
Follow existing repository patterns in `sales_repo.go`.

---

### CANC-005: Add New Inventory Transaction Type

**Priority**: P0 (Blocker)
**Estimate**: 1 hour
**Dependencies**: None

**Description**:
Add `cancellation_return` to valid inventory transaction types.

**Files to Modify**:
- `internal/services/inventory_service.go`

**Acceptance Criteria**:
- [ ] `cancellation_return` added to validTypes slice
- [ ] Documentation comment updated

**Code Change**:
```go
validTypes := []string{
    "import",
    "manual_add",
    "adjustment",
    "sale_deduction",
    "return_add",
    "transfer_in",
    "transfer_out",
    "cancellation_return", // NEW
}
```

---

### CANC-006: Update Sale Model with Cancellation Fields

**Priority**: P0 (Blocker)
**Estimate**: 1 hour
**Dependencies**: CANC-001

**Description**:
Add cancellation-related fields to the existing Sale model.

**Files to Modify**:
- `internal/database/models/sales.go`

**Acceptance Criteria**:
- [ ] `CancelledAt` field added (nullable timestamp)
- [ ] `CancellationReason` field added (nullable string)
- [ ] SaleResponse DTO updated with cancellation fields
- [ ] NewSale() constructor unchanged (backward compatible)

---

## Phase 2: Core Logic

### CANC-007: Implement CanCancelSale Validation

**Priority**: P0 (Blocker)
**Estimate**: 2 hours
**Dependencies**: CANC-006

**Description**:
Implement the business logic for determining if a sale can be cancelled.

**Files to Modify**:
- `internal/services/sales_service.go`

**Acceptance Criteria**:
- [ ] CanCancelSale() method implemented
- [ ] Returns (bool, string) for cancellability and reason
- [ ] Handles all status cases (pending, confirmed, processing, shipped, delivered, cancelled, returned)
- [ ] Unit tests written
- [ ] Edge cases handled

**Implementation**:
```go
func (s *SalesService) CanCancelSale(sale *models.Sale) (bool, string) {
    switch sale.Status {
    case "cancelled":
        return false, "Sale is already cancelled"
    case "shipped", "delivered":
        return false, "Cannot cancel shipped/delivered orders. Use Returns instead."
    case "returned":
        return false, "Sale has already been returned"
    case "pending", "confirmed", "processing":
        return true, ""
    default:
        return false, "Unknown sale status"
    }
}
```

---

### CANC-008: Implement CancelSale Service Method

**Priority**: P0 (Blocker)
**Estimate**: 6 hours
**Dependencies**: CANC-004, CANC-007

**Description**:
Implement the main CancelSale business logic with full transaction support.

**Files to Modify**:
- `internal/services/sales_service.go`
- `internal/services/interfaces/sales_service.go`

**Acceptance Criteria**:
- [ ] CancelSale() method signature defined in interface
- [ ] CancelSale() method implemented with transaction
- [ ] Uses pessimistic locking (SELECT FOR UPDATE)
- [ ] Creates SaleCancellation record
- [ ] Creates SaleCancellationItem records for each item
- [ ] Calls inventory restore for each item
- [ ] Calls discount reversal
- [ ] Calls tax voiding
- [ ] Updates Sale status to "cancelled"
- [ ] Updates Sale.CancelledAt and CancellationReason
- [ ] Returns proper error codes
- [ ] Structured logging added

**Transaction Flow**:
1. Lock sale record
2. Validate cancellability
3. Create cancellation record
4. For each item: create item record, restore inventory
5. Reverse discounts
6. Void taxes
7. Update sale status
8. Commit

---

### CANC-009: Implement Inventory Restore Logic

**Priority**: P0 (Blocker)
**Estimate**: 4 hours
**Dependencies**: CANC-005

**Description**:
Implement the logic to restore inventory to original batches during cancellation.

**Files to Modify**:
- `internal/services/sales_service.go` (or create dedicated file)
- `internal/database/repositories/inventory_repo.go`

**Acceptance Criteria**:
- [ ] RestoreInventoryForCancellation() method implemented
- [ ] Creates InventoryTransaction with type "cancellation_return"
- [ ] Uses UpdateBatchStockWithTx for atomic update
- [ ] References cancellation ID in transaction
- [ ] Handles edge case: batch no longer exists
- [ ] Returns created transaction IDs
- [ ] Unit tests written

**Implementation Notes**:
```go
func (s *SalesService) restoreInventoryForCancellation(
    tx *gorm.DB,
    saleItem *models.SaleItem,
    cancellationID string,
    performedBy *string,
) (*models.InventoryTransaction, error) {
    note := fmt.Sprintf("Inventory restored for cancellation %s", cancellationID)

    transaction := models.NewInventoryTransaction(
        saleItem.BatchID,
        "cancellation_return",
        saleItem.Quantity, // Positive to add back
        &cancellationID,
        performedBy,
        &note,
        time.Now(),
    )

    if err := s.inventoryRepo.CreateTransactionWithTx(tx, transaction); err != nil {
        return nil, err
    }

    if err := s.inventoryRepo.UpdateBatchStockWithTx(tx, saleItem.BatchID, saleItem.Quantity); err != nil {
        return nil, err
    }

    return transaction, nil
}
```

---

### CANC-010: Implement Discount Reversal Logic

**Priority**: P1 (High)
**Estimate**: 3 hours
**Dependencies**: CANC-008

**Description**:
Implement logic to reverse discounts when a sale is cancelled.

**Files to Modify**:
- `internal/services/sales_service.go`
- `internal/database/repositories/discounts_repo.go`
- `internal/database/models/discount.go` (add IsReversed field to DiscountUsage)

**Acceptance Criteria**:
- [ ] reverseDiscountsForCancellation() method implemented
- [ ] Decrements usage count on discount
- [ ] Marks DiscountUsage as reversed
- [ ] Handles case where usage limit changes
- [ ] Works within transaction
- [ ] Unit tests written

**Repository Methods Needed**:
```go
func (r *DiscountsRepository) DecrementUsageWithTx(tx *gorm.DB, discountID string) error
func (r *DiscountsRepository) GetUsagesBySaleID(saleID string) ([]models.DiscountUsage, error)
func (r *DiscountsRepository) UpdateUsageWithTx(tx *gorm.DB, usage *models.DiscountUsage) error
```

---

### CANC-011: Implement Tax Voiding Logic

**Priority**: P1 (High)
**Estimate**: 2 hours
**Dependencies**: CANC-008

**Description**:
Implement logic to void tax records when a sale is cancelled.

**Files to Modify**:
- `internal/services/sales_service.go`
- `internal/database/repositories/tax_repo.go`
- `internal/database/models/tax.go` (add IsVoided field to TaxSummary)

**Acceptance Criteria**:
- [ ] voidTaxesForCancellation() method implemented
- [ ] Marks TaxSummary as voided
- [ ] Preserves original amounts for audit
- [ ] Works within transaction
- [ ] Unit tests written

---

### CANC-012: Unit Tests for Cancellation Service

**Priority**: P0 (Blocker)
**Estimate**: 4 hours
**Dependencies**: CANC-008

**Description**:
Write comprehensive unit tests for the cancellation service.

**Files to Create**:
- `internal/services/sales_cancellation_test.go`

**Test Cases**:
- [ ] TestCancelSale_Success
- [ ] TestCancelSale_SaleNotFound
- [ ] TestCancelSale_AlreadyCancelled
- [ ] TestCancelSale_ShippedOrder
- [ ] TestCancelSale_DeliveredOrder
- [ ] TestCancelSale_InventoryRestore
- [ ] TestCancelSale_DiscountReversal
- [ ] TestCancelSale_TaxVoiding
- [ ] TestCancelSale_TransactionRollback

---

## Phase 3: Partial Cancellation

### CANC-013: Implement CancelItems Service Method

**Priority**: P1 (High)
**Estimate**: 6 hours
**Dependencies**: CANC-008

**Description**:
Implement partial cancellation for specific items.

**Files to Modify**:
- `internal/services/sales_service.go`
- `internal/services/interfaces/sales_service.go`

**Acceptance Criteria**:
- [ ] CancelItems() method signature defined
- [ ] Validates items belong to sale
- [ ] Validates quantities
- [ ] Supports partial quantity cancellation
- [ ] Creates cancellation records
- [ ] Updates sale total
- [ ] Updates sale status to "partially_cancelled" if items remain
- [ ] Full transaction support
- [ ] Unit tests written

---

### CANC-014: Implement Discount Recalculation

**Priority**: P1 (High)
**Estimate**: 4 hours
**Dependencies**: CANC-013

**Description**:
Recalculate discounts after partial cancellation.

**Files to Modify**:
- `internal/services/sales_service.go`

**Acceptance Criteria**:
- [ ] recalculateDiscountsAfterPartialCancellation() method
- [ ] Checks if remaining order value meets minimum thresholds
- [ ] Removes discounts that no longer apply
- [ ] Adjusts discount amounts proportionally where applicable
- [ ] Updates discount usage records
- [ ] Unit tests written

---

### CANC-015: Implement Tax Recalculation

**Priority**: P1 (High)
**Estimate**: 3 hours
**Dependencies**: CANC-013

**Description**:
Recalculate taxes after partial cancellation.

**Files to Modify**:
- `internal/services/sales_service.go`

**Acceptance Criteria**:
- [ ] recalculateTaxesAfterPartialCancellation() method
- [ ] Updates TaxSummary with new amounts
- [ ] Preserves audit trail of changes
- [ ] Unit tests written

---

### CANC-016: Unit Tests for Partial Cancellation

**Priority**: P1 (High)
**Estimate**: 4 hours
**Dependencies**: CANC-013

**Description**:
Write unit tests for partial cancellation functionality.

**Files to Create/Modify**:
- `internal/services/sales_cancellation_test.go` (extend)

**Test Cases**:
- [ ] TestCancelItems_Success_SingleItem
- [ ] TestCancelItems_Success_MultipleItems
- [ ] TestCancelItems_PartialQuantity
- [ ] TestCancelItems_InvalidItem
- [ ] TestCancelItems_ExceedsQuantity
- [ ] TestCancelItems_AllItems (should use full cancellation)
- [ ] TestCancelItems_DiscountRecalculation
- [ ] TestCancelItems_TaxRecalculation

---

## Phase 4: API Layer

### CANC-017: Implement CancelSale Handler

**Priority**: P0 (Blocker)
**Estimate**: 3 hours
**Dependencies**: CANC-008

**Description**:
Create HTTP handler for full sale cancellation.

**Files to Create/Modify**:
- `internal/api/handlers/sales_handler.go`

**Acceptance Criteria**:
- [ ] CancelSale() handler method
- [ ] Request validation
- [ ] Proper error mapping
- [ ] Structured logging
- [ ] Swagger documentation
- [ ] Returns CancellationResponse

**API Endpoint**:
```
POST /api/v1/sales/{id}/cancel
```

---

### CANC-018: Implement CancelItems Handler

**Priority**: P1 (High)
**Estimate**: 3 hours
**Dependencies**: CANC-013

**Description**:
Create HTTP handler for partial item cancellation.

**Files to Modify**:
- `internal/api/handlers/sales_handler.go`

**Acceptance Criteria**:
- [ ] CancelItems() handler method
- [ ] Request validation for items array
- [ ] Proper error mapping
- [ ] Structured logging
- [ ] Swagger documentation

**API Endpoint**:
```
POST /api/v1/sales/{id}/cancel-items
```

---

### CANC-019: Implement GetCancellations Handler

**Priority**: P2 (Medium)
**Estimate**: 2 hours
**Dependencies**: CANC-004

**Description**:
Create HTTP handler to retrieve cancellation history.

**Files to Modify**:
- `internal/api/handlers/sales_handler.go`

**Acceptance Criteria**:
- [ ] GetCancellations() handler method
- [ ] Swagger documentation
- [ ] Returns list of cancellations for a sale

**API Endpoint**:
```
GET /api/v1/sales/{id}/cancellations
```

---

### CANC-020: Add Routes and Middleware

**Priority**: P0 (Blocker)
**Estimate**: 2 hours
**Dependencies**: CANC-017, CANC-018, CANC-019

**Description**:
Register new routes with proper authentication and authorization.

**Files to Modify**:
- `internal/api/handlers/sales_handler.go` (RegisterRoutes method)

**Acceptance Criteria**:
- [ ] Cancel route registered: `POST /sales/:id/cancel`
- [ ] Cancel items route registered: `POST /sales/:id/cancel-items`
- [ ] Cancellations route registered: `GET /sales/:id/cancellations`
- [ ] All routes require authentication
- [ ] Cancel routes require `sale:cancel` permission
- [ ] Get cancellations requires `sale:read` permission

**Route Registration**:
```go
// In RegisterRoutes method
sales.POST("/:id/cancel", h.aaaMiddleware.RequireOrgPermission("sale", "cancel"), h.CancelSale)
sales.POST("/:id/cancel-items", h.aaaMiddleware.RequireOrgPermission("sale", "cancel"), h.CancelItems)
sales.GET("/:id/cancellations", h.aaaMiddleware.RequireOrgPermission("sale", "read"), h.GetCancellations)
```

---

### CANC-021: Add Swagger Documentation

**Priority**: P2 (Medium)
**Estimate**: 2 hours
**Dependencies**: CANC-020

**Description**:
Add complete Swagger/OpenAPI documentation for all cancellation endpoints.

**Files to Modify**:
- `internal/api/handlers/sales_handler.go`

**Acceptance Criteria**:
- [ ] @Summary annotations
- [ ] @Description annotations
- [ ] @Tags annotations
- [ ] @Param annotations (path, body)
- [ ] @Success annotations with response types
- [ ] @Failure annotations for all error cases
- [ ] @Security annotations

---

## Phase 5: Observability & Testing

### CANC-022: Add Prometheus Metrics

**Priority**: P1 (High)
**Estimate**: 3 hours
**Dependencies**: CANC-017, CANC-018

**Description**:
Add metrics for monitoring cancellation operations.

**Files to Create/Modify**:
- `internal/services/sales_service.go`
- (Optional) `internal/metrics/sales_metrics.go`

**Acceptance Criteria**:
- [ ] `sale_cancellations_total` counter (by type, reason, status)
- [ ] `sale_cancellation_amount_total` counter (by warehouse, reason)
- [ ] `sale_cancellation_duration_seconds` histogram
- [ ] `inventory_restore_failures_total` counter
- [ ] Metrics exposed on /metrics endpoint

---

### CANC-023: Add Structured Logging

**Priority**: P1 (High)
**Estimate**: 2 hours
**Dependencies**: CANC-008, CANC-013

**Description**:
Ensure comprehensive structured logging throughout cancellation flow.

**Files to Modify**:
- `internal/services/sales_service.go`
- `internal/api/handlers/sales_handler.go`

**Acceptance Criteria**:
- [ ] INFO log on cancellation initiated
- [ ] INFO log on cancellation completed
- [ ] DEBUG logs for each step
- [ ] ERROR logs for failures with context
- [ ] All logs include sale_id, cancellation_id, user_id

---

### CANC-024: Integration Tests

**Priority**: P0 (Blocker)
**Estimate**: 6 hours
**Dependencies**: CANC-020

**Description**:
Write end-to-end integration tests for cancellation flow.

**Files to Create**:
- `tests/integration/cancellation_test.go`

**Test Cases**:
- [ ] Full cancellation E2E
- [ ] Partial cancellation E2E
- [ ] Concurrent cancellation handling
- [ ] Transaction rollback on failure
- [ ] Inventory integrity verification

---

### CANC-025: Load Tests

**Priority**: P2 (Medium)
**Estimate**: 4 hours
**Dependencies**: CANC-024

**Description**:
Performance testing under load.

**Files to Create**:
- `tests/load/cancellation_load_test.go`

**Acceptance Criteria**:
- [ ] 100 concurrent cancellations
- [ ] P95 latency < 500ms
- [ ] No inventory discrepancies
- [ ] Error rate < 0.1%

---

### CANC-026: Security Review

**Priority**: P1 (High)
**Estimate**: 2 hours
**Dependencies**: CANC-020

**Description**:
Review security implications of cancellation feature.

**Checklist**:
- [ ] Authorization properly enforced
- [ ] Input validation complete
- [ ] No SQL injection risks
- [ ] Rate limiting in place
- [ ] Audit logging complete
- [ ] No sensitive data leakage in responses

---

## Phase 6: Documentation & Deployment

### CANC-027: Update API Documentation

**Priority**: P2 (Medium)
**Estimate**: 2 hours
**Dependencies**: CANC-021

**Description**:
Update external API documentation.

**Files to Update**:
- `docs/api/sales.md` (or equivalent)
- Postman collection

**Acceptance Criteria**:
- [ ] All new endpoints documented
- [ ] Request/response examples
- [ ] Error codes documented
- [ ] Postman collection updated

---

### CANC-028: Create Runbook

**Priority**: P2 (Medium)
**Estimate**: 2 hours
**Dependencies**: CANC-022

**Description**:
Create operational runbook for cancellation issues.

**Files to Create**:
- `docs/runbooks/sale-cancellation.md`

**Content**:
- [ ] Common failure scenarios
- [ ] How to investigate stuck cancellations
- [ ] How to manually reconcile inventory
- [ ] Alert response procedures
- [ ] Rollback instructions

---

### CANC-029: Staging Deployment & Testing

**Priority**: P0 (Blocker)
**Estimate**: 4 hours
**Dependencies**: All previous tasks

**Checklist**:
- [ ] Deploy to staging
- [ ] Run integration tests
- [ ] Manual smoke testing
- [ ] Load test on staging
- [ ] Verify metrics collection
- [ ] Verify logging
- [ ] Sign-off from QA

---

### CANC-030: Production Deployment

**Priority**: P0 (Blocker)
**Estimate**: 2 hours
**Dependencies**: CANC-029

**Checklist**:
- [ ] Deploy database migration
- [ ] Deploy application code
- [ ] Enable feature flag for limited users
- [ ] Monitor error rates
- [ ] Monitor latency
- [ ] Gradual rollout (10% -> 50% -> 100%)
- [ ] Remove feature flag after stable

---

## Summary

| Phase | Tasks | Total Estimate |
|-------|-------|----------------|
| Phase 1: Database & Models | 6 tasks | 12 hours |
| Phase 2: Core Logic | 6 tasks | 21 hours |
| Phase 3: Partial Cancellation | 4 tasks | 17 hours |
| Phase 4: API Layer | 5 tasks | 12 hours |
| Phase 5: Observability & Testing | 5 tasks | 17 hours |
| Phase 6: Documentation & Deployment | 4 tasks | 10 hours |

**Grand Total**: 30 tasks, ~89 hours (~2.5 weeks with buffer)

---

## Git Commit Guidelines

Each task should result in one meaningful commit:

```
feat(sales): add sale cancellation database migration

- Create sale_cancellations table
- Create sale_cancellation_items table
- Add cancelled_at and cancellation_reason to sales
- Add indexes for performance

Refs: CANC-001
```

```
feat(sales): implement CancelSale service method

- Add CanCancelSale validation
- Implement full cancellation with transaction
- Add inventory restore logic
- Add discount reversal
- Add tax voiding

Refs: CANC-008
```
