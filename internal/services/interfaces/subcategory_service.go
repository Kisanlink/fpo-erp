package interfaces

import (
	"context"
	"kisanlink-erp/internal/database/models"
)

type SubcategoryServiceInterface interface {
	CreateSubcategory(ctx context.Context, request *models.CreateSubcategoryRequest) (*models.SubcategoryResponse, error)
	GetSubcategory(ctx context.Context, id string) (*models.SubcategoryResponse, error)
	GetSubcategoryByName(ctx context.Context, name string) (*models.SubcategoryResponse, error)
	GetSubcategoriesByCategory(ctx context.Context, categoryName string) ([]models.SubcategoryResponse, error)
	GetAllSubcategories(ctx context.Context) ([]models.SubcategoryResponse, error)
	UpdateSubcategory(ctx context.Context, id string, request *models.UpdateSubcategoryRequest) (*models.SubcategoryResponse, error)
	DeleteSubcategory(ctx context.Context, id string) error
}
