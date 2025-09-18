// internal/services/discounts_service.go
package services

import (
	"encoding/json"
	"fmt"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	apperrors "kisanlink-erp/internal/errors"
)

type DiscountsService struct {
	discountRepo  *repositories.DiscountsRepository
	productRepo   *repositories.ProductRepository
	warehouseRepo *repositories.WarehouseRepository
}

func NewDiscountsService(discountRepo *repositories.DiscountsRepository, productRepo *repositories.ProductRepository, warehouseRepo *repositories.WarehouseRepository) *DiscountsService {
	return &DiscountsService{
		discountRepo:  discountRepo,
		productRepo:   productRepo,
		warehouseRepo: warehouseRepo,
	}
}

// CreateDiscount creates a new discount
func (s *DiscountsService) CreateDiscount(req *models.CreateDiscountRequest) (*models.DiscountResponse, error) {
	// Parse dates
	validFrom, err := time.Parse("2006-01-02T15:04:05Z", req.ValidFrom)
	if err != nil {
		return nil, apperrors.NewBadRequestError("invalid valid_from date format")
	}

	validUntil, err := time.Parse("2006-01-02T15:04:05Z", req.ValidUntil)
	if err != nil {
		return nil, apperrors.NewBadRequestError("invalid valid_until date format")
	}

	// Validate date range
	if validFrom.After(validUntil) {
		return nil, apperrors.NewBadRequestError("valid_from cannot be after valid_until")
	}

	// Validate business logic
	if err := s.validateDiscountRequest(req); err != nil {
		return nil, err
	}

	// Create discount
	discount := models.NewDiscount(req.Code, req.Name, req.Description, req.DiscountType, req.Value, validFrom, validUntil)

	// Set optional fields
	if req.MaxDiscountAmount != nil {
		discount.MaxDiscountAmount = req.MaxDiscountAmount
	}
	if req.MinOrderValue != nil {
		discount.MinOrderValue = req.MinOrderValue
	}
	if req.MaxOrderValue != nil {
		discount.MaxOrderValue = req.MaxOrderValue
	}
	if req.ApplicableProducts != nil {
		discount.ApplicableProducts = req.ApplicableProducts
	}
	if req.ExcludedProducts != nil {
		discount.ExcludedProducts = req.ExcludedProducts
	}
	if req.ApplicableCategories != nil {
		discount.ApplicableCategories = req.ApplicableCategories
	}
	if req.ExcludedCategories != nil {
		discount.ExcludedCategories = req.ExcludedCategories
	}
	if req.ApplicableWarehouses != nil {
		discount.ApplicableWarehouses = req.ApplicableWarehouses
	}
	if req.UsageLimit != nil {
		discount.UsageLimit = req.UsageLimit
	}
	if req.IsActive != nil {
		discount.IsActive = *req.IsActive
	}
	if req.IsStackable != nil {
		discount.IsStackable = *req.IsStackable
	}
	if req.Priority != nil {
		discount.Priority = *req.Priority
	}
	if req.Terms != nil {
		discount.Terms = req.Terms
	}

	// Set Buy X Get Y specific fields
	if req.BuyQuantity != nil {
		discount.BuyQuantity = req.BuyQuantity
	}
	if req.GetQuantity != nil {
		discount.GetQuantity = req.GetQuantity
	}
	if req.GetDiscountType != nil {
		discount.GetDiscountType = req.GetDiscountType
	}
	if req.GetDiscountValue != nil {
		discount.GetDiscountValue = req.GetDiscountValue
	}

	// Save to database
	if err := s.discountRepo.CreateDiscount(discount); err != nil {
		return nil, err
	}

	return discount.ToResponse(), nil
}

// GetDiscount retrieves a discount by ID
func (s *DiscountsService) GetDiscount(id string) (*models.DiscountResponse, error) {
	discount, err := s.discountRepo.GetDiscountByID(id)
	if err != nil {
		return nil, err
	}
	return discount.ToResponse(), nil
}

// GetAllDiscounts retrieves all discounts with pagination
func (s *DiscountsService) GetAllDiscounts(limit, offset int) ([]models.DiscountResponse, error) {
	discounts, err := s.discountRepo.GetAllDiscounts(limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []models.DiscountResponse
	for _, discount := range discounts {
		responses = append(responses, *discount.ToResponse())
	}

	return responses, nil
}

// GetActiveDiscounts retrieves all active discounts
func (s *DiscountsService) GetActiveDiscounts() ([]models.DiscountResponse, error) {
	discounts, err := s.discountRepo.GetActiveDiscounts()
	if err != nil {
		return nil, err
	}

	var responses []models.DiscountResponse
	for _, discount := range discounts {
		responses = append(responses, *discount.ToResponse())
	}

	return responses, nil
}

// UpdateDiscount updates a discount
func (s *DiscountsService) UpdateDiscount(id string, req *models.UpdateDiscountRequest) (*models.DiscountResponse, error) {
	discount, err := s.discountRepo.GetDiscountByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != nil {
		discount.Name = *req.Name
	}
	if req.Description != nil {
		discount.Description = req.Description
	}
	if req.Value != nil {
		discount.Value = *req.Value
	}
	if req.MaxDiscountAmount != nil {
		discount.MaxDiscountAmount = req.MaxDiscountAmount
	}
	if req.MinOrderValue != nil {
		discount.MinOrderValue = req.MinOrderValue
	}
	if req.MaxOrderValue != nil {
		discount.MaxOrderValue = req.MaxOrderValue
	}
	if req.ApplicableProducts != nil {
		discount.ApplicableProducts = req.ApplicableProducts
	}
	if req.ExcludedProducts != nil {
		discount.ExcludedProducts = req.ExcludedProducts
	}
	if req.ApplicableCategories != nil {
		discount.ApplicableCategories = req.ApplicableCategories
	}
	if req.ExcludedCategories != nil {
		discount.ExcludedCategories = req.ExcludedCategories
	}
	if req.ApplicableWarehouses != nil {
		discount.ApplicableWarehouses = req.ApplicableWarehouses
	}
	if req.UsageLimit != nil {
		discount.UsageLimit = req.UsageLimit
	}
	if req.ValidFrom != nil {
		validFrom, err := time.Parse("2006-01-02T15:04:05Z", *req.ValidFrom)
		if err != nil {
			return nil, apperrors.NewBadRequestError("invalid valid_from date format")
		}
		discount.ValidFrom = validFrom
	}
	if req.ValidUntil != nil {
		validUntil, err := time.Parse("2006-01-02T15:04:05Z", *req.ValidUntil)
		if err != nil {
			return nil, apperrors.NewBadRequestError("invalid valid_until date format")
		}
		discount.ValidUntil = validUntil
	}
	if req.IsActive != nil {
		discount.IsActive = *req.IsActive
	}
	if req.IsStackable != nil {
		discount.IsStackable = *req.IsStackable
	}
	if req.Priority != nil {
		discount.Priority = *req.Priority
	}
	if req.Terms != nil {
		discount.Terms = req.Terms
	}

	// Validate date range if both dates are being updated
	if req.ValidFrom != nil && req.ValidUntil != nil {
		if discount.ValidFrom.After(discount.ValidUntil) {
			return nil, apperrors.NewBadRequestError("valid_from cannot be after valid_until")
		}
	}

	// Save to database
	if err := s.discountRepo.UpdateDiscount(discount); err != nil {
		return nil, err
	}

	return discount.ToResponse(), nil
}

// DeleteDiscount deletes a discount
func (s *DiscountsService) DeleteDiscount(id string) error {
	return s.discountRepo.DeleteDiscount(id)
}

// GetDiscountsByType retrieves discounts by type
func (s *DiscountsService) GetDiscountsByType(discountType models.DiscountType) ([]models.DiscountResponse, error) {
	discounts, err := s.discountRepo.GetDiscountsByType(discountType)
	if err != nil {
		return nil, err
	}

	var responses []models.DiscountResponse
	for _, discount := range discounts {
		responses = append(responses, *discount.ToResponse())
	}

	return responses, nil
}

// GetDiscountsByStatus retrieves discounts by status
func (s *DiscountsService) GetDiscountsByStatus(status string) ([]models.DiscountResponse, error) {
	discounts, err := s.discountRepo.GetDiscountsByStatus(status)
	if err != nil {
		return nil, err
	}

	var responses []models.DiscountResponse
	for _, discount := range discounts {
		responses = append(responses, *discount.ToResponse())
	}

	return responses, nil
}

// ValidateDiscount validates a discount for a given order
func (s *DiscountsService) ValidateDiscount(req *models.ValidateDiscountRequest) (*models.DiscountValidationResponse, error) {
	// Validate discount with categories
	discount, err := s.discountRepo.ValidateDiscountWithCategories(req.DiscountCode, req.OrderValue, req.ProductIDs, req.CategoryIDs, req.WarehouseID)
	if err != nil {
		return &models.DiscountValidationResponse{
			IsValid: false,
			Message: err.Error(),
		}, nil
	}

	// Calculate discount amount
	calculatedDiscount := s.discountRepo.CalculateDiscount(discount, req.OrderValue)

	return &models.DiscountValidationResponse{
		IsValid:            true,
		DiscountID:         discount.ID,
		DiscountCode:       discount.Code,
		DiscountName:       discount.Name,
		DiscountType:       string(discount.DiscountType),
		Value:              discount.Value,
		MaxDiscountAmount:  discount.MaxDiscountAmount,
		CalculatedDiscount: calculatedDiscount,
		Message:            "Discount is valid",
	}, nil
}

// ApplyDiscount applies a discount to a sale and records usage
func (s *DiscountsService) ApplyDiscount(discountID, saleID string, discountAmount float64) error {
	// Create usage record
	usage := models.NewDiscountUsage(discountID, saleID, discountAmount)
	if err := s.discountRepo.CreateDiscountUsage(usage); err != nil {
		return err
	}

	// Increment usage count
	if err := s.discountRepo.IncrementUsage(discountID); err != nil {
		return err
	}

	return nil
}

// GetDiscountUsageBySale retrieves discount usage by sale
func (s *DiscountsService) GetDiscountUsageBySale(saleID string) ([]models.DiscountUsageResponse, error) {
	usages, err := s.discountRepo.GetDiscountUsageBySale(saleID)
	if err != nil {
		return nil, err
	}

	var responses []models.DiscountUsageResponse
	for _, usage := range usages {
		responses = append(responses, *usage.ToResponse())
	}

	return responses, nil
}

// validateDiscountRequest validates the discount creation request
func (s *DiscountsService) validateDiscountRequest(req *models.CreateDiscountRequest) error {
	// Validate discount type specific rules
	switch req.DiscountType {
	case models.DiscountTypePercentage:
		if req.Value <= 0 || req.Value > 100 {
			return apperrors.NewBadRequestError("percentage discount value must be between 0 and 100")
		}
	case models.DiscountTypeFlat:
		if req.Value <= 0 {
			return apperrors.NewBadRequestError("flat discount value must be greater than 0")
		}
	case models.DiscountTypeBuyXGetY:
		if req.ApplicableProducts == nil || *req.ApplicableProducts == "" {
			return apperrors.NewBadRequestError("buy_x_get_y discount requires applicable_products")
		}
		if req.BuyQuantity == nil || *req.BuyQuantity <= 0 {
			return apperrors.NewBadRequestError("buy_x_get_y discount requires valid buy_quantity")
		}
		if req.GetQuantity == nil || *req.GetQuantity <= 0 {
			return apperrors.NewBadRequestError("buy_x_get_y discount requires valid get_quantity")
		}
		if req.GetDiscountType != nil {
			allowedTypes := []string{"free", "percentage", "flat"}
			valid := false
			for _, t := range allowedTypes {
				if *req.GetDiscountType == t {
					valid = true
					break
				}
			}
			if !valid {
				return apperrors.NewBadRequestError("get_discount_type must be one of: free, percentage, flat")
			}
		}
	}

	// Validate percentage discounts have max_discount_amount for safety
	if req.DiscountType == models.DiscountTypePercentage && req.Value > 50 && req.MaxDiscountAmount == nil {
		return apperrors.NewBadRequestError("percentage discounts over 50% must have max_discount_amount set")
	}

	// Validate min/max order values
	if req.MinOrderValue != nil && req.MaxOrderValue != nil {
		if *req.MinOrderValue >= *req.MaxOrderValue {
			return apperrors.NewBadRequestError("min_order_value must be less than max_order_value")
		}
	}

	// Validate usage limits
	if req.UsageLimit != nil && *req.UsageLimit <= 0 {
		return apperrors.NewBadRequestError("usage_limit must be greater than 0")
	}

	// Validate JSON fields
	if err := s.validateJSONFields(req); err != nil {
		return err
	}

	// Validate referenced entities exist
	if err := s.validateReferencedEntities(req); err != nil {
		return err
	}

	return nil
}

// validateJSONFields validates that JSON string fields contain valid JSON arrays
func (s *DiscountsService) validateJSONFields(req *models.CreateDiscountRequest) error {
	if req.ApplicableProducts != nil && *req.ApplicableProducts != "" {
		var products []string
		if err := json.Unmarshal([]byte(*req.ApplicableProducts), &products); err != nil {
			return fmt.Errorf("applicable_products must be a valid JSON array: %v", err)
		}
		if len(products) == 0 {
			return apperrors.NewBadRequestError("applicable_products cannot be an empty array")
		}
	}

	if req.ExcludedProducts != nil && *req.ExcludedProducts != "" {
		var products []string
		if err := json.Unmarshal([]byte(*req.ExcludedProducts), &products); err != nil {
			return fmt.Errorf("excluded_products must be a valid JSON array: %v", err)
		}
	}

	if req.ApplicableCategories != nil && *req.ApplicableCategories != "" {
		var categories []string
		if err := json.Unmarshal([]byte(*req.ApplicableCategories), &categories); err != nil {
			return fmt.Errorf("applicable_categories must be a valid JSON array: %v", err)
		}
	}

	if req.ExcludedCategories != nil && *req.ExcludedCategories != "" {
		var categories []string
		if err := json.Unmarshal([]byte(*req.ExcludedCategories), &categories); err != nil {
			return fmt.Errorf("excluded_categories must be a valid JSON array: %v", err)
		}
	}

	if req.ApplicableWarehouses != nil && *req.ApplicableWarehouses != "" {
		var warehouses []string
		if err := json.Unmarshal([]byte(*req.ApplicableWarehouses), &warehouses); err != nil {
			return fmt.Errorf("applicable_warehouses must be a valid JSON array: %v", err)
		}
	}

	return nil
}

// validateReferencedEntities validates that referenced products, warehouses exist
func (s *DiscountsService) validateReferencedEntities(req *models.CreateDiscountRequest) error {
	// Validate applicable products exist
	if req.ApplicableProducts != nil && *req.ApplicableProducts != "" {
		var productIDs []string
		if err := json.Unmarshal([]byte(*req.ApplicableProducts), &productIDs); err != nil {
			return fmt.Errorf("invalid applicable_products JSON: %v", err)
		}
		for _, productID := range productIDs {
			if _, err := s.productRepo.GetByID(productID); err != nil {
				return fmt.Errorf("product %s not found in applicable_products", productID)
			}
		}
	}

	// Validate excluded products exist
	if req.ExcludedProducts != nil && *req.ExcludedProducts != "" {
		var productIDs []string
		if err := json.Unmarshal([]byte(*req.ExcludedProducts), &productIDs); err != nil {
			return fmt.Errorf("invalid excluded_products JSON: %v", err)
		}
		for _, productID := range productIDs {
			if _, err := s.productRepo.GetByID(productID); err != nil {
				return fmt.Errorf("product %s not found in excluded_products", productID)
			}
		}
	}

	// Validate applicable warehouses exist
	if req.ApplicableWarehouses != nil && *req.ApplicableWarehouses != "" {
		var warehouseIDs []string
		if err := json.Unmarshal([]byte(*req.ApplicableWarehouses), &warehouseIDs); err != nil {
			return fmt.Errorf("invalid applicable_warehouses JSON: %v", err)
		}
		for _, warehouseID := range warehouseIDs {
			if _, err := s.warehouseRepo.GetByID(warehouseID); err != nil {
				return fmt.Errorf("warehouse %s not found in applicable_warehouses", warehouseID)
			}
		}
	}

	// Validate conflicts
	if req.ApplicableProducts != nil && req.ApplicableCategories != nil {
		if *req.ApplicableProducts != "" && *req.ApplicableCategories != "" {
			return apperrors.NewBadRequestError("cannot specify both applicable_products and applicable_categories")
		}
	}

	return nil
}

// GetDiscountUsageStats retrieves usage statistics for a discount
func (s *DiscountsService) GetDiscountUsageStats(discountID string) (map[string]interface{}, error) {
	return s.discountRepo.GetDiscountUsageStats(discountID)
}

// GetApplicableDiscountsForOrder retrieves all applicable discounts for a given order
func (s *DiscountsService) GetApplicableDiscountsForOrder(orderValue float64, productIDs []string, categoryIDs []string, warehouseID string) ([]models.DiscountResponse, error) {
	discounts, err := s.discountRepo.GetApplicableDiscountsForOrder(orderValue, productIDs, categoryIDs, warehouseID)
	if err != nil {
		return nil, err
	}

	var responses []models.DiscountResponse
	for _, discount := range discounts {
		responses = append(responses, *discount.ToResponse())
	}

	return responses, nil
}

// CalculateOptimalDiscounts calculates the best combination of discounts for an order
func (s *DiscountsService) CalculateOptimalDiscounts(orderValue float64, productIDs []string, categoryIDs []string, warehouseID string) ([]models.DiscountResponse, float64, error) {
	discounts, totalDiscount, err := s.discountRepo.CalculateOptimalDiscounts(orderValue, productIDs, categoryIDs, warehouseID)
	if err != nil {
		return nil, 0, err
	}

	var responses []models.DiscountResponse
	for _, discount := range discounts {
		responses = append(responses, *discount.ToResponse())
	}

	return responses, totalDiscount, nil
}

// ValidateDiscountWithCategories validates a discount including category checks
func (s *DiscountsService) ValidateDiscountWithCategories(req *models.ValidateDiscountRequest, categoryIDs []string) (*models.DiscountValidationResponse, error) {
	// Validate discount with categories
	discount, err := s.discountRepo.ValidateDiscountWithCategories(req.DiscountCode, req.OrderValue, req.ProductIDs, categoryIDs, req.WarehouseID)
	if err != nil {
		return &models.DiscountValidationResponse{
			IsValid: false,
			Message: err.Error(),
		}, nil
	}

	// Calculate discount amount
	calculatedDiscount := s.discountRepo.CalculateDiscount(discount, req.OrderValue)

	return &models.DiscountValidationResponse{
		IsValid:            true,
		DiscountID:         discount.ID,
		DiscountCode:       discount.Code,
		DiscountName:       discount.Name,
		DiscountType:       string(discount.DiscountType),
		Value:              discount.Value,
		MaxDiscountAmount:  discount.MaxDiscountAmount,
		CalculatedDiscount: calculatedDiscount,
		Message:            "Discount is valid",
	}, nil
}
