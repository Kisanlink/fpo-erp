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
	// Initialize gRPC authorization client
	authzClient, err := NewAuthzClient(config.AAA.GRPCAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorization client: %w", err)
	}

	return &AAAMiddleware{
		config:      config,
		cache:       NewPermissionCache(config.AAA.CacheTTL),
		authzClient: authzClient,
	}, nil
}

// Authenticate validates JWT tokens from AAA service
func (m *AAAMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
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
			c.Next()
			return
		}

		// Not in cache, parse from token and cache it
		expiresAt := time.Now().Add(time.Duration(m.config.AAA.CacheTTL) * time.Minute)
		if claims.ExpiresAt != nil {
			expiresAt = claims.ExpiresAt.Time
		}

		// Extract permissions from the new role structure
		permissions := ExtractPermissions(claims.RoleIDs)

		cachedUser := &CachedUser{
			UserID:      claims.UserID,
			Username:    claims.Username,
			Roles:       claims.RoleIDs,
			Permissions: permissions,
			ExpiresAt:   expiresAt,
		}

		m.cache.Set(claims.UserID, cachedUser)

		// Set in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.RoleIDs)
		c.Set("permissions", permissions)

		c.Next()
	}
}

// RequirePermission checks if user has a specific permission using gRPC
func (m *AAAMiddleware) RequirePermission(resourceType, resourceID, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(string)

		// Check permission via gRPC
		allowed, err := m.authzClient.CheckPermission(c.Request.Context(), userID, resourceType, resourceID, action)
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

// RequireAnyPermission checks if user has any of the specified permissions using gRPC
func (m *AAAMiddleware) RequireAnyPermission(permissions []Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(string)

		// Check permissions via gRPC
		results, err := m.authzClient.CheckMultiplePermissions(c.Request.Context(), userID, permissions)
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
		userID := c.MustGet("user_id").(string)

		// Check permissions via gRPC
		results, err := m.authzClient.CheckMultiplePermissions(c.Request.Context(), userID, permissions)
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

		// Extract permissions for debugging
		permissions := ExtractPermissions(claims.RoleIDs)

		// Log token info for debugging
		utils.Debug("Token validated for user:", claims.Username)
		utils.Debug("User roles count:", len(claims.RoleIDs))
		utils.Debug("User permissions count:", len(permissions))
		utils.Debug("Token isvalidate field:", claims.IsValidated)

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
