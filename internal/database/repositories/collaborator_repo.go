package repositories

import (
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"

	"gorm.io/gorm"
)

type CollaboratorRepository struct {
	db *gorm.DB
}

func NewCollaboratorRepository(db *gorm.DB) *CollaboratorRepository {
	return &CollaboratorRepository{db: db}
}

// Create creates a new collaborator
func (r *CollaboratorRepository) Create(collaborator *models.Collaborator) error {
	if err := r.db.Create(collaborator).Error; err != nil {
		return errors.NewInternalServerError("Failed to create collaborator")
	}
	return nil
}

// GetByID retrieves a collaborator by ID
func (r *CollaboratorRepository) GetByID(id string) (*models.Collaborator, error) {
	var collaborator models.Collaborator
	if err := r.db.Where("id = ?", id).First(&collaborator).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Collaborator")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve collaborator")
	}
	return &collaborator, nil
}

// GetAll retrieves all collaborators
func (r *CollaboratorRepository) GetAll() ([]models.Collaborator, error) {
	var collaborators []models.Collaborator
	if err := r.db.Find(&collaborators).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve collaborators")
	}
	return collaborators, nil
}

// Update updates an existing collaborator
func (r *CollaboratorRepository) Update(collaborator *models.Collaborator) error {
	if err := r.db.Save(collaborator).Error; err != nil {
		return errors.NewInternalServerError("Failed to update collaborator")
	}
	return nil
}

// Delete deletes a collaborator (soft delete by setting is_active = false)
func (r *CollaboratorRepository) Delete(id string) error {
	if err := r.db.Model(&models.Collaborator{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return errors.NewInternalServerError("Failed to delete collaborator")
	}
	return nil
}

// GetByGSTNumber retrieves a collaborator by GST number
func (r *CollaboratorRepository) GetByGSTNumber(gstNumber string) (*models.Collaborator, error) {
	var collaborator models.Collaborator
	if err := r.db.Where("gst_number = ?", gstNumber).First(&collaborator).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Collaborator")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve collaborator by GST number")
	}
	return &collaborator, nil
}

// GetActiveCollaborators retrieves all active collaborators
func (r *CollaboratorRepository) GetActiveCollaborators() ([]models.Collaborator, error) {
	var collaborators []models.Collaborator
	if err := r.db.Where("is_active = ?", true).Find(&collaborators).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve active collaborators")
	}
	return collaborators, nil
}

// SearchByName searches collaborators by company name
func (r *CollaboratorRepository) SearchByName(name string) ([]models.Collaborator, error) {
	var collaborators []models.Collaborator
	if err := r.db.Where("company_name ILIKE ?", "%"+name+"%").Find(&collaborators).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to search collaborators by name")
	}
	return collaborators, nil
}

// GSTNumberExists checks if a GST number already exists
func (r *CollaboratorRepository) GSTNumberExists(gstNumber string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Collaborator{}).Where("gst_number = ?", gstNumber).Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check GST number existence")
	}
	return count > 0, nil
}
