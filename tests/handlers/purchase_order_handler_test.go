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

// =====================================================
// CreatePurchaseOrder Tests
// =====================================================

func TestPurchaseOrderHandler_CreatePurchaseOrder_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.PurchaseOrderResponse{
		ID:               "PORD00000001",
		PONumber:         "PO-2025-0001",
		CollaboratorID:   "CLAB00000001",
		CollaboratorName: "Test Vendor Inc",
		WarehouseID:      "WRHS00000001",
		WarehouseName:    "Main Warehouse",
		OrderDate:        "2025-01-15",
		ExpectedDelivery: "2025-01-20",
		Status:           "placed",
		TotalAmount:      15000.00,
		PaymentStatus:    "unpaid",
		PaidAmount:       0,
	}
	mockService.On("CreatePurchaseOrder", mock.Anything, mock.AnythingOfType("*models.CreatePurchaseOrderRequest"), mock.Anything).
		Return(expectedResponse, nil)

	// Create request
	reqBody := models.CreatePurchaseOrderRequest{
		CollaboratorID:   "CLAB00000001",
		WarehouseID:      "WRHS00000001",
		ExpectedDelivery: "2025-01-20",
		Items: []models.CreatePurchaseOrderItemRequest{
			{
				VariantID: "PVAR00000001",
				Quantity:  100,
				UnitPrice: 150.00,
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewReader(body))
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

func TestPurchaseOrderHandler_CreatePurchaseOrder_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing required fields
	reqBody := models.CreatePurchaseOrderRequest{
		CollaboratorID: "CLAB00000001",
		// Missing WarehouseID, ExpectedDelivery, Items
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPurchaseOrderHandler_CreatePurchaseOrder_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("CreatePurchaseOrder", mock.Anything, mock.AnythingOfType("*models.CreatePurchaseOrderRequest"), mock.Anything).
		Return(nil, errors.New("database error"))

	// Create request
	reqBody := models.CreatePurchaseOrderRequest{
		CollaboratorID:   "CLAB00000001",
		WarehouseID:      "WRHS00000001",
		ExpectedDelivery: "2025-01-20",
		Items: []models.CreatePurchaseOrderItemRequest{
			{
				VariantID: "PVAR00000001",
				Quantity:  100,
				UnitPrice: 150.00,
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// =====================================================
// GetAllPurchaseOrders Tests
// =====================================================

func TestPurchaseOrderHandler_GetAllPurchaseOrders_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.PurchaseOrderResponse{
		{
			ID:               "PORD00000001",
			PONumber:         "PO-2025-0001",
			CollaboratorName: "Vendor A",
			Status:           "placed",
			TotalAmount:      15000.00,
		},
		{
			ID:               "PORD00000002",
			PONumber:         "PO-2025-0002",
			CollaboratorName: "Vendor B",
			Status:           "confirmed",
			TotalAmount:      25000.00,
		},
	}
	mockService.On("GetAllPurchaseOrders", mock.Anything, 50, 0).
		Return(expectedResponse, int64(2), nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/purchase-orders", nil)

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

func TestPurchaseOrderHandler_GetAllPurchaseOrders_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetAllPurchaseOrders", mock.Anything, 50, 0).
		Return([]models.PurchaseOrderResponse{}, int64(0), nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/purchase-orders", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPurchaseOrderHandler_GetAllPurchaseOrders_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetAllPurchaseOrders", mock.Anything, 50, 0).
		Return(nil, int64(0), errors.New("database connection failed"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/purchase-orders", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// =====================================================
// GetPurchaseOrder Tests
// =====================================================

func TestPurchaseOrderHandler_GetPurchaseOrder_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.PurchaseOrderResponse{
		ID:               "PORD00000001",
		PONumber:         "PO-2025-0001",
		CollaboratorName: "Test Vendor Inc",
		Status:           "placed",
		TotalAmount:      15000.00,
	}
	mockService.On("GetPurchaseOrder", mock.Anything, "PORD00000001").
		Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/purchase-orders/PORD00000001", nil)

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

func TestPurchaseOrderHandler_GetPurchaseOrder_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetPurchaseOrder", mock.Anything, "PORD99999999").
		Return(nil, errors.New("purchase order not found"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/purchase-orders/PORD99999999", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// =====================================================
// GetPendingDeliveries Tests
// =====================================================

func TestPurchaseOrderHandler_GetPendingDeliveries_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.PurchaseOrderResponse{
		{
			ID:               "PORD00000001",
			PONumber:         "PO-2025-0001",
			Status:           "out_for_delivery",
			ExpectedDelivery: "2025-01-20",
		},
		{
			ID:               "PORD00000002",
			PONumber:         "PO-2025-0002",
			Status:           "confirmed",
			ExpectedDelivery: "2025-01-22",
		},
	}
	mockService.On("GetPendingDeliveries", mock.Anything, 50, 0).
		Return(expectedResponse, int64(2), nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/purchase-orders/pending-deliveries", nil)

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

func TestPurchaseOrderHandler_GetPendingDeliveries_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetPendingDeliveries", mock.Anything, 50, 0).
		Return([]models.PurchaseOrderResponse{}, int64(0), nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/purchase-orders/pending-deliveries", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// =====================================================
// GetPurchaseOrdersByStatus Tests
// =====================================================

func TestPurchaseOrderHandler_GetPurchaseOrdersByStatus_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.PurchaseOrderResponse{
		{
			ID:       "PORD00000001",
			PONumber: "PO-2025-0001",
			Status:   "delivered",
		},
	}
	mockService.On("GetPurchaseOrdersByStatus", mock.Anything, "delivered", 50, 0).
		Return(expectedResponse, int64(1), nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/purchase-orders/status/delivered", nil)

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

func TestPurchaseOrderHandler_GetPurchaseOrdersByStatus_MultipleStatuses(t *testing.T) {
	// Test different status values
	statuses := []string{"placed", "confirmed", "out_for_delivery", "delivered", "paid"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			// Setup
			mockService := new(mockServices.MockPurchaseOrderService)
			router := testutils.SetupTestRouter()
			mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
			handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
			handler.RegisterRoutes(router.Group("/api/v1"))

			// Mock expectations
			mockService.On("GetPurchaseOrdersByStatus", mock.Anything, status, 50, 0).
				Return([]models.PurchaseOrderResponse{}, int64(0), nil)

			// Create request
			req := httptest.NewRequest("GET", "/api/v1/purchase-orders/status/"+status, nil)

			// Execute
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

// =====================================================
// UpdatePurchaseOrderStatus Tests
// =====================================================

func TestPurchaseOrderHandler_UpdatePurchaseOrderStatus_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.PurchaseOrderResponse{
		ID:       "PORD00000001",
		PONumber: "PO-2025-0001",
		Status:   "confirmed",
	}
	mockService.On("UpdatePurchaseOrderStatus", mock.Anything, "PORD00000001", mock.AnythingOfType("*models.UpdatePOStatusRequest"), mock.Anything).
		Return(expectedResponse, nil)

	// Create request
	reqBody := models.UpdatePOStatusRequest{
		Status: "confirmed",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/purchase-orders/PORD00000001/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

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

func TestPurchaseOrderHandler_UpdatePurchaseOrderStatus_WithAcceptAll(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.PurchaseOrderResponse{
		ID:       "PORD00000001",
		PONumber: "PO-2025-0001",
		Status:   "delivered",
	}
	mockService.On("UpdatePurchaseOrderStatus", mock.Anything, "PORD00000001", mock.AnythingOfType("*models.UpdatePOStatusRequest"), mock.Anything).
		Return(expectedResponse, nil)

	// Create request with accept_all pattern
	acceptAll := true
	defaultExpiry := "2025-12-31"
	now := time.Now()
	reqBody := models.UpdatePOStatusRequest{
		Status:            "delivered",
		ActualDelivery:    &now,
		AcceptAll:         &acceptAll,
		DefaultExpiryDate: &defaultExpiry,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/purchase-orders/PORD00000001/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPurchaseOrderHandler_UpdatePurchaseOrderStatus_WithItemDetails(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.PurchaseOrderResponse{
		ID:       "PORD00000001",
		PONumber: "PO-2025-0001",
		Status:   "delivered",
	}
	mockService.On("UpdatePurchaseOrderStatus", mock.Anything, "PORD00000001", mock.AnythingOfType("*models.UpdatePOStatusRequest"), mock.Anything).
		Return(expectedResponse, nil)

	// Create request with item details
	accept := true
	received := int64(100)
	accepted := int64(95)
	batchNum := "BATCH-001"
	reqBody := models.UpdatePOStatusRequest{
		Status: "delivered",
		Items: []models.DeliveryItemRequest{
			{
				POItemID:         "POIM00000001",
				Accept:           &accept,
				ReceivedQuantity: &received,
				AcceptedQuantity: &accepted,
				ExpiryDate:       "2025-12-31",
				BatchNumber:      &batchNum,
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/purchase-orders/PORD00000001/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPurchaseOrderHandler_UpdatePurchaseOrderStatus_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with missing status
	reqBody := models.UpdatePOStatusRequest{
		// Missing required Status field
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/purchase-orders/PORD00000001/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPurchaseOrderHandler_UpdatePurchaseOrderStatus_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("UpdatePurchaseOrderStatus", mock.Anything, "PORD99999999", mock.AnythingOfType("*models.UpdatePOStatusRequest"), mock.Anything).
		Return(nil, errors.New("purchase order not found"))

	// Create request
	reqBody := models.UpdatePOStatusRequest{
		Status: "confirmed",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/purchase-orders/PORD99999999/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// =====================================================
// UpdatePaymentStatus Tests
// =====================================================

func TestPurchaseOrderHandler_UpdatePaymentStatus_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := &models.PurchaseOrderResponse{
		ID:            "PORD00000001",
		PONumber:      "PO-2025-0001",
		PaymentStatus: "partial",
		PaidAmount:    5000.00,
	}
	mockService.On("UpdatePaymentStatus", mock.Anything, "PORD00000001", mock.AnythingOfType("*models.UpdatePOPaymentRequest")).
		Return(expectedResponse, nil)

	// Create request
	reqBody := models.UpdatePOPaymentRequest{
		PaidAmount:    5000.00,
		PaymentStatus: "partial",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/purchase-orders/PORD00000001/payment", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

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

func TestPurchaseOrderHandler_UpdatePaymentStatus_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Create request with invalid data (missing required fields)
	reqBody := models.UpdatePOPaymentRequest{
		// Missing PaidAmount and PaymentStatus
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/purchase-orders/PORD00000001/payment", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPurchaseOrderHandler_UpdatePaymentStatus_NotFound(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("UpdatePaymentStatus", mock.Anything, "PORD99999999", mock.AnythingOfType("*models.UpdatePOPaymentRequest")).
		Return(nil, errors.New("purchase order not found"))

	// Create request
	reqBody := models.UpdatePOPaymentRequest{
		PaidAmount:    5000.00,
		PaymentStatus: "partial",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/purchase-orders/PORD99999999/payment", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// =====================================================
// GetPurchaseOrdersByCollaborator Tests
// =====================================================

func TestPurchaseOrderHandler_GetPurchaseOrdersByCollaborator_Success(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	expectedResponse := []models.PurchaseOrderResponse{
		{
			ID:               "PORD00000001",
			PONumber:         "PO-2025-0001",
			CollaboratorID:   "CLAB00000001",
			CollaboratorName: "Test Vendor Inc",
			Status:           "placed",
		},
		{
			ID:               "PORD00000002",
			PONumber:         "PO-2025-0002",
			CollaboratorID:   "CLAB00000001",
			CollaboratorName: "Test Vendor Inc",
			Status:           "confirmed",
		},
	}
	mockService.On("GetPurchaseOrdersByCollaborator", mock.Anything, "CLAB00000001", 50, 0).
		Return(expectedResponse, int64(2), nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/collaborators/CLAB00000001/purchase-orders", nil)

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

func TestPurchaseOrderHandler_GetPurchaseOrdersByCollaborator_EmptyList(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock expectations
	mockService.On("GetPurchaseOrdersByCollaborator", mock.Anything, "CLAB00000999", 50, 0).
		Return([]models.PurchaseOrderResponse{}, int64(0), nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/collaborators/CLAB00000999/purchase-orders", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPurchaseOrderHandler_GetPurchaseOrdersByCollaborator_ServiceError(t *testing.T) {
	// Setup
	mockService := new(mockServices.MockPurchaseOrderService)
	router := testutils.SetupTestRouter()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewPurchaseOrderHandler(mockService, testutils.NewMockAAAMiddleware(), mockLogger)
	handler.RegisterRoutes(router.Group("/api/v1"))

	// Mock service error
	mockService.On("GetPurchaseOrdersByCollaborator", mock.Anything, "CLAB00000001", 50, 0).
		Return(nil, int64(0), errors.New("database error"))

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/collaborators/CLAB00000001/purchase-orders", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
