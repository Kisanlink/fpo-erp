package interfaces

import (
	"context"
	"kisanlink-erp/internal/database/models"
)

type SubcategoryServiceInterface interface {
	CreateSubcategory(ctx context.Context, request *models.CreateSubcategoryRequest) (*models.SubcategoryResponse, error)
	GetSubcategory(ctx context.Context, id string) (*models.SubcategoryResponse, error)
	GetSubcategoryByName(ctx context.Context, name string) (*models.SubcategoryResponse, error)
	GetSubcategoriesByCategory(ctx context.Context, categoryID string) ([]models.SubcategoryResponse, error)
	GetSubcategoriesByCategoryPaginated(ctx context.Context, categoryID string, limit, offset int) ([]models.SubcategoryResponse, int64, error)
	GetAllSubcategories(ctx context.Context) ([]models.SubcategoryResponse, error)
	GetAllSubcategoriesPaginated(ctx context.Context, limit, offset int) ([]models.SubcategoryResponse, int64, error)
	SearchSubcategories(ctx context.Context, query string, limit, offset int) ([]models.SubcategoryResponse, int64, error)
	UpdateSubcategory(ctx context.Context, id string, request *models.UpdateSubcategoryRequest) (*models.SubcategoryResponse, error)
	DeleteSubcategory(ctx context.Context, id string) error
}
