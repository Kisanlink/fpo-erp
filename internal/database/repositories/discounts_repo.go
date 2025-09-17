package repositories

import (
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"

	"gorm.io/gorm"
)

type DiscountsRepository struct {
	db *gorm.DB
}

func NewDiscountsRepository(db *gorm.DB) *DiscountsRepository {
	return &DiscountsRepository{db: db}
}

// CreateDiscount creates a new discount
func (r *DiscountsRepository) CreateDiscount(discount *models.Discount) error {
	if err := r.db.Create(discount).Error; err != nil {
		return errors.NewInternalServerError("Failed to create discount")
	}
	return nil
}

// GetDiscountByID retrieves a discount by ID
func (r *DiscountsRepository) GetDiscountByID(id string) (*models.Discount, error) {
	var discount models.Discount
	if err := r.db.Where("id = ?", id).First(&discount).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Discount not found")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve discount")
	}
	return &discount, nil
}

// GetDiscountByCode retrieves a discount by code
func (r *DiscountsRepository) GetDiscountByCode(code string) (*models.Discount, error) {
	var discount models.Discount
	if err := r.db.Where("code = ?", code).First(&discount).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Discount not found")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve discount")
	}
	return &discount, nil
}

// GetAllDiscounts retrieves all discounts with pagination
func (r *DiscountsRepository) GetAllDiscounts(limit, offset int) ([]models.Discount, error) {
	var discounts []models.Discount
	if err := r.db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&discounts).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve discounts")
	}
	return discounts, nil
}

// GetActiveDiscounts retrieves all active discounts
func (r *DiscountsRepository) GetActiveDiscounts() ([]models.Discount, error) {
	var discounts []models.Discount
	now := time.Now()
	if err := r.db.Where("is_active = ? AND valid_from <= ? AND valid_until >= ?", true, now, now).Order("priority DESC, created_at DESC").Find(&discounts).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve active discounts")
	}
	return discounts, nil
}

// UpdateDiscount updates a discount
func (r *DiscountsRepository) UpdateDiscount(discount *models.Discount) error {
	if err := r.db.Save(discount).Error; err != nil {
		return errors.NewInternalServerError("Failed to update discount")
	}
	return nil
}

// DeleteDiscount deletes a discount
func (r *DiscountsRepository) DeleteDiscount(id string) error {
	if err := r.db.Where("id = ?", id).Delete(&models.Discount{}).Error; err != nil {
		return errors.NewInternalServerError("Failed to delete discount")
	}
	return nil
}

// GetDiscountsByType retrieves discounts by type
func (r *DiscountsRepository) GetDiscountsByType(discountType models.DiscountType) ([]models.Discount, error) {
	var discounts []models.Discount
	if err := r.db.Where("discount_type = ?", discountType).Order("priority DESC, created_at DESC").Find(&discounts).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve discounts by type")
	}
	return discounts, nil
}

// GetDiscountsByStatus retrieves discounts by status
func (r *DiscountsRepository) GetDiscountsByStatus(status string) ([]models.Discount, error) {
	var discounts []models.Discount
	now := time.Now()

	switch status {
	case "active":
		if err := r.db.Where("is_active = ? AND valid_from <= ? AND valid_until >= ?", true, now, now).Order("priority DESC, created_at DESC").Find(&discounts).Error; err != nil {
			return nil, errors.NewInternalServerError("Failed to retrieve active discounts")
		}
	case "expired":
		if err := r.db.Where("valid_until < ?", now).Order("valid_until DESC").Find(&discounts).Error; err != nil {
			return nil, errors.NewInternalServerError("Failed to retrieve expired discounts")
		}
	case "scheduled":
		if err := r.db.Where("valid_from > ?", now).Order("valid_from ASC").Find(&discounts).Error; err != nil {
			return nil, errors.NewInternalServerError("Failed to retrieve scheduled discounts")
		}
	case "inactive":
		if err := r.db.Where("is_active = ?", false).Order("created_at DESC").Find(&discounts).Error; err != nil {
			return nil, errors.NewInternalServerError("Failed to retrieve inactive discounts")
		}
	default:
		return nil, errors.NewBadRequestError("Invalid status")
	}

	return discounts, nil
}

// IncrementUsage increments the usage count of a discount
func (r *DiscountsRepository) IncrementUsage(id string) error {
	if err := r.db.Model(&models.Discount{}).Where("id = ?", id).UpdateColumn("current_usage", gorm.Expr("current_usage + ?", 1)).Error; err != nil {
		return errors.NewInternalServerError("Failed to increment discount usage")
	}
	return nil
}

// GetUsageByCustomer retrieves discount usage by customer
func (r *DiscountsRepository) GetUsageByCustomer(discountID, customerID string) (int, error) {
	var count int64
	if err := r.db.Model(&models.DiscountUsage{}).Where("discount_id = ? AND customer_id = ?", discountID, customerID).Count(&count).Error; err != nil {
		return 0, errors.NewInternalServerError("Failed to get discount usage by customer")
	}
	return int(count), nil
}

// CreateDiscountUsage creates a discount usage record
func (r *DiscountsRepository) CreateDiscountUsage(usage *models.DiscountUsage) error {
	if err := r.db.Create(usage).Error; err != nil {
		return errors.NewInternalServerError("Failed to create discount usage")
	}
	return nil
}

// GetDiscountUsageBySale retrieves discount usage by sale
func (r *DiscountsRepository) GetDiscountUsageBySale(saleID string) ([]models.DiscountUsage, error) {
	var usages []models.DiscountUsage
	if err := r.db.Where("sale_id = ?", saleID).Find(&usages).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve discount usage by sale")
	}
	return usages, nil
}

// ValidateDiscount validates a discount for a given order
func (r *DiscountsRepository) ValidateDiscount(code string, customerID *string, orderValue float64, productIDs []string, warehouseID string) (*models.Discount, error) {
	// Get discount by code
	discount, err := r.GetDiscountByCode(code)
	if err != nil {
		return nil, err
	}

	// Check if discount is active
	if !discount.IsActive {
		return nil, errors.NewBadRequestError("Discount is inactive")
	}

	// Check validity period
	now := time.Now()
	if now.Before(discount.ValidFrom) {
		return nil, errors.NewBadRequestError("Discount is not yet valid")
	}
	if now.After(discount.ValidUntil) {
		return nil, errors.NewBadRequestError("Discount has expired")
	}

	// Check usage limit
	if discount.UsageLimit != nil && discount.CurrentUsage >= *discount.UsageLimit {
		return nil, errors.NewBadRequestError("Discount usage limit reached")
	}

	// Check customer usage limit
	if customerID != nil && discount.UsagePerCustomer != nil {
		usageCount, err := r.GetUsageByCustomer(discount.ID, *customerID)
		if err != nil {
			return nil, err
		}
		if usageCount >= *discount.UsagePerCustomer {
			return nil, errors.NewBadRequestError("Customer usage limit reached for this discount")
		}
	}

	// Check minimum order value
	if discount.MinOrderValue != nil && orderValue < *discount.MinOrderValue {
		return nil, errors.NewBadRequestError("Order value does not meet minimum requirement")
	}

	// Check maximum order value
	if discount.MaxOrderValue != nil && orderValue > *discount.MaxOrderValue {
		return nil, errors.NewBadRequestError("Order value exceeds maximum limit for discount")
	}

	// Check warehouse applicability
	if discount.ApplicableWarehouses != nil {
		// TODO: Implement warehouse validation logic
		// This would require parsing the JSON array and checking if warehouseID is included
	}

	// Check product applicability
	if discount.ApplicableProducts != nil {
		// TODO: Implement product validation logic
		// This would require parsing the JSON array and checking if productIDs are included
	}

	// Check excluded products
	if discount.ExcludedProducts != nil {
		// TODO: Implement excluded product validation logic
		// This would require parsing the JSON array and checking if productIDs are excluded
	}

	return discount, nil
}

// CalculateDiscount calculates the discount amount for a given order
func (r *DiscountsRepository) CalculateDiscount(discount *models.Discount, orderValue float64) float64 {
	switch discount.DiscountType {
	case models.DiscountTypeFlat:
		// Flat discount - fixed amount
		return discount.Value
	case models.DiscountTypePercentage:
		// Percentage discount with optional maximum amount
		discountAmount := orderValue * (discount.Value / 100.0)
		if discount.MaxDiscountAmount != nil && discountAmount > *discount.MaxDiscountAmount {
			return *discount.MaxDiscountAmount
		}
		return discountAmount
	default:
		// For other discount types, return 0 for now
		// TODO: Implement other discount type calculations
		return 0
	}
}
