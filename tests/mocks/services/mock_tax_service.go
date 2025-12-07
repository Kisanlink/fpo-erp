package services

import (
	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockTaxService is a mock implementation of the TaxServiceInterface
// Simplified for GST-only tax system
type MockTaxService struct {
	mock.Mock
}

// CalculateGST calculates GST for a line item
func (m *MockTaxService) CalculateGST(lineTotal float64, gstRate float64, isInterState bool) *models.GSTCalculation {
	args := m.Called(lineTotal, gstRate, isInterState)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.GSTCalculation)
}

// GetTaxSummaryBySale retrieves the tax summary for a sale
func (m *MockTaxService) GetTaxSummaryBySale(saleID string) (*models.TaxSummary, error) {
	args := m.Called(saleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaxSummary), args.Error(1)
}

// GetTaxSummaryByReturn retrieves the tax summary for a return
func (m *MockTaxService) GetTaxSummaryByReturn(returnID string) (*models.TaxSummary, error) {
	args := m.Called(returnID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaxSummary), args.Error(1)
}
