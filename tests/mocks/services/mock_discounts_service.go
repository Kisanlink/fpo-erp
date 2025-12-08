package services

import (
	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockDiscountsService is a mock implementation of DiscountsServiceInterface
type MockDiscountsService struct {
	mock.Mock
}

func (m *MockDiscountsService) CreateDiscount(req *models.CreateDiscountRequest) (*models.DiscountResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiscountResponse), args.Error(1)
}

func (m *MockDiscountsService) GetDiscount(id string) (*models.DiscountResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiscountResponse), args.Error(1)
}

func (m *MockDiscountsService) GetAllDiscounts(limit, offset int) ([]models.DiscountResponse, int64, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.DiscountResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockDiscountsService) GetActiveDiscounts(limit, offset int) ([]models.DiscountResponse, int64, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.DiscountResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockDiscountsService) UpdateDiscount(id string, req *models.UpdateDiscountRequest) (*models.DiscountResponse, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiscountResponse), args.Error(1)
}

func (m *MockDiscountsService) DeleteDiscount(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockDiscountsService) GetDiscountsByType(discountType models.DiscountType) ([]models.DiscountResponse, error) {
	args := m.Called(discountType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DiscountResponse), args.Error(1)
}

func (m *MockDiscountsService) GetDiscountsByStatus(status string) ([]models.DiscountResponse, error) {
	args := m.Called(status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DiscountResponse), args.Error(1)
}

func (m *MockDiscountsService) ValidateDiscount(req *models.ValidateDiscountRequest) (*models.DiscountValidationResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiscountValidationResponse), args.Error(1)
}

func (m *MockDiscountsService) ApplyDiscount(discountID, saleID string, discountAmount float64) error {
	args := m.Called(discountID, saleID, discountAmount)
	return args.Error(0)
}

func (m *MockDiscountsService) GetDiscountUsageBySale(saleID string) ([]models.DiscountUsageResponse, error) {
	args := m.Called(saleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DiscountUsageResponse), args.Error(1)
}

func (m *MockDiscountsService) GetDiscountUsageStats(discountID string) (map[string]interface{}, error) {
	args := m.Called(discountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockDiscountsService) GetApplicableDiscountsForOrder(orderValue float64, productIDs []string, categoryIDs []string, warehouseID string) ([]models.DiscountResponse, error) {
	args := m.Called(orderValue, productIDs, categoryIDs, warehouseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DiscountResponse), args.Error(1)
}

func (m *MockDiscountsService) CalculateOptimalDiscounts(orderValue float64, productIDs []string, categoryIDs []string, warehouseID string) ([]models.DiscountResponse, float64, error) {
	args := m.Called(orderValue, productIDs, categoryIDs, warehouseID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(float64), args.Error(2)
	}
	return args.Get(0).([]models.DiscountResponse), args.Get(1).(float64), args.Error(2)
}

func (m *MockDiscountsService) ValidateDiscountWithCategories(req *models.ValidateDiscountRequest, categoryIDs []string) (*models.DiscountValidationResponse, error) {
	args := m.Called(req, categoryIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiscountValidationResponse), args.Error(1)
}
