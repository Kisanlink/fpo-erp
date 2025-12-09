package services

import (
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"

	"go.uber.org/zap"
)

// ProductPriceService handles product price business logic
type ProductPriceService struct {
	priceRepo   *repositories.ProductPriceRepository
	productRepo *repositories.ProductRepository
	variantRepo *repositories.ProductVariantRepository
	logger      interfaces.Logger
}

// NewProductPriceService creates a new product price service
func NewProductPriceService(priceRepo *repositories.ProductPriceRepository, productRepo *repositories.ProductRepository, variantRepo *repositories.ProductVariantRepository, logger interfaces.Logger) *ProductPriceService {
	return &ProductPriceService{
		priceRepo:   priceRepo,
		productRepo: productRepo,
		variantRepo: variantRepo,
		logger:      logger,
	}
}

// CreateProductPrice creates a new product price
func (s *ProductPriceService) CreateProductPrice(request *models.CreateProductPriceRequest) (*models.ProductPriceResponse, error) {
	s.logger.Info("Creating product price",
		zap.String("variant_id", request.VariantID),
		zap.String("price_type", request.PriceType),
		zap.Float64("price", request.Price))

	// Validate variant exists
	s.logger.Debug("Validating variant exists",
		zap.String("variant_id", request.VariantID))
	if _, err := s.variantRepo.GetByID(request.VariantID); err != nil {
		s.logger.Error("Variant not found",
			zap.Error(err),
			zap.String("variant_id", request.VariantID))
		return nil, errors.NewNotFoundError("ProductVariant")
	}

	// Parse effective dates
	s.logger.Debug("Parsing effective dates")
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
	currency := "INR"
	if request.Currency != "" {
		currency = request.Currency
	}

	s.logger.Debug("Setting price configuration",
		zap.String("currency", currency),
		zap.Time("effective_from", effectiveFrom))

	// Default to true if not specified
	isActive := true
	isActivePtr := &isActive
	if request.IsActive != nil {
		isActivePtr = request.IsActive
	}

	// Create price model using the proper constructor
	price := models.NewProductPrice(request.VariantID, request.PriceType, request.Price, currency, effectiveFrom, effectiveTo, isActivePtr)

	// Save to database
	s.logger.Debug("Saving price to database")
	if err := s.priceRepo.Create(price); err != nil {
		s.logger.Error("Failed to create product price",
			zap.Error(err),
			zap.String("variant_id", request.VariantID),
			zap.String("price_type", request.PriceType))
		return nil, err
	}

	// Convert to response
	isActiveValue := false
	if price.IsActive != nil {
		isActiveValue = *price.IsActive
	}
	response := &models.ProductPriceResponse{
		ID:            price.ID,
		VariantID:     price.VariantID,
		PriceType:     price.PriceType,
		Price:         price.Price,
		Currency:      price.Currency,
		EffectiveFrom: price.EffectiveFrom.Format(time.RFC3339),
		IsActive:      isActiveValue,
		CreatedAt:     price.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     price.UpdatedAt.Format(time.RFC3339),
	}

	if price.EffectiveTo != nil {
		effectiveToStr := price.EffectiveTo.Format(time.RFC3339)
		response.EffectiveTo = &effectiveToStr
	}

	s.logger.Info("Product price created successfully",
		zap.String("price_id", price.ID),
		zap.String("variant_id", price.VariantID),
		zap.String("price_type", price.PriceType),
		zap.Float64("price", price.Price))

	return response, nil
}

// GetProductPrice retrieves a product price by ID
func (s *ProductPriceService) GetProductPrice(id string) (*models.ProductPriceResponse, error) {
	s.logger.Info("Retrieving product price",
		zap.String("price_id", id))

	price, err := s.priceRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve product price",
			zap.Error(err),
			zap.String("price_id", id))
		return nil, err
	}

	isActiveValue := false
	if price.IsActive != nil {
		isActiveValue = *price.IsActive
	}
	response := &models.ProductPriceResponse{
		ID:            price.ID,
		VariantID:     price.VariantID,
		PriceType:     price.PriceType,
		Price:         price.Price,
		Currency:      price.Currency,
		EffectiveFrom: price.EffectiveFrom.Format(time.RFC3339),
		IsActive:      isActiveValue,
		CreatedAt:     price.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     price.UpdatedAt.Format(time.RFC3339),
	}

	if price.EffectiveTo != nil {
		effectiveToStr := price.EffectiveTo.Format(time.RFC3339)
		response.EffectiveTo = &effectiveToStr
	}

	s.logger.Debug("Product price retrieved successfully",
		zap.String("price_id", id),
		zap.String("price_type", price.PriceType),
		zap.Float64("price", price.Price))

	return response, nil
}

// GetVariantPrices retrieves all prices for a variant
func (s *ProductPriceService) GetVariantPrices(variantID string) ([]models.ProductPriceResponse, error) {
	s.logger.Info("Retrieving variant prices",
		zap.String("variant_id", variantID))

	prices, err := s.priceRepo.GetByVariantID(variantID)
	if err != nil {
		s.logger.Error("Failed to retrieve variant prices",
			zap.Error(err),
			zap.String("variant_id", variantID))
		return nil, err
	}

	var responses []models.ProductPriceResponse
	for _, price := range prices {
		isActiveValue := false
		if price.IsActive != nil {
			isActiveValue = *price.IsActive
		}

		response := models.ProductPriceResponse{
			ID:            price.ID,
			VariantID:     price.VariantID,
			PriceType:     price.PriceType,
			Price:         price.Price,
			Currency:      price.Currency,
			EffectiveFrom: price.EffectiveFrom.Format(time.RFC3339),
			IsActive:      isActiveValue,
			CreatedAt:     price.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     price.UpdatedAt.Format(time.RFC3339),
		}

		if price.EffectiveTo != nil {
			effectiveToStr := price.EffectiveTo.Format(time.RFC3339)
			response.EffectiveTo = &effectiveToStr
		}

		responses = append(responses, response)
	}

	s.logger.Info("Variant prices retrieved successfully",
		zap.String("variant_id", variantID),
		zap.Int("count", len(responses)))

	return responses, nil
}

// GetVariantPricesPaginated retrieves prices for a variant with pagination
func (s *ProductPriceService) GetVariantPricesPaginated(variantID string, limit, offset int) ([]models.ProductPriceResponse, int64, error) {
	s.logger.Info("Retrieving variant prices with pagination",
		zap.String("variant_id", variantID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	prices, total, err := s.priceRepo.GetByVariantIDPaginated(variantID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to retrieve paginated variant prices",
			zap.Error(err),
			zap.String("variant_id", variantID))
		return nil, 0, err
	}

	var responses []models.ProductPriceResponse
	for _, price := range prices {
		isActiveValue := false
		if price.IsActive != nil {
			isActiveValue = *price.IsActive
		}

		response := models.ProductPriceResponse{
			ID:            price.ID,
			VariantID:     price.VariantID,
			PriceType:     price.PriceType,
			Price:         price.Price,
			Currency:      price.Currency,
			EffectiveFrom: price.EffectiveFrom.Format(time.RFC3339),
			IsActive:      isActiveValue,
			CreatedAt:     price.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     price.UpdatedAt.Format(time.RFC3339),
		}

		if price.EffectiveTo != nil {
			effectiveToStr := price.EffectiveTo.Format(time.RFC3339)
			response.EffectiveTo = &effectiveToStr
		}

		responses = append(responses, response)
	}

	s.logger.Info("Paginated variant prices retrieved successfully",
		zap.String("variant_id", variantID),
		zap.Int("count", len(responses)),
		zap.Int64("total", total))

	return responses, total, nil
}

// GetCurrentPrice retrieves the current active price for a variant and type
func (s *ProductPriceService) GetCurrentPrice(variantID, priceType string) (*models.ProductPriceResponse, error) {
	s.logger.Info("Retrieving current price",
		zap.String("variant_id", variantID),
		zap.String("price_type", priceType))

	price, err := s.priceRepo.GetCurrentPrice(variantID, priceType)
	if err != nil {
		s.logger.Error("Failed to retrieve current price",
			zap.Error(err),
			zap.String("variant_id", variantID),
			zap.String("price_type", priceType))
		return nil, err
	}

	isActiveValue := false
	if price.IsActive != nil {
		isActiveValue = *price.IsActive
	}
	response := &models.ProductPriceResponse{
		ID:            price.ID,
		VariantID:     price.VariantID,
		PriceType:     price.PriceType,
		Price:         price.Price,
		Currency:      price.Currency,
		EffectiveFrom: price.EffectiveFrom.Format(time.RFC3339),
		IsActive:      isActiveValue,
		CreatedAt:     price.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     price.UpdatedAt.Format(time.RFC3339),
	}

	if price.EffectiveTo != nil {
		effectiveToStr := price.EffectiveTo.Format(time.RFC3339)
		response.EffectiveTo = &effectiveToStr
	}

	s.logger.Debug("Current price retrieved successfully",
		zap.String("price_id", price.ID),
		zap.Float64("price", price.Price))

	return response, nil
}

// UpdateProductPrice updates a product price
func (s *ProductPriceService) UpdateProductPrice(id string, request *models.UpdateProductPriceRequest) (*models.ProductPriceResponse, error) {
	s.logger.Info("Updating product price",
		zap.String("price_id", id))

	// Get existing price
	s.logger.Debug("Fetching existing price",
		zap.String("price_id", id))
	price, err := s.priceRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to fetch existing price",
			zap.Error(err),
			zap.String("price_id", id))
		return nil, err
	}

	// Update fields if provided
	s.logger.Debug("Updating price fields")
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
		price.IsActive = request.IsActive
	}

	// Save to database
	s.logger.Debug("Saving updated price to database")
	if err := s.priceRepo.Update(price); err != nil {
		s.logger.Error("Failed to update product price",
			zap.Error(err),
			zap.String("price_id", id))
		return nil, err
	}

	isActiveValue := false
	if price.IsActive != nil {
		isActiveValue = *price.IsActive
	}
	response := &models.ProductPriceResponse{
		ID:            price.ID,
		VariantID:     price.VariantID,
		PriceType:     price.PriceType,
		Price:         price.Price,
		Currency:      price.Currency,
		EffectiveFrom: price.EffectiveFrom.Format(time.RFC3339),
		IsActive:      isActiveValue,
		CreatedAt:     price.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     price.UpdatedAt.Format(time.RFC3339),
	}

	if price.EffectiveTo != nil {
		effectiveToStr := price.EffectiveTo.Format(time.RFC3339)
		response.EffectiveTo = &effectiveToStr
	}

	s.logger.Info("Product price updated successfully",
		zap.String("price_id", price.ID),
		zap.String("price_type", price.PriceType),
		zap.Float64("price", price.Price))

	return response, nil
}

// DeleteProductPrice deletes a product price
func (s *ProductPriceService) DeleteProductPrice(id string) error {
	s.logger.Info("Deleting product price",
		zap.String("price_id", id))

	// Check if price exists
	s.logger.Debug("Checking if price exists",
		zap.String("price_id", id))
	exists, err := s.priceRepo.Exists(id)
	if err != nil {
		s.logger.Error("Failed to check price existence",
			zap.Error(err),
			zap.String("price_id", id))
		return err
	}
	if !exists {
		s.logger.Error("Price not found",
			zap.String("price_id", id))
		return errors.NewNotFoundError("ProductPrice")
	}

	s.logger.Debug("Deleting price from database")
	if err := s.priceRepo.Delete(id); err != nil {
		s.logger.Error("Failed to delete product price",
			zap.Error(err),
			zap.String("price_id", id))
		return err
	}

	s.logger.Info("Product price deleted successfully",
		zap.String("price_id", id))

	return nil
}

// GetExpiredPrices retrieves prices that have expired
func (s *ProductPriceService) GetExpiredPrices() ([]models.ProductPriceResponse, error) {
	s.logger.Info("Retrieving expired prices")

	prices, err := s.priceRepo.GetExpiredPrices()
	if err != nil {
		s.logger.Error("Failed to retrieve expired prices",
			zap.Error(err))
		return nil, err
	}

	var responses []models.ProductPriceResponse
	for _, price := range prices {
		isActiveValue := false
		if price.IsActive != nil {
			isActiveValue = *price.IsActive
		}

		response := models.ProductPriceResponse{
			ID:            price.ID,
			VariantID:     price.VariantID,
			PriceType:     price.PriceType,
			Price:         price.Price,
			Currency:      price.Currency,
			EffectiveFrom: price.EffectiveFrom.Format(time.RFC3339),
			IsActive:      isActiveValue,
			CreatedAt:     price.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     price.UpdatedAt.Format(time.RFC3339),
		}

		if price.EffectiveTo != nil {
			effectiveToStr := price.EffectiveTo.Format(time.RFC3339)
			response.EffectiveTo = &effectiveToStr
		}

		responses = append(responses, response)
	}

	s.logger.Info("Expired prices retrieved successfully",
		zap.Int("count", len(responses)))

	return responses, nil
}

// GetExpiredPricesPaginated retrieves expired prices with pagination
func (s *ProductPriceService) GetExpiredPricesPaginated(limit, offset int) ([]models.ProductPriceResponse, int64, error) {
	s.logger.Info("Retrieving expired prices with pagination",
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	prices, total, err := s.priceRepo.GetExpiredPricesPaginated(limit, offset)
	if err != nil {
		s.logger.Error("Failed to retrieve paginated expired prices",
			zap.Error(err))
		return nil, 0, err
	}

	var responses []models.ProductPriceResponse
	for _, price := range prices {
		isActiveValue := false
		if price.IsActive != nil {
			isActiveValue = *price.IsActive
		}

		response := models.ProductPriceResponse{
			ID:            price.ID,
			VariantID:     price.VariantID,
			PriceType:     price.PriceType,
			Price:         price.Price,
			Currency:      price.Currency,
			EffectiveFrom: price.EffectiveFrom.Format(time.RFC3339),
			IsActive:      isActiveValue,
			CreatedAt:     price.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     price.UpdatedAt.Format(time.RFC3339),
		}

		if price.EffectiveTo != nil {
			effectiveToStr := price.EffectiveTo.Format(time.RFC3339)
			response.EffectiveTo = &effectiveToStr
		}

		responses = append(responses, response)
	}

	s.logger.Info("Paginated expired prices retrieved successfully",
		zap.Int("count", len(responses)),
		zap.Int64("total", total))

	return responses, total, nil
}
