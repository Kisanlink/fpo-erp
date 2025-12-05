package interfaces

import (
	"context"

	"kisanlink-erp/internal/database/models"
)

// ProductServiceInterface defines the contract for product service operations
type ProductServiceInterface interface {
	CreateProduct(request *models.CreateProductRequest) (*models.ProductResponse, error)
	GetProduct(ctx context.Context, id string) (*models.ProductResponse, error)
	GetAllProducts(ctx context.Context) ([]models.ProductResponse, error)
	UpdateProduct(id string, request *models.UpdateProductRequest) (*models.ProductResponse, error)
	DeleteProduct(id string) error
	SearchProducts(query string) ([]models.ProductResponse, error)
	GetProductWithPrices(id string) (*models.ProductWithPricesResponse, error)
}
