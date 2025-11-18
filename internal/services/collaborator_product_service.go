package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
)

// CollaboratorProductService handles collaborator-product association business logic
type CollaboratorProductService struct {
	collabProductRepo *repositories.CollaboratorProductRepository
	collaboratorRepo  *repositories.CollaboratorRepository
	productRepo       *repositories.ProductRepository
	variantRepo       *repositories.ProductVariantRepository // Added for unified architecture
}

// NewCollaboratorProductService creates a new collaborator product service
func NewCollaboratorProductService(
	collabProductRepo *repositories.CollaboratorProductRepository,
	collaboratorRepo *repositories.CollaboratorRepository,
	productRepo *repositories.ProductRepository,
	variantRepo *repositories.ProductVariantRepository,
) *CollaboratorProductService {
	return &CollaboratorProductService{
		collabProductRepo: collabProductRepo,
		collaboratorRepo:  collaboratorRepo,
		productRepo:       productRepo,
		variantRepo:       variantRepo,
	}
}

// AddProductToCollaborator adds a product to a collaborator with metadata
// This now creates a ProductVariant with collaborator_id set (unified architecture)
func (s *CollaboratorProductService) AddProductToCollaborator(ctx context.Context, collaboratorID string, request *models.CreateCollaboratorProductRequest) (*models.CollaboratorProductResponse, error) {
	// Validate collaborator exists
	collaborator, err := s.collaboratorRepo.GetByID(collaboratorID)
	if err != nil {
		return nil, err
	}
	if collaborator.IsActive != nil && !*collaborator.IsActive {
		return nil, errors.NewBadRequestError("collaborator is not active")
	}

	// Validate product exists
	product, err := s.productRepo.GetByID(request.ProductID)
	if err != nil {
		return nil, err
	}

	// Check if variant already exists for this collaborator+product combination
	// Query variants where product_id = X and collaborator_id = Y
	existingVariants, err := s.variantRepo.GetByProductID(request.ProductID)
	if err != nil {
		return nil, err
	}
	for _, variant := range existingVariants {
		if variant.CollaboratorID != nil && *variant.CollaboratorID == collaboratorID {
			return nil, errors.NewConflictError("product already associated with this collaborator")
		}
	}

	// Auto-generate variant details
	variantName := fmt.Sprintf("%s - %s", product.Name, request.BrandName)
	quantity := "1"        // Default quantity
	packSize := "Standard" // Default pack size

	// Serialize images to JSON if provided
	var imagesJSON *string
	if len(request.Images) > 0 {
		imagesBytes, err := json.Marshal(request.Images)
		if err != nil {
			return nil, errors.NewInternalServerError(fmt.Sprintf("failed to serialize images: %v", err))
		}
		imagesStr := string(imagesBytes)
		imagesJSON = &imagesStr
	}

	// Create product variant with collaborator fields (unified architecture)
	variant := models.NewCollaboratorVariant(
		request.ProductID,
		collaboratorID,
		variantName,
		quantity,
		packSize,
		request.BrandName,
		request.HSNCode,
		request.GSTRate,
	)
	variant.Images = imagesJSON
	variant.DosageInstructions = request.DosageInstructions
	variant.UsageDetails = request.UsageDetails

	// Save to database
	if err := s.variantRepo.Create(variant); err != nil {
		return nil, err
	}

	// Build response with related entities
	return s.buildCollaboratorProductResponseFromVariant(variant, collaborator, product)
}

// GetProductsByCollaborator retrieves all products for a collaborator
func (s *CollaboratorProductService) GetProductsByCollaborator(ctx context.Context, collaboratorID string) ([]models.CollaboratorProductResponse, error) {
	// Validate collaborator exists
	collaborator, err := s.collaboratorRepo.GetByID(collaboratorID)
	if err != nil {
		return nil, err
	}

	// Get all product variants for this collaborator (unified architecture)
	variants, err := s.variantRepo.GetByCollaboratorID(collaboratorID)
	if err != nil {
		return nil, err
	}

	var responses []models.CollaboratorProductResponse
	for _, variant := range variants {
		// Get product details
		product, err := s.productRepo.GetByID(variant.ProductID)
		if err != nil {
			continue // Skip on error
		}

		response, err := s.buildCollaboratorProductResponseFromVariant(&variant, collaborator, product)
		if err != nil {
			continue // Skip on error
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// GetCollaboratorsByProduct retrieves all collaborators for a product
func (s *CollaboratorProductService) GetCollaboratorsByProduct(ctx context.Context, productID string) ([]models.CollaboratorProductResponse, error) {
	// Validate product exists
	_, err := s.productRepo.GetByID(productID)
	if err != nil {
		return nil, err
	}

	// Get all collaborators for this product
	collabProducts, err := s.collabProductRepo.GetCollaboratorsByProduct(productID)
	if err != nil {
		return nil, err
	}

	var responses []models.CollaboratorProductResponse
	for _, cp := range collabProducts {
		// Get product details
		product, err := s.productRepo.GetByID(cp.ProductID)
		if err != nil {
			continue // Skip on error
		}

		response, err := s.buildCollaboratorProductResponse(&cp, &cp.Collaborator, product)
		if err != nil {
			continue // Skip on error
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// GetCollaboratorProduct retrieves a specific collaborator-product association
func (s *CollaboratorProductService) GetCollaboratorProduct(ctx context.Context, id string) (*models.CollaboratorProductResponse, error) {
	collabProduct, err := s.collabProductRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Get related entities
	collaborator, err := s.collaboratorRepo.GetByID(collabProduct.CollaboratorID)
	if err != nil {
		return nil, err
	}

	product, err := s.productRepo.GetByID(collabProduct.ProductID)
	if err != nil {
		return nil, err
	}

	return s.buildCollaboratorProductResponse(collabProduct, collaborator, product)
}

// UpdateCollaboratorProduct updates collaborator product metadata
func (s *CollaboratorProductService) UpdateCollaboratorProduct(ctx context.Context, id string, request *models.UpdateCollaboratorProductRequest) (*models.CollaboratorProductResponse, error) {
	// Get existing record
	collabProduct, err := s.collabProductRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if request.BrandName != nil {
		collabProduct.BrandName = *request.BrandName
	}
	if request.HSNCode != nil {
		collabProduct.HSNCode = *request.HSNCode
	}
	if request.GSTRate != nil {
		collabProduct.GSTRate = *request.GSTRate
	}
	if request.DosageInstructions != nil {
		collabProduct.DosageInstructions = request.DosageInstructions
	}
	if request.UsageDetails != nil {
		collabProduct.UsageDetails = request.UsageDetails
	}
	if request.IsActive != nil {
		collabProduct.IsActive = *request.IsActive
	}
	if request.Images != nil {
		// Serialize images to JSON
		imagesBytes, err := json.Marshal(*request.Images)
		if err != nil {
			return nil, errors.NewInternalServerError(fmt.Sprintf("failed to serialize images: %v", err))
		}
		imagesStr := string(imagesBytes)
		collabProduct.Images = &imagesStr
	}

	// Save to database
	if err := s.collabProductRepo.Update(collabProduct); err != nil {
		return nil, err
	}

	// Get related entities
	collaborator, err := s.collaboratorRepo.GetByID(collabProduct.CollaboratorID)
	if err != nil {
		return nil, err
	}

	product, err := s.productRepo.GetByID(collabProduct.ProductID)
	if err != nil {
		return nil, err
	}

	return s.buildCollaboratorProductResponse(collabProduct, collaborator, product)
}

// RemoveProductFromCollaborator removes a product from collaborator (soft delete)
func (s *CollaboratorProductService) RemoveProductFromCollaborator(ctx context.Context, collaboratorID, productID string) error {
	// Validate association exists
	exists, err := s.collabProductRepo.Exists(collaboratorID, productID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.NewNotFoundError("product association with this collaborator")
	}

	return s.collabProductRepo.DeleteByCollaboratorAndProduct(collaboratorID, productID)
}

// DeleteCollaboratorProduct deletes a collaborator product by ID (soft delete)
func (s *CollaboratorProductService) DeleteCollaboratorProduct(ctx context.Context, id string) error {
	// Validate exists
	_, err := s.collabProductRepo.GetByID(id)
	if err != nil {
		return err
	}

	return s.collabProductRepo.Delete(id)
}

// buildCollaboratorProductResponse builds a response with related entity details
func (s *CollaboratorProductService) buildCollaboratorProductResponse(
	collabProduct *models.CollaboratorProduct,
	collaborator *models.Collaborator,
	product *models.Product,
) (*models.CollaboratorProductResponse, error) {
	response := &models.CollaboratorProductResponse{
		ID:                 collabProduct.ID,
		CollaboratorID:     collabProduct.CollaboratorID,
		CollaboratorName:   collaborator.CompanyName,
		ProductID:          collabProduct.ProductID,
		ProductName:        product.Name,
		ProductSKU:         "", // Products no longer have SKU - SKU is at variant level
		BrandName:          collabProduct.BrandName,
		HSNCode:            collabProduct.HSNCode,
		GSTRate:            collabProduct.GSTRate,
		DosageInstructions: collabProduct.DosageInstructions,
		UsageDetails:       collabProduct.UsageDetails,
		IsActive:           collabProduct.IsActive,
		CreatedAt:          collabProduct.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:          collabProduct.UpdatedAt.UTC().Format(time.RFC3339),
	}

	// Parse images from JSON
	if collabProduct.Images != nil {
		var images []string
		if err := json.Unmarshal([]byte(*collabProduct.Images), &images); err == nil {
			response.Images = images
		}
	}

	// Add product summary
	response.Product = &models.ProductSummary{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
	}

	return response, nil
}

// buildCollaboratorProductResponseFromVariant builds a CollaboratorProductResponse from a ProductVariant
// This is used for the unified architecture where collaborator products are stored as variants
func (s *CollaboratorProductService) buildCollaboratorProductResponseFromVariant(
	variant *models.ProductVariant,
	collaborator *models.Collaborator,
	product *models.Product,
) (*models.CollaboratorProductResponse, error) {
	response := &models.CollaboratorProductResponse{
		ID:               variant.ID,
		CollaboratorID:   *variant.CollaboratorID,
		CollaboratorName: collaborator.CompanyName,
		ProductID:        variant.ProductID,
		ProductName:      product.Name,
		ProductSKU: func() string {
			if variant.SKU != nil {
				return *variant.SKU
			}
			return ""
		}(),
		IsActive:  variant.IsActive,
		CreatedAt: variant.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: variant.UpdatedAt.UTC().Format(time.RFC3339),
	}

	// Map collaborator-specific fields
	if variant.BrandName != nil {
		response.BrandName = *variant.BrandName
	}
	if variant.HSNCode != nil {
		response.HSNCode = *variant.HSNCode
	}
	if variant.GSTRate != nil {
		response.GSTRate = *variant.GSTRate
	}
	response.DosageInstructions = variant.DosageInstructions
	response.UsageDetails = variant.UsageDetails

	// Parse images from JSON
	if variant.Images != nil {
		var images []string
		if err := json.Unmarshal([]byte(*variant.Images), &images); err == nil {
			response.Images = images
		}
	}

	// Add product summary
	response.Product = &models.ProductSummary{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
	}

	return response, nil
}
