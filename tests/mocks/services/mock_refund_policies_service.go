package services

import (
	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockRefundPoliciesService is a mock implementation of RefundPoliciesServiceInterface
type MockRefundPoliciesService struct {
	mock.Mock
}

func (m *MockRefundPoliciesService) CreateRefundPolicy(req *models.CreateRefundPolicyRequest) (*models.RefundPolicyResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RefundPolicyResponse), args.Error(1)
}

func (m *MockRefundPoliciesService) GetRefundPolicy(id string) (*models.RefundPolicyResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RefundPolicyResponse), args.Error(1)
}

func (m *MockRefundPoliciesService) GetAllRefundPolicies(limit, offset int) ([]models.RefundPolicyResponse, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RefundPolicyResponse), args.Error(1)
}

func (m *MockRefundPoliciesService) UpdateRefundPolicy(id string, req *models.UpdateRefundPolicyRequest) (*models.RefundPolicyResponse, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RefundPolicyResponse), args.Error(1)
}

func (m *MockRefundPoliciesService) DeleteRefundPolicy(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
