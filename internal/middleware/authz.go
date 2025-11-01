package middleware

import (
	"fmt"
	"net/http"

	"kisanlink-erp/internal/auth"

	"github.com/gin-gonic/gin"
)

// Resource and Action types for permission checking
type Resource = string
type Action = string

// InferResourceID is a function type that extracts the resource ID from the gin context
type InferResourceID func(*gin.Context) (string, error)

// AuthZ creates authorization middleware that checks permissions
// This is the standard implementation from e-commerce service
func AuthZ(resourceType Resource, action Action, inferResourceID InferResourceID) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the authenticated user's ID from context
		subjectID, exists := GetSubjectID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authentication required",
				},
			})
			c.Abort()
			return
		}

		// Infer the resource ID from the context
		resourceID, err := inferResourceID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "INVALID_RESOURCE",
					"message": "Failed to identify resource",
					"details": gin.H{
						"reason": err.Error(),
					},
				},
			})
			c.Abort()
			return
		}

		// Get AAA client from context (should be set by authentication middleware)
		aaaClient, exists := c.Get("aaaClient")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Authorization service unavailable",
				},
			})
			c.Abort()
			return
		}

		client, ok := aaaClient.(auth.Client)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Invalid authorization client",
				},
			})
			c.Abort()
			return
		}

		// Get JWT token from context for gRPC authentication
		jwtToken, _ := GetJWTToken(c)

		// Check permission with AAA service
		req := &auth.AuthorizeRequest{
			UserID:     subjectID,
			Resource:   string(resourceType),
			Action:     string(action),
			ResourceID: resourceID,
			JWTToken:   jwtToken, // ✅ Pass JWT token for gRPC authentication
		}
		resp, err := client.Authorize(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "AUTHORIZATION_ERROR",
					"message": "Failed to check permissions",
					"details": gin.H{
						"reason": err.Error(),
					},
				},
			})
			c.Abort()
			return
		}

		if !resp.Allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Access denied",
					"details": gin.H{
						"reason":      "insufficient permissions",
						"resource":    resourceType,
						"action":      action,
						"resource_id": resourceID,
					},
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireOrgPermission checks if user has permission scoped to their organization
// This is the ERP-specific implementation that automatically uses organization_id from JWT
// This is the recommended method for multi-tenant resources (collaborators, products, sales, etc.)
func RequireOrgPermission(resourceType Resource, action Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the authenticated user's ID from context
		subjectID, exists := GetSubjectID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authentication required",
				},
			})
			c.Abort()
			return
		}

		// Extract organization ID from context (set by Authenticate middleware)
		organizationID := GetOrganizationID(c)
		if organizationID == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Organization context required. Ensure you're authenticated with a valid organization.",
				},
			})
			c.Abort()
			return
		}

		// Get AAA client from context
		aaaClient, exists := c.Get("aaaClient")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Authorization service unavailable",
				},
			})
			c.Abort()
			return
		}

		client, ok := aaaClient.(auth.Client)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Invalid authorization client",
				},
			})
			c.Abort()
			return
		}

		// Get JWT token from context for gRPC authentication
		jwtToken, _ := GetJWTToken(c)

		// Check permission with organization scope
		// This ensures users can only access resources within their organization
		req := &auth.AuthorizeRequest{
			UserID:     subjectID,
			Resource:   string(resourceType),
			Action:     string(action),
			ResourceID: organizationID, // ✅ Organization-scoped resource ID
			TenantID:   organizationID, // Include tenant context
			JWTToken:   jwtToken,       // ✅ Pass JWT token for gRPC authentication
		}
		resp, err := client.Authorize(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "AUTHORIZATION_ERROR",
					"message": "Failed to check permissions",
					"details": gin.H{
						"reason": err.Error(),
					},
				},
			})
			c.Abort()
			return
		}

		if !resp.Allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Insufficient permissions for this organization",
					"details": gin.H{
						"resource": resourceType,
						"action":   action,
					},
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Common resource ID inference functions

// InferOrgIDFromContext extracts org_id from auth context
func InferOrgIDFromContext() InferResourceID {
	return func(c *gin.Context) (string, error) {
		orgID := GetOrganizationID(c)
		if orgID == "" {
			return "", fmt.Errorf("organization ID not found in context")
		}
		return orgID, nil
	}
}

// InferOrgIDFromParam extracts org_id from URL parameter
func InferOrgIDFromParam(paramName string) InferResourceID {
	return func(c *gin.Context) (string, error) {
		orgID := c.Param(paramName)
		if orgID == "" {
			return "", fmt.Errorf("organization ID not found in URL parameter: %s", paramName)
		}
		return orgID, nil
	}
}

// InferResourceIDFromParam extracts resource_id from URL parameter
func InferResourceIDFromParam(paramName string) InferResourceID {
	return func(c *gin.Context) (string, error) {
		resourceID := c.Param(paramName)
		if resourceID == "" {
			return "", fmt.Errorf("resource ID not found in URL parameter: %s", paramName)
		}
		return resourceID, nil
	}
}
