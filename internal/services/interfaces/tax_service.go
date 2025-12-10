package interfaces

import (
	"kisanlink-erp/internal/database/models"
)

// TaxServiceInterface defines the simplified GST-only tax service
// Tax rates are now stored on ProductVariant (GSTRate field)
// This service only handles GST calculation (CGST+SGST for intra-state, IGST for inter-state)
type TaxServiceInterface interface {
	// CalculateGST calculates GST for a line item based on variant's GST rate
	// If isInterState is true, returns IGST; otherwise returns CGST+SGST (split 50-50)
	CalculateGST(lineTotal float64, gstRate float64, isInterState bool) *models.GSTCalculation

	// GetTaxSummaryBySale retrieves the tax summary for a sale
	GetTaxSummaryBySale(saleID string) (*models.TaxSummary, error)

	// GetTaxSummaryByReturn retrieves the tax summary for a return
	GetTaxSummaryByReturn(returnID string) (*models.TaxSummary, error)
}
