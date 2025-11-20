# API Changes - Purchase Order Workflow & Rejected Goods Tracking

**Date**: November 20, 2025
**Version**: 1.2.0
**Features**:
- Purchase Order "Verified" Status (Quality Inspection Stage)
- Rejected Goods Return Workflow

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
