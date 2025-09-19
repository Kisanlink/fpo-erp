# Sales Pricing Update

## Overview

The sales system has been updated to automatically calculate selling prices from the `product_prices` table instead of requiring manual price input. This ensures consistent pricing and reduces the chance of pricing errors.

## Changes Made

### 1. **Sales Service Updates**
- Added `priceRepo` dependency to `SalesService`
- Modified `CreateSale` method to automatically fetch selling prices from `product_prices` table
- Added `getSellingPrice` helper method to retrieve current retail prices
- Removed manual selling price validation from request validation

### 2. **Request Model Updates**
- Updated `CreateSaleItemRequest` to remove `SellingPrice` field
- Sales now only require `batch_id` and `quantity` for each item

### 3. **Dependency Injection Updates**
- Updated `SalesService` constructor to include `ProductPriceRepository`
- Updated routes registration to pass `priceRepo` to `SalesService`

## How It Works

### 1. **Price Retrieval Logic**
The system follows this priority order for getting selling prices:

1. **Primary**: Get active "retail" price for the product
2. **Fallback**: If no retail price exists, get any active price for the product
3. **Error**: If no prices exist, return an error

### 2. **Price Types**
The system supports different price types:
- `retail` - Standard retail price (primary choice for sales)
- `wholesale` - Wholesale price
- `bulk` - Bulk purchase price
- `cost` - Cost price (for inventory tracking)

## API Usage

### Before (Old Way)
```json
POST /api/v1/sales
{
  "warehouse_id": "WH_123",
  "customer_id": "CUST_456",
  "items": [
    {
      "batch_id": "BATCH_789",
      "quantity": 5,
      "selling_price": 25.99  // Manual price input required
    }
  ]
}
```

### After (New Way)
```json
POST /api/v1/sales
{
  "warehouse_id": "WH_123",
  "customer_id": "CUST_456",
  "items": [
    {
      "batch_id": "BATCH_789",
      "quantity": 5
      // No selling_price needed - calculated automatically
    }
  ]
}
```

## Setting Up Product Prices

### 1. **Create a Product Price**
```json
POST /api/v1/products/{product_id}/prices
{
  "price_type": "retail",
  "price": 25.99,
  "currency": "USD",
  "effective_from": "2024-01-01T00:00:00Z",
  "is_active": true
}
```

### 2. **Available Price Endpoints**
- `GET /api/v1/products/{id}/prices` - Get all prices for a product
- `GET /api/v1/products/{id}/prices/current?type=retail` - Get current retail price
- `POST /api/v1/products/{id}/prices` - Create new price
- `PATCH /api/v1/prices/{id}` - Update existing price
- `DELETE /api/v1/prices/{id}` - Delete price

## Error Handling

### Common Error Scenarios

1. **No Price Found**
   ```
   Error: "selling price not found for product"
   ```
   **Solution**: Create a price for the product using the price endpoints

2. **No Active Price**
   ```
   Error: "no active prices found for product"
   ```
   **Solution**: Set `is_active: true` for the price or create a new active price

3. **Price Expired**
   ```
   Error: "no active prices found for product"
   ```
   **Solution**: Update the `effective_to` date or create a new price

## Benefits

1. **Consistency**: All sales use the same pricing source
2. **Accuracy**: No manual price entry errors
3. **Flexibility**: Support for different price types and effective dates
4. **Audit Trail**: Price changes are tracked over time
5. **Automation**: Reduced manual work for sales creation

## Migration Notes

### For Existing Systems
1. Ensure all products have at least one active price in the `product_prices` table
2. Update any client applications to remove `selling_price` from sale requests
3. Test the new flow with existing products

### For New Systems
1. Create products first
2. Add prices for each product using the price endpoints
3. Create inventory batches
4. Create sales (prices will be calculated automatically)

## Testing

### Test Scenarios
1. **Valid Sale**: Product with active retail price
2. **Fallback Sale**: Product with no retail price but other active prices
3. **No Price Error**: Product with no prices
4. **Expired Price**: Product with only expired prices

### Example Test Flow
```bash
# 1. Create a product
POST /api/v1/products
{
  "sku": "TEST-001",
  "name": "Test Product",
  "description": "Test product for pricing"
}

# 2. Add a retail price
POST /api/v1/products/{product_id}/prices
{
  "price_type": "retail",
  "price": 29.99,
  "is_active": true
}

# 3. Create inventory batch
POST /api/v1/batches
{
  "warehouse_id": "WH_123",
  "product_id": "{product_id}",
  "cost_price": 20.00,
  "expiry_date": "2024-12-31",
  "quantity": 100
}

# 4. Create sale (price calculated automatically)
POST /api/v1/sales
{
  "warehouse_id": "WH_123",
  "items": [
    {
      "batch_id": "{batch_id}",
      "quantity": 5
    }
  ]
}
```

## Troubleshooting

### Issue: "selling price not found for product"
**Cause**: No active prices exist for the product
**Solution**: 
1. Check if product has any prices: `GET /api/v1/products/{id}/prices`
2. Create a new price if none exist
3. Ensure the price is active (`is_active: true`)

### Issue: "no active prices found for product"
**Cause**: All prices are inactive or expired
**Solution**:
1. Check price status: `GET /api/v1/products/{id}/prices`
2. Update price to active or create new active price
3. Check effective dates if price appears active

### Issue: Wrong price being used
**Cause**: Multiple active prices exist
**Solution**:
1. Review all prices for the product
2. Deactivate old prices or set proper effective dates
3. Ensure only one retail price is active at a time
