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

func TestCollaboratorHandler_CreateCollaborator_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorService)
	mockAAA := testutils.NewMockAAAMiddleware()
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorHandler(mockService, mockAAA)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	email := "test@vendor.com"
	expectedResponse := &models.CollaboratorResponse{
		ID:            "CLAB00000001",
		CompanyName:   "Test Vendor Inc",
		ContactPerson: "John Doe",
		ContactNumber: "1234567890",
		Email:         &email,
		GSTNumber:     "22AAAAA0000A1Z5",
		IsActive:      true,
	}
	mockService.On("CreateCollaborator", mock.Anything, mock.AnythingOfType("*models.CreateCollaboratorRequest"), mock.Anything, mock.Anything).
		Return(expectedResponse, nil)

	// Create request
	emailReq := "test@vendor.com"
	panNumber := "AAAAA0000A"
	bankName := "State Bank of India"
	reqBody := models.CreateCollaboratorRequest{
		CompanyName:   "Test Vendor Inc",
		ContactPerson: "John Doe",
		ContactNumber: "1234567890",
		Email:         &emailReq,
		GSTNumber:     "22AAAAA0000A1Z5",
		PANNumber:     &panNumber,
		BankAccountNo: "1234567890123456",
		BankIFSC:      "SBIN0001234",
		BankName:      &bankName,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/collaborators", bytes.NewReader(body))
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

func TestCollaboratorHandler_CreateCollaborator_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing required fields
	reqBody := models.CreateCollaboratorRequest{
		CompanyName: "Test Vendor Inc",
		// Missing ContactPerson, ContactNumber, GSTNumber, etc.
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/collaborators", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCollaboratorHandler_CreateCollaborator_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("CreateCollaborator", mock.Anything, mock.AnythingOfType("*models.CreateCollaboratorRequest"), mock.Anything, mock.Anything).
		Return(nil, errors.New("database error"))

	// Create request
	email := "test@vendor.com"
	panNumber := "AAAAA0000A"
	bankName := "State Bank of India"
	reqBody := models.CreateCollaboratorRequest{
		CompanyName:   "Test Vendor Inc",
		ContactPerson: "John Doe",
		ContactNumber: "1234567890",
		Email:         &email,
		GSTNumber:     "22AAAAA0000A1Z5",
		PANNumber:     &panNumber,
		BankAccountNo: "1234567890123456",
		BankIFSC:      "SBIN0001234",
		BankName:      &bankName,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/collaborators", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorHandler_GetAllCollaborators_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.CollaboratorResponse{
		{
			ID:            "CLAB00000001",
			CompanyName:   "Vendor 1",
			ContactPerson: "John Doe",
			IsActive:      true,
		},
		{
			ID:            "CLAB00000002",
			CompanyName:   "Vendor 2",
			ContactPerson: "Jane Smith",
			IsActive:      true,
		},
	}
	mockService.On("GetAllCollaborators", mock.Anything, mock.Anything).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/collaborators", nil)

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

func TestCollaboratorHandler_GetAllCollaborators_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetAllCollaborators", mock.Anything, mock.Anything).
		Return([]models.CollaboratorResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/collaborators", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorHandler_GetActiveCollaborators_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.CollaboratorResponse{
		{
			ID:          "CLAB00000001",
			CompanyName: "Active Vendor 1",
			IsActive:    true,
		},
	}
	mockService.On("GetActiveCollaborators", mock.Anything, mock.Anything).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/collaborators/active", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorHandler_SearchCollaborators_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.CollaboratorResponse{
		{
			ID:            "CLAB00000001",
			CompanyName:   "Test Vendor Inc",
			ContactPerson: "John Doe",
			IsActive:      true,
		},
	}
	mockService.On("SearchCollaborators", mock.Anything, "Test", mock.Anything).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/collaborators/search?q=Test", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorHandler_SearchCollaborators_NoResults(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("SearchCollaborators", mock.Anything, "NonExistent", mock.Anything).
		Return([]models.CollaboratorResponse{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/collaborators/search?q=NonExistent", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorHandler_GetCollaborator_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.CollaboratorResponse{
		ID:            "CLAB00000001",
		CompanyName:   "Test Vendor Inc",
		ContactPerson: "John Doe",
		IsActive:      true,
	}
	mockService.On("GetCollaborator", mock.Anything, "CLAB00000001", mock.Anything).
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/collaborators/CLAB00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorHandler_GetCollaborator_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service returns error
	mockService.On("GetCollaborator", mock.Anything, "CLAB99999999", mock.Anything).
		Return(nil, errors.New("collaborator not found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/collaborators/CLAB99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorHandler_UpdateCollaborator_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.CollaboratorResponse{
		ID:            "CLAB00000001",
		CompanyName:   "Updated Vendor Inc",
		ContactPerson: "John Doe",
		IsActive:      true,
	}
	mockService.On("UpdateCollaborator", mock.Anything, "CLAB00000001", mock.AnythingOfType("*models.UpdateCollaboratorRequest"), mock.Anything).
		Return(expectedResponse, nil)

	// Create request
	companyName := "Updated Vendor Inc"
	contactPerson := "John Doe"
	contactNumber := "1234567890"
	reqBody := models.UpdateCollaboratorRequest{
		CompanyName:   &companyName,
		ContactPerson: &contactPerson,
		ContactNumber: &contactNumber,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/collaborators/CLAB00000001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorHandler_UpdateCollaborator_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("UpdateCollaborator", mock.Anything, "CLAB99999999", mock.AnythingOfType("*models.UpdateCollaboratorRequest"), mock.Anything).
		Return(nil, errors.New("collaborator not found"))

	// Create request
	companyName := "Updated Vendor Inc"
	reqBody := models.UpdateCollaboratorRequest{
		CompanyName: &companyName,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/collaborators/CLAB99999999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorHandler_UpdateCollaborator_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid JSON
	req := httptest.NewRequest("PUT", "/api/v1/collaborators/CLAB00000001", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCollaboratorHandler_DeleteCollaborator_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("DeleteCollaborator", mock.Anything, "CLAB00000001", mock.Anything).
		Return(nil)

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/collaborators/CLAB00000001", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCollaboratorHandler_DeleteCollaborator_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockCollaboratorService)
	router := testutils.SetupTestRouter()
	handler := handlers.NewCollaboratorHandler(mockService, testutils.NewMockAAAMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("DeleteCollaborator", mock.Anything, "CLAB99999999", mock.Anything).
		Return(errors.New("collaborator not found"))

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/collaborators/CLAB99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
