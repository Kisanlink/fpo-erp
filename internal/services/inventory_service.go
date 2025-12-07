package services

import (
	"context"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"

	"go.uber.org/zap"
)

// InventoryService handles inventory business logic
type InventoryService struct {
	inventoryRepo *repositories.InventoryRepository
	warehouseRepo *repositories.WarehouseRepository
	productRepo   *repositories.ProductRepository
	variantRepo   *repositories.ProductVariantRepository
	addressClient *aaa.AddressGRPCClient
	logger        interfaces.Logger
}

// NewInventoryService creates a new inventory service
func NewInventoryService(inventoryRepo *repositories.InventoryRepository, warehouseRepo *repositories.WarehouseRepository, productRepo *repositories.ProductRepository, variantRepo *repositories.ProductVariantRepository, addressClient *aaa.AddressGRPCClient, logger interfaces.Logger) *InventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
		warehouseRepo: warehouseRepo,
		productRepo:   productRepo,
		variantRepo:   variantRepo,
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
		CostPrice:     batch.CostPrice,
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
		CostPrice:     batch.CostPrice,
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

// GetBatchesByWarehouse retrieves all batches for a warehouse
func (s *InventoryService) GetBatchesByWarehouse(warehouseID string) ([]models.InventoryBatchResponse, error) {
	s.logger.Info("Retrieving batches by warehouse",
		zap.String("warehouse_id", warehouseID))

	// Validate warehouse exists
	s.logger.Debug("Validating warehouse exists")
	_, err := s.warehouseRepo.GetByID(warehouseID)
	if err != nil {
		s.logger.Error("Warehouse not found",
			zap.Error(err),
			zap.String("warehouse_id", warehouseID))
		return nil, errors.NewNotFoundError("Warehouse")
	}

	batches, err := s.inventoryRepo.GetBatchesByWarehouse(warehouseID)
	if err != nil {
		s.logger.Error("Failed to retrieve batches by warehouse",
			zap.Error(err),
			zap.String("warehouse_id", warehouseID))
		return nil, err
	}

	var responses []models.InventoryBatchResponse
	for _, batch := range batches {
		response := s.batchToResponse(batch)
		responses = append(responses, response)
	}

	s.logger.Info("Batches retrieved successfully",
		zap.String("warehouse_id", warehouseID),
		zap.Int("count", len(responses)))

	return responses, nil
}

// GetBatchesByVariant retrieves all batches for a product variant
func (s *InventoryService) GetBatchesByVariant(variantID string) ([]models.InventoryBatchResponse, error) {
	s.logger.Info("Retrieving batches by variant",
		zap.String("variant_id", variantID))

	// Validate variant exists
	s.logger.Debug("Validating variant exists")
	_, err := s.variantRepo.GetByID(variantID)
	if err != nil {
		s.logger.Error("Variant not found",
			zap.Error(err),
			zap.String("variant_id", variantID))
		return nil, errors.NewNotFoundError("Variant")
	}

	batches, err := s.inventoryRepo.GetBatchesByVariant(variantID)
	if err != nil {
		s.logger.Error("Failed to retrieve batches by variant",
			zap.Error(err),
			zap.String("variant_id", variantID))
		return nil, err
	}

	var responses []models.InventoryBatchResponse
	for _, batch := range batches {
		response := s.batchToResponse(batch)
		responses = append(responses, response)
	}

	s.logger.Info("Batches retrieved successfully",
		zap.String("variant_id", variantID),
		zap.Int("count", len(responses)))

	return responses, nil
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

// GetTransactionsByBatch retrieves all transactions for a batch
func (s *InventoryService) GetTransactionsByBatch(batchID string) ([]models.InventoryTransactionResponse, error) {
	s.logger.Info("Retrieving transactions by batch",
		zap.String("batch_id", batchID))

	// Validate batch exists
	s.logger.Debug("Validating batch exists")
	if _, err := s.inventoryRepo.GetBatchByID(batchID); err != nil {
		s.logger.Error("Batch not found",
			zap.Error(err),
			zap.String("batch_id", batchID))
		return nil, errors.NewNotFoundError("Inventory batch")
	}

	transactions, err := s.inventoryRepo.GetTransactionsByBatch(batchID)
	if err != nil {
		s.logger.Error("Failed to retrieve transactions",
			zap.Error(err),
			zap.String("batch_id", batchID))
		return nil, err
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
		zap.Int("count", len(responses)))

	return responses, nil
}

// GetExpiringBatches retrieves batches that expire within a given timeframe
func (s *InventoryService) GetExpiringBatches(days int) ([]models.InventoryBatchResponse, error) {
	s.logger.Info("Retrieving expiring batches",
		zap.Int("days", days))

	batches, err := s.inventoryRepo.GetExpiringBatches(days)
	if err != nil {
		s.logger.Error("Failed to retrieve expiring batches",
			zap.Error(err),
			zap.Int("days", days))
		return nil, err
	}

	var responses []models.InventoryBatchResponse
	for _, batch := range batches {
		response := s.batchToResponse(batch)
		responses = append(responses, response)
	}

	s.logger.Info("Expiring batches retrieved successfully",
		zap.Int("days", days),
		zap.Int("count", len(responses)))

	return responses, nil
}

// GetLowStockBatches retrieves batches with low stock
func (s *InventoryService) GetLowStockBatches(threshold int64) ([]models.InventoryBatchResponse, error) {
	s.logger.Info("Retrieving low stock batches",
		zap.Int64("threshold", threshold))

	batches, err := s.inventoryRepo.GetLowStockBatches(threshold)
	if err != nil {
		s.logger.Error("Failed to retrieve low stock batches",
			zap.Error(err),
			zap.Int64("threshold", threshold))
		return nil, err
	}

	var responses []models.InventoryBatchResponse
	for _, batch := range batches {
		response := s.batchToResponse(batch)
		responses = append(responses, response)
	}

	s.logger.Info("Low stock batches retrieved successfully",
		zap.Int64("threshold", threshold),
		zap.Int("count", len(responses)))

	return responses, nil
}

// GetAllProductsAvailability retrieves all products available across all warehouses
func (s *InventoryService) GetAllProductsAvailability(ctx context.Context, jwtToken string) ([]models.ProductAvailabilityResponse, error) {
	s.logger.Info("Retrieving all products availability across warehouses")

	batches, err := s.inventoryRepo.GetAllBatches()
	if err != nil {
		s.logger.Error("Failed to retrieve all batches",
			zap.Error(err))
		return nil, err
	}

	s.logger.Debug("Processing batches for availability response",
		zap.Int("batch_count", len(batches)))

	var responses []models.ProductAvailabilityResponse
	for _, batch := range batches {
		response := models.ProductAvailabilityResponse{
			ID:            batch.ID,
			WarehouseID:   batch.WarehouseID,
			WarehouseName: batch.Warehouse.Name,
			VariantID:     batch.VariantID,
			ProductSKU: func() string {
				if batch.Variant.SKU != nil {
					return *batch.Variant.SKU
				}
				return ""
			}(),
			ProductName:        batch.Variant.VariantName,
			ProductDescription: batch.Variant.Description,
			CostPrice:          batch.CostPrice,
			ExpiryDate:         batch.ExpiryDate.Format("2006-01-02"),
			TotalQuantity:      batch.TotalQuantity,
			// GST-only tax system - tax rates are on ProductVariant, not on batches
			CreatedAt: batch.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: batch.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}

		// Fetch address details if warehouse has an address ID
		if batch.Warehouse.AddressID != nil {
			s.logger.Debug("Fetching warehouse address",
				zap.String("warehouse_id", batch.WarehouseID),
				zap.String("address_id", *batch.Warehouse.AddressID))
			address, err := s.addressClient.GetAddress(ctx, *batch.Warehouse.AddressID, jwtToken)
			if err == nil {
				response.Address = &models.AddressInfo{
					ID:          address.ID,
					Type:        address.Type,
					House:       address.House,
					Street:      address.Street,
					Landmark:    address.Landmark,
					PostOffice:  address.PostOffice,
					Subdistrict: address.Subdistrict,
					District:    address.District,
					VTC:         address.VTC,
					State:       address.State,
					Country:     address.Country,
					Pincode:     address.Pincode,
					FullAddress: address.BuildFullAddress(),
				}
			} else {
				s.logger.Error("Failed to fetch warehouse address",
					zap.Error(err),
					zap.String("warehouse_id", batch.WarehouseID),
					zap.String("address_id", *batch.Warehouse.AddressID))
			}
		}

		responses = append(responses, response)
	}

	s.logger.Info("All products availability retrieved successfully",
		zap.Int("count", len(responses)))

	return responses, nil
}

// batchToResponse converts a batch model to response model
// GST-only tax system - tax rates are on ProductVariant, not on batches
func (s *InventoryService) batchToResponse(batch models.InventoryBatch) models.InventoryBatchResponse {
	return models.InventoryBatchResponse{
		ID:            batch.ID,
		WarehouseID:   batch.WarehouseID,
		VariantID:     batch.VariantID,
		CostPrice:     batch.CostPrice,
		ExpiryDate:    batch.ExpiryDate.Format("2006-01-02"),
		TotalQuantity: batch.TotalQuantity,
		CreatedAt:     batch.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     batch.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
