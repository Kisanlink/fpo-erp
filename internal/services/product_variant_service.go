package services

import (
	"kisanlink-erp/internal/interfaces"

	"go.uber.org/zap"

	"context"
	"encoding/json"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
)

// ProductVariantService handles product variant business logic
type ProductVariantService struct {
	variantRepo *repositories.ProductVariantRepository
	productRepo *repositories.ProductRepository
	logger      interfaces.Logger
}

// NewProductVariantService creates a new product variant service
func NewProductVariantService(
	variantRepo *repositories.ProductVariantRepository,
	productRepo *repositories.ProductRepository,
	logger interfaces.Logger,
) *ProductVariantService {
	return &ProductVariantService{
		variantRepo: variantRepo,
		productRepo: productRepo,
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

	// Create variant
	variant := models.NewProductVariant(productID, request.VariantName, request.Quantity, request.PackSize)
	variant.Description = request.Description
	variant.SKU = request.SKU
	variant.Barcode = request.Barcode

	// Marshal attachment IDs to JSON
	if len(request.Images) > 0 {
		imagesBytes, err := json.Marshal(request.Images)
		if err != nil {
			return nil, errors.NewValidationError("Invalid images format")
		}
		imagesStr := string(imagesBytes)
		variant.Images = &imagesStr
	}

	// Save to database
	s.logger.Debug("Saving variant to database")
	if err := s.variantRepo.Create(variant); err != nil {
		s.logger.Error("Failed to create variant",
			zap.Error(err))
		return nil, err
	}
	s.logger.Info("Product variant created successfully",
		zap.String("variant_id", variant.ID))

	return s.buildProductVariantResponse(variant, product)
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

	return s.buildProductVariantResponse(variant, product)
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
		response, err := s.buildProductVariantResponse(&variant, product)
		if err != nil {
			continue // Skip on error
		}
		responses = append(responses, *response)
	}

	return responses, nil
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

	return s.buildProductVariantResponse(variant, product)
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

	return s.buildProductVariantResponse(variant, product)
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

	// Save to database
	s.logger.Debug("Saving updated variant")
	if err := s.variantRepo.Update(variant); err != nil {
		s.logger.Error("Failed to update variant",
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

	return s.buildProductVariantResponse(variant, product)
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

// buildProductVariantResponse builds a response with product details
func (s *ProductVariantService) buildProductVariantResponse(variant *models.ProductVariant, product *models.Product) (*models.ProductVariantResponse, error) {
	// Unmarshal attachment IDs from JSON
	var images []string
	if variant.Images != nil && *variant.Images != "" {
		if err := json.Unmarshal([]byte(*variant.Images), &images); err == nil {
			// Successfully unmarshaled
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
		IsActive:    variant.IsActive,
		CreatedAt:   variant.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   variant.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}
