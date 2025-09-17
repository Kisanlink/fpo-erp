package handlers

import (
	"strconv"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

// WebhookConfigHandler handles webhook configuration HTTP requests
type WebhookConfigHandler struct {
	webhookConfigService *services.WebhookConfigService
	aaaMiddleware        *aaa.AAAMiddleware
}

// NewWebhookConfigHandler creates a new webhook configuration handler
func NewWebhookConfigHandler(
	webhookConfigService *services.WebhookConfigService,
	aaaMiddleware *aaa.AAAMiddleware,
) *WebhookConfigHandler {
	return &WebhookConfigHandler{
		webhookConfigService: webhookConfigService,
		aaaMiddleware:        aaaMiddleware,
	}
}

// CreateConfig handles POST /api/v1/webhook-configs
// @Summary Create Webhook Configuration
// @Description Create a new webhook configuration for an FPO (requires authentication)
// @Tags Webhook Configuration
// @Accept json
// @Produce json
// @Param request body models.CreateWebhookConfigRequest true "Webhook configuration data"
// @Success 201 {object} utils.Response{data=models.WebhookConfigurationResponse} "Configuration created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 409 {object} utils.ErrorResponseModel "Configuration already exists for FPO"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/webhook-configs [post]
func (h *WebhookConfigHandler) CreateConfig(c *gin.Context) {
	var req models.CreateWebhookConfigRequest

	if err := utils.ValidateRequest(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	config, err := h.webhookConfigService.CreateConfig(&req)
	if err != nil {
		if err.Error() == "Configuration already exists for this FPO" {
			utils.ErrorResponse(c, 409, "Configuration already exists for FPO", err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to create webhook configuration", err)
		return
	}

	utils.CreatedResponse(c, "Webhook configuration created successfully", config)
}

// GetConfig handles GET /api/v1/webhook-configs/:id
// @Summary Get Webhook Configuration
// @Description Retrieve a webhook configuration by ID (requires authentication)
// @Tags Webhook Configuration
// @Produce json
// @Param id path string true "Configuration ID" example(WHCF_12345678)
// @Success 200 {object} utils.Response{data=models.WebhookConfigurationResponse} "Configuration retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Configuration not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/webhook-configs/{id} [get]
func (h *WebhookConfigHandler) GetConfig(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Configuration ID is required", nil)
		return
	}

	config, err := h.webhookConfigService.GetConfig(id)
	if err != nil {
		utils.NotFoundResponse(c, "Webhook configuration not found")
		return
	}

	utils.OKResponse(c, "Webhook configuration retrieved successfully", config)
}

// GetConfigByFPO handles GET /api/v1/webhook-configs/fpo/:fpo_id
// @Summary Get Webhook Configuration by FPO
// @Description Retrieve a webhook configuration by FPO ID (requires authentication)
// @Tags Webhook Configuration
// @Produce json
// @Param fpo_id path string true "FPO ID" example(fpo_123456789)
// @Success 200 {object} utils.Response{data=models.WebhookConfigurationResponse} "Configuration retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Configuration not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/webhook-configs/fpo/{fpo_id} [get]
func (h *WebhookConfigHandler) GetConfigByFPO(c *gin.Context) {
	fpoID := c.Param("fpo_id")
	if fpoID == "" {
		utils.BadRequestResponse(c, "FPO ID is required", nil)
		return
	}

	config, err := h.webhookConfigService.GetConfigByFPO(fpoID)
	if err != nil {
		utils.NotFoundResponse(c, "Webhook configuration not found for FPO")
		return
	}

	utils.OKResponse(c, "Webhook configuration retrieved successfully", config)
}

// UpdateConfig handles PUT /api/v1/webhook-configs/:id
// @Summary Update Webhook Configuration
// @Description Update an existing webhook configuration (requires authentication)
// @Tags Webhook Configuration
// @Accept json
// @Produce json
// @Param id path string true "Configuration ID" example(WHCF_12345678)
// @Param request body models.UpdateWebhookConfigRequest true "Updated configuration data"
// @Success 200 {object} utils.Response{data=models.WebhookConfigurationResponse} "Configuration updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Configuration not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/webhook-configs/{id} [put]
func (h *WebhookConfigHandler) UpdateConfig(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Configuration ID is required", nil)
		return
	}

	var req models.UpdateWebhookConfigRequest
	if err := utils.ValidatePartialRequest(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	config, err := h.webhookConfigService.UpdateConfig(id, &req)
	if err != nil {
		if err.Error() == "Webhook configuration not found" {
			utils.NotFoundResponse(c, "Webhook configuration not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update webhook configuration", err)
		return
	}

	utils.OKResponse(c, "Webhook configuration updated successfully", config)
}

// DeleteConfig handles DELETE /api/v1/webhook-configs/:id
// @Summary Delete Webhook Configuration
// @Description Delete a webhook configuration (requires authentication)
// @Tags Webhook Configuration
// @Produce json
// @Param id path string true "Configuration ID" example(WHCF_12345678)
// @Success 200 {object} utils.Response "Configuration deleted successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Configuration not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/webhook-configs/{id} [delete]
func (h *WebhookConfigHandler) DeleteConfig(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Configuration ID is required", nil)
		return
	}

	err := h.webhookConfigService.DeleteConfig(id)
	if err != nil {
		if err.Error() == "Webhook configuration not found" {
			utils.NotFoundResponse(c, "Webhook configuration not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to delete webhook configuration", err)
		return
	}

	utils.OKResponse(c, "Webhook configuration deleted successfully", nil)
}

// GetAllConfigs handles GET /api/v1/webhook-configs
// @Summary Get All Webhook Configurations
// @Description Retrieve all webhook configurations with pagination (requires authentication)
// @Tags Webhook Configuration
// @Produce json
// @Param limit query integer false "Number of records to return (default: 10)" example(10)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.Response{data=[]models.WebhookConfigurationResponse} "Configurations retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/webhook-configs [get]
func (h *WebhookConfigHandler) GetAllConfigs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid limit parameter", err)
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid offset parameter", err)
		return
	}

	configs, err := h.webhookConfigService.GetAllConfigs(limit, offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve webhook configurations", err)
		return
	}

	utils.OKResponse(c, "Webhook configurations retrieved successfully", configs)
}

// ToggleConfig handles PATCH /api/v1/webhook-configs/:id/toggle
// @Summary Toggle Webhook Configuration
// @Description Enable or disable a webhook configuration (requires authentication)
// @Tags Webhook Configuration
// @Produce json
// @Param id path string true "Configuration ID" example(WHCF_12345678)
// @Success 200 {object} utils.Response{data=models.WebhookConfigurationResponse} "Configuration toggled successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Configuration not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/webhook-configs/{id}/toggle [patch]
func (h *WebhookConfigHandler) ToggleConfig(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Configuration ID is required", nil)
		return
	}

	config, err := h.webhookConfigService.ToggleConfig(id)
	if err != nil {
		if err.Error() == "Webhook configuration not found" {
			utils.NotFoundResponse(c, "Webhook configuration not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to toggle webhook configuration", err)
		return
	}

	status := "disabled"
	if config.Enabled {
		status = "enabled"
	}

	utils.OKResponse(c, "Webhook configuration "+status+" successfully", config)
}

// GetConfigHistory handles GET /api/v1/webhook-configs/:id/history
// @Summary Get Configuration History
// @Description Retrieve delivery history for a webhook configuration (requires authentication)
// @Tags Webhook Configuration
// @Produce json
// @Param id path string true "Configuration ID" example(WHCF_12345678)
// @Param limit query integer false "Number of records to return (default: 50)" example(50)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.Response{data=[]models.WebhookHistoryResponse} "Configuration history retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Configuration not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/webhook-configs/{id}/history [get]
func (h *WebhookConfigHandler) GetConfigHistory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Configuration ID is required", nil)
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

	history, err := h.webhookConfigService.GetConfigHistory(id, limit, offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve configuration history", err)
		return
	}

	utils.OKResponse(c, "Configuration history retrieved successfully", history)
}

// RegisterRoutes registers all webhook configuration routes
func (h *WebhookConfigHandler) RegisterRoutes(router *gin.RouterGroup) {
	configs := router.Group("/webhook-configs")
	{
		// Apply authentication middleware
		configs.Use(h.aaaMiddleware.Authenticate())

		// Create/Update/Delete routes - CEO and Tech_Support only
		configs.POST("", h.aaaMiddleware.RequirePermission("aaa/webhook_config", "*", "create"), h.CreateConfig)
		configs.PUT("/:id", h.aaaMiddleware.RequirePermission("aaa/webhook_config", "*", "update"), h.UpdateConfig)
		configs.DELETE("/:id", h.aaaMiddleware.RequirePermission("aaa/webhook_config", "*", "delete"), h.DeleteConfig)
		configs.PATCH("/:id/toggle", h.aaaMiddleware.RequirePermission("aaa/webhook_config", "*", "update"), h.ToggleConfig)

		// Read routes - Operations team can view configurations
		configs.GET("", h.aaaMiddleware.RequirePermission("aaa/webhook_config", "*", "read"), h.GetAllConfigs)
		configs.GET("/:id", h.aaaMiddleware.RequirePermission("aaa/webhook_config", "*", "read"), h.GetConfig)
		configs.GET("/fpo/:fpo_id", h.aaaMiddleware.RequirePermission("aaa/webhook_config", "*", "read"), h.GetConfigByFPO)
		configs.GET("/:id/history", h.aaaMiddleware.RequirePermission("aaa/webhook_config", "*", "read"), h.GetConfigHistory)
	}
}