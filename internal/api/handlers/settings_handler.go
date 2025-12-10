package handlers

import (
	"kisanlink-erp/internal/database/models"
	logger "kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SettingsHandler handles settings HTTP requests
type SettingsHandler struct {
	settingsService interfaces.SettingsServiceInterface
	logger          logger.Logger
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(settingsService interfaces.SettingsServiceInterface, logger logger.Logger) *SettingsHandler {
	return &SettingsHandler{
		settingsService: settingsService,
		logger:          logger,
	}
}

// GetAllSettings handles GET /api/v1/settings
// @Summary Get All Settings
// @Description Retrieve all FPO settings
// @Tags Settings
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.SettingResponse} "Settings retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/settings [get]
func (h *SettingsHandler) GetAllSettings(c *gin.Context) {
	h.logger.Info("Handling get all settings request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get all settings
	response, err := h.settingsService.GetAllSettings(c.Request.Context())
	if err != nil {
		h.logger.Error("Service error retrieving all settings",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve settings", err)
		return
	}

	h.logger.Info("All settings retrieved successfully",
		zap.Int("settings_count", len(response)))

	utils.OKResponse(c, "Settings retrieved successfully", response)
}

// GetSetting handles GET /api/v1/settings/:key
// @Summary Get Setting
// @Description Retrieve a specific setting by key
// @Tags Settings
// @Produce json
// @Param key path string true "Setting key" example(fpo_name)
// @Success 200 {object} utils.Response{data=models.SettingResponse} "Setting retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Setting not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/settings/{key} [get]
func (h *SettingsHandler) GetSetting(c *gin.Context) {
	h.logger.Info("Handling get setting request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get key from URL
	key := c.Param("key")
	if key == "" {
		h.logger.Error("Setting key is required but not provided")
		utils.BadRequestResponse(c, "Setting key is required", nil)
		return
	}

	h.logger.Debug("Fetching setting by key",
		zap.String("key", key))

	// Get setting
	response, err := h.settingsService.GetSetting(c.Request.Context(), key)
	if err != nil {
		h.logger.Error("Setting not found",
			zap.Error(err),
			zap.String("key", key))
		utils.HandleServiceError(c, "Failed to retrieve setting", err)
		return
	}

	h.logger.Info("Setting retrieved successfully",
		zap.String("key", response.Key))

	utils.OKResponse(c, "Setting retrieved successfully", response)
}

// UpsertSetting handles PUT /api/v1/settings/:key
// @Summary Create or Update Setting
// @Description Create a new setting or update an existing one by key
// @Tags Settings
// @Accept json
// @Produce json
// @Param key path string true "Setting key" example(fpo_name)
// @Param request body models.CreateSettingRequest true "Setting data"
// @Success 200 {object} utils.Response{data=models.SettingResponse} "Setting saved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/settings/{key} [put]
func (h *SettingsHandler) UpsertSetting(c *gin.Context) {
	h.logger.Info("Handling upsert setting request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get key from URL
	key := c.Param("key")
	if key == "" {
		h.logger.Error("Setting key is required but not provided")
		utils.BadRequestResponse(c, "Setting key is required", nil)
		return
	}

	var request models.CreateSettingRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		h.logger.Error("Invalid request body for upsert setting",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	h.logger.Debug("Calling service to upsert setting",
		zap.String("key", key))

	// Upsert setting
	response, err := h.settingsService.UpsertSetting(c.Request.Context(), key, &request)
	if err != nil {
		h.logger.Error("Service error upserting setting",
			zap.Error(err),
			zap.String("key", key))
		utils.HandleServiceError(c, "Failed to save setting", err)
		return
	}

	h.logger.Info("Setting saved successfully",
		zap.String("key", response.Key))

	utils.OKResponse(c, "Setting saved successfully", response)
}

// DeleteSetting handles DELETE /api/v1/settings/:key
// @Summary Delete Setting
// @Description Delete a setting by key
// @Tags Settings
// @Produce json
// @Param key path string true "Setting key" example(fpo_name)
// @Success 200 {object} utils.Response "Setting deleted successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Setting not found"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/settings/{key} [delete]
func (h *SettingsHandler) DeleteSetting(c *gin.Context) {
	h.logger.Info("Handling delete setting request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get key from URL
	key := c.Param("key")
	if key == "" {
		h.logger.Error("Setting key is required but not provided")
		utils.BadRequestResponse(c, "Setting key is required", nil)
		return
	}

	h.logger.Debug("Calling service to delete setting",
		zap.String("key", key))

	// Delete setting
	if err := h.settingsService.DeleteSetting(c.Request.Context(), key); err != nil {
		h.logger.Error("Service error deleting setting",
			zap.Error(err),
			zap.String("key", key))
		utils.HandleServiceError(c, "Failed to delete setting", err)
		return
	}

	h.logger.Info("Setting deleted successfully",
		zap.String("key", key))

	utils.OKResponse(c, "Setting deleted successfully", nil)
}

// GetHeaderFields handles GET /api/v1/settings/header-fields
// @Summary Get Header Fields
// @Description Retrieve settings configured as invoice header fields, ordered by display_order
// @Tags Settings
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.HeaderFieldResponse} "Header fields retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/settings/header-fields [get]
func (h *SettingsHandler) GetHeaderFields(c *gin.Context) {
	h.logger.Info("Handling get header fields request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get header fields
	response, err := h.settingsService.GetHeaderFields(c.Request.Context())
	if err != nil {
		h.logger.Error("Service error retrieving header fields",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve header fields", err)
		return
	}

	h.logger.Info("Header fields retrieved successfully",
		zap.Int("field_count", len(response)))

	utils.OKResponse(c, "Header fields retrieved successfully", response)
}

// CheckInvoiceRequirements handles GET /api/v1/settings/invoice-requirements
// @Summary Check Invoice Requirements
// @Description Check if all required settings for invoice generation exist
// @Tags Settings
// @Produce json
// @Success 200 {object} utils.Response{data=models.InvoiceRequirementsResponse} "Invoice requirements check complete"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/settings/invoice-requirements [get]
func (h *SettingsHandler) CheckInvoiceRequirements(c *gin.Context) {
	h.logger.Info("Handling check invoice requirements request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Check invoice requirements
	ready, missing, err := h.settingsService.CheckInvoiceRequirements(c.Request.Context())
	if err != nil {
		h.logger.Error("Service error checking invoice requirements",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to check invoice requirements", err)
		return
	}

	response := models.InvoiceRequirementsResponse{
		Ready:           ready,
		MissingSettings: missing,
	}

	if ready {
		h.logger.Info("All invoice requirements satisfied")
		utils.OKResponse(c, "All invoice requirements satisfied", response)
	} else {
		h.logger.Warn("Missing required settings for invoice",
			zap.Strings("missing", missing))
		utils.OKResponse(c, "Missing required settings for invoice generation", response)
	}
}

// RegisterRoutes registers all settings routes
func (h *SettingsHandler) RegisterRoutes(v1 *gin.RouterGroup) {
	settings := v1.Group("/settings")
	{
		// Read operations
		settings.GET("", h.GetAllSettings)
		settings.GET("/header-fields", h.GetHeaderFields)
		settings.GET("/invoice-requirements", h.CheckInvoiceRequirements)
		settings.GET("/:key", h.GetSetting)

		// Write operations (requires authentication)
		settings.PUT("/:key", h.UpsertSetting)
		settings.DELETE("/:key", h.DeleteSetting)
	}
}
