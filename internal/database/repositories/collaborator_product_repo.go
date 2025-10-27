package repositories

import (
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"

	"gorm.io/gorm"
)

type CollaboratorProductRepository struct {
	db *gorm.DB
}

func NewCollaboratorProductRepository(db *gorm.DB) *CollaboratorProductRepository {
	return &CollaboratorProductRepository{db: db}
}

// Create creates a new collaborator product association
func (r *CollaboratorProductRepository) Create(collabProduct *models.CollaboratorProduct) error {
	if err := r.db.Create(collabProduct).Error; err != nil {
		return errors.NewInternalServerError("Failed to create collaborator product")
	}
	return nil
}

// GetByID retrieves a collaborator product by ID
func (r *CollaboratorProductRepository) GetByID(id string) (*models.CollaboratorProduct, error) {
	var collabProduct models.CollaboratorProduct
	if err := r.db.Preload("Collaborator").Preload("Product").Where("id = ?", id).First(&collabProduct).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("CollaboratorProduct")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve collaborator product")
	}
	return &collabProduct, nil
}

// GetByCollaboratorAndProduct retrieves a specific collaborator-product association
func (r *CollaboratorProductRepository) GetByCollaboratorAndProduct(collaboratorID, productID string) (*models.CollaboratorProduct, error) {
	var collabProduct models.CollaboratorProduct
	if err := r.db.Where("collaborator_id = ? AND product_id = ?", collaboratorID, productID).First(&collabProduct).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("CollaboratorProduct")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve collaborator product")
	}
	return &collabProduct, nil
}

// GetProductsByCollaborator retrieves all products for a collaborator
func (r *CollaboratorProductRepository) GetProductsByCollaborator(collaboratorID string) ([]models.CollaboratorProduct, error) {
	var collabProducts []models.CollaboratorProduct
	if err := r.db.Preload("Product").Where("collaborator_id = ? AND is_active = ?", collaboratorID, true).Find(&collabProducts).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve products by collaborator")
	}
	return collabProducts, nil
}

// GetCollaboratorsByProduct retrieves all collaborators for a product
func (r *CollaboratorProductRepository) GetCollaboratorsByProduct(productID string) ([]models.CollaboratorProduct, error) {
	var collabProducts []models.CollaboratorProduct
	if err := r.db.Preload("Collaborator").Where("product_id = ? AND is_active = ?", productID, true).Find(&collabProducts).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve collaborators by product")
	}
	return collabProducts, nil
}

// Update updates an existing collaborator product
func (r *CollaboratorProductRepository) Update(collabProduct *models.CollaboratorProduct) error {
	if err := r.db.Save(collabProduct).Error; err != nil {
		return errors.NewInternalServerError("Failed to update collaborator product")
	}
	return nil
}

// Delete deletes a collaborator product (soft delete by setting is_active = false)
func (r *CollaboratorProductRepository) Delete(id string) error {
	if err := r.db.Model(&models.CollaboratorProduct{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return errors.NewInternalServerError("Failed to delete collaborator product")
	}
	return nil
}

// DeleteByCollaboratorAndProduct deletes a specific association
func (r *CollaboratorProductRepository) DeleteByCollaboratorAndProduct(collaboratorID, productID string) error {
	if err := r.db.Model(&models.CollaboratorProduct{}).
		Where("collaborator_id = ? AND product_id = ?", collaboratorID, productID).
		Update("is_active", false).Error; err != nil {
		return errors.NewInternalServerError("Failed to delete collaborator product association")
	}
	return nil
}

// Exists checks if a collaborator-product association exists
func (r *CollaboratorProductRepository) Exists(collaboratorID, productID string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.CollaboratorProduct{}).
		Where("collaborator_id = ? AND product_id = ? AND is_active = ?", collaboratorID, productID, true).
		Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check collaborator product existence")
	}
	return count > 0, nil
}
