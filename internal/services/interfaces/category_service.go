package interfaces

import (
	"context"
	"kisanlink-erp/internal/database/models"
)

type CategoryServiceInterface interface {
	SeedCategories(ctx context.Context) (*models.SeedCategoriesResponse, error)
	CreateCategory(ctx context.Context, request *models.CreateCategoryRequest) (*models.CategoryResponse, error)
	GetCategory(ctx context.Context, id string) (*models.CategoryResponse, error)
	GetCategoryByName(ctx context.Context, name string) (*models.CategoryResponse, error)
	GetAllCategories(ctx context.Context) ([]models.CategoryResponse, error)
	GetAllCategoriesWithSubcategories(ctx context.Context) ([]models.CategoryResponse, error)
	UpdateCategory(ctx context.Context, id string, request *models.UpdateCategoryRequest) (*models.CategoryResponse, error)
	DeleteCategory(ctx context.Context, id string) error
}
