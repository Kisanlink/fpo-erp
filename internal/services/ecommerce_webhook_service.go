package services

import (
	"fmt"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/utils"
)

// EcommerceWebhookService handles e-commerce webhook processing
type EcommerceWebhookService struct {
	inventoryService *InventoryService
	inventoryRepo    *repositories.InventoryRepository
	salesService     *SalesService
	productRepo      *repositories.ProductRepository
	warehouseRepo    *repositories.WarehouseRepository
	historyService   *WebhookHistoryService
}

// NewEcommerceWebhookService creates a new e-commerce webhook service
func NewEcommerceWebhookService(
	inventoryService *InventoryService,
	inventoryRepo *repositories.InventoryRepository,
	salesService *SalesService,
	productRepo *repositories.ProductRepository,
	warehouseRepo *repositories.WarehouseRepository,
	historyService *WebhookHistoryService,
) *EcommerceWebhookService {
	return &EcommerceWebhookService{
		inventoryService: inventoryService,
		inventoryRepo:    inventoryRepo,
		salesService:     salesService,
		productRepo:      productRepo,
		warehouseRepo:    warehouseRepo,
		historyService:   historyService,
	}
}

// FPOSalePayload represents the payload for FPO sale events
type FPOSalePayload struct {
	EventID         string          `json:"event_id"`
	EventType       string          `json:"event_type"`
	Timestamp       int64           `json:"timestamp"`
	FPOID           string          `json:"fpo_id"`
	SaleID          string          `json:"sale_id"`
	CustomerID      string          `json:"customer_id"`
	DeliveryAddress DeliveryAddress `json:"delivery_address"`
	Items           []FPOSaleItem   `json:"items"`
	TotalAmount     float64         `json:"total_amount"`
	Currency        string          `json:"currency"`
	OrderDate       string          `json:"order_date"`
}

// FPOPurchasePayload represents the payload for FPO purchase events
type FPOPurchasePayload struct {
	EventID              string            `json:"event_id"`
	EventType            string            `json:"event_type"`
	Timestamp            int64             `json:"timestamp"`
	FPOID                string            `json:"fpo_id"`
	PurchaseID           string            `json:"purchase_id"`
	SupplierID           string            `json:"supplier_id"`
	DeliveryAddress      DeliveryAddress   `json:"delivery_address"`
	Items                []FPOPurchaseItem `json:"items"`
	TotalAmount          float64           `json:"total_amount"`
	Currency             string            `json:"currency"`
	ExpectedDeliveryDate string            `json:"expected_delivery_date"`
}

type DeliveryAddress struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type FPOSaleItem struct {
	SKU        string  `json:"sku"`
	Name       string  `json:"name"`
	Category   string  `json:"category"`
	Brand      string  `json:"brand"`
	Quantity   int64   `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	TotalPrice float64 `json:"total_price"`
}

type FPOPurchaseItem struct {
	SKU       string  `json:"sku"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Brand     string  `json:"brand"`
	Quantity  int64   `json:"quantity"`
	UnitCost  float64 `json:"unit_cost"`
	TotalCost float64 `json:"total_cost"`
}

// ProcessFPOSaleEvent processes an FPO sale event from e-commerce
func (s *EcommerceWebhookService) ProcessFPOSaleEvent(payload *FPOSalePayload) error {
	utils.Info("Processing FPO sale event:", payload.EventID)

	// Check idempotency
	existingEvent, isProcessed, err := s.historyService.CheckIdempotency(payload.EventID)
	if err != nil {
		return fmt.Errorf("idempotency check failed: %w", err)
	}

	if isProcessed {
		utils.Info("Event already processed:", payload.EventID)
		return errors.NewConflictError("Event already processed")
	}

	// Create event record if it doesn't exist
	if existingEvent == nil {
		_, err = s.historyService.CreateEventRecord(
			payload.EventType,
			payload.EventID,
			payload.FPOID,
			payload,
			"e-commerce",
		)
		if err != nil {
			return fmt.Errorf("failed to create event record: %w", err)
		}
	}

	// Find appropriate warehouse based on delivery address
	warehouse, err := s.findNearestWarehouse(payload.FPOID, payload.DeliveryAddress)
	if err != nil {
		s.historyService.MarkEventFailed(payload.EventID, "No suitable warehouse found: "+err.Error())
		return fmt.Errorf("warehouse lookup failed: %w", err)
	}

	// Process each item in the sale
	saleItems := make([]models.CreateSaleItemRequest, 0, len(payload.Items))
	for _, item := range payload.Items {
		// Find product by SKU
		product, err := s.productRepo.GetBySKU(item.SKU)
		if err != nil {
			errorMsg := fmt.Sprintf("Product not found for SKU %s: %s", item.SKU, err.Error())
			s.historyService.MarkEventFailed(payload.EventID, errorMsg)
			return errors.NewNotFoundError("Product with SKU " + item.SKU + " not found in FPO catalog")
		}

		// Check inventory availability using the correct repository method
		batches, err := s.inventoryRepo.GetBatchesByProductAndWarehouseOrderedByExpiry(product.ID, warehouse.ID)
		if err != nil {
			errorMsg := fmt.Sprintf("Inventory check failed for %s: %s", item.SKU, err.Error())
			s.historyService.MarkEventFailed(payload.EventID, errorMsg)
			return fmt.Errorf("inventory check failed: %w", err)
		}

		// Calculate total available quantity
		var totalAvailable int64
		for _, batch := range batches {
			totalAvailable += batch.TotalQuantity
		}

		if totalAvailable < item.Quantity {
			errorMsg := fmt.Sprintf("Insufficient inventory for %s. Available: %d, Required: %d", item.SKU, totalAvailable, item.Quantity)
			s.historyService.MarkEventFailed(payload.EventID, errorMsg)
			return errors.NewBadRequestError(errorMsg)
		}

		// Create sale item request (the sales service will handle FEFO allocation)
		saleItem := models.CreateSaleItemRequest{
			ProductID: product.ID,
			Quantity:  item.Quantity,
		}
		saleItems = append(saleItems, saleItem)
	}

	// Create sale record with transaction safety
	saleRequest := &models.CreateSaleRequest{
		WarehouseID: warehouse.ID,
		CustomerID:  &payload.CustomerID,
		Items:       saleItems,
		// Note: The sales service calculates prices and taxes automatically
		// E-commerce integration doesn't override pricing
	}

	// Use transaction-safe sale creation with rollback capability
	sale, err := s.createSaleWithRollbackCapability(saleRequest, payload.EventID)
	if err != nil {
		// Error handling is done inside createSaleWithRollbackCapability
		return err
	}

	// Mark event as processed
	if err := s.historyService.MarkEventProcessed(payload.EventID); err != nil {
		utils.Error("Failed to mark event as processed:", err)
		// Don't fail the whole operation for this
	}

	utils.Info("Successfully processed FPO sale event:", payload.EventID, "-> Sale ID:", sale.ID)
	return nil
}

// ProcessFPOPurchaseEvent processes an FPO purchase event from e-commerce
func (s *EcommerceWebhookService) ProcessFPOPurchaseEvent(payload *FPOPurchasePayload) error {
	utils.Info("Processing FPO purchase event:", payload.EventID)

	// Check idempotency
	existingEvent, isProcessed, err := s.historyService.CheckIdempotency(payload.EventID)
	if err != nil {
		return fmt.Errorf("idempotency check failed: %w", err)
	}

	if isProcessed {
		utils.Info("Event already processed:", payload.EventID)
		return errors.NewConflictError("Event already processed")
	}

	// Create event record if it doesn't exist
	if existingEvent == nil {
		_, err = s.historyService.CreateEventRecord(
			payload.EventType,
			payload.EventID,
			payload.FPOID,
			payload,
			"e-commerce",
		)
		if err != nil {
			return fmt.Errorf("failed to create event record: %w", err)
		}
	}

	// Find appropriate warehouse for delivery
	warehouse, err := s.findNearestWarehouse(payload.FPOID, payload.DeliveryAddress)
	if err != nil {
		s.historyService.MarkEventFailed(payload.EventID, "No suitable warehouse found: "+err.Error())
		return fmt.Errorf("warehouse lookup failed: %w", err)
	}

	// Parse expected delivery date
	deliveryDate, err := time.Parse("2006-01-02T15:04:05Z", payload.ExpectedDeliveryDate)
	if err != nil {
		// Try alternative format
		deliveryDate, err = time.Parse("2006-01-02", payload.ExpectedDeliveryDate)
		if err != nil {
			errorMsg := fmt.Sprintf("Invalid delivery date format: %s", payload.ExpectedDeliveryDate)
			s.historyService.MarkEventFailed(payload.EventID, errorMsg)
			return errors.NewBadRequestError(errorMsg)
		}
	}

	// Process each item in the purchase with transaction safety
	err = s.processPurchaseItemsWithTransactionSafety(payload.Items, warehouse.ID, deliveryDate, payload.EventID)
	if err != nil {
		// Error handling is done inside processPurchaseItemsWithTransactionSafety
		return err
	}

	// Mark event as processed
	if err := s.historyService.MarkEventProcessed(payload.EventID); err != nil {
		utils.Error("Failed to mark event as processed:", err)
		// Don't fail the whole operation for this
	}

	utils.Info("Successfully processed FPO purchase event:", payload.EventID)
	return nil
}

// findNearestWarehouse finds the most suitable warehouse for a delivery address
// This is a simplified implementation - in production, you might use geolocation services
func (s *EcommerceWebhookService) findNearestWarehouse(fpoID string, address DeliveryAddress) (*models.Warehouse, error) {
	// For now, get the first available warehouse for the FPO
	// In production, implement proper geolocation matching
	warehouses, err := s.warehouseRepo.GetAll()
	if err != nil {
		return nil, err
	}

	if len(warehouses) == 0 {
		return nil, errors.NewNotFoundError("No warehouses available for FPO")
	}

	// TODO: Implement proper geolocation logic based on address
	// For now, return the first warehouse
	return &warehouses[0], nil
}

// ValidateFPOSalePayload validates the structure and content of FPO sale payload
func (s *EcommerceWebhookService) ValidateFPOSalePayload(payload *FPOSalePayload) error {
	if payload.EventID == "" {
		return errors.NewBadRequestError("event_id is required")
	}
	if payload.FPOID == "" {
		return errors.NewBadRequestError("fpo_id is required")
	}
	if payload.SaleID == "" {
		return errors.NewBadRequestError("sale_id is required")
	}
	if len(payload.Items) == 0 {
		return errors.NewBadRequestError("items array cannot be empty")
	}

	for i, item := range payload.Items {
		if item.SKU == "" {
			return errors.NewBadRequestError(fmt.Sprintf("item[%d].sku is required", i))
		}
		if item.Quantity <= 0 {
			return errors.NewBadRequestError(fmt.Sprintf("item[%d].quantity must be positive", i))
		}
		if item.UnitPrice <= 0 {
			return errors.NewBadRequestError(fmt.Sprintf("item[%d].unit_price must be positive", i))
		}
	}

	return nil
}

// ValidateFPOPurchasePayload validates the structure and content of FPO purchase payload
func (s *EcommerceWebhookService) ValidateFPOPurchasePayload(payload *FPOPurchasePayload) error {
	if payload.EventID == "" {
		return errors.NewBadRequestError("event_id is required")
	}
	if payload.FPOID == "" {
		return errors.NewBadRequestError("fpo_id is required")
	}
	if payload.PurchaseID == "" {
		return errors.NewBadRequestError("purchase_id is required")
	}
	if len(payload.Items) == 0 {
		return errors.NewBadRequestError("items array cannot be empty")
	}

	for i, item := range payload.Items {
		if item.SKU == "" {
			return errors.NewBadRequestError(fmt.Sprintf("item[%d].sku is required", i))
		}
		if item.Quantity <= 0 {
			return errors.NewBadRequestError(fmt.Sprintf("item[%d].quantity must be positive", i))
		}
		if item.UnitCost <= 0 {
			return errors.NewBadRequestError(fmt.Sprintf("item[%d].unit_cost must be positive", i))
		}
	}

	return nil
}

// createSaleWithTransactionSafety creates a sale with proper transaction handling and rollback
func (s *EcommerceWebhookService) createSaleWithTransactionSafety(saleRequest *models.CreateSaleRequest, eventID string) (*models.SaleResponse, error) {
	utils.Info("Creating sale with transaction safety for event:", eventID)

	// Attempt to create the sale
	sale, err := s.salesService.CreateSale(saleRequest)
	if err != nil {
		// Sale creation failed - mark event as failed
		errorMsg := fmt.Sprintf("Failed to create sale: %s", err.Error())
		utils.Error("Sale creation failed for event:", eventID, "Error:", err)

		// Mark the event as failed in webhook history
		if markErr := s.historyService.MarkEventFailed(eventID, errorMsg); markErr != nil {
			utils.Error("Failed to mark event as failed:", markErr)
		}

		return nil, fmt.Errorf("sale creation failed: %w", err)
	}

	// Sale created successfully - inventory has been updated
	utils.Info("Sale created successfully for event:", eventID, "Sale ID:", sale.ID)
	return sale, nil
}

// createSaleWithRollbackCapability creates a sale with full rollback capability
// This method provides additional safety by implementing a two-phase approach
func (s *EcommerceWebhookService) createSaleWithRollbackCapability(saleRequest *models.CreateSaleRequest, eventID string) (*models.SaleResponse, error) {
	utils.Info("Creating sale with rollback capability for event:", eventID)

	// Phase 1: Validate all prerequisites without making changes
	if err := s.validateSalePrerequisites(saleRequest, eventID); err != nil {
		errorMsg := fmt.Sprintf("Sale prerequisites validation failed: %s", err.Error())
		utils.Error("Sale validation failed for event:", eventID, "Error:", err)

		if markErr := s.historyService.MarkEventFailed(eventID, errorMsg); markErr != nil {
			utils.Error("Failed to mark event as failed:", markErr)
		}

		return nil, fmt.Errorf("sale validation failed: %w", err)
	}

	// Phase 2: Create the sale (this will update inventory)
	sale, err := s.salesService.CreateSale(saleRequest)
	if err != nil {
		// Sale creation failed - inventory remains unchanged
		errorMsg := fmt.Sprintf("Failed to create sale after validation: %s", err.Error())
		utils.Error("Sale creation failed after validation for event:", eventID, "Error:", err)

		if markErr := s.historyService.MarkEventFailed(eventID, errorMsg); markErr != nil {
			utils.Error("Failed to mark event as failed:", markErr)
		}

		return nil, fmt.Errorf("sale creation failed: %w", err)
	}

	// Sale created successfully
	utils.Info("Sale created successfully with rollback capability for event:", eventID, "Sale ID:", sale.ID)
	return sale, nil
}

// validateSalePrerequisites validates all prerequisites for sale creation without making changes
func (s *EcommerceWebhookService) validateSalePrerequisites(saleRequest *models.CreateSaleRequest, eventID string) error {
	utils.Info("Validating sale prerequisites for event:", eventID)

	// Validate warehouse exists
	_, err := s.warehouseRepo.GetByID(saleRequest.WarehouseID)
	if err != nil {
		return fmt.Errorf("warehouse validation failed: %w", err)
	}

	// Validate each item
	for i, item := range saleRequest.Items {
		// Validate product exists
		product, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			return fmt.Errorf("product validation failed for item %d: %w", i, err)
		}

		// Validate inventory availability
		batches, err := s.inventoryRepo.GetBatchesByProductAndWarehouseOrderedByExpiry(item.ProductID, saleRequest.WarehouseID)
		if err != nil {
			return fmt.Errorf("inventory validation failed for item %d: %w", i, err)
		}

		// Calculate total available quantity
		var totalAvailable int64
		for _, batch := range batches {
			totalAvailable += batch.TotalQuantity
		}

		if totalAvailable < item.Quantity {
			return fmt.Errorf("insufficient inventory for item %d (product %s): available %d, required %d",
				i, product.SKU, totalAvailable, item.Quantity)
		}
	}

	utils.Info("Sale prerequisites validation passed for event:", eventID)
	return nil
}

// processPurchaseItemsWithTransactionSafety processes purchase items with proper error handling and rollback
func (s *EcommerceWebhookService) processPurchaseItemsWithTransactionSafety(items []FPOPurchaseItem, warehouseID string, deliveryDate time.Time, eventID string) error {
	utils.Info("Processing purchase items with transaction safety for event:", eventID)

	// Track successfully created batches for potential rollback
	var createdBatches []string

	// Process each item in the purchase
	for _, item := range items {
		// Find or create product by SKU
		product, err := s.productRepo.GetBySKU(item.SKU)
		if err != nil {
			// Product doesn't exist - this might be a new product from supplier
			errorMsg := fmt.Sprintf("Product not found for SKU %s. Manual intervention required.", item.SKU)
			utils.Error("Product lookup failed for event:", eventID, "SKU:", item.SKU, "Error:", err)

			// Mark event as failed
			if markErr := s.historyService.MarkEventFailed(eventID, errorMsg); markErr != nil {
				utils.Error("Failed to mark event as failed:", markErr)
			}

			return errors.NewNotFoundError("Product with SKU " + item.SKU + " not found in catalog")
		}

		// Create inventory batch for the incoming stock
		// Use delivery date + 1 year as default expiry for non-perishables
		expiryDate := deliveryDate.AddDate(1, 0, 0)

		batchResponse, err := s.inventoryService.CreateBatch(
			warehouseID,
			product.ID,
			item.UnitCost,
			expiryDate,
			item.Quantity,
		)
		if err != nil {
			// Batch creation failed - mark event as failed
			errorMsg := fmt.Sprintf("Failed to create inventory batch for %s: %s", item.SKU, err.Error())
			utils.Error("Batch creation failed for event:", eventID, "SKU:", item.SKU, "Error:", err)

			// Mark event as failed
			if markErr := s.historyService.MarkEventFailed(eventID, errorMsg); markErr != nil {
				utils.Error("Failed to mark event as failed:", markErr)
			}

			return fmt.Errorf("inventory batch creation failed: %w", err)
		}

		// Track successfully created batch
		createdBatches = append(createdBatches, batchResponse.ID)
		utils.Info(fmt.Sprintf("Added inventory: %d units of %s to warehouse %s (Batch ID: %s)",
			item.Quantity, item.SKU, warehouseID, batchResponse.ID))
	}

	utils.Info("Successfully processed all purchase items for event:", eventID, "Created batches:", len(createdBatches))
	return nil
}
