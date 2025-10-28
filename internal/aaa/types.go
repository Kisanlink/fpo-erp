package aaa

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AAAAddress represents address information in user profile
type AAAAddress struct {
	House       *string `json:"house"`
	Street      *string `json:"street"`
	Landmark    *string `json:"landmark"`
	PostOffice  *string `json:"post_office"`
	Subdistrict *string `json:"subdistrict"`
	District    *string `json:"district"`
	VTC         *string `json:"vtc"`
	State       *string `json:"state"`
	Country     *string `json:"country"`
	Pincode     *string `json:"pincode"`
	FullAddress *string `json:"full_address"`
}

// AAAProfile represents user profile information
type AAAProfile struct {
	UserID        string      `json:"user_id"`
	Name          *string     `json:"name"`
	CareOf        *string     `json:"care_of"`
	DateOfBirth   *time.Time  `json:"date_of_birth"`
	Photo         *string     `json:"photo"`
	YearOfBirth   *int        `json:"year_of_birth"`
	Message       *string     `json:"message"`
	AadhaarNumber *string     `json:"aadhaar_number"`
	EmailHash     *string     `json:"email_hash"`
	ShareCode     *string     `json:"share_code"`
	AddressID     *string     `json:"address_id"`
	Address       AAAAddress  `json:"address"`
}

// AAAUser represents user information in role
type AAAUser struct {
	PhoneNumber   string      `json:"phone_number"`
	CountryCode   string      `json:"country_code"`
	Username      *string     `json:"username"`
	Password      string      `json:"password"`
	MPIN          *string     `json:"mpin"`
	IsValidated   bool        `json:"is_validated"`
	Status        *string     `json:"status"`
	Tokens        int         `json:"tokens"`
	Profile       AAAProfile  `json:"profile"`
	Contacts      interface{} `json:"contacts"`
	Roles         interface{} `json:"roles"`
}

// AAAUserRole represents a role definition
type AAAUserRole struct {
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	Scope          string      `json:"scope"`
	IsActive       bool        `json:"is_active"`
	Version        int         `json:"version"`
	Metadata       interface{} `json:"metadata"`
	OrganizationID *string     `json:"organization_id"`
	GroupID        *string     `json:"group_id"`
	ParentID       *string     `json:"parent_id"`
	Children       interface{} `json:"children"`
	Users          interface{} `json:"users"`
	Permissions    interface{} `json:"permissions"`
}

// AAARole represents a role from the AAA service (new structure)
type AAARole struct {
	ID        string      `json:"id"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	CreatedBy string      `json:"created_by"`
	UpdatedBy string      `json:"updated_by"`
	UserID    string      `json:"UserID"`
	RoleID    string      `json:"RoleID"`
	IsActive  bool        `json:"is_active"`
	User      AAAUser     `json:"user"`
	Role      AAAUserRole `json:"role"`
}

// AAATokenClaims represents JWT claims from AAA service (new structure)
type AAATokenClaims struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	IsValidated bool      `json:"isvalidate"`
	RoleIDs     []AAARole `json:"roleIds"`
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

// ExtractPermissions extracts permissions from the new role structure
func ExtractPermissions(roles []AAARole) []string {
	var permissions []string
	for _, role := range roles {
		// Since permissions is interface{}, we need to handle it carefully
		if role.Role.Permissions != nil {
			// Try to cast to []string or []interface{}
			switch perms := role.Role.Permissions.(type) {
			case []string:
				permissions = append(permissions, perms...)
			case []interface{}:
				for _, perm := range perms {
					if permStr, ok := perm.(string); ok {
						permissions = append(permissions, permStr)
					}
				}
			}
		}
	}
	return permissions
}
