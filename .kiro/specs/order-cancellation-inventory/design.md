# Order Cancellation Inventory Return System

## Design Document

**Version**: 1.0
**Date**: 2024-11-25
**Author**: Backend Architecture Team
**Status**: Draft

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Problem Statement](#problem-statement)
3. [System Analysis](#system-analysis)
4. [Solution Design](#solution-design)
5. [API Specification](#api-specification)
6. [Data Models](#data-models)
7. [Business Rules](#business-rules)
8. [Transaction Flow](#transaction-flow)
9. [Security Considerations](#security-considerations)
10. [Observability](#observability)
11. [Testing Strategy](#testing-strategy)
12. [Implementation Tasks](#implementation-tasks)
13. [Risk Assessment](#risk-assessment)
14. [Rollback Strategy](#rollback-strategy)

---

## 1. Executive Summary

This document specifies the design for implementing order (sale) cancellation functionality that properly returns inventory to the warehouse. The system must handle:

- Full order cancellations
- Partial item cancellations
- Status-based cancellation eligibility
- Atomic inventory restoration
- Discount/tax reversal handling
- Complete audit trail

**Key Invariants**:
- Inventory integrity: `total_deducted = total_sold + total_returned_via_cancellation + total_returned_via_return`
- Financial integrity: Cancelled sales must reverse all associated financial records
- Idempotency: Cancellation operations must be idempotent to handle retries safely

---

## 2. Problem Statement

### Current State

The existing codebase has:
- **Sale creation** (`CreateSale` in `sales_service.go`): Deducts inventory using transactions within a DB transaction
- **Returns system** (`returns_service.go`): Handles customer returns AFTER a sale is completed
- **Missing**: Dedicated cancellation flow that returns inventory before fulfillment

### Gap Analysis

| Capability | Current State | Required State |
|------------|---------------|----------------|
| Full order cancellation | Not implemented | Atomic cancellation with inventory return |
| Partial cancellation | Not implemented | Cancel specific items while keeping others |
| Status validation | Basic status field | State machine with transition rules |
| Inventory return | Only via Returns | Dedicated cancellation path |
| Audit trail | Basic logs | Comprehensive cancellation audit |
| Discount reversal | Not implemented | Decrement usage, restore limits |

---

## 3. System Analysis

### Existing Data Flow (Sale Creation)

```
CreateSaleRequest
    |
    v
+-------------------+
| Validate Request  |
+-------------------+
    |
    v
+-------------------+
| Fetch Batches     |
| (FEFO ordered)    |
+-------------------+
    |
    v
+-------------------+
| Begin Transaction |
+-------------------+
    |
    +---> Create Sale record
    |
    +---> For each item:
    |       - Create SaleItem
    |       - Create InventoryTransaction (type: "sale", quantity: -N)
    |       - UpdateBatchStock (decrement)
    |
    +---> Apply discounts
    |       - Create DiscountUsage
    |       - Increment usage count
    |
    +---> Apply taxes
    |       - Create TaxSummary
    |
    +---> Update Sale total
    |
+-------------------+
| Commit Transaction|
+-------------------+
```

### Existing Inventory Transaction Types

From `inventory_service.go`:
```go
validTypes := []string{
    "import",
    "manual_add",
    "adjustment",
    "sale_deduction",  // Not currently used (using "sale" instead)
    "return_add",
    "transfer_in",
    "transfer_out"
}
```

Current usage in `sales_service.go`:
```go
transaction := models.NewInventoryTransaction(batch.ID, "sale", -quantityFromBatch, &sale.ID, nil, stringPtr("Sale transaction"), time.Now())
```

### Sale Status Values

Current implicit statuses (not formally defined):
- `pending` - Sale created, awaiting processing
- `completed` - Sale fulfilled
- (Missing: `cancelled`, `partially_cancelled`)

---

## 4. Solution Design

### 4.1 Sale Status State Machine

```
                   +------------+
                   |  pending   |
                   +------------+
                        |
         +--------------+--------------+
         |              |              |
         v              v              v
   +-----------+  +-----------+  +-----------+
   | confirmed |  | cancelled |  |  failed   |
   +-----------+  +-----------+  +-----------+
         |
         v
   +-----------+
   | processing|
   +-----------+
         |
    +----+----+
    |         |
    v         v
+-----------+ +-----------+
| shipped   | | cancelled |
+-----------+ +-----------+
    |
    v
+-----------+
| delivered |
+-----------+
    |
    +---> (Cannot cancel delivered orders)
    |
    v
+-----------+
| returned  | (via Returns flow, not cancellation)
+-----------+
```

### 4.2 Cancellation Eligibility Rules

| Current Status | Can Cancel? | Notes |
|----------------|-------------|-------|
| pending | Yes | Full inventory return |
| confirmed | Yes | Full inventory return |
| processing | Yes (with conditions) | May require warehouse verification |
| shipped | No | Use Returns flow instead |
| delivered | No | Use Returns flow instead |
| cancelled | No | Already cancelled |
| returned | No | Already processed via Returns |

### 4.3 Cancellation Types

1. **Full Cancellation**: Cancel entire order
   - All items returned to inventory
   - Sale status -> `cancelled`
   - All discounts reversed
   - All tax records voided

2. **Partial Cancellation**: Cancel specific items
   - Selected items returned to inventory
   - Sale remains active with reduced total
   - Discounts recalculated
   - Taxes recalculated

### 4.4 New Transaction Type

Add new inventory transaction type:
```go
"cancellation_return" // Stock returned due to order cancellation
```

---

## 5. API Specification

### 5.1 Cancel Full Order

```
POST /api/v1/sales/{id}/cancel
```

**Request Body**:
```json
{
    "reason": "customer_request",
    "reason_details": "Customer changed their mind",
    "skip_inventory_return": false,
    "performed_by": "USER_12345"
}
```

**Reason Codes** (enum):
- `customer_request` - Customer initiated cancellation
- `payment_failed` - Payment could not be processed
- `out_of_stock` - Stock unavailable after sale creation
- `pricing_error` - Incorrect pricing applied
- `duplicate_order` - Order was duplicated
- `fraud_suspected` - Fraud detection triggered
- `system_error` - System error during processing
- `other` - Other reason (requires `reason_details`)

**Response** (200 OK):
```json
{
    "status": "success",
    "data": {
        "sale": {
            "id": "SALE_12345678",
            "status": "cancelled",
            "cancelled_at": "2024-11-25T10:30:00Z",
            "cancellation_reason": "customer_request",
            "cancellation_details": "Customer changed their mind"
        },
        "inventory_restored": [
            {
                "batch_id": "BTCH_001",
                "variant_id": "PVAR_001",
                "quantity_restored": 50,
                "transaction_id": "TXN_87654321"
            }
        ],
        "financial_adjustments": {
            "discount_reversed": {
                "discount_id": "DISC_001",
                "amount_reversed": 25.00,
                "usage_decremented": true
            },
            "tax_voided": {
                "tax_summary_id": "TXSM_001",
                "amount_voided": 42.50
            }
        },
        "audit_id": "AUD_CANCEL_001"
    }
}
```

**Error Responses**:

| Status Code | Error Code | Description |
|-------------|------------|-------------|
| 400 | INVALID_REASON | Invalid cancellation reason |
| 404 | SALE_NOT_FOUND | Sale does not exist |
| 409 | SALE_NOT_CANCELLABLE | Sale status does not allow cancellation |
| 409 | SALE_ALREADY_CANCELLED | Sale is already cancelled |
| 422 | PARTIAL_SHIPMENT | Some items already shipped |
| 500 | INVENTORY_RESTORE_FAILED | Failed to restore inventory |

### 5.2 Cancel Specific Items (Partial Cancellation)

```
POST /api/v1/sales/{id}/cancel-items
```

**Request Body**:
```json
{
    "items": [
        {
            "sale_item_id": "SITEM_001",
            "quantity_to_cancel": 10,
            "reason": "customer_request"
        },
        {
            "sale_item_id": "SITEM_002",
            "quantity_to_cancel": 5,
            "reason": "out_of_stock"
        }
    ],
    "reason_details": "Partial cancellation per customer request",
    "performed_by": "USER_12345"
}
```

**Response** (200 OK):
```json
{
    "status": "success",
    "data": {
        "sale": {
            "id": "SALE_12345678",
            "status": "partially_cancelled",
            "original_total": 1500.00,
            "new_total": 850.00,
            "cancelled_amount": 650.00
        },
        "cancelled_items": [
            {
                "sale_item_id": "SITEM_001",
                "quantity_cancelled": 10,
                "inventory_restored": true,
                "batch_id": "BTCH_001",
                "transaction_id": "TXN_001"
            }
        ],
        "remaining_items": [
            {
                "sale_item_id": "SITEM_003",
                "quantity": 20,
                "status": "active"
            }
        ],
        "financial_adjustments": {
            "discount_recalculated": true,
            "new_discount_amount": 15.00,
            "tax_recalculated": true,
            "new_tax_amount": 35.00
        }
    }
}
```

### 5.3 Get Cancellation History

```
GET /api/v1/sales/{id}/cancellations
```

**Response**:
```json
{
    "status": "success",
    "data": {
        "sale_id": "SALE_12345678",
        "cancellations": [
            {
                "id": "CANC_001",
                "type": "partial",
                "cancelled_at": "2024-11-25T09:00:00Z",
                "cancelled_by": "USER_12345",
                "reason": "customer_request",
                "items_affected": 2,
                "amount_cancelled": 250.00,
                "inventory_transactions": ["TXN_001", "TXN_002"]
            }
        ]
    }
}
```

---

## 6. Data Models

### 6.1 New Model: SaleCancellation

```go
// SaleCancellation tracks cancellation events for sales
type SaleCancellation struct {
    base.BaseModel
    SaleID            string     `gorm:"type:varchar(100);not null;index" json:"sale_id"`
    CancellationType  string     `gorm:"type:varchar(20);not null" json:"cancellation_type"` // full, partial
    CancelledBy       *string    `gorm:"type:varchar(100)" json:"cancelled_by"`
    Reason            string     `gorm:"type:varchar(50);not null" json:"reason"`
    ReasonDetails     *string    `gorm:"type:text" json:"reason_details"`
    CancelledAt       time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"cancelled_at"`

    // Financial impact
    OriginalAmount    float64    `gorm:"type:numeric(14,4);not null" json:"original_amount"`
    CancelledAmount   float64    `gorm:"type:numeric(14,4);not null" json:"cancelled_amount"`
    DiscountReversed  float64    `gorm:"type:numeric(14,4);default:0" json:"discount_reversed"`
    TaxReversed       float64    `gorm:"type:numeric(14,4);default:0" json:"tax_reversed"`

    // Associations
    Sale              Sale                     `gorm:"foreignKey:SaleID" json:"sale,omitempty"`
    Items             []SaleCancellationItem   `gorm:"foreignKey:CancellationID" json:"items,omitempty"`
}

func (SaleCancellation) TableName() string {
    return "sale_cancellations"
}
```

### 6.2 New Model: SaleCancellationItem

```go
// SaleCancellationItem tracks individual items in a cancellation
type SaleCancellationItem struct {
    base.BaseModel
    CancellationID      string  `gorm:"type:varchar(100);not null;index" json:"cancellation_id"`
    SaleItemID          string  `gorm:"type:varchar(100);not null" json:"sale_item_id"`
    BatchID             string  `gorm:"type:varchar(100);not null" json:"batch_id"`
    QuantityCancelled   int64   `gorm:"type:bigint;not null" json:"quantity_cancelled"`
    RefundAmount        float64 `gorm:"type:numeric(14,4);not null" json:"refund_amount"`
    InventoryRestored   bool    `gorm:"type:boolean;not null;default:true" json:"inventory_restored"`
    TransactionID       *string `gorm:"type:varchar(100)" json:"transaction_id"` // Inventory transaction ID

    // Associations
    Cancellation        SaleCancellation `gorm:"foreignKey:CancellationID" json:"cancellation,omitempty"`
    SaleItem            SaleItem         `gorm:"foreignKey:SaleItemID" json:"sale_item,omitempty"`
    Batch               InventoryBatch   `gorm:"foreignKey:BatchID" json:"batch,omitempty"`
}

func (SaleCancellationItem) TableName() string {
    return "sale_cancellation_items"
}
```

### 6.3 Updated Sale Model

Add fields to existing `Sale` model:

```go
// Additional fields for Sale model
CancelledAt       *time.Time `gorm:"type:timestamptz" json:"cancelled_at,omitempty"`
CancellationReason *string   `gorm:"type:varchar(50)" json:"cancellation_reason,omitempty"`
```

### 6.4 Database Migration

```sql
-- Migration: create_sale_cancellation_tables
-- Version: 20241125000001

-- Add cancellation fields to sales table
ALTER TABLE sales ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ;
ALTER TABLE sales ADD COLUMN IF NOT EXISTS cancellation_reason VARCHAR(50);

-- Create sale_cancellations table
CREATE TABLE IF NOT EXISTS sale_cancellations (
    id VARCHAR(100) PRIMARY KEY,
    sale_id VARCHAR(100) NOT NULL REFERENCES sales(id),
    cancellation_type VARCHAR(20) NOT NULL CHECK (cancellation_type IN ('full', 'partial')),
    cancelled_by VARCHAR(100),
    reason VARCHAR(50) NOT NULL,
    reason_details TEXT,
    cancelled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    original_amount NUMERIC(14,4) NOT NULL,
    cancelled_amount NUMERIC(14,4) NOT NULL,
    discount_reversed NUMERIC(14,4) DEFAULT 0,
    tax_reversed NUMERIC(14,4) DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create sale_cancellation_items table
CREATE TABLE IF NOT EXISTS sale_cancellation_items (
    id VARCHAR(100) PRIMARY KEY,
    cancellation_id VARCHAR(100) NOT NULL REFERENCES sale_cancellations(id),
    sale_item_id VARCHAR(100) NOT NULL REFERENCES sale_items(id),
    batch_id VARCHAR(100) NOT NULL REFERENCES inventory_batches(id),
    quantity_cancelled BIGINT NOT NULL CHECK (quantity_cancelled > 0),
    refund_amount NUMERIC(14,4) NOT NULL,
    inventory_restored BOOLEAN NOT NULL DEFAULT TRUE,
    transaction_id VARCHAR(100) REFERENCES inventory_transactions(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_sale_cancellations_sale_id ON sale_cancellations(sale_id);
CREATE INDEX IF NOT EXISTS idx_sale_cancellations_cancelled_at ON sale_cancellations(cancelled_at DESC);
CREATE INDEX IF NOT EXISTS idx_sale_cancellation_items_cancellation_id ON sale_cancellation_items(cancellation_id);
CREATE INDEX IF NOT EXISTS idx_sale_cancellation_items_sale_item_id ON sale_cancellation_items(sale_item_id);
CREATE INDEX IF NOT EXISTS idx_sales_status ON sales(status) WHERE status IN ('cancelled', 'partially_cancelled');
```

---

## 7. Business Rules

### 7.1 Cancellation Eligibility

```go
func (s *SalesService) CanCancelSale(sale *models.Sale) (bool, string) {
    switch sale.Status {
    case "cancelled":
        return false, "Sale is already cancelled"
    case "shipped":
        return false, "Cannot cancel shipped orders. Use Returns instead."
    case "delivered":
        return false, "Cannot cancel delivered orders. Use Returns instead."
    case "returned":
        return false, "Sale has already been returned"
    case "pending", "confirmed", "processing":
        return true, ""
    default:
        return false, "Unknown sale status"
    }
}
```

### 7.2 Partial Cancellation Rules

1. Cannot cancel more quantity than originally ordered
2. Cannot cancel more quantity than remaining (after previous cancellations)
3. Must have at least one active item remaining (otherwise use full cancellation)
4. Discounts must be recalculated based on remaining items
5. Tax must be recalculated based on new total

### 7.3 Inventory Return Rules

1. Stock is returned to the ORIGINAL batch (maintains cost/expiry tracking)
2. If original batch no longer exists, create adjustment transaction
3. Inventory transaction must reference the cancellation ID
4. Transaction type: `cancellation_return`

### 7.4 Discount Reversal Rules

```go
func (s *SalesService) reverseDiscounts(tx *gorm.DB, saleID string) error {
    // Get all discount usages for this sale
    usages, err := s.discountsRepo.GetUsagesBySaleID(saleID)
    if err != nil {
        return err
    }

    for _, usage := range usages {
        // Decrement usage count on the discount
        if err := s.discountsRepo.DecrementUsageWithTx(tx, usage.DiscountID); err != nil {
            return err
        }

        // Mark usage as reversed
        usage.IsReversed = true
        usage.ReversedAt = time.Now()
        if err := s.discountsRepo.UpdateUsageWithTx(tx, &usage); err != nil {
            return err
        }
    }

    return nil
}
```

### 7.5 Tax Reversal Rules

1. Tax summary must be marked as voided
2. Original tax amounts preserved for audit
3. New field: `is_voided` on TaxSummary

---

## 8. Transaction Flow

### 8.1 Full Cancellation Flow

```
CancelSaleRequest
    |
    v
+-------------------------+
| 1. Validate Sale Exists |
+-------------------------+
    |
    v
+-------------------------+
| 2. Check Cancellability |
|    (status validation)  |
+-------------------------+
    |
    v
+-------------------------+
| 3. Begin Transaction    |
+-------------------------+
    |
    +---> 4. Lock Sale record (SELECT FOR UPDATE)
    |
    +---> 5. Create SaleCancellation record
    |
    +---> 6. For each SaleItem:
    |       a. Create SaleCancellationItem
    |       b. Create InventoryTransaction (type: "cancellation_return")
    |       c. Update batch stock (increment)
    |
    +---> 7. Reverse discounts
    |       a. Decrement usage counts
    |       b. Mark usages as reversed
    |
    +---> 8. Void tax records
    |       a. Mark TaxSummary as voided
    |
    +---> 9. Update Sale
    |       a. status = "cancelled"
    |       b. cancelled_at = now()
    |       c. cancellation_reason = reason
    |
+-------------------------+
| 10. Commit Transaction  |
+-------------------------+
    |
    v
+-------------------------+
| 11. Emit Events         |
|     (async, non-blocking)|
+-------------------------+
    |
    v
Return CancellationResponse
```

### 8.2 Partial Cancellation Flow

```
CancelItemsRequest
    |
    v
+-------------------------+
| 1. Validate Sale Exists |
+-------------------------+
    |
    v
+-------------------------+
| 2. Validate Items       |
|    - Items belong to sale
|    - Quantities valid   |
+-------------------------+
    |
    v
+-------------------------+
| 3. Begin Transaction    |
+-------------------------+
    |
    +---> 4. Lock Sale record
    |
    +---> 5. Create SaleCancellation record (type: partial)
    |
    +---> 6. For each item to cancel:
    |       a. Create SaleCancellationItem
    |       b. Create InventoryTransaction
    |       c. Update batch stock
    |       d. Update SaleItem cancelled_quantity
    |
    +---> 7. Recalculate discounts
    |       a. Get remaining items value
    |       b. Check if discounts still apply
    |       c. Adjust discount amounts
    |
    +---> 8. Recalculate taxes
    |       a. Calculate new tax on remaining
    |       b. Update TaxSummary
    |
    +---> 9. Update Sale
    |       a. total_amount = new total
    |       b. status = "partially_cancelled" (if items remain)
    |
+-------------------------+
| 10. Commit Transaction  |
+-------------------------+
    |
    v
Return PartialCancellationResponse
```

### 8.3 Concurrency Handling

```go
// Use pessimistic locking to prevent race conditions
func (r *SalesRepository) GetSaleForUpdateWithTx(tx *gorm.DB, id string) (*models.Sale, error) {
    var sale models.Sale
    if err := tx.Set("gorm:query_option", "FOR UPDATE").
        Preload("Items").
        Where("id = ?", id).
        First(&sale).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, errors.NewNotFoundError("Sale")
        }
        return nil, errors.NewInternalServerError("Failed to lock sale")
    }
    return &sale, nil
}
```

---

## 9. Security Considerations

### 9.1 Authorization

| Endpoint | Required Permission |
|----------|---------------------|
| POST /sales/{id}/cancel | `sale:cancel` |
| POST /sales/{id}/cancel-items | `sale:cancel` |
| GET /sales/{id}/cancellations | `sale:read` |

### 9.2 Role-Based Access

| Role | Can Cancel? | Restrictions |
|------|-------------|--------------|
| CEO | Yes | No restrictions |
| Store_Manager | Yes | Own store only |
| Store_Staff | Yes (with approval) | Requires manager approval for > threshold |
| Accountant | No | Read-only access to cancellation history |
| Auditor | No | Read-only access to cancellation history |

### 9.3 Fraud Prevention

1. **Rate limiting**: Max 10 cancellations per user per hour
2. **Amount threshold**: Cancellations > INR 50,000 require manager approval
3. **Audit logging**: All cancellations logged with IP, user agent, timestamp
4. **Anomaly detection**: Flag users with high cancellation rate

### 9.4 Input Validation

```go
type CancelSaleRequest struct {
    Reason        CancellationReason `json:"reason" binding:"required,oneof=customer_request payment_failed out_of_stock pricing_error duplicate_order fraud_suspected system_error other"`
    ReasonDetails *string            `json:"reason_details" binding:"required_if=Reason other,max=1000"`
    SkipInventoryReturn bool         `json:"skip_inventory_return"`
    PerformedBy   string             `json:"performed_by" binding:"required"`
}
```

---

## 10. Observability

### 10.1 Metrics

```go
// Prometheus metrics to track
var (
    cancellationTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "sale_cancellations_total",
            Help: "Total number of sale cancellations",
        },
        []string{"type", "reason", "status"},
    )

    cancellationAmountTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "sale_cancellation_amount_total",
            Help: "Total amount of cancelled sales",
        },
        []string{"warehouse_id", "reason"},
    )

    cancellationDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "sale_cancellation_duration_seconds",
            Help:    "Time taken to process cancellation",
            Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5},
        },
        []string{"type"},
    )

    inventoryRestoreFailures = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "inventory_restore_failures_total",
            Help: "Number of failed inventory restore operations",
        },
    )
)
```

### 10.2 Structured Logging

```go
s.logger.Info("Sale cancellation initiated",
    zap.String("sale_id", saleID),
    zap.String("cancellation_type", "full"),
    zap.String("reason", string(req.Reason)),
    zap.String("performed_by", req.PerformedBy),
    zap.Float64("sale_amount", sale.TotalAmount),
    zap.Int("item_count", len(sale.Items)),
)

s.logger.Info("Sale cancellation completed",
    zap.String("sale_id", saleID),
    zap.String("cancellation_id", cancellation.ID),
    zap.Duration("duration", time.Since(startTime)),
    zap.Int("inventory_transactions_created", len(transactions)),
    zap.Float64("discount_reversed", discountReversed),
    zap.Float64("tax_reversed", taxReversed),
)
```

### 10.3 Alerting Rules

```yaml
# Prometheus alerting rules
groups:
  - name: sale_cancellation_alerts
    rules:
      - alert: HighCancellationRate
        expr: rate(sale_cancellations_total[1h]) > 50
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High sale cancellation rate detected"

      - alert: InventoryRestoreFailure
        expr: increase(inventory_restore_failures_total[5m]) > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Inventory restore failed during cancellation"

      - alert: CancellationLatencyHigh
        expr: histogram_quantile(0.95, rate(sale_cancellation_duration_seconds_bucket[5m])) > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Sale cancellation P95 latency above 5 seconds"
```

---

## 11. Testing Strategy

### 11.1 Unit Tests

```go
// File: internal/services/sales_cancellation_service_test.go

func TestCancelSale_Success(t *testing.T) {
    // Setup mock repositories
    // Create sale with items
    // Call CancelSale
    // Assert: Sale status = cancelled
    // Assert: Inventory transactions created
    // Assert: Batch quantities restored
    // Assert: Discounts reversed
}

func TestCancelSale_AlreadyCancelled(t *testing.T) {
    // Setup sale with status = cancelled
    // Call CancelSale
    // Assert: Error SALE_ALREADY_CANCELLED
}

func TestCancelSale_ShippedOrder(t *testing.T) {
    // Setup sale with status = shipped
    // Call CancelSale
    // Assert: Error SALE_NOT_CANCELLABLE
}

func TestPartialCancellation_Success(t *testing.T) {
    // Setup sale with 3 items
    // Cancel 1 item
    // Assert: Sale still active
    // Assert: Total recalculated
    // Assert: Correct item returned to inventory
}

func TestPartialCancellation_ExceedsQuantity(t *testing.T) {
    // Setup sale with item quantity = 10
    // Try to cancel quantity = 15
    // Assert: Validation error
}

func TestCancellation_InventoryIntegrity(t *testing.T) {
    // Create sale that deducts inventory
    // Cancel sale
    // Assert: Original batch quantity restored exactly
}
```

### 11.2 Integration Tests

```go
// File: tests/integration/cancellation_test.go

func TestCancellationE2E_FullFlow(t *testing.T) {
    // 1. Create sale via API
    // 2. Verify inventory deducted
    // 3. Cancel sale via API
    // 4. Verify inventory restored
    // 5. Verify cancellation record created
    // 6. Verify discount usage decremented
}

func TestCancellationE2E_ConcurrentCancellation(t *testing.T) {
    // 1. Create sale
    // 2. Attempt concurrent cancellations
    // 3. Assert: Only one succeeds
    // 4. Assert: No duplicate inventory returns
}

func TestCancellationE2E_TransactionRollback(t *testing.T) {
    // 1. Create sale
    // 2. Mock inventory update to fail
    // 3. Attempt cancellation
    // 4. Assert: Sale status unchanged
    // 5. Assert: No partial cancellation records
}
```

### 11.3 Load Tests

```go
// Test cancellation under load
func TestCancellation_HighVolume(t *testing.T) {
    // Create 100 sales
    // Cancel 50 concurrently
    // Assert: All 50 cancelled successfully
    // Assert: No inventory discrepancies
    // Assert: P95 latency < 500ms
}
```

### 11.4 Business Logic Tests

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| Cancel pending sale | Sale in pending status | Status = cancelled, inventory restored |
| Cancel with FEFO batches | Sale from multiple batches | All batches restored correctly |
| Discount threshold breach | Partial cancel below min order | Discount removed |
| Multi-item partial | Cancel 2 of 5 items | Correct 2 returned, 3 remain |

---

## 12. Implementation Tasks

### Phase 1: Foundation (Week 1)

| Task ID | Task | Estimate | Dependencies |
|---------|------|----------|--------------|
| CANC-001 | Create database migration for new tables | 2h | None |
| CANC-002 | Implement SaleCancellation model | 2h | CANC-001 |
| CANC-003 | Implement SaleCancellationItem model | 2h | CANC-002 |
| CANC-004 | Add cancellation repository methods | 4h | CANC-003 |
| CANC-005 | Add new inventory transaction type | 1h | None |
| CANC-006 | Update Sale model with cancellation fields | 1h | CANC-001 |

### Phase 2: Core Logic (Week 1-2)

| Task ID | Task | Estimate | Dependencies |
|---------|------|----------|--------------|
| CANC-007 | Implement CanCancelSale validation | 2h | CANC-006 |
| CANC-008 | Implement CancelSale service method | 6h | CANC-004, CANC-007 |
| CANC-009 | Implement inventory restore logic | 4h | CANC-005 |
| CANC-010 | Implement discount reversal logic | 3h | CANC-008 |
| CANC-011 | Implement tax voiding logic | 2h | CANC-008 |
| CANC-012 | Unit tests for cancellation service | 4h | CANC-008 |

### Phase 3: Partial Cancellation (Week 2)

| Task ID | Task | Estimate | Dependencies |
|---------|------|----------|--------------|
| CANC-013 | Implement CancelItems service method | 6h | CANC-008 |
| CANC-014 | Implement discount recalculation | 4h | CANC-013 |
| CANC-015 | Implement tax recalculation | 3h | CANC-013 |
| CANC-016 | Unit tests for partial cancellation | 4h | CANC-013 |

### Phase 4: API Layer (Week 2-3)

| Task ID | Task | Estimate | Dependencies |
|---------|------|----------|--------------|
| CANC-017 | Implement CancelSale handler | 3h | CANC-008 |
| CANC-018 | Implement CancelItems handler | 3h | CANC-013 |
| CANC-019 | Implement GetCancellations handler | 2h | CANC-004 |
| CANC-020 | Add routes and middleware | 2h | CANC-017-019 |
| CANC-021 | Add Swagger documentation | 2h | CANC-020 |

### Phase 5: Observability & Testing (Week 3)

| Task ID | Task | Estimate | Dependencies |
|---------|------|----------|--------------|
| CANC-022 | Add Prometheus metrics | 3h | CANC-017, CANC-018 |
| CANC-023 | Add structured logging | 2h | CANC-008, CANC-013 |
| CANC-024 | Integration tests | 6h | CANC-020 |
| CANC-025 | Load tests | 4h | CANC-024 |
| CANC-026 | Security review | 2h | CANC-020 |

### Phase 6: Documentation & Deployment (Week 3)

| Task ID | Task | Estimate | Dependencies |
|---------|------|----------|--------------|
| CANC-027 | Update API documentation | 2h | CANC-021 |
| CANC-028 | Create runbook | 2h | CANC-022 |
| CANC-029 | Staging deployment & testing | 4h | All |
| CANC-030 | Production deployment | 2h | CANC-029 |

**Total Estimated Effort**: ~80 hours (2 weeks with buffer)

---

## 13. Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Inventory inconsistency | Medium | Critical | Atomic transactions, row locking, reconciliation job |
| Discount over-reversal | Low | High | Idempotency keys, usage tracking |
| Performance degradation | Medium | Medium | Batch processing, async operations where safe |
| Concurrent cancellation race | Medium | High | Pessimistic locking, unique constraints |
| Partial failure mid-transaction | Low | Critical | Full transaction rollback, retry mechanism |

---

## 14. Rollback Strategy

### Database Rollback

```sql
-- Rollback migration if needed
ALTER TABLE sales DROP COLUMN IF EXISTS cancelled_at;
ALTER TABLE sales DROP COLUMN IF EXISTS cancellation_reason;
DROP TABLE IF EXISTS sale_cancellation_items;
DROP TABLE IF EXISTS sale_cancellations;
```

### Feature Flag

```go
const FeatureCancellation = "feature.sale.cancellation"

func (s *SalesService) CancelSale(req *CancelSaleRequest) (*CancelSaleResponse, error) {
    if !config.IsFeatureEnabled(FeatureCancellation) {
        return nil, errors.NewBadRequestError("Cancellation feature is disabled")
    }
    // ...
}
```

### Rollback Procedure

1. Disable feature flag
2. Monitor for in-flight cancellations to complete
3. Roll back API deployment
4. (Optional) Roll back database migration if no data
5. Investigate issue
6. Re-deploy with fix

---

## Appendix A: Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| SALE_NOT_FOUND | 404 | Sale with given ID does not exist |
| SALE_NOT_CANCELLABLE | 409 | Sale status does not allow cancellation |
| SALE_ALREADY_CANCELLED | 409 | Sale has already been cancelled |
| INVALID_REASON | 400 | Cancellation reason is invalid |
| INVALID_QUANTITY | 400 | Quantity to cancel exceeds available |
| ITEM_NOT_FOUND | 404 | Sale item does not exist |
| INVENTORY_RESTORE_FAILED | 500 | Failed to restore inventory |
| DISCOUNT_REVERSAL_FAILED | 500 | Failed to reverse discount |
| TAX_VOID_FAILED | 500 | Failed to void tax records |

---

## Appendix B: Related Documents

- [Sales Context API](../api-contracts/sales-context-api.md)
- [Inventory List API](../api-contracts/inventory-list-api.md)
- [Implementation Summary](../IMPLEMENTATION-SUMMARY.md)

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2024-11-25 | Backend Architecture | Initial design |
