package services

import (
	"context"

	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockCollaboratorProductService is a mock implementation of CollaboratorProductServiceInterface
type MockCollaboratorProductService struct {
	mock.Mock
}

func (m *MockCollaboratorProductService) AddProductToCollaborator(ctx context.Context, collaboratorID string, request *models.CreateCollaboratorProductRequest) (*models.CollaboratorProductResponse, error) {
	args := m.Called(ctx, collaboratorID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CollaboratorProductResponse), args.Error(1)
}

func (m *MockCollaboratorProductService) GetProductsByCollaborator(ctx context.Context, collaboratorID string) ([]models.CollaboratorProductResponse, error) {
	args := m.Called(ctx, collaboratorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.CollaboratorProductResponse), args.Error(1)
}

func (m *MockCollaboratorProductService) GetCollaboratorsByProduct(ctx context.Context, productID string) ([]models.CollaboratorProductResponse, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.CollaboratorProductResponse), args.Error(1)
}

func (m *MockCollaboratorProductService) GetCollaboratorProduct(ctx context.Context, id string) (*models.CollaboratorProductResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CollaboratorProductResponse), args.Error(1)
}

func (m *MockCollaboratorProductService) UpdateCollaboratorProduct(ctx context.Context, id string, request *models.UpdateCollaboratorProductRequest) (*models.CollaboratorProductResponse, error) {
	args := m.Called(ctx, id, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CollaboratorProductResponse), args.Error(1)
}

func (m *MockCollaboratorProductService) RemoveProductFromCollaborator(ctx context.Context, collaboratorID, productID string) error {
	args := m.Called(ctx, collaboratorID, productID)
	return args.Error(0)
}

func (m *MockCollaboratorProductService) DeleteCollaboratorProduct(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
