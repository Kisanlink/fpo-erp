package services

import (
	"context"
	"encoding/json"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/utils"

	"go.uber.org/zap"
)

// InventoryService handles inventory business logic
type InventoryService struct {
	inventoryRepo *repositories.InventoryRepository
	warehouseRepo *repositories.WarehouseRepository
	productRepo   *repositories.ProductRepository
	variantRepo   *repositories.ProductVariantRepository
	priceRepo     *repositories.ProductPriceRepository
	addressClient *aaa.AddressGRPCClient
	logger        interfaces.Logger
}

// NewInventoryService creates a new inventory service
func NewInventoryService(inventoryRepo *repositories.InventoryRepository, warehouseRepo *repositories.WarehouseRepository, productRepo *repositories.ProductRepository, variantRepo *repositories.ProductVariantRepository, priceRepo *repositories.ProductPriceRepository, addressClient *aaa.AddressGRPCClient, logger interfaces.Logger) *InventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
		warehouseRepo: warehouseRepo,
		productRepo:   productRepo,
		variantRepo:   variantRepo,
		priceRepo:     priceRepo,
		addressClient: addressClient,
		logger:        logger,
	}
}

// CreateBatch creates a new inventory batch
// Simplified for GST-only tax system - tax rates are on ProductVariant, not on batches
func (s *InventoryService) CreateBatch(warehouseID, variantID string, costPrice float64, expiryDate time.Time, quantity int64) (*models.InventoryBatchResponse, error) {
	s.logger.Info("Creating inventory batch",
		zap.String("warehouse_id", warehouseID),
		zap.String("variant_id", variantID),
		zap.Int64("quantity", quantity),
		zap.Float64("cost_price", costPrice))

	// Validate warehouse exists
	s.logger.Debug("Validating warehouse exists",
		zap.String("warehouse_id", warehouseID))
	_, err := s.warehouseRepo.GetByID(warehouseID)
	if err != nil {
		s.logger.Error("Warehouse not found",
			zap.Error(err),
			zap.String("warehouse_id", warehouseID))
		return nil, errors.NewNotFoundError("Warehouse")
	}

	// Validate product variant exists
	s.logger.Debug("Validating product variant exists",
		zap.String("variant_id", variantID))
	_, err = s.variantRepo.GetByID(variantID)
	if err != nil {
		s.logger.Error("Product variant not found",
			zap.Error(err),
			zap.String("variant_id", variantID))
		return nil, errors.NewNotFoundError("ProductVariant")
	}

	// Validate expiry date is in the future
	s.logger.Debug("Validating expiry date",
		zap.Time("expiry_date", expiryDate))
	if expiryDate.Before(time.Now()) {
		s.logger.Error("Invalid expiry date - must be in the future",
			zap.Time("expiry_date", expiryDate))
		return nil, errors.NewBadRequestError("Expiry date must be in the future")
	}

	// Validate quantity is positive
	if quantity <= 0 {
		s.logger.Error("Invalid quantity - must be positive",
			zap.Int64("quantity", quantity))
		return nil, errors.NewBadRequestError("Quantity must be positive")
	}

	// Create batch using the simplified constructor (no tax params - GST on variant)
	batch := models.NewInventoryBatch(warehouseID, variantID, costPrice, expiryDate, quantity)

	// Create initial transaction using the proper constructor
	note := "Initial import"
	transaction := models.NewInventoryTransaction("", "import", quantity, nil, nil, &note, time.Now())

	// Create batch and initial transaction atomically
	s.logger.Debug("Saving batch and initial transaction to database")
	if err := s.inventoryRepo.CreateBatchWithTransaction(batch, transaction); err != nil {
		s.logger.Error("Failed to create inventory batch",
			zap.Error(err),
			zap.String("warehouse_id", warehouseID),
			zap.String("variant_id", variantID))
		return nil, err
	}

	response := &models.InventoryBatchResponse{
		ID:            batch.ID,
		WarehouseID:   batch.WarehouseID,
		VariantID:     batch.VariantID,
		CostPrice:     utils.RoundPrice(batch.CostPrice),
		ExpiryDate:    batch.ExpiryDate.Format("2006-01-02"),
		TotalQuantity: batch.TotalQuantity,
		CreatedAt:     batch.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     batch.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("Inventory batch created successfully",
		zap.String("batch_id", batch.ID),
		zap.String("warehouse_id", batch.WarehouseID),
		zap.String("variant_id", batch.VariantID),
		zap.Int64("quantity", batch.TotalQuantity))

	return response, nil
}

// GetBatch retrieves an inventory batch by ID
func (s *InventoryService) GetBatch(id string) (*models.InventoryBatchResponse, error) {
	s.logger.Info("Retrieving inventory batch",
		zap.String("batch_id", id))

	batch, err := s.inventoryRepo.GetBatchByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve inventory batch",
			zap.Error(err),
			zap.String("batch_id", id))
		return nil, err
	}

	// GST-only tax system - tax rates are on ProductVariant, not on batches
	response := &models.InventoryBatchResponse{
		ID:            batch.ID,
		WarehouseID:   batch.WarehouseID,
		VariantID:     batch.VariantID,
		CostPrice:     utils.RoundPrice(batch.CostPrice),
		ExpiryDate:    batch.ExpiryDate.Format("2006-01-02"),
		TotalQuantity: batch.TotalQuantity,
		CreatedAt:     batch.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     batch.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Debug("Inventory batch retrieved successfully",
		zap.String("batch_id", id),
		zap.Int64("quantity", batch.TotalQuantity))

	return response, nil
}

// GetBatchesByWarehouse retrieves all batches for a warehouse (paginated)
func (s *InventoryService) GetBatchesByWarehouse(warehouseID string, limit, offset int) ([]models.InventoryBatchResponse, int64, error) {
	s.logger.Info("Retrieving batches by warehouse",
		zap.String("warehouse_id", warehouseID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	// Validate warehouse exists
	s.logger.Debug("Validating warehouse exists")
	_, err := s.warehouseRepo.GetByID(warehouseID)
	if err != nil {
		s.logger.Error("Warehouse not found",
			zap.Error(err),
			zap.String("warehouse_id", warehouseID))
		return nil, 0, errors.NewNotFoundError("Warehouse")
	}

	batches, total, err := s.inventoryRepo.GetBatchesByWarehouse(warehouseID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to retrieve batches by warehouse",
			zap.Error(err),
			zap.String("warehouse_id", warehouseID))
		return nil, 0, err
	}

	var responses []models.InventoryBatchResponse
	for _, batch := range batches {
		response := s.batchToResponse(batch)
		responses = append(responses, response)
	}

	s.logger.Info("Batches retrieved successfully",
		zap.String("warehouse_id", warehouseID),
		zap.Int("count", len(responses)),
		zap.Int64("total", total))

	return responses, total, nil
}

// GetBatchesByVariant retrieves all batches for a product variant (paginated)
func (s *InventoryService) GetBatchesByVariant(variantID string, limit, offset int) ([]models.InventoryBatchResponse, int64, error) {
	s.logger.Info("Retrieving batches by variant",
		zap.String("variant_id", variantID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	// Validate variant exists
	s.logger.Debug("Validating variant exists")
	_, err := s.variantRepo.GetByID(variantID)
	if err != nil {
		s.logger.Error("Variant not found",
			zap.Error(err),
			zap.String("variant_id", variantID))
		return nil, 0, errors.NewNotFoundError("Variant")
	}

	batches, total, err := s.inventoryRepo.GetBatchesByVariant(variantID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to retrieve batches by variant",
			zap.Error(err),
			zap.String("variant_id", variantID))
		return nil, 0, err
	}

	var responses []models.InventoryBatchResponse
	for _, batch := range batches {
		response := s.batchToResponse(batch)
		responses = append(responses, response)
	}

	s.logger.Info("Batches retrieved successfully",
		zap.String("variant_id", variantID),
		zap.Int("count", len(responses)),
		zap.Int64("total", total))

	return responses, total, nil
}

// CreateTransaction creates a new inventory transaction
func (s *InventoryService) CreateTransaction(batchID string, request *models.CreateInventoryTransactionRequest) (*models.InventoryTransactionResponse, error) {
	s.logger.Info("Creating inventory transaction",
		zap.String("batch_id", batchID),
		zap.String("transaction_type", request.TransactionType),
		zap.Int64("quantity_change", request.QuantityChange))

	// Validate batch exists
	s.logger.Debug("Validating batch exists",
		zap.String("batch_id", batchID))
	if _, err := s.inventoryRepo.GetBatchByID(batchID); err != nil {
		s.logger.Error("Batch not found",
			zap.Error(err),
			zap.String("batch_id", batchID))
		return nil, errors.NewNotFoundError("Inventory batch")
	}

	// Validate transaction type
	s.logger.Debug("Validating transaction type",
		zap.String("transaction_type", request.TransactionType))
	validTypes := []string{"import", "manual_add", "adjustment", "sale_deduction", "return_add", "transfer_in", "transfer_out", "cancellation_return"}
	isValidType := false
	for _, t := range validTypes {
		if t == request.TransactionType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		s.logger.Error("Invalid transaction type",
			zap.String("transaction_type", request.TransactionType))
		return nil, errors.NewBadRequestError("Invalid transaction type")
	}

	// Create transaction using the proper constructor
	transaction := models.NewInventoryTransaction(batchID, request.TransactionType, request.QuantityChange, request.RelatedEntityID, nil, request.Note, time.Now())

	s.logger.Debug("Saving transaction to database")
	if err := s.inventoryRepo.CreateTransaction(transaction); err != nil {
		s.logger.Error("Failed to create transaction",
			zap.Error(err),
			zap.String("batch_id", batchID))
		return nil, err
	}

	// Update batch stock level
	s.logger.Debug("Updating batch stock level",
		zap.Int64("quantity_change", request.QuantityChange))
	if err := s.inventoryRepo.UpdateBatchStock(batchID, request.QuantityChange); err != nil {
		s.logger.Error("Failed to update batch stock",
			zap.Error(err),
			zap.String("batch_id", batchID))
		return nil, err
	}

	response := &models.InventoryTransactionResponse{
		ID:              transaction.ID,
		BatchID:         transaction.BatchID,
		TransactionType: transaction.TransactionType,
		QuantityChange:  transaction.QuantityChange,
		RelatedEntityID: transaction.RelatedEntityID,
		Note:            transaction.Note,
		OccurredAt:      transaction.OccurredAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("Inventory transaction created successfully",
		zap.String("transaction_id", transaction.ID),
		zap.String("batch_id", batchID),
		zap.String("transaction_type", transaction.TransactionType))

	return response, nil
}

// GetTransactionsByBatch retrieves all transactions for a batch (paginated)
func (s *InventoryService) GetTransactionsByBatch(batchID string, limit, offset int) ([]models.InventoryTransactionResponse, int64, error) {
	s.logger.Info("Retrieving transactions by batch",
		zap.String("batch_id", batchID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	// Validate batch exists
	s.logger.Debug("Validating batch exists")
	if _, err := s.inventoryRepo.GetBatchByID(batchID); err != nil {
		s.logger.Error("Batch not found",
			zap.Error(err),
			zap.String("batch_id", batchID))
		return nil, 0, errors.NewNotFoundError("Inventory batch")
	}

	transactions, total, err := s.inventoryRepo.GetTransactionsByBatch(batchID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to retrieve transactions",
			zap.Error(err),
			zap.String("batch_id", batchID))
		return nil, 0, err
	}

	var responses []models.InventoryTransactionResponse
	for _, transaction := range transactions {
		response := models.InventoryTransactionResponse{
			ID:              transaction.ID,
			BatchID:         transaction.BatchID,
			TransactionType: transaction.TransactionType,
			QuantityChange:  transaction.QuantityChange,
			RelatedEntityID: transaction.RelatedEntityID,
			Note:            transaction.Note,
			OccurredAt:      transaction.OccurredAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	s.logger.Info("Transactions retrieved successfully",
		zap.String("batch_id", batchID),
		zap.Int("count", len(responses)),
		zap.Int64("total", total))

	return responses, total, nil
}

// GetExpiringBatches retrieves batches that expire within a given timeframe (paginated)
func (s *InventoryService) GetExpiringBatches(days int, limit, offset int) ([]models.InventoryBatchResponse, int64, error) {
	s.logger.Info("Retrieving expiring batches",
		zap.Int("days", days),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	batches, total, err := s.inventoryRepo.GetExpiringBatches(days, limit, offset)
	if err != nil {
		s.logger.Error("Failed to retrieve expiring batches",
			zap.Error(err),
			zap.Int("days", days))
		return nil, 0, err
	}

	var responses []models.InventoryBatchResponse
	for _, batch := range batches {
		response := s.batchToResponse(batch)
		responses = append(responses, response)
	}

	s.logger.Info("Expiring batches retrieved successfully",
		zap.Int("days", days),
		zap.Int("count", len(responses)),
		zap.Int64("total", total))

	return responses, total, nil
}

// GetLowStockBatches retrieves batches with low stock (paginated)
func (s *InventoryService) GetLowStockBatches(threshold int64, limit, offset int) ([]models.InventoryBatchResponse, int64, error) {
	s.logger.Info("Retrieving low stock batches",
		zap.Int64("threshold", threshold),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	batches, total, err := s.inventoryRepo.GetLowStockBatches(threshold, limit, offset)
	if err != nil {
		s.logger.Error("Failed to retrieve low stock batches",
			zap.Error(err),
			zap.Int64("threshold", threshold))
		return nil, 0, err
	}

	var responses []models.InventoryBatchResponse
	for _, batch := range batches {
		response := s.batchToResponse(batch)
		responses = append(responses, response)
	}

	s.logger.Info("Low stock batches retrieved successfully",
		zap.Int64("threshold", threshold),
		zap.Int("count", len(responses)),
		zap.Int64("total", total))

	return responses, total, nil
}

// GetAllProductsAvailability retrieves all products available across all warehouses grouped by SKU
// Returns aggregated availability data with per-warehouse breakdown
// Only includes non-expired batches in availability counts, but shows expired batches separately
func (s *InventoryService) GetAllProductsAvailability(ctx context.Context, jwtToken string, limit, offset int) ([]models.ProductAvailabilityGroupedResponse, int64, error) {
	s.logger.Info("Retrieving all products availability across warehouses (grouped by SKU)",
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	// Get all batches including expired ones for full visibility
	batches, total, err := s.inventoryRepo.GetAllBatchesPaginatedWithExpired(limit, offset)
	if err != nil {
		s.logger.Error("Failed to retrieve all batches",
			zap.Error(err))
		return nil, 0, err
	}

	s.logger.Debug("Processing batches for grouped availability response",
		zap.Int("batch_count", len(batches)))

	// Group batches by variant (SKU)
	variantMap := make(map[string]*variantAvailability)

	for _, batch := range batches {
		sku := ""
		if batch.Variant.SKU != nil {
			sku = *batch.Variant.SKU
		}

		// Skip batches without SKU
		if sku == "" {
			s.logger.Warn("Batch has no SKU, skipping",
				zap.String("batch_id", batch.ID),
				zap.String("variant_id", batch.VariantID))
			continue
		}

		// Initialize variant entry if not exists
		if _, exists := variantMap[sku]; !exists {
			gstRate := batch.Variant.GSTRate
			// Parse images from JSON string (Issue 8)
			var images []string
			if batch.Variant.Images != nil && *batch.Variant.Images != "" {
				if err := json.Unmarshal([]byte(*batch.Variant.Images), &images); err != nil {
					s.logger.Warn("Failed to parse variant images",
						zap.Error(err),
						zap.String("variant_id", batch.VariantID))
				}
			}
			variantMap[sku] = &variantAvailability{
				VariantID:          batch.VariantID,
				ProductName:        batch.Variant.VariantName,
				ProductDescription: batch.Variant.Description,
				WarehouseDetails:   make(map[string]*warehouseDetail),
				// GST Details
				HSNCode:  batch.Variant.HSNCode,
				GSTRate:  gstRate,
				CGSTRate: gstRate / 2,
				SGSTRate: gstRate / 2,
				// Images (Issue 8)
				Images: images,
			}
		}

		variantEntry := variantMap[sku]

		// Calculate expiry status and determine if batch is expired
		isExpired := batch.ExpiryDate.Before(time.Now())
		expiryStatus := s.calculateExpiryStatus(batch.ExpiryDate)

		// Initialize warehouse entry if not exists
		warehouseKey := batch.WarehouseID
		if _, exists := variantEntry.WarehouseDetails[warehouseKey]; !exists {
			variantEntry.WarehouseDetails[warehouseKey] = &warehouseDetail{
				WarehouseID:   batch.WarehouseID,
				WarehouseName: batch.Warehouse.Name,
				AddressID:     batch.Warehouse.AddressID,
				Address:       batch.Warehouse.BuildAddressInfo(), // Uses local fields - NO gRPC call!
			}
		}

		warehouseEntry := variantEntry.WarehouseDetails[warehouseKey]

		// Add quantities based on expiry status
		if isExpired {
			warehouseEntry.ExpiredQuantity += batch.TotalQuantity
		} else {
			warehouseEntry.Quantity += batch.TotalQuantity
		}

		// Track earliest expiry for this warehouse (only non-expired)
		if !isExpired && (warehouseEntry.EarliestExpiry.IsZero() || batch.ExpiryDate.Before(warehouseEntry.EarliestExpiry)) {
			warehouseEntry.EarliestExpiry = batch.ExpiryDate
			warehouseEntry.ExpiryStatus = expiryStatus
		}
	}

	// Convert map to response array
	var responses []models.ProductAvailabilityGroupedResponse
	for sku, variantData := range variantMap {
		response := models.ProductAvailabilityGroupedResponse{
			SKU:                sku,
			VariantID:          variantData.VariantID,
			ProductName:        variantData.ProductName,
			ProductDescription: variantData.ProductDescription,
			WarehouseDetails:   []models.WarehouseAvailabilityDetail{},
			// GST Details
			HSNCode:  variantData.HSNCode,
			GSTRate:  variantData.GSTRate,
			CGSTRate: variantData.CGSTRate,
			SGSTRate: variantData.SGSTRate,
			// Images (Issue 8)
			Images: variantData.Images,
		}

		// Process warehouse details
		var earliestExpiry time.Time
		for _, warehouseData := range variantData.WarehouseDetails {
			detail := models.WarehouseAvailabilityDetail{
				WarehouseID:     warehouseData.WarehouseID,
				WarehouseName:   warehouseData.WarehouseName,
				Quantity:        warehouseData.Quantity,
				ExpiredQuantity: warehouseData.ExpiredQuantity,
				ExpiryStatus:    warehouseData.ExpiryStatus,
			}

			if !warehouseData.EarliestExpiry.IsZero() {
				detail.EarliestExpiry = warehouseData.EarliestExpiry.Format("2006-01-02")

				// Track overall earliest expiry
				if earliestExpiry.IsZero() || warehouseData.EarliestExpiry.Before(earliestExpiry) {
					earliestExpiry = warehouseData.EarliestExpiry
				}
			}

			// Use pre-populated address from local cache (NO gRPC call!)
			detail.Address = warehouseData.Address

			response.TotalQuantity += warehouseData.Quantity
			response.ExpiredQuantity += warehouseData.ExpiredQuantity
			response.WarehouseDetails = append(response.WarehouseDetails, detail)
		}

		// Set overall earliest expiry and status
		if !earliestExpiry.IsZero() {
			response.EarliestExpiry = earliestExpiry.Format("2006-01-02")
			response.ExpiryStatus = s.calculateExpiryStatus(earliestExpiry)
		} else if response.ExpiredQuantity > 0 {
			// All batches are expired
			response.ExpiryStatus = "expired"
		}

		// Sort warehouse details by earliest expiry (FEFO - First Expired First Out)
		s.sortWarehouseDetailsByExpiry(response.WarehouseDetails)

		// Fetch active prices for this variant
		if s.priceRepo != nil {
			prices, err := s.priceRepo.GetActiveByVariantID(variantData.VariantID)
			if err != nil {
				s.logger.Error("Failed to fetch prices for variant",
					zap.Error(err),
					zap.String("variant_id", variantData.VariantID),
					zap.String("sku", sku))
				// Don't fail the entire request, just log and continue without prices
			} else {
				var priceResponses []models.ProductPriceResponse
				for _, price := range prices {
					isActive := price.IsActive != nil && *price.IsActive
					priceResponse := models.ProductPriceResponse{
						ID:            price.ID,
						VariantID:     price.VariantID,
						PriceType:     price.PriceType,
						Price:         utils.RoundPrice(price.Price),
						Currency:      price.Currency,
						EffectiveFrom: price.EffectiveFrom.Format("2006-01-02T15:04:05Z"),
						IsActive:      isActive,
						CreatedAt:     price.CreatedAt.Format("2006-01-02T15:04:05Z"),
						UpdatedAt:     price.UpdatedAt.Format("2006-01-02T15:04:05Z"),
					}
					if price.EffectiveTo != nil {
						effectiveTo := price.EffectiveTo.Format("2006-01-02T15:04:05Z")
						priceResponse.EffectiveTo = &effectiveTo
					}
					priceResponses = append(priceResponses, priceResponse)
				}
				response.Prices = priceResponses
			}
		}

		responses = append(responses, response)
	}

	s.logger.Info("All products availability retrieved and grouped successfully",
		zap.Int("unique_products", len(responses)),
		zap.Int64("total_batches", total))

	return responses, total, nil
}

// Helper types for grouping logic
type variantAvailability struct {
	VariantID          string
	ProductName        string
	ProductDescription *string
	WarehouseDetails   map[string]*warehouseDetail
	// GST Details
	HSNCode  string
	GSTRate  float64
	CGSTRate float64
	SGSTRate float64
	// Images (Issue 8)
	Images []string
}

type warehouseDetail struct {
	WarehouseID     string
	WarehouseName   string
	AddressID       *string
	Address         *models.AddressInfo // Local cache - NO gRPC needed
	Quantity        int64
	ExpiredQuantity int64
	EarliestExpiry  time.Time
	ExpiryStatus    string
}

// calculateExpiryStatus determines expiry status based on expiry date
// Returns: "fresh", "expiring_soon" (within 30 days), or "expired"
func (s *InventoryService) calculateExpiryStatus(expiryDate time.Time) string {
	now := time.Now()

	if expiryDate.Before(now) {
		return "expired"
	}

	daysUntilExpiry := int(expiryDate.Sub(now).Hours() / 24)

	if daysUntilExpiry <= 30 {
		return "expiring_soon"
	}

	return "fresh"
}

// sortWarehouseDetailsByExpiry sorts warehouse details by earliest expiry (FEFO logic)
func (s *InventoryService) sortWarehouseDetailsByExpiry(details []models.WarehouseAvailabilityDetail) {
	// Simple bubble sort for small arrays (warehouse count is typically small)
	for i := 0; i < len(details)-1; i++ {
		for j := 0; j < len(details)-i-1; j++ {
			// Compare expiry dates (empty expiry goes to end)
			if details[j].EarliestExpiry == "" {
				continue
			}
			if details[j+1].EarliestExpiry == "" {
				// Swap if next is empty
				details[j], details[j+1] = details[j+1], details[j]
				continue
			}

			// Parse and compare dates
			expiry1, err1 := time.Parse("2006-01-02", details[j].EarliestExpiry)
			expiry2, err2 := time.Parse("2006-01-02", details[j+1].EarliestExpiry)

			if err1 == nil && err2 == nil && expiry1.After(expiry2) {
				details[j], details[j+1] = details[j+1], details[j]
			}
		}
	}
}

// batchToResponse converts a batch model to response model
// GST-only tax system - tax rates are on ProductVariant, not on batches
func (s *InventoryService) batchToResponse(batch models.InventoryBatch) models.InventoryBatchResponse {
	// Calculate expiry status
	isExpired := batch.ExpiryDate.Before(time.Now())
	expiryStatus := s.calculateExpiryStatus(batch.ExpiryDate)

	return models.InventoryBatchResponse{
		ID:                batch.ID,
		WarehouseID:       batch.WarehouseID,
		VariantID:         batch.VariantID,
		CostPrice:         utils.RoundPrice(batch.CostPrice),
		ExpiryDate:        batch.ExpiryDate.Format("2006-01-02"),
		TotalQuantity:     batch.TotalQuantity,
		ReservedQuantity:  batch.ReservedQuantity,
		AvailableQuantity: batch.AvailableQuantity(), // Total - Reserved
		IsExpired:         isExpired,
		ExpiryStatus:      expiryStatus,
		CreatedAt:         batch.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         batch.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
