package interfaces

import (
	"context"

	"kisanlink-erp/internal/database/models"
)

type CollaboratorProductServiceInterface interface {
	AddProductToCollaborator(ctx context.Context, collaboratorID string, request *models.CreateCollaboratorProductRequest) (*models.CollaboratorProductResponse, error)
	GetProductsByCollaborator(ctx context.Context, collaboratorID string) ([]models.CollaboratorProductResponse, error)
	GetCollaboratorsByProduct(ctx context.Context, productID string) ([]models.CollaboratorProductResponse, error)
	GetCollaboratorProduct(ctx context.Context, id string) (*models.CollaboratorProductResponse, error)
	UpdateCollaboratorProduct(ctx context.Context, id string, request *models.UpdateCollaboratorProductRequest) (*models.CollaboratorProductResponse, error)
	RemoveProductFromCollaborator(ctx context.Context, collaboratorID, productID string) error
	DeleteCollaboratorProduct(ctx context.Context, id string) error
}
