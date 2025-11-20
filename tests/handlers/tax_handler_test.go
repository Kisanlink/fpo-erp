package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

// TestTaxHandler_CreateTax_Success tests successful tax creation
func TestTaxHandler_CreateTax_Success(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	validFrom := time.Now()
	request := &models.CreateTaxRequest{
		Code:            "CGST9",
		Name:            "CGST",
		TaxType:         "cgst",
		CalculationType: "percentage",
		Rate:            9.0,
		ValidFrom:       validFrom,
	}

	expectedResponse := &models.TaxResponse{
		ID:              "TAX000000001",
		Name:            "CGST",
		TaxType:         "cgst",
		CalculationType: "percentage",
		Rate:            9.0,
		Status:          "active",
	}

	mockService.On("CreateTax", mock.AnythingOfType("*models.CreateTaxRequest"), "test-user-123").Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/taxes", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "test-user-123")

	handler.CreateTax(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Tax created successfully", response["message"])
	assert.NotNil(t, response["data"])
}

// TestTaxHandler_CreateTax_ValidationError tests validation errors
func TestTaxHandler_CreateTax_ValidationError(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	// Missing required fields
	request := &models.CreateTaxRequest{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/taxes", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "test-user-123")

	handler.CreateTax(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "CreateTax", mock.Anything, mock.Anything)
}

// TestTaxHandler_CreateTax_ServiceError tests service layer errors
func TestTaxHandler_CreateTax_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	validFrom := time.Now()
	request := &models.CreateTaxRequest{
		Code:            "CGST9",
		Name:            "CGST",
		TaxType:         "cgst",
		CalculationType: "percentage",
		Rate:            9.0,
		ValidFrom:       validFrom,
	}

	mockService.On("CreateTax", mock.AnythingOfType("*models.CreateTaxRequest"), "test-user-123").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/taxes", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "test-user-123")

	handler.CreateTax(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaxHandler_GetTax_Success tests fetching a single tax
func TestTaxHandler_GetTax_Success(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	expectedTax := &models.TaxResponse{
		ID:              "TAX000000001",
		Name:            "CGST",
		TaxType:         "cgst",
		CalculationType: "percentage",
		Rate:            9.0,
		Status:          "active",
	}

	mockService.On("GetTax", "TAX000000001").Return(expectedTax, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "TAX000000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/TAX000000001", nil)

	handler.GetTax(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Tax retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])
}

// TestTaxHandler_GetTax_NotFound tests tax not found
func TestTaxHandler_GetTax_NotFound(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetTax", "TAX999999999").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "TAX999999999"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/TAX999999999", nil)

	handler.GetTax(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaxHandler_GetAllTaxes_Success tests fetching all taxes
func TestTaxHandler_GetAllTaxes_Success(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	expectedTaxes := []models.TaxResponse{
		{
			ID:              "TAX000000001",
			Name:            "CGST",
			TaxType:         "cgst",
			CalculationType: "percentage",
			Rate:            9.0,
		},
		{
			ID:              "TAX000000002",
			Name:            "SGST",
			TaxType:         "sgst",
			CalculationType: "percentage",
			Rate:            9.0,
		},
	}

	mockService.On("GetAllTaxes", 10, 0).Return(expectedTaxes, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes", nil)

	handler.GetAllTaxes(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Taxes retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])
}

// TestTaxHandler_GetAllTaxes_WithPagination tests pagination
func TestTaxHandler_GetAllTaxes_WithPagination(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetAllTaxes", 20, 10).Return([]models.TaxResponse{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes?limit=20&offset=10", nil)

	handler.GetAllTaxes(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaxHandler_UpdateTax_Success tests successful tax update
func TestTaxHandler_UpdateTax_Success(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	rate := 10.0
	request := &models.UpdateTaxRequest{
		Rate: &rate,
	}

	expectedResponse := &models.TaxResponse{
		ID:              "TAX000000001",
		Name:            "CGST",
		TaxType:         "cgst",
		CalculationType: "percentage",
		Rate:            10.0,
	}

	mockService.On("UpdateTax", "TAX000000001", mock.AnythingOfType("*models.UpdateTaxRequest"), "test-user-123").Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "TAX000000001"}}
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("PUT", "/api/v1/taxes/TAX000000001", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "test-user-123")

	handler.UpdateTax(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Tax updated successfully", response["message"])
}

// TestTaxHandler_UpdateTax_ServiceError tests update service error
func TestTaxHandler_UpdateTax_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	rate := 10.0
	request := &models.UpdateTaxRequest{
		Rate: &rate,
	}

	mockService.On("UpdateTax", "TAX999999999", mock.AnythingOfType("*models.UpdateTaxRequest"), "test-user-123").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "TAX999999999"}}
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("PUT", "/api/v1/taxes/TAX999999999", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "test-user-123")

	handler.UpdateTax(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaxHandler_DeleteTax_Success tests successful tax deletion
func TestTaxHandler_DeleteTax_Success(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	mockService.On("DeleteTax", "TAX000000001").Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "TAX000000001"}}
	c.Request, _ = http.NewRequest("DELETE", "/api/v1/taxes/TAX000000001", nil)

	handler.DeleteTax(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Tax deleted successfully", response["message"])
}

// TestTaxHandler_DeleteTax_ServiceError tests deletion service error
func TestTaxHandler_DeleteTax_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	mockService.On("DeleteTax", "TAX999999999").Return(assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "TAX999999999"}}
	c.Request, _ = http.NewRequest("DELETE", "/api/v1/taxes/TAX999999999", nil)

	handler.DeleteTax(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaxHandler_GetActiveTaxes_Success tests fetching active taxes
func TestTaxHandler_GetActiveTaxes_Success(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	expectedTaxes := []models.TaxResponse{
		{ID: "TAX000000001", Name: "CGST", Status: "active"},
		{ID: "TAX000000002", Name: "SGST", Status: "active"},
	}

	mockService.On("GetActiveTaxes").Return(expectedTaxes, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/active", nil)

	handler.GetActiveTaxes(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Active taxes retrieved successfully", response["message"])
}

// TestTaxHandler_GetActiveTaxes_ServiceError tests active taxes service error
func TestTaxHandler_GetActiveTaxes_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetActiveTaxes").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/active", nil)

	handler.GetActiveTaxes(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaxHandler_GetTaxesByType_Success tests fetching taxes by type
func TestTaxHandler_GetTaxesByType_Success(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	expectedTaxes := []models.TaxResponse{
		{ID: "TAX000000001", Name: "CGST", TaxType: "cgst"},
	}

	mockService.On("GetTaxesByType", models.TaxType("cgst")).Return(expectedTaxes, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "type", Value: "cgst"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/type/cgst", nil)

	handler.GetTaxesByType(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Taxes retrieved successfully", response["message"])
}

// TestTaxHandler_GetTaxesByType_InvalidType tests invalid tax type
func TestTaxHandler_GetTaxesByType_InvalidType(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "type", Value: "invalid_type"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/type/invalid_type", nil)

	handler.GetTaxesByType(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "GetTaxesByType", mock.Anything)
}

// TestTaxHandler_GetTaxesByType_ServiceError tests type service error
func TestTaxHandler_GetTaxesByType_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetTaxesByType", models.TaxType("cgst")).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "type", Value: "cgst"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/type/cgst", nil)

	handler.GetTaxesByType(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaxHandler_GetTaxesByStatus_Success tests fetching taxes by status
func TestTaxHandler_GetTaxesByStatus_Success(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	expectedTaxes := []models.TaxResponse{
		{ID: "TAX000000001", Name: "CGST", Status: "active"},
	}

	mockService.On("GetTaxesByStatus", "active").Return(expectedTaxes, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "status", Value: "active"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/status/active", nil)

	handler.GetTaxesByStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Taxes retrieved successfully", response["message"])
}

// TestTaxHandler_GetTaxesByStatus_InvalidStatus tests invalid status
func TestTaxHandler_GetTaxesByStatus_InvalidStatus(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "status", Value: "invalid_status"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/status/invalid_status", nil)

	handler.GetTaxesByStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "GetTaxesByStatus", mock.Anything)
}

// TestTaxHandler_GetTaxesByStatus_ServiceError tests status service error
func TestTaxHandler_GetTaxesByStatus_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetTaxesByStatus", "active").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "status", Value: "active"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/status/active", nil)

	handler.GetTaxesByStatus(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaxHandler_CalculateTax_Success tests tax calculation
func TestTaxHandler_CalculateTax_Success(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	request := &models.TaxCalculationRequest{
		WarehouseID:    "WH000000001",
		WarehouseState: "Karnataka",
		Items: []models.TaxCalculationItem{
			{
				ProductID: "PROD00000001",
				Quantity:  10,
				UnitPrice: 100.0,
				LineTotal: 1000.0,
			},
		},
	}

	expectedResponse := &models.TaxCalculationResponse{
		SubTotal:       1000.0,
		TotalTaxAmount: 180.0,
		GrandTotal:     1180.0,
		TaxBreakdown:   []models.TaxBreakdown{},
	}

	mockService.On("CalculateTax", mock.AnythingOfType("*models.TaxCalculationRequest")).Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/taxes/calculate", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CalculateTax(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Tax calculation completed successfully", response["message"])
}

// TestTaxHandler_CalculateTax_EmptyItems tests calculation with no items
func TestTaxHandler_CalculateTax_EmptyItems(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	request := &models.TaxCalculationRequest{
		Items: []models.TaxCalculationItem{},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/taxes/calculate", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CalculateTax(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "CalculateTax", mock.Anything)
}

// TestTaxHandler_CalculateTax_ServiceError tests calculation service error
func TestTaxHandler_CalculateTax_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	request := &models.TaxCalculationRequest{
		WarehouseID:    "WH000000001",
		WarehouseState: "Karnataka",
		Items: []models.TaxCalculationItem{
			{
				ProductID: "PROD00000001",
				Quantity:  10,
				UnitPrice: 100.0,
				LineTotal: 1000.0,
			},
		},
	}

	mockService.On("CalculateTax", mock.AnythingOfType("*models.TaxCalculationRequest")).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/taxes/calculate", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CalculateTax(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaxHandler_GetTaxApplicationsBySale_Success tests fetching tax applications by sale
func TestTaxHandler_GetTaxApplicationsBySale_Success(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	expectedApplications := []models.TaxApplication{
		{
			TaxID:      "TAX000000001",
			SaleID:     stringPtr("SALE00000001"),
			BaseAmount: 1000.0,
			TaxRate:    9.0,
			TaxAmount:  90.0,
			TaxType:    "cgst",
		},
	}

	mockService.On("GetTaxApplicationsBySale", "SALE00000001").Return(expectedApplications, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "saleID", Value: "SALE00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/applications/sale/SALE00000001", nil)

	handler.GetTaxApplicationsBySale(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Tax applications retrieved successfully", response["message"])
}

// TestTaxHandler_GetTaxApplicationsBySale_ServiceError tests sale applications service error
func TestTaxHandler_GetTaxApplicationsBySale_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetTaxApplicationsBySale", "SALE999999999").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "saleID", Value: "SALE999999999"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/applications/sale/SALE999999999", nil)

	handler.GetTaxApplicationsBySale(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaxHandler_GetTaxApplicationsByReturn_Success tests fetching tax applications by return
func TestTaxHandler_GetTaxApplicationsByReturn_Success(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	expectedApplications := []models.TaxApplication{
		{
			TaxID:      "TAX000000001",
			ReturnID:   stringPtr("RETN00000001"),
			BaseAmount: 500.0,
			TaxRate:    9.0,
			TaxAmount:  45.0,
			TaxType:    "cgst",
		},
	}

	mockService.On("GetTaxApplicationsByReturn", "RETN00000001").Return(expectedApplications, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "returnID", Value: "RETN00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/applications/return/RETN00000001", nil)

	handler.GetTaxApplicationsByReturn(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Tax applications retrieved successfully", response["message"])
}

// TestTaxHandler_GetTaxApplicationsByReturn_ServiceError tests return applications service error
func TestTaxHandler_GetTaxApplicationsByReturn_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetTaxApplicationsByReturn", "RETN999999999").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "returnID", Value: "RETN999999999"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/applications/return/RETN999999999", nil)

	handler.GetTaxApplicationsByReturn(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaxHandler_GetTaxSummaryBySale_Success tests fetching tax summary by sale
func TestTaxHandler_GetTaxSummaryBySale_Success(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	expectedSummary := &models.TaxSummary{
		SaleID:         stringPtr("SALE00000001"),
		CGSTAmount:     90.0,
		SGSTAmount:     90.0,
		TotalTaxAmount: 180.0,
		SubTotal:       1000.0,
		GrandTotal:     1180.0,
	}

	mockService.On("GetTaxSummaryBySale", "SALE00000001").Return(expectedSummary, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "saleID", Value: "SALE00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/summary/sale/SALE00000001", nil)

	handler.GetTaxSummaryBySale(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Tax summary retrieved successfully", response["message"])
}

// TestTaxHandler_GetTaxSummaryBySale_NotFound tests sale summary not found
func TestTaxHandler_GetTaxSummaryBySale_NotFound(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetTaxSummaryBySale", "SALE999999999").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "saleID", Value: "SALE999999999"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/summary/sale/SALE999999999", nil)

	handler.GetTaxSummaryBySale(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaxHandler_GetTaxSummaryByReturn_Success tests fetching tax summary by return
func TestTaxHandler_GetTaxSummaryByReturn_Success(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	expectedSummary := &models.TaxSummary{
		ReturnID:       stringPtr("RETN00000001"),
		CGSTAmount:     45.0,
		SGSTAmount:     45.0,
		TotalTaxAmount: 90.0,
		SubTotal:       500.0,
		GrandTotal:     590.0,
	}

	mockService.On("GetTaxSummaryByReturn", "RETN00000001").Return(expectedSummary, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "returnID", Value: "RETN00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/summary/return/RETN00000001", nil)

	handler.GetTaxSummaryByReturn(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Tax summary retrieved successfully", response["message"])
}

// TestTaxHandler_GetTaxSummaryByReturn_NotFound tests return summary not found
func TestTaxHandler_GetTaxSummaryByReturn_NotFound(t *testing.T) {
	mockService := new(mockServices.MockTaxService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewTaxHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetTaxSummaryByReturn", "RETN999999999").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "returnID", Value: "RETN999999999"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/taxes/summary/return/RETN999999999", nil)

	handler.GetTaxSummaryByReturn(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}
