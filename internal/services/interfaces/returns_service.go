package interfaces

import (
	"time"

	"kisanlink-erp/internal/database/models"
)

// ReturnsServiceInterface defines the contract for returns service operations
type ReturnsServiceInterface interface {
	CreateReturn(req *models.CreateReturnRequest) (*models.ReturnResponse, error)
	GetReturn(id string) (*models.ReturnResponse, error)
	GetAllReturns(limit, offset int) ([]models.ReturnResponse, error)
	UpdateReturn(id string, req *models.UpdateReturnRequest) (*models.ReturnResponse, error)
	DeleteReturn(id string) error
	GetReturnsByCustomer(customerID string) ([]models.ReturnResponse, error)
	GetReturnsBySaleID(saleID string) ([]models.ReturnResponse, error)
	GetReturnsByDateRange(startDate, endDate time.Time) ([]models.ReturnResponse, error)
	GetReturnsByStatus(status string) ([]models.ReturnResponse, error)
	GetTotalReturnsAmount(startDate, endDate time.Time) (float64, error)
	GetReturnRateByProduct(productID string, startDate, endDate time.Time) (float64, error)
	GetMostReturnedProducts(limit int) ([]models.MostReturnedProductResponse, error)
}
