# Discount System Architecture

## Overview
The Kisanlink ERP discount system has been designed for external e-commerce integration compatibility by removing customer dependency and simplifying tracking to sale-based only.

## Key Architectural Decision: Sale-Based Discount Tracking

### Previous Design (Removed)
- DiscountUsage had `customer_id` field
- Enabled per-customer usage limits and tracking
- Created dependency on customer management system

### Current Design (Production)
- DiscountUsage tracks per-sale only (`sale_id` + `discount_id`)
- No customer dependency whatsoever
- Simplified integration with external e-commerce platforms
- Removes anti-abuse tracking complexity

## Database Schema

### DiscountUsage Table
```sql
CREATE TABLE discount_usages (
    id VARCHAR(100) PRIMARY KEY,
    discount_id VARCHAR(100) NOT NULL REFERENCES discounts(id),
    sale_id VARCHAR(100) NOT NULL REFERENCES sales(id),
    used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    amount NUMERIC(10,4) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Migration Applied
- Removed `customer_id` column from `discount_usages` table
- Dropped associated indexes and constraints
- No data migration needed (existing usage data remains valid per-sale)

## API Response Model

```json
{
  "id": "DISC_USE_12345678",
  "discount_id": "DISC_12345678",
  "sale_id": "SALE_12345678",
  "used_at": "2024-01-01T00:00:00Z",
  "amount": 10.50,
  "created_at": "2024-01-01T00:00:00Z"
}
```

## Benefits

### 1. External E-commerce Integration
- No customer management dependency
- E-commerce platforms can apply discounts without syncing customer data
- Simplified webhook integration

### 2. System Simplification
- Reduced complexity in discount validation
- Fewer database queries and joins
- Cleaner API contracts

### 3. Scalability
- No per-customer usage limits to track across distributed systems
- Sale-based tracking scales with transaction volume
- Reduced storage and indexing requirements

## Implementation Status

- ✅ **Database Schema**: Updated with migration script
- ✅ **Models**: DiscountUsage and DiscountUsageResponse models updated
- ✅ **API Documentation**: Auto-generated documentation reflects changes
- ✅ **Repository Code**: No customer_id queries remain
- ✅ **Service Layer**: Discount application logic updated

## Future Considerations

If per-customer limits are needed in the future, they can be implemented at the application layer by:
1. Querying sales by external customer identifier
2. Aggregating discount usage across those sales
3. Enforcing limits during discount validation

This approach maintains the decoupled architecture while providing usage controls when needed.