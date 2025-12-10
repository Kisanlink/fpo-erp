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

// AddressInfo represents address information from AAA service - Indian hierarchical format
type AddressInfo struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
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
	FullAddress string  `json:"full_address"`
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

// CreateAddressRequest for inline address creation - Indian hierarchical format
type CreateAddressRequest struct {
	Type        string  `json:"type" binding:"required"` // HOME, WORK, OTHER
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
	IsPrimary   bool    `json:"is_primary"`
}

// UpdateWarehouseRequest represents the request to update a warehouse
type UpdateWarehouseRequest struct {
	Name      *string               `json:"name,omitempty"`
	AddressID *string               `json:"address_id,omitempty"`
	Address   *UpdateAddressRequest `json:"address,omitempty"`
}

// UpdateAddressRequest for inline address updates - Indian hierarchical format
type UpdateAddressRequest struct {
	ID          string  `json:"id" binding:"required"`
	Type        string  `json:"type,omitempty"`
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
	IsPrimary   *bool   `json:"is_primary,omitempty"`
}
