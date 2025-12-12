# Frontend API Changes Catalog - December 12, 2025

## Issue 1: New Endpoint - Get Variants by Collaborator

**Type**: New Feature

**Endpoint**: `GET /api/v1/collaborators/{collaborator_id}/variants`

**Description**: Retrieve all product variants associated with a specific collaborator with pagination support.

**Query Parameters**:
- `limit` (optional, integer, default: 50, max: 200) - Number of records to return
- `offset` (optional, integer, default: 0) - Number of records to skip

**Path Parameters**:
- `collaborator_id` (required, string) - Collaborator ID (format: `CLAB00000001`)

**Authentication**: Required (Bearer token)

**Authorization**: Requires `variant:read` permission

**Request Example**:
```bash
GET /api/v1/collaborators/CLAB00000001/variants?limit=50&offset=0
Authorization: Bearer <token>
```

**Success Response** (200 OK):
```json
{
  "success": true,
  "message": "Variants retrieved successfully",
  "data": [
    {
      "id": "PVAR00000001",
      "product_id": "PROD00000001",
      "variant_name": "500g",
      "description": "Half kilogram pack",
      "quantity": "500",
      "pack_size": "500g",
      "sku": "TOM-500G-ABC123",
      "barcode": "1234567890123",
      "images": [
        "product-variants/PVAR00000001/image1.jpg"
      ],
      "image_urls": [
        "https://s3.amazonaws.com/bucket/product-variants/PVAR00000001/image1.jpg?..."
      ],
      "prices": [
        {
          "id": "PRIC00000001",
          "variant_id": "PVAR00000001",
          "price_type": "retail",
          "price": 100.50,
          "currency": "INR",
          "effective_from": "2025-01-01T00:00:00Z",
          "effective_to": null,
          "is_active": true,
          "created_at": "2025-12-12T10:00:00Z",
          "updated_at": "2025-12-12T10:00:00Z"
        }
      ],
      "is_active": true,
      "created_at": "2025-12-12T10:00:00Z",
      "updated_at": "2025-12-12T10:00:00Z"
    }
  ],
  "pagination": {
    "total": 15,
    "limit": 50,
    "offset": 0
  }
}
```

**Error Responses**:
- `400 Bad Request` - Invalid collaborator ID
- `401 Unauthorized` - Missing or invalid authentication token
- `403 Forbidden` - Insufficient permissions
- `500 Internal Server Error` - Server error

**Use Cases**:
- Display all variants sold by a specific vendor/supplier
- Filter products by collaborator on frontend
- Show collaborator's product catalog
- Export collaborator-specific product list

**Implementation Notes**:
- Searches within `collaborator_ids` JSON array using PostgreSQL `@>` operator
- Only returns active variants (`is_active = true`)
- Results ordered by `created_at DESC` (newest first)
- Pagination follows standard ERP pagination pattern
- Image URLs are presigned S3 URLs (valid for 1 hour)
- Prices fetched from `product_prices` table

**Frontend Integration Example**:
```typescript
// API call
const getVariantsByCollaborator = async (
  collaboratorId: string,
  limit: number = 50,
  offset: number = 0
) => {
  const response = await fetch(
    `/api/v1/collaborators/${collaboratorId}/variants?limit=${limit}&offset=${offset}`,
    {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    }
  );

  if (!response.ok) {
    throw new Error('Failed to fetch collaborator variants');
  }

  return response.json();
};

// React component usage
const CollaboratorVariantsList = ({ collaboratorId }) => {
  const [variants, setVariants] = useState([]);
  const [pagination, setPagination] = useState({ total: 0, limit: 50, offset: 0 });

  useEffect(() => {
    const fetchVariants = async () => {
      const result = await getVariantsByCollaborator(
        collaboratorId,
        pagination.limit,
        pagination.offset
      );
      setVariants(result.data);
      setPagination(result.pagination);
    };

    fetchVariants();
  }, [collaboratorId, pagination.offset]);

  return (
    <div>
      <h2>Variants for Collaborator {collaboratorId}</h2>
      <ul>
        {variants.map(variant => (
          <li key={variant.id}>
            {variant.variant_name} - SKU: {variant.sku}
            {variant.prices[0] && ` - ₹${variant.prices[0].price}`}
          </li>
        ))}
      </ul>
      <p>Total: {pagination.total} variants</p>
    </div>
  );
};
```

---

## Issue 2: Collaborator Logo Auto-Sync (BREAKING CHANGE)

**Type**: Breaking Change

**Changes**:
- Logo is now auto-synced to collaborator on upload (like product variant images)
- `POST /api/v1/attachments` with `entity_type: "logo"` now automatically updates the collaborator's `logo` field
- Logo field now stores presigned S3 URL (1-year expiration) instead of attachment ID
- On attachment deletion, collaborator's `logo` field is automatically cleared

**Backend Entity Types** (use these exact values):
| Entity Type | Folder | Description |
|-------------|--------|-------------|
| `logo` | `logos/` | Collaborator logos |
| `po` | `purchase-orders/{id}/` | Purchase order documents |
| `grn` | `grns/{id}/` | GRN documents |
| `variant` | `product-variants/{id}/` | Product variant images |

**Frontend Entity Type Fixes Required**:
```typescript
// BEFORE (WRONG - will go to misc folder)
entityType: "logos"              // ❌ WRONG
entityType: "purchase_order_docs" // ❌ WRONG
entityType: "grn_docs"           // ❌ WRONG

// AFTER (CORRECT)
entityType: "logo"    // ✅ CORRECT
entityType: "po"      // ✅ CORRECT
entityType: "grn"     // ✅ CORRECT
entityType: "variant" // ✅ CORRECT
```

**Upload Response** (AttachmentResponse):
```json
{
  "id": "ATCH00000001",
  "entity_type": "logo",
  "entity_id": "CLAB00000001",
  "file_path": "logos/uuid.png",
  "file_type": "image/png",
  "download_url": "https://s3.amazonaws.com/...",  // Presigned URL
  "uploaded_by": "USER00000001",
  "uploaded_at": "2025-12-12T10:00:00Z"
}
```

**Collaborator Response** (after logo upload):
```json
{
  "id": "CLAB00000001",
  "company_name": "ABC Suppliers",
  "logo": "https://s3.amazonaws.com/...",  // Auto-synced presigned URL
  ...
}
```

**Migration for CollaboratorForm.tsx**:
```jsx
// BEFORE (WRONG)
<FileUpload
  entityType="logos"  // ❌ Wrong entity type
  entityId={initialData.id}
  onSuccess={(attachments) => {
    const firstAttachment = attachments[0] as { url?: string };  // ❌ Wrong field
    if (firstAttachment.url) {
      setFormData((prev) => ({ ...prev, logo: firstAttachment.url }));
    }
  }}
/>

// AFTER (CORRECT)
<FileUpload
  entityType="logo"  // ✅ Fixed entity type
  entityId={initialData.id}
  onSuccess={(attachments) => {
    const firstAttachment = attachments[0] as { download_url?: string };  // ✅ Correct field
    if (firstAttachment.download_url) {
      // Logo is already auto-synced to collaborator by backend
      // Just update local form state for immediate display
      setFormData((prev) => ({ ...prev, logo: firstAttachment.download_url }));
    }
  }}
/>
```

**attachments.ts Schema Fix**:
```typescript
// BEFORE (WRONG)
export const entityTypeEnum = z.enum([
  'logos',              // ❌ WRONG
  'purchase_order_docs', // ❌ WRONG
  'grn_docs',           // ❌ WRONG
  'product',
  'variant',
  'misc',
]);

// AFTER (CORRECT)
export const entityTypeEnum = z.enum([
  'logo',    // ✅ CORRECT - matches backend
  'po',      // ✅ CORRECT - matches backend
  'grn',     // ✅ CORRECT - matches backend
  'variant', // ✅ CORRECT
  'misc',    // ✅ CORRECT
]);
```

**FileUpload.tsx Switch Case Fixes**:
```typescript
// getAcceptedTypes function - BEFORE
switch (entityType) {
  case 'logos':              // ❌ WRONG
  case 'product':
    return 'image/*';
  case 'purchase_order_docs': // ❌ WRONG
  case 'grn_docs':           // ❌ WRONG
    return '.pdf,.doc,.docx,.xls,.xlsx,.csv,image/*';

// getAcceptedTypes function - AFTER
switch (entityType) {
  case 'logo':    // ✅ CORRECT
  case 'product':
    return 'image/*';
  case 'po':      // ✅ CORRECT
  case 'grn':     // ✅ CORRECT
    return '.pdf,.doc,.docx,.xls,.xlsx,.csv,image/*';
```

**Benefits**:
- Logo uploads work automatically (like variant images)
- No manual collaborator update needed after upload
- Logo automatically cleared when attachment deleted
- Consistent entity types across all attachments

---

## Issue 2: Purchase Order Documents Auto-Sync (BREAKING CHANGE)

**Type**: Breaking Change

**Changes**:
- PO documents are now auto-synced to purchase order on upload (like product variant images)
- `POST /api/v1/attachments` with `entity_type: "po"` now automatically updates the PO's `documents` field
- `documents` field is now a JSON array of S3 keys (stored in DB), returned as presigned URLs in API response
- On attachment deletion, the S3 key is automatically removed from the PO's `documents` array
- **NO manual PO update needed after upload**

**Before vs After**:

| Aspect | Before | After |
|--------|--------|-------|
| Storage | N/A (manual) | JSON array of S3 keys |
| Response | N/A | Array of presigned URLs |
| Auto-sync | ❌ | ✅ |
| Delete sync | ❌ | ✅ |

**Purchase Order Response** (after document upload):
```json
{
  "id": "PORD00000001",
  "po_number": "PO-2025-0001",
  "documents": [
    "https://s3.amazonaws.com/bucket/purchase-orders/PORD00000001/invoice.pdf?...",
    "https://s3.amazonaws.com/bucket/purchase-orders/PORD00000001/receipt.jpg?..."
  ],
  "collaborator_id": "CLAB00000001",
  "warehouse_id": "WARE00000001",
  ...
}
```

**Upload Workflow**:
```typescript
// 1. Upload document to PO
const response = await uploadAttachment({
  entityType: "po",       // ✅ CORRECT entity type
  entityId: purchaseOrder.id,
  file: selectedFile
});

// 2. Document is automatically added to PO's documents array
// 3. Refresh PO data to see the new document in response
const updatedPO = await getPurchaseOrder(purchaseOrder.id);
// updatedPO.documents will contain presigned URL for uploaded file
```

**Delete Workflow**:
```typescript
// 1. Delete attachment
await deleteAttachment(attachmentId);

// 2. Document is automatically removed from PO's documents array
// 3. Refresh PO data to see updated documents list
const updatedPO = await getPurchaseOrder(purchaseOrder.id);
```

---

## Issue 3: GRN Documents Auto-Sync (BREAKING CHANGE)

**Type**: Breaking Change

**Changes**:
- **REMOVED**: `grn_document` field (single attachment ID)
- **ADDED**: `documents` field (JSON array of S3 keys, returned as presigned URLs)
- GRN documents are now auto-synced on upload (like product variant images)
- `POST /api/v1/attachments` with `entity_type: "grn"` now automatically updates the GRN's `documents` field
- On attachment deletion, the S3 key is automatically removed from the GRN's `documents` array
- **UpdateGRNRequest no longer has `grn_document` field** - documents are auto-synced via attachments API

**Before vs After**:

| Aspect | Before | After |
|--------|--------|-------|
| Field | `grn_document` (single attachment ID) | `documents` (array of presigned URLs) |
| Type | `string \| null` | `string[]` |
| Storage | Attachment ID reference | JSON array of S3 keys |
| Response | Attachment ID | Array of presigned URLs |
| Auto-sync | ❌ | ✅ |
| Multiple docs | ❌ | ✅ |

**GRN Response** (after document upload):
```json
{
  "id": "GRNX00000001",
  "grn_number": "GRN-VENDOR-001",
  "documents": [
    "https://s3.amazonaws.com/bucket/grns/GRNX00000001/vendor-grn.pdf?...",
    "https://s3.amazonaws.com/bucket/grns/GRNX00000001/quality-report.jpg?..."
  ],
  "po_id": "PORD00000001",
  "po_number": "PO-2025-0001",
  ...
}
```

**Update GRN Request** (CHANGED):
```typescript
// BEFORE (REMOVED)
interface UpdateGRNRequest {
  grn_document?: string;    // ❌ REMOVED - no longer works
  remarks?: string;
  quality_status?: string;
}

// AFTER (CURRENT)
interface UpdateGRNRequest {
  // NO grn_document field - documents are auto-synced via attachments API
  remarks?: string;
  quality_status?: string;  // "accepted" | "rejected" | "partial"
}
```

**Upload Workflow**:
```typescript
// 1. Upload document to GRN
const response = await uploadAttachment({
  entityType: "grn",       // ✅ CORRECT entity type
  entityId: grn.id,
  file: selectedFile
});

// 2. Document is automatically added to GRN's documents array
// 3. Refresh GRN data to see the new document in response
const updatedGRN = await getGRN(grn.id);
// updatedGRN.documents will contain presigned URLs for all uploaded files
```

**Migration for GRN Document Display**:
```typescript
// BEFORE (WRONG - using attachment ID)
const GRNDocumentDisplay = ({ grn }) => {
  // ❌ Old approach - fetching by attachment ID
  const { data: attachment } = useQuery(
    ['attachment', grn.grn_document],
    () => getAttachment(grn.grn_document)
  );
  return <a href={attachment?.download_url}>View Document</a>;
};

// AFTER (CORRECT - using presigned URLs directly)
const GRNDocumentDisplay = ({ grn }) => {
  // ✅ New approach - URLs are already in the response
  return (
    <div>
      {grn.documents?.map((url, index) => (
        <a key={index} href={url} target="_blank">
          Document {index + 1}
        </a>
      ))}
    </div>
  );
};
```

---

## Summary of Auto-Sync Entity Types

| Entity Type | Field | Storage | Response | Auto-Sync |
|-------------|-------|---------|----------|-----------|
| `logo` | `collaborator.logo` | S3 presigned URL | Single URL string | ✅ |
| `variant` | `variant.images` | JSON array of S3 keys | Array of presigned URLs | ✅ |
| `po` | `purchase_order.documents` | JSON array of S3 keys | Array of presigned URLs | ✅ |
| `grn` | `grn.documents` | JSON array of S3 keys | Array of presigned URLs | ✅ |
| `misc` | N/A | S3 key | Via attachment API | ❌ |

**Benefits**:
- Frontend doesn't need to store attachment IDs
- No extra API calls to get presigned URLs
- Delete automatically removes from parent entity
- Multiple documents supported (PO, GRN, variant)
- Consistent pattern across all entity types

---

## Issue 4: Deleting Attachments (Important for Storage Management)

**Why Delete Matters**:
- S3 storage costs money - unused files should be deleted
- Deleting an attachment automatically:
  1. Removes file from S3 (frees storage space)
  2. Removes attachment record from database
  3. Auto-removes from parent entity's field (logo, images, documents)

### Delete API Endpoint

```
DELETE /api/v1/attachments/:id
```

**Required**: Attachment ID (e.g., `ATCH00000001`)

**Response**: 204 No Content (success) or error

### How to Get Attachment ID for Deletion

Since the API responses now return presigned URLs (not attachment IDs), you need to query attachments by entity:

```
GET /api/v1/attachments/entity/:entity_type/:entity_id
```

**Example**: Get all attachments for a GRN
```typescript
// Get attachments for GRN to find their IDs for deletion
const response = await fetch(`/api/v1/attachments/entity/grn/${grnId}`);
const attachments = await response.json();

// Response:
[
  {
    "id": "ATCH00000001",           // Use this ID for deletion
    "entity_type": "grn",
    "entity_id": "GRNX00000001",
    "file_path": "grns/GRNX00000001/invoice.pdf",
    "file_type": "application/pdf",
    "download_url": "https://s3...",  // Same URL as in grn.documents[]
    "uploaded_by": "USER00000001",
    "uploaded_at": "2025-12-12T10:00:00Z"
  },
  {
    "id": "ATCH00000002",
    "entity_type": "grn",
    "entity_id": "GRNX00000001",
    "file_path": "grns/GRNX00000001/photo.jpg",
    ...
  }
]
```

### Delete Workflow by Entity Type

#### Deleting Collaborator Logo
```typescript
// 1. Get collaborator's logo attachment
const attachments = await getAttachmentsByEntity('logo', collaboratorId);
const logoAttachment = attachments[0];

// 2. Delete the attachment
await deleteAttachment(logoAttachment.id);

// 3. Logo is automatically cleared from collaborator
// collaborator.logo will be null/empty after refresh
```

#### Deleting Product Variant Image
```typescript
// 1. Get variant's image attachments
const attachments = await getAttachmentsByEntity('variant', variantId);

// 2. Find the specific image to delete (match by URL or index)
const imageToDelete = attachments.find(att =>
  variant.images.includes(att.download_url) ||
  att.file_path.includes('specific-filename.jpg')
);

// 3. Delete the attachment
await deleteAttachment(imageToDelete.id);

// 4. Image is automatically removed from variant.images array
const updatedVariant = await getVariant(variantId);
// updatedVariant.images will no longer contain the deleted image URL
```

#### Deleting PO Document
```typescript
// 1. Get PO's document attachments
const attachments = await getAttachmentsByEntity('po', purchaseOrderId);

// 2. Delete specific document
await deleteAttachment(attachments[0].id);

// 3. Document is automatically removed from purchase_order.documents array
const updatedPO = await getPurchaseOrder(purchaseOrderId);
```

#### Deleting GRN Document
```typescript
// 1. Get GRN's document attachments
const attachments = await getAttachmentsByEntity('grn', grnId);

// 2. Delete specific document
await deleteAttachment(attachments[0].id);

// 3. Document is automatically removed from grn.documents array
const updatedGRN = await getGRN(grnId);
```

### React Component Example: Document List with Delete

```tsx
interface DocumentListProps {
  entityType: 'po' | 'grn' | 'variant' | 'logo';
  entityId: string;
  documents: string[];  // Array of presigned URLs from parent entity
  onDocumentDeleted: () => void;  // Callback to refresh parent
}

const DocumentList: React.FC<DocumentListProps> = ({
  entityType,
  entityId,
  documents,
  onDocumentDeleted
}) => {
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [deleting, setDeleting] = useState<string | null>(null);

  // Fetch attachments to get IDs for deletion
  useEffect(() => {
    const fetchAttachments = async () => {
      const response = await fetch(
        `/api/v1/attachments/entity/${entityType}/${entityId}`
      );
      const data = await response.json();
      setAttachments(data);
    };
    fetchAttachments();
  }, [entityType, entityId, documents]);

  const handleDelete = async (attachmentId: string) => {
    if (!confirm('Delete this document? This cannot be undone.')) return;

    setDeleting(attachmentId);
    try {
      await fetch(`/api/v1/attachments/${attachmentId}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` }
      });

      // Refresh parent entity to get updated documents list
      onDocumentDeleted();
    } catch (error) {
      console.error('Failed to delete document:', error);
      alert('Failed to delete document');
    } finally {
      setDeleting(null);
    }
  };

  // Match presigned URLs to attachments for deletion
  const getAttachmentForUrl = (url: string) => {
    // URLs contain the file_path, so we can match
    return attachments.find(att => url.includes(encodeURIComponent(att.file_path)));
  };

  return (
    <div className="document-list">
      {documents.map((url, index) => {
        const attachment = getAttachmentForUrl(url);
        const fileName = url.split('/').pop()?.split('?')[0] || `Document ${index + 1}`;

        return (
          <div key={index} className="document-item">
            <a href={url} target="_blank" rel="noopener noreferrer">
              📄 {decodeURIComponent(fileName)}
            </a>
            {attachment && (
              <button
                onClick={() => handleDelete(attachment.id)}
                disabled={deleting === attachment.id}
                className="delete-btn"
              >
                {deleting === attachment.id ? '⏳' : '🗑️'}
              </button>
            )}
          </div>
        );
      })}
    </div>
  );
};
```

### API Helper Functions

```typescript
// attachments.api.ts

export const getAttachmentsByEntity = async (
  entityType: string,
  entityId: string
): Promise<Attachment[]> => {
  const response = await fetch(
    `/api/v1/attachments/entity/${entityType}/${entityId}`,
    { headers: { 'Authorization': `Bearer ${getToken()}` } }
  );
  if (!response.ok) throw new Error('Failed to fetch attachments');
  return response.json();
};

export const deleteAttachment = async (attachmentId: string): Promise<void> => {
  const response = await fetch(
    `/api/v1/attachments/${attachmentId}`,
    {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${getToken()}` }
    }
  );
  if (!response.ok) throw new Error('Failed to delete attachment');
};

// Convenience function: Delete all documents for an entity
export const deleteAllAttachmentsForEntity = async (
  entityType: string,
  entityId: string
): Promise<void> => {
  const attachments = await getAttachmentsByEntity(entityType, entityId);
  await Promise.all(attachments.map(att => deleteAttachment(att.id)));
};
```

### Delete Behavior Summary

| Entity Type | Delete Effect |
|-------------|---------------|
| `logo` | S3 file deleted, `collaborator.logo` cleared to null |
| `variant` | S3 file deleted, URL removed from `variant.images[]` |
| `po` | S3 file deleted, URL removed from `purchase_order.documents[]` |
| `grn` | S3 file deleted, URL removed from `grn.documents[]` |
| `misc` | S3 file deleted, attachment record removed (no parent update) |

**Important Notes**:
1. Deletion is permanent - files cannot be recovered
2. Always confirm with user before deleting
3. Refresh parent entity after deletion to get updated list
4. Presigned URLs in parent entity will be invalid after deletion

---

