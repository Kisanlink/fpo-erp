package services

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"

	"go.uber.org/zap"
)

// AttachmentService handles attachment business logic
type AttachmentService struct {
	attachmentRepo   *repositories.AttachmentRepository
	variantRepo      *repositories.ProductVariantRepository
	collaboratorRepo *repositories.CollaboratorRepository
	poRepo           *repositories.PurchaseOrderRepository
	grnRepo          *repositories.GRNRepository
	s3Service        *S3Service
	logger           interfaces.Logger
}

// NewAttachmentService creates a new attachment service
func NewAttachmentService(attachmentRepo *repositories.AttachmentRepository, variantRepo *repositories.ProductVariantRepository, collaboratorRepo *repositories.CollaboratorRepository, poRepo *repositories.PurchaseOrderRepository, grnRepo *repositories.GRNRepository, s3Service *S3Service, logger interfaces.Logger) *AttachmentService {
	return &AttachmentService{
		attachmentRepo:   attachmentRepo,
		variantRepo:      variantRepo,
		collaboratorRepo: collaboratorRepo,
		poRepo:           poRepo,
		grnRepo:          grnRepo,
		s3Service:        s3Service,
		logger:           logger,
	}
}

// UploadAttachment uploads a file and creates an attachment record
func (s *AttachmentService) UploadAttachment(ctx context.Context, file *multipart.FileHeader, entityType, entityID, uploadedBy string) (*models.AttachmentResponse, error) {
	s.logger.Info("Uploading attachment",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
		zap.String("filename", file.Filename),
		zap.Int64("file_size", file.Size),
		zap.String("uploaded_by", uploadedBy))

	// Validate file
	if err := s.validateFile(file); err != nil {
		s.logger.Warn("File validation failed",
			zap.Error(err),
			zap.String("filename", file.Filename))
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
		s.logger.Error("Failed to upload file to S3",
			zap.Error(err),
			zap.String("filename", file.Filename))
		return nil, errors.NewInternalServerError("failed to upload file")
	}

	s.logger.Debug("File uploaded to S3",
		zap.String("s3_key", s3Key))

	// Create attachment record using the proper constructor
	var uploadedByPtr *string
	if uploadedBy != "" {
		uploadedByPtr = &uploadedBy
	}

	attachment := models.NewAttachment(entityType, entityID, s3Key, s.s3Service.GetContentType(file), uploadedByPtr, time.Now())

	if err := s.attachmentRepo.Create(attachment); err != nil {
		s.logger.Error("Failed to create attachment record",
			zap.Error(err),
			zap.String("s3_key", s3Key))
		// If database creation fails, delete the uploaded file
		if deleteErr := s.s3Service.DeleteFile(ctx, s3Key); deleteErr != nil {
			s.logger.Error("Failed to delete S3 file after database error",
				zap.Error(deleteErr),
				zap.String("s3_key", s3Key))
			// Log the deletion error but return the original error
			fmt.Printf("Failed to delete S3 file after database error: %v", deleteErr)
		}
		return nil, errors.NewInternalServerError("failed to create attachment record")
	}

	s.logger.Info("Attachment created successfully",
		zap.String("attachment_id", attachment.ID),
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID))

	// Auto-update variant images if entity_type is "variant"
	if entityType == "variant" && s.variantRepo != nil {
		if err := s.addImageToVariant(entityID, s3Key); err != nil {
			s.logger.Warn("Failed to auto-update variant images",
				zap.Error(err),
				zap.String("variant_id", entityID),
				zap.String("s3_key", s3Key))
			// Don't fail the upload, just log the warning
		} else {
			s.logger.Info("Auto-updated variant images",
				zap.String("variant_id", entityID),
				zap.String("s3_key", s3Key))
		}
	}

	// Auto-update collaborator logo if entity_type is "logo"
	if entityType == "logo" && s.collaboratorRepo != nil {
		// Generate a long-lived presigned URL (1 year) for the logo
		if logoURL, err := s.s3Service.GeneratePresignedURLForKey(ctx, s3Key, 365*24*time.Hour); err == nil {
			if err := s.updateCollaboratorLogo(entityID, logoURL); err != nil {
				s.logger.Warn("Failed to auto-update collaborator logo",
					zap.Error(err),
					zap.String("collaborator_id", entityID),
					zap.String("s3_key", s3Key))
				// Don't fail the upload, just log the warning
			} else {
				s.logger.Info("Auto-updated collaborator logo",
					zap.String("collaborator_id", entityID),
					zap.String("logo_url", logoURL))
			}
		}
	}

	// Auto-update PO documents if entity_type is "po"
	if entityType == "po" && s.poRepo != nil {
		if err := s.addDocumentToPO(entityID, s3Key); err != nil {
			s.logger.Warn("Failed to add document to PO",
				zap.Error(err),
				zap.String("po_id", entityID),
				zap.String("s3_key", s3Key))
			// Don't fail the upload, just log the warning
		} else {
			s.logger.Info("Auto-added document to PO",
				zap.String("po_id", entityID),
				zap.String("s3_key", s3Key))
		}
	}

	// Auto-update GRN documents if entity_type is "grn"
	if entityType == "grn" && s.grnRepo != nil {
		if err := s.addDocumentToGRN(entityID, s3Key); err != nil {
			s.logger.Warn("Failed to add document to GRN",
				zap.Error(err),
				zap.String("grn_id", entityID),
				zap.String("s3_key", s3Key))
			// Don't fail the upload, just log the warning
		} else {
			s.logger.Info("Auto-added document to GRN",
				zap.String("grn_id", entityID),
				zap.String("s3_key", s3Key))
		}
	}

	// Build response with presigned URL
	response := s.buildAttachmentResponse(ctx, attachment)
	return &response, nil
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
func (s *AttachmentService) GetAttachments(ctx context.Context, entityType, entityID *string, limit, offset int) ([]models.AttachmentResponse, error) {
	attachments, err := s.attachmentRepo.GetAll(entityType, entityID, limit, offset)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to retrieve attachments")
	}

	// Build response with presigned URLs
	responses := make([]models.AttachmentResponse, len(attachments))
	for i, att := range attachments {
		responses[i] = s.buildAttachmentResponse(ctx, &att)
	}

	return responses, nil
}

// GetAttachmentsByEntity retrieves all attachments for a specific entity
func (s *AttachmentService) GetAttachmentsByEntity(ctx context.Context, entityType, entityID string) ([]models.AttachmentResponse, error) {
	attachments, err := s.attachmentRepo.GetByEntity(entityType, entityID)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to retrieve attachments")
	}

	// Build response with presigned URLs
	responses := make([]models.AttachmentResponse, len(attachments))
	for i, att := range attachments {
		responses[i] = s.buildAttachmentResponse(ctx, &att)
	}

	return responses, nil
}

// DeleteAttachment deletes an attachment and its associated file
func (s *AttachmentService) DeleteAttachment(ctx context.Context, id string) error {
	s.logger.Info("Deleting attachment",
		zap.String("attachment_id", id))

	// Get attachment to get the S3 key
	attachment, err := s.attachmentRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve attachment for deletion",
			zap.Error(err),
			zap.String("attachment_id", id))
		return errors.NewNotFoundError("attachment not found")
	}

	// Delete file from S3 using key directly
	if err := s.s3Service.DeleteFileByKey(ctx, attachment.FilePath); err != nil {
		s.logger.Error("Failed to delete file from S3",
			zap.Error(err),
			zap.String("s3_key", attachment.FilePath))
		return errors.NewInternalServerError("failed to delete file from S3")
	}

	s.logger.Debug("S3 file deleted",
		zap.String("s3_key", attachment.FilePath))

	// Remove image from variant if entity_type is "variant"
	if attachment.EntityType == "variant" && s.variantRepo != nil {
		if err := s.removeImageFromVariant(attachment.EntityID, attachment.FilePath); err != nil {
			s.logger.Warn("Failed to remove image from variant",
				zap.Error(err),
				zap.String("variant_id", attachment.EntityID),
				zap.String("s3_key", attachment.FilePath))
			// Don't fail the deletion, just log the warning
		} else {
			s.logger.Info("Removed image from variant",
				zap.String("variant_id", attachment.EntityID),
				zap.String("s3_key", attachment.FilePath))
		}
	}

	// Remove logo from collaborator if entity_type is "logo"
	if attachment.EntityType == "logo" && s.collaboratorRepo != nil {
		if err := s.clearCollaboratorLogo(attachment.EntityID); err != nil {
			s.logger.Warn("Failed to clear collaborator logo",
				zap.Error(err),
				zap.String("collaborator_id", attachment.EntityID))
			// Don't fail the deletion, just log the warning
		} else {
			s.logger.Info("Cleared collaborator logo",
				zap.String("collaborator_id", attachment.EntityID))
		}
	}

	// Remove document from PO if entity_type is "po"
	if attachment.EntityType == "po" && s.poRepo != nil {
		if err := s.removeDocumentFromPO(attachment.EntityID, attachment.FilePath); err != nil {
			s.logger.Warn("Failed to remove document from PO",
				zap.Error(err),
				zap.String("po_id", attachment.EntityID),
				zap.String("s3_key", attachment.FilePath))
			// Don't fail the deletion, just log the warning
		} else {
			s.logger.Info("Removed document from PO",
				zap.String("po_id", attachment.EntityID),
				zap.String("s3_key", attachment.FilePath))
		}
	}

	// Remove document from GRN if entity_type is "grn"
	if attachment.EntityType == "grn" && s.grnRepo != nil {
		if err := s.removeDocumentFromGRN(attachment.EntityID, attachment.FilePath); err != nil {
			s.logger.Warn("Failed to remove document from GRN",
				zap.Error(err),
				zap.String("grn_id", attachment.EntityID),
				zap.String("s3_key", attachment.FilePath))
			// Don't fail the deletion, just log the warning
		} else {
			s.logger.Info("Removed document from GRN",
				zap.String("grn_id", attachment.EntityID),
				zap.String("s3_key", attachment.FilePath))
		}
	}

	// Delete attachment record
	if err := s.attachmentRepo.Delete(id); err != nil {
		s.logger.Error("Failed to delete attachment record",
			zap.Error(err),
			zap.String("attachment_id", id))
		return errors.NewInternalServerError("failed to delete attachment record")
	}

	s.logger.Info("Attachment deleted successfully",
		zap.String("attachment_id", id))

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

	// Generate presigned URL using S3 key directly
	url, err := s.s3Service.GeneratePresignedURLForKey(ctx, attachment.FilePath, expiration)
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

	// Get file info from S3 using key directly
	fileInfo, err := s.s3Service.GetFileInfoByKey(ctx, attachment.FilePath)
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

// buildAttachmentResponse builds an AttachmentResponse with presigned URL
func (s *AttachmentService) buildAttachmentResponse(ctx context.Context, att *models.Attachment) models.AttachmentResponse {
	response := models.AttachmentResponse{
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

	// Generate presigned URL (1 hour expiration)
	if url, err := s.s3Service.GeneratePresignedURLForKey(ctx, att.FilePath, time.Hour); err == nil {
		response.DownloadURL = url
	}

	return response
}

// addImageToVariant adds an S3 key to the variant's images JSON array
func (s *AttachmentService) addImageToVariant(variantID, s3Key string) error {
	// Get the variant
	variant, err := s.variantRepo.GetByID(variantID)
	if err != nil {
		return fmt.Errorf("variant not found: %w", err)
	}

	// Parse existing images
	var images []string
	if variant.Images != nil && *variant.Images != "" {
		if err := json.Unmarshal([]byte(*variant.Images), &images); err != nil {
			// If parsing fails, start fresh
			images = []string{}
		}
	}

	// Check if image already exists (avoid duplicates)
	for _, img := range images {
		if img == s3Key {
			return nil // Already exists, no update needed
		}
	}

	// Add new image
	images = append(images, s3Key)

	// Marshal back to JSON
	imagesJSON, err := json.Marshal(images)
	if err != nil {
		return fmt.Errorf("failed to marshal images: %w", err)
	}

	// Update variant
	imagesStr := string(imagesJSON)
	variant.Images = &imagesStr

	if err := s.variantRepo.Update(variant); err != nil {
		return fmt.Errorf("failed to update variant: %w", err)
	}

	return nil
}

// removeImageFromVariant removes an S3 key from the variant's images JSON array
func (s *AttachmentService) removeImageFromVariant(variantID, s3Key string) error {
	// Get the variant
	variant, err := s.variantRepo.GetByID(variantID)
	if err != nil {
		return fmt.Errorf("variant not found: %w", err)
	}

	// Parse existing images
	var images []string
	if variant.Images != nil && *variant.Images != "" {
		if err := json.Unmarshal([]byte(*variant.Images), &images); err != nil {
			// If parsing fails, nothing to remove
			return nil
		}
	}

	// Find and remove the image
	found := false
	newImages := make([]string, 0, len(images))
	for _, img := range images {
		if img == s3Key {
			found = true
			continue
		}
		newImages = append(newImages, img)
	}

	if !found {
		return nil // Image wasn't in the array, nothing to do
	}

	// Marshal back to JSON
	imagesJSON, err := json.Marshal(newImages)
	if err != nil {
		return fmt.Errorf("failed to marshal images: %w", err)
	}

	// Update variant
	imagesStr := string(imagesJSON)
	variant.Images = &imagesStr

	if err := s.variantRepo.Update(variant); err != nil {
		return fmt.Errorf("failed to update variant: %w", err)
	}

	return nil
}

// updateCollaboratorLogo updates the collaborator's logo field with a presigned URL
func (s *AttachmentService) updateCollaboratorLogo(collaboratorID, logoURL string) error {
	collaborator, err := s.collaboratorRepo.GetByID(collaboratorID)
	if err != nil {
		return fmt.Errorf("collaborator not found: %w", err)
	}
	collaborator.Logo = &logoURL
	if err := s.collaboratorRepo.Update(collaborator); err != nil {
		return fmt.Errorf("failed to update collaborator: %w", err)
	}
	return nil
}

// clearCollaboratorLogo removes the logo from a collaborator
func (s *AttachmentService) clearCollaboratorLogo(collaboratorID string) error {
	collaborator, err := s.collaboratorRepo.GetByID(collaboratorID)
	if err != nil {
		return fmt.Errorf("collaborator not found: %w", err)
	}
	collaborator.Logo = nil
	if err := s.collaboratorRepo.Update(collaborator); err != nil {
		return fmt.Errorf("failed to update collaborator: %w", err)
	}
	return nil
}

// addDocumentToPO adds an S3 key to the purchase order's documents JSON array
func (s *AttachmentService) addDocumentToPO(poID, s3Key string) error {
	po, err := s.poRepo.GetByID(poID)
	if err != nil {
		return fmt.Errorf("purchase order not found: %w", err)
	}

	// Parse existing documents
	var docs []string
	if po.Documents != nil && *po.Documents != "" {
		if err := json.Unmarshal([]byte(*po.Documents), &docs); err != nil {
			docs = []string{}
		}
	}

	// Check if document already exists (avoid duplicates)
	for _, doc := range docs {
		if doc == s3Key {
			return nil
		}
	}

	// Add new document
	docs = append(docs, s3Key)

	// Marshal back to JSON
	docsJSON, err := json.Marshal(docs)
	if err != nil {
		return fmt.Errorf("failed to marshal documents: %w", err)
	}

	docsStr := string(docsJSON)
	po.Documents = &docsStr

	if err := s.poRepo.Update(po); err != nil {
		return fmt.Errorf("failed to update purchase order: %w", err)
	}

	return nil
}

// removeDocumentFromPO removes an S3 key from the purchase order's documents JSON array
func (s *AttachmentService) removeDocumentFromPO(poID, s3Key string) error {
	po, err := s.poRepo.GetByID(poID)
	if err != nil {
		return fmt.Errorf("purchase order not found: %w", err)
	}

	// Parse existing documents
	var docs []string
	if po.Documents != nil && *po.Documents != "" {
		if err := json.Unmarshal([]byte(*po.Documents), &docs); err != nil {
			return nil
		}
	}

	// Find and remove the document
	found := false
	newDocs := make([]string, 0, len(docs))
	for _, doc := range docs {
		if doc == s3Key {
			found = true
			continue
		}
		newDocs = append(newDocs, doc)
	}

	if !found {
		return nil
	}

	// Marshal back to JSON
	docsJSON, err := json.Marshal(newDocs)
	if err != nil {
		return fmt.Errorf("failed to marshal documents: %w", err)
	}

	docsStr := string(docsJSON)
	po.Documents = &docsStr

	if err := s.poRepo.Update(po); err != nil {
		return fmt.Errorf("failed to update purchase order: %w", err)
	}

	return nil
}

// addDocumentToGRN adds an S3 key to the GRN's documents JSON array
func (s *AttachmentService) addDocumentToGRN(grnID, s3Key string) error {
	grn, err := s.grnRepo.GetByID(grnID)
	if err != nil {
		return fmt.Errorf("GRN not found: %w", err)
	}

	// Parse existing documents
	var docs []string
	if grn.Documents != nil && *grn.Documents != "" {
		if err := json.Unmarshal([]byte(*grn.Documents), &docs); err != nil {
			docs = []string{}
		}
	}

	// Check if document already exists (avoid duplicates)
	for _, doc := range docs {
		if doc == s3Key {
			return nil
		}
	}

	// Add new document
	docs = append(docs, s3Key)

	// Marshal back to JSON
	docsJSON, err := json.Marshal(docs)
	if err != nil {
		return fmt.Errorf("failed to marshal documents: %w", err)
	}

	docsStr := string(docsJSON)

	updates := map[string]interface{}{
		"documents": docsStr,
	}
	if err := s.grnRepo.Update(grnID, updates); err != nil {
		return fmt.Errorf("failed to update GRN: %w", err)
	}

	return nil
}

// removeDocumentFromGRN removes an S3 key from the GRN's documents JSON array
func (s *AttachmentService) removeDocumentFromGRN(grnID, s3Key string) error {
	grn, err := s.grnRepo.GetByID(grnID)
	if err != nil {
		return fmt.Errorf("GRN not found: %w", err)
	}

	// Parse existing documents
	var docs []string
	if grn.Documents != nil && *grn.Documents != "" {
		if err := json.Unmarshal([]byte(*grn.Documents), &docs); err != nil {
			return nil
		}
	}

	// Find and remove the document
	found := false
	newDocs := make([]string, 0, len(docs))
	for _, doc := range docs {
		if doc == s3Key {
			found = true
			continue
		}
		newDocs = append(newDocs, doc)
	}

	if !found {
		return nil
	}

	// Marshal back to JSON
	docsJSON, err := json.Marshal(newDocs)
	if err != nil {
		return fmt.Errorf("failed to marshal documents: %w", err)
	}

	docsStr := string(docsJSON)

	updates := map[string]interface{}{
		"documents": docsStr,
	}
	if err := s.grnRepo.Update(grnID, updates); err != nil {
		return fmt.Errorf("failed to update GRN: %w", err)
	}

	return nil
}
