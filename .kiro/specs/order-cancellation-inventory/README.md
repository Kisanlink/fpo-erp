# Order Cancellation Inventory Return System

## Specification Overview

This folder contains the complete design specification for implementing order (sale) cancellation functionality with proper inventory return handling.

## Documents

| Document | Purpose |
|----------|---------|
| [design.md](./design.md) | Complete technical design document |
| [tasks.md](./tasks.md) | Implementation task breakdown |
| [adr-001-cancellation-design.md](./adr-001-cancellation-design.md) | Architecture Decision Record |

## Feature Summary

### Capabilities

1. **Full Order Cancellation**: Cancel entire order, return all inventory
2. **Partial Item Cancellation**: Cancel specific items, keep others active
3. **Inventory Restoration**: Stock returned to original batches
4. **Discount Reversal**: Usage counts decremented, limits restored
5. **Tax Voiding**: Tax records marked as voided
6. **Audit Trail**: Complete history of all cancellations

### API Endpoints

```
POST /api/v1/sales/{id}/cancel          # Cancel full order
POST /api/v1/sales/{id}/cancel-items    # Cancel specific items
GET  /api/v1/sales/{id}/cancellations   # Get cancellation history
```

### New Data Models

- `SaleCancellation` - Tracks cancellation events
- `SaleCancellationItem` - Tracks individual cancelled items

### New Transaction Type

- `cancellation_return` - Inventory transaction for restored stock

## Implementation Timeline

| Phase | Duration | Focus |
|-------|----------|-------|
| Phase 1 | Week 1 | Database & Models |
| Phase 2 | Week 1-2 | Core Logic |
| Phase 3 | Week 2 | Partial Cancellation |
| Phase 4 | Week 2-3 | API Layer |
| Phase 5 | Week 3 | Testing & Observability |
| Phase 6 | Week 3 | Documentation & Deployment |

**Total Estimated Effort**: ~90 hours (~2.5 weeks)

## Key Design Decisions

1. **Separate from Returns**: Cancellation is distinct from customer returns
2. **Atomic Transactions**: All cancellation operations in single DB transaction
3. **Pessimistic Locking**: Prevents race conditions on concurrent cancellations
4. **Soft Status Updates**: Sales marked as cancelled, not deleted
5. **Complete Audit Trail**: All actions recorded with timestamps and user IDs

## Cancellation Rules

| Current Status | Can Cancel? |
|----------------|-------------|
| pending | Yes |
| confirmed | Yes |
| processing | Yes (with conditions) |
| shipped | No (use Returns) |
| delivered | No (use Returns) |
| cancelled | No (already cancelled) |

## Dependencies

This feature depends on:
- Existing `Sale` and `SaleItem` models
- Existing `InventoryBatch` and `InventoryTransaction` models
- Existing `Discount` and `DiscountUsage` models
- Existing `TaxSummary` model

## Quick Start for Implementation

1. Read [design.md](./design.md) for full context
2. Review [tasks.md](./tasks.md) for task list
3. Start with Phase 1 tasks (CANC-001 through CANC-006)
4. Each task is designed to be a single commit

## Related Documentation

- [Sales Context API](../api-contracts/sales-context-api.md)
- [Inventory List API](../api-contracts/inventory-list-api.md)
- [Implementation Summary](../IMPLEMENTATION-SUMMARY.md)
