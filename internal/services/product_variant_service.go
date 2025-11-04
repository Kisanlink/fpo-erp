package services

import (
	"context"
	"fmt"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
)

// ProductVariantService handles product variant business logic
type ProductVariantService struct {
	variantRepo *repositories.ProductVariantRepository
	productRepo *repositories.ProductRepository
}

// NewProductVariantService creates a new product variant service
func NewProductVariantService(
	variantRepo *repositories.ProductVariantRepository,
	productRepo *repositories.ProductRepository,
) *ProductVariantService {
	return &ProductVariantService{
		variantRepo: variantRepo,
		productRepo: productRepo,
	}
}

// CreateProductVariant creates a new product variant
func (s *ProductVariantService) CreateProductVariant(ctx context.Context, productID string, request *models.CreateProductVariantRequest) (*models.ProductVariantResponse, error) {
	// Validate product exists
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		return nil, err
	}

	// Validate SKU uniqueness if provided
	if request.SKU != nil && *request.SKU != "" {
		exists, err := s.variantRepo.SKUExists(*request.SKU)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("variant with SKU %s already exists", *request.SKU)
		}
	}

	// Validate barcode uniqueness if provided
	if request.Barcode != nil && *request.Barcode != "" {
		exists, err := s.variantRepo.BarcodeExists(*request.Barcode)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("variant with barcode %s already exists", *request.Barcode)
		}
	}

	// Create variant
	variant := models.NewProductVariant(productID, request.VariantName, request.Quantity, request.PackSize)
	variant.Description = request.Description
	variant.SKU = request.SKU
	variant.Barcode = request.Barcode

	// Save to database
	if err := s.variantRepo.Create(variant); err != nil {
		return nil, err
	}

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
		return nil, err
	}

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
	// Get existing variant
	variant, err := s.variantRepo.GetByID(id)
	if err != nil {
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
				return nil, fmt.Errorf("variant with SKU %s already exists", *request.SKU)
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
				return nil, fmt.Errorf("variant with barcode %s already exists", *request.Barcode)
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

	// Save to database
	if err := s.variantRepo.Update(variant); err != nil {
		return nil, err
	}

	// Get product details
	product, err := s.productRepo.GetByID(variant.ProductID)
	if err != nil {
		return nil, err
	}

	return s.buildProductVariantResponse(variant, product)
}

// DeleteProductVariant deletes a product variant (soft delete)
func (s *ProductVariantService) DeleteProductVariant(ctx context.Context, id string) error {
	// Validate variant exists
	_, err := s.variantRepo.GetByID(id)
	if err != nil {
		return err
	}

	return s.variantRepo.Delete(id)
}

// buildProductVariantResponse builds a response with product details
func (s *ProductVariantService) buildProductVariantResponse(variant *models.ProductVariant, product *models.Product) (*models.ProductVariantResponse, error) {
	return &models.ProductVariantResponse{
		ID:          variant.ID,
		ProductID:   variant.ProductID,
		VariantName: variant.VariantName,
		Description: variant.Description,
		Quantity:    variant.Quantity,
		PackSize:    variant.PackSize,
		SKU:         variant.SKU,
		Barcode:     variant.Barcode,
		IsActive:    variant.IsActive,
		CreatedAt:   variant.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   variant.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}
