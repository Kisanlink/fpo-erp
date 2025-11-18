package handlers

import (
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

// GRNHandler handles goods receipt note HTTP requests
type GRNHandler struct {
	grnService    interfaces.GRNServiceInterface
	aaaMiddleware *aaa.AAAMiddleware
}

// NewGRNHandler creates a new GRN handler
func NewGRNHandler(grnService interfaces.GRNServiceInterface, aaaMiddleware *aaa.AAAMiddleware) *GRNHandler {
	return &GRNHandler{
		grnService:    grnService,
		aaaMiddleware: aaaMiddleware,
	}
}

// CreateGRN handles POST /api/v1/grns
// @Summary Create Goods Receipt Note
// @Description Create a new goods receipt note and inventory batches (requires authentication)
// @Tags GRNs
// @Accept json
// @Produce json
// @Param request body models.CreateGRNRequest true "GRN data"
// @Success 201 {object} utils.Response{data=models.GRNResponse} "GRN created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/grns [post]
func (h *GRNHandler) CreateGRN(c *gin.Context) {
	var request models.CreateGRNRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Create GRN
	response, err := h.grnService.CreateGRN(c.Request.Context(), &request)
	if err != nil {
		utils.HandleServiceError(c, "Failed to create GRN", err)
		return
	}

	utils.CreatedResponse(c, "GRN created successfully", response)
}

// GetGRN handles GET /api/v1/grns/:id
// @Summary Get Goods Receipt Note
// @Description Retrieve a specific GRN by ID
// @Tags GRNs
// @Produce json
// @Param id path string true "GRN ID (format: GRNX_xxxxxxxx)" example(GRNX_12345678)
// @Success 200 {object} utils.Response{data=models.GRNResponse} "GRN details"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "GRN not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/grns/{id} [get]
func (h *GRNHandler) GetGRN(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "GRN ID is required", nil)
		return
	}

	// Get GRN
	response, err := h.grnService.GetGRN(c.Request.Context(), id)
	if err != nil {
		utils.NotFoundResponse(c, "GRN not found")
		return
	}

	utils.OKResponse(c, "GRN retrieved successfully", response)
}

// GetAllGRNs handles GET /api/v1/grns
// @Summary Get All Goods Receipt Notes
// @Description Retrieve all GRNs (requires authentication)
// @Tags GRNs
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.GRNResponse} "GRNs retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/grns [get]
func (h *GRNHandler) GetAllGRNs(c *gin.Context) {
	// Get all GRNs
	response, err := h.grnService.GetAllGRNs(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve GRNs", err)
		return
	}

	utils.OKResponse(c, "GRNs retrieved successfully", response)
}

// GetGRNsByWarehouse handles GET /api/v1/warehouses/:id/grns
// @Summary Get GRNs by Warehouse
// @Description Retrieve all GRNs for a specific warehouse
// @Tags GRNs
// @Produce json
// @Param id path string true "Warehouse ID (format: WHSE_xxxxxxxx)" example(WHSE_12345678)
// @Success 200 {object} utils.Response{data=[]models.GRNResponse} "GRNs retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Warehouse not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/warehouses/{id}/grns [get]
func (h *GRNHandler) GetGRNsByWarehouse(c *gin.Context) {
	// Get warehouse ID from URL
	warehouseID := c.Param("id")
	if warehouseID == "" {
		utils.BadRequestResponse(c, "Warehouse ID is required", nil)
		return
	}

	// Get GRNs
	response, err := h.grnService.GetGRNsByWarehouse(c.Request.Context(), warehouseID)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve GRNs", err)
		return
	}

	utils.OKResponse(c, "GRNs retrieved successfully", response)
}

// GetGRNByPurchaseOrder handles GET /api/v1/purchase-orders/:id/grn
// @Summary Get GRN by Purchase Order
// @Description Retrieve GRN for a specific purchase order
// @Tags GRNs
// @Produce json
// @Param id path string true "Purchase Order ID (format: PORD_xxxxxxxx)" example(PORD_12345678)
// @Success 200 {object} utils.Response{data=models.GRNResponse} "GRN retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "GRN not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/purchase-orders/{id}/grn [get]
func (h *GRNHandler) GetGRNByPurchaseOrder(c *gin.Context) {
	// Get PO ID from URL
	poID := c.Param("id")
	if poID == "" {
		utils.BadRequestResponse(c, "Purchase Order ID is required", nil)
		return
	}

	// Get GRN
	response, err := h.grnService.GetGRNByPurchaseOrder(c.Request.Context(), poID)
	if err != nil {
		utils.NotFoundResponse(c, "GRN not found for this purchase order")
		return
	}

	utils.OKResponse(c, "GRN retrieved successfully", response)
}

// UpdateGRN handles PUT /api/v1/grns/:id
// @Summary Update Goods Receipt Note
// @Description Update GRN details (attach PDF, update remarks, quality status) (requires authentication)
// @Tags GRNs
// @Accept json
// @Produce json
// @Param id path string true "GRN ID (format: GRNX_xxxxxxxx)" example(GRNX_12345678)
// @Param request body models.UpdateGRNRequest true "Update data"
// @Success 200 {object} utils.Response{data=models.GRNResponse} "GRN updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "GRN not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/grns/{id} [put]
func (h *GRNHandler) UpdateGRN(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "GRN ID is required", nil)
		return
	}

	// Validate request
	var request models.UpdateGRNRequest
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Update GRN
	response, err := h.grnService.UpdateGRN(c.Request.Context(), id, &request)
	if err != nil {
		utils.HandleServiceError(c, "Failed to update GRN", err)
		return
	}

	utils.OKResponse(c, "GRN updated successfully", response)
}

// RegisterRoutes registers all GRN routes
func (h *GRNHandler) RegisterRoutes(router *gin.RouterGroup) {
	// GRN routes
	grns := router.Group("/grns")
	grns.Use(h.aaaMiddleware.Authenticate())
	{
		grns.POST("", h.aaaMiddleware.RequireOrgPermission("grn", "create"), h.CreateGRN)
		grns.GET("", h.aaaMiddleware.RequireOrgPermission("grn", "read"), h.GetAllGRNs)
		grns.GET("/:id", h.aaaMiddleware.RequireOrgPermission("grn", "read"), h.GetGRN)
		grns.PUT("/:id", h.aaaMiddleware.RequireOrgPermission("grn", "update"), h.UpdateGRN)
	}

	// Nested routes under warehouses
	warehouses := router.Group("/warehouses")
	warehouses.Use(h.aaaMiddleware.Authenticate())
	{
		warehouses.GET("/:id/grns", h.aaaMiddleware.RequireOrgPermission("grn", "read"), h.GetGRNsByWarehouse)
	}

	// Nested routes under purchase orders
	purchaseOrders := router.Group("/purchase-orders")
	purchaseOrders.Use(h.aaaMiddleware.Authenticate())
	{
		purchaseOrders.GET("/:id/grn", h.aaaMiddleware.RequireOrgPermission("grn", "read"), h.GetGRNByPurchaseOrder)
	}
}
