package services

import (
	"math"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/interfaces"

	"go.uber.org/zap"
)

// TaxService handles GST tax calculations
// Simplified for GST-only tax system - tax rates are on ProductVariant.GSTRate
type TaxService struct {
	taxRepo *repositories.TaxRepository
	logger  interfaces.Logger
}

func NewTaxService(taxRepo *repositories.TaxRepository, logger interfaces.Logger) *TaxService {
	return &TaxService{
		taxRepo: taxRepo,
		logger:  logger,
	}
}

// CalculateGST calculates GST for a line item based on variant's GST rate
// If isInterState is true, returns IGST; otherwise returns CGST+SGST (split 50-50)
// GST Law: CGST and SGST are always equal (50-50 split of total GST rate)
func (s *TaxService) CalculateGST(lineTotal float64, gstRate float64, isInterState bool) *models.GSTCalculation {
	s.logger.Debug("Calculating GST",
		zap.Float64("line_total", lineTotal),
		zap.Float64("gst_rate", gstRate),
		zap.Bool("is_inter_state", isInterState))

	if gstRate == 0 {
		return &models.GSTCalculation{
			CGSTAmount:     0,
			SGSTAmount:     0,
			IGSTAmount:     0,
			TotalTaxAmount: 0,
			IsInterState:   isInterState,
		}
	}

	if isInterState {
		// Inter-state: Apply full rate as IGST
		igstAmount := s.roundToNearestPaisa(lineTotal * (gstRate / 100))
		return &models.GSTCalculation{
			CGSTAmount:     0,
			SGSTAmount:     0,
			IGSTAmount:     igstAmount,
			TotalTaxAmount: igstAmount,
			IsInterState:   true,
		}
	}

	// Intra-state: Split 50-50 between CGST and SGST
	halfRate := gstRate / 2
	cgstAmount := s.roundToNearestPaisa(lineTotal * (halfRate / 100))
	sgstAmount := s.roundToNearestPaisa(lineTotal * (halfRate / 100))

	return &models.GSTCalculation{
		CGSTAmount:     cgstAmount,
		SGSTAmount:     sgstAmount,
		IGSTAmount:     0,
		TotalTaxAmount: cgstAmount + sgstAmount,
		IsInterState:   false,
	}
}

// GetTaxSummaryBySale retrieves the tax summary for a sale
func (s *TaxService) GetTaxSummaryBySale(saleID string) (*models.TaxSummary, error) {
	s.logger.Debug("Getting tax summary by sale", zap.String("sale_id", saleID))
	return s.taxRepo.GetTaxSummaryBySale(saleID)
}

// GetTaxSummaryByReturn retrieves the tax summary for a return
func (s *TaxService) GetTaxSummaryByReturn(returnID string) (*models.TaxSummary, error) {
	s.logger.Debug("Getting tax summary by return", zap.String("return_id", returnID))
	return s.taxRepo.GetTaxSummaryByReturn(returnID)
}

// roundToNearestPaisa rounds amount to nearest paisa (2 decimal places) for GST compliance
func (s *TaxService) roundToNearestPaisa(amount float64) float64 {
	return math.Round(amount*100) / 100
}
