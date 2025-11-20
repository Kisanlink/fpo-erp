package services

import (
	"context"
	"fmt"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PurchaseOrderService handles purchase order business logic
type PurchaseOrderService struct {
	poRepo           *repositories.PurchaseOrderRepository
	collaboratorRepo *repositories.CollaboratorRepository
	warehouseRepo    *repositories.WarehouseRepository
	productRepo      *repositories.ProductRepository
	variantRepo      *repositories.ProductVariantRepository
	grnRepo          *repositories.GRNRepository
	inventoryRepo    *repositories.InventoryRepository
	logger           interfaces.Logger
}

// NewPurchaseOrderService creates a new purchase order service
func NewPurchaseOrderService(
	poRepo *repositories.PurchaseOrderRepository,
	collaboratorRepo *repositories.CollaboratorRepository,
	warehouseRepo *repositories.WarehouseRepository,
	productRepo *repositories.ProductRepository,
	variantRepo *repositories.ProductVariantRepository,
	grnRepo *repositories.GRNRepository,
	inventoryRepo *repositories.InventoryRepository,
	logger interfaces.Logger,
) *PurchaseOrderService {
	return &PurchaseOrderService{
		poRepo:           poRepo,
		collaboratorRepo: collaboratorRepo,
		warehouseRepo:    warehouseRepo,
		productRepo:      productRepo,
		variantRepo:      variantRepo,
		grnRepo:          grnRepo,
		inventoryRepo:    inventoryRepo,
		logger:           logger,
	}
}

// CreatePurchaseOrder creates a new purchase order with items
func (s *PurchaseOrderService) CreatePurchaseOrder(ctx context.Context, request *models.CreatePurchaseOrderRequest) (*models.PurchaseOrderResponse, error) {
	s.logger.Info("Creating purchase order",
		zap.String("collaborator_id", request.CollaboratorID),
		zap.String("warehouse_id", request.WarehouseID),
		zap.Int("items_count", len(request.Items)))

	// Validate collaborator exists and is active
	collaborator, err := s.collaboratorRepo.GetByID(request.CollaboratorID)
	if err != nil {
		s.logger.Error("Failed to retrieve collaborator for PO",
			zap.Error(err),
			zap.String("collaborator_id", request.CollaboratorID))
		return nil, err
	}
	if collaborator.IsActive != nil && !*collaborator.IsActive {
		s.logger.Warn("Attempted to create PO with inactive collaborator",
			zap.String("collaborator_id", request.CollaboratorID))
		return nil, errors.NewBadRequestError("collaborator is not active")
	}

	// Validate warehouse exists
	_, err = s.warehouseRepo.GetByID(request.WarehouseID)
	if err != nil {
		return nil, err
	}

	// Parse dates
	var orderDate time.Time
	if request.OrderDate != nil {
		orderDate, err = time.Parse("2006-01-02", *request.OrderDate)
		if err != nil {
			return nil, errors.NewValidationError("invalid order_date format")
		}
	} else {
		orderDate = time.Now().UTC()
	}

	expectedDelivery, err := time.Parse("2006-01-02", request.ExpectedDelivery)
	if err != nil {
		return nil, errors.NewValidationError("invalid expected_delivery_date format")
	}

	// Validate expected delivery is after order date
	if expectedDelivery.Before(orderDate) {
		return nil, errors.NewValidationError("expected delivery date must be after order date")
	}

	// Validate items and calculate total
	if len(request.Items) == 0 {
		return nil, errors.NewValidationError("purchase order must have at least one item")
	}

	var totalAmount float64
	variantDetails := make(map[string]*models.ProductVariant)

	for _, item := range request.Items {
		// Validate variant exists
		variant, err := s.variantRepo.GetByID(item.VariantID)
		if err != nil {
			return nil, errors.NewNotFoundError("Product variant")
		}
		variantDetails[item.VariantID] = variant

		// Calculate line total
		lineTotal := float64(item.Quantity) * item.UnitPrice
		totalAmount += lineTotal
	}

	// Generate PO number
	poNumber, err := s.generatePONumber()
	if err != nil {
		s.logger.Error("Failed to generate PO number", zap.Error(err))
		return nil, err
	}

	s.logger.Debug("Generated PO number",
		zap.String("po_number", poNumber),
		zap.Float64("total_amount", totalAmount))

	// Create purchase order and items in a transaction
	var po *models.PurchaseOrder
	err = s.poRepo.WithTransaction(func(tx *gorm.DB) error {
		// Create purchase order
		po = models.NewPurchaseOrder(
			poNumber,
			request.CollaboratorID,
			request.WarehouseID,
			orderDate,
			expectedDelivery,
		)
		po.TotalAmount = totalAmount

		if err := s.poRepo.CreateWithTx(tx, po); err != nil {
			return err
		}

		// Create purchase order items
		for _, itemReq := range request.Items {
			variant := variantDetails[itemReq.VariantID]
			item := models.NewPurchaseOrderItem(
				po.ID,
				itemReq.VariantID,
				itemReq.Quantity,
				itemReq.UnitPrice,
			)
			// Snapshot variant details
			item.ProductName = &variant.VariantName
			item.ProductSKU = variant.SKU

			if err := s.poRepo.CreateItemWithTx(tx, item); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to create purchase order in transaction",
			zap.Error(err),
			zap.String("po_number", poNumber))
		return nil, err
	}

	s.logger.Info("Purchase order created successfully",
		zap.String("po_id", po.ID),
		zap.String("po_number", po.PONumber),
		zap.Float64("total_amount", po.TotalAmount))

	// Fetch complete PO with items
	return s.GetPurchaseOrder(ctx, po.ID)
}

// GetPurchaseOrder retrieves a purchase order by ID with items
func (s *PurchaseOrderService) GetPurchaseOrder(ctx context.Context, id string) (*models.PurchaseOrderResponse, error) {
	s.logger.Debug("Retrieving purchase order",
		zap.String("po_id", id))

	po, err := s.poRepo.GetByIDWithItems(id)
	if err != nil {
		s.logger.Error("Failed to retrieve purchase order",
			zap.Error(err),
			zap.String("po_id", id))
		return nil, err
	}

	return s.buildPurchaseOrderResponse(po)
}

// GetAllPurchaseOrders retrieves all purchase orders
func (s *PurchaseOrderService) GetAllPurchaseOrders(ctx context.Context) ([]models.PurchaseOrderResponse, error) {
	s.logger.Info("Retrieving all purchase orders")

	pos, err := s.poRepo.GetAll()
	if err != nil {
		s.logger.Error("Failed to retrieve all purchase orders", zap.Error(err))
		return nil, err
	}

	s.logger.Debug("Retrieved purchase orders",
		zap.Int("count", len(pos)))

	var responses []models.PurchaseOrderResponse
	for _, po := range pos {
		// Get with items
		poWithItems, err := s.poRepo.GetByIDWithItems(po.ID)
		if err != nil {
			continue
		}
		response, err := s.buildPurchaseOrderResponse(poWithItems)
		if err != nil {
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// GetPurchaseOrdersByCollaborator retrieves purchase orders by collaborator
func (s *PurchaseOrderService) GetPurchaseOrdersByCollaborator(ctx context.Context, collaboratorID string) ([]models.PurchaseOrderResponse, error) {
	// Validate collaborator exists
	_, err := s.collaboratorRepo.GetByID(collaboratorID)
	if err != nil {
		return nil, err
	}

	pos, err := s.poRepo.GetByCollaborator(collaboratorID)
	if err != nil {
		return nil, err
	}

	var responses []models.PurchaseOrderResponse
	for _, po := range pos {
		// Get with items
		poWithItems, err := s.poRepo.GetByIDWithItems(po.ID)
		if err != nil {
			continue
		}
		response, err := s.buildPurchaseOrderResponse(poWithItems)
		if err != nil {
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// GetPurchaseOrdersByStatus retrieves purchase orders by status
func (s *PurchaseOrderService) GetPurchaseOrdersByStatus(ctx context.Context, status string) ([]models.PurchaseOrderResponse, error) {
	// Validate status
	if !isValidPOStatus(status) {
		return nil, errors.NewValidationError(fmt.Sprintf("invalid status: %s", status))
	}

	pos, err := s.poRepo.GetByStatus(status)
	if err != nil {
		return nil, err
	}

	var responses []models.PurchaseOrderResponse
	for _, po := range pos {
		// Get with items
		poWithItems, err := s.poRepo.GetByIDWithItems(po.ID)
		if err != nil {
			continue
		}
		response, err := s.buildPurchaseOrderResponse(poWithItems)
		if err != nil {
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// GetPendingDeliveries retrieves all pending purchase orders
func (s *PurchaseOrderService) GetPendingDeliveries(ctx context.Context) ([]models.PurchaseOrderResponse, error) {
	pos, err := s.poRepo.GetPendingDeliveries()
	if err != nil {
		return nil, err
	}

	var responses []models.PurchaseOrderResponse
	for _, po := range pos {
		// Get with items
		poWithItems, err := s.poRepo.GetByIDWithItems(po.ID)
		if err != nil {
			continue
		}
		response, err := s.buildPurchaseOrderResponse(poWithItems)
		if err != nil {
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// UpdatePurchaseOrderStatus updates the status of a purchase order
// Supports auto-GRN creation when status = "verified" with delivery details
func (s *PurchaseOrderService) UpdatePurchaseOrderStatus(ctx context.Context, id string, request *models.UpdatePOStatusRequest, userID string) (*models.PurchaseOrderResponse, error) {
	s.logger.Info("Updating purchase order status",
		zap.String("po_id", id),
		zap.String("new_status", request.Status),
		zap.String("user_id", userID),
		zap.Bool("has_items", len(request.Items) > 0),
		zap.Bool("accept_all", request.AcceptAll != nil && *request.AcceptAll))

	// Validate status
	if !isValidPOStatus(request.Status) {
		s.logger.Warn("Invalid PO status provided",
			zap.String("status", request.Status))
		return nil, errors.NewValidationError(fmt.Sprintf("invalid status: %s", request.Status))
	}

	// Get existing PO with items
	po, err := s.poRepo.GetByIDWithItems(id)
	if err != nil {
		s.logger.Error("Failed to retrieve PO for status update",
			zap.Error(err),
			zap.String("po_id", id))
		return nil, err
	}

	// Validate status transition
	if !isValidPOStatusTransition(po.Status, request.Status) {
		return nil, errors.NewBadRequestError(fmt.Sprintf("invalid status transition from %s to %s", po.Status, request.Status))
	}

	// Set actual delivery date if status is delivered
	var actualDelivery time.Time
	if request.Status == "delivered" {
		if request.ActualDelivery != nil {
			actualDelivery = *request.ActualDelivery
		} else {
			actualDelivery = time.Now().UTC()
		}
	}

	// Pattern Detection: Auto-create GRN if status = "verified" and delivery details provided
	if request.Status == "verified" && (request.AcceptAll != nil || len(request.Items) > 0) {
		s.logger.Info("Auto-GRN trigger detected",
			zap.String("po_id", po.ID),
			zap.Bool("accept_all", request.AcceptAll != nil && *request.AcceptAll))

		// Check if GRN already exists for this PO
		grnExists, err := s.grnRepo.GRNExistsForPO(po.ID)
		if err != nil {
			s.logger.Error("Failed to check GRN existence",
				zap.Error(err),
				zap.String("po_id", po.ID))
			return nil, err
		}
		if grnExists {
			s.logger.Warn("GRN already exists for this PO",
				zap.String("po_id", po.ID))
			return nil, errors.NewConflictError("GRN already exists for this purchase order")
		}

		// Pattern 1: Accept All
		if request.AcceptAll != nil && *request.AcceptAll {
			s.logger.Debug("Processing Pattern 1: Accept All",
				zap.String("po_id", po.ID))
			if err := s.processAcceptAll(ctx, po, actualDelivery, request.DefaultExpiryDate, userID); err != nil {
				s.logger.Error("Failed to process Accept All pattern",
					zap.Error(err),
					zap.String("po_id", po.ID))
				return nil, err
			}
		} else if len(request.Items) > 0 {
			s.logger.Debug("Processing Pattern 2/3: Per-item details",
				zap.String("po_id", po.ID),
				zap.Int("items_count", len(request.Items)))
			// Pattern 2 & 3: Per-item details
			if err := s.processDeliveryItems(ctx, po, actualDelivery, request.Items, userID); err != nil {
				s.logger.Error("Failed to process delivery items",
					zap.Error(err),
					zap.String("po_id", po.ID))
				return nil, err
			}
		}

		s.logger.Info("Auto-GRN created successfully",
			zap.String("po_id", po.ID))

		// Update PO status and delivery date (already done in transaction above)
		return s.GetPurchaseOrder(ctx, id)
	}

	// Traditional flow: Just update status without auto-GRN
	po.Status = request.Status
	if request.Status == "delivered" {
		po.ActualDelivery = &actualDelivery
	}

	// Save to database
	if err := s.poRepo.Update(po); err != nil {
		return nil, err
	}

	return s.GetPurchaseOrder(ctx, id)
}

// UpdatePaymentStatus updates the payment status of a purchase order
func (s *PurchaseOrderService) UpdatePaymentStatus(ctx context.Context, id string, request *models.UpdatePOPaymentRequest) (*models.PurchaseOrderResponse, error) {
	s.logger.Info("Updating purchase order payment status",
		zap.String("po_id", id),
		zap.String("payment_status", request.PaymentStatus),
		zap.Float64("paid_amount", request.PaidAmount))

	// Validate payment status
	if !isValidPaymentStatus(request.PaymentStatus) {
		s.logger.Warn("Invalid payment status provided",
			zap.String("payment_status", request.PaymentStatus))
		return nil, errors.NewValidationError(fmt.Sprintf("invalid payment status: %s", request.PaymentStatus))
	}

	// Get existing PO
	po, err := s.poRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve PO for payment update",
			zap.Error(err),
			zap.String("po_id", id))
		return nil, err
	}

	// Validate paid amount doesn't exceed total
	if request.PaidAmount > po.TotalAmount {
		return nil, errors.NewValidationError("paid amount cannot exceed total amount")
	}

	// Update payment fields
	po.PaidAmount = request.PaidAmount
	po.PaymentStatus = request.PaymentStatus

	// Auto-update status to paid if fully paid
	if request.PaidAmount >= po.TotalAmount {
		po.PaymentStatus = "paid"
		if po.Status == "delivered" {
			po.Status = "paid"
		}
	}

	// Save to database
	if err := s.poRepo.Update(po); err != nil {
		s.logger.Error("Failed to update payment status",
			zap.Error(err),
			zap.String("po_id", id))
		return nil, err
	}

	s.logger.Info("Payment status updated successfully",
		zap.String("po_id", id),
		zap.String("payment_status", po.PaymentStatus),
		zap.Float64("paid_amount", po.PaidAmount))

	return s.GetPurchaseOrder(ctx, id)
}

// processAcceptAll handles Pattern 1: Accept all items with default expiry
func (s *PurchaseOrderService) processAcceptAll(ctx context.Context, po *models.PurchaseOrder, actualDelivery time.Time, defaultExpiryDate *string, userID string) error {
	// Validate default expiry date
	if defaultExpiryDate == nil || *defaultExpiryDate == "" {
		return errors.NewValidationError("default_expiry_date is required when accept_all is true")
	}

	_, err := time.Parse("2006-01-02", *defaultExpiryDate)
	if err != nil {
		return errors.NewValidationError("invalid default_expiry_date format")
	}

	// Build delivery items from all PO items
	var deliveryItems []models.DeliveryItemRequest
	for _, poItem := range po.Items {
		acceptTrue := true
		deliveryItems = append(deliveryItems, models.DeliveryItemRequest{
			POItemID:    poItem.ID,
			Accept:      &acceptTrue, // Pattern 1: Accept field only - system auto-derives quantities
			ExpiryDate:  *defaultExpiryDate,
			BatchNumber: nil, // Optional - not available in AcceptAll pattern
		})
	}

	// Process with standard flow
	return s.processDeliveryItems(ctx, po, actualDelivery, deliveryItems, userID)
}

// processDeliveryItems handles Pattern 2 & 3: Per-item details
func (s *PurchaseOrderService) processDeliveryItems(ctx context.Context, po *models.PurchaseOrder, actualDelivery time.Time, items []models.DeliveryItemRequest, userID string) error {
	s.logger.Info("Processing delivery items for auto-GRN",
		zap.String("po_id", po.ID),
		zap.Int("items_count", len(items)),
		zap.String("user_id", userID))

	// Validate delivery items
	if err := s.validateDeliveryItems(po, items); err != nil {
		s.logger.Error("Delivery items validation failed",
			zap.Error(err),
			zap.String("po_id", po.ID))
		return err
	}

	// Generate GRN number
	grnNumber, err := s.generateGRNNumber()
	if err != nil {
		s.logger.Error("Failed to generate GRN number",
			zap.Error(err),
			zap.String("po_id", po.ID))
		return err
	}

	s.logger.Debug("Generated GRN number for auto-GRN",
		zap.String("grn_number", grnNumber),
		zap.String("po_id", po.ID))

	// Process in transaction
	return s.poRepo.WithTransaction(func(tx *gorm.DB) error {
		// Create GRN
		grn := models.NewGRN(
			grnNumber,
			po.ID,
			po.WarehouseID,
			userID,
			actualDelivery,
			"", // Quality status calculated later
		)

		if err := s.grnRepo.CreateWithTx(tx, grn); err != nil {
			return err
		}

		// Process each item
		for _, itemReq := range items {
			// Find corresponding PO item
			var poItem *models.PurchaseOrderItem
			for i := range po.Items {
				if po.Items[i].ID == itemReq.POItemID {
					poItem = &po.Items[i]
					break
				}
			}
			if poItem == nil {
				return errors.NewNotFoundError("Purchase order item")
			}

			// Determine quantities based on pattern
			var receivedQty, acceptedQty int64
			if itemReq.Accept != nil {
				// Pattern 2: Simple Accept/Reject
				receivedQty = poItem.Quantity
				if *itemReq.Accept {
					acceptedQty = poItem.Quantity
				} else {
					acceptedQty = 0
				}
			} else if itemReq.ReceivedQuantity != nil && itemReq.AcceptedQuantity != nil {
				// Pattern 3: Detailed Quantities
				receivedQty = *itemReq.ReceivedQuantity
				acceptedQty = *itemReq.AcceptedQuantity
			} else {
				return errors.NewValidationError(fmt.Sprintf("item %s must have either accept field or quantity fields", itemReq.POItemID))
			}

			// Parse expiry date
			expiryDate, err := time.Parse("2006-01-02", itemReq.ExpiryDate)
			if err != nil {
				return errors.NewValidationError(fmt.Sprintf("invalid expiry_date for item %s", itemReq.POItemID))
			}

			// Create GRN item
			grnItem := models.NewGRNItem(
				grn.ID,
				itemReq.POItemID,
				poItem.VariantID,
				poItem.Quantity,
				receivedQty,
				acceptedQty,
				expiryDate,
			)
			grnItem.BatchNumber = itemReq.BatchNumber

			// Create inventory batch for accepted quantity
			if acceptedQty > 0 {
				s.logger.Debug("Creating inventory batch for accepted quantity",
					zap.String("grn_number", grnNumber),
					zap.String("variant_id", poItem.VariantID),
					zap.Int64("accepted_qty", acceptedQty),
					zap.Float64("cost_price", poItem.UnitPrice))

				batch := models.NewInventoryBatch(
					po.WarehouseID,
					poItem.VariantID,
					poItem.UnitPrice, // ALL-IN cost price from PO
					expiryDate,
					acceptedQty,
					0,          // CGST rate 0 (price is ALL-IN)
					0,          // SGST rate 0 (price is ALL-IN)
					[]string{}, // No custom taxes
					false,      // Not tax exempt
				)

				if err := s.inventoryRepo.CreateBatchWithTx(tx, batch); err != nil {
					s.logger.Error("Failed to create inventory batch",
						zap.Error(err),
						zap.String("grn_number", grnNumber))
					return err
				}

				s.logger.Debug("Inventory batch created",
					zap.String("batch_id", batch.ID),
					zap.String("grn_number", grnNumber))

				// Link inventory batch to GRN item
				grnItem.InventoryBatchID = &batch.ID

				// Create initial inventory transaction
				note := fmt.Sprintf("Initial stock from GRN %s", grnNumber)
				transaction := models.NewInventoryTransaction(
					batch.ID,
					"purchase",
					acceptedQty,
					&grn.ID,
					&userID,
					&note,
					actualDelivery,
				)
				if err := s.inventoryRepo.CreateTransactionWithTx(tx, transaction); err != nil {
					s.logger.Error("Failed to create inventory transaction",
						zap.Error(err),
						zap.String("batch_id", batch.ID))
					return err
				}

				s.logger.Debug("Inventory transaction created",
					zap.String("batch_id", batch.ID),
					zap.Int64("quantity", acceptedQty))
			}

			// Save GRN item
			if err := s.grnRepo.CreateItemWithTx(tx, grnItem); err != nil {
				return err
			}

			// Update PO item received quantity
			if err := s.poRepo.UpdateItemReceivedQuantityWithTx(tx, poItem.ID, receivedQty); err != nil {
				return err
			}
		}

		// Calculate quality status
		qualityStatus := s.calculateQualityStatus(items)
		grn.QualityStatus = qualityStatus

		// Update GRN with quality status
		if err := tx.Model(grn).Update("quality_status", qualityStatus).Error; err != nil {
			return err
		}

		// Update PO status and delivery date
		po.Status = "delivered"
		po.ActualDelivery = &actualDelivery
		if err := s.poRepo.UpdateWithTx(tx, po); err != nil {
			return err
		}

		return nil
	})
}

// validateDeliveryItems validates delivery items against PO
func (s *PurchaseOrderService) validateDeliveryItems(po *models.PurchaseOrder, items []models.DeliveryItemRequest) error {
	// Build PO item map
	poItemMap := make(map[string]*models.PurchaseOrderItem)
	for i := range po.Items {
		poItemMap[po.Items[i].ID] = &po.Items[i]
	}

	// Track seen items to prevent duplicates
	seenItems := make(map[string]bool)

	for _, item := range items {
		// Check duplicate
		if seenItems[item.POItemID] {
			return errors.NewValidationError(fmt.Sprintf("duplicate po_item_id: %s", item.POItemID))
		}
		seenItems[item.POItemID] = true

		// Validate item belongs to this PO
		poItem, exists := poItemMap[item.POItemID]
		if !exists {
			return errors.NewValidationError(fmt.Sprintf("po_item_id %s does not belong to this purchase order", item.POItemID))
		}

		// Validate pattern usage
		hasAccept := item.Accept != nil
		hasQuantities := item.ReceivedQuantity != nil && item.AcceptedQuantity != nil

		if !hasAccept && !hasQuantities {
			return errors.NewValidationError(fmt.Sprintf("item %s must have either accept field or quantity fields", item.POItemID))
		}

		if hasAccept && hasQuantities {
			return errors.NewValidationError(fmt.Sprintf("item %s cannot have both accept field and quantity fields", item.POItemID))
		}

		// Validate quantities if using Pattern 3
		if hasQuantities {
			if *item.ReceivedQuantity <= 0 {
				return errors.NewValidationError(fmt.Sprintf("received_quantity must be greater than 0 for item %s", item.POItemID))
			}
			if *item.AcceptedQuantity < 0 {
				return errors.NewValidationError(fmt.Sprintf("accepted_quantity cannot be negative for item %s", item.POItemID))
			}
			if *item.AcceptedQuantity > *item.ReceivedQuantity {
				return errors.NewValidationError(fmt.Sprintf("accepted_quantity cannot exceed received_quantity for item %s", item.POItemID))
			}
			if *item.ReceivedQuantity > poItem.Quantity {
				return errors.NewValidationError(fmt.Sprintf("received_quantity (%d) cannot exceed ordered quantity (%d) for item %s", *item.ReceivedQuantity, poItem.Quantity, item.POItemID))
			}
		}

		// Validate expiry date format
		if item.ExpiryDate != "" {
			if _, err := time.Parse("2006-01-02", item.ExpiryDate); err != nil {
				return errors.NewValidationError(fmt.Sprintf("invalid expiry_date format for item %s: must be YYYY-MM-DD", item.POItemID))
			}
		}
	}

	return nil
}

// calculateQualityStatus determines overall quality status from items
func (s *PurchaseOrderService) calculateQualityStatus(items []models.DeliveryItemRequest) string {
	allAccepted := true
	allRejected := true

	for _, item := range items {
		if item.Accept != nil {
			if *item.Accept {
				allRejected = false
			} else {
				allAccepted = false
			}
		} else if item.ReceivedQuantity != nil && item.AcceptedQuantity != nil {
			if *item.AcceptedQuantity > 0 {
				allRejected = false
			}
			if *item.AcceptedQuantity < *item.ReceivedQuantity {
				allAccepted = false
			}
		}
	}

	if allAccepted {
		return "accepted"
	}
	if allRejected {
		return "rejected"
	}
	return "partial"
}

// generateGRNNumber generates a unique GRN number in format: GRN-YYYY-NNNN
func (s *PurchaseOrderService) generateGRNNumber() (string, error) {
	year := time.Now().UTC().Year()

	// Try to find the next available number
	for i := 1; i <= 9999; i++ {
		grnNumber := fmt.Sprintf("GRN-%d-%04d", year, i)
		exists, err := s.grnRepo.GRNNumberExists(grnNumber)
		if err != nil {
			return "", err
		}
		if !exists {
			return grnNumber, nil
		}
	}

	return "", errors.NewInternalServerError("failed to generate unique GRN number")
}

// generatePONumber generates a unique PO number in format: PO-YYYY-NNNN
func (s *PurchaseOrderService) generatePONumber() (string, error) {
	year := time.Now().UTC().Year()

	// Try to find the next available number
	for i := 1; i <= 9999; i++ {
		poNumber := fmt.Sprintf("PO-%d-%04d", year, i)
		exists, err := s.poRepo.PONumberExists(poNumber)
		if err != nil {
			return "", err
		}
		if !exists {
			return poNumber, nil
		}
	}

	return "", errors.NewInternalServerError("failed to generate unique PO number")
}

// buildPurchaseOrderResponse builds a response with related entity details
func (s *PurchaseOrderService) buildPurchaseOrderResponse(po *models.PurchaseOrder) (*models.PurchaseOrderResponse, error) {
	// Calculate total rejected amount from GRN (if GRN exists)
	totalRejectedAmount, err := s.grnRepo.GetTotalRejectedAmountByPO(po.ID)
	if err != nil {
		// Log error but don't fail - treat as 0 if no GRN exists yet
		s.logger.Warn("Failed to calculate rejected amount for PO",
			zap.String("po_id", po.ID),
			zap.Error(err))
		totalRejectedAmount = 0
	}

	// Calculate amount owed (Option A: keep TotalAmount unchanged, subtract rejections)
	amountOwed := po.TotalAmount - totalRejectedAmount

	response := &models.PurchaseOrderResponse{
		ID:                  po.ID,
		PONumber:            po.PONumber,
		CollaboratorID:      po.CollaboratorID,
		CollaboratorName:    po.Collaborator.CompanyName,
		WarehouseID:         po.WarehouseID,
		WarehouseName:       po.Warehouse.Name,
		OrderDate:           po.OrderDate.Format("2006-01-02"),
		ExpectedDelivery:    po.ExpectedDelivery.Format("2006-01-02"),
		Status:              po.Status,
		TotalAmount:         po.TotalAmount,
		TotalRejectedAmount: totalRejectedAmount,
		AmountOwed:          amountOwed,
		PaymentStatus:       po.PaymentStatus,
		PaidAmount:          po.PaidAmount,
		CreatedAt:           po.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:           po.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if po.ActualDelivery != nil {
		actualDelivery := po.ActualDelivery.Format("2006-01-02")
		response.ActualDelivery = &actualDelivery
	}

	// Add items
	var items []models.PurchaseOrderItemResponse
	for _, item := range po.Items {
		items = append(items, models.PurchaseOrderItemResponse{
			ID:               item.ID,
			POID:             item.POID,
			VariantID:        item.VariantID,
			ProductName:      item.ProductName,
			ProductSKU:       item.ProductSKU,
			Quantity:         item.Quantity,
			UnitPrice:        item.UnitPrice,
			LineTotal:        item.LineTotal,
			ReceivedQuantity: item.ReceivedQuantity,
			CreatedAt:        item.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	response.Items = items

	return response, nil
}

// isValidPOStatus validates purchase order status
func isValidPOStatus(status string) bool {
	validStatuses := []string{"placed", "confirmed", "out_for_delivery", "delivered", "verified", "paid"}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// isValidPOStatusTransition validates status transition
func isValidPOStatusTransition(from, to string) bool {
	// Define valid transitions
	transitions := map[string][]string{
		"placed":           {"confirmed"},
		"confirmed":        {"out_for_delivery"},
		"out_for_delivery": {"delivered"},
		"delivered":        {"verified"},
		"verified":         {"paid"},
	}

	validNextStatuses, ok := transitions[from]
	if !ok {
		return false
	}

	for _, validStatus := range validNextStatuses {
		if validStatus == to {
			return true
		}
	}

	return false
}

// isValidPaymentStatus validates payment status
func isValidPaymentStatus(status string) bool {
	validStatuses := []string{"unpaid", "partial", "paid"}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}
