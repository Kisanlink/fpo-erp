package interfaces

import (
	"kisanlink-erp/internal/database/models"
)

// CollaboratorInterface defines the contract for collaborator repository operations
type CollaboratorInterface interface {
	// Create creates a new collaborator
	Create(collaborator *models.Collaborator) error

	// GetByID retrieves a collaborator by ID
	GetByID(id string) (*models.Collaborator, error)

	// GetAll retrieves all collaborators
	GetAll() ([]models.Collaborator, error)

	// Update updates an existing collaborator
	Update(collaborator *models.Collaborator) error

	// Delete deletes a collaborator (soft delete by setting is_active = false)
	Delete(id string) error

	// GetByGSTNumber retrieves a collaborator by GST number
	GetByGSTNumber(gstNumber string) (*models.Collaborator, error)

	// GetActiveCollaborators retrieves all active collaborators
	GetActiveCollaborators() ([]models.Collaborator, error)

	// SearchByName searches collaborators by company name
	SearchByName(name string) ([]models.Collaborator, error)

	// GSTNumberExists checks if a GST number already exists
	GSTNumberExists(gstNumber string) (bool, error)

	// FindByExternalID finds a collaborator by external_id (for webhook integration)
	FindByExternalID(externalID string) (*models.Collaborator, error)
}
