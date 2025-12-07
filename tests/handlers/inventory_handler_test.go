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

// ============================================================================
// CreateBatch Tests
// ============================================================================

func TestInventoryHandler_CreateBatch_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	mockAAA := testutils.NewMockAAAMiddleware()
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, mockAAA, mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.InventoryBatchResponse{
		ID:            "BATC00000001",
		WarehouseID:   "WHSE00000001",
		VariantID:     "PVAR00000001",
		CostPrice:     100.50,
		ExpiryDate:    "2025-12-31",
		TotalQuantity: 500,
	}
	mockService.On("CreateBatch",
		"WHSE00000001", "PVAR00000001", 100.50, mock.Anything, int64(500)).
		Return(expectedResponse, nil)

	// Create request
	reqBody := models.CreateInventoryBatchRequest{
		WarehouseID: "WHSE00000001",
		VariantID:   "PVAR00000001",
		CostPrice:   100.50,
		ExpiryDate:  "2025-12-31",
		Quantity:    500,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/batches", bytes.NewReader(body))
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

func TestInventoryHandler_CreateBatch_ValidationError_MissingFields(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing required fields
	reqBody := models.CreateInventoryBatchRequest{
		WarehouseID: "WHSE00000001",
		// Missing VariantID, CostPrice, ExpiryDate, Quantity
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/batches", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInventoryHandler_CreateBatch_InvalidExpiryDateFormat(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid date format
	reqBody := models.CreateInventoryBatchRequest{
		WarehouseID: "WHSE00000001",
		VariantID:   "PVAR00000001",
		CostPrice:   100.50,
		ExpiryDate:  "31-12-2025", // Invalid format
		Quantity:    500,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/batches", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInventoryHandler_CreateBatch_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("CreateBatch",
		"WHSE00000001", "PVAR00000001", 100.50, mock.Anything, int64(500)).
		Return(nil, errors.New("warehouse not found"))

	// Create request
	reqBody := models.CreateInventoryBatchRequest{
		WarehouseID: "WHSE00000001",
		VariantID:   "PVAR00000001",
		CostPrice:   100.50,
		ExpiryDate:  "2025-12-31",
		Quantity:    500,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/batches", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// GetBatch Tests
// ============================================================================

func TestInventoryHandler_GetBatch_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.InventoryBatchResponse{
		ID:            "BATC00000001",
		WarehouseID:   "WHSE00000001",
		VariantID:     "PVAR00000001",
		CostPrice:     100.50,
		ExpiryDate:    "2025-12-31",
		TotalQuantity: 500,
	}
	mockService.On("GetBatch", "BATC00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/batches/BATC00000001", nil)

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

func TestInventoryHandler_GetBatch_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetBatch", "BATC99999999").
		Return(nil, errors.New("batch not found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/batches/BATC99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestInventoryHandler_GetBatch_MissingID(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request without ID
	req := httptest.NewRequest("GET", "/api/v1/batches/", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code) // Router returns 404 for missing route param
}

// ============================================================================
// GetBatchesByWarehouse Tests
// ============================================================================

func TestInventoryHandler_GetBatchesByWarehouse_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.InventoryBatchResponse{
		{
			ID:            "BATC00000001",
			WarehouseID:   "WHSE00000001",
			VariantID:     "PVAR00000001",
			TotalQuantity: 500,
		},
		{
			ID:            "BATC00000002",
			WarehouseID:   "WHSE00000001",
			VariantID:     "PVAR00000002",
			TotalQuantity: 300,
		},
	}
	mockService.On("GetBatchesByWarehouse", "WHSE00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/warehouses/WHSE00000001/batches", nil)

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

func TestInventoryHandler_GetBatchesByWarehouse_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetBatchesByWarehouse", "WHSE00000001").
		Return([]models.InventoryBatchResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/warehouses/WHSE00000001/batches", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestInventoryHandler_GetBatchesByWarehouse_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetBatchesByWarehouse", "WHSE00000001").
		Return(nil, errors.New("database error"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/warehouses/WHSE00000001/batches", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// GetBatchesByVariant Tests
// ============================================================================

func TestInventoryHandler_GetBatchesByVariant_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.InventoryBatchResponse{
		{
			ID:            "BATC00000001",
			WarehouseID:   "WHSE00000001",
			VariantID:     "PVAR00000001",
			TotalQuantity: 500,
		},
		{
			ID:            "BATC00000002",
			WarehouseID:   "WHSE00000002",
			VariantID:     "PVAR00000001",
			TotalQuantity: 300,
		},
	}
	mockService.On("GetBatchesByVariant", "PVAR00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/variants/PVAR00000001/batches", nil)

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

func TestInventoryHandler_GetBatchesByVariant_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetBatchesByVariant", "PVAR00000001").
		Return([]models.InventoryBatchResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/variants/PVAR00000001/batches", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestInventoryHandler_GetBatchesByVariant_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetBatchesByVariant", "PVAR00000001").
		Return(nil, errors.New("database error"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/variants/PVAR00000001/batches", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// CreateTransaction Tests
// ============================================================================

func TestInventoryHandler_CreateTransaction_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	mockAAA := testutils.NewMockAAAMiddleware()
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, mockAAA, mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	relatedEntityID := "SALE00000001"
	note := "Sale transaction"
	expectedResponse := &models.InventoryTransactionResponse{
		ID:              "TRNS00000001",
		BatchID:         "BATC00000001",
		TransactionType: "sale",
		QuantityChange:  -50,
		RelatedEntityID: &relatedEntityID,
		Note:            &note,
	}
	mockService.On("CreateTransaction", "BATC00000001", mock.AnythingOfType("*models.CreateInventoryTransactionRequest")).
		Return(expectedResponse, nil)

	// Create request
	reqBody := models.CreateInventoryTransactionRequest{
		TransactionType: "sale",
		QuantityChange:  -50,
		RelatedEntityID: &relatedEntityID,
		Note:            &note,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/batches/BATC00000001/transactions", bytes.NewReader(body))
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

func TestInventoryHandler_CreateTransaction_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing required fields
	reqBody := models.CreateInventoryTransactionRequest{
		// Missing TransactionType and QuantityChange
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/batches/BATC00000001/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInventoryHandler_CreateTransaction_MissingBatchID(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request without batch ID in URL
	note := "Test transaction"
	reqBody := models.CreateInventoryTransactionRequest{
		TransactionType: "adjustment",
		QuantityChange:  10,
		Note:            &note,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/batches//transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	// Handler validates empty batch ID and returns 400 instead of 404
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInventoryHandler_CreateTransaction_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("CreateTransaction", "BATC00000001", mock.AnythingOfType("*models.CreateInventoryTransactionRequest")).
		Return(nil, errors.New("batch not found"))

	// Create request
	note := "Test transaction"
	reqBody := models.CreateInventoryTransactionRequest{
		TransactionType: "adjustment",
		QuantityChange:  10,
		Note:            &note,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/batches/BATC00000001/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// GetTransactionsByBatch Tests
// ============================================================================

func TestInventoryHandler_GetTransactionsByBatch_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.InventoryTransactionResponse{
		{
			ID:              "TRNS00000001",
			BatchID:         "BATC00000001",
			TransactionType: "purchase",
			QuantityChange:  100,
		},
		{
			ID:              "TRNS00000002",
			BatchID:         "BATC00000001",
			TransactionType: "sale",
			QuantityChange:  -50,
		},
	}
	mockService.On("GetTransactionsByBatch", "BATC00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/batches/BATC00000001/transactions", nil)

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

func TestInventoryHandler_GetTransactionsByBatch_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetTransactionsByBatch", "BATC00000001").
		Return([]models.InventoryTransactionResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/batches/BATC00000001/transactions", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestInventoryHandler_GetTransactionsByBatch_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetTransactionsByBatch", "BATC00000001").
		Return(nil, errors.New("database error"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/batches/BATC00000001/transactions", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// GetExpiringBatches Tests
// ============================================================================

func TestInventoryHandler_GetExpiringBatches_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.InventoryBatchResponse{
		{
			ID:         "BATC00000001",
			ExpiryDate: "2025-02-15",
		},
		{
			ID:         "BATC00000002",
			ExpiryDate: "2025-03-01",
		},
	}
	mockService.On("GetExpiringBatches", 30).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/batches/expiring?days=30", nil)

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

func TestInventoryHandler_GetExpiringBatches_DefaultDays(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations (default 30 days)
	mockService.On("GetExpiringBatches", 30).
		Return([]models.InventoryBatchResponse{}, nil)

	// Create request without days parameter
	req := httptest.NewRequest("GET", "/api/v1/batches/expiring", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestInventoryHandler_GetExpiringBatches_InvalidDays(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid days parameter
	req := httptest.NewRequest("GET", "/api/v1/batches/expiring?days=invalid", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInventoryHandler_GetExpiringBatches_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetExpiringBatches", 30).
		Return(nil, errors.New("database error"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/batches/expiring?days=30", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// GetLowStockBatches Tests
// ============================================================================

func TestInventoryHandler_GetLowStockBatches_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.InventoryBatchResponse{
		{
			ID:            "BATC00000001",
			TotalQuantity: 5,
		},
		{
			ID:            "BATC00000002",
			TotalQuantity: 8,
		},
	}
	mockService.On("GetLowStockBatches", int64(10)).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/batches/low-stock?threshold=10", nil)

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

func TestInventoryHandler_GetLowStockBatches_DefaultThreshold(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations (default threshold 10)
	mockService.On("GetLowStockBatches", int64(10)).
		Return([]models.InventoryBatchResponse{}, nil)

	// Create request without threshold parameter
	req := httptest.NewRequest("GET", "/api/v1/batches/low-stock", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestInventoryHandler_GetLowStockBatches_InvalidThreshold(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid threshold parameter
	req := httptest.NewRequest("GET", "/api/v1/batches/low-stock?threshold=abc", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInventoryHandler_GetLowStockBatches_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetLowStockBatches", int64(10)).
		Return(nil, errors.New("database error"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/batches/low-stock?threshold=10", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// GetAllProductsAvailability Tests
// ============================================================================

func TestInventoryHandler_GetAllProductsAvailability_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.ProductAvailabilityResponse{
		{
			ID:            "BATC00000001",
			WarehouseID:   "WHSE00000001",
			WarehouseName: "Main Warehouse",
			VariantID:     "PVAR00000001",
			ProductSKU:    "PROD-SKU-001",
			ProductName:   "Rice 1kg",
			TotalQuantity: 500,
		},
		{
			ID:            "BATC00000002",
			WarehouseID:   "WHSE00000002",
			WarehouseName: "Branch Warehouse",
			VariantID:     "PVAR00000002",
			ProductSKU:    "PROD-SKU-002",
			ProductName:   "Wheat 1kg",
			TotalQuantity: 300,
		},
	}
	mockService.On("GetAllProductsAvailability", mock.Anything, mock.AnythingOfType("string")).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/availability", nil)

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

func TestInventoryHandler_GetAllProductsAvailability_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetAllProductsAvailability", mock.Anything, mock.AnythingOfType("string")).
		Return([]models.ProductAvailabilityResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/availability", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestInventoryHandler_GetAllProductsAvailability_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockInventoryService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewInventoryHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetAllProductsAvailability", mock.Anything, mock.AnythingOfType("string")).
		Return(nil, errors.New("AAA service error"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/availability", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
