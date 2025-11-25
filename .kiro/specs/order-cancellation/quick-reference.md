# Order Cancellation - Quick Reference

**Assessment Date**: 2025-11-25
**Implementation Estimate**: 3-5 days
**Complexity**: Medium

---

## TL;DR

Implement order cancellation with automatic inventory return by:
1. Adding cancellation fields to `Sale` model (DB migration)
2. Creating `CancelSale()` service method with transaction-based inventory reversal
3. Adding POST `/api/v1/sales/:id/cancel` endpoint
4. Implementing status transition validation
5. Adding `"sale_cancellation"` transaction type

**Pattern**: Mirror existing purchase order flow, use GORM transactions, follow FEFO reversal.

---

## Key Files to Modify

### Models (1 file)
- `/internal/database/models/sales.go`
  - Add: `CancellationDate`, `CancellationNote`, `CancelledBy` fields
  - Add: `CancelSaleRequest`, update `SaleResponse`

### Repositories (2 files)
- `/internal/database/repositories/sales_repo.go`
  - Add: `GetSaleWithItemsAndBatches()` - preload items with batches
  - Add: `UpdateSaleStatusWithTx()` - update within transaction

### Services (2 files)
- `/internal/services/sales_service.go`
  - Add: `CancelSale(saleID, request)` - main cancellation logic
  - Add: `validateSaleCancellation(sale)` - business rules
  - Add: `isValidSaleStatusTransition(from, to)` - state machine

- `/internal/services/inventory_service.go`
  - Update: `validTypes` array to include `"sale_cancellation"`

### Handlers (1 file)
- `/internal/api/handlers/sales_handler.go`
  - Add: `CancelSale(c *gin.Context)` handler

### Routes (1 file)
- `/internal/api/routes/routes.go`
  - Add: `salesGroup.POST("/:id/cancel", ...)`

### Migrations (1 new file)
- `/migrations/YYYYMMDD_add_sale_cancellation_fields.sql`

---

## Core Implementation Logic

```go
// Service: sales_service.go
func (s *SalesService) CancelSale(saleID string, request *models.CancelSaleRequest) (*models.SaleResponse, error) {
    err := s.salesRepo.WithTransaction(func(tx *gorm.DB) error {
        // 1. Load sale with items and batch details
        sale, err := s.salesRepo.GetSaleWithItemsAndBatches(saleID)

        // 2. Validate cancellable (status == "pending")
        if err := s.validateSaleCancellation(sale); err != nil {
            return err
        }

        // 3. Reverse inventory for each sale item
        for _, saleItem := range sale.Items {
            // 3a. Create reverse transaction
            transaction := models.NewInventoryTransaction(
                saleItem.BatchID,
                "sale_cancellation",
                saleItem.Quantity,  // POSITIVE to add back
                &sale.ID,
                &request.CancelledBy,
                &request.Reason,
                time.Now(),
            )
            s.inventoryRepo.CreateTransactionWithTx(tx, transaction)

            // 3b. Update batch quantity
            s.inventoryRepo.UpdateBatchStockWithTx(tx, saleItem.BatchID, saleItem.Quantity)
        }

        // 4. Update sale status
        now := time.Now()
        sale.Status = "cancelled"
        sale.CancellationDate = &now
        sale.CancellationNote = &request.Reason
        sale.CancelledBy = &request.CancelledBy

        return s.salesRepo.UpdateSaleStatusWithTx(tx, sale)
    })

    return s.GetSaleByID(saleID)
}
```

---

## Status Transition State Machine

```
pending ──→ completed
   │
   └──────→ cancelled (terminal)

completed ──→ refunded (future)
```

**Validation**:
- ✅ pending → cancelled (ALLOWED)
- ✅ pending → completed (ALLOWED)
- ❌ completed → cancelled (NOT ALLOWED - use refund flow)
- ❌ cancelled → * (TERMINAL STATE)

---

## Database Changes

```sql
ALTER TABLE sales ADD COLUMN cancellation_date TIMESTAMPTZ;
ALTER TABLE sales ADD COLUMN cancellation_note TEXT;
ALTER TABLE sales ADD COLUMN cancelled_by VARCHAR(100);

CREATE INDEX idx_sales_status_cancellation
ON sales(status, cancellation_date)
WHERE status IN ('cancelled', 'refunded');

ALTER TABLE sales ADD CONSTRAINT chk_cancellation_metadata
CHECK (
    (status = 'cancelled' AND cancellation_date IS NOT NULL) OR
    (status != 'cancelled')
);
```

---

## API Endpoint

**Request**:
```bash
POST /api/v1/sales/{sale_id}/cancel
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "reason": "Customer requested cancellation"
}
```

**Response (200)**:
```json
{
  "id": "SALE_abc123",
  "status": "cancelled",
  "cancellation_date": "2025-11-25T11:00:00Z",
  "cancellation_note": "Customer requested cancellation",
  "cancelled_by": "USER_def456",
  "items": [...],
  ...
}
```

**Errors**:
- 400: Invalid status (not "pending")
- 403: Permission denied
- 404: Sale not found
- 409: Already cancelled

---

## Testing Checklist

### Unit Tests
- [ ] `TestCancelSale_Success`
- [ ] `TestCancelSale_InvalidStatus` (completed sale)
- [ ] `TestCancelSale_AlreadyCancelled`
- [ ] `TestCancelSale_InventoryReturned` (verify batch quantities)
- [ ] `TestCancelSale_MultipleItems` (FEFO reversal)
- [ ] `TestValidateSaleCancellation` (business rules)
- [ ] `TestIsValidSaleStatusTransition` (state machine)

### Integration Tests
- [ ] End-to-end: Create → Cancel → Verify inventory
- [ ] Concurrent cancellations (no race conditions)
- [ ] Transaction rollback on failure
- [ ] FEFO reversal correctness

### Manual Testing
- [ ] Cancel pending sale via API
- [ ] Verify inventory returned to correct batches
- [ ] Check audit trail in inventory_transactions
- [ ] Verify cancellation metadata saved
- [ ] Test permission denied scenario
- [ ] Test invalid status transition

---

## Security & Permissions

**New Permission Required**: `sales:cancel`

**Permission Matrix Update**:
| Role | Current | With Cancellation |
|------|---------|------------------|
| CEO | CRUD | CRUD + Cancel |
| Store Manager | CRUD | CRUD + Cancel |
| Store Staff | CRUD (limited) | Create + Read (no cancel) |
| Auditor | Read | Read |
| Accountant | Read | Read |

**Route Protection**:
```go
salesRoutes.POST("/:id/cancel",
    middleware.RequireOrgPermission("sales", "cancel"),
    salesHandler.CancelSale,
)
```

---

## Existing Patterns to Follow

### 1. Purchase Order Status Updates
**File**: `/internal/services/purchase_order_service.go:315-418`
- ✅ Status transition validation
- ✅ Transaction-based updates
- ✅ State machine pattern

### 2. GRN Inventory Addition
**File**: `/internal/services/purchase_order_service.go:599-653`
- ✅ Batch creation with transactions
- ✅ Inventory transaction logging
- ✅ Error handling

### 3. Sale Creation with FEFO
**File**: `/internal/services/sales_service.go:207-296`
- ✅ Multi-batch allocation
- ✅ Transaction rollback on error
- ✅ Pessimistic locking

---

## Potential Gotchas

1. **FEFO Reversal**: Sale items store batch_id, so reversal is straightforward
2. **Transaction Isolation**: Use GORM's `WithTransaction()` - already handles BEGIN/COMMIT/ROLLBACK
3. **Status Validation**: Add validator BEFORE implementing handler
4. **Discount Usage**: Need to reverse discount usage (not in initial scope, document as future work)
5. **Payment Integration**: Limit to pending/unpaid sales initially

---

## Metrics to Track

**Implementation Metrics**:
- Lines of code: ~300-400 (service + handler + tests)
- Files modified: 7
- New files: 1 (migration)
- Test coverage target: > 80%

**Runtime Metrics**:
- `sale_cancellation_requests_total` (counter)
- `sale_cancellation_duration_seconds` (histogram)
- `inventory_returned_quantity_total` (counter)

---

## Open Questions (Require Design Decisions)

1. Allow cancellation of "completed" sales? → **Suggest: No, use refund flow**
2. Time limit for cancellation? → **Suggest: No limit for pending sales**
3. Partial cancellation support? → **Suggest: No, full cancellation only (Phase 1)**
4. Discount usage reversal? → **Suggest: Phase 2 feature**
5. Approval workflow? → **Suggest: Direct cancellation with permission check**

---

## Next Steps

1. ✅ Read technical assessment (`technical-assessment.md`)
2. ⏳ Review with tech lead
3. ⏳ Create design document with business rules
4. ⏳ Get stakeholder approval
5. ⏳ Create implementation tickets
6. ⏳ Start Phase 1: Database migration

---

## Resources

**Related Documents**:
- Full Technical Assessment: `.kiro/specs/order-cancellation/technical-assessment.md`
- API Contracts: `.kiro/specs/api-contracts/`
- Implementation Summary: `.kiro/specs/IMPLEMENTATION-SUMMARY.md`

**Key Code References**:
- Purchase Order Service: `/internal/services/purchase_order_service.go`
- Sales Service: `/internal/services/sales_service.go`
- Inventory Repository: `/internal/database/repositories/inventory_repo.go`

**Architecture Patterns**:
- Hexagonal architecture (handlers → services → repositories)
- GORM transaction management
- AAA service integration for permissions
- Organization-scoped access control
