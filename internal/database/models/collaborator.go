package models

import (
	"kisanlink-erp/internal/constants"

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
	Logo          *string `gorm:"type:varchar(500)" json:"logo"` // Attachment ID (ATT_xxxxxxxx) - Use /api/v1/attachments/{id}/url to get image URL
	ContactPerson string  `gorm:"type:varchar(100);not null" json:"contact_person"`
	ContactNumber string  `gorm:"type:varchar(20);not null" json:"contact_number"`
	Email         *string `gorm:"type:varchar(255)" json:"email"`

	// Legal/Compliance
	GSTNumber string  `gorm:"type:varchar(15)" json:"gst_number"` // GSTIN
	PANNumber *string `gorm:"type:varchar(10)" json:"pan_number"`

	// Banking
	BankAccountNo string  `gorm:"type:varchar(50);not null" json:"bank_account_no"`
	BankIFSC      string  `gorm:"type:varchar(11);not null" json:"bank_ifsc"`
	BankName      *string `gorm:"type:varchar(100)" json:"bank_name"`

	// Address (AAA integration - similar to Warehouse)
	AddressID *string `gorm:"type:varchar(50)" json:"address_id"` // Reference to AAA address

	// Metadata
	Experience *string `gorm:"type:text" json:"experience"` // Description
	IsActive   bool    `gorm:"default:true" json:"is_active"`

	// Associations
	Products []CollaboratorProduct `gorm:"foreignKey:CollaboratorID" json:"products,omitempty"`
}

// NewCollaborator creates a new Collaborator with initialized fields
func NewCollaborator(companyName, contactPerson, contactNumber, bankAccountNo, bankIFSC string, addressID *string) *Collaborator {
	baseModel := base.NewBaseModel(constants.TableCollaborator, hash.Medium)
	return &Collaborator{
		BaseModel:     *baseModel,
		CompanyName:   companyName,
		ContactPerson: contactPerson,
		ContactNumber: contactNumber,
		BankAccountNo: bankAccountNo,
		BankIFSC:      bankIFSC,
		AddressID:     addressID,
		IsActive:      true,
	}
}

func (Collaborator) TableName() string {
	return "collaborators"
}

// CollaboratorResponse represents the API response for collaborator
type CollaboratorResponse struct {
	ID            string       `json:"id"`
	CompanyName   string       `json:"company_name"`
	Logo          *string      `json:"logo"`
	ContactPerson string       `json:"contact_person"`
	ContactNumber string       `json:"contact_number"`
	Email         *string      `json:"email"`
	GSTNumber     string       `json:"gst_number"`
	PANNumber     *string      `json:"pan_number"`
	BankAccountNo string       `json:"bank_account_no"`
	BankIFSC      string       `json:"bank_ifsc"`
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
	Email         *string               `json:"email"`
	GSTNumber     string                `json:"gst_number"`
	PANNumber     *string               `json:"pan_number"`
	BankAccountNo string                `json:"bank_account_no" binding:"required"`
	BankIFSC      string                `json:"bank_ifsc" binding:"required,len=11"`
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
	Email         *string               `json:"email,omitempty"`
	GSTNumber     *string               `json:"gst_number,omitempty"`
	PANNumber     *string               `json:"pan_number,omitempty"`
	BankAccountNo *string               `json:"bank_account_no,omitempty"`
	BankIFSC      *string               `json:"bank_ifsc,omitempty"`
	BankName      *string               `json:"bank_name,omitempty"`
	Experience    *string               `json:"experience,omitempty"`
	IsActive      *bool                 `json:"is_active,omitempty"`
	Address       *UpdateAddressRequest `json:"address,omitempty"`
}
