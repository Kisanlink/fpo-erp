# API Changes: Beta → Development/Variants-Fix

This document describes all breaking changes, new features, and response format changes between the `beta` branch and the current `development/variants-fix` branch.

**Target Audience:** Frontend developers integrating with the ERP API

---

## Table of Contents

1. [Breaking Changes Summary](#breaking-changes-summary)
2. [Removed Endpoints](#removed-endpoints)
3. [New Endpoints](#new-endpoints)
   - [Aggregation API](#1-aggregation-api-performance-optimization)
   - [Sales Cancellation](#2-sales-cancellation-endpoints)
   - [GRN Rejection Tracking](#3-grn-rejection-tracking)
   - [Categories API](#4-categories-api-new)
   - [Subcategories API](#5-subcategories-api-new)
   - [Products by Category](#6-products-by-category-new)
4. [Response Format Changes](#response-format-changes)
5. [Endpoints with Query Parameters](#endpoints-with-query-parameters)
6. [Frontend Migration Guide](#frontend-migration-guide)

---

## Breaking Changes Summary

| Change | Impact | Action Required |
|--------|--------|-----------------|
| **Tax endpoints removed (13 endpoints)** | **HIGH** | Remove all `/api/v1/taxes/*` API calls |
| **`apply_taxes` field on sales** | **MEDIUM** | Set `apply_taxes: true` to calculate taxes (default: false) |
| **Inventory batch tax fields removed** | **MEDIUM** | Remove `cgst_rate`, `sgst_rate`, `is_tax_exempt` from batch handling |
| **PO item GST breakdown fields** | **LOW** | Handle new GST fields in PO item responses (`base_price`, `gst_rate`, etc.) |
| **PO `is_inter_state` field** | **LOW** | Handle new inter-state flag on PO responses |
| Collaborator-Product endpoints removed | HIGH | Use variants API instead |
| `collaborator_id` → `collaborator_ids` | HIGH | Update to array format |
| Prices in variant response | MEDIUM | Handle new `prices` array |
| Reserved quantity in inventory | MEDIUM | Use `available_quantity` for stock checks |
| Sales pending status | MEDIUM | Handle `pending` → `completed` workflow |
| PO `verified` status | LOW | Add to status handling |
| Product `category_name` → `category_id` | HIGH | Use ID-based category reference (changed from name-based) |
| Product `subcategory_name` → `subcategory_id` | HIGH | Use ID-based subcategory reference (changed from name-based) |
| Subcategory `category_name` → `category_id` | HIGH | Subcategory model uses ID-based FK (changed from name-based) |
| Subcategory endpoint `/category/:category` → `/category/:categoryId` | HIGH | Use category ID in URL path |
| **`customer_id` → Customer Details** | **HIGH** | Replace `customer_id` with `customer_phone`, `customer_name`, `is_org_member` |

---

## Removed Endpoints

### Tax Management API (ALL 13 ENDPOINTS REMOVED)

**Reason:** Simplified to GST-only tax system. GST rate is now stored directly on `ProductVariant.gst_rate` field. No separate tax configuration needed.

```
REMOVED: POST   /api/v1/taxes                    - Create tax (no longer needed)
REMOVED: GET    /api/v1/taxes                    - Get all taxes
REMOVED: GET    /api/v1/taxes/:id                - Get tax by ID
REMOVED: PATCH  /api/v1/taxes/:id                - Update tax
REMOVED: DELETE /api/v1/taxes/:id                - Delete tax
REMOVED: GET    /api/v1/taxes/active             - Get active taxes
REMOVED: GET    /api/v1/taxes/type/:type         - Get taxes by type
REMOVED: POST   /api/v1/taxes/:id/tiers          - Create tax tiers
REMOVED: GET    /api/v1/taxes/:id/tiers          - Get tax tiers
REMOVED: PATCH  /api/v1/taxes/:id/tiers/:tierId  - Update tax tier
REMOVED: DELETE /api/v1/taxes/:id/tiers/:tierId  - Delete tax tier
REMOVED: POST   /api/v1/taxes/calculate          - Calculate tax (now automatic)
REMOVED: GET    /api/v1/taxes/hsn/:hsnCode       - Get taxes by HSN code
```

**Migration Path:**
```javascript
// OLD (Beta) - Tax configuration via API
const tax = await api.post('/taxes', {
  name: 'GST 18%',
  type: 'GST',
  rate: 18.0,
  hsn_codes: ['1234']
});

// Assign tax to batch
await api.post('/batches', {
  // ...
  cgst_rate: 9.0,
  sgst_rate: 9.0,
  custom_tax_ids: [tax.id]
});

// NEW (Current) - GST rate on variant
// NO tax API calls needed!
// GST rate is set on the product variant:
await api.post(`/products/${productId}/variants`, {
  variant_name: '500g',
  quantity: '500g',
  pack_size: 'Single',
  hsn_code: '12345678',   // Required: HSN code for GST
  gst_rate: 18.0,         // Required: GST rate (0, 5, 12, 18, or 28)
  prices: [...]
});

// Tax calculation happens automatically during sale if apply_taxes: true
const sale = await api.post('/sales', {
  warehouse_id: 'WREH00000001',
  payment_mode: 'cash',
  sale_type: 'in_store',
  apply_taxes: true,  // Enable tax calculation (default: false)
  items: [...]
});
```

---

### Collaborator Products API (ALL REMOVED)

**Reason:** Unified variant architecture - collaborator products are now managed through ProductVariant with `collaborator_ids` array.

```
REMOVED: POST   /api/v1/collaborators/:id/products
REMOVED: GET    /api/v1/collaborator-products
REMOVED: GET    /api/v1/collaborator-products/:id
REMOVED: GET    /api/v1/collaborator-products/collaborator/:id
REMOVED: GET    /api/v1/collaborator-products/product/:id
REMOVED: PATCH  /api/v1/collaborator-products/:id
REMOVED: PATCH  /api/v1/collaborator-products/:id/status
REMOVED: PATCH  /api/v1/collaborator-products/:id/images
REMOVED: DELETE /api/v1/collaborator-products/:id
```

**Migration Path:**
```javascript
// OLD (Beta) - REMOVED
POST /api/v1/collaborators/:id/products
{
  "product_id": "PROD_123",
  "brand_name": "Fresh Farms"
}

// NEW (Current) - Use variants API
POST /api/v1/products/:productId/variants
{
  "variant_name": "500g",
  "quantity": "500g",
  "pack_size": "Single",
  "collaborator_ids": ["CLAB_123", "CLAB_456"],
  "brand_name": "Fresh Farms",
  "prices": [
    {"price_type": "MRP", "price": 199.99, "currency": "INR"}
  ]
}
```

---

## New Endpoints

### 1. Aggregation API (Performance Optimization)

Reduces frontend API calls by 75-85% through data aggregation.

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/aggregation/products/:id` | GET | Full product with variants, prices, inventory |
| `/api/v1/aggregation/variants/:id` | GET | Variant with product and collaborator context |
| `/api/v1/aggregation/sales-context/:warehouseId` | GET | POS/sales context data |
| `/api/v1/aggregation/purchase-orders/:id` | GET | Full PO with items, GRNs, payments |
| `/api/v1/aggregation/inventory` | GET | Inventory list with filters |

### 2. Sales Cancellation Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/sales/:id/cancel-items` | POST | Partial item cancellation |
| `/api/v1/sales/:id/cancellations` | GET | View cancellation history |

### 3. GRN Rejection Tracking

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/grns/:id/rejected-items` | GET | Get rejected items from GRN |
| `/api/v1/grns/items/:id/return-status` | PATCH | Update item return status |

### 4. Categories API (NEW)

Product categorization system with predefined hierarchy.

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/categories` | GET | List all categories |
| `/api/v1/categories/:id` | GET | Get category by ID |
| `/api/v1/categories/name/:name` | GET | Get category by name |
| `/api/v1/categories/with-subcategories` | GET | Get all categories with their subcategories |
| `/api/v1/categories` | POST | Create new category |
| `/api/v1/categories/:id` | PATCH | Update category |
| `/api/v1/categories/:id` | DELETE | Delete category |
| `/api/v1/categories/seed` | POST | Seed predefined categories (admin-only, idempotent) |

### 5. Subcategories API (UPDATED - ID-BASED)

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/subcategories` | GET | List all subcategories |
| `/api/v1/subcategories/:id` | GET | Get subcategory by ID |
| `/api/v1/subcategories/name/:name` | GET | Get subcategory by name |
| `/api/v1/subcategories/category/:categoryId` | GET | Get subcategories by category ID (**CHANGED** from `:category` name) |
| `/api/v1/subcategories` | POST | Create new subcategory |
| `/api/v1/subcategories/:id` | PATCH | Update subcategory |
| `/api/v1/subcategories/:id` | DELETE | Delete subcategory |

**Breaking Change:** Subcategory model now uses `category_id` instead of `category_name`.

**CreateSubcategoryRequest:**
```json
{
  "name": "WATER_SOLUBLE",
  "category_id": "CATG00000001",
  "description": "Water soluble fertilizers"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | **Yes** | Subcategory name (ALL_CAPS_SNAKE_CASE recommended) |
| `category_id` | `string` | **Yes** | Category ID (NOT category name) |
| `description` | `*string` | No | Subcategory description |

**SubcategoryResponse:**
```json
{
  "id": "SCAT00000001",
  "name": "WATER_SOLUBLE",
  "description": "Water soluble fertilizers",
  "category_id": "CATG00000001",
  "created_at": "2025-12-08T10:30:00Z",
  "updated_at": "2025-12-08T10:30:00Z"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Subcategory ID |
| `name` | `string` | Subcategory name |
| `description` | `*string` | Subcategory description (nullable) |
| `category_id` | `string` | Category ID (**CHANGED** from `category_name`) |
| `created_at` | `string` | Creation timestamp |
| `updated_at` | `string` | Last update timestamp |

### 6. Products by Category (NEW)

Get products filtered by category with optional subcategory filter.

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/products/category/:categoryId` | GET | Get all products in a category |
| `/api/v1/products/category/:categoryId?subcategory_id=SUBC00000001` | GET | Get products filtered by category and subcategory |

**Request Parameters:**

| Type | Name | Required | Description |
|------|------|----------|-------------|
| Path | `categoryId` | Yes | Category ID (e.g., `CATG00000001`) |
| Query | `subcategory_id` | No | Optional subcategory filter |

**Response:** Array of `ProductResponse` objects (same format as `GET /api/v1/products`)

**Example:**
```javascript
// Get all products in a category
const products = await api.get('/products/category/CATG00000001');

// Get products filtered by subcategory
const filtered = await api.get('/products/category/CATG00000001?subcategory_id=SUBC00000001');
```

---

## Response Format Changes

### 1. ProductVariantResponse

**Old (Beta):**
```json
{
  "id": "PVAR00000001",
  "product_id": "PROD00000001",
  "variant_name": "500g",
  "sku": "TOM-500G",
  "quantity": "500g",
  "pack_size": "Single",
  "collaborator_id": "CLAB00000001",
  "brand_name": "Fresh Farms",
  "is_active": true,
  "created_at": "2025-11-10T10:30:00Z",
  "updated_at": "2025-11-10T10:30:00Z"
}
```

**New (Current):**
```json
{
  "id": "PVAR00000001",
  "product_id": "PROD00000001",
  "variant_name": "500g",
  "sku": "TOM-500G",
  "quantity": "500g",
  "pack_size": "Single",
  "hsn_code": "12345678",
  "gst_rate": 18.0,
  "collaborator_ids": ["CLAB00000001", "CLAB00000002"],
  "brand_name": "Fresh Farms",
  "prices": [
    {
      "id": "PRIC00000001",
      "variant_id": "PVAR00000001",
      "price_type": "MRP",
      "price": 199.99,
      "currency": "INR",
      "effective_from": "2025-11-01T00:00:00Z",
      "effective_to": null,
      "is_active": true,
      "created_at": "2025-11-01T00:00:00Z",
      "updated_at": "2025-11-01T00:00:00Z"
    },
    {
      "id": "PRIC00000002",
      "variant_id": "PVAR00000001",
      "price_type": "retail",
      "price": 179.99,
      "currency": "INR",
      "effective_from": "2025-11-01T00:00:00Z",
      "effective_to": null,
      "is_active": true,
      "created_at": "2025-11-01T00:00:00Z",
      "updated_at": "2025-11-01T00:00:00Z"
    }
  ],
  "is_active": true,
  "created_at": "2025-11-10T10:30:00Z",
  "updated_at": "2025-11-10T10:30:00Z"
}
```

**Changes:**
| Field | Old | New |
|-------|-----|-----|
| `collaborator_id` | `string` (nullable) | REMOVED |
| `collaborator_ids` | N/A | `[]string` (NEW) |
| `hsn_code` | N/A | `string` (NEW, required) - HSN code for GST classification |
| `gst_rate` | N/A | `float64` (NEW, required) - GST rate (0, 5, 12, 18, or 28) |
| `prices` | N/A | `[]ProductPriceResponse` (NEW) |

---

### 2. InventoryBatchResponse (Tax Fields Removed)

**Old (Beta):**
```json
{
  "id": "BATC00000001",
  "warehouse_id": "WREH00000001",
  "variant_id": "PVAR00000001",
  "cost_price": 100.00,
  "expiry_date": "2025-12-31",
  "total_quantity": 500,
  "cgst_rate": 9.0,
  "sgst_rate": 9.0,
  "is_tax_exempt": false,
  "custom_tax_ids": ["TAX00000001"],
  "created_at": "2025-11-10T10:30:00Z"
}
```

**New (Current):**
```json
{
  "id": "BATC00000001",
  "warehouse_id": "WREH00000001",
  "variant_id": "PVAR00000001",
  "cost_price": 100.00,
  "expiry_date": "2025-12-31",
  "total_quantity": 500,
  "reserved_quantity": 150,
  "available_quantity": 350,
  "created_at": "2025-11-10T10:30:00Z"
}
```

**Changes:**
| Field | Old | New |
|-------|-----|-----|
| `cgst_rate` | `float64` | **REMOVED** - Tax comes from variant now |
| `sgst_rate` | `float64` | **REMOVED** - Tax comes from variant now |
| `is_tax_exempt` | `bool` | **REMOVED** - Use `apply_taxes` on sale |
| `custom_tax_ids` | `[]string` | **REMOVED** - No custom taxes |
| `reserved_quantity` | N/A | `int64` (NEW) - Stock reserved by pending sales |
| `available_quantity` | N/A | `int64` (NEW) - `total_quantity - reserved_quantity` |

**Important:**
- Always use `available_quantity` for stock availability checks, NOT `total_quantity`.
- Tax is now determined by the product variant's `gst_rate`, NOT the inventory batch.

---

### 3. SaleResponse - Status Workflow & Tax Control

**Old (Beta):**
```json
{
  "id": "SALE00000001",
  "status": "completed"
}
```

**New (Current):**
```json
{
  "id": "SALE00000001",
  "status": "pending",
  "apply_taxes": false
}
```

**Status Workflow Change:**
```
OLD (Beta):
  Sale Created → "completed" (inventory deducted immediately)

NEW (Current):
  Sale Created → "pending" (inventory reserved)
       ↓
  Complete Sale → "completed" (inventory deducted)
       OR
  Cancel Sale → "cancelled" (reservation released)
```

**Valid Status Values:**
- `pending` - Sale created, inventory reserved (NEW)
- `completed` - Sale finalized, inventory deducted
- `cancelled` - Sale cancelled, inventory released

**New `apply_taxes` Field (GST Control):**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `apply_taxes` | `bool` | `false` | Controls whether GST is calculated for this sale |

**Tax Calculation Behavior:**
- `apply_taxes: false` (default) → No taxes calculated, all tax amounts = 0
- `apply_taxes: true` → GST calculated using variant's `gst_rate`
  - CGST + SGST (intra-state) or IGST (inter-state) applied
  - Tax amounts included in sale response

**Example - Sale WITHOUT taxes:**
```javascript
const sale = await api.post('/sales', {
  warehouse_id: 'WREH00000001',
  payment_mode: 'cash',
  sale_type: 'in_store',
  // apply_taxes not set → defaults to false
  items: [{ variant_id: 'PVAR00000001', quantity: 10 }]
});
// sale.items[0].cgst_amount = 0
// sale.items[0].sgst_amount = 0
// sale.items[0].total_tax_amount = 0
```

**Example - Sale WITH taxes:**
```javascript
const sale = await api.post('/sales', {
  warehouse_id: 'WREH00000001',
  payment_mode: 'cash',
  sale_type: 'in_store',
  apply_taxes: true,  // Enable GST calculation
  items: [{ variant_id: 'PVAR00000001', quantity: 10 }]
});
// sale.items[0].cgst_amount = 90.00 (9% of line total)
// sale.items[0].sgst_amount = 90.00 (9% of line total)
// sale.items[0].total_tax_amount = 180.00
```

---

### 3.1 SaleResponse - Customer Details (BREAKING CHANGE)

**Old (Beta) - Full Response:**
```json
{
  "id": "SALE00000001",
  "warehouse_id": "WREH00000001",
  "sale_date": "2025-12-08T10:30:00Z",
  "total_amount": 5000.00,
  "status": "completed",
  "customer_id": "CUST_123",
  "payment_mode": "cash",
  "sale_type": "in_store",
  "apply_taxes": false,
  "items": [
    {
      "id": "SITM00000001",
      "sale_id": "SALE00000001",
      "batch_id": "BATC00000001",
      "quantity": 10,
      "selling_price": 500.00,
      "line_total": 5000.00,
      "cost_price": 400.00,
      "margin": 100.00,
      "cgst_amount": 0,
      "sgst_amount": 0,
      "igst_amount": 0,
      "total_tax_amount": 0,
      "created_at": "2025-12-08T10:30:00Z"
    }
  ],
  "breakdown": {
    "base_amount": 5000.00,
    "applied_discounts": [],
    "discount_amount": 0,
    "tax_breakdown": {
      "cgst_amount": 0,
      "sgst_amount": 0,
      "igst_amount": 0,
      "total_tax_amount": 0,
      "is_inter_state": false
    },
    "tax_amount": 0,
    "total_savings": 0,
    "final_amount": 5000.00
  },
  "created_at": "2025-12-08T10:30:00Z",
  "updated_at": "2025-12-08T10:30:00Z"
}
```

**New (Current) - Full Response:**
```json
{
  "id": "SALE00000001",
  "warehouse_id": "WREH00000001",
  "sale_date": "2025-12-08T10:30:00Z",
  "total_amount": 5000.00,
  "status": "pending",
  "customer_phone": "9876543210",
  "customer_name": "John Doe",
  "is_org_member": false,
  "payment_mode": "cash",
  "sale_type": "in_store",
  "apply_taxes": false,
  "items": [
    {
      "id": "SITM00000001",
      "sale_id": "SALE00000001",
      "batch_id": "BATC00000001",
      "quantity": 10,
      "selling_price": 500.00,
      "line_total": 5000.00,
      "cost_price": 400.00,
      "margin": 100.00,
      "cgst_amount": 0,
      "sgst_amount": 0,
      "igst_amount": 0,
      "total_tax_amount": 0,
      "created_at": "2025-12-08T10:30:00Z"
    }
  ],
  "breakdown": {
    "base_amount": 5000.00,
    "applied_discounts": [],
    "discount_amount": 0,
    "tax_breakdown": {
      "cgst_amount": 0,
      "sgst_amount": 0,
      "igst_amount": 0,
      "total_tax_amount": 0,
      "is_inter_state": false
    },
    "tax_amount": 0,
    "total_savings": 0,
    "final_amount": 5000.00
  },
  "created_at": "2025-12-08T10:30:00Z",
  "updated_at": "2025-12-08T10:30:00Z"
}
```

**Field Changes:**
| Field | Old | New |
|-------|-----|-----|
| `customer_id` | `string` (nullable) | **REMOVED** |
| `customer_phone` | N/A | `*string` (NEW, nullable) - Customer phone number |
| `customer_name` | N/A | `*string` (NEW, nullable) - Customer name |
| `is_org_member` | N/A | `bool` (NEW, required) - Whether customer is FPO member |

**Rationale:**
- `customer_id` was an external reference that didn't provide direct customer information
- New fields capture customer details directly for:
  - Better customer identification without external lookups
  - FPO member tracking for differential pricing/benefits
  - Audit trail with human-readable customer info

**`is_org_member` Field:**
- `false` (default): External customer (non-member)
- `true`: FPO member (eligible for member pricing/benefits)

---

### 4. PurchaseOrderResponse

**Old (Beta):**
```json
{
  "id": "PORD00000001",
  "po_number": "PO-2025-0001",
  "status": "delivered",
  "total_amount": 50000.00,
  "payment_status": "unpaid",
  "paid_amount": 0.00
}
```

**New (Current):**
```json
{
  "id": "PORD00000001",
  "po_number": "PO-2025-0001",
  "status": "verified",
  "total_amount": 50000.00,
  "total_rejected_amount": 5000.00,
  "amount_owed": 45000.00,
  "payment_status": "partial",
  "paid_amount": 20000.00,
  "is_inter_state": false,
  "items": [
    {
      "id": "POIM00000001",
      "po_id": "PORD00000001",
      "variant_id": "PVAR00000001",
      "product_name": "Rice - Premium",
      "product_sku": "RICE-PREM-5KG",
      "quantity": 100,
      "unit_price": 118.00,
      "line_total": 11800.00,
      "received_quantity": 100,
      "base_price": 100.00,
      "gst_rate": 18.0,
      "gst_amount": 18.00,
      "cgst_rate": 9.0,
      "cgst_amount": 9.00,
      "sgst_rate": 9.0,
      "sgst_amount": 9.00,
      "igst_rate": 0,
      "igst_amount": 0,
      "created_at": "2025-12-01T10:00:00Z"
    }
  ]
}
```

**Status Workflow Change:**
```
OLD (Beta):
  placed → confirmed → out_for_delivery → delivered → paid

NEW (Current):
  placed → confirmed → out_for_delivery → delivered → verified → paid
```

**New Fields (PurchaseOrderResponse):**
| Field | Type | Description |
|-------|------|-------------|
| `total_rejected_amount` | `float64` | Value of rejected items from GRN |
| `amount_owed` | `float64` | `total_amount - total_rejected_amount` |
| `is_inter_state` | `*bool` | `true` = inter-state (IGST), `false` = intra-state (CGST+SGST), `null` = unknown |

**New Fields (PurchaseOrderItemResponse - GST Breakdown):**
| Field | Type | Description |
|-------|------|-------------|
| `base_price` | `float64` | Price before GST (reverse-calculated from unit_price) |
| `gst_rate` | `float64` | GST rate from variant (e.g., 5, 12, 18, 28) |
| `gst_amount` | `float64` | Total GST per unit |
| `cgst_rate` | `float64` | Central GST rate (0 if inter-state) |
| `cgst_amount` | `float64` | Central GST amount per unit (0 if inter-state) |
| `sgst_rate` | `float64` | State GST rate (0 if inter-state) |
| `sgst_amount` | `float64` | State GST amount per unit (0 if inter-state) |
| `igst_rate` | `float64` | Integrated GST rate (0 if intra-state) |
| `igst_amount` | `float64` | Integrated GST amount per unit (0 if intra-state) |

**GST Calculation Logic:**
- User enters ALL-IN `unit_price` (includes GST)
- System reverse-calculates: `base_price = unit_price / (1 + gst_rate/100)`
- GST split depends on collaborator vs warehouse state:
  - **Intra-state** (same state): CGST + SGST at 50/50 split
  - **Inter-state** (different states): IGST at full rate
  - **Unknown** (no state info): Only `gst_amount` populated, no split

**Example - Intra-state (Maharashtra → Maharashtra):**
```
unit_price: 118.00 (ALL-IN)
gst_rate: 18%
base_price: 100.00
gst_amount: 18.00
cgst_rate: 9%, cgst_amount: 9.00
sgst_rate: 9%, sgst_amount: 9.00
igst_rate: 0, igst_amount: 0
is_inter_state: false
```

**Example - Inter-state (Gujarat → Maharashtra):**
```
unit_price: 118.00 (ALL-IN)
gst_rate: 18%
base_price: 100.00
gst_amount: 18.00
cgst_rate: 0, cgst_amount: 0
sgst_rate: 0, sgst_amount: 0
igst_rate: 18%, igst_amount: 18.00
is_inter_state: true
```

---

### 5. GRNItemResponse - Return Tracking

**Old (Beta):**
```json
{
  "id": "GRIT00000001",
  "po_item_id": "POIM00000001",
  "variant_id": "PVAR00000001",
  "received_quantity": 95,
  "accepted_quantity": 90,
  "rejected_quantity": 5,
  "expiry_date": "2025-12-31"
}
```

**New (Current):**
```json
{
  "id": "GRIT00000001",
  "po_item_id": "POIM00000001",
  "variant_id": "PVAR00000001",
  "received_quantity": 95,
  "accepted_quantity": 90,
  "rejected_quantity": 5,
  "expiry_date": "2025-12-31",
  "return_status": "pending",
  "return_sent_date": null,
  "return_received_date": null,
  "return_closed_date": null,
  "return_remarks": "Damaged packaging"
}
```

**New Fields for Rejected Item Tracking:**
| Field | Type | Description |
|-------|------|-------------|
| `return_status` | `string` | `pending`, `sent`, `received_by_vendor`, `closed` |
| `return_sent_date` | `string` (nullable) | When shipped to vendor |
| `return_received_date` | `string` (nullable) | When vendor confirmed receipt |
| `return_closed_date` | `string` (nullable) | When return process closed |
| `return_remarks` | `string` (nullable) | Notes about return |

---

### 6. ProductResponse - Category Fields (CHANGED TO ID-BASED)

**Old (Beta):**
```json
{
  "id": "PROD00000001",
  "name": "NPK 19-19-19",
  "description": "Balanced fertilizer",
  "variants": [...]
}
```

**New (Current) - ID-BASED REFERENCES:**
```json
{
  "id": "PROD00000001",
  "name": "NPK 19-19-19",
  "description": "Balanced fertilizer",
  "category_id": "CATG00000001",
  "subcategory_id": "SCAT00000002",
  "variants": [...]
}
```

**New Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `category_id` | `*string` (nullable) | Category ID reference (optional) |
| `subcategory_id` | `*string` (nullable) | Subcategory ID reference (optional) |

**BREAKING CHANGE:** Changed from name-based (`category_name`, `subcategory_name`) to ID-based (`category_id`, `subcategory_id`) references.

### 6.1 CreateProductRequest

**Request Body:**
```json
{
  "name": "NPK 19-19-19",
  "description": "Balanced fertilizer for crops",
  "category_id": "CATG00000001",
  "subcategory_id": "SCAT00000002"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | **Yes** | Product name |
| `description` | `*string` | No | Product description |
| `category_id` | `*string` | No | Category ID (nullable) |
| `subcategory_id` | `*string` | No | Subcategory ID (nullable) |

### 6.2 UpdateProductRequest

**Request Body:**
```json
{
  "name": "NPK 19-19-19 Updated",
  "description": "Updated description",
  "category_id": "CATG00000003",
  "subcategory_id": "SCAT00000005"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `*string` | No | Product name |
| `description` | `*string` | No | Product description |
| `category_id` | `*string` | No | Category ID |
| `subcategory_id` | `*string` | No | Subcategory ID |

### 6.3 ProductResponse (Full Example)

**Response Body:**
```json
{
  "id": "PROD00000001",
  "name": "NPK 19-19-19",
  "description": "Balanced fertilizer for crops",
  "category_id": "CATG00000001",
  "subcategory_id": "SCAT00000002",
  "variants": [
    {
      "id": "PVAR00000001",
      "product_id": "PROD00000001",
      "variant_name": "500g",
      "sku": "NPK-500G",
      "quantity": "500g",
      "pack_size": "Single",
      "is_active": true,
      "prices": [...],
      "created_at": "2025-12-07T10:30:00Z",
      "updated_at": "2025-12-07T10:30:00Z"
    }
  ],
  "created_at": "2025-12-07T10:30:00Z",
  "updated_at": "2025-12-07T10:30:00Z"
}
```

---

## Endpoints with Query Parameters

### Products Module
| Endpoint | Query Parameters |
|----------|------------------|
| `GET /api/v1/products/search` | `q` (required) - Search term |

### Warehouses Module
| Endpoint | Query Parameters |
|----------|------------------|
| `GET /api/v1/warehouses/search` | `q` (required) - Search term |

### Inventory Module
| Endpoint | Query Parameters |
|----------|------------------|
| `GET /api/v1/batches/expiring` | `days` (optional, default: 30) |
| `GET /api/v1/batches/low-stock` | `threshold` (optional, default: 10) |

### Sales Module
| Endpoint | Query Parameters |
|----------|------------------|
| `GET /api/v1/sales` | `limit` (default: 10), `offset` (default: 0) |
| `GET /api/v1/sales/date-range` | `start_date` (required), `end_date` (required) - YYYY-MM-DD |
| `GET /api/v1/sales/total-amount` | `start_date` (required), `end_date` (required) - YYYY-MM-DD |
| `GET /api/v1/sales/top-selling` | `limit` (default: 10) |

### Returns Module
| Endpoint | Query Parameters |
|----------|------------------|
| `GET /api/v1/returns` | `limit` (default: 10), `offset` (default: 0) |
| `GET /api/v1/returns/date-range` | `start_date` (required), `end_date` (required) - YYYY-MM-DD |

### Discounts Module
| Endpoint | Query Parameters |
|----------|------------------|
| `GET /api/v1/discounts` | `limit` (default: 10), `offset` (default: 0) |

### ~~Taxes Module~~ (REMOVED)
**All tax endpoints have been removed.** Tax calculation is now automatic based on variant's `gst_rate` and sale's `apply_taxes` field.

### Attachments Module
| Endpoint | Query Parameters |
|----------|------------------|
| `GET /api/v1/attachments` | `entity_type` (optional), `entity_id` (optional) |

### Aggregation API (NEW)
| Endpoint | Query Parameters |
|----------|------------------|
| `GET /api/v1/aggregation/products/:id` | `include` (variants,prices,inventory,collaborators,taxes), `warehouse_id`, `price_type` (retail,wholesale,bulk,all), `active_only` (default: true), `in_stock_only` |
| `GET /api/v1/aggregation/variants/:id` | `include` (product,collaborators,prices,inventory) |
| `GET /api/v1/aggregation/sales-context/:warehouseId` | `include_zero_stock`, `price_type`, `effective_date` (ISO date) |
| `GET /api/v1/aggregation/purchase-orders/:id` | `include` (collaborator,warehouse,items,grns,inventory,payments) |
| `GET /api/v1/aggregation/inventory` | `warehouse_id`, `variant_id`, `product_id`, `expiring_within_days`, `low_stock_threshold`, `include_zero_stock`, `page`, `page_size` |

#### Aggregation Response Model Refactoring (December 2025)

**Change**: Aggregation endpoints now use the same response models as other APIs for consistency.

| Aggregation Object | Now Uses | New Fields Available |
|--------------------|----------|---------------------|
| Product info | `ProductResponse` | `category_id`, `subcategory_id`, `variants` |
| Variant info | `ProductVariantResponse` | `collaborator_ids` (array), `product_id`, `hsn_code`, `gst_rate`, full timestamps |
| Collaborator info | `CollaboratorResponse` | `gst_number`, `pan_number`, banking details, full timestamps |
| Warehouse info | `WarehouseResponse` | `address` object (if populated), standard format |
| PO info | `PurchaseOrderResponse` | `is_inter_state`, GST breakdown fields, `amount_owed` |
| PO item info | `PurchaseOrderItemResponse` | Full GST breakdown (`base_price`, `cgst_amount`, `sgst_amount`, `igst_amount`) |

**Benefits**:
- Consistent field names across all APIs
- Additional fields automatically available (category IDs, GST fields, timestamps)
- Easier frontend integration - same types everywhere

**No Breaking Changes**: This is additive - all existing fields remain, new fields are added.

#### Complete Response Body Examples for Aggregation Endpoints

##### 1. GET /api/v1/aggregation/products/:id - Product Detail

**Query Parameters:**
- `include` - Comma-separated list: `variants`, `prices`, `inventory`, `collaborators`, `taxes`
- `warehouse_id` - Filter inventory by warehouse
- `price_type` - Filter prices: `retail`, `wholesale`, `bulk`, `all` (default: all)
- `active_only` - Only active variants (default: true)
- `in_stock_only` - Only variants with stock

**Response:**
```json
{
  "product": {
    "id": "PROD00000001",
    "name": "Organic Tomatoes",
    "description": "Fresh organic tomatoes from local farms",
    "category_id": "CAT000000001",
    "subcategory_id": "SCAT00000001",
    "variants": [],
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-20T14:45:00Z"
  },
  "variants": [
    {
      "variant": {
        "id": "PVAR00000001",
        "product_id": "PROD00000001",
        "variant_name": "500g Pack",
        "description": "Half kilogram pack",
        "quantity": "500g",
        "pack_size": "Small",
        "sku": "TOM-500G-ORG",
        "barcode": "8901234567890",
        "hsn_code": "0702",
        "gst_rate": 5.0,
        "collaborator_ids": ["CLAB00000001", "CLAB00000002"],
        "brand_name": "FreshFarms",
        "images": null,
        "image_urls": null,
        "dosage_instructions": null,
        "usage_details": "Store in cool, dry place",
        "prices": [],
        "is_active": true,
        "created_at": "2025-01-15T10:35:00Z",
        "updated_at": "2025-01-20T14:50:00Z"
      },
      "prices": [
        {
          "id": "PRIC00000001",
          "variant_id": "PVAR00000001",
          "price_type": "MRP",
          "price": 99.00,
          "currency": "INR",
          "effective_from": "2025-01-01T00:00:00Z",
          "effective_to": null,
          "is_active": true,
          "created_at": "2025-01-15T10:40:00Z",
          "updated_at": "2025-01-15T10:40:00Z"
        },
        {
          "id": "PRIC00000002",
          "variant_id": "PVAR00000001",
          "price_type": "retail",
          "price": 89.00,
          "currency": "INR",
          "effective_from": "2025-01-01T00:00:00Z",
          "effective_to": null,
          "is_active": true,
          "created_at": "2025-01-15T10:40:00Z",
          "updated_at": "2025-01-15T10:40:00Z"
        }
      ],
      "inventory": {
        "total_quantity": 500,
        "available_quantity": 450,
        "reserved_quantity": 50,
        "warehouse_breakdown": [
          {
            "warehouse_id": "WRHS00000001",
            "warehouse_name": "Main Warehouse",
            "total_quantity": 300,
            "available_quantity": 270,
            "reserved_quantity": 30,
            "batches": [
              {
                "batch_id": "BATC00000001",
                "quantity": 300,
                "available_quantity": 270,
                "reserved_quantity": 30,
                "cost_price": 60.00,
                "expiry_date": "2025-03-15",
                "batch_number": "LOT-2025-001"
              }
            ]
          },
          {
            "warehouse_id": "WRHS00000002",
            "warehouse_name": "Secondary Warehouse",
            "total_quantity": 200,
            "available_quantity": 180,
            "reserved_quantity": 20,
            "batches": []
          }
        ]
      },
      "tax_info": {
        "hsn_code": "0702",
        "gst_rate": 5.0,
        "cgst_rate": 2.5,
        "sgst_rate": 2.5,
        "igst_rate": 5.0
      }
    }
  ],
  "collaborators": [
    {
      "id": "CLAB00000001",
      "company_name": "Organic Farms Ltd",
      "contact_person": "Rajesh Kumar",
      "contact_number": "9876543210",
      "email": "rajesh@organicfarms.com",
      "gst_number": "29AABCU9603R1ZM",
      "pan_number": "AABCU9603R",
      "bank_account_no": "1234567890123456",
      "bank_ifsc": "HDFC0001234",
      "bank_name": "HDFC Bank",
      "address_id": "ADDR00000001",
      "address": {
        "type": "business",
        "street": "123 Farm Road",
        "village": "Green Village",
        "taluk": "South Taluk",
        "district": "Bangalore Rural",
        "state": "Karnataka",
        "state_code": "KA",
        "pincode": "560001"
      },
      "is_active": true,
      "created_at": "2025-01-10T08:00:00Z",
      "updated_at": "2025-01-10T08:00:00Z"
    }
  ],
  "metadata": {
    "total_variants": 2,
    "active_variants": 2,
    "total_stock": 500,
    "available_stock": 450,
    "warehouses_with_stock": 2,
    "price_range": {
      "min": 89.00,
      "max": 99.00,
      "currency": "INR"
    }
  }
}
```

##### 2. GET /api/v1/aggregation/variants/:id - Variant Detail

**Query Parameters:**
- `include` - Comma-separated list: `product`, `collaborators`, `prices`, `inventory`

**Response:**
```json
{
  "variant": {
    "id": "PVAR00000001",
    "product_id": "PROD00000001",
    "variant_name": "500g Pack",
    "description": "Half kilogram pack",
    "quantity": "500g",
    "pack_size": "Small",
    "sku": "TOM-500G-ORG",
    "barcode": "8901234567890",
    "hsn_code": "0702",
    "gst_rate": 5.0,
    "collaborator_ids": ["CLAB00000001"],
    "brand_name": "FreshFarms",
    "images": ["products/tomato/500g-front.jpg"],
    "image_urls": ["https://s3.amazonaws.com/bucket/products/tomato/500g-front.jpg?signature=..."],
    "dosage_instructions": null,
    "usage_details": "Store in cool, dry place",
    "prices": [
      {
        "id": "PRIC00000001",
        "variant_id": "PVAR00000001",
        "price_type": "MRP",
        "price": 99.00,
        "currency": "INR",
        "effective_from": "2025-01-01T00:00:00Z",
        "effective_to": null,
        "is_active": true,
        "created_at": "2025-01-15T10:40:00Z",
        "updated_at": "2025-01-15T10:40:00Z"
      }
    ],
    "is_active": true,
    "created_at": "2025-01-15T10:35:00Z",
    "updated_at": "2025-01-20T14:50:00Z"
  },
  "product": {
    "id": "PROD00000001",
    "name": "Organic Tomatoes",
    "description": "Fresh organic tomatoes from local farms",
    "category_id": "CAT000000001",
    "subcategory_id": "SCAT00000001",
    "variants": [],
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-20T14:45:00Z"
  },
  "collaborators": [
    {
      "id": "CLAB00000001",
      "company_name": "Organic Farms Ltd",
      "contact_person": "Rajesh Kumar",
      "contact_number": "9876543210",
      "email": "rajesh@organicfarms.com",
      "gst_number": "29AABCU9603R1ZM",
      "pan_number": "AABCU9603R",
      "bank_account_no": "1234567890123456",
      "bank_ifsc": "HDFC0001234",
      "bank_name": "HDFC Bank",
      "address_id": "ADDR00000001",
      "address": null,
      "is_active": true,
      "created_at": "2025-01-10T08:00:00Z",
      "updated_at": "2025-01-10T08:00:00Z"
    }
  ],
  "inventory": {
    "total_quantity": 500,
    "available_quantity": 450,
    "reserved_quantity": 50,
    "warehouse_breakdown": [
      {
        "warehouse_id": "WRHS00000001",
        "warehouse_name": "Main Warehouse",
        "total_quantity": 500,
        "available_quantity": 450,
        "reserved_quantity": 50,
        "batches": [
          {
            "batch_id": "BATC00000001",
            "quantity": 500,
            "available_quantity": 450,
            "reserved_quantity": 50,
            "cost_price": 60.00,
            "expiry_date": "2025-03-15",
            "batch_number": "LOT-2025-001"
          }
        ]
      }
    ]
  },
  "metadata": {
    "has_stock": true,
    "is_available_for_sale": true,
    "earliest_expiry": "2025-03-15",
    "lowest_price": 89.00,
    "price_currency": "INR"
  }
}
```

##### 3. GET /api/v1/aggregation/sales-context/:warehouseId - Sales Context

**Query Parameters:**
- `include_zero_stock` - Include items with zero stock (default: false)
- `price_type` - Filter by price type: `retail`, `wholesale`, `bulk`, `all`
- `effective_date` - Price effective date (ISO format, default: now)

**Response:**
```json
{
  "warehouse": {
    "id": "WRHS00000001",
    "name": "Main Warehouse",
    "address": {
      "type": "warehouse",
      "street": "456 Industrial Area",
      "village": null,
      "taluk": "North Taluk",
      "district": "Bangalore Urban",
      "state": "Karnataka",
      "state_code": "KA",
      "pincode": "560002"
    },
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  },
  "inventory": [
    {
      "variant": {
        "id": "PVAR00000001",
        "product_id": "PROD00000001",
        "variant_name": "500g Pack",
        "description": "Half kilogram pack",
        "quantity": "500g",
        "pack_size": "Small",
        "sku": "TOM-500G-ORG",
        "barcode": "8901234567890",
        "hsn_code": "0702",
        "gst_rate": 5.0,
        "collaborator_ids": ["CLAB00000001"],
        "brand_name": "FreshFarms",
        "images": null,
        "image_urls": null,
        "dosage_instructions": null,
        "usage_details": null,
        "prices": [],
        "is_active": true,
        "created_at": "2025-01-15T10:35:00Z",
        "updated_at": "2025-01-20T14:50:00Z"
      },
      "product": {
        "id": "PROD00000001",
        "name": "Organic Tomatoes",
        "description": "Fresh organic tomatoes",
        "category_id": "CAT000000001",
        "subcategory_id": "SCAT00000001",
        "variants": [],
        "created_at": "2025-01-15T10:30:00Z",
        "updated_at": "2025-01-20T14:45:00Z"
      },
      "prices": [
        {
          "id": "PRIC00000001",
          "variant_id": "PVAR00000001",
          "price_type": "MRP",
          "price": 99.00,
          "currency": "INR",
          "effective_from": "2025-01-01T00:00:00Z",
          "effective_to": null,
          "is_active": true,
          "created_at": "2025-01-15T10:40:00Z",
          "updated_at": "2025-01-15T10:40:00Z"
        },
        {
          "id": "PRIC00000002",
          "variant_id": "PVAR00000001",
          "price_type": "retail",
          "price": 89.00,
          "currency": "INR",
          "effective_from": "2025-01-01T00:00:00Z",
          "effective_to": null,
          "is_active": true,
          "created_at": "2025-01-15T10:40:00Z",
          "updated_at": "2025-01-15T10:40:00Z"
        }
      ],
      "stock": {
        "total_quantity": 500,
        "available_quantity": 450,
        "reserved_quantity": 50,
        "batches": [
          {
            "batch_id": "BATC00000001",
            "quantity": 500,
            "available_quantity": 450,
            "reserved_quantity": 50,
            "cost_price": 60.00,
            "expiry_date": "2025-03-15",
            "batch_number": "LOT-2025-001"
          }
        ]
      },
      "tax_config": {
        "hsn_code": "0702",
        "gst_rate": 5.0,
        "cgst_rate": 2.5,
        "sgst_rate": 2.5,
        "igst_rate": 5.0
      }
    }
  ],
  "discounts": [
    {
      "id": "DISC00000001",
      "name": "Winter Sale",
      "code": "WINTER25",
      "discount_type": "percentage",
      "discount_value": 10.0,
      "min_order_value": 500.0,
      "max_discount_amount": 100.0,
      "applicable_products": ["PROD00000001"],
      "applicable_categories": [],
      "applicable_warehouses": ["WRHS00000001"],
      "valid_from": "2025-01-01T00:00:00Z",
      "valid_to": "2025-01-31T23:59:59Z",
      "is_active": true,
      "usage_limit": 100,
      "used_count": 25
    }
  ],
  "payment_methods": [
    {
      "code": "cash",
      "name": "Cash",
      "is_enabled": true
    },
    {
      "code": "upi",
      "name": "UPI",
      "is_enabled": true
    },
    {
      "code": "online",
      "name": "Online Payment",
      "is_enabled": true
    }
  ],
  "sale_types": [
    {
      "code": "in_store",
      "name": "In-Store",
      "is_enabled": true
    },
    {
      "code": "delivery",
      "name": "Delivery",
      "is_enabled": true
    }
  ],
  "tax_config": {
    "apply_taxes_default": false,
    "warehouse_state": "Karnataka",
    "warehouse_state_code": "KA",
    "gst_rates_available": [0, 5, 12, 18, 28]
  },
  "metadata": {
    "total_products": 15,
    "total_variants": 42,
    "variants_in_stock": 38,
    "active_discounts": 2,
    "context_generated_at": "2025-01-20T15:00:00Z"
  }
}
```

##### 4. GET /api/v1/aggregation/purchase-orders/:id - PO Detail

**Query Parameters:**
- `include` - Comma-separated list: `collaborator`, `warehouse`, `items`, `grns`, `inventory`, `payments`

**Response:**
```json
{
  "purchase_order": {
    "id": "PORD00000001",
    "po_number": "PO-2025-0001",
    "collaborator_id": "CLAB00000001",
    "collaborator_name": "Organic Farms Ltd",
    "warehouse_id": "WRHS00000001",
    "warehouse_name": "Main Warehouse",
    "order_date": "2025-01-15",
    "expected_delivery_date": "2025-01-20",
    "actual_delivery_date": "2025-01-19",
    "status": "delivered",
    "total_amount": 50000.00,
    "total_rejected_amount": 2500.00,
    "amount_owed": 47500.00,
    "payment_status": "partial",
    "paid_amount": 25000.00,
    "is_inter_state": false,
    "items": [
      {
        "id": "POIM00000001",
        "po_id": "PORD00000001",
        "variant_id": "PVAR00000001",
        "product_name": "Organic Tomatoes - 500g",
        "product_sku": "TOM-500G-ORG",
        "quantity": 1000,
        "unit_price": 50.00,
        "line_total": 50000.00,
        "received_quantity": 950,
        "base_price": 47.62,
        "gst_rate": 5.0,
        "gst_amount": 2.38,
        "cgst_rate": 2.5,
        "cgst_amount": 1.19,
        "sgst_rate": 2.5,
        "sgst_amount": 1.19,
        "igst_rate": 0,
        "igst_amount": 0,
        "created_at": "2025-01-15T09:00:00Z"
      }
    ],
    "created_at": "2025-01-15T09:00:00Z",
    "updated_at": "2025-01-19T14:30:00Z"
  },
  "collaborator": {
    "id": "CLAB00000001",
    "company_name": "Organic Farms Ltd",
    "contact_person": "Rajesh Kumar",
    "contact_number": "9876543210",
    "email": "rajesh@organicfarms.com",
    "gst_number": "29AABCU9603R1ZM",
    "pan_number": "AABCU9603R",
    "bank_account_no": "1234567890123456",
    "bank_ifsc": "HDFC0001234",
    "bank_name": "HDFC Bank",
    "address_id": "ADDR00000001",
    "address": {
      "type": "business",
      "street": "123 Farm Road",
      "village": "Green Village",
      "taluk": "South Taluk",
      "district": "Bangalore Rural",
      "state": "Karnataka",
      "state_code": "KA",
      "pincode": "560001"
    },
    "is_active": true,
    "created_at": "2025-01-10T08:00:00Z",
    "updated_at": "2025-01-10T08:00:00Z"
  },
  "warehouse": {
    "id": "WRHS00000001",
    "name": "Main Warehouse",
    "address": {
      "type": "warehouse",
      "street": "456 Industrial Area",
      "village": null,
      "taluk": "North Taluk",
      "district": "Bangalore Urban",
      "state": "Karnataka",
      "state_code": "KA",
      "pincode": "560002"
    },
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  },
  "grns": [
    {
      "id": "GRNX00000001",
      "grn_number": "GRN-VENDOR-2025-001",
      "po_id": "PORD00000001",
      "warehouse_id": "WRHS00000001",
      "received_date": "2025-01-19",
      "quality_status": "partial",
      "remarks": "5% items damaged during transit",
      "grn_document": "ATCH00000001",
      "items": [
        {
          "id": "GRIT00000001",
          "grn_id": "GRNX00000001",
          "po_item_id": "POIM00000001",
          "variant_id": "PVAR00000001",
          "received_quantity": 950,
          "accepted_quantity": 900,
          "rejected_quantity": 50,
          "rejection_reason": "Damaged packaging",
          "expiry_date": "2025-03-15",
          "batch_number": "LOT-2025-001",
          "inventory_batch_id": "BATC00000001",
          "created_at": "2025-01-19T14:30:00Z"
        }
      ],
      "created_at": "2025-01-19T14:30:00Z",
      "updated_at": "2025-01-19T14:30:00Z"
    }
  ],
  "inventory_batches": [
    {
      "id": "BATC00000001",
      "warehouse_id": "WRHS00000001",
      "variant_id": "PVAR00000001",
      "total_quantity": 900,
      "available_quantity": 850,
      "reserved_quantity": 50,
      "cost_price": 50.00,
      "expiry_date": "2025-03-15",
      "batch_number": "LOT-2025-001",
      "grn_id": "GRNX00000001",
      "created_at": "2025-01-19T14:30:00Z",
      "updated_at": "2025-01-20T10:00:00Z"
    }
  ],
  "payments": [
    {
      "id": "BPMT00000001",
      "entity_type": "purchase_order",
      "entity_id": "PORD00000001",
      "amount": 25000.00,
      "payment_method": "bank_transfer",
      "transaction_reference": "UTR123456789",
      "payment_date": "2025-01-20",
      "notes": "Partial payment - first installment",
      "created_at": "2025-01-20T10:00:00Z"
    }
  ],
  "summary": {
    "ordered_value": 50000.00,
    "received_value": 47500.00,
    "rejected_value": 2500.00,
    "paid_amount": 25000.00,
    "pending_amount": 22500.00,
    "grn_count": 1,
    "batch_count": 1,
    "payment_count": 1,
    "completion_percentage": 95.0,
    "payment_percentage": 52.63
  }
}
```

##### 5. GET /api/v1/aggregation/inventory - Inventory List

**Query Parameters:**
- `warehouse_id` - Filter by warehouse
- `variant_id` - Filter by variant
- `product_id` - Filter by product (all variants)
- `expiring_within_days` - Filter batches expiring within N days
- `low_stock_threshold` - Filter items below threshold
- `include_zero_stock` - Include zero-quantity batches (default: false)
- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 20, max: 100)

**Response:**
```json
{
  "batches": [
    {
      "batch": {
        "id": "BATC00000001",
        "warehouse_id": "WRHS00000001",
        "variant_id": "PVAR00000001",
        "total_quantity": 900,
        "available_quantity": 850,
        "reserved_quantity": 50,
        "cost_price": 50.00,
        "expiry_date": "2025-03-15",
        "batch_number": "LOT-2025-001",
        "cgst_rate": 2.5,
        "sgst_rate": 2.5,
        "custom_tax_rate": 0,
        "created_at": "2025-01-19T14:30:00Z",
        "updated_at": "2025-01-20T10:00:00Z"
      },
      "warehouse": {
        "id": "WRHS00000001",
        "name": "Main Warehouse",
        "address": null,
        "created_at": "2025-01-01T00:00:00Z",
        "updated_at": "2025-01-01T00:00:00Z"
      },
      "variant": {
        "id": "PVAR00000001",
        "product_id": "PROD00000001",
        "variant_name": "500g Pack",
        "description": "Half kilogram pack",
        "quantity": "500g",
        "pack_size": "Small",
        "sku": "TOM-500G-ORG",
        "barcode": "8901234567890",
        "hsn_code": "0702",
        "gst_rate": 5.0,
        "collaborator_ids": ["CLAB00000001"],
        "brand_name": "FreshFarms",
        "images": null,
        "image_urls": null,
        "dosage_instructions": null,
        "usage_details": null,
        "prices": [],
        "is_active": true,
        "created_at": "2025-01-15T10:35:00Z",
        "updated_at": "2025-01-20T14:50:00Z"
      },
      "product": {
        "id": "PROD00000001",
        "name": "Organic Tomatoes",
        "description": "Fresh organic tomatoes",
        "category_id": "CAT000000001",
        "subcategory_id": "SCAT00000001",
        "variants": [],
        "created_at": "2025-01-15T10:30:00Z",
        "updated_at": "2025-01-20T14:45:00Z"
      },
      "days_until_expiry": 54,
      "is_expiring_soon": false,
      "is_low_stock": false
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 156,
    "total_pages": 8,
    "has_next": true,
    "has_previous": false
  },
  "summary": {
    "total_batches": 156,
    "total_quantity": 45000,
    "total_available": 42500,
    "total_reserved": 2500,
    "total_value": 2250000.00,
    "warehouses_count": 3,
    "variants_count": 42,
    "expiring_soon_count": 12,
    "low_stock_count": 5,
    "zero_stock_count": 3
  },
  "filters_applied": {
    "warehouse_id": null,
    "variant_id": null,
    "product_id": null,
    "expiring_within_days": null,
    "low_stock_threshold": null,
    "include_zero_stock": false
  }
}
```

#### Key Field Notes for Frontend

**ID Formats** (all use kisanlink-db hash IDs, NO underscores):
- Products: `PROD00000001`
- Variants: `PVAR00000001`
- Warehouses: `WRHS00000001`
- Collaborators: `CLAB00000001`
- Purchase Orders: `PORD00000001`
- PO Items: `POIM00000001`
- GRNs: `GRNX00000001`
- GRN Items: `GRIT00000001`
- Batches: `BATC00000001`
- Prices: `PRIC00000001`
- Discounts: `DISC00000001`
- Attachments: `ATCH00000001`
- Bank Payments: `BPMT00000001`

**GST Tax Fields**:
- `gst_rate`: Total GST rate (e.g., 5%, 12%, 18%, 28%)
- `cgst_rate`: Central GST (half of gst_rate for intra-state)
- `sgst_rate`: State GST (half of gst_rate for intra-state)
- `igst_rate`: Integrated GST (full rate for inter-state, else 0)
- `is_inter_state`: Boolean indicating IGST vs CGST+SGST

**Optional Fields** (may be `null`):
- `description`, `barcode`, `batch_number`
- `address` (if not populated from AAA)
- `effective_to` (open-ended prices)
- `image_urls` (if no images uploaded)
- `rejection_reason` (if fully accepted)

**Date Formats**:
- Timestamps: ISO 8601 with timezone (`2025-01-20T15:00:00Z`)
- Date-only fields: `YYYY-MM-DD` (`2025-01-20`)

---

## Frontend Migration Guide

### Step 0: Remove ALL Tax API Calls (CRITICAL)

**Delete all calls to tax endpoints - they will return 404:**
```javascript
// REMOVE ALL THESE - ENDPOINTS NO LONGER EXIST
api.post('/taxes', ...)                    // ❌ REMOVED
api.get('/taxes', ...)                     // ❌ REMOVED
api.get('/taxes/:id', ...)                 // ❌ REMOVED
api.patch('/taxes/:id', ...)               // ❌ REMOVED
api.delete('/taxes/:id', ...)              // ❌ REMOVED
api.get('/taxes/active', ...)              // ❌ REMOVED
api.get('/taxes/type/:type', ...)          // ❌ REMOVED
api.post('/taxes/:id/tiers', ...)          // ❌ REMOVED
api.get('/taxes/:id/tiers', ...)           // ❌ REMOVED
api.patch('/taxes/:id/tiers/:tierId', ...) // ❌ REMOVED
api.delete('/taxes/:id/tiers/:tierId', ..) // ❌ REMOVED
api.post('/taxes/calculate', ...)          // ❌ REMOVED
api.get('/taxes/hsn/:hsnCode', ...)        // ❌ REMOVED
```

**New Tax Approach:**
1. Set `hsn_code` and `gst_rate` when creating variants
2. Set `apply_taxes: true` when creating sales that need GST
3. Tax amounts appear automatically in sale response

---

### Step 1: Remove Collaborator-Product API Calls

**Delete all calls to these endpoints (they will return 404):**
```javascript
// REMOVE ALL THESE
api.post('/collaborators/:id/products', ...)
api.get('/collaborator-products', ...)
api.get('/collaborator-products/:id', ...)
api.patch('/collaborator-products/:id', ...)
api.delete('/collaborator-products/:id', ...)
```

### Step 2: Update Variant Creation

**Before (Beta):**
```javascript
// Create variant
const variant = await api.post('/product-variants', {
  product_id: 'PROD00000001',
  variant_name: '500g',
  quantity: '500g',
  collaborator_id: 'CLAB00000001',  // Single collaborator
  brand_name: 'Fresh Farms'
});

// Create prices separately
await api.post('/prices', {
  variant_id: variant.id,
  price_type: 'MRP',
  price: 199.99,
  currency: 'INR'
});
```

**After (Current):**
```javascript
// Create variant with prices AND multiple collaborators
const variant = await api.post(`/products/${productId}/variants`, {
  variant_name: '500g',
  quantity: '500g',
  collaborator_ids: ['CLAB00000001', 'CLAB00000002'],  // Array!
  brand_name: 'Fresh Farms',
  prices: [  // Prices included in creation
    { price_type: 'MRP', price: 199.99, currency: 'INR' },
    { price_type: 'retail', price: 179.99, currency: 'INR' }
  ]
});

// Prices automatically included in response
console.log(variant.prices);  // Array of ProductPriceResponse
```

### Step 3: Update Inventory Stock Checks

**Before (Beta):**
```javascript
const batch = await api.get(`/batches/${batchId}`);
const canSell = batch.total_quantity >= requestedQuantity;  // WRONG NOW!
```

**After (Current):**
```javascript
const batch = await api.get(`/batches/${batchId}`);
const canSell = batch.available_quantity >= requestedQuantity;  // CORRECT

// Display inventory status
console.log(`Total Stock: ${batch.total_quantity}`);
console.log(`Reserved: ${batch.reserved_quantity}`);
console.log(`Available for Sale: ${batch.available_quantity}`);
```

### Step 4: Update Sales Workflow

**Before (Beta):**
```javascript
// Sale created as completed immediately
const sale = await api.post('/sales', {
  warehouse_id: 'WREH00000001',
  items: [...]
});
// sale.status === 'completed'
```

**After (Current):**
```javascript
// Sale created as pending (inventory reserved)
const sale = await api.post('/sales', {
  warehouse_id: 'WREH00000001',
  payment_mode: 'cash',
  sale_type: 'in_store',
  apply_taxes: false,
  items: [...]
});
// sale.status === 'pending'

// After payment confirmation, complete the sale
const completedSale = await api.patch(`/sales/${sale.id}/status`, {
  status: 'completed'
});

// OR cancel the sale
await api.post(`/sales/${sale.id}/cancel`, {
  reason: 'customer_request',
  reason_details: 'Customer changed mind'
});

// OR cancel specific items (partial cancellation)
await api.post(`/sales/${sale.id}/cancel-items`, {
  reason: 'out_of_stock',
  items: [
    { sale_item_id: 'SITM00000001', quantity: 25 }
  ]
});
```

### Step 4.1: Update Customer Tracking in Sales (BREAKING CHANGE)

**Before (Beta) - Full CreateSaleRequest with `customer_id`:**
```json
{
  "warehouse_id": "WREH00000001",
  "sale_date": "2025-12-08",
  "customer_id": "CUST_123",
  "payment_mode": "cash",
  "sale_type": "in_store",
  "apply_taxes": false,
  "delivery_state": null,
  "discount_id": null,
  "coupon_code": null,
  "auto_apply_discounts": true,
  "items": [
    {
      "variant_id": "PVAR00000001",
      "quantity": 10
    }
  ]
}
```

**After (Current) - Full CreateSaleRequest with customer details:**
```json
{
  "warehouse_id": "WREH00000001",
  "sale_date": "2025-12-08",
  "customer_phone": "9876543210",
  "customer_name": "John Doe",
  "is_org_member": false,
  "payment_mode": "cash",
  "sale_type": "in_store",
  "apply_taxes": false,
  "delivery_state": null,
  "discount_id": null,
  "coupon_code": null,
  "auto_apply_discounts": true,
  "items": [
    {
      "variant_id": "PVAR00000001",
      "quantity": 10
    }
  ]
}
```

**CreateSaleRequest Field Reference:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `warehouse_id` | `string` | **Yes** | Warehouse ID where sale occurs |
| `sale_date` | `string` | No | Sale date (YYYY-MM-DD), defaults to today |
| `customer_phone` | `*string` | No | Customer phone number (replaces `customer_id`) |
| `customer_name` | `*string` | No | Customer name (replaces `customer_id`) |
| `is_org_member` | `bool` | No | Whether customer is FPO member (default: `false`) |
| `payment_mode` | `string` | **Yes** | Payment method: `cash`, `upi`, `online` |
| `sale_type` | `string` | **Yes** | Sale type: `in_store`, `delivery` |
| `apply_taxes` | `*bool` | No | Enable GST calculation (default: `false`) |
| `delivery_state` | `*string` | No | State code for inter-state IGST detection |
| `discount_id` | `*string` | No | Manual discount ID (highest priority) |
| `coupon_code` | `*string` | No | Coupon code (second priority) |
| `auto_apply_discounts` | `*bool` | No | Auto-discover discounts (default: `true`) |
| `items` | `array` | **Yes** | Array of sale items |

**CreateSaleItemRequest:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `variant_id` | `string` | **Yes** | Product variant ID |
| `quantity` | `int64` | **Yes** | Quantity (must be > 0) |

**JavaScript Migration Example:**
```javascript
// OLD: Using customer_id (external reference)
const sale = await api.post('/sales', {
  warehouse_id: 'WREH00000001',
  customer_id: 'CUST_123',  // ❌ REMOVED - No longer supported
  payment_mode: 'cash',
  sale_type: 'in_store',
  items: [{ variant_id: 'PVAR00000001', quantity: 10 }]
});
console.log(sale.customer_id);  // 'CUST_123'

// NEW: Using customer_phone, customer_name, is_org_member
const sale = await api.post('/sales', {
  warehouse_id: 'WREH00000001',
  customer_phone: '9876543210',    // Optional: Phone number
  customer_name: 'John Doe',        // Optional: Customer name
  is_org_member: false,             // FPO member status (default: false)
  payment_mode: 'cash',
  sale_type: 'in_store',
  apply_taxes: false,
  items: [{ variant_id: 'PVAR00000001', quantity: 10 }]
});
console.log(sale.customer_phone);  // '9876543210'
console.log(sale.customer_name);   // 'John Doe'
console.log(sale.is_org_member);   // false
```

**Migration Checklist:**
1. Replace `customer_id` with `customer_phone` and `customer_name` in all sale creation calls
2. Add `is_org_member` field (defaults to `false` for non-FPO customers)
3. Update response handling to use new fields instead of `customer_id`
4. Update any customer display components to show phone/name

**FPO Member Sales (Differential Pricing):**
```javascript
// Sale to FPO member (may get member pricing)
const memberSale = await api.post('/sales', {
  warehouse_id: 'WREH00000001',
  customer_phone: '9876543210',
  customer_name: 'Farmer Ramesh',
  is_org_member: true,  // ✅ FPO member
  payment_mode: 'cash',
  sale_type: 'in_store',
  apply_taxes: false,
  items: [{ variant_id: 'PVAR00000001', quantity: 5 }]
});
```

### Step 5: Use Aggregation API for Performance

**Before (Beta) - 5+ API calls:**
```javascript
const product = await api.get(`/products/${productId}`);
const variants = await api.get(`/product-variants?product_id=${productId}`);
const prices = await Promise.all(variants.map(v =>
  api.get(`/prices?variant_id=${v.id}`)
));
const batches = await Promise.all(variants.map(v =>
  api.get(`/batches?variant_id=${v.id}`)
));
// Manual data combination...
```

**After (Current) - 1 API call:**
```javascript
const productDetail = await api.get(
  `/aggregation/products/${productId}?include=variants,prices,inventory`
);

// All data included
console.log(productDetail.product);
console.log(productDetail.variants[0].prices);
console.log(productDetail.variants[0].stock_summary);
```

### Step 6: Update PO Status Handling

**Before (Beta):**
```javascript
const statusOrder = ['placed', 'confirmed', 'out_for_delivery', 'delivered', 'paid'];
```

**After (Current):**
```javascript
const statusOrder = ['placed', 'confirmed', 'out_for_delivery', 'delivered', 'verified', 'paid'];
// Added 'verified' status between 'delivered' and 'paid'
```

### Step 7: Use Categories for Product Organization (ID-BASED)

**Seed Categories (One-Time Admin Action):**
```javascript
// Call once to populate predefined categories
await api.post('/categories/seed');
// Returns { "message": "Seeded X categories and Y subcategories" }
```

**Get Categories for Dropdowns:**
```javascript
// Get all categories with their subcategories
const categories = await api.get('/categories/with-subcategories');
// Returns: [{ id: "CATG00000001", name: "FERTILIZERS", description: "Chemical and organic fertilizers", subcategories: [...] }, ...]

// Get subcategories for a selected category (CHANGED - now uses categoryId)
const subcategories = await api.get(`/subcategories/category/${categoryId}`);
// Returns: [{ id: "SCAT00000001", name: "WATER_SOLUBLE", description: "...", category_id: "CATG00000001" }, ...]
```

**Create Product with Category (ID-BASED):**
```javascript
// UPDATED: Use category_id and subcategory_id (NOT category_name/subcategory_name)
const product = await api.post('/products', {
  name: 'NPK 19-19-19',
  description: 'Balanced fertilizer for crops',
  category_id: 'CATG00000001',       // Optional (nullable)
  subcategory_id: 'SCAT00000002'     // Optional (nullable)
});
```

**Update Product Category:**
```javascript
// Update product to change category
const updated = await api.patch(`/products/${productId}`, {
  category_id: 'CATG00000003',       // New category ID
  subcategory_id: 'SCAT00000005'     // New subcategory ID
});
```

**MIGRATION NOTE:** If you were using `category_name` and `subcategory_name`, change to `category_id` and `subcategory_id`:
```javascript
// OLD (NO LONGER WORKS)
{ category_name: 'Fertilizers', subcategory_name: 'Water Soluble' }

// NEW (USE THIS)
{ category_id: 'CATG00000001', subcategory_id: 'SCAT00000002' }
```

**Predefined Categories (Seeded via `/categories/seed`):**

**Naming Convention:** ALL_CAPS_SNAKE_CASE (e.g., `WATER_SOLUBLE`, `BIO_PRODUCTS`)

**Idempotency:** Seeding is case-insensitive idempotent (won't create duplicates regardless of case)

| Category | Description | Subcategories |
|----------|-------------|---------------|
| SEEDS | Agricultural seeds and planting materials | - |
| FERTILIZERS | Chemical and organic fertilizers | BULK, WATER_SOLUBLE, MICRONUTRIENTS, MACRONUTRIENTS |
| PESTICIDES | Pest control products | WEEDICIDES, INSECTICIDES, FUNGICIDES, ORGANIC |
| BIO_PRODUCTS | Biological and eco-friendly products | BULK, LIQUIDS, OTHER |
| IMPLEMENTS | Agricultural tools and implements | WEEDING, SOWING, SPRAYERS |
| IRRIGATION | Irrigation equipment and systems | PIPES, DRIPPERS, SPRINKLERS, AUTOMATION_MACHINES, OTHER |
| OTHER | Miscellaneous agricultural products | - |

**Note:** Same subcategory name can exist in different categories (e.g., `BULK` exists in both `FERTILIZERS` and `BIO_PRODUCTS`). This is enforced by composite unique index: `UNIQUE(name, category_id)`.

---

## Commit History (Beta → Current)

```
4769aac feat: uses unified product prices table
480e9e5 fix: fixes the prices issue in products response body
e6a5ee7 fix: fixes the tests for the changed responses
126bdb0 feat: adds test cases for aggregate services
cdba137 feat: implements new aggregate api endpoints
1341e1c merge: bring reserved stock and partial cancellation
a8c999b doc: documented the implementation status
0aa2d29 fix: fixes the logical errors
b27d6e4 fix: fixes the sales service issues
4c9e103 feat: implements reserved stock for sales
05de9d4 fix: changes swagger tags in the aggregation endpoints
86caa14 feat: adds the aggregation endpoints handlers and routes
197b9d5 feat: merges beta into variants-fix development branch
9eba3ae deletion: deletes the collaborator-product endpoints
3f78d6c fix: fixes the optional sku json issue
160d3c0 feat: implements the feature to add prices to the variants
16fc979 feat: implements the feature to add multiple collaborators to same variant
9baafae fix: fixes the grn issue
e99c9d3 feat: implements grn rejection features
```

---

## Summary

**Total Files Changed:** 53+ files
**Lines Added:** 12,059+
**Lines Deleted:** 36,028+

**Key Changes:**
1. **TAX SYSTEM SIMPLIFIED TO GST-ONLY (13 endpoints removed)**
   - All `/api/v1/taxes/*` endpoints removed
   - GST rate stored on `ProductVariant.gst_rate`
   - Tax calculation controlled by `apply_taxes` field on sales
   - Inventory batches no longer have tax configuration
2. Collaborator-product endpoints completely removed
3. Multiple collaborators per variant supported
4. Prices included in variant responses
5. Reserved/available quantity tracking in inventory
6. Pending status workflow for sales
7. Partial sale cancellation support
8. GRN rejection tracking with return status
9. New aggregation API for performance optimization
10. **Categories & Subcategories system for product organization**

**GST Tax Quick Reference:**
| Where | Field | Purpose |
|-------|-------|---------|
| Variant | `hsn_code` | HSN code for GST classification |
| Variant | `gst_rate` | GST rate (0, 5, 12, 18, or 28) |
| Sale | `apply_taxes` | Enable/disable GST calculation |
| Sale Response | `cgst_amount`, `sgst_amount`, `igst_amount` | Calculated tax amounts |

**Questions?** Contact the backend team.
