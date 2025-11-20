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

func TestProductHandler_CreateProduct_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewProductHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	desc := "Premium quality organic rice"
	expectedResponse := &models.ProductResponse{
		ID:          "PROD00000001",
		Name:        "Organic Rice",
		Description: &desc,
	}
	mockService.On("CreateProduct", mock.AnythingOfType("*models.CreateProductRequest")).
		Return(expectedResponse, nil)

	// Create request
	description := "Premium quality organic rice"
	reqBody := models.CreateProductRequest{
		Name:        "Organic Rice",
		Description: &description,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/products", bytes.NewReader(body))
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

func TestProductHandler_CreateProduct_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewProductHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing required fields
	reqBody := models.CreateProductRequest{
		// Missing Name (required field)
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProductHandler_CreateProduct_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewProductHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("CreateProduct", mock.AnythingOfType("*models.CreateProductRequest")).
		Return(nil, errors.New("database error"))

	// Create request
	description := "Test product description"
	reqBody := models.CreateProductRequest{
		Name:        "Organic Rice",
		Description: &description,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_UpdateProduct_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewProductHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	updatedDesc := "Updated description"
	expectedResponse := &models.ProductResponse{
		ID:          "PROD00000001",
		Name:        "Updated Rice",
		Description: &updatedDesc,
	}
	mockService.On("UpdateProduct", "PROD00000001", mock.AnythingOfType("*models.UpdateProductRequest")).
		Return(expectedResponse, nil)

	// Create request
	name := "Updated Rice"
	description := "Updated description"
	reqBody := models.UpdateProductRequest{
		Name:        &name,
		Description: &description,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/products/PROD00000001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_UpdateProduct_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewProductHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("UpdateProduct", "PROD99999999", mock.AnythingOfType("*models.UpdateProductRequest")).
		Return(nil, errors.New("product not found"))

	// Create request
	name := "Updated Rice"
	reqBody := models.UpdateProductRequest{
		Name: &name,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/products/PROD99999999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_DeleteProduct_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewProductHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("DeleteProduct", "PROD00000001").
		Return(nil)

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/products/PROD00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_DeleteProduct_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewProductHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("DeleteProduct", "PROD99999999").
		Return(errors.New("product not found"))

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/products/PROD99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_GetAllProducts_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewProductHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.ProductResponse{
		{
			ID:   "PROD00000001",
			Name: "Organic Rice",
		},
		{
			ID:   "PROD00000002",
			Name: "Wheat Flour",
		},
	}
	mockService.On("GetAllProducts", mock.Anything).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products", nil)

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

func TestProductHandler_GetAllProducts_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewProductHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetAllProducts", mock.Anything).
		Return([]models.ProductResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_SearchProducts_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewProductHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.ProductResponse{
		{
			ID:   "PROD00000001",
			Name: "Organic Rice",
		},
	}
	mockService.On("SearchProducts", "Rice").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/search?q=Rice", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_SearchProducts_NoResults(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewProductHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("SearchProducts", "NonExistent").
		Return([]models.ProductResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/search?q=NonExistent", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_GetProduct_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewProductHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	desc := "Premium quality organic rice"
	expectedResponse := &models.ProductResponse{
		ID:          "PROD00000001",
		Name:        "Organic Rice",
		Description: &desc,
	}
	mockService.On("GetProduct", "PROD00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/PROD00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_GetProduct_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewProductHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetProduct", "PROD99999999").
		Return(nil, errors.New("product not found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/PROD99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_GetProductWithPrices_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewProductHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.ProductWithPricesResponse{
		ID:   "PROD00000001",
		Name: "Organic Rice",
		Prices: []models.ProductPriceResponse{
			{
				ID:        "PRIC00000001",
				PriceType: "retail",
				Price:     100.50,
			},
		},
	}
	mockService.On("GetProductWithPrices", "PROD00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/PROD00000001/with-prices", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_GetProductWithPrices_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewProductHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetProductWithPrices", "PROD99999999").
		Return(nil, errors.New("product not found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/PROD99999999/with-prices", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}
