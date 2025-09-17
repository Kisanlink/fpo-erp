package aaa

// Address represents an address from the AAA service
type Address struct {
	ID           string            `json:"id"`
	UserID       string            `json:"user_id"`
	Type         string            `json:"type"` // HOME, WORK, OTHER
	AddressLine1 string            `json:"address_line_1"`
	AddressLine2 string            `json:"address_line_2"`
	City         string            `json:"city"`
	State        string            `json:"state"`
	PostalCode   string            `json:"postal_code"`
	Country      string            `json:"country"`
	IsPrimary    bool              `json:"is_primary"`
	IsActive     bool              `json:"is_active"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// CreateAddressRequest represents a request to create an address
type CreateAddressRequest struct {
	UserID       string            `json:"user_id"`
	Type         string            `json:"type"`
	AddressLine1 string            `json:"address_line_1"`
	AddressLine2 string            `json:"address_line_2"`
	City         string            `json:"city"`
	State        string            `json:"state"`
	PostalCode   string            `json:"postal_code"`
	Country      string            `json:"country"`
	IsPrimary    bool              `json:"is_primary"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// UpdateAddressRequest represents a request to update an address
type UpdateAddressRequest struct {
	ID           string            `json:"id"`
	Type         string            `json:"type,omitempty"`
	AddressLine1 string            `json:"address_line_1,omitempty"`
	AddressLine2 string            `json:"address_line_2,omitempty"`
	City         string            `json:"city,omitempty"`
	State        string            `json:"state,omitempty"`
	PostalCode   string            `json:"postal_code,omitempty"`
	Country      string            `json:"country,omitempty"`
	IsPrimary    bool              `json:"is_primary,omitempty"`
	IsActive     bool              `json:"is_active,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// GetAddressRequest represents a request to get an address
type GetAddressRequest struct {
	ID string `json:"id"`
}

// DeleteAddressRequest represents a request to delete an address
type DeleteAddressRequest struct {
	ID         string `json:"id"`
	SoftDelete bool   `json:"soft_delete"`
}
