# Attachments Flow Documentation

## Collaborator Logo Upload Flow

### Overview
Creating a collaborator with a logo is a **two-step process**:
1. First, create the collaborator (to get the collaborator ID)
2. Then, upload the logo (using the collaborator ID as `entity_id`)

The logo is automatically synced to the collaborator's `logo` field by the backend.

---

## Step-by-Step Flow

### Step 1: Create Collaborator (without logo)

**API:** `POST /api/v1/collaborators`

**Request:**
```json
{
  "company_name": "ABC Suppliers",
  "contact_person": "John Doe",
  "contact_number": "9876543210",
  "email": "john@abc.com",
  "gst_number": "29ABCDE1234F1Z5",
  "pan_number": "ABCDE1234F",
  "bank_name": "State Bank",
  "bank_account_no": "1234567890",
  "bank_ifsc": "SBIN0001234",
  "experience": "5 years in agriculture",
  "address": {
    "type": "business",
    "street": "123 Main Road",
    "district": "Bangalore Urban",
    "state": "Karnataka",
    "pincode": "560001",
    "country": "India"
  }
}
```

**Response:**
```json
{
  "id": "CLAB00000001",           // <-- Save this ID for Step 2
  "company_name": "ABC Suppliers",
  "contact_person": "John Doe",
  "contact_number": "9876543210",
  "email": "john@abc.com",
  "gst_number": "29ABCDE1234F1Z5",
  "pan_number": "ABCDE1234F",
  "bank_name": "State Bank",
  "bank_account_no": "1234567890",
  "bank_ifsc": "SBIN0001234",
  "experience": "5 years in agriculture",
  "logo": null,                   // <-- No logo yet
  "is_active": true,
  "address": {
    "id": "ADDR00000001",
    "type": "business",
    "street": "123 Main Road",
    "district": "Bangalore Urban",
    "state": "Karnataka",
    "pincode": "560001",
    "country": "India"
  },
  "created_at": "2025-12-12T10:00:00Z",
  "updated_at": "2025-12-12T10:00:00Z"
}
```

---

### Step 2: Upload Logo

**API:** `POST /api/v1/attachments`

**Content-Type:** `multipart/form-data`

**Request Form Data:**
| Field | Value | Description |
|-------|-------|-------------|
| `entity_type` | `logo` | Must be exactly `"logo"` (not "logos") |
| `entity_id` | `CLAB00000001` | The collaborator ID from Step 1 |
| `file` | (binary) | The image file (PNG, JPG, etc.) |

**cURL Example:**
```bash
curl -X POST "http://localhost:8080/api/v1/attachments" \
  -H "Authorization: Bearer {token}" \
  -F "entity_type=logo" \
  -F "entity_id=CLAB00000001" \
  -F "file=@/path/to/logo.png"
```

**Response:**
```json
{
  "id": "ATCH00000001",
  "entity_type": "logo",
  "entity_id": "CLAB00000001",
  "file_path": "logos/abc123-uuid.png",
  "file_type": "image/png",
  "download_url": "https://s3.amazonaws.com/bucket/logos/abc123-uuid.png?X-Amz-...",
  "uploaded_by": "USER00000001",
  "uploaded_at": "2025-12-12T10:01:00Z",
  "created_at": "2025-12-12T10:01:00Z",
  "updated_at": "2025-12-12T10:01:00Z"
}
```

**What happens automatically:**
1. File is uploaded to S3 in `logos/` folder
2. Attachment record is created in database
3. Backend generates 1-year presigned URL
4. Backend auto-updates `collaborator.logo` field with the presigned URL

---

### Step 3: Verify (Optional)

**API:** `GET /api/v1/collaborators/{id}`

**Request:**
```
GET /api/v1/collaborators/CLAB00000001
```

**Response:**
```json
{
  "id": "CLAB00000001",
  "company_name": "ABC Suppliers",
  "contact_person": "John Doe",
  "contact_number": "9876543210",
  "email": "john@abc.com",
  "gst_number": "29ABCDE1234F1Z5",
  "logo": "https://s3.amazonaws.com/bucket/logos/abc123-uuid.png?X-Amz-...",  // <-- Logo URL auto-synced!
  "is_active": true,
  ...
}
```

---

## Visual Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     COLLABORATOR + LOGO CREATION FLOW                    │
└─────────────────────────────────────────────────────────────────────────┘

    ┌──────────────────┐
    │  Frontend Form   │
    │  (Company info)  │
    └────────┬─────────┘
             │
             ▼
    ┌──────────────────────────────────────────────────────────────────┐
    │ STEP 1: POST /api/v1/collaborators                                │
    │                                                                   │
    │ Request: { company_name, contact_person, ... }                   │
    │ Response: { id: "CLAB00000001", logo: null, ... }                │
    └────────┬─────────────────────────────────────────────────────────┘
             │
             │ Save collaborator ID
             ▼
    ┌──────────────────────────────────────────────────────────────────┐
    │ STEP 2: POST /api/v1/attachments (multipart/form-data)           │
    │                                                                   │
    │ Form Data:                                                        │
    │   entity_type = "logo"                                            │
    │   entity_id = "CLAB00000001"                                      │
    │   file = [logo.png binary]                                        │
    │                                                                   │
    │ Response: { id: "ATCH00000001", download_url: "https://s3..." }  │
    └────────┬─────────────────────────────────────────────────────────┘
             │
             │ Backend Auto-Sync:
             │ ┌─────────────────────────────────────────────────────┐
             │ │ 1. Upload file to S3 (logos/uuid.png)              │
             │ │ 2. Create attachment record                         │
             │ │ 3. Generate 1-year presigned URL                    │
             │ │ 4. UPDATE collaborators SET logo = {presigned_url}  │
             │ └─────────────────────────────────────────────────────┘
             │
             ▼
    ┌──────────────────────────────────────────────────────────────────┐
    │ DONE: Collaborator now has logo URL                               │
    │                                                                   │
    │ GET /api/v1/collaborators/CLAB00000001                           │
    │ Response: { ..., logo: "https://s3.amazonaws.com/..." }          │
    └──────────────────────────────────────────────────────────────────┘
```

---

## Why Two Steps?

**Problem:** You can't upload a logo for a collaborator that doesn't exist yet.

**Solution:**
1. Create collaborator first -> Get ID
2. Upload logo using that ID -> Backend auto-syncs

**Alternative (Edit Flow):**
For existing collaborators, you only need Step 2:
1. User clicks "Change Logo" on existing collaborator
2. Frontend uploads with `entity_type: "logo"`, `entity_id: {existing_collaborator_id}`
3. Logo is automatically updated

---

## Entity Types Reference

| Entity Type | S3 Folder | Used For | Auto-Sync |
|-------------|-----------|----------|-----------|
| `logo` | `logos/` | Collaborator logos | Yes - updates `collaborator.logo` |
| `variant` | `product-variants/{id}/` | Product variant images | Yes - updates `variant.images` array |
| `po` | `purchase-orders/{id}/` | PO documents | No |
| `grn` | `grns/{id}/` | GRN documents | No |
| `misc` | `misc/` | Other files | No |

---

## Frontend Implementation Notes

### In CollaboratorForm.tsx:

```jsx
// After creating/having a collaborator:
<FileUpload
  entityType="logo"                    // MUST be "logo" (not "logos")
  entityId={collaborator.id}           // The collaborator ID
  accept="image/*"
  onSuccess={(attachments) => {
    // Logo is already auto-synced to collaborator by backend
    // Just update local state for immediate display
    const { download_url } = attachments[0];
    setFormData(prev => ({ ...prev, logo: download_url }));
  }}
/>
```

### Key Points:
1. **entity_type must be exactly `"logo"`** - not "logos" (plural)
2. **entity_id is the collaborator ID** - get from create response or existing record
3. **No need to update collaborator** - backend auto-syncs the logo URL
4. **Use `download_url` from response** - for immediate display in form

---

## Delete Logo Flow

When a logo attachment is deleted:

**API:** `DELETE /api/v1/attachments/{attachment_id}`

**What happens automatically:**
1. File is deleted from S3
2. Attachment record is deleted from database
3. Backend auto-clears `collaborator.logo` to `null`

```
DELETE /api/v1/attachments/ATCH00000001

Result:
- S3 file deleted
- Attachment record deleted
- collaborator.logo = null (auto-cleared)
```

---

## Product Variant Images Flow

Similar to collaborator logos, but uses `entity_type: "variant"`:

### Upload Variant Image

**API:** `POST /api/v1/attachments`

**Request Form Data:**
| Field | Value |
|-------|-------|
| `entity_type` | `variant` |
| `entity_id` | `PVAR00000001` |
| `file` | (image file) |

**What happens:**
1. File uploaded to S3 in `product-variants/{variant_id}/` folder
2. S3 key auto-added to `variant.images` JSON array
3. Presigned URLs generated in API responses

---

## Summary

| Action | API | Auto-Sync |
|--------|-----|-----------|
| Create Collaborator | `POST /api/v1/collaborators` | N/A |
| Upload Logo | `POST /api/v1/attachments` (entity_type=logo) | Updates `collaborator.logo` |
| Delete Logo | `DELETE /api/v1/attachments/{id}` | Clears `collaborator.logo` |
| Upload Variant Image | `POST /api/v1/attachments` (entity_type=variant) | Adds to `variant.images` |
| Delete Variant Image | `DELETE /api/v1/attachments/{id}` | Removes from `variant.images` |
