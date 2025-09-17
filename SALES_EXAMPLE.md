# Sales System Example - Product-Based with FEFO

## Complete Workflow Example

### Step 1: Create a Product
```bash
POST /api/v1/products
{
  "sku": "RICE-001",
  "name": "Basmati Rice",
  "description": "Premium quality basmati rice"
}
```

### Step 2: Add Product Price
```bash
POST /api/v1/products/{product_id}/prices
{
  "price_type": "retail",
  "price": 45.99,
  "currency": "USD",
  "is_active": true
}
```

### Step 3: Create Warehouse
```bash
POST /api/v1/warehouses
{
  "name": "Main Warehouse",
  "location": "Mumbai, Maharashtra"
}
```

### Step 4: Create Multiple Inventory Batches (Different Expiry Dates)
```bash
# Batch 1 - Expires first (FEFO priority)
POST /api/v1/batches
{
  "warehouse_id": "WH_123",
  "product_id": "PROD_456",
  "cost_price": 35.00,
  "expiry_date": "2024-01-15",
  "quantity": 100
}

# Batch 2 - Expires second
POST /api/v1/batches
{
  "warehouse_id": "WH_123",
  "product_id": "PROD_456",
  "cost_price": 38.00,
  "expiry_date": "2024-02-01",
  "quantity": 150
}

# Batch 3 - Expires last
POST /api/v1/batches
{
  "warehouse_id": "WH_123",
  "product_id": "PROD_456",
  "cost_price": 40.00,
  "expiry_date": "2024-03-01",
  "quantity": 200
}
```

### Step 5: Create Sale (Automatic FEFO + Pricing)
```bash
POST /api/v1/sales
{
  "warehouse_id": "WH_123",
  "customer_id": "CUST_789",
  "items": [
    {
      "product_id": "PROD_456",
      "quantity": 120
    }
  ]
}
```

## What Happens Automatically

### 1. **Price Calculation**
- System gets retail price: ₹45.99 from `product_prices` table
- No manual price input needed

### 2. **Batch Selection (FEFO)**
- System finds all batches for PROD_456 in WH_123
- Orders by expiry date (earliest first)
- Available batches:
  - Batch 1: 100 units (expires 2024-01-15) ← **First priority**
  - Batch 2: 150 units (expires 2024-02-01) ← **Second priority**
  - Batch 3: 200 units (expires 2024-03-01) ← **Last priority**

### 3. **Quantity Allocation**
- Request: 120 units
- Allocation:
  - 100 units from Batch 1 (uses all available)
  - 20 units from Batch 2 (remaining needed)
  - 0 units from Batch 3 (not needed)

### 4. **Sale Items Created**
```json
{
  "sale_id": "SALE_123",
  "items": [
    {
      "batch_id": "BATCH_1",
      "quantity": 100,
      "selling_price": 45.99,
      "line_total": 4599.00
    },
    {
      "batch_id": "BATCH_2", 
      "quantity": 20,
      "selling_price": 45.99,
      "line_total": 919.80
    }
  ],
  "total_amount": 5518.80
}
```

### 5. **Inventory Updates**
- Batch 1: 100 → 0 units (fully consumed)
- Batch 2: 150 → 130 units (20 consumed)
- Batch 3: 200 → 200 units (unchanged)

## Benefits Demonstrated

### ✅ **User-Friendly**
- No need to know batch IDs
- Just specify product and quantity

### ✅ **Automatic FEFO**
- Oldest stock used first
- Prevents waste and spoilage

### ✅ **Multi-Batch Support**
- Can fulfill orders across multiple batches
- Handles complex inventory scenarios

### ✅ **Consistent Pricing**
- All sales use same retail price
- No manual price entry errors

### ✅ **Automatic Calculations**
- Prices from product_prices table
- Line totals calculated automatically

## Error Handling Examples

### Scenario 1: No Inventory
```bash
POST /api/v1/sales
{
  "warehouse_id": "WH_123",
  "items": [
    {
      "product_id": "PROD_999",  # Product with no batches
      "quantity": 10
    }
  ]
}
```
**Response**: `"no inventory available for product in this warehouse"`

### Scenario 2: Insufficient Stock
```bash
POST /api/v1/sales
{
  "warehouse_id": "WH_123",
  "items": [
    {
      "product_id": "PROD_456",
      "quantity": 500  # Only 450 available total
    }
  ]
}
```
**Response**: `"insufficient stock for product"`

### Scenario 3: No Price Set
```bash
POST /api/v1/sales
{
  "warehouse_id": "WH_123",
  "items": [
    {
      "product_id": "PROD_789",  # Product with no prices
      "quantity": 10
    }
  ]
}
```
**Response**: `"selling price not found for product"`

## Real-World Use Case

### Restaurant Order Scenario
```bash
# Restaurant wants to order rice
POST /api/v1/sales
{
  "warehouse_id": "WH_MAIN",
  "customer_id": "REST_001",
  "items": [
    {
      "product_id": "RICE_BASMATI",
      "quantity": 50  # 50kg of rice
    },
    {
      "product_id": "RICE_JASMINE", 
      "quantity": 30  # 30kg of jasmine rice
    }
  ]
}
```

### What Happens:
1. **Rice Basmati**: System finds batches, uses FEFO, calculates price
2. **Rice Jasmine**: System finds batches, uses FEFO, calculates price
3. **Total**: Combined total calculated automatically
4. **Inventory**: Updated across multiple batches as needed

This makes the system much more practical for real-world use!
