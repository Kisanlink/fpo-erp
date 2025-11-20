# Inventory List API Contract

## Overview

**Purpose**: Eliminate N+1 query problem by providing inventory list with complete context (product, variant, warehouse, pricing) in a single paginated response.

**Current Problem**:
- Fetching 100 batches requires 1 + 200+ API calls (N+1 pattern)
- Sequential fetches for product names, warehouse details, prices
- 2-4 second page load time

**Solution Impact**:
- **95%+ reduction** in API calls (200+ → 1)
- **2-4 seconds** faster page loads
- Single paginated query with all context

---

## API Specification

### Endpoint: Get Inventory with Context

```
GET /api/v1/inventory/batches/list
```

**Authentication**: Required
**Authorization**: `inventory_batch:read` permission

#### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `warehouse_id` | string | No | all | Filter by warehouse |
| `variant_id` | string | No | all | Filter by variant |
| `product_id` | string | No | all | Filter by product |
| `category` | string | No | all | Filter by product category |
| `in_stock_only` | boolean | No | true | Show only batches with stock > 0 |
| `expiring_soon` | boolean | No | false | Show batches expiring within 30 days |
| `low_stock_threshold` | integer | No | null | Show batches below threshold |
| `include` | string[] | No | all | `variant,product,warehouse,prices,taxes` |
| `sort_by` | string | No | expiry_date | Sort field: `expiry_date`, `quantity`, `cost_price` |
| `sort_order` | string | No | asc | `asc` or `desc` |
| `limit` | integer | No | 50 | Results per page (max 200) |
| `offset` | integer | No | 0 | Pagination offset |

#### Request Example

```http
GET /api/v1/inventory/batches/list?warehouse_id=WH_001&in_stock_only=true&limit=50&offset=0
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### Response Schema

```json
{
  "batches": [
    {
      "id": "BTCH_001",
      "warehouse": {
        "id": "WH_001",
        "name": "Main Warehouse - Delhi",
        "location": {
          "city": "New Delhi",
          "state": "Delhi"
        }
      },
      "variant": {
        "id": "PVAR_001",
        "variant_name": "Premium Basmati Rice - 1kg Pack",
        "sku": "PBR-1KG-001",
        "barcode": "8901234567890",
        "brand_name": "Golden Harvest",
        "quantity": "1kg",
        "pack_size": "Standard Pack",
        "images": ["https://cdn.example.com/products/rice-1kg-front.jpg"],
        "hsn_code": "1006"
      },
      "product": {
        "id": "PROD_12345",
        "name": "Premium Basmati Rice",
        "category": "Grains",
        "description": "Aged premium basmati rice"
      },
      "quantity_details": {
        "total_quantity": 3000,
        "available_quantity": 2850,
        "reserved_quantity": 150,
        "sold_quantity": 0,
        "in_stock": true
      },
      "pricing": {
        "cost_price": 55.00,
        "selling_prices": {
          "retail": 85.00,
          "wholesale": 75.00,
          "bulk": 70.00
        },
        "margin": {
          "retail_margin": 30.00,
          "retail_margin_percentage": 54.55
        },
        "currency": "INR"
      },
      "batch_info": {
        "batch_number": "BATCH-2024-12-001",
        "manufacturing_date": "2024-12-01",
        "expiry_date": "2025-12-31",
        "days_until_expiry": 405,
        "expiry_status": "good"
      },
      "tax_config": {
        "cgst_rate": 2.5,
        "sgst_rate": 2.5,
        "total_gst_rate": 5.0,
        "is_tax_exempt": false
      },
      "metadata": {
        "created_at": "2024-11-15T10:15:00Z",
        "updated_at": "2024-11-20T14:30:00Z",
        "created_by": "user_warehouse_manager_001"
      }
    }
  ],

  "pagination": {
    "total": 234,
    "limit": 50,
    "offset": 0,
    "has_more": true,
    "next_offset": 50
  },

  "summary": {
    "total_batches": 234,
    "total_products": 156,
    "total_variants": 234,
    "total_stock_quantity": 125000,
    "total_stock_value": 6875000.00,
    "expiring_soon_count": 15,
    "low_stock_count": 8,
    "zero_stock_count": 0
  },

  "metadata": {
    "read_timestamp": "2024-11-21T10:30:00Z",
    "filters_applied": {
      "warehouse_id": "WH_001",
      "in_stock_only": true
    }
  }
}
```

---

## Business Rules

### Expiry Status Calculation

```javascript
const daysUntilExpiry = (expiryDate - today) / (1000 * 60 * 60 * 24);

if (daysUntilExpiry < 0) return "expired";
if (daysUntilExpiry <= 30) return "critical";
if (daysUntilExpiry <= 90) return "warning";
return "good";
```

### Low Stock Detection

- Configurable threshold per variant
- `low_stock_count` = batches below threshold
- Alert icon in UI for low stock items

### Reserved Quantity

- Stock allocated to pending sales
- `available_quantity = total_quantity - reserved_quantity`
- Updated real-time as sales are created/completed

---

## Performance

- **Target P95**: < 250ms for 50 records
- **Database**: Single query with LATERAL JOINs
- **Indexes Required**:
  ```sql
  CREATE INDEX idx_inv_batch_warehouse_qty ON inventory_batches(warehouse_id, total_quantity DESC);
  CREATE INDEX idx_inv_batch_expiry ON inventory_batches(expiry_date ASC) WHERE total_quantity > 0;
  CREATE INDEX idx_product_variants_active ON product_variants(product_id, is_active);
  ```

---

## Use Cases

### Use Case 1: Inventory Manager Dashboard

**Old Flow** (N+1 queries):
```
GET /inventory/batches → 100 batches
GET /products/:id (x 100) → Product details
GET /warehouses/:id (x 50) → Warehouse details
GET /prices?variant_id (x 100) → Pricing
```
**Total**: 250+ API calls

**New Flow**:
```
GET /inventory/batches/list?limit=100&include=all
```
**Total**: 1 API call

**Benefit**: 99.6% reduction in API calls

---

### Use Case 2: Expiring Stock Report

**Request**:
```
GET /inventory/batches/list?expiring_soon=true&sort_by=expiry_date&sort_order=asc
```

**Response**: All batches expiring within 30 days, sorted by expiry date

---

### Use Case 3: Low Stock Alert

**Request**:
```
GET /inventory/batches/list?low_stock_threshold=100&in_stock_only=true
```

**Response**: All batches with quantity below 100 units

---

## Pagination Best Practices

### Cursor-Based Pagination (Future)

Instead of offset-based, use cursor for better performance:

```
GET /inventory/batches/list?after_cursor=BTCH_050&limit=50
```

### Response Size Management

- Default limit: 50 records
- Max limit: 200 records
- Typical response size: 100-300KB
- Large response size: Up to 2MB (with images)

---

## Related Contracts

- [Aggregated Product API](./aggregated-product-api.md)
- [Sales Context API](./sales-context-api.md)
- [Optional Includes Pattern](./optional-includes-pattern.md)
