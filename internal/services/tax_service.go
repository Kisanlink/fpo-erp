package services

import (
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
)

type TaxService struct {
	taxRepo *repositories.TaxRepository
}

func NewTaxService(taxRepo *repositories.TaxRepository) *TaxService {
	return &TaxService{
		taxRepo: taxRepo,
	}
}

// CreateTax creates a new tax
func (s *TaxService) CreateTax(req *models.CreateTaxRequest, userID string) (*models.TaxResponse, error) {
	// Check if tax code already exists
	existingTax, _ := s.taxRepo.GetTaxByCode(req.Code)
	if existingTax != nil {
		return nil, errors.NewBadRequestError("Tax code already exists")
	}

	// Validate date range
	if req.ValidUntil != nil && req.ValidFrom.After(*req.ValidUntil) {
		return nil, errors.NewBadRequestError("Valid from date cannot be after valid until date")
	}

	tax := models.NewTax()
	tax.Code = req.Code
	tax.Name = req.Name
	tax.Description = req.Description
	tax.TaxType = req.TaxType
	tax.CalculationType = req.CalculationType
	tax.Rate = req.Rate
	tax.MinAmount = req.MinAmount
	tax.MaxAmount = req.MaxAmount
	tax.MinOrderValue = req.MinOrderValue
	tax.MaxOrderValue = req.MaxOrderValue
	tax.ApplicableProducts = req.ApplicableProducts
	tax.ExcludedProducts = req.ExcludedProducts
	tax.ApplicableCategories = req.ApplicableCategories
	tax.ExcludedCategories = req.ExcludedCategories
	tax.ApplicableWarehouses = req.ApplicableWarehouses
	tax.ExcludedWarehouses = req.ExcludedWarehouses
	tax.ApplicableStates = req.ApplicableStates
	tax.ExcludedStates = req.ExcludedStates
	tax.ApplicableCustomerGroups = req.ApplicableCustomerGroups
	tax.ExcludedCustomerGroups = req.ExcludedCustomerGroups
	tax.ValidFrom = req.ValidFrom
	tax.ValidUntil = req.ValidUntil
	tax.IsActive = req.IsActive
	tax.Priority = req.Priority
	tax.IsStackable = req.IsStackable
	tax.StackingOrder = req.StackingOrder
	tax.RequiresGSTIN = req.RequiresGSTIN
	tax.RequiresPAN = req.RequiresPAN
	tax.IsInterState = req.IsInterState
	tax.HSNCode = req.HSNCode
	tax.SACCode = req.SACCode
	tax.TaxCategory = req.TaxCategory
	tax.CreatedBy = userID
	tax.UpdatedBy = userID

	if err := s.taxRepo.CreateTax(tax); err != nil {
		return nil, err
	}

	return tax.ToResponse(), nil
}

// GetTax retrieves a tax by ID
func (s *TaxService) GetTax(id string) (*models.TaxResponse, error) {
	tax, err := s.taxRepo.GetTaxByID(id)
	if err != nil {
		return nil, err
	}

	return tax.ToResponse(), nil
}

// GetAllTaxes retrieves all taxes with pagination
func (s *TaxService) GetAllTaxes(limit, offset int) ([]models.TaxResponse, error) {
	taxes, err := s.taxRepo.GetAllTaxes(limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []models.TaxResponse
	for _, tax := range taxes {
		responses = append(responses, *tax.ToResponse())
	}

	return responses, nil
}

// GetActiveTaxes retrieves all currently active taxes
func (s *TaxService) GetActiveTaxes() ([]models.TaxResponse, error) {
	taxes, err := s.taxRepo.GetActiveTaxes()
	if err != nil {
		return nil, err
	}

	var responses []models.TaxResponse
	for _, tax := range taxes {
		responses = append(responses, *tax.ToResponse())
	}

	return responses, nil
}

// GetTaxesByType retrieves taxes by type
func (s *TaxService) GetTaxesByType(taxType models.TaxType) ([]models.TaxResponse, error) {
	taxes, err := s.taxRepo.GetTaxesByType(taxType)
	if err != nil {
		return nil, err
	}

	var responses []models.TaxResponse
	for _, tax := range taxes {
		responses = append(responses, *tax.ToResponse())
	}

	return responses, nil
}

// GetTaxesByStatus retrieves taxes by status
func (s *TaxService) GetTaxesByStatus(status string) ([]models.TaxResponse, error) {
	taxes, err := s.taxRepo.GetTaxesByStatus(status)
	if err != nil {
		return nil, err
	}

	var responses []models.TaxResponse
	for _, tax := range taxes {
		responses = append(responses, *tax.ToResponse())
	}

	return responses, nil
}

// UpdateTax updates an existing tax
func (s *TaxService) UpdateTax(id string, req *models.UpdateTaxRequest, userID string) (*models.TaxResponse, error) {
	// Check if tax exists
	_, err := s.taxRepo.GetTaxByID(id)
	if err != nil {
		return nil, err
	}

	// Validate date range if both dates are provided
	if req.ValidFrom != nil && req.ValidUntil != nil && req.ValidFrom.After(*req.ValidUntil) {
		return nil, errors.NewBadRequestError("Valid from date cannot be after valid until date")
	}

	// Build updates map
	updates := make(map[string]interface{})
	updates["updated_by"] = userID
	updates["updated_at"] = time.Now()

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.CalculationType != nil {
		updates["calculation_type"] = *req.CalculationType
	}
	if req.Rate != nil {
		updates["rate"] = *req.Rate
	}
	if req.MinAmount != nil {
		updates["min_amount"] = *req.MinAmount
	}
	if req.MaxAmount != nil {
		updates["max_amount"] = *req.MaxAmount
	}
	if req.MinOrderValue != nil {
		updates["min_order_value"] = *req.MinOrderValue
	}
	if req.MaxOrderValue != nil {
		updates["max_order_value"] = *req.MaxOrderValue
	}
	if req.ApplicableProducts != nil {
		updates["applicable_products"] = req.ApplicableProducts
	}
	if req.ExcludedProducts != nil {
		updates["excluded_products"] = req.ExcludedProducts
	}
	if req.ApplicableCategories != nil {
		updates["applicable_categories"] = req.ApplicableCategories
	}
	if req.ExcludedCategories != nil {
		updates["excluded_categories"] = req.ExcludedCategories
	}
	if req.ApplicableWarehouses != nil {
		updates["applicable_warehouses"] = req.ApplicableWarehouses
	}
	if req.ExcludedWarehouses != nil {
		updates["excluded_warehouses"] = req.ExcludedWarehouses
	}
	if req.ApplicableStates != nil {
		updates["applicable_states"] = req.ApplicableStates
	}
	if req.ExcludedStates != nil {
		updates["excluded_states"] = req.ExcludedStates
	}
	if req.ApplicableCustomerGroups != nil {
		updates["applicable_customer_groups"] = req.ApplicableCustomerGroups
	}
	if req.ExcludedCustomerGroups != nil {
		updates["excluded_customer_groups"] = req.ExcludedCustomerGroups
	}
	if req.ValidFrom != nil {
		updates["valid_from"] = *req.ValidFrom
	}
	if req.ValidUntil != nil {
		updates["valid_until"] = *req.ValidUntil
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.IsStackable != nil {
		updates["is_stackable"] = *req.IsStackable
	}
	if req.StackingOrder != nil {
		updates["stacking_order"] = *req.StackingOrder
	}
	if req.RequiresGSTIN != nil {
		updates["requires_gstin"] = *req.RequiresGSTIN
	}
	if req.RequiresPAN != nil {
		updates["requires_pan"] = *req.RequiresPAN
	}
	if req.IsInterState != nil {
		updates["is_inter_state"] = *req.IsInterState
	}
	if req.HSNCode != nil {
		updates["hsn_code"] = *req.HSNCode
	}
	if req.SACCode != nil {
		updates["sac_code"] = *req.SACCode
	}
	if req.TaxCategory != nil {
		updates["tax_category"] = *req.TaxCategory
	}

	if err := s.taxRepo.UpdateTax(id, updates); err != nil {
		return nil, err
	}

	// Get updated tax
	updatedTax, err := s.taxRepo.GetTaxByID(id)
	if err != nil {
		return nil, err
	}

	return updatedTax.ToResponse(), nil
}

// DeleteTax deletes a tax
func (s *TaxService) DeleteTax(id string) error {
	// Check if tax exists
	_, err := s.taxRepo.GetTaxByID(id)
	if err != nil {
		return err
	}

	return s.taxRepo.DeleteTax(id)
}

// CalculateTax calculates taxes for a given transaction
func (s *TaxService) CalculateTax(req *models.TaxCalculationRequest) (*models.TaxCalculationResponse, error) {
	// Get applicable taxes
	applicableTaxes, err := s.taxRepo.GetApplicableTaxes(*req)
	if err != nil {
		return nil, err
	}

	// Calculate subtotal
	var subTotal float64
	for _, item := range req.Items {
		subTotal += item.LineTotal
	}

	// Check order value limits for taxes
	var filteredTaxes []models.Tax
	for _, tax := range applicableTaxes {
		if tax.MinOrderValue != nil && subTotal < *tax.MinOrderValue {
			continue
		}
		if tax.MaxOrderValue != nil && subTotal > *tax.MaxOrderValue {
			continue
		}
		filteredTaxes = append(filteredTaxes, tax)
	}

	// Calculate taxes for each item
	var taxBreakdown []models.TaxBreakdown
	var appliedTaxes []models.AppliedTax
	var totalTaxAmount float64

	// Group taxes by type for breakdown
	taxBreakdownMap := make(map[models.TaxType]*models.TaxBreakdown)

	for _, item := range req.Items {
		itemTaxes, err := s.calculateItemTaxes(item, req, filteredTaxes)
		if err != nil {
			return nil, err
		}

		for _, appliedTax := range itemTaxes {
			// Add to applied taxes
			appliedTaxes = append(appliedTaxes, appliedTax)
			totalTaxAmount += appliedTax.Amount

			// Update tax breakdown
			if breakdown, exists := taxBreakdownMap[appliedTax.TaxType]; exists {
				breakdown.Amount += appliedTax.Amount
			} else {
				taxBreakdownMap[appliedTax.TaxType] = &models.TaxBreakdown{
					TaxType: appliedTax.TaxType,
					TaxName: appliedTax.TaxName,
					TaxCode: appliedTax.TaxCode,
					Rate:    appliedTax.Rate,
					Amount:  appliedTax.Amount,
					HSNCode: nil, // Will be set from tax configuration
					SACCode: nil, // Will be set from tax configuration
				}
			}
		}
	}

	// Convert breakdown map to slice
	for _, breakdown := range taxBreakdownMap {
		taxBreakdown = append(taxBreakdown, *breakdown)
	}

	grandTotal := subTotal + totalTaxAmount

	return &models.TaxCalculationResponse{
		SubTotal:       subTotal,
		TaxBreakdown:   taxBreakdown,
		TotalTaxAmount: totalTaxAmount,
		GrandTotal:     grandTotal,
		AppliedTaxes:   appliedTaxes,
	}, nil
}

// calculateItemTaxes calculates taxes for a specific item
func (s *TaxService) calculateItemTaxes(item models.TaxCalculationItem, req *models.TaxCalculationRequest, applicableTaxes []models.Tax) ([]models.AppliedTax, error) {
	var appliedTaxes []models.AppliedTax

	// Get product-specific taxes
	customerState := ""
	if req.CustomerState != nil {
		customerState = *req.CustomerState
	}
	productTaxes, err := s.taxRepo.GetTaxesByProduct(item.ProductID, req.WarehouseID, customerState, req.IsInterState)
	if err != nil {
		return nil, err
	}

	// Get category-specific taxes if category is provided
	var categoryTaxes []models.Tax
	if item.CategoryID != nil {
		categoryTaxes, err = s.taxRepo.GetTaxesByCategory(*item.CategoryID, req.WarehouseID, customerState, req.IsInterState)
		if err != nil {
			return nil, err
		}
	}

	// Combine all applicable taxes
	allTaxes := append(applicableTaxes, productTaxes...)
	allTaxes = append(allTaxes, categoryTaxes...)

	// Remove duplicates and sort by priority
	uniqueTaxes := s.removeDuplicateTaxes(allTaxes)
	sortedTaxes := s.sortTaxesByPriority(uniqueTaxes)

	// Calculate taxes for this item
	for _, tax := range sortedTaxes {
		// Check if tax is applicable to this specific item
		if !s.isTaxApplicableToItem(tax, item, req) {
			continue
		}

		taxAmount := s.taxRepo.CalculateTax(tax, item.LineTotal, item.Quantity)

		if taxAmount > 0 {
			appliedTax := models.AppliedTax{
				TaxID:      tax.ID,
				TaxCode:    tax.Code,
				TaxName:    tax.Name,
				TaxType:    tax.TaxType,
				Rate:       tax.Rate,
				Amount:     taxAmount,
				BaseAmount: item.LineTotal,
			}

			appliedTaxes = append(appliedTaxes, appliedTax)
		}
	}

	return appliedTaxes, nil
}

// isTaxApplicableToItem checks if a tax is applicable to a specific item
func (s *TaxService) isTaxApplicableToItem(tax models.Tax, item models.TaxCalculationItem, req *models.TaxCalculationRequest) bool {
	// Check product applicability
	if len(tax.ApplicableProducts) > 0 {
		productApplicable := false
		for _, productID := range tax.ApplicableProducts {
			if productID == item.ProductID {
				productApplicable = true
				break
			}
		}
		if !productApplicable {
			return false
		}
	}

	// Check product exclusions
	if len(tax.ExcludedProducts) > 0 {
		for _, productID := range tax.ExcludedProducts {
			if productID == item.ProductID {
				return false
			}
		}
	}

	// Check category applicability
	if item.CategoryID != nil && len(tax.ApplicableCategories) > 0 {
		categoryApplicable := false
		for _, categoryID := range tax.ApplicableCategories {
			if categoryID == *item.CategoryID {
				categoryApplicable = true
				break
			}
		}
		if !categoryApplicable {
			return false
		}
	}

	// Check category exclusions
	if item.CategoryID != nil && len(tax.ExcludedCategories) > 0 {
		for _, categoryID := range tax.ExcludedCategories {
			if categoryID == *item.CategoryID {
				return false
			}
		}
	}

	return true
}

// removeDuplicateTaxes removes duplicate taxes based on ID
func (s *TaxService) removeDuplicateTaxes(taxes []models.Tax) []models.Tax {
	seen := make(map[string]bool)
	var uniqueTaxes []models.Tax

	for _, tax := range taxes {
		if !seen[tax.ID] {
			seen[tax.ID] = true
			uniqueTaxes = append(uniqueTaxes, tax)
		}
	}

	return uniqueTaxes
}

// sortTaxesByPriority sorts taxes by priority and stacking order
func (s *TaxService) sortTaxesByPriority(taxes []models.Tax) []models.Tax {
	// Simple bubble sort for priority and stacking order
	for i := 0; i < len(taxes)-1; i++ {
		for j := 0; j < len(taxes)-i-1; j++ {
			if taxes[j].Priority < taxes[j+1].Priority ||
				(taxes[j].Priority == taxes[j+1].Priority && taxes[j].StackingOrder > taxes[j+1].StackingOrder) {
				taxes[j], taxes[j+1] = taxes[j+1], taxes[j]
			}
		}
	}

	return taxes
}

// ApplyTaxesToSale applies taxes to a sale and creates tax applications
func (s *TaxService) ApplyTaxesToSale(saleID string, items []models.SaleItem, req *models.TaxCalculationRequest, userID string) (*models.TaxSummary, error) {
	// Calculate taxes
	taxCalculation, err := s.CalculateTax(req)
	if err != nil {
		return nil, err
	}

	// Create tax summary
	taxSummary := models.NewTaxSummary()
	taxSummary.SaleID = &saleID
	taxSummary.SubTotal = taxCalculation.SubTotal
	taxSummary.TotalTaxAmount = taxCalculation.TotalTaxAmount
	taxSummary.GrandTotal = taxCalculation.GrandTotal

	// Calculate tax breakdown by type
	for _, breakdown := range taxCalculation.TaxBreakdown {
		switch breakdown.TaxType {
		case models.TaxTypeCGST:
			taxSummary.CGSTAmount = breakdown.Amount
		case models.TaxTypeSGST:
			taxSummary.SGSTAmount = breakdown.Amount
		case models.TaxTypeIGST:
			taxSummary.IGSTAmount = breakdown.Amount
		case models.TaxTypeVAT:
			taxSummary.VATAmount = breakdown.Amount
		case models.TaxTypeSTT:
			taxSummary.STTAmount = breakdown.Amount
		case models.TaxTypeTDS:
			taxSummary.TDSAmount = breakdown.Amount
		case models.TaxTypeTCS:
			taxSummary.TCSAmount = breakdown.Amount
		case models.TaxTypeExcise:
			taxSummary.ExciseAmount = breakdown.Amount
		case models.TaxTypeCustoms:
			taxSummary.CustomsAmount = breakdown.Amount
		default:
			taxSummary.OtherTaxAmount += breakdown.Amount
		}
	}

	// Save tax summary
	if err := s.taxRepo.CreateTaxSummary(taxSummary); err != nil {
		return nil, err
	}

	// Create tax applications for each applied tax
	for _, appliedTax := range taxCalculation.AppliedTaxes {
		taxApp := models.NewTaxApplication()
		taxApp.TaxID = appliedTax.TaxID
		taxApp.SaleID = &saleID
		taxApp.BaseAmount = appliedTax.BaseAmount
		taxApp.TaxRate = appliedTax.Rate
		taxApp.TaxAmount = appliedTax.Amount
		taxApp.TaxType = appliedTax.TaxType

		if err := s.taxRepo.CreateTaxApplication(taxApp); err != nil {
			return nil, err
		}
	}

	return taxSummary, nil
}

// ApplyTaxesToReturn applies taxes to a return and creates tax applications
func (s *TaxService) ApplyTaxesToReturn(returnID string, items []models.ReturnItem, req *models.TaxCalculationRequest, userID string) (*models.TaxSummary, error) {
	// Calculate taxes (same logic as sale but for returns)
	taxCalculation, err := s.CalculateTax(req)
	if err != nil {
		return nil, err
	}

	// Create tax summary
	taxSummary := models.NewTaxSummary()
	taxSummary.ReturnID = &returnID
	taxSummary.SubTotal = taxCalculation.SubTotal
	taxSummary.TotalTaxAmount = taxCalculation.TotalTaxAmount
	taxSummary.GrandTotal = taxCalculation.GrandTotal

	// Calculate tax breakdown by type
	for _, breakdown := range taxCalculation.TaxBreakdown {
		switch breakdown.TaxType {
		case models.TaxTypeCGST:
			taxSummary.CGSTAmount = breakdown.Amount
		case models.TaxTypeSGST:
			taxSummary.SGSTAmount = breakdown.Amount
		case models.TaxTypeIGST:
			taxSummary.IGSTAmount = breakdown.Amount
		case models.TaxTypeVAT:
			taxSummary.VATAmount = breakdown.Amount
		case models.TaxTypeSTT:
			taxSummary.STTAmount = breakdown.Amount
		case models.TaxTypeTDS:
			taxSummary.TDSAmount = breakdown.Amount
		case models.TaxTypeTCS:
			taxSummary.TCSAmount = breakdown.Amount
		case models.TaxTypeExcise:
			taxSummary.ExciseAmount = breakdown.Amount
		case models.TaxTypeCustoms:
			taxSummary.CustomsAmount = breakdown.Amount
		default:
			taxSummary.OtherTaxAmount += breakdown.Amount
		}
	}

	// Save tax summary
	if err := s.taxRepo.CreateTaxSummary(taxSummary); err != nil {
		return nil, err
	}

	// Create tax applications for each applied tax
	for _, appliedTax := range taxCalculation.AppliedTaxes {
		taxApp := models.NewTaxApplication()
		taxApp.TaxID = appliedTax.TaxID
		taxApp.ReturnID = &returnID
		taxApp.BaseAmount = appliedTax.BaseAmount
		taxApp.TaxRate = appliedTax.Rate
		taxApp.TaxAmount = appliedTax.Amount
		taxApp.TaxType = appliedTax.TaxType

		if err := s.taxRepo.CreateTaxApplication(taxApp); err != nil {
			return nil, err
		}
	}

	return taxSummary, nil
}

// GetTaxSummaryBySale retrieves tax summary for a sale
func (s *TaxService) GetTaxSummaryBySale(saleID string) (*models.TaxSummary, error) {
	return s.taxRepo.GetTaxSummaryBySale(saleID)
}

// GetTaxSummaryByReturn retrieves tax summary for a return
func (s *TaxService) GetTaxSummaryByReturn(returnID string) (*models.TaxSummary, error) {
	return s.taxRepo.GetTaxSummaryByReturn(returnID)
}

// GetTaxApplicationsBySale retrieves tax applications for a sale
func (s *TaxService) GetTaxApplicationsBySale(saleID string) ([]models.TaxApplication, error) {
	return s.taxRepo.GetTaxApplicationsBySale(saleID)
}

// GetTaxApplicationsByReturn retrieves tax applications for a return
func (s *TaxService) GetTaxApplicationsByReturn(returnID string) ([]models.TaxApplication, error) {
	return s.taxRepo.GetTaxApplicationsByReturn(returnID)
}

// CalculateBatchTax calculates taxes for a sale item based on its inventory batch
func (s *TaxService) CalculateBatchTax(batch models.InventoryBatch, quantity int64, unitPrice float64) (*models.BatchTaxCalculation, error) {
	lineTotal := float64(quantity) * unitPrice

	// Check if batch is tax exempt
	if batch.IsTaxExempt {
		return &models.BatchTaxCalculation{
			BatchID:         batch.ID,
			LineTotal:       lineTotal,
			CGSTAmount:      0,
			SGSTAmount:      0,
			CustomTaxAmount: 0,
			TotalTaxAmount:  0,
		}, nil
	}

	// Calculate CGST
	cgstAmount := s.roundToNearestPaisa(lineTotal * (batch.CGSTRate / 100))

	// Calculate SGST
	sgstAmount := s.roundToNearestPaisa(lineTotal * (batch.SGSTRate / 100))

	// Calculate custom taxes
	customTaxAmount := float64(0)
	if len(batch.CustomTaxIDs) > 0 {
		customTaxes, err := s.taxRepo.GetTaxesByIDs(batch.CustomTaxIDs)
		if err != nil {
			return nil, err
		}

		for _, tax := range customTaxes {
			if tax.IsActive {
				taxAmount := s.calculateCustomTaxAmount(tax, lineTotal, quantity)
				customTaxAmount += taxAmount
			}
		}
	}

	totalTaxAmount := cgstAmount + sgstAmount + customTaxAmount

	return &models.BatchTaxCalculation{
		BatchID:         batch.ID,
		LineTotal:       lineTotal,
		CGSTAmount:      cgstAmount,
		SGSTAmount:      sgstAmount,
		CustomTaxAmount: customTaxAmount,
		TotalTaxAmount:  totalTaxAmount,
	}, nil
}

// calculateCustomTaxAmount calculates the amount for a custom tax
func (s *TaxService) calculateCustomTaxAmount(tax models.Tax, lineTotal float64, quantity int64) float64 {
	switch tax.CalculationType {
	case models.TaxCalculationPercentage:
		amount := lineTotal * (tax.Rate / 100)
		if tax.MinAmount != nil && amount < *tax.MinAmount {
			amount = *tax.MinAmount
		}
		if tax.MaxAmount != nil && amount > *tax.MaxAmount {
			amount = *tax.MaxAmount
		}
		return s.roundToNearestPaisa(amount)

	case models.TaxCalculationFixed:
		amount := tax.Rate * float64(quantity)
		if tax.MinAmount != nil && amount < *tax.MinAmount {
			amount = *tax.MinAmount
		}
		if tax.MaxAmount != nil && amount > *tax.MaxAmount {
			amount = *tax.MaxAmount
		}
		return s.roundToNearestPaisa(amount)

	default:
		return 0
	}
}

// roundToNearestPaisa rounds amount to nearest paisa (2 decimal places) for GST compliance
func (s *TaxService) roundToNearestPaisa(amount float64) float64 {
	return float64(int(amount*100+0.5)) / 100
}
