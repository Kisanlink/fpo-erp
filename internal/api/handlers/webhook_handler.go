package handlers

import (
	"encoding/json"
	"io"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

// WebhookHandler handles webhook-related HTTP requests
type WebhookHandler struct {
	webhookSecurityService   *services.WebhookSecurityService
	ecommerceWebhookService  *services.EcommerceWebhookService
	webhookHistoryService    *services.WebhookHistoryService
	webhookRepo              *repositories.WebhookRepository
	aaaMiddleware            *aaa.AAAMiddleware
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(
	webhookSecurityService *services.WebhookSecurityService,
	ecommerceWebhookService *services.EcommerceWebhookService,
	webhookHistoryService *services.WebhookHistoryService,
	webhookRepo *repositories.WebhookRepository,
	aaaMiddleware *aaa.AAAMiddleware,
) *WebhookHandler {
	return &WebhookHandler{
		webhookSecurityService:   webhookSecurityService,
		ecommerceWebhookService:  ecommerceWebhookService,
		webhookHistoryService:    webhookHistoryService,
		webhookRepo:              webhookRepo,
		aaaMiddleware:            aaaMiddleware,
	}
}

// ReceiveEcommerceWebhook handles POST /api/v1/ecommerce/webhook
// @Summary Receive E-commerce Webhook
// @Description Receive and process webhooks from e-commerce platforms
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param X-Kisanlink-Signature header string true "HMAC-SHA256 signature"
// @Param X-Kisanlink-Timestamp header string true "Unix timestamp"
// @Param request body object true "Webhook payload"
// @Success 200 {object} utils.Response "Webhook processed successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request or validation error"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized - signature validation failed"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - event already processed"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable entity - business logic error"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/ecommerce/webhook [post]
func (h *WebhookHandler) ReceiveEcommerceWebhook(c *gin.Context) {
	// Get headers
	signature := c.GetHeader("X-Kisanlink-Signature")
	timestamp := c.GetHeader("X-Kisanlink-Timestamp")

	// Read raw body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to read request body", err)
		return
	}

	bodyString := string(body)

	// Parse the payload to get FPO ID for webhook secret lookup
	var basicPayload struct {
		FPOID     string `json:"fpo_id"`
		EventType string `json:"event_type"`
		EventID   string `json:"event_id"`
	}

	if err := json.Unmarshal(body, &basicPayload); err != nil {
		utils.BadRequestResponse(c, "Invalid JSON payload", err)
		return
	}

	// Validate required fields
	if basicPayload.FPOID == "" {
		utils.BadRequestResponse(c, "fpo_id is required", nil)
		return
	}

	if basicPayload.EventType == "" {
		utils.BadRequestResponse(c, "event_type is required", nil)
		return
	}

	if basicPayload.EventID == "" {
		utils.BadRequestResponse(c, "event_id is required", nil)
		return
	}

	// Get webhook configuration for this FPO
	config, err := h.webhookRepo.GetConfigByFPO(basicPayload.FPOID)
	if err != nil {
		utils.BadRequestResponse(c, "No webhook configuration found for FPO", err)
		return
	}

	if !config.Enabled {
		utils.BadRequestResponse(c, "Webhook processing is disabled for this FPO", nil)
		return
	}

	// Validate webhook security (HMAC + timestamp)
	if err := h.webhookSecurityService.SecureWebhookValidation(bodyString, signature, timestamp, config.SecretKey); err != nil {
		utils.UnauthorizedResponse(c, "Webhook security validation failed: "+err.Error())
		return
	}

	utils.Info("Webhook security validation passed for event:", basicPayload.EventID)

	// Process based on event type
	switch basicPayload.EventType {
	case "fpo_sale":
		if err := h.processFPOSaleWebhook(body); err != nil {
			if err.Error() == "Event already processed" {
				utils.ErrorResponse(c, 409, "Event already processed", err)
				return
			}
			utils.ErrorResponse(c, 422, "Failed to process FPO sale event", err)
			return
		}

	case "fpo_purchase":
		if err := h.processFPOPurchaseWebhook(body); err != nil {
			if err.Error() == "Event already processed" {
				utils.ErrorResponse(c, 409, "Event already processed", err)
				return
			}
			utils.ErrorResponse(c, 422, "Failed to process FPO purchase event", err)
			return
		}

	default:
		utils.BadRequestResponse(c, "Unsupported event type: "+basicPayload.EventType, nil)
		return
	}

	utils.OKResponse(c, "Webhook processed successfully", gin.H{
		"event_id":   basicPayload.EventID,
		"event_type": basicPayload.EventType,
		"status":     "processed",
	})
}

// processFPOSaleWebhook processes FPO sale webhook events
func (h *WebhookHandler) processFPOSaleWebhook(body []byte) error {
	var payload services.FPOSalePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	// Validate payload structure
	if err := h.ecommerceWebhookService.ValidateFPOSalePayload(&payload); err != nil {
		return err
	}

	// Process the sale event
	return h.ecommerceWebhookService.ProcessFPOSaleEvent(&payload)
}

// processFPOPurchaseWebhook processes FPO purchase webhook events
func (h *WebhookHandler) processFPOPurchaseWebhook(body []byte) error {
	var payload services.FPOPurchasePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	// Validate payload structure
	if err := h.ecommerceWebhookService.ValidateFPOPurchasePayload(&payload); err != nil {
		return err
	}

	// Process the purchase event
	return h.ecommerceWebhookService.ProcessFPOPurchaseEvent(&payload)
}

// GetWebhookStats handles GET /api/v1/webhooks/stats
// @Summary Get Webhook Statistics
// @Description Retrieve webhook processing statistics (requires authentication)
// @Tags Webhooks
// @Produce json
// @Success 200 {object} utils.Response{data=object} "Webhook statistics retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/webhooks/stats [get]
func (h *WebhookHandler) GetWebhookStats(c *gin.Context) {
	stats, err := h.webhookHistoryService.GetWebhookStats()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve webhook statistics", err)
		return
	}

	utils.OKResponse(c, "Webhook statistics retrieved successfully", stats)
}

// GetFailedWebhooks handles GET /api/v1/webhooks/failed
// @Summary Get Failed Webhook Deliveries
// @Description Retrieve failed webhook deliveries for monitoring (requires authentication)
// @Tags Webhooks
// @Produce json
// @Param limit query integer false "Number of records to return (default: 50)" example(50)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.Response{data=[]models.WebhookHistoryResponse} "Failed webhooks retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/webhooks/failed [get]
func (h *WebhookHandler) GetFailedWebhooks(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := utils.ParseIntParam(l); err == nil {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := utils.ParseIntParam(o); err == nil {
			offset = parsed
		}
	}

	failed, err := h.webhookHistoryService.GetFailedDeliveries(limit, offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve failed webhooks", err)
		return
	}

	utils.OKResponse(c, "Failed webhooks retrieved successfully", failed)
}

// GetEventHistory handles GET /api/v1/webhooks/events/:event_id/history
// @Summary Get Event History
// @Description Retrieve delivery history for a specific event (requires authentication)
// @Tags Webhooks
// @Produce json
// @Param event_id path string true "Event ID" example(evt_123456789)
// @Success 200 {object} utils.Response{data=[]models.WebhookHistoryResponse} "Event history retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Event not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/webhooks/events/{event_id}/history [get]
func (h *WebhookHandler) GetEventHistory(c *gin.Context) {
	eventID := c.Param("event_id")
	if eventID == "" {
		utils.BadRequestResponse(c, "Event ID is required", nil)
		return
	}

	history, err := h.webhookHistoryService.GetEventHistory(eventID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve event history", err)
		return
	}

	utils.OKResponse(c, "Event history retrieved successfully", history)
}

// GetEventsByStatus handles GET /api/v1/webhooks/events/status/:status
// @Summary Get Events by Status
// @Description Retrieve webhook events by processing status (requires authentication)
// @Tags Webhooks
// @Produce json
// @Param status path string true "Processing status" example(pending)
// @Param limit query integer false "Number of records to return (default: 50)" example(50)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.Response{data=[]models.WebhookEventResponse} "Events retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/webhooks/events/status/{status} [get]
func (h *WebhookHandler) GetEventsByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		utils.BadRequestResponse(c, "Status is required", nil)
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := utils.ParseIntParam(l); err == nil {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := utils.ParseIntParam(o); err == nil {
			offset = parsed
		}
	}

	events, err := h.webhookHistoryService.GetEventsByStatus(status, limit, offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve events", err)
		return
	}

	utils.OKResponse(c, "Events retrieved successfully", events)
}

// RegisterRoutes registers all webhook routes
func (h *WebhookHandler) RegisterRoutes(router *gin.RouterGroup) {
	webhooks := router.Group("/webhooks")
	{
		// Public endpoint for receiving e-commerce webhooks (no auth required)
		// Security is handled via HMAC signature validation
		router.POST("/ecommerce/webhook", h.ReceiveEcommerceWebhook)

		// Protected routes for monitoring and management
		webhooks.Use(h.aaaMiddleware.Authenticate())

		// Monitoring routes - read access for operations team
		webhooks.GET("/stats", h.aaaMiddleware.RequirePermission("aaa/webhook", "*", "read"), h.GetWebhookStats)
		webhooks.GET("/failed", h.aaaMiddleware.RequirePermission("aaa/webhook", "*", "read"), h.GetFailedWebhooks)
		webhooks.GET("/events/:event_id/history", h.aaaMiddleware.RequirePermission("aaa/webhook", "*", "read"), h.GetEventHistory)
		webhooks.GET("/events/status/:status", h.aaaMiddleware.RequirePermission("aaa/webhook", "*", "read"), h.GetEventsByStatus)
	}
}