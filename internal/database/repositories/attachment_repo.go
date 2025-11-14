package repositories

import (
	"kisanlink-erp/internal/database/models"

	"gorm.io/gorm"
)

// AttachmentRepository handles database operations for attachments
type AttachmentRepository struct {
	db *gorm.DB
}

// NewAttachmentRepository creates a new attachment repository
func NewAttachmentRepository(db *gorm.DB) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

// Create creates a new attachment
func (r *AttachmentRepository) Create(attachment *models.Attachment) error {
	return r.db.Create(attachment).Error
}

// GetByID retrieves an attachment by ID
func (r *AttachmentRepository) GetByID(id string) (*models.Attachment, error) {
	var attachment models.Attachment
	err := r.db.First(&attachment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

// GetByEntity retrieves all attachments for a specific entity
func (r *AttachmentRepository) GetByEntity(entityType, entityID string) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := r.db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("uploaded_at DESC").
		Find(&attachments).Error
	return attachments, err
}

// GetAll retrieves all attachments with optional filters
func (r *AttachmentRepository) GetAll(entityType, entityID *string, limit, offset int) ([]models.Attachment, error) {
	var attachments []models.Attachment
	query := r.db.Model(&models.Attachment{})

	if entityType != nil && *entityType != "" {
		query = query.Where("entity_type = ?", *entityType)
	}
	if entityID != nil && *entityID != "" {
		query = query.Where("entity_id = ?", *entityID)
	}

	err := query.Order("uploaded_at DESC").
		Limit(limit).Offset(offset).
		Find(&attachments).Error
	return attachments, err
}

// Update updates an attachment
func (r *AttachmentRepository) Update(attachment *models.Attachment) error {
	return r.db.Save(attachment).Error
}

// Delete deletes an attachment by ID
func (r *AttachmentRepository) Delete(id string) error {
	return r.db.Delete(&models.Attachment{}, "id = ?", id).Error
}

// DeleteByEntity deletes all attachments for an entity
func (r *AttachmentRepository) DeleteByEntity(entityType, entityID string) error {
	return r.db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).Delete(&models.Attachment{}).Error
}

// CountByEntity counts attachments for a specific entity
func (r *AttachmentRepository) CountByEntity(entityType, entityID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Attachment{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Count(&count).Error
	return count, err
}

// GetByFileType retrieves attachments by file type
func (r *AttachmentRepository) GetByFileType(fileType string, limit, offset int) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := r.db.Where("file_type LIKE ?", "%"+fileType+"%").
		Order("uploaded_at DESC").
		Limit(limit).Offset(offset).
		Find(&attachments).Error
	return attachments, err
}

// GetByUploadedBy retrieves attachments uploaded by a specific user
func (r *AttachmentRepository) GetByUploadedBy(uploadedBy string, limit, offset int) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := r.db.Where("uploaded_by = ?", uploadedBy).
		Order("uploaded_at DESC").
		Limit(limit).Offset(offset).
		Find(&attachments).Error
	return attachments, err
}

// GetRecentAttachments retrieves recent attachments
func (r *AttachmentRepository) GetRecentAttachments(limit int) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := r.db.Order("uploaded_at DESC").
		Limit(limit).
		Find(&attachments).Error
	return attachments, err
}

// GetAttachmentsByDateRange retrieves attachments within a date range
func (r *AttachmentRepository) GetAttachmentsByDateRange(startDate, endDate string, limit, offset int) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := r.db.Where("uploaded_at BETWEEN ? AND ?", startDate, endDate).
		Order("uploaded_at DESC").
		Limit(limit).Offset(offset).
		Find(&attachments).Error
	return attachments, err
}
