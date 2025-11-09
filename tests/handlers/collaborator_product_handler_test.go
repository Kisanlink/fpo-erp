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

// ============================================================================
// AddProductToCollaborator Tests
// ============================================================================

func TestCollaboratorProductHandler_AddProductToCollaborator_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.CollaboratorProductResponse{
		ID:               "CPRD00000001",
		CollaboratorID:   "CLAB00000001",
		CollaboratorName: "Test Vendor",
		ProductID:        "PROD00000001",
		ProductName:      "Organic Rice",
		BrandName:        "Premium Rice Co",
		HSNCode:          "10063010",
		GSTRate:          5.0,
		IsActive:         true,
	}
	mockService.On("AddProductToCollaborator", mock.Anything, "CLAB00000001", mock.AnythingOfType("*models.CreateCollaboratorProductRequest")).
		Return(expectedResponse, nil)

	// Create request
	reqBody := models.CreateCollaboratorProductRequest{
		ProductID: "PROD00000001",
		BrandName: "Premium Rice Co",
		HSNCode:   "10063010",
		GSTRate:   5.0,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/collaborators/CLAB00000001/products", bytes.NewReader(body))
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

func TestCollaboratorProductHandler_AddProductToCollaborator_MissingCollaboratorID(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing ID (empty path param)
	reqBody := models.CreateCollaboratorProductRequest{
		ProductID: "PROD00000001",
		BrandName: "Premium Rice Co",
		HSNCode:   "10063010",
		GSTRate:   5.0,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/collaborators//products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - should 404 as route won't match
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCollaboratorProductHandler_AddProductToCollaborator_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing required fields
	reqBody := models.CreateCollaboratorProductRequest{
		// Missing ProductID, BrandName, HSNCode, GSTRate
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/collaborators/CLAB00000001/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCollaboratorProductHandler_AddProductToCollaborator_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("AddProductToCollaborator", mock.Anything, "CLAB00000001", mock.AnythingOfType("*models.CreateCollaboratorProductRequest")).
		Return(nil, errors.New("database error"))

	// Create request
	reqBody := models.CreateCollaboratorProductRequest{
		ProductID: "PROD00000001",
		BrandName: "Premium Rice Co",
		HSNCode:   "10063010",
		GSTRate:   5.0,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/collaborators/CLAB00000001/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// GetProductsByCollaborator Tests
// ============================================================================

func TestCollaboratorProductHandler_GetProductsByCollaborator_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.CollaboratorProductResponse{
		{
			ID:               "CPRD00000001",
			CollaboratorID:   "CLAB00000001",
			CollaboratorName: "Test Vendor",
			ProductID:        "PROD00000001",
			ProductName:      "Organic Rice",
			BrandName:        "Premium Rice Co",
			HSNCode:          "10063010",
			GSTRate:          5.0,
			IsActive:         true,
		},
		{
			ID:               "CPRD00000002",
			CollaboratorID:   "CLAB00000001",
			CollaboratorName: "Test Vendor",
			ProductID:        "PROD00000002",
			ProductName:      "Wheat Flour",
			BrandName:        "Golden Wheat",
			HSNCode:          "11010010",
			GSTRate:          5.0,
			IsActive:         true,
		},
	}
	mockService.On("GetProductsByCollaborator", mock.Anything, "CLAB00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/collaborators/CLAB00000001/products", nil)

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

func TestCollaboratorProductHandler_GetProductsByCollaborator_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetProductsByCollaborator", mock.Anything, "CLAB00000001").
		Return([]models.CollaboratorProductResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/collaborators/CLAB00000001/products", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorProductHandler_GetProductsByCollaborator_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetProductsByCollaborator", mock.Anything, "CLAB00000001").
		Return(nil, errors.New("database error"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/collaborators/CLAB00000001/products", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// GetCollaboratorsByProduct Tests
// ============================================================================

func TestCollaboratorProductHandler_GetCollaboratorsByProduct_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.CollaboratorProductResponse{
		{
			ID:               "CPRD00000001",
			CollaboratorID:   "CLAB00000001",
			CollaboratorName: "Test Vendor 1",
			ProductID:        "PROD00000001",
			ProductName:      "Organic Rice",
			BrandName:        "Premium Rice Co",
			HSNCode:          "10063010",
			GSTRate:          5.0,
			IsActive:         true,
		},
		{
			ID:               "CPRD00000002",
			CollaboratorID:   "CLAB00000002",
			CollaboratorName: "Test Vendor 2",
			ProductID:        "PROD00000001",
			ProductName:      "Organic Rice",
			BrandName:        "Golden Harvest",
			HSNCode:          "10063010",
			GSTRate:          5.0,
			IsActive:         true,
		},
	}
	mockService.On("GetCollaboratorsByProduct", mock.Anything, "PROD00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/PROD00000001/collaborators", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorProductHandler_GetCollaboratorsByProduct_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetCollaboratorsByProduct", mock.Anything, "PROD00000001").
		Return([]models.CollaboratorProductResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/PROD00000001/collaborators", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorProductHandler_GetCollaboratorsByProduct_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetCollaboratorsByProduct", mock.Anything, "PROD00000001").
		Return(nil, errors.New("database error"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/PROD00000001/collaborators", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// GetCollaboratorProduct Tests
// ============================================================================

func TestCollaboratorProductHandler_GetCollaboratorProduct_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.CollaboratorProductResponse{
		ID:               "CPRD00000001",
		CollaboratorID:   "CLAB00000001",
		CollaboratorName: "Test Vendor",
		ProductID:        "PROD00000001",
		ProductName:      "Organic Rice",
		BrandName:        "Premium Rice Co",
		HSNCode:          "10063010",
		GSTRate:          5.0,
		IsActive:         true,
	}
	mockService.On("GetCollaboratorProduct", mock.Anything, "CPRD00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/collaborator-products/CPRD00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorProductHandler_GetCollaboratorProduct_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetCollaboratorProduct", mock.Anything, "CPRD99999999").
		Return(nil, errors.New("not found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/collaborator-products/CPRD99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// UpdateCollaboratorProduct Tests
// ============================================================================

func TestCollaboratorProductHandler_UpdateCollaboratorProduct_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.CollaboratorProductResponse{
		ID:               "CPRD00000001",
		CollaboratorID:   "CLAB00000001",
		CollaboratorName: "Test Vendor",
		ProductID:        "PROD00000001",
		ProductName:      "Organic Rice",
		BrandName:        "Updated Brand",
		HSNCode:          "10063010",
		GSTRate:          12.0,
		IsActive:         true,
	}
	mockService.On("UpdateCollaboratorProduct", mock.Anything, "CPRD00000001", mock.AnythingOfType("*models.UpdateCollaboratorProductRequest")).
		Return(expectedResponse, nil)

	// Create request
	brandName := "Updated Brand"
	gstRate := 12.0
	reqBody := models.UpdateCollaboratorProductRequest{
		BrandName: &brandName,
		GSTRate:   &gstRate,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/collaborator-products/CPRD00000001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorProductHandler_UpdateCollaboratorProduct_MissingID(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing ID
	brandName := "Updated Brand"
	reqBody := models.UpdateCollaboratorProductRequest{
		BrandName: &brandName,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/collaborator-products/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - should 404 as route won't match
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCollaboratorProductHandler_UpdateCollaboratorProduct_InvalidJSON(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid JSON
	req := httptest.NewRequest("PUT", "/api/v1/collaborator-products/CPRD00000001", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCollaboratorProductHandler_UpdateCollaboratorProduct_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("UpdateCollaboratorProduct", mock.Anything, "CPRD00000001", mock.AnythingOfType("*models.UpdateCollaboratorProductRequest")).
		Return(nil, errors.New("database error"))

	// Create request
	brandName := "Updated Brand"
	reqBody := models.UpdateCollaboratorProductRequest{
		BrandName: &brandName,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/collaborator-products/CPRD00000001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// RemoveProductFromCollaborator Tests
// ============================================================================

func TestCollaboratorProductHandler_RemoveProductFromCollaborator_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("RemoveProductFromCollaborator", mock.Anything, "CLAB00000001", "PROD00000001").
		Return(nil)

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/collaborators/CLAB00000001/products/PROD00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorProductHandler_RemoveProductFromCollaborator_MissingCollaboratorID(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing collaborator ID
	req := httptest.NewRequest("DELETE", "/api/v1/collaborators//products/PROD00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - handler validates IDs and returns 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCollaboratorProductHandler_RemoveProductFromCollaborator_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("RemoveProductFromCollaborator", mock.Anything, "CLAB00000001", "PROD00000001").
		Return(errors.New("database error"))

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/collaborators/CLAB00000001/products/PROD00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// DeleteCollaboratorProduct Tests
// ============================================================================

func TestCollaboratorProductHandler_DeleteCollaboratorProduct_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("DeleteCollaboratorProduct", mock.Anything, "CPRD00000001").
		Return(nil)

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/collaborator-products/CPRD00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorProductHandler_DeleteCollaboratorProduct_MissingID(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing ID
	req := httptest.NewRequest("DELETE", "/api/v1/collaborator-products/", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - should 404 as route won't match
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCollaboratorProductHandler_DeleteCollaboratorProduct_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorProductService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorProductHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("DeleteCollaboratorProduct", mock.Anything, "CPRD00000001").
		Return(errors.New("database error"))

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/collaborator-products/CPRD00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
