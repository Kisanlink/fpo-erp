package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// EcommerceWebhookHandler handles incoming webhooks from e-commerce platform
type EcommerceWebhookHandler struct {
	webhookService  interfaces.EcommerceWebhookServiceInterface
	securityService *services.WebhookSecurityService
	historyService  *services.WebhookHistoryService
	webhookRepo     *repositories.WebhookRepository
	aaaMiddleware   interface{ Authenticate() gin.HandlerFunc }
}

// NewEcommerceWebhookHandler creates a new e-commerce webhook handler
func NewEcommerceWebhookHandler(
	webhookService interfaces.EcommerceWebhookServiceInterface,
	securityService *services.WebhookSecurityService,
	historyService *services.WebhookHistoryService,
	webhookRepo *repositories.WebhookRepository,
	aaaMiddleware interface{ Authenticate() gin.HandlerFunc },
) *EcommerceWebhookHandler {
	return &EcommerceWebhookHandler{
		webhookService:  webhookService,
		securityService: securityService,
		historyService:  historyService,
		webhookRepo:     webhookRepo,
		aaaMiddleware:   aaaMiddleware,
	}
}

// ========================================
// Helper Functions
// ========================================

// isUniqueConstraintError checks if the error is a UNIQUE constraint violation
// Works with both SQLite and PostgreSQL error messages
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	// SQLite: "UNIQUE constraint failed: webhook_events.event_id"
	if strings.Contains(errMsg, "UNIQUE constraint failed") {
		return true
	}
	// PostgreSQL: "duplicate key value violates unique constraint"
	if strings.Contains(errMsg, "duplicate key value violates unique constraint") {
		return true
	}
	return false
}

// ========================================
// Common Webhook Processing Logic
// ========================================

// processWebhook is a generic handler for all webhook events
func (h *EcommerceWebhookHandler) processWebhook(
	c *gin.Context,
	eventType string,
	parseFunc func([]byte) (interface{}, error),
	processFunc func(context.Context, interface{}) (string, error),
) {
	// Read raw request body for signature verification
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to read request body", err)
		return
	}

	// Verify webhook signature and headers
	headers, err := h.securityService.ValidateWebhook(c, bodyBytes)
	if err != nil {
		utils.UnauthorizedResponse(c, "Webhook validation failed: "+err.Error())
		return
	}

	// Check idempotency (duplicate event detection)
	alreadyProcessed, existingEvent, err := h.historyService.CheckIdempotency(c.Request.Context(), headers.EventID)
	if err != nil {
		utils.HandleServiceError(c, "Failed to check webhook idempotency", err)
		return
	}

	if alreadyProcessed {
		// Return success for duplicate events (idempotency)
		utils.OKResponse(c, "Webhook already processed", gin.H{
			"event_id":     headers.EventID,
			"status":       existingEvent.Status,
			"processed_at": existingEvent.ProcessedAt,
		})
		return
	}

	// Calculate payload hash for deduplication
	hash := sha256.Sum256(bodyBytes)
	payloadHash := hex.EncodeToString(hash[:])

	// Create or reuse webhook event record
	var webhookEvent *models.WebhookEvent
	sourceIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	if existingEvent != nil {
		// Reuse existing event for retry (status is "failed")
		// This prevents UNIQUE constraint violations on retry attempts
		webhookEvent = existingEvent
		webhookEvent.Status = "processing"
		webhookEvent.ErrorMessage = nil
		webhookEvent.ProcessedAt = nil
	} else {
		// Create new webhook event record using constructor
		webhookEvent = models.NewWebhookEvent(headers.EventID, eventType, payloadHash, string(bodyBytes))
		webhookEvent.SourceIP = &sourceIP
		webhookEvent.UserAgent = &userAgent
		webhookEvent.SignatureValid = true

		// Record webhook in history
		if err := h.historyService.RecordWebhook(c.Request.Context(), webhookEvent); err != nil {
			// Check if error is UNIQUE constraint violation (race condition)
			if isUniqueConstraintError(err) {
				// Race condition: Another concurrent request already created this event
				// Treat as idempotent - refetch the existing event and return success
				existingEvent, fetchErr := h.webhookRepo.FindByEventID(c.Request.Context(), headers.EventID)
				if fetchErr == nil && existingEvent != nil {
					utils.OKResponse(c, "Webhook already processed", gin.H{
						"event_id":     headers.EventID,
						"status":       existingEvent.Status,
						"processed_at": existingEvent.ProcessedAt,
					})
					return
				}
				// If refetch also fails, fall through to generic error
			}
			utils.HandleServiceError(c, "Failed to record webhook", err)
			return
		}
	}

	// Parse webhook payload
	webhookData, err := parseFunc(bodyBytes)
	if err != nil {
		// Mark webhook as failed
		h.historyService.MarkFailed(c.Request.Context(), headers.EventID, err)
		utils.BadRequestResponse(c, "Invalid webhook payload", err)
		return
	}

	// Process webhook
	poID, err := processFunc(c.Request.Context(), webhookData)
	if err != nil {
		// Mark webhook as failed
		h.historyService.MarkFailed(c.Request.Context(), headers.EventID, err)
		utils.HandleServiceError(c, "Failed to process webhook", err)
		return
	}

	// Mark webhook as successfully processed
	if err := h.historyService.MarkProcessed(c.Request.Context(), headers.EventID, poID); err != nil {
		utils.Error("Failed to mark webhook as processed:", err)
		// Don't fail the request - webhook was processed successfully
	}

	// Return success response
	utils.OKResponse(c, "Webhook processed successfully", gin.H{
		"event_id":          headers.EventID,
		"event_type":        eventType,
		"purchase_order_id": poID,
		"processed_at":      time.Now(),
	})
}

// ========================================
// 1. Order Created Webhook Handler
// ========================================

// HandleOrderCreated handles POST /webhooks/ecommerce/order/created
// @Summary E-commerce Order Created Webhook
// @Description Receives order.created webhook from e-commerce platform
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param X-Webhook-Signature header string true "HMAC-SHA256 signature of request body"
// @Param X-Webhook-Event-ID header string true "Unique event identifier (for idempotency)"
// @Param X-Webhook-Timestamp header string true "Unix timestamp of webhook creation"
// @Param request body models.OrderCreatedWebhook true "Order created webhook payload"
// @Success 200 {object} utils.Response{data=map[string]interface{}} "Webhook processed successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request or invalid payload"
// @Failure 401 {object} utils.ErrorResponseModel "Invalid webhook signature"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/webhooks/ecommerce/order/created [post]
func (h *EcommerceWebhookHandler) HandleOrderCreated(c *gin.Context) {
	h.processWebhook(c, "order.created",
		// Parse function
		func(bodyBytes []byte) (interface{}, error) {
			var webhook models.OrderCreatedWebhook
			if err := json.Unmarshal(bodyBytes, &webhook); err != nil {
				return nil, err
			}
			return &webhook, nil
		},
		// Process function
		func(ctx context.Context, data interface{}) (string, error) {
			webhook := data.(*models.OrderCreatedWebhook)
			return h.webhookService.ProcessOrderCreated(ctx, webhook)
		},
	)
}

// ========================================
// 2. Order Confirmed Webhook Handler
// ========================================

// HandleOrderConfirmed handles POST /webhooks/ecommerce/order/confirmed
// @Summary E-commerce Order Confirmed Webhook
// @Description Receives order.confirmed webhook from e-commerce platform
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param X-Webhook-Signature header string true "HMAC-SHA256 signature of request body"
// @Param X-Webhook-Event-ID header string true "Unique event identifier (for idempotency)"
// @Param X-Webhook-Timestamp header string true "Unix timestamp of webhook creation"
// @Param request body models.OrderConfirmedWebhook true "Order confirmed webhook payload"
// @Success 200 {object} utils.Response{data=map[string]interface{}} "Webhook processed successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request or invalid payload"
// @Failure 401 {object} utils.ErrorResponseModel "Invalid webhook signature"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/webhooks/ecommerce/order/confirmed [post]
func (h *EcommerceWebhookHandler) HandleOrderConfirmed(c *gin.Context) {
	h.processWebhook(c, "order.confirmed",
		// Parse function
		func(bodyBytes []byte) (interface{}, error) {
			var webhook models.OrderConfirmedWebhook
			if err := json.Unmarshal(bodyBytes, &webhook); err != nil {
				return nil, err
			}
			return &webhook, nil
		},
		// Process function - returns empty string for non-PO-creating webhooks
		func(ctx context.Context, data interface{}) (string, error) {
			webhook := data.(*models.OrderConfirmedWebhook)
			err := h.webhookService.ProcessOrderConfirmed(ctx, webhook)
			return "", err // No new PO created, just status update
		},
	)
}

// ========================================
// 3. Order Shipped Webhook Handler
// ========================================

// HandleOrderShipped handles POST /webhooks/ecommerce/order/shipped
// @Summary E-commerce Order Shipped Webhook
// @Description Receives order.shipped webhook from e-commerce platform
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param X-Webhook-Signature header string true "HMAC-SHA256 signature of request body"
// @Param X-Webhook-Event-ID header string true "Unique event identifier (for idempotency)"
// @Param X-Webhook-Timestamp header string true "Unix timestamp of webhook creation"
// @Param request body models.OrderShippedWebhook true "Order shipped webhook payload"
// @Success 200 {object} utils.Response{data=map[string]interface{}} "Webhook processed successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request or invalid payload"
// @Failure 401 {object} utils.ErrorResponseModel "Invalid webhook signature"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/webhooks/ecommerce/order/shipped [post]
func (h *EcommerceWebhookHandler) HandleOrderShipped(c *gin.Context) {
	h.processWebhook(c, "order.shipped",
		// Parse function
		func(bodyBytes []byte) (interface{}, error) {
			var webhook models.OrderShippedWebhook
			if err := json.Unmarshal(bodyBytes, &webhook); err != nil {
				return nil, err
			}
			return &webhook, nil
		},
		// Process function
		func(ctx context.Context, data interface{}) (string, error) {
			webhook := data.(*models.OrderShippedWebhook)
			err := h.webhookService.ProcessOrderShipped(ctx, webhook)
			return "", err // No new PO created, just status update
		},
	)
}

// ========================================
// 4. Order Delivered Webhook Handler
// ========================================

// HandleOrderDelivered handles POST /webhooks/ecommerce/order/delivered
// @Summary E-commerce Order Delivered Webhook
// @Description Receives order.delivered webhook from e-commerce platform
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param X-Webhook-Signature header string true "HMAC-SHA256 signature of request body"
// @Param X-Webhook-Event-ID header string true "Unique event identifier (for idempotency)"
// @Param X-Webhook-Timestamp header string true "Unix timestamp of webhook creation"
// @Param request body models.OrderDeliveredWebhook true "Order delivered webhook payload"
// @Success 200 {object} utils.Response{data=map[string]interface{}} "Webhook processed successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request or invalid payload"
// @Failure 401 {object} utils.ErrorResponseModel "Invalid webhook signature"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/webhooks/ecommerce/order/delivered [post]
func (h *EcommerceWebhookHandler) HandleOrderDelivered(c *gin.Context) {
	h.processWebhook(c, "order.delivered",
		// Parse function
		func(bodyBytes []byte) (interface{}, error) {
			var webhook models.OrderDeliveredWebhook
			if err := json.Unmarshal(bodyBytes, &webhook); err != nil {
				return nil, err
			}
			return &webhook, nil
		},
		// Process function
		func(ctx context.Context, data interface{}) (string, error) {
			webhook := data.(*models.OrderDeliveredWebhook)
			err := h.webhookService.ProcessOrderDelivered(ctx, webhook)
			return "", err // No new PO created, creates GRN instead
		},
	)
}

// ========================================
// 5. Order Payment Webhook Handler
// ========================================

// HandleOrderPayment handles POST /webhooks/ecommerce/order/payment
// @Summary E-commerce Order Payment Webhook
// @Description Receives order.payment webhook from e-commerce platform
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param X-Webhook-Signature header string true "HMAC-SHA256 signature of request body"
// @Param X-Webhook-Event-ID header string true "Unique event identifier (for idempotency)"
// @Param X-Webhook-Timestamp header string true "Unix timestamp of webhook creation"
// @Param request body models.OrderPaymentWebhook true "Order payment webhook payload"
// @Success 200 {object} utils.Response{data=map[string]interface{}} "Webhook processed successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request or invalid payload"
// @Failure 401 {object} utils.ErrorResponseModel "Invalid webhook signature"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/webhooks/ecommerce/order/payment [post]
func (h *EcommerceWebhookHandler) HandleOrderPayment(c *gin.Context) {
	h.processWebhook(c, "order.payment",
		// Parse function
		func(bodyBytes []byte) (interface{}, error) {
			var webhook models.OrderPaymentWebhook
			if err := json.Unmarshal(bodyBytes, &webhook); err != nil {
				return nil, err
			}
			return &webhook, nil
		},
		// Process function
		func(ctx context.Context, data interface{}) (string, error) {
			webhook := data.(*models.OrderPaymentWebhook)
			err := h.webhookService.ProcessOrderPayment(ctx, webhook)
			return "", err // No new PO created, just payment update
		},
	)
}

// ========================================
// 6. Webhook Status and History Endpoints
// ========================================

// GetWebhookHistory handles GET /webhooks/history
// @Summary Get Webhook History
// @Description Retrieve webhook processing history (requires authentication)
// @Tags Webhooks
// @Produce json
// @Param external_order_id query string false "Filter by external order ID"
// @Param status query string false "Filter by status (processing, success, failed)"
// @Param limit query int false "Limit results" default(50)
// @Success 200 {object} utils.Response{data=[]models.WebhookEvent} "Webhook history retrieved"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/webhooks/history [get]
func (h *EcommerceWebhookHandler) GetWebhookHistory(c *gin.Context) {
	// Get query parameters
	externalOrderID := c.Query("external_order_id")
	status := c.Query("status")
	limit := 50
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := utils.ParseInt(limitParam); err == nil {
			limit = parsedLimit
		}
	}

	var events []models.WebhookEvent
	var err error

	// Query based on filters
	if externalOrderID != "" {
		events, err = h.webhookRepo.FindByExternalOrderID(c.Request.Context(), externalOrderID)
	} else if status != "" {
		events, err = h.webhookRepo.FindByStatus(c.Request.Context(), status, limit)
	} else {
		events, err = h.webhookRepo.FindRecent(c.Request.Context(), limit)
	}

	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve webhook history", err)
		return
	}

	utils.OKResponse(c, "Webhook history retrieved successfully", events)
}

// GetWebhookStats handles GET /webhooks/stats
// @Summary Get Webhook Statistics
// @Description Retrieve webhook processing statistics (requires authentication)
// @Tags Webhooks
// @Produce json
// @Success 200 {object} utils.Response{data=map[string]interface{}} "Webhook statistics"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/webhooks/stats [get]
func (h *EcommerceWebhookHandler) GetWebhookStats(c *gin.Context) {
	stats, err := h.historyService.GetWebhookStats(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve webhook stats", err)
		return
	}

	utils.OKResponse(c, "Webhook statistics retrieved successfully", stats)
}

// ========================================
// Route Registration
// ========================================

// RegisterRoutes registers all webhook routes
func (h *EcommerceWebhookHandler) RegisterRoutes(router *gin.RouterGroup) {
	// E-commerce webhook routes (no authentication - uses HMAC signature verification)
	webhooks := router.Group("/webhooks")
	{
		// Order lifecycle webhooks
		webhooks.POST("/ecommerce/order/created", h.HandleOrderCreated)
		webhooks.POST("/ecommerce/order/confirmed", h.HandleOrderConfirmed)
		webhooks.POST("/ecommerce/order/shipped", h.HandleOrderShipped)
		webhooks.POST("/ecommerce/order/delivered", h.HandleOrderDelivered)
		webhooks.POST("/ecommerce/order/payment", h.HandleOrderPayment)

		// Webhook management endpoints (requires authentication)
		webhooks.GET("/history", h.aaaMiddleware.Authenticate(), h.GetWebhookHistory)
		webhooks.GET("/stats", h.aaaMiddleware.Authenticate(), h.GetWebhookStats)
	}
}
