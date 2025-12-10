package services

import (
	"kisanlink-erp/internal/interfaces"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/utils"
)

// ProductVariantService handles product variant business logic
type ProductVariantService struct {
	variantRepo *repositories.ProductVariantRepository
	productRepo *repositories.ProductRepository
	priceRepo   *repositories.ProductPriceRepository
	s3Service   *S3Service
	logger      interfaces.Logger
}

// NewProductVariantService creates a new product variant service
func NewProductVariantService(
	variantRepo *repositories.ProductVariantRepository,
	productRepo *repositories.ProductRepository,
	priceRepo *repositories.ProductPriceRepository,
	s3Service *S3Service,
	logger interfaces.Logger,
) *ProductVariantService {
	return &ProductVariantService{
		variantRepo: variantRepo,
		productRepo: productRepo,
		priceRepo:   priceRepo,
		s3Service:   s3Service,
		logger:      logger,
	}
}

// CreateProductVariant creates a new product variant
func (s *ProductVariantService) CreateProductVariant(ctx context.Context, productID string, request *models.CreateProductVariantRequest) (*models.ProductVariantResponse, error) {
	s.logger.Info("Creating product variant",
		zap.String("product_id", productID),
		zap.String("variant_name", request.VariantName))

	// Validate product exists
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		s.logger.Error("Product not found",
			zap.Error(err),
			zap.String("product_id", productID))
		return nil, err
	}
	s.logger.Debug("Product found")

	// Validate SKU uniqueness if provided
	if request.SKU != nil && *request.SKU != "" {
		s.logger.Debug("Validating SKU uniqueness",
			zap.String("sku", *request.SKU))
		exists, err := s.variantRepo.SKUExists(*request.SKU)
		if err != nil {
			s.logger.Error("Failed to check SKU existence",
				zap.Error(err))
			return nil, err
		}
		if exists {
			s.logger.Error("SKU already exists",
				zap.String("sku", *request.SKU))
			return nil, errors.NewConflictError("variant with SKU " + *request.SKU + " already exists")
		}
	}

	// Validate barcode uniqueness if provided
	if request.Barcode != nil && *request.Barcode != "" {
		exists, err := s.variantRepo.BarcodeExists(*request.Barcode)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.NewConflictError("variant with barcode " + *request.Barcode + " already exists")
		}
	}

	// Validate prices if provided
	if err := s.validatePrices(request.Prices); err != nil {
		s.logger.Error("Price validation failed",
			zap.Error(err))
		return nil, err
	}

	// Create variant with required HSNCode and GSTRate (GST-only tax system)
	variant := models.NewProductVariant(productID, request.VariantName, request.Quantity, request.PackSize, request.HSNCode, request.GSTRate)
	variant.Description = request.Description
	variant.SKU = request.SKU
	variant.Barcode = request.Barcode
	// Note: Prices are stored in product_prices table, not embedded in variant

	// Assign collaborator-specific fields
	variant.CollaboratorIDs = request.CollaboratorIDs
	variant.BrandName = request.BrandName
	variant.DosageInstructions = request.DosageInstructions
	variant.UsageDetails = request.UsageDetails

	// Marshal attachment IDs to JSON
	if len(request.Images) > 0 {
		imagesBytes, err := json.Marshal(request.Images)
		if err != nil {
			return nil, errors.NewValidationError("Invalid images format")
		}
		imagesStr := string(imagesBytes)
		variant.Images = &imagesStr
	}

	// Save variant and prices atomically in a transaction
	s.logger.Debug("Saving variant and prices to database")
	err = s.variantRepo.WithTransaction(func(tx *gorm.DB) error {
		// Create variant
		if err := s.variantRepo.CreateWithTx(tx, variant); err != nil {
			s.logger.Error("Failed to create variant in transaction",
				zap.Error(err))
			return err
		}

		// Create ProductPrice records for each price in request
		if len(request.Prices) > 0 && s.priceRepo != nil {
			for _, price := range request.Prices {
				// Validate price is positive
				if price.Price <= 0 {
					err := fmt.Errorf("price must be greater than 0 for type %s", price.PriceType)
					s.logger.Error("Invalid price value",
						zap.Error(err),
						zap.String("price_type", price.PriceType),
						zap.Float64("price", price.Price))
					return errors.NewValidationError(err.Error())
				}

				// Set defaults
				currency := price.Currency
				if currency == "" {
					currency = "INR"
				}

				isActive := true
				if price.IsActive != nil {
					isActive = *price.IsActive
				}

				// Parse effective dates
				effectiveFrom := time.Now()
				if price.EffectiveFrom != nil {
					if parsed, err := time.Parse("2006-01-02", *price.EffectiveFrom); err == nil {
						effectiveFrom = parsed
					}
				}

				var effectiveTo *time.Time
				if price.EffectiveTo != nil {
					if parsed, err := time.Parse("2006-01-02", *price.EffectiveTo); err == nil {
						effectiveTo = &parsed
					}
				}

				productPrice := models.NewProductPrice(
					variant.ID,
					price.PriceType,
					price.Price,
					currency,
					effectiveFrom,
					effectiveTo,
					&isActive,
				)
				if err := s.priceRepo.CreateWithTx(tx, productPrice); err != nil {
					s.logger.Error("Failed to create price record in transaction",
						zap.Error(err),
						zap.String("variant_id", variant.ID),
						zap.String("price_type", price.PriceType))
					return fmt.Errorf("failed to create price for type %s: %w", price.PriceType, err)
				}
				s.logger.Debug("Price record created",
					zap.String("price_id", productPrice.ID),
					zap.String("price_type", price.PriceType))
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Transaction failed",
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("Product variant created successfully",
		zap.String("variant_id", variant.ID))

	return s.buildProductVariantResponse(ctx, variant, product)
}

// GetProductVariant retrieves a product variant by ID
func (s *ProductVariantService) GetProductVariant(ctx context.Context, id string) (*models.ProductVariantResponse, error) {
	variant, err := s.variantRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Get product details
	product, err := s.productRepo.GetByID(variant.ProductID)
	if err != nil {
		return nil, err
	}

	return s.buildProductVariantResponse(ctx, variant, product)
}

// GetVariantsByProduct retrieves all variants for a product
func (s *ProductVariantService) GetVariantsByProduct(ctx context.Context, productID string) ([]models.ProductVariantResponse, error) {
	// Validate product exists
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		s.logger.Error("Product not found",
			zap.Error(err),
			zap.String("product_id", productID))
		return nil, err
	}
	s.logger.Debug("Product found")

	// Get all variants for this product
	variants, err := s.variantRepo.GetByProductID(productID)
	if err != nil {
		return nil, err
	}

	var responses []models.ProductVariantResponse
	for _, variant := range variants {
		response, err := s.buildProductVariantResponse(ctx, &variant, product)
		if err != nil {
			continue // Skip on error
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// GetVariantsByProductPaginated retrieves variants for a product with pagination
func (s *ProductVariantService) GetVariantsByProductPaginated(ctx context.Context, productID string, limit, offset int) ([]models.ProductVariantResponse, int64, error) {
	s.logger.Info("Retrieving variants by product with pagination",
		zap.String("product_id", productID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	// Validate product exists
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		s.logger.Error("Product not found",
			zap.Error(err),
			zap.String("product_id", productID))
		return nil, 0, err
	}

	// Get paginated variants for this product
	variants, total, err := s.variantRepo.GetByProductIDPaginated(productID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to retrieve paginated variants",
			zap.Error(err),
			zap.String("product_id", productID))
		return nil, 0, err
	}

	var responses []models.ProductVariantResponse
	for _, variant := range variants {
		response, err := s.buildProductVariantResponse(ctx, &variant, product)
		if err != nil {
			continue // Skip on error
		}
		responses = append(responses, *response)
	}

	s.logger.Info("Paginated variants retrieved successfully",
		zap.String("product_id", productID),
		zap.Int("count", len(responses)),
		zap.Int64("total", total))

	return responses, total, nil
}

// GetVariantBySKU retrieves a product variant by SKU
func (s *ProductVariantService) GetVariantBySKU(ctx context.Context, sku string) (*models.ProductVariantResponse, error) {
	variant, err := s.variantRepo.GetBySKU(sku)
	if err != nil {
		return nil, err
	}

	// Get product details
	product, err := s.productRepo.GetByID(variant.ProductID)
	if err != nil {
		return nil, err
	}

	return s.buildProductVariantResponse(ctx, variant, product)
}

// GetVariantByBarcode retrieves a product variant by barcode
func (s *ProductVariantService) GetVariantByBarcode(ctx context.Context, barcode string) (*models.ProductVariantResponse, error) {
	variant, err := s.variantRepo.GetByBarcode(barcode)
	if err != nil {
		return nil, err
	}

	// Get product details
	product, err := s.productRepo.GetByID(variant.ProductID)
	if err != nil {
		return nil, err
	}

	return s.buildProductVariantResponse(ctx, variant, product)
}

// UpdateProductVariant updates a product variant
func (s *ProductVariantService) UpdateProductVariant(ctx context.Context, id string, request *models.UpdateProductVariantRequest) (*models.ProductVariantResponse, error) {
	s.logger.Info("Updating product variant",
		zap.String("variant_id", id))

	// Get existing variant
	variant, err := s.variantRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Variant not found",
			zap.Error(err),
			zap.String("variant_id", id))
		return nil, err
	}

	// Validate SKU uniqueness if being updated
	if request.SKU != nil && *request.SKU != "" {
		// Only check if SKU is different from current
		if variant.SKU == nil || *variant.SKU != *request.SKU {
			exists, err := s.variantRepo.SKUExists(*request.SKU)
			if err != nil {
				return nil, err
			}
			if exists {
				s.logger.Error("SKU already exists",
					zap.String("sku", *request.SKU))
				return nil, errors.NewConflictError("variant with SKU " + *request.SKU + " already exists")
			}
		}
	}

	// Validate barcode uniqueness if being updated
	if request.Barcode != nil && *request.Barcode != "" {
		// Only check if barcode is different from current
		if variant.Barcode == nil || *variant.Barcode != *request.Barcode {
			exists, err := s.variantRepo.BarcodeExists(*request.Barcode)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, errors.NewConflictError("variant with barcode " + *request.Barcode + " already exists")
			}
		}
	}

	// Update fields if provided
	if request.VariantName != nil {
		variant.VariantName = *request.VariantName
	}
	if request.Description != nil {
		variant.Description = request.Description
	}
	if request.Quantity != nil {
		variant.Quantity = *request.Quantity
	}
	if request.PackSize != nil {
		variant.PackSize = *request.PackSize
	}
	if request.SKU != nil {
		variant.SKU = request.SKU
	}
	if request.Barcode != nil {
		variant.Barcode = request.Barcode
	}
	if request.IsActive != nil {
		variant.IsActive = *request.IsActive
	}
	if request.Images != nil {
		// Marshal attachment IDs to JSON
		imagesBytes, err := json.Marshal(*request.Images)
		if err != nil {
			return nil, errors.NewValidationError("Invalid images format")
		}
		imagesStr := string(imagesBytes)
		variant.Images = &imagesStr
	}
	if request.BrandName != nil {
		variant.BrandName = request.BrandName
	}
	if request.HSNCode != nil {
		variant.HSNCode = *request.HSNCode
	}
	if request.GSTRate != nil {
		variant.GSTRate = *request.GSTRate
	}
	if request.DosageInstructions != nil {
		variant.DosageInstructions = request.DosageInstructions
	}
	if request.UsageDetails != nil {
		variant.UsageDetails = request.UsageDetails
	}
	if request.CollaboratorIDs != nil {
		variant.CollaboratorIDs = *request.CollaboratorIDs
	}

	// Wrap variant update and price updates in a transaction
	s.logger.Debug("Updating variant and prices in transaction")
	err = s.variantRepo.WithTransaction(func(tx *gorm.DB) error {
		// Update variant using tx.Save() directly
		if err := tx.Save(variant).Error; err != nil {
			s.logger.Error("Failed to update variant in transaction",
				zap.Error(err))
			return errors.NewInternalServerError("Failed to update product variant")
		}

		// Update prices if provided
		if request.Prices != nil && s.priceRepo != nil {
			// Validate prices before updating
			if err := s.validatePrices(*request.Prices); err != nil {
				s.logger.Error("Price validation failed during update",
					zap.Error(err))
				return err
			}

			// Update prices in product_prices table
			for _, price := range *request.Prices {
				// Set defaults
				currency := price.Currency
				if currency == "" {
					currency = "INR"
				}

				isActive := true
				if price.IsActive != nil {
					isActive = *price.IsActive
				}

				// Parse effective dates
				effectiveFrom := time.Now()
				if price.EffectiveFrom != nil {
					if parsed, err := time.Parse("2006-01-02", *price.EffectiveFrom); err == nil {
						effectiveFrom = parsed
					}
				}

				var effectiveTo *time.Time
				if price.EffectiveTo != nil {
					if parsed, err := time.Parse("2006-01-02", *price.EffectiveTo); err == nil {
						effectiveTo = &parsed
					}
				}

				// Try to get existing price for this type
				existingPrice, err := s.priceRepo.GetCurrentPrice(variant.ID, price.PriceType)
				if err == nil && existingPrice != nil {
					// Update existing price using tx.Save() directly
					existingPrice.Price = price.Price
					existingPrice.Currency = currency
					existingPrice.EffectiveFrom = effectiveFrom
					existingPrice.EffectiveTo = effectiveTo
					existingPrice.IsActive = &isActive
					if err := tx.Save(existingPrice).Error; err != nil {
						s.logger.Error("Failed to update price record in transaction",
							zap.Error(err),
							zap.String("variant_id", variant.ID),
							zap.String("price_type", price.PriceType))
						return fmt.Errorf("failed to update price for type %s: %w", price.PriceType, err)
					}
				} else {
					// Create new price record using CreateWithTx
					productPrice := models.NewProductPrice(
						variant.ID,
						price.PriceType,
						price.Price,
						currency,
						effectiveFrom,
						effectiveTo,
						&isActive,
					)
					if err := s.priceRepo.CreateWithTx(tx, productPrice); err != nil {
						s.logger.Error("Failed to create price record in transaction",
							zap.Error(err),
							zap.String("variant_id", variant.ID),
							zap.String("price_type", price.PriceType))
						return fmt.Errorf("failed to create price for type %s: %w", price.PriceType, err)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Transaction failed",
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("Product variant updated successfully",
		zap.String("variant_id", id))

	// Get product details
	product, err := s.productRepo.GetByID(variant.ProductID)
	if err != nil {
		return nil, err
	}

	return s.buildProductVariantResponse(ctx, variant, product)
}

// DeleteProductVariant deletes a product variant (soft delete)
func (s *ProductVariantService) DeleteProductVariant(ctx context.Context, id string) error {
	s.logger.Info("Deleting product variant",
		zap.String("variant_id", id))

	// Validate variant exists
	_, err := s.variantRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Variant not found",
			zap.Error(err),
			zap.String("variant_id", id))
		return err
	}

	if err = s.variantRepo.Delete(id); err != nil {
		s.logger.Error("Failed to delete variant",
			zap.Error(err),
			zap.String("variant_id", id))
		return err
	}
	s.logger.Info("Product variant deleted successfully",
		zap.String("variant_id", id))
	return nil
}

// GetPriceByType returns the price for a specific price type from a variant
func (s *ProductVariantService) GetPriceByType(ctx context.Context, variantID string, priceType string) (*models.VariantPrice, error) {
	// Validate variant exists
	_, err := s.variantRepo.GetByID(variantID)
	if err != nil {
		return nil, err
	}

	// Search for the price type in product_prices table
	if s.priceRepo == nil {
		return nil, errors.NewInternalServerError("price repository not configured")
	}

	productPrice, err := s.priceRepo.GetCurrentPrice(variantID, priceType)
	if err != nil {
		return nil, errors.NewNotFoundError("price type '" + priceType + "' not found for variant")
	}

	// Convert ProductPrice to VariantPrice for API compatibility
	return &models.VariantPrice{
		PriceType: productPrice.PriceType,
		Price:     productPrice.Price,
		Currency:  productPrice.Currency,
	}, nil
}

// validatePrices validates the prices array
func (s *ProductVariantService) validatePrices(prices []models.VariantPrice) error {
	if len(prices) == 0 {
		return nil // Prices are optional
	}

	validPriceTypes := map[string]bool{
		models.PriceTypeMRP:    true,
		models.PriceTypeMSP:    true,
		models.PriceTypeMember: true,
		models.PriceTypeRetail: true,
	}

	priceTypeSeen := make(map[string]bool)

	for i, price := range prices {
		// Validate price_type
		if !validPriceTypes[price.PriceType] {
			return errors.NewValidationError("Invalid price_type at index " + strconv.Itoa(i) + ": must be 'MRP', 'MSP', 'member', or 'retail'")
		}

		// Check for duplicate price types
		if priceTypeSeen[price.PriceType] {
			return errors.NewValidationError("Duplicate price_type '" + price.PriceType + "' found")
		}
		priceTypeSeen[price.PriceType] = true

		// Validate price is positive
		if price.Price <= 0 {
			return errors.NewValidationError("Price must be greater than 0 for " + price.PriceType)
		}

		// Validate effective_from date format if provided
		if price.EffectiveFrom != nil {
			if _, err := time.Parse("2006-01-02", *price.EffectiveFrom); err != nil {
				return errors.NewValidationError("Invalid effective_from date format at index " + strconv.Itoa(i) + ": must be YYYY-MM-DD")
			}
		}

		// Validate effective_to date format if provided
		if price.EffectiveTo != nil {
			if _, err := time.Parse("2006-01-02", *price.EffectiveTo); err != nil {
				return errors.NewValidationError("Invalid effective_to date format at index " + strconv.Itoa(i) + ": must be YYYY-MM-DD")
			}
		}

		// Validate effective_to > effective_from if both provided
		if price.EffectiveFrom != nil && price.EffectiveTo != nil {
			effectiveFrom, _ := time.Parse("2006-01-02", *price.EffectiveFrom)
			effectiveTo, _ := time.Parse("2006-01-02", *price.EffectiveTo)
			if !effectiveTo.After(effectiveFrom) {
				return errors.NewValidationError("effective_to must be after effective_from at index " + strconv.Itoa(i))
			}
		}
	}

	return nil
}

// buildProductVariantResponse builds a response with product details and presigned image URLs
func (s *ProductVariantService) buildProductVariantResponse(ctx context.Context, variant *models.ProductVariant, product *models.Product) (*models.ProductVariantResponse, error) {
	// Unmarshal image paths from JSON
	var images []string
	if variant.Images != nil && *variant.Images != "" {
		if err := json.Unmarshal([]byte(*variant.Images), &images); err == nil {
			// Successfully unmarshaled
		}
	}

	// Generate presigned URLs for each image
	var imageURLs []string
	if s.s3Service != nil && len(images) > 0 {
		for _, imagePath := range images {
			if url, err := s.s3Service.GeneratePresignedURLForKey(ctx, imagePath, time.Hour); err == nil {
				imageURLs = append(imageURLs, url)
			}
		}
	}

	// Fetch prices from product_prices table
	var priceResponses []models.ProductPriceResponse
	if s.priceRepo != nil {
		prices, err := s.priceRepo.GetActiveByVariantID(variant.ID)
		if err == nil {
			for _, price := range prices {
				effectiveTo := ""
				if price.EffectiveTo != nil {
					effectiveTo = price.EffectiveTo.Format(time.RFC3339)
				}

				isActive := false
				if price.IsActive != nil {
					isActive = *price.IsActive
				}

				priceResponses = append(priceResponses, models.ProductPriceResponse{
					ID:            price.ID,
					VariantID:     price.VariantID,
					PriceType:     price.PriceType,
					Price:         utils.RoundPrice(price.Price),
					Currency:      price.Currency,
					EffectiveFrom: price.EffectiveFrom.Format(time.RFC3339),
					EffectiveTo:   &effectiveTo,
					IsActive:      isActive,
					CreatedAt:     price.CreatedAt.Format(time.RFC3339),
					UpdatedAt:     price.UpdatedAt.Format(time.RFC3339),
				})
			}
		}
	}

	return &models.ProductVariantResponse{
		ID:          variant.ID,
		ProductID:   variant.ProductID,
		VariantName: variant.VariantName,
		Description: variant.Description,
		Quantity:    variant.Quantity,
		PackSize:    variant.PackSize,
		SKU:         variant.SKU,
		Barcode:     variant.Barcode,
		Images:      images,
		ImageURLs:   imageURLs,
		Prices:      priceResponses, // Fetch from product_prices table
		IsActive:    variant.IsActive,
		CreatedAt:   variant.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   variant.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}
