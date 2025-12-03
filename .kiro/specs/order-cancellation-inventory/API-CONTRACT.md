# Order Cancellation API Contract

**Version:** 1.0.0
**Last Updated:** 2025-11-25
**Status:** Production Ready

## Overview

This document provides the API contract for downstream services to implement order cancellation functionality. When a sale/order is cancelled, inventory is automatically returned to the original batches.

---

## Base URL

```
Production: https://api.kisanlink.in/api/v1
Staging:    https://api-staging.kisanlink.in/api/v1
```

---

## Authentication

All endpoints require Bearer token authentication.

```
Authorization: Bearer <JWT_TOKEN>
```

### Required Permission
- `sale:cancel` - Permission to cancel sales

---

## Endpoints

### 1. Cancel Sale

Cancels a sale and returns inventory to original batches.

#### Request

```
POST /sales/{sale_id}/cancel
Content-Type: application/json
Authorization: Bearer <token>
```

#### Path Parameters

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| sale_id   | string | Yes      | The unique identifier of the sale (e.g., `SALE_12345678`) |

#### Request Body

```json
{
  "reason": "customer_request",
  "reason_details": "Customer changed their mind about the purchase",
  "skip_inventory_return": false,
  "performed_by": "USER_abc123"
}
```

| Field                | Type    | Required | Description                                           |
|----------------------|---------|----------|-------------------------------------------------------|
| reason               | string  | Yes      | Cancellation reason code (see enum below)             |
| reason_details       | string  | No       | Additional details about the cancellation (max 1000 chars) |
| skip_inventory_return| boolean | No       | If true, inventory won't be restored (default: false) |
| performed_by         | string  | Yes      | User ID of the person performing the cancellation     |

#### Reason Enum Values

| Value             | Description                                    |
|-------------------|------------------------------------------------|
| `customer_request`| Customer requested cancellation                |
| `payment_failed`  | Payment processing failed                      |
| `out_of_stock`    | Items went out of stock after order           |
| `pricing_error`   | Incorrect pricing was applied                  |
| `duplicate_order` | Order was a duplicate                          |
| `fraud_suspected` | Potential fraudulent activity detected         |
| `system_error`    | System error occurred during processing        |
| `other`           | Other reason (provide details in reason_details)|

#### Response - Success (200 OK)

```json
{
  "status": "success",
  "message": "Sale cancelled successfully",
  "data": {
    "sale": {
      "id": "SALE_12345678",
      "organization_id": "ORG_xyz789",
      "warehouse_id": "WH_abc123",
      "status": "cancelled",
      "cancelled_at": "2025-11-25T10:30:00Z",
      "cancellation_reason": "customer_request",
      "total_amount": 1500.00,
      "discount_amount": 100.00,
      "tax_amount": 210.00,
      "net_amount": 1610.00,
      "items": [
        {
          "id": "SLITM_item001",
          "variant_id": "VAR_prod123",
          "batch_id": "BTCH_batch456",
          "quantity": 10,
          "unit_price": 150.00,
          "total_price": 1500.00
        }
      ],
      "created_at": "2025-11-24T15:00:00Z",
      "updated_at": "2025-11-25T10:30:00Z"
    },
    "inventory_restored": [
      {
        "batch_id": "BTCH_batch456",
        "variant_id": "VAR_prod123",
        "quantity_restored": 10,
        "transaction_id": "INVTXN_txn789"
      }
    ],
    "financial_adjustments": {
      "discount_reversed": {
        "discount_id": "DISC_promo100",
        "amount_reversed": 100.00,
        "usage_decremented": true
      },
      "tax_voided": {
        "tax_summary_id": "TAXSUM_tax001",
        "amount_voided": 210.00
      }
    },
    "cancellation_id": "SLCAN_can123"
  }
}
```

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| sale | object | The updated sale object with cancelled status |
| inventory_restored | array | List of inventory items that were restored |
| financial_adjustments | object | Discount and tax reversals (if applicable) |
| cancellation_id | string | Unique ID for this cancellation record |

#### Response - Error Cases

##### 400 Bad Request - Invalid Request

```json
{
  "status": "error",
  "message": "Invalid request data",
  "errors": [
    {
      "field": "reason",
      "message": "reason is required"
    }
  ]
}
```

##### 401 Unauthorized

```json
{
  "status": "error",
  "message": "Unauthorized",
  "error": "Invalid or expired token"
}
```

##### 403 Forbidden - Insufficient Permissions

```json
{
  "status": "error",
  "message": "Forbidden - insufficient permissions",
  "error": "User does not have sale:cancel permission"
}
```

##### 404 Not Found - Sale Not Found

```json
{
  "status": "error",
  "message": "Sale not found",
  "error": "Sale with ID SALE_invalid does not exist"
}
```

##### 409 Conflict - Sale Cannot Be Cancelled

```json
{
  "status": "error",
  "message": "Sale cannot be cancelled",
  "error": "Sale status is 'shipped' - cannot cancel shipped orders"
}
```

---

## Business Rules

### Cancellation Eligibility

| Sale Status   | Can Cancel? | Notes                                    |
|---------------|-------------|------------------------------------------|
| `pending`     | Yes         | Full cancellation allowed                |
| `confirmed`   | Yes         | Full cancellation allowed                |
| `processing`  | Yes         | Full cancellation allowed                |
| `shipped`     | No          | Use Returns API instead                  |
| `delivered`   | No          | Use Returns API instead                  |
| `cancelled`   | No          | Already cancelled                        |
| `returned`    | No          | Already returned                         |

### Inventory Restoration

1. When a sale is cancelled, inventory is returned to the **original batches**
2. FEFO (First Expiry First Out) tracking is maintained
3. Each restored item creates an `InventoryTransaction` with type `cancellation_return`
4. If `skip_inventory_return: true`, no inventory transactions are created

### Financial Reversals

1. **Discounts**: If a discount was applied, usage count is decremented
2. **Taxes**: Tax records are marked as voided (not deleted for audit purposes)
3. All financial amounts are preserved in the cancellation record for audit

### Concurrency Handling

- The API uses pessimistic locking (`SELECT FOR UPDATE`) to prevent race conditions
- If another cancellation is in progress, request will wait or timeout
- Recommended timeout: 30 seconds

---

## Code Examples

### cURL

```bash
curl -X POST "https://api.kisanlink.in/api/v1/sales/SALE_12345678/cancel" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "reason": "customer_request",
    "reason_details": "Customer changed their mind",
    "performed_by": "USER_admin001"
  }'
```

### JavaScript/TypeScript

```typescript
interface CancelSaleRequest {
  reason:
    | 'customer_request'
    | 'payment_failed'
    | 'out_of_stock'
    | 'pricing_error'
    | 'duplicate_order'
    | 'fraud_suspected'
    | 'system_error'
    | 'other';
  reason_details?: string;
  skip_inventory_return?: boolean;
  performed_by: string;
}

interface InventoryRestoredItem {
  batch_id: string;
  variant_id: string;
  quantity_restored: number;
  transaction_id: string;
}

interface CancelSaleResponse {
  sale: Sale;
  inventory_restored: InventoryRestoredItem[];
  financial_adjustments?: {
    discount_reversed?: {
      discount_id: string;
      amount_reversed: number;
      usage_decremented: boolean;
    };
    tax_voided?: {
      tax_summary_id: string;
      amount_voided: number;
    };
  };
  cancellation_id: string;
}

async function cancelSale(
  saleId: string,
  request: CancelSaleRequest,
  token: string
): Promise<CancelSaleResponse> {
  const response = await fetch(
    `https://api.kisanlink.in/api/v1/sales/${saleId}/cancel`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify(request),
    }
  );

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'Failed to cancel sale');
  }

  const result = await response.json();
  return result.data;
}

// Usage
const result = await cancelSale(
  'SALE_12345678',
  {
    reason: 'customer_request',
    reason_details: 'Customer changed their mind',
    performed_by: 'USER_admin001',
  },
  'your_jwt_token'
);

console.log(`Sale cancelled. Cancellation ID: ${result.cancellation_id}`);
console.log(`Items restored: ${result.inventory_restored.length}`);
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type CancelSaleRequest struct {
    Reason              string `json:"reason"`
    ReasonDetails       string `json:"reason_details,omitempty"`
    SkipInventoryReturn bool   `json:"skip_inventory_return,omitempty"`
    PerformedBy         string `json:"performed_by"`
}

type InventoryRestoredItem struct {
    BatchID          string `json:"batch_id"`
    VariantID        string `json:"variant_id"`
    QuantityRestored int64  `json:"quantity_restored"`
    TransactionID    string `json:"transaction_id"`
}

type CancelSaleResponse struct {
    Sale                Sale                    `json:"sale"`
    InventoryRestored   []InventoryRestoredItem `json:"inventory_restored"`
    CancellationID      string                  `json:"cancellation_id"`
}

func CancelSale(saleID string, req CancelSaleRequest, token string) (*CancelSaleResponse, error) {
    url := fmt.Sprintf("https://api.kisanlink.in/api/v1/sales/%s/cancel", saleID)

    body, _ := json.Marshal(req)
    request, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
    request.Header.Set("Content-Type", "application/json")
    request.Header.Set("Authorization", "Bearer "+token)

    client := &http.Client{}
    resp, err := client.Do(request)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
    }

    var result struct {
        Data CancelSaleResponse `json:"data"`
    }
    json.NewDecoder(resp.Body).Decode(&result)

    return &result.Data, nil
}
```

### Python

```python
import requests
from typing import Optional, List
from dataclasses import dataclass
from enum import Enum

class CancellationReason(Enum):
    CUSTOMER_REQUEST = "customer_request"
    PAYMENT_FAILED = "payment_failed"
    OUT_OF_STOCK = "out_of_stock"
    PRICING_ERROR = "pricing_error"
    DUPLICATE_ORDER = "duplicate_order"
    FRAUD_SUSPECTED = "fraud_suspected"
    SYSTEM_ERROR = "system_error"
    OTHER = "other"

@dataclass
class InventoryRestoredItem:
    batch_id: str
    variant_id: str
    quantity_restored: int
    transaction_id: str

def cancel_sale(
    sale_id: str,
    reason: CancellationReason,
    performed_by: str,
    token: str,
    reason_details: Optional[str] = None,
    skip_inventory_return: bool = False
) -> dict:
    """Cancel a sale and return inventory to original batches."""

    url = f"https://api.kisanlink.in/api/v1/sales/{sale_id}/cancel"

    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {token}"
    }

    payload = {
        "reason": reason.value,
        "performed_by": performed_by,
        "skip_inventory_return": skip_inventory_return
    }

    if reason_details:
        payload["reason_details"] = reason_details

    response = requests.post(url, json=payload, headers=headers)
    response.raise_for_status()

    return response.json()["data"]

# Usage
result = cancel_sale(
    sale_id="SALE_12345678",
    reason=CancellationReason.CUSTOMER_REQUEST,
    performed_by="USER_admin001",
    token="your_jwt_token",
    reason_details="Customer changed their mind"
)

print(f"Sale cancelled. Cancellation ID: {result['cancellation_id']}")
for item in result["inventory_restored"]:
    print(f"  - Restored {item['quantity_restored']} units to batch {item['batch_id']}")
```

---

## Webhook Events (Future)

When order cancellation webhooks are enabled, the following events will be emitted:

### `sale.cancelled`

```json
{
  "event": "sale.cancelled",
  "timestamp": "2025-11-25T10:30:00Z",
  "data": {
    "sale_id": "SALE_12345678",
    "cancellation_id": "SLCAN_can123",
    "reason": "customer_request",
    "cancelled_by": "USER_admin001",
    "inventory_restored": true,
    "total_amount": 1610.00
  }
}
```

---

## Rate Limits

| Endpoint              | Rate Limit     | Window   |
|-----------------------|----------------|----------|
| POST /sales/{id}/cancel | 100 requests | 1 minute |

---

## Changelog

| Version | Date       | Changes                                  |
|---------|------------|------------------------------------------|
| 1.0.0   | 2025-11-25 | Initial release                          |

---

## Support

For questions or issues:
- Technical Support: tech@kisanlink.in
- API Documentation: https://docs.kisanlink.in/api
- GitHub Issues: https://github.com/Kisanlink/fpo-erp/issues
