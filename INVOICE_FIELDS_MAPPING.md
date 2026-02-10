# Invoice Fields Mapping

This document lists all fields displayed in the invoice PDF and their data sources.

## Invoice Fields and Their Sources

### 1. Header Section (Top of Invoice)

| Field | Source | Setting Key | Notes |
|-------|--------|-------------|-------|
| FPO Name | Settings | `fpo_name` | Required |
| Branch Address | Settings | `fpo_branch_address` | JSON format, required |
| Registered Address | Settings | `fpo_registered_address` | JSON format, optional |
| Logo | Settings | `fpo_logo_url` | Attachment ID or URL, required |

### 2. Invoice Details Section (Two-Column Layout)

**Left Column:**
| Field | Source | Notes |
|-------|--------|-------|
| Invoice No | Sale | `sale.InvoiceNumber` |
| Invoice Date | Sale | `sale.SaleDate` formatted as DD/MM/YYYY |
| State | Warehouse | `warehouse.State` (falls back to "Telangana" if not set) |

**Right Column (Header Fields from Settings):**
| Field | Source | Setting Key | Display Label | Display Order |
|-------|--------|-------------|---------------|---------------|
| GSTIN | Settings (Header Field) | `fpo_gstin` | "GSTIN" | 0 |
| Fert. Lic. | Settings (Header Field) | Custom key | "Fert. Lic." | 1 |
| Pest. Lic. / Seeds Lic. | Settings (Header Field) | Custom key | "Pest. Lic." / "Seeds Lic." | 2 |

**Note:** Header fields are configured in the `settings` table with:
- `is_header_field = true`
- `display_label` set to the label (e.g., "GSTIN", "Fert. Lic.")
- `display_order` set to control the order (0, 1, 2, etc.)

### 3. Receiver Section (Two-Column Layout)

**Left Column - "Details of Receiver (Billed To)":**
| Field | Source | Notes |
|-------|--------|-------|
| Customer Name | Sale | `sale.CustomerName` (defaults to "Walk-in Customer") |
| Phone | Sale | `sale.CustomerPhone` (defaults to "N/A") |
| Customer Type | Sale | "Member" or "Non-Member" based on `sale.IsOrgMember` |

**Right Column - "Place of Supply / Shipped To":**
| Field | Source | Notes |
|-------|--------|-------|
| Place of Supply | Warehouse | `warehouse.State` (falls back to "Telangana") |
| Ship To Name | Sale | Same as customer name |
| Ship To Address | Sale | Currently empty (can be extended) |
| Ship To GSTIN | Sale | Currently empty (can be extended) |

### 4. Line Items Table

| Field | Source | Notes |
|-------|--------|-------|
| S.No | Calculated | Sequential number |
| Item Name | Product + Variant | `product.Name + variant.Quantity` |
| HSN Code | Variant | `variant.HSNCode` |
| Units | Variant | `variant.PackSize` |
| Qty | Sale Item | `saleItem.Quantity` |
| Excl. Rate | Sale Item | `saleItem.SellingPrice` |
| Net Value | Calculated | `Excl. Rate × Qty` |
| GST% | Variant | `variant.GSTRate` |
| CGST | Sale Item | `saleItem.CGSTAmount` |
| SGST | Sale Item | `saleItem.SGSTAmount` |
| Total Value | Calculated | `Net Value + CGST + SGST` |

### 5. Summary Section

| Field | Source | Notes |
|-------|--------|-------|
| Grand Total (in words) | Sale | `sale.TotalAmount` converted to words |
| Total Amount Before Tax | Calculated | Sum of all `NetValue` |
| Add: CGST | Calculated | Sum of all `CGSTAmount` |
| Add: SGST | Calculated | Sum of all `SGSTAmount` |
| Total Amount | Sale | `sale.TotalAmount` |

### 6. Footer Section

**Left Side - "Virtual Payment Details":**
| Field | Source | Setting Key | Notes |
|-------|--------|-------------|-------|
| Account Name | Settings | `fpo_bank_account` (JSON) | `account_name` field |
| Account Number | Settings | `fpo_bank_account` (JSON) | `account_number` field |
| IFSC Code | Settings | `fpo_bank_account` (JSON) | `ifsc_code` field |
| Branch | Settings | `fpo_bank_account` (JSON) | `branch` field |

**Right Side:**
| Field | Source | Notes |
|-------|--------|-------|
| Authorized Signatory | Static | "For {FPO Name}" + "Authorised Signatory" |

**Terms and Conditions:**
- Hardcoded in the invoice service (6 standard terms)

## Required Settings Configuration

### Basic Settings (Required)
1. `fpo_name` - FPO/Company name
2. `fpo_logo_url` - Logo attachment ID (starts with "ATCH_") or URL
3. `fpo_branch_address` - JSON: `{"street": "...", "city": "...", "state": "...", "pincode": "...", "country": "..."}`
4. `fpo_bank_account` - JSON: `{"account_name": "...", "account_number": "...", "ifsc_code": "...", "bank_name": "...", "branch": "..."}`

### Header Fields (Required for Invoice Details Right Column)
Configure these as header fields (`is_header_field = true`):

1. **GSTIN** (Display Order: 0)
   - Key: `fpo_gstin` (or custom)
   - Display Label: `GSTIN`
   - Value: e.g., `36AAWCA6575D1Z7`

2. **Fertilizer License** (Display Order: 1)
   - Key: e.g., `fpo_fert_license`
   - Display Label: `Fert. Lic.`
   - Value: e.g., `FERT-W-250132`

3. **Pest/Seeds License** (Display Order: 2)
   - Key: e.g., `fpo_pest_seeds_license`
   - Display Label: `Pest. Lic.` or `Pest. Lic./Seeds Lic.`
   - Value: e.g., `PP-N-250251/1141046` or `PP-N-250251/1141046  Seeds Lic.: SEED-N-250209/1141112`

### Optional Settings
- `fpo_registered_address` - JSON format (same structure as branch address)

## Example Settings Configuration

```json
{
  "fpo_name": "AGROS FARMER PRODUCER ORGANIZATION LIMITED",
  "fpo_logo_url": "ATCH_1234567890",
  "fpo_branch_address": "{\"street\":\"Main Street\",\"city\":\"Hyderabad\",\"state\":\"Telangana\",\"pincode\":\"500001\",\"country\":\"India\"}",
  "fpo_registered_address": "{\"street\":\"Reg Street\",\"city\":\"Hyderabad\",\"state\":\"Telangana\",\"pincode\":\"500001\",\"country\":\"India\"}",
  "fpo_bank_account": "{\"account_name\":\"AGROS FPO\",\"account_number\":\"1234567890\",\"ifsc_code\":\"BANK0001234\",\"bank_name\":\"Bank Name\",\"branch\":\"Branch Name\"}"
}
```

### Header Fields Example (via API)
```json
{
  "key": "fpo_gstin",
  "value": "36AAWCA6575D1Z7",
  "display_label": "GSTIN",
  "display_order": 0,
  "is_header_field": true
}
```

## Notes

1. **State Field**: Currently uses `warehouse.State`. If warehouse doesn't have state, defaults to "Telangana". Can be overridden via settings if needed.

2. **Customer Address & GSTIN**: Currently not stored in the Sale model. These fields are prepared in the code but will be empty until added to the sale model or linked to a collaborator.

3. **Header Fields**: The system uses a flexible header field mechanism where any setting with `is_header_field = true` will be displayed in the invoice details right column, ordered by `display_order`.

4. **Terms and Conditions**: Currently hardcoded. Can be moved to settings if needed for customization.

