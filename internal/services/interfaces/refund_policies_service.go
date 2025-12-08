package interfaces

import (
	"kisanlink-erp/internal/database/models"
)

type RefundPoliciesServiceInterface interface {
	CreateRefundPolicy(req *models.CreateRefundPolicyRequest) (*models.RefundPolicyResponse, error)
	GetRefundPolicy(id string) (*models.RefundPolicyResponse, error)
	GetAllRefundPolicies(limit, offset int) ([]models.RefundPolicyResponse, int64, error)
	UpdateRefundPolicy(id string, req *models.UpdateRefundPolicyRequest) (*models.RefundPolicyResponse, error)
	DeleteRefundPolicy(id string) error
}
