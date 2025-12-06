package services

import (
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"

	"go.uber.org/zap"
)

type RefundPoliciesService struct {
	refundPoliciesRepo *repositories.RefundPoliciesRepository
	logger             interfaces.Logger
}

func NewRefundPoliciesService(refundPoliciesRepo *repositories.RefundPoliciesRepository, logger interfaces.Logger) *RefundPoliciesService {
	return &RefundPoliciesService{
		refundPoliciesRepo: refundPoliciesRepo,
		logger:             logger,
	}
}

// CreateRefundPolicy creates a new refund policy
func (s *RefundPoliciesService) CreateRefundPolicy(req *models.CreateRefundPolicyRequest) (*models.RefundPolicyResponse, error) {
	s.logger.Info("Creating refund policy",
		zap.String("policy_name", req.PolicyName),
		zap.Int("max_days", req.MaxDays),
		zap.Float64("restocking_fee", req.RestockingFee))

	// Validate request
	if err := s.validateRefundPolicyRequest(req); err != nil {
		s.logger.Warn("Refund policy validation failed",
			zap.Error(err))
		return nil, err
	}

	// Create refund policy using the proper constructor
	policy := models.NewRefundPolicy(req.PolicyName, req.Description, req.MaxDays, req.RestockingFee)

	if err := s.refundPoliciesRepo.CreateRefundPolicy(policy); err != nil {
		s.logger.Error("Failed to create refund policy",
			zap.Error(err),
			zap.String("policy_name", req.PolicyName))
		return nil, err
	}

	s.logger.Info("Refund policy created successfully",
		zap.String("policy_id", policy.ID),
		zap.String("policy_name", policy.PolicyName))

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
func (s *RefundPoliciesService) GetAllRefundPolicies(limit, offset int) ([]models.RefundPolicyResponse, int64, error) {
	policies, total, err := s.refundPoliciesRepo.GetAllRefundPolicies(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var responses []models.RefundPolicyResponse
	for _, policy := range policies {
		responses = append(responses, *s.mapRefundPolicyToResponse(&policy))
	}

	return responses, total, nil
}

// UpdateRefundPolicy updates a refund policy
func (s *RefundPoliciesService) UpdateRefundPolicy(id string, req *models.UpdateRefundPolicyRequest) (*models.RefundPolicyResponse, error) {
	s.logger.Info("Updating refund policy",
		zap.String("policy_id", id),
		zap.Bool("has_description", req.Description != nil),
		zap.Bool("has_max_days", req.MaxDays != nil),
		zap.Bool("has_restocking_fee", req.RestockingFee != nil))

	policy, err := s.refundPoliciesRepo.GetRefundPolicyByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve refund policy for update",
			zap.Error(err),
			zap.String("policy_id", id))
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
		s.logger.Error("Failed to update refund policy",
			zap.Error(err),
			zap.String("policy_id", id))
		return nil, err
	}

	s.logger.Info("Refund policy updated successfully",
		zap.String("policy_id", id))

	return s.mapRefundPolicyToResponse(policy), nil
}

// DeleteRefundPolicy deletes a refund policy
func (s *RefundPoliciesService) DeleteRefundPolicy(id string) error {
	s.logger.Info("Deleting refund policy",
		zap.String("policy_id", id))

	if err := s.refundPoliciesRepo.DeleteRefundPolicy(id); err != nil {
		s.logger.Error("Failed to delete refund policy",
			zap.Error(err),
			zap.String("policy_id", id))
		return err
	}

	s.logger.Info("Refund policy deleted successfully",
		zap.String("policy_id", id))

	return nil
}

// Helper methods
func (s *RefundPoliciesService) validateRefundPolicyRequest(req *models.CreateRefundPolicyRequest) error {
	if req.PolicyName == "" {
		return errors.NewValidationError("policy name is required")
	}
	if req.MaxDays <= 0 {
		return errors.NewValidationError("max days must be greater than 0")
	}
	if req.RestockingFee < 0 {
		return errors.NewValidationError("restocking fee cannot be negative")
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
