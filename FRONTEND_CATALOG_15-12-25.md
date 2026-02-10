## Discounts & Validation (SKU-Level Support)

### Create Discount – `POST /api/v1/discounts`

- **Purpose**: Create a discount that can target variants (SKUs), products, categories, and/or warehouses.
- **Content-Type**: `application/json`

#### Request (percentage discount with variant-level targeting)

```json
{
  "code": "TEST20",
  "name": "20% off selected variants",
  "description": "20% discount on specific variants",
  "discount_type": "percentage",
  "value": 20,
  "max_discount_amount": 1000,
  "valid_from": "2025-12-15T00:00:00Z",
  "valid_until": "2025-12-31T23:59:59Z",
  "is_active": true,
  "is_stackable": true,
  "priority": 1,
  "applicable_variants": "[\"PVAR00000005\",\"PVAR00000006\"]"
}
```

- **Field notes**:
  - **discount_type**: `"percentage"`, `"flat"`, `"buy_x_get_y"`, etc.
  - **value**: for percentage discounts, must be in `(0, 100]`.
  - **valid_from** / **valid_until**: RFC3339 timestamps (for example `2025-12-15T00:00:00Z`).
  - **applicable_variants**:
    - Optional.
    - Stringified JSON array of variant IDs (for example `PVAR00000005`).
    - If present and non-empty, must be a valid non-empty JSON array.
  - Existing fields like **applicable_products**, **applicable_categories**, **applicable_warehouses** continue to work.

#### Response

```json
{
  "success": true,
  "message": "Discount created successfully",
  "data": {
    "id": "DISC00000001",
    "code": "TEST20",
    "name": "20% off selected variants",
    "description": "20% discount on specific variants",
    "discount_type": "percentage",
    "value": 20,
    "max_discount_amount": 1000,
    "min_order_value": null,
    "max_order_value": null,
    "applicable_variants": "[\"PVAR00000005\",\"PVAR00000006\"]",
    "applicable_products": null,
    "excluded_products": null,
    "applicable_categories": null,
    "excluded_categories": null,
    "applicable_warehouses": null,
    "usage_limit": null,
    "current_usage": 0,
    "valid_from": "2025-12-15T00:00:00Z",
    "valid_until": "2025-12-31T23:59:59Z",
    "is_active": true,
    "is_stackable": true,
    "priority": 1,
    "terms": null,
    "status": "active",
    "buy_quantity": null,
    "get_quantity": null,
    "get_discount_type": null,
    "get_discount_value": null,
    "created_at": "2025-12-15T10:00:00Z",
    "updated_at": "2025-12-15T10:00:00Z"
  }
}
```

---

### Validate Discount – `POST /api/v1/discounts/validate`

- **Purpose**: Check whether a discount code can be applied for a given cart.
- **Content-Type**: `application/json`

#### Request (variant/SKU-based)

```json
{
  "discount_code": "TEST20",
  "order_value": 50,
  "variant_ids": ["PVAR00000005"],
  "product_ids": ["PROD00000003"],
  "category_ids": [],
  "warehouse_id": "WHSE00000004"
}
```

- **Field notes**:
  - **discount_code** (string, required): Code like `"TEST20"`.
  - **order_value** (number, required, `> 0`): Cart total to evaluate the discount against (sum of `price * quantity` per line).
  - **variant_ids** (string[], required):
    - Variant IDs in the cart (for example `PVAR00000005`).
    - Used for matching `applicable_variants` on discounts.
  - **product_ids** (string[], optional):
    - Still supported for product-level discounts (`applicable_products` / `excluded_products`).
  - **category_ids** (string[], optional):
    - Used when discounts are defined by categories.
  - **warehouse_id** (string, required): Warehouse where the order is fulfilled.

#### Response

```json
{
  "success": true,
  "message": "Discount validation completed",
  "data": {
    "is_valid": true,
    "discount_id": "DISC00000001",
    "discount_code": "TEST20",
    "discount_name": "20% off selected variants",
    "discount_type": "percentage",
    "value": 20,
    "max_discount_amount": 1000,
    "calculated_discount": 10,
    "message": "Discount is valid"
  }
}
```

- If validation fails, `is_valid` will be `false` and `message` will contain the reason (for example: not applicable to these variants, warehouse, categories, order value, or period).

---

### Applicable Discounts – `GET /api/v1/discounts/applicable`

- **Purpose**: Get all discounts that could apply to a given order.
- **Query params**:
  - `order_value` (required, number)
  - `warehouse_id` (required, string)
  - `product_ids` (optional, comma-separated string of product IDs)
  - `category_ids` (optional, comma-separated string of category IDs)
- **Note**: This endpoint is still product/category based. Variant-level targeting is primarily enforced via `POST /discounts/validate`.

---

### Frontend Integration Notes

- For **cart-level validation**:
  - Always send:
    - `discount_code`
    - `order_value` (cart total)
    - `variant_ids` (variants present in the cart)
    - `warehouse_id`
  - Optionally send:
    - `product_ids` if product-level discounts are used.
    - `category_ids` if category-level discounts are used.
- Discount targeting priority on the backend:
  - If `applicable_variants` is set on the discount → variant IDs are used.
  - Else if `applicable_products` is set → product IDs are used.
  - Category and warehouse rules always apply when configured.

---

## Sales – Tax, Discount & Margin Breakdown

### Get Sale Details – `GET /api/v1/sales/{id}`

- **Purpose**: Retrieve full sale details, including item-wise prices, discounts, taxes, and margins.
- **Content-Type**: `application/json`

#### Response (important fields)

- **Top-level**:
  - `total_amount`: Final amount **after discount and tax**.
  - `apply_taxes`: Whether GST was applied to this sale.
  - `breakdown`:
    - `base_amount`: Total before discounts and taxes (sum of gross line totals).
    - `discount_amount`: Total discount applied to the sale.
    - `net_amount_before_tax`: `base_amount - discount_amount` (taxable base).
    - `tax_amount`: Total GST amount calculated **after discount**.
    - `total_savings`: Same as `discount_amount`.
    - `final_amount`: `net_amount_before_tax + tax_amount` (equals `total_amount`).

- **Items (`items[]`)**:
  - `selling_price`: Gross unit price (before discount).
  - `line_total`: Gross line total = `selling_price * quantity`.
  - `discount_amount`: Discount allocated to this line.
  - `net_line_total`: `line_total - discount_amount` (base for tax).
  - `net_selling_price`: `net_line_total / quantity` (effective unit price after discount).
  - `cost_price`: Cost price per unit.
  - `margin`: Per-unit margin **after discount** = `net_selling_price - cost_price`.
  - `cgst_amount`, `sgst_amount`, `igst_amount`, `total_tax_amount`: GST amounts calculated on `net_line_total` (after discount).

#### Example (simplified)

```json
{
  "data": {
    "id": "SALE00000001",
    "warehouse_id": "WHSE00000001",
    "invoice_number": "122500000001",
    "total_amount": 1180,
    "items": [
      {
        "id": "SITEM00000001",
        "sku": "SKU-001",
        "quantity": 2,
        "selling_price": 600,
        "line_total": 1200,
        "discount_amount": 200,
        "net_line_total": 1000,
        "net_selling_price": 500,
        "cost_price": 400,
        "margin": 100,
        "cgst_amount": 90,
        "sgst_amount": 90,
        "igst_amount": 0,
        "total_tax_amount": 180
      }
    ],
    "breakdown": {
      "base_amount": 1200,
      "discount_amount": 200,
      "net_amount_before_tax": 1000,
      "tax_amount": 180,
      "total_savings": 200,
      "final_amount": 1180
    }
  }
}
```

**Key behavior**:

- **Discounts are applied first**, then **GST is calculated on the discounted (net) line amounts**.
- `margin` in responses is always **after discount**, using `net_selling_price`.

---

## Sales – Cancellation & Refunds (Full and Partial)

### Cancel Full Sale – `POST /api/v1/sales/{id}/cancel`

- **Purpose**: Cancel an entire sale and reverse inventory, discounts, and taxes.
- **Refund principle**: The effective refund equals the **final paid amount** (`sale.total_amount`), which is net after discounts and taxes.

#### Response (key points)

- `sale.total_amount`: Final amount of the sale (before cancellation, for reference).
- `financial_adjustments`:
  - `discount_reversed.amount_reversed`: Total discount reversed for this cancellation.
  - `tax_voided.amount_voided`: Total GST voided.
- `cancellation_id`: ID for the `SaleCancellation` record.
- In `GET /api/v1/sales/{id}/cancellations`, each `SaleCancellationItem` has:
  - `refund_amount`: Per-line refund based on **net after discount + tax**.

### Cancel Items (Partial) – `POST /api/v1/sales/{id}/cancel-items`

- **Purpose**: Cancel specific items or quantities in a sale.
- **Refund principle**: Each cancelled unit is refunded at its **final paid price**:
  - `refund_per_unit = net_unit_price_after_discount + tax_per_unit`.

#### Per-line refund math (backend)

- For a sale item:
  - `line_total` = gross base (`selling_price * quantity`).
  - `discount_amount` = discount allocated to this line.
  - `total_tax_amount` = GST on the net line after discount.
- Backend computes:
  - `net_line_total = line_total - discount_amount`.
  - `net_base_per_unit = net_line_total / quantity`.
  - `tax_per_unit = total_tax_amount / quantity`.
  - `refund_per_unit = net_base_per_unit + tax_per_unit`.
  - For a cancelled quantity `q`: `refund_amount = refund_per_unit * q`.

#### Cancellation record & sale total

- `cancellation.cancelled_amount` (in history):
  - Sum of all per-line `refund_amount` values for that cancellation.
- `sale.total_amount` after a partial cancellation:
  - Updated to `old_total_amount - cancelled_amount`.
  - Always reflects the **remaining final amount** after all discounts, taxes, and cancellations.

#### Frontend expectations

- For **full cancellation**:
  - Treat the effective refund as the original `sale.total_amount` at the time of cancellation.
  - Per-line refund breakdown is visible via the cancellation history.
- For **partial cancellation**:
  - Use `CancelItemsResponse`:
    - `new_sale_total`: Updated final amount after cancellation.
    - `items_cancelled[*].amount_refunded`: Per-line refunded amount (matches backend formulas above).
    - `financial_adjustments` for proportional discount and tax reversals.


