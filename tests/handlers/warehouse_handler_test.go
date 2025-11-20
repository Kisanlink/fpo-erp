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

func TestWarehouseHandler_CreateWarehouse_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockWarehouseService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewWarehouseHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.WarehouseResponse{
		ID:   "WARE00000001",
		Name: "Main Warehouse",
	}
	mockService.On("CreateWarehouse", mock.Anything, mock.AnythingOfType("*models.CreateWarehouseRequest"), mock.Anything, mock.Anything).
		Return(expectedResponse, nil)

	// Create request
	addressID := "ADDR_12345678"
	reqBody := models.CreateWarehouseRequest{
		Name:      "Main Warehouse",
		AddressID: &addressID,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/warehouses", bytes.NewReader(body))
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

func TestWarehouseHandler_CreateWarehouse_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockWarehouseService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewWarehouseHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing required fields
	reqBody := models.CreateWarehouseRequest{
		// Missing Name
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/warehouses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWarehouseHandler_CreateWarehouse_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockWarehouseService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewWarehouseHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("CreateWarehouse", mock.Anything, mock.AnythingOfType("*models.CreateWarehouseRequest"), mock.Anything, mock.Anything).
		Return(nil, errors.New("database error"))

	// Create request
	addressID := "ADDR_12345678"
	reqBody := models.CreateWarehouseRequest{
		Name:      "Main Warehouse",
		AddressID: &addressID,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/warehouses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestWarehouseHandler_GetAllWarehouses_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockWarehouseService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewWarehouseHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.WarehouseResponse{
		{
			ID:   "WARE00000001",
			Name: "Main Warehouse",
		},
		{
			ID:   "WARE00000002",
			Name: "Secondary Warehouse",
		},
	}
	mockService.On("GetAllWarehouses", mock.Anything, mock.Anything).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/warehouses", nil)

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

func TestWarehouseHandler_GetAllWarehouses_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockWarehouseService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewWarehouseHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetAllWarehouses", mock.Anything, mock.Anything).
		Return([]models.WarehouseResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/warehouses", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestWarehouseHandler_SearchWarehouses_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockWarehouseService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewWarehouseHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.WarehouseResponse{
		{
			ID:   "WARE00000001",
			Name: "Main Warehouse",
		},
	}
	mockService.On("SearchWarehouses", mock.Anything, "Main", mock.Anything).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/warehouses/search?q=Main", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestWarehouseHandler_SearchWarehouses_NoResults(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockWarehouseService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewWarehouseHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("SearchWarehouses", mock.Anything, "NonExistent", mock.Anything).
		Return([]models.WarehouseResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/warehouses/search?q=NonExistent", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestWarehouseHandler_GetWarehouse_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockWarehouseService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewWarehouseHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.WarehouseResponse{
		ID:   "WARE00000001",
		Name: "Main Warehouse",
	}
	mockService.On("GetWarehouse", mock.Anything, "WARE00000001", mock.Anything).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/warehouses/WARE00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestWarehouseHandler_GetWarehouse_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockWarehouseService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewWarehouseHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetWarehouse", mock.Anything, "WARE99999999", mock.Anything).
		Return(nil, errors.New("warehouse not found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/warehouses/WARE99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestWarehouseHandler_UpdateWarehouse_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockWarehouseService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewWarehouseHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.WarehouseResponse{
		ID:   "WARE00000001",
		Name: "Updated Warehouse",
	}
	mockService.On("UpdateWarehouse", mock.Anything, "WARE00000001", mock.AnythingOfType("*models.UpdateWarehouseRequest"), mock.Anything).
		Return(expectedResponse, nil)

	// Create request
	name := "Updated Warehouse"
	reqBody := models.UpdateWarehouseRequest{
		Name: &name,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/warehouses/WARE00000001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestWarehouseHandler_UpdateWarehouse_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockWarehouseService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewWarehouseHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("UpdateWarehouse", mock.Anything, "WARE99999999", mock.AnythingOfType("*models.UpdateWarehouseRequest"), mock.Anything).
		Return(nil, errors.New("warehouse not found"))

	// Create request
	name := "Updated Warehouse"
	reqBody := models.UpdateWarehouseRequest{
		Name: &name,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/warehouses/WARE99999999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestWarehouseHandler_UpdateWarehouse_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockWarehouseService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewWarehouseHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid JSON
	req := httptest.NewRequest("PATCH", "/api/v1/warehouses/WARE00000001", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWarehouseHandler_DeleteWarehouse_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockWarehouseService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewWarehouseHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("DeleteWarehouse", mock.Anything, "WARE00000001", mock.Anything).
		Return(nil)

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/warehouses/WARE00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestWarehouseHandler_DeleteWarehouse_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockWarehouseService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewWarehouseHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("DeleteWarehouse", mock.Anything, "WARE99999999", mock.Anything).
		Return(errors.New("warehouse not found"))

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/warehouses/WARE99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}
