# Frontend API Guide: Invoice & Attachments

This document provides complete API reference for implementing invoice generation with logo support in the frontend.

---

## Quick Overview

To generate invoices with logo, you need to:
1. **Upload a logo** using the Attachments API
2. **Save the logo attachment ID** to Settings API with key `fpo_logo`
3. **Configure other required settings** (FPO name, addresses, bank details)
4. **Generate invoice** by calling the Invoice API

---

## API Endpoints Summary

| API | Endpoints | Purpose |
|-----|-----------|---------|
| **Settings** | 6 endpoints | Store FPO configuration (name, logo ID, addresses, bank) |
| **Attachments** | 8 endpoints | Upload/download files, get presigned URLs |
| **Invoice** | 2 endpoints | Check requirements, download PDF |

---

## 1. Settings API

Base URL: `/api/v1/settings`

Settings are stored as key-value pairs. The invoice feature requires these keys:
- `fpo_name` - Company name shown on invoice header
- `fpo_logo_attachment_id` - Attachment ID of the uploaded logo (e.g., `ATCH00000001`)
- `fpo_branch_address` - Branch office address (JSON)
- `fpo_registered_address` - Registered office address (JSON)
- `fpo_bank_details` - Bank account information (JSON)

### 1.1 Create/Update Setting (Upsert)

```http
PUT /api/v1/settings/fpo_name
Content-Type: application/json
Authorization: Bearer <token>

{
  "value": "Kisanlink Farmer Producer Organization"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Setting saved successfully",
  "data": {
    "id": "STNG00000001",
    "key": "fpo_name",
    "value": "Kisanlink Farmer Producer Organization",
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  }
}
```

### 1.2 Get Setting by Key

```http
GET /api/v1/settings/fpo_name
Authorization: Bearer <token>
```

**Response:**
```json
{
  "success": true,
  "message": "Setting retrieved successfully",
  "data": {
    "id": "STNG00000001",
    "key": "fpo_name",
    "value": "Kisanlink Farmer Producer Organization",
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  }
}
```

### 1.3 Get All Settings

```http
GET /api/v1/settings
Authorization: Bearer <token>
```

### 1.4 Get Header Fields

```http
GET /api/v1/settings/header-fields
Authorization: Bearer <token>
```

**Response:**
```json
{
  "success": true,
  "message": "Header fields retrieved successfully",
  "data": [
    {
      "key": "gstin",
      "label": "GSTIN",
      "value": "36AABCU9603R1ZM",
      "display_order": 1
    }
  ]
}
```

### 1.5 Delete Setting

```http
DELETE /api/v1/settings/fpo_name
Authorization: Bearer <token>
```

---

## 2. Attachments API

Base URL: `/api/v1/attachments`

### 2.1 Upload File (Logo)

```http
POST /api/v1/attachments
Content-Type: multipart/form-data
Authorization: Bearer <token>

Form Fields:
- file: <binary file data>
- entity_type: "logo"
- entity_id: "fpo"
```

**JavaScript Example:**
```javascript
async function uploadLogo(file) {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('entity_type', 'logo');
  formData.append('entity_id', 'fpo');

  const response = await fetch('/api/v1/attachments', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    },
    body: formData
  });

  const result = await response.json();
  // result.data.id contains the attachment ID (e.g., "ATCH00000001")
  return result.data.id;
}
```

**Response:**
```json
{
  "success": true,
  "message": "Attachment uploaded successfully",
  "data": {
    "id": "ATCH00000001",
    "entity_type": "logo",
    "entity_id": "fpo",
    "file_path": "logos/uuid-filename.png",
    "file_type": "image/png",
    "uploaded_by": "USER00000001",
    "uploaded_at": "2025-01-15T10:30:00Z",
    "download_url": "https://s3.amazonaws.com/bucket/logos/uuid-filename.png?X-Amz-..."
  }
}
```

### 2.2 Get Presigned URL (for displaying images)

```http
GET /api/v1/attachments/ATCH00000001/url
Authorization: Bearer <token>
```

**Response:**
```json
{
  "success": true,
  "message": "Download URL generated successfully",
  "data": {
    "download_url": "https://s3.amazonaws.com/bucket/logos/uuid-filename.png?X-Amz-...",
    "expires_in": 3600
  }
}
```

**JavaScript Example:**
```javascript
async function getLogoUrl(attachmentId) {
  const response = await fetch(`/api/v1/attachments/${attachmentId}/url`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  const result = await response.json();
  return result.data.download_url;
}
```

### 2.3 Get Attachment Metadata

```http
GET /api/v1/attachments/ATCH00000001
Authorization: Bearer <token>
```

### 2.4 Get Attachment Info (with file size)

```http
GET /api/v1/attachments/ATCH00000001/info
Authorization: Bearer <token>
```

### 2.5 Download File

```http
GET /api/v1/attachments/ATCH00000001/download
Authorization: Bearer <token>
```

### 2.6 Get Attachments by Entity

```http
GET /api/v1/attachments/entity/logo/fpo
Authorization: Bearer <token>
```

### 2.7 List All Attachments

```http
GET /api/v1/attachments?entity_type=logo&limit=10&offset=0
Authorization: Bearer <token>
```

### 2.8 Delete Attachment

```http
DELETE /api/v1/attachments/ATCH00000001
Authorization: Bearer <token>
```

---

## 3. Invoice API

Base URL: `/api/v1/sales`

### 3.1 Check Invoice Requirements

**Call this before showing the "Download Invoice" button to verify setup is complete.**

```http
GET /api/v1/sales/invoice-requirements
Authorization: Bearer <token>
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
    "missing_settings": ["fpo_name", "fpo_logo_attachment_id", "fpo_bank_details"]
  }
}
```

### 3.2 Download Invoice PDF

```http
GET /api/v1/sales/SALE00000001/invoice
Authorization: Bearer <token>
```

**Response:** Binary PDF file with headers:
- `Content-Type: application/pdf`
- `Content-Disposition: attachment; filename=Invoice_12250001.pdf`

**JavaScript Example:**
```javascript
async function downloadInvoice(saleId) {
  const response = await fetch(`/api/v1/sales/${saleId}/invoice`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message);
  }

  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = response.headers.get('Content-Disposition')?.split('filename=')[1] || 'invoice.pdf';
  a.click();
  window.URL.revokeObjectURL(url);
}
```

---

## 4. Complete Setup Flow

### Step 1: Upload Logo

```javascript
// 1. User selects logo file
const logoFile = document.getElementById('logoInput').files[0];

// 2. Upload to attachments API
const formData = new FormData();
formData.append('file', logoFile);
formData.append('entity_type', 'logo');
formData.append('entity_id', 'fpo');

const uploadResponse = await fetch('/api/v1/attachments', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` },
  body: formData
});

const uploadResult = await uploadResponse.json();
const logoAttachmentId = uploadResult.data.id; // e.g., "ATCH00000001"
```

### Step 2: Save Settings

```javascript
// Save all required settings using PUT (upsert)
const settings = [
  { key: 'fpo_name', value: 'Kisanlink FPO' },
  { key: 'fpo_logo_attachment_id', value: logoAttachmentId }, // From step 1
  {
    key: 'fpo_branch_address',
    value: JSON.stringify({
      type: 'branch',
      line1: '123 Main Street',
      line2: 'Floor 2',
      city: 'Hyderabad',
      district: 'Hyderabad',
      state: 'Telangana',
      pincode: '500001',
      country: 'India'
    })
  },
  {
    key: 'fpo_registered_address',
    value: JSON.stringify({
      type: 'registered',
      line1: '456 Business Park',
      city: 'Hyderabad',
      district: 'Hyderabad',
      state: 'Telangana',
      pincode: '500002',
      country: 'India'
    })
  },
  {
    key: 'fpo_bank_details',
    value: JSON.stringify({
      bank_name: 'State Bank of India',
      account_number: '1234567890',
      ifsc_code: 'SBIN0001234',
      branch_name: 'Hyderabad Main Branch',
      account_type: 'Current'
    })
  }
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
```

### Step 3: Verify Setup

```javascript
const checkResponse = await fetch('/api/v1/sales/invoice-requirements', {
  headers: { 'Authorization': `Bearer ${token}` }
});
const checkResult = await checkResponse.json();

if (checkResult.data.ready) {
  console.log('Invoice setup complete!');
} else {
  console.log('Missing settings:', checkResult.data.missing_settings);
}
```

### Step 4: Generate Invoice

```javascript
// On sales detail page, show download button if ready
if (invoiceReady) {
  downloadButton.onclick = () => downloadInvoice(saleId);
}
```

---

## 5. JSON Data Structures

### Address JSON

```json
{
  "type": "branch",
  "line1": "123 Main Street",
  "line2": "Floor 2, Unit 5",
  "city": "Hyderabad",
  "district": "Hyderabad",
  "state": "Telangana",
  "pincode": "500001",
  "country": "India"
}
```

### Bank Details JSON

```json
{
  "bank_name": "State Bank of India",
  "account_number": "1234567890123",
  "ifsc_code": "SBIN0001234",
  "branch_name": "Hyderabad Main Branch",
  "account_type": "Current"
}
```

---

## 6. UI Implementation Guide

### Settings Page Layout

```
+------------------------------------------+
|  FPO Settings                            |
+------------------------------------------+
|                                          |
|  Company Logo:                           |
|  +----------+  [Upload New Logo]         |
|  |  LOGO    |                            |
|  |  IMAGE   |  Current: ATCH00000001     |
|  +----------+                            |
|                                          |
|  Company Name:                           |
|  [Kisanlink FPO                    ]     |
|                                          |
|  Branch Address:                         |
|  [123 Main Street                  ]     |
|  [Floor 2                          ]     |
|  [Hyderabad    ] [Telangana  ] [500001]  |
|                                          |
|  Registered Address:                     |
|  [456 Business Park                ]     |
|  [                                 ]     |
|  [Hyderabad    ] [Telangana  ] [500002]  |
|                                          |
|  Bank Details:                           |
|  Bank: [State Bank of India        ]     |
|  A/C:  [1234567890123              ]     |
|  IFSC: [SBIN0001234                ]     |
|  Branch: [Hyderabad Main Branch    ]     |
|                                          |
|  [Save Settings]                         |
+------------------------------------------+
```

### Sales Detail Page (Invoice Button)

```
+------------------------------------------+
|  Sale: SALE00000001                      |
+------------------------------------------+
|  Date: 15 Jan 2025                       |
|  Warehouse: Main Warehouse               |
|  Total: Rs. 1,500.00                     |
|                                          |
|  Items:                                  |
|  - Tomato 500g x 10 = Rs. 500.00         |
|  - Onion 1kg x 5 = Rs. 1,000.00          |
|                                          |
|  [Download Invoice]  <-- Only show if    |
|                          requirements met |
+------------------------------------------+
```

---

## 7. Error Handling

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| 400 "Missing required settings" | Settings not configured | Check `/invoice-requirements` first |
| 404 "Sale not found" | Invalid sale ID | Verify sale ID exists |
| 401 "Unauthorized" | Missing/invalid token | Re-authenticate user |
| 500 "Failed to generate invoice" | Server error | Check server logs |

### Error Response Format

```json
{
  "success": false,
  "message": "Missing required settings for invoice generation",
  "error": {
    "code": "MISSING_SETTINGS",
    "details": ["fpo_name", "fpo_logo"]
  }
}
```

---

## 8. Notes

1. **Logo Support**: Logo appears in PDF header (top-left corner, 25mm width)
2. **Invoice Number**: Generated automatically using format `MMYYNNNN` (e.g., `12250001` for December 2025, invoice #1)
3. **Presigned URLs**: Expire after 1 hour by default (configurable via `expiration` query param in seconds, max 86400)
4. **File Types**: Supported logo formats: PNG, JPG, JPEG, GIF (max 10MB)
5. **State in Invoice**: Uses state from sale's warehouse address
6. **Net Value Calculation**: Rate × Quantity (before tax)

---

## 9. API Response Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created (new resource) |
| 400 | Bad Request (validation error) |
| 401 | Unauthorized (invalid/missing token) |
| 404 | Not Found (resource doesn't exist) |
| 500 | Internal Server Error |

---

## Contact

For backend API issues or questions, contact the backend team.
