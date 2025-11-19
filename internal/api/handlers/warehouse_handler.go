package handlers

import (
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	logger "kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// WarehouseHandler handles warehouse HTTP requests
type WarehouseHandler struct {
	warehouseService interfaces.WarehouseServiceInterface
	aaaMiddleware    *aaa.AAAMiddleware
	logger           logger.Logger
}

// NewWarehouseHandler creates a new warehouse handler
func NewWarehouseHandler(warehouseService interfaces.WarehouseServiceInterface, aaaMiddleware *aaa.AAAMiddleware, logger logger.Logger) *WarehouseHandler {
	return &WarehouseHandler{
		warehouseService: warehouseService,
		aaaMiddleware:    aaaMiddleware,
		logger:           logger,
	}
}

// CreateWarehouse handles POST /api/v1/warehouses
// @Summary Create Warehouse
// @Description Create a new warehouse (requires authentication)
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param request body models.CreateWarehouseRequest true "Warehouse data"
// @Success 201 {object} utils.Response{data=models.WarehouseResponse} "Warehouse created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/warehouses [post]
func (h *WarehouseHandler) CreateWarehouse(c *gin.Context) {
	h.logger.Info("Handling create warehouse request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	var request models.CreateWarehouseRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		h.logger.Error("Invalid request body for create warehouse",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Get authenticated user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system" // Fallback for unauthenticated contexts
	}

	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		h.logger.Error("Missing authentication token for create warehouse")
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	h.logger.Debug("Calling service to create warehouse",
		zap.String("warehouse_name", request.Name),
		zap.String("user_id", userID))

	// Create warehouse
	response, err := h.warehouseService.CreateWarehouse(c.Request.Context(), &request, userID, jwtToken)
	if err != nil {
		h.logger.Error("Service error creating warehouse",
			zap.Error(err),
			zap.String("warehouse_name", request.Name))
		utils.HandleServiceError(c, "Failed to create warehouse", err)
		return
	}

	h.logger.Info("Warehouse created successfully",
		zap.String("warehouse_id", response.ID),
		zap.String("warehouse_name", response.Name))

	utils.CreatedResponse(c, "Warehouse created successfully", response)
}

// GetWarehouse handles GET /api/v1/warehouses/:id
// @Summary Get Warehouse
// @Description Retrieve a specific warehouse by ID
// @Tags Warehouses
// @Produce json
// @Param id path string true "Warehouse ID (format: WHSE_xxxxxxxx)" example(WHSE_12345678)
// @Success 200 {object} utils.Response{data=models.WarehouseResponse} "Warehouse details"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Warehouse not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/warehouses/{id} [get]
func (h *WarehouseHandler) GetWarehouse(c *gin.Context) {
	h.logger.Info("Handling get warehouse request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Warehouse ID is required but not provided")
		utils.BadRequestResponse(c, "Warehouse ID is required", nil)
		return
	}

	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		h.logger.Error("Missing authentication token for get warehouse")
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	h.logger.Debug("Fetching warehouse by ID",
		zap.String("warehouse_id", id))

	// Get warehouse
	response, err := h.warehouseService.GetWarehouse(c.Request.Context(), id, jwtToken)
	if err != nil {
		h.logger.Error("Warehouse not found",
			zap.Error(err),
			zap.String("warehouse_id", id))
		utils.NotFoundResponse(c, "Warehouse not found")
		return
	}

	h.logger.Info("Warehouse retrieved successfully",
		zap.String("warehouse_id", response.ID),
		zap.String("warehouse_name", response.Name))

	utils.OKResponse(c, "Warehouse retrieved successfully", response)
}

// GetAllWarehouses handles GET /api/v1/warehouses
// @Summary Get All Warehouses
// @Description Retrieve all warehouses (requires authentication)
// @Tags Warehouses
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.WarehouseResponse} "Warehouses retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/warehouses [get]
func (h *WarehouseHandler) GetAllWarehouses(c *gin.Context) {
	h.logger.Info("Handling get all warehouses request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		h.logger.Error("Missing authentication token for get all warehouses")
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	// Get all warehouses
	response, err := h.warehouseService.GetAllWarehouses(c.Request.Context(), jwtToken)
	if err != nil {
		h.logger.Error("Service error retrieving all warehouses",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve warehouses", err)
		return
	}

	h.logger.Info("All warehouses retrieved successfully",
		zap.Int("warehouse_count", len(response)))

	utils.OKResponse(c, "Warehouses retrieved successfully", response)
}

// UpdateWarehouse handles PATCH /api/v1/warehouses/:id
// @Summary Update Warehouse
// @Description Update an existing warehouse by ID (requires authentication)
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param id path string true "Warehouse ID (format: WHSE_xxxxxxxx)" example(WHSE_12345678)
// @Param request body models.UpdateWarehouseRequest true "Updated warehouse data"
// @Success 200 {object} utils.Response{data=models.WarehouseResponse} "Warehouse updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Warehouse not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/warehouses/{id} [patch]
func (h *WarehouseHandler) UpdateWarehouse(c *gin.Context) {
	h.logger.Info("Handling update warehouse request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Warehouse ID is required but not provided for update")
		utils.BadRequestResponse(c, "Warehouse ID is required", nil)
		return
	}

	var request models.UpdateWarehouseRequest

	// Validate request
	if err := utils.ValidatePartialRequest(c, &request); err != nil {
		h.logger.Error("Invalid request body for update warehouse",
			zap.Error(err),
			zap.String("warehouse_id", id))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		h.logger.Error("Missing authentication token for update warehouse")
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	h.logger.Debug("Calling service to update warehouse",
		zap.String("warehouse_id", id))

	// Update warehouse
	response, err := h.warehouseService.UpdateWarehouse(c.Request.Context(), id, &request, jwtToken)
	if err != nil {
		h.logger.Error("Service error updating warehouse",
			zap.Error(err),
			zap.String("warehouse_id", id))
		utils.HandleServiceError(c, "Failed to update warehouse", err)
		return
	}

	h.logger.Info("Warehouse updated successfully",
		zap.String("warehouse_id", response.ID),
		zap.String("warehouse_name", response.Name))

	utils.OKResponse(c, "Warehouse updated successfully", response)
}

// DeleteWarehouse handles DELETE /api/v1/warehouses/:id
// @Summary Delete Warehouse
// @Description Delete a warehouse by ID (requires authentication)
// @Tags Warehouses
// @Produce json
// @Param id path string true "Warehouse ID (format: WHSE_xxxxxxxx)" example(WHSE_12345678)
// @Success 200 {object} utils.Response "Warehouse deleted successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Warehouse not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/warehouses/{id} [delete]
func (h *WarehouseHandler) DeleteWarehouse(c *gin.Context) {
	h.logger.Info("Handling delete warehouse request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Warehouse ID is required but not provided for delete")
		utils.BadRequestResponse(c, "Warehouse ID is required", nil)
		return
	}

	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		h.logger.Error("Missing authentication token for delete warehouse")
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	h.logger.Debug("Calling service to delete warehouse",
		zap.String("warehouse_id", id))

	// Delete warehouse
	err := h.warehouseService.DeleteWarehouse(c.Request.Context(), id, jwtToken)
	if err != nil {
		h.logger.Error("Service error deleting warehouse",
			zap.Error(err),
			zap.String("warehouse_id", id))
		utils.HandleServiceError(c, "Failed to delete warehouse", err)
		return
	}

	h.logger.Info("Warehouse deleted successfully",
		zap.String("warehouse_id", id))

	utils.OKResponse(c, "Warehouse deleted successfully", nil)
}

// SearchWarehouses handles GET /api/v1/warehouses/search
// @Summary Search Warehouses
// @Description Search warehouses by query string (requires authentication)
// @Tags Warehouses
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {object} utils.Response{data=[]models.WarehouseResponse} "Warehouses search completed"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/warehouses/search [get]
func (h *WarehouseHandler) SearchWarehouses(c *gin.Context) {
	h.logger.Info("Handling search warehouses request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get query parameter
	query := c.Query("q")
	if query == "" {
		h.logger.Error("Search query is required but not provided")
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		h.logger.Error("Missing authentication token for search warehouses")
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	h.logger.Debug("Searching warehouses",
		zap.String("query", query))

	// Search warehouses
	response, err := h.warehouseService.SearchWarehouses(c.Request.Context(), query, jwtToken)
	if err != nil {
		h.logger.Error("Service error searching warehouses",
			zap.Error(err),
			zap.String("query", query))
		utils.HandleServiceError(c, "Failed to search warehouses", err)
		return
	}

	h.logger.Info("Warehouses search completed",
		zap.String("query", query),
		zap.Int("results_count", len(response)))

	utils.OKResponse(c, "Warehouses search completed", response)
}

// RegisterRoutes registers warehouse routes
func (h *WarehouseHandler) RegisterRoutes(router *gin.RouterGroup) {
	warehouses := router.Group("/warehouses")
	{
		// Apply authentication middleware to all routes
		warehouses.Use(h.aaaMiddleware.Authenticate())

		// Create: AAA HTTP service will validate addresses permissions internally
		warehouses.POST("", h.aaaMiddleware.RequireOrgPermission("warehouse", "create"), h.CreateWarehouse)

		// Read operations: AAA HTTP service will validate addresses permissions internally
		warehouses.GET("", h.aaaMiddleware.RequireOrgPermission("warehouse", "read"), h.GetAllWarehouses)
		warehouses.GET("/search", h.aaaMiddleware.RequireOrgPermission("warehouse", "read"), h.SearchWarehouses)
		warehouses.GET("/:id", h.aaaMiddleware.RequireOrgPermission("warehouse", "read"), h.GetWarehouse)

		// Update: AAA HTTP service will validate addresses permissions internally
		warehouses.PATCH("/:id", h.aaaMiddleware.RequireOrgPermission("warehouse", "update"), h.UpdateWarehouse)

		// Delete: AAA HTTP service will validate addresses permissions internally
		warehouses.DELETE("/:id", h.aaaMiddleware.RequireOrgPermission("warehouse", "delete"), h.DeleteWarehouse)
	}
}
