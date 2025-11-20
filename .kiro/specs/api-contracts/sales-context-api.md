# Sales Context API Contract

## Overview

**Purpose**: Provide all necessary data for sale/checkout operations in a single API call, eliminating multiple pre-fetch operations.

**Current Problem**:
- Checkout flow requires 5-6 sequential API calls
- Fetching warehouses, inventory, prices, taxes separately
- 400-800ms total latency
- Poor UX with multiple loading states

**Solution Impact**:
- **80-83% reduction** in API calls (6 → 1)
- **400-800ms** faster checkout initialization
- **Single loading state** for entire checkout context
- **Consistent data snapshot** for transaction

---

## API Specification

### Endpoint: Get Sales Context

```
GET /api/v1/sales/context
```

**Description**: Retrieves all data needed to create a sale including available inventory, active prices, tax configuration, warehouse details, and applicable discounts.

**Authentication**: Required (Bearer Token)

**Authorization**: Requires `sale:create` permission

#### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `warehouse_id` | string | No | user's default | Filter inventory to specific warehouse |
| `include_zero_stock` | boolean | No | false | Include products with zero stock |
| `price_type` | string | No | retail | Price type to use: `retail`, `wholesale`, `bulk` |
| `customer_id` | string | No | null | For customer-specific pricing (future) |
| `effective_date` | string | No | now | ISO date for price effective date |

#### Request Example

```http
GET /api/v1/sales/context?warehouse_id=WH_001&price_type=retail
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### Response Schema

```json
{
  "warehouse": {
    "id": "WH_001",
    "name": "Main Warehouse - Delhi",
    "address": {
      "street": "Plot 45, Industrial Area",
      "city": "New Delhi",
      "state": "Delhi",
      "pincode": "110020",
      "country": "India"
    },
    "contact_phone": "+91-11-12345678",
    "is_active": true,
    "organization_id": "ORG_001"
  },

  "available_inventory": [
    {
      "batch_id": "BTCH_001",
      "variant_id": "PVAR_001",
      "variant": {
        "id": "PVAR_001",
        "variant_name": "Premium Basmati Rice - 1kg Pack",
        "sku": "PBR-1KG-001",
        "barcode": "8901234567890",
        "brand_name": "Golden Harvest",
        "quantity": "1kg",
        "pack_size": "Standard Pack",
        "images": ["https://cdn.example.com/products/rice-1kg-front.jpg"],
        "hsn_code": "1006",
        "is_active": true
      },
      "product": {
        "id": "PROD_12345",
        "name": "Premium Basmati Rice",
        "category": "Grains",
        "description": "Aged premium basmati rice from Punjab region"
      },
      "quantity_available": 3000,
      "quantity_reserved": 150,
      "quantity_sellable": 2850,
      "cost_price": 55.00,
      "expiry_date": "2025-12-31",
      "manufacturing_date": "2024-12-01",
      "batch_number": "BATCH-2024-12-001",

      "selling_price": {
        "price_id": "PRICE_001",
        "price": 85.00,
        "price_type": "retail",
        "currency": "INR",
        "effective_from": "2024-11-01T00:00:00Z",
        "effective_to": null,
        "is_active": true
      },

      "alternate_prices": [
        {
          "price": 75.00,
          "price_type": "wholesale",
          "min_quantity": 50
        },
        {
          "price": 70.00,
          "price_type": "bulk",
          "min_quantity": 100
        }
      ],

      "tax_config": {
        "cgst_rate": 2.5,
        "sgst_rate": 2.5,
        "total_gst_rate": 5.0,
        "is_tax_exempt": false,
        "custom_taxes": [],
        "hsn_code": "1006"
      },

      "margin": {
        "cost_price": 55.00,
        "selling_price": 85.00,
        "margin_amount": 30.00,
        "margin_percentage": 54.55
      }
    },
    {
      "batch_id": "BTCH_002",
      "variant_id": "PVAR_002",
      "variant": {
        "id": "PVAR_002",
        "variant_name": "Premium Basmati Rice - 5kg Pack",
        "sku": "PBR-5KG-001",
        "barcode": "8901234567891",
        "brand_name": "Golden Harvest",
        "quantity": "5kg",
        "pack_size": "Bulk Pack",
        "images": ["https://cdn.example.com/products/rice-5kg-front.jpg"],
        "hsn_code": "1006",
        "is_active": true
      },
      "product": {
        "id": "PROD_12345",
        "name": "Premium Basmati Rice",
        "category": "Grains",
        "description": "Aged premium basmati rice from Punjab region"
      },
      "quantity_available": 1500,
      "quantity_reserved": 50,
      "quantity_sellable": 1450,
      "cost_price": 270.00,
      "expiry_date": "2025-11-30",
      "manufacturing_date": "2024-11-01",
      "batch_number": "BATCH-2024-11-002",

      "selling_price": {
        "price_id": "PRICE_002",
        "price": 400.00,
        "price_type": "retail",
        "currency": "INR",
        "effective_from": "2024-11-01T00:00:00Z",
        "effective_to": null,
        "is_active": true
      },

      "alternate_prices": [
        {
          "price": 360.00,
          "price_type": "wholesale",
          "min_quantity": 20
        }
      ],

      "tax_config": {
        "cgst_rate": 2.5,
        "sgst_rate": 2.5,
        "total_gst_rate": 5.0,
        "is_tax_exempt": false,
        "custom_taxes": [],
        "hsn_code": "1006"
      },

      "margin": {
        "cost_price": 270.00,
        "selling_price": 400.00,
        "margin_amount": 130.00,
        "margin_percentage": 48.15
      }
    }
  ],

  "global_tax_configuration": {
    "default_cgst_rate": 2.5,
    "default_sgst_rate": 2.5,
    "tax_calculation_method": "inclusive",
    "active_taxes": [
      {
        "id": "TAX_001",
        "name": "Standard GST",
        "tax_type": "GST",
        "cgst_rate": 2.5,
        "sgst_rate": 2.5,
        "is_active": true
      }
    ]
  },

  "discount_policies": [
    {
      "id": "DISC_001",
      "name": "Bulk Purchase Discount",
      "discount_type": "percentage",
      "discount_value": 5.0,
      "min_quantity": 100,
      "min_amount": null,
      "applicable_categories": ["Grains"],
      "start_date": "2024-11-01T00:00:00Z",
      "end_date": "2024-12-31T23:59:59Z",
      "is_active": true
    },
    {
      "id": "DISC_002",
      "name": "Festival Sale",
      "discount_type": "fixed",
      "discount_value": 50.0,
      "min_quantity": null,
      "min_amount": 1000.0,
      "applicable_categories": null,
      "start_date": "2024-11-20T00:00:00Z",
      "end_date": "2024-11-25T23:59:59Z",
      "is_active": true
    }
  ],

  "refund_policies": [
    {
      "id": "REF_001",
      "name": "30-Day Return Policy",
      "description": "Full refund within 30 days for unopened packages",
      "refund_percentage": 100.0,
      "valid_days": 30,
      "conditions": ["Unopened package", "Original receipt required"],
      "is_active": true
    }
  ],

  "payment_methods": [
    {
      "id": "PAY_CASH",
      "name": "Cash",
      "type": "cash",
      "is_active": true
    },
    {
      "id": "PAY_CARD",
      "name": "Card Payment",
      "type": "card",
      "is_active": true
    },
    {
      "id": "PAY_UPI",
      "name": "UPI",
      "type": "upi",
      "is_active": true
    },
    {
      "id": "PAY_CREDIT",
      "name": "Credit (30 days)",
      "type": "credit",
      "credit_days": 30,
      "is_active": true
    }
  ],

  "metadata": {
    "total_products": 156,
    "total_variants": 234,
    "total_batches": 189,
    "total_stock_value": 1256789.50,
    "warehouse_capacity_used_percent": 67.5,
    "read_timestamp": "2024-11-21T10:30:00Z",
    "consistency_token": "CT_sales_xyz789",
    "expires_at": "2024-11-21T10:35:00Z"
  }
}
```

#### Response Status Codes

| Status Code | Description |
|-------------|-------------|
| 200 OK | Sales context retrieved successfully |
| 400 Bad Request | Invalid query parameters |
| 401 Unauthorized | Authentication token missing or invalid |
| 403 Forbidden | User lacks `sale:create` permission |
| 404 Not Found | Specified warehouse not found |
| 500 Internal Server Error | Server-side error occurred |

#### Error Response Schema

```json
{
  "status": "error",
  "error": {
    "code": "WAREHOUSE_NOT_FOUND",
    "message": "Warehouse with ID WH_001 not found or not accessible",
    "details": {
      "warehouse_id": "WH_001",
      "organization_id": "ORG_001"
    },
    "timestamp": "2024-11-21T10:30:00Z",
    "request_id": "req_abc123"
  }
}
```

---

## Business Rules & Constraints

### Inventory Allocation (FEFO)

1. **Batch Ordering**: Batches are ordered by expiry date (earliest first)
2. **Sellable Quantity**: `quantity_sellable = quantity_available - quantity_reserved`
3. **Reserved Stock**: Stock allocated to pending/in-progress sales is reserved
4. **Zero Stock Handling**: Excluded by default unless `include_zero_stock=true`

### Pricing Rules

1. **Price Hierarchy**:
   - Primary: Requested `price_type` (e.g., retail)
   - Fallback: Next available price type
   - Last resort: Cost price + default margin
2. **Effective Date Validation**: Only prices active on `effective_date` are returned
3. **Currency Consistency**: All prices in response use same currency
4. **Alternate Prices**: Displayed for quantity-based pricing tiers

### Tax Calculation

1. **Batch-Level Tax**: Each batch carries its own tax configuration
2. **Tax Exemptions**: `is_tax_exempt` flag overrides all tax calculations
3. **Custom Taxes**: Additional taxes (cess, surcharge) applied on top of GST
4. **Tax Method**: `inclusive` or `exclusive` determines calculation approach

### Discount Application

1. **Eligibility Check**: Discounts checked against quantity, amount, category
2. **Stackability**: Multiple discounts can apply (documented in policy)
3. **Priority Order**: Category-specific → Amount-based → Quantity-based
4. **Active Period**: Only discounts within start_date to end_date are returned

### Margin Calculation

1. **Formula**: `margin_amount = selling_price - cost_price`
2. **Percentage**: `margin_percentage = (margin_amount / cost_price) * 100`
3. **Visibility**: Only shown to roles with `financial:read` permission
4. **Currency**: Follows selling price currency

---

## Data Consistency Guarantees

### Transaction Isolation

1. **Read Snapshot**: All data read within single database transaction
2. **Consistency Token**: Token validates data hasn't changed when sale is created
3. **Token Expiry**: Context expires after 5 minutes (configurable)
4. **Validation on Submit**: Sale creation validates quantities and prices against token

### Optimistic Locking

```javascript
// Frontend Usage Example
const context = await fetchSalesContext();
const sale = {
  warehouse_id: context.warehouse.id,
  consistency_token: context.metadata.consistency_token,
  items: [
    {
      batch_id: "BTCH_001",
      quantity: 50,
      unit_price: 85.00
    }
  ]
};

const result = await createSale(sale);
// If consistency_token invalid, API returns 409 Conflict
// Frontend refreshes context and retry
```

### Race Condition Handling

1. **Stock Reservation**: Reserve stock immediately when added to cart (optional)
2. **Price Lock**: Lock prices for 5 minutes after fetching context
3. **Concurrent Sales**: Database-level row locking prevents overselling
4. **Rollback**: Failed sales release all reservations atomically

---

## Performance Characteristics

### Response Time Targets

- **P50**: < 150ms
- **P95**: < 400ms
- **P99**: < 800ms

### Response Size

- **Typical**: 50-200KB (50-100 products)
- **Large**: Up to 2MB (500+ products)
- **Mitigation**: Pagination or warehouse-specific filtering

### Caching Strategy

- **Warehouse Metadata**: Cache 10 minutes
- **Tax Configuration**: Cache 5 minutes
- **Discount Policies**: Cache 1 minute
- **Inventory & Prices**: NO CACHE (real-time)
- **Cache Invalidation**: Webhook-based on inventory/price updates

### Database Optimization

1. **Composite Indexes**:
   - `inventory_batches(warehouse_id, variant_id, total_quantity DESC, expiry_date ASC)`
   - `product_prices(variant_id, is_active, price_type, effective_from)`
2. **Query Optimization**: Single query with LATERAL JOINs
3. **Connection Pooling**: Dedicated pool for sales operations
4. **Read Replicas**: Use replica for heavy read operations

---

## Security Considerations

### Authorization

1. **Warehouse Access**: Verify user has access to specified warehouse
2. **Organization Isolation**: Filter all data by user's organization_id
3. **Role-Based Fields**:
   - `cost_price` and `margin`: Only for Manager/CEO roles
   - `quantity_reserved`: Only for Inventory Manager role
   - `alternate_prices`: Only for Sales Manager role

### Data Sensitivity

1. **Cost Price Protection**: Masked for unauthorized roles
2. **Margin Exposure**: Business-critical data, strict access control
3. **Customer Data**: Future customer-specific pricing requires additional auth
4. **Audit Trail**: Log all context fetches with user_id and timestamp

### Rate Limiting

- **Per User**: 60 requests per minute
- **Per Warehouse**: 100 requests per minute
- **Burst**: Allow 10 requests burst, then throttle

---

## Frontend Integration Example

### React/TypeScript Example

```typescript
interface SalesContext {
  warehouse: Warehouse;
  available_inventory: InventoryItem[];
  global_tax_configuration: TaxConfig;
  discount_policies: Discount[];
  refund_policies: RefundPolicy[];
  payment_methods: PaymentMethod[];
  metadata: ContextMetadata;
}

const CheckoutPage = () => {
  const [context, setContext] = useState<SalesContext | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchContext = async () => {
      try {
        const response = await fetch(
          '/api/v1/sales/context?warehouse_id=WH_001&price_type=retail',
          {
            headers: {
              'Authorization': `Bearer ${token}`,
            },
          }
        );
        const data = await response.json();
        setContext(data);
      } catch (error) {
        console.error('Failed to fetch sales context:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchContext();
  }, []);

  if (loading) return <LoadingSpinner />;

  return (
    <CheckoutForm
      inventory={context.available_inventory}
      taxConfig={context.global_tax_configuration}
      discounts={context.discount_policies}
      consistencyToken={context.metadata.consistency_token}
    />
  );
};
```

---

## Migration Strategy

### Backward Compatibility

**Existing Endpoints Maintained**:
- `GET /api/v1/warehouses` - Warehouse list
- `GET /api/v1/inventory/batches` - Batch inventory
- `GET /api/v1/prices` - Product prices
- `GET /api/v1/taxes` - Tax configuration
- `GET /api/v1/discounts` - Discount policies

### Migration Path

**Phase 1** (Week 1-2):
- Deploy new sales context endpoint
- Feature flag for gradual rollout
- A/B test with 10% of traffic

**Phase 2** (Week 3-4):
- Update POS frontend to use new endpoint
- Monitor performance and error rates
- Increase traffic to 50%

**Phase 3** (Week 5-6):
- Full traffic migration
- Deprecation warnings on old endpoints
- Performance optimization based on metrics

**Phase 4** (Month 3+):
- Remove old endpoints after 6 months
- Update API documentation
- Archive legacy code

---

## Testing Requirements

### Unit Tests

- Organization boundary enforcement
- Role-based field filtering (cost_price, margin)
- FEFO batch ordering
- Price effective date validation
- Discount eligibility logic
- Consistency token generation and validation

### Integration Tests

- Full checkout flow with context
- Concurrent sale creation (race conditions)
- Stock reservation and release
- Price changes during checkout
- Expired consistency token handling

### Load Tests

- 1000+ concurrent context fetches
- Large warehouse with 500+ products
- Database query performance under load
- Response time distribution (P50, P95, P99)

### Business Logic Tests

- **Test Case 1**: Verify FEFO compliance (earliest expiry first)
- **Test Case 2**: Validate stock reservation prevents overselling
- **Test Case 3**: Ensure price lock during context validity period
- **Test Case 4**: Confirm discount stacking rules
- **Test Case 5**: Verify margin calculation accuracy
- **Test Case 6**: Test organization isolation (no data leakage)

---

## Monitoring & Alerts

### Key Metrics

- Context fetch latency (by warehouse size)
- Cache hit ratio
- Consistency token validation success rate
- Stock reservation vs actual sales ratio
- Discount application rate

### Alert Conditions

- **Critical**: P95 latency > 800ms
- **High**: Consistency token expiry rate > 5%
- **Medium**: Cache hit rate < 60%
- **Low**: Overselling incidents (should be zero)

### Dashboard Widgets

1. **Real-time Sales Context Performance**
2. **Stock Reservation Trends**
3. **Popular Products (by fetch count)**
4. **Warehouse Load Distribution**
5. **Error Rate by Error Type**

---

## Future Enhancements

1. **Customer-Specific Pricing**: Support for `customer_id` parameter
2. **Multi-Warehouse Selection**: Allow cart items from multiple warehouses
3. **Real-Time Stock Updates**: WebSocket for live inventory changes
4. **Predictive Stock Alerts**: Warn about low stock before adding to cart
5. **Dynamic Pricing**: AI-based pricing recommendations
6. **Bundled Products**: Support for product bundles and kits
7. **Subscription Pricing**: Recurring purchase discounts

---

## Related API Contracts

- [Aggregated Product API](./aggregated-product-api.md) - For product browsing
- [Purchase Order Detail API](./purchase-order-detail-api.md) - For procurement
- [Inventory List API](./inventory-list-api.md) - For inventory management
