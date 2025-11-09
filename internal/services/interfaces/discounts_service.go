package interfaces

import (
	"kisanlink-erp/internal/database/models"
)

type DiscountsServiceInterface interface {
	CreateDiscount(req *models.CreateDiscountRequest) (*models.DiscountResponse, error)
	GetDiscount(id string) (*models.DiscountResponse, error)
	GetAllDiscounts(limit, offset int) ([]models.DiscountResponse, error)
	GetActiveDiscounts() ([]models.DiscountResponse, error)
	UpdateDiscount(id string, req *models.UpdateDiscountRequest) (*models.DiscountResponse, error)
	DeleteDiscount(id string) error
	GetDiscountsByType(discountType models.DiscountType) ([]models.DiscountResponse, error)
	GetDiscountsByStatus(status string) ([]models.DiscountResponse, error)
	ValidateDiscount(req *models.ValidateDiscountRequest) (*models.DiscountValidationResponse, error)
	ApplyDiscount(discountID, saleID string, discountAmount float64) error
	GetDiscountUsageBySale(saleID string) ([]models.DiscountUsageResponse, error)
	GetDiscountUsageStats(discountID string) (map[string]interface{}, error)
	GetApplicableDiscountsForOrder(orderValue float64, productIDs []string, categoryIDs []string, warehouseID string) ([]models.DiscountResponse, error)
	CalculateOptimalDiscounts(orderValue float64, productIDs []string, categoryIDs []string, warehouseID string) ([]models.DiscountResponse, float64, error)
	ValidateDiscountWithCategories(req *models.ValidateDiscountRequest, categoryIDs []string) (*models.DiscountValidationResponse, error)
}
