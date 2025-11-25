# Order Cancellation Flow Diagrams

---

## 1. High-Level Flow

```
┌─────────────┐
│   CLIENT    │
│  (Frontend) │
└─────┬───────┘
      │
      │ POST /api/v1/sales/{id}/cancel
      │ { "reason": "..." }
      │
      ▼
┌─────────────────────────────────────┐
│     SALES HANDLER                   │
│  ─────────────────────────────────  │
│  1. Extract user_id from JWT        │
│  2. Validate request                │
│  3. Check permissions               │
└─────────┬───────────────────────────┘
          │
          │ CancelSale(saleID, request)
          │
          ▼
┌─────────────────────────────────────┐
│     SALES SERVICE                   │
│  ─────────────────────────────────  │
│  1. Load sale with items & batches  │
│  2. Validate cancellable            │
│  3. Start DB transaction            │
│     ├─ Reverse inventory            │
│     ├─ Update sale status           │
│     └─ Commit transaction           │
└─────────┬───────────────────────────┘
          │
          │ Success: SaleResponse
          │
          ▼
┌─────────────┐
│   CLIENT    │
│  (Frontend) │
└─────────────┘
```

---

## 2. Detailed Service Layer Flow

```
┌──────────────────────────────────────────────────────────────┐
│                     CancelSale Method                        │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────┐
        │  1. Load Sale with Items & Batches  │
        │     GetSaleWithItemsAndBatches()    │
        └─────────────────┬───────────────────┘
                          │
                          ▼
        ┌─────────────────────────────────────┐
        │  2. Validate Cancellation Rules     │
        │     - Status must be "pending"      │
        │     - Not already cancelled         │
        │     - Time limits (if any)          │
        └─────────────────┬───────────────────┘
                          │
                    ✓ Valid │ ✗ Invalid
                          │
                          ├────────────────────► Error Response
                          │                      (400 Bad Request)
                          ▼
        ┌─────────────────────────────────────┐
        │  3. Start Database Transaction      │
        │     salesRepo.WithTransaction()     │
        └─────────────────┬───────────────────┘
                          │
                          ▼
        ┌─────────────────────────────────────┐
        │  4. For Each Sale Item:             │
        │                                     │
        │  ┌───────────────────────────────┐ │
        │  │ Item 1: Batch A (Qty: 10)     │ │
        │  │   ├─ Create Inventory Txn     │ │
        │  │   │  (type: sale_cancellation)│ │
        │  │   │  (change: +10)            │ │
        │  │   └─ Update Batch Stock       │ │
        │  │      (TotalQuantity += 10)    │ │
        │  └───────────────────────────────┘ │
        │                                     │
        │  ┌───────────────────────────────┐ │
        │  │ Item 2: Batch B (Qty: 5)      │ │
        │  │   ├─ Create Inventory Txn     │ │
        │  │   └─ Update Batch Stock       │ │
        │  └───────────────────────────────┘ │
        │                                     │
        │  ┌───────────────────────────────┐ │
        │  │ Item 2: Batch C (Qty: 3)      │ │
        │  │   ├─ Create Inventory Txn     │ │
        │  │   └─ Update Batch Stock       │ │
        │  └───────────────────────────────┘ │
        │                                     │
        └─────────────────┬───────────────────┘
                          │
                          ▼
        ┌─────────────────────────────────────┐
        │  5. Update Sale Record              │
        │     - status = "cancelled"          │
        │     - cancellation_date = now       │
        │     - cancellation_note = reason    │
        │     - cancelled_by = user_id        │
        └─────────────────┬───────────────────┘
                          │
                          ▼
        ┌─────────────────────────────────────┐
        │  6. Commit Transaction              │
        └─────────────────┬───────────────────┘
                          │
                    Success │ Failure
                          │
                          ├────────────────────► Rollback
                          │                      All Changes
                          │                      Return Error
                          ▼
        ┌─────────────────────────────────────┐
        │  7. Return Updated Sale Response    │
        └─────────────────────────────────────┘
```

---

## 3. Database Transaction Flow

```
┌─────────────────────────────────────────────────────────────┐
│                  PostgreSQL Transaction                     │
└─────────────────────────────────────────────────────────────┘

   BEGIN TRANSACTION
      │
      │
      ├─── SELECT * FROM sales WHERE id = ?                (Row Lock)
      │    JOIN sale_items ON sale_items.sale_id = sales.id
      │    JOIN inventory_batches ON inventory_batches.id = sale_items.batch_id
      │
      ├─── Validation: Check status = "pending"
      │
      │
      ├─── FOR EACH sale_item:
      │       │
      │       ├─── INSERT INTO inventory_transactions (      (Write)
      │       │        batch_id = sale_item.batch_id,
      │       │        transaction_type = "sale_cancellation",
      │       │        quantity_change = +sale_item.quantity,
      │       │        related_entity_id = sale.id,
      │       │        performed_by = user_id,
      │       │        note = cancellation_reason
      │       │    )
      │       │
      │       └─── UPDATE inventory_batches                  (Write)
      │            SET total_quantity = total_quantity + sale_item.quantity
      │            WHERE id = sale_item.batch_id
      │
      │
      ├─── UPDATE sales SET                                  (Write)
      │        status = "cancelled",
      │        cancellation_date = NOW(),
      │        cancellation_note = reason,
      │        cancelled_by = user_id
      │    WHERE id = sale.id
      │
      │
   COMMIT
      │
      ▼
   ┌─────────────────────────────────────────┐
   │  Database State Updated Atomically      │
   │  ✓ Sale status = cancelled              │
   │  ✓ Inventory restored to batches        │
   │  ✓ Audit trail in transactions table    │
   └─────────────────────────────────────────┘
```

---

## 4. FEFO Reversal Example

### Scenario: Sale with 18 units allocated across 3 batches

```
BEFORE SALE:
┌──────────────────────────────────────────────────┐
│  Batch A: Expiry 2025-12-01 │ Quantity: 10      │
│  Batch B: Expiry 2025-12-15 │ Quantity: 20      │
│  Batch C: Expiry 2025-12-31 │ Quantity: 15      │
└──────────────────────────────────────────────────┘

DURING SALE (FEFO allocation):
Customer orders 18 units
  ├─ Allocate 10 from Batch A (oldest)
  └─ Allocate 8 from Batch B (second oldest)

AFTER SALE:
┌──────────────────────────────────────────────────┐
│  Batch A: Expiry 2025-12-01 │ Quantity: 0       │  ◄─ Fully consumed
│  Batch B: Expiry 2025-12-15 │ Quantity: 12      │  ◄─ Partially consumed
│  Batch C: Expiry 2025-12-31 │ Quantity: 15      │  ◄─ Untouched
└──────────────────────────────────────────────────┘

Sale Items Created:
┌────────────────────────────────────────┐
│ Sale Item 1: Batch A, Qty: 10         │
│ Sale Item 2: Batch B, Qty: 8          │
└────────────────────────────────────────┘

CANCELLATION (Reverse allocation):
  ├─ Return 10 to Batch A
  └─ Return 8 to Batch B

AFTER CANCELLATION:
┌──────────────────────────────────────────────────┐
│  Batch A: Expiry 2025-12-01 │ Quantity: 10      │  ◄─ Restored
│  Batch B: Expiry 2025-12-15 │ Quantity: 20      │  ◄─ Restored
│  Batch C: Expiry 2025-12-31 │ Quantity: 15      │  ◄─ Unchanged
└──────────────────────────────────────────────────┘

✓ Inventory fully restored to original state
✓ Audit trail maintained in inventory_transactions table
```

---

## 5. Status State Machine

```
                    ┌─────────────┐
                    │  PENDING    │  ◄──── Initial state after sale creation
                    └──────┬──────┘
                           │
                ┏━━━━━━━━━━┻━━━━━━━━━━┓
                ▼                      ▼
         ┌─────────────┐        ┌─────────────┐
         │  COMPLETED  │        │  CANCELLED  │  ◄──── Terminal state
         └──────┬──────┘        └─────────────┘
                │
                │ (Future feature)
                ▼
         ┌─────────────┐
         │  REFUNDED   │  ◄──── Terminal state
         └─────────────┘


VALID TRANSITIONS:
✓ pending → completed      (Normal sale flow)
✓ pending → cancelled      (Cancellation - THIS FEATURE)
✓ completed → refunded     (Refund flow - Future)

INVALID TRANSITIONS:
✗ completed → cancelled    (Use refund flow instead)
✗ cancelled → pending      (Terminal state - cannot revert)
✗ cancelled → completed    (Terminal state - cannot revert)
✗ refunded → *            (Terminal state - cannot revert)
```

---

## 6. Error Handling Flow

```
┌────────────────────────────────────────────┐
│         CancelSale Request                 │
└────────────────┬───────────────────────────┘
                 │
                 ▼
        ┌─────────────────┐
        │  Validate Input │
        └────────┬─────────┘
                 │
         ✗ Invalid
      ┌──────────┴──────────┐
      │ 400 Bad Request     │
      │ "Invalid sale ID"   │
      └─────────────────────┘
                 │
                 ▼ ✓ Valid
        ┌─────────────────┐
        │   Load Sale     │
        └────────┬─────────┘
                 │
         ✗ Not Found
      ┌──────────┴──────────┐
      │ 404 Not Found       │
      │ "Sale not found"    │
      └─────────────────────┘
                 │
                 ▼ ✓ Found
        ┌─────────────────┐
        │ Check Permission│
        └────────┬─────────┘
                 │
         ✗ Denied
      ┌──────────┴──────────┐
      │ 403 Forbidden       │
      │ "Permission denied" │
      └─────────────────────┘
                 │
                 ▼ ✓ Allowed
        ┌─────────────────┐
        │ Validate Status │
        └────────┬─────────┘
                 │
         ✗ Not "pending"
      ┌──────────┴──────────┐
      │ 400 Bad Request     │
      │ "Only pending sales │
      │  can be cancelled"  │
      └─────────────────────┘
                 │
                 ▼ ✓ Valid
        ┌─────────────────┐
        │ Already Cancelled│
        └────────┬─────────┘
                 │
         ✗ Yes
      ┌──────────┴──────────┐
      │ 409 Conflict        │
      │ "Already cancelled" │
      └─────────────────────┘
                 │
                 ▼ ✓ Not cancelled
        ┌─────────────────────┐
        │ Execute Transaction │
        └────────┬────────────┘
                 │
         ✗ DB Error
      ┌──────────┴──────────┐
      │ 500 Internal Error  │
      │ (Rollback complete) │
      └─────────────────────┘
                 │
                 ▼ ✓ Success
        ┌─────────────────────┐
        │ 200 OK              │
        │ Return cancelled    │
        │ sale response       │
        └─────────────────────┘
```

---

## 7. Concurrent Cancellation Handling

```
Two users try to cancel the same sale simultaneously:

User A                          User B
  │                               │
  │  POST /sales/123/cancel       │
  ├────────────────────────────►  │
  │                               │  POST /sales/123/cancel
  │                               ├────────────────────►
  │                               │
  ▼                               ▼
┌─────────────────┐         ┌─────────────────┐
│ Start Txn (A)   │         │ Start Txn (B)   │
│ Lock sale row   │         │ Wait for lock   │
└────────┬────────┘         └────────┬────────┘
         │                           │
         │                           │ (Blocked)
         │                           │
         ▼                           │
┌─────────────────┐                  │
│ Check status    │                  │
│ status="pending"│                  │
└────────┬────────┘                  │
         │                           │
         ▼                           │
┌─────────────────┐                  │
│ Update status   │                  │
│ to "cancelled"  │                  │
└────────┬────────┘                  │
         │                           │
         ▼                           │
┌─────────────────┐                  │
│ Commit Txn (A)  │                  │
│ Release lock    │                  │
└────────┬────────┘                  │
         │                           │
         ▼                           ▼
      Success                 ┌─────────────────┐
         │                    │ Acquire lock    │
         │                    │ Check status    │
         │                    │ status="cancelled"│
         │                    └────────┬────────┘
         │                             │
         │                             ▼
         │                    ┌─────────────────┐
         │                    │ Validation Fails│
         │                    │ 409 Conflict    │
         │                    │ "Already        │
         │                    │  cancelled"     │
         │                    └─────────────────┘
         │                             │
         │                             ▼
         ▼                          Failure
    200 OK                         409 Conflict

✓ No race condition: Database row-level locking prevents double cancellation
✓ Second request gets clear error message
✓ Inventory is only adjusted once
```

---

## 8. Comparison: Sale Creation vs Cancellation

```
┌────────────────────────────────────────────────────────────────┐
│                    SALE CREATION (Existing)                    │
└────────────────────────────────────────────────────────────────┘

For each item:
  ├─ Create sale_item record
  ├─ Create inventory_transaction (type: "sale", change: -qty)
  └─ Update inventory_batch (total_quantity -= qty)

Result: Inventory DEDUCTED from batches

┌────────────────────────────────────────────────────────────────┐
│                 SALE CANCELLATION (New Feature)                │
└────────────────────────────────────────────────────────────────┘

For each existing sale_item:
  ├─ sale_item record unchanged (kept for audit)
  ├─ Create inventory_transaction (type: "sale_cancellation", change: +qty)
  └─ Update inventory_batch (total_quantity += qty)

Result: Inventory RETURNED to batches

┌────────────────────────────────────────────────────────────────┐
│                        KEY DIFFERENCES                         │
└────────────────────────────────────────────────────────────────┘

Sale Creation:
  - Creates NEW sale_items
  - Negative quantity change (-qty)
  - Multiple batches allocated via FEFO
  - Status: "pending"

Sale Cancellation:
  - Reads EXISTING sale_items
  - Positive quantity change (+qty)
  - Returns to SAME batches (via batch_id in sale_item)
  - Status: "pending" → "cancelled"

Common Pattern:
  ✓ Both use database transactions
  ✓ Both create inventory_transactions for audit
  ✓ Both update batch total_quantity
  ✓ Both are atomic (all-or-nothing)
```

---

## Legend

```
┌────────┐
│ Box    │  = Process/Step
└────────┘

  │
  ▼        = Flow direction

  ├─────►  = Conditional branch

  ✓        = Success/Valid
  ✗        = Failure/Invalid

┏━━━━━┓
┃     ┃   = Decision point
┗━━━━━┛
```

---

**Note**: These diagrams illustrate the complete flow. For implementation details, refer to `technical-assessment.md`.
