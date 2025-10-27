package services

import (
	"context"
	"fmt"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"

	"gorm.io/gorm"
)

// GRNService handles goods receipt note business logic
type GRNService struct {
	grnRepo       *repositories.GRNRepository
	poRepo        *repositories.PurchaseOrderRepository
	warehouseRepo *repositories.WarehouseRepository
	productRepo   *repositories.ProductRepository
	inventoryRepo *repositories.InventoryRepository
}

// NewGRNService creates a new GRN service
func NewGRNService(
	grnRepo *repositories.GRNRepository,
	poRepo *repositories.PurchaseOrderRepository,
	warehouseRepo *repositories.WarehouseRepository,
	productRepo *repositories.ProductRepository,
	inventoryRepo *repositories.InventoryRepository,
) *GRNService {
	return &GRNService{
		grnRepo:       grnRepo,
		poRepo:        poRepo,
		warehouseRepo: warehouseRepo,
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
	}
}

// CreateGRN creates a new goods receipt note and inventory batches
func (s *GRNService) CreateGRN(ctx context.Context, request *models.CreateGRNRequest) (*models.GRNResponse, error) {
	// Validate purchase order exists with items
	po, err := s.poRepo.GetByIDWithItems(request.POID)
	if err != nil {
		return nil, err
	}

	// Validate PO status is delivered
	if po.Status != "delivered" {
		return nil, fmt.Errorf("purchase order must be in 'delivered' status to create GRN")
	}

	// Check if GRN already exists for this PO
	exists, err := s.grnRepo.GRNExistsForPO(request.POID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("GRN already exists for this purchase order")
	}

	// Validate warehouse exists
	warehouse, err := s.warehouseRepo.GetByID(po.WarehouseID)
	if err != nil {
		return nil, err
	}

	// Validate quality status
	if !isValidQualityStatus(request.QualityStatus) {
		return nil, fmt.Errorf("invalid quality status: %s", request.QualityStatus)
	}

	// Parse received date
	var receivedDate time.Time
	if request.ReceivedDate != nil {
		receivedDate, err = time.Parse("2006-01-02", *request.ReceivedDate)
		if err != nil {
			return nil, fmt.Errorf("invalid received_date format: %w", err)
		}
	} else {
		receivedDate = time.Now().UTC()
	}

	// Validate items and map to PO items
	if len(request.Items) == 0 {
		return nil, fmt.Errorf("GRN must have at least one item")
	}

	poItemMap := make(map[string]*models.PurchaseOrderItem)
	for i := range po.Items {
		poItemMap[po.Items[i].ID] = &po.Items[i]
	}

	// Validate all GRN items reference valid PO items
	for _, grnItem := range request.Items {
		if _, exists := poItemMap[grnItem.POItemID]; !exists {
			return nil, fmt.Errorf("invalid PO item ID: %s", grnItem.POItemID)
		}

		// Validate quantities
		if grnItem.ReceivedQuantity <= 0 {
			return nil, fmt.Errorf("received quantity must be greater than 0")
		}
		if grnItem.AcceptedQuantity < 0 {
			return nil, fmt.Errorf("accepted quantity cannot be negative")
		}
		if grnItem.AcceptedQuantity > grnItem.ReceivedQuantity {
			return nil, fmt.Errorf("accepted quantity cannot exceed received quantity")
		}
	}

	// Generate GRN number
	grnNumber, err := s.generateGRNNumber()
	if err != nil {
		return nil, err
	}

	// Create GRN, items, and inventory batches in a transaction
	var grn *models.GRN
	err = s.grnRepo.WithTransaction(func(tx *gorm.DB) error {
		// Create GRN
		grn = models.NewGRN(
			grnNumber,
			request.POID,
			po.WarehouseID,
			request.ReceivedBy,
			receivedDate,
			request.QualityStatus,
		)
		grn.Remarks = request.Remarks

		if err := s.grnRepo.CreateWithTx(tx, grn); err != nil {
			return err
		}

		// Create GRN items and inventory batches
		for _, itemReq := range request.Items {
			poItem := poItemMap[itemReq.POItemID]

			// Parse expiry date
			expiryDate, err := time.Parse("2006-01-02", itemReq.ExpiryDate)
			if err != nil {
				return fmt.Errorf("invalid expiry_date format for item %s: %w", itemReq.POItemID, err)
			}

			// Create GRN item
			grnItem := models.NewGRNItem(
				grn.ID,
				itemReq.POItemID,
				poItem.ProductID,
				poItem.Quantity,
				itemReq.ReceivedQuantity,
				itemReq.AcceptedQuantity,
				expiryDate,
			)
			grnItem.BatchNumber = itemReq.BatchNumber

			// Create inventory batch for accepted quantity
			if itemReq.AcceptedQuantity > 0 {
				// Create inventory batch with ALL-IN cost price from PO
				// For procurement, we use the PO unit price as cost price
				// Tax rates are 0 because PO price already includes all taxes
				batch := models.NewInventoryBatch(
					po.WarehouseID,
					poItem.ProductID,
					poItem.UnitPrice, // ALL-IN cost price from PO
					expiryDate,
					itemReq.AcceptedQuantity,
					0, // CGST rate 0 (price is ALL-IN)
					0, // SGST rate 0 (price is ALL-IN)
					[]string{}, // No custom taxes
					false, // Not tax exempt
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
					itemReq.AcceptedQuantity,
					&grn.ID,
					&request.ReceivedBy,
					&note,
					receivedDate,
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
			if err := s.poRepo.UpdateItemReceivedQuantity(poItem.ID, itemReq.ReceivedQuantity); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Fetch complete GRN with items
	return s.GetGRN(ctx, grn.ID)
}

// GetGRN retrieves a GRN by ID with items
func (s *GRNService) GetGRN(ctx context.Context, id string) (*models.GRNResponse, error) {
	grn, err := s.grnRepo.GetByIDWithItems(id)
	if err != nil {
		return nil, err
	}

	return s.buildGRNResponse(grn)
}

// GetAllGRNs retrieves all GRNs
func (s *GRNService) GetAllGRNs(ctx context.Context) ([]models.GRNResponse, error) {
	grns, err := s.grnRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []models.GRNResponse
	for _, grn := range grns {
		// Get with items
		grnWithItems, err := s.grnRepo.GetByIDWithItems(grn.ID)
		if err != nil {
			continue
		}
		response, err := s.buildGRNResponse(grnWithItems)
		if err != nil {
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// GetGRNsByWarehouse retrieves GRNs by warehouse
func (s *GRNService) GetGRNsByWarehouse(ctx context.Context, warehouseID string) ([]models.GRNResponse, error) {
	// Validate warehouse exists
	_, err := s.warehouseRepo.GetByID(warehouseID)
	if err != nil {
		return nil, err
	}

	grns, err := s.grnRepo.GetByWarehouse(warehouseID)
	if err != nil {
		return nil, err
	}

	var responses []models.GRNResponse
	for _, grn := range grns {
		// Get with items
		grnWithItems, err := s.grnRepo.GetByIDWithItems(grn.ID)
		if err != nil {
			continue
		}
		response, err := s.buildGRNResponse(grnWithItems)
		if err != nil {
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// GetGRNByPurchaseOrder retrieves GRN by purchase order ID
func (s *GRNService) GetGRNByPurchaseOrder(ctx context.Context, poID string) (*models.GRNResponse, error) {
	// Validate PO exists
	_, err := s.poRepo.GetByID(poID)
	if err != nil {
		return nil, err
	}

	grn, err := s.grnRepo.GetByPurchaseOrder(poID)
	if err != nil {
		return nil, err
	}

	// Get with items
	grnWithItems, err := s.grnRepo.GetByIDWithItems(grn.ID)
	if err != nil {
		return nil, err
	}

	return s.buildGRNResponse(grnWithItems)
}

// generateGRNNumber generates a unique GRN number in format: GRN-YYYY-NNNN
func (s *GRNService) generateGRNNumber() (string, error) {
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

// buildGRNResponse builds a response with related entity details
func (s *GRNService) buildGRNResponse(grn *models.GRN) (*models.GRNResponse, error) {
	response := &models.GRNResponse{
		ID:            grn.ID,
		GRNNumber:     grn.GRNNumber,
		POID:          grn.POID,
		PONumber:      grn.PurchaseOrder.PONumber,
		WarehouseID:   grn.WarehouseID,
		WarehouseName: grn.Warehouse.Name,
		ReceivedDate:  grn.ReceivedDate.Format("2006-01-02"),
		ReceivedBy:    grn.ReceivedBy,
		QualityStatus: grn.QualityStatus,
		Remarks:       grn.Remarks,
		CreatedAt:     grn.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     grn.UpdatedAt.UTC().Format(time.RFC3339),
	}

	// Add items
	var items []models.GRNItemResponse
	for _, item := range grn.Items {
		items = append(items, models.GRNItemResponse{
			ID:               item.ID,
			GRNID:            item.GRNID,
			POItemID:         item.POItemID,
			ProductID:        item.ProductID,
			ProductName:      item.Product.Name,
			ProductSKU:       item.Product.SKU,
			OrderedQuantity:  item.OrderedQuantity,
			ReceivedQuantity: item.ReceivedQuantity,
			AcceptedQuantity: item.AcceptedQuantity,
			RejectedQuantity: item.RejectedQuantity,
			ExpiryDate:       item.ExpiryDate.Format("2006-01-02"),
			BatchNumber:      item.BatchNumber,
			InventoryBatchID: item.InventoryBatchID,
			CreatedAt:        item.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	response.Items = items

	return response, nil
}

// isValidQualityStatus validates quality status
func isValidQualityStatus(status string) bool {
	validStatuses := []string{"accepted", "rejected", "partial"}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}
