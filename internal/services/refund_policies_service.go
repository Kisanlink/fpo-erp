package services

import (
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
)

type RefundPoliciesService struct {
	refundPoliciesRepo *repositories.RefundPoliciesRepository
}

func NewRefundPoliciesService(refundPoliciesRepo *repositories.RefundPoliciesRepository) *RefundPoliciesService {
	return &RefundPoliciesService{
		refundPoliciesRepo: refundPoliciesRepo,
	}
}

// CreateRefundPolicy creates a new refund policy
func (s *RefundPoliciesService) CreateRefundPolicy(req *models.CreateRefundPolicyRequest) (*models.RefundPolicyResponse, error) {
	// Validate request
	if err := s.validateRefundPolicyRequest(req); err != nil {
		return nil, err
	}

	// Create refund policy using the proper constructor
	policy := models.NewRefundPolicy(req.PolicyName, req.Description, req.MaxDays, req.RestockingFee)

	if err := s.refundPoliciesRepo.CreateRefundPolicy(policy); err != nil {
		return nil, err
	}

	return s.mapRefundPolicyToResponse(policy), nil
}

// GetRefundPolicy retrieves a refund policy by ID
func (s *RefundPoliciesService) GetRefundPolicy(id string) (*models.RefundPolicyResponse, error) {
	policy, err := s.refundPoliciesRepo.GetRefundPolicyByID(id)
	if err != nil {
		return nil, errors.NewNotFoundError("Refund policy not found")
	}
	return s.mapRefundPolicyToResponse(policy), nil
}

// GetAllRefundPolicies retrieves all refund policies with pagination
func (s *RefundPoliciesService) GetAllRefundPolicies(limit, offset int) ([]models.RefundPolicyResponse, error) {
	policies, err := s.refundPoliciesRepo.GetAllRefundPolicies(limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []models.RefundPolicyResponse
	for _, policy := range policies {
		responses = append(responses, *s.mapRefundPolicyToResponse(&policy))
	}

	return responses, nil
}

// UpdateRefundPolicy updates a refund policy
func (s *RefundPoliciesService) UpdateRefundPolicy(id string, req *models.UpdateRefundPolicyRequest) (*models.RefundPolicyResponse, error) {
	policy, err := s.refundPoliciesRepo.GetRefundPolicyByID(id)
	if err != nil {
		return nil, errors.NewNotFoundError("Refund policy not found")
	}

	// Update fields
	if req.Description != nil {
		policy.Description = req.Description
	}
	if req.MaxDays != nil {
		policy.MaxDays = *req.MaxDays
	}
	if req.RestockingFee != nil {
		policy.RestockingFee = *req.RestockingFee
	}

	if err := s.refundPoliciesRepo.UpdateRefundPolicy(policy); err != nil {
		return nil, err
	}

	return s.mapRefundPolicyToResponse(policy), nil
}

// DeleteRefundPolicy deletes a refund policy
func (s *RefundPoliciesService) DeleteRefundPolicy(id string) error {
	return s.refundPoliciesRepo.DeleteRefundPolicy(id)
}

// Helper methods
func (s *RefundPoliciesService) validateRefundPolicyRequest(req *models.CreateRefundPolicyRequest) error {
	if req.PolicyName == "" {
		return errors.NewNotFoundError("policy name is required")
	}
	if req.MaxDays <= 0 {
		return errors.NewNotFoundError("max days must be greater than 0")
	}
	if req.RestockingFee < 0 {
		return errors.NewNotFoundError("restocking fee cannot be negative")
	}
	return nil
}

func (s *RefundPoliciesService) mapRefundPolicyToResponse(policy *models.RefundPolicy) *models.RefundPolicyResponse {
	return &models.RefundPolicyResponse{
		ID:            policy.ID,
		PolicyName:    policy.PolicyName,
		Description:   policy.Description,
		MaxDays:       policy.MaxDays,
		RestockingFee: policy.RestockingFee,
		CreatedAt:     policy.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     policy.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
