package services

import (
	"errors"
	"log"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"

	"gorm.io/gorm"
)

// Helper function to convert string to pointer
func stringPtr(s string) *string {
	return &s
}

type SalesService struct {
	salesRepo     *repositories.SalesRepository
	productRepo   *repositories.ProductRepository
	inventoryRepo *repositories.InventoryRepository
	priceRepo     *repositories.ProductPriceRepository
	discountsRepo *repositories.DiscountsRepository
	taxRepo       *repositories.TaxRepository
	taxService    *TaxService
	warehouseRepo *repositories.WarehouseRepository
}

func NewSalesService(salesRepo *repositories.SalesRepository, productRepo *repositories.ProductRepository, inventoryRepo *repositories.InventoryRepository, priceRepo *repositories.ProductPriceRepository, discountsRepo *repositories.DiscountsRepository, taxRepo *repositories.TaxRepository, warehouseRepo *repositories.WarehouseRepository) *SalesService {
	return &SalesService{
		salesRepo:     salesRepo,
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
		priceRepo:     priceRepo,
		discountsRepo: discountsRepo,
		taxRepo:       taxRepo,
		taxService:    NewTaxService(taxRepo),
		warehouseRepo: warehouseRepo,
	}
}

// CreateSale creates a new sale with items and summary using database transaction
func (s *SalesService) CreateSale(req *models.CreateSaleRequest) (*models.SaleResponse, error) {
	log.Printf("[DEBUG] Starting transactional sale creation for warehouse: %s", req.WarehouseID)

	// Validate sale request
	if err := s.validateSaleRequest(req); err != nil {
		log.Printf("[ERROR] Sale validation failed: %v", err)
		return nil, err
	}
	log.Printf("[DEBUG] Sale validation passed")

	var response *models.SaleResponse

	// Execute everything within a database transaction
	err := s.salesRepo.WithTransaction(func(tx *gorm.DB) error {
		// Parse sale date or use current time
		var saleDate time.Time
		if req.SaleDate != nil {
			log.Printf("[DEBUG] Parsing sale date: %s", *req.SaleDate)
			if parsedDate, err := time.Parse(time.RFC3339, *req.SaleDate); err == nil {
				saleDate = parsedDate
				log.Printf("[DEBUG] Sale date parsed successfully: %v", saleDate)
			} else {
				// If parsing fails, use current time
				log.Printf("[WARN] Date parsing failed, using current time: %v", err)
				saleDate = time.Now()
			}
		} else {
			// If no sale date provided, use current time
			log.Printf("[DEBUG] No sale date provided, using current time")
			saleDate = time.Now()
		}

		// Create sale using the proper constructor
		log.Printf("[DEBUG] Creating sale with warehouse: %s, date: %v", req.WarehouseID, saleDate)
		sale := models.NewSale(req.WarehouseID, saleDate, 0, "pending")
		log.Printf("[DEBUG] Sale created with ID: %s", sale.ID)

		if err := s.salesRepo.CreateSaleWithTx(tx, sale); err != nil {
			log.Printf("[ERROR] Failed to create sale in database: %v", err)
			return err
		}
		log.Printf("[DEBUG] Sale created successfully in database")

	// Create sale items
	log.Printf("[DEBUG] Starting to process %d sale items", len(req.Items))
	var totalAmount float64
	var saleItems []models.SaleItem // Collect sale items for tax calculation
	for i, itemReq := range req.Items {
		log.Printf("[DEBUG] Processing item %d: product=%s, qty=%d", i+1, itemReq.ProductID, itemReq.Quantity)

		// Get selling price from product_prices table
		log.Printf("[DEBUG] Getting selling price for product: %s", itemReq.ProductID)
		sellingPrice, err := s.getSellingPrice(itemReq.ProductID)
		if err != nil {
			log.Printf("[ERROR] Failed to get selling price: %v", err)
			return errors.New("selling price not found for product")
		}
		log.Printf("[DEBUG] Selling price retrieved: %.2f", sellingPrice)

		// Get batches for this product in the warehouse ordered by expiry date (FEFO)
		log.Printf("[DEBUG] Getting batches for product: %s in warehouse: %s", itemReq.ProductID, req.WarehouseID)
		batches, err := s.inventoryRepo.GetBatchesByProductAndWarehouseOrderedByExpiry(itemReq.ProductID, req.WarehouseID)
		if err != nil {
			log.Printf("[ERROR] Failed to get batches for product: %v", err)
			return errors.New("failed to retrieve product batches")
		}

		if len(batches) == 0 {
			log.Printf("[ERROR] No batches found for product: %s in warehouse: %s", itemReq.ProductID, req.WarehouseID)
			return errors.New("no inventory available for product in this warehouse")
		}

		log.Printf("[DEBUG] Found %d batches for product", len(batches))

		// Calculate total available quantity across all batches
		totalAvailable := int64(0)
		for _, batch := range batches {
			totalAvailable += batch.TotalQuantity
		}

		if totalAvailable < itemReq.Quantity {
			log.Printf("[ERROR] Insufficient stock: available %d, requested %d", totalAvailable, itemReq.Quantity)
			return errors.New("insufficient stock for product")
		}

		log.Printf("[DEBUG] Stock validation passed - available: %d, requested: %d", totalAvailable, itemReq.Quantity)

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

			log.Printf("[DEBUG] Allocating %d units from batch %s (expires: %s)", quantityFromBatch, batch.ID, batch.ExpiryDate.Format("2006-01-02"))

			// Calculate line total for this batch allocation
			batchLineTotal := sellingPrice * float64(quantityFromBatch)

			// Calculate taxes for this batch allocation
			log.Printf("[DEBUG] Calculating taxes for batch allocation")
			taxCalculation, err := s.taxService.CalculateBatchTax(batch, quantityFromBatch, sellingPrice)
			if err != nil {
				log.Printf("[ERROR] Failed to calculate taxes: %v", err)
				return err
			}
			log.Printf("[DEBUG] Tax calculation completed: CGST=%.2f, SGST=%.2f, Custom=%.2f, Total=%.2f",
				taxCalculation.CGSTAmount, taxCalculation.SGSTAmount, taxCalculation.CustomTaxAmount, taxCalculation.TotalTaxAmount)

			itemTotal += batchLineTotal

			// Create sale item with tax amounts
			saleItem := models.NewSaleItemWithTax(sale.ID, batch.ID, quantityFromBatch, sellingPrice, batchLineTotal,
				taxCalculation.CGSTAmount, taxCalculation.SGSTAmount, taxCalculation.CustomTaxAmount)
			log.Printf("[DEBUG] Sale item created with ID: %s (includes tax amounts)", saleItem.ID)

			if err := s.salesRepo.CreateSaleItemWithTx(tx, saleItem); err != nil {
				log.Printf("[ERROR] Failed to create sale item: %v", err)
				return err
			}
			log.Printf("[DEBUG] Sale item created successfully in database")

			// Add to collection for tax calculation
			saleItems = append(saleItems, *saleItem)

			// Update inventory using the proper constructor
			log.Printf("[DEBUG] Creating inventory transaction for batch: %s", batch.ID)
			transaction := models.NewInventoryTransaction(batch.ID, "sale", -quantityFromBatch, &sale.ID, nil, stringPtr("Sale transaction"), time.Now())
			log.Printf("[DEBUG] Inventory transaction created with ID: %s", transaction.ID)

			if err := s.inventoryRepo.CreateTransactionWithTx(tx, transaction); err != nil {
				log.Printf("[ERROR] Failed to create inventory transaction: %v", err)
				return err
			}
			log.Printf("[DEBUG] Inventory transaction created successfully")

			// Update batch stock level with row lock to prevent race conditions
			log.Printf("[DEBUG] Updating batch stock: %s, change: %d", batch.ID, -quantityFromBatch)
			if err := s.inventoryRepo.UpdateBatchStockWithTx(tx, batch.ID, -quantityFromBatch); err != nil {
				log.Printf("[ERROR] Failed to update batch stock: %v", err)
				return err
			}
			log.Printf("[DEBUG] Batch stock updated successfully")

			remainingQuantity -= quantityFromBatch
		}

		totalAmount += itemTotal
		log.Printf("[DEBUG] Running total: %.2f", totalAmount)
	}

	// Collect product IDs for discount discovery
	var productIDs []string
	for _, item := range saleItems {
		batch, err := s.inventoryRepo.GetBatchByID(item.BatchID)
		if err == nil {
			productIDs = append(productIDs, batch.ProductID)
		}
	}

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
		log.Printf("[ERROR] Failed to resolve discounts: %v", err)
		return err
	}

	discountAmount = totalDiscountAmount
	appliedDiscounts = applications

		// Create discount usage records for applied discounts
		for _, discount := range finalDiscounts {
			discountUsage := s.discountsRepo.CalculateDiscount(&discount, totalAmount)
			usage := models.NewDiscountUsage(discount.ID, sale.ID, discountUsage)
			if err := s.discountsRepo.CreateDiscountUsageWithTx(tx, usage); err != nil {
				log.Printf("[ERROR] Failed to create discount usage record: %v", err)
				return err
			}
			if err := s.discountsRepo.IncrementUsageWithTx(tx, discount.ID); err != nil {
				log.Printf("[ERROR] Failed to increment discount usage: %v", err)
				return err
			}
		}

	log.Printf("[DEBUG] Total discount applied: %.2f from %d discounts", discountAmount, len(finalDiscounts))

	// Calculate final amount after discount
	finalAmount := totalAmount - discountAmount
	log.Printf("[DEBUG] Final amount after discount: %.2f", finalAmount)

		// Apply taxes using the existing tax service (no customer data needed)
		var taxAmount float64
		if len(saleItems) > 0 {
			taxSummary, err := s.applyTaxesToSaleWithTx(tx, sale.ID, saleItems, req.WarehouseID)
			if err != nil {
				log.Printf("[ERROR] Tax calculation failed: %v", err)
				return err
			}
			if taxSummary != nil {
				taxAmount = taxSummary.TotalTaxAmount
				finalAmount += taxAmount
				log.Printf("[DEBUG] Tax applied successfully: %.2f", taxAmount)
			}
		}

		// Update sale with final amount
		log.Printf("[DEBUG] Updating sale with final amount: %.2f", finalAmount)
		sale.TotalAmount = finalAmount
		if err := s.salesRepo.UpdateSaleWithTx(tx, sale); err != nil {
			log.Printf("[ERROR] Failed to update sale with final amount: %v", err)
			return err
		}
		log.Printf("[DEBUG] Sale updated with final amount successfully")

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
		log.Printf("[ERROR] Transaction failed: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Transactional sale creation completed successfully")
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

// getSellingPrice retrieves the current retail price for a product
func (s *SalesService) getSellingPrice(productID string) (float64, error) {
	// Get the active retail price for the product
	price, err := s.priceRepo.GetCurrentPrice(productID, "retail")
	if err != nil {
		// Try to get any active price if retail price is not found
		prices, err2 := s.priceRepo.GetActiveByProductID(productID)
		if err2 != nil {
			return 0, errors.New("no pricing information found for product")
		}
		if len(prices) == 0 {
			return 0, errors.New("no active prices found for product")
		}
		// Use the first active price as fallback
		return prices[0].Price, nil
	}
	return price.Price, nil
}

// Helper methods
func (s *SalesService) validateSaleRequest(req *models.CreateSaleRequest) error {
	log.Printf("[DEBUG] Validating sale request: warehouse=%s, items=%d", req.WarehouseID, len(req.Items))

	if req.WarehouseID == "" {
		log.Printf("[ERROR] Validation failed: warehouse ID is empty")
		return errors.New("warehouse ID is required")
	}
	if len(req.Items) == 0 {
		log.Printf("[ERROR] Validation failed: no items provided")
		return errors.New("at least one item is required")
	}

	for i, item := range req.Items {
		log.Printf("[DEBUG] Validating item %d: product=%s, qty=%d", i+1, item.ProductID, item.Quantity)

		if item.ProductID == "" {
			log.Printf("[ERROR] Validation failed: product ID is empty for item %d", i+1)
			return errors.New("product ID is required for all items")
		}
		if item.Quantity <= 0 {
			log.Printf("[ERROR] Validation failed: quantity <= 0 for item %d", i+1)
			return errors.New("quantity must be greater than 0")
		}
		// Remove selling price validation since it will be calculated automatically from product_prices table
	}

	log.Printf("[DEBUG] Sale request validation passed")
	return nil
}

func (s *SalesService) mapSaleToResponse(sale *models.Sale) *models.SaleResponse {
	response := &models.SaleResponse{
		ID:          sale.ID,
		WarehouseID: sale.WarehouseID,
		SaleDate:    sale.SaleDate.Format("2006-01-02T15:04:05Z07:00"),
		TotalAmount: sale.TotalAmount,
		Status:      sale.Status,
		CreatedAt:   sale.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   sale.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Map items
	for _, item := range sale.Items {
		response.Items = append(response.Items, models.SaleItemResponse{
			ID:              item.ID,
			SaleID:          item.SaleID,
			BatchID:         item.BatchID,
			Quantity:        item.Quantity,
			SellingPrice:    item.SellingPrice,
			LineTotal:       item.LineTotal,
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
		return 0, errors.New("discount is not active")
	}

	// Check usage limits
	if discount.UsageLimit != nil && discount.CurrentUsage >= *discount.UsageLimit {
		return 0, errors.New("discount usage limit reached")
	}

	// Check date validity
	now := time.Now()
	if now.Before(discount.ValidFrom) || now.After(discount.ValidUntil) {
		return 0, errors.New("discount is not valid for the current date")
	}

	// Check minimum order value
	if discount.MinOrderValue != nil && orderValue < *discount.MinOrderValue {
		return 0, errors.New("order value does not meet minimum requirement")
	}

	// Check maximum order value
	if discount.MaxOrderValue != nil && orderValue > *discount.MaxOrderValue {
		return 0, errors.New("order value exceeds maximum limit")
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
			log.Printf("[WARN] Failed to discover automatic discounts: %v", err)
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
			ProductID:  batch.ProductID,
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
			ProductID:  batch.ProductID,
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
				productID = batch.ProductID
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
