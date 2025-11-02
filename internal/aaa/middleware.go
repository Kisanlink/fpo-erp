package aaa

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"kisanlink-erp/internal/config"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AAAMiddleware handles AAA service authentication and authorization
type AAAMiddleware struct {
	config      *config.Config
	cache       *PermissionCache
	authzClient *AuthzClient
}

// NewAAAMiddleware creates a new AAA middleware
func NewAAAMiddleware(config *config.Config) (*AAAMiddleware, error) {
	var authzClient *AuthzClient
	var err error

	// Only initialize gRPC client if AAA is enabled
	if config.AAA.Enabled {
		authzClient, err = NewAuthzClient(config.AAA.GRPCAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to create authorization client: %w", err)
		}
		utils.Info("AAA middleware initialized with gRPC client")
	} else {
		utils.Info("⚠️  AAA middleware initialized in BYPASS mode (disabled)")
	}

	return &AAAMiddleware{
		config:      config,
		cache:       NewPermissionCache(config.AAA.CacheTTL),
		authzClient: authzClient, // Will be nil when AAA disabled
	}, nil
}

// Authenticate validates JWT tokens from AAA service
func (m *AAAMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Early return if AAA disabled - set mock user context for testing
		if !m.config.AAA.Enabled {
			mockRole := AAARole{
				ID:       "test-role",
				RoleID:   "test-role-id",
				IsActive: true,
				Role: AAAUserRole{
					Name:        "Admin",
					Description: "Mock admin role for testing",
					Scope:       "global",
					IsActive:    true,
				},
			}
			c.Set("user_id", "test-user-123")
			c.Set("username", "test-user")
			c.Set("roles", []AAARole{mockRole})
			c.Set("permissions", []string{"*:*"})
			c.Set("jwt_token", "mock-jwt-token")
			c.Set("organization_id", "test-org-123")
			c.Set("organization_name", "Test Organization")
			c.Set("organization_ids", []string{"test-org-123"})
			c.Set("organizations", []interface{}{})
			c.Set("groups", []string{})
			c.Next()
			return
		}

		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedResponse(c, "Authorization header required")
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.UnauthorizedResponse(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate JWT token
		claims, err := m.parseToken(tokenString)
		if err != nil {
			utils.UnauthorizedResponse(c, "Invalid token")
			c.Abort()
			return
		}

		// Check cache first
		if cached, exists := m.cache.Get(claims.UserID); exists {
			// Use cached data
			c.Set("user_id", cached.UserID)
			c.Set("username", cached.Username)
			c.Set("roles", cached.Roles)
			c.Set("permissions", cached.Permissions)
			c.Set("jwt_token", tokenString) // Store JWT token for gRPC calls

			// ✅ Extract organization context from claims (not cached)
			// We don't cache organization context because organization memberships can change
			c.Set("organization_id", claims.GetPrimaryOrganizationID())
			c.Set("organization_name", claims.GetOrganizationName())
			c.Set("organization_ids", claims.GetOrganizationIDs())

			// Combine organizations from BOTH top-level and user_context
			allOrgs := claims.Organizations
			if claims.UserContext != nil && len(claims.UserContext.Organizations) > 0 {
				allOrgs = append(allOrgs, claims.UserContext.Organizations...)
			}
			c.Set("organizations", allOrgs)

			// Combine groups from BOTH top-level and user_context
			allGroups := claims.Groups
			if claims.UserContext != nil && len(claims.UserContext.Groups) > 0 {
				allGroups = append(allGroups, claims.UserContext.Groups...)
			}
			c.Set("groups", allGroups)

			c.Next()
			return
		}

		// Not in cache, parse from token and cache it
		expiresAt := time.Now().Add(time.Duration(m.config.AAA.CacheTTL) * time.Minute)
		if claims.ExpiresAt != nil {
			expiresAt = claims.ExpiresAt.Time
		}

		// Convert JWT roles to AAARole structure for backward compatibility
		var roles []AAARole
		if claims.UserContext != nil && claims.UserContext.Roles != nil {
			roles = ConvertJWTRolesToAAARole(claims.UserContext.Roles)
		} else {
			roles = []AAARole{} // Empty array if no user context
		}

		// Extract permissions from the role structure
		// Note: JWT roles don't contain permissions - actual checking is done via gRPC
		permissions := ExtractPermissions(roles)

		// ✅ Extract organization context from token claims
		organizationID := claims.GetPrimaryOrganizationID()
		organizationName := claims.GetOrganizationName()
		organizationIDs := claims.GetOrganizationIDs()

		cachedUser := &CachedUser{
			UserID:      claims.UserID,
			Username:    claims.Username,
			Roles:       roles,
			Permissions: permissions,
			ExpiresAt:   expiresAt,
		}

		m.cache.Set(claims.UserID, cachedUser)

		// Set in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", roles)
		c.Set("permissions", permissions)
		c.Set("jwt_token", tokenString) // Store JWT token for gRPC calls

		// ✅ Set organization context in request context
		c.Set("organization_id", organizationID)     // Primary organization ID
		c.Set("organization_name", organizationName) // Organization name
		c.Set("organization_ids", organizationIDs)   // All organizations

		// Combine organizations from BOTH top-level and user_context
		allOrgs := claims.Organizations
		if claims.UserContext != nil && len(claims.UserContext.Organizations) > 0 {
			allOrgs = append(allOrgs, claims.UserContext.Organizations...)
		}
		c.Set("organizations", allOrgs)

		// Combine groups from BOTH top-level and user_context
		allGroups := claims.Groups
		if claims.UserContext != nil && len(claims.UserContext.Groups) > 0 {
			allGroups = append(allGroups, claims.UserContext.Groups...)
		}
		c.Set("groups", allGroups)

		c.Next()
	}
}

// RequirePermission checks if user has a specific permission using gRPC
func (m *AAAMiddleware) RequirePermission(resourceType, resourceID, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Early return if AAA disabled - allow all permissions
		if !m.config.AAA.Enabled {
			c.Next()
			return
		}

		userID := c.MustGet("user_id").(string)
		jwtToken := c.GetString("jwt_token") // Get JWT token from context

		// Check permission via gRPC with JWT token
		allowed, err := m.authzClient.CheckPermissionWithToken(c.Request.Context(), userID, resourceType, resourceID, action, jwtToken)
		if err != nil {
			utils.ErrorResponse(c, 500, "Permission check failed: "+err.Error(), err)
			c.Abort()
			return
		}

		if !allowed {
			utils.ForbiddenResponse(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireOrgPermission checks if user has permission scoped to their organization
// This is the recommended method for multi-tenant resources (collaborators, products, sales, etc.)
// It automatically uses the organization_id from the JWT token context
func (m *AAAMiddleware) RequireOrgPermission(resourceType, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Early return if AAA disabled - allow all permissions
		if !m.config.AAA.Enabled {
			c.Next()
			return
		}

		userID := c.MustGet("user_id").(string)
		jwtToken := c.GetString("jwt_token")

		// Extract organization ID from context (set by Authenticate middleware)
		organizationID := c.GetString("organization_id")
		if organizationID == "" {
			utils.ErrorResponse(c, 403, "Organization context required. Ensure you're authenticated with a valid organization.", nil)
			c.Abort()
			return
		}

		// Check permission with organization scope
		// This ensures users can only access resources within their organization
		allowed, err := m.authzClient.CheckPermissionWithToken(
			c.Request.Context(),
			userID,
			resourceType,
			organizationID, // ✅ Organization-scoped resource ID
			action,
			jwtToken,
		)
		if err != nil {
			utils.ErrorResponse(c, 500, "Permission check failed: "+err.Error(), err)
			c.Abort()
			return
		}

		if !allowed {
			utils.ForbiddenResponse(c, "Insufficient permissions for this organization")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireMultipleOrgPermissions checks if user has multiple permissions scoped to their organization
// This is useful when an operation requires permissions on multiple resource types
// For example, creating a collaborator with address requires both "collaborator:create" and "address:create"
func (m *AAAMiddleware) RequireMultipleOrgPermissions(resourceActions []ResourceAction) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Early return if AAA disabled - allow all permissions
		if !m.config.AAA.Enabled {
			c.Next()
			return
		}

		userID := c.MustGet("user_id").(string)
		jwtToken := c.GetString("jwt_token")

		// Extract organization ID from context (set by Authenticate middleware)
		organizationID := c.GetString("organization_id")
		if organizationID == "" {
			utils.ErrorResponse(c, 403, "Organization context required. Ensure you're authenticated with a valid organization.", nil)
			c.Abort()
			return
		}

		// Build permissions with organization scope
		var permissions []Permission
		for _, ra := range resourceActions {
			// Special case: addresses resource uses empty resourceID for global scope
			resourceID := organizationID
			if ra.ResourceType == "addresses" {
				resourceID = ""
			}

			permissions = append(permissions, Permission{
				ResourceType: ra.ResourceType,
				ResourceID:   resourceID,
				Action:       ra.Action,
			})
		}

		// Check all permissions via batch gRPC call
		results, err := m.authzClient.CheckMultiplePermissionsWithToken(
			c.Request.Context(),
			userID,
			permissions,
			jwtToken,
		)
		if err != nil {
			utils.ErrorResponse(c, 500, "Permission check failed: "+err.Error(), err)
			c.Abort()
			return
		}

		// Check if all permissions are granted
		// results is a map[string]bool with keys like "resourceType:resourceID:action"
		for _, ra := range resourceActions {
			// Special case: addresses resource uses empty resourceID
			resourceID := organizationID
			if ra.ResourceType == "addresses" {
				resourceID = ""
			}

			key := fmt.Sprintf("%s:%s:%s", ra.ResourceType, resourceID, ra.Action)
			if allowed, exists := results[key]; !exists || !allowed {
				utils.ForbiddenResponse(c, fmt.Sprintf("Insufficient permissions: missing %s:%s", ra.ResourceType, ra.Action))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequireAnyPermission checks if user has any of the specified permissions using gRPC
func (m *AAAMiddleware) RequireAnyPermission(permissions []Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Early return if AAA disabled - allow all permissions
		if !m.config.AAA.Enabled {
			c.Next()
			return
		}

		userID := c.MustGet("user_id").(string)
		jwtToken := c.GetString("jwt_token") // Get JWT token from context

		// Check permissions via gRPC with JWT token
		results, err := m.authzClient.CheckMultiplePermissionsWithToken(c.Request.Context(), userID, permissions, jwtToken)
		if err != nil {
			utils.ErrorResponse(c, 500, "Permission check failed: "+err.Error(), err)
			c.Abort()
			return
		}

		// Check if any permission is granted
		for _, allowed := range results {
			if allowed {
				c.Next()
				return
			}
		}

		utils.ForbiddenResponse(c, "Insufficient permissions")
		c.Abort()
	}
}

// RequireAllPermissions checks if user has all specified permissions using gRPC
func (m *AAAMiddleware) RequireAllPermissions(permissions []Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Early return if AAA disabled - allow all permissions
		if !m.config.AAA.Enabled {
			c.Next()
			return
		}

		userID := c.MustGet("user_id").(string)
		jwtToken := c.GetString("jwt_token") // Get JWT token from context

		// Check permissions via gRPC with JWT token
		results, err := m.authzClient.CheckMultiplePermissionsWithToken(c.Request.Context(), userID, permissions, jwtToken)
		if err != nil {
			utils.ErrorResponse(c, 500, "Permission check failed: "+err.Error(), err)
			c.Abort()
			return
		}

		// Check if all permissions are granted
		for _, allowed := range results {
			if !allowed {
				utils.ForbiddenResponse(c, "Insufficient permissions")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequireRole checks if user has a specific role
func (m *AAAMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Early return if AAA disabled - allow all roles
		if !m.config.AAA.Enabled {
			c.Next()
			return
		}

		userRoles := c.MustGet("roles").([]AAARole)

		if !containsRole(userRoles, role) {
			utils.ForbiddenResponse(c, "Insufficient role")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole checks if user has any of the specified roles
func (m *AAAMiddleware) RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Early return if AAA disabled - allow all roles
		if !m.config.AAA.Enabled {
			c.Next()
			return
		}

		userRoles := c.MustGet("roles").([]AAARole)

		for _, role := range roles {
			if containsRole(userRoles, role) {
				c.Next()
				return
			}
		}

		utils.ForbiddenResponse(c, "Insufficient role")
		c.Abort()
	}
}

// parseToken parses and validates JWT token from AAA service
func (m *AAAMiddleware) parseToken(tokenString string) (*AAATokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AAATokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token signing method")
		}
		return []byte(m.config.AAA.JWTSecret), nil
	})

	if err != nil {
		utils.Error("JWT parsing error:", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*AAATokenClaims); ok && token.Valid {
		// Validate required fields
		if claims.UserID == "" {
			return nil, errors.New("missing user_id in token")
		}
		if claims.Username == "" {
			return nil, errors.New("missing username in token")
		}

		// Convert JWT roles for debugging
		var roles []AAARole
		roleCount := 0
		if claims.UserContext != nil && claims.UserContext.Roles != nil {
			roles = ConvertJWTRolesToAAARole(claims.UserContext.Roles)
			roleCount = len(claims.UserContext.Roles)
		}
		permissions := ExtractPermissions(roles)

		// Log token info for debugging
		utils.Debug("Token validated for user:", claims.Username)
		utils.Debug("User role IDs from token:", claims.RoleIDs)
		utils.Debug("User roles from user_context:", roleCount)
		utils.Debug("User permissions count:", len(permissions))
		utils.Debug("Token isvalidate field:", claims.IsValidated)
		utils.Debug("UserContext present:", claims.UserContext != nil)
		// ✅ Log organization context
		utils.Debug("Organizations count:", len(claims.Organizations))
		utils.Debug("Groups count:", len(claims.Groups))
		utils.Debug("Primary Organization ID:", claims.GetPrimaryOrganizationID())
		utils.Debug("Organization Name:", claims.GetOrganizationName())
		utils.Debug("Token version:", claims.TokenVersion)

		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// GetUserContext extracts user context from gin context
func GetUserContext(c *gin.Context) *UserContext {
	return &UserContext{
		UserID:      c.MustGet("user_id").(string),
		Username:    c.MustGet("username").(string),
		Roles:       c.MustGet("roles").([]AAARole),
		Permissions: c.MustGet("permissions").([]string),
	}
}

// GetCacheStats returns cache statistics
func (m *AAAMiddleware) GetCacheStats() map[string]interface{} {
	return m.cache.GetStats()
}

// Close closes the gRPC connection
func (m *AAAMiddleware) Close() error {
	if m.authzClient != nil {
		return m.authzClient.Close()
	}
	return nil
}
