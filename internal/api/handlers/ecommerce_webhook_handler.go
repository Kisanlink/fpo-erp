package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	logger "kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EcommerceWebhookHandler handles incoming webhooks from e-commerce platform
type EcommerceWebhookHandler struct {
	webhookService  interfaces.EcommerceWebhookServiceInterface
	securityService *services.WebhookSecurityService
	historyService  *services.WebhookHistoryService
	webhookRepo     *repositories.WebhookRepository
	aaaMiddleware   interface{ Authenticate() gin.HandlerFunc }
	logger          logger.Logger
}

// NewEcommerceWebhookHandler creates a new e-commerce webhook handler
func NewEcommerceWebhookHandler(
	webhookService interfaces.EcommerceWebhookServiceInterface,
	securityService *services.WebhookSecurityService,
	historyService *services.WebhookHistoryService,
	webhookRepo *repositories.WebhookRepository,
	aaaMiddleware interface{ Authenticate() gin.HandlerFunc },
	logger logger.Logger,
) *EcommerceWebhookHandler {
	return &EcommerceWebhookHandler{
		webhookService:  webhookService,
		securityService: securityService,
		historyService:  historyService,
		logger:          logger,
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
	// 1. Entry Log
	h.logger.Info("Handling webhook request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("event_type", eventType))

	// Read raw request body for signature verification
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read webhook request body",
			zap.Error(err),
			zap.String("event_type", eventType))
		utils.BadRequestResponse(c, "Failed to read request body", err)
		return
	}

	// Verify webhook signature and headers
	h.logger.Debug("Validating webhook signature and headers",
		zap.String("event_type", eventType))

	headers, err := h.securityService.ValidateWebhook(c, bodyBytes)
	if err != nil {
		h.logger.Error("Webhook validation failed",
			zap.Error(err),
			zap.String("event_type", eventType))
		utils.UnauthorizedResponse(c, "Webhook validation failed: "+err.Error())
		return
	}

	h.logger.Debug("Webhook signature validated successfully",
		zap.String("event_type", eventType),
		zap.String("event_id", headers.EventID))

	// Check idempotency (duplicate event detection)
	h.logger.Debug("Checking webhook idempotency",
		zap.String("event_type", eventType),
		zap.String("event_id", headers.EventID))

	alreadyProcessed, existingEvent, err := h.historyService.CheckIdempotency(c.Request.Context(), headers.EventID)
	if err != nil {
		h.logger.Error("Failed to check webhook idempotency",
			zap.Error(err),
			zap.String("event_type", eventType),
			zap.String("event_id", headers.EventID))
		utils.HandleServiceError(c, "Failed to check webhook idempotency", err)
		return
	}

	if alreadyProcessed {
		h.logger.Info("Webhook already processed (idempotent)",
			zap.String("event_type", eventType),
			zap.String("event_id", headers.EventID),
			zap.String("status", existingEvent.Status))
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
	h.logger.Debug("Parsing webhook payload",
		zap.String("event_type", eventType),
		zap.String("event_id", headers.EventID))

	webhookData, err := parseFunc(bodyBytes)
	if err != nil {
		h.logger.Error("Invalid webhook payload",
			zap.Error(err),
			zap.String("event_type", eventType),
			zap.String("event_id", headers.EventID))
		// Mark webhook as failed
		h.historyService.MarkFailed(c.Request.Context(), headers.EventID, err)
		utils.BadRequestResponse(c, "Invalid webhook payload", err)
		return
	}

	// Process webhook
	h.logger.Debug("Processing webhook via service",
		zap.String("event_type", eventType),
		zap.String("event_id", headers.EventID))

	poID, err := processFunc(c.Request.Context(), webhookData)
	if err != nil {
		h.logger.Error("Failed to process webhook via service",
			zap.Error(err),
			zap.String("event_type", eventType),
			zap.String("event_id", headers.EventID))
		// Mark webhook as failed
		h.historyService.MarkFailed(c.Request.Context(), headers.EventID, err)
		utils.HandleServiceError(c, "Failed to process webhook", err)
		return
	}

	// Mark webhook as successfully processed
	if err := h.historyService.MarkProcessed(c.Request.Context(), headers.EventID, poID); err != nil {
		h.logger.Error("Failed to mark webhook as processed",
			zap.Error(err),
			zap.String("event_type", eventType),
			zap.String("event_id", headers.EventID))
		utils.Error("Failed to mark webhook as processed:", err)
		// Don't fail the request - webhook was processed successfully
	}

	// Return success response
	h.logger.Info("Webhook processed successfully via handler",
		zap.String("event_type", eventType),
		zap.String("event_id", headers.EventID),
		zap.String("purchase_order_id", poID))
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
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/webhooks/history [get]
func (h *EcommerceWebhookHandler) GetWebhookHistory(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get webhook history request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

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

	// 3. Service Call
	h.logger.Debug("Querying webhook history",
		zap.String("external_order_id", externalOrderID),
		zap.String("status", status),
		zap.Int("limit", limit))

	// Query based on filters
	if externalOrderID != "" {
		events, err = h.webhookRepo.FindByExternalOrderID(c.Request.Context(), externalOrderID)
	} else if status != "" {
		events, err = h.webhookRepo.FindByStatus(c.Request.Context(), status, limit)
	} else {
		events, err = h.webhookRepo.FindRecent(c.Request.Context(), limit)
	}

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to retrieve webhook history",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve webhook history", err)
		return
	}

	// 5. Success
	h.logger.Info("Webhook history retrieved successfully via handler",
		zap.Int("count", len(events)))
	utils.OKResponse(c, "Webhook history retrieved successfully", events)
}

// GetWebhookStats handles GET /webhooks/stats
// @Summary Get Webhook Statistics
// @Description Retrieve webhook processing statistics (requires authentication)
// @Tags Webhooks
// @Produce json
// @Success 200 {object} utils.Response{data=map[string]interface{}} "Webhook statistics"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/webhooks/stats [get]
func (h *EcommerceWebhookHandler) GetWebhookStats(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get webhook stats request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// 3. Service Call
	h.logger.Debug("Calling service to get webhook stats")

	stats, err := h.historyService.GetWebhookStats(c.Request.Context())

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to retrieve webhook stats via service",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve webhook stats", err)
		return
	}

	// 5. Success
	h.logger.Info("Webhook statistics retrieved successfully via handler")
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
