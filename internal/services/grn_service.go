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

// GRNService handles goods receipt note business logic
type GRNService struct {
	grnRepo       *repositories.GRNRepository
	poRepo        *repositories.PurchaseOrderRepository
	warehouseRepo *repositories.WarehouseRepository
	productRepo   *repositories.ProductRepository
	inventoryRepo *repositories.InventoryRepository
	logger        interfaces.Logger
}

// NewGRNService creates a new GRN service
func NewGRNService(
	grnRepo *repositories.GRNRepository,
	poRepo *repositories.PurchaseOrderRepository,
	warehouseRepo *repositories.WarehouseRepository,
	productRepo *repositories.ProductRepository,
	inventoryRepo *repositories.InventoryRepository,
	logger interfaces.Logger,
) *GRNService {
	return &GRNService{
		grnRepo:       grnRepo,
		poRepo:        poRepo,
		warehouseRepo: warehouseRepo,
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
		logger:        logger,
	}
}

// CreateGRN creates a new goods receipt note and inventory batches
func (s *GRNService) CreateGRN(ctx context.Context, request *models.CreateGRNRequest) (*models.GRNResponse, error) {
	s.logger.Info("Creating manual GRN",
		zap.String("grn_number", request.GRNNumber),
		zap.String("po_id", request.POID),
		zap.String("received_by", request.ReceivedBy),
		zap.Int("items_count", len(request.Items)))

	// Validate purchase order exists with items
	po, err := s.poRepo.GetByIDWithItems(request.POID)
	if err != nil {
		s.logger.Error("Failed to retrieve purchase order for GRN",
			zap.Error(err),
			zap.String("po_id", request.POID))
		return nil, err
	}

	// Validate PO status is delivered
	if po.Status != "delivered" {
		s.logger.Warn("Attempted to create GRN for non-delivered PO",
			zap.String("po_id", request.POID),
			zap.String("po_status", po.Status))
		return nil, errors.NewBadRequestError("Purchase order must be in 'delivered' status to create GRN")
	}

	// Check if GRN already exists for this PO
	exists, err := s.grnRepo.GRNExistsForPO(request.POID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.NewConflictError("GRN already exists for this purchase order")
	}

	// Validate warehouse exists
	_, err = s.warehouseRepo.GetByID(po.WarehouseID)
	if err != nil {
		return nil, err
	}

	// Validate quality status
	if !isValidQualityStatus(request.QualityStatus) {
		return nil, errors.NewValidationError(fmt.Sprintf("Invalid quality status: %s", request.QualityStatus))
	}

	// Parse received date
	var receivedDate time.Time
	if request.ReceivedDate != nil {
		receivedDate, err = time.Parse("2006-01-02", *request.ReceivedDate)
		if err != nil {
			return nil, errors.NewValidationError(fmt.Sprintf("Invalid received_date format: %v", err))
		}
	} else {
		receivedDate = time.Now().UTC()
	}

	// Validate items and map to PO items
	if len(request.Items) == 0 {
		return nil, errors.NewValidationError("GRN must have at least one item")
	}

	poItemMap := make(map[string]*models.PurchaseOrderItem)
	for i := range po.Items {
		poItemMap[po.Items[i].ID] = &po.Items[i]
	}

	// Validate all GRN items reference valid PO items
	for _, grnItem := range request.Items {
		if _, exists := poItemMap[grnItem.POItemID]; !exists {
			return nil, errors.NewValidationError(fmt.Sprintf("Invalid PO item ID: %s", grnItem.POItemID))
		}

		// Validate quantities
		if grnItem.ReceivedQuantity <= 0 {
			return nil, errors.NewValidationError("Received quantity must be greater than 0")
		}
		if grnItem.AcceptedQuantity < 0 {
			return nil, errors.NewValidationError("Accepted quantity cannot be negative")
		}
		if grnItem.AcceptedQuantity > grnItem.ReceivedQuantity {
			return nil, errors.NewValidationError("Accepted quantity cannot exceed received quantity")
		}
	}

	// Use user-provided GRN number (instead of auto-generation)
	grnNumber := request.GRNNumber

	// Validate GRN number is unique
	exists, err = s.grnRepo.GRNNumberExists(grnNumber)
	if err != nil {
		s.logger.Error("Failed to check GRN number uniqueness",
			zap.Error(err),
			zap.String("grn_number", grnNumber))
		return nil, err
	}
	if exists {
		s.logger.Warn("Duplicate GRN number provided",
			zap.String("grn_number", grnNumber))
		return nil, errors.NewConflictError(fmt.Sprintf("GRN number '%s' already exists", grnNumber))
	}

	s.logger.Debug("GRN number validation passed",
		zap.String("grn_number", grnNumber))

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
				return errors.NewValidationError(fmt.Sprintf("Invalid expiry_date format for item %s: %v", itemReq.POItemID, err))
			}

			// Create GRN item
			grnItem := models.NewGRNItem(
				grn.ID,
				itemReq.POItemID,
				poItem.VariantID,
				poItem.Quantity,
				itemReq.ReceivedQuantity,
				itemReq.AcceptedQuantity,
				expiryDate,
			)
			grnItem.BatchNumber = itemReq.BatchNumber

			// Create inventory batch for accepted quantity
			if itemReq.AcceptedQuantity > 0 {
				s.logger.Debug("Creating inventory batch from manual GRN",
					zap.String("grn_number", grnNumber),
					zap.String("variant_id", poItem.VariantID),
					zap.Int64("accepted_qty", itemReq.AcceptedQuantity),
					zap.Float64("cost_price", poItem.UnitPrice),
					zap.String("expiry_date", expiryDate.Format("2006-01-02")),
					zap.Stringp("batch_number", itemReq.BatchNumber))

				// Create inventory batch with ALL-IN cost price from PO
				// For procurement, we use the PO unit price as cost price
				// Tax rates are 0 because PO price already includes all taxes
				batch := models.NewInventoryBatch(
					po.WarehouseID,
					poItem.VariantID,
					poItem.UnitPrice, // ALL-IN cost price from PO
					expiryDate,
					itemReq.AcceptedQuantity,
					0,          // CGST rate 0 (price is ALL-IN)
					0,          // SGST rate 0 (price is ALL-IN)
					[]string{}, // No custom taxes
					false,      // Not tax exempt
				)

				if err := s.inventoryRepo.CreateBatchWithTx(tx, batch); err != nil {
					s.logger.Error("Failed to create inventory batch in manual GRN",
						zap.Error(err),
						zap.String("grn_number", grnNumber))
					return err
				}

				s.logger.Debug("Inventory batch created from manual GRN",
					zap.String("batch_id", batch.ID),
					zap.String("grn_number", grnNumber))

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
				if err := s.inventoryRepo.CreateTransactionWithTx(tx, transaction); err != nil {
					s.logger.Error("Failed to create inventory transaction",
						zap.Error(err),
						zap.String("batch_id", batch.ID))
					return err
				}

				s.logger.Debug("Inventory transaction created",
					zap.String("batch_id", batch.ID),
					zap.Int64("quantity", itemReq.AcceptedQuantity))
			}

			// Save GRN item
			if err := s.grnRepo.CreateItemWithTx(tx, grnItem); err != nil {
				return err
			}

			// Update PO item received quantity
			if err := s.poRepo.UpdateItemReceivedQuantityWithTx(tx, poItem.ID, itemReq.ReceivedQuantity); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to create GRN in transaction",
			zap.Error(err),
			zap.String("grn_number", grnNumber))
		return nil, err
	}

	s.logger.Info("Manual GRN created successfully",
		zap.String("grn_id", grn.ID),
		zap.String("grn_number", grn.GRNNumber),
		zap.String("quality_status", grn.QualityStatus))

	// Fetch complete GRN with items
	return s.GetGRN(ctx, grn.ID)
}

// GetGRN retrieves a GRN by ID with items
func (s *GRNService) GetGRN(ctx context.Context, id string) (*models.GRNResponse, error) {
	s.logger.Debug("Retrieving GRN",
		zap.String("grn_id", id))

	grn, err := s.grnRepo.GetByIDWithItems(id)
	if err != nil {
		s.logger.Error("Failed to retrieve GRN",
			zap.Error(err),
			zap.String("grn_id", id))
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

// UpdateGRN updates a GRN
func (s *GRNService) UpdateGRN(ctx context.Context, id string, request *models.UpdateGRNRequest) (*models.GRNResponse, error) {
	s.logger.Info("Updating GRN",
		zap.String("grn_id", id),
		zap.Bool("has_document", request.GRNDocument != nil),
		zap.Bool("has_quality_status", request.QualityStatus != nil))

	// Validate GRN exists
	_, err := s.grnRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve GRN for update",
			zap.Error(err),
			zap.String("grn_id", id))
		return nil, err
	}

	// Build updates map
	updates := make(map[string]interface{})
	if request.GRNDocument != nil {
		updates["grn_document"] = *request.GRNDocument
	}
	if request.Remarks != nil {
		updates["remarks"] = *request.Remarks
	}
	if request.QualityStatus != nil {
		if !isValidQualityStatus(*request.QualityStatus) {
			return nil, errors.NewValidationError(fmt.Sprintf("Invalid quality status: %s", *request.QualityStatus))
		}
		updates["quality_status"] = *request.QualityStatus
	}

	// Update GRN
	if err := s.grnRepo.Update(id, updates); err != nil {
		return nil, err
	}

	// Return updated GRN
	return s.GetGRN(ctx, id)
}

// buildGRNResponse builds a response with related entity details
func (s *GRNService) buildGRNResponse(grn *models.GRN) (*models.GRNResponse, error) {
	response := &models.GRNResponse{
		ID:            grn.ID,
		GRNNumber:     grn.GRNNumber,
		GRNDocument:   grn.GRNDocument, // Attachment ID for vendor's GRN PDF
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
			ID:          item.ID,
			GRNID:       item.GRNID,
			POItemID:    item.POItemID,
			VariantID:   item.VariantID,
			ProductName: item.Variant.VariantName, // Using variant name instead of product name
			ProductSKU: func() string {
				if item.Variant.SKU != nil {
					return *item.Variant.SKU
				}
				return ""
			}(),
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
