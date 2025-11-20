package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"kisanlink-erp/internal/api/handlers"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/utils"
	mockServices "kisanlink-erp/tests/mocks/services"
	"kisanlink-erp/tests/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	// Bypass AAA for testing
	os.Setenv("AAA_ENABLED", "false")
	gin.SetMode(gin.TestMode)
}

// TestBankPaymentsHandler_CreateBankPayment_Success tests successful bank payment creation
func TestBankPaymentsHandler_CreateBankPayment_Success(t *testing.T) {
	mockService := new(mockServices.MockBankPaymentsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewBankPaymentsHandler(mockService, mockAAA, mockLogger)

	request := &models.CreateBankPaymentRequest{
		SaleID:        stringPtr("SALE00000001"),
		PaymentMethod: "upi",
		Amount:        1500.00,
	}

	expectedResponse := &models.BankPaymentResponse{
		ID:             "BPAY00000001",
		SaleID:         stringPtr("SALE00000001"),
		PaymentMethod:  "upi",
		TransactionRef: "TXN123456",
		Amount:         1500.00,
		PaidAt:         "2025-11-10T10:00:00Z",
	}

	mockService.On("CreateBankPayment", mock.AnythingOfType("*models.CreateBankPaymentRequest")).Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/bank-payments", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateBankPayment(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Bank payment created successfully", response["message"])
	assert.NotNil(t, response["data"])
}

// TestBankPaymentsHandler_CreateBankPayment_ValidationError tests validation errors
func TestBankPaymentsHandler_CreateBankPayment_ValidationError(t *testing.T) {
	mockService := new(mockServices.MockBankPaymentsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewBankPaymentsHandler(mockService, mockAAA, mockLogger)

	// Missing required fields
	request := &models.CreateBankPaymentRequest{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/bank-payments", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateBankPayment(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "CreateBankPayment", mock.Anything)
}

// TestBankPaymentsHandler_CreateBankPayment_ServiceError tests service layer errors
func TestBankPaymentsHandler_CreateBankPayment_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockBankPaymentsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewBankPaymentsHandler(mockService, mockAAA, mockLogger)

	request := &models.CreateBankPaymentRequest{
		SaleID:        stringPtr("SALE00000001"),
		PaymentMethod: "upi",
		Amount:        1500.00,
	}

	mockService.On("CreateBankPayment", mock.AnythingOfType("*models.CreateBankPaymentRequest")).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody, _ := json.Marshal(request)
	c.Request, _ = http.NewRequest("POST", "/api/v1/bank-payments", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateBankPayment(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestBankPaymentsHandler_GetAllBankPayments_Success tests fetching all payments
func TestBankPaymentsHandler_GetAllBankPayments_Success(t *testing.T) {
	mockService := new(mockServices.MockBankPaymentsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewBankPaymentsHandler(mockService, mockAAA, mockLogger)

	expectedPayments := []models.BankPaymentResponse{
		{
			ID:            "BPAY00000001",
			SaleID:        stringPtr("SALE00000001"),
			PaymentMethod: "upi",
			Amount:        1500.00,
		},
		{
			ID:            "BPAY00000002",
			ReturnID:      stringPtr("RETN00000001"),
			PaymentMethod: "cash",
			Amount:        500.00,
		},
	}

	mockService.On("GetAllBankPayments", 10, 0).Return(expectedPayments, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/bank-payments", nil)

	handler.GetAllBankPayments(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Bank payments retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])
}

// TestBankPaymentsHandler_GetAllBankPayments_WithPagination tests pagination
func TestBankPaymentsHandler_GetAllBankPayments_WithPagination(t *testing.T) {
	mockService := new(mockServices.MockBankPaymentsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewBankPaymentsHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetAllBankPayments", 10, 20).Return([]models.BankPaymentResponse{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/bank-payments?limit=10&offset=20", nil)

	handler.GetAllBankPayments(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestBankPaymentsHandler_GetBankPayment_Success tests fetching a single payment
func TestBankPaymentsHandler_GetBankPayment_Success(t *testing.T) {
	mockService := new(mockServices.MockBankPaymentsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewBankPaymentsHandler(mockService, mockAAA, mockLogger)

	expectedPayment := &models.BankPaymentResponse{
		ID:            "BPAY00000001",
		SaleID:        stringPtr("SALE00000001"),
		PaymentMethod: "upi",
		Amount:        1500.00,
	}

	mockService.On("GetBankPayment", "BPAY00000001").Return(expectedPayment, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "BPAY00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/bank-payments/BPAY00000001", nil)

	handler.GetBankPayment(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Bank payment retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])
}

// TestBankPaymentsHandler_GetBankPayment_NotFound tests payment not found
func TestBankPaymentsHandler_GetBankPayment_NotFound(t *testing.T) {
	mockService := new(mockServices.MockBankPaymentsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewBankPaymentsHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetBankPayment", "BPAY99999999").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "BPAY99999999"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/bank-payments/BPAY99999999", nil)

	handler.GetBankPayment(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// TestBankPaymentsHandler_GetBankPaymentsBySale_Success tests fetching payments by sale
func TestBankPaymentsHandler_GetBankPaymentsBySale_Success(t *testing.T) {
	mockService := new(mockServices.MockBankPaymentsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewBankPaymentsHandler(mockService, mockAAA, mockLogger)

	expectedPayments := []models.BankPaymentResponse{
		{
			ID:            "BPAY00000001",
			SaleID:        stringPtr("SALE00000001"),
			PaymentMethod: "upi",
			Amount:        1500.00,
		},
	}

	mockService.On("GetBankPaymentsBySaleID", "SALE00000001").Return(expectedPayments, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "saleID", Value: "SALE00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/bank-payments/sale/SALE00000001", nil)

	handler.GetBankPaymentsBySale(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Bank payments for sale retrieved successfully", response["message"])
}

// TestBankPaymentsHandler_GetBankPaymentsBySale_NoPayments tests sale with no payments
func TestBankPaymentsHandler_GetBankPaymentsBySale_NoPayments(t *testing.T) {
	mockService := new(mockServices.MockBankPaymentsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewBankPaymentsHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetBankPaymentsBySaleID", "SALE00000999").Return([]models.BankPaymentResponse{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "saleID", Value: "SALE00000999"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/bank-payments/sale/SALE00000999", nil)

	handler.GetBankPaymentsBySale(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestBankPaymentsHandler_GetBankPaymentsByReturn_Success tests fetching payments by return
func TestBankPaymentsHandler_GetBankPaymentsByReturn_Success(t *testing.T) {
	mockService := new(mockServices.MockBankPaymentsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewBankPaymentsHandler(mockService, mockAAA, mockLogger)

	expectedPayments := []models.BankPaymentResponse{
		{
			ID:            "BPAY00000002",
			ReturnID:      stringPtr("RETN00000001"),
			PaymentMethod: "cash",
			Amount:        500.00,
		},
	}

	mockService.On("GetBankPaymentsByReturnID", "RETN00000001").Return(expectedPayments, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "returnID", Value: "RETN00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/bank-payments/return/RETN00000001", nil)

	handler.GetBankPaymentsByReturn(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Bank payments for return retrieved successfully", response["message"])
}

// TestBankPaymentsHandler_GetBankPaymentsByReturn_ServiceError tests service error
func TestBankPaymentsHandler_GetBankPaymentsByReturn_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockBankPaymentsService)
	mockAAA := testutils.NewMockAAAMiddleware()
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	handler := handlers.NewBankPaymentsHandler(mockService, mockAAA, mockLogger)

	mockService.On("GetBankPaymentsByReturnID", "RETN99999999").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "returnID", Value: "RETN99999999"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/bank-payments/return/RETN99999999", nil)

	handler.GetBankPaymentsByReturn(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}
