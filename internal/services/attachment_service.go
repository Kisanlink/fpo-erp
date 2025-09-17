package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
)

// AttachmentService handles attachment business logic
type AttachmentService struct {
	attachmentRepo *repositories.AttachmentRepository
	s3Service      *S3Service
}

// NewAttachmentService creates a new attachment service
func NewAttachmentService(attachmentRepo *repositories.AttachmentRepository, s3Service *S3Service) *AttachmentService {
	return &AttachmentService{
		attachmentRepo: attachmentRepo,
		s3Service:      s3Service,
	}
}

// UploadAttachment uploads a file and creates an attachment record
func (s *AttachmentService) UploadAttachment(ctx context.Context, file *multipart.FileHeader, saleID, returnID *string, uploadedBy string) (*models.Attachment, error) {
	// Validate file
	if err := s.validateFile(file); err != nil {
		return nil, err
	}

	// Validate that either saleID or returnID is provided, but not both
	if saleID == nil && returnID == nil {
		return nil, errors.NewBadRequestError("either sale_id or return_id must be provided")
	}
	if saleID != nil && returnID != nil {
		return nil, errors.NewBadRequestError("cannot provide both sale_id and return_id")
	}

	// Use uploadedBy string directly
	uploadedByStr := uploadedBy

	// Upload file to S3
	s3URL, err := s.s3Service.UploadFile(ctx, file, saleID, returnID)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create attachment record using the proper constructor
	attachment := models.NewAttachment(saleID, returnID, s3URL, s.s3Service.GetContentType(file), &uploadedByStr, time.Now())

	if err := s.attachmentRepo.Create(attachment); err != nil {
		// If database creation fails, delete the uploaded file
		if deleteErr := s.s3Service.DeleteFile(ctx, s3URL); deleteErr != nil {
			// Log the deletion error but return the original error
			fmt.Printf("Failed to delete S3 file after database error: %v", deleteErr)
		}
		return nil, fmt.Errorf("failed to create attachment record: %w", err)
	}

	return attachment, nil
}

// GetAttachment retrieves an attachment by ID
func (s *AttachmentService) GetAttachment(id string) (*models.Attachment, error) {
	attachment, err := s.attachmentRepo.GetByID(id)
	if err != nil {
		return nil, errors.NewNotFoundError("attachment not found")
	}

	return attachment, nil
}

// GetAttachments retrieves attachments with optional filters
func (s *AttachmentService) GetAttachments(saleID, returnID *string, limit, offset int) ([]models.Attachment, error) {
	attachments, err := s.attachmentRepo.GetAll(saleID, returnID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve attachments: %w", err)
	}

	return attachments, nil
}

// DeleteAttachment deletes an attachment and its associated file
func (s *AttachmentService) DeleteAttachment(ctx context.Context, id string) error {
	// Get attachment to get the S3 URL
	attachment, err := s.attachmentRepo.GetByID(id)
	if err != nil {
		return errors.NewNotFoundError("attachment not found")
	}

	// Delete file from S3
	if err := s.s3Service.DeleteFile(ctx, attachment.FilePath); err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	// Delete attachment record
	if err := s.attachmentRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete attachment record: %w", err)
	}

	return nil
}

// DownloadAttachment downloads an attachment file
func (s *AttachmentService) DownloadAttachment(ctx context.Context, id string) (interface{}, string, error) {
	// Get attachment
	attachment, err := s.attachmentRepo.GetByID(id)
	if err != nil {
		return nil, "", errors.NewNotFoundError("attachment not found")
	}

	// Download file from S3
	fileReader, contentType, err := s.s3Service.DownloadFile(ctx, attachment.FilePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download file: %w", err)
	}

	return fileReader, contentType, nil
}

// GenerateDownloadURL generates a presigned URL for file download
func (s *AttachmentService) GenerateDownloadURL(ctx context.Context, id string, expiration time.Duration) (string, error) {
	// Get attachment
	attachment, err := s.attachmentRepo.GetByID(id)
	if err != nil {
		return "", errors.NewNotFoundError("attachment not found")
	}

	// Generate presigned URL
	url, err := s.s3Service.GeneratePresignedURL(ctx, attachment.FilePath, expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return url, nil
}

// GetAttachmentInfo gets information about an attachment
func (s *AttachmentService) GetAttachmentInfo(ctx context.Context, id string) (*AttachmentInfo, error) {
	// Get attachment
	attachment, err := s.attachmentRepo.GetByID(id)
	if err != nil {
		return nil, errors.NewNotFoundError("attachment not found")
	}

	// Get file info from S3
	fileInfo, err := s.s3Service.GetFileInfo(ctx, attachment.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &AttachmentInfo{
		ID:           attachment.ID,
		SaleID:       attachment.SaleID,
		ReturnID:     attachment.ReturnID,
		FilePath:     attachment.FilePath,
		FileType:     attachment.FileType,
		UploadedBy:   attachment.UploadedBy,
		UploadedAt:   attachment.UploadedAt,
		FileSize:     fileInfo.Size,
		LastModified: fileInfo.LastModified,
		Metadata:     fileInfo.Metadata,
	}, nil
}

// GetAttachmentsBySale retrieves all attachments for a sale
func (s *AttachmentService) GetAttachmentsBySale(saleID string) ([]models.Attachment, error) {
	attachments, err := s.attachmentRepo.GetBySaleID(saleID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sale attachments: %w", err)
	}

	return attachments, nil
}

// GetAttachmentsByReturn retrieves all attachments for a return
func (s *AttachmentService) GetAttachmentsByReturn(returnID string) ([]models.Attachment, error) {
	attachments, err := s.attachmentRepo.GetByReturnID(returnID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve return attachments: %w", err)
	}

	return attachments, nil
}

// validateFile validates the uploaded file
func (s *AttachmentService) validateFile(file *multipart.FileHeader) error {
	// Check file size (10MB limit)
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if file.Size > maxFileSize {
		return errors.NewBadRequestError("file size exceeds maximum limit of 10MB")
	}

	// Check file type
	if err := s.s3Service.ValidateFileType(file.Filename); err != nil {
		return errors.NewBadRequestError(err.Error())
	}

	return nil
}

// AttachmentInfo represents comprehensive attachment information
type AttachmentInfo struct {
	ID           string            `json:"id"`
	SaleID       *string           `json:"sale_id"`
	ReturnID     *string           `json:"return_id"`
	FilePath     string            `json:"file_path"`
	FileType     string            `json:"file_type"`
	UploadedBy   *string           `json:"uploaded_by"`
	UploadedAt   time.Time         `json:"uploaded_at"`
	FileSize     int64             `json:"file_size"`
	LastModified time.Time         `json:"last_modified"`
	Metadata     map[string]string `json:"metadata"`
}
