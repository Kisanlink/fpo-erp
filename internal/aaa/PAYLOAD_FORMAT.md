# AAA Service Payload Format

This document describes the JWT token payload format that the ERP service expects from the AAA service.

## Expected JWT Token Structure

```json
{
  "user_id": "USER_abc123def456ghi789",
  "username": "john.doe",
  "is_validated": true,
  "roles": [
    {
      "id": "ROL_xyz789abc123",
      "user_id": "USER_abc123def456ghi789",
      "role_id": "ROL_admin_role",
      "is_active": true,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": "ROL_def456ghi789",
      "user_id": "USER_abc123def456ghi789", 
      "role_id": "ROL_user_manager",
      "is_active": true,
      "created_at": "2024-01-15T11:00:00Z",
      "updated_at": "2024-01-15T11:00:00Z"
    }
  ],
  "permissions": [
    "sale_summaries:read",
    "warehouses:read",
    "warehouses:create",
    "warehouses:update",
    "warehouses:delete",
    "inventory_batches:read",
    "inventory_batches:create",
    "inventory_batches:update",
    "inventory_batches:delete",
    "sale_items:read",
    "sale_items:create",
    "sale_items:update",
    "sale_items:delete",
    "inventory_transactions:read",
    "inventory_transactions:create",
    "sales:read",
    "sales:create",
    "sales:update",
    "sales:delete",
    "returns:read",
    "returns:create",
    "returns:update",
    "returns:delete",
    "sku:read",
    "sku:create",
    "sku:update",
    "sku:delete",
    "return_items:read",
    "return_items:create",
    "return_items:update",
    "return_items:delete",
    "return_summaries:read",
    "return_summaries:create",
    "return_summaries:update",
    "return_summaries:delete",
    "refund_policy:read",
    "refund_policy:create",
    "refund_policy:update",
    "refund_policy:delete",
    "bank_payments:read",
    "bank_payments:create",
    "bank_payments:update",
    "attachments:read",
    "attachments:create",
    "attachments:delete"
  ],
  "token_type": "access",
  "iss": "aaa-service",
  "aud": "aaa-clients",
  "exp": 1705229400,
  "iat": 1705225800,
  "nbf": 1705225800,
  "jti": "jwt_123456789abcdef"
}
```

## Required Fields

### Custom Claims
- `user_id` (string) - Unique user identifier
- `username` (string) - User's username/email
- `is_validated` (boolean) - Whether user account is validated
- `roles` (array) - Array of role objects
- `permissions` (array) - Array of permission strings
- `token_type` (string) - Type of token (e.g., "access")

### Role Object Structure
```json
{
  "id": "ROL_xyz789abc123",
  "user_id": "USER_abc123def456ghi789",
  "role_id": "ROL_admin_role",
  "is_active": true,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Standard JWT Claims
- `iss` (issuer) - Should be "aaa-service" or your AAA service identifier
- `aud` (audience) - Can be "aaa-clients" or any audience value
- `exp` (expiration) - Unix timestamp when token expires
- `iat` (issued at) - Unix timestamp when token was issued
- `nbf` (not before) - Unix timestamp when token becomes valid
- `jti` (JWT ID) - Unique identifier for the token

## Permission Format

The ERP expects permissions in the format: `{entity}:{action}`

### Available Permissions

#### Sale Summaries
- `sale_summaries:read` - Read sale summary data

#### Warehouses
- `warehouses:read` - Read warehouse data
- `warehouses:create` - Create warehouses
- `warehouses:update` - Update warehouses
- `warehouses:delete` - Delete warehouses

#### Inventory Batches
- `inventory_batches:read` - Read inventory batches
- `inventory_batches:create` - Create inventory batches
- `inventory_batches:update` - Update inventory batches
- `inventory_batches:delete` - Delete inventory batches

#### Sale Items
- `sale_items:read` - Read sale items data
- `sale_items:create` - Create sale items
- `sale_items:update` - Update sale items
- `sale_items:delete` - Delete sale items

#### Inventory Transactions
- `inventory_transactions:read` - Read inventory transactions
- `inventory_transactions:create` - Create inventory transactions

#### Sales
- `sales:read` - Read sales data
- `sales:create` - Create sales records
- `sales:update` - Update sales records
- `sales:delete` - Delete sales records

#### Returns
- `returns:read` - Read returns data
- `returns:create` - Create return records
- `returns:update` - Update return records
- `returns:delete` - Delete return records

#### Products (SKU)
- `sku:read` - Read product data
- `sku:create` - Create products
- `sku:update` - Update products
- `sku:delete` - Delete products

#### Return Items
- `return_items:read` - Read return items data
- `return_items:create` - Create return items
- `return_items:update` - Update return items
- `return_items:delete` - Delete return items

#### Return Summaries
- `return_summaries:read` - Read return summary data
- `return_summaries:create` - Create return summaries
- `return_summaries:update` - Update return summaries
- `return_summaries:delete` - Delete return summaries

#### Refund Policies
- `refund_policy:read` - Read refund policies
- `refund_policy:create` - Create refund policies
- `refund_policy:update` - Update refund policies
- `refund_policy:delete` - Delete refund policies

#### Bank Payments
- `bank_payments:read` - Read bank payments
- `bank_payments:create` - Create bank payments
- `bank_payments:update` - Update bank payments

#### Attachments
- `attachments:read` - Read attachments
- `attachments:create` - Create/upload attachments
- `attachments:delete` - Delete attachments

## Role Mapping

The ERP service uses the `role_id` field from the roles array to determine user permissions. Common role IDs:

- `ROL_admin_role` - Full access to all operations
- `ROL_user_manager` - Manager-level access
- `ROL_store_manager` - Store manager access
- `ROL_store_staff` - Store staff access
- `ROL_auditor` - Read-only access for auditing
- `ROL_accountant` - Financial operations access

## Permission Matrix by Role

| Entity | Director | CEO | Auditor | Accountant | Tech_Support | Store_Manager | Store_Staff |
|--------|----------|-----|---------|------------|--------------|---------------|-------------|
| sale_summaries | R | CRUD | R | R | R/W (temp) | R | R |
| warehouses | R | CRUD | R | – | R/W (temp) | CRUD | R |
| inventory_batches | R | CRUD | R | – | R/W (temp) | CRUD | R |
| sale_items | R | CRUD | R | R | R/W (temp) | R | CRUD |
| inventory_transactions | R | CRUD | R | – | R/W (temp) | CRUD | R |
| sales | R | CRUD | R | R | R/W (temp) | R | CRUD |
| returns | R | CRUD | R | R | R/W (temp) | R | CRUD |
| sku | R | CRUD | R | – | R/W (temp) | CRUD | R |
| return_items | R | CRUD | R | R | R/W (temp) | R | CRUD |
| return_summaries | R | CRUD | R | R | R/W (temp) | R | CRUD |
| refund_policy | R | CRUD | R | CRUD | R/W (temp) | – | – |
| bank_payments | R | CRUD | R | CRUD | R/W (temp) | – | – |
| attachments | R | CRUD | R | R | R/W (temp) | R | R |

**Legend:**
- **R** = Read access
- **CRUD** = Create, Read, Update, Delete access
- **R/W (temp)** = Read/Write access (temporary for Tech Support)
- **–** = No access

## Example Usage

```bash
# API request with JWT token
curl -X GET http://localhost:8080/api/v1/warehouses \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

The ERP service will:
1. Validate the JWT token using the configured secret
2. Extract user information and permissions from the token
3. Check if the user has the required permission for the requested operation
4. Allow or deny the request based on the user's permissions
5. Cache the user's permissions for performance (TTL-based)

## Error Handling

If the token is invalid or missing required fields, the ERP service will return:
- `401 Unauthorized` - Invalid or missing token
- `403 Forbidden` - Insufficient permissions
- `400 Bad Request` - Malformed token structure
