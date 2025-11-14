package interfaces

import (
	"time"

	"kisanlink-erp/internal/database/models"
)

// SalesServiceInterface defines the contract for sales service operations
type SalesServiceInterface interface {
	CreateSale(req *models.CreateSaleRequest) (*models.SaleResponse, error)
	GetSale(id string) (*models.SaleResponse, error)
	GetAllSales(limit, offset int) ([]models.SaleResponse, error)
	UpdateSale(id string, req *models.UpdateSaleRequest) (*models.SaleResponse, error)
	DeleteSale(id string) error
	GetSalesByDateRange(startDate, endDate time.Time) ([]models.SaleResponse, error)
	GetSalesByStatus(status string) ([]models.SaleResponse, error)
	GetTotalSalesAmount(startDate, endDate time.Time) (float64, error)
	GetTopSellingProducts(limit int) ([]models.TopSellingProductResponse, error)
}
