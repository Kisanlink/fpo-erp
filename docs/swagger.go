package docs

// @title Kisanlink ERP API
// @version 1.0
// @description Comprehensive ERP system for agricultural cooperatives with multi-tenant architecture
// @termsOfService http://swagger.io/terms/

// @contact.name Kisanlink Support
// @contact.url http://www.kisanlink.com/support
// @contact.email support@kisanlink.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3000
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name Warehouses
// @tag.description Warehouse management operations

// @tag.name Products
// @tag.description Product catalog and inventory management

// @tag.name Inventory
// @tag.description Inventory batch and transaction management

// @tag.name Sales
// @tag.description Sales transaction management

// @tag.name Returns
// @tag.description Return and refund processing

// @tag.name Taxes
// @tag.description Tax configuration and calculation

// @tag.name Discounts
// @tag.description Discount management and validation

// @tag.name Product Prices
// @tag.description Product pricing management

// @tag.name Attachments
// @tag.description File attachment management

// @tag.name Bank Payments
// @tag.description Bank payment processing

// @tag.name Refund Policies
// @tag.description Refund policy management
