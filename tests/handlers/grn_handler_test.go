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

func TestGRNHandler_CreateGRN_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.GRNResponse{
		ID:            "GRNX00000001",
		GRNNumber:     "GRN-2025-0001",
		POID:          "PORD00000001",
		PONumber:      "PO-2025-001",
		WarehouseID:   "WHSE00000001",
		WarehouseName: "Main Warehouse",
		ReceivedBy:    "test-user-123",
		QualityStatus: "accepted",
	}
	mockService.On("CreateGRN", mock.Anything, mock.AnythingOfType("*models.CreateGRNRequest")).
		Return(expectedResponse, nil)

	// Create request
	reqBody := models.CreateGRNRequest{
		GRNNumber:     "GRN-2025-0001",
		POID:          "PORD00000001",
		ReceivedBy:    "test-user-123",
		QualityStatus: "accepted",
		Items: []models.CreateGRNItemRequest{
			{
				POItemID:         "POIM00000001",
				ReceivedQuantity: 100,
				AcceptedQuantity: 95,
				RejectedQuantity: 5,
				ExpiryDate:       "2025-12-31",
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/grns", bytes.NewReader(body))
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

func TestGRNHandler_CreateGRN_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing required fields
	reqBody := models.CreateGRNRequest{
		GRNNumber: "GRN-2025-0001",
		// Missing POID, ReceivedBy, QualityStatus, Items
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/grns", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGRNHandler_CreateGRN_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("CreateGRN", mock.Anything, mock.AnythingOfType("*models.CreateGRNRequest")).
		Return(nil, errors.New("database error"))

	// Create request
	reqBody := models.CreateGRNRequest{
		GRNNumber:     "GRN-2025-0001",
		POID:          "PORD00000001",
		ReceivedBy:    "test-user-123",
		QualityStatus: "accepted",
		Items: []models.CreateGRNItemRequest{
			{
				POItemID:         "POIM00000001",
				ReceivedQuantity: 100,
				AcceptedQuantity: 95,
				ExpiryDate:       "2025-12-31",
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/grns", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestGRNHandler_GetGRN_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.GRNResponse{
		ID:            "GRNX00000001",
		GRNNumber:     "GRN-2025-0001",
		POID:          "PORD00000001",
		PONumber:      "PO-2025-001",
		WarehouseID:   "WHSE00000001",
		WarehouseName: "Main Warehouse",
		ReceivedBy:    "test-user-123",
		QualityStatus: "accepted",
	}
	mockService.On("GetGRN", mock.Anything, "GRNX00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/grns/GRNX00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGRNHandler_GetGRN_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetGRN", mock.Anything, "GRNX99999999").
		Return(nil, errors.New("grn not found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/grns/GRNX99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestGRNHandler_GetAllGRNs_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.GRNResponse{
		{
			ID:            "GRNX00000001",
			GRNNumber:     "GRN-2025-0001",
			POID:          "PORD00000001",
			PONumber:      "PO-2025-001",
			WarehouseID:   "WHSE00000001",
			WarehouseName: "Main Warehouse",
			QualityStatus: "accepted",
		},
		{
			ID:            "GRNX00000002",
			GRNNumber:     "GRN-2025-0002",
			POID:          "PORD00000002",
			PONumber:      "PO-2025-002",
			WarehouseID:   "WHSE00000001",
			WarehouseName: "Main Warehouse",
			QualityStatus: "partial",
		},
	}
	mockService.On("GetAllGRNs", mock.Anything, 50, 0).
		Return(expectedResponse, int64(2), nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/grns", nil)

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
	assert.NotNil(t, response["pagination"])
}

func TestGRNHandler_GetAllGRNs_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetAllGRNs", mock.Anything, 50, 0).
		Return([]models.GRNResponse{}, int64(0), nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/grns", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGRNHandler_GetGRNsByWarehouse_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.GRNResponse{
		{
			ID:            "GRNX00000001",
			GRNNumber:     "GRN-2025-0001",
			POID:          "PORD00000001",
			WarehouseID:   "WHSE00000001",
			WarehouseName: "Main Warehouse",
			QualityStatus: "accepted",
		},
	}
	mockService.On("GetGRNsByWarehouse", mock.Anything, "WHSE00000001", 50, 0).
		Return(expectedResponse, int64(1), nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/warehouses/WHSE00000001/grns", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGRNHandler_GetGRNsByWarehouse_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetGRNsByWarehouse", mock.Anything, "WHSE00000002", 50, 0).
		Return([]models.GRNResponse{}, int64(0), nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/warehouses/WHSE00000002/grns", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGRNHandler_GetGRNsByWarehouse_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetGRNsByWarehouse", mock.Anything, "WHSE99999999", 50, 0).
		Return(nil, int64(0), errors.New("warehouse not found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/warehouses/WHSE99999999/grns", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestGRNHandler_GetGRNByPurchaseOrder_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.GRNResponse{
		ID:            "GRNX00000001",
		GRNNumber:     "GRN-2025-0001",
		POID:          "PORD00000001",
		PONumber:      "PO-2025-001",
		WarehouseID:   "WHSE00000001",
		WarehouseName: "Main Warehouse",
		QualityStatus: "accepted",
	}
	mockService.On("GetGRNByPurchaseOrder", mock.Anything, "PORD00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/purchase-orders/PORD00000001/grn", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGRNHandler_GetGRNByPurchaseOrder_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetGRNByPurchaseOrder", mock.Anything, "PORD99999999").
		Return(nil, errors.New("grn not found for this purchase order"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/purchase-orders/PORD99999999/grn", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestGRNHandler_UpdateGRN_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	grnDocument := "ATT_xyz789"
	remarks := "Updated remarks"
	expectedResponse := &models.GRNResponse{
		ID:            "GRNX00000001",
		GRNNumber:     "GRN-2025-0001",
		GRNDocument:   &grnDocument,
		POID:          "PORD00000001",
		QualityStatus: "accepted",
		Remarks:       &remarks,
	}
	mockService.On("UpdateGRN", mock.Anything, "GRNX00000001", mock.AnythingOfType("*models.UpdateGRNRequest")).
		Return(expectedResponse, nil)

	// Create request
	reqBody := models.UpdateGRNRequest{
		GRNDocument: &grnDocument,
		Remarks:     &remarks,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/grns/GRNX00000001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGRNHandler_UpdateGRN_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("UpdateGRN", mock.Anything, "GRNX99999999", mock.AnythingOfType("*models.UpdateGRNRequest")).
		Return(nil, errors.New("grn not found"))

	// Create request
	remarks := "Updated remarks"
	reqBody := models.UpdateGRNRequest{
		Remarks: &remarks,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/grns/GRNX99999999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestGRNHandler_UpdateGRN_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid JSON
	req := httptest.NewRequest("PUT", "/api/v1/grns/GRNX00000001", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGRNHandler_CreateGRN_EmptyItems(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with empty items array (should fail validation)
	reqBody := models.CreateGRNRequest{
		GRNNumber:     "GRN-2025-0001",
		POID:          "PORD00000001",
		ReceivedBy:    "test-user-123",
		QualityStatus: "accepted",
		Items:         []models.CreateGRNItemRequest{},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/grns", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGRNHandler_GetGRNsByWarehouse_MissingID(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations - handler should check if ID is empty
	// This tests the handler's parameter validation
	mockService.On("GetGRNsByWarehouse", mock.Anything, "").
		Return([]models.GRNResponse{}, nil)

	// Create request with empty warehouse ID
	req := httptest.NewRequest("GET", "/api/v1/warehouses//grns", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - handler checks for empty ID and returns BadRequest
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGRNHandler_GetGRNByPurchaseOrder_MissingID(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockGRNService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewGRNHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetGRNByPurchaseOrder", mock.Anything, "").
		Return(nil, errors.New("purchase order id required"))

	// Create request with empty PO ID
	req := httptest.NewRequest("GET", "/api/v1/purchase-orders//grn", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - handler checks for empty ID and returns BadRequest
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
