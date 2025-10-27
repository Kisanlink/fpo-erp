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
}

// NewPurchaseOrderService creates a new purchase order service
func NewPurchaseOrderService(
	poRepo *repositories.PurchaseOrderRepository,
	collaboratorRepo *repositories.CollaboratorRepository,
	warehouseRepo *repositories.WarehouseRepository,
	productRepo *repositories.ProductRepository,
) *PurchaseOrderService {
	return &PurchaseOrderService{
		poRepo:           poRepo,
		collaboratorRepo: collaboratorRepo,
		warehouseRepo:    warehouseRepo,
		productRepo:      productRepo,
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
	warehouse, err := s.warehouseRepo.GetByID(request.WarehouseID)
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
func (s *PurchaseOrderService) UpdatePurchaseOrderStatus(ctx context.Context, id string, request *models.UpdatePOStatusRequest) (*models.PurchaseOrderResponse, error) {
	// Validate status
	if !isValidPOStatus(request.Status) {
		return nil, fmt.Errorf("invalid status: %s", request.Status)
	}

	// Get existing PO
	po, err := s.poRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if !isValidPOStatusTransition(po.Status, request.Status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", po.Status, request.Status)
	}

	// Update status
	po.Status = request.Status

	// Set actual delivery date if status is delivered
	if request.Status == "delivered" && request.ActualDelivery != nil {
		po.ActualDelivery = request.ActualDelivery
	} else if request.Status == "delivered" && request.ActualDelivery == nil {
		now := time.Now().UTC()
		po.ActualDelivery = &now
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
