package services

import (
	"errors"
	"log"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
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
	discountsRepo *repositories.DiscountsRepository // Add discount repository
	taxRepo       *repositories.TaxRepository       // Add tax repository
	taxService    *TaxService                       // Add tax service
}

func NewSalesService(salesRepo *repositories.SalesRepository, productRepo *repositories.ProductRepository, inventoryRepo *repositories.InventoryRepository, priceRepo *repositories.ProductPriceRepository, discountsRepo *repositories.DiscountsRepository, taxRepo *repositories.TaxRepository) *SalesService {
	return &SalesService{
		salesRepo:     salesRepo,
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
		priceRepo:     priceRepo,
		discountsRepo: discountsRepo,          // Add discount repository
		taxRepo:       taxRepo,                // Add tax repository
		taxService:    NewTaxService(taxRepo), // Initialize tax service
	}
}

// CreateSale creates a new sale with items and summary
func (s *SalesService) CreateSale(req *models.CreateSaleRequest) (*models.SaleResponse, error) {
	log.Printf("[DEBUG] Starting sale creation for warehouse: %s", req.WarehouseID)

	// Validate sale request
	if err := s.validateSaleRequest(req); err != nil {
		log.Printf("[ERROR] Sale validation failed: %v", err)
		return nil, err
	}
	log.Printf("[DEBUG] Sale validation passed")

	// Parse sale date or use current time
	var saleDate time.Time
	if req.SaleDate != nil {
		log.Printf("[DEBUG] Parsing sale date: %s", *req.SaleDate)
		if parsedDate, err := time.Parse("2006-01-02T15:04:05Z", *req.SaleDate); err == nil {
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
	log.Printf("[DEBUG] Creating sale with warehouse: %s, customer: %v, date: %v", req.WarehouseID, req.CustomerID, saleDate)
	sale := models.NewSale(req.WarehouseID, req.CustomerID, saleDate, 0, "pending")
	log.Printf("[DEBUG] Sale created with ID: %s", sale.ID)

	if err := s.salesRepo.CreateSale(sale); err != nil {
		log.Printf("[ERROR] Failed to create sale in database: %v", err)
		return nil, err
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
			return nil, errors.New("selling price not found for product")
		}
		log.Printf("[DEBUG] Selling price retrieved: %.2f", sellingPrice)

		// Get batches for this product in the warehouse ordered by expiry date (FEFO)
		log.Printf("[DEBUG] Getting batches for product: %s in warehouse: %s", itemReq.ProductID, req.WarehouseID)
		batches, err := s.inventoryRepo.GetBatchesByProductAndWarehouseOrderedByExpiry(itemReq.ProductID, req.WarehouseID)
		if err != nil {
			log.Printf("[ERROR] Failed to get batches for product: %v", err)
			return nil, errors.New("failed to retrieve product batches")
		}

		if len(batches) == 0 {
			log.Printf("[ERROR] No batches found for product: %s in warehouse: %s", itemReq.ProductID, req.WarehouseID)
			return nil, errors.New("no inventory available for product in this warehouse")
		}

		log.Printf("[DEBUG] Found %d batches for product", len(batches))

		// Calculate total available quantity across all batches
		totalAvailable := int64(0)
		for _, batch := range batches {
			totalAvailable += batch.TotalQuantity
		}

		if totalAvailable < itemReq.Quantity {
			log.Printf("[ERROR] Insufficient stock: available %d, requested %d", totalAvailable, itemReq.Quantity)
			return nil, errors.New("insufficient stock for product")
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
			itemTotal += batchLineTotal

			// Create sale item for this batch allocation
			saleItem := models.NewSaleItem(sale.ID, batch.ID, quantityFromBatch, sellingPrice, batchLineTotal)
			log.Printf("[DEBUG] Sale item created with ID: %s", saleItem.ID)

			if err := s.salesRepo.CreateSaleItem(saleItem); err != nil {
				log.Printf("[ERROR] Failed to create sale item: %v", err)
				return nil, err
			}
			log.Printf("[DEBUG] Sale item created successfully in database")

			// Add to collection for tax calculation
			saleItems = append(saleItems, *saleItem)

			// Update inventory using the proper constructor
			log.Printf("[DEBUG] Creating inventory transaction for batch: %s", batch.ID)
			transaction := models.NewInventoryTransaction(batch.ID, "sale", -quantityFromBatch, &sale.ID, nil, stringPtr("Sale transaction"), time.Now())
			log.Printf("[DEBUG] Inventory transaction created with ID: %s", transaction.ID)

			if err := s.inventoryRepo.CreateTransaction(transaction); err != nil {
				log.Printf("[ERROR] Failed to create inventory transaction: %v", err)
				return nil, err
			}
			log.Printf("[DEBUG] Inventory transaction created successfully")

			// Update batch stock level
			log.Printf("[DEBUG] Updating batch stock: %s, change: %d", batch.ID, -quantityFromBatch)
			if err := s.inventoryRepo.UpdateBatchStock(batch.ID, -quantityFromBatch); err != nil {
				log.Printf("[ERROR] Failed to update batch stock: %v", err)
				return nil, err
			}
			log.Printf("[DEBUG] Batch stock updated successfully")

			remainingQuantity -= quantityFromBatch
		}

		totalAmount += itemTotal
		log.Printf("[DEBUG] Running total: %.2f", totalAmount)
	}

	// Apply discount if provided
	var discountAmount float64
	if req.DiscountID != nil && *req.DiscountID != "" {
		log.Printf("[DEBUG] Applying discount: %s", *req.DiscountID)
		var err error
		discountAmount, err = s.applyDiscountToSale(*req.DiscountID, req.CustomerID, sale.ID, totalAmount)
		if err != nil {
			log.Printf("[ERROR] Failed to apply discount: %v", err)
			return nil, err
		}
		log.Printf("[DEBUG] Discount applied: %.2f", discountAmount)
	}

	// Calculate final amount after discount
	finalAmount := totalAmount - discountAmount
	log.Printf("[DEBUG] Final amount after discount: %.2f", finalAmount)

	// Calculate and apply taxes
	var taxAmount float64
	if req.CustomerState != nil && req.WarehouseState != nil {
		var err error
		taxAmount, _, err = s.applyTaxesToSale(sale.ID, saleItems, req)
		if err != nil {
			return nil, err
		}
		finalAmount += taxAmount
		log.Printf("[DEBUG] Final amount after tax: %.2f", finalAmount)
	}

	// Update sale with final amount
	log.Printf("[DEBUG] Updating sale with final amount: %.2f", finalAmount)
	sale.TotalAmount = finalAmount
	if err := s.salesRepo.UpdateSale(sale); err != nil {
		log.Printf("[ERROR] Failed to update sale with final amount: %v", err)
		return nil, err
	}
	log.Printf("[DEBUG] Sale updated with final amount successfully")

	// Get complete sale with items and summary
	log.Printf("[DEBUG] Getting complete sale by ID: %s", sale.ID)
	completeSale, err := s.salesRepo.GetSaleByID(sale.ID)
	if err != nil {
		log.Printf("[ERROR] Failed to get complete sale: %v", err)
		return nil, err
	}
	log.Printf("[DEBUG] Complete sale retrieved successfully")

	log.Printf("[DEBUG] Sale creation completed successfully")
	return s.mapSaleToResponse(completeSale), nil
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

// GetSalesByCustomer retrieves sales for a specific customer
func (s *SalesService) GetSalesByCustomer(customerID string) ([]models.SaleResponse, error) {
	sales, err := s.salesRepo.GetSalesByCustomer(customerID)
	if err != nil {
		return nil, err
	}

	var responses []models.SaleResponse
	for _, sale := range sales {
		responses = append(responses, *s.mapSaleToResponse(&sale))
	}

	return responses, nil
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
		CustomerID:  sale.CustomerID,
		SaleDate:    sale.SaleDate.Format("2006-01-02T15:04:05Z07:00"),
		TotalAmount: sale.TotalAmount,
		Status:      sale.Status,
		CreatedAt:   sale.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   sale.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
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
			CreatedAt:    item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return response
}

// applyTaxesToSale applies taxes to a sale and returns the tax amount and summary
func (s *SalesService) applyTaxesToSale(saleID string, saleItems []models.SaleItem, req *models.CreateSaleRequest) (float64, *models.TaxSummary, error) {
	// Convert sale items to tax calculation items
	var taxItems []models.TaxCalculationItem
	for _, item := range saleItems {
		// Get batch to retrieve product ID
		batch, err := s.inventoryRepo.GetBatchByID(item.BatchID)
		if err != nil {
			return 0, nil, err
		}

		taxItem := models.TaxCalculationItem{
			ProductID: batch.ProductID,
			Quantity:  int(item.Quantity),
			UnitPrice: item.SellingPrice,
			LineTotal: item.LineTotal,
		}
		taxItems = append(taxItems, taxItem)
	}

	// Create tax calculation request
	taxReq := &models.TaxCalculationRequest{
		CustomerID:     *req.CustomerID,
		CustomerState:  *req.CustomerState,
		CustomerGSTIN:  req.CustomerGSTIN,
		CustomerPAN:    req.CustomerPAN,
		WarehouseID:    req.WarehouseID,
		WarehouseState: *req.WarehouseState,
		Items:          taxItems,
		IsInterState:   req.IsInterState != nil && *req.IsInterState,
	}

	// Calculate taxes
	taxCalculation, err := s.taxService.CalculateTax(taxReq)
	if err != nil {
		return 0, nil, err
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
		return 0, nil, err
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
			return 0, nil, err
		}
	}

	return taxCalculation.TotalTaxAmount, taxSummary, nil
}

// applyDiscountToSale applies a discount to a sale and returns the discount amount
func (s *SalesService) applyDiscountToSale(discountID string, customerID *string, saleID string, orderValue float64) (float64, error) {
	// Get discount by ID
	discount, err := s.discountsRepo.GetDiscountByID(discountID)
	if err != nil {
		return 0, err
	}

	// Validate discount for this order
	_, err = s.discountsRepo.ValidateDiscount(discount.Code, customerID, orderValue, nil, "") // TODO: Add product IDs and warehouse ID
	if err != nil {
		return 0, err
	}

	// Calculate discount amount
	discountAmount := s.discountsRepo.CalculateDiscount(discount, orderValue)

	// Create discount usage record
	usage := models.NewDiscountUsage(discountID, *customerID, saleID, discountAmount)
	if err := s.discountsRepo.CreateDiscountUsage(usage); err != nil {
		return 0, err
	}

	// Increment usage count
	if err := s.discountsRepo.IncrementUsage(discountID); err != nil {
		return 0, err
	}

	return discountAmount, nil
}
