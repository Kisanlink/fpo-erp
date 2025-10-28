package handlers

import (
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

// WarehouseHandler handles warehouse HTTP requests
type WarehouseHandler struct {
	warehouseService *services.WarehouseService
	aaaMiddleware    *aaa.AAAMiddleware
}

// NewWarehouseHandler creates a new warehouse handler
func NewWarehouseHandler(warehouseService *services.WarehouseService, aaaMiddleware *aaa.AAAMiddleware) *WarehouseHandler {
	return &WarehouseHandler{
		warehouseService: warehouseService,
		aaaMiddleware:    aaaMiddleware,
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
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/warehouses [post]
func (h *WarehouseHandler) CreateWarehouse(c *gin.Context) {
	var request models.CreateWarehouseRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Get authenticated user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system" // Fallback for unauthenticated contexts
	}

	// Create warehouse
	response, err := h.warehouseService.CreateWarehouse(c.Request.Context(), &request, userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create warehouse", err)
		return
	}

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
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Warehouse ID is required", nil)
		return
	}

	// Get warehouse
	response, err := h.warehouseService.GetWarehouse(c.Request.Context(), id)
	if err != nil {
		utils.NotFoundResponse(c, "Warehouse not found")
		return
	}

	utils.OKResponse(c, "Warehouse retrieved successfully", response)
}

// GetAllWarehouses handles GET /api/v1/warehouses
// @Summary Get All Warehouses
// @Description Retrieve all warehouses (requires authentication)
// @Tags Warehouses
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.WarehouseResponse} "Warehouses retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/warehouses [get]
func (h *WarehouseHandler) GetAllWarehouses(c *gin.Context) {
	// Get all warehouses
	response, err := h.warehouseService.GetAllWarehouses(c.Request.Context())
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve warehouses", err)
		return
	}

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
// @Failure 404 {object} utils.ErrorResponseModel "Warehouse not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/warehouses/{id} [patch]
func (h *WarehouseHandler) UpdateWarehouse(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Warehouse ID is required", nil)
		return
	}

	var request models.UpdateWarehouseRequest

	// Validate request
	if err := utils.ValidatePartialRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Update warehouse
	response, err := h.warehouseService.UpdateWarehouse(c.Request.Context(), id, &request)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update warehouse", err)
		return
	}

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
// @Failure 404 {object} utils.ErrorResponseModel "Warehouse not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/warehouses/{id} [delete]
func (h *WarehouseHandler) DeleteWarehouse(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Warehouse ID is required", nil)
		return
	}

	// Delete warehouse
	err := h.warehouseService.DeleteWarehouse(c.Request.Context(), id)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete warehouse", err)
		return
	}

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
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/warehouses/search [get]
func (h *WarehouseHandler) SearchWarehouses(c *gin.Context) {
	// Get query parameter
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	// Search warehouses
	response, err := h.warehouseService.SearchWarehouses(c.Request.Context(), query)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to search warehouses", err)
		return
	}

	utils.OKResponse(c, "Warehouses search completed", response)
}

// RegisterRoutes registers warehouse routes
func (h *WarehouseHandler) RegisterRoutes(router *gin.RouterGroup) {
	warehouses := router.Group("/warehouses")
	{
		// Apply authentication middleware to all routes
		warehouses.Use(h.aaaMiddleware.Authenticate())

		warehouses.POST("", h.aaaMiddleware.RequireOrgPermission("warehouse", "create"), h.CreateWarehouse)
		warehouses.GET("", h.aaaMiddleware.RequireOrgPermission("warehouse", "read"), h.GetAllWarehouses)
		warehouses.GET("/search", h.aaaMiddleware.RequireOrgPermission("warehouse", "read"), h.SearchWarehouses)
		warehouses.GET("/:id", h.aaaMiddleware.RequireOrgPermission("warehouse", "read"), h.GetWarehouse)
		warehouses.PATCH("/:id", h.aaaMiddleware.RequireOrgPermission("warehouse", "update"), h.UpdateWarehouse)
		warehouses.DELETE("/:id", h.aaaMiddleware.RequireOrgPermission("warehouse", "delete"), h.DeleteWarehouse)
	}
}
