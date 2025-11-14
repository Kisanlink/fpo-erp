package testutils

import (
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/config"

	"github.com/gin-gonic/gin"
)

// MockAAAMiddleware creates a middleware that injects mock user context
// for testing handlers without requiring actual AAA service authentication
func MockAAAMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Inject mock user context that matches AAA authentication middleware
		c.Set("user_id", "test-user-123")
		c.Set("username", "test-user")
		c.Set("organization_id", "test-org-123")
		c.Set("organization_name", "Test Organization")
		c.Set("organization_ids", []string{"test-org-123"})

		// Mock permissions - grant all permissions for testing
		c.Set("permissions", map[string]bool{
			"*:*": true, // Wildcard permission for all operations
		})

		// Mock roles
		c.Set("roles", []map[string]interface{}{
			{
				"id":   "test-role-123",
				"name": "admin",
			},
		})

		// Continue to next handler
		c.Next()
	}
}

// MockAAAMiddlewareWithCustomUser creates a middleware with custom user context
// Useful for testing specific user scenarios or permissions
func MockAAAMiddlewareWithCustomUser(userID, username, orgID string, permissions map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("username", username)
		c.Set("organization_id", orgID)
		c.Set("organization_name", "Test Organization")
		c.Set("organization_ids", []string{orgID})
		c.Set("permissions", permissions)
		c.Set("roles", []map[string]interface{}{
			{
				"id":   "custom-role",
				"name": "custom",
			},
		})
		c.Next()
	}
}

// SetupTestRouter creates a Gin router with mock AAA middleware for testing
func SetupTestRouter() *gin.Engine {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create router
	router := gin.New()

	// Add mock AAA middleware
	router.Use(MockAAAMiddleware())

	return router
}

// SetupTestRouterWithCustomAuth creates a router with custom authentication context
func SetupTestRouterWithCustomAuth(userID, username, orgID string, permissions map[string]bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(MockAAAMiddlewareWithCustomUser(userID, username, orgID, permissions))
	return router
}

// NewMockAAAMiddleware creates a mock AAA middleware for handler testing
// This creates an AAA middleware with AAA disabled so it bypasses all checks
func NewMockAAAMiddleware() *aaa.AAAMiddleware {
	// Create config with AAA disabled
	cfg := &config.Config{
		AAA: config.AAAConfig{
			Enabled: false, // Disable AAA for testing
		},
	}

	// Create middleware with AAA disabled - this will bypass all auth checks
	middleware, _ := aaa.NewAAAMiddleware(cfg)
	return middleware
}
