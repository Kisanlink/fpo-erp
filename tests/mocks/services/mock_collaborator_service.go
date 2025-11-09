package services

import (
	"context"

	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockCollaboratorService is a mock implementation of CollaboratorServiceInterface
type MockCollaboratorService struct {
	mock.Mock
}

func (m *MockCollaboratorService) CreateCollaborator(ctx context.Context, request *models.CreateCollaboratorRequest, userID string, jwtToken string) (*models.CollaboratorResponse, error) {
	args := m.Called(ctx, request, userID, jwtToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CollaboratorResponse), args.Error(1)
}

func (m *MockCollaboratorService) GetCollaborator(ctx context.Context, id string, jwtToken string) (*models.CollaboratorResponse, error) {
	args := m.Called(ctx, id, jwtToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CollaboratorResponse), args.Error(1)
}

func (m *MockCollaboratorService) GetAllCollaborators(ctx context.Context, jwtToken string) ([]models.CollaboratorResponse, error) {
	args := m.Called(ctx, jwtToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.CollaboratorResponse), args.Error(1)
}

func (m *MockCollaboratorService) GetActiveCollaborators(ctx context.Context, jwtToken string) ([]models.CollaboratorResponse, error) {
	args := m.Called(ctx, jwtToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.CollaboratorResponse), args.Error(1)
}

func (m *MockCollaboratorService) UpdateCollaborator(ctx context.Context, id string, request *models.UpdateCollaboratorRequest, jwtToken string) (*models.CollaboratorResponse, error) {
	args := m.Called(ctx, id, request, jwtToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CollaboratorResponse), args.Error(1)
}

func (m *MockCollaboratorService) DeleteCollaborator(ctx context.Context, id string, jwtToken string) error {
	args := m.Called(ctx, id, jwtToken)
	return args.Error(0)
}

func (m *MockCollaboratorService) SearchCollaborators(ctx context.Context, query string, jwtToken string) ([]models.CollaboratorResponse, error) {
	args := m.Called(ctx, query, jwtToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.CollaboratorResponse), args.Error(1)
}
