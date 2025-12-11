package services

import (
	"fmt"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Helper function to convert string to pointer
func stringPtr(s string) *string {
	return &s
}

// generateInvoiceNumber generates an invoice number in the format MMYYNNNN
// - MM: 2-digit month (01-12)
// - YY: 2-digit year (e.g., 25 for 2025)
// - NNNN: 4-digit sequence number (does NOT reset each month)
// Example: 12250001 = December 2025, sequence 1
func generateInvoiceNumber(lastSequence int) string {
	now := time.Now()
	month := now.Format("01") // MM
	year := now.Format("06")  // YY
	sequence := lastSequence + 1
	return fmt.Sprintf("%s%s%04d", month, year, sequence)
}

type SalesService struct {
	salesRepo            *repositories.SalesRepository
	productRepo          *repositories.ProductRepository
	inventoryRepo        *repositories.InventoryRepository
	variantRepo          *repositories.ProductVariantRepository
	priceRepo            *repositories.ProductPriceRepository // Prices from product_prices table
	discountsRepo        *repositories.DiscountsRepository
	taxRepo              *repositories.TaxRepository
	taxService           *TaxService
	warehouseRepo        *repositories.WarehouseRepository
	saleCancellationRepo *repositories.SaleCancellationRepository // Order cancellation feature
	logger               interfaces.Logger
}

func NewSalesService(salesRepo *repositories.SalesRepository, productRepo *repositories.ProductRepository, inventoryRepo *repositories.InventoryRepository, variantRepo *repositories.ProductVariantRepository, priceRepo *repositories.ProductPriceRepository, discountsRepo *repositories.DiscountsRepository, taxRepo *repositories.TaxRepository, warehouseRepo *repositories.WarehouseRepository, saleCancellationRepo *repositories.SaleCancellationRepository, logger interfaces.Logger) *SalesService {
	return &SalesService{
		salesRepo:            salesRepo,
		productRepo:          productRepo,
		inventoryRepo:        inventoryRepo,
		variantRepo:          variantRepo,
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
		variant      *models.ProductVariant // For GST rate
	}

	itemDataMap := make(map[string]*itemData)
	for _, itemReq := range req.Items {
		// Get variant for GST rate (GST-only tax system)
		s.logger.Debug("Getting variant for GST rate",
			zap.String("variant_id", itemReq.VariantID))
		variant, err := s.variantRepo.GetByID(itemReq.VariantID)
		if err != nil {
			s.logger.Error("Failed to get variant",
				zap.Error(err),
				zap.String("variant_id", itemReq.VariantID))
			return nil, errors.NewNotFoundError("variant not found")
		}

		// Get selling price from product_prices table (by variant_id)
		// Price selection depends on membership status: member price for members, retail for non-members
		s.logger.Debug("Getting selling price for variant",
			zap.String("variant_id", itemReq.VariantID),
			zap.Bool("is_org_member", req.IsOrgMember))
		sellingPrice, err := s.getSellingPrice(itemReq.VariantID, req.IsOrgMember)
		if err != nil {
			s.logger.Error("Failed to get selling price",
				zap.Error(err),
				zap.String("variant_id", itemReq.VariantID))
			return nil, errors.NewNotFoundError("selling price not found for product")
		}
		s.logger.Debug("Selling price retrieved",
			zap.Float64("selling_price", sellingPrice),
			zap.Bool("is_org_member", req.IsOrgMember))

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
		// Available = TotalQuantity - ReservedQuantity (reservation system)
		// NOTE: This pre-transaction check is an optimization for early failure detection.
		// The real race-condition-safe check happens in ReserveBatchStockWithTx (atomic conditional update).
		// Concurrent sales may pass this check but the atomic reservation will fail safely.
		totalAvailable := int64(0)
		for _, batch := range batches {
			totalAvailable += batch.AvailableQuantity()
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
			variant:      variant,
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

		// Handle ApplyTaxes - default to true if not provided (Issue 6: Breaking Change)
		applyTaxes := true
		if req.ApplyTaxes != nil {
			applyTaxes = *req.ApplyTaxes
		}

		// Generate invoice number within transaction for concurrency safety
		s.logger.Debug("Generating invoice number")
		lastSequence, err := s.salesRepo.GetLastInvoiceSequenceWithTx(tx)
		if err != nil {
			s.logger.Error("Failed to get last invoice sequence",
				zap.Error(err))
			return errors.NewInternalServerError("failed to generate invoice number")
		}
		invoiceNumber := generateInvoiceNumber(lastSequence)
		s.logger.Debug("Invoice number generated",
			zap.String("invoice_number", invoiceNumber),
			zap.Int("last_sequence", lastSequence))

		// Create sale using the proper constructor with BRD requirements
		s.logger.Debug("Creating sale",
			zap.String("warehouse_id", req.WarehouseID),
			zap.String("invoice_number", invoiceNumber),
			zap.Time("sale_date", saleDate),
			zap.String("payment_mode", req.PaymentMode),
			zap.String("sale_type", req.SaleType),
			zap.Bool("apply_taxes", applyTaxes),
			zap.Bool("is_org_member", req.IsOrgMember))
		sale := models.NewSale(req.WarehouseID, invoiceNumber, saleDate, 0, models.SaleStatusPending, req.CustomerPhone, req.CustomerName, req.IsOrgMember, req.PaymentMode, req.SaleType, applyTaxes)
		s.logger.Info("Sale created",
			zap.String("sale_id", sale.ID),
			zap.String("invoice_number", sale.InvoiceNumber),
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
			variant := itemData.variant

			// Allocate quantity across batches using FEFO (First Expired, First Out)
			remainingQuantity := itemReq.Quantity
			var itemTotal float64

			for _, batch := range batches {
				if remainingQuantity <= 0 {
					break
				}

				// Calculate how much to take from this batch using available quantity
				availableInBatch := batch.AvailableQuantity()
				if availableInBatch <= 0 {
					continue // Skip batches with no available stock
				}

				quantityFromBatch := remainingQuantity
				if availableInBatch < remainingQuantity {
					quantityFromBatch = availableInBatch
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

				// Calculate GST using variant's GSTRate (GST-only tax system)
				// For now, assume intra-state (CGST+SGST) - inter-state detection can be added later
				// TODO: Implement inter-state detection via warehouse state vs delivery state
				isInterState := false
				s.logger.Debug("Calculating GST for batch allocation",
					zap.Float64("gst_rate", variant.GSTRate),
					zap.Bool("is_inter_state", isInterState))
				taxCalculation := s.taxService.CalculateGST(batchLineTotal, variant.GSTRate, isInterState)
				s.logger.Debug("GST calculation completed",
					zap.Float64("cgst_amount", taxCalculation.CGSTAmount),
					zap.Float64("sgst_amount", taxCalculation.SGSTAmount),
					zap.Float64("igst_amount", taxCalculation.IGSTAmount),
					zap.Float64("total_tax_amount", taxCalculation.TotalTaxAmount))

				itemTotal += batchLineTotal

				// Create sale item with tax amounts and cost price (BRD requirement)
				saleItem := models.NewSaleItemWithTax(sale.ID, batch.ID, quantityFromBatch, sellingPrice, costPrice, batchLineTotal,
					taxCalculation.CGSTAmount, taxCalculation.SGSTAmount, taxCalculation.IGSTAmount)
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

				// Create reservation transaction (not deduction - that happens on CompleteSale)
				// Use positive quantity to indicate amount reserved (negative deduction happens on CompleteSale)
				s.logger.Debug("Creating reservation transaction",
					zap.String("batch_id", batch.ID))
				transaction := models.NewInventoryTransaction(batch.ID, models.TransactionTypeReservation, quantityFromBatch, &sale.ID, nil, stringPtr("Stock reserved for pending sale"), time.Now())
				s.logger.Debug("Reservation transaction created",
					zap.String("transaction_id", transaction.ID))

				if err := s.inventoryRepo.CreateTransactionWithTx(tx, transaction); err != nil {
					s.logger.Error("Failed to create reservation transaction",
						zap.Error(err))
					return err
				}
				s.logger.Debug("Reservation transaction created successfully")

				// Reserve stock instead of deducting (actual deduction happens on CompleteSale)
				s.logger.Debug("Reserving batch stock",
					zap.String("batch_id", batch.ID),
					zap.Int64("quantity_reserved", quantityFromBatch))
				if err := s.inventoryRepo.ReserveBatchStockWithTx(tx, batch.ID, quantityFromBatch); err != nil {
					s.logger.Error("Failed to reserve batch stock",
						zap.Error(err))
					return err
				}
				s.logger.Debug("Batch stock reserved successfully")

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
		// GST-only tax system - no VAT or other taxes
		var taxBreakdown *models.TaxSummaryBreakdown
		if taxAmount > 0 {
			taxSummary, err := s.taxRepo.GetTaxSummaryBySale(sale.ID)
			if err == nil {
				taxBreakdown = &models.TaxSummaryBreakdown{
					CGSTAmount:     taxSummary.CGSTAmount,
					SGSTAmount:     taxSummary.SGSTAmount,
					IGSTAmount:     taxSummary.IGSTAmount,
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

// GetAllSales retrieves all sales with pagination (returns list without items for performance)
// Use GetSale(id) to get full details with items
func (s *SalesService) GetAllSales(limit, offset int) ([]models.SaleListResponse, int64, error) {
	sales, total, err := s.salesRepo.GetAllSales(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var responses []models.SaleListResponse
	for _, sale := range sales {
		responses = append(responses, *s.mapSaleToListResponse(&sale))
	}

	return responses, total, nil
}

// UpdateSale updates a sale with proper inventory handling for status transitions
func (s *SalesService) UpdateSale(id string, req *models.UpdateSaleRequest) (*models.SaleResponse, error) {
	s.logger.Info("Starting sale update",
		zap.String("sale_id", id))

	var response *models.SaleResponse

	err := s.salesRepo.WithTransaction(func(tx *gorm.DB) error {
		// 1. Lock and retrieve sale with items
		sale, err := s.salesRepo.GetSaleForUpdateWithTx(tx, id)
		if err != nil {
			s.logger.Error("Failed to get sale for update",
				zap.Error(err),
				zap.String("sale_id", id))
			if err == gorm.ErrRecordNotFound {
				return errors.NewNotFoundError("Sale")
			}
			return errors.NewInternalServerError("Failed to lock sale")
		}

		// 2. If status change requested, validate and handle inventory
		if req.Status != nil && *req.Status != sale.Status {
			newStatus := *req.Status
			oldStatus := sale.Status

			s.logger.Info("Status change requested",
				zap.String("sale_id", id),
				zap.String("old_status", oldStatus),
				zap.String("new_status", newStatus))

			// Validate the status transition
			if err := models.ValidateStatusTransition(oldStatus, newStatus); err != nil {
				s.logger.Error("Invalid status transition",
					zap.Error(err),
					zap.String("old_status", oldStatus),
					zap.String("new_status", newStatus))
				return errors.NewBadRequestError(err.Error())
			}

			// Get performedBy from request or default to system
			performedBy := req.PerformedBy
			if performedBy == "" {
				performedBy = "system"
			}

			// Handle inventory based on the transition
			switch newStatus {
			case models.SaleStatusCompleted:
				// pending → completed: Convert reservations to actual deductions
				s.logger.Info("Processing completion inventory transition",
					zap.String("sale_id", id))
				if err := s.handleCompletionInventory(tx, sale, performedBy); err != nil {
					return err
				}

			case models.SaleStatusCancelled:
				// pending/completed → cancelled: Release reservations or restore stock
				s.logger.Info("Processing cancellation inventory transition",
					zap.String("sale_id", id),
					zap.String("current_status", oldStatus))
				if err := s.handleCancellationInventory(tx, sale, performedBy); err != nil {
					return err
				}
			}

			sale.Status = newStatus
		}

		// 3. Save the sale
		if err := s.salesRepo.UpdateSaleWithTx(tx, sale); err != nil {
			s.logger.Error("Failed to save sale",
				zap.Error(err),
				zap.String("sale_id", id))
			return errors.NewInternalServerError("Failed to update sale")
		}

		response = s.mapSaleToResponse(sale)
		return nil
	})

	if err != nil {
		s.logger.Error("Sale update failed",
			zap.Error(err),
			zap.String("sale_id", id))
		return nil, err
	}

	s.logger.Info("Sale updated successfully",
		zap.String("sale_id", id))

	return response, nil
}

// DeleteSale deletes a sale
func (s *SalesService) DeleteSale(id string) error {
	return s.salesRepo.DeleteSale(id)
}

// handleCompletionInventory converts reservations to actual deductions when sale status changes to completed
func (s *SalesService) handleCompletionInventory(tx *gorm.DB, sale *models.Sale, performedBy string) error {
	s.logger.Debug("Converting reservations to deductions for sale completion",
		zap.String("sale_id", sale.ID),
		zap.Int("item_count", len(sale.Items)))

	for _, saleItem := range sale.Items {
		// Convert reservation to deduction (decrements both reserved_quantity and total_quantity)
		if err := s.inventoryRepo.ConvertReservationToDeductionWithTx(tx, saleItem.BatchID, saleItem.Quantity); err != nil {
			s.logger.Error("Failed to convert reservation to deduction",
				zap.Error(err),
				zap.String("batch_id", saleItem.BatchID),
				zap.Int64("quantity", saleItem.Quantity))
			return errors.NewInternalServerError("Failed to convert reservation to stock deduction")
		}

		// Create inventory transaction for the sale
		note := "Sale completed - stock deducted from reservation"
		transaction := models.NewInventoryTransaction(
			saleItem.BatchID,
			models.TransactionTypeSale,
			-saleItem.Quantity, // Negative for deduction
			&sale.ID,
			&performedBy,
			&note,
			time.Now(),
		)

		if err := s.inventoryRepo.CreateTransactionWithTx(tx, transaction); err != nil {
			s.logger.Error("Failed to create sale transaction",
				zap.Error(err),
				zap.String("batch_id", saleItem.BatchID))
			return errors.NewInternalServerError("Failed to record sale transaction")
		}

		s.logger.Debug("Reservation converted to deduction",
			zap.String("batch_id", saleItem.BatchID),
			zap.Int64("quantity", saleItem.Quantity))
	}

	return nil
}

// handleCancellationInventory releases reservations or restores stock based on current sale status
func (s *SalesService) handleCancellationInventory(tx *gorm.DB, sale *models.Sale, performedBy string) error {
	isPendingSale := models.IsReservationStatus(sale.Status)

	s.logger.Debug("Handling cancellation inventory",
		zap.String("sale_id", sale.ID),
		zap.Bool("is_pending_sale", isPendingSale),
		zap.Int("item_count", len(sale.Items)))

	for _, saleItem := range sale.Items {
		var transactionType string
		var note string

		if isPendingSale {
			// Pending sale: release reservation (reserved_quantity decreases, total unchanged)
			if err := s.inventoryRepo.ReleaseBatchReservationWithTx(tx, saleItem.BatchID, saleItem.Quantity); err != nil {
				s.logger.Error("Failed to release reservation",
					zap.Error(err),
					zap.String("batch_id", saleItem.BatchID))
				return errors.NewInternalServerError("Failed to release reservation")
			}
			transactionType = models.TransactionTypeReservationRelease
			note = "Reservation released - sale cancelled via status update"
		} else {
			// Completed sale: restore actual stock (total_quantity increases)
			if err := s.inventoryRepo.UpdateBatchStockWithTx(tx, saleItem.BatchID, saleItem.Quantity); err != nil {
				s.logger.Error("Failed to restore stock",
					zap.Error(err),
					zap.String("batch_id", saleItem.BatchID))
				return errors.NewInternalServerError("Failed to restore stock")
			}
			transactionType = models.TransactionTypeCancellationReturn
			note = "Stock restored - completed sale cancelled via status update"
		}

		// Create inventory transaction
		transaction := models.NewInventoryTransaction(
			saleItem.BatchID,
			transactionType,
			saleItem.Quantity, // Positive for restoration/release
			&sale.ID,
			&performedBy,
			&note,
			time.Now(),
		)

		if err := s.inventoryRepo.CreateTransactionWithTx(tx, transaction); err != nil {
			s.logger.Error("Failed to create inventory transaction",
				zap.Error(err),
				zap.String("batch_id", saleItem.BatchID))
			return errors.NewInternalServerError("Failed to record inventory transaction")
		}

		s.logger.Debug("Inventory operation completed for cancellation",
			zap.String("batch_id", saleItem.BatchID),
			zap.String("transaction_type", transactionType),
			zap.Int64("quantity", saleItem.Quantity))
	}

	// Update cancellation timestamp
	now := time.Now()
	sale.CancelledAt = &now

	return nil
}

// GetSalesByDateRange retrieves sales within a date range
func (s *SalesService) GetSalesByDateRange(startDate, endDate time.Time, limit, offset int) ([]models.SaleResponse, int64, error) {
	sales, total, err := s.salesRepo.GetSalesByDateRange(startDate, endDate, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var responses []models.SaleResponse
	for _, sale := range sales {
		responses = append(responses, *s.mapSaleToResponse(&sale))
	}

	return responses, total, nil
}

// GetSalesByStatus retrieves sales by status
func (s *SalesService) GetSalesByStatus(status string, limit, offset int) ([]models.SaleResponse, int64, error) {
	sales, total, err := s.salesRepo.GetSalesByStatus(status, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var responses []models.SaleResponse
	for _, sale := range sales {
		responses = append(responses, *s.mapSaleToResponse(&sale))
	}

	return responses, total, nil
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

// getSellingPrice retrieves the appropriate selling price for a variant based on membership status
// For members (isOrgMember=true): member price → retail → MRP → any active
// For non-members (isOrgMember=false): retail → MRP → any active
func (s *SalesService) getSellingPrice(variantID string, isOrgMember bool) (float64, error) {
	// Get prices from product_prices table
	if s.priceRepo == nil {
		return 0, errors.NewInternalServerError("price repository not configured")
	}

	// 1. For FPO members: try member price first
	if isOrgMember {
		memberPrice, err := s.priceRepo.GetCurrentPrice(variantID, models.PriceTypeMember)
		if err == nil && memberPrice != nil {
			s.logger.Debug("Using member price",
				zap.String("variant_id", variantID),
				zap.Float64("price", memberPrice.Price))
			return memberPrice.Price, nil
		}
		// Member fallback continues to retail below
	}

	// 2. For non-members OR member fallback: use retail price
	retailPrice, err := s.priceRepo.GetCurrentPrice(variantID, models.PriceTypeRetail)
	if err == nil && retailPrice != nil {
		s.logger.Debug("Using retail price",
			zap.String("variant_id", variantID),
			zap.Float64("price", retailPrice.Price),
			zap.Bool("is_org_member", isOrgMember))
		return retailPrice.Price, nil
	}

	// 3. Fallback to MRP (for backwards compatibility)
	mrpPrice, err := s.priceRepo.GetCurrentPrice(variantID, models.PriceTypeMRP)
	if err == nil && mrpPrice != nil {
		s.logger.Debug("Using MRP price (fallback)",
			zap.String("variant_id", variantID),
			zap.Float64("price", mrpPrice.Price))
		return mrpPrice.Price, nil
	}

	// 4. Final fallback: any active price
	prices, err := s.priceRepo.GetActiveByVariantID(variantID)
	if err != nil || len(prices) == 0 {
		return 0, errors.NewNotFoundError("no pricing information found for variant")
	}

	s.logger.Debug("Using first available price (final fallback)",
		zap.String("variant_id", variantID),
		zap.Float64("price", prices[0].Price))
	return prices[0].Price, nil
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
		ID:            sale.ID,
		InvoiceNumber: sale.InvoiceNumber,
		WarehouseID:   sale.WarehouseID,
		SaleDate:      sale.SaleDate.Format("2006-01-02T15:04:05Z07:00"),
		TotalAmount:   sale.TotalAmount,
		Status:        sale.Status,
		// BRD Requirements - Customer tracking
		CustomerPhone: sale.CustomerPhone,
		CustomerName:  sale.CustomerName,
		IsOrgMember:   sale.IsOrgMember,
		PaymentMode:   sale.PaymentMode,
		SaleType:      sale.SaleType,
		ApplyTaxes:    sale.ApplyTaxes,
		CreatedAt:     sale.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     sale.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
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
		// Get SKU from batch's variant (Issue 4)
		sku := ""
		if item.Batch.Variant.SKU != nil {
			sku = *item.Batch.Variant.SKU
		}
		response.Items = append(response.Items, models.SaleItemResponse{
			ID:           item.ID,
			SaleID:       item.SaleID,
			BatchID:      item.BatchID,
			SKU:          sku,
			Quantity:     item.Quantity,
			SellingPrice: utils.RoundPrice(item.SellingPrice),
			LineTotal:    utils.RoundPrice(item.LineTotal),
			// BRD Requirements - Cost and Margin
			CostPrice:      utils.RoundPrice(item.CostPrice),
			Margin:         utils.RoundPrice(item.Margin),
			CGSTAmount:     utils.RoundPrice(item.CGSTAmount),
			SGSTAmount:     utils.RoundPrice(item.SGSTAmount),
			IGSTAmount:     utils.RoundPrice(item.IGSTAmount),
			TotalTaxAmount: utils.RoundPrice(item.TotalTaxAmount),
			CreatedAt:      item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return response
}

// mapSaleToListResponse maps a Sale model to SaleListResponse (without items for performance)
func (s *SalesService) mapSaleToListResponse(sale *models.Sale) *models.SaleListResponse {
	response := &models.SaleListResponse{
		ID:            sale.ID,
		InvoiceNumber: sale.InvoiceNumber,
		WarehouseID:   sale.WarehouseID,
		SaleDate:      sale.SaleDate.Format("2006-01-02T15:04:05Z07:00"),
		TotalAmount:   sale.TotalAmount,
		Status:        sale.Status,
		// BRD Requirements - Customer tracking
		CustomerPhone: sale.CustomerPhone,
		CustomerName:  sale.CustomerName,
		IsOrgMember:   sale.IsOrgMember,
		PaymentMode:   sale.PaymentMode,
		SaleType:      sale.SaleType,
		ApplyTaxes:    sale.ApplyTaxes,
		CreatedAt:     sale.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     sale.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Add cancellation fields if present
	if sale.CancelledAt != nil {
		cancelledAtStr := sale.CancelledAt.Format("2006-01-02T15:04:05Z07:00")
		response.CancelledAt = &cancelledAtStr
	}
	if sale.CancellationReason != nil {
		response.CancellationReason = sale.CancellationReason
	}

	// Note: Items and Breakdown are omitted for performance
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

// applyTaxesToSaleWithTx aggregates taxes from sale items and creates TaxSummary within a transaction
// GST-only tax system: taxes are already calculated per-item, this just aggregates them
func (s *SalesService) applyTaxesToSaleWithTx(tx *gorm.DB, saleID string, saleItems []models.SaleItem, warehouseID string) (*models.TaxSummary, error) {
	// Aggregate tax amounts from sale items (already calculated during item creation)
	var totalCGST, totalSGST, totalIGST, totalTax, subTotal float64
	isInterState := false

	for _, item := range saleItems {
		totalCGST += item.CGSTAmount
		totalSGST += item.SGSTAmount
		totalIGST += item.IGSTAmount
		totalTax += item.TotalTaxAmount
		subTotal += item.LineTotal
		// If any item has IGST, it's an inter-state sale
		if item.IGSTAmount > 0 {
			isInterState = true
		}
	}

	grandTotal := subTotal + totalTax

	// Create tax summary
	taxSummary := models.NewTaxSummary()
	taxSummary.SaleID = &saleID
	taxSummary.CGSTAmount = totalCGST
	taxSummary.SGSTAmount = totalSGST
	taxSummary.IGSTAmount = totalIGST
	taxSummary.TotalTaxAmount = totalTax
	taxSummary.SubTotal = subTotal
	taxSummary.GrandTotal = grandTotal
	taxSummary.IsInterState = isInterState

	// Save tax summary using transaction
	if err := s.taxRepo.CreateTaxSummaryWithTx(tx, taxSummary); err != nil {
		return nil, err
	}

	return taxSummary, nil
}

// applyTaxesToSale aggregates taxes from sale items and creates TaxSummary (no transaction)
// GST-only tax system: taxes are already calculated per-item, this just aggregates them
func (s *SalesService) applyTaxesToSale(saleID string, saleItems []models.SaleItem, warehouseID string) (*models.TaxSummary, error) {
	// Aggregate tax amounts from sale items (already calculated during item creation)
	var totalCGST, totalSGST, totalIGST, totalTax, subTotal float64
	isInterState := false

	for _, item := range saleItems {
		totalCGST += item.CGSTAmount
		totalSGST += item.SGSTAmount
		totalIGST += item.IGSTAmount
		totalTax += item.TotalTaxAmount
		subTotal += item.LineTotal
		// If any item has IGST, it's an inter-state sale
		if item.IGSTAmount > 0 {
			isInterState = true
		}
	}

	grandTotal := subTotal + totalTax

	// Create tax summary
	taxSummary := models.NewTaxSummary()
	taxSummary.SaleID = &saleID
	taxSummary.CGSTAmount = totalCGST
	taxSummary.SGSTAmount = totalSGST
	taxSummary.IGSTAmount = totalIGST
	taxSummary.TotalTaxAmount = totalTax
	taxSummary.SubTotal = subTotal
	taxSummary.GrandTotal = grandTotal
	taxSummary.IsInterState = isInterState

	// Save tax summary
	if err := s.taxRepo.CreateTaxSummary(taxSummary); err != nil {
		return nil, err
	}

	return taxSummary, nil
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
// Uses ValidStatusTransitions from models to determine if cancellation is allowed
func (s *SalesService) CanCancelSale(sale *models.Sale) (bool, string) {
	switch sale.Status {
	case models.SaleStatusCancelled:
		return false, "Sale is already cancelled"
	case models.SaleStatusPending, models.SaleStatusCompleted:
		// Both pending and completed sales can be cancelled per state machine
		return true, ""
	default:
		return false, "Unknown sale status: " + sale.Status
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

			// Bug #3 Fix: Use transactional batch read (includes soft-deleted batches)
			batch, err := s.inventoryRepo.GetBatchByIDWithTx(tx, saleItem.BatchID)
			if err != nil {
				s.logger.Error("Failed to get batch",
					zap.Error(err),
					zap.String("batch_id", saleItem.BatchID))
				return errors.NewInternalServerError("Failed to get batch information")
			}

			// Track whether inventory was actually restored
			inventoryWasRestored := false
			var transactionID *string

			// Smart release/restore based on sale status and SkipInventoryReturn flag
			if !req.SkipInventoryReturn {
				// Determine transaction type and action based on sale status
				isPendingSale := models.IsReservationStatus(sale.Status)

				var transactionType string
				var note string

				if isPendingSale {
					// PENDING SALE: Release reservation (reserved_quantity decreases, total_quantity unchanged)
					transactionType = "reservation_release"
					note = "Reservation released for cancelled pending sale " + sale.ID

					if err := s.inventoryRepo.ReleaseBatchReservationWithTx(tx, saleItem.BatchID, saleItem.Quantity); err != nil {
						s.logger.Error("Failed to release batch reservation",
							zap.Error(err),
							zap.String("batch_id", saleItem.BatchID))
						return errors.NewInternalServerError("Failed to release reservation")
					}
					s.logger.Info("Reservation released from batch",
						zap.String("batch_id", saleItem.BatchID),
						zap.Int64("quantity_released", saleItem.Quantity))

					// For pending sales, we released reservation but didn't restore stock (it was never deducted)
					inventoryWasRestored = false
				} else {
					// COMPLETED SALE: Restore actual stock (total_quantity increases)
					transactionType = "cancellation_return"
					note = "Inventory restored for cancelled completed sale " + sale.ID

					if err := s.inventoryRepo.UpdateBatchStockWithTx(tx, saleItem.BatchID, saleItem.Quantity); err != nil {
						s.logger.Error("Failed to restore batch stock",
							zap.Error(err),
							zap.String("batch_id", saleItem.BatchID))
						return errors.NewInternalServerError("Failed to restore batch stock")
					}
					s.logger.Info("Inventory restored to batch",
						zap.String("batch_id", saleItem.BatchID),
						zap.Int64("quantity_restored", saleItem.Quantity))

					inventoryWasRestored = true
				}

				// Create inventory transaction for audit trail
				transaction := models.NewInventoryTransaction(
					saleItem.BatchID,
					transactionType,
					saleItem.Quantity, // Positive for release/restore
					&cancellation.ID,
					&req.PerformedBy,
					&note,
					time.Now(),
				)

				if err := s.inventoryRepo.CreateTransactionWithTx(tx, transaction); err != nil {
					s.logger.Error("Failed to create inventory transaction",
						zap.Error(err),
						zap.String("batch_id", saleItem.BatchID))
					return errors.NewInternalServerError("Failed to record inventory transaction")
				}
				s.logger.Debug("Inventory transaction created",
					zap.String("transaction_id", transaction.ID),
					zap.String("transaction_type", transactionType))

				transactionID = &transaction.ID
			} else {
				s.logger.Debug("Skipping inventory restoration as requested",
					zap.String("batch_id", saleItem.BatchID))
			}

			// Bug #2 Fix: Create SaleCancellationItem with accurate inventoryRestored flag
			cancellationItem := models.NewSaleCancellationItem(
				cancellation.ID,
				saleItem.ID,
				saleItem.BatchID,
				saleItem.Quantity,
				saleItem.LineTotal,
				transactionID,
				inventoryWasRestored, // Now accurately reflects whether inventory was restored
			)

			if err := s.saleCancellationRepo.CreateCancellationItemWithTx(tx, cancellationItem); err != nil {
				s.logger.Error("Failed to create cancellation item",
					zap.Error(err),
					zap.String("sale_item_id", saleItem.ID))
				return errors.NewInternalServerError("Failed to create cancellation item")
			}

			// Add to restored items list (only if inventory was actually restored)
			if inventoryWasRestored {
				inventoryRestored = append(inventoryRestored, models.InventoryRestoredItem{
					BatchID:          saleItem.BatchID,
					VariantID:        batch.VariantID,
					QuantityRestored: saleItem.Quantity,
					TransactionID:    *transactionID,
				})
			}
		}

		// 5. Reverse discounts (if any were applied)
		s.logger.Debug("Checking for discounts to reverse", zap.String("sale_id", sale.ID))
		var discountReversedInfo *models.DiscountReversedInfo
		discountUsages, err := s.discountsRepo.GetDiscountUsageBySaleWithTx(tx, sale.ID)
		if err != nil {
			s.logger.Warn("Failed to get discount usages for sale",
				zap.Error(err),
				zap.String("sale_id", sale.ID))
			// Non-critical - continue with cancellation
		} else if len(discountUsages) > 0 {
			s.logger.Info("Found discounts to reverse",
				zap.String("sale_id", sale.ID),
				zap.Int("discount_count", len(discountUsages)))

			var totalDiscountReversed float64
			var lastDiscountID string
			for _, usage := range discountUsages {
				// Decrement usage count for the discount
				if err := s.discountsRepo.DecrementUsageWithTx(tx, usage.DiscountID); err != nil {
					s.logger.Warn("Failed to decrement discount usage",
						zap.Error(err),
						zap.String("discount_id", usage.DiscountID))
					// Non-critical - continue
				} else {
					s.logger.Debug("Decremented discount usage",
						zap.String("discount_id", usage.DiscountID))
				}
				totalDiscountReversed += usage.Amount
				lastDiscountID = usage.DiscountID
			}

			// Delete the discount usage records
			if err := s.discountsRepo.DeleteDiscountUsagesBySaleWithTx(tx, sale.ID); err != nil {
				s.logger.Warn("Failed to delete discount usage records",
					zap.Error(err),
					zap.String("sale_id", sale.ID))
				// Non-critical - continue
			} else {
				s.logger.Debug("Deleted discount usage records",
					zap.String("sale_id", sale.ID))
			}

			// Update cancellation record with discount reversal info
			cancellation.DiscountReversed = totalDiscountReversed
			if err := s.saleCancellationRepo.UpdateCancellationWithTx(tx, cancellation); err != nil {
				s.logger.Warn("Failed to update cancellation with discount info",
					zap.Error(err),
					zap.String("cancellation_id", cancellation.ID))
				// Non-critical - continue
			}

			// Build discount reversal info for response
			discountReversedInfo = &models.DiscountReversedInfo{
				DiscountID:       lastDiscountID, // Use last discount ID (if multiple, frontend should query for full details)
				AmountReversed:   totalDiscountReversed,
				UsageDecremented: true,
			}
			s.logger.Info("Discount reversal completed",
				zap.Float64("total_reversed", totalDiscountReversed))
		}

		// 6. Void tax records (if taxes were applied)
		s.logger.Debug("Checking for taxes to void", zap.String("sale_id", sale.ID))
		var taxVoidedInfo *models.TaxVoidedInfo
		taxSummary, err := s.taxRepo.GetTaxSummaryBySaleWithTx(tx, sale.ID)
		if err != nil {
			s.logger.Warn("Failed to get tax summary for sale",
				zap.Error(err),
				zap.String("sale_id", sale.ID))
			// Non-critical - continue with cancellation
		} else if taxSummary != nil && taxSummary.TotalTaxAmount > 0 {
			s.logger.Info("Found taxes to void",
				zap.String("sale_id", sale.ID),
				zap.Float64("total_tax_amount", taxSummary.TotalTaxAmount))

			// Store tax summary ID and amount before deleting
			taxSummaryID := taxSummary.ID
			totalTaxVoided := taxSummary.TotalTaxAmount

			// GST-only tax system: No TaxApplications table, only TaxSummary
			// Delete tax summary for this sale
			if err := s.taxRepo.DeleteTaxSummaryBySaleWithTx(tx, sale.ID); err != nil {
				s.logger.Warn("Failed to delete tax summary",
					zap.Error(err),
					zap.String("sale_id", sale.ID))
				// Non-critical - continue
			} else {
				s.logger.Debug("Deleted tax summary",
					zap.String("sale_id", sale.ID))
			}

			// Update cancellation record with tax reversal info
			cancellation.TaxReversed = totalTaxVoided
			if err := s.saleCancellationRepo.UpdateCancellationWithTx(tx, cancellation); err != nil {
				s.logger.Warn("Failed to update cancellation with tax info",
					zap.Error(err),
					zap.String("cancellation_id", cancellation.ID))
				// Non-critical - continue
			}

			// Build tax voided info for response
			taxVoidedInfo = &models.TaxVoidedInfo{
				TaxSummaryID: taxSummaryID,
				AmountVoided: totalTaxVoided,
			}
			s.logger.Info("Tax voiding completed",
				zap.Float64("total_voided", totalTaxVoided))
		}

		// 7. Update Sale status to cancelled
		s.logger.Debug("Updating sale status to cancelled",
			zap.String("sale_id", sale.ID))
		sale.Status = models.SaleStatusCancelled
		now := time.Now()
		sale.CancelledAt = &now
		sale.CancellationReason = &req.Reason

		if err := s.salesRepo.UpdateSaleWithTx(tx, sale); err != nil {
			s.logger.Error("Failed to update sale status",
				zap.Error(err),
				zap.String("sale_id", sale.ID))
			return errors.NewInternalServerError("Failed to update sale status")
		}

		// Build financial adjustments if any
		var financialAdjustments *models.FinancialAdjustmentsResponse
		if discountReversedInfo != nil || taxVoidedInfo != nil {
			financialAdjustments = &models.FinancialAdjustmentsResponse{
				DiscountReversed: discountReversedInfo,
				TaxVoided:        taxVoidedInfo,
			}
		}

		// Build response
		response = &models.CancelSaleResponse{
			Sale:                 *s.mapSaleToResponse(sale),
			InventoryRestored:    inventoryRestored,
			FinancialAdjustments: financialAdjustments,
			CancellationID:       cancellation.ID,
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

// GetCancellations retrieves all cancellations for a sale
func (s *SalesService) GetCancellations(saleID string) (*models.GetCancellationsResponse, error) {
	s.logger.Info("Getting cancellations for sale", zap.String("sale_id", saleID))

	// Verify sale exists
	sale, err := s.salesRepo.GetSaleByID(saleID)
	if err != nil {
		s.logger.Error("Failed to get sale", zap.Error(err), zap.String("sale_id", saleID))
		return nil, errors.NewNotFoundError("Sale")
	}

	// Get all cancellations for this sale
	cancellations, err := s.saleCancellationRepo.GetCancellationsBySaleID(sale.ID)
	if err != nil {
		s.logger.Error("Failed to get cancellations", zap.Error(err), zap.String("sale_id", saleID))
		return nil, errors.NewInternalServerError("Failed to retrieve cancellations")
	}

	// Map to response
	var cancellationResponses []models.SaleCancellationResponse
	for _, c := range cancellations {
		var itemResponses []models.SaleCancellationItemResponse
		for _, item := range c.Items {
			itemResponses = append(itemResponses, models.SaleCancellationItemResponse{
				ID:                item.ID,
				CancellationID:    item.CancellationID,
				SaleItemID:        item.SaleItemID,
				BatchID:           item.BatchID,
				QuantityCancelled: item.QuantityCancelled,
				RefundAmount:      item.RefundAmount,
				InventoryRestored: item.InventoryRestored,
				TransactionID:     item.TransactionID,
				CreatedAt:         item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		cancellationResponses = append(cancellationResponses, models.SaleCancellationResponse{
			ID:               c.ID,
			SaleID:           c.SaleID,
			CancellationType: c.CancellationType,
			CancelledBy:      c.CancelledBy,
			Reason:           c.Reason,
			ReasonDetails:    c.ReasonDetails,
			CancelledAt:      c.CancelledAt.Format("2006-01-02T15:04:05Z07:00"),
			OriginalAmount:   c.OriginalAmount,
			CancelledAmount:  c.CancelledAmount,
			DiscountReversed: c.DiscountReversed,
			TaxReversed:      c.TaxReversed,
			Items:            itemResponses,
			CreatedAt:        c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:        c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	s.logger.Info("Retrieved cancellations successfully",
		zap.String("sale_id", saleID),
		zap.Int("count", len(cancellations)))

	return &models.GetCancellationsResponse{
		SaleID:        saleID,
		Cancellations: cancellationResponses,
		TotalCount:    len(cancellations),
	}, nil
}

// CancelItems cancels specific items in a sale (partial cancellation)
func (s *SalesService) CancelItems(saleID string, req *models.CancelItemsRequest) (*models.CancelItemsResponse, error) {
	s.logger.Info("Starting partial sale cancellation",
		zap.String("sale_id", saleID),
		zap.String("reason", req.Reason),
		zap.Int("items_count", len(req.Items)))

	// Validate reason - if "other", reason_details is required
	if req.Reason == models.ReasonOther && (req.ReasonDetails == nil || *req.ReasonDetails == "") {
		s.logger.Error("Reason details required for 'other' reason", zap.String("sale_id", saleID))
		return nil, errors.NewValidationError("reason_details is required when reason is 'other'")
	}

	var response *models.CancelItemsResponse

	err := s.salesRepo.WithTransaction(func(tx *gorm.DB) error {
		// 1. Lock sale record (SELECT FOR UPDATE)
		sale, err := s.salesRepo.GetSaleForUpdateWithTx(tx, saleID)
		if err != nil {
			s.logger.Error("Failed to get sale for update", zap.Error(err), zap.String("sale_id", saleID))
			if err == gorm.ErrRecordNotFound {
				return errors.NewNotFoundError("Sale")
			}
			return errors.NewInternalServerError("Failed to lock sale")
		}

		// 2. Check if sale can have items cancelled
		// Valid statuses are: pending, completed, cancelled (see models/sales.go)
		if sale.Status == models.SaleStatusCancelled {
			return errors.NewBadRequestError("Sale is already fully cancelled")
		}
		// Note: Only pending and completed sales can have items cancelled
		// - pending: releases reservation
		// - completed: restores stock

		// 3. Build a map of sale items for quick lookup
		saleItemMap := make(map[string]*models.SaleItem)
		for i := range sale.Items {
			saleItemMap[sale.Items[i].ID] = &sale.Items[i]
		}

		// 4. Validate all requested items
		var totalCancelledAmount float64
		for _, itemReq := range req.Items {
			// Validate quantity is positive
			if itemReq.Quantity <= 0 {
				return errors.NewBadRequestError("Cancellation quantity must be greater than 0 for item " + itemReq.SaleItemID)
			}
			saleItem, exists := saleItemMap[itemReq.SaleItemID]
			if !exists {
				return errors.NewBadRequestError("Sale item " + itemReq.SaleItemID + " not found in this sale")
			}
			if itemReq.Quantity > saleItem.Quantity {
				return errors.NewBadRequestError("Cannot cancel more than available quantity for item " + itemReq.SaleItemID)
			}
		}

		// 5. Create SaleCancellation record (partial)
		cancellation := models.NewSaleCancellation(
			sale.ID,
			models.CancellationTypePartial,
			req.Reason,
			&req.PerformedBy,
			req.ReasonDetails,
			sale.TotalAmount,
			0, // Will be calculated
		)

		if err := s.saleCancellationRepo.CreateCancellationWithTx(tx, cancellation); err != nil {
			s.logger.Error("Failed to create cancellation record", zap.Error(err))
			return errors.NewInternalServerError("Failed to create cancellation record")
		}

		// 6. Process each item
		var inventoryRestored []models.InventoryRestoredItem
		var cancelledItems []models.CancelledItemInfo

		for _, itemReq := range req.Items {
			saleItem := saleItemMap[itemReq.SaleItemID]

			// Guard against division by zero (data integrity check)
			if saleItem.Quantity <= 0 {
				s.logger.Error("Invalid sale item: quantity is zero or negative",
					zap.String("sale_item_id", saleItem.ID),
					zap.Int64("quantity", saleItem.Quantity))
				return errors.NewInternalServerError("Data integrity error: sale item has invalid quantity")
			}

			// Calculate refund amount for this item (including proportional taxes)
			basePricePerUnit := saleItem.LineTotal / float64(saleItem.Quantity)
			taxPerUnit := saleItem.TotalTaxAmount / float64(saleItem.Quantity)
			totalPricePerUnit := basePricePerUnit + taxPerUnit
			refundAmount := totalPricePerUnit * float64(itemReq.Quantity)
			totalCancelledAmount += refundAmount

			// Use transactional batch read (includes soft-deleted batches)
			batch, err := s.inventoryRepo.GetBatchByIDWithTx(tx, saleItem.BatchID)
			if err != nil {
				s.logger.Error("Failed to get batch for partial cancellation",
					zap.Error(err),
					zap.String("batch_id", saleItem.BatchID))
				return errors.NewInternalServerError("Failed to get batch information")
			}

			// Handle inventory based on sale status
			isPendingSale := models.IsReservationStatus(sale.Status)
			var transactionType string
			var note string
			var inventoryWasRestored bool

			if isPendingSale {
				// PENDING: Release reservation (reserved_quantity decreases, total unchanged)
				if err := s.inventoryRepo.ReleaseBatchReservationWithTx(tx, saleItem.BatchID, itemReq.Quantity); err != nil {
					s.logger.Error("Failed to release reservation for partial cancellation",
						zap.Error(err),
						zap.String("batch_id", saleItem.BatchID))
					return errors.NewInternalServerError("Failed to release reservation")
				}
				transactionType = models.TransactionTypeReservationRelease
				note = "Reservation released for partial cancellation " + cancellation.ID
				inventoryWasRestored = false // Reservation released, not stock restored
			} else {
				// COMPLETED: Restore actual stock (total_quantity increases)
				if err := s.inventoryRepo.UpdateBatchStockWithTx(tx, saleItem.BatchID, itemReq.Quantity); err != nil {
					s.logger.Error("Failed to restore stock for partial cancellation",
						zap.Error(err),
						zap.String("batch_id", saleItem.BatchID))
					return errors.NewInternalServerError("Failed to restore stock")
				}
				transactionType = models.TransactionTypeCancellationReturn
				note = "Inventory restored for partial cancellation " + cancellation.ID
				inventoryWasRestored = true // Actual stock was restored
			}

			// Create inventory transaction
			transaction := models.NewInventoryTransaction(
				saleItem.BatchID,
				transactionType,
				itemReq.Quantity,
				&cancellation.ID,
				&req.PerformedBy,
				&note,
				time.Now(),
			)

			if err := s.inventoryRepo.CreateTransactionWithTx(tx, transaction); err != nil {
				s.logger.Error("Failed to create inventory transaction",
					zap.Error(err),
					zap.String("batch_id", saleItem.BatchID))
				return errors.NewInternalServerError("Failed to record inventory change")
			}

			// Create cancellation item record with accurate inventoryRestored flag
			cancellationItem := models.NewSaleCancellationItem(
				cancellation.ID,
				saleItem.ID,
				saleItem.BatchID,
				itemReq.Quantity,
				refundAmount,
				&transaction.ID,
				inventoryWasRestored, // Accurate based on actual operation performed
			)

			if err := s.saleCancellationRepo.CreateCancellationItemWithTx(tx, cancellationItem); err != nil {
				s.logger.Error("Failed to create cancellation item", zap.Error(err))
				return errors.NewInternalServerError("Failed to create cancellation item")
			}

			// Update sale item quantity if partial
			if itemReq.Quantity < saleItem.Quantity {
				newQuantity := saleItem.Quantity - itemReq.Quantity
				// Recalculate all proportional fields (GST-only tax system)
				newLineTotal := basePricePerUnit * float64(newQuantity)
				marginPerUnit := saleItem.Margin / float64(saleItem.Quantity)
				cgstPerUnit := saleItem.CGSTAmount / float64(saleItem.Quantity)
				sgstPerUnit := saleItem.SGSTAmount / float64(saleItem.Quantity)
				igstPerUnit := saleItem.IGSTAmount / float64(saleItem.Quantity)

				newMargin := marginPerUnit * float64(newQuantity)
				newCGST := cgstPerUnit * float64(newQuantity)
				newSGST := sgstPerUnit * float64(newQuantity)
				newIGST := igstPerUnit * float64(newQuantity)
				newTotalTax := newCGST + newSGST + newIGST

				if err := tx.Model(&models.SaleItem{}).Where("id = ?", saleItem.ID).Updates(map[string]interface{}{
					"quantity":         newQuantity,
					"line_total":       newLineTotal,
					"margin":           newMargin,
					"cgst_amount":      newCGST,
					"sgst_amount":      newSGST,
					"igst_amount":      newIGST,
					"total_tax_amount": newTotalTax,
				}).Error; err != nil {
					s.logger.Error("Failed to update sale item quantity", zap.Error(err))
					return errors.NewInternalServerError("Failed to update sale item")
				}
			} else {
				// Full item cancelled - manually clean up all FK references before deleting
				// Delete all SaleCancellationItem records that reference this SaleItem
				// (they were just created in this same transaction for audit purposes)
				if err := tx.Where("sale_item_id = ?", saleItem.ID).Delete(&models.SaleCancellationItem{}).Error; err != nil {
					s.logger.Error("Failed to delete sale cancellation item references", zap.Error(err))
					return errors.NewInternalServerError("Failed to clean up cancellation records")
				}

				// Now safe to delete the SaleItem (no foreign key references remain)
				if err := tx.Delete(&models.SaleItem{}, "id = ?", saleItem.ID).Error; err != nil {
					s.logger.Error("Failed to delete sale item", zap.Error(err))
					return errors.NewInternalServerError("Failed to remove cancelled item")
				}
			}

			// Add to response lists - only include in InventoryRestored if stock was actually restored
			// (not for pending sales where only reservation was released)
			if inventoryWasRestored && batch != nil {
				inventoryRestored = append(inventoryRestored, models.InventoryRestoredItem{
					BatchID:          saleItem.BatchID,
					VariantID:        batch.VariantID,
					QuantityRestored: itemReq.Quantity,
					TransactionID:    transaction.ID,
				})
			}

			cancelledItems = append(cancelledItems, models.CancelledItemInfo{
				SaleItemID:        saleItem.ID,
				QuantityCancelled: itemReq.Quantity,
				AmountRefunded:    refundAmount,
			})
		}

		// 7. Update cancellation with total amount
		cancellation.CancelledAmount = totalCancelledAmount

		// 7a. Calculate proportional discount/tax reversal for partial cancellation
		var financialAdjustments *models.FinancialAdjustmentsResponse
		if totalCancelledAmount > 0 && sale.TotalAmount > 0 {
			cancelRatio := totalCancelledAmount / sale.TotalAmount

			// Calculate proportional discount reversal (don't decrement usage - sale is still active)
			discountUsages, err := s.discountsRepo.GetDiscountUsageBySaleWithTx(tx, sale.ID)
			if err != nil {
				s.logger.Warn("Failed to get discount usages for partial cancellation", zap.Error(err))
			} else if len(discountUsages) > 0 {
				var totalDiscountAmount float64
				var lastDiscountID string
				for _, usage := range discountUsages {
					totalDiscountAmount += usage.Amount
					lastDiscountID = usage.DiscountID
				}
				proportionalDiscount := totalDiscountAmount * cancelRatio
				cancellation.DiscountReversed = proportionalDiscount

				if financialAdjustments == nil {
					financialAdjustments = &models.FinancialAdjustmentsResponse{}
				}
				financialAdjustments.DiscountReversed = &models.DiscountReversedInfo{
					DiscountID:       lastDiscountID,
					AmountReversed:   proportionalDiscount,
					UsageDecremented: false, // Partial cancellation doesn't decrement usage
				}
				s.logger.Debug("Calculated proportional discount reversal",
					zap.Float64("cancel_ratio", cancelRatio),
					zap.Float64("proportional_discount", proportionalDiscount))
			}

			// Calculate proportional tax reversal (don't delete records - sale is still active)
			taxSummary, err := s.taxRepo.GetTaxSummaryBySaleWithTx(tx, sale.ID)
			if err != nil {
				s.logger.Warn("Failed to get tax summary for partial cancellation", zap.Error(err))
			} else if taxSummary != nil && taxSummary.TotalTaxAmount > 0 {
				proportionalTax := taxSummary.TotalTaxAmount * cancelRatio
				cancellation.TaxReversed = proportionalTax

				if financialAdjustments == nil {
					financialAdjustments = &models.FinancialAdjustmentsResponse{}
				}
				financialAdjustments.TaxVoided = &models.TaxVoidedInfo{
					TaxSummaryID: taxSummary.ID,
					AmountVoided: proportionalTax,
				}
				s.logger.Debug("Calculated proportional tax reversal",
					zap.Float64("cancel_ratio", cancelRatio),
					zap.Float64("proportional_tax", proportionalTax))
			}
		}

		if err := s.saleCancellationRepo.UpdateCancellationWithTx(tx, cancellation); err != nil {
			s.logger.Warn("Failed to update cancellation amount", zap.Error(err))
		}

		// 8. Update sale total (keep original status for partial cancellations)
		newSaleTotal := sale.TotalAmount - totalCancelledAmount
		if err := tx.Model(&models.Sale{}).Where("id = ?", sale.ID).Update("total_amount", newSaleTotal).Error; err != nil {
			s.logger.Error("Failed to update sale total", zap.Error(err))
			return errors.NewInternalServerError("Failed to update sale")
		}
		// Note: Status remains unchanged for partial cancellations (pending/completed)

		// 9. Re-fetch sale to get updated items (avoids stale data in response)
		updatedSale, err := s.salesRepo.GetSaleForUpdateWithTx(tx, sale.ID)
		if err != nil {
			s.logger.Warn("Failed to re-fetch sale for response, using original data", zap.Error(err))
			updatedSale = sale
			updatedSale.TotalAmount = newSaleTotal
		}

		// Build response with fresh data
		response = &models.CancelItemsResponse{
			Sale:                 *s.mapSaleToResponse(updatedSale),
			ItemsCancelled:       cancelledItems,
			InventoryRestored:    inventoryRestored,
			FinancialAdjustments: financialAdjustments,
			CancellationID:       cancellation.ID,
			NewSaleTotal:         newSaleTotal,
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Partial sale cancellation failed", zap.Error(err), zap.String("sale_id", saleID))
		return nil, err
	}

	s.logger.Info("Partial sale cancellation completed successfully",
		zap.String("sale_id", saleID),
		zap.String("cancellation_id", response.CancellationID))

	return response, nil
}

// CompleteSale converts a pending sale to completed by converting reservations to actual deductions
// This is called when a pending sale is ready to be fulfilled (e.g., payment confirmed, order shipped)
func (s *SalesService) CompleteSale(saleID string, performedBy string) (*models.SaleResponse, error) {
	s.logger.Info("Starting sale completion",
		zap.String("sale_id", saleID),
		zap.String("performed_by", performedBy))

	var response *models.SaleResponse

	err := s.salesRepo.WithTransaction(func(tx *gorm.DB) error {
		// 1. Lock and retrieve sale
		sale, err := s.salesRepo.GetSaleForUpdateWithTx(tx, saleID)
		if err != nil {
			s.logger.Error("Failed to get sale for completion",
				zap.Error(err),
				zap.String("sale_id", saleID))
			if err == gorm.ErrRecordNotFound {
				return errors.NewNotFoundError("Sale")
			}
			return errors.NewInternalServerError("Failed to lock sale")
		}

		// 2. Validate sale status - must be pending to complete
		if sale.Status != models.SaleStatusPending {
			s.logger.Error("Sale is not in pending status",
				zap.String("sale_id", saleID),
				zap.String("current_status", sale.Status))
			return errors.NewBadRequestError("Only pending sales can be completed. Current status: " + sale.Status)
		}

		// 3. Convert reservations to actual deductions using helper method
		if err := s.handleCompletionInventory(tx, sale, performedBy); err != nil {
			return err
		}

		// 4. Update sale status to completed
		sale.Status = models.SaleStatusCompleted
		if err := s.salesRepo.UpdateSaleWithTx(tx, sale); err != nil {
			s.logger.Error("Failed to update sale status to completed",
				zap.Error(err),
				zap.String("sale_id", saleID))
			return errors.NewInternalServerError("Failed to update sale status")
		}

		s.logger.Info("Sale status updated to completed",
			zap.String("sale_id", saleID))

		// Build response
		response = s.mapSaleToResponse(sale)
		return nil
	})

	if err != nil {
		s.logger.Error("Sale completion failed",
			zap.Error(err),
			zap.String("sale_id", saleID))
		return nil, err
	}

	s.logger.Info("Sale completion finished successfully",
		zap.String("sale_id", saleID))

	return response, nil
}
