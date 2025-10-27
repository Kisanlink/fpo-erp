package services

import (
	"context"
	"fmt"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"

	"gorm.io/gorm"
)

// PurchaseOrderService handles purchase order business logic
type PurchaseOrderService struct {
	poRepo           *repositories.PurchaseOrderRepository
	collaboratorRepo *repositories.CollaboratorRepository
	warehouseRepo    *repositories.WarehouseRepository
	productRepo      *repositories.ProductRepository
	grnRepo          *repositories.GRNRepository
	inventoryRepo    *repositories.InventoryRepository
}

// NewPurchaseOrderService creates a new purchase order service
func NewPurchaseOrderService(
	poRepo *repositories.PurchaseOrderRepository,
	collaboratorRepo *repositories.CollaboratorRepository,
	warehouseRepo *repositories.WarehouseRepository,
	productRepo *repositories.ProductRepository,
	grnRepo *repositories.GRNRepository,
	inventoryRepo *repositories.InventoryRepository,
) *PurchaseOrderService {
	return &PurchaseOrderService{
		poRepo:           poRepo,
		collaboratorRepo: collaboratorRepo,
		warehouseRepo:    warehouseRepo,
		productRepo:      productRepo,
		grnRepo:          grnRepo,
		inventoryRepo:    inventoryRepo,
	}
}

// CreatePurchaseOrder creates a new purchase order with items
func (s *PurchaseOrderService) CreatePurchaseOrder(ctx context.Context, request *models.CreatePurchaseOrderRequest) (*models.PurchaseOrderResponse, error) {
	// Validate collaborator exists and is active
	collaborator, err := s.collaboratorRepo.GetByID(request.CollaboratorID)
	if err != nil {
		return nil, err
	}
	if !collaborator.IsActive {
		return nil, fmt.Errorf("collaborator is not active")
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
			return nil, fmt.Errorf("invalid order_date format: %w", err)
		}
	} else {
		orderDate = time.Now().UTC()
	}

	expectedDelivery, err := time.Parse("2006-01-02", request.ExpectedDelivery)
	if err != nil {
		return nil, fmt.Errorf("invalid expected_delivery_date format: %w", err)
	}

	// Validate expected delivery is after order date
	if expectedDelivery.Before(orderDate) {
		return nil, fmt.Errorf("expected delivery date must be after order date")
	}

	// Validate items and calculate total
	if len(request.Items) == 0 {
		return nil, fmt.Errorf("purchase order must have at least one item")
	}

	var totalAmount float64
	productDetails := make(map[string]*models.Product)

	for _, item := range request.Items {
		// Validate product exists
		product, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product %s: %w", item.ProductID, err)
		}
		productDetails[item.ProductID] = product

		// Calculate line total
		lineTotal := float64(item.Quantity) * item.UnitPrice
		totalAmount += lineTotal
	}

	// Generate PO number
	poNumber, err := s.generatePONumber()
	if err != nil {
		return nil, err
	}

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
			product := productDetails[itemReq.ProductID]
			item := models.NewPurchaseOrderItem(
				po.ID,
				itemReq.ProductID,
				itemReq.Quantity,
				itemReq.UnitPrice,
			)
			// Snapshot product details
			item.ProductName = &product.Name
			item.ProductSKU = &product.SKU

			if err := s.poRepo.CreateItemWithTx(tx, item); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Fetch complete PO with items
	return s.GetPurchaseOrder(ctx, po.ID)
}

// GetPurchaseOrder retrieves a purchase order by ID with items
func (s *PurchaseOrderService) GetPurchaseOrder(ctx context.Context, id string) (*models.PurchaseOrderResponse, error) {
	po, err := s.poRepo.GetByIDWithItems(id)
	if err != nil {
		return nil, err
	}

	return s.buildPurchaseOrderResponse(po)
}

// GetAllPurchaseOrders retrieves all purchase orders
func (s *PurchaseOrderService) GetAllPurchaseOrders(ctx context.Context) ([]models.PurchaseOrderResponse, error) {
	pos, err := s.poRepo.GetAll()
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
		return nil, fmt.Errorf("invalid status: %s", status)
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
// Supports auto-GRN creation when status = "delivered" with delivery details
func (s *PurchaseOrderService) UpdatePurchaseOrderStatus(ctx context.Context, id string, request *models.UpdatePOStatusRequest, userID string) (*models.PurchaseOrderResponse, error) {
	// Validate status
	if !isValidPOStatus(request.Status) {
		return nil, fmt.Errorf("invalid status: %s", request.Status)
	}

	// Get existing PO with items
	po, err := s.poRepo.GetByIDWithItems(id)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if !isValidPOStatusTransition(po.Status, request.Status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", po.Status, request.Status)
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

	// Pattern Detection: Auto-create GRN if status = "delivered" and delivery details provided
	if request.Status == "delivered" && (request.AcceptAll != nil || len(request.Items) > 0) {
		// Check if GRN already exists for this PO
		grnExists, err := s.grnRepo.GRNExistsForPO(po.ID)
		if err != nil {
			return nil, err
		}
		if grnExists {
			return nil, fmt.Errorf("GRN already exists for this purchase order")
		}

		// Pattern 1: Accept All
		if request.AcceptAll != nil && *request.AcceptAll {
			if err := s.processAcceptAll(ctx, po, actualDelivery, request.DefaultExpiryDate, userID); err != nil {
				return nil, fmt.Errorf("failed to process accept all: %w", err)
			}
		} else if len(request.Items) > 0 {
			// Pattern 2 & 3: Per-item details
			if err := s.processDeliveryItems(ctx, po, actualDelivery, request.Items, userID); err != nil {
				return nil, fmt.Errorf("failed to process delivery items: %w", err)
			}
		}

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
	// Validate payment status
	if !isValidPaymentStatus(request.PaymentStatus) {
		return nil, fmt.Errorf("invalid payment status: %s", request.PaymentStatus)
	}

	// Get existing PO
	po, err := s.poRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Validate paid amount doesn't exceed total
	if request.PaidAmount > po.TotalAmount {
		return nil, fmt.Errorf("paid amount cannot exceed total amount")
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
		return nil, err
	}

	return s.GetPurchaseOrder(ctx, id)
}

// processAcceptAll handles Pattern 1: Accept all items with default expiry
func (s *PurchaseOrderService) processAcceptAll(ctx context.Context, po *models.PurchaseOrder, actualDelivery time.Time, defaultExpiryDate *string, userID string) error {
	// Validate default expiry date
	if defaultExpiryDate == nil || *defaultExpiryDate == "" {
		return fmt.Errorf("default_expiry_date is required when accept_all is true")
	}

	_, err := time.Parse("2006-01-02", *defaultExpiryDate)
	if err != nil {
		return fmt.Errorf("invalid default_expiry_date format: %w", err)
	}

	// Build delivery items from all PO items
	var deliveryItems []models.DeliveryItemRequest
	for _, poItem := range po.Items {
		acceptTrue := true
		deliveryItems = append(deliveryItems, models.DeliveryItemRequest{
			POItemID:         poItem.ID,
			Accept:           &acceptTrue,
			ReceivedQuantity: &poItem.Quantity,
			AcceptedQuantity: &poItem.Quantity,
			ExpiryDate:       *defaultExpiryDate,
			BatchNumber:      nil,
		})
	}

	// Process with standard flow
	return s.processDeliveryItems(ctx, po, actualDelivery, deliveryItems, userID)
}

// processDeliveryItems handles Pattern 2 & 3: Per-item details
func (s *PurchaseOrderService) processDeliveryItems(ctx context.Context, po *models.PurchaseOrder, actualDelivery time.Time, items []models.DeliveryItemRequest, userID string) error {
	// Validate delivery items
	if err := s.validateDeliveryItems(po, items); err != nil {
		return err
	}

	// Generate GRN number
	grnNumber, err := s.generateGRNNumber()
	if err != nil {
		return err
	}

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
				return fmt.Errorf("PO item %s not found", itemReq.POItemID)
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
				return fmt.Errorf("item %s must have either accept field or quantity fields", itemReq.POItemID)
			}

			// Parse expiry date
			expiryDate, err := time.Parse("2006-01-02", itemReq.ExpiryDate)
			if err != nil {
				return fmt.Errorf("invalid expiry_date for item %s: %w", itemReq.POItemID, err)
			}

			// Create GRN item
			grnItem := models.NewGRNItem(
				grn.ID,
				itemReq.POItemID,
				poItem.ProductID,
				poItem.Quantity,
				receivedQty,
				acceptedQty,
				expiryDate,
			)
			grnItem.BatchNumber = itemReq.BatchNumber

			// Create inventory batch for accepted quantity
			if acceptedQty > 0 {
				batch := models.NewInventoryBatch(
					po.WarehouseID,
					poItem.ProductID,
					poItem.UnitPrice, // ALL-IN cost price from PO
					expiryDate,
					acceptedQty,
					0, // CGST rate 0 (price is ALL-IN)
					0, // SGST rate 0 (price is ALL-IN)
					[]string{}, // No custom taxes
					false,      // Not tax exempt
				)

				if err := s.inventoryRepo.CreateBatch(batch); err != nil {
					return err
				}

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
				if err := s.inventoryRepo.CreateTransaction(transaction); err != nil {
					return err
				}
			}

			// Save GRN item
			if err := s.grnRepo.CreateItemWithTx(tx, grnItem); err != nil {
				return err
			}

			// Update PO item received quantity
			if err := s.poRepo.UpdateItemReceivedQuantity(poItem.ID, receivedQty); err != nil {
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
			return fmt.Errorf("duplicate po_item_id: %s", item.POItemID)
		}
		seenItems[item.POItemID] = true

		// Validate item belongs to this PO
		poItem, exists := poItemMap[item.POItemID]
		if !exists {
			return fmt.Errorf("po_item_id %s does not belong to this purchase order", item.POItemID)
		}

		// Validate pattern usage
		hasAccept := item.Accept != nil
		hasQuantities := item.ReceivedQuantity != nil && item.AcceptedQuantity != nil

		if !hasAccept && !hasQuantities {
			return fmt.Errorf("item %s must have either accept field or quantity fields", item.POItemID)
		}

		if hasAccept && hasQuantities {
			return fmt.Errorf("item %s cannot have both accept field and quantity fields", item.POItemID)
		}

		// Validate quantities if using Pattern 3
		if hasQuantities {
			if *item.ReceivedQuantity <= 0 {
				return fmt.Errorf("received_quantity must be greater than 0 for item %s", item.POItemID)
			}
			if *item.AcceptedQuantity < 0 {
				return fmt.Errorf("accepted_quantity cannot be negative for item %s", item.POItemID)
			}
			if *item.AcceptedQuantity > *item.ReceivedQuantity {
				return fmt.Errorf("accepted_quantity cannot exceed received_quantity for item %s", item.POItemID)
			}
			if *item.ReceivedQuantity > poItem.Quantity {
				return fmt.Errorf("received_quantity (%d) cannot exceed ordered quantity (%d) for item %s", *item.ReceivedQuantity, poItem.Quantity, item.POItemID)
			}
		}

		// Validate expiry date format
		if item.ExpiryDate != "" {
			if _, err := time.Parse("2006-01-02", item.ExpiryDate); err != nil {
				return fmt.Errorf("invalid expiry_date format for item %s: must be YYYY-MM-DD", item.POItemID)
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

	return "", fmt.Errorf("failed to generate unique GRN number")
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

	return "", fmt.Errorf("failed to generate unique PO number")
}

// buildPurchaseOrderResponse builds a response with related entity details
func (s *PurchaseOrderService) buildPurchaseOrderResponse(po *models.PurchaseOrder) (*models.PurchaseOrderResponse, error) {
	response := &models.PurchaseOrderResponse{
		ID:               po.ID,
		PONumber:         po.PONumber,
		CollaboratorID:   po.CollaboratorID,
		CollaboratorName: po.Collaborator.CompanyName,
		WarehouseID:      po.WarehouseID,
		WarehouseName:    po.Warehouse.Name,
		OrderDate:        po.OrderDate.Format("2006-01-02"),
		ExpectedDelivery: po.ExpectedDelivery.Format("2006-01-02"),
		Status:           po.Status,
		TotalAmount:      po.TotalAmount,
		PaymentStatus:    po.PaymentStatus,
		PaidAmount:       po.PaidAmount,
		CreatedAt:        po.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        po.UpdatedAt.UTC().Format(time.RFC3339),
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
			ProductID:        item.ProductID,
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
	validStatuses := []string{"placed", "confirmed", "out_for_delivery", "delivered", "paid"}
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
		"delivered":        {"paid"},
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
