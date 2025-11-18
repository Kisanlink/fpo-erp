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
	mockServices "kisanlink-erp/tests/mocks/services"
	"kisanlink-erp/tests/testutils"
)

func TestProductVariantHandler_GetProductVariant_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductVariantService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductVariantHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	sku := "RICE-1KG-001"
	expectedResponse := &models.ProductVariantResponse{
		ID:          "PVAR00000001",
		ProductID:   "PROD00000001",
		VariantName: "1kg Pack",
		SKU:         &sku,
		PackSize:    "1kg",
		IsActive:    true,
	}
	mockService.On("GetProductVariant", mock.Anything, "PVAR00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/variants/PVAR00000001", nil)

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

func TestProductVariantHandler_GetProductVariant_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductVariantService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductVariantHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetProductVariant", mock.Anything, "PVAR99999999").
		Return(nil, errors.New("variant not found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/variants/PVAR99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductVariantHandler_UpdateProductVariant_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductVariantService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductVariantHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	sku := "RICE-1KG-001"
	expectedResponse := &models.ProductVariantResponse{
		ID:          "PVAR00000001",
		ProductID:   "PROD00000001",
		VariantName: "Updated 1kg Pack",
		SKU:         &sku,
		IsActive:    true,
	}
	mockService.On("UpdateProductVariant", mock.Anything, "PVAR00000001", mock.AnythingOfType("*models.UpdateProductVariantRequest")).
		Return(expectedResponse, nil)

	// Create request
	variantName := "Updated 1kg Pack"
	packSize := "1000g"
	reqBody := models.UpdateProductVariantRequest{
		VariantName: &variantName,
		PackSize:    &packSize,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/variants/PVAR00000001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductVariantHandler_UpdateProductVariant_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductVariantService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductVariantHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("UpdateProductVariant", mock.Anything, "PVAR99999999", mock.AnythingOfType("*models.UpdateProductVariantRequest")).
		Return(nil, errors.New("variant not found"))

	// Create request
	variantName := "Updated Pack"
	reqBody := models.UpdateProductVariantRequest{
		VariantName: &variantName,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/variants/PVAR99999999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductVariantHandler_DeleteProductVariant_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductVariantService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductVariantHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("DeleteProductVariant", mock.Anything, "PVAR00000001").
		Return(nil)

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/variants/PVAR00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductVariantHandler_DeleteProductVariant_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductVariantService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductVariantHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("DeleteProductVariant", mock.Anything, "PVAR99999999").
		Return(errors.New("variant not found"))

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/variants/PVAR99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductVariantHandler_GetVariantBySKU_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductVariantService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductVariantHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	sku := "RICE-1KG-001"
	expectedResponse := &models.ProductVariantResponse{
		ID:          "PVAR00000001",
		ProductID:   "PROD00000001",
		VariantName: "1kg Pack",
		SKU:         &sku,
		IsActive:    true,
	}
	mockService.On("GetVariantBySKU", mock.Anything, "RICE-1KG-001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/variants/sku/RICE-1KG-001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductVariantHandler_GetVariantBySKU_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductVariantService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductVariantHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetVariantBySKU", mock.Anything, "NONEXISTENT-SKU").
		Return(nil, errors.New("variant not found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/variants/sku/NONEXISTENT-SKU", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductVariantHandler_GetVariantByBarcode_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductVariantService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductVariantHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	sku := "RICE-1KG-001"
	barcode := "1234567890123"
	expectedResponse := &models.ProductVariantResponse{
		ID:          "PVAR00000001",
		ProductID:   "PROD00000001",
		VariantName: "1kg Pack",
		SKU:         &sku,
		Barcode:     &barcode,
		IsActive:    true,
	}
	mockService.On("GetVariantByBarcode", mock.Anything, "1234567890123").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/variants/barcode/1234567890123", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductVariantHandler_GetVariantByBarcode_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductVariantService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductVariantHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetVariantByBarcode", mock.Anything, "0000000000000").
		Return(nil, errors.New("variant not found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/variants/barcode/0000000000000", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductVariantHandler_CreateProductVariant_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductVariantService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductVariantHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	sku := "RICE-1KG-001"
	expectedResponse := &models.ProductVariantResponse{
		ID:          "PVAR00000001",
		ProductID:   "PROD00000001",
		VariantName: "1kg Pack",
		SKU:         &sku,
		PackSize:    "1kg",
		IsActive:    true,
	}
	mockService.On("CreateProductVariant", mock.Anything, mock.Anything, mock.AnythingOfType("*models.CreateProductVariantRequest")).
		Return(expectedResponse, nil)

	// Create request
	reqSKU := "RICE-1KG-001"
	reqBody := models.CreateProductVariantRequest{
		VariantName: "1kg Pack",
		Quantity:    "1",
		PackSize:    "1kg",
		SKU:         &reqSKU,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/products/PROD00000001/variants", bytes.NewReader(body))
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

func TestProductVariantHandler_CreateProductVariant_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductVariantService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductVariantHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing required fields
	reqBody := models.CreateProductVariantRequest{
		// Missing VariantName, Quantity, and PackSize
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/products/PROD00000001/variants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProductVariantHandler_CreateProductVariant_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductVariantService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductVariantHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("CreateProductVariant", mock.Anything, mock.Anything, mock.AnythingOfType("*models.CreateProductVariantRequest")).
		Return(nil, errors.New("database error"))

	// Create request
	reqSKU := "RICE-1KG-001"
	reqBody := models.CreateProductVariantRequest{
		VariantName: "1kg Pack",
		Quantity:    "1",
		PackSize:    "1kg",
		SKU:         &reqSKU,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/products/PROD00000001/variants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductVariantHandler_GetVariantsByProduct_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductVariantService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductVariantHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	sku1 := "RICE-1KG-001"
	sku2 := "RICE-5KG-001"
	expectedResponse := []models.ProductVariantResponse{
		{
			ID:          "PVAR00000001",
			ProductID:   "PROD00000001",
			VariantName: "1kg Pack",
			SKU:         &sku1,
			IsActive:    true,
		},
		{
			ID:          "PVAR00000002",
			ProductID:   "PROD00000001",
			VariantName: "5kg Pack",
			SKU:         &sku2,
			IsActive:    true,
		},
	}
	mockService.On("GetVariantsByProduct", mock.Anything, "PROD00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/PROD00000001/variants", nil)

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

func TestProductVariantHandler_GetVariantsByProduct_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductVariantService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductVariantHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetVariantsByProduct", mock.Anything, "PROD00000001").
		Return([]models.ProductVariantResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/PROD00000001/variants", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
