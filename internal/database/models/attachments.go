package models

import (
	"time"

	"kisanlink-erp/internal/constants"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Attachment represents a file attachment
type Attachment struct {
	base.BaseModel
	SaleID     *string   `gorm:"type:varchar(100)" json:"sale_id"`
	ReturnID   *string   `gorm:"type:varchar(100)" json:"return_id"`
	FilePath   string    `gorm:"type:text;not null" json:"file_path"`
	FileType   string    `gorm:"type:varchar(50);not null" json:"file_type"`
	UploadedBy *string   `gorm:"type:varchar(100)" json:"uploaded_by"`
	UploadedAt time.Time `gorm:"type:timestamptz;not null;default:now()" json:"uploaded_at"`

	// Associations
	Sale   *Sale   `gorm:"foreignKey:SaleID" json:"sale,omitempty"`
	Return *Return `gorm:"foreignKey:ReturnID" json:"return,omitempty"`
}

func (Attachment) TableName() string {
	return "attachments"
}

// NewAttachment creates a new Attachment with initialized fields
func NewAttachment(saleID, returnID *string, filePath, fileType string, uploadedBy *string, uploadedAt time.Time) *Attachment {
	baseModel := base.NewBaseModel(constants.TableAttachment, hash.Medium)
	return &Attachment{
		BaseModel:  *baseModel,
		SaleID:     saleID,
		ReturnID:   returnID,
		FilePath:   filePath,
		FileType:   fileType,
		UploadedBy: uploadedBy,
		UploadedAt: uploadedAt,
	}
}

// AttachmentResponse represents the API response for attachment
type AttachmentResponse struct {
	ID         string  `json:"id"`
	SaleID     *string `json:"sale_id"`
	ReturnID   *string `json:"return_id"`
	FilePath   string  `json:"file_path"`
	FileType   string  `json:"file_type"`
	UploadedBy *string `json:"uploaded_by"`
	UploadedAt string  `json:"uploaded_at"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// AttachmentInfoResponse represents detailed attachment information
type AttachmentInfoResponse struct {
	ID         string  `json:"id"`
	SaleID     *string `json:"sale_id"`
	ReturnID   *string `json:"return_id"`
	FilePath   string  `json:"file_path"`
	FileType   string  `json:"file_type"`
	FileSize   int64   `json:"file_size"`
	UploadedBy *string `json:"uploaded_by"`
	UploadedAt string  `json:"uploaded_at"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}
