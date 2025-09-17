# Sales Pricing Update Summary

## Changes Made

### 1. **Sales Service**
- Added `priceRepo` dependency to automatically fetch prices
- Modified `CreateSale` to get selling prices from `product_prices` table
- Added `getSellingPrice()` method with fallback logic
- Removed manual selling price validation
- **NEW**: Changed from batch-based to product-based sales
- **NEW**: Automatic batch selection using FEFO (First Expired, First Out)

### 2. **Request Model**
- Changed from `batch_id` to `product_id` in `CreateSaleItemRequest`
- Sales now only need `product_id` and `quantity`
- Batch selection is automatic based on expiry date

### 3. **Inventory Repository**
- Added `GetBatchesByProductOrderedByExpiry()` method
- Added `GetBatchesByProductAndWarehouseOrderedByExpiry()` method
- Implements FEFO (First Expired, First Out) logic

### 4. **Price Retrieval Logic**
1. Try to get active "retail" price
2. Fallback to any active price if retail not found
3. Error if no prices exist

### 5. **Batch Selection Logic**
1. Get all batches for the product in the specified warehouse
2. Order by expiry date (earliest first - FEFO)
3. Filter only batches with available stock
4. Allocate quantity across multiple batches if needed

## API Usage

### Before (Batch-based)
```json
{
  "items": [
    {
      "batch_id": "BATCH_123",
      "quantity": 5,
      "selling_price": 25.99  // Manual input
    }
  ]
}
```

### After (Product-based with FEFO)
```json
{
  "items": [
    {
      "product_id": "PROD_123",
      "quantity": 5
      // Batch selected automatically (FEFO)
      // Price calculated automatically
    }
  ]
}
```

## How FEFO Works

### Example Scenario
- Product A has 3 batches:
  - Batch 1: 10 units, expires 2024-01-15
  - Batch 2: 15 units, expires 2024-02-01
  - Batch 3: 20 units, expires 2024-03-01

### Sale Request: 25 units
The system will automatically allocate:
- 10 units from Batch 1 (expires first)
- 15 units from Batch 2 (expires second)
- 0 units from Batch 3 (not needed)

## Product Prices Endpoints (Already Available)

- `GET /api/v1/products/{id}/prices` - Get all prices
- `GET /api/v1/products/{id}/prices/current?type=retail` - Get current price
- `POST /api/v1/products/{id}/prices` - Create price
- `PATCH /api/v1/prices/{id}` - Update price
- `DELETE /api/v1/prices/{id}` - Delete price

## Benefits
- **User-friendly**: No need to know batch IDs
- **Automatic FEFO**: Prevents waste by using oldest stock first
- **Multi-batch support**: Can fulfill orders across multiple batches
- **Consistent pricing**: All sales use the same pricing source
- **No manual price entry**: Reduces errors
- **Automatic price calculation**: From product_prices table

## Error Scenarios

### 1. **No Inventory Available**
```
Error: "no inventory available for product in this warehouse"
```
**Solution**: Check if product has batches in the specified warehouse

### 2. **Insufficient Stock**
```
Error: "insufficient stock for product"
```
**Solution**: Check total available quantity across all batches

### 3. **No Price Found**
```
Error: "selling price not found for product"
```
**Solution**: Create a price for the product using price endpoints
