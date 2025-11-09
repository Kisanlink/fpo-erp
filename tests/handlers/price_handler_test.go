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
	mockServices "kisanlink-erp/tests/mocks/services"
	"kisanlink-erp/tests/testutils"
)

func TestPriceHandler_CreateProductPrice_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	validFrom := time.Now().Format(time.RFC3339)
	validTo := time.Now().AddDate(0, 1, 0).Format(time.RFC3339)
	expectedResponse := &models.ProductPriceResponse{
		ID:            "PRIC00000001",
		VariantID:     "PVAR00000001",
		PriceType:     "retail",
		Price:         100.50,
		EffectiveFrom: validFrom,
		EffectiveTo:   &validTo,
	}
	mockService.On("CreateProductPrice", mock.AnythingOfType("*models.CreateProductPriceRequest")).
		Return(expectedResponse, nil)

	// Create request
	effectiveFrom := time.Now().Format(time.RFC3339)
	effectiveTo := time.Now().AddDate(0, 1, 0).Format(time.RFC3339)
	reqBody := models.CreateProductPriceRequest{
		VariantID:     "PVAR00000001",
		PriceType:     "retail",
		Price:         100.50,
		EffectiveFrom: &effectiveFrom,
		EffectiveTo:   &effectiveTo,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/prices", bytes.NewReader(body))
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

func TestPriceHandler_CreateProductPrice_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing required fields
	reqBody := models.CreateProductPriceRequest{
		VariantID: "PVAR00000001",
		// Missing PriceType, Amount, ValidFrom
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/prices", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPriceHandler_CreateProductPrice_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("CreateProductPrice", mock.AnythingOfType("*models.CreateProductPriceRequest")).
		Return(nil, errors.New("database error"))

	// Create request
	effectiveFrom := time.Now().Format(time.RFC3339)
	reqBody := models.CreateProductPriceRequest{
		VariantID:     "PVAR00000001",
		PriceType:     "retail",
		Price:         100.50,
		EffectiveFrom: &effectiveFrom,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/prices", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestPriceHandler_UpdateProductPrice_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	validFrom := time.Now().Format(time.RFC3339)
	validTo := time.Now().AddDate(0, 1, 0).Format(time.RFC3339)
	expectedResponse := &models.ProductPriceResponse{
		ID:            "PRIC00000001",
		VariantID:     "PVAR00000001",
		PriceType:     "retail",
		Price:         120.00,
		EffectiveFrom: validFrom,
		EffectiveTo:   &validTo,
	}
	mockService.On("UpdateProductPrice", "PRIC00000001", mock.AnythingOfType("*models.UpdateProductPriceRequest")).
		Return(expectedResponse, nil)

	// Create request
	price := 120.00
	effectiveTo := time.Now().AddDate(0, 1, 0).Format(time.RFC3339)
	reqBody := models.UpdateProductPriceRequest{
		Price:       &price,
		EffectiveTo: &effectiveTo,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/prices/PRIC00000001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPriceHandler_UpdateProductPrice_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("UpdateProductPrice", "PRIC99999999", mock.AnythingOfType("*models.UpdateProductPriceRequest")).
		Return(nil, errors.New("price not found"))

	// Create request
	price := 120.00
	reqBody := models.UpdateProductPriceRequest{
		Price: &price,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/prices/PRIC99999999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestPriceHandler_DeleteProductPrice_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("DeleteProductPrice", "PRIC00000001").
		Return(nil)

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/prices/PRIC00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPriceHandler_DeleteProductPrice_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("DeleteProductPrice", "PRIC99999999").
		Return(errors.New("price not found"))

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/prices/PRIC99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestPriceHandler_GetProductPrice_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	validFrom := time.Now().Format(time.RFC3339)
	expectedResponse := &models.ProductPriceResponse{
		ID:            "PRIC00000001",
		VariantID:     "PVAR00000001",
		PriceType:     "retail",
		Price:         100.50,
		EffectiveFrom: validFrom,
	}
	mockService.On("GetProductPrice", "PRIC00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/prices/PRIC00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPriceHandler_GetProductPrice_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetProductPrice", "PRIC99999999").
		Return(nil, errors.New("price not found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/prices/PRIC99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestPriceHandler_GetExpiredPrices_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	validFrom := time.Now().AddDate(0, -2, 0).Format(time.RFC3339)
	validTo := time.Now().AddDate(0, -1, 0).Format(time.RFC3339)
	expectedResponse := []models.ProductPriceResponse{
		{
			ID:            "PRIC00000001",
			VariantID:     "PVAR00000001",
			PriceType:     "retail",
			Price:         100.50,
			EffectiveFrom: validFrom,
			EffectiveTo:   &validTo,
		},
	}
	mockService.On("GetExpiredPrices").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/prices/expired", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPriceHandler_GetExpiredPrices_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetExpiredPrices").
		Return([]models.ProductPriceResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/prices/expired", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPriceHandler_GetVariantPrices_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	validFrom := time.Now().Format(time.RFC3339)
	expectedResponse := []models.ProductPriceResponse{
		{
			ID:            "PRIC00000001",
			VariantID:     "PVAR00000001",
			PriceType:     "retail",
			Price:         100.50,
			EffectiveFrom: validFrom,
		},
		{
			ID:            "PRIC00000002",
			VariantID:     "PVAR00000001",
			PriceType:     "wholesale",
			Price:         90.00,
			EffectiveFrom: validFrom,
		},
	}
	mockService.On("GetVariantPrices", "PVAR00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/variants/PVAR00000001/prices", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPriceHandler_GetVariantPrices_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetVariantPrices", "PVAR00000001").
		Return([]models.ProductPriceResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/variants/PVAR00000001/prices", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPriceHandler_GetCurrentPrice_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	validFrom := time.Now().Format(time.RFC3339)
	expectedResponse := &models.ProductPriceResponse{
		ID:            "PRIC00000001",
		VariantID:     "PVAR00000001",
		PriceType:     "retail",
		Price:         100.50,
		EffectiveFrom: validFrom,
	}
	mockService.On("GetCurrentPrice", "PVAR00000001", "retail").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/variants/PVAR00000001/prices/current?price_type=retail", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPriceHandler_GetCurrentPrice_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetCurrentPrice", "PVAR99999999", "retail").
		Return(nil, errors.New("no current price found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/variants/PVAR99999999/prices/current?price_type=retail", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestPriceHandler_GetCurrentPrice_MissingPriceType(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request without price_type parameter
	req := httptest.NewRequest("GET", "/api/v1/variants/PVAR00000001/prices/current", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPriceHandler_CreateProductPriceForVariant_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	validFrom := time.Now().Format(time.RFC3339)
	validTo := time.Now().AddDate(0, 1, 0).Format(time.RFC3339)
	expectedResponse := &models.ProductPriceResponse{
		ID:            "PRIC00000001",
		VariantID:     "PVAR00000001",
		PriceType:     "retail",
		Price:         100.50,
		EffectiveFrom: validFrom,
		EffectiveTo:   &validTo,
	}
	mockService.On("CreateProductPrice", mock.AnythingOfType("*models.CreateProductPriceRequest")).
		Return(expectedResponse, nil)

	// Create request
	effectiveFrom := time.Now().Format(time.RFC3339)
	effectiveTo := time.Now().AddDate(0, 1, 0).Format(time.RFC3339)
	reqBody := models.CreateProductPriceRequest{
		VariantID:     "PVAR00000001",
		PriceType:     "retail",
		Price:         100.50,
		EffectiveFrom: &effectiveFrom,
		EffectiveTo:   &effectiveTo,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/variants/PVAR00000001/prices", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestPriceHandler_CreateProductPriceForVariant_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockProductPriceService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewProductPriceHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing required fields
	reqBody := models.CreateProductPriceRequest{
		VariantID: "PVAR00000001",
		// Missing PriceType, Amount, ValidFrom
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/variants/PVAR00000001/prices", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
