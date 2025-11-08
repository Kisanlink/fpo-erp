package repositories

import (
	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockCollaboratorRepository is a mock implementation of CollaboratorInterface
// Following the e-commerce codebase pattern for manual mocks
type MockCollaboratorRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockCollaboratorRepository) Create(collaborator *models.Collaborator) error {
	args := m.Called(collaborator)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *MockCollaboratorRepository) GetByID(id string) (*models.Collaborator, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Collaborator), args.Error(1)
}

// GetAll mocks the GetAll method
func (m *MockCollaboratorRepository) GetAll() ([]models.Collaborator, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Collaborator), args.Error(1)
}

// Update mocks the Update method
func (m *MockCollaboratorRepository) Update(collaborator *models.Collaborator) error {
	args := m.Called(collaborator)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockCollaboratorRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// GetByGSTNumber mocks the GetByGSTNumber method
func (m *MockCollaboratorRepository) GetByGSTNumber(gstNumber string) (*models.Collaborator, error) {
	args := m.Called(gstNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Collaborator), args.Error(1)
}

// GetActiveCollaborators mocks the GetActiveCollaborators method
func (m *MockCollaboratorRepository) GetActiveCollaborators() ([]models.Collaborator, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Collaborator), args.Error(1)
}

// SearchByName mocks the SearchByName method
func (m *MockCollaboratorRepository) SearchByName(name string) ([]models.Collaborator, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Collaborator), args.Error(1)
}

// GSTNumberExists mocks the GSTNumberExists method
func (m *MockCollaboratorRepository) GSTNumberExists(gstNumber string) (bool, error) {
	args := m.Called(gstNumber)
	return args.Bool(0), args.Error(1)
}

// FindByExternalID mocks the FindByExternalID method
func (m *MockCollaboratorRepository) FindByExternalID(externalID string) (*models.Collaborator, error) {
	args := m.Called(externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Collaborator), args.Error(1)
}
