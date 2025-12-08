package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"kisanlink-erp/internal/api/handlers"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/utils"
	mockServices "kisanlink-erp/tests/mocks/services"
	"kisanlink-erp/tests/testutils"
)

// ====================================
// CreateReturn Tests
// ====================================

func TestReturnsHandler_CreateReturn_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.ReturnResponse{
		ID:          "RETN00000001",
		SaleID:      "SALE00000001",
		ReturnDate:  "2024-01-15T10:00:00Z",
		TotalRefund: 150.00,
		Status:      "pending",
	}
	mockService.On("CreateReturn", mock.AnythingOfType("*models.CreateReturnRequest")).
		Return(expectedResponse, nil)

	// Create request
	reqBody := models.CreateReturnRequest{
		SaleID: "SALE00000001",
		Items: []models.CreateReturnItemRequest{
			{
				BatchID:      "BATC00000001",
				Quantity:     2,
				RefundAmount: 150.00,
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/returns", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])
	assert.NotNil(t, response["data"])
}

func TestReturnsHandler_CreateReturn_ValidationError_MissingSaleID(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing sale_id
	reqBody := models.CreateReturnRequest{
		// Missing SaleID (required field)
		Items: []models.CreateReturnItemRequest{
			{
				BatchID:      "BATC00000001",
				Quantity:     2,
				RefundAmount: 150.00,
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/returns", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReturnsHandler_CreateReturn_ValidationError_MissingItems(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing items
	reqBody := models.CreateReturnRequest{
		SaleID: "SALE00000001",
		// Missing Items (required field)
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/returns", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReturnsHandler_CreateReturn_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("CreateReturn", mock.AnythingOfType("*models.CreateReturnRequest")).
		Return(nil, errors.New("failed to create return"))

	// Create request
	reqBody := models.CreateReturnRequest{
		SaleID: "SALE00000001",
		Items: []models.CreateReturnItemRequest{
			{
				BatchID:      "BATC00000001",
				Quantity:     2,
				RefundAmount: 150.00,
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/returns", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// ====================================
// GetReturn Tests
// ====================================

func TestReturnsHandler_GetReturn_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.ReturnResponse{
		ID:          "RETN00000001",
		SaleID:      "SALE00000001",
		ReturnDate:  "2024-01-15T10:00:00Z",
		TotalRefund: 150.00,
		Status:      "pending",
	}
	mockService.On("GetReturn", "RETN00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/returns/RETN00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])
	assert.NotNil(t, response["data"])
}

func TestReturnsHandler_GetReturn_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetReturn", "RETN99999999").
		Return(nil, errors.New("return not found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/returns/RETN99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// ====================================
// GetAllReturns Tests
// ====================================

func TestReturnsHandler_GetAllReturns_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.ReturnResponse{
		{
			ID:          "RETN00000001",
			SaleID:      "SALE00000001",
			TotalRefund: 150.00,
			Status:      "pending",
		},
		{
			ID:          "RETN00000002",
			SaleID:      "SALE00000002",
			TotalRefund: 200.00,
			Status:      "processed",
		},
	}
	mockService.On("GetAllReturns", 50, 0).
		Return(expectedResponse, int64(2), nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/returns", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])
	assert.NotNil(t, response["data"])
}

func TestReturnsHandler_GetAllReturns_WithPagination(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.ReturnResponse{
		{
			ID:          "RETN00000003",
			SaleID:      "SALE00000003",
			TotalRefund: 100.00,
			Status:      "completed",
		},
	}
	mockService.On("GetAllReturns", 5, 10).
		Return(expectedResponse, int64(1), nil)

	// Create request with pagination
	req := httptest.NewRequest("GET", "/api/v1/returns?limit=5&offset=10", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestReturnsHandler_GetAllReturns_InvalidLimit(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations (in case handler still calls service despite invalid limit)
	mockService.On("GetAllReturns", 50, 0).
		Return([]models.ReturnResponse{}, int64(0), nil).Maybe()

	// Create request with invalid limit
	req := httptest.NewRequest("GET", "/api/v1/returns?limit=invalid", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - handler ignores invalid limit and uses default
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReturnsHandler_GetAllReturns_InvalidOffset(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations (in case handler still calls service despite invalid offset)
	mockService.On("GetAllReturns", 50, 0).
		Return([]models.ReturnResponse{}, int64(0), nil).Maybe()

	// Create request with invalid offset
	req := httptest.NewRequest("GET", "/api/v1/returns?offset=invalid", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - handler ignores invalid offset and uses default
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReturnsHandler_GetAllReturns_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetAllReturns", 50, 0).
		Return([]models.ReturnResponse{}, int64(0), nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/returns", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// ====================================
// UpdateReturn Tests
// ====================================

func TestReturnsHandler_UpdateReturn_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.ReturnResponse{
		ID:          "RETN00000001",
		SaleID:      "SALE00000001",
		TotalRefund: 150.00,
		Status:      "processed",
	}
	mockService.On("UpdateReturn", "RETN00000001", mock.AnythingOfType("*models.UpdateReturnRequest")).
		Return(expectedResponse, nil)

	// Create request
	status := "processed"
	reqBody := models.UpdateReturnRequest{
		Status: &status,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/returns/RETN00000001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])
}

func TestReturnsHandler_UpdateReturn_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("UpdateReturn", "RETN99999999", mock.AnythingOfType("*models.UpdateReturnRequest")).
		Return(nil, errors.New("return not found"))

	// Create request
	status := "processed"
	reqBody := models.UpdateReturnRequest{
		Status: &status,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/returns/RETN99999999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// ====================================
// DeleteReturn Tests
// ====================================

func TestReturnsHandler_DeleteReturn_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("DeleteReturn", "RETN00000001").
		Return(nil)

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/returns/RETN00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestReturnsHandler_DeleteReturn_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("DeleteReturn", "RETN99999999").
		Return(errors.New("return not found"))

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/returns/RETN99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// ====================================
// GetReturnsByDateRange Tests
// ====================================

func TestReturnsHandler_GetReturnsByDateRange_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.ReturnResponse{
		{
			ID:          "RETN00000001",
			SaleID:      "SALE00000001",
			TotalRefund: 150.00,
			Status:      "processed",
		},
	}
	mockService.On("GetReturnsByDateRange", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/returns/date-range?start_date=2024-01-01&end_date=2024-01-31", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])
}

func TestReturnsHandler_GetReturnsByDateRange_MissingStartDate(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing start_date
	req := httptest.NewRequest("GET", "/api/v1/returns/date-range?end_date=2024-01-31", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReturnsHandler_GetReturnsByDateRange_MissingEndDate(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing end_date
	req := httptest.NewRequest("GET", "/api/v1/returns/date-range?start_date=2024-01-01", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReturnsHandler_GetReturnsByDateRange_InvalidDateFormat(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid date format
	req := httptest.NewRequest("GET", "/api/v1/returns/date-range?start_date=01/01/2024&end_date=2024-01-31", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ====================================
// GetReturnsByStatus Tests
// ====================================

func TestReturnsHandler_GetReturnsByStatus_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.ReturnResponse{
		{
			ID:          "RETN00000001",
			SaleID:      "SALE00000001",
			TotalRefund: 150.00,
			Status:      "pending",
		},
	}
	mockService.On("GetReturnsByStatus", "pending").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/returns/status/pending", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])
}

func TestReturnsHandler_GetReturnsByStatus_EmptyResult(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetReturnsByStatus", "completed").
		Return([]models.ReturnResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/returns/status/completed", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// ====================================
// GetTotalReturnsAmount Tests
// ====================================

func TestReturnsHandler_GetTotalReturnsAmount_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetTotalReturnsAmount", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(1500.50, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/returns/total-amount?start_date=2024-01-01&end_date=2024-01-31", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])
	assert.NotNil(t, response["data"])
}

func TestReturnsHandler_GetTotalReturnsAmount_MissingDates(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing dates
	req := httptest.NewRequest("GET", "/api/v1/returns/total-amount", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReturnsHandler_GetTotalReturnsAmount_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetTotalReturnsAmount", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(0.0, errors.New("database error"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/returns/total-amount?start_date=2024-01-01&end_date=2024-01-31", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// ====================================
// GetMostReturnedProducts Tests
// ====================================

func TestReturnsHandler_GetMostReturnedProducts_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.MostReturnedProductResponse{
		{
			ProductID:     "PROD00000001",
			ProductName:   "Rice",
			TotalReturned: 50,
			ReturnAmount:  500.00,
		},
		{
			ProductID:     "PROD00000002",
			ProductName:   "Wheat",
			TotalReturned: 30,
			ReturnAmount:  300.00,
		},
	}
	mockService.On("GetMostReturnedProducts", 10).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/returns/most-returned", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])
	assert.NotNil(t, response["data"])
}

func TestReturnsHandler_GetMostReturnedProducts_WithLimit(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.MostReturnedProductResponse{
		{
			ProductID:     "PROD00000001",
			ProductName:   "Rice",
			TotalReturned: 50,
			ReturnAmount:  500.00,
		},
	}
	mockService.On("GetMostReturnedProducts", 5).
		Return(expectedResponse, nil)

	// Create request with custom limit
	req := httptest.NewRequest("GET", "/api/v1/returns/most-returned?limit=5", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestReturnsHandler_GetMostReturnedProducts_InvalidLimit(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid limit
	req := httptest.NewRequest("GET", "/api/v1/returns/most-returned?limit=invalid", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReturnsHandler_GetMostReturnedProducts_EmptyResult(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockReturnsService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewReturnsHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetMostReturnedProducts", 10).
		Return([]models.MostReturnedProductResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/returns/most-returned", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
