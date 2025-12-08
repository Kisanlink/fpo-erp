package interfaces

import (
	"context"

	"kisanlink-erp/internal/database/models"
)

type WarehouseServiceInterface interface {
	CreateWarehouse(ctx context.Context, request *models.CreateWarehouseRequest, userID string, jwtToken string) (*models.WarehouseResponse, error)
	GetWarehouse(ctx context.Context, id string, jwtToken string) (*models.WarehouseResponse, error)
	GetAllWarehouses(ctx context.Context, limit, offset int, jwtToken string) ([]models.WarehouseResponse, int64, error)
	UpdateWarehouse(ctx context.Context, id string, request *models.UpdateWarehouseRequest, jwtToken string) (*models.WarehouseResponse, error)
	DeleteWarehouse(ctx context.Context, id string, jwtToken string) error
	SearchWarehouses(ctx context.Context, query string, jwtToken string) ([]models.WarehouseResponse, error)
}
