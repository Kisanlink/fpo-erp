package services

import (
	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockTaxService is a mock implementation of TaxServiceInterface
type MockTaxService struct {
	mock.Mock
}

func (m *MockTaxService) CreateTax(req *models.CreateTaxRequest, userID string) (*models.TaxResponse, error) {
	args := m.Called(req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaxResponse), args.Error(1)
}

func (m *MockTaxService) GetTax(id string) (*models.TaxResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaxResponse), args.Error(1)
}

func (m *MockTaxService) GetAllTaxes(limit, offset int) ([]models.TaxResponse, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TaxResponse), args.Error(1)
}

func (m *MockTaxService) GetActiveTaxes() ([]models.TaxResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TaxResponse), args.Error(1)
}

func (m *MockTaxService) GetTaxesByType(taxType models.TaxType) ([]models.TaxResponse, error) {
	args := m.Called(taxType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TaxResponse), args.Error(1)
}

func (m *MockTaxService) GetTaxesByStatus(status string) ([]models.TaxResponse, error) {
	args := m.Called(status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TaxResponse), args.Error(1)
}

func (m *MockTaxService) UpdateTax(id string, req *models.UpdateTaxRequest, userID string) (*models.TaxResponse, error) {
	args := m.Called(id, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaxResponse), args.Error(1)
}

func (m *MockTaxService) DeleteTax(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTaxService) CalculateTax(req *models.TaxCalculationRequest) (*models.TaxCalculationResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaxCalculationResponse), args.Error(1)
}

func (m *MockTaxService) ApplyTaxesToSale(saleID string, items []models.SaleItem, req *models.TaxCalculationRequest, userID string) (*models.TaxSummary, error) {
	args := m.Called(saleID, items, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaxSummary), args.Error(1)
}

func (m *MockTaxService) ApplyTaxesToSaleWithTx(tx *gorm.DB, saleID string, items []models.SaleItem, req *models.TaxCalculationRequest, userID string) (*models.TaxSummary, error) {
	args := m.Called(tx, saleID, items, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaxSummary), args.Error(1)
}

func (m *MockTaxService) ApplyTaxesToReturn(returnID string, items []models.ReturnItem, req *models.TaxCalculationRequest, userID string) (*models.TaxSummary, error) {
	args := m.Called(returnID, items, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaxSummary), args.Error(1)
}

func (m *MockTaxService) GetTaxSummaryBySale(saleID string) (*models.TaxSummary, error) {
	args := m.Called(saleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaxSummary), args.Error(1)
}

func (m *MockTaxService) GetTaxSummaryByReturn(returnID string) (*models.TaxSummary, error) {
	args := m.Called(returnID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaxSummary), args.Error(1)
}

func (m *MockTaxService) GetTaxApplicationsBySale(saleID string) ([]models.TaxApplication, error) {
	args := m.Called(saleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TaxApplication), args.Error(1)
}

func (m *MockTaxService) GetTaxApplicationsByReturn(returnID string) ([]models.TaxApplication, error) {
	args := m.Called(returnID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TaxApplication), args.Error(1)
}

func (m *MockTaxService) CalculateBatchTax(batch models.InventoryBatch, quantity int64, unitPrice float64) (*models.BatchTaxCalculation, error) {
	args := m.Called(batch, quantity, unitPrice)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BatchTaxCalculation), args.Error(1)
}
