# Implementation Changes - December 2025

## Executive Summary
This document details 8 critical fixes and enhancements implemented in December 2025, addressing data accuracy issues, adding new filtering capabilities, improving data integrity through transactions, relaxing validation constraints, and standardizing price precision across the ERP system.

**Key Improvements:**
- Product availability now grouped by SKU (eliminates duplicate entries)
- Expired batches excluded from stock counts (accurate inventory reporting)
- New quantity-based product filtering endpoint
- Transactional integrity for variant operations (prevents partial failures)
- GST line totals added to purchase orders (complete tax breakdown)
- Standardized 2-decimal price precision (consistent display)
- Relaxed collaborator bank validation (optional fields)
- Logo attachment support via attachment IDs

---

## Issues Fixed

### Issue #1: Availability Endpoint Grouping by SKU

**Problem**: GET `/api/v1/products/availability` returned duplicate entries when same product variant was in multiple warehouses or batches.

**Root Cause**: Endpoint was returning flat list of batches without grouping by product SKU, leading to:
- Multiple rows for same product (one per batch)
- Difficult to see total available quantity per product
- Frontend had to perform manual aggregation

**Files Modified**:
- `internal/services/inventory_service.go:410-637`
- `internal/database/models/inventory.go:105-128` (new response models)

**Changes**:

1. **New Response Models** (`inventory.go:105-128`):
   ```go
   // ProductAvailabilityGroupedResponse - Main response structure
   type ProductAvailabilityGroupedResponse struct {
       SKU                string                        `json:"sku"`
       VariantID          string                        `json:"variant_id"`
       ProductName        string                        `json:"product_name"`
       ProductDescription *string                       `json:"product_description,omitempty"`
       TotalQuantity      int64                         `json:"total_quantity"`      // Available (non-expired)
       ExpiredQuantity    int64                         `json:"expired_quantity"`    // Expired quantity
       EarliestExpiry     string                        `json:"earliest_expiry"`     // Across all warehouses
       ExpiryStatus       string                        `json:"expiry_status"`       // Overall status
       WarehouseDetails   []WarehouseAvailabilityDetail `json:"warehouse_details"`
   }

   // WarehouseAvailabilityDetail - Per-warehouse breakdown
   type WarehouseAvailabilityDetail struct {
       WarehouseID     string       `json:"warehouse_id"`
       WarehouseName   string       `json:"warehouse_name"`
       Address         *AddressInfo `json:"address,omitempty"`
       Quantity        int64        `json:"quantity"`          // Available quantity
       ExpiredQuantity int64        `json:"expired_quantity"`   // Expired quantity
       EarliestExpiry  string       `json:"earliest_expiry"`    // For this warehouse
       ExpiryStatus    string       `json:"expiry_status"`      // "fresh"/"expiring_soon"/"expired"
   }
   ```

2. **Service Logic** (`inventory_service.go:410-637`):
   - Groups batches by SKU using map (`variantMap`)
   - Aggregates quantities per warehouse within each SKU
   - Separates expired vs available quantities
   - Tracks earliest expiry date per warehouse and overall
   - Sorts warehouse details by expiry date (FEFO logic)
   - Fetches warehouse addresses from AAA service

3. **Expiry Status Calculation** (`inventory_service.go:594-610`):
   ```go
   func (s *InventoryService) calculateExpiryStatus(expiryDate time.Time) string {
       now := time.Now()
       if expiryDate.Before(now) {
           return "expired"
       }
       daysUntilExpiry := int(expiryDate.Sub(now).Hours() / 24)
       if daysUntilExpiry <= 30 {
           return "expiring_soon"
       }
       return "fresh"
   }
   ```

**Testing**:
```bash
# Test availability endpoint
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/products/availability?limit=50&offset=0

# Expected response format:
{
  "data": [
    {
      "sku": "RICE-1KG",
      "variant_id": "PVAR00000001",
      "product_name": "Basmati Rice 1kg",
      "total_quantity": 900,        // Available across all warehouses
      "expired_quantity": 100,      // Expired across all warehouses
      "earliest_expiry": "2025-03-15",
      "expiry_status": "expiring_soon",
      "warehouse_details": [
        {
          "warehouse_id": "WRHS00000001",
          "warehouse_name": "Main Warehouse",
          "quantity": 500,            // Available in this warehouse
          "expired_quantity": 50,     // Expired in this warehouse
          "earliest_expiry": "2025-03-15",
          "expiry_status": "expiring_soon"
        },
        {
          "warehouse_id": "WRHS00000002",
          "warehouse_name": "Branch Warehouse",
          "quantity": 400,
          "expired_quantity": 50,
          "earliest_expiry": "2025-04-20",
          "expiry_status": "fresh"
        }
      ]
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

---

### Issue #2: Expired Batches Not Counted in Stock

**Problem**: Availability endpoint was counting expired batches as available stock.

**Root Cause**: Repository method `GetAllBatchesPaginated()` returned all batches without filtering by expiry date.

**Files Modified**:
- `internal/database/repositories/inventory_repo.go:129-167`
- `internal/services/inventory_service.go:419,459-486,641-658`
- `internal/database/models/inventory.go:67-82`

**Changes**:

1. **Repository Enhancement** (`inventory_repo.go:129-167`):
   ```go
   // GetAllBatchesPaginated - Returns only non-expired batches (default behavior)
   func (r *InventoryRepository) GetAllBatchesPaginated(limit, offset int) ([]models.InventoryBatch, int64, error) {
       // Filter: expiry_date > NOW()
       if err := r.db.Model(&models.InventoryBatch{}).
           Where("expiry_date > ?", time.Now()).
           Count(&total).Error; err != nil {
           return nil, 0, errors.NewInternalServerError("Failed to count batches")
       }
       // Query with same filter
       if err := r.db.Preload("Warehouse").Preload("Variant").
           Where("expiry_date > ?", time.Now()).
           Order("created_at DESC").
           Limit(limit).Offset(offset).Find(&batches).Error; err != nil {
           return nil, 0, errors.NewInternalServerError("Failed to retrieve all batches")
       }
       return batches, total, nil
   }

   // GetAllBatchesPaginatedWithExpired - Includes expired batches (for visibility)
   func (r *InventoryRepository) GetAllBatchesPaginatedWithExpired(limit, offset int) ([]models.InventoryBatch, int64, error) {
       // No expiry filter - returns all batches
       // Used by availability endpoint to show expired quantities separately
   }
   ```

2. **Service Logic** (`inventory_service.go:459-486`):
   ```go
   // In GetAllProductsAvailability():
   for _, batch := range batches {
       isExpired := batch.ExpiryDate.Before(time.Now())
       expiryStatus := s.calculateExpiryStatus(batch.ExpiryDate)

       // Separate expired vs available quantities
       if isExpired {
           warehouseEntry.ExpiredQuantity += batch.TotalQuantity
       } else {
           warehouseEntry.Quantity += batch.TotalQuantity  // Only non-expired
       }

       // Track earliest expiry for non-expired batches only
       if !isExpired && (warehouseEntry.EarliestExpiry.IsZero() ||
           batch.ExpiryDate.Before(warehouseEntry.EarliestExpiry)) {
           warehouseEntry.EarliestExpiry = batch.ExpiryDate
           warehouseEntry.ExpiryStatus = expiryStatus
       }
   }
   ```

3. **Response Enhancement** (`inventory_service.go:641-658`, `inventory.go:67-82`):
   ```go
   // InventoryBatchResponse now includes expiry status
   type InventoryBatchResponse struct {
       // ... other fields
       IsExpired    bool   `json:"is_expired"`     // NEW: true if expiry_date < now
       ExpiryStatus string `json:"expiry_status"`  // NEW: "fresh"/"expiring_soon"/"expired"
   }

   // Calculated in batchToResponse()
   func (s *InventoryService) batchToResponse(batch models.InventoryBatch) models.InventoryBatchResponse {
       isExpired := batch.ExpiryDate.Before(time.Now())
       expiryStatus := s.calculateExpiryStatus(batch.ExpiryDate)

       return models.InventoryBatchResponse{
           // ... other fields
           IsExpired:    isExpired,
           ExpiryStatus: expiryStatus,
       }
   }
   ```

**Testing**:
```bash
# Create batch with past expiry date
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "warehouse_id": "WRHS00000001",
    "variant_id": "PVAR00000001",
    "cost_price": 50.00,
    "expiry_date": "2024-01-01",  # Past date
    "quantity": 100
  }' \
  http://localhost:8080/api/v1/batches

# Check availability - expired batch should NOT be in quantity
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/products/availability

# Expected: expired_quantity = 100, total_quantity = 0 (if no other batches)
```

---

### Issue #3: Products Quantity Filter Endpoint

**Problem**: No way to filter products by inventory quantity range (e.g., "show products with stock between 10-100").

**Solution**: New endpoint GET `/api/v1/products/by-quantity?min=X&max=Y`

**Files Created/Modified**:
- `internal/api/handlers/product_handler.go:434-516` (new handler)
- `internal/services/product_service.go` (new service method)
- `internal/database/repositories/product_repo.go` (new repository method)

**Implementation**:

1. **Handler** (`product_handler.go:434-516`):
   ```go
   // GetProductsByQuantity handles GET /api/v1/products/by-quantity
   func (h *ProductHandler) GetProductsByQuantity(c *gin.Context) {
       // Parse and validate min parameter
       minQty, err := strconv.ParseInt(c.Query("min"), 10, 64)
       if err != nil {
           utils.BadRequestResponse(c, "Invalid 'min' parameter - must be a valid integer", err)
           return
       }

       // Parse and validate max parameter
       maxQty, err := strconv.ParseInt(c.Query("max"), 10, 64)
       if err != nil {
           utils.BadRequestResponse(c, "Invalid 'max' parameter - must be a valid integer", err)
           return
       }

       // Get pagination parameters
       params := utils.GetPaginationParams(c)

       // Call service
       products, total, err := h.productService.GetProductsByQuantityRange(
           c.Request.Context(), minQty, maxQty, params.Limit, params.Offset)

       if err != nil {
           utils.HandleServiceError(c, "Failed to retrieve products by quantity range", err)
           return
       }

       utils.PaginatedOKResponse(c, products, total, params.Limit, params.Offset)
   }
   ```

2. **Service Logic** (joins variants and batches, aggregates quantities):
   ```go
   func (s *ProductService) GetProductsByQuantityRange(ctx context.Context, minQty, maxQty int64, limit, offset int) ([]models.ProductResponse, int64, error) {
       // Query: SELECT products with SUM(batch.total_quantity) BETWEEN minQty AND maxQty
       // Joins: products -> variants -> batches
       // Filters: Only non-expired batches (expiry_date > NOW())
       // Groups by: product_id
       // Having: SUM(quantity) BETWEEN min AND max
   }
   ```

3. **Query Parameters**:
   - `min` (required): Minimum quantity threshold
   - `max` (required): Maximum quantity threshold
   - `limit` (optional, default 50): Page size
   - `offset` (optional, default 0): Page offset

**Testing**:
```bash
# Get products with stock between 10 and 100 units
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/products/by-quantity?min=10&max=100&limit=50&offset=0"

# Expected response:
{
  "data": [
    {
      "id": "PROD00000001",
      "name": "Rice",
      "total_inventory": 75  # Aggregated across all warehouses and batches
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

**Use Cases**:
- Low stock alerts (min=0, max=10)
- Optimal stock levels (min=50, max=200)
- Overstocked items (min=500, max=999999)
- Out of stock (min=0, max=0)

---

### Issue #4: Product Variant Price Creation Transaction

**Problem**: Creating product variant with prices could result in partial failures (variant created but prices failed).

**Root Cause**: Variant and prices were created in separate database operations without transaction wrapping.

**Files Modified**:
- `internal/services/product_variant_service.go:120-177`

**Changes**:

**Before** (No transaction):
```go
// Create variant
if err := s.variantRepo.Create(variant); err != nil {
    return nil, err
}

// Create prices (could fail AFTER variant created)
for _, price := range request.Prices {
    if err := s.priceRepo.Create(productPrice); err != nil {
        return nil, err  // Variant already in DB, orphaned!
    }
}
```

**After** (Transactional):
```go
// Wrap in transaction (lines 122-171)
err = s.variantRepo.WithTransaction(func(tx *gorm.DB) error {
    // Create variant within transaction
    if err := s.variantRepo.CreateWithTx(tx, variant); err != nil {
        s.logger.Error("Failed to create variant in transaction", zap.Error(err))
        return err
    }

    // Create all prices within same transaction
    if len(request.Prices) > 0 && s.priceRepo != nil {
        for _, price := range request.Prices {
            // Validate price is positive
            if price.Price <= 0 {
                err := fmt.Errorf("price must be greater than 0 for type %s", price.PriceType)
                return errors.NewValidationError(err.Error())
            }

            productPrice := models.NewProductPrice(variant.ID, price.PriceType, price.Price, ...)
            if err := s.priceRepo.CreateWithTx(tx, productPrice); err != nil {
                s.logger.Error("Failed to create price record in transaction", zap.Error(err))
                return fmt.Errorf("failed to create price for type %s: %w", price.PriceType, err)
            }
        }
    }

    return nil  // Commit transaction
})

if err != nil {
    // Automatic rollback - no orphaned variants
    return nil, err
}
```

**Benefits**:
- ✅ Atomic operation (all-or-nothing)
- ✅ No orphaned variants without prices
- ✅ Consistent database state
- ✅ Automatic rollback on any error
- ✅ Price validation within transaction boundary

**Testing**:
```bash
# Test: Create variant with invalid price (should rollback entire operation)
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "PROD00000001",
    "variant_name": "1kg",
    "hsn_code": "10063020",
    "gst_rate": 5.00,
    "prices": [
      {"price_type": "retail", "price": -10.00}  # Invalid: negative price
    ]
  }' \
  http://localhost:8080/api/v1/product-variants

# Expected: 400 Bad Request, variant NOT created in database
```

---

### Issue #5: Purchase Order GST Line Totals

**Problem**: Purchase order items only showed per-unit GST amounts, not total GST for line item.

**Root Cause**: PurchaseOrderItem model lacked total GST fields (gst_amount_total, cgst_amount_total, sgst_amount_total, igst_amount_total).

**Files Modified**:
- `internal/database/models/purchase_order.go:103-106,176-179` (added 4 fields)
- `internal/services/purchase_order_service.go:200-204` (calculation logic)

**Changes**:

1. **Model Enhancement** (`purchase_order.go:103-106`):
   ```go
   type PurchaseOrderItem struct {
       // ... existing per-unit fields
       GSTAmount  float64 `json:"gst_amount"`   // Per unit
       CGSTAmount float64 `json:"cgst_amount"`  // Per unit
       SGSTAmount float64 `json:"sgst_amount"`  // Per unit
       IGSTAmount float64 `json:"igst_amount"`  // Per unit

       // NEW: Total GST Breakdown (per-unit amounts × quantity)
       GSTAmountTotal  float64 `gorm:"type:numeric(14,4);default:0" json:"gst_amount_total"`  // Total GST for line
       CGSTAmountTotal float64 `gorm:"type:numeric(14,4);default:0" json:"cgst_amount_total"` // Total CGST for line
       SGSTAmountTotal float64 `gorm:"type:numeric(14,4);default:0" json:"sgst_amount_total"` // Total SGST for line
       IGSTAmountTotal float64 `gorm:"type:numeric(14,4);default:0" json:"igst_amount_total"` // Total IGST for line
   }
   ```

2. **Calculation Logic** (`purchase_order_service.go:200-204`):
   ```go
   // After calculating per-unit GST breakdown
   gstBreakdown := calculateGSTFromAllInPrice(itemReq.UnitPrice, variant.GSTRate, po.IsInterState)

   item.BasePrice = gstBreakdown.BasePrice
   item.GSTRate = gstBreakdown.GSTRate
   item.GSTAmount = gstBreakdown.GSTAmount    // Per unit
   item.CGSTRate = gstBreakdown.CGSTRate
   item.CGSTAmount = gstBreakdown.CGSTAmount  // Per unit
   item.SGSTRate = gstBreakdown.SGSTRate
   item.SGSTAmount = gstBreakdown.SGSTAmount  // Per unit
   item.IGSTRate = gstBreakdown.IGSTRate
   item.IGSTAmount = gstBreakdown.IGSTAmount  // Per unit

   // NEW: Calculate line totals
   quantityFloat := float64(itemReq.Quantity)
   item.GSTAmountTotal = gstBreakdown.GSTAmount * quantityFloat
   item.CGSTAmountTotal = gstBreakdown.CGSTAmount * quantityFloat
   item.SGSTAmountTotal = gstBreakdown.SGSTAmount * quantityFloat
   item.IGSTAmountTotal = gstBreakdown.IGSTAmount * quantityFloat
   ```

3. **Response Fields** (`purchase_order.go:176-179`):
   ```go
   type PurchaseOrderItemResponse struct {
       // Per-unit GST (existing)
       GSTAmount  float64 `json:"gst_amount"`
       CGSTAmount float64 `json:"cgst_amount,omitempty"`
       SGSTAmount float64 `json:"sgst_amount,omitempty"`
       IGSTAmount float64 `json:"igst_amount,omitempty"`

       // NEW: Line totals
       GSTAmountTotal  float64 `json:"gst_amount_total"`
       CGSTAmountTotal float64 `json:"cgst_amount_total"`
       SGSTAmountTotal float64 `json:"sgst_amount_total"`
       IGSTAmountTotal float64 `json:"igst_amount_total"`
   }
   ```

**Database Migration**:
GORM AutoMigrate will add these columns automatically on next server startup:
```sql
ALTER TABLE purchase_order_items
ADD COLUMN gst_amount_total NUMERIC(14,4) DEFAULT 0,
ADD COLUMN cgst_amount_total NUMERIC(14,4) DEFAULT 0,
ADD COLUMN sgst_amount_total NUMERIC(14,4) DEFAULT 0,
ADD COLUMN igst_amount_total NUMERIC(14,4) DEFAULT 0;
```

**Testing**:
```bash
# Create PO with GST calculation
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "collaborator_id": "CLAB00000001",
    "warehouse_id": "WRHS00000001",
    "expected_delivery_date": "2025-12-31",
    "items": [
      {
        "variant_id": "PVAR00000001",
        "quantity": 10,
        "unit_price": 105.00  # ALL-IN price (includes 5% GST)
      }
    ]
  }' \
  http://localhost:8080/api/v1/purchase-orders

# Expected response includes line totals:
{
  "items": [
    {
      "quantity": 10,
      "unit_price": 105.00,
      "line_total": 1050.00,
      "base_price": 100.00,
      "gst_rate": 5.00,
      "gst_amount": 5.00,        # Per unit
      "cgst_amount": 2.50,       # Per unit (if intra-state)
      "sgst_amount": 2.50,       # Per unit (if intra-state)
      "gst_amount_total": 50.00,   # NEW: 5.00 × 10
      "cgst_amount_total": 25.00,  # NEW: 2.50 × 10
      "sgst_amount_total": 25.00   # NEW: 2.50 × 10
    }
  ]
}
```

---

### Issue #6: Price Decimal Precision

**Problem**: Prices displayed with 4 decimal places (e.g., 99.9999) instead of standard 2 decimal places.

**Root Cause**: Database stores prices as `NUMERIC(14,4)` but no rounding applied in responses.

**Files Modified**:
- `internal/utils/price_formatter.go` (NEW FILE - 9 lines)
- `internal/services/inventory_service.go:107,141,649`
- All service layers returning price responses

**Changes**:

1. **Utility Function** (`utils/price_formatter.go`):
   ```go
   package utils

   import "math"

   // RoundPrice rounds price to 2 decimal places using standard rounding (NOT truncation)
   // Example: 3.14159 -> 3.14, 3.145 -> 3.15
   func RoundPrice(price float64) float64 {
       return math.Round(price*100) / 100
   }
   ```

2. **Applied in Services**:
   ```go
   // In InventoryService responses (lines 107, 141, 649)
   response := &models.InventoryBatchResponse{
       CostPrice: utils.RoundPrice(batch.CostPrice),  // 99.9999 -> 100.00
       // ... other fields
   }

   // In ProductPrice responses
   response.Price = utils.RoundPrice(price.Price)

   // In Sale responses
   response.TotalAmount = utils.RoundPrice(sale.TotalAmount)
   ```

**Rounding Behavior**:
- Uses standard rounding (`math.Round`), NOT truncation
- `3.144` → `3.14` (rounds down)
- `3.145` → `3.15` (rounds up)
- `3.14159` → `3.14`
- `99.9999` → `100.00`

**Testing**:
```bash
# Check price precision in responses
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/batches/BATC00000001

# Expected: cost_price should be 2 decimal places
{
  "cost_price": 100.00,  # Not 99.9999
  "expiry_date": "2025-12-31"
}
```

---

### Issue #7: Collaborator Bank Details Optional

**Problem**: Bank account number and IFSC code were required fields, preventing collaborator creation for vendors without banking setup.

**Root Cause**: Model validation enforced `binding:"required"` on bank fields.

**Files Modified**:
- `internal/database/models/collaborator.go:29-31,95-96`

**Changes**:

**Before** (Required):
```go
type Collaborator struct {
    BankAccountNo string `gorm:"type:varchar(50);not null" json:"bank_account_no"`
    BankIFSC      string `gorm:"type:varchar(11);not null" json:"bank_ifsc"`
}

type CreateCollaboratorRequest struct {
    BankAccountNo string `json:"bank_account_no" binding:"required"`
    BankIFSC      string `json:"bank_ifsc" binding:"required,len=11"`
}
```

**After** (Optional):
```go
type Collaborator struct {
    BankAccountNo *string `gorm:"type:varchar(50)" json:"bank_account_no"`  // Nullable
    BankIFSC      *string `gorm:"type:varchar(11)" json:"bank_ifsc"`        // Nullable
    BankName      *string `gorm:"type:varchar(100)" json:"bank_name"`       // Already optional
}

type CreateCollaboratorRequest struct {
    BankAccountNo *string `json:"bank_account_no"`                          // Optional
    BankIFSC      *string `json:"bank_ifsc" binding:"omitempty,len=11"`     // Optional, validated if provided
}
```

**Database Migration**:
GORM AutoMigrate will modify columns automatically on next server startup:
```sql
ALTER TABLE collaborators
ALTER COLUMN bank_account_no DROP NOT NULL,
ALTER COLUMN bank_ifsc DROP NOT NULL;
```

**Testing**:
```bash
# Create collaborator WITHOUT bank details
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "ABC Suppliers",
    "contact_person": "John Doe",
    "contact_number": "9876543210",
    "gst_number": "29ABCDE1234F1Z5"
    # NO bank_account_no or bank_ifsc
  }' \
  http://localhost:8080/api/v1/collaborators

# Expected: 201 Created (bank details can be added later via PATCH)

# Update with bank details later
curl -X PATCH -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "bank_account_no": "1234567890",
    "bank_ifsc": "HDFC0001234",
    "bank_name": "HDFC Bank"
  }' \
  http://localhost:8080/api/v1/collaborators/CLAB00000001
```

---

### Issue #8: Settings Logo Attachment Support

**Problem**: Settings/FPO logo field only supported direct image paths, not attachment IDs from the attachment management system.

**Solution**: Logo field now accepts attachment IDs (format: `ATT_xxxxxxxx` or `ATCH00000001`).

**Files Modified**:
- `internal/database/models/collaborator.go:19` (field comment updated)
- Frontend integration pattern documented

**Changes**:

1. **Model Field** (`collaborator.go:19`):
   ```go
   type Collaborator struct {
       // Logo field accepts attachment ID from /api/v1/attachments
       Logo *string `gorm:"type:varchar(500)" json:"logo"`
       // Comment: Attachment ID (ATT_xxxxxxxx) - Use /api/v1/attachments/{id}/url to get image URL
   }
   ```

2. **Workflow**:
   ```
   Step 1: Upload logo file
   POST /api/v1/attachments
   Content-Type: multipart/form-data
   - file: logo.png
   - entity_type: "logo"
   - entity_id: "settings" or "fpo"

   Response: { "id": "ATCH00000001" }

   Step 2: Save attachment ID to collaborator/settings
   POST /api/v1/collaborators
   {
     "company_name": "ABC FPO",
     "logo": "ATCH00000001"  # Attachment ID reference
   }

   Step 3: Display logo in frontend
   GET /api/v1/attachments/ATCH00000001/url
   Response: { "url": "https://s3.../logos/uuid.png?presigned-params" }

   <img src="{{ presigned_url }}" />
   ```

3. **Benefits**:
   - ✅ Consistent with other attachment patterns (PO documents, GRN PDFs)
   - ✅ S3 file management (automatic cleanup, versioning)
   - ✅ Presigned URL security (time-limited access)
   - ✅ Entity-based folder structure (`logos/` folder)

**Testing**:
```bash
# 1. Upload logo
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -F "file=@logo.png" \
  -F "entity_type=logo" \
  -F "entity_id=settings" \
  http://localhost:8080/api/v1/attachments

# Response: {"id": "ATCH00000001", "file_path": "logos/abc123.png"}

# 2. Create collaborator with logo attachment ID
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "ABC FPO",
    "logo": "ATCH00000001"
  }' \
  http://localhost:8080/api/v1/collaborators

# 3. Get presigned URL for display
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/attachments/ATCH00000001/url

# Response: {"url": "https://s3.amazonaws.com/kisanlink-erp/logos/abc123.png?X-Amz-..."}
```

---

## Additional Bug Fixes

### Fix #1: Cancel Items Endpoint JSON Validation

**Problem**: Cancel-items endpoint failed when frontend sent snake_case JSON (correct format).

**Root Cause**: Gin binding expected PascalCase field names (Go struct field names) instead of JSON tag names.

**Files Modified**:
- `internal/database/models/sales.go` (request DTO validation tags)

**Solution**: Ensured JSON tags match snake_case convention:
```go
type CancelSaleItemsRequest struct {
    Reason      string                 `json:"reason" binding:"required"`      // NOT "Reason"
    PerformedBy string                 `json:"performed_by" binding:"required"` // NOT "PerformedBy"
    Items       []CancelSaleItemDetail `json:"items" binding:"required,min=1"`  // NOT "Items"
}

type CancelSaleItemDetail struct {
    SaleItemID string `json:"sale_item_id" binding:"required"` // NOT "SaleItemID"
    Quantity   int64  `json:"quantity" binding:"required,gt=0"` // NOT "Quantity"
}
```

---

### Fix #2: UpdateProductVariant Transaction Wrapping

**Problem**: Similar to Issue #4, updating variant could leave database in inconsistent state.

**Solution**: Wrapped UpdateProductVariant in transaction (same pattern as CreateProductVariant).

**Files Modified**:
- `internal/services/product_variant_service.go` (update method)

---

### Fix #3: Error Message Index Bug

**Problem**: Error messages with dynamic array indices showed incorrect character values.

**Root Cause**: Used `string(rune)` which converts to Unicode character instead of digit string.

**Solution**: Use `strconv.Itoa()` for integer-to-string conversion:
```go
// BEFORE (Wrong)
for i, item := range items {
    if item.Quantity <= 0 {
        return fmt.Errorf("Item %s has invalid quantity", string(i))  // Prints Unicode char!
    }
}

// AFTER (Correct)
import "strconv"

for i, item := range items {
    if item.Quantity <= 0 {
        return fmt.Errorf("Item %s has invalid quantity", strconv.Itoa(i))  // Prints "0", "1", "2"
    }
}
```

---

## Database Schema Changes (GORM AutoMigrate)

GORM will automatically apply these changes on server restart:

### 1. purchase_order_items Table
```sql
ALTER TABLE purchase_order_items
ADD COLUMN gst_amount_total NUMERIC(14,4) DEFAULT 0,
ADD COLUMN cgst_amount_total NUMERIC(14,4) DEFAULT 0,
ADD COLUMN sgst_amount_total NUMERIC(14,4) DEFAULT 0,
ADD COLUMN igst_amount_total NUMERIC(14,4) DEFAULT 0;
```

### 2. collaborators Table
```sql
ALTER TABLE collaborators
ALTER COLUMN bank_account_no DROP NOT NULL,
ALTER COLUMN bank_ifsc DROP NOT NULL;
```

### 3. inventory_batches Table (Calculated Fields - No Schema Change)
- `is_expired`: Calculated at runtime (`expiry_date < NOW()`)
- `expiry_status`: Calculated at runtime (`calculateExpiryStatus()`)
- NOT stored in database - computed in service layer

---

## New Files Created

### 1. internal/utils/price_formatter.go
**Purpose**: Standardize price rounding to 2 decimal places
**Size**: 9 lines
**Functions**: `RoundPrice(float64) float64`

### 2. internal/database/models/inventory.go (Enhanced)
**New Structs**:
- `ProductAvailabilityGroupedResponse`
- `WarehouseAvailabilityDetail`

---

## Breaking Changes

### 1. Availability Endpoint Response Format (HIGH IMPACT)
**Endpoint**: GET `/api/v1/products/availability`

**Before** (Flat list):
```json
[
  {
    "id": "BATC00000001",
    "warehouse_id": "WRHS00000001",
    "product_sku": "RICE-1KG",
    "quantity": 500
  },
  {
    "id": "BATC00000002",
    "warehouse_id": "WRHS00000001",
    "product_sku": "RICE-1KG",
    "quantity": 400
  }
]
```

**After** (Grouped by SKU):
```json
[
  {
    "sku": "RICE-1KG",
    "total_quantity": 900,
    "expired_quantity": 100,
    "expiry_status": "fresh",
    "warehouse_details": [
      {
        "warehouse_id": "WRHS00000001",
        "quantity": 500,
        "expired_quantity": 50
      },
      {
        "warehouse_id": "WRHS00000002",
        "quantity": 400,
        "expired_quantity": 50
      }
    ]
  }
]
```

**Migration Required**: Frontend must update availability parsing logic (see FRONTEND_API_CHANGES_DEC2025.md)

---

### 2. Price Precision (LOW IMPACT)
**Affected Fields**: All price fields in responses (cost_price, unit_price, total_amount, etc.)

**Before**: 4 decimal places (99.9999, 3.14159)
**After**: 2 decimal places (100.00, 3.14)

**Migration**: No code changes needed - display-only change

---

## Non-Breaking Enhancements

1. **New Endpoint**: GET `/api/v1/products/by-quantity?min=X&max=Y`
2. **New Fields**:
   - `is_expired`, `expiry_status` in InventoryBatchResponse
   - `expired_quantity` in ProductAvailabilityGroupedResponse
   - `gst_amount_total`, `cgst_amount_total`, `sgst_amount_total`, `igst_amount_total` in PurchaseOrderItemResponse
3. **Relaxed Validation**: Collaborator `bank_account_no` and `bank_ifsc` now optional

---

## Critical Bug Fixes

1. **UpdateProductVariant Transaction**: Prevents partial updates if price operations fail
2. **Error Message Index**: Fixes incorrect array index display in error messages (uses `strconv.Itoa()`)
3. **Cancel-Items JSON**: Ensures snake_case JSON keys work correctly with Gin binding

---

## Summary

| Category | Count |
|----------|-------|
| Issues Fixed | 8 |
| Additional Bug Fixes | 3 |
| New Files | 1 |
| Breaking Changes | 1 (availability endpoint) |
| Database Migrations | 2 (auto-applied by GORM) |
| Lines of Code Modified | ~500 |

**Impact Assessment**:
- ✅ High: Availability endpoint (requires frontend update)
- ⚠️ Medium: New endpoint (requires frontend integration if used)
- ✅ Low: Price precision, bank validation, logo support (backward compatible)
