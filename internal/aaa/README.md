# AAA Service Integration

This package provides integration with the AAA (Authentication, Authorization, and Accounting) service for the Kisanlink ERP system.

## Features

- **JWT Token Validation**: Validates JWT tokens from AAA service
- **Permission-Based Access Control**: Checks user permissions for route access
- **Role-Based Access Control**: Checks user roles for route access
- **TTL Caching**: Caches user permissions with automatic expiration
- **Audit Logging**: Optional audit event logging

## Components

### 1. Permission Cache (`cache.go`)

Provides TTL-based caching for user permissions to improve performance.

```go
// Create cache with 30-minute TTL
cache := NewPermissionCache(30)

// Store user data
cache.Set(userID, userData)

// Retrieve user data
if cached, exists := cache.Get(userID); exists {
    // Use cached data
}
```

### 2. AAA Middleware (`middleware.go`)

Handles authentication and authorization using JWT tokens from AAA service.

```go
// Create middleware
aaaMiddleware := aaa.NewAAAMiddleware(config)

// Authenticate and check permissions
router.GET("/warehouses", 
    aaaMiddleware.Authenticate(),
    aaaMiddleware.RequirePermission("warehouses:read"),
    handler.GetWarehouses)
```

### 3. Types (`types.go`)

Defines data structures for AAA service integration.

- `AAARole`: Represents a user role
- `AAATokenClaims`: JWT token claims structure
- `UserContext`: User information in request context

### 4. Audit Logger (`audit.go`)

Optional audit logging functionality.

```go
logger := aaa.NewAuditLogger(true)
logger.LogEvent(ctx, userID, "create", "warehouse", "wh_123", "Created warehouse")
```

## Configuration

Add the following environment variables:

```bash
# AAA Service Configuration
AAA_JWT_SECRET=your-jwt-secret-here
AAA_CACHE_TTL=30  # Cache TTL in minutes
```

## Usage

### 1. Update Route Registration

```go
// Initialize AAA middleware
aaaMiddleware := aaa.NewAAAMiddleware(cfg)

// Register routes with permission checks
warehouses := router.Group("/warehouses")
{
    warehouses.GET("", aaaMiddleware.RequirePermission("warehouses:read"), handler.GetAllWarehouses)
    warehouses.POST("", aaaMiddleware.RequirePermission("warehouses:create"), handler.CreateWarehouse)
    warehouses.PATCH("/:id", aaaMiddleware.RequirePermission("warehouses:update"), handler.UpdateWarehouse)
    warehouses.DELETE("/:id", aaaMiddleware.RequirePermission("warehouses:delete"), handler.DeleteWarehouse)
}
```

### 2. Permission Names

Use these permission names based on your permission matrix:

- `sale_summaries:read`, `sale_summaries:create`, `sale_summaries:update`, `sale_summaries:delete`
- `warehouses:read`, `warehouses:create`, `warehouses:update`, `warehouses:delete`
- `inventory_batches:read`, `inventory_batches:create`, `inventory_batches:update`, `inventory_batches:delete`
- `sale_items:read`, `sale_items:create`, `sale_items:update`, `sale_items:delete`
- `inventory_transactions:read`, `inventory_transactions:create`, `inventory_transactions:update`, `inventory_transactions:delete`
- `sales:read`, `sales:create`, `sales:update`, `sales:delete`
- `returns:read`, `returns:create`, `returns:update`, `returns:delete`
- `sku:read`, `sku:create`, `sku:update`, `sku:delete`
- `return_items:read`, `return_items:create`, `return_items:update`, `return_items:delete`
- `return_summaries:read`, `return_summaries:create`, `return_summaries:update`, `return_summaries:delete`
- `refund_policy:read`, `refund_policy:create`, `refund_policy:update`, `refund_policy:delete`
- `bank_payments:read`, `bank_payments:create`, `bank_payments:update`, `bank_payments:delete`
- `attachments:read`, `attachments:create`, `attachments:update`, `attachments:delete`
- `users:assign_roles`, `users:create`, `users:deactivate`, `users:view_all`
- `fpo_settings:manage`

### 3. Middleware Functions

- `Authenticate()`: Validates JWT token and sets user context
- `RequirePermission(permission)`: Checks if user has specific permission
- `RequireAnyPermission(permissions...)`: Checks if user has any of the specified permissions
- `RequireAllPermissions(permissions...)`: Checks if user has all specified permissions
- `RequireRole(role)`: Checks if user has specific role
- `RequireAnyRole(roles...)`: Checks if user has any of the specified roles

### 4. Getting User Context

```go
func (h *Handler) SomeHandler(c *gin.Context) {
    userContext := aaa.GetUserContext(c)
    
    // Access user information
    userID := userContext.UserID
    username := userContext.Username
    roles := userContext.Roles
    permissions := userContext.Permissions
}
```

## JWT Token Structure

The AAA service provides JWT tokens with this structure:

```json
{
  "user_id": "usr_abc123",
  "username": "john.doe",
  "is_validated": true,
  "roles": [
    {
      "id": "rol_admin",
      "name": "admin",
      "description": "Administrator role",
      "is_active": true
    }
  ],
  "permissions": [
    "user:view",
    "user:edit",
    "user:delete",
    "role:manage",
    "system:manage_users"
  ],
  "token_type": "access",
  "exp": 1703980800,
  "iat": 1703894400,
  "sub": "usr_abc123"
}
```

## Testing

Run the tests to verify the integration:

```bash
go test ./internal/...
```

## Benefits

1. **High Performance**: TTL caching reduces JWT parsing overhead
2. **Security**: Centralized authentication and authorization
3. **Scalability**: No network calls for permission checks
4. **Maintainability**: Clean separation of concerns
5. **Flexibility**: Easy to add new permissions and roles

## Future Enhancements

1. **gRPC Client**: ✅ Implemented real gRPC client for AAA service communication
2. **Audit Integration**: Send audit events to AAA service
3. **Token Refresh**: Automatic token refresh functionality
4. **Metrics**: Add performance metrics and monitoring
