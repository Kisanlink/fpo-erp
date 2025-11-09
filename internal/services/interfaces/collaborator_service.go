package interfaces

import (
	"context"
	"kisanlink-erp/internal/database/models"
)

type CollaboratorServiceInterface interface {
	CreateCollaborator(ctx context.Context, request *models.CreateCollaboratorRequest, userID string, jwtToken string) (*models.CollaboratorResponse, error)
	GetCollaborator(ctx context.Context, id string, jwtToken string) (*models.CollaboratorResponse, error)
	GetAllCollaborators(ctx context.Context, jwtToken string) ([]models.CollaboratorResponse, error)
	GetActiveCollaborators(ctx context.Context, jwtToken string) ([]models.CollaboratorResponse, error)
	UpdateCollaborator(ctx context.Context, id string, request *models.UpdateCollaboratorRequest, jwtToken string) (*models.CollaboratorResponse, error)
	DeleteCollaborator(ctx context.Context, id string, jwtToken string) error
	SearchCollaborators(ctx context.Context, query string, jwtToken string) ([]models.CollaboratorResponse, error)
}
