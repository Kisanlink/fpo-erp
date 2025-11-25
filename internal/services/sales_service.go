package services

import (
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Helper function to convert string to pointer
func stringPtr(s string) *string {
	return &s
}

type SalesService struct {
	salesRepo            *repositories.SalesRepository
	productRepo          *repositories.ProductRepository
	inventoryRepo        *repositories.InventoryRepository
	priceRepo            *repositories.ProductPriceRepository
	discountsRepo        *repositories.DiscountsRepository
	taxRepo              *repositories.TaxRepository
	taxService           *TaxService
	warehouseRepo        *repositories.WarehouseRepository
	saleCancellationRepo *repositories.SaleCancellationRepository
	logger               interfaces.Logger
}

func NewSalesService(salesRepo *repositories.SalesRepository, productRepo *repositories.ProductRepository, inventoryRepo *repositories.InventoryRepository, priceRepo *repositories.ProductPriceRepository, discountsRepo *repositories.DiscountsRepository, taxRepo *repositories.TaxRepository, warehouseRepo *repositories.WarehouseRepository, saleCancellationRepo *repositories.SaleCancellationRepository, logger interfaces.Logger) *SalesService {
	return &SalesService{
		salesRepo:            salesRepo,
		productRepo:          productRepo,
		inventoryRepo:        inventoryRepo,
		priceRepo:            priceRepo,
		discountsRepo:        discountsRepo,
		taxRepo:              taxRepo,
		taxService:           NewTaxService(taxRepo, logger),
		warehouseRepo:        warehouseRepo,
		saleCancellationRepo: saleCancellationRepo,
		logger:               logger,
	}
}

// CreateSale creates a new sale with items and summary using database transaction
func (s *SalesService) CreateSale(req *models.CreateSaleRequest) (*models.SaleResponse, error) {
	s.logger.Info("Starting transactional sale creation",
		zap.String("warehouse_id", req.WarehouseID),
		zap.Int("item_count", len(req.Items)))

	// Validate sale request
	if err := s.validateSaleRequest(req); err != nil {
		s.logger.Error("Sale validation failed",
			zap.Error(err),
			zap.String("warehouse_id", req.WarehouseID))
		return nil, err
	}
	s.logger.Debug("Sale validation passed")

	// Pre-fetch all required data BEFORE transaction to avoid SQLite deadlocks
	// This follows the same pattern as CreatePurchaseOrder
	type itemData struct {
		sellingPrice float64
		batches      []models.InventoryBatch
	}

	itemDataMap := make(map[string]*itemData)
	for _, itemReq := range req.Items {
		// Get selling price from product_prices table (by variant_id)
		s.logger.Debug("Getting selling price for variant",
			zap.String("variant_id", itemReq.VariantID))
		sellingPrice, err := s.getSellingPrice(itemReq.VariantID)
		if err != nil {
			s.logger.Error("Failed to get selling price",
				zap.Error(err),
				zap.String("variant_id", itemReq.VariantID))
			return nil, errors.NewNotFoundError("selling price not found for product")
		}
		s.logger.Debug("Selling price retrieved",
			zap.Float64("selling_price", sellingPrice))

		// Get batches for this variant in the warehouse ordered by expiry date (FEFO)
		s.logger.Debug("Getting batches for variant",
			zap.String("variant_id", itemReq.VariantID),
			zap.String("warehouse_id", req.WarehouseID))
		batches, err := s.inventoryRepo.GetBatchesByVariantAndWarehouseOrderedByExpiry(itemReq.VariantID, req.WarehouseID)
		if err != nil {
			s.logger.Error("Failed to get batches for variant",
				zap.Error(err),
				zap.String("variant_id", itemReq.VariantID))
			return nil, errors.NewInternalServerError("failed to retrieve variant batches")
		}

		if len(batches) == 0 {
			s.logger.Error("No batches found for variant",
				zap.String("variant_id", itemReq.VariantID),
				zap.String("warehouse_id", req.WarehouseID))
			return nil, errors.NewNotFoundError("no inventory available for variant in this warehouse")
		}

		s.logger.Debug("Found batches for product",
			zap.Int("batch_count", len(batches)))

		// Calculate total available quantity across all batches
		totalAvailable := int64(0)
		for _, batch := range batches {
			totalAvailable += batch.TotalQuantity
		}

		if totalAvailable < itemReq.Quantity {
			s.logger.Error("Insufficient stock",
				zap.Int64("available", totalAvailable),
				zap.Int64("requested", itemReq.Quantity))
			return nil, errors.NewBadRequestError("insufficient stock for product")
		}

		s.logger.Debug("Stock validation passed",
			zap.Int64("available", totalAvailable),
			zap.Int64("requested", itemReq.Quantity))

		itemDataMap[itemReq.VariantID] = &itemData{
			sellingPrice: sellingPrice,
			batches:      batches,
		}
	}

	// Pre-build product IDs map for discount discovery (using pre-fetched batch data)
	// This avoids calling GetBatchByID inside the transaction
	productIDsMap := make(map[string]bool)
	for _, itemData := range itemDataMap {
		for _, batch := range itemData.batches {
			productIDsMap[batch.VariantID] = true
		}
	}
	var productIDs []string
	for variantID := range productIDsMap {
		productIDs = append(productIDs, variantID)
	}

	var response *models.SaleResponse

	// Execute everything within a database transaction
	err := s.salesRepo.WithTransaction(func(tx *gorm.DB) error {
		// Parse sale date or use current time
		var saleDate time.Time
		if req.SaleDate != nil {
			s.logger.Debug("Parsing sale date",
				zap.String("sale_date", *req.SaleDate))
			if parsedDate, err := time.Parse(time.RFC3339, *req.SaleDate); err == nil {
				saleDate = parsedDate
				s.logger.Debug("Sale date parsed successfully",
					zap.Time("sale_date", saleDate))
			} else {
				// If parsing fails, use current time
				s.logger.Warn("Date parsing failed, using current time",
					zap.Error(err))
				saleDate = time.Now()
			}
		} else {
			// If no sale date provided, use current time
			s.logger.Debug("No sale date provided, using current time")
			saleDate = time.Now()
		}

		// Handle ApplyTaxes - default to false if not provided
		applyTaxes := false
		if req.ApplyTaxes != nil {
			applyTaxes = *req.ApplyTaxes
		}

		// Create sale using the proper constructor with BRD requirements
		s.logger.Debug("Creating sale",
			zap.String("warehouse_id", req.WarehouseID),
			zap.Time("sale_date", saleDate),
			zap.String("payment_mode", req.PaymentMode),
			zap.String("sale_type", req.SaleType),
			zap.Bool("apply_taxes", applyTaxes))
		sale := models.NewSale(req.WarehouseID, saleDate, 0, "pending", req.FarmerID, req.PaymentMode, req.SaleType, applyTaxes)
		s.logger.Info("Sale created",
			zap.String("sale_id", sale.ID),
			zap.Bool("apply_taxes", sale.ApplyTaxes))

		if err := s.salesRepo.CreateSaleWithTx(tx, sale); err != nil {
			s.logger.Error("Failed to create sale in database",
				zap.Error(err))
			return err
		}
		s.logger.Debug("Sale created successfully in database")

		// Create sale items using pre-fetched data
		s.logger.Debug("Starting to process sale items",
			zap.Int("item_count", len(req.Items)))
		var totalAmount float64
		var saleItems []models.SaleItem // Collect sale items for tax calculation
		for i, itemReq := range req.Items {
			s.logger.Debug("Processing sale item",
				zap.Int("item_number", i+1),
				zap.String("variant_id", itemReq.VariantID),
				zap.Int64("quantity", itemReq.Quantity))

			// Get pre-fetched data
			itemData := itemDataMap[itemReq.VariantID]
			sellingPrice := itemData.sellingPrice
			batches := itemData.batches

			// Allocate quantity across batches using FEFO (First Expired, First Out)
			remainingQuantity := itemReq.Quantity
			var itemTotal float64

			for _, batch := range batches {
				if remainingQuantity <= 0 {
					break
				}

				// Calculate how much to take from this batch
				quantityFromBatch := remainingQuantity
				if batch.TotalQuantity < remainingQuantity {
					quantityFromBatch = batch.TotalQuantity
				}

				s.logger.Debug("FEFO: Allocating from batch",
					zap.Int64("quantity", quantityFromBatch),
					zap.String("batch_id", batch.ID),
					zap.String("expiry_date", batch.ExpiryDate.Format("2006-01-02")))

				// BRD Requirement: Capture cost price from batch for margin calculation
				costPrice := batch.CostPrice
				s.logger.Debug("Cost price and margin calculation",
					zap.Float64("cost_price", costPrice),
					zap.Float64("selling_price", sellingPrice),
					zap.Float64("margin", sellingPrice-costPrice))

				// Calculate line total for this batch allocation
				batchLineTotal := sellingPrice * float64(quantityFromBatch)

				// Calculate taxes for this batch allocation
				s.logger.Debug("Calculating taxes for batch allocation")
				taxCalculation, err := s.taxService.CalculateBatchTax(batch, quantityFromBatch, sellingPrice)
				if err != nil {
					s.logger.Error("Failed to calculate taxes",
						zap.Error(err))
					return err
				}
				s.logger.Debug("Tax calculation completed",
					zap.Float64("cgst_amount", taxCalculation.CGSTAmount),
					zap.Float64("sgst_amount", taxCalculation.SGSTAmount),
					zap.Float64("custom_tax_amount", taxCalculation.CustomTaxAmount),
					zap.Float64("total_tax_amount", taxCalculation.TotalTaxAmount))

				itemTotal += batchLineTotal

				// Create sale item with tax amounts and cost price (BRD requirement)
				saleItem := models.NewSaleItemWithTax(sale.ID, batch.ID, quantityFromBatch, sellingPrice, costPrice, batchLineTotal,
					taxCalculation.CGSTAmount, taxCalculation.SGSTAmount, taxCalculation.CustomTaxAmount)
				s.logger.Info("Sale item created",
					zap.String("sale_item_id", saleItem.ID),
					zap.Float64("cost_price", costPrice),
					zap.Float64("margin", saleItem.Margin))

				if err := s.salesRepo.CreateSaleItemWithTx(tx, saleItem); err != nil {
					s.logger.Error("Failed to create sale item",
						zap.Error(err))
					return err
				}
				s.logger.Debug("Sale item created successfully in database")

				// Add to collection for tax calculation
				saleItems = append(saleItems, *saleItem)

				// Update inventory using the proper constructor
				s.logger.Debug("Creating inventory transaction",
					zap.String("batch_id", batch.ID))
				transaction := models.NewInventoryTransaction(batch.ID, "sale", -quantityFromBatch, &sale.ID, nil, stringPtr("Sale transaction"), time.Now())
				s.logger.Debug("Inventory transaction created",
					zap.String("transaction_id", transaction.ID))

				if err := s.inventoryRepo.CreateTransactionWithTx(tx, transaction); err != nil {
					s.logger.Error("Failed to create inventory transaction",
						zap.Error(err))
					return err
				}
				s.logger.Debug("Inventory transaction created successfully")

				// Update batch stock level with row lock to prevent race conditions
				s.logger.Debug("Updating batch stock",
					zap.String("batch_id", batch.ID),
					zap.Int64("quantity_change", -quantityFromBatch))
				if err := s.inventoryRepo.UpdateBatchStockWithTx(tx, batch.ID, -quantityFromBatch); err != nil {
					s.logger.Error("Failed to update batch stock",
						zap.Error(err))
					return err
				}
				s.logger.Debug("Batch stock updated successfully")

				remainingQuantity -= quantityFromBatch
			}

			totalAmount += itemTotal
			s.logger.Debug("Running total",
				zap.Float64("total_amount", totalAmount))
		}

		// Use pre-built product IDs for discount discovery (already built before transaction)

		// Apply discounts using priority-based resolution
		var discountAmount float64
		var appliedDiscounts []models.DiscountApplication
		// Convert saleItems to pointer slice for discount calculation
		var saleItemPtrs []*models.SaleItem
		for i := range saleItems {
			saleItemPtrs = append(saleItemPtrs, &saleItems[i])
		}

		finalDiscounts, applications, totalDiscountAmount, err := s.resolveDiscountsWithPriority(req, totalAmount, productIDs, saleItemPtrs)
		if err != nil {
			s.logger.Error("Failed to resolve discounts",
				zap.Error(err))
			return err
		}

		discountAmount = totalDiscountAmount
		appliedDiscounts = applications

		// Create discount usage records for applied discounts
		for _, discount := range finalDiscounts {
			discountUsage := s.discountsRepo.CalculateDiscount(&discount, totalAmount)
			usage := models.NewDiscountUsage(discount.ID, sale.ID, discountUsage)
			if err := s.discountsRepo.CreateDiscountUsageWithTx(tx, usage); err != nil {
				s.logger.Error("Failed to create discount usage record",
					zap.Error(err))
				return err
			}
			if err := s.discountsRepo.IncrementUsageWithTx(tx, discount.ID); err != nil {
				s.logger.Error("Failed to increment discount usage",
					zap.Error(err))
				return err
			}
		}

		s.logger.Info("Discounts applied",
			zap.Float64("discount_amount", discountAmount),
			zap.Int("discount_count", len(finalDiscounts)))

		// Calculate final amount after discount
		finalAmount := totalAmount - discountAmount
		s.logger.Debug("Final amount after discount",
			zap.Float64("final_amount", finalAmount))

		// Apply taxes using the existing tax service (no customer data needed)
		// Only apply taxes if ApplyTaxes field is true
		var taxAmount float64
		if sale.ApplyTaxes && len(saleItems) > 0 {
			s.logger.Debug("ApplyTaxes is true, calculating taxes")
			taxSummary, err := s.applyTaxesToSaleWithTx(tx, sale.ID, saleItems, req.WarehouseID)
			if err != nil {
				s.logger.Error("Tax calculation failed",
					zap.Error(err))
				return err
			}
			if taxSummary != nil {
				taxAmount = taxSummary.TotalTaxAmount
				finalAmount += taxAmount
				s.logger.Info("Tax applied successfully",
					zap.Float64("tax_amount", taxAmount))
			}
		} else {
			s.logger.Debug("ApplyTaxes is false, skipping tax calculation")
		}

		// Update sale with final amount
		s.logger.Debug("Updating sale with final amount",
			zap.Float64("final_amount", finalAmount))
		sale.TotalAmount = finalAmount
		if err := s.salesRepo.UpdateSaleWithTx(tx, sale); err != nil {
			s.logger.Error("Failed to update sale with final amount",
				zap.Error(err))
			return err
		}
		s.logger.Debug("Sale updated with final amount successfully")

		// Load sale items into sale.Items for response mapping
		// The sale object doesn't have items preloaded, so we set them from the saleItems we created
		sale.Items = saleItems

		// Build response within transaction (before committing)
		response = s.mapSaleToResponse(sale)

		// Add breakdown information
		var taxBreakdown *models.TaxSummaryBreakdown
		if taxAmount > 0 {
			taxSummary, err := s.taxRepo.GetTaxSummaryBySale(sale.ID)
			if err == nil {
				taxBreakdown = &models.TaxSummaryBreakdown{
					CGSTAmount:     taxSummary.CGSTAmount,
					SGSTAmount:     taxSummary.SGSTAmount,
					IGSTAmount:     taxSummary.IGSTAmount,
					VATAmount:      taxSummary.VATAmount,
					OtherTaxAmount: taxSummary.OtherTaxAmount,
					TotalTaxAmount: taxSummary.TotalTaxAmount,
				}
			}
		}

		response.Breakdown = &models.SaleBreakdown{
			BaseAmount:       totalAmount,
			AppliedDiscounts: appliedDiscounts,
			DiscountAmount:   discountAmount,
			TaxBreakdown:     taxBreakdown,
			TaxAmount:        taxAmount,
			TotalSavings:     discountAmount,
			FinalAmount:      finalAmount,
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Transaction failed",
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("Transactional sale creation completed successfully")
	return response, nil
}

// GetSale retrieves a sale by ID
func (s *SalesService) GetSale(id string) (*models.SaleResponse, error) {
	sale, err := s.salesRepo.GetSaleByID(id)
	if err != nil {
		return nil, err
	}
	return s.mapSaleToResponse(sale), nil
}

// GetAllSales retrieves all sales with pagination
func (s *SalesService) GetAllSales(limit, offset int) ([]models.SaleResponse, error) {
	sales, err := s.salesRepo.GetAllSales(limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []models.SaleResponse
	for _, sale := range sales {
		responses = append(responses, *s.mapSaleToResponse(&sale))
	}

	return responses, nil
}

// UpdateSale updates a sale
func (s *SalesService) UpdateSale(id string, req *models.UpdateSaleRequest) (*models.SaleResponse, error) {
	sale, err := s.salesRepo.GetSaleByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Status != nil {
		sale.Status = *req.Status
	}

	if err := s.salesRepo.UpdateSale(sale); err != nil {
		return nil, err
	}

	return s.mapSaleToResponse(sale), nil
}

// DeleteSale deletes a sale
func (s *SalesService) DeleteSale(id string) error {
	return s.salesRepo.DeleteSale(id)
}

// GetSalesByDateRange retrieves sales within a date range
func (s *SalesService) GetSalesByDateRange(startDate, endDate time.Time) ([]models.SaleResponse, error) {
	sales, err := s.salesRepo.GetSalesByDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	var responses []models.SaleResponse
	for _, sale := range sales {
		responses = append(responses, *s.mapSaleToResponse(&sale))
	}

	return responses, nil
}

// GetSalesByStatus retrieves sales by status
func (s *SalesService) GetSalesByStatus(status string) ([]models.SaleResponse, error) {
	sales, err := s.salesRepo.GetSalesByStatus(status)
	if err != nil {
		return nil, err
	}

	var responses []models.SaleResponse
	for _, sale := range sales {
		responses = append(responses, *s.mapSaleToResponse(&sale))
	}

	return responses, nil
}

// GetTotalSalesAmount calculates total sales amount for a date range
func (s *SalesService) GetTotalSalesAmount(startDate, endDate time.Time) (float64, error) {
	return s.salesRepo.GetTotalSalesAmount(startDate, endDate)
}

// GetTopSellingProducts retrieves top selling products
func (s *SalesService) GetTopSellingProducts(limit int) ([]models.TopSellingProductResponse, error) {
	results, err := s.salesRepo.GetTopSellingProducts(limit)
	if err != nil {
		return nil, err
	}

	var responses []models.TopSellingProductResponse
	for _, result := range results {
		responses = append(responses, models.TopSellingProductResponse{
			ProductID:   result.ProductID,
			ProductName: result.ProductName,
			TotalSold:   result.TotalSold,
			TotalAmount: result.TotalAmount,
		})
	}

	return responses, nil
}

// getSellingPrice retrieves the current retail price for a variant
func (s *SalesService) getSellingPrice(variantID string) (float64, error) {
	// Get the active retail price for the variant
	price, err := s.priceRepo.GetCurrentPrice(variantID, "retail")
	if err != nil {
		// Try to get any active price if retail price is not found
		prices, err2 := s.priceRepo.GetActiveByVariantID(variantID)
		if err2 != nil {
			return 0, errors.NewNotFoundError("no pricing information found for variant")
		}
		if len(prices) == 0 {
			return 0, errors.NewNotFoundError("no active prices found for variant")
		}
		// Use the first active price as fallback
		return prices[0].Price, nil
	}
	return price.Price, nil
}

// isValidPaymentMode validates payment mode (BRD requirement)
func isValidPaymentMode(mode string) bool {
	validModes := []string{"cash", "upi", "online"}
	for _, validMode := range validModes {
		if mode == validMode {
			return true
		}
	}
	return false
}

// isValidSaleType validates sale type (BRD requirement)
func isValidSaleType(saleType string) bool {
	validTypes := []string{"in_store", "delivery"}
	for _, validType := range validTypes {
		if saleType == validType {
			return true
		}
	}
	return false
}

// Helper methods
func (s *SalesService) validateSaleRequest(req *models.CreateSaleRequest) error {
	s.logger.Debug("Validating sale request",
		zap.String("warehouse_id", req.WarehouseID),
		zap.Int("item_count", len(req.Items)))

	if req.WarehouseID == "" {
		s.logger.Error("Validation failed: warehouse ID is empty")
		return errors.NewValidationError("warehouse ID is required")
	}
	if len(req.Items) == 0 {
		s.logger.Error("Validation failed: no items provided")
		return errors.NewValidationError("at least one item is required")
	}

	// BRD Requirements: Validate payment_mode and sale_type
	if !isValidPaymentMode(req.PaymentMode) {
		s.logger.Error("Validation failed: invalid payment mode",
			zap.String("payment_mode", req.PaymentMode))
		return errors.NewValidationError("payment_mode must be one of: cash, upi, online")
	}
	if !isValidSaleType(req.SaleType) {
		s.logger.Error("Validation failed: invalid sale type",
			zap.String("sale_type", req.SaleType))
		return errors.NewValidationError("sale_type must be one of: in_store, delivery")
	}

	for i, item := range req.Items {
		s.logger.Debug("Validating item",
			zap.Int("item_number", i+1),
			zap.String("variant_id", item.VariantID),
			zap.Int64("quantity", item.Quantity))

		if item.VariantID == "" {
			s.logger.Error("Validation failed: variant ID is empty",
				zap.Int("item_number", i+1))
			return errors.NewValidationError("variant ID is required for all items")
		}
		if item.Quantity <= 0 {
			s.logger.Error("Validation failed: quantity <= 0",
				zap.Int("item_number", i+1))
			return errors.NewValidationError("quantity must be greater than 0")
		}
		// Remove selling price validation since it will be calculated automatically from product_prices table
	}

	s.logger.Debug("Sale request validation passed")
	return nil
}

func (s *SalesService) mapSaleToResponse(sale *models.Sale) *models.SaleResponse {
	response := &models.SaleResponse{
		ID:          sale.ID,
		WarehouseID: sale.WarehouseID,
		SaleDate:    sale.SaleDate.Format("2006-01-02T15:04:05Z07:00"),
		TotalAmount: sale.TotalAmount,
		Status:      sale.Status,
		// BRD Requirements
		FarmerID:    sale.FarmerID,
		PaymentMode: sale.PaymentMode,
		SaleType:    sale.SaleType,
		ApplyTaxes:  sale.ApplyTaxes,
		CreatedAt:   sale.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   sale.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Add cancellation fields if present
	if sale.CancelledAt != nil {
		cancelledAtStr := sale.CancelledAt.Format("2006-01-02T15:04:05Z07:00")
		response.CancelledAt = &cancelledAtStr
	}
	if sale.CancellationReason != nil {
		response.CancellationReason = sale.CancellationReason
	}

	// Map items
	for _, item := range sale.Items {
		response.Items = append(response.Items, models.SaleItemResponse{
			ID:           item.ID,
			SaleID:       item.SaleID,
			BatchID:      item.BatchID,
			Quantity:     item.Quantity,
			SellingPrice: item.SellingPrice,
			LineTotal:    item.LineTotal,
			// BRD Requirements - Cost and Margin
			CostPrice:       item.CostPrice,
			Margin:          item.Margin,
			CGSTAmount:      item.CGSTAmount,
			SGSTAmount:      item.SGSTAmount,
			CustomTaxAmount: item.CustomTaxAmount,
			TotalTaxAmount:  item.TotalTaxAmount,
			CreatedAt:       item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return response
}

// applyDiscountToSale applies a discount to a sale and returns the discount amount
func (s *SalesService) applyDiscountToSale(discountID string, saleID string, orderValue float64) (float64, error) {
	// Get discount by ID
	discount, err := s.discountsRepo.GetDiscountByID(discountID)
	if err != nil {
		return 0, err
	}

	// Basic discount validation (no customer-specific validation)
	if !discount.IsActive {
		return 0, errors.NewBadRequestError("discount is not active")
	}

	// Check usage limits
	if discount.UsageLimit != nil && discount.CurrentUsage >= *discount.UsageLimit {
		return 0, errors.NewBadRequestError("discount usage limit reached")
	}

	// Check date validity
	now := time.Now()
	if now.Before(discount.ValidFrom) || now.After(discount.ValidUntil) {
		return 0, errors.NewBadRequestError("discount is not valid for the current date")
	}

	// Check minimum order value
	if discount.MinOrderValue != nil && orderValue < *discount.MinOrderValue {
		return 0, errors.NewBadRequestError("order value does not meet minimum requirement")
	}

	// Check maximum order value
	if discount.MaxOrderValue != nil && orderValue > *discount.MaxOrderValue {
		return 0, errors.NewBadRequestError("order value exceeds maximum limit")
	}

	// Calculate discount amount
	discountAmount := s.discountsRepo.CalculateDiscount(discount, orderValue)

	// Create discount usage record (without customer ID)
	usage := models.NewDiscountUsage(discountID, saleID, discountAmount)
	if err := s.discountsRepo.CreateDiscountUsage(usage); err != nil {
		return 0, err
	}

	// Increment usage count
	if err := s.discountsRepo.IncrementUsage(discountID); err != nil {
		return 0, err
	}

	return discountAmount, nil
}

// discoverApplicableDiscounts automatically finds the best applicable discounts for an order
func (s *SalesService) discoverApplicableDiscounts(orderValue float64, productIDs []string, warehouseID string) ([]models.Discount, error) {
	// Get all applicable discounts for this order
	applicableDiscounts, err := s.discountsRepo.GetApplicableDiscountsForOrder(orderValue, productIDs, nil, warehouseID)
	if err != nil {
		return nil, err
	}

	if len(applicableDiscounts) == 0 {
		return []models.Discount{}, nil
	}

	// Calculate optimal discount combination
	optimalDiscounts, _, err := s.discountsRepo.CalculateOptimalDiscounts(orderValue, productIDs, nil, warehouseID)
	if err != nil {
		// If optimal calculation fails, return single best discount
		bestDiscount := s.findBestSingleDiscount(applicableDiscounts, orderValue)
		if bestDiscount != nil {
			return []models.Discount{*bestDiscount}, nil
		}
		return []models.Discount{}, nil
	}

	// Convert responses to models for internal use
	var result []models.Discount
	for _, discountResp := range optimalDiscounts {
		discount, err := s.discountsRepo.GetDiscountByID(discountResp.ID)
		if err == nil {
			result = append(result, *discount)
		}
	}

	return result, nil
}

// findBestSingleDiscount finds the single best discount from available options
func (s *SalesService) findBestSingleDiscount(discounts []models.Discount, orderValue float64) *models.Discount {
	var bestDiscount *models.Discount
	var maxSavings float64

	for _, discount := range discounts {
		savings := s.discountsRepo.CalculateDiscount(&discount, orderValue)
		if savings > maxSavings {
			maxSavings = savings
			bestDiscount = &discount
		}
	}

	return bestDiscount
}

// resolveDiscountsWithPriority resolves which discounts to apply based on priority
func (s *SalesService) resolveDiscountsWithPriority(req *models.CreateSaleRequest, orderValue float64, productIDs []string, saleItems []*models.SaleItem) ([]models.Discount, []models.DiscountApplication, float64, error) {
	var finalDiscounts []models.Discount
	var applications []models.DiscountApplication
	var totalDiscountAmount float64

	// Priority 1: Manual discount by ID
	if req.DiscountID != nil && *req.DiscountID != "" {
		discount, err := s.discountsRepo.GetDiscountByID(*req.DiscountID)
		if err != nil {
			return nil, nil, 0, err
		}

		discountAmount := s.calculateDiscountAmount(discount, orderValue, saleItems)
		finalDiscounts = append(finalDiscounts, *discount)
		applications = append(applications, models.DiscountApplication{
			DiscountID:   discount.ID,
			DiscountCode: discount.Code,
			DiscountName: discount.Name,
			DiscountType: string(discount.DiscountType),
			Amount:       discountAmount,
			AppliedBy:    "manual",
		})
		totalDiscountAmount += discountAmount

		return finalDiscounts, applications, totalDiscountAmount, nil
	}

	// Priority 2: Manual discount by coupon code
	if req.CouponCode != nil && *req.CouponCode != "" {
		discount, err := s.discountsRepo.GetDiscountByCode(*req.CouponCode)
		if err != nil {
			return nil, nil, 0, err
		}

		discountAmount := s.calculateDiscountAmount(discount, orderValue, saleItems)
		finalDiscounts = append(finalDiscounts, *discount)
		applications = append(applications, models.DiscountApplication{
			DiscountID:   discount.ID,
			DiscountCode: discount.Code,
			DiscountName: discount.Name,
			DiscountType: string(discount.DiscountType),
			Amount:       discountAmount,
			AppliedBy:    "coupon",
		})
		totalDiscountAmount += discountAmount

		return finalDiscounts, applications, totalDiscountAmount, nil
	}

	// Priority 3: Auto-discovered discounts (default enabled)
	autoApply := true // Default value
	if req.AutoApplyDiscounts != nil {
		autoApply = *req.AutoApplyDiscounts
	}

	if autoApply {
		autoDiscounts, err := s.discoverApplicableDiscounts(orderValue, productIDs, req.WarehouseID)
		if err != nil {
			s.logger.Warn("Failed to discover automatic discounts",
				zap.Error(err))
			return []models.Discount{}, []models.DiscountApplication{}, 0, nil
		}

		for _, discount := range autoDiscounts {
			discountAmount := s.discountsRepo.CalculateDiscount(&discount, orderValue)
			finalDiscounts = append(finalDiscounts, discount)
			applications = append(applications, models.DiscountApplication{
				DiscountID:   discount.ID,
				DiscountCode: discount.Code,
				DiscountName: discount.Name,
				DiscountType: string(discount.DiscountType),
				Amount:       discountAmount,
				AppliedBy:    "auto",
			})
			totalDiscountAmount += discountAmount
		}
	}

	return finalDiscounts, applications, totalDiscountAmount, nil
}

// applyTaxesToSaleWithTx applies taxes to a sale using warehouse-based calculation within a transaction
func (s *SalesService) applyTaxesToSaleWithTx(tx *gorm.DB, saleID string, saleItems []models.SaleItem, warehouseID string) (*models.TaxSummary, error) {
	// Convert sale items to tax calculation items
	var taxItems []models.TaxCalculationItem
	for _, item := range saleItems {
		// Get batch to retrieve product ID
		batch, err := s.inventoryRepo.GetBatchByID(item.BatchID)
		if err != nil {
			return nil, err
		}

		taxItem := models.TaxCalculationItem{
			ProductID:  batch.VariantID,
			CategoryID: nil, // No category management in current model
			Quantity:   int(item.Quantity),
			UnitPrice:  item.SellingPrice,
			LineTotal:  item.LineTotal,
		}
		taxItems = append(taxItems, taxItem)
	}

	// Default state since warehouse doesn't store state directly (can be configurable)
	defaultState := "DefaultState" // This should be configurable or fetched from address service

	// Create tax calculation request without customer information
	taxReq := &models.TaxCalculationRequest{
		CustomerID:     nil, // No customer management
		CustomerState:  nil, // No customer management
		CustomerGSTIN:  nil, // No customer GSTIN
		CustomerPAN:    nil, // No customer PAN
		WarehouseID:    warehouseID,
		WarehouseState: defaultState, // Use default state for warehouse
		Items:          taxItems,
		IsInterState:   false, // Default to intra-state for warehouse-based taxation
	}

	// Use the tax service to apply taxes and create summary within transaction
	return s.taxService.ApplyTaxesToSaleWithTx(tx, saleID, saleItems, taxReq, "system")
}

// applyTaxesToSale applies taxes to a sale using warehouse-based calculation (no customer data needed)
func (s *SalesService) applyTaxesToSale(saleID string, saleItems []models.SaleItem, warehouseID string) (*models.TaxSummary, error) {
	// Convert sale items to tax calculation items
	var taxItems []models.TaxCalculationItem
	for _, item := range saleItems {
		// Get batch to retrieve product ID
		batch, err := s.inventoryRepo.GetBatchByID(item.BatchID)
		if err != nil {
			return nil, err
		}

		taxItem := models.TaxCalculationItem{
			ProductID:  batch.VariantID,
			CategoryID: nil, // No category management in current model
			Quantity:   int(item.Quantity),
			UnitPrice:  item.SellingPrice,
			LineTotal:  item.LineTotal,
		}
		taxItems = append(taxItems, taxItem)
	}

	// Default state since warehouse doesn't store state directly (can be configurable)
	defaultState := "DefaultState" // This should be configurable or fetched from address service

	// Create tax calculation request without customer information
	taxReq := &models.TaxCalculationRequest{
		CustomerID:     nil, // No customer management
		CustomerState:  nil, // No customer management
		CustomerGSTIN:  nil, // No customer GSTIN
		CustomerPAN:    nil, // No customer PAN
		WarehouseID:    warehouseID,
		WarehouseState: defaultState, // Use default state for warehouse
		Items:          taxItems,
		IsInterState:   false, // Default to intra-state for warehouse-based taxation
	}

	// Use the existing tax service to apply taxes
	return s.taxService.ApplyTaxesToSale(saleID, saleItems, taxReq, "system")
}

// calculateDiscountAmount calculates discount amount based on discount type
func (s *SalesService) calculateDiscountAmount(discount *models.Discount, orderValue float64, saleItems []*models.SaleItem) float64 {
	if discount.DiscountType == models.DiscountTypeBuyXGetY {
		// Convert SaleItem to repository SaleItem format
		var repoItems []repositories.SaleItem
		for _, item := range saleItems {
			// Get batch to extract product ID
			batch, err := s.inventoryRepo.GetBatchByID(item.BatchID)
			productID := item.BatchID // fallback to batch ID if product ID can't be retrieved
			if err == nil && batch != nil {
				productID = batch.VariantID
			}

			repoItems = append(repoItems, repositories.SaleItem{
				ProductID: productID,
				Quantity:  item.Quantity,
				Price:     item.SellingPrice,
			})
		}
		return s.discountsRepo.CalculateBuyXGetYDiscount(*discount, repoItems)
	}

	// For other discount types, use the regular calculation
	return s.discountsRepo.CalculateDiscount(discount, orderValue)
}

// CanCancelSale determines if a sale can be cancelled based on its status
func (s *SalesService) CanCancelSale(sale *models.Sale) (bool, string) {
	switch sale.Status {
	case "cancelled":
		return false, "Sale is already cancelled"
	case "shipped", "delivered":
		return false, "Cannot cancel shipped/delivered orders. Use Returns instead."
	case "returned":
		return false, "Sale has already been returned"
	case "pending", "confirmed", "processing":
		return true, ""
	default:
		return false, "Unknown sale status"
	}
}

// CancelSale cancels a full sale and returns inventory to original batches
func (s *SalesService) CancelSale(saleID string, req *models.CancelSaleRequest) (*models.CancelSaleResponse, error) {
	s.logger.Info("Starting sale cancellation",
		zap.String("sale_id", saleID),
		zap.String("reason", req.Reason),
		zap.String("performed_by", req.PerformedBy))

	// Validate reason - if "other", reason_details is required
	if req.Reason == models.ReasonOther && (req.ReasonDetails == nil || *req.ReasonDetails == "") {
		s.logger.Error("Reason details required for 'other' reason",
			zap.String("sale_id", saleID))
		return nil, errors.NewValidationError("reason_details is required when reason is 'other'")
	}

	var response *models.CancelSaleResponse

	// Execute everything within a database transaction
	err := s.salesRepo.WithTransaction(func(tx *gorm.DB) error {
		// 1. Lock sale record (SELECT FOR UPDATE)
		s.logger.Debug("Locking sale record for update",
			zap.String("sale_id", saleID))
		sale, err := s.salesRepo.GetSaleForUpdateWithTx(tx, saleID)
		if err != nil {
			s.logger.Error("Failed to get sale for update",
				zap.Error(err),
				zap.String("sale_id", saleID))
			if err == gorm.ErrRecordNotFound {
				return errors.NewNotFoundError("Sale")
			}
			return errors.NewInternalServerError("Failed to lock sale")
		}

		// 2. Check cancellability
		s.logger.Debug("Checking if sale can be cancelled",
			zap.String("sale_id", saleID),
			zap.String("status", sale.Status))
		canCancel, reason := s.CanCancelSale(sale)
		if !canCancel {
			s.logger.Error("Sale cannot be cancelled",
				zap.String("sale_id", saleID),
				zap.String("reason", reason))
			return errors.NewBadRequestError(reason)
		}

		// 3. Create SaleCancellation record
		s.logger.Debug("Creating cancellation record",
			zap.String("sale_id", saleID))
		cancellation := models.NewSaleCancellation(
			sale.ID,
			models.CancellationTypeFull,
			req.Reason,
			&req.PerformedBy,
			req.ReasonDetails,
			sale.TotalAmount,
			sale.TotalAmount, // Full cancellation
		)

		if err := s.saleCancellationRepo.CreateCancellationWithTx(tx, cancellation); err != nil {
			s.logger.Error("Failed to create cancellation record",
				zap.Error(err),
				zap.String("sale_id", saleID))
			return errors.NewInternalServerError("Failed to create cancellation record")
		}
		s.logger.Info("Cancellation record created",
			zap.String("cancellation_id", cancellation.ID))

		// 4. For each SaleItem: restore inventory
		s.logger.Debug("Processing sale items for inventory restoration",
			zap.Int("item_count", len(sale.Items)))
		var inventoryRestored []models.InventoryRestoredItem

		for i, saleItem := range sale.Items {
			s.logger.Debug("Processing sale item",
				zap.Int("item_number", i+1),
				zap.String("sale_item_id", saleItem.ID),
				zap.String("batch_id", saleItem.BatchID),
				zap.Int64("quantity", saleItem.Quantity))

			// Get batch to retrieve variant ID for response
			batch, err := s.inventoryRepo.GetBatchByID(saleItem.BatchID)
			if err != nil {
				s.logger.Error("Failed to get batch",
					zap.Error(err),
					zap.String("batch_id", saleItem.BatchID))
				return errors.NewInternalServerError("Failed to get batch information")
			}

			// Restore inventory: Create inventory transaction
			note := "Inventory restored for cancelled sale " + sale.ID
			transaction := models.NewInventoryTransaction(
				saleItem.BatchID,
				"cancellation_return",
				saleItem.Quantity, // Positive to add back
				&cancellation.ID,
				&req.PerformedBy,
				&note,
				time.Now(),
			)

			if err := s.inventoryRepo.CreateTransactionWithTx(tx, transaction); err != nil {
				s.logger.Error("Failed to create inventory transaction",
					zap.Error(err),
					zap.String("batch_id", saleItem.BatchID))
				return errors.NewInternalServerError("Failed to restore inventory")
			}
			s.logger.Debug("Inventory transaction created",
				zap.String("transaction_id", transaction.ID))

			// Update batch stock level (add back the quantity)
			if !req.SkipInventoryReturn {
				if err := s.inventoryRepo.UpdateBatchStockWithTx(tx, saleItem.BatchID, saleItem.Quantity); err != nil {
					s.logger.Error("Failed to update batch stock",
						zap.Error(err),
						zap.String("batch_id", saleItem.BatchID))
					return errors.NewInternalServerError("Failed to restore batch stock")
				}
				s.logger.Info("Inventory restored to batch",
					zap.String("batch_id", saleItem.BatchID),
					zap.Int64("quantity_restored", saleItem.Quantity))
			}

			// Create SaleCancellationItem record
			cancellationItem := models.NewSaleCancellationItem(
				cancellation.ID,
				saleItem.ID,
				saleItem.BatchID,
				saleItem.Quantity,
				saleItem.LineTotal,
				&transaction.ID,
			)

			if err := s.saleCancellationRepo.CreateCancellationItemWithTx(tx, cancellationItem); err != nil {
				s.logger.Error("Failed to create cancellation item",
					zap.Error(err),
					zap.String("sale_item_id", saleItem.ID))
				return errors.NewInternalServerError("Failed to create cancellation item")
			}

			// Add to restored items list
			inventoryRestored = append(inventoryRestored, models.InventoryRestoredItem{
				BatchID:          saleItem.BatchID,
				VariantID:        batch.VariantID,
				QuantityRestored: saleItem.Quantity,
				TransactionID:    transaction.ID,
			})
		}

		// 5. Reverse discounts (if any were applied)
		// TODO: Implement discount reversal logic
		s.logger.Debug("Discount reversal not yet implemented - skipping")

		// 6. Void tax records (if taxes were applied)
		// TODO: Implement tax voiding logic
		s.logger.Debug("Tax voiding not yet implemented - skipping")

		// 7. Update Sale status to "cancelled"
		s.logger.Debug("Updating sale status to cancelled",
			zap.String("sale_id", sale.ID))
		sale.Status = "cancelled"
		now := time.Now()
		sale.CancelledAt = &now
		sale.CancellationReason = &req.Reason

		if err := s.salesRepo.UpdateSaleWithTx(tx, sale); err != nil {
			s.logger.Error("Failed to update sale status",
				zap.Error(err),
				zap.String("sale_id", sale.ID))
			return errors.NewInternalServerError("Failed to update sale status")
		}

		// Build response
		response = &models.CancelSaleResponse{
			Sale:              *s.mapSaleToResponse(sale),
			InventoryRestored: inventoryRestored,
			CancellationID:    cancellation.ID,
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Sale cancellation transaction failed",
			zap.Error(err),
			zap.String("sale_id", saleID))
		return nil, err
	}

	s.logger.Info("Sale cancellation completed successfully",
		zap.String("sale_id", saleID),
		zap.String("cancellation_id", response.CancellationID))

	return response, nil
}
