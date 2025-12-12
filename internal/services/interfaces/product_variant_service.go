package interfaces

import (
	"context"

	"kisanlink-erp/internal/database/models"
)

type ProductVariantServiceInterface interface {
	CreateProductVariant(ctx context.Context, productID string, request *models.CreateProductVariantRequest) (*models.ProductVariantResponse, error)
	GetProductVariant(ctx context.Context, id string) (*models.ProductVariantResponse, error)
	GetVariantsByProduct(ctx context.Context, productID string) ([]models.ProductVariantResponse, error)
	GetVariantsByProductPaginated(ctx context.Context, productID string, limit, offset int) ([]models.ProductVariantResponse, int64, error)
	GetVariantBySKU(ctx context.Context, sku string) (*models.ProductVariantResponse, error)
	GetVariantByBarcode(ctx context.Context, barcode string) (*models.ProductVariantResponse, error)
	GetVariantsByCollaborator(ctx context.Context, collaboratorID string) ([]models.ProductVariantResponse, error)
	GetVariantsByCollaboratorPaginated(ctx context.Context, collaboratorID string, limit, offset int) ([]models.ProductVariantResponse, int64, error)
	UpdateProductVariant(ctx context.Context, id string, request *models.UpdateProductVariantRequest) (*models.ProductVariantResponse, error)
	DeleteProductVariant(ctx context.Context, id string) error
}
