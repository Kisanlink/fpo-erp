# Frontend Invoice Setup Guide

Complete guide for frontend developers to configure and use the invoice generation system.

---

## Table of Contentss

1. [Quick Start](#quick-start)
2. [API Endpoints](#api-endpoints)
3. [Step-by-Step Setup](#step-by-step-setup)
4. [Data Structures](#data-structures)
5. [Header Fields Configuration](#header-fields-configuration)
6. [Invoice Generation](#invoice-generation)
7. [Troubleshooting](#troubleshooting)

---

## Quick Start

To enable invoice generation, you need to:

1. **Upload a logo** (optional but recommended)
2. **Configure basic settings** (FPO name, addresses, bank details)
3. **Set up header fields** (GSTIN, licenses)
4. **Check requirements** before generating invoices
5. **Generate invoice** for any sale

**Minimum Required Settings:**
- `fpo_name` - Company name
- `fpo_logo_url` - Logo attachment ID or URL
- `fpo_branch_address` - Branch office address (JSON)

---

## API Endpoints

### Base URL
All endpoints are under `/api/v1`

### Authentication
All endpoints require Bearer token authentication:
```
Authorization: Bearer <your_jwt_token>
```

---

### 1. Settings API

#### 1.1 Get All Settings
```http
GET /api/v1/settings
```

**Response:**
```json
{
  "success": true,
  "message": "Settings retrieved successfully",
  "data": [
    {
      "key": "fpo_name",
      "value": "AGROS FARMER PRODUCER ORGANIZATION LIMITED",
      "display_label": null,
      "display_order": 0,
      "is_header_field": false,
      "created_at": "2025-01-15T10:30:00Z",
      "updated_at": "2025-01-15T10:30:00Z"
    },
    {
      "key": "fpo_gstin",
      "value": "36AAWCA6575D1Z7",
      "display_label": "GSTIN",
      "display_order": 0,
      "is_header_field": true,
      "created_at": "2025-01-15T10:30:00Z",
      "updated_at": "2025-01-15T10:30:00Z"
    }
  ]
}
```

#### 1.2 Get Single Setting
```http
GET /api/v1/settings/{key}
```

**Example:**
```http
GET /api/v1/settings/fpo_name
```

**Response:**
```json
{
  "success": true,
  "message": "Setting retrieved successfully",
  "data": {
    "key": "fpo_name",
    "value": "AGROS FARMER PRODUCER ORGANIZATION LIMITED",
    "display_label": null,
    "display_order": 0,
    "is_header_field": false,
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  }
}
```

#### 1.3 Create/Update Setting (Upsert)
```http
PUT /api/v1/settings/{key}
Content-Type: application/json
```

**Request Body:**
```json
{
  "value": "Your FPO Name",
  "display_label": "Optional Label",
  "display_order": 0,
  "is_header_field": false
}
```

**Example - Set FPO Name:**
```http
PUT /api/v1/settings/fpo_name
Content-Type: application/json

{
  "value": "AGROS FARMER PRODUCER ORGANIZATION LIMITED"
}
```

**Example - Set Branch Address (JSON):**
```http
PUT /api/v1/settings/fpo_branch_address
Content-Type: application/json

{
  "value": "{\"street\":\"Main Street\",\"city\":\"Hyderabad\",\"state\":\"Telangana\",\"pincode\":\"500001\",\"country\":\"India\"}"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Setting saved successfully",
  "data": {
    "key": "fpo_name",
    "value": "AGROS FARMER PRODUCER ORGANIZATION LIMITED",
    "display_label": null,
    "display_order": 0,
    "is_header_field": false,
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  }
}
```

#### 1.4 Get Header Fields
```http
GET /api/v1/settings/header-fields
```

**Response:**
```json
{
  "success": true,
  "message": "Header fields retrieved successfully",
  "data": [
    {
      "key": "fpo_gstin",
      "value": "36AAWCA6575D1Z7",
      "display_label": "GSTIN",
      "display_order": 0
    },
    {
      "key": "fpo_fert_license",
      "value": "FERT-W-250132",
      "display_label": "Fert. Lic.",
      "display_order": 1
    },
    {
      "key": "fpo_pest_seeds_license",
      "value": "PP-N-250251/1141046  Seeds Lic.: SEED-N-250209/1141112",
      "display_label": "Pest. Lic./Seeds Lic.",
      "display_order": 2
    }
  ]
}
```

#### 1.5 Check Invoice Requirements
```http
GET /api/v1/settings/invoice-requirementsimage.png
```

**Response (Ready):**
```json
{
  "success": true,
  "message": "All invoice requirements satisfied",
  "data": {
    "ready": true,
    "missing_settings": []
  }
}
```

**Response (Not Ready):**
```json
{
  "success": true,
  "message": "Missing required settings for invoice generation",
  "data": {
    "ready": false,
    "missing_settings": ["fpo_name", "fpo_logo_url", "fpo_branch_address"]
  }
}
```

#### 1.6 Delete Setting
```http
DELETE /api/v1/settings/{key}
```

**Response:**
```json
{
  "success": true,
  "message": "Setting deleted successfully",
  "data": null
}
```

---

### 2. Invoice API

#### 2.1 Check Invoice Requirements
```http
GET /api/v1/sales/invoice-requirements
```

**Response:** Same as Settings API endpoint above.

#### 2.2 Download Invoice PDF
```http
GET /api/v1/sales/{sale_id}/invoice
Accept: application/pdf
```

**Example:**
```http
GET /api/v1/sales/SALE00000001/invoice
```

**Response:**
- Content-Type: `application/pdf`
- Content-Disposition: `attachment; filename=Invoice_INV001.pdf`
- Binary PDF data

**Frontend Implementation (JavaScript):**
```javascript
async function downloadInvoice(saleId) {
  const response = await fetch(`/api/v1/sales/${saleId}/invoice`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Accept': 'application/pdf'
    }
  });

  if (!response.ok) {
    throw new Error('Failed to generate invoice');
  }

  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `Invoice_${saleId}.pdf`;
  document.body.appendChild(a);
  a.click();
  window.URL.revokeObjectURL(url);
  document.body.removeChild(a);
}
```

---

## Step-by-Step Setup

### Step 1: Upload Logo (Optional but Recommended)

First, upload your FPO logo using the Attachments API, then save the attachment ID.

**Example:**
```javascript
// 1. Upload logo file
const formData = new FormData();
formData.append('file', logoFile);
const uploadResponse = await fetch('/api/v1/attachments', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` },
  body: formData
});
const { data: attachment } = await uploadResponse.json();

// 2. Save logo attachment ID to settings
await fetch(`/api/v1/settings/fpo_logo_url`, {
  method: 'PUT',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    value: attachment.id // e.g., "ATCH00000001"
  })
});
```

### Step 2: Configure Basic Settings

#### 2.1 Set FPO Name
```javascript
await fetch('/api/v1/settings/fpo_name', {
  method: 'PUT',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    value: 'AGROS FARMER PRODUCER ORGANIZATION LIMITED'
  })
});
```

#### 2.2 Set Branch Address
```javascript
const branchAddress = {
  street: 'Main Street',
  city: 'Hyderabad',
  state: 'Telangana',
  pincode: '500001',
  country: 'India'
};

await fetch('/api/v1/settings/fpo_branch_address', {
  method: 'PUT',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    value: JSON.stringify(branchAddress)
  })
});
```

#### 2.3 Set Registered Address (Optional)
```javascript
const registeredAddress = {
  street: 'Registered Office Street',
  city: 'Hyderabad',
  state: 'Telangana',
  pincode: '500001',
  country: 'India'
};

await fetch('/api/v1/settings/fpo_registered_address', {
  method: 'PUT',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    value: JSON.stringify(registeredAddress)
  })
});
```

#### 2.4 Set Bank Account Details
```javascript
const bankDetails = {
  account_name: 'AGROS FPO',
  account_number: '1234567890',
  ifsc_code: 'BANK0001234',
  bank_name: 'Bank Name',
  branch: 'Branch Name'
};

await fetch('/api/v1/settings/fpo_bank_account', {
  method: 'PUT',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    value: JSON.stringify(bankDetails)
  })
});
```

### Step 3: Configure Header Fields

Header fields appear in the invoice details section (right column). These should be configured with `is_header_field: true`.

#### 3.1 Set GSTIN
```javascript
await fetch('/api/v1/settings/fpo_gstin', {
  method: 'PUT',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    value: '36AAWCA6575D1Z7',
    display_label: 'GSTIN',
    display_order: 0,
    is_header_field: true
  })
});
```

#### 3.2 Set Fertilizer License
```javascript
await fetch('/api/v1/settings/fpo_fert_license', {
  method: 'PUT',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    value: 'FERT-W-250132',
    display_label: 'Fert. Lic.',
    display_order: 1,
    is_header_field: true
  })
});
```

#### 3.3 Set Pest/Seeds License
```javascript
await fetch('/api/v1/settings/fpo_pest_seeds_license', {
  method: 'PUT',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    value: 'PP-N-250251/1141046  Seeds Lic.: SEED-N-250209/1141112',
    display_label: 'Pest. Lic./Seeds Lic.',
    display_order: 2,
    is_header_field: true
  })
});
```

### Step 4: Verify Setup

```javascript
const response = await fetch('/api/v1/settings/invoice-requirements', {
  headers: { 'Authorization': `Bearer ${token}` }
});
const { data } = await response.json();

if (data.ready) {
  console.log('✅ All invoice requirements satisfied');
} else {
  console.log('❌ Missing settings:', data.missing_settings);
}
```

### Step 5: Generate Invoice

```javascript
async function generateInvoice(saleId) {
  // Check requirements first
  const checkResponse = await fetch('/api/v1/sales/invoice-requirements', {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  const { data: requirements } = await checkResponse.json();

  if (!requirements.ready) {
    alert(`Missing settings: ${requirements.missing_settings.join(', ')}`);
    return;
  }

  // Download invoice
  const response = await fetch(`/api/v1/sales/${saleId}/invoice`, {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Accept': 'application/pdf'
    }
  });

  if (!response.ok) {
    throw new Error('Failed to generate invoice');
  }

  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `Invoice_${saleId}.pdf`;
  a.click();
  window.URL.revokeObjectURL(url);
}
```

---

## Data Structures

### Setting Keys Reference

| Key | Type | Required | Description | Example |
|-----|------|----------|------------|---------|
| `fpo_name` | string | ✅ Yes | FPO/Company name | `"AGROS FPO"` |
| `fpo_logo_url` | string | ✅ Yes | Logo attachment ID or URL | `"ATCH00000001"` |
| `fpo_branch_address` | JSON string | ✅ Yes | Branch office address | See below |
| `fpo_registered_address` | JSON string | ❌ No | Registered office address | See below |
| `fpo_bank_account` | JSON string | ❌ No | Bank account details | See below |
| `fpo_gstin` | string | ❌ No | GSTIN (header field) | `"36AAWCA6575D1Z7"` |
| `fpo_fert_license` | string | ❌ No | Fertilizer license (header field) | `"FERT-W-250132"` |
| `fpo_pest_seeds_license` | string | ❌ No | Pest/Seeds license (header field) | `"PP-N-250251/1141046"` |

### Address JSON Structure

```json
{
  "street": "Main Street",
  "city": "Hyderabad",
  "state": "Telangana",
  "pincode": "500001",
  "country": "India"
}
```

**Important:** When saving to settings, you must stringify the JSON:
```javascript
const address = { street: "...", city: "...", ... };
const value = JSON.stringify(address); // Convert to string
```

### Bank Account JSON Structure

```json
{
  "account_name": "AGROS FPO",
  "account_number": "1234567890",
  "ifsc_code": "BANK0001234",
  "bank_name": "Bank Name",
  "branch": "Branch Name"
}
```

**Important:** When saving to settings, you must stringify the JSON:
```javascript
const bankDetails = { account_name: "...", ... };
const value = JSON.stringify(bankDetails); // Convert to string
```

### Header Field Structure

Header fields are settings with `is_header_field: true`. They appear in the invoice details section (right column) in the order specified by `display_order`.

**Required fields for header field:**
- `value` - The actual value to display
- `display_label` - Label shown before the value (e.g., "GSTIN:", "Fert. Lic.:")
- `display_order` - Order in which fields appear (0, 1, 2, ...)
- `is_header_field` - Must be `true`

---

## Header Fields Configuration

Header fields are displayed in the invoice details section (right column) alongside:
- Row 1: Invoice No | **GSTIN** (display_order: 0)
- Row 2: Invoice Date | **Fert. Lic.** (display_order: 1)
- Row 3: State | **Pest. Lic./Seeds Lic.** (display_order: 2)

### Creating Header Fields

```javascript
// Example: Create GSTIN header field
async function createHeaderField(key, value, label, order) {
  await fetch(`/api/v1/settings/${key}`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      value: value,
      display_label: label,
      display_order: order,
      is_header_field: true
    })
  });
}

// Usage
await createHeaderField('fpo_gstin', '36AAWCA6575D1Z7', 'GSTIN', 0);
await createHeaderField('fpo_fert_license', 'FERT-W-250132', 'Fert. Lic.', 1);
await createHeaderField('fpo_pest_seeds_license', 'PP-N-250251/1141046', 'Pest. Lic./Seeds Lic.', 2);
```

### Retrieving Header Fields

```javascript
const response = await fetch('/api/v1/settings/header-fields', {
  headers: { 'Authorization': `Bearer ${token}` }
});
const { data: headerFields } = await response.json();

// headerFields is an array sorted by display_order
headerFields.forEach(field => {
  console.log(`${field.display_label}: ${field.value}`);
});
```

---

## Invoice Generation

### Pre-flight Check

Always check invoice requirements before attempting to generate:

```javascript
async function canGenerateInvoice() {
  const response = await fetch('/api/v1/sales/invoice-requirements', {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  const { data } = await response.json();
  return data.ready;
}
```

### Generate and Download

```javascript
async function downloadInvoicePDF(saleId) {
  try {
    // Check requirements
    const canGenerate = await canGenerateInvoice();
    if (!canGenerate) {
      throw new Error('Invoice requirements not met');
    }

    // Generate and download
    const response = await fetch(`/api/v1/sales/${saleId}/invoice`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Accept': 'application/pdf'
      }
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to generate invoice');
    }

    // Get filename from Content-Disposition header
    const contentDisposition = response.headers.get('Content-Disposition');
    const filename = contentDisposition
      ? contentDisposition.split('filename=')[1].replace(/"/g, '')
      : `Invoice_${saleId}.pdf`;

    // Create blob and download
    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);
  } catch (error) {
    console.error('Error generating invoice:', error);
    alert(`Failed to generate invoice: ${error.message}`);
  }
}
```

### Generate and Display in Browser

```javascript
async function viewInvoicePDF(saleId) {
  const response = await fetch(`/api/v1/sales/${saleId}/invoice`, {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Accept': 'application/pdf'
    }
  });

  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  
  // Open in new tab
  window.open(url, '_blank');
  
  // Or embed in iframe
  // const iframe = document.createElement('iframe');
  // iframe.src = url;
  // document.body.appendChild(iframe);
}
```

---

## Troubleshooting

### Issue: "Missing required settings" error

**Solution:**
1. Check which settings are missing:
   ```javascript
   const response = await fetch('/api/v1/settings/invoice-requirements');
   const { data } = await response.json();
   console.log('Missing:', data.missing_settings);
   ```

2. Ensure all required settings are configured:
   - `fpo_name`
   - `fpo_logo_url`
   - `fpo_branch_address`

### Issue: Logo not appearing in invoice

**Possible causes:**
1. Logo attachment ID is incorrect
2. Logo file was deleted
3. Logo URL is invalid

**Solution:**
1. Verify logo setting:
   ```javascript
   const response = await fetch('/api/v1/settings/fpo_logo_url');
   const { data } = await response.json();
   console.log('Logo setting:', data.value);
   ```

2. Ensure logo attachment ID starts with `ATCH_` or is a valid URL

### Issue: JSON parsing error for addresses/bank details

**Cause:** JSON was not stringified before saving

**Solution:**
```javascript
// ❌ Wrong
const address = { street: "...", city: "..." };
await saveSetting('fpo_branch_address', address);

// ✅ Correct
const address = { street: "...", city: "..." };
await saveSetting('fpo_branch_address', JSON.stringify(address));
```

### Issue: Header fields not appearing in correct order

**Solution:**
Ensure `display_order` is set correctly:
- GSTIN: `display_order: 0`
- Fert. Lic.: `display_order: 1`
- Pest/Seeds Lic.: `display_order: 2`

### Issue: Invoice PDF download fails

**Check:**
1. Sale ID is valid
2. Sale exists and has items
3. All required settings are configured
4. User has proper permissions

**Debug:**
```javascript
try {
  const response = await fetch(`/api/v1/sales/${saleId}/invoice`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  
  if (!response.ok) {
    const error = await response.json();
    console.error('Error:', error);
  }
} catch (error) {
  console.error('Network error:', error);
}
```

---

## Complete Setup Example

```javascript
// Complete invoice setup function
async function setupInvoiceSystem(config) {
  const { token, logoFile } = config;

  // 1. Upload logo
  const formData = new FormData();
  formData.append('file', logoFile);
  const uploadRes = await fetch('/api/v1/attachments', {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` },
    body: formData
  });
  const { data: attachment } = await uploadRes.json();

  // 2. Configure basic settings
  const settings = [
    { key: 'fpo_name', value: config.fpoName },
    { key: 'fpo_logo_url', value: attachment.id },
    { key: 'fpo_branch_address', value: JSON.stringify(config.branchAddress) },
    { key: 'fpo_registered_address', value: JSON.stringify(config.registeredAddress) },
    { key: 'fpo_bank_account', value: JSON.stringify(config.bankDetails) }
  ];

  for (const setting of settings) {
    await fetch(`/api/v1/settings/${setting.key}`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ value: setting.value })
    });
  }

  // 3. Configure header fields
  const headerFields = [
    { key: 'fpo_gstin', value: config.gstin, label: 'GSTIN', order: 0 },
    { key: 'fpo_fert_license', value: config.fertLicense, label: 'Fert. Lic.', order: 1 },
    { key: 'fpo_pest_seeds_license', value: config.pestSeedsLicense, label: 'Pest. Lic./Seeds Lic.', order: 2 }
  ];

  for (const field of headerFields) {
    await fetch(`/api/v1/settings/${field.key}`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        value: field.value,
        display_label: field.label,
        display_order: field.order,
        is_header_field: true
      })
    });
  }

  // 4. Verify setup
  const checkRes = await fetch('/api/v1/settings/invoice-requirements', {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  const { data } = await checkRes.json();

  if (data.ready) {
    console.log('✅ Invoice system configured successfully!');
  } else {
    console.error('❌ Missing settings:', data.missing_settings);
  }
}

// Usage
await setupInvoiceSystem({
  token: 'your_jwt_token',
  logoFile: logoFileObject,
  fpoName: 'AGROS FARMER PRODUCER ORGANIZATION LIMITED',
  branchAddress: {
    street: 'Main Street',
    city: 'Hyderabad',
    state: 'Telangana',
    pincode: '500001',
    country: 'India'
  },
  registeredAddress: { /* ... */ },
  bankDetails: { /* ... */ },
  gstin: '36AAWCA6575D1Z7',
  fertLicense: 'FERT-W-250132',
  pestSeedsLicense: 'PP-N-250251/1141046  Seeds Lic.: SEED-N-250209/1141112'
});
```

---

## Support

For issues or questions:
1. Check the troubleshooting section above
2. Verify all required settings are configured
3. Check API response errors for specific details
4. Ensure authentication token is valid

---

**Last Updated:** January 2025

