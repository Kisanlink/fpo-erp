package services

import (
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"time"
)

// ProductPriceService handles product price business logic
type ProductPriceService struct {
	priceRepo   *repositories.ProductPriceRepository
	productRepo *repositories.ProductRepository
}

// NewProductPriceService creates a new product price service
func NewProductPriceService(priceRepo *repositories.ProductPriceRepository, productRepo *repositories.ProductRepository) *ProductPriceService {
	return &ProductPriceService{
		priceRepo:   priceRepo,
		productRepo: productRepo,
	}
}

// CreateProductPrice creates a new product price
func (s *ProductPriceService) CreateProductPrice(request *models.CreateProductPriceRequest) (*models.ProductPriceResponse, error) {
	// Check if product exists
	_, err := s.productRepo.GetByID(request.ProductID)
	if err != nil {
		return nil, err
	}

	// Parse effective dates
	effectiveFrom := time.Now()
	if request.EffectiveFrom != nil {
		if parsed, err := time.Parse(time.RFC3339, *request.EffectiveFrom); err == nil {
			effectiveFrom = parsed
		}
	}

	var effectiveTo *time.Time
	if request.EffectiveTo != nil {
		if parsed, err := time.Parse(time.RFC3339, *request.EffectiveTo); err == nil {
			effectiveTo = &parsed
		}
	}

	// Set default values
	currency := "USD"
	if request.Currency != "" {
		currency = request.Currency
	}

	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	// Create price model using the proper constructor
	price := models.NewProductPrice(request.ProductID, request.PriceType, request.Price, currency, effectiveFrom, effectiveTo, isActive)

	// Save to database
	if err := s.priceRepo.Create(price); err != nil {
		return nil, err
	}

	// Convert to response
	response := &models.ProductPriceResponse{
		ID:            price.ID,
		ProductID:     price.ProductID,
		PriceType:     price.PriceType,
		Price:         price.Price,
		Currency:      price.Currency,
		EffectiveFrom: price.EffectiveFrom.Format(time.RFC3339),
		IsActive:      price.IsActive,
		CreatedAt:     price.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     price.UpdatedAt.Format(time.RFC3339),
	}

	if price.EffectiveTo != nil {
		effectiveToStr := price.EffectiveTo.Format(time.RFC3339)
		response.EffectiveTo = &effectiveToStr
	}

	return response, nil
}

// GetProductPrice retrieves a product price by ID
func (s *ProductPriceService) GetProductPrice(id string) (*models.ProductPriceResponse, error) {
	price, err := s.priceRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := &models.ProductPriceResponse{
		ID:            price.ID,
		ProductID:     price.ProductID,
		PriceType:     price.PriceType,
		Price:         price.Price,
		Currency:      price.Currency,
		EffectiveFrom: price.EffectiveFrom.Format(time.RFC3339),
		IsActive:      price.IsActive,
		CreatedAt:     price.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     price.UpdatedAt.Format(time.RFC3339),
	}

	if price.EffectiveTo != nil {
		effectiveToStr := price.EffectiveTo.Format(time.RFC3339)
		response.EffectiveTo = &effectiveToStr
	}

	return response, nil
}

// GetProductPrices retrieves all prices for a product
func (s *ProductPriceService) GetProductPrices(productID string) ([]models.ProductPriceResponse, error) {
	prices, err := s.priceRepo.GetByProductID(productID)
	if err != nil {
		return nil, err
	}

	var responses []models.ProductPriceResponse
	for _, price := range prices {
		response := models.ProductPriceResponse{
			ID:            price.ID,
			ProductID:     price.ProductID,
			PriceType:     price.PriceType,
			Price:         price.Price,
			Currency:      price.Currency,
			EffectiveFrom: price.EffectiveFrom.Format(time.RFC3339),
			IsActive:      price.IsActive,
			CreatedAt:     price.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     price.UpdatedAt.Format(time.RFC3339),
		}

		if price.EffectiveTo != nil {
			effectiveToStr := price.EffectiveTo.Format(time.RFC3339)
			response.EffectiveTo = &effectiveToStr
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// GetCurrentPrice retrieves the current active price for a product and type
func (s *ProductPriceService) GetCurrentPrice(productID, priceType string) (*models.ProductPriceResponse, error) {
	price, err := s.priceRepo.GetCurrentPrice(productID, priceType)
	if err != nil {
		return nil, err
	}

	response := &models.ProductPriceResponse{
		ID:            price.ID,
		ProductID:     price.ProductID,
		PriceType:     price.PriceType,
		Price:         price.Price,
		Currency:      price.Currency,
		EffectiveFrom: price.EffectiveFrom.Format(time.RFC3339),
		IsActive:      price.IsActive,
		CreatedAt:     price.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     price.UpdatedAt.Format(time.RFC3339),
	}

	if price.EffectiveTo != nil {
		effectiveToStr := price.EffectiveTo.Format(time.RFC3339)
		response.EffectiveTo = &effectiveToStr
	}

	return response, nil
}

// UpdateProductPrice updates a product price
func (s *ProductPriceService) UpdateProductPrice(id string, request *models.UpdateProductPriceRequest) (*models.ProductPriceResponse, error) {
	// Get existing price
	price, err := s.priceRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if request.PriceType != nil {
		price.PriceType = *request.PriceType
	}
	if request.Price != nil {
		price.Price = *request.Price
	}
	if request.Currency != nil {
		price.Currency = *request.Currency
	}
	if request.EffectiveFrom != nil {
		if parsed, err := time.Parse(time.RFC3339, *request.EffectiveFrom); err == nil {
			price.EffectiveFrom = parsed
		}
	}
	if request.EffectiveTo != nil {
		if parsed, err := time.Parse(time.RFC3339, *request.EffectiveTo); err == nil {
			price.EffectiveTo = &parsed
		}
	}
	if request.IsActive != nil {
		price.IsActive = *request.IsActive
	}

	// Save to database
	if err := s.priceRepo.Update(price); err != nil {
		return nil, err
	}

	response := &models.ProductPriceResponse{
		ID:            price.ID,
		ProductID:     price.ProductID,
		PriceType:     price.PriceType,
		Price:         price.Price,
		Currency:      price.Currency,
		EffectiveFrom: price.EffectiveFrom.Format(time.RFC3339),
		IsActive:      price.IsActive,
		CreatedAt:     price.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     price.UpdatedAt.Format(time.RFC3339),
	}

	if price.EffectiveTo != nil {
		effectiveToStr := price.EffectiveTo.Format(time.RFC3339)
		response.EffectiveTo = &effectiveToStr
	}

	return response, nil
}

// DeleteProductPrice deletes a product price
func (s *ProductPriceService) DeleteProductPrice(id string) error {
	// Check if price exists
	exists, err := s.priceRepo.Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.NewNotFoundError("ProductPrice")
	}

	return s.priceRepo.Delete(id)
}

// GetExpiredPrices retrieves prices that have expired
func (s *ProductPriceService) GetExpiredPrices() ([]models.ProductPriceResponse, error) {
	prices, err := s.priceRepo.GetExpiredPrices()
	if err != nil {
		return nil, err
	}

	var responses []models.ProductPriceResponse
	for _, price := range prices {
		response := models.ProductPriceResponse{
			ID:            price.ID,
			ProductID:     price.ProductID,
			PriceType:     price.PriceType,
			Price:         price.Price,
			Currency:      price.Currency,
			EffectiveFrom: price.EffectiveFrom.Format(time.RFC3339),
			IsActive:      price.IsActive,
			CreatedAt:     price.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     price.UpdatedAt.Format(time.RFC3339),
		}

		if price.EffectiveTo != nil {
			effectiveToStr := price.EffectiveTo.Format(time.RFC3339)
			response.EffectiveTo = &effectiveToStr
		}

		responses = append(responses, response)
	}

	return responses, nil
}
