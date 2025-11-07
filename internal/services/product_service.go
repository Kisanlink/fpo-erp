package services

import (
	"fmt"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
)

// ProductService handles product business logic
type ProductService struct {
	productRepo *repositories.ProductRepository
	priceRepo   *repositories.ProductPriceRepository
	variantRepo *repositories.ProductVariantRepository
}

// NewProductService creates a new product service
func NewProductService(productRepo *repositories.ProductRepository, priceRepo *repositories.ProductPriceRepository, variantRepo *repositories.ProductVariantRepository) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		priceRepo:   priceRepo,
		variantRepo: variantRepo,
	}
}

// CreateProduct creates a new product (generic product category)
func (s *ProductService) CreateProduct(request *models.CreateProductRequest) (*models.ProductResponse, error) {
	// Create product model using the proper constructor
	product := models.NewProduct(request.Name, request.Description)

	// Save to database
	if err := s.productRepo.Create(product); err != nil {
		// Log the actual error for debugging
		fmt.Printf("DEBUG: Product creation failed with error: %v\n", err)
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

	return response, nil
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(id string) (*models.ProductResponse, error) {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := &models.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		CreatedAt:   product.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return response, nil
}

// GetAllProducts retrieves all products
func (s *ProductService) GetAllProducts() ([]models.ProductResponse, error) {
	products, err := s.productRepo.GetAll()
	if err != nil {
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

	return responses, nil
}

// UpdateProduct updates a product
func (s *ProductService) UpdateProduct(id string, request *models.UpdateProductRequest) (*models.ProductResponse, error) {
	// Get existing product
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if request.Name != nil {
		product.Name = *request.Name
	}
	if request.Description != nil {
		product.Description = request.Description
	}

	// Save to database
	if err := s.productRepo.Update(product); err != nil {
		return nil, err
	}

	response := &models.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		CreatedAt:   product.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return response, nil
}

// DeleteProduct deletes a product
func (s *ProductService) DeleteProduct(id string) error {
	// Check if product exists
	exists, err := s.productRepo.Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.NewNotFoundError("Product")
	}

	return s.productRepo.Delete(id)
}

// SearchProducts searches products by name
func (s *ProductService) SearchProducts(query string) ([]models.ProductResponse, error) {
	products, err := s.productRepo.GetByName(query)
	if err != nil {
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

	return responses, nil
}

// GetProductWithPrices retrieves a product with all its prices
func (s *ProductService) GetProductWithPrices(id string) (*models.ProductWithPricesResponse, error) {
	// Get product
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Get product prices (if priceRepo is available)
	// NOTE: Prices are now variant-specific, not product-specific
	// Query all variants of the product and aggregate their prices
	var prices []models.ProductPrice
	if s.priceRepo != nil && s.variantRepo != nil {
		// Get all variants for this product
		variants, err := s.variantRepo.GetByProductID(id)
		if err != nil {
			return nil, err
		}

		// Get prices for each variant
		for _, variant := range variants {
			variantPrices, err := s.priceRepo.GetByVariantID(variant.ID)
			if err != nil {
				// Continue to next variant if price query fails
				continue
			}
			prices = append(prices, variantPrices...)
		}
	}

	// Convert prices to response format
	var priceResponses []models.ProductPriceResponse
	for _, price := range prices {
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

	return response, nil
}
