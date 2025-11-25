# ADR-001: Order Cancellation Inventory Return Design

## Status

Proposed

## Date

2024-11-25

## Context

The FPO ERP system currently handles:
1. **Sale Creation**: Deducts inventory from batches using FEFO (First Expired, First Out)
2. **Returns**: Allows customers to return items AFTER delivery

However, there is no mechanism to:
- Cancel an order before it ships
- Return inventory to the original batch when an order is cancelled
- Track cancellation history for audit purposes
- Handle partial cancellations (cancel some items, keep others)

### Current System Behavior

When a sale is created:
1. Inventory is deducted from batches (negative `InventoryTransaction`)
2. Discount usage is incremented
3. Tax summary is created
4. Sale status is set to "pending"

The only way to "undo" a sale is through the Returns flow, which:
- Creates a new `Return` record
- Creates positive inventory transactions
- Does NOT reverse discounts or taxes
- Is designed for post-delivery scenarios

### Business Requirements

1. Allow cancellation of orders that haven't shipped
2. Restore inventory to original batches
3. Reverse discount usage counts (so limits work correctly)
4. Void tax records for cancelled sales
5. Maintain complete audit trail
6. Support partial cancellation (cancel some items, keep others)

## Decision Drivers

1. **Data Integrity**: Inventory counts must be accurate
2. **Financial Accuracy**: Cancelled sales should not affect revenue/tax reports
3. **Auditability**: Complete history of what happened
4. **Idempotency**: Safe to retry failed cancellation operations
5. **Performance**: Cancellation should not block other operations
6. **Backward Compatibility**: Existing APIs and flows unchanged

## Considered Options

### Option 1: Reuse Returns System

**Description**: Extend the existing Returns system to handle cancellations.

**Pros**:
- Less code to write
- Single system for all "reversals"

**Cons**:
- Returns is designed for post-delivery scenarios
- Would conflate two different business operations
- Returns don't reverse discounts
- Returns create refund records, not just status changes
- Harder to query/report on cancellations vs returns

**Decision**: Rejected - Different business semantics

### Option 2: Simple Status Change

**Description**: Just update sale status to "cancelled", no inventory handling.

**Pros**:
- Very simple implementation
- Quick to develop

**Cons**:
- Inventory becomes permanently out of sync
- No audit trail
- Discounts remain used
- Tax records remain
- Violates data integrity invariants

**Decision**: Rejected - Violates core requirements

### Option 3: Soft Delete with Reversal (Selected)

**Description**: Create dedicated cancellation system that:
- Creates explicit cancellation records
- Creates reversal inventory transactions
- Updates discount and tax records
- Maintains full audit trail

**Pros**:
- Clear separation of concerns
- Complete audit trail
- Maintains data integrity
- Supports both full and partial cancellation
- Easy to query and report

**Cons**:
- More complex implementation
- New tables and models required
- More code to maintain

**Decision**: Selected - Best balance of requirements

### Option 4: Event Sourcing

**Description**: Implement full event sourcing where cancellation is just another event.

**Pros**:
- Ultimate auditability
- Can replay and rebuild state
- Flexible for future needs

**Cons**:
- Major architectural change
- Overkill for current requirements
- Significant development effort
- Team unfamiliar with pattern

**Decision**: Rejected - Over-engineering for current needs

## Decision

Implement **Option 3: Soft Delete with Reversal** with the following architecture:

### New Data Model

```
Sale 1 ---> * SaleItem
  |
  +-------> * SaleCancellation 1 ---> * SaleCancellationItem
```

### New Transaction Type

Add `cancellation_return` to inventory transaction types.

### State Machine

```
pending -> confirmed -> processing -> shipped -> delivered
    |          |            |
    v          v            v
cancelled  cancelled    cancelled (with conditions)
```

### Transaction Flow

1. Lock sale record (pessimistic locking)
2. Validate cancellation eligibility
3. Create SaleCancellation record
4. For each item:
   - Create SaleCancellationItem
   - Create InventoryTransaction (cancellation_return, +quantity)
   - Update batch stock
5. Reverse discount usage
6. Void tax summary
7. Update sale status
8. Commit transaction

### Error Handling

- All operations within single database transaction
- Full rollback on any failure
- Idempotency via unique constraints
- Detailed error codes for each failure mode

## Consequences

### Positive

1. **Data Integrity Maintained**: Inventory always accurate
2. **Financial Accuracy**: Clear separation of active vs cancelled sales
3. **Auditability**: Complete history of cancellations
4. **Flexibility**: Supports both full and partial cancellation
5. **Reportability**: Easy to query cancellation metrics
6. **Backward Compatible**: No changes to existing APIs

### Negative

1. **Increased Complexity**: New tables, models, services
2. **Migration Required**: Database schema changes
3. **More Testing**: Additional test cases needed
4. **Learning Curve**: Team needs to understand new flow

### Neutral

1. **Discount System Changes**: Need to add reversal capability
2. **Tax System Changes**: Need to add voiding capability
3. **Reporting Impact**: New reports possible/required

## Technical Debt Created

1. The existing `DeleteSale` method does NOT restore inventory - should we deprecate it?
2. Returns system has similar but different inventory handling - should they share code?
3. Discount reversal could be useful for Returns too - potential future refactoring

## Compliance

- **Audit Requirements**: Met by SaleCancellation and SaleCancellationItem records
- **Data Retention**: Cancellation records are never deleted, only soft-deleted if needed
- **Financial Reporting**: Cancelled sales clearly distinguishable from completed sales

## References

- [Design Document](./design.md)
- [Task Breakdown](./tasks.md)
- [Sales Context API](../api-contracts/sales-context-api.md)
- [Inventory List API](../api-contracts/inventory-list-api.md)

## Notes

This ADR was created as part of the Order Cancellation Inventory Return System feature request.

The decision prioritizes data integrity and auditability over simplicity, as these are critical for a financial/inventory system.
