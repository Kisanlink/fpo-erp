package models

import (
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/constants"
	"strings"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Warehouse represents a storage location
type Warehouse struct {
	base.BaseModel
	Name      string  `gorm:"type:varchar(100);not null" json:"name"`
	AddressID *string `gorm:"type:varchar(50)" json:"address_id"` // Reference to AAA address

	// Local address cache (synced on write operations) - eliminates gRPC calls on GET
	AddressType  *string `gorm:"type:varchar(20);column:address_type" json:"-"`
	House        *string `gorm:"type:varchar(255);column:house" json:"-"`
	Street       *string `gorm:"type:varchar(255);column:street" json:"-"`
	Landmark     *string `gorm:"type:varchar(255);column:landmark" json:"-"`
	PostOffice   *string `gorm:"type:varchar(255);column:post_office" json:"-"`
	Subdistrict  *string `gorm:"type:varchar(255);column:subdistrict" json:"-"`
	District     *string `gorm:"type:varchar(255);column:district" json:"-"`
	VTC          *string `gorm:"type:varchar(255);column:vtc" json:"-"`
	State        *string `gorm:"type:varchar(100);column:state" json:"-"`
	Country      *string `gorm:"type:varchar(100);column:country" json:"-"`
	Pincode      *string `gorm:"type:varchar(10);column:pincode" json:"-"`
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

// BuildAddressInfo constructs AddressInfo from local fields (no gRPC needed)
func (w *Warehouse) BuildAddressInfo() *AddressInfo {
	if w.AddressID == nil {
		return nil
	}
	return &AddressInfo{
		ID:          *w.AddressID,
		Type:        ptrStringValue(w.AddressType),
		House:       w.House,
		Street:      w.Street,
		Landmark:    w.Landmark,
		PostOffice:  w.PostOffice,
		Subdistrict: w.Subdistrict,
		District:    w.District,
		VTC:         w.VTC,
		State:       w.State,
		Country:     w.Country,
		Pincode:     w.Pincode,
		FullAddress: w.buildFullAddress(),
	}
}

// buildFullAddress constructs a full address string from local fields
func (w *Warehouse) buildFullAddress() string {
	parts := []string{}
	addPart := func(s *string) {
		if s != nil && *s != "" {
			parts = append(parts, *s)
		}
	}
	addPart(w.House)
	addPart(w.Street)
	addPart(w.Landmark)
	addPart(w.VTC)
	addPart(w.District)
	addPart(w.State)
	addPart(w.Pincode)
	addPart(w.Country)
	return strings.Join(parts, ", ")
}

// SyncFromAAA populates local fields from AAA address response
func (w *Warehouse) SyncFromAAA(addr *aaa.Address) {
	if addr == nil {
		return
	}
	w.AddressType = &addr.Type
	w.House = addr.House
	w.Street = addr.Street
	w.Landmark = addr.Landmark
	w.PostOffice = addr.PostOffice
	w.Subdistrict = addr.Subdistrict
	w.District = addr.District
	w.VTC = addr.VTC
	w.State = addr.State
	w.Country = addr.Country
	w.Pincode = addr.Pincode
}

// ptrStringValue returns the value of a string pointer or empty string if nil
func ptrStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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
