package handlers

import (
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

// PurchaseOrderHandler handles purchase order HTTP requests
type PurchaseOrderHandler struct {
	poService     interfaces.PurchaseOrderServiceInterface
	aaaMiddleware *aaa.AAAMiddleware
}

// NewPurchaseOrderHandler creates a new purchase order handler
func NewPurchaseOrderHandler(poService interfaces.PurchaseOrderServiceInterface, aaaMiddleware *aaa.AAAMiddleware) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{
		poService:     poService,
		aaaMiddleware: aaaMiddleware,
	}
}

// CreatePurchaseOrder handles POST /api/v1/purchase-orders
// @Summary Create Purchase Order
// @Description Create a new purchase order (requires authentication)
// @Tags Purchase Orders
// @Accept json
// @Produce json
// @Param request body models.CreatePurchaseOrderRequest true "Purchase order data"
// @Success 201 {object} utils.Response{data=models.PurchaseOrderResponse} "Purchase order created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/purchase-orders [post]
func (h *PurchaseOrderHandler) CreatePurchaseOrder(c *gin.Context) {
	var request models.CreatePurchaseOrderRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Create purchase order
	response, err := h.poService.CreatePurchaseOrder(c.Request.Context(), &request)
	if err != nil {
		utils.HandleServiceError(c, "Failed to create purchase order", err)
		return
	}

	utils.CreatedResponse(c, "Purchase order created successfully", response)
}

// GetPurchaseOrder handles GET /api/v1/purchase-orders/:id
// @Summary Get Purchase Order
// @Description Retrieve a specific purchase order by ID
// @Tags Purchase Orders
// @Produce json
// @Param id path string true "Purchase Order ID (format: PORD_xxxxxxxx)" example(PORD_12345678)
// @Success 200 {object} utils.Response{data=models.PurchaseOrderResponse} "Purchase order details"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Purchase order not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/purchase-orders/{id} [get]
func (h *PurchaseOrderHandler) GetPurchaseOrder(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Purchase Order ID is required", nil)
		return
	}

	// Get purchase order
	response, err := h.poService.GetPurchaseOrder(c.Request.Context(), id)
	if err != nil {
		utils.NotFoundResponse(c, "Purchase order not found")
		return
	}

	utils.OKResponse(c, "Purchase order retrieved successfully", response)
}

// GetAllPurchaseOrders handles GET /api/v1/purchase-orders
// @Summary Get All Purchase Orders
// @Description Retrieve all purchase orders (requires authentication)
// @Tags Purchase Orders
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.PurchaseOrderResponse} "Purchase orders retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/purchase-orders [get]
func (h *PurchaseOrderHandler) GetAllPurchaseOrders(c *gin.Context) {
	// Get all purchase orders
	response, err := h.poService.GetAllPurchaseOrders(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve purchase orders", err)
		return
	}

	utils.OKResponse(c, "Purchase orders retrieved successfully", response)
}

// GetPurchaseOrdersByCollaborator handles GET /api/v1/collaborators/:id/purchase-orders
// @Summary Get Purchase Orders by Collaborator
// @Description Retrieve all purchase orders for a specific collaborator
// @Tags Purchase Orders
// @Produce json
// @Param id path string true "Collaborator ID (format: CLAB_xxxxxxxx)" example(CLAB_12345678)
// @Success 200 {object} utils.Response{data=[]models.PurchaseOrderResponse} "Purchase orders retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Collaborator not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/collaborators/{id}/purchase-orders [get]
func (h *PurchaseOrderHandler) GetPurchaseOrdersByCollaborator(c *gin.Context) {
	// Get collaborator ID from URL
	collaboratorID := c.Param("id")
	if collaboratorID == "" {
		utils.BadRequestResponse(c, "Collaborator ID is required", nil)
		return
	}

	// Get purchase orders
	response, err := h.poService.GetPurchaseOrdersByCollaborator(c.Request.Context(), collaboratorID)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve purchase orders", err)
		return
	}

	utils.OKResponse(c, "Purchase orders retrieved successfully", response)
}

// GetPurchaseOrdersByStatus handles GET /api/v1/purchase-orders/status/:status
// @Summary Get Purchase Orders by Status
// @Description Retrieve all purchase orders with a specific status
// @Tags Purchase Orders
// @Produce json
// @Param status path string true "Status" Enums(placed, confirmed, out_for_delivery, delivered, paid)
// @Success 200 {object} utils.Response{data=[]models.PurchaseOrderResponse} "Purchase orders retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/purchase-orders/status/{status} [get]
func (h *PurchaseOrderHandler) GetPurchaseOrdersByStatus(c *gin.Context) {
	// Get status from URL
	status := c.Param("status")
	if status == "" {
		utils.BadRequestResponse(c, "Status is required", nil)
		return
	}

	// Get purchase orders
	response, err := h.poService.GetPurchaseOrdersByStatus(c.Request.Context(), status)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve purchase orders", err)
		return
	}

	utils.OKResponse(c, "Purchase orders retrieved successfully", response)
}

// GetPendingDeliveries handles GET /api/v1/purchase-orders/pending-deliveries
// @Summary Get Pending Deliveries
// @Description Retrieve all purchase orders with pending deliveries
// @Tags Purchase Orders
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.PurchaseOrderResponse} "Pending deliveries retrieved successfully"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/purchase-orders/pending-deliveries [get]
func (h *PurchaseOrderHandler) GetPendingDeliveries(c *gin.Context) {
	// Get pending deliveries
	response, err := h.poService.GetPendingDeliveries(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve pending deliveries", err)
		return
	}

	utils.OKResponse(c, "Pending deliveries retrieved successfully", response)
}

// UpdatePurchaseOrderStatus handles PATCH /api/v1/purchase-orders/:id/status
// @Summary Update Purchase Order Status
// @Description Update the status of a purchase order (requires authentication)
// @Tags Purchase Orders
// @Accept json
// @Produce json
// @Param id path string true "Purchase Order ID (format: PORD_xxxxxxxx)" example(PORD_12345678)
// @Param request body models.UpdatePOStatusRequest true "Status update data"
// @Success 200 {object} utils.Response{data=models.PurchaseOrderResponse} "Status updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Purchase order not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/purchase-orders/{id}/status [patch]
func (h *PurchaseOrderHandler) UpdatePurchaseOrderStatus(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Purchase Order ID is required", nil)
		return
	}

	var request models.UpdatePOStatusRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Get authenticated user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system" // Fallback if user_id not found in context
	}

	// Update status
	response, err := h.poService.UpdatePurchaseOrderStatus(c.Request.Context(), id, &request, userID)
	if err != nil {
		utils.HandleServiceError(c, "Failed to update status", err)
		return
	}

	utils.OKResponse(c, "Status updated successfully", response)
}

// UpdatePaymentStatus handles PATCH /api/v1/purchase-orders/:id/payment
// @Summary Update Payment Status
// @Description Update the payment status of a purchase order (requires authentication)
// @Tags Purchase Orders
// @Accept json
// @Produce json
// @Param id path string true "Purchase Order ID (format: PORD_xxxxxxxx)" example(PORD_12345678)
// @Param request body models.UpdatePOPaymentRequest true "Payment update data"
// @Success 200 {object} utils.Response{data=models.PurchaseOrderResponse} "Payment status updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Purchase order not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/purchase-orders/{id}/payment [patch]
func (h *PurchaseOrderHandler) UpdatePaymentStatus(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Purchase Order ID is required", nil)
		return
	}

	var request models.UpdatePOPaymentRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Update payment status
	response, err := h.poService.UpdatePaymentStatus(c.Request.Context(), id, &request)
	if err != nil {
		utils.HandleServiceError(c, "Failed to update payment status", err)
		return
	}

	utils.OKResponse(c, "Payment status updated successfully", response)
}

// RegisterRoutes registers all purchase order routes
func (h *PurchaseOrderHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Purchase order routes
	purchaseOrders := router.Group("/purchase-orders")
	purchaseOrders.Use(h.aaaMiddleware.Authenticate())
	{
		purchaseOrders.POST("", h.aaaMiddleware.RequireOrgPermission("purchase_order", "create"), h.CreatePurchaseOrder)
		purchaseOrders.GET("", h.aaaMiddleware.RequireOrgPermission("purchase_order", "read"), h.GetAllPurchaseOrders)
		purchaseOrders.GET("/pending-deliveries", h.aaaMiddleware.RequireOrgPermission("purchase_order", "read"), h.GetPendingDeliveries)
		purchaseOrders.GET("/status/:status", h.aaaMiddleware.RequireOrgPermission("purchase_order", "read"), h.GetPurchaseOrdersByStatus)
		purchaseOrders.GET("/:id", h.aaaMiddleware.RequireOrgPermission("purchase_order", "read"), h.GetPurchaseOrder)
		purchaseOrders.PATCH("/:id/status", h.aaaMiddleware.RequireOrgPermission("purchase_order", "update"), h.UpdatePurchaseOrderStatus)
		purchaseOrders.PATCH("/:id/payment", h.aaaMiddleware.RequireOrgPermission("purchase_order", "update"), h.UpdatePaymentStatus)
	}

	// Nested routes under collaborators
	collaborators := router.Group("/collaborators")
	collaborators.Use(h.aaaMiddleware.Authenticate())
	{
		collaborators.GET("/:id/purchase-orders", h.aaaMiddleware.RequireOrgPermission("purchase_order", "read"), h.GetPurchaseOrdersByCollaborator)
	}
}
