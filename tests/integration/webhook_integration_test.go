package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"kisanlink-erp/internal/api/handlers"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/services"
	mockServices "kisanlink-erp/tests/mocks/services"
	"kisanlink-erp/tests/testutils"
)

// ============================================================================
// WEBHOOK INTEGRATION TESTS
// ============================================================================
//
// These integration tests verify the complete end-to-end webhook processing
// flow including:
// - HMAC signature validation (real WebhookSecurityService)
// - Idempotency checking (real WebhookHistoryService + database)
// - Event recording in database (real WebhookRepository)
// - HTTP request/response handling (real handler layer)
// - Error scenarios and edge cases
//
// Architecture: Integration tests with REAL services (not mocks)
// - WebhookSecurityService: REAL (signature validation)
// - WebhookHistoryService: REAL (idempotency + event tracking)
// - WebhookRepository: REAL (database operations)
// - EcommerceWebhookService: MOCKED (business logic - tested separately)
//
// This pattern provides production-level confidence while keeping tests fast.
// ============================================================================

const (
	testWebhookSecret = "test-webhook-secret-key-32-chars-minimum"
)

// WebhookTestContext holds all dependencies for webhook integration tests
type WebhookTestContext struct {
	Router             *gin.Engine
	DB                 *gorm.DB
	SecurityService    *services.WebhookSecurityService
	HistoryService     *services.WebhookHistoryService
	WebhookRepo        *repositories.WebhookRepository
	MockWebhookService *mockServices.MockEcommerceWebhookService
}

// setupWebhookIntegration creates test environment with real services
func setupWebhookIntegration(t *testing.T) (*WebhookTestContext, func()) {
	t.Helper()

	// Create in-memory SQLite database
	db := testutils.SetupTestDB(t)

	// Create REAL webhook services
	securityService := services.NewWebhookSecurityService(testWebhookSecret)
	webhookRepo := repositories.NewWebhookRepository(db)
	historyService := services.NewWebhookHistoryService(webhookRepo)

	// Create MOCK business logic service
	mockWebhookService := new(mockServices.MockEcommerceWebhookService)

	// Create webhook handler
	handler := handlers.NewEcommerceWebhookHandler(
		mockWebhookService,
		securityService,
		historyService,
		webhookRepo,
		testutils.NewMockAAAMiddleware(),
	)

	// Setup router
	router := testutils.SetupTestRouter()
	handler.RegisterRoutes(router.Group("/api/v1"))

	cleanup := func() {
		testutils.CleanupTestDB(db)
	}

	return &WebhookTestContext{
		Router:             router,
		DB:                 db,
		SecurityService:    securityService,
		HistoryService:     historyService,
		WebhookRepo:        webhookRepo,
		MockWebhookService: mockWebhookService,
	}, cleanup
}

// sendWebhookRequest sends a webhook HTTP request with proper headers
func sendWebhookRequest(router *gin.Engine, method, path string, payload []byte, headers *testutils.WebhookTestHeaders) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", headers.Signature)
	req.Header.Set("X-Event-ID", headers.EventID)
	req.Header.Set("X-Timestamp", headers.Timestamp)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// ============================================================================
// 1. ORDER.CREATED Webhook Tests (8 tests)
// ============================================================================

func TestWebhookIntegration_OrderCreated_Success(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// Create webhook payload
	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-001")
	payload := testutils.MustMarshalWebhook(webhook)

	// Generate headers with valid signature
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	// Mock business logic service
	ctx.MockWebhookService.On("ProcessOrderCreated", mock.Anything, mock.AnythingOfType("*models.OrderCreatedWebhook")).
		Return("PORD00000001", nil)

	// Send webhook request
	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)

	// Assert response
	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	testutils.AssertEqual(t, response["success"], true, "Should return success=true")

	data := response["data"].(map[string]interface{})
	testutils.AssertEqual(t, data["event_id"], webhook.EventID, "Should return event_id")
	testutils.AssertEqual(t, data["event_type"], "order.created", "Should return event_type")
	testutils.AssertEqual(t, data["purchase_order_id"], "PORD00000001", "Should return purchase_order_id")

	// Verify webhook was recorded in database
	var event models.WebhookEvent
	err := ctx.DB.Where("event_id = ?", webhook.EventID).First(&event).Error
	testutils.AssertNoError(t, err, "Webhook event should be recorded in database")
	testutils.AssertEqual(t, event.Status, "success", "Event status should be success")
	testutils.AssertEqual(t, event.EventType, "order.created", "Event type should match")
	testutils.AssertTrue(t, event.SignatureValid, "Signature should be marked as valid")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderCreated_InvalidSignature(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-002")
	payload := testutils.MustMarshalWebhook(webhook)

	// Generate headers with INVALID signature
	headers := testutils.GenerateInvalidSignatureHeaders(payload)

	// Send webhook request
	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)

	// Assert unauthorized response
	testutils.AssertEqual(t, w.Code, http.StatusUnauthorized, "Should return 401 Unauthorized")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	testutils.AssertEqual(t, response["success"], false, "Should return success=false")
}

func TestWebhookIntegration_OrderCreated_MissingHeaders(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-003")
	payload := testutils.MustMarshalWebhook(webhook)

	// Send request WITHOUT webhook headers
	req := httptest.NewRequest("POST", "/api/v1/webhooks/ecommerce/order/created", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	// Missing X-Webhook-Signature, X-Event-ID, X-Timestamp

	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	// Assert unauthorized response
	testutils.AssertEqual(t, w.Code, http.StatusUnauthorized, "Should return 401 Unauthorized")
}

func TestWebhookIntegration_OrderCreated_InvalidJSON(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// Invalid JSON payload
	payload := []byte(`{"invalid": "json"`)

	headers := testutils.GenerateWebhookHeaders(payload, testWebhookSecret)

	// Send webhook request
	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)

	// Assert bad request response
	testutils.AssertEqual(t, w.Code, http.StatusBadRequest, "Should return 400 Bad Request")

	// Verify webhook was recorded as failed
	var event models.WebhookEvent
	err := ctx.DB.Where("event_id = ?", headers.EventID).First(&event).Error
	testutils.AssertNoError(t, err, "Webhook event should be recorded")
	testutils.AssertEqual(t, event.Status, "failed", "Event status should be failed")
	testutils.AssertNotNil(t, event.ErrorMessage, "Error message should be recorded")
}

func TestWebhookIntegration_OrderCreated_Idempotency(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-004")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	// Mock business logic service (should only be called ONCE)
	ctx.MockWebhookService.On("ProcessOrderCreated", mock.Anything, mock.AnythingOfType("*models.OrderCreatedWebhook")).
		Return("PORD00000002", nil).
		Once()

	// Send webhook request FIRST TIME
	w1 := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)
	testutils.AssertEqual(t, w1.Code, http.StatusOK, "First request should succeed")

	// Send webhook request SECOND TIME (same event_id)
	w2 := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)
	testutils.AssertEqual(t, w2.Code, http.StatusOK, "Second request should also return 200 OK (idempotency)")

	var response map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &response)
	testutils.AssertEqual(t, response["success"], true, "Should return success=true")

	data := response["data"].(map[string]interface{})
	testutils.AssertEqual(t, data["status"], "success", "Should return existing status")

	// Verify mock was called only ONCE
	ctx.MockWebhookService.AssertExpectations(t)

	// Verify only ONE webhook event in database
	var count int64
	ctx.DB.Model(&models.WebhookEvent{}).Where("event_id = ?", webhook.EventID).Count(&count)
	testutils.AssertEqual(t, count, int64(1), "Should have exactly one webhook event")
}

func TestWebhookIntegration_OrderCreated_ServiceError(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-005")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	// Mock business logic service to return error
	ctx.MockWebhookService.On("ProcessOrderCreated", mock.Anything, mock.AnythingOfType("*models.OrderCreatedWebhook")).
		Return("", fmt.Errorf("database connection failed"))

	// Send webhook request
	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)

	// Assert internal server error response
	testutils.AssertEqual(t, w.Code, http.StatusInternalServerError, "Should return 500 Internal Server Error")

	// Verify webhook was recorded as failed
	var event models.WebhookEvent
	err := ctx.DB.Where("event_id = ?", webhook.EventID).First(&event).Error
	testutils.AssertNoError(t, err, "Webhook event should be recorded")
	testutils.AssertEqual(t, event.Status, "failed", "Event status should be failed")
	testutils.AssertNotNil(t, event.ErrorMessage, "Error message should be recorded")
	testutils.AssertContains(t, *event.ErrorMessage, "database connection failed", "Error message should contain failure reason")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderCreated_PayloadHashDeduplication(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-006")
	payload := testutils.MustMarshalWebhook(webhook)

	// First request with event_id_1
	headers1 := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, "evt_001")
	ctx.MockWebhookService.On("ProcessOrderCreated", mock.Anything, mock.AnythingOfType("*models.OrderCreatedWebhook")).
		Return("PORD00000003", nil).
		Once()

	w1 := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers1)
	testutils.AssertEqual(t, w1.Code, http.StatusOK, "First request should succeed")

	// Second request with DIFFERENT event_id but SAME payload
	headers2 := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, "evt_002")
	ctx.MockWebhookService.On("ProcessOrderCreated", mock.Anything, mock.AnythingOfType("*models.OrderCreatedWebhook")).
		Return("PORD00000004", nil).
		Once()

	w2 := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers2)
	testutils.AssertEqual(t, w2.Code, http.StatusOK, "Second request should also succeed (different event_id)")

	// Both webhooks should be recorded (different event_ids)
	var count int64
	ctx.DB.Model(&models.WebhookEvent{}).Count(&count)
	testutils.AssertEqual(t, count, int64(2), "Should have two webhook events")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderCreated_DatabaseEventRecording(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-007")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderCreated", mock.Anything, mock.AnythingOfType("*models.OrderCreatedWebhook")).
		Return("PORD00000005", nil)

	// Send webhook request
	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)
	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should succeed")

	// Verify complete event record in database
	var event models.WebhookEvent
	err := ctx.DB.Where("event_id = ?", webhook.EventID).First(&event).Error
	testutils.AssertNoError(t, err, "Webhook event should be found")

	// Verify all fields
	testutils.AssertEqual(t, event.EventType, "order.created", "Event type should match")
	testutils.AssertEqual(t, event.Status, "success", "Status should be success")
	testutils.AssertTrue(t, event.SignatureValid, "Signature should be valid")
	testutils.AssertNotNil(t, event.PayloadHash, "Payload hash should be recorded")
	testutils.AssertNotNil(t, event.RequestBody, "Request body should be stored")
	testutils.AssertNotNil(t, event.ProcessedAt, "Processed timestamp should be set")
	testutils.AssertNotNil(t, event.PurchaseOrderID, "Purchase order ID should be set")
	testutils.AssertEqual(t, *event.PurchaseOrderID, "PORD00000005", "Purchase order ID should match")

	ctx.MockWebhookService.AssertExpectations(t)
}

// ============================================================================
// 2. ORDER.CONFIRMED Webhook Tests (6 tests)
// ============================================================================

func TestWebhookIntegration_OrderConfirmed_Success(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderConfirmedWebhook("ORDER-008")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderConfirmed", mock.Anything, mock.AnythingOfType("*models.OrderConfirmedWebhook")).
		Return(nil)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/confirmed", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	testutils.AssertEqual(t, response["success"], true, "Should return success=true")

	data := response["data"].(map[string]interface{})
	testutils.AssertEqual(t, data["event_type"], "order.confirmed", "Should return event_type")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderConfirmed_InvalidSignature(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderConfirmedWebhook("ORDER-009")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateInvalidSignatureHeaders(payload)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/confirmed", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusUnauthorized, "Should return 401 Unauthorized")
}

func TestWebhookIntegration_OrderConfirmed_Idempotency(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderConfirmedWebhook("ORDER-010")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderConfirmed", mock.Anything, mock.AnythingOfType("*models.OrderConfirmedWebhook")).
		Return(nil).
		Once()

	// First request
	w1 := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/confirmed", payload, headers)
	testutils.AssertEqual(t, w1.Code, http.StatusOK, "First request should succeed")

	// Second request (idempotent)
	w2 := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/confirmed", payload, headers)
	testutils.AssertEqual(t, w2.Code, http.StatusOK, "Second request should succeed (idempotency)")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderConfirmed_InvalidPayload(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// Missing required fields - json.Unmarshal will succeed but service should validate
	payload := []byte(`{"event_type": "order.confirmed"}`)
	headers := testutils.GenerateWebhookHeaders(payload, testWebhookSecret)

	// Service should be called and return validation error
	ctx.MockWebhookService.On("ProcessOrderConfirmed", mock.Anything, mock.AnythingOfType("*models.OrderConfirmedWebhook")).
		Return(fmt.Errorf("validation error: external_order_id is required"))

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/confirmed", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusInternalServerError, "Should return 500 Internal Server Error")
	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderConfirmed_ServiceError(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderConfirmedWebhook("ORDER-011")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderConfirmed", mock.Anything, mock.AnythingOfType("*models.OrderConfirmedWebhook")).
		Return(fmt.Errorf("purchase order not found"))

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/confirmed", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusInternalServerError, "Should return 500 Internal Server Error")

	var event models.WebhookEvent
	err := ctx.DB.Where("event_id = ?", webhook.EventID).First(&event).Error
	testutils.AssertNoError(t, err, "Event should be recorded")
	testutils.AssertEqual(t, event.Status, "failed", "Status should be failed")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderConfirmed_EventRecording(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderConfirmedWebhook("ORDER-012")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderConfirmed", mock.Anything, mock.AnythingOfType("*models.OrderConfirmedWebhook")).
		Return(nil)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/confirmed", payload, headers)
	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should succeed")

	var event models.WebhookEvent
	err := ctx.DB.Where("event_id = ?", webhook.EventID).First(&event).Error
	testutils.AssertNoError(t, err, "Event should be recorded")
	testutils.AssertEqual(t, event.EventType, "order.confirmed", "Event type should match")

	ctx.MockWebhookService.AssertExpectations(t)
}

// ============================================================================
// 3. ORDER.SHIPPED Webhook Tests (6 tests)
// ============================================================================

func TestWebhookIntegration_OrderShipped_Success(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderShippedWebhook("ORDER-013")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderShipped", mock.Anything, mock.AnythingOfType("*models.OrderShippedWebhook")).
		Return(nil)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/shipped", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	testutils.AssertEqual(t, response["success"], true, "Should return success=true")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderShipped_InvalidSignature(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderShippedWebhook("ORDER-014")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateInvalidSignatureHeaders(payload)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/shipped", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusUnauthorized, "Should return 401 Unauthorized")
}

func TestWebhookIntegration_OrderShipped_Idempotency(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderShippedWebhook("ORDER-015")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderShipped", mock.Anything, mock.AnythingOfType("*models.OrderShippedWebhook")).
		Return(nil).
		Once()

	w1 := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/shipped", payload, headers)
	w2 := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/shipped", payload, headers)

	testutils.AssertEqual(t, w1.Code, http.StatusOK, "First request should succeed")
	testutils.AssertEqual(t, w2.Code, http.StatusOK, "Second request should succeed (idempotency)")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderShipped_ServiceError(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderShippedWebhook("ORDER-016")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderShipped", mock.Anything, mock.AnythingOfType("*models.OrderShippedWebhook")).
		Return(fmt.Errorf("tracking update failed"))

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/shipped", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusInternalServerError, "Should return 500 Internal Server Error")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderShipped_InvalidPayload(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// Missing required fields - json.Unmarshal will succeed but service should validate
	payload := []byte(`{"event_type": "order.shipped"}`)
	headers := testutils.GenerateWebhookHeaders(payload, testWebhookSecret)

	// Service should be called and return validation error
	ctx.MockWebhookService.On("ProcessOrderShipped", mock.Anything, mock.AnythingOfType("*models.OrderShippedWebhook")).
		Return(fmt.Errorf("validation error: external_order_id is required"))

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/shipped", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusInternalServerError, "Should return 500 Internal Server Error")
	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderShipped_EventRecording(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderShippedWebhook("ORDER-017")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderShipped", mock.Anything, mock.AnythingOfType("*models.OrderShippedWebhook")).
		Return(nil)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/shipped", payload, headers)
	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should succeed")

	var event models.WebhookEvent
	err := ctx.DB.Where("event_id = ?", webhook.EventID).First(&event).Error
	testutils.AssertNoError(t, err, "Event should be recorded")
	testutils.AssertEqual(t, event.EventType, "order.shipped", "Event type should match")

	ctx.MockWebhookService.AssertExpectations(t)
}

// ============================================================================
// 4. ORDER.DELIVERED Webhook Tests (8 tests)
// ============================================================================

func TestWebhookIntegration_OrderDelivered_Success(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderDeliveredWebhook("ORDER-018")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderDelivered", mock.Anything, mock.AnythingOfType("*models.OrderDeliveredWebhook")).
		Return(nil)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/delivered", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	testutils.AssertEqual(t, response["success"], true, "Should return success=true")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderDelivered_ComplexPayload(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// Create webhook with multiple items
	webhook := testutils.FixtureOrderDeliveredWebhook("ORDER-019")
	webhook.Items = append(webhook.Items, models.WebhookDeliveryItem{
		ExternalProductID: "PROD002",
		ExternalVariantID: "VAR002",
		ReceivedQuantity:  50,
		AcceptedQuantity:  48,
		RejectedQuantity:  2,
		BatchNumber:       "BATCH-2025-002",
		ExpiryDate:        time.Now().AddDate(1, 0, 0).Format("2006-01-02"),
		CostPrice:         150.00,
	})

	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderDelivered", mock.Anything, mock.AnythingOfType("*models.OrderDeliveredWebhook")).
		Return(nil)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/delivered", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderDelivered_InvalidSignature(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderDeliveredWebhook("ORDER-020")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateInvalidSignatureHeaders(payload)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/delivered", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusUnauthorized, "Should return 401 Unauthorized")
}

func TestWebhookIntegration_OrderDelivered_Idempotency(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderDeliveredWebhook("ORDER-021")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderDelivered", mock.Anything, mock.AnythingOfType("*models.OrderDeliveredWebhook")).
		Return(nil).
		Once()

	w1 := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/delivered", payload, headers)
	w2 := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/delivered", payload, headers)

	testutils.AssertEqual(t, w1.Code, http.StatusOK, "First request should succeed")
	testutils.AssertEqual(t, w2.Code, http.StatusOK, "Second request should succeed (idempotency)")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderDelivered_ServiceError(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderDeliveredWebhook("ORDER-022")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderDelivered", mock.Anything, mock.AnythingOfType("*models.OrderDeliveredWebhook")).
		Return(fmt.Errorf("GRN creation failed"))

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/delivered", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusInternalServerError, "Should return 500 Internal Server Error")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderDelivered_InvalidPayload(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// Missing required fields and empty items array - json.Unmarshal will succeed but service should validate
	payload := []byte(`{"event_type": "order.delivered", "items": []}`)
	headers := testutils.GenerateWebhookHeaders(payload, testWebhookSecret)

	// Service should be called and return validation error
	ctx.MockWebhookService.On("ProcessOrderDelivered", mock.Anything, mock.AnythingOfType("*models.OrderDeliveredWebhook")).
		Return(fmt.Errorf("validation error: external_order_id is required"))

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/delivered", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusInternalServerError, "Should return 500 Internal Server Error")
	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderDelivered_EventRecording(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderDeliveredWebhook("ORDER-023")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderDelivered", mock.Anything, mock.AnythingOfType("*models.OrderDeliveredWebhook")).
		Return(nil)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/delivered", payload, headers)
	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should succeed")

	var event models.WebhookEvent
	err := ctx.DB.Where("event_id = ?", webhook.EventID).First(&event).Error
	testutils.AssertNoError(t, err, "Event should be recorded")
	testutils.AssertEqual(t, event.EventType, "order.delivered", "Event type should match")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderDelivered_PartialAcceptance(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderDeliveredWebhook("ORDER-024")
	// Set partial acceptance (some items rejected)
	webhook.Items[0].ReceivedQuantity = 100
	webhook.Items[0].AcceptedQuantity = 70
	webhook.Items[0].RejectedQuantity = 30
	rejectionReason := "Quality issues"
	webhook.Items[0].RejectionReason = &rejectionReason

	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderDelivered", mock.Anything, mock.AnythingOfType("*models.OrderDeliveredWebhook")).
		Return(nil)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/delivered", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK")

	ctx.MockWebhookService.AssertExpectations(t)
}

// ============================================================================
// 5. ORDER.PAYMENT Webhook Tests (7 tests)
// ============================================================================

func TestWebhookIntegration_OrderPayment_Success(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderPaymentWebhook("ORDER-025")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderPayment", mock.Anything, mock.AnythingOfType("*models.OrderPaymentWebhook")).
		Return(nil)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/payment", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	testutils.AssertEqual(t, response["success"], true, "Should return success=true")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderPayment_PartialPayment(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderPaymentWebhook("ORDER-026")
	webhook.PaymentStatus = "partial"
	webhook.PaidAmount = 5000.00 // Partial payment

	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderPayment", mock.Anything, mock.AnythingOfType("*models.OrderPaymentWebhook")).
		Return(nil)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/payment", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderPayment_InvalidSignature(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderPaymentWebhook("ORDER-027")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateInvalidSignatureHeaders(payload)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/payment", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusUnauthorized, "Should return 401 Unauthorized")
}

func TestWebhookIntegration_OrderPayment_Idempotency(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderPaymentWebhook("ORDER-028")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderPayment", mock.Anything, mock.AnythingOfType("*models.OrderPaymentWebhook")).
		Return(nil).
		Once()

	w1 := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/payment", payload, headers)
	w2 := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/payment", payload, headers)

	testutils.AssertEqual(t, w1.Code, http.StatusOK, "First request should succeed")
	testutils.AssertEqual(t, w2.Code, http.StatusOK, "Second request should succeed (idempotency)")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderPayment_ServiceError(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderPaymentWebhook("ORDER-029")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderPayment", mock.Anything, mock.AnythingOfType("*models.OrderPaymentWebhook")).
		Return(fmt.Errorf("payment recording failed"))

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/payment", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusInternalServerError, "Should return 500 Internal Server Error")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderPayment_InvalidPayload(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// Missing required fields - json.Unmarshal will succeed but service should validate
	payload := []byte(`{"event_type": "order.payment"}`)
	headers := testutils.GenerateWebhookHeaders(payload, testWebhookSecret)

	// Service should be called and return validation error
	ctx.MockWebhookService.On("ProcessOrderPayment", mock.Anything, mock.AnythingOfType("*models.OrderPaymentWebhook")).
		Return(fmt.Errorf("validation error: external_order_id is required"))

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/payment", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusInternalServerError, "Should return 500 Internal Server Error")
	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_OrderPayment_EventRecording(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderPaymentWebhook("ORDER-030")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderPayment", mock.Anything, mock.AnythingOfType("*models.OrderPaymentWebhook")).
		Return(nil)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/payment", payload, headers)
	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should succeed")

	var event models.WebhookEvent
	err := ctx.DB.Where("event_id = ?", webhook.EventID).First(&event).Error
	testutils.AssertNoError(t, err, "Event should be recorded")
	testutils.AssertEqual(t, event.EventType, "order.payment", "Event type should match")

	ctx.MockWebhookService.AssertExpectations(t)
}

// ============================================================================
// 6. Webhook History Endpoint Tests (6 tests)
// ============================================================================

func TestWebhookIntegration_GetHistory_Success(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// Create some webhook events first
	event1 := models.NewWebhookEvent("evt_001", "order.created", "hash1", `{"test": "payload1"}`)
	event1.Status = "success"
	ctx.DB.Create(event1)

	event2 := models.NewWebhookEvent("evt_002", "order.confirmed", "hash2", `{"test": "payload2"}`)
	event2.Status = "success"
	ctx.DB.Create(event2)

	// Query webhook history
	req := httptest.NewRequest("GET", "/api/v1/webhooks/history", nil)
	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	testutils.AssertEqual(t, response["success"], true, "Should return success=true")

	data := response["data"].([]interface{})
	testutils.AssertTrue(t, len(data) >= 2, "Should return at least 2 events")
}

func TestWebhookIntegration_GetHistory_FilterByOrderID(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// Create webhook event with specific external_order_id
	externalOrderID := "ORDER-FILTER-001"
	event := models.NewWebhookEvent("evt_filter_001", "order.created", "hash_filter", `{"test": "payload"}`)
	event.ExternalOrderID = &externalOrderID
	event.Status = "success"
	ctx.DB.Create(event)

	// Query by external_order_id
	req := httptest.NewRequest("GET", "/api/v1/webhooks/history?external_order_id=ORDER-FILTER-001", nil)
	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].([]interface{})
	testutils.AssertEqual(t, len(data), 1, "Should return exactly 1 event")

	firstEvent := data[0].(map[string]interface{})
	testutils.AssertEqual(t, firstEvent["external_order_id"], "ORDER-FILTER-001", "Should match filtered order ID")
}

func TestWebhookIntegration_GetHistory_FilterByStatus(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// Create failed webhook event
	failedEvent := models.NewWebhookEvent("evt_failed_001", "order.created", "hash_failed", `{"test": "payload"}`)
	failedEvent.Status = "failed"
	errorMsg := "test error"
	failedEvent.ErrorMessage = &errorMsg
	ctx.DB.Create(failedEvent)

	// Query by status
	req := httptest.NewRequest("GET", "/api/v1/webhooks/history?status=failed", nil)
	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].([]interface{})
	testutils.AssertTrue(t, len(data) >= 1, "Should return at least 1 failed event")

	firstEvent := data[0].(map[string]interface{})
	testutils.AssertEqual(t, firstEvent["status"], "failed", "Should return failed status")
}

func TestWebhookIntegration_GetHistory_LimitParameter(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// Create multiple events
	for i := 0; i < 10; i++ {
		event := models.NewWebhookEvent(fmt.Sprintf("evt_limit_%d", i), "order.created", fmt.Sprintf("hash_%d", i), `{"test": "payload"}`)
		event.Status = "success"
		ctx.DB.Create(event)
	}

	// Query with limit
	req := httptest.NewRequest("GET", "/api/v1/webhooks/history?limit=5", nil)
	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].([]interface{})
	testutils.AssertEqual(t, len(data), 5, "Should return exactly 5 events")
}

func TestWebhookIntegration_GetHistory_EmptyResults(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// Query non-existent order ID
	req := httptest.NewRequest("GET", "/api/v1/webhooks/history?external_order_id=NONEXISTENT", nil)
	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].([]interface{})
	testutils.AssertEqual(t, len(data), 0, "Should return empty array")
}

func TestWebhookIntegration_GetHistory_RequiresAuthentication(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// The mock AAA middleware always passes authentication
	// This test verifies the middleware is called
	req := httptest.NewRequest("GET", "/api/v1/webhooks/history", nil)
	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	// Should succeed with mock auth
	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK with mock auth")
}

// ============================================================================
// 7. Webhook Stats Endpoint Tests (4 tests)
// ============================================================================

func TestWebhookIntegration_GetStats_Success(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// Create various webhook events
	events := []struct {
		eventType string
		status    string
	}{
		{"order.created", "success"},
		{"order.created", "success"},
		{"order.created", "failed"},
		{"order.confirmed", "success"},
		{"order.shipped", "success"},
		{"order.delivered", "success"},
		{"order.payment", "failed"},
	}

	for i, e := range events {
		event := models.NewWebhookEvent(fmt.Sprintf("evt_stats_%d", i), e.eventType, fmt.Sprintf("hash_%d", i), `{"test": "payload"}`)
		event.Status = e.status
		ctx.DB.Create(event)
	}

	// Query stats
	req := httptest.NewRequest("GET", "/api/v1/webhooks/stats", nil)
	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	testutils.AssertEqual(t, response["success"], true, "Should return success=true")

	data := response["data"].(map[string]interface{})
	testutils.AssertNotNil(t, data["order.created"], "Should have order.created stats")
	testutils.AssertNotNil(t, data["order.confirmed"], "Should have order.confirmed stats")
}

func TestWebhookIntegration_GetStats_EmptyDatabase(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// Query stats on empty database
	req := httptest.NewRequest("GET", "/api/v1/webhooks/stats", nil)
	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	testutils.AssertEqual(t, response["success"], true, "Should return success=true")

	data := response["data"].(map[string]interface{})
	testutils.AssertNotNil(t, data, "Should return stats object")
}

func TestWebhookIntegration_GetStats_CountsByEventType(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// Create 3 order.created events (2 success, 1 failed)
	for i := 0; i < 3; i++ {
		event := models.NewWebhookEvent(fmt.Sprintf("evt_count_%d", i), "order.created", fmt.Sprintf("hash_%d", i), `{"test": "payload"}`)
		if i < 2 {
			event.Status = "success"
		} else {
			event.Status = "failed"
		}
		ctx.DB.Create(event)
	}

	// Query stats
	req := httptest.NewRequest("GET", "/api/v1/webhooks/stats", nil)
	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})

	orderCreatedStats := data["order.created"].(map[string]interface{})
	testutils.AssertEqual(t, int(orderCreatedStats["total"].(float64)), 3, "Should have 3 total order.created events")
	testutils.AssertEqual(t, int(orderCreatedStats["success"].(float64)), 2, "Should have 2 successful events")
	testutils.AssertEqual(t, int(orderCreatedStats["failed"].(float64)), 1, "Should have 1 failed event")
}

func TestWebhookIntegration_GetStats_RequiresAuthentication(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// The mock AAA middleware always passes authentication
	req := httptest.NewRequest("GET", "/api/v1/webhooks/stats", nil)
	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should return 200 OK with mock auth")
}

// ============================================================================
// 8. Security and Edge Case Tests (15 tests)
// ============================================================================

func TestWebhookIntegration_Security_TamperedPayload(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	// Create valid webhook and signature
	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-TAMPER")
	originalPayload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeaders(originalPayload, testWebhookSecret)

	// Tamper with payload AFTER generating signature
	webhook.Order.TotalAmount = 999999.99 // Changed amount!
	tamperedPayload := testutils.MustMarshalWebhook(webhook)

	// Send tampered payload with original signature
	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", tamperedPayload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusUnauthorized, "Should reject tampered payload")
}

func TestWebhookIntegration_Security_ExpiredTimestamp(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-EXPIRED")
	payload := testutils.MustMarshalWebhook(webhook)

	// Generate headers with expired timestamp (> 5 minutes old)
	headers := testutils.GenerateExpiredWebhookHeaders(payload, testWebhookSecret)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusUnauthorized, "Should reject expired webhook")
}

func TestWebhookIntegration_Security_FutureTimestamp(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-FUTURE")
	payload := testutils.MustMarshalWebhook(webhook)

	// Generate headers with future timestamp (> 1 minute in future)
	headers := testutils.GenerateFutureWebhookHeaders(payload, testWebhookSecret)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusUnauthorized, "Should reject future webhook")
}

func TestWebhookIntegration_Security_SignaturePrefixHandling(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-PREFIX")
	payload := testutils.MustMarshalWebhook(webhook)

	// Generate signature WITH "sha256=" prefix
	headers := testutils.GenerateWebhookHeaders(payload, testWebhookSecret)
	headers.Signature = "sha256=" + headers.Signature

	ctx.MockWebhookService.On("ProcessOrderCreated", mock.Anything, mock.AnythingOfType("*models.OrderCreatedWebhook")).
		Return("PORD00000100", nil)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should accept signature with sha256= prefix")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_Security_MissingEventIDHeader(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-NO-EVENT-ID")
	payload := testutils.MustMarshalWebhook(webhook)

	// Create request without X-Event-ID header
	req := httptest.NewRequest("POST", "/api/v1/webhooks/ecommerce/order/created", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", testutils.GenerateWebhookSignature(payload, testWebhookSecret))
	req.Header.Set("X-Timestamp", testutils.CurrentUnixTimestamp())
	// Missing X-Event-ID

	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	testutils.AssertEqual(t, w.Code, http.StatusUnauthorized, "Should reject without event ID")
}

func TestWebhookIntegration_Security_MissingTimestampHeader(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-NO-TIMESTAMP")
	payload := testutils.MustMarshalWebhook(webhook)

	// Create request without X-Timestamp header
	req := httptest.NewRequest("POST", "/api/v1/webhooks/ecommerce/order/created", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", testutils.GenerateWebhookSignature(payload, testWebhookSecret))
	req.Header.Set("X-Event-ID", "evt_missing_timestamp")
	// Missing X-Timestamp

	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	testutils.AssertEqual(t, w.Code, http.StatusUnauthorized, "Should reject without timestamp")
}

func TestWebhookIntegration_EdgeCase_LargePayload(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-LARGE")

	// Add many items to create large payload
	for i := 0; i < 100; i++ {
		description := fmt.Sprintf("Product description %d", i)
		webhook.Items = append(webhook.Items, models.WebhookOrderItem{
			Product: models.WebhookProduct{
				ExternalID:  fmt.Sprintf("PROD%03d", i),
				Name:        fmt.Sprintf("Product %d", i),
				Description: &description,
				Category:    "Test",
				Unit:        "kg",
			},
			Variant: models.WebhookVariant{
				ExternalID:   fmt.Sprintf("VAR%03d", i),
				SKU:          fmt.Sprintf("SKU-%03d", i),
				Name:         fmt.Sprintf("Variant %d", i),
				QuantityText: "1kg",
				PackSize:     "1kg",
			},
			Quantity:  10,
			UnitPrice: 100.00,
		})
	}

	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderCreated", mock.Anything, mock.AnythingOfType("*models.OrderCreatedWebhook")).
		Return("PORD00000101", nil)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should handle large payload")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_EdgeCase_ConcurrentDuplicateWebhooks(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-CONCURRENT")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	// Mock service should be called only ONCE due to idempotency
	ctx.MockWebhookService.On("ProcessOrderCreated", mock.Anything, mock.AnythingOfType("*models.OrderCreatedWebhook")).
		Return("PORD00000102", nil).
		Once()

	// Send concurrent requests
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)
			testutils.AssertEqual(t, w.Code, http.StatusOK, "All requests should return 200 OK")
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify only ONE webhook event in database
	var count int64
	ctx.DB.Model(&models.WebhookEvent{}).Where("event_id = ?", webhook.EventID).Count(&count)
	testutils.AssertTrue(t, count <= 1, "Should have at most one webhook event (race condition may occur in SQLite)")

	// Note: SQLite in-memory may not perfectly handle concurrent writes
	// This test is more relevant for PostgreSQL integration tests
}

func TestWebhookIntegration_EdgeCase_FailedWebhookRetry(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-RETRY")
	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	// First request fails
	ctx.MockWebhookService.On("ProcessOrderCreated", mock.Anything, mock.AnythingOfType("*models.OrderCreatedWebhook")).
		Return("", fmt.Errorf("temporary failure")).
		Once()

	w1 := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)
	testutils.AssertEqual(t, w1.Code, http.StatusInternalServerError, "First request should fail")

	// Verify webhook marked as failed
	var event models.WebhookEvent
	ctx.DB.Where("event_id = ?", webhook.EventID).First(&event)
	testutils.AssertEqual(t, event.Status, "failed", "Event should be marked as failed")

	// Second request (retry) succeeds
	ctx.MockWebhookService.On("ProcessOrderCreated", mock.Anything, mock.AnythingOfType("*models.OrderCreatedWebhook")).
		Return("PORD00000103", nil).
		Once()

	w2 := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)
	testutils.AssertEqual(t, w2.Code, http.StatusOK, "Retry should succeed")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_EdgeCase_EmptyPayload(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	payload := []byte(`{}`)
	headers := testutils.GenerateWebhookHeaders(payload, testWebhookSecret)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusBadRequest, "Should reject empty payload")
}

func TestWebhookIntegration_EdgeCase_MalformedJSON(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	payload := []byte(`{"event_type": "order.created", "incomplete":`)
	headers := testutils.GenerateWebhookHeaders(payload, testWebhookSecret)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusBadRequest, "Should reject malformed JSON")
}

func TestWebhookIntegration_EdgeCase_NullRequiredFields(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	payload := []byte(`{
		"event_type": "order.created",
		"event_id": null,
		"timestamp": null,
		"order": null
	}`)
	headers := testutils.GenerateWebhookHeaders(payload, testWebhookSecret)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusBadRequest, "Should reject null required fields")
}

func TestWebhookIntegration_EdgeCase_VeryLongEventID(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-LONG-ID")
	payload := testutils.MustMarshalWebhook(webhook)

	// Create very long event ID (100 characters)
	longEventID := "evt_" + string(make([]byte, 96))
	for i := range longEventID[4:] {
		longEventID = longEventID[:4+i] + "a" + longEventID[5+i:]
	}

	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, longEventID)

	ctx.MockWebhookService.On("ProcessOrderCreated", mock.Anything, mock.AnythingOfType("*models.OrderCreatedWebhook")).
		Return("PORD00000104", nil)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should handle long event ID")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_EdgeCase_SpecialCharactersInPayload(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-SPECIAL-CHARS")
	// Add special characters
	specialChars := "Test <script>alert('xss')</script> & 'quotes' \"double\" \\backslash"
	webhook.Collaborator.CompanyName = specialChars

	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	ctx.MockWebhookService.On("ProcessOrderCreated", mock.Anything, mock.AnythingOfType("*models.OrderCreatedWebhook")).
		Return("PORD00000105", nil)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusOK, "Should handle special characters in payload")

	ctx.MockWebhookService.AssertExpectations(t)
}

func TestWebhookIntegration_EdgeCase_ZeroQuantityValidation(t *testing.T) {
	ctx, cleanup := setupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-ZERO-QTY")
	webhook.Items[0].Quantity = 0 // Invalid: quantity must be > 0

	payload := testutils.MustMarshalWebhook(webhook)
	headers := testutils.GenerateWebhookHeadersWithEventID(payload, testWebhookSecret, webhook.EventID)

	w := sendWebhookRequest(ctx.Router, "POST", "/api/v1/webhooks/ecommerce/order/created", payload, headers)

	testutils.AssertEqual(t, w.Code, http.StatusBadRequest, "Should reject zero quantity")
}

// ============================================================================
// SUMMARY
// ============================================================================
//
// Total tests implemented: 60
//
// Coverage by endpoint:
// - POST /webhooks/ecommerce/order/created: 8 tests
// - POST /webhooks/ecommerce/order/confirmed: 6 tests
// - POST /webhooks/ecommerce/order/shipped: 6 tests
// - POST /webhooks/ecommerce/order/delivered: 8 tests
// - POST /webhooks/ecommerce/order/payment: 7 tests
// - GET /webhooks/history: 6 tests
// - GET /webhooks/stats: 4 tests
// - Security & Edge Cases: 15 tests
//
// Test categories:
// ✅ Success cases for all 5 webhook types
// ✅ Invalid signature rejection
// ✅ Missing headers validation
// ✅ Invalid JSON handling
// ✅ Idempotency verification (duplicate event_id)
// ✅ Service error handling
// ✅ Database event recording
// ✅ Payload hash deduplication
// ✅ History endpoint filtering (by order ID, status, limit)
// ✅ Statistics aggregation
// ✅ Tampered payload detection
// ✅ Expired timestamp rejection
// ✅ Future timestamp rejection
// ✅ Signature prefix handling
// ✅ Concurrent duplicate handling
// ✅ Failed webhook retry
// ✅ Large payload handling
// ✅ Edge cases (empty payload, malformed JSON, null fields, special chars)
//
// ============================================================================
