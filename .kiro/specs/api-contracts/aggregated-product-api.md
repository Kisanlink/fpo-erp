# Aggregated Product API Contract

## Overview

**Purpose**: Provide complete product information in a single API call, eliminating the need for multiple sequential requests.

**Current Problem**:
- Product detail page requires 4+ API calls
- Sequential waterfall loading (400-800ms total)
- Multiple loading spinners affect UX

**Solution Impact**:
- **75% reduction** in API calls (4 → 1)
- **300-600ms** faster page loads
- **Single loading state** instead of 4 spinners

---

## API Specification

### Endpoint 1: Get Product Detail with Aggregated Data

```
GET /api/v1/products/{product_id}/detail
```

**Description**: Retrieves complete product information including all variants, prices, inventory availability, and collaborator details in a single call.

**Authentication**: Required (Bearer Token)

**Authorization**: Requires `product:read` permission

#### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `product_id` | string | Yes | Unique product identifier |

#### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `include` | string[] | No | all | Comma-separated list: `variants,prices,inventory,collaborators,taxes` |
| `warehouse_id` | string | No | all | Filter inventory by specific warehouse |
| `price_type` | string | No | all | Filter prices: `retail`, `wholesale`, `bulk`, or `all` |
| `active_only` | boolean | No | true | Show only active variants |
| `in_stock_only` | boolean | No | false | Show only variants with available stock |

#### Request Example

```http
GET /api/v1/products/PROD_12345/detail?include=variants,prices,inventory&warehouse_id=WH_001
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### Response Schema

```json
{
  "product": {
    "id": "PROD_12345",
    "name": "Premium Basmati Rice",
    "description": "Aged premium basmati rice from Punjab region",
    "category": "Grains",
    "organization_id": "ORG_001",
    "is_active": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-11-20T15:45:00Z"
  },
  "collaborator": {
    "id": "CLAB_789",
    "company_name": "Punjab Rice Mills",
    "contact_person": "Rajesh Kumar",
    "phone": "+91-9876543210",
    "email": "contact@punjabrice.com",
    "address": {
      "street": "Industrial Area, Phase 2",
      "city": "Ludhiana",
      "state": "Punjab",
      "pincode": "141003",
      "country": "India"
    },
    "is_active": true
  },
  "variants": [
    {
      "id": "PVAR_001",
      "product_id": "PROD_12345",
      "variant_name": "Premium Basmati Rice - 1kg Pack",
      "description": "1kg premium packaging suitable for retail",
      "sku": "PBR-1KG-001",
      "barcode": "8901234567890",
      "external_id": "SHOPIFY_VAR_12345",
      "brand_name": "Golden Harvest",
      "quantity": "1kg",
      "pack_size": "Standard Pack",
      "images": [
        "https://cdn.example.com/products/rice-1kg-front.jpg",
        "https://cdn.example.com/products/rice-1kg-back.jpg"
      ],
      "hsn_code": "1006",
      "gst_rate": 5.0,
      "dosage_instructions": null,
      "usage_details": "Store in cool, dry place. Best before 12 months from packaging",
      "is_active": true,
      "created_at": "2024-01-15T10:35:00Z",
      "updated_at": "2024-11-20T15:45:00Z",

      "prices": {
        "currency": "INR",
        "has_active_price": true,
        "retail_price": {
          "price": 85.00,
          "effective_from": "2024-11-01T00:00:00Z",
          "effective_to": null
        },
        "wholesale_price": {
          "price": 75.00,
          "effective_from": "2024-11-01T00:00:00Z",
          "effective_to": null
        },
        "bulk_price": {
          "price": 70.00,
          "effective_from": "2024-11-01T00:00:00Z",
          "effective_to": "2024-12-31T23:59:59Z"
        }
      },

      "stock_summary": {
        "total_quantity": 5000,
        "available_quantity": 4800,
        "reserved_quantity": 200,
        "in_stock": true,
        "warehouse_count": 3,
        "min_cost_price": 55.00,
        "max_cost_price": 58.00,
        "earliest_expiry": "2025-12-31"
      },

      "warehouse_stock": [
        {
          "warehouse_id": "WH_001",
          "warehouse_name": "Main Warehouse - Delhi",
          "quantity": 3000,
          "cost_price": 55.00,
          "expiry_date": "2025-12-31",
          "batch_count": 2
        },
        {
          "warehouse_id": "WH_002",
          "warehouse_name": "Branch Warehouse - Mumbai",
          "quantity": 1500,
          "cost_price": 56.50,
          "expiry_date": "2026-01-15",
          "batch_count": 1
        },
        {
          "warehouse_id": "WH_003",
          "warehouse_name": "Regional Hub - Bangalore",
          "quantity": 500,
          "cost_price": 58.00,
          "expiry_date": "2025-12-31",
          "batch_count": 1
        }
      ],

      "tax_configuration": {
        "cgst_rate": 2.5,
        "sgst_rate": 2.5,
        "is_tax_exempt": false,
        "custom_tax_ids": []
      }
    },
    {
      "id": "PVAR_002",
      "product_id": "PROD_12345",
      "variant_name": "Premium Basmati Rice - 5kg Pack",
      "description": "5kg bulk packaging",
      "sku": "PBR-5KG-001",
      "barcode": "8901234567891",
      "external_id": "SHOPIFY_VAR_12346",
      "brand_name": "Golden Harvest",
      "quantity": "5kg",
      "pack_size": "Bulk Pack",
      "images": [
        "https://cdn.example.com/products/rice-5kg-front.jpg"
      ],
      "hsn_code": "1006",
      "gst_rate": 5.0,
      "dosage_instructions": null,
      "usage_details": "Store in cool, dry place. Best before 12 months from packaging",
      "is_active": true,
      "created_at": "2024-01-15T10:36:00Z",
      "updated_at": "2024-11-20T15:45:00Z",

      "prices": {
        "currency": "INR",
        "has_active_price": true,
        "retail_price": {
          "price": 400.00,
          "effective_from": "2024-11-01T00:00:00Z",
          "effective_to": null
        },
        "wholesale_price": {
          "price": 360.00,
          "effective_from": "2024-11-01T00:00:00Z",
          "effective_to": null
        },
        "bulk_price": null
      },

      "stock_summary": {
        "total_quantity": 2000,
        "available_quantity": 1950,
        "reserved_quantity": 50,
        "in_stock": true,
        "warehouse_count": 2,
        "min_cost_price": 270.00,
        "max_cost_price": 275.00,
        "earliest_expiry": "2025-11-30"
      },

      "warehouse_stock": [
        {
          "warehouse_id": "WH_001",
          "warehouse_name": "Main Warehouse - Delhi",
          "quantity": 1500,
          "cost_price": 270.00,
          "expiry_date": "2025-11-30",
          "batch_count": 1
        },
        {
          "warehouse_id": "WH_002",
          "warehouse_name": "Branch Warehouse - Mumbai",
          "quantity": 500,
          "cost_price": 275.00,
          "expiry_date": "2026-01-31",
          "batch_count": 1
        }
      ],

      "tax_configuration": {
        "cgst_rate": 2.5,
        "sgst_rate": 2.5,
        "is_tax_exempt": false,
        "custom_tax_ids": []
      }
    }
  ],

  "metadata": {
    "total_variants": 2,
    "active_variants": 2,
    "total_stock_value": 324500.00,
    "read_timestamp": "2024-11-21T10:30:00Z",
    "consistency_token": "CT_abc123def456"
  }
}
```

#### Response Status Codes

| Status Code | Description |
|-------------|-------------|
| 200 OK | Product details retrieved successfully |
| 400 Bad Request | Invalid query parameters |
| 401 Unauthorized | Authentication token missing or invalid |
| 403 Forbidden | User lacks required permissions |
| 404 Not Found | Product with specified ID not found |
| 500 Internal Server Error | Server-side error occurred |

#### Error Response Schema

```json
{
  "status": "error",
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "message": "Product with ID PROD_12345 not found",
    "details": {
      "product_id": "PROD_12345",
      "organization_id": "ORG_001"
    },
    "timestamp": "2024-11-21T10:30:00Z",
    "request_id": "req_xyz789"
  }
}
```

---

### Endpoint 2: Get Variant Detail (Single Variant View)

```
GET /api/v1/products/variants/{variant_id}/detail
```

**Description**: Retrieves complete information for a specific product variant.

#### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `variant_id` | string | Yes | Unique variant identifier |

#### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `include` | string[] | No | all | `prices,inventory,product,collaborator,taxes` |
| `warehouse_id` | string | No | all | Filter inventory by warehouse |

#### Response Example

```json
{
  "variant": {
    "id": "PVAR_001",
    "product_id": "PROD_12345",
    "variant_name": "Premium Basmati Rice - 1kg Pack",
    "sku": "PBR-1KG-001",
    "barcode": "8901234567890",
    "brand_name": "Golden Harvest",
    "quantity": "1kg",
    "pack_size": "Standard Pack",
    "images": ["..."],
    "is_active": true,

    "product": {
      "id": "PROD_12345",
      "name": "Premium Basmati Rice",
      "category": "Grains"
    },

    "collaborator": {
      "id": "CLAB_789",
      "company_name": "Punjab Rice Mills"
    },

    "prices": { "..." },
    "stock_summary": { "..." },
    "warehouse_stock": [ "..." ],
    "tax_configuration": { "..." }
  }
}
```

---

## Business Rules & Constraints

### Data Consistency

1. **Read Consistency**: All data is fetched within a single database transaction to ensure consistency
2. **Consistency Token**: Response includes a token that can be used for optimistic locking
3. **Timestamp**: `read_timestamp` indicates the point-in-time for the data snapshot

### Authorization

1. **Organization Isolation**: Users can only access products within their organization
2. **Role-Based Filtering**: Different roles see different fields:
   - **CEO/Manager**: Sees cost prices, margins, full data
   - **Store Staff**: Sees selling prices, stock, no cost prices
   - **Auditor**: Read-only access to all data
3. **Permission Cascade**: Requires `product:read` permission. Optional data (prices, inventory) requires additional permissions

### Price Logic

1. **Active Prices Only**: Only prices within `effective_from` to `effective_to` range are returned
2. **Multiple Price Types**: Retail, wholesale, and bulk prices can coexist
3. **Currency Consistency**: All prices for a variant use the same currency
4. **No Price Fallback**: If no active price exists, `has_active_price` is false and price fields are null

### Inventory Logic

1. **FEFO Compliance**: Earliest expiry date is shown for stock planning
2. **Reserved Quantity**: Deducted from available quantity (e.g., in pending sales)
3. **Warehouse Aggregation**: Stock summary aggregates across all warehouses
4. **Zero Stock Filtering**: `in_stock_only=true` excludes variants with zero total_quantity

### Tax Configuration

1. **Batch-Level Taxes**: Tax rates are averaged across batches
2. **CGST + SGST**: Combined rates should equal GST rate
3. **Custom Taxes**: Additional tax IDs for special levies

---

## Performance Characteristics

### Response Time Targets

- **P50**: < 100ms
- **P95**: < 250ms
- **P99**: < 500ms

### Caching Strategy

- **Product & Variant Metadata**: Cache TTL 5 minutes
- **Prices**: Cache TTL 1 minute (dynamic)
- **Inventory**: No cache (real-time)
- **Cache Key**: `product:detail:{product_id}:{include}:{warehouse_id}`

### Database Optimization

1. **Single Query**: Uses LATERAL JOINs to fetch all data in one query
2. **Indexes Required**:
   - `product_variants(product_id, is_active)`
   - `product_prices(variant_id, is_active, price_type)`
   - `inventory_batches(variant_id, warehouse_id, total_quantity)`
3. **Connection Pooling**: Read replicas for heavy aggregation queries

---

## Migration Strategy

### Backward Compatibility

**Existing Endpoints Maintained**:
- `GET /api/v1/products/{id}` - Returns only product data
- `GET /api/v1/product-variants?product_id={id}` - Returns only variants
- `GET /api/v1/prices?variant_id={id}` - Returns only prices
- `GET /api/v1/inventory/batches?variant_id={id}` - Returns only inventory

### Migration Path

**Phase 1**: Deploy new aggregated endpoint alongside existing endpoints
**Phase 2**: Update frontend to use new endpoint for product detail pages
**Phase 3**: Monitor usage, optimize based on real-world patterns
**Phase 4**: Deprecate old endpoints after 6 months (with warnings)

### Rollback Plan

If issues arise:
1. Feature flag to disable aggregated endpoint
2. Frontend automatically falls back to individual API calls
3. No data migration needed (backward compatible)

---

## Testing Requirements

### Unit Tests

- Permission validation for each included resource
- Organization boundary enforcement
- Price effective date filtering
- Inventory aggregation across warehouses
- Null/empty data handling

### Integration Tests

- Full flow with real database
- Concurrent read consistency
- Large product catalogs (100+ variants)
- Missing/partial data scenarios

### Performance Tests

- Load test with 1000+ concurrent requests
- Response time under various `include` combinations
- Database query performance with large datasets

---

## Example Use Cases

### Use Case 1: E-commerce Product Detail Page

**Frontend Needs**: Show product with all variants, prices, and availability

**Old Approach** (4 API calls):
```
GET /products/PROD_12345
GET /product-variants?product_id=PROD_12345
GET /prices?variant_id=PVAR_001 (per variant)
GET /batches?variant_id=PVAR_001 (per variant)
```

**New Approach** (1 API call):
```
GET /products/PROD_12345/detail?include=variants,prices,inventory
```

**Benefit**: 75% reduction in API calls, 300-600ms faster

---

### Use Case 2: Store Staff Inventory Check

**Frontend Needs**: Check specific warehouse stock for a product

**Request**:
```
GET /products/PROD_12345/detail?warehouse_id=WH_001&include=variants,inventory
```

**Response**: Only variants with stock in WH_001, no pricing data (permission-based filtering)

---

### Use Case 3: Price Comparison Tool

**Frontend Needs**: Compare retail vs wholesale prices

**Request**:
```
GET /products/PROD_12345/detail?include=variants,prices&price_type=retail,wholesale
```

**Response**: Variants with retail and wholesale prices, no inventory data

---

## Security Considerations

1. **Rate Limiting**: 100 requests per minute per user
2. **Response Size**: Max 5MB per response (use pagination for large catalogs)
3. **Sensitive Data**: Cost prices filtered based on role
4. **Audit Logging**: All access logged with user_id, organization_id, timestamp
5. **HTTPS Only**: Enforce TLS 1.2+

---

## Monitoring & Alerts

### Metrics to Track

- Request count by `include` parameter combinations
- Response time by included resources
- Error rate by error code
- Cache hit/miss ratio
- Database query execution time

### Alert Conditions

- P95 response time > 500ms
- Error rate > 1%
- Cache hit rate < 70%
- Zero stock displayed but sale created (consistency issue)

---

## Open Questions & Future Enhancements

1. **GraphQL Alternative**: Consider GraphQL for even more flexible field selection
2. **Real-time Updates**: WebSocket subscriptions for inventory changes
3. **Predictive Prefetch**: Machine learning to predict next API calls
4. **Partial Responses**: HTTP 206 for extremely large responses
5. **Compression**: Gzip/Brotli for large JSON payloads
