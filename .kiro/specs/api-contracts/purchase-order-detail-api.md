# Purchase Order Detail API Contract

## Overview

**Purpose**: Provide complete purchase order information with related collaborator, warehouse, GRN, and inventory data in a single API call.

**Current Problem**:
- PO detail page requires 5 API calls
- Sequential fetches for collaborator, warehouse, GRN status, inventory
- 400-600ms total latency

**Solution Impact**:
- **80% reduction** in API calls (5 → 1)
- **400-600ms** faster page loads
- Complete PO lifecycle view in single response

---

## API Specification

### Endpoint: Get Purchase Order Detail

```
GET /api/v1/purchase-orders/{po_id}/detail
```

**Authentication**: Required
**Authorization**: `purchase_order:read` permission

#### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `po_id` | string | Yes | Purchase order identifier |

#### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `include` | string[] | No | all | `collaborator,warehouse,items,grns,inventory,payments` |

#### Response Example

```json
{
  "purchase_order": {
    "id": "PO_2024_001",
    "po_number": "PO-2024-001",
    "organization_id": "ORG_001",
    "collaborator_id": "CLAB_789",
    "warehouse_id": "WH_001",
    "status": "partially_received",
    "order_date": "2024-11-01T00:00:00Z",
    "expected_delivery_date": "2024-11-15T00:00:00Z",
    "actual_delivery_date": null,
    "total_amount": 125000.00,
    "paid_amount": 50000.00,
    "pending_amount": 75000.00,
    "currency": "INR",
    "payment_terms": "30 days net",
    "notes": "Urgent delivery required",
    "created_by": "user_ceo_001",
    "created_at": "2024-11-01T10:30:00Z",
    "updated_at": "2024-11-15T14:20:00Z"
  },

  "collaborator": {
    "id": "CLAB_789",
    "company_name": "Punjab Rice Mills",
    "contact_person": "Rajesh Kumar",
    "phone": "+91-9876543210",
    "email": "contact@punjabrice.com",
    "gstin": "03AABCP1234F1Z5",
    "pan": "AABCP1234F",
    "address": {
      "street": "Industrial Area, Phase 2",
      "city": "Ludhiana",
      "state": "Punjab",
      "pincode": "141003",
      "country": "India"
    },
    "payment_terms": "30 days",
    "credit_limit": 500000.00,
    "outstanding_balance": 125000.00,
    "is_active": true
  },

  "warehouse": {
    "id": "WH_001",
    "name": "Main Warehouse - Delhi",
    "address": {
      "street": "Plot 45, Industrial Area",
      "city": "New Delhi",
      "state": "Delhi",
      "pincode": "110020",
      "country": "India"
    },
    "contact_phone": "+91-11-12345678",
    "capacity": {
      "total_capacity_kg": 100000,
      "used_capacity_kg": 67500,
      "available_capacity_kg": 32500,
      "utilization_percent": 67.5
    },
    "is_active": true
  },

  "items": [
    {
      "id": "POITEM_001",
      "variant_id": "PVAR_001",
      "variant": {
        "id": "PVAR_001",
        "variant_name": "Premium Basmati Rice - 1kg Pack",
        "sku": "PBR-1KG-001",
        "brand_name": "Golden Harvest",
        "quantity": "1kg",
        "pack_size": "Standard Pack",
        "images": ["https://cdn.example.com/products/rice-1kg-front.jpg"]
      },
      "product": {
        "id": "PROD_12345",
        "name": "Premium Basmati Rice",
        "category": "Grains"
      },
      "ordered_quantity": 5000,
      "received_quantity": 3000,
      "pending_quantity": 2000,
      "unit_cost": 55.00,
      "total_cost": 275000.00,
      "received_status": "partially_received",
      "expected_delivery_date": "2024-11-15T00:00:00Z"
    },
    {
      "id": "POITEM_002",
      "variant_id": "PVAR_002",
      "variant": {
        "id": "PVAR_002",
        "variant_name": "Premium Basmati Rice - 5kg Pack",
        "sku": "PBR-5KG-001",
        "brand_name": "Golden Harvest",
        "quantity": "5kg",
        "pack_size": "Bulk Pack",
        "images": ["https://cdn.example.com/products/rice-5kg-front.jpg"]
      },
      "product": {
        "id": "PROD_12345",
        "name": "Premium Basmati Rice",
        "category": "Grains"
      },
      "ordered_quantity": 2000,
      "received_quantity": 2000,
      "pending_quantity": 0,
      "unit_cost": 270.00,
      "total_cost": 540000.00,
      "received_status": "fully_received",
      "expected_delivery_date": "2024-11-15T00:00:00Z"
    }
  ],

  "grns": [
    {
      "id": "GRN_2024_001",
      "grn_number": "GRN-2024-001",
      "purchase_order_id": "PO_2024_001",
      "received_date": "2024-11-10T14:30:00Z",
      "status": "accepted",
      "received_by": "user_warehouse_manager_001",
      "notes": "First batch received and inspected",
      "items": [
        {
          "po_item_id": "POITEM_002",
          "variant_id": "PVAR_002",
          "ordered_quantity": 2000,
          "received_quantity": 2000,
          "accepted_quantity": 2000,
          "rejected_quantity": 0,
          "damage_quantity": 0,
          "unit_cost": 270.00,
          "total_cost": 540000.00
        }
      ],
      "inventory_created": [
        {
          "batch_id": "BTCH_002",
          "variant_id": "PVAR_002",
          "warehouse_id": "WH_001",
          "quantity": 2000,
          "cost_price": 270.00,
          "expiry_date": "2025-11-30",
          "manufacturing_date": "2024-11-01",
          "batch_number": "BATCH-2024-11-002"
        }
      ]
    },
    {
      "id": "GRN_2024_002",
      "grn_number": "GRN-2024-002",
      "purchase_order_id": "PO_2024_001",
      "received_date": "2024-11-15T10:15:00Z",
      "status": "accepted",
      "received_by": "user_warehouse_manager_001",
      "notes": "Second batch - partial delivery",
      "items": [
        {
          "po_item_id": "POITEM_001",
          "variant_id": "PVAR_001",
          "ordered_quantity": 5000,
          "received_quantity": 3000,
          "accepted_quantity": 3000,
          "rejected_quantity": 0,
          "damage_quantity": 0,
          "unit_cost": 55.00,
          "total_cost": 165000.00
        }
      ],
      "inventory_created": [
        {
          "batch_id": "BTCH_001",
          "variant_id": "PVAR_001",
          "warehouse_id": "WH_001",
          "quantity": 3000,
          "cost_price": 55.00,
          "expiry_date": "2025-12-31",
          "manufacturing_date": "2024-12-01",
          "batch_number": "BATCH-2024-12-001"
        }
      ]
    }
  ],

  "payments": [
    {
      "id": "PAY_001",
      "purchase_order_id": "PO_2024_001",
      "payment_date": "2024-11-05T00:00:00Z",
      "amount": 50000.00,
      "payment_method": "bank_transfer",
      "reference_number": "TXN-2024-11-05-001",
      "notes": "Advance payment",
      "status": "completed",
      "created_by": "user_accountant_001"
    }
  ],

  "summary": {
    "total_order_value": 815000.00,
    "total_received_value": 705000.00,
    "total_pending_value": 110000.00,
    "completion_percentage": 86.5,
    "total_items_ordered": 7000,
    "total_items_received": 5000,
    "total_items_pending": 2000,
    "payment_status": "partially_paid",
    "fulfillment_status": "partially_received"
  },

  "timeline": [
    {
      "timestamp": "2024-11-01T10:30:00Z",
      "event": "purchase_order_created",
      "description": "Purchase order created",
      "actor": "user_ceo_001"
    },
    {
      "timestamp": "2024-11-05T15:00:00Z",
      "event": "payment_received",
      "description": "Advance payment of ₹50,000 received",
      "actor": "user_accountant_001"
    },
    {
      "timestamp": "2024-11-10T14:30:00Z",
      "event": "grn_created",
      "description": "GRN-2024-001 created - 2000 units received",
      "actor": "user_warehouse_manager_001"
    },
    {
      "timestamp": "2024-11-15T10:15:00Z",
      "event": "grn_created",
      "description": "GRN-2024-002 created - 3000 units received",
      "actor": "user_warehouse_manager_001"
    }
  ],

  "metadata": {
    "read_timestamp": "2024-11-21T10:30:00Z",
    "consistency_token": "CT_po_xyz789"
  }
}
```

---

## Business Rules

### PO Status Transitions

```
draft → submitted → approved → partially_received → fully_received → closed
                             ↓
                          cancelled
```

### Fulfillment Calculation

- `completion_percentage = (received_quantity / ordered_quantity) * 100`
- `fulfillment_status`:
  - 0%: `pending`
  - 1-99%: `partially_received`
  - 100%: `fully_received`

### Payment Status

- `payment_status`:
  - `unpaid`: paid_amount = 0
  - `partially_paid`: 0 < paid_amount < total_amount
  - `fully_paid`: paid_amount >= total_amount

### GRN-Inventory Linkage

- Each GRN creates corresponding inventory batches
- Batch IDs linked to GRN for traceability
- FEFO applies from manufacturing/expiry dates

---

## Performance

- **Target P95**: < 300ms
- **Caching**: PO metadata (5 min), collaborator (10 min)
- **No Cache**: GRN status, inventory (real-time)

---

## Use Cases

### Use Case 1: Procurement Manager Reviews PO

**Old Flow** (5 calls):
```
GET /purchase-orders/:id
GET /collaborators/:id
GET /warehouses/:id
GET /grns?po_id=:id
GET /batches?grn_id=:id (per GRN)
```

**New Flow** (1 call):
```
GET /purchase-orders/:id/detail
```

**Benefit**: 80% reduction, complete context in one response

---

## Related Contracts

- [Sales Context API](./sales-context-api.md)
- [Inventory List API](./inventory-list-api.md)
