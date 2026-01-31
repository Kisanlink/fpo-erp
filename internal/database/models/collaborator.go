package models

import (
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/constants"
	"strings"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Collaborator represents a vendor/supplier in the system
type Collaborator struct {
	base.BaseModel

	// E-commerce Integration
	ExternalID *string `gorm:"type:varchar(100);unique;index" json:"external_id"` // E-commerce supplier ID for webhook matching

	// Business Info
	CompanyName   string  `gorm:"type:varchar(150);not null" json:"company_name"`
	Logo          *string `gorm:"type:varchar(2000)" json:"logo"` // Presigned S3 URL (auto-synced on upload)
	ContactPerson string  `gorm:"type:varchar(100);not null" json:"contact_person"`
	ContactNumber string  `gorm:"type:varchar(20);not null" json:"contact_number"`
	Email         *string `gorm:"type:varchar(255)" json:"email"`

	// Legal/Compliance
	GSTNumber string  `gorm:"type:varchar(15)" json:"gst_number"` // GSTIN
	PANNumber *string `gorm:"type:varchar(10)" json:"pan_number"`

	// Banking
	BankAccountNo *string `gorm:"type:varchar(50)" json:"bank_account_no"`
	BankIFSC      *string `gorm:"type:varchar(11)" json:"bank_ifsc"`
	BankName      *string `gorm:"type:varchar(100)" json:"bank_name"`

	// Address (AAA integration - similar to Warehouse)
	AddressID *string `gorm:"type:varchar(50)" json:"address_id"` // Reference to AAA address

	// Local address cache (synced on write operations) - eliminates gRPC calls on GET
	AddressType *string `gorm:"type:varchar(20);column:address_type" json:"-"`
	House       *string `gorm:"type:varchar(255);column:house" json:"-"`
	Street      *string `gorm:"type:varchar(255);column:street" json:"-"`
	Landmark    *string `gorm:"type:varchar(255);column:landmark" json:"-"`
	PostOffice  *string `gorm:"type:varchar(255);column:post_office" json:"-"`
	Subdistrict *string `gorm:"type:varchar(255);column:subdistrict" json:"-"`
	District    *string `gorm:"type:varchar(255);column:district" json:"-"`
	VTC         *string `gorm:"type:varchar(255);column:vtc" json:"-"`
	State       *string `gorm:"type:varchar(100);column:state" json:"-"`
	Country     *string `gorm:"type:varchar(100);column:country" json:"-"`
	Pincode     *string `gorm:"type:varchar(10);column:pincode" json:"-"`

	// Metadata
	Experience *string `gorm:"type:text" json:"experience"`   // Description
	IsActive   *bool   `gorm:"type:boolean" json:"is_active"` // Pointer to allow explicit false values

	// Associations
	Products []CollaboratorProduct `gorm:"foreignKey:CollaboratorID" json:"products,omitempty"`
}

// NewCollaborator creates a new Collaborator with initialized fields
func NewCollaborator(companyName, contactPerson, contactNumber string, bankAccountNo, bankIFSC *string, addressID *string) *Collaborator {
	baseModel := base.NewBaseModel(constants.TableCollaborator, hash.Medium)
	isActive := true
	return &Collaborator{
		BaseModel:     *baseModel,
		CompanyName:   companyName,
		ContactPerson: contactPerson,
		ContactNumber: contactNumber,
		BankAccountNo: bankAccountNo,
		BankIFSC:      bankIFSC,
		AddressID:     addressID,
		IsActive:      &isActive,
	}
}

func (Collaborator) TableName() string {
	return "collaborators"
}

// BuildAddressInfo constructs AddressInfo from local fields (no gRPC needed)
func (c *Collaborator) BuildAddressInfo() *AddressInfo {
	if c.AddressID == nil {
		return nil
	}
	return &AddressInfo{
		ID:          *c.AddressID,
		Type:        ptrStringValue(c.AddressType),
		House:       c.House,
		Street:      c.Street,
		Landmark:    c.Landmark,
		PostOffice:  c.PostOffice,
		Subdistrict: c.Subdistrict,
		District:    c.District,
		VTC:         c.VTC,
		State:       c.State,
		Country:     c.Country,
		Pincode:     c.Pincode,
		FullAddress: c.buildFullAddress(),
	}
}

// buildFullAddress constructs a full address string from local fields
func (c *Collaborator) buildFullAddress() string {
	parts := []string{}
	addPart := func(s *string) {
		if s != nil && *s != "" {
			parts = append(parts, *s)
		}
	}
	addPart(c.House)
	addPart(c.Street)
	addPart(c.Landmark)
	addPart(c.VTC)
	addPart(c.District)
	addPart(c.State)
	addPart(c.Pincode)
	addPart(c.Country)
	return strings.Join(parts, ", ")
}

// SyncFromAAA populates local fields from AAA address response
func (c *Collaborator) SyncFromAAA(addr *aaa.Address) {
	if addr == nil {
		return
	}
	c.AddressType = &addr.Type
	c.House = addr.House
	c.Street = addr.Street
	c.Landmark = addr.Landmark
	c.PostOffice = addr.PostOffice
	c.Subdistrict = addr.Subdistrict
	c.District = addr.District
	c.VTC = addr.VTC
	c.State = addr.State
	c.Country = addr.Country
	c.Pincode = addr.Pincode
}

// HasAddressCache returns true if local address fields are populated
// Used for lazy-fetch detection on GET operations for legacy records
func (c *Collaborator) HasAddressCache() bool {
	if c.AddressID == nil {
		return true // No address = nothing to cache
	}
	// State is always required for valid Indian addresses
	return c.State != nil && *c.State != ""
}

// CollaboratorResponse represents the API response for collaborator
type CollaboratorResponse struct {
	ID            string       `json:"id"`
	ExternalID    *string      `json:"external_id,omitempty"`
	AddressID     *string      `json:"address_id,omitempty"`
	CompanyName   string       `json:"company_name"`
	Logo          *string      `json:"logo"`
	ContactPerson string       `json:"contact_person"`
	ContactNumber string       `json:"contact_number"`
	Email         *string      `json:"email"`
	GSTNumber     string       `json:"gst_number"`
	PANNumber     *string      `json:"pan_number"`
	BankAccountNo *string      `json:"bank_account_no"`
	BankIFSC      *string      `json:"bank_ifsc"`
	BankName      *string      `json:"bank_name"`
	Experience    *string      `json:"experience"`
	IsActive      bool         `json:"is_active"`
	Address       *AddressInfo `json:"address,omitempty"` // Embedded from AAA
	CreatedAt     string       `json:"created_at"`
	UpdatedAt     string       `json:"updated_at"`
}

// CreateCollaboratorRequest represents the request to create a collaborator
type CreateCollaboratorRequest struct {
	CompanyName   string                `json:"company_name" binding:"required"`
	Logo          *string               `json:"logo"`
	ContactPerson string                `json:"contact_person" binding:"required"`
	ContactNumber string                `json:"contact_number" binding:"required"`
	Email         *string               `json:"email" binding:"omitempty,email"`
	GSTNumber     string                `json:"gst_number"`
	PANNumber     *string               `json:"pan_number"`
	BankAccountNo *string               `json:"bank_account_no"`
	BankIFSC      *string               `json:"bank_ifsc" binding:"omitempty,len=11"`
	BankName      *string               `json:"bank_name"`
	Experience    *string               `json:"experience"`
	Address       *CreateAddressRequest `json:"address"` // Inline address creation via AAA
}

// UpdateCollaboratorRequest represents the request to update a collaborator
type UpdateCollaboratorRequest struct {
	CompanyName   *string               `json:"company_name,omitempty"`
	Logo          *string               `json:"logo,omitempty"`
	ContactPerson *string               `json:"contact_person,omitempty"`
	ContactNumber *string               `json:"contact_number,omitempty"`
	Email         *string               `json:"email,omitempty" binding:"omitempty,email"`
	GSTNumber     *string               `json:"gst_number,omitempty"`
	PANNumber     *string               `json:"pan_number,omitempty"`
	BankAccountNo *string               `json:"bank_account_no,omitempty"`
	BankIFSC      *string               `json:"bank_ifsc,omitempty" binding:"omitempty,len=11"`
	BankName      *string               `json:"bank_name,omitempty"`
	Experience    *string               `json:"experience,omitempty"`
	IsActive      *bool                 `json:"is_active,omitempty"`
	Address       *UpdateAddressRequest `json:"address,omitempty"`
}

// CollaboratorStats holds transaction statistics for a collaborator
type CollaboratorStats struct {
	CollaboratorID string  `json:"collaborator_id"`
	CompanyName    string  `json:"company_name"`
	POCount        int64   `json:"po_count"`
	GRNCount       int64   `json:"grn_count"`
	TotalAmount    float64 `json:"total_amount"`
	ActivePOCount  int64   `json:"active_po_count"`
	LastPODate     *string `json:"last_po_date"`
}

// CollaboratorStatsSummary holds simplified stats for a collaborator (for bulk stats endpoint)
type CollaboratorStatsSummary struct {
	CollaboratorID string `json:"collaborator_id"`
	CompanyName    string `json:"company_name"`
	POCount        int64  `json:"po_count"`
}

// AllCollaboratorsStatsResponse holds stats for all collaborators
type AllCollaboratorsStatsResponse struct {
	Collaborators []CollaboratorStatsSummary `json:"collaborators"`
	TotalPOCount  int64                      `json:"total_po_count"`
}
