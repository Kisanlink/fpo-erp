package services

import (
	"context"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
)

// InventoryService handles inventory business logic
type InventoryService struct {
	inventoryRepo *repositories.InventoryRepository
	warehouseRepo *repositories.WarehouseRepository
	productRepo   *repositories.ProductRepository
	addressClient *aaa.AddressClient
}

// NewInventoryService creates a new inventory service
func NewInventoryService(inventoryRepo *repositories.InventoryRepository, warehouseRepo *repositories.WarehouseRepository, productRepo *repositories.ProductRepository, addressClient *aaa.AddressClient) *InventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
		warehouseRepo: warehouseRepo,
		productRepo:   productRepo,
		addressClient: addressClient,
	}
}

// CreateBatch creates a new inventory batch with tax configuration
func (s *InventoryService) CreateBatch(warehouseID, productID string, costPrice float64, expiryDate time.Time, quantity int64, cgstRate, sgstRate float64, customTaxIDs []string, isTaxExempt bool) (*models.InventoryBatchResponse, error) {
	// Validate warehouse exists
	_, err := s.warehouseRepo.GetByID(warehouseID)
	if err != nil {
		return nil, errors.NewNotFoundError("Warehouse")
	}

	// Validate product exists
	_, err = s.productRepo.GetByID(productID)
	if err != nil {
		return nil, errors.NewNotFoundError("Product")
	}

	// Validate expiry date is in the future
	if expiryDate.Before(time.Now()) {
		return nil, errors.NewBadRequestError("Expiry date must be in the future")
	}

	// Validate quantity is positive
	if quantity <= 0 {
		return nil, errors.NewBadRequestError("Quantity must be positive")
	}

	// Validate tax rates
	if cgstRate < 0 || cgstRate > 100 {
		return nil, errors.NewBadRequestError("CGST rate must be between 0 and 100")
	}
	if sgstRate < 0 || sgstRate > 100 {
		return nil, errors.NewBadRequestError("SGST rate must be between 0 and 100")
	}

	// Create batch using the updated constructor
	batch := models.NewInventoryBatch(warehouseID, productID, costPrice, expiryDate, quantity, cgstRate, sgstRate, customTaxIDs, isTaxExempt)

	// Create initial transaction using the proper constructor
	note := "Initial import"
	transaction := models.NewInventoryTransaction("", "import", quantity, nil, nil, &note, time.Now())

	// Create batch and initial transaction atomically
	if err := s.inventoryRepo.CreateBatchWithTransaction(batch, transaction); err != nil {
		return nil, err
	}

	response := &models.InventoryBatchResponse{
		ID:            batch.ID,
		WarehouseID:   batch.WarehouseID,
		ProductID:     batch.ProductID,
		CostPrice:     batch.CostPrice,
		ExpiryDate:    batch.ExpiryDate.Format("2006-01-02"),
		TotalQuantity: batch.TotalQuantity,
		CGSTRate:      batch.CGSTRate,
		SGSTRate:      batch.SGSTRate,
		CustomTaxIDs:  batch.CustomTaxIDs,
		IsTaxExempt:   batch.IsTaxExempt,
		CreatedAt:     batch.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     batch.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return response, nil
}

// GetBatch retrieves an inventory batch by ID
func (s *InventoryService) GetBatch(id string) (*models.InventoryBatchResponse, error) {
	batch, err := s.inventoryRepo.GetBatchByID(id)
	if err != nil {
		return nil, err
	}

	response := &models.InventoryBatchResponse{
		ID:            batch.ID,
		WarehouseID:   batch.WarehouseID,
		ProductID:     batch.ProductID,
		CostPrice:     batch.CostPrice,
		ExpiryDate:    batch.ExpiryDate.Format("2006-01-02"),
		TotalQuantity: batch.TotalQuantity,
		CGSTRate:      batch.CGSTRate,
		SGSTRate:      batch.SGSTRate,
		CustomTaxIDs:  batch.CustomTaxIDs,
		IsTaxExempt:   batch.IsTaxExempt,
		CreatedAt:     batch.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     batch.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return response, nil
}

// GetBatchesByWarehouse retrieves all batches for a warehouse
func (s *InventoryService) GetBatchesByWarehouse(warehouseID string) ([]models.InventoryBatchResponse, error) {
	// Validate warehouse exists
	_, err := s.warehouseRepo.GetByID(warehouseID)
	if err != nil {
		return nil, errors.NewNotFoundError("Warehouse")
	}

	batches, err := s.inventoryRepo.GetBatchesByWarehouse(warehouseID)
	if err != nil {
		return nil, err
	}

	var responses []models.InventoryBatchResponse
	for _, batch := range batches {
		response := s.batchToResponse(batch)
		responses = append(responses, response)
	}

	return responses, nil
}

// GetBatchesByProduct retrieves all batches for a product
func (s *InventoryService) GetBatchesByProduct(productID string) ([]models.InventoryBatchResponse, error) {
	// Validate product exists
	_, err := s.productRepo.GetByID(productID)
	if err != nil {
		return nil, errors.NewNotFoundError("Product")
	}

	batches, err := s.inventoryRepo.GetBatchesByProduct(productID)
	if err != nil {
		return nil, err
	}

	var responses []models.InventoryBatchResponse
	for _, batch := range batches {
		response := s.batchToResponse(batch)
		responses = append(responses, response)
	}

	return responses, nil
}

// CreateTransaction creates a new inventory transaction
func (s *InventoryService) CreateTransaction(batchID string, request *models.CreateInventoryTransactionRequest) (*models.InventoryTransactionResponse, error) {
	// Validate batch exists
	if _, err := s.inventoryRepo.GetBatchByID(batchID); err != nil {
		return nil, errors.NewNotFoundError("Inventory batch")
	}

	// Validate transaction type
	validTypes := []string{"import", "manual_add", "adjustment", "sale_deduction", "return_add", "transfer_in", "transfer_out"}
	isValidType := false
	for _, t := range validTypes {
		if t == request.TransactionType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return nil, errors.NewBadRequestError("Invalid transaction type")
	}

	// Create transaction using the proper constructor
	transaction := models.NewInventoryTransaction(batchID, request.TransactionType, request.QuantityChange, request.RelatedEntityID, nil, request.Note, time.Now())

	if err := s.inventoryRepo.CreateTransaction(transaction); err != nil {
		return nil, err
	}

	// Update batch stock level
	if err := s.inventoryRepo.UpdateBatchStock(batchID, request.QuantityChange); err != nil {
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

	return response, nil
}

// GetTransactionsByBatch retrieves all transactions for a batch
func (s *InventoryService) GetTransactionsByBatch(batchID string) ([]models.InventoryTransactionResponse, error) {
	// Validate batch exists
	if _, err := s.inventoryRepo.GetBatchByID(batchID); err != nil {
		return nil, errors.NewNotFoundError("Inventory batch")
	}

	transactions, err := s.inventoryRepo.GetTransactionsByBatch(batchID)
	if err != nil {
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

	return responses, nil
}

// GetExpiringBatches retrieves batches that expire within a given timeframe
func (s *InventoryService) GetExpiringBatches(days int) ([]models.InventoryBatchResponse, error) {
	batches, err := s.inventoryRepo.GetExpiringBatches(days)
	if err != nil {
		return nil, err
	}

	var responses []models.InventoryBatchResponse
	for _, batch := range batches {
		response := s.batchToResponse(batch)
		responses = append(responses, response)
	}

	return responses, nil
}

// GetLowStockBatches retrieves batches with low stock
func (s *InventoryService) GetLowStockBatches(threshold int64) ([]models.InventoryBatchResponse, error) {
	batches, err := s.inventoryRepo.GetLowStockBatches(threshold)
	if err != nil {
		return nil, err
	}

	var responses []models.InventoryBatchResponse
	for _, batch := range batches {
		response := s.batchToResponse(batch)
		responses = append(responses, response)
	}

	return responses, nil
}

// GetAllProductsAvailability retrieves all products available across all warehouses
func (s *InventoryService) GetAllProductsAvailability(ctx context.Context) ([]models.ProductAvailabilityResponse, error) {
	batches, err := s.inventoryRepo.GetAllBatches()
	if err != nil {
		return nil, err
	}

	var responses []models.ProductAvailabilityResponse
	for _, batch := range batches {
		response := models.ProductAvailabilityResponse{
			ID:                 batch.ID,
			WarehouseID:        batch.WarehouseID,
			WarehouseName:      batch.Warehouse.Name,
			ProductID:          batch.ProductID,
			ProductSKU:         batch.Product.SKU,
			ProductName:        batch.Product.Name,
			ProductDescription: batch.Product.Description,
			CostPrice:          batch.CostPrice,
			ExpiryDate:         batch.ExpiryDate.Format("2006-01-02"),
			TotalQuantity:      batch.TotalQuantity,
			CGSTRate:           batch.CGSTRate,
			SGSTRate:           batch.SGSTRate,
			CustomTaxIDs:       batch.CustomTaxIDs,
			IsTaxExempt:        batch.IsTaxExempt,
			CreatedAt:          batch.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:          batch.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}

		// Fetch address details if warehouse has an address ID
		if batch.Warehouse.AddressID != nil {
			address, err := s.addressClient.GetAddress(ctx, *batch.Warehouse.AddressID)
			if err == nil {
				response.Address = &models.AddressInfo{
					ID:           address.ID,
					Type:         address.Type,
					AddressLine1: address.AddressLine1,
					AddressLine2: address.AddressLine2,
					City:         address.City,
					State:        address.State,
					PostalCode:   address.PostalCode,
					Country:      address.Country,
					FullAddress:  s.buildFullAddress(address),
				}
			}
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// batchToResponse converts a batch model to response model
func (s *InventoryService) batchToResponse(batch models.InventoryBatch) models.InventoryBatchResponse {
	return models.InventoryBatchResponse{
		ID:            batch.ID,
		WarehouseID:   batch.WarehouseID,
		ProductID:     batch.ProductID,
		CostPrice:     batch.CostPrice,
		ExpiryDate:    batch.ExpiryDate.Format("2006-01-02"),
		TotalQuantity: batch.TotalQuantity,
		CGSTRate:      batch.CGSTRate,
		SGSTRate:      batch.SGSTRate,
		CustomTaxIDs:  batch.CustomTaxIDs,
		IsTaxExempt:   batch.IsTaxExempt,
		CreatedAt:     batch.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     batch.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// buildFullAddress builds a full address string from address components
func (s *InventoryService) buildFullAddress(address *aaa.Address) string {
	parts := []string{}
	if address.AddressLine1 != "" {
		parts = append(parts, address.AddressLine1)
	}
	if address.AddressLine2 != "" {
		parts = append(parts, address.AddressLine2)
	}
	if address.City != "" {
		parts = append(parts, address.City)
	}
	if address.State != "" {
		parts = append(parts, address.State)
	}
	if address.PostalCode != "" {
		parts = append(parts, address.PostalCode)
	}
	if address.Country != "" {
		parts = append(parts, address.Country)
	}

	fullAddress := ""
	for i, part := range parts {
		if i > 0 {
			fullAddress += ", "
		}
		fullAddress += part
	}
	return fullAddress
}
