package aaa

import (
	"strings"
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

// AAARole represents a role from the AAA service (full structure for cache/context)
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

// JWTRole represents simplified role structure in JWT token
type JWTRole struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Scope          string  `json:"scope"`
	OrganizationID *string `json:"organization_id,omitempty"`
	GroupID        *string `json:"group_id,omitempty"`
	IsActive       bool    `json:"is_active"`
}

// JWTOrganization represents organization data in JWT token
type JWTOrganization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// JWTGroup represents group data in JWT token
type JWTGroup struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OrganizationID string `json:"organization_id"`
}

// JWTUserContext represents the user_context field in JWT token
type JWTUserContext struct {
	ID             string            `json:"id"`
	Username       string            `json:"username"`
	PhoneNumber    string            `json:"phone_number"`
	CountryCode    string            `json:"country_code"`
	IsValidated    bool              `json:"is_validated"`
	Roles          []JWTRole         `json:"roles"`
	Organizations  []JWTOrganization `json:"organizations"`  // Support nested organizations in user_context
	Groups         []JWTGroup        `json:"groups"`         // Support nested groups in user_context
}

// AAATokenClaims represents JWT claims from AAA service (updated to match actual JWT structure)
type AAATokenClaims struct {
	UserID        string            `json:"user_id"`
	Username      string            `json:"username"`
	IsValidated   bool              `json:"isvalidate"`
	RoleIDs       []string          `json:"roleIds"`        // Array of role ID strings (legacy)
	Permissions   interface{}       `json:"permissions"`    // Permissions (may be null)
	UserContext   *JWTUserContext   `json:"user_context"`   // Contains actual role objects and user info
	Organizations []JWTOrganization `json:"organizations"`  // ✅ Organizations (FPO ID is here)
	Groups        []JWTGroup        `json:"groups"`         // ✅ Groups
	Roles         []JWTRole         `json:"roles"`          // ✅ Roles with org/group context
	Scopes        []string          `json:"scopes"`         // ✅ Scopes (org:xxx, group:xxx)
	TokenType     string            `json:"token_type"`
	TokenVersion  string            `json:"token_version"`
	SessionID     string            `json:"session_id"`
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

// ConvertJWTRolesToAAARole converts JWT roles to AAARole structure for backward compatibility
// Note: JWT roles don't contain full role information, so we create minimal AAARole objects
func ConvertJWTRolesToAAARole(jwtRoles []JWTRole) []AAARole {
	if jwtRoles == nil {
		return []AAARole{}
	}

	aaaRoles := make([]AAARole, len(jwtRoles))
	for i, jwtRole := range jwtRoles {
		aaaRoles[i] = AAARole{
			ID:       jwtRole.ID,
			RoleID:   jwtRole.ID,
			IsActive: jwtRole.IsActive,
			Role: AAAUserRole{
				Name:     jwtRole.Name,
				Scope:    jwtRole.Scope,
				IsActive: jwtRole.IsActive,
			},
		}
	}
	return aaaRoles
}

// ExtractPermissions extracts permissions from the full AAARole structure
// Note: JWT tokens don't contain permissions - actual permission checking is done via gRPC
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

// GetOrganizationIDs extracts all organization IDs from token claims
// Returns organization IDs from both direct memberships and role-based access
// Supports BOTH top-level organizations and user_context.organizations for backward compatibility
func (c *AAATokenClaims) GetOrganizationIDs() []string {
	orgMap := make(map[string]bool)

	// Get from top-level direct organizations (new format)
	for _, org := range c.Organizations {
		if org.ID != "" {
			orgMap[org.ID] = true
		}
	}

	// ✅ ALSO check user_context.organizations (AAA service nested format)
	if c.UserContext != nil {
		for _, org := range c.UserContext.Organizations {
			if org.ID != "" {
				orgMap[org.ID] = true
			}
		}
	}

	// Get from top-level roles with organization scope
	for _, role := range c.Roles {
		if role.OrganizationID != nil && *role.OrganizationID != "" {
			orgMap[*role.OrganizationID] = true
		}
	}

	// Also check UserContext roles (for backward compatibility)
	if c.UserContext != nil {
		for _, role := range c.UserContext.Roles {
			if role.OrganizationID != nil && *role.OrganizationID != "" {
				orgMap[*role.OrganizationID] = true
			}
		}
	}

	// Convert map to slice
	orgs := make([]string, 0, len(orgMap))
	for orgID := range orgMap {
		orgs = append(orgs, orgID)
	}

	return orgs
}

// GetPrimaryOrganizationID returns the primary organization ID
// For single-organization users, they should have exactly ONE organization
// Returns empty string if no organizations found
func (c *AAATokenClaims) GetPrimaryOrganizationID() string {
	orgs := c.GetOrganizationIDs()

	if len(orgs) == 0 {
		return ""
	}

	// Return first organization as primary
	// For single-organization users, there should only be one
	return orgs[0]
}

// GetOrganizationName returns the name of the primary organization
// Checks both top-level organizations and user_context.organizations
func (c *AAATokenClaims) GetOrganizationName() string {
	// Check top-level organizations first (new format)
	if len(c.Organizations) > 0 {
		return c.Organizations[0].Name
	}

	// ✅ Fallback to user_context.organizations (AAA service nested format)
	if c.UserContext != nil && len(c.UserContext.Organizations) > 0 {
		return c.UserContext.Organizations[0].Name
	}

	return ""
}

// HasOrganizationAccess checks if user has access to a specific organization
func (c *AAATokenClaims) HasOrganizationAccess(organizationID string) bool {
	// Check direct memberships
	for _, org := range c.Organizations {
		if org.ID == organizationID {
			return true
		}
	}

	// Check role-based access
	for _, role := range c.Roles {
		if role.IsActive && role.OrganizationID != nil && *role.OrganizationID == organizationID {
			return true
		}
	}

	// Check UserContext roles
	if c.UserContext != nil {
		for _, role := range c.UserContext.Roles {
			if role.IsActive && role.OrganizationID != nil && *role.OrganizationID == organizationID {
				return true
			}
		}
	}

	return false
}

// Address represents an address from AAA service (for HTTP API)
type Address struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`

	// Hierarchical Indian Address Fields
	House       *string `json:"house,omitempty"`
	Street      *string `json:"street,omitempty"`
	Landmark    *string `json:"landmark,omitempty"`
	PostOffice  *string `json:"post_office,omitempty"`
	Subdistrict *string `json:"subdistrict,omitempty"`
	District    *string `json:"district,omitempty"`
	VTC         *string `json:"vtc,omitempty"` // Village/Town/City
	State       *string `json:"state,omitempty"`
	Country     *string `json:"country,omitempty"`
	Pincode     *string `json:"pincode,omitempty"`
	FullAddress *string `json:"full_address,omitempty"`

	// Metadata
	Type      string                 `json:"type"`       // HOME, WORK, OTHER
	IsPrimary bool                   `json:"is_primary"`
	IsActive  bool                   `json:"is_active"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateAddressRequest represents a request to create an address
type CreateAddressRequest struct {
	UserID      string                 `json:"user_id"`
	Type        string                 `json:"type"`
	House       *string                `json:"house,omitempty"`
	Street      *string                `json:"street,omitempty"`
	Landmark    *string                `json:"landmark,omitempty"`
	PostOffice  *string                `json:"post_office,omitempty"`
	Subdistrict *string                `json:"subdistrict,omitempty"`
	District    *string                `json:"district,omitempty"`
	VTC         *string                `json:"vtc,omitempty"`
	State       *string                `json:"state,omitempty"`
	Country     *string                `json:"country,omitempty"`
	Pincode     *string                `json:"pincode,omitempty"`
	IsPrimary   bool                   `json:"is_primary"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateAddressRequest represents a request to update an address
type UpdateAddressRequest struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	House       *string                `json:"house,omitempty"`
	Street      *string                `json:"street,omitempty"`
	Landmark    *string                `json:"landmark,omitempty"`
	PostOffice  *string                `json:"post_office,omitempty"`
	Subdistrict *string                `json:"subdistrict,omitempty"`
	District    *string                `json:"district,omitempty"`
	VTC         *string                `json:"vtc,omitempty"`
	State       *string                `json:"state,omitempty"`
	Country     *string                `json:"country,omitempty"`
	Pincode     *string                `json:"pincode,omitempty"`
	IsPrimary   bool                   `json:"is_primary"`
	IsActive    bool                   `json:"is_active"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BuildFullAddress builds a full address string from components
func (a *Address) BuildFullAddress() string {
	parts := []string{}

	if a.House != nil && *a.House != "" {
		parts = append(parts, *a.House)
	}
	if a.Street != nil && *a.Street != "" {
		parts = append(parts, *a.Street)
	}
	if a.Landmark != nil && *a.Landmark != "" {
		parts = append(parts, *a.Landmark)
	}
	if a.PostOffice != nil && *a.PostOffice != "" {
		parts = append(parts, *a.PostOffice)
	}
	if a.Subdistrict != nil && *a.Subdistrict != "" {
		parts = append(parts, *a.Subdistrict)
	}
	if a.District != nil && *a.District != "" {
		parts = append(parts, *a.District)
	}
	if a.VTC != nil && *a.VTC != "" {
		parts = append(parts, *a.VTC)
	}
	if a.State != nil && *a.State != "" {
		parts = append(parts, *a.State)
	}
	if a.Country != nil && *a.Country != "" {
		parts = append(parts, *a.Country)
	}
	if a.Pincode != nil && *a.Pincode != "" {
		parts = append(parts, *a.Pincode)
	}

	return strings.Join(parts, ", ")
}

// IsComplete checks if address has all required fields
func (a *Address) IsComplete() bool {
	return a.House != nil && *a.House != "" &&
		a.Street != nil && *a.Street != "" &&
		a.District != nil && *a.District != "" &&
		a.State != nil && *a.State != "" &&
		a.Country != nil && *a.Country != "" &&
		a.Pincode != nil && *a.Pincode != ""
}
