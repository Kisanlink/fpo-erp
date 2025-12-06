package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"kisanlink-erp/internal/api/handlers"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/utils"
	mockServices "kisanlink-erp/tests/mocks/services"
	"kisanlink-erp/tests/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	// Bypass AAA for testing
	os.Setenv("AAA_ENABLED", "false")
	gin.SetMode(gin.TestMode)
}

// TestDiscountsHandler_CreateDiscount_Success tests successful discount creation
func TestDiscountsHandler_CreateDiscount_Success(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	request := &models.CreateDiscountRequest{
		Code:         "SUMMER20",
		Name:         "Summer Sale",
		DiscountType: "percentage",
		Value:        20.0,
		ValidFrom:    "2025-06-01",
		ValidUntil:   "2025-08-31",
	}

	expectedResponse := &models.DiscountResponse{
		ID:           "DISC00000001",
		Code:         "SUMMER20",
		Name:         "Summer Sale",
		DiscountType: "percentage",
		Value:        20.0,
		IsActive:     true,
	}

	mockService.On("CreateDiscount", mock.AnythingOfType("*models.CreateDiscountRequest")).Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/discounts", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateDiscount(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Discount created successfully", response["message"])
	assert.NotNil(t, response["data"])
}

// TestDiscountsHandler_CreateDiscount_ValidationError tests validation errors
func TestDiscountsHandler_CreateDiscount_ValidationError(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	// Missing required fields
	request := &models.CreateDiscountRequest{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/discounts", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateDiscount(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "CreateDiscount", mock.Anything)
}

// TestDiscountsHandler_CreateDiscount_ServiceError tests service layer errors
func TestDiscountsHandler_CreateDiscount_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	request := &models.CreateDiscountRequest{
		Code:         "SUMMER20",
		Name:         "Summer Sale",
		DiscountType: "percentage",
		Value:        20.0,
		ValidFrom:    "2025-06-01",
		ValidUntil:   "2025-08-31",
	}

	mockService.On("CreateDiscount", mock.AnythingOfType("*models.CreateDiscountRequest")).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/discounts", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateDiscount(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_GetDiscount_Success tests fetching a single discount
func TestDiscountsHandler_GetDiscount_Success(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	expectedDiscount := &models.DiscountResponse{
		ID:           "DISC00000001",
		Code:         "SUMMER20",
		Name:         "Summer Sale",
		DiscountType: "percentage",
		Value:        20.0,
	}

	mockService.On("GetDiscount", "DISC00000001").Return(expectedDiscount, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "DISC00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/DISC00000001", nil)

	handler.GetDiscount(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Discount retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])
}

// TestDiscountsHandler_GetDiscount_NotFound tests discount not found
func TestDiscountsHandler_GetDiscount_NotFound(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetDiscount", "DISC99999999").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "DISC99999999"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/DISC99999999", nil)

	handler.GetDiscount(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_GetAllDiscounts_Success tests fetching all discounts
func TestDiscountsHandler_GetAllDiscounts_Success(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	expectedDiscounts := []models.DiscountResponse{
		{
			ID:           "DISC00000001",
			Code:         "SUMMER20",
			Name:         "Summer Sale",
			DiscountType: "percentage",
			Value:        20.0,
		},
		{
			ID:           "DISC00000002",
			Code:         "FLAT100",
			Name:         "Flat 100 Off",
			DiscountType: "flat",
			Value:        100.0,
		},
	}

	mockService.On("GetAllDiscounts", 50, 0).Return(expectedDiscounts, int64(2), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts", nil)

	handler.GetAllDiscounts(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["data"])
	assert.NotNil(t, response["pagination"])
}

// TestDiscountsHandler_GetAllDiscounts_WithPagination tests pagination
func TestDiscountsHandler_GetAllDiscounts_WithPagination(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetAllDiscounts", 10, 20).Return([]models.DiscountResponse{}, int64(0), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts?limit=10&offset=20", nil)

	handler.GetAllDiscounts(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_GetActiveDiscounts_Success tests fetching active discounts
func TestDiscountsHandler_GetActiveDiscounts_Success(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	expectedDiscounts := []models.DiscountResponse{
		{
			ID:       "DISC00000001",
			Code:     "SUMMER20",
			IsActive: true,
		},
	}

	mockService.On("GetActiveDiscounts", 50, 0).Return(expectedDiscounts, int64(1), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/active", nil)

	handler.GetActiveDiscounts(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_GetActiveDiscounts_ServiceError tests service error
func TestDiscountsHandler_GetActiveDiscounts_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetActiveDiscounts", 50, 0).Return(nil, int64(0), assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/active", nil)

	handler.GetActiveDiscounts(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_UpdateDiscount_Success tests successful discount update
func TestDiscountsHandler_UpdateDiscount_Success(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	value := 25.0
	request := &models.UpdateDiscountRequest{
		Name:  testutils.StringPtr("Updated Summer Sale"),
		Value: &value,
	}

	expectedResponse := &models.DiscountResponse{
		ID:    "DISC00000001",
		Name:  "Updated Summer Sale",
		Value: 25.0,
	}

	mockService.On("UpdateDiscount", "DISC00000001", mock.AnythingOfType("*models.UpdateDiscountRequest")).Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "DISC00000001"}}
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("PUT", "/api/v1/discounts/DISC00000001", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateDiscount(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_UpdateDiscount_ServiceError tests service error
func TestDiscountsHandler_UpdateDiscount_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	request := &models.UpdateDiscountRequest{
		Name: testutils.StringPtr("Updated Name"),
	}

	mockService.On("UpdateDiscount", "DISC00000001", mock.AnythingOfType("*models.UpdateDiscountRequest")).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "DISC00000001"}}
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("PUT", "/api/v1/discounts/DISC00000001", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateDiscount(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_DeleteDiscount_Success tests successful discount deletion
func TestDiscountsHandler_DeleteDiscount_Success(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	mockService.On("DeleteDiscount", "DISC00000001").Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "DISC00000001"}}
	c.Request, _ = http.NewRequest("DELETE", "/api/v1/discounts/DISC00000001", nil)

	handler.DeleteDiscount(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_DeleteDiscount_ServiceError tests service error
func TestDiscountsHandler_DeleteDiscount_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	mockService.On("DeleteDiscount", "DISC00000001").Return(assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "DISC00000001"}}
	c.Request, _ = http.NewRequest("DELETE", "/api/v1/discounts/DISC00000001", nil)

	handler.DeleteDiscount(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_GetDiscountsByType_Success tests fetching discounts by type
func TestDiscountsHandler_GetDiscountsByType_Success(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	expectedDiscounts := []models.DiscountResponse{
		{
			ID:           "DISC00000001",
			DiscountType: "percentage",
		},
	}

	mockService.On("GetDiscountsByType", models.DiscountType("percentage")).Return(expectedDiscounts, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "type", Value: "percentage"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/type/percentage", nil)

	handler.GetDiscountsByType(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_GetDiscountsByType_InvalidType tests invalid discount type
func TestDiscountsHandler_GetDiscountsByType_InvalidType(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	// Handler doesn't validate type, so service is called and returns empty results
	mockService.On("GetDiscountsByType", models.DiscountType("invalid_type")).Return([]models.DiscountResponse{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "type", Value: "invalid_type"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/type/invalid_type", nil)

	handler.GetDiscountsByType(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_GetDiscountsByType_ServiceError tests service error
func TestDiscountsHandler_GetDiscountsByType_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetDiscountsByType", models.DiscountType("percentage")).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "type", Value: "percentage"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/type/percentage", nil)

	handler.GetDiscountsByType(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_GetDiscountsByStatus_Success tests fetching discounts by status
func TestDiscountsHandler_GetDiscountsByStatus_Success(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	expectedDiscounts := []models.DiscountResponse{
		{
			ID:       "DISC00000001",
			IsActive: true,
		},
	}

	mockService.On("GetDiscountsByStatus", "active").Return(expectedDiscounts, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "status", Value: "active"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/status/active", nil)

	handler.GetDiscountsByStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_GetDiscountsByStatus_InvalidStatus tests invalid status
func TestDiscountsHandler_GetDiscountsByStatus_InvalidStatus(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	// Handler doesn't validate status, so service is called and returns empty results
	mockService.On("GetDiscountsByStatus", "invalid_status").Return([]models.DiscountResponse{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "status", Value: "invalid_status"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/status/invalid_status", nil)

	handler.GetDiscountsByStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_GetDiscountsByStatus_ServiceError tests service error
func TestDiscountsHandler_GetDiscountsByStatus_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetDiscountsByStatus", "active").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "status", Value: "active"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/status/active", nil)

	handler.GetDiscountsByStatus(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_ValidateDiscount_Success tests successful discount validation
func TestDiscountsHandler_ValidateDiscount_Success(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	request := &models.ValidateDiscountRequest{
		DiscountCode: "SUMMER20",
		OrderValue:   1000.0,
		WarehouseID:  "WARE00000001",
	}

	expectedResponse := &models.DiscountValidationResponse{
		IsValid:            true,
		CalculatedDiscount: 200.0,
		Message:            "Discount is valid",
	}

	mockService.On("ValidateDiscount", mock.AnythingOfType("*models.ValidateDiscountRequest")).Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/discounts/validate", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ValidateDiscount(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_ValidateDiscount_ValidationError tests validation error
func TestDiscountsHandler_ValidateDiscount_ValidationError(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	request := &models.ValidateDiscountRequest{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/discounts/validate", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ValidateDiscount(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "ValidateDiscount", mock.Anything)
}

// TestDiscountsHandler_ValidateDiscount_ServiceError tests service error
func TestDiscountsHandler_ValidateDiscount_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	request := &models.ValidateDiscountRequest{
		DiscountCode: "SUMMER20",
		OrderValue:   1000.0,
		WarehouseID:  "WARE00000001",
	}

	mockService.On("ValidateDiscount", mock.AnythingOfType("*models.ValidateDiscountRequest")).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/discounts/validate", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ValidateDiscount(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_GetDiscountUsageBySale_Success tests fetching discount usage by sale
func TestDiscountsHandler_GetDiscountUsageBySale_Success(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	expectedUsage := []models.DiscountUsageResponse{
		{
			ID:         "DUSG00000001",
			DiscountID: "DISC00000001",
			SaleID:     "SALE00000001",
			Amount:     200.0,
		},
	}

	mockService.On("GetDiscountUsageBySale", "SALE00000001").Return(expectedUsage, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "saleID", Value: "SALE00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/usage/sale/SALE00000001", nil)

	handler.GetDiscountUsageBySale(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_GetDiscountUsageBySale_ServiceError tests service error
func TestDiscountsHandler_GetDiscountUsageBySale_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetDiscountUsageBySale", "SALE00000001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "saleID", Value: "SALE00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/usage/sale/SALE00000001", nil)

	handler.GetDiscountUsageBySale(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_GetApplicableDiscounts_Success tests fetching applicable discounts
func TestDiscountsHandler_GetApplicableDiscounts_Success(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	expectedDiscounts := []models.DiscountResponse{
		{
			ID:           "DISC00000001",
			Code:         "SUMMER20",
			DiscountType: "percentage",
			Value:        20.0,
		},
	}

	mockService.On("GetApplicableDiscountsForOrder", 1000.0, []string{"PROD00000001"}, []string(nil), "WHSE00000001").Return(expectedDiscounts, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/applicable?order_value=1000.0&warehouse_id=WHSE00000001&product_ids=PROD00000001", nil)

	handler.GetApplicableDiscounts(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_GetApplicableDiscounts_MissingParams tests missing required params
func TestDiscountsHandler_GetApplicableDiscounts_MissingParams(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/applicable", nil)

	handler.GetApplicableDiscounts(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "GetApplicableDiscounts", mock.Anything, mock.Anything)
}

// TestDiscountsHandler_GetApplicableDiscounts_ServiceError tests service error
func TestDiscountsHandler_GetApplicableDiscounts_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetApplicableDiscountsForOrder", 1000.0, []string{"PROD00000001"}, []string(nil), "WHSE00000001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/applicable?order_value=1000.0&warehouse_id=WHSE00000001&product_ids=PROD00000001", nil)

	handler.GetApplicableDiscounts(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_CalculateOptimalDiscounts_Success tests optimal discount calculation
func TestDiscountsHandler_CalculateOptimalDiscounts_Success(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	request := &models.ValidateDiscountRequest{
		DiscountCode: "SUMMER20",
		OrderValue:   1000.0,
		WarehouseID:  "WARE00000001",
		ProductIDs:   []string{"PROD00000001"},
	}

	expectedDiscounts := []models.DiscountResponse{
		{
			ID:   "DISC00000001",
			Code: "BEST_DEAL",
		},
	}

	mockService.On("CalculateOptimalDiscounts", 1000.0, []string{"PROD00000001"}, []string(nil), "WARE00000001").Return(expectedDiscounts, 100.0, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/discounts/optimal", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CalculateOptimalDiscounts(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_CalculateOptimalDiscounts_InvalidOrderValue tests invalid order value
func TestDiscountsHandler_CalculateOptimalDiscounts_InvalidOrderValue(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	// Empty request - missing required fields
	request := &models.ValidateDiscountRequest{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/discounts/optimal", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CalculateOptimalDiscounts(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "CalculateOptimalDiscounts", mock.Anything, mock.Anything, mock.Anything)
}

// TestDiscountsHandler_CalculateOptimalDiscounts_ServiceError tests service error
func TestDiscountsHandler_CalculateOptimalDiscounts_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	request := &models.ValidateDiscountRequest{
		DiscountCode: "SUMMER20",
		OrderValue:   1000.0,
		WarehouseID:  "WARE00000001",
		ProductIDs:   []string{"PROD00000001"},
	}

	mockService.On("CalculateOptimalDiscounts", 1000.0, []string{"PROD00000001"}, []string(nil), "WARE00000001").Return(nil, 0.0, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/discounts/optimal", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CalculateOptimalDiscounts(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_GetDiscountUsageStats_Success tests fetching discount usage stats
func TestDiscountsHandler_GetDiscountUsageStats_Success(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	expectedStats := map[string]interface{}{
		"total_usage":           50,
		"total_discount_amount": 10000.0,
	}

	mockService.On("GetDiscountUsageStats", "DISC00000001").Return(expectedStats, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "discountID", Value: "DISC00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/DISC00000001/stats", nil)

	handler.GetDiscountUsageStats(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestDiscountsHandler_GetDiscountUsageStats_ServiceError tests service error
func TestDiscountsHandler_GetDiscountUsageStats_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockDiscountsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewDiscountsHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetDiscountUsageStats", "DISC00000001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "discountID", Value: "DISC00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/discounts/DISC00000001/stats", nil)

	handler.GetDiscountUsageStats(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
