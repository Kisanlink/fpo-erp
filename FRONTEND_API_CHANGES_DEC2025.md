# Frontend API Changes - December 2025

## Breaking Changes Summary

| Change | Impact | Action Required | Deadline |
|--------|--------|-----------------|----------|
| Availability response format | **HIGH** | Update to handle SKU grouping | Immediate |
| Price precision | **LOW** | No action (display only) | None |
| PO GST line totals | **MEDIUM** | Use new total fields | Optional |

---

## Table of Contents
1. [New Endpoints](#new-endpoints)
2. [Modified Response Formats](#modified-response-formats)
3. [Request Format Changes](#request-format-changes)
4. [Migration Guide](#migration-guide)
5. [Example Code](#example-code)

---

## New Endpoints

### 1. GET /api/v1/products/by-quantity

**Purpose**: Filter products by total inventory quantity across all warehouses

**Authentication**: Required (Bearer token)

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `min` | integer | ✅ Yes | - | Minimum quantity (inclusive) |
| `max` | integer | ✅ Yes | - | Maximum quantity (inclusive) |
| `limit` | integer | ❌ No | 50 | Page size (max: 200) |
| `offset` | integer | ❌ No | 0 | Number of records to skip |

**Example Requests**:
```javascript
// Low stock products (0-10 units)
GET /api/v1/products/by-quantity?min=0&max=10&limit=50&offset=0

// Optimal stock (50-200 units)
GET /api/v1/products/by-quantity?min=50&max=200

// Overstocked (500+ units)
GET /api/v1/products/by-quantity?min=500&max=999999
```

**Response Format**:
```json
{
  "success": true,
  "message": "Products filtered by quantity",
  "data": [
    {
      "id": "PROD00000001",
      "name": "Basmati Rice",
      "description": "Premium quality rice",
      "category_id": "CATG00000001",
      "is_active": true,
      "total_inventory_quantity": 75,  // Aggregated across all warehouses/batches
      "created_at": "2025-01-15T10:00:00Z",
      "updated_at": "2025-12-01T15:30:00Z"
    },
    {
      "id": "PROD00000002",
      "name": "Wheat Flour",
      "total_inventory_quantity": 120
    }
  ],
  "total": 2,
  "limit": 50,
  "offset": 0
}
```

**Use Cases**:
```javascript
// Low Stock Alert Dashboard
async function getLowStockProducts() {
  const response = await fetch(
    '/api/v1/products/by-quantity?min=0&max=10',
    { headers: { 'Authorization': `Bearer ${token}` } }
  );
  const data = await response.json();

  // Show alert badge if any products
  if (data.total > 0) {
    showLowStockAlert(data.data);
  }
}

// Inventory Report Filter
function filterByStockLevel(level) {
  const ranges = {
    'out-of-stock': { min: 0, max: 0 },
    'low': { min: 1, max: 10 },
    'optimal': { min: 11, max: 100 },
    'overstocked': { min: 101, max: 999999 }
  };

  const range = ranges[level];
  return fetch(`/api/v1/products/by-quantity?min=${range.min}&max=${range.max}`);
}
```

**Error Responses**:
```json
// Missing required parameter
{
  "success": false,
  "message": "Query parameter 'min' is required",
  "error": null
}

// Invalid parameter type
{
  "success": false,
  "message": "Invalid 'min' parameter - must be a valid integer",
  "error": "strconv.ParseInt: parsing \"abc\": invalid syntax"
}
```

---

## Modified Response Formats

### 1. Product Availability Endpoint (BREAKING CHANGE)

**Endpoint**: GET `/api/v1/products/availability`

**What Changed**: Response now groups batches by SKU instead of returning flat list

**Migration Required**: ✅ YES - Update frontend parsing logic

#### Before (Flat List - DEPRECATED)
```json
[
  {
    "id": "BATC00000001",
    "warehouse_id": "WRHS00000001",
    "warehouse_name": "Main Warehouse",
    "product_sku": "RICE-1KG",
    "product_name": "Basmati Rice 1kg",
    "quantity": 500,
    "cost_price": 50.00,
    "expiry_date": "2025-06-15"
  },
  {
    "id": "BATC00000002",
    "warehouse_id": "WRHS00000001",
    "warehouse_name": "Main Warehouse",
    "product_sku": "RICE-1KG",   // DUPLICATE SKU!
    "product_name": "Basmati Rice 1kg",
    "quantity": 400,
    "cost_price": 52.00,
    "expiry_date": "2025-07-20"
  },
  {
    "id": "BATC00000003",
    "warehouse_id": "WRHS00000002",
    "warehouse_name": "Branch Warehouse",
    "product_sku": "RICE-1KG",   // DUPLICATE SKU!
    "product_name": "Basmati Rice 1kg",
    "quantity": 300,
    "cost_price": 51.00,
    "expiry_date": "2025-05-10"
  }
]
```

**Problems with Old Format**:
- ❌ Multiple entries for same product (frontend had to aggregate manually)
- ❌ No total quantity visibility
- ❌ Expired batches counted as available stock
- ❌ No warehouse-level breakdown

#### After (Grouped by SKU - CURRENT)
```json
[
  {
    "sku": "RICE-1KG",
    "variant_id": "PVAR00000001",
    "product_name": "Basmati Rice 1kg",
    "product_description": "Premium quality basmati rice",
    "total_quantity": 1200,        // Total AVAILABLE (non-expired) across all warehouses
    "expired_quantity": 100,       // Total EXPIRED across all warehouses
    "earliest_expiry": "2025-05-10",  // Earliest expiry across all warehouses
    "expiry_status": "expiring_soon",  // "fresh" | "expiring_soon" | "expired"
    "warehouse_details": [
      {
        "warehouse_id": "WRHS00000001",
        "warehouse_name": "Main Warehouse",
        "address": {
          "id": "ADDR_12345678",
          "state": "Karnataka",
          "district": "Bangalore",
          "full_address": "123 Main St, Bangalore, Karnataka - 560001"
        },
        "quantity": 900,           // Available (non-expired) in this warehouse
        "expired_quantity": 50,    // Expired in this warehouse
        "earliest_expiry": "2025-06-15",
        "expiry_status": "fresh"
      },
      {
        "warehouse_id": "WRHS00000002",
        "warehouse_name": "Branch Warehouse",
        "address": {
          "id": "ADDR_87654321",
          "state": "Karnataka",
          "district": "Mysore"
        },
        "quantity": 300,
        "expired_quantity": 50,
        "earliest_expiry": "2025-05-10",
        "expiry_status": "expiring_soon"  // < 30 days until expiry
      }
    ]
  },
  {
    "sku": "WHEAT-1KG",
    "variant_id": "PVAR00000002",
    "product_name": "Wheat Flour 1kg",
    "total_quantity": 500,
    "expired_quantity": 0,
    "earliest_expiry": "2026-01-15",
    "expiry_status": "fresh",
    "warehouse_details": [
      {
        "warehouse_id": "WRHS00000001",
        "warehouse_name": "Main Warehouse",
        "quantity": 500,
        "expired_quantity": 0,
        "earliest_expiry": "2026-01-15",
        "expiry_status": "fresh"
      }
    ]
  }
]
```

**Benefits of New Format**:
- ✅ One entry per product (no duplicates)
- ✅ Clear total quantity at product level
- ✅ Expired vs available stock separated
- ✅ Warehouse-level breakdown in nested array
- ✅ Expiry status indicators
- ✅ Earliest expiry date tracking
- ✅ Address information included

**Expiry Status Values**:
| Status | Description | Color Code |
|--------|-------------|------------|
| `fresh` | More than 30 days until expiry | 🟢 Green |
| `expiring_soon` | 30 days or less until expiry | 🟡 Yellow |
| `expired` | Past expiry date | 🔴 Red |

---

### 2. Inventory Batch Response (NEW FIELDS)

**Endpoint**: GET `/api/v1/batches/:id`

**New Fields Added**:
```json
{
  "id": "BATC00000001",
  "warehouse_id": "WRHS00000001",
  "variant_id": "PVAR00000001",
  "cost_price": 50.00,
  "expiry_date": "2025-01-15",
  "total_quantity": 100,
  "reserved_quantity": 10,
  "available_quantity": 90,

  // NEW: Expiry status indicators
  "is_expired": false,                 // true if expiry_date < current date
  "expiry_status": "expiring_soon",    // "fresh" | "expiring_soon" | "expired"

  "created_at": "2025-11-01T10:00:00Z",
  "updated_at": "2025-12-01T15:30:00Z"
}
```

**Usage in UI**:
```javascript
function renderBatchCard(batch) {
  // Show expiry badge
  const expiryBadge = {
    'fresh': { color: 'green', icon: '✓', text: 'Fresh' },
    'expiring_soon': { color: 'orange', icon: '⚠', text: 'Expiring Soon' },
    'expired': { color: 'red', icon: '✗', text: 'Expired' }
  }[batch.expiry_status];

  return `
    <div class="batch-card ${batch.is_expired ? 'disabled' : ''}">
      <span class="badge ${expiryBadge.color}">
        ${expiryBadge.icon} ${expiryBadge.text}
      </span>
      <p>Quantity: ${batch.available_quantity}</p>
      <p>Expires: ${batch.expiry_date}</p>
    </div>
  `;
}

// Filter out expired batches
const activeBatches = batches.filter(b => !b.is_expired);
```

---

### 3. Purchase Order Items (NEW GST TOTALS)

**Endpoint**: GET `/api/v1/purchase-orders/:id`

**New Fields Added**:
```json
{
  "id": "PORD00000001",
  "po_number": "PO-2025-0001",
  "total_amount": 5250.00,
  "items": [
    {
      "id": "POIM00000001",
      "variant_id": "PVAR00000001",
      "quantity": 10,
      "unit_price": 525.00,      // ALL-IN price (includes GST)
      "line_total": 5250.00,      // unit_price × quantity

      // Per-unit GST breakdown (existing)
      "base_price": 500.00,       // Price before GST
      "gst_rate": 5.00,
      "gst_amount": 25.00,        // GST per unit
      "cgst_amount": 12.50,       // CGST per unit (if intra-state)
      "sgst_amount": 12.50,       // SGST per unit (if intra-state)
      "igst_amount": 0.00,        // IGST per unit (if inter-state)

      // NEW: Line item GST totals (per-unit × quantity)
      "gst_amount_total": 250.00,   // 25.00 × 10
      "cgst_amount_total": 125.00,  // 12.50 × 10
      "sgst_amount_total": 125.00,  // 12.50 × 10
      "igst_amount_total": 0.00     // 0.00 × 10
    }
  ]
}
```

**Usage in Invoice Display**:
```javascript
function renderPOInvoice(purchaseOrder) {
  let totalGST = 0;
  let totalCGST = 0;
  let totalSGST = 0;

  const itemsHTML = purchaseOrder.items.map(item => {
    // Use new total fields instead of calculating manually
    totalGST += item.gst_amount_total;
    totalCGST += item.cgst_amount_total;
    totalSGST += item.sgst_amount_total;

    return `
      <tr>
        <td>${item.product_name}</td>
        <td>${item.quantity}</td>
        <td>₹${item.base_price.toFixed(2)}</td>
        <td>₹${item.gst_amount_total.toFixed(2)}</td>  <!-- Use total, not per-unit -->
        <td>₹${item.line_total.toFixed(2)}</td>
      </tr>
    `;
  }).join('');

  return `
    <table>
      ${itemsHTML}
    </table>
    <div class="summary">
      <p>Total CGST: ₹${totalCGST.toFixed(2)}</p>
      <p>Total SGST: ₹${totalSGST.toFixed(2)}</p>
      <p>Grand Total: ₹${purchaseOrder.total_amount.toFixed(2)}</p>
    </div>
  `;
}
```

---

### 4. Price Precision (ALL PRICE FIELDS)

**Affected Endpoints**: All endpoints returning price data

**What Changed**: All price fields now rounded to 2 decimal places

**Examples**:
```json
// Product prices
{
  "id": "PRIC00000001",
  "price": 99.99    // Before: 99.9987
}

// Inventory batches
{
  "id": "BATC00000001",
  "cost_price": 50.00   // Before: 49.9999
}

// Sales
{
  "id": "SALE00000001",
  "total_amount": 1234.56   // Before: 1234.5678
}

// Purchase orders
{
  "id": "PORD00000001",
  "unit_price": 525.00   // Before: 524.9999
}
```

**Impact**: Display-only change, no code changes required

**CSS Recommendation**:
```css
/* Always display 2 decimal places in price fields */
.price {
  font-variant-numeric: tabular-nums;  /* Align digits */
}

/* JavaScript formatting (if needed) */
const formatPrice = (price) => price.toFixed(2);
```

---

## Request Format Changes

### 1. Cancel Items Endpoint (JSON KEY FORMAT)

**Endpoint**: POST `/api/v1/sales/:id/cancel-items`

**Critical**: Use **snake_case** JSON keys (NOT PascalCase)

**Correct Format** ✅:
```json
{
  "reason": "pricing_error",
  "performed_by": "USER00000001",
  "items": [
    {
      "sale_item_id": "SITM00000003",
      "quantity": 1
    },
    {
      "sale_item_id": "SITM00000005",
      "quantity": 2
    }
  ]
}
```

**Wrong Format** ❌:
```json
{
  "Reason": "pricing_error",         // Wrong: PascalCase
  "PerformedBy": "USER00000001",     // Wrong: PascalCase
  "Items": [                         // Wrong: PascalCase
    {
      "SaleItemID": "SITM00000003",  // Wrong: PascalCase
      "Quantity": 1                  // Wrong: PascalCase
    }
  ]
}
```

**Error Response** (if wrong format):
```json
{
  "success": false,
  "message": "Invalid request data",
  "error": "Key: 'CancelSaleItemsRequest.reason' Error:Field validation for 'reason' failed on the 'required' tag"
}
```

**JavaScript Example**:
```javascript
async function cancelSaleItems(saleId, itemsToCcancel, reason) {
  const response = await fetch(`/api/v1/sales/${saleId}/cancel-items`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      reason: reason,                    // snake_case ✅
      performed_by: getCurrentUserId(),  // snake_case ✅
      items: itemsToCancel.map(item => ({
        sale_item_id: item.id,           // snake_case ✅
        quantity: item.quantity          // snake_case ✅
      }))
    })
  });

  if (!response.ok) {
    const error = await response.json();
    console.error('Cancel failed:', error);
  }
}
```

---

### 2. Collaborator Creation (RELAXED VALIDATION)

**Endpoint**: POST `/api/v1/collaborators`

**What Changed**: Bank details now optional (can be added later)

**Before** (Required):
```json
{
  "company_name": "ABC Suppliers",
  "contact_person": "John Doe",
  "contact_number": "9876543210",
  "gst_number": "29ABCDE1234F1Z5",
  "bank_account_no": "1234567890",      // REQUIRED ❌
  "bank_ifsc": "HDFC0001234",           // REQUIRED ❌
  "bank_name": "HDFC Bank"
}
```

**After** (Optional):
```json
{
  "company_name": "ABC Suppliers",
  "contact_person": "John Doe",
  "contact_number": "9876543210",
  "gst_number": "29ABCDE1234F1Z5"
  // bank details can be omitted during creation ✅
}
```

**Add Bank Details Later**:
```javascript
// Create collaborator without bank details
const createResponse = await fetch('/api/v1/collaborators', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
  body: JSON.stringify({
    company_name: 'ABC Suppliers',
    contact_person: 'John Doe',
    contact_number: '9876543210',
    gst_number: '29ABCDE1234F1Z5'
    // No bank details
  })
});

const collaborator = await createResponse.json();
console.log('Created:', collaborator.data.id);

// Update with bank details later (separate API call)
await fetch(`/api/v1/collaborators/${collaborator.data.id}`, {
  method: 'PATCH',
  headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
  body: JSON.stringify({
    bank_account_no: '1234567890',
    bank_ifsc: 'HDFC0001234',
    bank_name: 'HDFC Bank'
  })
});
```

---

### 3. Logo Attachment Support

**Endpoints**:
- POST `/api/v1/collaborators`
- PATCH `/api/v1/collaborators/:id`

**What Changed**: `logo` field now accepts attachment IDs from attachment management system

**Workflow**:
```javascript
// Step 1: Upload logo file
async function uploadLogo(file) {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('entity_type', 'logo');
  formData.append('entity_id', 'settings');  // or collaborator ID

  const response = await fetch('/api/v1/attachments', {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` },
    body: formData
  });

  const attachment = await response.json();
  return attachment.data.id;  // Returns: "ATCH00000001"
}

// Step 2: Create/update collaborator with attachment ID
async function saveCollaborator(logoFile) {
  // Upload logo first
  const attachmentId = await uploadLogo(logoFile);

  // Create collaborator with attachment reference
  const response = await fetch('/api/v1/collaborators', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      company_name: 'ABC FPO',
      contact_person: 'John Doe',
      contact_number: '9876543210',
      logo: attachmentId  // Store attachment ID, not file path
    })
  });

  return response.json();
}

// Step 3: Display logo using presigned URL
async function displayLogo(collaborator) {
  if (!collaborator.logo) return;

  // Get presigned URL for display
  const urlResponse = await fetch(`/api/v1/attachments/${collaborator.logo}/url`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });

  const urlData = await urlResponse.json();

  // Use presigned URL in img tag
  document.querySelector('#logo').src = urlData.data.url;
}
```

**Benefits**:
- ✅ Consistent with other attachment patterns (PO docs, GRN PDFs)
- ✅ S3 file management (versioning, cleanup)
- ✅ Secure presigned URLs (time-limited access)
- ✅ Entity-based folder structure

---

## Migration Guide

### Priority 1: Availability Endpoint (CRITICAL)

**Impact**: HIGH - Response structure completely changed

**Timeline**: Update within 1 week

**Steps**:

1. **Identify Usage**:
```bash
# Search codebase for availability endpoint usage
grep -r "products/availability" src/
```

2. **Update Parsing Logic**:

**Before** (Flat array iteration):
```javascript
// OLD CODE - WILL BREAK
function displayAvailability(data) {
  data.forEach(batch => {
    renderProductRow({
      sku: batch.product_sku,
      warehouse: batch.warehouse_name,
      quantity: batch.quantity
    });
  });
}
```

**After** (Nested structure):
```javascript
// NEW CODE - WORKS WITH GROUPED RESPONSE
function displayAvailability(data) {
  data.forEach(product => {
    // Product-level totals
    const totalQuantity = product.total_quantity;
    const expiredQuantity = product.expired_quantity;

    // Warehouse breakdown
    product.warehouse_details.forEach(warehouse => {
      renderProductRow({
        sku: product.sku,
        productName: product.product_name,
        warehouseName: warehouse.warehouse_name,
        quantity: warehouse.quantity,           // Available in this warehouse
        expiredQuantity: warehouse.expired_quantity,  // Expired in this warehouse
        expiryStatus: warehouse.expiry_status,
        earliestExpiry: warehouse.earliest_expiry
      });
    });
  });
}
```

3. **Add Expiry Status Indicators**:
```javascript
function getExpiryBadge(status) {
  const badges = {
    'fresh': { color: 'green', icon: '✓', text: 'Fresh' },
    'expiring_soon': { color: 'orange', icon: '⚠', text: 'Expiring Soon' },
    'expired': { color: 'red', icon: '✗', text: 'Expired' }
  };
  return badges[status];
}

// In template
<span class="badge badge-${getExpiryBadge(product.expiry_status).color}">
  ${getExpiryBadge(product.expiry_status).icon}
  ${getExpiryBadge(product.expiry_status).text}
</span>
```

4. **Handle Expired Quantities**:
```javascript
function renderInventorySummary(product) {
  return `
    <div class="inventory-summary">
      <h3>${product.product_name} (${product.sku})</h3>

      <div class="quantity-breakdown">
        <div class="available">
          <label>Available:</label>
          <span class="value">${product.total_quantity}</span>
        </div>

        ${product.expired_quantity > 0 ? `
          <div class="expired">
            <label>Expired:</label>
            <span class="value text-danger">${product.expired_quantity}</span>
            <span class="help-text">Not available for sale</span>
          </div>
        ` : ''}
      </div>

      <table class="warehouse-details">
        <thead>
          <tr>
            <th>Warehouse</th>
            <th>Available</th>
            <th>Expired</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          ${product.warehouse_details.map(wh => `
            <tr>
              <td>${wh.warehouse_name}</td>
              <td>${wh.quantity}</td>
              <td>${wh.expired_quantity}</td>
              <td>
                <span class="badge badge-${getExpiryBadge(wh.expiry_status).color}">
                  ${wh.expiry_status}
                </span>
              </td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    </div>
  `;
}
```

---

### Priority 2: Use New Quantity Filter (OPTIONAL)

**Impact**: MEDIUM - New functionality

**Timeline**: Implement when needed

**Implementation**:

```javascript
// Inventory Dashboard Component
class InventoryDashboard {
  constructor() {
    this.filters = {
      'out-of-stock': { min: 0, max: 0 },
      'low-stock': { min: 1, max: 10 },
      'optimal': { min: 11, max: 100 },
      'overstocked': { min: 101, max: 999999 }
    };
  }

  async loadProducts(filter = 'all', page = 0, pageSize = 50) {
    let url = `/api/v1/products?limit=${pageSize}&offset=${page * pageSize}`;

    // Apply quantity filter if selected
    if (filter !== 'all' && this.filters[filter]) {
      const range = this.filters[filter];
      url = `/api/v1/products/by-quantity?min=${range.min}&max=${range.max}&limit=${pageSize}&offset=${page * pageSize}`;
    }

    const response = await fetch(url, {
      headers: { 'Authorization': `Bearer ${this.token}` }
    });

    const data = await response.json();
    this.renderProducts(data.data);
    this.renderPagination(data.total, page, pageSize);
  }

  renderFilterButtons() {
    return `
      <div class="filter-buttons">
        <button onclick="dashboard.loadProducts('all')">
          All Products
        </button>
        <button onclick="dashboard.loadProducts('out-of-stock')" class="btn-danger">
          Out of Stock
        </button>
        <button onclick="dashboard.loadProducts('low-stock')" class="btn-warning">
          Low Stock (1-10)
        </button>
        <button onclick="dashboard.loadProducts('optimal')" class="btn-success">
          Optimal (11-100)
        </button>
        <button onclick="dashboard.loadProducts('overstocked')" class="btn-info">
          Overstocked (100+)
        </button>
      </div>
    `;
  }
}
```

---

### Priority 3: Update PO Invoice Display (MEDIUM)

**Impact**: MEDIUM - Better tax breakdown

**Timeline**: Update when redesigning PO UI

**Changes**:

```javascript
// Before: Manual calculation
function calculateLineTotals(items) {
  return items.map(item => ({
    ...item,
    gst_total: item.gst_amount * item.quantity  // Manual calculation
  }));
}

// After: Use provided totals
function renderPOInvoice(purchaseOrder) {
  const items = purchaseOrder.items.map(item => ({
    ...item,
    // Use pre-calculated totals from API
    gst_total: item.gst_amount_total,
    cgst_total: item.cgst_amount_total,
    sgst_total: item.sgst_amount_total
  }));

  return renderInvoiceTable(items);
}
```

---

## Example Code

### Complete Availability Component

```javascript
// React Component Example
import React, { useState, useEffect } from 'react';
import { Badge, Table, Alert } from 'react-bootstrap';

function ProductAvailabilityTable() {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadAvailability();
  }, []);

  async function loadAvailability() {
    try {
      const response = await fetch('/api/v1/products/availability', {
        headers: { 'Authorization': `Bearer ${getToken()}` }
      });

      if (!response.ok) throw new Error('Failed to load availability');

      const data = await response.json();
      setProducts(data.data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  function getExpiryBadgeVariant(status) {
    const variants = {
      'fresh': 'success',
      'expiring_soon': 'warning',
      'expired': 'danger'
    };
    return variants[status] || 'secondary';
  }

  if (loading) return <div>Loading availability...</div>;
  if (error) return <Alert variant="danger">{error}</Alert>;

  return (
    <div className="availability-container">
      <h2>Product Availability</h2>

      {products.map(product => (
        <div key={product.sku} className="product-card mb-4">
          <div className="product-header">
            <h4>{product.product_name}</h4>
            <span className="text-muted">SKU: {product.sku}</span>
            <Badge bg={getExpiryBadgeVariant(product.expiry_status)}>
              {product.expiry_status.replace('_', ' ').toUpperCase()}
            </Badge>
          </div>

          <div className="quantity-summary">
            <div className="stat">
              <label>Available Stock:</label>
              <span className="value text-success">{product.total_quantity}</span>
            </div>

            {product.expired_quantity > 0 && (
              <div className="stat">
                <label>Expired Stock:</label>
                <span className="value text-danger">{product.expired_quantity}</span>
              </div>
            )}

            {product.earliest_expiry && (
              <div className="stat">
                <label>Earliest Expiry:</label>
                <span className="value">{product.earliest_expiry}</span>
              </div>
            )}
          </div>

          <Table striped bordered hover size="sm">
            <thead>
              <tr>
                <th>Warehouse</th>
                <th>Available</th>
                <th>Expired</th>
                <th>Expiry Date</th>
                <th>Status</th>
                <th>Location</th>
              </tr>
            </thead>
            <tbody>
              {product.warehouse_details.map(warehouse => (
                <tr key={warehouse.warehouse_id}>
                  <td>{warehouse.warehouse_name}</td>
                  <td className="text-end">{warehouse.quantity}</td>
                  <td className="text-end text-danger">
                    {warehouse.expired_quantity > 0 ? warehouse.expired_quantity : '-'}
                  </td>
                  <td>{warehouse.earliest_expiry || '-'}</td>
                  <td>
                    <Badge bg={getExpiryBadgeVariant(warehouse.expiry_status)}>
                      {warehouse.expiry_status}
                    </Badge>
                  </td>
                  <td className="text-muted small">
                    {warehouse.address?.full_address || 'No address'}
                  </td>
                </tr>
              ))}
            </tbody>
          </Table>
        </div>
      ))}
    </div>
  );
}

export default ProductAvailabilityTable;
```

---

### Quantity Filter Integration

```javascript
// Inventory Filter Hook
import { useState, useEffect } from 'react';

function useInventoryFilter() {
  const [filter, setFilter] = useState('all');
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(false);

  const filterRanges = {
    'all': null,
    'out-of-stock': { min: 0, max: 0 },
    'low-stock': { min: 1, max: 10 },
    'optimal': { min: 11, max: 100 },
    'overstocked': { min: 101, max: 999999 }
  };

  useEffect(() => {
    loadProducts();
  }, [filter]);

  async function loadProducts() {
    setLoading(true);

    try {
      const range = filterRanges[filter];
      const url = range
        ? `/api/v1/products/by-quantity?min=${range.min}&max=${range.max}`
        : `/api/v1/products`;

      const response = await fetch(url, {
        headers: { 'Authorization': `Bearer ${getToken()}` }
      });

      const data = await response.json();
      setProducts(data.data);
    } catch (error) {
      console.error('Failed to load products:', error);
    } finally {
      setLoading(false);
    }
  }

  return { filter, setFilter, products, loading, filterRanges };
}

// Usage in component
function InventoryDashboard() {
  const { filter, setFilter, products, loading, filterRanges } = useInventoryFilter();

  return (
    <div>
      <div className="filter-buttons">
        {Object.keys(filterRanges).map(key => (
          <button
            key={key}
            onClick={() => setFilter(key)}
            className={filter === key ? 'active' : ''}
          >
            {key.replace('-', ' ').toUpperCase()}
          </button>
        ))}
      </div>

      {loading ? (
        <div>Loading...</div>
      ) : (
        <ProductGrid products={products} />
      )}
    </div>
  );
}
```

---

## Summary

### Action Items by Priority

**Immediate (1 week)**:
- ✅ Update availability endpoint parsing logic
- ✅ Add expiry status UI indicators
- ✅ Handle expired quantity display

**Optional (when needed)**:
- ⚠️ Integrate quantity filter endpoint
- ⚠️ Use PO GST line totals in invoices
- ⚠️ Implement logo attachment workflow

**No Action Required**:
- ✅ Price precision (automatic)
- ✅ Collaborator bank validation (relaxed)

### Testing Checklist

- [ ] Availability page loads without errors
- [ ] Product totals match warehouse breakdowns
- [ ] Expired quantities shown separately
- [ ] Expiry status badges display correctly
- [ ] Warehouse address information renders
- [ ] Quantity filter works for all ranges
- [ ] PO invoice shows GST line totals
- [ ] Logo upload and display works
- [ ] Cancel items uses snake_case JSON
- [ ] Collaborators can be created without bank details

---

## Support

For issues or questions:
1. Check IMPLEMENTATION_CHANGES_DEC2025.md for technical details
2. Review API_DOCUMENTATION.md for complete API reference
3. Contact backend team for clarifications
