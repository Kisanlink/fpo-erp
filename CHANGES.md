# API Changes - Collaborator Management & Unified Variant Architecture

**Date**: November 20, 2025
**Version**: 1.5.0
**Breaking Changes**:
- Removed all `/api/v1/collaborator-products` endpoints (7 endpoints deleted)
- Deprecated `CollaboratorProduct` model (will be removed in v2.0.0)
- Unified architecture: All collaborator products now use `ProductVariant` table

**New Features**:
- Can now update collaborator associations via `PUT /api/v1/variants/{id}`
- Support for add/remove/delete all collaborators on existing variants
- Simplified many-to-many relationship via JSON array

---

## BREAKING CHANGE: Unified Variant Architecture - CollaboratorProduct Deprecated

### Overview

All collaborator-specific products are now stored in the `product_variants` table using the `collaborator_ids` field (JSON array). The separate `CollaboratorProduct` model, service, handler, and repository have been removed. This provides a unified architecture where both regular variants and collaborator-specific variants use the same table and endpoints.

### What Changed

**REMOVED Endpoints** (v1.5.0 - no longer available):
1. ~~`GET /api/v1/collaborator-products/{id}`~~ (removed)
2. ~~`PUT /api/v1/collaborator-products/{id}`~~ (removed)
3. ~~`DELETE /api/v1/collaborator-products/{id}`~~ (removed)
4. ~~`GET /api/v1/collaborators/{id}/products`~~ (removed)
5. ~~`POST /api/v1/collaborators/{id}/products`~~ (removed)
6. ~~`DELETE /api/v1/collaborators/{id}/products/{product_id}`~~ (removed)
7. ~~`GET /api/v1/products/{id}/collaborators`~~ (removed)

**NEW/CORRECT Endpoints** (v1.5.0 - use these instead):
1. **Get variant by ID**: `GET /api/v1/variants/{id}`
2. **Update variant** (including collaborators): `PUT /api/v1/variants/{id}`
3. **Delete variant**: `DELETE /api/v1/variants/{id}`
4. **Get variant by SKU**: `GET /api/v1/variants/sku/{sku}`
5. **Get variant by barcode**: `GET /api/v1/variants/barcode/{barcode}`
6. **Create variant for product**: `POST /api/v1/products/{id}/variants`
7. **Get all variants for product**: `GET /api/v1/products/{id}/variants`

**NEW Capability** (v1.5.0):
- `CollaboratorIDs` field added to `UpdateProductVariantRequest`
- Can now **add, remove, or delete all** collaborators from existing variants via `PUT /api/v1/variants/{id}`
- Full collaborator association management without needing separate endpoints

### Deleted Files

The following files have been removed:
- `internal/api/handlers/collaborator_product_handler.go`
- `internal/services/collaborator_product_service.go`
- `internal/services/interfaces/collaborator_product_service.go`
- `internal/database/repositories/collaborator_product_repo.go`
- `tests/handlers/collaborator_product_handler_test.go`

**Kept for Backward Compatibility**:
- `internal/database/models/collaborator_product.go` (deprecated, not migrated)

### New Workflow

#### Creating Collaborator-Specific Variants

**Endpoint**: `POST /api/v1/products/{product_id}/variants`

```json
{
  "variant_name": "1kg Premium Pack",
  "quantity": "1kg",
  "pack_size": "Premium Pack",
  "collaborator_ids": ["CLAB00000001", "CLAB00000002"],
  "brand_name": "Vendor Brand",
  "hsn_code": "12345678",
  "gst_rate": 18.0,
  "images": ["s3://bucket/image1.jpg"],
  "dosage_instructions": "Take 2 tablets daily",
  "usage_details": "Best used after meals",
  "prices": [
    {"price_type": "MRP", "price": 150.00, "currency": "INR"},
    {"price_type": "MSP", "price": 120.00, "currency": "INR"}
  ]
}
```

#### Adding Collaborators to Existing Variant

**Endpoint**: `PUT /api/v1/variants/{variant_id}`

```json
{
  "collaborator_ids": ["CLAB00000001", "CLAB00000002", "CLAB00000003"]
}
```

**Result**: Variant now associated with 3 collaborators

#### Removing Specific Collaborators

**Endpoint**: `PUT /api/v1/variants/{variant_id}`

```json
{
  "collaborator_ids": ["CLAB00000001"]
}
```

**Result**: Only CLAB00000001 remains, others removed

#### Deleting ALL Collaborators

**Endpoint**: `PUT /api/v1/variants/{variant_id}`

```json
{
  "collaborator_ids": []
}
```

**Result**: Variant no longer associated with any collaborators

#### Getting Collaborator's Products

**Endpoint**: `GET /api/v1/variants?collaborator_id={collaborator_id}`

**Response**: All variants where `collaborator_ids` array contains the specified ID

### Migration Guide for Frontend

**Before (v1.4.x)**:
```javascript
// Creating collaborator product
POST /api/v1/collaborators/CLAB001/products
{
  "product_id": "PROD001",
  "brand_name": "Brand X",
  "hsn_code": "12345678",
  "gst_rate": 18.0
}

// Updating collaborator product
PUT /api/v1/collaborator-products/CPRD001
{
  "brand_name": "New Brand"
}

// Getting collaborator products
GET /api/v1/collaborators/CLAB001/products
```

**After (v1.5.0)**:
```javascript
// Creating variant with collaborator
POST /api/v1/products/PROD001/variants
{
  "variant_name": "1kg Pack",
  "quantity": "1kg",
  "pack_size": "Standard",
  "collaborator_ids": ["CLAB001"],
  "brand_name": "Brand X",
  "hsn_code": "12345678",
  "gst_rate": 18.0,
  "prices": [{"price_type": "MRP", "price": 100, "currency": "INR"}]
}

// Updating variant (including collaborators)
PUT /api/v1/variants/PVAR001
{
  "brand_name": "New Brand",
  "collaborator_ids": ["CLAB001", "CLAB002"]  // Can update collaborators too!
}

// Getting collaborator's variants
GET /api/v1/variants?collaborator_id=CLAB001
```

### Technical Changes

**Model Changes**:
```go
// UpdateProductVariantRequest - NEW field added
type UpdateProductVariantRequest struct {
    // ... existing fields ...
    CollaboratorIDs *[]string `json:"collaborator_ids,omitempty"` // NEW: Can update collaborators
}
```

**Service Changes** (`product_variant_service.go:303-305`):
```go
// NEW: Support for updating collaborator associations
if request.CollaboratorIDs != nil {
    variant.CollaboratorIDs = *request.CollaboratorIDs
}
```

**Route Changes** (`routes.go`):
- Removed: `collaboratorProductRepo` initialization
- Removed: `collaboratorProductService` initialization
- Removed: `collaboratorProductHandler` initialization
- Removed: `collaboratorProductHandler.RegisterRoutes(v1)`

### Deprecation Notice

The `CollaboratorProduct` model in `internal/database/models/collaborator_product.go` is **deprecated** and will be removed in **v2.0.0**. The file has been kept for backward compatibility but is no longer used by the system.

**Migration Path**:
- All data should be migrated to `product_variants` table with `collaborator_ids` field
- Update all API calls to use `/api/v1/variants` and `/api/v1/products/:id/variants` endpoints
- CollaboratorProduct table will not be auto-migrated (already removed from migrator)

### Benefits of Unified Architecture

1. **Simplified Queries**: Single table for all variants (regular + collaborator)
2. **Better Performance**: No JOIN needed between tables
3. **Flexible Associations**: Many-to-many via JSON array
4. **Easier Maintenance**: One codebase for all variant operations
5. **Scalable**: Can easily support multiple collaborators per variant

---

# API Changes - Price Type Update (MRP/MSP)

**Date**: November 20, 2025
**Version**: 1.4.1
**Breaking Changes**:
- Changed price types from retail/wholesale/bulk to MRP/MSP
- Old price types will be rejected with validation error

---

## BREAKING CHANGE: Price Types Changed to MRP/MSP

### Overview

Price types have been changed from generic retail/wholesale/bulk to Indian standard MRP (Maximum Retail Price) and MSP (Minimum Selling Price). This aligns with Indian retail practices where MRP is printed on products and MSP is the minimum price at which a retailer can sell.

### What Changed

**OLD Price Types** (v1.4.0 - no longer accepted):
- `"retail"` - Retail/consumer price
- `"wholesale"` - Wholesale price for bulk buyers
- `"bulk"` - Bulk/distributor price

**NEW Price Types** (v1.4.1 - only accepted):
- `"MRP"` - Maximum Retail Price (printed price on product)
- `"MSP"` - Minimum Selling Price (floor price for retailers)

### String Constants

Price types are now defined as string constants in the models package:

```go
// internal/database/models/product_variant.go
const (
    PriceTypeMRP = "MRP"  // Maximum Retail Price
    PriceTypeMSP = "MSP"  // Minimum Selling Price
)
```

### Validation Changes

**Price validation now ONLY accepts "MRP" or "MSP":**

```go
// Before (v1.4.0)
validPriceTypes := []string{"retail", "wholesale", "bulk"}

// After (v1.4.1)
validPriceTypes := []string{"MRP", "MSP"}
```

**Error Messages:**

**OLD** (v1.4.0):
```json
{
  "status": "error",
  "message": "Invalid price_type at index 0: must be 'retail', 'wholesale', or 'bulk'"
}
```

**NEW** (v1.4.1):
```json
{
  "status": "error",
  "message": "Invalid price_type at index 0: must be 'MRP' or 'MSP'"
}
```

---

### API Examples

#### Create Variant with MRP/MSP

**Endpoint**: `POST /api/v1/products/:id/variants`

**Request Body**:
```json
{
  "variant_name": "1kg Premium Pack",
  "quantity": "1kg",
  "pack_size": "Premium Pack",
  "prices": [
    {
      "price_type": "MRP",
      "price": 150.00,
      "currency": "INR"
    },
    {
      "price_type": "MSP",
      "price": 130.00,
      "currency": "INR"
    }
  ],
  "sku": "PROD-1KG"
}
```

**Response**:
```json
{
  "status": "success",
  "message": "Product variant created successfully",
  "data": {
    "id": "PVAR_abc12345",
    "variant_name": "1kg Premium Pack",
    "prices": [
      {
        "price_type": "MRP",
        "price": 150.00,
        "currency": "INR"
      },
      {
        "price_type": "MSP",
        "price": 130.00,
        "currency": "INR"
      }
    ]
  }
}
```

#### Update Variant Prices

**Endpoint**: `PUT /api/v1/variants/:id`

**Request Body**:
```json
{
  "prices": [
    {
      "price_type": "MRP",
      "price": 175.00,
      "currency": "INR"
    },
    {
      "price_type": "MSP",
      "price": 150.00,
      "currency": "INR"
    }
  ]
}
```

---

### Sales Behavior Change

**Sales service now looks for MRP price first** (instead of retail):

**Before** (v1.4.0):
```go
// Looked for "retail" price
if price.PriceType == "retail" {
    return price.Price, nil
}
```

**After** (v1.4.1):
```go
// Looks for "MRP" price
if price.PriceType == models.PriceTypeMRP {
    return price.Price, nil
}
```

**Fallback Behavior**: If MRP is not found, uses first available price (MSP).

---

### Frontend Migration Guide

#### Step 1: Update Price Type Values

**Before** (v1.4.0):
```javascript
const prices = [
  { price_type: "retail", price: 150, currency: "INR" },
  { price_type: "wholesale", price: 130, currency: "INR" }
];
```

**After** (v1.4.1):
```javascript
const prices = [
  { price_type: "MRP", price: 150, currency: "INR" },
  { price_type: "MSP", price: 130, currency: "INR" }
];
```

#### Step 2: Update Form Labels

```jsx
<div className="price-inputs">
  <div className="form-group">
    <label>MRP (Maximum Retail Price) ₹</label>
    <input
      type="number"
      value={prices.find(p => p.price_type === 'MRP')?.price || ''}
      onChange={(e) => handlePriceChange('MRP', e.target.value)}
    />
    <small className="text-muted">Price printed on product packaging</small>
  </div>

  <div className="form-group">
    <label>MSP (Minimum Selling Price) ₹</label>
    <input
      type="number"
      value={prices.find(p => p.price_type === 'MSP')?.price || ''}
      onChange={(e) => handlePriceChange('MSP', e.target.value)}
    />
    <small className="text-muted">Minimum price for retailers</small>
  </div>
</div>
```

#### Step 3: Update Display Components

```jsx
const PriceDisplay = ({ prices }) => {
  const mrp = prices.find(p => p.price_type === 'MRP');
  const msp = prices.find(p => p.price_type === 'MSP');

  return (
    <div className="price-info">
      {mrp && (
        <div className="mrp">
          <strong>MRP:</strong> ₹{mrp.price.toLocaleString('en-IN')}
        </div>
      )}
      {msp && (
        <div className="msp">
          <strong>MSP:</strong> ₹{msp.price.toLocaleString('en-IN')}
        </div>
      )}
      {mrp && msp && (
        <div className="margin text-muted">
          Margin: ₹{(mrp.price - msp.price).toFixed(2)}
          ({(((mrp.price - msp.price) / mrp.price) * 100).toFixed(1)}%)
        </div>
      )}
    </div>
  );
};
```

---

### Business Logic

#### MRP (Maximum Retail Price)
- **Definition**: The maximum price at which a product can be sold to consumers in India
- **Legal Requirement**: Must be printed on product packaging (Legal Metrology Act, 2009)
- **Use Case**: This is the price customers see on the product
- **Example**: ₹150 printed on a 1kg rice packet

#### MSP (Minimum Selling Price)
- **Definition**: The minimum price at which retailers/distributors can sell the product
- **Use Case**: Protects profit margins, prevents undercutting
- **Example**: Retailer can sell between ₹130 (MSP) and ₹150 (MRP)

#### Margin Calculation
- **Margin** = MRP - MSP
- **Margin %** = ((MRP - MSP) / MRP) × 100

**Example**:
- MRP = ₹150
- MSP = ₹130
- Margin = ₹20 (13.3%)

---

### Breaking Changes Summary

**⚠️ CRITICAL**: This is a **breaking change** requiring immediate updates:

1. **Price Type Values Changed**: `"retail"/"wholesale"/"bulk"` → `"MRP"/"MSP"`
2. **Only 2 Price Types Supported**: Reduced from 3 to 2 price points
3. **Validation Will Reject Old Types**: Sending `"retail"` will return 400 error
4. **Sales Logic Changed**: Uses MRP instead of retail price
5. **No Backward Compatibility**: Old price types not supported

---

### Migration Checklist

#### Frontend (Required)
- [ ] Update all price type values from retail/wholesale/bulk to MRP/MSP
- [ ] Change form labels to "MRP" and "MSP"
- [ ] Update validation to only allow MRP/MSP
- [ ] Update display components to show MRP/MSP labels
- [ ] Update price comparison logic (if any)
- [ ] Test variant creation with new price types
- [ ] Test price validation errors

#### Backend (Already Complete)
- [x] Added PriceTypeMRP and PriceTypeMSP constants
- [x] Updated validation to only accept MRP/MSP
- [x] Updated sales service to use MRP
- [x] Updated error messages
- [x] Code compiles successfully

---

### Files Modified

**Models**:
- `internal/database/models/product_variant.go` - Added price type constants and updated comments

**Services**:
- `internal/services/product_variant_service.go` - Updated validation to accept only MRP/MSP
- `internal/services/sales_service.go` - Updated getSellingPrice() to use MRP

**Total Changes**: 3 files modified, ~15 lines changed

---

# API Changes - Product Pricing Consolidation

**Date**: November 20, 2025
**Version**: 1.4.0
**Breaking Changes**:
- Removed separate pricing API endpoints (8 endpoints)
- Pricing now embedded in Product Variant API
- Simplified pricing model (no price history, current prices only)

---

## BREAKING CHANGE: Pricing API Removed

### Overview

The separate pricing API has been **completely removed** and pricing functionality has been consolidated into the Product Variant API. Prices are now stored directly within variants as a JSON array, eliminating the need for a separate price management system.

### What Changed

**REMOVED: 8 Pricing Endpoints**
1. `POST /api/v1/prices` - Create price
2. `GET /api/v1/prices/:id` - Get price by ID
3. `GET /api/v1/prices` - Get all prices
4. `GET /api/v1/prices/variant/:id` - Get prices by variant
5. `GET /api/v1/prices/variant/:id/current` - Get current price
6. `GET /api/v1/prices/variant/:id/active` - Get active prices
7. `PATCH /api/v1/prices/:id` - Update price
8. `DELETE /api/v1/prices/:id` - Delete price

**NEW: Pricing in Variant API**

Prices are now managed through variant endpoints:
- `POST /api/v1/products/:id/variants` - Create variant with prices
- `GET /api/v1/variants/:id` - Get variant with prices
- `PUT /api/v1/variants/:id` - Update variant prices

---

### Database Schema Changes

**Table Removed**: `product_prices` table is **no longer auto-migrated**

**Table Modified**: `product_variants`

| Field | Type | Description |
|-------|------|-------------|
| `prices` | `JSON` | Array of price objects with price_type, price, and currency |

**Price Structure** (stored as JSON):
```json
{
  "prices": [
    {
      "price_type": "retail",
      "price": 100.50,
      "currency": "INR"
    },
    {
      "price_type": "wholesale",
      "price": 85.00,
      "currency": "INR"
    },
    {
      "price_type": "bulk",
      "price": 75.00,
      "currency": "INR"
    }
  ]
}
```

---

### API Changes

#### 1. Create Product Variant (with prices)

**Endpoint**: `POST /api/v1/products/:id/variants`

**Request Body**:
```json
{
  "variant_name": "1kg Premium Pack",
  "quantity": "1kg",
  "pack_size": "Premium Pack",
  "prices": [
    {
      "price_type": "retail",
      "price": 150.00,
      "currency": "INR"
    },
    {
      "price_type": "wholesale",
      "price": 130.00,
      "currency": "INR"
    }
  ],
  "sku": "PROD-1KG",
  "barcode": "1234567890"
}
```

**Note**: Prices are **optional** at creation. You can create a variant without prices.

#### 2. Get Product Variant (includes prices)

**Endpoint**: `GET /api/v1/variants/:id`

**Response**:
```json
{
  "status": "success",
  "message": "Product variant retrieved successfully",
  "data": {
    "id": "PVAR_abc12345",
    "product_id": "PROD_product1",
    "variant_name": "1kg Premium Pack",
    "quantity": "1kg",
    "pack_size": "Premium Pack",
    "prices": [
      {
        "price_type": "retail",
        "price": 150.00,
        "currency": "INR"
      },
      {
        "price_type": "wholesale",
        "price": 130.00,
        "currency": "INR"
      }
    ],
    "sku": "PROD-1KG",
    "is_active": true,
    "created_at": "2025-11-20T10:00:00Z",
    "updated_at": "2025-11-20T10:00:00Z"
  }
}
```

**Before** (v1.3.0):
- Variant response did NOT include prices
- Had to make separate API call to `/api/v1/prices/variant/:id` to get prices

**After** (v1.4.0):
- Variant response ALWAYS includes `prices` array
- No additional API calls needed

#### 3. Update Product Variant (update prices)

**Endpoint**: `PUT /api/v1/variants/:id`

**Request Body** (partial update):
```json
{
  "prices": [
    {
      "price_type": "retail",
      "price": 175.00,
      "currency": "INR"
    }
  ]
}
```

**Important**:
- Updating prices **replaces** the entire prices array (no price history)
- To keep existing prices, send the complete array with modifications
- Omit `prices` field to leave prices unchanged

---

### Pricing Rules & Validation

#### Valid Price Types
- `retail` - Retail/consumer price
- `wholesale` - Wholesale price for bulk buyers
- `bulk` - Bulk/distributor price

**Note**: Other price types will be rejected with validation error.

#### Validation Rules
1. **No Duplicates**: Cannot have multiple prices with the same `price_type`
2. **Positive Prices**: Price must be > 0
3. **Currency Required**: Currency field cannot be empty
4. **Optional**: Can create/update variant without prices (empty array)

#### Example Validation Errors

**Duplicate Price Type**:
```json
{
  "status": "error",
  "message": "Duplicate price_type 'retail' found"
}
```

**Invalid Price Type**:
```json
{
  "status": "error",
  "message": "Invalid price_type at index 0: must be 'retail', 'wholesale', or 'bulk'"
}
```

**Negative Price**:
```json
{
  "status": "error",
  "message": "Price must be greater than 0 for retail"
}
```

---

### Migration Path for Frontend

#### Step 1: Remove Old Pricing API Calls

**Before** (v1.3.0):
```javascript
// Create variant (WITHOUT prices)
const variant = await createVariant({
  variant_name: "1kg Pack",
  quantity: "1kg",
  pack_size: "Regular"
});

// THEN create price separately
await createPrice({
  variant_id: variant.id,
  price_type: "retail",
  price: 150.00,
  currency: "INR",
  effective_from: "2025-11-20"
});
```

**After** (v1.4.0):
```javascript
// Create variant WITH prices (single API call)
const variant = await createVariant({
  variant_name: "1kg Pack",
  quantity: "1kg",
  pack_size: "Regular",
  prices: [
    {
      price_type: "retail",
      price: 150.00,
      currency: "INR"
    }
  ]
});
```

#### Step 2: Update Variant Display Components

**Before** (v1.3.0):
```jsx
// Fetch variant
const variant = await fetchVariant(variantId);

// Fetch prices separately
const prices = await fetchPricesByVariant(variantId);

// Display
<div>
  <h3>{variant.variant_name}</h3>
  <p>Retail Price: ₹{prices.find(p => p.price_type === 'retail')?.price}</p>
</div>
```

**After** (v1.4.0):
```jsx
// Fetch variant (prices included)
const variant = await fetchVariant(variantId);

// Display
<div>
  <h3>{variant.variant_name}</h3>
  <p>Retail Price: ₹{variant.prices.find(p => p.price_type === 'retail')?.price || 'N/A'}</p>
</div>
```

#### Step 3: Update Price Management UI

**Price Input Form**:
```jsx
const PriceInputs = ({ prices, setPrices }) => {
  const handlePriceChange = (priceType, value) => {
    const updated = prices.filter(p => p.price_type !== priceType);
    if (value > 0) {
      updated.push({
        price_type: priceType,
        price: parseFloat(value),
        currency: "INR"
      });
    }
    setPrices(updated);
  };

  return (
    <div className="price-inputs">
      <div className="form-group">
        <label>Retail Price (₹)</label>
        <input
          type="number"
          value={prices.find(p => p.price_type === 'retail')?.price || ''}
          onChange={(e) => handlePriceChange('retail', e.target.value)}
          step="0.01"
        />
      </div>

      <div className="form-group">
        <label>Wholesale Price (₹)</label>
        <input
          type="number"
          value={prices.find(p => p.price_type === 'wholesale')?.price || ''}
          onChange={(e) => handlePriceChange('wholesale', e.target.value)}
          step="0.01"
        />
      </div>

      <div className="form-group">
        <label>Bulk Price (₹)</label>
        <input
          type="number"
          value={prices.find(p => p.price_type === 'bulk')?.price || ''}
          onChange={(e) => handlePriceChange('bulk', e.target.value)}
          step="0.01"
        />
      </div>
    </div>
  );
};
```

#### Step 4: Update Price Display Components

**Price Table Component**:
```jsx
const PriceTable = ({ prices }) => {
  if (!prices || prices.length === 0) {
    return <p className="text-muted">No pricing information available</p>;
  }

  const priceTypeLabels = {
    retail: 'Retail Price',
    wholesale: 'Wholesale Price',
    bulk: 'Bulk Price'
  };

  return (
    <table className="table">
      <thead>
        <tr>
          <th>Price Type</th>
          <th>Amount</th>
          <th>Currency</th>
        </tr>
      </thead>
      <tbody>
        {prices.map(price => (
          <tr key={price.price_type}>
            <td>{priceTypeLabels[price.price_type] || price.price_type}</td>
            <td>₹{price.price.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
            <td>{price.currency}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
};
```

---

### Simplified Pricing Model

#### What Was Removed
- ❌ Price history tracking
- ❌ Date-based price ranges (effective_from, effective_to)
- ❌ Separate price activation/deactivation
- ❌ Price update audit trail
- ❌ Time-based pricing queries

#### What Remains
- ✅ Multiple price points per variant (retail, wholesale, bulk)
- ✅ Currency support
- ✅ Price validation
- ✅ Embedded pricing in variant responses

#### Rationale
The separate pricing system added unnecessary complexity for this use case:
- Most users don't need price history
- Price changes can be tracked via variant update history
- Simpler data model = faster queries and easier maintenance
- Reduces API calls from 2 (variant + prices) to 1 (variant with prices)

---

### Breaking Changes Summary

**⚠️ CRITICAL**: This is a **major breaking change** requiring frontend updates:

1. **Endpoints Removed**: All 8 pricing endpoints no longer exist (404 error)
2. **Data Structure Changed**: Prices moved from separate table to JSON field in variants
3. **No Price History**: Old price data will NOT be migrated (only current/latest prices should be preserved)
4. **Response Format Changed**: Variant responses now include `prices` array
5. **Request Format Changed**: Must include `prices` array in variant create/update requests

---

### Migration Checklist

#### Backend (Already Complete)
- [x] Removed ProductPrice model from auto-migration
- [x] Updated ProductVariant model with `prices` JSON field
- [x] Updated ProductVariantService with price validation
- [x] Updated SalesService to use variant prices (not price repository)
- [x] Removed price routes from routes.go
- [x] Removed ProductService dependency on priceRepo
- [x] Code compiles without errors

#### Frontend (Required)
- [ ] Remove all calls to `/api/v1/prices/*` endpoints
- [ ] Update variant creation forms to include prices array
- [ ] Update variant display components to show embedded prices
- [ ] Update variant edit forms to manage prices array
- [ ] Remove price management pages/components
- [ ] Update sales/order flows to read prices from variants
- [ ] Test price validation (duplicate types, negative prices, etc.)

---

### Files Modified

**Models**:
- `internal/database/models/product_variant.go` - Added `Prices []VariantPrice` field

**Services**:
- `internal/services/product_variant_service.go` - Added `validatePrices()` and `GetPriceByType()` methods
- `internal/services/sales_service.go` - Changed from `priceRepo` to `variantRepo`, updated `getSellingPrice()` method
- `internal/services/product_service.go` - Removed `priceRepo` dependency

**Routes**:
- `internal/api/routes/routes.go` - Removed price-related initialization and route registration

**Database**:
- `internal/database/migrator.go` - Removed `&models.ProductPrice{}` from auto-migration

**Total Changes**: 5 files modified, ProductPrice table no longer managed

---

# API Changes - Product Variant Multi-Collaborator Support

**Date**: November 20, 2025
**Version**: 1.3.0
**Features**:
- Multiple Collaborators Per Product Variant
- Purchase Order "Verified" Status (Quality Inspection Stage)
- Rejected Goods Return Workflow

---

## NEW: Multiple Collaborators Per Product Variant

### Overview

Product variants can now be associated with **multiple collaborators (vendors/suppliers)**. Previously, each variant could only be linked to a single collaborator. This update allows the same variant to be supplied by multiple vendors, enabling better price comparison and supply chain flexibility.

### Key Changes

**Before**: One variant → One collaborator (1:1 relationship)
**After**: One variant → Multiple collaborators (1:N relationship)

### Database Schema Changes

**Table**: `product_variants`

| Field | Old Type | New Type | Description |
|-------|----------|----------|-------------|
| `collaborator_id` | `VARCHAR(100)` (nullable) | **REMOVED** | Single collaborator ID |
| `collaborator_ids` | N/A | `JSON` (array of strings) | **NEW**: Multiple collaborator IDs stored as JSON array |

**Migration**: GORM AutoMigrate will automatically handle the schema change when the server restarts. Existing data will be migrated from `collaborator_id` to `collaborator_ids` array.

---

### API Impact

#### 1. Product Variant Response

**Endpoint**: All variant endpoints (GET /api/v1/variants/*, etc.)

**Changed Field**:
```json
{
  "id": "PVAR_abc12345",
  "product_id": "PROD_product1",
  "variant_name": "1kg Premium Pack",
  "collaborator_ids": ["CLAB_vendor1", "CLAB_vendor2", "CLAB_vendor3"],
  "brand_name": "Premium Brand",
  "sku": "RICE-PREM-1KG",
  "is_active": true
}
```

**Before**: `"collaborator_id": "CLAB_vendor1"` (single string, nullable)
**After**: `"collaborator_ids": ["CLAB_vendor1", "CLAB_vendor2"]` (array of strings)

#### 2. Create Product Variant Request

**Endpoint**: `POST /api/v1/products/:id/variants`

**Changed Field**:
```json
{
  "variant_name": "1kg Premium Pack",
  "quantity": "1kg",
  "pack_size": "Premium Pack",
  "collaborator_ids": ["CLAB_vendor1", "CLAB_vendor2"],
  "brand_name": "Premium Brand",
  "hsn_code": "10063020",
  "gst_rate": 5.0,
  "images": ["s3://path/to/image1.jpg", "s3://path/to/image2.jpg"]
}
```

**Before**: `"collaborator_id": "CLAB_vendor1"` (single string, optional)
**After**: `"collaborator_ids": ["CLAB_vendor1", "CLAB_vendor2"]` (array of strings, optional)

#### 3. Add Product to Collaborator

**Endpoint**: `POST /api/v1/collaborators/:id/products`

**Behavior Change**:
- **Before**: Creating a variant for a collaborator would fail if the product was already associated with that collaborator
- **After**: System checks if the collaborator ID already exists in the variant's `collaborator_ids` array
  - If exists: Returns `409 Conflict` error
  - If not exists: Adds collaborator ID to the existing variant's array

**Example Flow**:
```javascript
// Step 1: Vendor A associates Product X
POST /api/v1/collaborators/CLAB_vendorA/products
{ "product_id": "PROD_X", "brand_name": "Brand A", ... }
// Creates variant with collaborator_ids: ["CLAB_vendorA"]

// Step 2: Vendor B associates the same Product X
POST /api/v1/collaborators/CLAB_vendorB/products
{ "product_id": "PROD_X", "brand_name": "Brand B", ... }
// Updates variant with collaborator_ids: ["CLAB_vendorA", "CLAB_vendorB"]

// Step 3: Vendor A tries to associate Product X again
POST /api/v1/collaborators/CLAB_vendorA/products
{ "product_id": "PROD_X", ... }
// Returns 409 Conflict: "product already associated with this collaborator"
```

#### 4. Get Products by Collaborator

**Endpoint**: `GET /api/v1/collaborators/:id/products`

**Query Change**: Repository now uses PostgreSQL JSON contains operator (`@>`) to search within the `collaborator_ids` array.

**SQL Query**:
```sql
-- Before
WHERE collaborator_id = 'CLAB_vendor1' AND is_active = true

-- After (PostgreSQL)
WHERE collaborator_ids @> '["CLAB_vendor1"]' AND is_active = true
```

---

### Frontend Integration Guide

#### 1. Display Multiple Collaborators

**Variant List View**:
```jsx
const VariantCollaborators = ({ collaboratorIds }) => {
  if (!collaboratorIds || collaboratorIds.length === 0) {
    return <span className="text-muted">No collaborators</span>;
  }

  return (
    <div className="collaborator-badges">
      {collaboratorIds.map(collabId => (
        <span key={collabId} className="badge badge-primary me-1">
          {collabId}
        </span>
      ))}
      <span className="text-muted">
        ({collaboratorIds.length} vendor{collaboratorIds.length > 1 ? 's' : ''})
      </span>
    </div>
  );
};
```

#### 2. Create Variant Form

**Update form to accept multiple collaborators**:
```jsx
const [collaboratorIds, setCollaboratorIds] = useState([]);

const handleAddCollaborator = (collaboratorId) => {
  if (!collaboratorIds.includes(collaboratorId)) {
    setCollaboratorIds([...collaboratorIds, collaboratorId]);
  }
};

const handleRemoveCollaborator = (collaboratorId) => {
  setCollaboratorIds(collaboratorIds.filter(id => id !== collaboratorId));
};

// In form submission
const createVariant = async () => {
  const payload = {
    variant_name: variantName,
    quantity: quantity,
    pack_size: packSize,
    collaborator_ids: collaboratorIds, // Send as array
    brand_name: brandName,
    hsn_code: hsnCode,
    gst_rate: gstRate
  };

  const response = await fetch(`/api/v1/products/${productId}/variants`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
};
```

#### 3. Collaborator Selection UI

**Multi-select dropdown example**:
```jsx
<div className="form-group">
  <label>Suppliers/Vendors</label>
  <Select
    isMulti
    options={collaboratorOptions}
    value={selectedCollaborators}
    onChange={setSelectedCollaborators}
    placeholder="Select vendors that supply this variant..."
  />

  <div className="selected-vendors mt-2">
    {selectedCollaborators.map(collab => (
      <span key={collab.value} className="badge badge-info me-1">
        {collab.label}
        <button
          className="btn-close btn-sm ms-1"
          onClick={() => handleRemoveCollaborator(collab.value)}
        />
      </span>
    ))}
  </div>
</div>
```

#### 4. Variant Detail Page

**Show all suppliers for a variant**:
```jsx
const VariantDetailPage = ({ variant }) => {
  return (
    <div className="variant-details">
      <h3>{variant.variant_name}</h3>

      <div className="section">
        <h5>Suppliers ({variant.collaborator_ids.length})</h5>
        <table className="table">
          <thead>
            <tr>
              <th>Vendor ID</th>
              <th>Vendor Name</th>
              <th>Price</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {variant.collaborator_ids.map(collabId => (
              <tr key={collabId}>
                <td>{collabId}</td>
                <td>{getCollaboratorName(collabId)}</td>
                <td>{getLatestPrice(variant.id, collabId)}</td>
                <td>
                  <button onClick={() => createPO(collabId, variant.id)}>
                    Create PO
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};
```

---

### Use Cases

#### 1. Price Comparison

**Scenario**: Compare prices from multiple vendors for the same variant

```javascript
// Get variant with all suppliers
const variant = await fetchVariant("PVAR_abc123");

// Fetch latest prices from each supplier
const priceComparison = await Promise.all(
  variant.collaborator_ids.map(async (collabId) => {
    const price = await fetchLatestPrice(variant.id, collabId);
    const collaborator = await fetchCollaborator(collabId);

    return {
      vendor: collaborator.company_name,
      vendor_id: collabId,
      price: price.unit_price,
      last_updated: price.updated_at
    };
  })
);

// Sort by price (lowest first)
priceComparison.sort((a, b) => a.price - b.price);

console.log("Best price:", priceComparison[0]);
```

#### 2. Supply Chain Redundancy

**Scenario**: Maintain backup suppliers for critical products

```javascript
const ensureSupplyChainRedundancy = (variant) => {
  if (variant.collaborator_ids.length < 2) {
    console.warn(`⚠️ Variant ${variant.variant_name} has only ${variant.collaborator_ids.length} supplier(s). Consider adding backup vendors.`);
  } else {
    console.log(`✅ Variant ${variant.variant_name} has ${variant.collaborator_ids.length} suppliers (redundancy maintained)`);
  }
};
```

#### 3. Vendor Performance Tracking

**Scenario**: Track which vendors supply which products

```javascript
const getVendorPerformance = async (vendorId) => {
  // Get all variants supplied by this vendor
  const variants = await fetchVariantsByCollaborator(vendorId);

  return {
    vendor_id: vendorId,
    total_products: variants.length,
    product_ids: variants.map(v => v.product_id),
    variants: variants.map(v => ({
      id: v.id,
      name: v.variant_name,
      sku: v.sku
    }))
  };
};
```

---

### Breaking Changes

**⚠️ IMPORTANT**: This is a **breaking change** for existing integrations:

1. **Response Field Changed**: `collaborator_id` (string) → `collaborator_ids` (array)
2. **Request Field Changed**: `collaborator_id` (string) → `collaborator_ids` (array)
3. **Query Logic Changed**: Repository uses JSON array search instead of simple equality

**Migration Path for Frontend**:

**Step 1**: Update all variant displays to handle arrays:
```javascript
// Before
<span>{variant.collaborator_id}</span>

// After
<span>
  {variant.collaborator_ids.length > 0
    ? variant.collaborator_ids.join(', ')
    : 'No collaborators'}
</span>
```

**Step 2**: Update variant creation forms to send arrays:
```javascript
// Before
{ collaborator_id: selectedVendor }

// After
{ collaborator_ids: [selectedVendor] }  // Wrap single ID in array
```

**Step 3**: Update filtering/search logic:
```javascript
// Before
variants.filter(v => v.collaborator_id === vendorId)

// After
variants.filter(v => v.collaborator_ids.includes(vendorId))
```

---

### Database Migration Notes

**Automatic Migration**: When the server restarts, GORM will automatically:
1. Add new `collaborator_ids` column (JSON type)
2. Migrate existing `collaborator_id` values to `collaborator_ids` array
3. Keep old `collaborator_id` column for reference (will be removed in future version)

**Manual Migration** (if needed):
```sql
-- PostgreSQL: Migrate existing data
UPDATE product_variants
SET collaborator_ids = CASE
  WHEN collaborator_id IS NOT NULL THEN json_build_array(collaborator_id)::jsonb
  ELSE '[]'::jsonb
END
WHERE collaborator_ids IS NULL;
```

---

### Testing Checklist

#### Backend Verification
- [x] Code compiles without errors
- [ ] GORM AutoMigrate adds `collaborator_ids` column on server restart
- [ ] Can create variant with multiple collaborator IDs
- [ ] Can query variants by collaborator ID (JSON array search)
- [ ] Duplicate check works (prevents same collaborator from being added twice)
- [ ] Webhook integration updated (e-commerce order processing)

#### Frontend Integration
- [ ] Variant list shows multiple collaborators per variant
- [ ] Create variant form accepts array of collaborator IDs
- [ ] Edit variant form allows adding/removing collaborators
- [ ] Collaborator detail page shows all variants they supply
- [ ] Purchase order creation works with new structure
- [ ] Price comparison feature displays all vendor prices

---

### Files Modified

**Models**:
- `internal/database/models/product_variant.go` - Changed `CollaboratorID` to `CollaboratorIDs` (JSON array)

**Services**:
- `internal/services/collaborator_product_service.go` - Updated duplicate check and variant creation logic
- `internal/services/ecommerce_webhook_service.go` - Updated webhook variant creation

**Repositories**:
- `internal/database/repositories/product_variant_repo.go` - Updated `GetByCollaboratorID()` to use JSON array search

**Total Changes**: 5 files modified, ~20 lines changed

---

## Previous Changes (Version 1.2.0)

---

## NEW: Purchase Order Workflow Update - "Verified" Status

### Overview

A new **"verified"** status has been added to the Purchase Order workflow to create a quality inspection stage between delivery and payment. This allows the team to test/inspect products after delivery before creating a GRN and proceeding with payment.

### Updated Status Workflow

**Previous Workflow**:
```
placed → confirmed → out_for_delivery → delivered → paid
```

**New Workflow**:
```
placed → confirmed → out_for_delivery → delivered → verified → paid
```

### Key Changes

1. **New Status**: `"verified"` added between `"delivered"` and `"paid"`
2. **GRN Creation**: GRNs can now ONLY be created when PO status = `"verified"` (previously required `"delivered"`)
3. **Quality Inspection Stage**: After goods are delivered, change status to `"verified"` once inspection/testing is complete
4. **Manual GRN Only**: Auto-GRN creation now triggers at `"verified"` status (not `"delivered"`)

### API Impact

#### 1. Purchase Order Status Values

All PO status fields now accept/return the new `"verified"` status:

**Valid Status Values**:
- `"placed"` - Order created
- `"confirmed"` - Vendor confirmed order
- `"out_for_delivery"` - Shipment dispatched
- `"delivered"` - Goods arrived at warehouse
- **`"verified"` (NEW)** - Quality inspection completed
- `"paid"` - Payment completed

#### 2. Status Transitions

**Valid Transitions**:
| From | To | Purpose |
|------|-----|---------|
| `placed` | `confirmed` | Vendor confirms order |
| `confirmed` | `out_for_delivery` | Vendor ships goods |
| `out_for_delivery` | `delivered` | Goods arrive at warehouse |
| **`delivered`** | **`verified`** | **Quality inspection completed (NEW)** |
| **`verified`** | **`paid`** | **Payment processed (NEW)** |

**Invalid Transitions** (will return `400 Bad Request`):
- `delivered` → `paid` (must go through `verified` first)
- Any status → `delivered` → `paid` (skipping `verified`)

#### 3. GRN Creation Endpoint

**Endpoint**: `POST /api/v1/grns`

**Change**: Purchase Order must be in `"verified"` status (previously `"delivered"`)

**Error Response** (if PO not verified):
```json
{
  "status": "error",
  "message": "Purchase order must be in 'verified' status to create GRN"
}
```

#### 4. Auto-GRN Creation

**Endpoint**: `PATCH /api/v1/purchase-orders/:id/status`

**Change**: Auto-GRN creation now triggers when changing status to `"verified"` (with delivery details provided)

**Example Request**:
```json
{
  "status": "verified",
  "accept_all": true,
  "default_expiry_date": "2025-12-31"
}
```

### Frontend Integration Guide

#### 1. Update PO Status Dropdown

**Before**:
```jsx
<select name="status">
  <option value="placed">Placed</option>
  <option value="confirmed">Confirmed</option>
  <option value="out_for_delivery">Out for Delivery</option>
  <option value="delivered">Delivered</option>
  <option value="paid">Paid</option>
</select>
```

**After**:
```jsx
<select name="status">
  <option value="placed">Placed</option>
  <option value="confirmed">Confirmed</option>
  <option value="out_for_delivery">Out for Delivery</option>
  <option value="delivered">Delivered</option>
  <option value="verified">Verified (Quality Checked)</option>
  <option value="paid">Paid</option>
</select>
```

#### 2. Status Flow UI

**Add visual workflow indicator**:
```
┌────────┐   ┌───────────┐   ┌──────────────────┐   ┌───────────┐   ┌──────────┐   ┌──────┐
│ Placed │ → │ Confirmed │ → │ Out for Delivery │ → │ Delivered │ → │ Verified │ → │ Paid │
└────────┘   └───────────┘   └──────────────────┘   └───────────┘   └──────────┘   └──────┘
                                                                           ↓
                                                                      [Create GRN]
```

#### 3. GRN Creation Button Logic

**Before**:
```jsx
<button
  disabled={po.status !== 'delivered'}
  onClick={createGRN}
>
  Create GRN
</button>
```

**After**:
```jsx
<button
  disabled={po.status !== 'verified'}
  onClick={createGRN}
  title={po.status !== 'verified' ? 'PO must be verified first' : 'Create GRN'}
>
  Create GRN
</button>
```

#### 4. Status Transition Validation

```javascript
const isValidTransition = (currentStatus, newStatus) => {
  const transitions = {
    'placed': ['confirmed'],
    'confirmed': ['out_for_delivery'],
    'out_for_delivery': ['delivered'],
    'delivered': ['verified'],    // NEW
    'verified': ['paid'],         // NEW
  };

  return transitions[currentStatus]?.includes(newStatus) || false;
};
```

### Workflow Example

**Scenario**: Receiving a purchase order

1. **Goods Arrive**: Change PO status to `"delivered"`
   ```
   PATCH /api/v1/purchase-orders/{id}/status
   { "status": "delivered" }
   ```

2. **Quality Inspection**: Warehouse team tests/inspects products
   - Check for damages
   - Verify quantities
   - Test product quality

3. **Mark as Verified**: After inspection passes, change status to `"verified"`
   ```
   PATCH /api/v1/purchase-orders/{id}/status
   {
     "status": "verified",
     "accept_all": true,
     "default_expiry_date": "2025-12-31"
   }
   ```
   - This auto-creates GRN with accepted quantities
   - Inventory batches created automatically

4. **Process Payment**: After GRN is created, mark as `"paid"`
   ```
   PATCH /api/v1/purchase-orders/{id}/status
   { "status": "paid" }
   ```

### Migration Notes

**Existing POs**: No data migration required. Existing purchase orders in `"delivered"` status can:
- Option A: Move directly to `"verified"` status (recommended)
- Option B: Update workflow code to allow one-time `delivered` → `paid` transition for legacy orders

**Database**: No schema changes required. `status` field already supports VARCHAR(30) which accommodates `"verified"`.

### Breaking Changes

**❗ IMPORTANT**: This is a **breaking change** for existing integrations:

1. **Cannot skip "verified" status**: Direct transition from `delivered` → `paid` is no longer allowed
2. **GRN creation blocked**: Cannot create GRN for POs in `"delivered"` status (must be `"verified"`)
3. **Auto-GRN trigger changed**: Auto-GRN creation moved from `"delivered"` to `"verified"` status

**Migration Path for Frontend**:
1. Update all PO status displays to include `"verified"`
2. Update status transition logic to enforce new workflow
3. Update GRN creation buttons to check for `"verified"` status
4. Add messaging for users: "Goods delivered. Complete quality inspection to verify."

---

## Rejected Goods Return Tracking

---

## Overview

This update adds a complete workflow for tracking and managing rejected goods from Goods Receipt Notes (GRNs). When goods are rejected during quality inspection, the system now tracks the return process to the vendor and automatically adjusts the purchase order payment calculations.

## Database Changes (Auto-Applied via GORM)

When you restart the backend server, GORM will automatically add the following fields to the `grn_items` table:

| Field Name | Type | Nullable | Description |
|------------|------|----------|-------------|
| `return_status` | VARCHAR(30) | Yes | Current return status: `pending`, `sent`, `received_by_vendor`, `closed` |
| `return_sent_date` | TIMESTAMPTZ | Yes | Date when rejected items were shipped back to vendor |
| `return_received_date` | TIMESTAMPTZ | Yes | Date when vendor confirmed receipt |
| `return_closed_date` | TIMESTAMPTZ | Yes | Date when return process was closed |
| `return_remarks` | TEXT | Yes | Notes about the return (reason, condition, etc.) |

**Important**: Existing GRN items with `rejected_quantity > 0` will automatically have `return_status` set to `"pending"` when first accessed through the new endpoints.

---

## New API Endpoints

### 1. Get Rejected Items for a GRN

**Endpoint**: `GET /api/v1/grns/:id/rejected-items`

**Description**: Retrieves all rejected items from a GRN with detailed return tracking information.

**Authentication**: Required (Bearer token)

**Authorization**: `grn:read` permission

**URL Parameters**:
- `id` (required): GRN ID (format: `GRNX_xxxxxxxx`)

**Response**: `200 OK`
```json
{
  "status": "success",
  "message": "Rejected items retrieved successfully",
  "data": {
    "grn_id": "GRNX_abc12345",
    "grn_number": "GRN-2025-0023",
    "po_id": "PORD_xyz78901",
    "po_number": "PO-2025-0015",
    "rejected_items": [
      {
        "id": "GRIT_item001",
        "variant_id": "PVAR_variant1",
        "product_name": "Premium Rice",
        "product_sku": "RICE-PREM-25KG",
        "rejected_quantity": 50,
        "unit_price": 1200.00,
        "total_value": 60000.00,
        "return_status": "pending",
        "return_sent_date": null,
        "return_received_date": null,
        "return_closed_date": null,
        "return_remarks": null
      }
    ],
    "total_rejected_value": 60000.00,
    "return_status_breakdown": {
      "pending": 2,
      "sent": 1,
      "received_by_vendor": 0,
      "closed": 0
    }
  }
}
```

**Error Responses**:
- `400 Bad Request`: Missing or invalid GRN ID
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: GRN not found or no rejected items exist
- `500 Internal Server Error`: Server error

---

### 2. Update Return Status for Rejected Item

**Endpoint**: `PATCH /api/v1/grns/items/:item_id/return-status`

**Description**: Updates the return status of a rejected GRN item. Status transitions follow a strict workflow.

**Authentication**: Required (Bearer token)

**Authorization**: `grn:update` permission

**URL Parameters**:
- `item_id` (required): GRN Item ID (format: `GRIT_xxxxxxxx`)

**Request Body**:
```json
{
  "return_status": "sent",
  "return_remarks": "Shipped via FedEx, tracking: 123456789"
}
```

**Request Fields**:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `return_status` | string | Yes | New status: `pending`, `sent`, `received_by_vendor`, or `closed` |
| `return_remarks` | string | No | Optional notes about the status change |

**Response**: `200 OK`
```json
{
  "status": "success",
  "message": "Return status updated successfully",
  "data": {
    "id": "GRIT_item001",
    "grn_id": "GRNX_abc12345",
    "po_item_id": "POIM_item123",
    "variant_id": "PVAR_variant1",
    "product_name": "Premium Rice",
    "product_sku": "RICE-PREM-25KG",
    "ordered_quantity": 100,
    "received_quantity": 100,
    "accepted_quantity": 50,
    "rejected_quantity": 50,
    "expiry_date": "2025-12-31",
    "batch_number": "BATCH-2025-001",
    "inventory_batch_id": "BATC_batch123",
    "return_status": "sent",
    "return_sent_date": "2025-11-20T10:30:00Z",
    "return_received_date": null,
    "return_closed_date": null,
    "return_remarks": "Shipped via FedEx, tracking: 123456789",
    "created_at": "2025-11-15T08:00:00Z"
  }
}
```

**Error Responses**:
- `400 Bad Request`: Invalid status transition (see workflow below)
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: GRN item not found
- `422 Unprocessable Entity`: Validation failed (missing required fields)
- `500 Internal Server Error`: Server error

---

## Status Workflow & Validation

### Valid Status Transitions

The return status follows a strict state machine:

```
pending → sent → received_by_vendor → closed
```

**Rules**:
1. **Cannot skip states**: Must progress through each state in order
2. **Cannot go backwards**: Once advanced, cannot revert to previous state
3. **Automatic date setting**: Each transition automatically sets the corresponding date field:
   - `pending` → `sent`: Sets `return_sent_date` to current timestamp
   - `sent` → `received_by_vendor`: Sets `return_received_date` to current timestamp
   - `received_by_vendor` → `closed`: Sets `return_closed_date` to current timestamp
4. **No transitions from closed**: Once closed, status cannot change

### Examples of Invalid Transitions

| Current Status | Attempted New Status | Result |
|----------------|---------------------|---------|
| `pending` | `received_by_vendor` | ❌ Error: "Invalid status transition" |
| `pending` | `closed` | ❌ Error: "Invalid status transition" |
| `sent` | `pending` | ❌ Error: "Invalid status transition" |
| `closed` | `sent` | ❌ Error: "Cannot transition from closed status" |

### Valid Transition Examples

| Current Status | New Status | Result |
|----------------|-----------|--------|
| `pending` | `sent` | ✅ Success, `return_sent_date` set |
| `sent` | `received_by_vendor` | ✅ Success, `return_received_date` set |
| `received_by_vendor` | `closed` | ✅ Success, `return_closed_date` set |

---

## Updated Response: Purchase Order

### Modified Fields

The Purchase Order response (`GET /api/v1/purchase-orders/:id` and all other PO endpoints) now includes two new calculated fields:

**New Fields**:

| Field Name | Type | Description |
|------------|------|-------------|
| `total_rejected_amount` | number | Total value of all rejected items from the GRN (sum of rejected_quantity × unit_price) |
| `amount_owed` | number | Amount the buyer owes to vendor: `total_amount - total_rejected_amount` |

**Example**:
```json
{
  "status": "success",
  "message": "Purchase order retrieved successfully",
  "data": {
    "id": "PORD_xyz78901",
    "po_number": "PO-2025-0015",
    "collaborator_id": "CLAB_vendor123",
    "collaborator_name": "ABC Suppliers Ltd",
    "warehouse_id": "WHSE_warehouse1",
    "warehouse_name": "Main Warehouse",
    "order_date": "2025-11-10",
    "expected_delivery_date": "2025-11-20",
    "actual_delivery_date": "2025-11-18",
    "status": "delivered",
    "total_amount": 250000.00,
    "total_rejected_amount": 60000.00,
    "amount_owed": 190000.00,
    "payment_status": "unpaid",
    "paid_amount": 0.00,
    "items": [...],
    "created_at": "2025-11-10T09:00:00Z",
    "updated_at": "2025-11-18T14:30:00Z"
  }
}
```

**Financial Calculation Logic**:
- `total_amount`: Original purchase order total (NEVER changes - historical record)
- `total_rejected_amount`: Automatically calculated from GRN rejected items
- `amount_owed`: `total_amount - total_rejected_amount`
- Payment is complete when: `paid_amount >= amount_owed`

---

## Updated Response: GRN Item

### Modified Fields

All GRN Item responses now include the new return tracking fields:

**New Fields in GRNItemResponse**:

```json
{
  "id": "GRIT_item001",
  "grn_id": "GRNX_abc12345",
  "po_item_id": "POIM_item123",
  "variant_id": "PVAR_variant1",
  "product_name": "Premium Rice",
  "product_sku": "RICE-PREM-25KG",
  "ordered_quantity": 100,
  "received_quantity": 100,
  "accepted_quantity": 50,
  "rejected_quantity": 50,
  "expiry_date": "2025-12-31",
  "batch_number": "BATCH-2025-001",
  "inventory_batch_id": "BATC_batch123",
  "return_status": "pending",
  "return_sent_date": null,
  "return_received_date": null,
  "return_closed_date": null,
  "return_remarks": null,
  "created_at": "2025-11-15T08:00:00Z"
}
```

**Note**: Return status fields will be `null` for items with `rejected_quantity = 0` (fully accepted items).

---

## Frontend Implementation Guide

### 1. Display Rejected Items

**When to show**: On GRN detail page, add a "Rejected Items" section if `rejected_quantity > 0` exists.

**Example UI**:
```
┌─────────────────────────────────────────────────┐
│ Rejected Items (3 items, ₹85,000 value)       │
├─────────────────────────────────────────────────┤
│ Product: Premium Rice (50 kg)                   │
│ Status: ⏳ Pending Return                       │
│ Value: ₹60,000                                  │
│ [Mark as Sent] [Add Remarks]                    │
├─────────────────────────────────────────────────┤
│ Product: Wheat Flour (20 bags)                  │
│ Status: 📦 Sent to Vendor                       │
│ Sent Date: Nov 18, 2025                         │
│ Tracking: Shipped via FedEx                     │
│ [Mark as Received by Vendor]                    │
└─────────────────────────────────────────────────┘
```

### 2. Update Return Status Flow

**Step 1**: User clicks "Mark as Sent" button

**Step 2**: Show modal/form:
```
┌─────────────────────────────────────┐
│ Update Return Status                │
├─────────────────────────────────────┤
│ Current Status: Pending             │
│ New Status: Sent ▼                  │
│                                     │
│ Remarks (optional):                 │
│ ┌─────────────────────────────────┐ │
│ │ Shipped via FedEx               │ │
│ │ Tracking: 123456789             │ │
│ └─────────────────────────────────┘ │
│                                     │
│ [Cancel]  [Update Status]           │
└─────────────────────────────────────┘
```

**Step 3**: Make API call:
```javascript
const updateReturnStatus = async (itemId, status, remarks) => {
  try {
    const response = await fetch(
      `/api/v1/grns/items/${itemId}/return-status`,
      {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          return_status: status,
          return_remarks: remarks || undefined
        })
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message);
    }

    const result = await response.json();
    // Update UI with new status and dates
    updateItemInUI(result.data);
  } catch (error) {
    // Show error toast: "Invalid status transition" or other error
    showErrorToast(error.message);
  }
};
```

### 3. Purchase Order Payment Display

**Update your PO display to show the correct amount owed**:

```jsx
<div className="po-financial-summary">
  <div className="row">
    <span>Original Order Amount:</span>
    <span>₹{po.total_amount.toLocaleString()}</span>
  </div>

  {po.total_rejected_amount > 0 && (
    <div className="row text-danger">
      <span>Less: Rejected Items:</span>
      <span>- ₹{po.total_rejected_amount.toLocaleString()}</span>
    </div>
  )}

  <div className="row total-row">
    <span><strong>Amount Owed to Vendor:</strong></span>
    <span><strong>₹{po.amount_owed.toLocaleString()}</strong></span>
  </div>

  <div className="row">
    <span>Paid Amount:</span>
    <span>₹{po.paid_amount.toLocaleString()}</span>
  </div>

  <div className="row balance-row">
    <span>Balance Due:</span>
    <span className={po.paid_amount >= po.amount_owed ? 'text-success' : 'text-warning'}>
      ₹{(po.amount_owed - po.paid_amount).toLocaleString()}
    </span>
  </div>
</div>
```

### 4. Status Badge Component

**Create a reusable status badge**:

```jsx
const ReturnStatusBadge = ({ status }) => {
  const statusConfig = {
    pending: { color: 'warning', icon: '⏳', text: 'Pending Return' },
    sent: { color: 'info', icon: '📦', text: 'Sent to Vendor' },
    received_by_vendor: { color: 'primary', icon: '✅', text: 'Received by Vendor' },
    closed: { color: 'success', icon: '🔒', text: 'Closed' }
  };

  const config = statusConfig[status] || statusConfig.pending;

  return (
    <span className={`badge badge-${config.color}`}>
      {config.icon} {config.text}
    </span>
  );
};
```

### 5. Next Action Buttons

**Show appropriate action based on current status**:

```jsx
const ReturnActionButtons = ({ item, onStatusUpdate }) => {
  const nextActions = {
    pending: { status: 'sent', label: 'Mark as Sent', icon: '📦' },
    sent: { status: 'received_by_vendor', label: 'Vendor Received', icon: '✅' },
    received_by_vendor: { status: 'closed', label: 'Close Return', icon: '🔒' },
    closed: null
  };

  const action = nextActions[item.return_status];

  if (!action) {
    return <span className="text-muted">Return Closed</span>;
  }

  return (
    <button
      className="btn btn-sm btn-primary"
      onClick={() => onStatusUpdate(item.id, action.status)}
    >
      {action.icon} {action.label}
    </button>
  );
};
```

---

## Error Handling

### Common Errors and Solutions

| Error Code | Message | Frontend Action |
|------------|---------|-----------------|
| 400 | Invalid status transition | Show error toast explaining valid next status |
| 400 | Cannot transition from closed status | Disable action buttons for closed items |
| 404 | GRN item not found | Refresh page or show "Item not found" message |
| 404 | No rejected items found for this GRN | Hide "Rejected Items" section |
| 422 | return_status is required | Validate form before submission |

### Example Error Handling

```javascript
try {
  await updateReturnStatus(itemId, newStatus, remarks);
  showSuccessToast('Return status updated successfully');
} catch (error) {
  if (error.message.includes('Invalid status transition')) {
    showErrorToast(`Cannot change status from ${currentStatus} to ${newStatus}. Please follow the workflow: pending → sent → received_by_vendor → closed`);
  } else if (error.message.includes('Cannot transition from closed')) {
    showErrorToast('This return is already closed and cannot be modified');
  } else {
    showErrorToast('Failed to update return status. Please try again.');
  }
}
```

---

## Testing Checklist

### Backend Verification
- [ ] Server starts without errors (GORM auto-migration runs)
- [ ] New columns added to `grn_items` table
- [ ] GET `/api/v1/grns/:id/rejected-items` returns rejected items
- [ ] PATCH `/api/v1/grns/items/:item_id/return-status` updates status
- [ ] Invalid status transitions are rejected with 400 error
- [ ] Purchase order responses include `total_rejected_amount` and `amount_owed`

### Frontend Integration
- [ ] Rejected items section appears on GRN detail page
- [ ] Status badges display correctly with icons
- [ ] Action buttons show appropriate next status
- [ ] Status update modal/form captures remarks
- [ ] API calls include authentication token
- [ ] Success messages shown after status update
- [ ] Error messages shown for invalid transitions
- [ ] PO financial summary shows rejected amount and amount owed
- [ ] Status transitions follow workflow (pending → sent → received_by_vendor → closed)

---

## Breaking Changes

**None**. All changes are additive:
- New optional fields in GRN item responses
- New calculated fields in PO responses
- New endpoints (existing endpoints unchanged)

## Backward Compatibility

- Existing GRN items will have `return_status = null` until accessed via new endpoints
- Existing PO responses will show `total_rejected_amount = 0` if no GRN exists
- Frontend can safely ignore new fields if not implementing the feature

---

## Support

For questions or issues:
- Backend API: Check server logs for detailed error messages
- Database: All changes handled via GORM AutoMigrate on server restart
- Integration: See Frontend Implementation Guide above

---

**End of Changes Document**
