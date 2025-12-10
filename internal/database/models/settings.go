package models

import (
	"time"
)

// Setting represents a key-value configuration entry for FPO settings
// Used for dynamic invoice header fields and FPO configuration
type Setting struct {
	Key           string    `gorm:"type:varchar(50);primaryKey" json:"key"`
	Value         string    `gorm:"type:text;not null" json:"value"`
	DisplayLabel  *string   `gorm:"type:varchar(100)" json:"display_label"`  // Label for invoice header (e.g., "GSTIN", "Fert. Lic.")
	DisplayOrder  int       `gorm:"type:int;default:0" json:"display_order"` // Order in invoice header
	IsHeaderField bool      `gorm:"type:boolean;default:false" json:"is_header_field"` // Show in invoice header?
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// NewSetting creates a new Setting with the given key and value
func NewSetting(key, value string) *Setting {
	return &Setting{
		Key:           key,
		Value:         value,
		DisplayOrder:  0,
		IsHeaderField: false,
	}
}

// NewHeaderSetting creates a new Setting configured as an invoice header field
func NewHeaderSetting(key, value, displayLabel string, displayOrder int) *Setting {
	return &Setting{
		Key:           key,
		Value:         value,
		DisplayLabel:  &displayLabel,
		DisplayOrder:  displayOrder,
		IsHeaderField: true,
	}
}

func (Setting) TableName() string {
	return "settings"
}

// SettingResponse represents the API response for a setting
type SettingResponse struct {
	Key           string  `json:"key"`
	Value         string  `json:"value"`
	DisplayLabel  *string `json:"display_label,omitempty"`
	DisplayOrder  int     `json:"display_order"`
	IsHeaderField bool    `json:"is_header_field"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// HeaderFieldResponse represents a header field for invoice display
type HeaderFieldResponse struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	DisplayLabel string `json:"display_label"`
	DisplayOrder int    `json:"display_order"`
}

// CreateSettingRequest represents the request to create a setting
type CreateSettingRequest struct {
	Value         string  `json:"value" binding:"required"`
	DisplayLabel  *string `json:"display_label"`
	DisplayOrder  *int    `json:"display_order"`
	IsHeaderField *bool   `json:"is_header_field"`
}

// UpdateSettingRequest represents the request to update a setting
type UpdateSettingRequest struct {
	Value         *string `json:"value"`
	DisplayLabel  *string `json:"display_label"`
	DisplayOrder  *int    `json:"display_order"`
	IsHeaderField *bool   `json:"is_header_field"`
}

// ToResponse converts a Setting to SettingResponse
func (s *Setting) ToResponse() *SettingResponse {
	return &SettingResponse{
		Key:           s.Key,
		Value:         s.Value,
		DisplayLabel:  s.DisplayLabel,
		DisplayOrder:  s.DisplayOrder,
		IsHeaderField: s.IsHeaderField,
		CreatedAt:     s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     s.UpdatedAt.Format(time.RFC3339),
	}
}

// ToHeaderFieldResponse converts a Setting to HeaderFieldResponse
// Returns nil if not a header field or has no display label
func (s *Setting) ToHeaderFieldResponse() *HeaderFieldResponse {
	if !s.IsHeaderField || s.DisplayLabel == nil {
		return nil
	}
	return &HeaderFieldResponse{
		Key:          s.Key,
		Value:        s.Value,
		DisplayLabel: *s.DisplayLabel,
		DisplayOrder: s.DisplayOrder,
	}
}

// InvoiceRequirementsResponse represents the response for invoice requirements check
type InvoiceRequirementsResponse struct {
	Ready           bool     `json:"ready"`
	MissingSettings []string `json:"missing_settings,omitempty"`
}

// Predefined setting keys for FPO configuration
const (
	SettingKeyFPOName              = "fpo_name"
	SettingKeyFPOLogoURL           = "fpo_logo_url"
	SettingKeyFPOBranchAddress     = "fpo_branch_address"
	SettingKeyFPORegisteredAddress = "fpo_registered_address"
	SettingKeyFPOGSTIN             = "fpo_gstin"
	SettingKeyFPOBankAccount       = "fpo_bank_account"
)

// RequiredSettingsForInvoice returns the list of required setting keys for invoice generation
func RequiredSettingsForInvoice() []string {
	return []string{
		SettingKeyFPOName,
		SettingKeyFPOLogoURL,
		SettingKeyFPOBranchAddress,
	}
}
