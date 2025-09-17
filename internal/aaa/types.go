package aaa

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AAARole represents a role from the AAA service
type AAARole struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	RoleID    string    `json:"role_id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AAATokenClaims represents JWT claims from AAA service
type AAATokenClaims struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	IsValidated bool      `json:"is_validated"`
	Roles       []AAARole `json:"roles"`
	Permissions []string  `json:"permissions"`
	TokenType   string    `json:"token_type"`
	jwt.RegisteredClaims
}

// UserContext represents user information in request context
type UserContext struct {
	UserID      string
	Username    string
	Roles       []AAARole
	Permissions []string
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// containsRole checks if a slice contains a specific role ID
func containsRole(roles []AAARole, roleID string) bool {
	for _, role := range roles {
		if role.RoleID == roleID {
			return true
		}
	}
	return false
}

// GetRoleNames returns a slice of role names/IDs for easier permission checking
func GetRoleNames(roles []AAARole) []string {
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.RoleID
	}
	return roleNames
}
