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
// CreateRefundPolicy Tests
// ============================================================================

func TestRefundPoliciesHandler_CreateRefundPolicy_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockRefundPoliciesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewRefundPoliciesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	desc := "Standard 30-day return policy"
	expectedResponse := &models.RefundPolicyResponse{
		ID:            "RFPL00000001",
		PolicyName:    "30-Day Standard Return",
		Description:   &desc,
		MaxDays:       30,
		RestockingFee: 5.0,
	}
	mockService.On("CreateRefundPolicy", mock.AnythingOfType("*models.CreateRefundPolicyRequest")).
		Return(expectedResponse, nil)

	// Create request
	description := "Standard 30-day return policy"
	reqBody := models.CreateRefundPolicyRequest{
		PolicyName:    "30-Day Standard Return",
		Description:   &description,
		MaxDays:       30,
		RestockingFee: 5.0,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/refund-policies", bytes.NewReader(body))
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

func TestRefundPoliciesHandler_CreateRefundPolicy_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockRefundPoliciesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewRefundPoliciesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing required fields
	reqBody := models.CreateRefundPolicyRequest{
		// Missing PolicyName and MaxDays (required fields)
		RestockingFee: 5.0,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/refund-policies", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRefundPoliciesHandler_CreateRefundPolicy_InvalidJSON(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockRefundPoliciesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewRefundPoliciesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/api/v1/refund-policies", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRefundPoliciesHandler_CreateRefundPolicy_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockRefundPoliciesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewRefundPoliciesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("CreateRefundPolicy", mock.AnythingOfType("*models.CreateRefundPolicyRequest")).
		Return(nil, errors.New("database error"))

	// Create request
	reqBody := models.CreateRefundPolicyRequest{
		PolicyName: "30-Day Standard Return",
		MaxDays:    30,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/refund-policies", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// GetAllRefundPolicies Tests
// ============================================================================

func TestRefundPoliciesHandler_GetAllRefundPolicies_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockRefundPoliciesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewRefundPoliciesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	desc1 := "Standard 30-day return policy"
	desc2 := "Extended 60-day return policy"
	expectedResponse := []models.RefundPolicyResponse{
		{
			ID:            "RFPL00000001",
			PolicyName:    "30-Day Standard Return",
			Description:   &desc1,
			MaxDays:       30,
			RestockingFee: 5.0,
		},
		{
			ID:            "RFPL00000002",
			PolicyName:    "60-Day Extended Return",
			Description:   &desc2,
			MaxDays:       60,
			RestockingFee: 10.0,
		},
	}
	mockService.On("GetAllRefundPolicies", 10, 0).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/refund-policies", nil)

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

func TestRefundPoliciesHandler_GetAllRefundPolicies_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockRefundPoliciesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewRefundPoliciesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetAllRefundPolicies", 10, 0).
		Return([]models.RefundPolicyResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/refund-policies", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestRefundPoliciesHandler_GetAllRefundPolicies_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockRefundPoliciesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewRefundPoliciesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetAllRefundPolicies", 10, 0).
		Return(nil, errors.New("database error"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/refund-policies", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// GetRefundPolicy Tests
// ============================================================================

func TestRefundPoliciesHandler_GetRefundPolicy_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockRefundPoliciesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewRefundPoliciesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	desc := "Standard 30-day return policy"
	expectedResponse := &models.RefundPolicyResponse{
		ID:            "RFPL00000001",
		PolicyName:    "30-Day Standard Return",
		Description:   &desc,
		MaxDays:       30,
		RestockingFee: 5.0,
	}
	mockService.On("GetRefundPolicy", "RFPL00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/refund-policies/RFPL00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestRefundPoliciesHandler_GetRefundPolicy_MissingID(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockRefundPoliciesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewRefundPoliciesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing ID
	req := httptest.NewRequest("GET", "/api/v1/refund-policies/", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - Gin redirects trailing slash (301)
	assert.Equal(t, http.StatusMovedPermanently, w.Code)
}

func TestRefundPoliciesHandler_GetRefundPolicy_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockRefundPoliciesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewRefundPoliciesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetRefundPolicy", "RFPL99999999").
		Return(nil, errors.New("policy not found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/refund-policies/RFPL99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// UpdateRefundPolicy Tests
// ============================================================================

func TestRefundPoliciesHandler_UpdateRefundPolicy_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockRefundPoliciesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewRefundPoliciesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	updatedDesc := "Updated return policy description"
	expectedResponse := &models.RefundPolicyResponse{
		ID:            "RFPL00000001",
		PolicyName:    "30-Day Standard Return",
		Description:   &updatedDesc,
		MaxDays:       45,
		RestockingFee: 7.5,
	}
	mockService.On("UpdateRefundPolicy", "RFPL00000001", mock.AnythingOfType("*models.UpdateRefundPolicyRequest")).
		Return(expectedResponse, nil)

	// Create request
	newMaxDays := 45
	newFee := 7.5
	reqBody := models.UpdateRefundPolicyRequest{
		MaxDays:       &newMaxDays,
		RestockingFee: &newFee,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/refund-policies/RFPL00000001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestRefundPoliciesHandler_UpdateRefundPolicy_MissingID(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockRefundPoliciesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewRefundPoliciesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing ID
	newMaxDays := 45
	reqBody := models.UpdateRefundPolicyRequest{
		MaxDays: &newMaxDays,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/refund-policies/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - should 404 as route won't match
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRefundPoliciesHandler_UpdateRefundPolicy_InvalidJSON(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockRefundPoliciesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewRefundPoliciesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid JSON
	req := httptest.NewRequest("PATCH", "/api/v1/refund-policies/RFPL00000001", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRefundPoliciesHandler_UpdateRefundPolicy_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockRefundPoliciesService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewRefundPoliciesHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("UpdateRefundPolicy", "RFPL00000001", mock.AnythingOfType("*models.UpdateRefundPolicyRequest")).
		Return(nil, errors.New("database error"))

	// Create request
	newMaxDays := 45
	reqBody := models.UpdateRefundPolicyRequest{
		MaxDays: &newMaxDays,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/refund-policies/RFPL00000001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
