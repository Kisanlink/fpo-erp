package repositories

import (
	"kisanlink-erp/internal/database/models"

	"gorm.io/gorm"
)

type RefundPoliciesRepository struct {
	db *gorm.DB
}

func NewRefundPoliciesRepository(db *gorm.DB) *RefundPoliciesRepository {
	return &RefundPoliciesRepository{db: db}
}

// CreateRefundPolicy creates a new refund policy
func (r *RefundPoliciesRepository) CreateRefundPolicy(policy *models.RefundPolicy) error {
	return r.db.Create(policy).Error
}

// GetRefundPolicyByID retrieves a refund policy by ID
func (r *RefundPoliciesRepository) GetRefundPolicyByID(id string) (*models.RefundPolicy, error) {
	var policy models.RefundPolicy
	err := r.db.First(&policy, "id = ?", id).Error
	return &policy, err
}

// GetAllRefundPolicies retrieves all refund policies with pagination
func (r *RefundPoliciesRepository) GetAllRefundPolicies(limit, offset int) ([]models.RefundPolicy, int64, error) {
	var policies []models.RefundPolicy
	var total int64

	// Get total count
	if err := r.db.Model(&models.RefundPolicy{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated records
	err := r.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&policies).Error
	return policies, total, err
}

// UpdateRefundPolicy updates a refund policy
func (r *RefundPoliciesRepository) UpdateRefundPolicy(policy *models.RefundPolicy) error {
	return r.db.Save(policy).Error
}

// DeleteRefundPolicy deletes a refund policy
func (r *RefundPoliciesRepository) DeleteRefundPolicy(id string) error {
	return r.db.Delete(&models.RefundPolicy{}, "id = ?", id).Error
}
