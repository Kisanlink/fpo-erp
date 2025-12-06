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

// PurchaseOrderHandler handles purchase order HTTP requests
type PurchaseOrderHandler struct {
	poService     interfaces.PurchaseOrderServiceInterface
	aaaMiddleware *aaa.AAAMiddleware
	logger        logger.Logger
}

// NewPurchaseOrderHandler creates a new purchase order handler
func NewPurchaseOrderHandler(poService interfaces.PurchaseOrderServiceInterface, aaaMiddleware *aaa.AAAMiddleware, logger logger.Logger) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{
		poService:     poService,
		aaaMiddleware: aaaMiddleware,
		logger:        logger,
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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/purchase-orders [post]
func (h *PurchaseOrderHandler) CreatePurchaseOrder(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Creating purchase order",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	var request models.CreatePurchaseOrderRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for create purchase order",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling CreatePurchaseOrder service",
		zap.String("collaborator_id", request.CollaboratorID),
		zap.String("warehouse_id", request.WarehouseID))

	// Create purchase order
	response, err := h.poService.CreatePurchaseOrder(c.Request.Context(), &request)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to create purchase order",
			zap.Error(err),
			zap.String("collaborator_id", request.CollaboratorID),
			zap.String("warehouse_id", request.WarehouseID))
		utils.HandleServiceError(c, "Failed to create purchase order", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Purchase order created successfully",
		zap.String("purchase_order_id", response.ID),
		zap.Float64("total_amount", response.TotalAmount))

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
	// 1. Entry Log
	id := c.Param("id")
	h.logger.Info("Getting purchase order",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("po_id", id))

	if id == "" {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for get purchase order",
			zap.String("error", "purchase order ID is required"))
		utils.BadRequestResponse(c, "Purchase Order ID is required", nil)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling GetPurchaseOrder service",
		zap.String("po_id", id))

	// Get purchase order
	response, err := h.poService.GetPurchaseOrder(c.Request.Context(), id)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to get purchase order",
			zap.Error(err),
			zap.String("po_id", id))
		utils.NotFoundResponse(c, "Purchase order not found")
		return
	}

	// 5. Success Log
	h.logger.Info("Purchase order retrieved successfully",
		zap.String("po_id", response.ID),
		zap.String("status", response.Status))

	utils.OKResponse(c, "Purchase order retrieved successfully", response)
}

// GetAllPurchaseOrders handles GET /api/v1/purchase-orders
// @Summary Get All Purchase Orders
// @Description Retrieve all purchase orders with pagination (requires authentication)
// @Tags Purchase Orders
// @Produce json
// @Param limit query integer false "Number of records to return (default: 50, max: 200)" example(50)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.PaginatedResponseModel{data=[]models.PurchaseOrderResponse} "Purchase orders retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/purchase-orders [get]
func (h *PurchaseOrderHandler) GetAllPurchaseOrders(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Getting all purchase orders",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// 3. Service Call Log
	h.logger.Debug("Calling GetAllPurchaseOrders service",
		zap.Int("limit", params.Limit),
		zap.Int("offset", params.Offset))

	// Get all purchase orders
	response, total, err := h.poService.GetAllPurchaseOrders(c.Request.Context(), params.Limit, params.Offset)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to retrieve purchase orders",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve purchase orders", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Purchase orders retrieved successfully",
		zap.Int("count", len(response)),
		zap.Int64("total", total))

	utils.PaginatedOKResponse(c, response, total, params.Limit, params.Offset)
}

// GetPurchaseOrdersByCollaborator handles GET /api/v1/collaborators/:id/purchase-orders
// @Summary Get Purchase Orders by Collaborator
// @Description Retrieve all purchase orders for a specific collaborator with pagination
// @Tags Purchase Orders
// @Produce json
// @Param id path string true "Collaborator ID (format: CLAB_xxxxxxxx)" example(CLAB_12345678)
// @Param limit query integer false "Number of records to return (default: 50, max: 200)" example(50)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.PaginatedResponseModel{data=[]models.PurchaseOrderResponse} "Purchase orders retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Collaborator not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/collaborators/{id}/purchase-orders [get]
func (h *PurchaseOrderHandler) GetPurchaseOrdersByCollaborator(c *gin.Context) {
	// 1. Entry Log
	collaboratorID := c.Param("id")
	h.logger.Info("Getting purchase orders by collaborator",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("collaborator_id", collaboratorID))

	if collaboratorID == "" {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for get POs by collaborator",
			zap.String("error", "collaborator ID is required"))
		utils.BadRequestResponse(c, "Collaborator ID is required", nil)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// 3. Service Call Log
	h.logger.Debug("Calling GetPurchaseOrdersByCollaborator service",
		zap.String("collaborator_id", collaboratorID),
		zap.Int("limit", params.Limit),
		zap.Int("offset", params.Offset))

	// Get purchase orders
	response, total, err := h.poService.GetPurchaseOrdersByCollaborator(c.Request.Context(), collaboratorID, params.Limit, params.Offset)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to retrieve purchase orders",
			zap.Error(err),
			zap.String("collaborator_id", collaboratorID))
		utils.HandleServiceError(c, "Failed to retrieve purchase orders", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Purchase orders retrieved successfully",
		zap.String("collaborator_id", collaboratorID),
		zap.Int("count", len(response)),
		zap.Int64("total", total))

	utils.PaginatedOKResponse(c, response, total, params.Limit, params.Offset)
}

// GetPurchaseOrdersByStatus handles GET /api/v1/purchase-orders/status/:status
// @Summary Get Purchase Orders by Status
// @Description Retrieve all purchase orders with a specific status with pagination
// @Tags Purchase Orders
// @Produce json
// @Param status path string true "Status" Enums(placed, confirmed, out_for_delivery, delivered, paid)
// @Param limit query integer false "Number of records to return (default: 50, max: 200)" example(50)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.PaginatedResponseModel{data=[]models.PurchaseOrderResponse} "Purchase orders retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/purchase-orders/status/{status} [get]
func (h *PurchaseOrderHandler) GetPurchaseOrdersByStatus(c *gin.Context) {
	// 1. Entry Log
	status := c.Param("status")
	h.logger.Info("Getting purchase orders by status",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("status", status))

	if status == "" {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for get POs by status",
			zap.String("error", "status is required"))
		utils.BadRequestResponse(c, "Status is required", nil)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// 3. Service Call Log
	h.logger.Debug("Calling GetPurchaseOrdersByStatus service",
		zap.String("status", status),
		zap.Int("limit", params.Limit),
		zap.Int("offset", params.Offset))

	// Get purchase orders
	response, total, err := h.poService.GetPurchaseOrdersByStatus(c.Request.Context(), status, params.Limit, params.Offset)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to retrieve purchase orders",
			zap.Error(err),
			zap.String("status", status))
		utils.HandleServiceError(c, "Failed to retrieve purchase orders", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Purchase orders retrieved successfully",
		zap.String("status", status),
		zap.Int("count", len(response)),
		zap.Int64("total", total))

	utils.PaginatedOKResponse(c, response, total, params.Limit, params.Offset)
}

// GetPendingDeliveries handles GET /api/v1/purchase-orders/pending-deliveries
// @Summary Get Pending Deliveries
// @Description Retrieve all purchase orders with pending deliveries with pagination
// @Tags Purchase Orders
// @Produce json
// @Param limit query integer false "Number of records to return (default: 50, max: 200)" example(50)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.PaginatedResponseModel{data=[]models.PurchaseOrderResponse} "Pending deliveries retrieved successfully"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/purchase-orders/pending-deliveries [get]
func (h *PurchaseOrderHandler) GetPendingDeliveries(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Getting pending deliveries",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// 3. Service Call Log
	h.logger.Debug("Calling GetPendingDeliveries service",
		zap.Int("limit", params.Limit),
		zap.Int("offset", params.Offset))

	// Get pending deliveries
	response, total, err := h.poService.GetPendingDeliveries(c.Request.Context(), params.Limit, params.Offset)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to retrieve pending deliveries",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve pending deliveries", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Pending deliveries retrieved successfully",
		zap.Int("count", len(response)),
		zap.Int64("total", total))

	utils.PaginatedOKResponse(c, response, total, params.Limit, params.Offset)
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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Purchase order not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/purchase-orders/{id}/status [patch]
func (h *PurchaseOrderHandler) UpdatePurchaseOrderStatus(c *gin.Context) {
	// 1. Entry Log
	id := c.Param("id")
	h.logger.Info("Updating purchase order status",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("po_id", id))

	if id == "" {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for update PO status",
			zap.String("error", "purchase order ID is required"))
		utils.BadRequestResponse(c, "Purchase Order ID is required", nil)
		return
	}

	var request models.UpdatePOStatusRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for update PO status",
			zap.Error(err),
			zap.String("po_id", id))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Get authenticated user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system" // Fallback if user_id not found in context
	}

	// 3. Service Call Log (CRITICAL - status transitions)
	h.logger.Debug("Calling UpdatePurchaseOrderStatus service",
		zap.String("po_id", id),
		zap.String("new_status", request.Status),
		zap.String("user_id", userID))

	// Update status
	response, err := h.poService.UpdatePurchaseOrderStatus(c.Request.Context(), id, &request, userID)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to update PO status",
			zap.Error(err),
			zap.String("po_id", id),
			zap.String("new_status", request.Status))
		utils.HandleServiceError(c, "Failed to update status", err)
		return
	}

	// 5. Success Log (log status transition)
	h.logger.Info("Purchase order status updated successfully",
		zap.String("po_id", response.ID),
		zap.String("new_status", response.Status))

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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Purchase order not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/purchase-orders/{id}/payment [patch]
func (h *PurchaseOrderHandler) UpdatePaymentStatus(c *gin.Context) {
	// 1. Entry Log
	id := c.Param("id")
	h.logger.Info("Updating payment status",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("po_id", id))

	if id == "" {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for update payment status",
			zap.String("error", "purchase order ID is required"))
		utils.BadRequestResponse(c, "Purchase Order ID is required", nil)
		return
	}

	var request models.UpdatePOPaymentRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for update payment status",
			zap.Error(err),
			zap.String("po_id", id))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling UpdatePaymentStatus service",
		zap.String("po_id", id),
		zap.String("payment_status", request.PaymentStatus),
		zap.Float64("paid_amount", request.PaidAmount))

	// Update payment status
	response, err := h.poService.UpdatePaymentStatus(c.Request.Context(), id, &request)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to update payment status",
			zap.Error(err),
			zap.String("po_id", id))
		utils.HandleServiceError(c, "Failed to update payment status", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Payment status updated successfully",
		zap.String("po_id", response.ID),
		zap.String("payment_status", response.PaymentStatus),
		zap.Float64("paid_amount", response.PaidAmount))

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
