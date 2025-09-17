// internal/services/discounts_service.go
package services

import (
	"errors"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
)

type DiscountsService struct {
	discountRepo *repositories.DiscountsRepository
}

func NewDiscountsService(discountRepo *repositories.DiscountsRepository) *DiscountsService {
	return &DiscountsService{
		discountRepo: discountRepo,
	}
}

// CreateDiscount creates a new discount
func (s *DiscountsService) CreateDiscount(req *models.CreateDiscountRequest) (*models.DiscountResponse, error) {
	// Parse dates
	validFrom, err := time.Parse("2006-01-02T15:04:05Z", req.ValidFrom)
	if err != nil {
		return nil, errors.New("invalid valid_from date format")
	}

	validUntil, err := time.Parse("2006-01-02T15:04:05Z", req.ValidUntil)
	if err != nil {
		return nil, errors.New("invalid valid_until date format")
	}

	// Validate date range
	if validFrom.After(validUntil) {
		return nil, errors.New("valid_from cannot be after valid_until")
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
	if req.CustomerGroups != nil {
		discount.CustomerGroups = req.CustomerGroups
	}
	if req.UsageLimit != nil {
		discount.UsageLimit = req.UsageLimit
	}
	if req.UsagePerCustomer != nil {
		discount.UsagePerCustomer = req.UsagePerCustomer
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
	if req.CustomerGroups != nil {
		discount.CustomerGroups = req.CustomerGroups
	}
	if req.UsageLimit != nil {
		discount.UsageLimit = req.UsageLimit
	}
	if req.UsagePerCustomer != nil {
		discount.UsagePerCustomer = req.UsagePerCustomer
	}
	if req.ValidFrom != nil {
		validFrom, err := time.Parse("2006-01-02T15:04:05Z", *req.ValidFrom)
		if err != nil {
			return nil, errors.New("invalid valid_from date format")
		}
		discount.ValidFrom = validFrom
	}
	if req.ValidUntil != nil {
		validUntil, err := time.Parse("2006-01-02T15:04:05Z", *req.ValidUntil)
		if err != nil {
			return nil, errors.New("invalid valid_until date format")
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
			return nil, errors.New("valid_from cannot be after valid_until")
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
	// Validate discount
	discount, err := s.discountRepo.ValidateDiscount(req.DiscountCode, req.CustomerID, req.OrderValue, req.ProductIDs, req.WarehouseID)
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
func (s *DiscountsService) ApplyDiscount(discountID, customerID, saleID string, discountAmount float64) error {
	// Create usage record
	usage := models.NewDiscountUsage(discountID, customerID, saleID, discountAmount)
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
