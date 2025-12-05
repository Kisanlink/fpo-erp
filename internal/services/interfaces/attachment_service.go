package interfaces

import (
	"context"
	"kisanlink-erp/internal/database/models"
	"mime/multipart"
	"time"
)

type AttachmentServiceInterface interface {
	UploadAttachment(ctx context.Context, file *multipart.FileHeader, entityType, entityID, uploadedBy string) (*models.AttachmentResponse, error)
	GetAttachment(id string) (*models.Attachment, error)
	GetAttachments(ctx context.Context, entityType, entityID *string, limit, offset int) ([]models.AttachmentResponse, error)
	GetAttachmentsByEntity(ctx context.Context, entityType, entityID string) ([]models.AttachmentResponse, error)
	DeleteAttachment(ctx context.Context, id string) error
	DownloadAttachment(ctx context.Context, id string) (interface{}, string, error)
	GenerateDownloadURL(ctx context.Context, id string, expiration time.Duration) (string, error)
	GetAttachmentInfo(ctx context.Context, id string) (*models.AttachmentInfoResponse, error)
}
