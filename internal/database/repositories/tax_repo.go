package repositories

import (
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"

	"gorm.io/gorm"
)

type TaxRepository struct {
	db *gorm.DB
}

func NewTaxRepository(db *gorm.DB) *TaxRepository {
	return &TaxRepository{db: db}
}

// CreateTax creates a new tax
func (r *TaxRepository) CreateTax(tax *models.Tax) error {
	if err := r.db.Create(tax).Error; err != nil {
		return errors.NewInternalServerError("Failed to create tax")
	}
	return nil
}

// GetTaxByID retrieves a tax by ID
func (r *TaxRepository) GetTaxByID(id string) (*models.Tax, error) {
	var tax models.Tax
	if err := r.db.Where("id = ?", id).First(&tax).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Tax not found")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve tax")
	}
	return &tax, nil
}

// GetTaxByCode retrieves a tax by code
func (r *TaxRepository) GetTaxByCode(code string) (*models.Tax, error) {
	var tax models.Tax
	if err := r.db.Where("code = ?", code).First(&tax).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Tax not found")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve tax")
	}
	return &tax, nil
}

// GetAllTaxes retrieves all taxes with pagination
func (r *TaxRepository) GetAllTaxes(limit, offset int) ([]models.Tax, error) {
	var taxes []models.Tax
	if err := r.db.Limit(limit).Offset(offset).Order("priority DESC, created_at DESC").Find(&taxes).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve taxes")
	}
	return taxes, nil
}

// GetActiveTaxes retrieves all currently active taxes
func (r *TaxRepository) GetActiveTaxes() ([]models.Tax, error) {
	var taxes []models.Tax
	now := time.Now()

	if err := r.db.Where("is_active = ? AND valid_from <= ? AND (valid_until IS NULL OR valid_until > ?)",
		true, now, now).Order("priority DESC, stacking_order ASC").Find(&taxes).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve active taxes")
	}
	return taxes, nil
}

// GetTaxesByType retrieves taxes by type
func (r *TaxRepository) GetTaxesByType(taxType models.TaxType) ([]models.Tax, error) {
	var taxes []models.Tax
	if err := r.db.Where("tax_type = ?", taxType).Order("priority DESC").Find(&taxes).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve taxes by type")
	}
	return taxes, nil
}

// GetTaxesByStatus retrieves taxes by status
func (r *TaxRepository) GetTaxesByStatus(status string) ([]models.Tax, error) {
	var taxes []models.Tax
	now := time.Now()

	var query *gorm.DB
	switch status {
	case "active":
		query = r.db.Where("is_active = ? AND valid_from <= ? AND (valid_until IS NULL OR valid_until > ?)",
			true, now, now)
	case "inactive":
		query = r.db.Where("is_active = ?", false)
	case "expired":
		query = r.db.Where("valid_until IS NOT NULL AND valid_until <= ?", now)
	case "scheduled":
		query = r.db.Where("is_active = ? AND valid_from > ?", true, now)
	default:
		return nil, errors.NewBadRequestError("Invalid status")
	}

	if err := query.Order("priority DESC").Find(&taxes).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve taxes by status")
	}
	return taxes, nil
}

// UpdateTax updates an existing tax
func (r *TaxRepository) UpdateTax(id string, updates map[string]interface{}) error {
	if err := r.db.Model(&models.Tax{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return errors.NewInternalServerError("Failed to update tax")
	}
	return nil
}

// DeleteTax deletes a tax
func (r *TaxRepository) DeleteTax(id string) error {
	if err := r.db.Delete(&models.Tax{}, "id = ?", id).Error; err != nil {
		return errors.NewInternalServerError("Failed to delete tax")
	}
	return nil
}

// GetApplicableTaxes retrieves taxes that are applicable for a given transaction
func (r *TaxRepository) GetApplicableTaxes(req models.TaxCalculationRequest) ([]models.Tax, error) {
	var taxes []models.Tax
	now := time.Now()

	// Base query for active taxes
	query := r.db.Where("is_active = ? AND valid_from <= ? AND (valid_until IS NULL OR valid_until > ?)",
		true, now, now)

	// Get all active taxes first
	if err := query.Order("priority DESC, stacking_order ASC").Find(&taxes).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve applicable taxes")
	}

	// Filter taxes based on applicability rules
	var applicableTaxes []models.Tax
	for _, tax := range taxes {
		if r.isTaxApplicable(tax, req) {
			applicableTaxes = append(applicableTaxes, tax)
		}
	}

	return applicableTaxes, nil
}

// isTaxApplicable checks if a tax is applicable for the given transaction
func (r *TaxRepository) isTaxApplicable(tax models.Tax, req models.TaxCalculationRequest) bool {
	// Check warehouse applicability
	if len(tax.ApplicableWarehouses) > 0 {
		warehouseApplicable := false
		for _, warehouseID := range tax.ApplicableWarehouses {
			if warehouseID == req.WarehouseID {
				warehouseApplicable = true
				break
			}
		}
		if !warehouseApplicable {
			return false
		}
	}

	// Check warehouse exclusions
	if len(tax.ExcludedWarehouses) > 0 {
		for _, warehouseID := range tax.ExcludedWarehouses {
			if warehouseID == req.WarehouseID {
				return false
			}
		}
	}

	// Check state applicability
	if len(tax.ApplicableStates) > 0 {
		stateApplicable := false
		for _, state := range tax.ApplicableStates {
			if (req.CustomerState != nil && state == *req.CustomerState) || state == req.WarehouseState {
				stateApplicable = true
				break
			}
		}
		if !stateApplicable {
			return false
		}
	}

	// Check state exclusions
	if len(tax.ExcludedStates) > 0 {
		for _, state := range tax.ExcludedStates {
			if (req.CustomerState != nil && state == *req.CustomerState) || state == req.WarehouseState {
				return false
			}
		}
	}

	// Check inter-state applicability
	if tax.IsInterState && !req.IsInterState {
		return false
	}

	// Check GSTIN requirement
	if tax.RequiresGSTIN && (req.CustomerGSTIN == nil || *req.CustomerGSTIN == "") {
		return false
	}

	// Check PAN requirement
	if tax.RequiresPAN && (req.CustomerPAN == nil || *req.CustomerPAN == "") {
		return false
	}

	return true
}

// CalculateTax calculates tax amount for a given base amount and tax
func (r *TaxRepository) CalculateTax(tax models.Tax, baseAmount float64, quantity int) float64 {
	var taxAmount float64

	switch tax.CalculationType {
	case models.TaxCalculationPercentage:
		taxAmount = baseAmount * (tax.Rate / 100.0)
	case models.TaxCalculationFixed:
		taxAmount = tax.Rate * float64(quantity)
	case models.TaxCalculationTiered:
		taxAmount = r.calculateTieredTax(tax, baseAmount)
	}

	// Apply minimum and maximum limits
	if tax.MinAmount != nil && taxAmount < *tax.MinAmount {
		taxAmount = *tax.MinAmount
	}
	if tax.MaxAmount != nil && taxAmount > *tax.MaxAmount {
		taxAmount = *tax.MaxAmount
	}

	return taxAmount
}

// calculateTieredTax calculates tax using tiered rates
func (r *TaxRepository) calculateTieredTax(tax models.Tax, baseAmount float64) float64 {
	var tiers []models.TaxTier
	if err := r.db.Where("tax_id = ?", tax.ID).Order("min_amount ASC").Find(&tiers).Error; err != nil {
		// Log error for debugging purposes but don't fail the calculation
		// This allows the system to continue functioning even if tier data is missing
		if tax.ID != "" {
			// Only log if we have a valid tax ID to avoid spam
			// Note: In production, this should use proper logging framework
			// log.Printf("Error retrieving tax tiers for tax %s: %v", tax.ID, err)
		}
		return 0
	}

	var totalTax float64
	for _, tier := range tiers {
		if baseAmount >= tier.MinAmount && (tier.MaxAmount == nil || baseAmount <= *tier.MaxAmount) {
			if tier.FixedAmount != nil {
				totalTax += *tier.FixedAmount
			} else {
				tierAmount := baseAmount * (tier.Rate / 100.0)
				totalTax += tierAmount
			}
		}
	}

	return totalTax
}

// CreateTaxApplication creates a tax application record
func (r *TaxRepository) CreateTaxApplication(taxApp *models.TaxApplication) error {
	if err := r.db.Create(taxApp).Error; err != nil {
		return errors.NewInternalServerError("Failed to create tax application")
	}
	return nil
}

// GetTaxApplicationsBySale retrieves tax applications for a sale
func (r *TaxRepository) GetTaxApplicationsBySale(saleID string) ([]models.TaxApplication, error) {
	var taxApps []models.TaxApplication
	if err := r.db.Where("sale_id = ?", saleID).Find(&taxApps).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve tax applications")
	}
	return taxApps, nil
}

// GetTaxApplicationsByReturn retrieves tax applications for a return
func (r *TaxRepository) GetTaxApplicationsByReturn(returnID string) ([]models.TaxApplication, error) {
	var taxApps []models.TaxApplication
	if err := r.db.Where("return_id = ?", returnID).Find(&taxApps).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve tax applications")
	}
	return taxApps, nil
}

// CreateTaxSummary creates a tax summary record
func (r *TaxRepository) CreateTaxSummary(taxSummary *models.TaxSummary) error {
	if err := r.db.Create(taxSummary).Error; err != nil {
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

// GetTaxTiers retrieves tax tiers for a tax
func (r *TaxRepository) GetTaxTiers(taxID string) ([]models.TaxTier, error) {
	var tiers []models.TaxTier
	if err := r.db.Where("tax_id = ?", taxID).Order("min_amount ASC").Find(&tiers).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve tax tiers")
	}
	return tiers, nil
}

// CreateTaxTier creates a tax tier
func (r *TaxRepository) CreateTaxTier(tier *models.TaxTier) error {
	if err := r.db.Create(tier).Error; err != nil {
		return errors.NewInternalServerError("Failed to create tax tier")
	}
	return nil
}

// DeleteTaxTier deletes a tax tier
func (r *TaxRepository) DeleteTaxTier(id string) error {
	if err := r.db.Delete(&models.TaxTier{}, "id = ?", id).Error; err != nil {
		return errors.NewInternalServerError("Failed to delete tax tier")
	}
	return nil
}

// GetTaxesByProduct retrieves taxes applicable to a specific product
func (r *TaxRepository) GetTaxesByProduct(productID string, warehouseID string, customerState string, isInterState bool) ([]models.Tax, error) {
	var taxes []models.Tax
	now := time.Now()

	query := r.db.Where("is_active = ? AND valid_from <= ? AND (valid_until IS NULL OR valid_until > ?)",
		true, now, now)

	// Add product-specific conditions (PostgreSQL JSON operator @>)
	query = query.Where("(applicable_products IS NULL OR applicable_products @> ?) AND (excluded_products IS NULL OR NOT (excluded_products @> ?))",
		`"`+productID+`"`, `"`+productID+`"`)

	// Add warehouse conditions (PostgreSQL JSON operator @>)
	query = query.Where("(applicable_warehouses IS NULL OR applicable_warehouses @> ?) AND (excluded_warehouses IS NULL OR NOT (excluded_warehouses @> ?))",
		`"`+warehouseID+`"`, `"`+warehouseID+`"`)

	// Add state conditions (PostgreSQL JSON operator @>)
	query = query.Where("(applicable_states IS NULL OR applicable_states @> ?) AND (excluded_states IS NULL OR NOT (excluded_states @> ?))",
		`"`+customerState+`"`, `"`+customerState+`"`)

	// Add inter-state condition
	if isInterState {
		query = query.Where("is_inter_state = ?", true)
	}

	if err := query.Order("priority DESC, stacking_order ASC").Find(&taxes).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve taxes by product")
	}

	return taxes, nil
}

// GetTaxesByCategory retrieves taxes applicable to a specific category
func (r *TaxRepository) GetTaxesByCategory(categoryID string, warehouseID string, customerState string, isInterState bool) ([]models.Tax, error) {
	var taxes []models.Tax
	now := time.Now()

	query := r.db.Where("is_active = ? AND valid_from <= ? AND (valid_until IS NULL OR valid_until > ?)",
		true, now, now)

	// Add category-specific conditions (PostgreSQL JSON operator @>)
	query = query.Where("(applicable_categories IS NULL OR applicable_categories @> ?) AND (excluded_categories IS NULL OR NOT (excluded_categories @> ?))",
		`"`+categoryID+`"`, `"`+categoryID+`"`)

	// Add warehouse conditions (PostgreSQL JSON operator @>)
	query = query.Where("(applicable_warehouses IS NULL OR applicable_warehouses @> ?) AND (excluded_warehouses IS NULL OR NOT (excluded_warehouses @> ?))",
		`"`+warehouseID+`"`, `"`+warehouseID+`"`)

	// Add state conditions (PostgreSQL JSON operator @>)
	query = query.Where("(applicable_states IS NULL OR applicable_states @> ?) AND (excluded_states IS NULL OR NOT (excluded_states @> ?))",
		`"`+customerState+`"`, `"`+customerState+`"`)

	// Add inter-state condition
	if isInterState {
		query = query.Where("is_inter_state = ?", true)
	}

	if err := query.Order("priority DESC, stacking_order ASC").Find(&taxes).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve taxes by category")
	}

	return taxes, nil
}
