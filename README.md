# 🌾 Kisanlink ERP System

A comprehensive Enterprise Resource Planning (ERP) system built with Go, designed for agricultural businesses and FPOs (Farmer Producer Organizations).

[![Production Ready](https://img.shields.io/badge/Status-Production%20Ready-green)](https://github.com/kisanlink/fpo-erp)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow)](LICENSE)

## 📚 API Documentation

- **🌟 Official API Documentation**: [Scalar Registry](https://registry.scalar.com/@kisanlink/apis/kisanlink-erp-api/latest)
- **Interactive Documentation**: `/docs` endpoint (Scalar Go-powered)
- **Swagger UI**: `/swagger/index.html` endpoint
- **OpenAPI Specification**: `/api-docs` endpoint

## 🚀 Features

### Core ERP Modules
- **Procurement Management**: Complete vendor-to-inventory workflow
  - Vendor/Supplier (Collaborator) Management
  - Vendor-Product Associations (many-to-many)
  - Product Variants (separate table)
  - Purchase Order workflow with ALL-IN pricing
  - GRN (Goods Receipt Notes) with 3 input patterns
  - Auto-GRN creation with quality inspection
  - Automatic inventory batch creation from GRN
- **Inventory Management**: Track products, batches, and stock levels with automatic batch creation from GRN and FEFO integration
- **Sales Management**: Handle sales orders, invoices, and customer data
- **Returns Management**: Process returns and refunds
- **Warehouse Management**: Multi-warehouse support with location tracking
- **Product Management**: SKU management with pricing and categorization
- **Tax Management**: Comprehensive tax calculation and compliance system
- **Discount System**: Advanced discount management with validation
- **File Attachments**: Generic entity-based system (S3) for logos, POs, GRNs, and documents
- **Bank Payments**: Track payment transactions for sales and returns
- **Refund Policies**: Manage return and refund policies
- **Reporting**: Sales analytics and inventory reports

### Security & Access Control
- **AAA Service Integration**: Authentication, Authorization, and Accounting
- **Organization-Scoped Permissions**: Multi-tenant isolation with automatic organization-level access control
- **Role-Based Access Control (RBAC)**: Granular permissions for different user roles
- **JWT Token Validation**: Secure token-based authentication from external AAA service with organization context
- **TTL Caching**: High-performance permission caching
- **Audit Logging**: Comprehensive activity tracking
- **External User Management**: User management handled by separate AAA service
- **Multi-Tenant Support**: Complete isolation between organizations with organization-scoped permission checks

## 🛠️ Technology Stack

- **Backend**: Go 1.21+
- **Framework**: Gin (HTTP framework)
- **Database**: PostgreSQL with GORM ORM
- **Authentication**: JWT with external AAA service integration
- **File Storage**: AWS S3
- **HTTP API**: RESTful API for all operations
- **Documentation**: Scalar Go package + Swagger UI
- **Tax System**: Production-ready GST compliance system
- **Configuration**: Environment-based configuration
- **Logging**: Structured logging with levels
- **Validation**: Comprehensive input validation and business rules

## 🏗️ Project Structure

```
Kisanlink-erp-v1/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── aaa/                        # AAA service integration
│   │   ├── cache.go               # TTL permission caching
│   │   ├── middleware.go          # AAA authentication middleware
│   │   ├── types.go               # AAA data structures
│   │   ├── audit.go               # Audit logging
│   │   └── README.md              # AAA integration docs
│   ├── api/
│   │   ├── handlers/              # HTTP request handlers
│   │   ├── middleware/            # HTTP middleware (CORS, logging, rate limiting)
│   │   ├── routes/                # Route definitions
│   │   └── server/                # HTTP server setup
│   ├── config/                    # Configuration management
│   ├── database/
│   │   ├── models/                # Database models (organized by domain)
│   │   │   ├── attachments.go     # Generic entity-based attachment models
│   │   │   ├── bank_payments.go   # Bank payment models
│   │   │   ├── collaborator.go    # Vendor/supplier models
│   │   │   ├── collaborator_product.go  # Vendor-product associations
│   │   │   ├── product_variant.go # Product size/quantity variants
│   │   │   ├── purchase_order.go  # Purchase order models
│   │   │   ├── grn.go            # Goods receipt note models
│   │   │   ├── inventory.go       # Inventory batch and transaction models
│   │   │   ├── price.go          # Product pricing models
│   │   │   ├── product.go        # Product/SKU models
│   │   │   ├── returns.go        # Return and refund policy models
│   │   │   ├── sales.go          # Sales models
│   │   │   └── warehouse.go      # Warehouse models
│   │   ├── repositories/          # Data access layer
│   │   └── migrator.go            # Database migrations
│   ├── services/                  # Business logic layer
│   ├── utils/                     # Utility functions
│   └── aaa/                       # AAA service integration
├── proto/                         # Protocol Buffer definitions
├── scripts/                       # Build and deployment scripts
└── docs/                          # API documentation
```

## AAA Service Integration

The ERP system integrates with an external AAA (Authentication, Authorization, and Accounting) service for centralized security management. **User management is handled by a separate service** - this ERP service only handles business operations and uses tokens from the header for authentication.

### Features
- **JWT Token Validation**: Validates tokens from AAA service
- **Permission-Based Access Control**: Route-level permission checks
- **Role-Based Access Control**: Role-based route protection
- **TTL Caching**: Caches user permissions for performance
- **Audit Logging**: Optional audit event logging

### Organization-Scoped Permissions (January 2025)

The system implements **organization-scoped permissions** for multi-tenant isolation:

#### How It Works
1. **JWT Token Contains Organization Context**: AAA LoginV2 includes `organizations` array in JWT
2. **Automatic Extraction**: Middleware extracts `organization_id` from JWT token
3. **Scoped Permission Checks**: All routes check permissions against user's organization
4. **Multi-Tenant Isolation**: Users can only access resources within their organization

#### Permission Hierarchy
```
Global Permissions (Super Admin):
  ResourceType: "collaborator"
  ResourceId:   "*"              // Access all organizations
  Action:       "read"

Organization-Scoped Permissions (FPO Users):
  ResourceType: "collaborator"
  ResourceId:   "ORG_12345"      // Access only ORG_12345
  Action:       "read"
```

#### Implementation
- **16 Handler Files Updated**: ~115 routes now use organization-scoped permissions
- **Automatic Scoping**: Middleware handles organization context extraction
- **Backward Compatible**: Super admin roles can still use wildcard (`*`) permissions
- **Security Enhancement**: Critical for multi-tenant security and data isolation

**Documentation**: See `files/ORG_SCOPED_PERMISSIONS_IMPLEMENTATION.md` for complete details

### Permission Matrix

The system supports the following permissions based on user roles:

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

### Permission Names

The system uses the following permission naming convention:
- `{entity}:read` - Read access to entity
- `{entity}:create` - Create access to entity
- `{entity}:update` - Update access to entity
- `{entity}:delete` - Delete access to entity

**Examples:**
- `sale_summaries:read` - Read sale summary data
- `warehouses:read` - Read warehouse data
- `sales:create` - Create sales records
- `sku:update` - Update product information
- `returns:delete` - Delete return records
- `refund_policy:create` - Create refund policies
- `bank_payments:read` - Read bank payment records
- `attachments:read` - Read attachments

## 🏢 Multi-Tenant Architecture

The ERP system is **fully multi-tenant** with organization-level isolation:

### Organization Context
- **JWT Token**: Contains user's organization ID(s) from AAA LoginV2
- **Automatic Extraction**: Middleware extracts `organization_id` from JWT
- **Scoped Access**: All API calls are automatically scoped to user's organization
- **Data Isolation**: Users cannot access resources from other organizations

### Permission Model

#### For FPO Users
```bash
# User from Organization ORG_A creates a collaborator
POST /api/v1/collaborators
Authorization: Bearer <token-with-org-A>

# Permission Check:
# - ResourceType: "collaborator"
# - ResourceId:   "ORG_A"          ← Organization-scoped
# - Action:       "create"
# Result: Collaborator created in ORG_A only
```

#### For Super Admins
```bash
# Super admin with global permissions
POST /api/v1/collaborators
Authorization: Bearer <super-admin-token>

# Permission Check:
# - ResourceType: "collaborator"
# - ResourceId:   "*"              ← Global access
# - Action:       "create"
# Result: Can create collaborators in any organization
```

### Security Benefits
✅ **Organization Isolation**: Users cannot access data from other organizations
✅ **Automatic Scoping**: No manual filtering needed in handlers
✅ **Audit Trail**: All operations logged with organization context
✅ **Flexible Access**: Supports single-org users and multi-org admins

### Implementation Details
- **Middleware**: `RequireOrgPermission(resourceType, action)` in all handlers
- **Coverage**: 16 handler files, ~115 routes
- **Documentation**: Complete guide in `files/ORG_SCOPED_PERMISSIONS_IMPLEMENTATION.md`

## 🚦 Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- AWS S3 bucket (for file attachments)
- AAA service (for authentication)

### Environment Variables

Create a `.env` file in the project root:

```bash
# Server Configuration
SERVER_HTTP_PORT=8080
AAA_SERVICE_URL=localhost:9091
SERVER_MODE=release

# Database Configuration
DB_POSTGRES_HOST=localhost
DB_POSTGRES_PORT=5432
DB_POSTGRES_USER=postgres
DB_POSTGRES_PASSWORD=your_password
DB_POSTGRES_DBNAME=erp_database
DB_POSTGRES_SSLMODE=disable

# JWT Configuration
JWT_SECRET=your-jwt-secret
JWT_EXPIRY_HOURS=24

# AWS Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_S3_BUCKET=your-s3-bucket

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
CORS_ALLOWED_HEADERS=Origin,Content-Type,Accept,Authorization,X-Requested-With,X-Request-ID

# AAA Service Configuration
AAA_JWT_SECRET=your-aaa-jwt-secret
AAA_CACHE_TTL=30
```

### JWT Token Generation

For testing purposes, you can generate JWT tokens using the provided script:

1. **Create the token generator script:**
   ```bash
   # Create generate_token.go file with your JWT secret
   # Replace "REPLACE_WITH_YOUR_ACTUAL_JWT_SECRET" with your actual secret
   ```

2. **Generate a token:**
   ```bash
   go run generate_token.go
   ```

3. **Use the token in API requests:**
   ```bash
   curl -X GET http://localhost:8080/api/v1/warehouses \
     -H "Authorization: Bearer <generated-token>" \
     -H "Content-Type: application/json"
   ```

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/Kisanlink/fpo-erp.git
   cd Kisanlink-erp-v1
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up the database**
   ```bash
   # Create PostgreSQL database
   createdb erp_database
   ```

4. **Run the application**
   ```bash
   go run cmd/server/main.go
   ```

The server will start on:
- HTTP API: `http://localhost:3000` (default)
- API Documentation: `http://localhost:3000/docs` (Scalar)
- Swagger UI: `http://localhost:3000/swagger/index.html`
- OpenAPI Spec: `http://localhost:3000/api-docs`
- AAA Service: `localhost:9091`

## API Endpoints

### Authentication
All API endpoints require a valid JWT token from the AAA service in the Authorization header:
```
Authorization: Bearer <jwt-token>
```

### Core Endpoints

#### Warehouses
- `GET /api/v1/warehouses` - List all warehouses
- `POST /api/v1/warehouses` - Create warehouse
- `GET /api/v1/warehouses/:id` - Get warehouse details
- `PATCH /api/v1/warehouses/:id` - Update warehouse
- `DELETE /api/v1/warehouses/:id` - Delete warehouse

#### Products (SKU)
- `GET /api/v1/products` - List all products
- `POST /api/v1/products` - Create product
- `GET /api/v1/products/:id` - Get product details
- `PATCH /api/v1/products/:id` - Update product
- `DELETE /api/v1/products/:id` - Delete product

#### Inventory Batches (with Tax Support)
- `GET /api/v1/batches` - List all batches
- `POST /api/v1/batches` - Create batch **with tax configuration**
- `GET /api/v1/batches/:id` - Get batch details
- `GET /api/v1/batches/expiring` - Get expiring batches
- `GET /api/v1/batches/low-stock` - Get low stock batches

#### Tax Management
- `GET /api/v1/taxes` - List all taxes
- `POST /api/v1/taxes` - Create custom tax
- `GET /api/v1/taxes/:id` - Get tax details
- `PATCH /api/v1/taxes/:id` - Update tax
- `DELETE /api/v1/taxes/:id` - Delete tax
- `POST /api/v1/taxes/calculate` - Calculate tax for items

#### Sales
- `GET /api/v1/sales` - List all sales
- `POST /api/v1/sales` - Create sale
- `GET /api/v1/sales/:id` - Get sale details
- `PUT /api/v1/sales/:id` - Update sale
- `DELETE /api/v1/sales/:id` - Delete sale

#### Returns
- `GET /api/v1/returns` - List all returns
- `POST /api/v1/returns` - Create return
- `GET /api/v1/returns/:id` - Get return details
- `PUT /api/v1/returns/:id` - Update return
- `DELETE /api/v1/returns/:id` - Delete return

#### Refund Policies
- `GET /api/v1/refund-policies` - List all refund policies
- `POST /api/v1/refund-policies` - Create refund policy
- `GET /api/v1/refund-policies/:id` - Get refund policy details
- `PUT /api/v1/refund-policies/:id` - Update refund policy
- `DELETE /api/v1/refund-policies/:id` - Delete refund policy

#### Collaborators (Vendors/Suppliers)
- `GET /api/v1/collaborators` - List all vendors
- `POST /api/v1/collaborators` - Create vendor
- `GET /api/v1/collaborators/:id` - Get vendor details
- `GET /api/v1/collaborators/active` - Get active vendors
- `PATCH /api/v1/collaborators/:id` - Update vendor
- `PATCH /api/v1/collaborators/:id/status` - Activate/deactivate vendor
- `DELETE /api/v1/collaborators/:id` - Delete vendor

#### Collaborator Products (Vendor-Product Associations)
- `GET /api/v1/collaborator-products` - List vendor-product associations
- `POST /api/v1/collaborator-products` - Create association
- `GET /api/v1/collaborator-products/:id` - Get association details
- `GET /api/v1/collaborator-products/collaborator/:id` - Get products by vendor
- `GET /api/v1/collaborator-products/product/:id` - Get vendors by product
- `PATCH /api/v1/collaborator-products/:id` - Update association
- `PATCH /api/v1/collaborator-products/:id/status` - Activate/deactivate
- `DELETE /api/v1/collaborator-products/:id` - Delete association

#### Product Variants
- `GET /api/v1/product-variants` - List all variants
- `POST /api/v1/product-variants` - Create variant
- `GET /api/v1/product-variants/:id` - Get variant details
- `GET /api/v1/product-variants/product/:id` - Get variants by product
- `PATCH /api/v1/product-variants/:id` - Update variant
- `PATCH /api/v1/product-variants/:id/status` - Activate/deactivate
- `DELETE /api/v1/product-variants/:id` - Delete variant

#### Purchase Orders
- `GET /api/v1/purchase-orders` - List all purchase orders
- `POST /api/v1/purchase-orders` - Create purchase order
- `GET /api/v1/purchase-orders/:id` - Get PO details
- `GET /api/v1/purchase-orders/pending-deliveries` - Get pending deliveries
- `GET /api/v1/purchase-orders/status/:status` - Get POs by status
- `PATCH /api/v1/purchase-orders/:id/status` - Update status (with auto-GRN support)
- `PATCH /api/v1/purchase-orders/:id/payment` - Update payment status
- `GET /api/v1/collaborators/:id/purchase-orders` - Get POs by vendor

#### Goods Receipt Notes (GRN)
- `GET /api/v1/grns` - List all GRNs
- `POST /api/v1/grns` - Create GRN manually
- `GET /api/v1/grns/:id` - Get GRN details
- `GET /api/v1/grns/warehouse/:id` - Get GRNs by warehouse
- `GET /api/v1/grns/purchase-order/:id` - Get GRN by purchase order

#### Bank Payments
- `GET /api/v1/bank-payments` - List all bank payments
- `POST /api/v1/bank-payments` - Create bank payment
- `GET /api/v1/bank-payments/:id` - Get bank payment details
- `PUT /api/v1/bank-payments/:id` - Update bank payment

#### Attachments (Entity-Based)
- `GET /api/v1/attachments` - List attachments (with entity filters)
- `POST /api/v1/attachments` - Upload attachment (requires entity_type & entity_id)
- `GET /api/v1/attachments/:id` - Get attachment metadata
- `GET /api/v1/attachments/:id/download` - Download attachment file
- `GET /api/v1/attachments/:id/url` - Get presigned URL for display
- `GET /api/v1/attachments/:id/info` - Get detailed file information
- `GET /api/v1/attachments/entity/:entity_type/:entity_id` - Get all attachments for entity
- `DELETE /api/v1/attachments/:id` - Delete attachment

## Usage Examples

### Creating a Warehouse
```bash
curl -X POST http://localhost:8080/api/v1/warehouses \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Main Warehouse",
    "location": "Mumbai, Maharashtra",
    "capacity": 10000,
    "manager_name": "John Doe",
    "contact_number": "+91-9876543210"
  }'
```

### Creating an Inventory Batch (with Tax Configuration)
```bash
curl -X POST http://localhost:3000/api/v1/batches \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "warehouse_id": "WH1234567890",
    "product_id": "PROD00000001",
    "cost_price": 85.50,
    "expiry_date": "2025-12-31",
    "quantity": 1000,
    "cgst_rate": 2.5,
    "sgst_rate": 2.5,
    "custom_tax_ids": ["TAX_CESS_ENV_001", "TAX_MANDI_FEE_001"],
    "is_tax_exempt": false
  }'
```

### Creating a Custom Tax
```bash
curl -X POST http://localhost:3000/api/v1/taxes \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "TAX_CESS_ENV_001",
    "name": "Environmental Cess",
    "tax_type": "item_specific",
    "calculation_type": "percentage",
    "rate": 1.0,
    "valid_from": "2024-01-01T00:00:00Z",
    "is_active": true
  }'
```

### Creating a Sale
```bash
curl -X POST http://localhost:8080/api/v1/sales \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "warehouse_id": "WH1234567890",
    "customer_id": "CUST00000001",
    "batch_id": "BATCH00000001",
    "selling_price": 25.00
  }'
```

### Creating a Refund Policy
```bash
curl -X POST http://localhost:8080/api/v1/refund-policies \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "policy_name": "Standard Return Policy",
    "description": "30-day return policy with 10% restocking fee",
    "max_days": 30,
    "restocking_fee": 10.00
  }'
```

### Creating a Collaborator (Vendor)
```bash
curl -X POST http://localhost:8080/api/v1/collaborators \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "ABC Traders",
    "contact_person": "Rajesh Kumar",
    "contact_number": "+91-9876543210",
    "email": "rajesh@abctraders.com",
    "gst_number": "27AABCU9603R1ZM",
    "bank_account_no": "123456789012",
    "bank_ifsc": "SBIN0001234",
    "address": {
      "line1": "123 Market Street",
      "city": "Mumbai",
      "state": "Maharashtra",
      "pincode": "400001",
      "country": "India"
    }
  }'
```

### Creating a Purchase Order
```bash
curl -X POST http://localhost:8080/api/v1/purchase-orders \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "collaborator_id": "CLAB_abc12345",
    "warehouse_id": "WH_xyz67890",
    "expected_delivery_date": "2025-02-15",
    "items": [
      {
        "product_id": "SKU_rice001",
        "quantity": 1000,
        "unit_price": 45.50
      },
      {
        "product_id": "SKU_wheat002",
        "quantity": 500,
        "unit_price": 38.00
      }
    ]
  }'
```

### Creating GRN with Auto-Acceptance (Pattern 1)
```bash
curl -X PATCH http://localhost:8080/api/v1/purchase-orders/PO_123/status \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "delivered",
    "accept_all": true,
    "default_expiry_date": "2025-12-31"
  }'
```

### Uploading an Attachment (Logo)
```bash
curl -X POST http://localhost:8080/api/v1/attachments \
  -H "Authorization: Bearer <your-jwt-token>" \
  -F "file=@company_logo.png" \
  -F "entity_type=logo" \
  -F "entity_id=CLAB_abc12345"
```

### Getting Presigned URL for Logo Display
```bash
curl -X GET http://localhost:8080/api/v1/attachments/ATT_xyz789/url \
  -H "Authorization: Bearer <your-jwt-token>"
```

## Testing

Run the test suite:
```bash
go test ./...
```

Run specific test packages:
```bash
go test ./internal/aaa/...
go test ./internal/api/handlers/...
go test ./internal/services/...
```

### Testing with JWT Tokens

1. **Generate a test token** using the token generator script
2. **Test different roles** by modifying the permissions in the token
3. **Verify permission enforcement** by testing endpoints with different permission sets
4. **Check audit logs** for authentication and authorization events

### Example Test Scenarios

- **Store Manager**: Test warehouse and inventory management permissions
- **CEO**: Test full CRUD access to all entities
- **Auditor**: Test read-only access to all entities
- **Accountant**: Test financial operations (bank payments, refund policies)
- **Store Staff**: Test limited CRUD operations on sales and returns

## Deployment

### Docker Deployment
```bash
# Build the Docker image
docker build -t kisanlink-erp .

# Run the container
docker run -p 8080:8080 -p 9090:9090 \
  -e DB_POSTGRES_HOST=your-db-host \
  -e DB_POSTGRES_PASSWORD=your-db-password \
  kisanlink-erp
```

### Production Deployment
1. Set up a PostgreSQL database
2. Configure AWS S3 for file storage
3. Set up the AAA service
4. Configure environment variables
5. Deploy using your preferred method (Docker, Kubernetes, etc.)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions:
- Create an issue in the GitHub repository
- Contact the development team
- Check the documentation in the `docs/` directory

## 📈 Recent Updates

### ✅ Latest Completed Features (January 2025)

#### 🔐 Organization-Scoped Permissions (January 2025)
- **Multi-Tenant Security**: Complete organization-level isolation for all resources
- **Automatic Organization Scoping**: Middleware extracts `organization_id` from JWT and validates permissions
- **16 Handlers Updated**: All major handlers (~115 routes) now use organization-scoped permission checks
- **Defense in Depth**: 4-layer security (JWT → Context → Org Permissions → Business Logic)
- **AAA Integration**: Leverages AAA service's hierarchical permission model (global/organization/instance)
- **Zero Performance Impact**: Organization ID already in JWT token, no additional queries needed
- **Super Admin Support**: Global wildcard permissions still work for cross-organization access
- **Production Implementation**: `RequireOrgPermission(resourceType, action)` replaces `RequirePermission(resourceType, "*", action)`

#### 🏭 Complete Procurement Module
- **Vendor/Supplier Management**: Full collaborator (vendor) profiles with AAA address integration
- **Vendor-Product Associations**: Many-to-many relationships with brand info and compliance data
- **Product Variants**: Separate table for size/quantity variants (500g, 1kg, 5kg, etc.)
- **Purchase Order Workflow**: ALL-IN pricing with status tracking (placed → confirmed → delivered → paid)
- **GRN (Goods Receipt Notes)**: 3 flexible input patterns (Accept All, Simple Accept/Reject, Detailed Quantities)
- **Auto-GRN Creation**: Automatic GRN generation with quality inspection and batch tracking
- **Automatic Inventory Updates**: GRN accepted quantities automatically create inventory batches

#### 📎 Generic Attachment System
- **Entity-Based Architecture**: Refactored from sale/return-specific to generic `entity_type` + `entity_id` pattern
- **Multi-Entity Support**: Works for logos, purchase orders, GRNs, and any future entity types
- **Frontend-Friendly**: Two-step workflow (upload → get attachment ID → create entity) with presigned URLs
- **S3 Folder Structure**: Organized by entity type (logos/, purchase-orders/{ID}/, grns/{ID}/, misc/)
- **Database Migration**: Complete migration script for existing data (`migrations/20250128_refactor_attachments_to_entity.sql`)

#### 📦 GRN → Inventory Integration
- **Automatic Batch Creation**: Accepted quantities from GRN automatically create inventory batches
- **FEFO Integration**: New batches immediately participate in First-Expired-First-Out allocation
- **Transaction Audit Trail**: Complete history of inventory movements from GRN to sales
- **Batch Number Tracking**: Vendor batch numbers stored separately from system batch IDs
- **ALL-IN Cost Pricing**: PO unit price becomes inventory batch cost price

#### 🔐 JWT Structure Updates
- **Nested Role Support**: Enhanced parsing for AAA tokens with `user_context.roles` structure
- **Backward Compatibility**: Conversion layer maintains compatibility with existing code
- **Improved Permission Extraction**: Better handling of roleIds and permissions fields

### ✅ Previously Completed Features (2024)
- **🏷️ Production-Ready Tax System**: Complete inventory-based tax management with CGST, SGST, and custom taxes
- **📚 Scalar Go Documentation**: Interactive API documentation using Scalar Go package
- **🔧 Advanced Tax Calculation**: Automatic tax computation during sales with GST compliance
- **🎯 Real-world Tax Scenarios**: Support for multiple tax rates and exemptions
- **🔐 Enhanced Validation**: Complete input validation and business rule enforcement
- **Permission Matrix Implementation**: Complete role-based access control for all entities
- **AAA Service Integration**: Full integration with external AAA service
- **JWT Token Support**: Support for external JWT tokens with custom payload format
- **Discount System**: Comprehensive discount management and validation

### 🎯 Current Status: Production Ready
- **✅ Procurement Module**: Complete vendor-to-inventory workflow (40+ endpoints)
- **✅ Tax System**: Fully implemented and tested
- **✅ Inventory Management**: Automatic updates from GRN with FEFO integration
- **✅ API Documentation**: Complete with Scalar Registry integration
- **✅ GST Compliance**: Indian tax regulations supported
- **✅ Performance**: Optimized with TTL caching and efficient queries

## Roadmap

- [ ] Advanced reporting and analytics
- [ ] Mobile application
- [x] Multi-tenant support (✅ Completed - Organization-scoped permissions implemented)
- [ ] Advanced inventory forecasting
- [ ] Integration with external systems
- [ ] Real-time notifications
- [ ] Advanced audit logging
- [ ] Performance monitoring and metrics
- [ ] Bulk operations support
- [ ] Advanced search and filtering
- [ ] Export functionality (CSV, PDF)
- [ ] Dashboard and analytics
