package services

import (
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"

	"go.uber.org/zap"
)

// Helper function to convert string to pointer (local to this package)
func stringPtrReturns(s string) *string {
	return &s
}

type ReturnsService struct {
	returnsRepo   *repositories.ReturnsRepository
	salesRepo     *repositories.SalesRepository
	inventoryRepo *repositories.InventoryRepository
	logger        interfaces.Logger
}

func NewReturnsService(returnsRepo *repositories.ReturnsRepository, salesRepo *repositories.SalesRepository, inventoryRepo *repositories.InventoryRepository, logger interfaces.Logger) *ReturnsService {
	return &ReturnsService{
		returnsRepo:   returnsRepo,
		salesRepo:     salesRepo,
		inventoryRepo: inventoryRepo,
		logger:        logger,
	}
}

// CreateReturn creates a new return with items and summary
func (s *ReturnsService) CreateReturn(req *models.CreateReturnRequest) (*models.ReturnResponse, error) {
	s.logger.Info("Creating return",
		zap.String("sale_id", req.SaleID),
		zap.Int("item_count", len(req.Items)))

	// Validate return request
	if err := s.validateReturnRequest(req); err != nil {
		s.logger.Error("Return validation failed",
			zap.Error(err),
			zap.String("sale_id", req.SaleID))
		return nil, err
	}
	s.logger.Debug("Return validation passed")

	// Validate original sale exists
	_, err := s.salesRepo.GetSaleByID(req.SaleID)
	if err != nil {
		s.logger.Error("Original sale not found",
			zap.Error(err),
			zap.String("sale_id", req.SaleID))
		return nil, errors.NewNotFoundError("Original sale")
	}
	s.logger.Debug("Original sale found")

	// Parse return date or use current time
	var returnDate time.Time
	if req.ReturnDate != nil {
		s.logger.Debug("Parsing return date",
			zap.String("return_date", *req.ReturnDate))
		if parsedDate, err := time.Parse("2006-01-02T15:04:05Z", *req.ReturnDate); err == nil {
			returnDate = parsedDate
			s.logger.Debug("Return date parsed successfully",
				zap.Time("return_date", returnDate))
		} else {
			// If parsing fails, use current time
			s.logger.Warn("Date parsing failed, using current time",
				zap.Error(err))
			returnDate = time.Now()
		}
	} else {
		// If no return date provided, use current time
		s.logger.Debug("No return date provided, using current time")
		returnDate = time.Now()
	}

	// Create return using the proper constructor
	ret := models.NewReturn(req.SaleID, returnDate, 0, "pending")

	s.logger.Debug("Saving return to database")
	if err := s.returnsRepo.CreateReturn(ret); err != nil {
		s.logger.Error("Failed to create return",
			zap.Error(err),
			zap.String("sale_id", req.SaleID))
		return nil, err
	}
	s.logger.Info("Return created",
		zap.String("return_id", ret.ID))

	// Create return items
	s.logger.Debug("Processing return items",
		zap.Int("item_count", len(req.Items)))
	var totalAmount float64
	for i, itemReq := range req.Items {
		s.logger.Debug("Processing return item",
			zap.Int("item_number", i+1),
			zap.String("batch_id", itemReq.BatchID),
			zap.Int64("quantity", itemReq.Quantity),
			zap.Float64("refund_amount", itemReq.RefundAmount))

		// Validate sale item exists
		saleItems, err := s.salesRepo.GetSaleItemsBySaleID(req.SaleID)
		if err != nil {
			s.logger.Error("Failed to get sale items",
				zap.Error(err),
				zap.String("sale_id", req.SaleID))
			return nil, err
		}

		// Find matching sale item
		var matchingSaleItem *models.SaleItem
		for _, saleItem := range saleItems {
			if saleItem.BatchID == itemReq.BatchID {
				matchingSaleItem = &saleItem
				break
			}
		}

		if matchingSaleItem == nil {
			s.logger.Error("Sale item not found for batch",
				zap.String("batch_id", itemReq.BatchID),
				zap.String("sale_id", req.SaleID))
			return nil, errors.NewNotFoundError("Sale item for batch")
		}

		// Validate return quantity
		if itemReq.Quantity > matchingSaleItem.Quantity {
			s.logger.Error("Return quantity exceeds original sale quantity",
				zap.Int64("return_quantity", itemReq.Quantity),
				zap.Int64("sale_quantity", matchingSaleItem.Quantity),
				zap.String("batch_id", itemReq.BatchID))
			return nil, errors.NewBadRequestError("Return quantity cannot exceed original sale quantity")
		}
		s.logger.Debug("Return quantity validated against sale",
			zap.Int64("return_quantity", itemReq.Quantity),
			zap.Int64("sale_quantity", matchingSaleItem.Quantity))

		// Calculate item total
		itemTotal := itemReq.RefundAmount * float64(itemReq.Quantity)
		s.logger.Debug("Calculated refund for item",
			zap.Float64("item_total", itemTotal))

		// Create return item using the proper constructor
		returnItem := models.NewReturnItem(ret.ID, itemReq.BatchID, itemReq.Quantity, itemReq.RefundAmount)

		if err := s.returnsRepo.CreateReturnItem(returnItem); err != nil {
			s.logger.Error("Failed to create return item",
				zap.Error(err),
				zap.String("return_id", ret.ID),
				zap.String("batch_id", itemReq.BatchID))
			return nil, err
		}
		s.logger.Debug("Return item created",
			zap.String("return_item_id", returnItem.ID))

		// Update inventory (restore stock) using the proper constructor
		s.logger.Debug("Restoring inventory stock",
			zap.String("batch_id", itemReq.BatchID),
			zap.Int64("quantity", itemReq.Quantity))
		transaction := models.NewInventoryTransaction(itemReq.BatchID, "return", itemReq.Quantity, &ret.ID, nil, stringPtrReturns("Return transaction"), time.Now())

		if err := s.inventoryRepo.CreateTransaction(transaction); err != nil {
			s.logger.Error("Failed to create inventory transaction",
				zap.Error(err),
				zap.String("batch_id", itemReq.BatchID))
			return nil, err
		}
		s.logger.Debug("Inventory transaction created",
			zap.String("transaction_id", transaction.ID))

		// Update batch stock level (restore stock)
		if err := s.inventoryRepo.UpdateBatchStock(itemReq.BatchID, itemReq.Quantity); err != nil {
			s.logger.Error("Failed to update batch stock",
				zap.Error(err),
				zap.String("batch_id", itemReq.BatchID))
			return nil, err
		}
		s.logger.Info("Inventory restored",
			zap.String("batch_id", itemReq.BatchID),
			zap.Int64("quantity_restored", itemReq.Quantity))

		totalAmount += itemTotal
	}

	// Update return with total refund amount
	s.logger.Debug("Updating return with total refund amount",
		zap.Float64("total_refund", totalAmount))
	ret.TotalRefund = totalAmount
	if err := s.returnsRepo.UpdateReturn(ret); err != nil {
		s.logger.Error("Failed to update return with total refund",
			zap.Error(err),
			zap.String("return_id", ret.ID))
		return nil, err
	}

	// Get complete return with items and summary
	completeReturn, err := s.returnsRepo.GetReturnByID(ret.ID)
	if err != nil {
		s.logger.Error("Failed to get complete return",
			zap.Error(err),
			zap.String("return_id", ret.ID))
		return nil, err
	}

	s.logger.Info("Return creation completed successfully",
		zap.String("return_id", ret.ID),
		zap.Float64("total_refund", totalAmount))

	return s.mapReturnToResponse(completeReturn), nil
}

// GetReturn retrieves a return by ID
func (s *ReturnsService) GetReturn(id string) (*models.ReturnResponse, error) {
	ret, err := s.returnsRepo.GetReturnByID(id)
	if err != nil {
		return nil, err
	}
	return s.mapReturnToResponse(ret), nil
}

// GetAllReturns retrieves all returns with pagination
func (s *ReturnsService) GetAllReturns(limit, offset int) ([]models.ReturnResponse, error) {
	returns, err := s.returnsRepo.GetAllReturns(limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []models.ReturnResponse
	for _, ret := range returns {
		responses = append(responses, *s.mapReturnToResponse(&ret))
	}

	return responses, nil
}

// UpdateReturn updates a return
func (s *ReturnsService) UpdateReturn(id string, req *models.UpdateReturnRequest) (*models.ReturnResponse, error) {
	s.logger.Info("Updating return",
		zap.String("return_id", id))

	ret, err := s.returnsRepo.GetReturnByID(id)
	if err != nil {
		s.logger.Error("Return not found",
			zap.Error(err),
			zap.String("return_id", id))
		return nil, err
	}

	// Update fields
	if req.Status != nil {
		s.logger.Debug("Updating return status",
			zap.String("old_status", ret.Status),
			zap.String("new_status", *req.Status))
		ret.Status = *req.Status
	}

	if err := s.returnsRepo.UpdateReturn(ret); err != nil {
		s.logger.Error("Failed to update return",
			zap.Error(err),
			zap.String("return_id", id))
		return nil, err
	}

	s.logger.Info("Return updated successfully",
		zap.String("return_id", id))

	return s.mapReturnToResponse(ret), nil
}

// DeleteReturn deletes a return
func (s *ReturnsService) DeleteReturn(id string) error {
	s.logger.Info("Deleting return",
		zap.String("return_id", id))

	err := s.returnsRepo.DeleteReturn(id)
	if err != nil {
		s.logger.Error("Failed to delete return",
			zap.Error(err),
			zap.String("return_id", id))
		return err
	}

	s.logger.Info("Return deleted successfully",
		zap.String("return_id", id))
	return nil
}

// GetReturnsByCustomer retrieves returns for a specific customer
func (s *ReturnsService) GetReturnsByCustomer(customerID string) ([]models.ReturnResponse, error) {
	returns, err := s.returnsRepo.GetReturnsByCustomer(customerID)
	if err != nil {
		return nil, err
	}

	var responses []models.ReturnResponse
	for _, ret := range returns {
		responses = append(responses, *s.mapReturnToResponse(&ret))
	}

	return responses, nil
}

// GetReturnsBySaleID retrieves returns for a specific sale
func (s *ReturnsService) GetReturnsBySaleID(saleID string) ([]models.ReturnResponse, error) {
	returns, err := s.returnsRepo.GetReturnsBySaleID(saleID)
	if err != nil {
		return nil, err
	}

	var responses []models.ReturnResponse
	for _, ret := range returns {
		responses = append(responses, *s.mapReturnToResponse(&ret))
	}

	return responses, nil
}

// GetReturnsByDateRange retrieves returns within a date range
func (s *ReturnsService) GetReturnsByDateRange(startDate, endDate time.Time) ([]models.ReturnResponse, error) {
	returns, err := s.returnsRepo.GetReturnsByDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	var responses []models.ReturnResponse
	for _, ret := range returns {
		responses = append(responses, *s.mapReturnToResponse(&ret))
	}

	return responses, nil
}

// GetReturnsByStatus retrieves returns by status
func (s *ReturnsService) GetReturnsByStatus(status string) ([]models.ReturnResponse, error) {
	returns, err := s.returnsRepo.GetReturnsByStatus(status)
	if err != nil {
		return nil, err
	}

	var responses []models.ReturnResponse
	for _, ret := range returns {
		responses = append(responses, *s.mapReturnToResponse(&ret))
	}

	return responses, nil
}

// GetTotalReturnsAmount calculates total returns amount for a date range
func (s *ReturnsService) GetTotalReturnsAmount(startDate, endDate time.Time) (float64, error) {
	return s.returnsRepo.GetTotalReturnsAmount(startDate, endDate)
}

// GetReturnRateByProduct calculates return rate for a product
func (s *ReturnsService) GetReturnRateByProduct(productID string, startDate, endDate time.Time) (float64, error) {
	return s.returnsRepo.GetReturnRateByProduct(productID, startDate, endDate)
}

// GetMostReturnedProducts retrieves most returned products
func (s *ReturnsService) GetMostReturnedProducts(limit int) ([]models.MostReturnedProductResponse, error) {
	results, err := s.returnsRepo.GetMostReturnedProducts(limit)
	if err != nil {
		return nil, err
	}

	var responses []models.MostReturnedProductResponse
	for _, result := range results {
		responses = append(responses, models.MostReturnedProductResponse{
			ProductID:     result.ProductID,
			ProductName:   result.ProductName,
			TotalReturned: result.TotalReturned,
			ReturnAmount:  result.ReturnAmount,
		})
	}

	return responses, nil
}

// Helper methods
func (s *ReturnsService) validateReturnRequest(req *models.CreateReturnRequest) error {
	s.logger.Debug("Validating return request",
		zap.String("sale_id", req.SaleID),
		zap.Int("item_count", len(req.Items)))

	if req.SaleID == "" {
		s.logger.Error("Validation failed: sale ID is empty")
		return errors.NewValidationError("Sale ID is required")
	}
	if len(req.Items) == 0 {
		s.logger.Error("Validation failed: no items provided")
		return errors.NewValidationError("At least one item is required")
	}

	for i, item := range req.Items {
		s.logger.Debug("Validating return item",
			zap.Int("item_number", i+1),
			zap.String("batch_id", item.BatchID),
			zap.Int64("quantity", item.Quantity))

		if item.BatchID == "" {
			s.logger.Error("Validation failed: batch ID is empty",
				zap.Int("item_number", i+1))
			return errors.NewValidationError("Batch ID is required for all items")
		}
		if item.Quantity <= 0 {
			s.logger.Error("Validation failed: quantity <= 0",
				zap.Int("item_number", i+1))
			return errors.NewValidationError("Quantity must be greater than 0")
		}
	}

	s.logger.Debug("Return request validation passed")
	return nil
}

func (s *ReturnsService) mapReturnToResponse(ret *models.Return) *models.ReturnResponse {
	response := &models.ReturnResponse{
		ID:          ret.ID,
		SaleID:      ret.SaleID,
		ReturnDate:  ret.ReturnDate.Format("2006-01-02T15:04:05Z07:00"),
		TotalRefund: ret.TotalRefund,
		Status:      ret.Status,
		CreatedAt:   ret.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   ret.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Map items
	for _, item := range ret.Items {
		response.Items = append(response.Items, models.ReturnItemResponse{
			ID:           item.ID,
			ReturnID:     item.ReturnID,
			BatchID:      item.BatchID,
			Quantity:     item.Quantity,
			RefundAmount: item.RefundAmount,
			CreatedAt:    item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return response
}
