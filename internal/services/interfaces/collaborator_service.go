package interfaces

import (
	"context"

	"kisanlink-erp/internal/database/models"
)

type CollaboratorServiceInterface interface {
	CreateCollaborator(ctx context.Context, request *models.CreateCollaboratorRequest, organizationID string, userID string, jwtToken string) (*models.CollaboratorResponse, error)
	GetCollaborator(ctx context.Context, id string, jwtToken string) (*models.CollaboratorResponse, error)
	GetAllCollaborators(ctx context.Context, jwtToken string, limit, offset int) ([]models.CollaboratorResponse, int64, error)
	GetActiveCollaborators(ctx context.Context, jwtToken string, limit, offset int) ([]models.CollaboratorResponse, int64, error)
	UpdateCollaborator(ctx context.Context, id string, request *models.UpdateCollaboratorRequest, organizationID string, userID string, jwtToken string) (*models.CollaboratorResponse, error)
	DeleteCollaborator(ctx context.Context, id string, organizationID string, jwtToken string) error
	SearchCollaborators(ctx context.Context, query string, jwtToken string, limit, offset int) ([]models.CollaboratorResponse, int64, error)
	GetCollaboratorStats(ctx context.Context, collaboratorID string) (*models.CollaboratorStats, error)
}
