package services

import (
	"errors"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
)

// Helper function to convert string to pointer (local to this package)
func stringPtrReturns(s string) *string {
	return &s
}

type ReturnsService struct {
	returnsRepo   *repositories.ReturnsRepository
	salesRepo     *repositories.SalesRepository
	inventoryRepo *repositories.InventoryRepository
}

func NewReturnsService(returnsRepo *repositories.ReturnsRepository, salesRepo *repositories.SalesRepository, inventoryRepo *repositories.InventoryRepository) *ReturnsService {
	return &ReturnsService{
		returnsRepo:   returnsRepo,
		salesRepo:     salesRepo,
		inventoryRepo: inventoryRepo,
	}
}

// CreateReturn creates a new return with items and summary
func (s *ReturnsService) CreateReturn(req *models.CreateReturnRequest) (*models.ReturnResponse, error) {
	// Validate return request
	if err := s.validateReturnRequest(req); err != nil {
		return nil, err
	}

	// Validate original sale exists
	_, err := s.salesRepo.GetSaleByID(req.SaleID)
	if err != nil {
		return nil, errors.New("original sale not found")
	}

	// Parse return date or use current time
	var returnDate time.Time
	if req.ReturnDate != nil {
		if parsedDate, err := time.Parse("2006-01-02T15:04:05Z", *req.ReturnDate); err == nil {
			returnDate = parsedDate
		} else {
			// If parsing fails, use current time
			returnDate = time.Now()
		}
	} else {
		// If no return date provided, use current time
		returnDate = time.Now()
	}

	// Create return using the proper constructor
	ret := models.NewReturn(req.SaleID, returnDate, 0, "pending")

	if err := s.returnsRepo.CreateReturn(ret); err != nil {
		return nil, err
	}

	// Create return items
	var totalAmount float64
	for _, itemReq := range req.Items {
		// Validate sale item exists
		saleItems, err := s.salesRepo.GetSaleItemsBySaleID(req.SaleID)
		if err != nil {
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
			return nil, errors.New("sale item not found for batch")
		}

		if itemReq.Quantity > matchingSaleItem.Quantity {
			return nil, errors.New("return quantity cannot exceed original sale quantity")
		}

		// Calculate item total
		itemTotal := itemReq.RefundAmount * float64(itemReq.Quantity)

		// Create return item using the proper constructor
		returnItem := models.NewReturnItem(ret.ID, itemReq.BatchID, itemReq.Quantity, itemReq.RefundAmount)

		if err := s.returnsRepo.CreateReturnItem(returnItem); err != nil {
			return nil, err
		}

		// Update inventory (restore stock) using the proper constructor
		transaction := models.NewInventoryTransaction(itemReq.BatchID, "return", itemReq.Quantity, &ret.ID, nil, stringPtrReturns("Return transaction"), time.Now())

		if err := s.inventoryRepo.CreateTransaction(transaction); err != nil {
			return nil, err
		}

		// Update batch stock level (restore stock)
		if err := s.inventoryRepo.UpdateBatchStock(itemReq.BatchID, itemReq.Quantity); err != nil {
			return nil, err
		}

		totalAmount += itemTotal
	}

	// Update return with total refund amount
	ret.TotalRefund = totalAmount
	if err := s.returnsRepo.UpdateReturn(ret); err != nil {
		return nil, err
	}

	// Get complete return with items and summary
	completeReturn, err := s.returnsRepo.GetReturnByID(ret.ID)
	if err != nil {
		return nil, err
	}

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
	ret, err := s.returnsRepo.GetReturnByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Status != nil {
		ret.Status = *req.Status
	}

	if err := s.returnsRepo.UpdateReturn(ret); err != nil {
		return nil, err
	}

	return s.mapReturnToResponse(ret), nil
}

// DeleteReturn deletes a return
func (s *ReturnsService) DeleteReturn(id string) error {
	return s.returnsRepo.DeleteReturn(id)
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
	if req.SaleID == "" {
		return errors.New("sale ID is required")
	}
	if len(req.Items) == 0 {
		return errors.New("at least one item is required")
	}

	for _, item := range req.Items {
		if item.BatchID == "" {
			return errors.New("batch ID is required for all items")
		}
		if item.Quantity <= 0 {
			return errors.New("quantity must be greater than 0")
		}
	}

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
