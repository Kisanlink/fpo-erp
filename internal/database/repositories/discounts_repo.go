package repositories

import (
	"encoding/json"
	"sort"
	"strings"
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

// CreateDiscountUsage creates a discount usage record
func (r *DiscountsRepository) CreateDiscountUsage(usage *models.DiscountUsage) error {
	if err := r.db.Create(usage).Error; err != nil {
		return errors.NewInternalServerError("Failed to create discount usage")
	}
	return nil
}

// CreateDiscountUsageWithTx creates a discount usage record within a transaction
func (r *DiscountsRepository) CreateDiscountUsageWithTx(tx *gorm.DB, usage *models.DiscountUsage) error {
	if err := tx.Create(usage).Error; err != nil {
		return errors.NewInternalServerError("Failed to create discount usage")
	}
	return nil
}

// IncrementUsageWithTx increments discount usage count within a transaction
func (r *DiscountsRepository) IncrementUsageWithTx(tx *gorm.DB, discountID string) error {
	if err := tx.Model(&models.Discount{}).Where("id = ?", discountID).Update("current_usage", gorm.Expr("current_usage + ?", 1)).Error; err != nil {
		return errors.NewInternalServerError("Failed to increment discount usage")
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
func (r *DiscountsRepository) ValidateDiscount(code string, orderValue float64, productIDs []string, warehouseID string) (*models.Discount, error) {
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

	// Check minimum order value
	if discount.MinOrderValue != nil && orderValue < *discount.MinOrderValue {
		return nil, errors.NewBadRequestError("Order value does not meet minimum requirement")
	}

	// Check maximum order value
	if discount.MaxOrderValue != nil && orderValue > *discount.MaxOrderValue {
		return nil, errors.NewBadRequestError("Order value exceeds maximum limit for discount")
	}

	// Check warehouse applicability
	if discount.ApplicableWarehouses != nil && *discount.ApplicableWarehouses != "" {
		if !r.isIDInJSONArray(*discount.ApplicableWarehouses, warehouseID) {
			return nil, errors.NewBadRequestError("Discount is not applicable to this warehouse")
		}
	}

	// Check product applicability
	if discount.ApplicableProducts != nil && *discount.ApplicableProducts != "" {
		if !r.areProductIDsApplicable(*discount.ApplicableProducts, productIDs) {
			return nil, errors.NewBadRequestError("Discount is not applicable to the products in this order")
		}
	}

	// Check excluded products
	if discount.ExcludedProducts != nil && *discount.ExcludedProducts != "" {
		if r.areProductIDsExcluded(*discount.ExcludedProducts, productIDs) {
			return nil, errors.NewBadRequestError("Discount cannot be applied to excluded products in this order")
		}
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
	case models.DiscountTypeBuyXGetY:
		// Buy X Get Y discount - requires item-level analysis
		// This method doesn't have access to individual items, so return 0
		// Use CalculateBuyXGetYDiscount method for proper calculation
		return 0
	case models.DiscountTypeSeasonal:
		// Seasonal discount - time-based percentage or flat amount
		discountAmount := orderValue * (discount.Value / 100.0)
		if discount.MaxDiscountAmount != nil && discountAmount > *discount.MaxDiscountAmount {
			return *discount.MaxDiscountAmount
		}
		return discountAmount
	case models.DiscountTypeBulk:
		// Bulk discount - typically percentage based on quantity thresholds
		// This would need product quantity information
		discountAmount := orderValue * (discount.Value / 100.0)
		if discount.MaxDiscountAmount != nil && discountAmount > *discount.MaxDiscountAmount {
			return *discount.MaxDiscountAmount
		}
		return discountAmount
	default:
		// Unknown discount type
		return 0
	}
}

// ValidateDiscountWithCategories validates a discount for a given order including category checks
func (r *DiscountsRepository) ValidateDiscountWithCategories(code string, orderValue float64, productIDs []string, categoryIDs []string, warehouseID string) (*models.Discount, error) {
	// First perform basic validation
	discount, err := r.ValidateDiscount(code, orderValue, productIDs, warehouseID)
	if err != nil {
		return nil, err
	}

	// Check applicable categories
	if discount.ApplicableCategories != nil && *discount.ApplicableCategories != "" {
		if !r.areCategoryIDsApplicable(*discount.ApplicableCategories, categoryIDs) {
			return nil, errors.NewBadRequestError("Discount is not applicable to the product categories in this order")
		}
	}

	// Check excluded categories
	if discount.ExcludedCategories != nil && *discount.ExcludedCategories != "" {
		if r.areCategoryIDsExcluded(*discount.ExcludedCategories, categoryIDs) {
			return nil, errors.NewBadRequestError("Discount cannot be applied to excluded product categories in this order")
		}
	}

	return discount, nil
}

// Helper methods for JSON parsing and validation

// isIDInJSONArray checks if a given ID exists in a JSON array string
func (r *DiscountsRepository) isIDInJSONArray(jsonArrayStr, targetID string) bool {
	if jsonArrayStr == "" || targetID == "" {
		return false
	}

	var ids []string
	if err := json.Unmarshal([]byte(jsonArrayStr), &ids); err != nil {
		// If JSON parsing fails, try treating it as a comma-separated string
		ids = strings.Split(jsonArrayStr, ",")
		for i, id := range ids {
			ids[i] = strings.TrimSpace(id)
		}
	}

	for _, id := range ids {
		if strings.TrimSpace(id) == targetID {
			return true
		}
	}
	return false
}

// areProductIDsApplicable checks if any of the given product IDs are in the applicable products JSON array
func (r *DiscountsRepository) areProductIDsApplicable(jsonArrayStr string, productIDs []string) bool {
	if jsonArrayStr == "" || len(productIDs) == 0 {
		return true // If no restrictions, all products are applicable
	}

	var applicableIDs []string
	if err := json.Unmarshal([]byte(jsonArrayStr), &applicableIDs); err != nil {
		// If JSON parsing fails, try treating it as a comma-separated string
		applicableIDs = strings.Split(jsonArrayStr, ",")
		for i, id := range applicableIDs {
			applicableIDs[i] = strings.TrimSpace(id)
		}
	}

	// Check if any of the order's product IDs are in the applicable list
	for _, orderProductID := range productIDs {
		for _, applicableID := range applicableIDs {
			if strings.TrimSpace(applicableID) == orderProductID {
				return true // At least one product is applicable
			}
		}
	}
	return false
}

// areProductIDsExcluded checks if any of the given product IDs are in the excluded products JSON array
func (r *DiscountsRepository) areProductIDsExcluded(jsonArrayStr string, productIDs []string) bool {
	if jsonArrayStr == "" || len(productIDs) == 0 {
		return false // If no exclusions, no products are excluded
	}

	var excludedIDs []string
	if err := json.Unmarshal([]byte(jsonArrayStr), &excludedIDs); err != nil {
		// If JSON parsing fails, try treating it as a comma-separated string
		excludedIDs = strings.Split(jsonArrayStr, ",")
		for i, id := range excludedIDs {
			excludedIDs[i] = strings.TrimSpace(id)
		}
	}

	// Check if any of the order's product IDs are in the excluded list
	for _, orderProductID := range productIDs {
		for _, excludedID := range excludedIDs {
			if strings.TrimSpace(excludedID) == orderProductID {
				return true // At least one product is excluded
			}
		}
	}
	return false
}

// areCategoryIDsApplicable checks if any of the given category IDs are in the applicable categories JSON array
func (r *DiscountsRepository) areCategoryIDsApplicable(jsonArrayStr string, categoryIDs []string) bool {
	if jsonArrayStr == "" || len(categoryIDs) == 0 {
		return true // If no restrictions, all categories are applicable
	}

	var applicableIDs []string
	if err := json.Unmarshal([]byte(jsonArrayStr), &applicableIDs); err != nil {
		// If JSON parsing fails, try treating it as a comma-separated string
		applicableIDs = strings.Split(jsonArrayStr, ",")
		for i, id := range applicableIDs {
			applicableIDs[i] = strings.TrimSpace(id)
		}
	}

	// Check if any of the order's category IDs are in the applicable list
	for _, orderCategoryID := range categoryIDs {
		for _, applicableID := range applicableIDs {
			if strings.TrimSpace(applicableID) == orderCategoryID {
				return true // At least one category is applicable
			}
		}
	}
	return false
}

// areCategoryIDsExcluded checks if any of the given category IDs are in the excluded categories JSON array
func (r *DiscountsRepository) areCategoryIDsExcluded(jsonArrayStr string, categoryIDs []string) bool {
	if jsonArrayStr == "" || len(categoryIDs) == 0 {
		return false // If no exclusions, no categories are excluded
	}

	var excludedIDs []string
	if err := json.Unmarshal([]byte(jsonArrayStr), &excludedIDs); err != nil {
		// If JSON parsing fails, try treating it as a comma-separated string
		excludedIDs = strings.Split(jsonArrayStr, ",")
		for i, id := range excludedIDs {
			excludedIDs[i] = strings.TrimSpace(id)
		}
	}

	// Check if any of the order's category IDs are in the excluded list
	for _, orderCategoryID := range categoryIDs {
		for _, excludedID := range excludedIDs {
			if strings.TrimSpace(excludedID) == orderCategoryID {
				return true // At least one category is excluded
			}
		}
	}
	return false
}

// Additional utility methods for advanced discount operations

// GetApplicableDiscountsForOrder retrieves all applicable discounts for a given order
func (r *DiscountsRepository) GetApplicableDiscountsForOrder(orderValue float64, productIDs []string, categoryIDs []string, warehouseID string) ([]models.Discount, error) {
	// Get all active discounts
	activeDiscounts, err := r.GetActiveDiscounts()
	if err != nil {
		return nil, err
	}

	var applicableDiscounts []models.Discount

	for _, discount := range activeDiscounts {
		// Validate each discount for the order
		_, err := r.ValidateDiscountWithCategories(discount.Code, orderValue, productIDs, categoryIDs, warehouseID)
		if err == nil {
			applicableDiscounts = append(applicableDiscounts, discount)
		}
	}

	// Sort by priority (higher priority first)
	sort.Slice(applicableDiscounts, func(i, j int) bool {
		return applicableDiscounts[i].Priority > applicableDiscounts[j].Priority
	})

	return applicableDiscounts, nil
}

// CalculateOptimalDiscounts calculates the best combination of discounts for an order
func (r *DiscountsRepository) CalculateOptimalDiscounts(orderValue float64, productIDs []string, categoryIDs []string, warehouseID string) ([]models.Discount, float64, error) {
	applicableDiscounts, err := r.GetApplicableDiscountsForOrder(orderValue, productIDs, categoryIDs, warehouseID)
	if err != nil {
		return nil, 0, err
	}

	if len(applicableDiscounts) == 0 {
		return []models.Discount{}, 0, nil
	}

	// For now, implement a simple strategy: use the highest priority discount
	// In a more complex system, you might want to implement:
	// - Stackable discount combinations
	// - Dynamic programming for optimal discount selection
	// - Customer preference-based selection

	bestDiscount := applicableDiscounts[0]
	totalDiscount := r.CalculateDiscount(&bestDiscount, orderValue)

	return []models.Discount{bestDiscount}, totalDiscount, nil
}

// GetDiscountUsageStats retrieves usage statistics for a discount
func (r *DiscountsRepository) GetDiscountUsageStats(discountID string) (map[string]interface{}, error) {
	var totalUsage int64
	var totalAmount float64

	// Get total usage count
	if err := r.db.Model(&models.DiscountUsage{}).Where("discount_id = ?", discountID).Count(&totalUsage).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to get total usage count")
	}

	// Get total discount amount
	if err := r.db.Model(&models.DiscountUsage{}).Where("discount_id = ?", discountID).Select("COALESCE(SUM(amount), 0)").Scan(&totalAmount).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to get total discount amount")
	}

	stats := map[string]interface{}{
		"total_usage":              totalUsage,
		"total_discount_amount":    totalAmount,
		"average_discount_per_use": 0.0,
	}

	if totalUsage > 0 {
		stats["average_discount_per_use"] = totalAmount / float64(totalUsage)
	}

	return stats, nil
}

// SaleItem represents an item in a sale for discount calculation
type SaleItem struct {
	ProductID string
	Quantity  int64
	Price     float64
}

// CalculateBuyXGetYDiscount calculates Buy X Get Y discount for specific sale items
func (r *DiscountsRepository) CalculateBuyXGetYDiscount(discount models.Discount, items []SaleItem) float64 {
	// Validate Buy X Get Y configuration
	if discount.BuyQuantity == nil || discount.GetQuantity == nil ||
		*discount.BuyQuantity <= 0 || *discount.GetQuantity <= 0 {
		return 0
	}

	// Parse applicable products
	var applicableProducts []string
	if discount.ApplicableProducts != nil {
		if err := json.Unmarshal([]byte(*discount.ApplicableProducts), &applicableProducts); err != nil {
			return 0
		}
	}

	// Count eligible items
	eligibleQuantity := int64(0)
	eligibleItems := make([]SaleItem, 0)

	for _, item := range items {
		// Check if product is applicable
		if len(applicableProducts) > 0 {
			isApplicable := false
			for _, productID := range applicableProducts {
				if productID == item.ProductID {
					isApplicable = true
					break
				}
			}
			if !isApplicable {
				continue
			}
		}

		eligibleQuantity += item.Quantity
		eligibleItems = append(eligibleItems, item)
	}

	// Calculate how many complete sets of Buy X Get Y we can apply
	buyQuantity := int64(*discount.BuyQuantity)
	getQuantity := int64(*discount.GetQuantity)

	// Number of complete Buy X sets
	completeSets := eligibleQuantity / buyQuantity

	if completeSets == 0 {
		return 0
	}

	// Calculate discount amount based on Get Y items
	totalGetQuantity := completeSets * getQuantity

	// Limit get quantity to available eligible items
	if totalGetQuantity > eligibleQuantity {
		totalGetQuantity = eligibleQuantity
	}

	// Calculate discount value
	discountAmount := 0.0
	getDiscountType := "free" // default
	if discount.GetDiscountType != nil {
		getDiscountType = *discount.GetDiscountType
	}

	// Sort items by price (ascending) to apply discount to cheapest eligible items
	// This is a common Buy X Get Y strategy
	sortedItems := make([]SaleItem, len(eligibleItems))
	copy(sortedItems, eligibleItems)

	// Sort by price (ascending) to discount cheapest items first
	sort.Slice(sortedItems, func(i, j int) bool { return sortedItems[i].Price < sortedItems[j].Price })

	// Apply discount to cheapest items up to totalGetQuantity
	remainingGetQuantity := totalGetQuantity
	for _, item := range sortedItems {
		if remainingGetQuantity <= 0 {
			break
		}

		itemGetQuantity := item.Quantity
		if itemGetQuantity > remainingGetQuantity {
			itemGetQuantity = remainingGetQuantity
		}

		switch getDiscountType {
		case "free":
			discountAmount += float64(itemGetQuantity) * item.Price
		case "percentage":
			if discount.GetDiscountValue != nil {
				discountAmount += float64(itemGetQuantity) * item.Price * (*discount.GetDiscountValue / 100.0)
			}
		case "flat":
			if discount.GetDiscountValue != nil {
				discountAmount += float64(itemGetQuantity) * (*discount.GetDiscountValue)
			}
		}

		remainingGetQuantity -= itemGetQuantity
	}

	// Apply max discount limit if set
	if discount.MaxDiscountAmount != nil && discountAmount > *discount.MaxDiscountAmount {
		discountAmount = *discount.MaxDiscountAmount
	}

	return discountAmount
}
