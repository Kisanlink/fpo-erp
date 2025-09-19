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
- **Inventory Management**: Track products, batches, and stock levels
- **Sales Management**: Handle sales orders, invoices, and customer data
- **Returns Management**: Process returns and refunds
- **Warehouse Management**: Multi-warehouse support with location tracking
- **Product Management**: SKU management with pricing and categorization
- **Tax Management**: Comprehensive tax calculation and compliance system
- **Discount System**: Advanced discount management with validation
- **File Attachments**: S3 integration for document storage
- **Bank Payments**: Track payment transactions for sales and returns
- **Refund Policies**: Manage return and refund policies
- **Reporting**: Sales analytics and inventory reports

### Security & Access Control
- **AAA Service Integration**: Authentication, Authorization, and Accounting
- **Role-Based Access Control (RBAC)**: Granular permissions for different user roles
- **JWT Token Validation**: Secure token-based authentication from external AAA service
- **TTL Caching**: High-performance permission caching
- **Audit Logging**: Comprehensive activity tracking
- **External User Management**: User management handled by separate AAA service

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
│   │   │   ├── attachments.go     # File attachment models
│   │   │   ├── bank_payments.go   # Bank payment models
│   │   │   ├── inventory.go       # Inventory models
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

#### Bank Payments
- `GET /api/v1/bank-payments` - List all bank payments
- `POST /api/v1/bank-payments` - Create bank payment
- `GET /api/v1/bank-payments/:id` - Get bank payment details
- `PUT /api/v1/bank-payments/:id` - Update bank payment

#### Attachments
- `GET /api/v1/attachments` - List attachments
- `POST /api/v1/attachments` - Upload attachment
- `GET /api/v1/attachments/:id` - Get attachment details
- `GET /api/v1/attachments/:id/download` - Download attachment
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

### ✅ Latest Completed Features (2024)
- **🏷️ Production-Ready Tax System**: Complete inventory-based tax management with CGST, SGST, and custom taxes
- **📚 Scalar Go Documentation**: Interactive API documentation using Scalar Go package
- **🔧 Advanced Tax Calculation**: Automatic tax computation during sales with GST compliance
- **📋 Comprehensive API Docs**: Updated with tax endpoints and examples
- **🎯 Real-world Tax Scenarios**: Support for multiple tax rates and exemptions
- **🔐 Enhanced Validation**: Complete input validation and business rule enforcement

### ✅ Previously Completed Features
- **Permission Matrix Implementation**: Complete role-based access control for all entities
- **Model Reorganization**: Moved models from misc files to domain-specific files
- **AAA Service Integration**: Full integration with external AAA service
- **User Management Removal**: Removed local user management (handled by AAA service)
- **JWT Token Support**: Support for external JWT tokens with custom payload format
- **Enhanced Security**: Route-level permission enforcement
- **Discount System**: Comprehensive discount management and validation

### 🎯 Current Status: Production Ready
- **✅ Tax System**: Fully implemented and tested
- **✅ API Documentation**: Complete with Scalar Registry integration
- **✅ GST Compliance**: Indian tax regulations supported
- **✅ Performance**: Optimized with TTL caching and efficient queries

## Roadmap

- [ ] Advanced reporting and analytics
- [ ] Mobile application
- [ ] Multi-tenant support
- [ ] Advanced inventory forecasting
- [ ] Integration with external systems
- [ ] Real-time notifications
- [ ] Advanced audit logging
- [ ] Performance monitoring and metrics
- [ ] Bulk operations support
- [ ] Advanced search and filtering
- [ ] Export functionality (CSV, PDF)
- [ ] Dashboard and analytics
