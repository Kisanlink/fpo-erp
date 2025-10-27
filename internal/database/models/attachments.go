package models

import (
	"time"

	"kisanlink-erp/internal/constants"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Attachment represents a file attachment (generic entity-based system)
type Attachment struct {
	base.BaseModel
	EntityType string    `gorm:"type:varchar(50);not null;index:idx_attachment_entity" json:"entity_type"` // "logo", "po", "grn", etc.
	EntityID   string    `gorm:"type:varchar(100);not null;index:idx_attachment_entity" json:"entity_id"`  // Entity ID (CLAB_xxx, PO_xxx, GRN_xxx, etc.)
	FilePath   string    `gorm:"type:text;not null" json:"file_path"`                                       // S3 key/path
	FileType   string    `gorm:"type:varchar(50);not null" json:"file_type"`                                // MIME type
	UploadedBy *string   `gorm:"type:varchar(100)" json:"uploaded_by"`                                      // User ID from AAA
	UploadedAt time.Time `gorm:"type:timestamptz;not null;default:now()" json:"uploaded_at"`
}

func (Attachment) TableName() string {
	return "attachments"
}

// NewAttachment creates a new Attachment with initialized fields
func NewAttachment(entityType, entityID, filePath, fileType string, uploadedBy *string, uploadedAt time.Time) *Attachment {
	baseModel := base.NewBaseModel(constants.TableAttachment, hash.Medium)
	return &Attachment{
		BaseModel:  *baseModel,
		EntityType: entityType,
		EntityID:   entityID,
		FilePath:   filePath,
		FileType:   fileType,
		UploadedBy: uploadedBy,
		UploadedAt: uploadedAt,
	}
}

// AttachmentResponse represents the API response for attachment
type AttachmentResponse struct {
	ID         string  `json:"id"`
	EntityType string  `json:"entity_type"` // "logo", "po", "grn", etc.
	EntityID   string  `json:"entity_id"`   // Entity ID (CLAB_xxx, PO_xxx, etc.)
	FilePath   string  `json:"file_path"`   // S3 key/path
	FileType   string  `json:"file_type"`   // MIME type
	UploadedBy *string `json:"uploaded_by"` // User ID
	UploadedAt string  `json:"uploaded_at"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// AttachmentInfoResponse represents detailed attachment information
type AttachmentInfoResponse struct {
	ID         string  `json:"id"`
	EntityType string  `json:"entity_type"` // "logo", "po", "grn", etc.
	EntityID   string  `json:"entity_id"`   // Entity ID (CLAB_xxx, PO_xxx, etc.)
	FilePath   string  `json:"file_path"`   // S3 key/path
	FileType   string  `json:"file_type"`   // MIME type
	FileSize   int64   `json:"file_size"`   // File size in bytes
	UploadedBy *string `json:"uploaded_by"` // User ID
	UploadedAt string  `json:"uploaded_at"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}
