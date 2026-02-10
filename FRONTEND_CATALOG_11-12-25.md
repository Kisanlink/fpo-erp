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

## Issue 4: SKU in Sale Items

**Type**: Enhancement

**Changes**:
- `GET /api/v1/sales/{id}` - Added `sku` field to each item in `items` array
- SKU is fetched from the product variant through the batch

**Response Change**:
```json
// BEFORE - GET /api/v1/sales/{id}
{
  "items": [
    {
      "id": "SITM00000001",
      "batch_id": "BATC00000001",
      "quantity": 5,
      "selling_price": 100.00
    }
  ]
}

// AFTER - GET /api/v1/sales/{id}
{
  "items": [
    {
      "id": "SITM00000001",
      "batch_id": "BATC00000001",
      "sku": "SKU-VEG-00000001",  // NEW - Product variant SKU
      "quantity": 5,
      "selling_price": 100.00
    }
  ]
}
```

**Usage**:
- Use `sku` to display product identification in sale details
- SKU links to the specific product variant sold

---

## Issue 5: Phone Number Storage (Investigation)

**Type**: Bug Investigation

**Status**: VERIFIED - Working Correctly

**Investigation Results**:
- `POST /api/v1/sales` - `customer_phone` field is properly stored
- Phone number flows correctly: Request → Handler → Service → Model → Database
- No code changes required

**How to Use**:
```json
// POST /api/v1/sales
{
  "warehouse_id": "WHSE00000001",
  "customer_phone": "9876543210",    // Optional - stored correctly
  "customer_name": "John Doe",       // Optional
  "payment_mode": "cash",
  "sale_type": "in_store",
  "items": [...]
}
```

**Response**:
```json
{
  "id": "SALE00000001",
  "customer_phone": "9876543210",    // Returned correctly
  "customer_name": "John Doe"
}
```

---

## Issue 6: apply_taxes Default True (BREAKING CHANGE)

**Type**: Breaking Change

**Changes**:
- `POST /api/v1/sales` - `apply_taxes` default changed from `false` to `true`
- If `apply_taxes` is not provided, taxes will be calculated by default

**Behavior Change**:
```json
// BEFORE - Omitting apply_taxes
{
  "warehouse_id": "WHSE00000001",
  "payment_mode": "cash",
  "sale_type": "in_store",
  "items": [...]
  // apply_taxes NOT provided → defaulted to FALSE (no tax calculation)
}

// AFTER - Omitting apply_taxes
{
  "warehouse_id": "WHSE00000001",
  "payment_mode": "cash",
  "sale_type": "in_store",
  "items": [...]
  // apply_taxes NOT provided → defaults to TRUE (taxes ARE calculated)
}
```

**Migration**:
- Review all sale creation calls
- If taxes should NOT be applied, explicitly set `"apply_taxes": false`
- Existing sales are not affected (only new sales)

**Example - Opting out of taxes**:
```json
{
  "warehouse_id": "WHSE00000001",
  "payment_mode": "cash",
  "sale_type": "in_store",
  "apply_taxes": false,  // Explicitly disable taxes
  "items": [...]
}
```

---

## Issue 7: Phone Filter Query Parameter

**Type**: New Feature

**Changes**:
- `GET /api/v1/sales` - Added `customer_phone` query parameter
- Filter sales by customer phone number

**Usage**:
```
GET /api/v1/sales?customer_phone=9876543210
GET /api/v1/sales?customer_phone=9876543210&limit=10&offset=0
```

**Response**:
```json
{
  "data": [
    {
      "id": "SALE00000001",
      "invoice_number": "12250001",
      "customer_phone": "9876543210",
      "customer_name": "John Doe",
      "total_amount": 500.00,
      "status": "completed"
    }
  ],
  "total": 5,
  "limit": 20,
  "offset": 0
}
```

**Notes**:
- Exact match on phone number (not partial search)
- Works with pagination parameters (limit, offset)
- Returns empty array if no sales found for phone

---

## Issue 8: Availability Variant Images

**Type**: Enhancement

**Changes**:
- `GET /api/v1/products/availability` - Added `images` field to response
- Images are S3 paths from the product variant

**Response Change**:
```json
// BEFORE
{
  "sku": "SKU-VEG-00000001",
  "variant_id": "PVAR00000001",
  "product_name": "Tomato 1kg",
  "hsn_code": "07020000",
  "gst_rate": 5.0
}

// AFTER
{
  "sku": "SKU-VEG-00000001",
  "variant_id": "PVAR00000001",
  "product_name": "Tomato 1kg",
  "hsn_code": "07020000",
  "gst_rate": 5.0,
  "images": [                            // NEW - S3 paths for variant images
    "variants/PVAR00000001/image1.jpg",
    "variants/PVAR00000001/image2.jpg"
  ]
}
```

**Notes**:
- `images` is an array of S3 paths (not presigned URLs)
- Use `/api/v1/attachments/:id/url` to get presigned URLs for display
- Empty array if no images exist

---

## Issue 9: PATCH Sale Endpoint

**Type**: New Feature

**Changes**:
- `PATCH /api/v1/sales/{id}` - New endpoint to partially update a sale
- Only pending sales can be updated
- Updateable fields: `payment_mode`, `sale_type`, `customer_phone`, `customer_name`

**Endpoint**:
```
PATCH /api/v1/sales/{id}
```

**Request Body** (all fields optional):
```json
{
  "payment_mode": "upi",              // Optional - "cash", "upi", or "online"
  "sale_type": "delivery",            // Optional - "in_store" or "delivery"
  "customer_phone": "9876543210",     // Optional - customer phone number
  "customer_name": "Jane Doe"         // Optional - customer name
}
```

**Response** (full sale object):
```json
{
  "id": "SALE00000001",
  "invoice_number": "12250001",
  "warehouse_id": "WHSE00000001",
  "sale_date": "2025-12-11",
  "total_amount": 500.00,
  "status": "pending",
  "customer_phone": "9876543210",
  "customer_name": "Jane Doe",
  "payment_mode": "upi",
  "sale_type": "delivery",
  "apply_taxes": true,
  "items": [...],
  "breakdown": {...},
  "created_at": "2025-12-11T10:00:00Z",
  "updated_at": "2025-12-11T10:30:00Z"
}
```

**Error Responses**:
- `400 Bad Request` - Invalid request data (invalid payment_mode or sale_type values)
- `404 Not Found` - Sale with given ID not found

**Usage Examples**:
```bash
# Update payment mode only
curl -X PATCH /api/v1/sales/SALE00000001 \
  -H "Content-Type: application/json" \
  -d '{"payment_mode": "upi"}'

# Update customer info only
curl -X PATCH /api/v1/sales/SALE00000001 \
  -H "Content-Type: application/json" \
  -d '{"customer_phone": "9876543210", "customer_name": "Jane Doe"}'

# Update multiple fields
curl -X PATCH /api/v1/sales/SALE00000001 \
  -H "Content-Type: application/json" \
  -d '{"payment_mode": "online", "sale_type": "delivery"}'
```

**Notes**:
- Only fields included in the request body will be updated
- Missing fields are left unchanged
- Validation enforces valid enum values for payment_mode and sale_type

---

