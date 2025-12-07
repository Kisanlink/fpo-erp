package repositories

import (
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"

	"gorm.io/gorm"
)

// TaxRepository handles TaxSummary database operations
// Simplified for GST-only tax system - tax rates are on ProductVariant
type TaxRepository struct {
	db *gorm.DB
}

func NewTaxRepository(db *gorm.DB) *TaxRepository {
	return &TaxRepository{db: db}
}

// CreateTaxSummary creates a tax summary record
func (r *TaxRepository) CreateTaxSummary(taxSummary *models.TaxSummary) error {
	if err := r.db.Create(taxSummary).Error; err != nil {
		return errors.NewInternalServerError("Failed to create tax summary")
	}
	return nil
}

// CreateTaxSummaryWithTx creates a tax summary record within a transaction
func (r *TaxRepository) CreateTaxSummaryWithTx(tx *gorm.DB, taxSummary *models.TaxSummary) error {
	if err := tx.Create(taxSummary).Error; err != nil {
		return errors.NewInternalServerError("Failed to create tax summary")
	}
	return nil
}

// GetTaxSummaryBySale retrieves tax summary for a sale
func (r *TaxRepository) GetTaxSummaryBySale(saleID string) (*models.TaxSummary, error) {
	var taxSummary models.TaxSummary
	if err := r.db.Where("sale_id = ?", saleID).First(&taxSummary).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Tax summary not found")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve tax summary")
	}
	return &taxSummary, nil
}

// GetTaxSummaryBySaleWithTx retrieves tax summary for a sale within a transaction
func (r *TaxRepository) GetTaxSummaryBySaleWithTx(tx *gorm.DB, saleID string) (*models.TaxSummary, error) {
	var taxSummary models.TaxSummary
	if err := tx.Where("sale_id = ?", saleID).First(&taxSummary).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found is OK for voiding
		}
		return nil, errors.NewInternalServerError("Failed to retrieve tax summary")
	}
	return &taxSummary, nil
}

// GetTaxSummaryByReturn retrieves tax summary for a return
func (r *TaxRepository) GetTaxSummaryByReturn(returnID string) (*models.TaxSummary, error) {
	var taxSummary models.TaxSummary
	if err := r.db.Where("return_id = ?", returnID).First(&taxSummary).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Tax summary not found")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve tax summary")
	}
	return &taxSummary, nil
}

// UpdateTaxSummary updates a tax summary
func (r *TaxRepository) UpdateTaxSummary(id string, updates map[string]interface{}) error {
	if err := r.db.Model(&models.TaxSummary{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return errors.NewInternalServerError("Failed to update tax summary")
	}
	return nil
}

// DeleteTaxSummaryBySaleWithTx deletes tax summary for a sale within a transaction
func (r *TaxRepository) DeleteTaxSummaryBySaleWithTx(tx *gorm.DB, saleID string) error {
	if err := tx.Where("sale_id = ?", saleID).Delete(&models.TaxSummary{}).Error; err != nil {
		return errors.NewInternalServerError("Failed to delete tax summary")
	}
	return nil
}
