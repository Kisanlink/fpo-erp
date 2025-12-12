# Frontend API Changes Catalog - December 12, 2025

## Issue 1: Collaborator Logo Auto-Sync (BREAKING CHANGE)

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

## Issue 5: Purchase Order Address Detection Optimization (Backend Only)

**Type**: Backend Optimization (No Frontend Changes Required)

**What Changed**:
- Purchase order creation now uses locally cached address data instead of making gRPC calls to AAA service
- Inter-state GST detection (CGST+SGST vs IGST) now reads from `collaborator.state` and `warehouse.state` fields in local database
- Eliminates 2 gRPC calls per purchase order creation (1 for collaborator address, 1 for warehouse address)

**Frontend Impact**: None - API request/response unchanged

**Technical Details**:
- **OLD**: `determineInterState()` made 2 gRPC calls to AAA service: `GetAddress(collaboratorAddressID)` and `GetAddress(warehouseAddressID)`
- **NEW**: `determineInterState()` reads from local cache: `collaborator.State` and `warehouse.State` (no gRPC calls)
- Write-through cache pattern: addresses are synced to local DB on collaborator/warehouse CREATE/UPDATE operations
- Fallback behavior unchanged: defaults to intra-state (CGST+SGST) if state data unavailable

**Performance Improvement**:
- Saves 2 gRPC round-trips per PO creation (~50-100ms total latency reduction)
- Faster PO creation response time
- Reduced load on AAA service
- No additional database queries (collaborator and warehouse already fetched during PO creation)

**Why This Works**:
- Collaborator and warehouse models have local address cache fields (including `State`) that are synced on write operations
- Address data is already available when fetching collaborator/warehouse for PO validation
- State comparison logic remains identical - only the data source changed (local DB vs gRPC)

**Files Modified**:
- `internal/services/purchase_order_service.go:111` - Updated call to `determineInterState()`
- `internal/services/purchase_order_service.go:1024-1056` - Rewritten function to use local cache

**No Frontend Changes Required** - API contract unchanged.

---

