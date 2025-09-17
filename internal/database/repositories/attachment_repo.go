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
	err := r.db.Preload("Sale").Preload("Return").First(&attachment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

// GetBySaleID retrieves all attachments for a sale
func (r *AttachmentRepository) GetBySaleID(saleID string) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := r.db.Where("sale_id = ?", saleID).Find(&attachments).Error
	return attachments, err
}

// GetByReturnID retrieves all attachments for a return
func (r *AttachmentRepository) GetByReturnID(returnID string) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := r.db.Where("return_id = ?", returnID).Find(&attachments).Error
	return attachments, err
}

// GetAll retrieves all attachments with optional filters
func (r *AttachmentRepository) GetAll(saleID, returnID *string, limit, offset int) ([]models.Attachment, error) {
	var attachments []models.Attachment
	query := r.db.Preload("Sale").Preload("Return")

	if saleID != nil {
		query = query.Where("sale_id = ?", *saleID)
	}
	if returnID != nil {
		query = query.Where("return_id = ?", *returnID)
	}

	err := query.Limit(limit).Offset(offset).Find(&attachments).Error
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

// DeleteBySaleID deletes all attachments for a sale
func (r *AttachmentRepository) DeleteBySaleID(saleID string) error {
	return r.db.Where("sale_id = ?", saleID).Delete(&models.Attachment{}).Error
}

// DeleteByReturnID deletes all attachments for a return
func (r *AttachmentRepository) DeleteByReturnID(returnID string) error {
	return r.db.Where("return_id = ?", returnID).Delete(&models.Attachment{}).Error
}

// CountBySaleID counts attachments for a sale
func (r *AttachmentRepository) CountBySaleID(saleID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Attachment{}).Where("sale_id = ?", saleID).Count(&count).Error
	return count, err
}

// CountByReturnID counts attachments for a return
func (r *AttachmentRepository) CountByReturnID(returnID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Attachment{}).Where("return_id = ?", returnID).Count(&count).Error
	return count, err
}

// GetByFileType retrieves attachments by file type
func (r *AttachmentRepository) GetByFileType(fileType string, limit, offset int) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := r.db.Where("file_type LIKE ?", "%"+fileType+"%").
		Preload("Sale").Preload("Return").
		Limit(limit).Offset(offset).
		Find(&attachments).Error
	return attachments, err
}

// GetByUploadedBy retrieves attachments uploaded by a specific user
func (r *AttachmentRepository) GetByUploadedBy(uploadedBy string, limit, offset int) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := r.db.Where("uploaded_by = ?", uploadedBy).
		Preload("Sale").Preload("Return").
		Limit(limit).Offset(offset).
		Find(&attachments).Error
	return attachments, err
}

// GetRecentAttachments retrieves recent attachments
func (r *AttachmentRepository) GetRecentAttachments(limit int) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := r.db.Order("uploaded_at DESC").
		Preload("Sale").Preload("Return").
		Limit(limit).
		Find(&attachments).Error
	return attachments, err
}

// GetAttachmentsByDateRange retrieves attachments within a date range
func (r *AttachmentRepository) GetAttachmentsByDateRange(startDate, endDate string, limit, offset int) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := r.db.Where("uploaded_at BETWEEN ? AND ?", startDate, endDate).
		Preload("Sale").Preload("Return").
		Order("uploaded_at DESC").
		Limit(limit).Offset(offset).
		Find(&attachments).Error
	return attachments, err
}
