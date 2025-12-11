# Frontend API Changes Catalog - December 11, 2025

## Issue 1: SKU Auto-Generation

**Type**: Enhancement

**Changes**:
- `POST /api/v1/products/{id}/variants` - SKU field removed from request body
- SKU is now auto-generated with pattern: `SKU-{CATEGORY_CODE}-{8-CHAR-HASH}`
- Category code: First 3 characters of category name (uppercase), defaults to "OTH"
- SKU is returned in response body

**Request Change**:
```json
// BEFORE
{
  "variant_name": "1kg Pack",
  "sku": "CUSTOM-SKU-001",  // REMOVED - no longer accepted
  "quantity": "1",
  "pack_size": "kg"
}

// AFTER
{
  "variant_name": "1kg Pack",
  "quantity": "1",
  "pack_size": "kg"
}
```

**Response** (includes auto-generated SKU):
```json
{
  "id": "PVAR00000001",
  "sku": "SKU-VEG-00000001",
  "variant_name": "1kg Pack"
}
```

---

## Issue 2: Sales List Optimization (BREAKING CHANGE)

**Type**: Breaking Change

**Changes**:
- `GET /api/v1/sales` - `items` array REMOVED from list response
- `GET /api/v1/sales/{id}` - Full details with items still available
- Performance improvement: 10x faster list queries

**Response Change**:
```json
// BEFORE - GET /api/v1/sales
{
  "data": [
    {
      "id": "SALE00000001",
      "invoice_number": "12250001",
      "total_amount": 500.00,
      "items": [...],        // REMOVED
      "breakdown": {...}     // REMOVED
    }
  ]
}

// AFTER - GET /api/v1/sales
{
  "data": [
    {
      "id": "SALE00000001",
      "invoice_number": "12250001",
      "total_amount": 500.00
      // items and breakdown NOT included
    }
  ]
}
```

**Migration**:
- Update sales list component to NOT expect `items` array
- Add click handler to fetch details via `GET /api/v1/sales/{id}`

---

## Issue 3: Availability GST Details

**Type**: Enhancement

**Changes**:
- `GET /api/v1/products/availability` - Added GST details to response
- New fields: `hsn_code`, `gst_rate`, `cgst_rate`, `sgst_rate`
- CGST and SGST rates are automatically calculated as GSTRate / 2

**Response Change**:
```json
// BEFORE
{
  "sku": "SKU-VEG-00000001",
  "variant_id": "PVAR00000001",
  "product_name": "Tomato 1kg",
  "total_quantity": 100,
  "warehouse_details": [...]
}

// AFTER
{
  "sku": "SKU-VEG-00000001",
  "variant_id": "PVAR00000001",
  "product_name": "Tomato 1kg",
  "total_quantity": 100,
  "warehouse_details": [...],
  "hsn_code": "07020000",    // NEW - HSN code for GST classification
  "gst_rate": 5.0,           // NEW - Total GST rate (0, 5, 12, 18, 28)
  "cgst_rate": 2.5,          // NEW - Central GST rate (gst_rate / 2)
  "sgst_rate": 2.5           // NEW - State GST rate (gst_rate / 2)
}
```

**Usage**:
- Use `hsn_code` for GST invoice generation
- Use `cgst_rate` and `sgst_rate` for intra-state sales
- For inter-state sales, use `gst_rate` as IGST

---

