package models

import (
	"kisanlink-erp/internal/constants"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Warehouse represents a storage location
type Warehouse struct {
	base.BaseModel
	Name      string  `gorm:"type:varchar(100);not null" json:"name"`
	AddressID *string `gorm:"type:varchar(50)" json:"address_id"` // Reference to AAA address
}

// NewWarehouse creates a new Warehouse with initialized fields
func NewWarehouse(name string, addressID *string) *Warehouse {
	baseModel := base.NewBaseModel(constants.TableWarehouse, hash.Medium)
	return &Warehouse{
		BaseModel: *baseModel,
		Name:      name,
		AddressID: addressID,
	}
}

func (Warehouse) TableName() string {
	return "warehouses"
}

// AddressInfo represents address information from AAA service
type AddressInfo struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
	FullAddress  string `json:"full_address"`
}

// WarehouseResponse represents the API response for warehouse
type WarehouseResponse struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Address   *AddressInfo `json:"address,omitempty"` // Embedded address info
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at"`
}

// CreateWarehouseRequest represents the request to create a warehouse
type CreateWarehouseRequest struct {
	Name      string                `json:"name" binding:"required"`
	AddressID *string               `json:"address_id,omitempty"` // Optional: can create address inline
	Address   *CreateAddressRequest `json:"address,omitempty"`    // Inline address creation
}

// CreateAddressRequest for inline address creation
type CreateAddressRequest struct {
	Type         string `json:"type" binding:"required"` // HOME, WORK, OTHER
	AddressLine1 string `json:"address_line_1" binding:"required"`
	AddressLine2 string `json:"address_line_2,omitempty"`
	City         string `json:"city" binding:"required"`
	State        string `json:"state" binding:"required"`
	PostalCode   string `json:"postal_code" binding:"required"`
	Country      string `json:"country" binding:"required"`
	IsPrimary    bool   `json:"is_primary"`
}

// UpdateWarehouseRequest represents the request to update a warehouse
type UpdateWarehouseRequest struct {
	Name      *string               `json:"name,omitempty"`
	AddressID *string               `json:"address_id,omitempty"`
	Address   *UpdateAddressRequest `json:"address,omitempty"`
}

// UpdateAddressRequest for inline address updates
type UpdateAddressRequest struct {
	ID           string `json:"id" binding:"required"`
	Type         string `json:"type,omitempty"`
	AddressLine1 string `json:"address_line_1,omitempty"`
	AddressLine2 string `json:"address_line_2,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
	Country      string `json:"country,omitempty"`
	IsPrimary    *bool  `json:"is_primary,omitempty"`
}
