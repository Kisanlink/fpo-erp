package services

import (
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"

	"go.uber.org/zap"
)

// ProductService handles product business logic
type ProductService struct {
	productRepo *repositories.ProductRepository
	variantRepo *repositories.ProductVariantRepository
	logger      interfaces.Logger
}

// NewProductService creates a new product service
func NewProductService(productRepo *repositories.ProductRepository, variantRepo *repositories.ProductVariantRepository, logger interfaces.Logger) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		variantRepo: variantRepo,
		logger:      logger,
	}
}

// CreateProduct creates a new product (generic product category)
func (s *ProductService) CreateProduct(request *models.CreateProductRequest) (*models.ProductResponse, error) {
	s.logger.Info("Creating product",
		zap.String("name", request.Name))

	// Create product model using the proper constructor
	product := models.NewProduct(request.Name, request.Description)

	s.logger.Debug("Saving product to database")

	// Save to database
	if err := s.productRepo.Create(product); err != nil {
		s.logger.Error("Failed to create product",
			zap.Error(err),
			zap.String("name", request.Name))
		return nil, err
	}

	// Convert to response
	response := &models.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		CreatedAt:   product.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("Product created successfully",
		zap.String("product_id", product.ID),
		zap.String("name", product.Name))

	return response, nil
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(id string) (*models.ProductResponse, error) {
	s.logger.Info("Retrieving product",
		zap.String("product_id", id))

	product, err := s.productRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve product",
			zap.Error(err),
			zap.String("product_id", id))
		return nil, err
	}

	response := &models.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		CreatedAt:   product.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Debug("Product retrieved successfully",
		zap.String("product_id", id),
		zap.String("name", product.Name))

	return response, nil
}

// GetAllProducts retrieves all products
func (s *ProductService) GetAllProducts() ([]models.ProductResponse, error) {
	s.logger.Info("Retrieving all products")

	products, err := s.productRepo.GetAll()
	if err != nil {
		s.logger.Error("Failed to retrieve all products",
			zap.Error(err))
		return nil, err
	}

	var responses []models.ProductResponse
	for _, product := range products {
		response := models.ProductResponse{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			CreatedAt:   product.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   product.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	s.logger.Info("Retrieved all products successfully",
		zap.Int("count", len(responses)))

	return responses, nil
}

// UpdateProduct updates a product
func (s *ProductService) UpdateProduct(id string, request *models.UpdateProductRequest) (*models.ProductResponse, error) {
	s.logger.Info("Updating product",
		zap.String("product_id", id))

	// Get existing product
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve product for update",
			zap.Error(err),
			zap.String("product_id", id))
		return nil, err
	}

	s.logger.Debug("Applying product updates",
		zap.String("product_id", id))

	// Update fields if provided
	if request.Name != nil {
		product.Name = *request.Name
	}
	if request.Description != nil {
		product.Description = request.Description
	}

	// Save to database
	if err := s.productRepo.Update(product); err != nil {
		s.logger.Error("Failed to update product",
			zap.Error(err),
			zap.String("product_id", id))
		return nil, err
	}

	response := &models.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		CreatedAt:   product.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("Product updated successfully",
		zap.String("product_id", id),
		zap.String("name", product.Name))

	return response, nil
}

// DeleteProduct deletes a product
func (s *ProductService) DeleteProduct(id string) error {
	s.logger.Info("Deleting product",
		zap.String("product_id", id))

	// Check if product exists
	exists, err := s.productRepo.Exists(id)
	if err != nil {
		s.logger.Error("Failed to check product existence",
			zap.Error(err),
			zap.String("product_id", id))
		return err
	}
	if !exists {
		s.logger.Warn("Product not found for deletion",
			zap.String("product_id", id))
		return errors.NewNotFoundError("Product")
	}

	if err := s.productRepo.Delete(id); err != nil {
		s.logger.Error("Failed to delete product",
			zap.Error(err),
			zap.String("product_id", id))
		return err
	}

	s.logger.Info("Product deleted successfully",
		zap.String("product_id", id))

	return nil
}

// SearchProducts searches products by name
func (s *ProductService) SearchProducts(query string) ([]models.ProductResponse, error) {
	s.logger.Info("Searching products",
		zap.String("query", query))

	products, err := s.productRepo.GetByName(query)
	if err != nil {
		s.logger.Error("Failed to search products",
			zap.Error(err),
			zap.String("query", query))
		return nil, err
	}

	var responses []models.ProductResponse
	for _, product := range products {
		response := models.ProductResponse{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			CreatedAt:   product.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   product.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	s.logger.Info("Product search completed",
		zap.String("query", query),
		zap.Int("results", len(responses)))

	return responses, nil
}

// GetProductWithPrices retrieves a product with all its prices
func (s *ProductService) GetProductWithPrices(id string) (*models.ProductWithPricesResponse, error) {
	s.logger.Info("Retrieving product with prices",
		zap.String("product_id", id))

	// Get product
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve product",
			zap.Error(err),
			zap.String("product_id", id))
		return nil, err
	}

	// NOTE: Prices are now embedded in product variants
	// Users should query individual variants to get their prices
	// This simplifies the product response and avoids complex aggregation
	s.logger.Debug("Prices are available at variant level",
		zap.String("product_id", id))

	// Price responses are now empty - prices moved to variants
	var priceResponses []models.ProductPriceResponse
	for _, price := range []models.ProductPrice{} {
		effectiveTo := ""
		if price.EffectiveTo != nil {
			effectiveTo = price.EffectiveTo.Format("2006-01-02T15:04:05Z")
		}

		isActiveValue := false
		if price.IsActive != nil {
			isActiveValue = *price.IsActive
		}

		priceResponse := models.ProductPriceResponse{
			ID:            price.ID,
			VariantID:     price.VariantID,
			PriceType:     price.PriceType,
			Price:         price.Price,
			Currency:      price.Currency,
			EffectiveFrom: price.EffectiveFrom.Format("2006-01-02T15:04:05Z"),
			EffectiveTo:   &effectiveTo,
			IsActive:      isActiveValue,
			CreatedAt:     price.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     price.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		priceResponses = append(priceResponses, priceResponse)
	}

	// Create response
	response := &models.ProductWithPricesResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Prices:      priceResponses,
		CreatedAt:   product.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("Product with prices retrieved successfully",
		zap.String("product_id", id),
		zap.String("name", product.Name),
		zap.Int("price_count", len(priceResponses)))

	return response, nil
}
