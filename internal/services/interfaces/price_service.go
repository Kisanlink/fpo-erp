package interfaces

import (
	"kisanlink-erp/internal/database/models"
)

type ProductPriceServiceInterface interface {
	CreateProductPrice(request *models.CreateProductPriceRequest) (*models.ProductPriceResponse, error)
	GetProductPrice(id string) (*models.ProductPriceResponse, error)
	GetVariantPrices(variantID string) ([]models.ProductPriceResponse, error)
	GetVariantPricesPaginated(variantID string, limit, offset int) ([]models.ProductPriceResponse, int64, error)
	GetCurrentPrice(variantID, priceType string) (*models.ProductPriceResponse, error)
	UpdateProductPrice(id string, request *models.UpdateProductPriceRequest) (*models.ProductPriceResponse, error)
	DeleteProductPrice(id string) error
	GetExpiredPrices() ([]models.ProductPriceResponse, error)
	GetExpiredPricesPaginated(limit, offset int) ([]models.ProductPriceResponse, int64, error)
}
