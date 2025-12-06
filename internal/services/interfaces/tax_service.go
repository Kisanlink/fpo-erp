package interfaces

import (
	"kisanlink-erp/internal/database/models"

	"gorm.io/gorm"
)

type TaxServiceInterface interface {
	CreateTax(req *models.CreateTaxRequest, userID string) (*models.TaxResponse, error)
	GetTax(id string) (*models.TaxResponse, error)
	GetAllTaxes(limit, offset int) ([]models.TaxResponse, int64, error)
	GetActiveTaxes(limit, offset int) ([]models.TaxResponse, int64, error)
	GetTaxesByType(taxType models.TaxType, limit, offset int) ([]models.TaxResponse, int64, error)
	GetTaxesByStatus(status string, limit, offset int) ([]models.TaxResponse, int64, error)
	UpdateTax(id string, req *models.UpdateTaxRequest, userID string) (*models.TaxResponse, error)
	DeleteTax(id string) error
	CalculateTax(req *models.TaxCalculationRequest) (*models.TaxCalculationResponse, error)
	ApplyTaxesToSale(saleID string, items []models.SaleItem, req *models.TaxCalculationRequest, userID string) (*models.TaxSummary, error)
	ApplyTaxesToSaleWithTx(tx *gorm.DB, saleID string, items []models.SaleItem, req *models.TaxCalculationRequest, userID string) (*models.TaxSummary, error)
	ApplyTaxesToReturn(returnID string, items []models.ReturnItem, req *models.TaxCalculationRequest, userID string) (*models.TaxSummary, error)
	GetTaxSummaryBySale(saleID string) (*models.TaxSummary, error)
	GetTaxSummaryByReturn(returnID string) (*models.TaxSummary, error)
	GetTaxApplicationsBySale(saleID string) ([]models.TaxApplication, error)
	GetTaxApplicationsByReturn(returnID string) ([]models.TaxApplication, error)
	CalculateBatchTax(batch models.InventoryBatch, quantity int64, unitPrice float64) (*models.BatchTaxCalculation, error)
}
