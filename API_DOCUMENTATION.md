# Kisanlink ERP API Documentation

## Overview

The Kisanlink ERP API provides comprehensive endpoints for managing products, warehouses, inventory, sales, returns, and file attachments. The API follows RESTful principles and uses JSON for data exchange.

**Base URL**: `http://localhost:8080/api/v1`

**Authentication**: JWT Bearer Token (for protected endpoints)

## Authentication

Protected endpoints require a valid JWT token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Response Format

All API responses follow a standard format:

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": { ... },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

Error responses:

```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error message",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## HTTP Status Codes

- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `404` - Not Found
- `429` - Too Many Requests
- `500` - Internal Server Error

---

## 1. Products API

### 1.1 Get All Products

**Endpoint**: `GET /api/v1/products`

**Description**: Retrieve all products in the system

**Authentication**: Not required

**Query Parameters**:
- None

**Response**:
```json
{
  "success": true,
  "message": "Products retrieved successfully",
  "data": [
    {
      "id": "PROD_abc123",
      "sku": "PROD-001",
      "name": "Sample Product",
      "description": "A sample product description",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 1.2 Get Product by ID

**Endpoint**: `GET /api/v1/products/{id}`

**Description**: Retrieve a specific product by its ID

**Authentication**: Not required

**Path Parameters**:
- `id` (string, required) - Product ID

**Response**:
```json
{
  "success": true,
  "message": "Product retrieved successfully",
  "data": {
    "id": "PROD_abc123",
    "sku": "PROD-001",
    "name": "Sample Product",
    "description": "A sample product description",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 1.3 Get Product by SKU

**Endpoint**: `GET /api/v1/products/sku/{sku}`

**Description**: Retrieve a specific product by its SKU

**Authentication**: Not required

**Path Parameters**:
- `sku` (string, required) - Product SKU

**Response**:
```json
{
  "success": true,
  "message": "Product retrieved successfully",
  "data": {
    "id": "PROD_abc123",
    "sku": "PROD-001",
    "name": "Sample Product",
    "description": "A sample product description",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 1.4 Search Products

**Endpoint**: `GET /api/v1/products/search`

**Description**: Search products by name

**Authentication**: Not required

**Query Parameters**:
- `q` (string, required) - Search query

**Response**:
```json
{
  "success": true,
  "message": "Products found",
  "data": [
    {
      "id": "PROD_abc123",
      "sku": "PROD-001",
      "name": "Sample Product",
      "description": "A sample product description",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 1.5 Create Product

**Endpoint**: `POST /api/v1/products`

**Description**: Create a new product

**Authentication**: Required

**Request Body**:
```json
{
  "sku": "PROD-001",
  "name": "Sample Product",
  "description": "A sample product description"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Product created successfully",
  "data": {
    "id": "PROD_abc123",
    "sku": "PROD-001",
    "name": "Sample Product",
    "description": "A sample product description",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 1.6 Update Product

**Endpoint**: `PATCH /api/v1/products/{id}`

**Description**: Update an existing product

**Authentication**: Required

**Path Parameters**:
- `id` (string, required) - Product ID

**Request Body**:
```json
{
  "name": "Updated Product Name",
  "description": "Updated product description"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Product updated successfully",
  "data": {
    "id": "PROD_abc123",
    "sku": "PROD-001",
    "name": "Updated Product Name",
    "description": "Updated product description",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 1.7 Delete Product

**Endpoint**: `DELETE /api/v1/products/{id}`

**Description**: Delete a product

**Authentication**: Required

**Path Parameters**:
- `id` (string, required) - Product ID

**Response**:
```json
{
  "success": true,
  "message": "Product deleted successfully",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z"
}
```

---

## 2. Warehouses API

### 2.1 Get All Warehouses

**Endpoint**: `GET /api/v1/warehouses`

**Description**: Retrieve all warehouses in the system

**Authentication**: Required

**Response**:
```json
{
  "success": true,
  "message": "Warehouses retrieved successfully",
  "data": [
    {
      "id": "WH_abc123",
      "name": "Main Warehouse",
      "address": {
        "id": "ADDR_abc123",
        "type": "WORK",
        "address_line_1": "123 Main St",
        "address_line_2": "Suite 100",
        "city": "New York",
        "state": "NY",
        "postal_code": "10001",
        "country": "USA",
        "full_address": "123 Main St, Suite 100, New York, NY, 10001, USA"
      },
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 2.2 Get Warehouse by ID

**Endpoint**: `GET /api/v1/warehouses/{id}`

**Description**: Retrieve a specific warehouse by its ID

**Authentication**: Required

**Path Parameters**:
- `id` (string, required) - Warehouse ID

**Response**:
```json
{
  "success": true,
  "message": "Warehouse retrieved successfully",
  "data": {
    "id": "WH_abc123",
    "name": "Main Warehouse",
    "address": {
      "id": "ADDR_abc123",
      "type": "WORK",
      "address_line_1": "123 Main St",
      "address_line_2": "Suite 100",
      "city": "New York",
      "state": "NY",
      "postal_code": "10001",
      "country": "USA",
      "full_address": "123 Main St, Suite 100, New York, NY, 10001, USA"
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 2.3 Search Warehouses

**Endpoint**: `GET /api/v1/warehouses/search`

**Description**: Search warehouses by name

**Authentication**: Required

**Query Parameters**:
- `q` (string, required) - Search query

**Response**:
```json
{
  "success": true,
  "message": "Warehouses search completed",
  "data": [
    {
      "id": "WH_abc123",
      "name": "Main Warehouse",
      "address": {
        "id": "ADDR_abc123",
        "type": "WORK",
        "address_line_1": "123 Main St",
        "address_line_2": "Suite 100",
        "city": "New York",
        "state": "NY",
        "postal_code": "10001",
        "country": "USA",
        "full_address": "123 Main St, Suite 100, New York, NY, 10001, USA"
      },
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 2.4 Create Warehouse

**Endpoint**: `POST /api/v1/warehouses`

**Description**: Create a new warehouse with address information

**Authentication**: Required

**Request Body**:
```json
{
  "name": "New Warehouse",
  "address_id": "ADDR_existing123"
}
```

**OR with inline address creation**:
```json
{
  "name": "New Warehouse",
  "address": {
    "type": "WORK",
    "address_line_1": "456 Oak St",
    "address_line_2": "Building A",
    "city": "Los Angeles",
    "state": "CA",
    "postal_code": "90210",
    "country": "USA",
    "is_primary": true
  }
}
```

**Response**:
```json
{
  "success": true,
  "message": "Warehouse created successfully",
  "data": {
    "id": "WH_def456",
    "name": "New Warehouse",
    "address": {
      "id": "ADDR_def456",
      "type": "WORK",
      "address_line_1": "456 Oak St",
      "address_line_2": "Building A",
      "city": "Los Angeles",
      "state": "CA",
      "postal_code": "90210",
      "country": "USA",
      "full_address": "456 Oak St, Building A, Los Angeles, CA, 90210, USA"
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 2.5 Update Warehouse

**Endpoint**: `PATCH /api/v1/warehouses/{id}`

**Description**: Update an existing warehouse

**Authentication**: Required

**Path Parameters**:
- `id` (string, required) - Warehouse ID

**Request Body**:
```json
{
  "name": "Updated Warehouse Name",
  "address_id": "ADDR_new123"
}
```

**OR with inline address update**:
```json
{
  "name": "Updated Warehouse Name",
  "address": {
    "id": "ADDR_existing123",
    "type": "WORK",
    "address_line_1": "789 Pine St",
    "city": "Chicago",
    "state": "IL",
    "postal_code": "60601",
    "country": "USA"
  }
}
```

**Response**:
```json
{
  "success": true,
  "message": "Warehouse updated successfully",
  "data": {
    "id": "WH_abc123",
    "name": "Updated Warehouse Name",
    "address": {
      "id": "ADDR_existing123",
      "type": "WORK",
      "address_line_1": "789 Pine St",
      "city": "Chicago",
      "state": "IL",
      "postal_code": "60601",
      "country": "USA",
      "full_address": "789 Pine St, Chicago, IL, 60601, USA"
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 2.6 Delete Warehouse

**Endpoint**: `DELETE /api/v1/warehouses/{id}`

**Description**: Delete a warehouse (also deletes associated address)

**Authentication**: Required

**Path Parameters**:
- `id` (string, required) - Warehouse ID

**Response**:
```json
{
  "success": true,
  "message": "Warehouse deleted successfully",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z"
}
```

---

## Error Responses

### 400 Bad Request
```json
{
  "success": false,
  "message": "Invalid request data",
  "error": "Field 'name' is required",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 401 Unauthorized
```json
{
  "success": false,
  "message": "Authorization header required",
  "error": "Invalid token",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 404 Not Found
```json
{
  "success": false,
  "message": "Product not found",
  "error": "Product with ID 'PROD_abc123' not found",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 429 Too Many Requests
```json
{
  "success": false,
  "message": "Rate limit exceeded",
  "error": "Rate limit exceeded",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 500 Internal Server Error
```json
{
  "success": false,
  "message": "Failed to create product",
  "error": "Database connection failed",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

---

---

## 3. Inventory API

### 3.1 Create Inventory Batch

**Endpoint**: `POST /api/v1/batches`

**Description**: Create a new inventory batch

**Authentication**: Required

**Request Body**:
```json
{
  "warehouse_id": "WH_abc123",
  "product_id": "PROD_abc123",
  "cost_price": 25.50,
  "expiry_date": "2024-12-31T00:00:00Z",
  "total_quantity": 100
}
```

**Response**:
```json
{
  "success": true,
  "message": "Inventory batch created successfully",
  "data": {
    "id": "BATCH_abc123",
    "warehouse_id": "WH_abc123",
    "product_id": "PROD_abc123",
    "cost_price": 25.50,
    "expiry_date": "2024-12-31T00:00:00Z",
    "total_quantity": 100,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 3.2 Get Batch by ID

**Endpoint**: `GET /api/v1/batches/{id}`

**Description**: Retrieve a specific inventory batch

**Authentication**: Not required

**Path Parameters**:
- `id` (string, required) - Batch ID

**Response**:
```json
{
  "success": true,
  "message": "Batch retrieved successfully",
  "data": {
    "id": "BATCH_abc123",
    "warehouse_id": "WH_abc123",
    "product_id": "PROD_abc123",
    "cost_price": 25.50,
    "expiry_date": "2024-12-31T00:00:00Z",
    "total_quantity": 100,
    "warehouse": {
      "id": "WH_abc123",
      "name": "Main Warehouse",
      "location": "123 Main St"
    },
    "product": {
      "id": "PROD_abc123",
      "sku": "PROD-001",
      "name": "Sample Product"
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 3.3 Get Batches by Warehouse

**Endpoint**: `GET /api/v1/warehouses/{warehouse_id}/batches`

**Description**: Retrieve all batches for a specific warehouse

**Authentication**: Not required

**Path Parameters**:
- `warehouse_id` (string, required) - Warehouse ID

**Response**:
```json
{
  "success": true,
  "message": "Warehouse batches retrieved successfully",
  "data": [
    {
      "id": "BATCH_abc123",
      "warehouse_id": "WH_abc123",
      "product_id": "PROD_abc123",
      "cost_price": 25.50,
      "expiry_date": "2024-12-31T00:00:00Z",
      "total_quantity": 100,
      "product": {
        "id": "PROD_abc123",
        "sku": "PROD-001",
        "name": "Sample Product"
      },
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 3.4 Get Batches by Product

**Endpoint**: `GET /api/v1/products/{product_id}/batches`

**Description**: Retrieve all batches for a specific product

**Authentication**: Not required

**Path Parameters**:
- `product_id` (string, required) - Product ID

**Response**:
```json
{
  "success": true,
  "message": "Product batches retrieved successfully",
  "data": [
    {
      "id": "BATCH_abc123",
      "warehouse_id": "WH_abc123",
      "product_id": "PROD_abc123",
      "cost_price": 25.50,
      "expiry_date": "2024-12-31T00:00:00Z",
      "total_quantity": 100,
      "warehouse": {
        "id": "WH_abc123",
        "name": "Main Warehouse",
        "location": "123 Main St"
      },
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 3.5 Create Inventory Transaction

**Endpoint**: `POST /api/v1/batches/{batch_id}/transactions`

**Description**: Create a new inventory transaction (stock movement)

**Authentication**: Required

**Path Parameters**:
- `batch_id` (string, required) - Batch ID

**Request Body**:
```json
{
  "transaction_type": "SALE",
  "quantity_change": -5,
  "related_entity_id": "SALE_abc123",
  "note": "Sold 5 units"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Inventory transaction created successfully",
  "data": {
    "id": "TRANS_abc123",
    "batch_id": "BATCH_abc123",
    "transaction_type": "SALE",
    "quantity_change": -5,
    "related_entity_id": "SALE_abc123",
    "performed_by": "USER_abc123",
    "note": "Sold 5 units",
    "occurred_at": "2024-01-01T00:00:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 3.6 Get Batch Transactions

**Endpoint**: `GET /api/v1/batches/{batch_id}/transactions`

**Description**: Retrieve all transactions for a specific batch

**Authentication**: Not required

**Path Parameters**:
- `batch_id` (string, required) - Batch ID

**Response**:
```json
{
  "success": true,
  "message": "Batch transactions retrieved successfully",
  "data": [
    {
      "id": "TRANS_abc123",
      "batch_id": "BATCH_abc123",
      "transaction_type": "SALE",
      "quantity_change": -5,
      "related_entity_id": "SALE_abc123",
      "performed_by": "USER_abc123",
      "note": "Sold 5 units",
      "occurred_at": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 3.7 Get Expiring Batches

**Endpoint**: `GET /api/v1/batches/expiring`

**Description**: Retrieve batches that expire within a specified number of days

**Authentication**: Not required

**Query Parameters**:
- `days` (integer, optional) - Number of days (default: 30)

**Response**:
```json
{
  "success": true,
  "message": "Expiring batches retrieved successfully",
  "data": [
    {
      "id": "BATCH_abc123",
      "warehouse_id": "WH_abc123",
      "product_id": "PROD_abc123",
      "cost_price": 25.50,
      "expiry_date": "2024-01-15T00:00:00Z",
      "total_quantity": 50,
      "warehouse": {
        "id": "WH_abc123",
        "name": "Main Warehouse"
      },
      "product": {
        "id": "PROD_abc123",
        "sku": "PROD-001",
        "name": "Sample Product"
      },
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 3.8 Get Low Stock Batches

**Endpoint**: `GET /api/v1/batches/low-stock`

**Description**: Retrieve batches with stock below a specified threshold

**Authentication**: Not required

**Query Parameters**:
- `threshold` (integer, optional) - Stock threshold (default: 10)

**Response**:
```json
{
  "success": true,
  "message": "Low stock batches retrieved successfully",
  "data": [
    {
      "id": "BATCH_abc123",
      "warehouse_id": "WH_abc123",
      "product_id": "PROD_abc123",
      "cost_price": 25.50,
      "expiry_date": "2024-12-31T00:00:00Z",
      "total_quantity": 5,
      "warehouse": {
        "id": "WH_abc123",
        "name": "Main Warehouse"
      },
      "product": {
        "id": "PROD_abc123",
        "sku": "PROD-001",
        "name": "Sample Product"
      },
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 3.9 Get All Products Availability

**Endpoint**: `GET /api/v1/products/availability`

**Description**: Retrieve all available products across all warehouses with detailed information

**Authentication**: Required

**Query Parameters**:
- None

**Response**:
```json
{
  "success": true,
  "message": "Products availability retrieved successfully",
  "data": [
    {
      "id": "BATCH_abc123",
      "warehouse_id": "WH_abc123",
      "warehouse_name": "Main Warehouse",
      "warehouse_location": "123 Main St, Mumbai, Maharashtra",
      "product_id": "PROD_abc123",
      "product_sku": "PROD-001",
      "product_name": "Sample Product",
      "product_description": "A sample product description",
      "cost_price": 25.50,
      "expiry_date": "2024-12-31",
      "total_quantity": 100,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": "BATCH_def456",
      "warehouse_id": "WH_def456",
      "warehouse_name": "Secondary Warehouse",
      "warehouse_location": "456 Oak Ave, Delhi, Delhi",
      "product_id": "PROD_def456",
      "product_sku": "PROD-002",
      "product_name": "Another Product",
      "product_description": "Another product description",
      "cost_price": 30.00,
      "expiry_date": "2024-11-30",
      "total_quantity": 75,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

**Notes**:
- This endpoint returns all inventory batches with complete product and warehouse information
- Each record represents a batch of a specific product in a specific warehouse
- The response includes warehouse details (name, location) and product details (SKU, name, description)
- Cost price, expiry date, and available quantity are included for each batch
- Authentication is required to access this endpoint

---

## 4. Product Pricing API

### 4.1 Create Product Price

**Endpoint**: `POST /api/v1/products/{product_id}/prices`

**Description**: Create a new price for a product

**Authentication**: Required

**Path Parameters**:
- `product_id` (string, required) - Product ID

**Request Body**:
```json
{
  "price_type": "retail",
  "price": 29.99,
  "currency": "USD",
  "effective_from": "2024-01-01T00:00:00Z",
  "effective_to": "2024-12-31T00:00:00Z",
  "is_active": true
}
```

**Response**:
```json
{
  "success": true,
  "message": "Product price created successfully",
  "data": {
    "id": "PRICE_abc123",
    "product_id": "PROD_abc123",
    "price_type": "retail",
    "price": 29.99,
    "currency": "USD",
    "effective_from": "2024-01-01T00:00:00Z",
    "effective_to": "2024-12-31T00:00:00Z",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 4.2 Get Product Prices

**Endpoint**: `GET /api/v1/products/{product_id}/prices`

**Description**: Retrieve all prices for a specific product

**Authentication**: Not required

**Path Parameters**:
- `product_id` (string, required) - Product ID

**Response**:
```json
{
  "success": true,
  "message": "Product prices retrieved successfully",
  "data": [
    {
      "id": "PRICE_abc123",
      "product_id": "PROD_abc123",
      "price_type": "retail",
      "price": 29.99,
      "currency": "USD",
      "effective_from": "2024-01-01T00:00:00Z",
      "effective_to": "2024-12-31T00:00:00Z",
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 4.3 Get Product with Prices

**Endpoint**: `GET /api/v1/products/{product_id}/with-prices`

**Description**: Retrieve a product with all its prices

**Authentication**: Not required

**Path Parameters**:
- `product_id` (string, required) - Product ID

**Response**:
```json
{
  "success": true,
  "message": "Product with prices retrieved successfully",
  "data": {
    "id": "PROD_abc123",
    "sku": "PROD-001",
    "name": "Sample Product",
    "description": "A sample product description",
    "prices": [
      {
        "id": "PRICE_abc123",
        "product_id": "PROD_abc123",
        "price_type": "retail",
        "price": 29.99,
        "currency": "USD",
        "effective_from": "2024-01-01T00:00:00Z",
        "effective_to": "2024-12-31T00:00:00Z",
        "is_active": true,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 4.4 Update Product Price

**Endpoint**: `PATCH /api/v1/prices/{price_id}`

**Description**: Update an existing product price

**Authentication**: Required

**Path Parameters**:
- `price_id` (string, required) - Price ID

**Request Body**:
```json
{
  "price": 34.99,
  "effective_to": "2024-12-31T00:00:00Z",
  "is_active": false
}
```

**Response**:
```json
{
  "success": true,
  "message": "Product price updated successfully",
  "data": {
    "id": "PRICE_abc123",
    "product_id": "PROD_abc123",
    "price_type": "retail",
    "price": 34.99,
    "currency": "USD",
    "effective_from": "2024-01-01T00:00:00Z",
    "effective_to": "2024-12-31T00:00:00Z",
    "is_active": false,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 4.5 Delete Product Price

**Endpoint**: `DELETE /api/v1/prices/{price_id}`

**Description**: Delete a product price

**Authentication**: Required

**Path Parameters**:
- `price_id` (string, required) - Price ID

**Response**:
```json
{
  "success": true,
  "message": "Product price deleted successfully",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z"
}
```

---

## 5. Sales API

### 5.1 Create Sale

**Endpoint**: `POST /api/v1/sales`

**Description**: Create a new sale transaction with optional discount application

**Authentication**: Required

**Notes**:
- The system automatically selects inventory batches using FEFO (First Expired, First Out) logic based on `product_id`
- Selling prices are automatically fetched from the `product_prices` table
- If `discount_id` is provided, the discount will be validated and applied automatically
- The `final_amount` will be calculated as `total_amount - discount_applied`

**Request Body**:
```json
{
  "warehouse_id": "WH_abc123",
  "customer_id": "CUST_abc123",
  "sale_date": "2024-01-01T00:00:00Z",
  "discount_id": "DISC_abc123",
  "items": [
    {
      "product_id": "PROD_abc123",
      "quantity": 5
    }
  ]
}
```

**Response**:
```json
{
  "success": true,
  "message": "Sale created successfully",
  "data": {
    "id": "SALE_abc123",
    "warehouse_id": "WH_abc123",
    "customer_id": "CUST_abc123",
    "sale_date": "2024-01-01T00:00:00Z",
    "total_amount": 299.99,
    "final_amount": 249.99,
    "discount_applied": 50.00,
    "status": "completed",
    "warehouse": {
      "id": "WH_abc123",
      "name": "Main Warehouse"
    },
    "items": [
      {
        "id": "SITEM_abc123",
        "sale_id": "SALE_abc123",
        "batch_id": "BATCH_abc123",
        "product_id": "PROD_abc123",
        "quantity": 5,
        "selling_price": 29.99,
        "line_total": 149.95
      }
    ],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 5.2 Get All Sales

**Endpoint**: `GET /api/v1/sales`

**Description**: Retrieve all sales transactions

**Authentication**: Not required

**Query Parameters**:
- `warehouse_id` (string, optional) - Filter by warehouse
- `status` (string, optional) - Filter by status
- `start_date` (string, optional) - Filter by start date (ISO format)
- `end_date` (string, optional) - Filter by end date (ISO format)

**Response**:
```json
{
  "success": true,
  "message": "Sales retrieved successfully",
  "data": [
    {
      "id": "SALE_abc123",
      "warehouse_id": "WH_abc123",
      "customer_id": "CUST_abc123",
      "sale_date": "2024-01-01T00:00:00Z",
      "total_amount": 299.99,
      "final_amount": 249.99,
      "discount_applied": 50.00,
      "status": "completed",
      "warehouse": {
        "id": "WH_abc123",
        "name": "Main Warehouse"
      },
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 5.3 Get Sale by ID

**Endpoint**: `GET /api/v1/sales/{id}`

**Description**: Retrieve a specific sale by its ID

**Authentication**: Not required

**Path Parameters**:
- `id` (string, required) - Sale ID

**Response**:
```json
{
  "success": true,
  "message": "Sale retrieved successfully",
  "data": {
    "id": "SALE_abc123",
    "warehouse_id": "WH_abc123",
    "customer_id": "CUST_abc123",
    "sale_date": "2024-01-01T00:00:00Z",
    "total_amount": 299.99,
    "final_amount": 249.99,
    "discount_applied": 50.00,
    "status": "completed",
    "warehouse": {
      "id": "WH_abc123",
      "name": "Main Warehouse"
    },
    "items": [
      {
        "id": "SITEM_abc123",
        "sale_id": "SALE_abc123",
        "batch_id": "BATCH_abc123",
        "product_id": "PROD_abc123",
        "quantity": 5,
        "selling_price": 29.99,
        "line_total": 149.95,
        "batch": {
          "id": "BATCH_abc123",
          "product": {
            "id": "PROD_abc123",
            "sku": "PROD-001",
            "name": "Sample Product"
          }
        }
      }
    ],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 5.4 Update Sale Status

**Endpoint**: `PATCH /api/v1/sales/{id}/status`

**Description**: Update the status of a sale

**Authentication**: Required

**Path Parameters**:
- `id` (string, required) - Sale ID

**Request Body**:
```json
{
  "status": "cancelled"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Sale status updated successfully",
  "data": {
    "id": "SALE_abc123",
    "status": "cancelled",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 5.5 Get Sales Summary

**Endpoint**: `GET /api/v1/sales/summary`

**Description**: Retrieve sales summary data

**Authentication**: Not required

**Query Parameters**:
- `warehouse_id` (string, optional) - Filter by warehouse
- `start_date` (string, optional) - Start date (ISO format)
- `end_date` (string, optional) - End date (ISO format)

**Response**:
```json
{
  "success": true,
  "message": "Sales summary retrieved successfully",
  "data": {
    "total_sales": 15000.00,
    "total_transactions": 150,
    "average_sale": 100.00,
    "top_products": [
      {
        "product_id": "PROD_abc123",
        "product_name": "Sample Product",
        "total_quantity": 500,
        "total_revenue": 5000.00
      }
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

---

## 6. Returns API

### 6.1 Create Return

**Endpoint**: `POST /api/v1/returns`

**Description**: Create a new return transaction

**Authentication**: Required

**Request Body**:
```json
{
  "sale_id": "SALE_abc123",
  "return_date": "2024-01-01T00:00:00Z",
  "items": [
    {
      "batch_id": "BATCH_abc123",
      "quantity": 2,
      "refund_amount": 25.00
    }
  ]
}
```

**Response**:
```json
{
  "success": true,
  "message": "Return created successfully",
  "data": {
    "id": "RET_abc123",
    "sale_id": "SALE_abc123",
    "return_date": "2024-01-01T00:00:00Z",
    "total_refund": 149.95,
    "status": "pending",
    "sale": {
      "id": "SALE_abc123",
      "total_amount": 299.99
    },
    "items": [
      {
        "id": "RITEM_abc123",
        "return_id": "RET_abc123",
        "batch_id": "BATCH_abc123",
        "quantity": 2,
        "refund_amount": 59.98
      }
    ],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 6.2 Get All Returns

**Endpoint**: `GET /api/v1/returns`

**Description**: Retrieve all return transactions

**Authentication**: Not required

**Query Parameters**:
- `sale_id` (string, optional) - Filter by sale ID
- `status` (string, optional) - Filter by status
- `start_date` (string, optional) - Filter by start date (ISO format)
- `end_date` (string, optional) - Filter by end date (ISO format)

**Response**:
```json
{
  "success": true,
  "message": "Returns retrieved successfully",
  "data": [
    {
      "id": "RET_abc123",
      "sale_id": "SALE_abc123",
      "return_date": "2024-01-01T00:00:00Z",
      "total_refund": 149.95,
      "status": "pending",
      "sale": {
        "id": "SALE_abc123",
        "total_amount": 299.99
      },
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 6.3 Get Return by ID

**Endpoint**: `GET /api/v1/returns/{id}`

**Description**: Retrieve a specific return by its ID

**Authentication**: Not required

**Path Parameters**:
- `id` (string, required) - Return ID

**Response**:
```json
{
  "success": true,
  "message": "Return retrieved successfully",
  "data": {
    "id": "RET_abc123",
    "sale_id": "SALE_abc123",
    "return_date": "2024-01-01T00:00:00Z",
    "total_refund": 149.95,
    "status": "pending",
    "sale": {
      "id": "SALE_abc123",
      "total_amount": 299.99
    },
    "items": [
      {
        "id": "RITEM_abc123",
        "return_id": "RET_abc123",
        "batch_id": "BATCH_abc123",
        "quantity": 2,
        "refund_amount": 59.98,
        "batch": {
          "id": "BATCH_abc123",
          "product": {
            "id": "PROD_abc123",
            "sku": "PROD-001",
            "name": "Sample Product"
          }
        }
      }
    ],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 6.4 Update Return Status

**Endpoint**: `PATCH /api/v1/returns/{id}/status`

**Description**: Update the status of a return

**Authentication**: Required

**Path Parameters**:
- `id` (string, required) - Return ID

**Request Body**:
```json
{
  "status": "approved"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Return status updated successfully",
  "data": {
    "id": "RET_abc123",
    "status": "approved",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 6.5 Get Returns by Sale

**Endpoint**: `GET /api/v1/sales/{sale_id}/returns`

**Description**: Retrieve all returns for a specific sale

**Authentication**: Not required

**Path Parameters**:
- `sale_id` (string, required) - Sale ID

**Response**:
```json
{
  "success": true,
  "message": "Sale returns retrieved successfully",
  "data": [
    {
      "id": "RET_abc123",
      "sale_id": "SALE_abc123",
      "return_date": "2024-01-01T00:00:00Z",
      "total_refund": 149.95,
      "status": "pending",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

---

## 7. Attachments API

### 7.1 Upload Attachment

**Endpoint**: `POST /api/v1/attachments`

**Description**: Upload a file attachment for a sale or return

**Authentication**: Required

**Request Body** (multipart/form-data):
- `file` (file, required) - File to upload
- `sale_id` (string, optional) - Associated sale ID
- `return_id` (string, optional) - Associated return ID
- `file_type` (string, required) - Type of file (e.g., "invoice", "receipt", "document")

**Response**:
```json
{
  "success": true,
  "message": "Attachment uploaded successfully",
  "data": {
    "id": "ATCH_abc123",
    "sale_id": "SALE_abc123",
    "return_id": null,
    "file_path": "uploads/sales/SALE_abc123/invoice.pdf",
    "file_type": "invoice",
    "uploaded_by": "USER_abc123",
    "uploaded_at": "2024-01-01T00:00:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 7.2 Get Attachments by Sale

**Endpoint**: `GET /api/v1/sales/{sale_id}/attachments`

**Description**: Retrieve all attachments for a specific sale

**Authentication**: Not required

**Path Parameters**:
- `sale_id` (string, required) - Sale ID

**Response**:
```json
{
  "success": true,
  "message": "Sale attachments retrieved successfully",
  "data": [
    {
      "id": "ATCH_abc123",
      "sale_id": "SALE_abc123",
      "return_id": null,
      "file_path": "uploads/sales/SALE_abc123/invoice.pdf",
      "file_type": "invoice",
      "uploaded_by": "USER_abc123",
      "uploaded_at": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 7.3 Get Attachments by Return

**Endpoint**: `GET /api/v1/returns/{return_id}/attachments`

**Description**: Retrieve all attachments for a specific return

**Authentication**: Not required

**Path Parameters**:
- `return_id` (string, required) - Return ID

**Response**:
```json
{
  "success": true,
  "message": "Return attachments retrieved successfully",
  "data": [
    {
      "id": "ATCH_abc123",
      "sale_id": null,
      "return_id": "RET_abc123",
      "file_path": "uploads/returns/RET_abc123/receipt.pdf",
      "file_type": "receipt",
      "uploaded_by": "USER_abc123",
      "uploaded_at": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 7.4 Download Attachment

**Endpoint**: `GET /api/v1/attachments/{id}/download`

**Description**: Download a specific attachment file

**Authentication**: Not required

**Path Parameters**:
- `id` (string, required) - Attachment ID

**Response**: File download (binary data)

### 7.5 Delete Attachment

**Endpoint**: `DELETE /api/v1/attachments/{id}`

**Description**: Delete a specific attachment

**Authentication**: Required

**Path Parameters**:
- `id` (string, required) - Attachment ID

**Response**:
```json
{
  "success": true,
  "message": "Attachment deleted successfully",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z"
}
```

---

## 8. Users API

### 8.1 Create User

**Endpoint**: `POST /api/v1/users`

**Description**: Create a new user account

**Authentication**: Required (Admin only)

**Request Body**:
```json
{
  "username": "john_doe",
  "full_name": "John Doe",
  "email": "john.doe@example.com",
  "role": "manager",
  "password": "securepassword123"
}
```

**Response**:
```json
{
  "success": true,
  "message": "User created successfully",
  "data": {
    "id": "USER_abc123",
    "username": "john_doe",
    "full_name": "John Doe",
    "email": "john.doe@example.com",
    "role": "manager",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 8.2 Get All Users

**Endpoint**: `GET /api/v1/users`

**Description**: Retrieve all users in the system

**Authentication**: Required (Admin only)

**Response**:
```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": "USER_abc123",
      "username": "john_doe",
      "full_name": "John Doe",
      "email": "john.doe@example.com",
      "role": "manager",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 8.3 Get User by ID

**Endpoint**: `GET /api/v1/users/{id}`

**Description**: Retrieve a specific user by ID

**Authentication**: Required (Admin or self)

**Path Parameters**:
- `id` (string, required) - User ID

**Response**:
```json
{
  "success": true,
  "message": "User retrieved successfully",
  "data": {
    "id": "USER_abc123",
    "username": "john_doe",
    "full_name": "John Doe",
    "email": "john.doe@example.com",
    "role": "manager",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 8.4 Update User

**Endpoint**: `PATCH /api/v1/users/{id}`

**Description**: Update user information

**Authentication**: Required (Admin or self)

**Path Parameters**:
- `id` (string, required) - User ID

**Request Body**:
```json
{
  "full_name": "John Smith",
  "email": "john.smith@example.com",
  "role": "admin"
}
```

**Response**:
```json
{
  "success": true,
  "message": "User updated successfully",
  "data": {
    "id": "USER_abc123",
    "username": "john_doe",
    "full_name": "John Smith",
    "email": "john.smith@example.com",
    "role": "admin",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 8.5 Delete User

**Endpoint**: `DELETE /api/v1/users/{id}`

**Description**: Delete a user account

**Authentication**: Required (Admin only)

**Path Parameters**:
- `id` (string, required) - User ID

**Response**:
```json
{
  "success": true,
  "message": "User deleted successfully",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z"
}
```

---

## 9. Refund Policies API

### 9.1 Create Refund Policy

**Endpoint**: `POST /api/v1/refund-policies`

**Description**: Create a new refund policy

**Authentication**: Required (Admin only)

**Request Body**:
```json
{
  "policy_name": "Standard Return Policy",
  "description": "Standard 30-day return policy",
  "max_days": 30,
  "restocking_fee": 5.00
}
```

**Response**:
```json
{
  "success": true,
  "message": "Refund policy created successfully",
  "data": {
    "id": "RPOL_abc123",
    "policy_name": "Standard Return Policy",
    "description": "Standard 30-day return policy",
    "max_days": 30,
    "restocking_fee": 5.00,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 9.2 Get All Refund Policies

**Endpoint**: `GET /api/v1/refund-policies`

**Description**: Retrieve all refund policies

**Authentication**: Not required

**Response**:
```json
{
  "success": true,
  "message": "Refund policies retrieved successfully",
  "data": [
    {
      "id": "RPOL_abc123",
      "policy_name": "Standard Return Policy",
      "description": "Standard 30-day return policy",
      "max_days": 30,
      "restocking_fee": 5.00,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 9.3 Get Refund Policy by ID

**Endpoint**: `GET /api/v1/refund-policies/{id}`

**Description**: Retrieve a specific refund policy

**Authentication**: Not required

**Path Parameters**:
- `id` (string, required) - Policy ID

**Response**:
```json
{
  "success": true,
  "message": "Refund policy retrieved successfully",
  "data": {
    "id": "RPOL_abc123",
    "policy_name": "Standard Return Policy",
    "description": "Standard 30-day return policy",
    "max_days": 30,
    "restocking_fee": 5.00,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 9.4 Update Refund Policy

**Endpoint**: `PATCH /api/v1/refund-policies/{id}`

**Description**: Update a refund policy

**Authentication**: Required (Admin only)

**Path Parameters**:
- `id` (string, required) - Policy ID

**Request Body**:
```json
{
  "max_days": 45,
  "restocking_fee": 10.00
}
```

**Response**:
```json
{
  "success": true,
  "message": "Refund policy updated successfully",
  "data": {
    "id": "RPOL_abc123",
    "policy_name": "Standard Return Policy",
    "description": "Standard 30-day return policy",
    "max_days": 45,
    "restocking_fee": 10.00,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 9.5 Delete Refund Policy

**Endpoint**: `DELETE /api/v1/refund-policies/{id}`

**Description**: Delete a refund policy

**Authentication**: Required (Admin only)

**Path Parameters**:
- `id` (string, required) - Policy ID

**Response**:
```json
{
  "success": true,
  "message": "Refund policy deleted successfully",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z"
}
```

---

## 10. Bank Payments API

### 10.1 Create Bank Payment

**Endpoint**: `POST /api/v1/bank-payments`

**Description**: Create a new bank payment record

**Authentication**: Required

**Request Body**:
```json
{
  "sale_id": "SALE_abc123",
  "return_id": null,
  "payment_method": "bank_transfer",
  "transaction_ref": "TXN123456789",
  "amount": 299.99,
  "paid_at": "2024-01-01T00:00:00Z"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Bank payment created successfully",
  "data": {
    "id": "PAY_abc123",
    "sale_id": "SALE_abc123",
    "return_id": null,
    "payment_method": "bank_transfer",
    "transaction_ref": "TXN123456789",
    "amount": 299.99,
    "paid_at": "2024-01-01T00:00:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 10.2 Get Bank Payments by Sale

**Endpoint**: `GET /api/v1/sales/{sale_id}/payments`

**Description**: Retrieve all bank payments for a specific sale

**Authentication**: Not required

**Path Parameters**:
- `sale_id` (string, required) - Sale ID

**Response**:
```json
{
  "success": true,
  "message": "Sale payments retrieved successfully",
  "data": [
    {
      "id": "PAY_abc123",
      "sale_id": "SALE_abc123",
      "return_id": null,
      "payment_method": "bank_transfer",
      "transaction_ref": "TXN123456789",
      "amount": 299.99,
      "paid_at": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 10.3 Get Bank Payments by Return

**Endpoint**: `GET /api/v1/returns/{return_id}/payments`

**Description**: Retrieve all bank payments for a specific return

**Authentication**: Not required

**Path Parameters**:
- `return_id` (string, required) - Return ID

**Response**:
```json
{
  "success": true,
  "message": "Return payments retrieved successfully",
  "data": [
    {
      "id": "PAY_abc123",
      "sale_id": null,
      "return_id": "RET_abc123",
      "payment_method": "bank_transfer",
      "transaction_ref": "TXN123456789",
      "amount": 149.95,
      "paid_at": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

---

## 11. Discounts API

### 11.1 Create Discount

**Endpoint**: `POST /api/v1/discounts`

**Description**: Create a new discount offer

**Authentication**: Required

**Request Body**:
```json
{
  "code": "FLAT50",
  "name": "Flat ₹50 Off",
  "description": "Get ₹50 off on orders above ₹500",
  "discount_type": "flat",
  "value": 50.00,
  "max_discount_amount": null,
  "min_order_value": 500.00,
  "max_order_value": null,
  "applicable_products": null,
  "excluded_products": null,
  "applicable_categories": null,
  "excluded_categories": null,
  "applicable_warehouses": null,
  "customer_groups": null,
  "usage_limit": 1000,
  "usage_per_customer": 1,
  "valid_from": "2024-01-01T00:00:00Z",
  "valid_until": "2024-12-31T23:59:59Z",
  "is_active": true,
  "is_stackable": false,
  "priority": 0,
  "terms": "Terms and conditions apply"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Discount created successfully",
  "data": {
    "id": "DISC_abc123",
    "code": "FLAT50",
    "name": "Flat ₹50 Off",
    "description": "Get ₹50 off on orders above ₹500",
    "discount_type": "flat",
    "value": 50.00,
    "max_discount_amount": null,
    "min_order_value": 500.00,
    "max_order_value": null,
    "applicable_products": null,
    "excluded_products": null,
    "applicable_categories": null,
    "excluded_categories": null,
    "applicable_warehouses": null,
    "customer_groups": null,
    "usage_limit": 1000,
    "usage_per_customer": 1,
    "current_usage": 0,
    "valid_from": "2024-01-01T00:00:00Z",
    "valid_until": "2024-12-31T23:59:59Z",
    "is_active": true,
    "is_stackable": false,
    "priority": 0,
    "terms": "Terms and conditions apply",
    "status": "active",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 11.2 Get All Discounts

**Endpoint**: `GET /api/v1/discounts`

**Description**: Retrieve all discounts with pagination

**Authentication**: Required

**Query Parameters**:
- `limit` (integer, optional) - Number of items per page (default: 10)
- `offset` (integer, optional) - Number of items to skip (default: 0)

**Response**:
```json
{
  "success": true,
  "message": "Discounts retrieved successfully",
  "data": [
    {
      "id": "DISC_abc123",
      "code": "FLAT50",
      "name": "Flat ₹50 Off",
      "description": "Get ₹50 off on orders above ₹500",
      "discount_type": "flat",
      "value": 50.00,
      "max_discount_amount": null,
      "min_order_value": 500.00,
      "max_order_value": null,
      "applicable_products": null,
      "excluded_products": null,
      "applicable_categories": null,
      "excluded_categories": null,
      "applicable_warehouses": null,
      "customer_groups": null,
      "usage_limit": 1000,
      "usage_per_customer": 1,
      "current_usage": 0,
      "valid_from": "2024-01-01T00:00:00Z",
      "valid_until": "2024-12-31T23:59:59Z",
      "is_active": true,
      "is_stackable": false,
      "priority": 0,
      "terms": "Terms and conditions apply",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 11.3 Get Discount by ID

**Endpoint**: `GET /api/v1/discounts/{id}`

**Description**: Retrieve a specific discount by its ID

**Authentication**: Required

**Path Parameters**:
- `id` (string, required) - Discount ID

**Response**:
```json
{
  "success": true,
  "message": "Discount retrieved successfully",
  "data": {
    "id": "DISC_abc123",
    "code": "FLAT50",
    "name": "Flat ₹50 Off",
    "description": "Get ₹50 off on orders above ₹500",
    "discount_type": "flat",
    "value": 50.00,
    "max_discount_amount": null,
    "min_order_value": 500.00,
    "max_order_value": null,
    "applicable_products": null,
    "excluded_products": null,
    "applicable_categories": null,
    "excluded_categories": null,
    "applicable_warehouses": null,
    "customer_groups": null,
    "usage_limit": 1000,
    "usage_per_customer": 1,
    "current_usage": 0,
    "valid_from": "2024-01-01T00:00:00Z",
    "valid_until": "2024-12-31T23:59:59Z",
    "is_active": true,
    "is_stackable": false,
    "priority": 0,
    "terms": "Terms and conditions apply",
    "status": "active",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 11.4 Update Discount

**Endpoint**: `PUT /api/v1/discounts/{id}`

**Description**: Update an existing discount

**Authentication**: Required

**Path Parameters**:
- `id` (string, required) - Discount ID

**Request Body**:
```json
{
  "name": "Updated Flat ₹50 Off",
  "description": "Updated description",
  "value": 60.00,
  "min_order_value": 600.00,
  "usage_limit": 1500,
  "is_active": true
}
```

**Response**:
```json
{
  "success": true,
  "message": "Discount updated successfully",
  "data": {
    "id": "DISC_abc123",
    "code": "FLAT50",
    "name": "Updated Flat ₹50 Off",
    "description": "Updated description",
    "discount_type": "flat",
    "value": 60.00,
    "max_discount_amount": null,
    "min_order_value": 600.00,
    "max_order_value": null,
    "applicable_products": null,
    "excluded_products": null,
    "applicable_categories": null,
    "excluded_categories": null,
    "applicable_warehouses": null,
    "customer_groups": null,
    "usage_limit": 1500,
    "usage_per_customer": 1,
    "current_usage": 0,
    "valid_from": "2024-01-01T00:00:00Z",
    "valid_until": "2024-12-31T23:59:59Z",
    "is_active": true,
    "is_stackable": false,
    "priority": 0,
    "terms": "Terms and conditions apply",
    "status": "active",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 11.5 Delete Discount

**Endpoint**: `DELETE /api/v1/discounts/{id}`

**Description**: Delete a discount

**Authentication**: Required

**Path Parameters**:
- `id` (string, required) - Discount ID

**Response**:
```json
{
  "success": true,
  "message": "Discount deleted successfully",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 11.6 Get Active Discounts

**Endpoint**: `GET /api/v1/discounts/active`

**Description**: Retrieve all currently active discounts

**Authentication**: Required

**Response**:
```json
{
  "success": true,
  "message": "Active discounts retrieved successfully",
  "data": [
    {
      "id": "DISC_abc123",
      "code": "FLAT50",
      "name": "Flat ₹50 Off",
      "description": "Get ₹50 off on orders above ₹500",
      "discount_type": "flat",
      "value": 50.00,
      "max_discount_amount": null,
      "min_order_value": 500.00,
      "max_order_value": null,
      "applicable_products": null,
      "excluded_products": null,
      "applicable_categories": null,
      "excluded_categories": null,
      "applicable_warehouses": null,
      "customer_groups": null,
      "usage_limit": 1000,
      "usage_per_customer": 1,
      "current_usage": 0,
      "valid_from": "2024-01-01T00:00:00Z",
      "valid_until": "2024-12-31T23:59:59Z",
      "is_active": true,
      "is_stackable": false,
      "priority": 0,
      "terms": "Terms and conditions apply",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 11.7 Get Discounts by Type

**Endpoint**: `GET /api/v1/discounts/type/{type}`

**Description**: Retrieve discounts by type

**Authentication**: Required

**Path Parameters**:
- `type` (string, required) - Discount type (flat, percentage, buy_x_get_y, first_order, loyalty, seasonal, bulk, referral)

**Response**:
```json
{
  "success": true,
  "message": "Discounts retrieved successfully",
  "data": [
    {
      "id": "DISC_abc123",
      "code": "FLAT50",
      "name": "Flat ₹50 Off",
      "description": "Get ₹50 off on orders above ₹500",
      "discount_type": "flat",
      "value": 50.00,
      "max_discount_amount": null,
      "min_order_value": 500.00,
      "max_order_value": null,
      "applicable_products": null,
      "excluded_products": null,
      "applicable_categories": null,
      "excluded_categories": null,
      "applicable_warehouses": null,
      "customer_groups": null,
      "usage_limit": 1000,
      "usage_per_customer": 1,
      "current_usage": 0,
      "valid_from": "2024-01-01T00:00:00Z",
      "valid_until": "2024-12-31T23:59:59Z",
      "is_active": true,
      "is_stackable": false,
      "priority": 0,
      "terms": "Terms and conditions apply",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 11.8 Get Discounts by Status

**Endpoint**: `GET /api/v1/discounts/status/{status}`

**Description**: Retrieve discounts by status

**Authentication**: Required

**Path Parameters**:
- `status` (string, required) - Discount status (active, expired, scheduled, inactive, usage_limit_reached)

**Response**:
```json
{
  "success": true,
  "message": "Discounts retrieved successfully",
  "data": [
    {
      "id": "DISC_abc123",
      "code": "FLAT50",
      "name": "Flat ₹50 Off",
      "description": "Get ₹50 off on orders above ₹500",
      "discount_type": "flat",
      "value": 50.00,
      "max_discount_amount": null,
      "min_order_value": 500.00,
      "max_order_value": null,
      "applicable_products": null,
      "excluded_products": null,
      "applicable_categories": null,
      "excluded_categories": null,
      "applicable_warehouses": null,
      "customer_groups": null,
      "usage_limit": 1000,
      "usage_per_customer": 1,
      "current_usage": 0,
      "valid_from": "2024-01-01T00:00:00Z",
      "valid_until": "2024-12-31T23:59:59Z",
      "is_active": true,
      "is_stackable": false,
      "priority": 0,
      "terms": "Terms and conditions apply",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 11.9 Validate Discount

**Endpoint**: `POST /api/v1/discounts/validate`

**Description**: Validate a discount for a given order

**Authentication**: Required

**Request Body**:
```json
{
  "discount_code": "FLAT50",
  "customer_id": "CUST_abc123",
  "order_value": 750.00,
  "product_ids": ["PROD_abc123", "PROD_def456"],
  "warehouse_id": "WH_abc123"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Discount validation completed",
  "data": {
    "is_valid": true,
    "discount_id": "DISC_abc123",
    "discount_code": "FLAT50",
    "discount_name": "Flat ₹50 Off",
    "discount_type": "flat",
    "value": 50.00,
    "max_discount_amount": null,
    "calculated_discount": 50.00,
    "message": "Discount is valid"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 11.10 Get Discount Usage by Sale

**Endpoint**: `GET /api/v1/discounts/usage/sale/{saleID}`

**Description**: Retrieve discount usage records for a specific sale

**Authentication**: Required

**Path Parameters**:
- `saleID` (string, required) - Sale ID

**Response**:
```json
{
  "success": true,
  "message": "Discount usage retrieved successfully",
  "data": [
    {
      "id": "DUSE_abc123",
      "discount_id": "DISC_abc123",
      "customer_id": "CUST_abc123",
      "sale_id": "SALE_abc123",
      "used_at": "2024-01-01T00:00:00Z",
      "amount": 50.00,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

---

## 12. Taxes API

### 12.1 Create Tax

**Endpoint**: `POST /api/v1/taxes`

**Description**: Create a new tax configuration

**Authentication**: Required

**Request Body**:
```json
{
  "code": "CGST_18",
  "name": "Central GST 18%",
  "description": "Central Goods and Services Tax at 18%",
  "tax_type": "CGST",
  "calculation_type": "PERCENTAGE",
  "rate": 18.0,
  "min_amount": 0.0,
  "max_amount": 10000.0,
  "min_order_value": 100.0,
  "max_order_value": 50000.0,
  "applicable_products": ["PROD_abc123", "PROD_def456"],
  "excluded_products": [],
  "applicable_categories": ["CAT_abc123"],
  "excluded_categories": [],
  "applicable_warehouses": ["WH_abc123"],
  "excluded_warehouses": [],
  "applicable_states": ["Karnataka", "Maharashtra"],
  "excluded_states": [],
  "applicable_customer_groups": ["RETAIL", "WHOLESALE"],
  "excluded_customer_groups": [],
  "valid_from": "2024-01-01T00:00:00Z",
  "valid_until": "2024-12-31T23:59:59Z",
  "is_active": true,
  "priority": 1,
  "is_stackable": true,
  "stacking_order": 1,
  "requires_gstin": true,
  "requires_pan": false,
  "is_inter_state": false,
  "hsn_code": "998314",
  "sac_code": "998314",
  "tax_category": "GST"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Tax created successfully",
  "data": {
    "id": "TAX_abc123",
    "code": "CGST_18",
    "name": "Central GST 18%",
    "description": "Central Goods and Services Tax at 18%",
    "tax_type": "CGST",
    "calculation_type": "PERCENTAGE",
    "rate": 18.0,
    "min_amount": 0.0,
    "max_amount": 10000.0,
    "min_order_value": 100.0,
    "max_order_value": 50000.0,
    "applicable_products": ["PROD_abc123", "PROD_def456"],
    "excluded_products": [],
    "applicable_categories": ["CAT_abc123"],
    "excluded_categories": [],
    "applicable_warehouses": ["WH_abc123"],
    "excluded_warehouses": [],
    "applicable_states": ["Karnataka", "Maharashtra"],
    "excluded_states": [],
    "applicable_customer_groups": ["RETAIL", "WHOLESALE"],
    "excluded_customer_groups": [],
    "valid_from": "2024-01-01T00:00:00Z",
    "valid_until": "2024-12-31T23:59:59Z",
    "is_active": true,
    "priority": 1,
    "is_stackable": true,
    "stacking_order": 1,
    "requires_gstin": true,
    "requires_pan": false,
    "is_inter_state": false,
    "hsn_code": "998314",
    "sac_code": "998314",
    "tax_category": "GST",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 12.2 Get Tax by ID

**Endpoint**: `GET /api/v1/taxes/{id}`

**Description**: Retrieve a specific tax by its ID

**Authentication**: Required

**Path Parameters**:
- `id` (string, required) - Tax ID

**Response**:
```json
{
  "success": true,
  "message": "Tax retrieved successfully",
  "data": {
    "id": "TAX_abc123",
    "code": "CGST_18",
    "name": "Central GST 18%",
    "description": "Central Goods and Services Tax at 18%",
    "tax_type": "CGST",
    "calculation_type": "PERCENTAGE",
    "rate": 18.0,
    "min_amount": 0.0,
    "max_amount": 10000.0,
    "min_order_value": 100.0,
    "max_order_value": 50000.0,
    "applicable_products": ["PROD_abc123", "PROD_def456"],
    "excluded_products": [],
    "applicable_categories": ["CAT_abc123"],
    "excluded_categories": [],
    "applicable_warehouses": ["WH_abc123"],
    "excluded_warehouses": [],
    "applicable_states": ["Karnataka", "Maharashtra"],
    "excluded_states": [],
    "applicable_customer_groups": ["RETAIL", "WHOLESALE"],
    "excluded_customer_groups": [],
    "valid_from": "2024-01-01T00:00:00Z",
    "valid_until": "2024-12-31T23:59:59Z",
    "is_active": true,
    "priority": 1,
    "is_stackable": true,
    "stacking_order": 1,
    "requires_gstin": true,
    "requires_pan": false,
    "is_inter_state": false,
    "hsn_code": "998314",
    "sac_code": "998314",
    "tax_category": "GST",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 12.3 Get All Taxes

**Endpoint**: `GET /api/v1/taxes`

**Description**: Retrieve all taxes with pagination

**Authentication**: Required

**Query Parameters**:
- `limit` (int, optional) - Number of items per page (default: 20, max: 100)
- `offset` (int, optional) - Number of items to skip (default: 0)

**Response**:
```json
{
  "success": true,
  "message": "Taxes retrieved successfully",
  "data": [
    {
      "id": "TAX_abc123",
      "code": "CGST_18",
      "name": "Central GST 18%",
      "description": "Central Goods and Services Tax at 18%",
      "tax_type": "CGST",
      "calculation_type": "PERCENTAGE",
      "rate": 18.0,
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 12.4 Get Active Taxes

**Endpoint**: `GET /api/v1/taxes/active`

**Description**: Retrieve all currently active taxes

**Authentication**: Required

**Response**:
```json
{
  "success": true,
  "message": "Active taxes retrieved successfully",
  "data": [
    {
      "id": "TAX_abc123",
      "code": "CGST_18",
      "name": "Central GST 18%",
      "description": "Central Goods and Services Tax at 18%",
      "tax_type": "CGST",
      "calculation_type": "PERCENTAGE",
      "rate": 18.0,
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 12.5 Get Taxes by Type

**Endpoint**: `GET /api/v1/taxes/type/{taxType}`

**Description**: Retrieve taxes by specific type (CGST, SGST, IGST, VAT, etc.)

**Authentication**: Required

**Path Parameters**:
- `taxType` (string, required) - Tax type (CGST, SGST, IGST, VAT, STT, TDS, TCS, EXCISE, CUSTOMS, ITEM_SPECIFIC, CATEGORY_BASED, FLAT)

**Response**:
```json
{
  "success": true,
  "message": "Taxes by type retrieved successfully",
  "data": [
    {
      "id": "TAX_abc123",
      "code": "CGST_18",
      "name": "Central GST 18%",
      "description": "Central Goods and Services Tax at 18%",
      "tax_type": "CGST",
      "calculation_type": "PERCENTAGE",
      "rate": 18.0,
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 12.6 Get Taxes by Status

**Endpoint**: `GET /api/v1/taxes/status/{status}`

**Description**: Retrieve taxes by status (active, inactive, expired)

**Authentication**: Required

**Path Parameters**:
- `status` (string, required) - Tax status (active, inactive, expired)

**Response**:
```json
{
  "success": true,
  "message": "Taxes by status retrieved successfully",
  "data": [
    {
      "id": "TAX_abc123",
      "code": "CGST_18",
      "name": "Central GST 18%",
      "description": "Central Goods and Services Tax at 18%",
      "tax_type": "CGST",
      "calculation_type": "PERCENTAGE",
      "rate": 18.0,
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 12.7 Update Tax

**Endpoint**: `PUT /api/v1/taxes/{id}`

**Description**: Update an existing tax configuration

**Authentication**: Required

**Path Parameters**:
- `id` (string, required) - Tax ID

**Request Body**:
```json
{
  "name": "Updated Central GST 18%",
  "description": "Updated description",
  "rate": 20.0,
  "is_active": false
}
```

**Response**:
```json
{
  "success": true,
  "message": "Tax updated successfully",
  "data": {
    "id": "TAX_abc123",
    "code": "CGST_18",
    "name": "Updated Central GST 18%",
    "description": "Updated description",
    "tax_type": "CGST",
    "calculation_type": "PERCENTAGE",
    "rate": 20.0,
    "is_active": false,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 12.8 Delete Tax

**Endpoint**: `DELETE /api/v1/taxes/{id}`

**Description**: Delete a tax configuration

**Authentication**: Required

**Path Parameters**:
- `id` (string, required) - Tax ID

**Response**:
```json
{
  "success": true,
  "message": "Tax deleted successfully",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 12.9 Calculate Tax

**Endpoint**: `POST /api/v1/taxes/calculate`

**Description**: Calculate taxes for a given transaction

**Authentication**: Required

**Request Body**:
```json
{
  "warehouse_id": "WH_abc123",
  "customer_state": "Karnataka",
  "warehouse_state": "Karnataka",
  "customer_gstin": "29ABCDE1234F1Z5",
  "customer_pan": "ABCDE1234F",
  "is_inter_state": false,
  "items": [
    {
      "product_id": "PROD_abc123",
      "category_id": "CAT_abc123",
      "quantity": 2,
      "line_total": 1000.00
    }
  ]
}
```

**Response**:
```json
{
  "success": true,
  "message": "Tax calculation completed successfully",
  "data": {
    "sub_total": 1000.00,
    "tax_breakdown": [
      {
        "tax_type": "CGST",
        "tax_name": "Central GST 18%",
        "tax_code": "CGST_18",
        "rate": 18.0,
        "amount": 90.00,
        "hsn_code": "998314",
        "sac_code": "998314"
      },
      {
        "tax_type": "SGST",
        "tax_name": "State GST 18%",
        "tax_code": "SGST_18",
        "rate": 18.0,
        "amount": 90.00,
        "hsn_code": "998314",
        "sac_code": "998314"
      }
    ],
    "total_tax_amount": 180.00,
    "grand_total": 1180.00,
    "applied_taxes": [
      {
        "tax_id": "TAX_abc123",
        "tax_code": "CGST_18",
        "tax_name": "Central GST 18%",
        "tax_type": "CGST",
        "rate": 18.0,
        "amount": 90.00,
        "base_amount": 1000.00
      }
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 12.10 Get Tax Summary by Sale

**Endpoint**: `GET /api/v1/taxes/summary/sale/{saleID}`

**Description**: Retrieve tax summary for a specific sale

**Authentication**: Required

**Path Parameters**:
- `saleID` (string, required) - Sale ID

**Response**:
```json
{
  "success": true,
  "message": "Tax summary retrieved successfully",
  "data": {
    "id": "TSUM_abc123",
    "sale_id": "SALE_abc123",
    "return_id": null,
    "sub_total": 1000.00,
    "total_tax_amount": 180.00,
    "grand_total": 1180.00,
    "cgst_amount": 90.00,
    "sgst_amount": 90.00,
    "igst_amount": 0.00,
    "vat_amount": 0.00,
    "stt_amount": 0.00,
    "tds_amount": 0.00,
    "tcs_amount": 0.00,
    "excise_amount": 0.00,
    "customs_amount": 0.00,
    "other_tax_amount": 0.00,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 12.11 Get Tax Summary by Return

**Endpoint**: `GET /api/v1/taxes/summary/return/{returnID}`

**Description**: Retrieve tax summary for a specific return

**Authentication**: Required

**Path Parameters**:
- `returnID` (string, required) - Return ID

**Response**:
```json
{
  "success": true,
  "message": "Tax summary retrieved successfully",
  "data": {
    "id": "TSUM_abc123",
    "sale_id": null,
    "return_id": "RET_abc123",
    "sub_total": 500.00,
    "total_tax_amount": 90.00,
    "grand_total": 590.00,
    "cgst_amount": 45.00,
    "sgst_amount": 45.00,
    "igst_amount": 0.00,
    "vat_amount": 0.00,
    "stt_amount": 0.00,
    "tds_amount": 0.00,
    "tcs_amount": 0.00,
    "excise_amount": 0.00,
    "customs_amount": 0.00,
    "other_tax_amount": 0.00,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 12.12 Get Tax Applications by Sale

**Endpoint**: `GET /api/v1/taxes/applications/sale/{saleID}`

**Description**: Retrieve tax applications for a specific sale

**Authentication**: Required

**Path Parameters**:
- `saleID` (string, required) - Sale ID

**Response**:
```json
{
  "success": true,
  "message": "Tax applications retrieved successfully",
  "data": [
    {
      "id": "TAPP_abc123",
      "tax_id": "TAX_abc123",
      "sale_id": "SALE_abc123",
      "return_id": null,
      "base_amount": 1000.00,
      "tax_rate": 18.0,
      "tax_amount": 180.00,
      "tax_type": "CGST",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 12.13 Get Tax Applications by Return

**Endpoint**: `GET /api/v1/taxes/applications/return/{returnID}`

**Description**: Retrieve tax applications for a specific return

**Authentication**: Required

**Path Parameters**:
- `returnID` (string, required) - Return ID

**Response**:
```json
{
  "success": true,
  "message": "Tax applications retrieved successfully",
  "data": [
    {
      "id": "TAPP_abc123",
      "tax_id": "TAX_abc123",
      "sale_id": null,
      "return_id": "RET_abc123",
      "base_amount": 500.00,
      "tax_rate": 18.0,
      "tax_amount": 90.00,
      "tax_type": "CGST",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 12.14 Get Taxes by Product

**Endpoint**: `GET /api/v1/taxes/product/{productID}`

**Description**: Retrieve taxes applicable to a specific product

**Authentication**: Required

**Path Parameters**:
- `productID` (string, required) - Product ID

**Query Parameters**:
- `warehouse_id` (string, optional) - Warehouse ID for filtering
- `customer_state` (string, optional) - Customer state for filtering
- `is_inter_state` (boolean, optional) - Whether it's an inter-state transaction

**Response**:
```json
{
  "success": true,
  "message": "Product taxes retrieved successfully",
  "data": [
    {
      "id": "TAX_abc123",
      "code": "CGST_18",
      "name": "Central GST 18%",
      "description": "Central Goods and Services Tax at 18%",
      "tax_type": "CGST",
      "calculation_type": "PERCENTAGE",
      "rate": 18.0,
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 12.15 Get Taxes by Category

**Endpoint**: `GET /api/v1/taxes/category/{categoryID}`

**Description**: Retrieve taxes applicable to a specific category

**Authentication**: Required

**Path Parameters**:
- `categoryID` (string, required) - Category ID

**Query Parameters**:
- `warehouse_id` (string, optional) - Warehouse ID for filtering
- `customer_state` (string, optional) - Customer state for filtering
- `is_inter_state` (boolean, optional) - Whether it's an inter-state transaction

**Response**:
```json
{
  "success": true,
  "message": "Category taxes retrieved successfully",
  "data": [
    {
      "id": "TAX_abc123",
      "code": "CGST_18",
      "name": "Central GST 18%",
      "description": "Central Goods and Services Tax at 18%",
      "tax_type": "CGST",
      "calculation_type": "PERCENTAGE",
      "rate": 18.0,
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```



## 13. Additional Notes

### 13.1 Rate Limiting

The API implements rate limiting to prevent abuse:
- **Default**: 100 requests per minute per IP
- **Authenticated users**: 1000 requests per minute
- **Admin users**: 5000 requests per minute

### 13.2 File Upload Limits

For attachment uploads:
- **Maximum file size**: 10MB
- **Allowed formats**: PDF, JPG, PNG, DOC, DOCX
- **Storage**: AWS S3 (configurable)

### 13.3 Pagination

For endpoints that return lists, pagination is supported:
- **Default page size**: 20 items
- **Maximum page size**: 100 items
- **Query parameters**: `page` and `limit`

### 13.4 Filtering and Sorting

Many endpoints support filtering and sorting:
- **Filtering**: Use query parameters like `status`, `warehouse_id`, etc.
- **Sorting**: Use `sort_by` and `sort_order` parameters
- **Date ranges**: Use `start_date` and `end_date` parameters

### 13.5 Webhook Support

The API supports webhooks for real-time notifications:
- **Events**: Sale created, Return processed, Low stock alert
- **Configuration**: Set webhook URL in environment variables
- **Authentication**: Webhook signatures for security

---

*This completes the comprehensive API documentation for the Kisanlink ERP system. All endpoints are documented with request/response examples, authentication requirements, and error handling.*
