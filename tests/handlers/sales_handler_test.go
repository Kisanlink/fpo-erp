package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"kisanlink-erp/internal/api/handlers"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/utils"
	mockServices "kisanlink-erp/tests/mocks/services"
	"kisanlink-erp/tests/testutils"
)

// ==================== CreateSale Tests ====================

func TestSalesHandler_CreateSale_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	customerID := "CUST_123"
	applyTaxes := false
	expectedResponse := &models.SaleResponse{
		ID:          "SALE00000001",
		WarehouseID: "WRHS00000001",
		SaleDate:    time.Now().Format(time.RFC3339),
		TotalAmount: 1500.00,
		Status:      "completed",
		CustomerID:  &customerID,
		PaymentMode: "cash",
		SaleType:    "in_store",
		ApplyTaxes:  applyTaxes,
	}
	mockService.On("CreateSale", mock.AnythingOfType("*models.CreateSaleRequest")).
		Return(expectedResponse, nil)

	// Create request
	reqBody := models.CreateSaleRequest{
		WarehouseID: "WRHS00000001",
		CustomerID:  &customerID,
		PaymentMode: "cash",
		SaleType:    "in_store",
		ApplyTaxes:  &applyTaxes,
		Items: []models.CreateSaleItemRequest{
			{
				VariantID: "PVAR00000001",
				Quantity:  10,
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/sales", bytes.NewReader(body))
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

func TestSalesHandler_CreateSale_ValidationError_MissingWarehouseID(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing warehouse_id
	reqBody := models.CreateSaleRequest{
		// Missing WarehouseID (required)
		PaymentMode: "cash",
		SaleType:    "in_store",
		Items: []models.CreateSaleItemRequest{
			{
				VariantID: "PVAR00000001",
				Quantity:  10,
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/sales", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSalesHandler_CreateSale_ValidationError_MissingPaymentMode(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing payment_mode
	reqBody := models.CreateSaleRequest{
		WarehouseID: "WRHS00000001",
		// Missing PaymentMode (required)
		SaleType: "in_store",
		Items: []models.CreateSaleItemRequest{
			{
				VariantID: "PVAR00000001",
				Quantity:  10,
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/sales", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSalesHandler_CreateSale_ValidationError_MissingSaleType(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing sale_type
	reqBody := models.CreateSaleRequest{
		WarehouseID: "WRHS00000001",
		PaymentMode: "cash",
		// Missing SaleType (required)
		Items: []models.CreateSaleItemRequest{
			{
				VariantID: "PVAR00000001",
				Quantity:  10,
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/sales", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSalesHandler_CreateSale_ValidationError_MissingItems(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing items
	reqBody := models.CreateSaleRequest{
		WarehouseID: "WRHS00000001",
		PaymentMode: "cash",
		SaleType:    "in_store",
		// Missing Items (required)
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/sales", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSalesHandler_CreateSale_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("CreateSale", mock.AnythingOfType("*models.CreateSaleRequest")).
		Return(nil, errors.New("insufficient inventory"))

	// Create request
	applyTaxes := false
	reqBody := models.CreateSaleRequest{
		WarehouseID: "WRHS00000001",
		PaymentMode: "upi",
		SaleType:    "delivery",
		ApplyTaxes:  &applyTaxes,
		Items: []models.CreateSaleItemRequest{
			{
				VariantID: "PVAR00000001",
				Quantity:  10,
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/sales", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== GetSale Tests ====================

func TestSalesHandler_GetSale_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	customerID := "CUST_123"
	expectedResponse := &models.SaleResponse{
		ID:          "SALE00000001",
		WarehouseID: "WRHS00000001",
		SaleDate:    time.Now().Format(time.RFC3339),
		TotalAmount: 1500.00,
		Status:      "completed",
		CustomerID:  &customerID,
		PaymentMode: "cash",
		SaleType:    "in_store",
		ApplyTaxes:  false,
	}
	mockService.On("GetSale", "SALE00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sales/SALE00000001", nil)

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

func TestSalesHandler_GetSale_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetSale", "SALE99999999").
		Return(nil, errors.New("sale not found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sales/SALE99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestSalesHandler_GetSale_MissingID(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with empty ID
	req := httptest.NewRequest("GET", "/api/v1/sales/", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert (Will match GetAllSales route instead)
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

// ==================== GetAllSales Tests ====================

func TestSalesHandler_GetAllSales_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.SaleResponse{
		{
			ID:          "SALE00000001",
			WarehouseID: "WRHS00000001",
			SaleDate:    time.Now().Format(time.RFC3339),
			TotalAmount: 1500.00,
			Status:      "completed",
			PaymentMode: "cash",
			SaleType:    "in_store",
			ApplyTaxes:  false,
		},
		{
			ID:          "SALE00000002",
			WarehouseID: "WRHS00000002",
			SaleDate:    time.Now().Format(time.RFC3339),
			TotalAmount: 2000.00,
			Status:      "pending",
			PaymentMode: "upi",
			SaleType:    "delivery",
			ApplyTaxes:  true,
		},
	}
	mockService.On("GetAllSales", 10, 0).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sales?limit=10&offset=0", nil)

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

func TestSalesHandler_GetAllSales_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetAllSales", 10, 0).
		Return([]models.SaleResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sales", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestSalesHandler_GetAllSales_CustomPagination(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetAllSales", 20, 10).
		Return([]models.SaleResponse{}, nil)

	// Create request with custom pagination
	req := httptest.NewRequest("GET", "/api/v1/sales?limit=20&offset=10", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestSalesHandler_GetAllSales_InvalidLimit(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid limit
	req := httptest.NewRequest("GET", "/api/v1/sales?limit=invalid&offset=0", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSalesHandler_GetAllSales_InvalidOffset(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid offset
	req := httptest.NewRequest("GET", "/api/v1/sales?limit=10&offset=invalid", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== UpdateSale Tests ====================

func TestSalesHandler_UpdateSale_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.SaleResponse{
		ID:          "SALE00000001",
		WarehouseID: "WRHS00000001",
		SaleDate:    time.Now().Format(time.RFC3339),
		TotalAmount: 1500.00,
		Status:      "completed",
		PaymentMode: "cash",
		SaleType:    "in_store",
		ApplyTaxes:  false,
	}
	mockService.On("UpdateSale", "SALE00000001", mock.AnythingOfType("*models.UpdateSaleRequest")).
		Return(expectedResponse, nil)

	// Create request
	status := "completed"
	reqBody := models.UpdateSaleRequest{
		Status: &status,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/sales/SALE00000001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestSalesHandler_UpdateSale_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("UpdateSale", "SALE99999999", mock.AnythingOfType("*models.UpdateSaleRequest")).
		Return(nil, errors.New("sale not found"))

	// Create request
	status := "completed"
	reqBody := models.UpdateSaleRequest{
		Status: &status,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/sales/SALE99999999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== DeleteSale Tests ====================

func TestSalesHandler_DeleteSale_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("DeleteSale", "SALE00000001").
		Return(nil)

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/sales/SALE00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestSalesHandler_DeleteSale_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("DeleteSale", "SALE99999999").
		Return(errors.New("sale not found"))

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/sales/SALE99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== GetSalesByDateRange Tests ====================

func TestSalesHandler_GetSalesByDateRange_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	expectedResponse := []models.SaleResponse{
		{
			ID:          "SALE00000001",
			WarehouseID: "WRHS00000001",
			SaleDate:    time.Now().Format(time.RFC3339),
			TotalAmount: 1500.00,
			Status:      "completed",
			PaymentMode: "cash",
			SaleType:    "in_store",
			ApplyTaxes:  false,
		},
	}
	mockService.On("GetSalesByDateRange", startDate, endDate).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sales/date-range?start_date=2024-01-01&end_date=2024-12-31", nil)

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

func TestSalesHandler_GetSalesByDateRange_MissingStartDate(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request without start_date
	req := httptest.NewRequest("GET", "/api/v1/sales/date-range?end_date=2024-12-31", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSalesHandler_GetSalesByDateRange_MissingEndDate(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request without end_date
	req := httptest.NewRequest("GET", "/api/v1/sales/date-range?start_date=2024-01-01", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSalesHandler_GetSalesByDateRange_InvalidStartDate(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid start_date format
	req := httptest.NewRequest("GET", "/api/v1/sales/date-range?start_date=invalid&end_date=2024-12-31", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSalesHandler_GetSalesByDateRange_InvalidEndDate(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid end_date format
	req := httptest.NewRequest("GET", "/api/v1/sales/date-range?start_date=2024-01-01&end_date=invalid", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== GetSalesByStatus Tests ====================

func TestSalesHandler_GetSalesByStatus_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.SaleResponse{
		{
			ID:          "SALE00000001",
			WarehouseID: "WRHS00000001",
			SaleDate:    time.Now().Format(time.RFC3339),
			TotalAmount: 1500.00,
			Status:      "completed",
			PaymentMode: "cash",
			SaleType:    "in_store",
			ApplyTaxes:  false,
		},
	}
	mockService.On("GetSalesByStatus", "completed").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sales/status/completed", nil)

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

func TestSalesHandler_GetSalesByStatus_EmptyResult(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetSalesByStatus", "cancelled").
		Return([]models.SaleResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sales/status/cancelled", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== GetTotalSalesAmount Tests ====================

func TestSalesHandler_GetTotalSalesAmount_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	mockService.On("GetTotalSalesAmount", startDate, endDate).
		Return(50000.00, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sales/total-amount?start_date=2024-01-01&end_date=2024-12-31", nil)

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

func TestSalesHandler_GetTotalSalesAmount_MissingDates(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request without dates
	req := httptest.NewRequest("GET", "/api/v1/sales/total-amount", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSalesHandler_GetTotalSalesAmount_InvalidDateFormat(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid date format
	req := httptest.NewRequest("GET", "/api/v1/sales/total-amount?start_date=01/01/2024&end_date=12/31/2024", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== GetTopSellingProducts Tests ====================

func TestSalesHandler_GetTopSellingProducts_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.TopSellingProductResponse{
		{
			ProductID:   "PROD00000001",
			ProductName: "Organic Rice",
			TotalSold:   1000,
			TotalAmount: 25000.00,
		},
		{
			ProductID:   "PROD00000002",
			ProductName: "Wheat Flour",
			TotalSold:   800,
			TotalAmount: 20000.00,
		},
	}
	mockService.On("GetTopSellingProducts", 10).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sales/top-selling?limit=10", nil)

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

func TestSalesHandler_GetTopSellingProducts_CustomLimit(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetTopSellingProducts", 5).
		Return([]models.TopSellingProductResponse{}, nil)

	// Create request with custom limit
	req := httptest.NewRequest("GET", "/api/v1/sales/top-selling?limit=5", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestSalesHandler_GetTopSellingProducts_InvalidLimit(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid limit
	req := httptest.NewRequest("GET", "/api/v1/sales/top-selling?limit=invalid", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== GetSalesSummary Tests ====================

func TestSalesHandler_GetSalesSummary_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sales/summary", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])
	assert.NotNil(t, response["data"])
}

// ==================== UpdateSaleStatus Tests ====================

func TestSalesHandler_UpdateSaleStatus_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.SaleResponse{
		ID:          "SALE00000001",
		WarehouseID: "WRHS00000001",
		SaleDate:    time.Now().Format(time.RFC3339),
		TotalAmount: 1500.00,
		Status:      "cancelled",
		PaymentMode: "cash",
		SaleType:    "in_store",
		ApplyTaxes:  false,
	}
	mockService.On("UpdateSale", "SALE00000001", mock.AnythingOfType("*models.UpdateSaleRequest")).
		Return(expectedResponse, nil)

	// Create request
	reqBody := map[string]string{
		"status": "cancelled",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/sales/SALE00000001/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestSalesHandler_UpdateSaleStatus_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockSalesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewSalesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request without status
	reqBody := map[string]string{}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/sales/SALE00000001/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
