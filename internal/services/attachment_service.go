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
func (s *AttachmentService) UploadAttachment(ctx context.Context, file *multipart.FileHeader, entityType, entityID, uploadedBy string) (*models.AttachmentResponse, error) {
	// Validate file
	if err := s.validateFile(file); err != nil {
		return nil, err
	}

	// Validate entity_type and entity_id
	if entityType == "" {
		return nil, errors.NewBadRequestError("entity_type is required")
	}
	if entityID == "" {
		return nil, errors.NewBadRequestError("entity_id is required")
	}

	// Upload file to S3 with entity-based folder structure
	s3Key, err := s.s3Service.UploadFile(ctx, file, entityType, entityID)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to upload file")
	}

	// Create attachment record using the proper constructor
	var uploadedByPtr *string
	if uploadedBy != "" {
		uploadedByPtr = &uploadedBy
	}

	attachment := models.NewAttachment(entityType, entityID, s3Key, s.s3Service.GetContentType(file), uploadedByPtr, time.Now())

	if err := s.attachmentRepo.Create(attachment); err != nil {
		// If database creation fails, delete the uploaded file
		if deleteErr := s.s3Service.DeleteFile(ctx, s3Key); deleteErr != nil {
			// Log the deletion error but return the original error
			fmt.Printf("Failed to delete S3 file after database error: %v", deleteErr)
		}
		return nil, errors.NewInternalServerError("failed to create attachment record")
	}

	// Build response
	return &models.AttachmentResponse{
		ID:         attachment.ID,
		EntityType: attachment.EntityType,
		EntityID:   attachment.EntityID,
		FilePath:   attachment.FilePath,
		FileType:   attachment.FileType,
		UploadedBy: attachment.UploadedBy,
		UploadedAt: attachment.UploadedAt.UTC().Format(time.RFC3339),
		CreatedAt:  attachment.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  attachment.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
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
func (s *AttachmentService) GetAttachments(entityType, entityID *string, limit, offset int) ([]models.AttachmentResponse, error) {
	attachments, err := s.attachmentRepo.GetAll(entityType, entityID, limit, offset)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to retrieve attachments")
	}

	// Build response
	responses := make([]models.AttachmentResponse, len(attachments))
	for i, att := range attachments {
		responses[i] = models.AttachmentResponse{
			ID:         att.ID,
			EntityType: att.EntityType,
			EntityID:   att.EntityID,
			FilePath:   att.FilePath,
			FileType:   att.FileType,
			UploadedBy: att.UploadedBy,
			UploadedAt: att.UploadedAt.UTC().Format(time.RFC3339),
			CreatedAt:  att.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:  att.UpdatedAt.UTC().Format(time.RFC3339),
		}
	}

	return responses, nil
}

// GetAttachmentsByEntity retrieves all attachments for a specific entity
func (s *AttachmentService) GetAttachmentsByEntity(entityType, entityID string) ([]models.AttachmentResponse, error) {
	attachments, err := s.attachmentRepo.GetByEntity(entityType, entityID)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to retrieve attachments")
	}

	// Build response
	responses := make([]models.AttachmentResponse, len(attachments))
	for i, att := range attachments {
		responses[i] = models.AttachmentResponse{
			ID:         att.ID,
			EntityType: att.EntityType,
			EntityID:   att.EntityID,
			FilePath:   att.FilePath,
			FileType:   att.FileType,
			UploadedBy: att.UploadedBy,
			UploadedAt: att.UploadedAt.UTC().Format(time.RFC3339),
			CreatedAt:  att.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:  att.UpdatedAt.UTC().Format(time.RFC3339),
		}
	}

	return responses, nil
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
		return errors.NewInternalServerError("failed to delete file from S3")
	}

	// Delete attachment record
	if err := s.attachmentRepo.Delete(id); err != nil {
		return errors.NewInternalServerError("failed to delete attachment record")
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
		return nil, "", errors.NewInternalServerError("failed to download file")
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
		return "", errors.NewInternalServerError("failed to generate download URL")
	}

	return url, nil
}

// GetAttachmentInfo gets information about an attachment
func (s *AttachmentService) GetAttachmentInfo(ctx context.Context, id string) (*models.AttachmentInfoResponse, error) {
	// Get attachment
	attachment, err := s.attachmentRepo.GetByID(id)
	if err != nil {
		return nil, errors.NewNotFoundError("attachment not found")
	}

	// Get file info from S3
	fileInfo, err := s.s3Service.GetFileInfo(ctx, attachment.FilePath)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to get file info")
	}

	return &models.AttachmentInfoResponse{
		ID:         attachment.ID,
		EntityType: attachment.EntityType,
		EntityID:   attachment.EntityID,
		FilePath:   attachment.FilePath,
		FileType:   attachment.FileType,
		UploadedBy: attachment.UploadedBy,
		UploadedAt: attachment.UploadedAt.UTC().Format(time.RFC3339),
		FileSize:   fileInfo.Size,
		CreatedAt:  attachment.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  attachment.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
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
