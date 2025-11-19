package handlers

import (
	"strconv"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	logger "kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InventoryHandler handles inventory HTTP requests
type InventoryHandler struct {
	inventoryService interfaces.InventoryServiceInterface
	aaaMiddleware    *aaa.AAAMiddleware
	logger           logger.Logger
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(inventoryService interfaces.InventoryServiceInterface, aaaMiddleware *aaa.AAAMiddleware, logger logger.Logger) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
		aaaMiddleware:    aaaMiddleware,
		logger:           logger,
	}
}

// CreateBatch handles POST /api/v1/batches
// @Summary Create Inventory Batch
// @Description Create a new inventory batch (requires authentication)
// @Tags Inventory
// @Accept json
// @Produce json
// @Param request body models.CreateInventoryBatchRequest true "Batch data"
// @Success 201 {object} utils.Response{data=models.InventoryBatchResponse} "Batch created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/batches [post]
func (h *InventoryHandler) CreateBatch(c *gin.Context) {
	h.logger.Info("Handling create inventory batch request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	var request models.CreateInventoryBatchRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		h.logger.Error("Invalid request body for create inventory batch",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Parse expiry date
	expiryDate, err := time.Parse("2006-01-02", request.ExpiryDate)
	if err != nil {
		h.logger.Error("Invalid expiry date format",
			zap.Error(err),
			zap.String("expiry_date", request.ExpiryDate))
		utils.BadRequestResponse(c, "Invalid expiry date format (YYYY-MM-DD)", err)
		return
	}

	h.logger.Debug("Calling inventory service to create batch",
		zap.String("warehouse_id", request.WarehouseID),
		zap.String("variant_id", request.VariantID),
		zap.Int64("quantity", request.Quantity),
		zap.String("expiry_date", request.ExpiryDate))

	// Create batch with tax configuration
	response, err := h.inventoryService.CreateBatch(
		request.WarehouseID,
		request.VariantID,
		request.CostPrice,
		expiryDate,
		request.Quantity,
		request.CGSTRate,
		request.SGSTRate,
		request.CustomTaxIDs,
		request.IsTaxExempt,
	)
	if err != nil {
		h.logger.Error("Failed to create inventory batch via service",
			zap.Error(err),
			zap.String("warehouse_id", request.WarehouseID),
			zap.String("variant_id", request.VariantID))
		utils.HandleServiceError(c, "Failed to create batch", err)
		return
	}

	h.logger.Info("Inventory batch created successfully via handler",
		zap.String("batch_id", response.ID),
		zap.String("warehouse_id", request.WarehouseID),
		zap.Int64("quantity", request.Quantity))

	utils.CreatedResponse(c, "Batch created successfully", response)
}

// GetBatch handles GET /api/v1/batches/:id
// @Summary Get Inventory Batch
// @Description Retrieve a specific inventory batch by ID
// @Tags Inventory
// @Produce json
// @Param id path string true "Batch ID" example(BTCH_12345678)
// @Success 200 {object} utils.Response{data=models.InventoryBatchResponse} "Batch retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Batch not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/batches/{id} [get]
func (h *InventoryHandler) GetBatch(c *gin.Context) {
	h.logger.Info("Handling get inventory batch request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Batch ID is required but not provided")
		utils.BadRequestResponse(c, "Batch ID is required", nil)
		return
	}

	h.logger.Debug("Calling inventory service to get batch",
		zap.String("batch_id", id))

	// Get batch
	response, err := h.inventoryService.GetBatch(id)
	if err != nil {
		h.logger.Error("Inventory batch not found",
			zap.Error(err),
			zap.String("batch_id", id))
		utils.NotFoundResponse(c, "Batch not found")
		return
	}

	h.logger.Info("Inventory batch retrieved successfully via handler",
		zap.String("batch_id", response.ID),
		zap.String("warehouse_id", response.WarehouseID))

	utils.OKResponse(c, "Batch retrieved successfully", response)
}

// GetBatchesByWarehouse handles GET /api/v1/warehouses/:id/batches
// @Summary Get Batches by Warehouse
// @Description Retrieve all inventory batches for a specific warehouse
// @Tags Inventory
// @Produce json
// @Param id path string true "Warehouse ID" example(WHSE_12345678)
// @Success 200 {object} utils.Response{data=[]models.InventoryBatchResponse} "Batches retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/warehouses/{id}/batches [get]
func (h *InventoryHandler) GetBatchesByWarehouse(c *gin.Context) {
	h.logger.Info("Handling get batches by warehouse request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get warehouse ID from URL
	warehouseID := c.Param("id")
	if warehouseID == "" {
		h.logger.Error("Warehouse ID is required but not provided")
		utils.BadRequestResponse(c, "Warehouse ID is required", nil)
		return
	}

	h.logger.Debug("Calling inventory service to get batches by warehouse",
		zap.String("warehouse_id", warehouseID))

	// Get batches by warehouse
	response, err := h.inventoryService.GetBatchesByWarehouse(warehouseID)
	if err != nil {
		h.logger.Error("Failed to retrieve batches by warehouse via service",
			zap.Error(err),
			zap.String("warehouse_id", warehouseID))
		utils.HandleServiceError(c, "Failed to retrieve batches", err)
		return
	}

	h.logger.Info("Batches retrieved successfully by warehouse via handler",
		zap.String("warehouse_id", warehouseID),
		zap.Int("count", len(response)))

	utils.OKResponse(c, "Batches retrieved successfully", response)
}

// GetBatchesByVariant handles GET /api/v1/variants/:id/batches
// @Summary Get Batches by Variant
// @Description Retrieve all inventory batches for a specific product variant
// @Tags Inventory
// @Produce json
// @Param id path string true "Variant ID" example(PVAR_12345678)
// @Success 200 {object} utils.Response{data=[]models.InventoryBatchResponse} "Batches retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/variants/{id}/batches [get]
func (h *InventoryHandler) GetBatchesByVariant(c *gin.Context) {
	h.logger.Info("Handling get batches by variant request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get variant ID from URL
	variantID := c.Param("id")
	if variantID == "" {
		h.logger.Error("Variant ID is required but not provided")
		utils.BadRequestResponse(c, "Variant ID is required", nil)
		return
	}

	h.logger.Debug("Calling inventory service to get batches by variant",
		zap.String("variant_id", variantID))

	// Get batches by variant
	response, err := h.inventoryService.GetBatchesByVariant(variantID)
	if err != nil {
		h.logger.Error("Failed to retrieve batches by variant via service",
			zap.Error(err),
			zap.String("variant_id", variantID))
		utils.HandleServiceError(c, "Failed to retrieve batches", err)
		return
	}

	h.logger.Info("Batches retrieved successfully by variant via handler",
		zap.String("variant_id", variantID),
		zap.Int("count", len(response)))

	utils.OKResponse(c, "Batches retrieved successfully", response)
}

// CreateTransaction handles POST /api/v1/batches/:id/transactions
// @Summary Create Inventory Transaction
// @Description Create a new inventory transaction for a batch (requires authentication)
// @Tags Inventory
// @Accept json
// @Produce json
// @Param id path string true "Batch ID" example(BTCH_12345678)
// @Param request body models.CreateInventoryTransactionRequest true "Transaction data"
// @Success 201 {object} utils.Response{data=models.InventoryTransactionResponse} "Transaction created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/batches/{id}/transactions [post]
func (h *InventoryHandler) CreateTransaction(c *gin.Context) {
	h.logger.Info("Handling create inventory transaction request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get batch ID from URL
	batchID := c.Param("id")
	if batchID == "" {
		h.logger.Error("Batch ID is required but not provided")
		utils.BadRequestResponse(c, "Batch ID is required", nil)
		return
	}

	var request models.CreateInventoryTransactionRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		h.logger.Error("Invalid request body for create inventory transaction",
			zap.Error(err),
			zap.String("batch_id", batchID))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	h.logger.Debug("Calling inventory service to create transaction",
		zap.String("batch_id", batchID),
		zap.String("transaction_type", request.TransactionType),
		zap.Int64("quantity_change", request.QuantityChange))

	// Create transaction
	response, err := h.inventoryService.CreateTransaction(batchID, &request)
	if err != nil {
		h.logger.Error("Failed to create inventory transaction via service",
			zap.Error(err),
			zap.String("batch_id", batchID),
			zap.String("transaction_type", request.TransactionType))
		utils.HandleServiceError(c, "Failed to create transaction", err)
		return
	}

	h.logger.Info("Inventory transaction created successfully via handler",
		zap.String("transaction_id", response.ID),
		zap.String("batch_id", batchID),
		zap.Int64("quantity_change", request.QuantityChange))

	utils.CreatedResponse(c, "Transaction created successfully", response)
}

// GetTransactionsByBatch handles GET /api/v1/batches/:id/transactions
// @Summary Get Transactions by Batch
// @Description Retrieve all inventory transactions for a specific batch
// @Tags Inventory
// @Produce json
// @Param id path string true "Batch ID" example(BTCH_12345678)
// @Success 200 {object} utils.Response{data=[]models.InventoryTransactionResponse} "Transactions retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/batches/{id}/transactions [get]
func (h *InventoryHandler) GetTransactionsByBatch(c *gin.Context) {
	h.logger.Info("Handling get transactions by batch request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get batch ID from URL
	batchID := c.Param("id")
	if batchID == "" {
		h.logger.Error("Batch ID is required but not provided")
		utils.BadRequestResponse(c, "Batch ID is required", nil)
		return
	}

	h.logger.Debug("Calling inventory service to get transactions by batch",
		zap.String("batch_id", batchID))

	// Get transactions by batch
	response, err := h.inventoryService.GetTransactionsByBatch(batchID)
	if err != nil {
		h.logger.Error("Failed to retrieve transactions by batch via service",
			zap.Error(err),
			zap.String("batch_id", batchID))
		utils.HandleServiceError(c, "Failed to retrieve transactions", err)
		return
	}

	h.logger.Info("Transactions retrieved successfully by batch via handler",
		zap.String("batch_id", batchID),
		zap.Int("count", len(response)))

	utils.OKResponse(c, "Transactions retrieved successfully", response)
}

// GetExpiringBatches handles GET /api/v1/batches/expiring
// @Summary Get Expiring Batches
// @Description Retrieve inventory batches that are expiring within specified days
// @Tags Inventory
// @Produce json
// @Param days query integer false "Number of days to check (default: 30)" example(30)
// @Success 200 {object} utils.Response{data=[]models.InventoryBatchResponse} "Expiring batches retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/batches/expiring [get]
func (h *InventoryHandler) GetExpiringBatches(c *gin.Context) {
	h.logger.Info("Handling get expiring batches request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get days parameter
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		h.logger.Error("Invalid days parameter",
			zap.Error(err),
			zap.String("days", daysStr))
		utils.BadRequestResponse(c, "Invalid days parameter", err)
		return
	}

	h.logger.Debug("Calling inventory service to get expiring batches",
		zap.Int("days", days))

	// Get expiring batches
	response, err := h.inventoryService.GetExpiringBatches(days)
	if err != nil {
		h.logger.Error("Failed to retrieve expiring batches via service",
			zap.Error(err),
			zap.Int("days", days))
		utils.HandleServiceError(c, "Failed to retrieve expiring batches", err)
		return
	}

	h.logger.Info("Expiring batches retrieved successfully via handler",
		zap.Int("days", days),
		zap.Int("count", len(response)))

	utils.OKResponse(c, "Expiring batches retrieved successfully", response)
}

// GetLowStockBatches handles GET /api/v1/batches/low-stock
// @Summary Get Low Stock Batches
// @Description Retrieve inventory batches with stock below threshold
// @Tags Inventory
// @Produce json
// @Param threshold query integer false "Stock threshold (default: 10)" example(10)
// @Success 200 {object} utils.Response{data=[]models.InventoryBatchResponse} "Low stock batches retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/batches/low-stock [get]
func (h *InventoryHandler) GetLowStockBatches(c *gin.Context) {
	h.logger.Info("Handling get low stock batches request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get threshold parameter
	thresholdStr := c.DefaultQuery("threshold", "10")
	threshold, err := strconv.ParseInt(thresholdStr, 10, 64)
	if err != nil {
		h.logger.Error("Invalid threshold parameter",
			zap.Error(err),
			zap.String("threshold", thresholdStr))
		utils.BadRequestResponse(c, "Invalid threshold parameter", err)
		return
	}

	h.logger.Debug("Calling inventory service to get low stock batches",
		zap.Int64("threshold", threshold))

	// Get low stock batches
	response, err := h.inventoryService.GetLowStockBatches(threshold)
	if err != nil {
		h.logger.Error("Failed to retrieve low stock batches via service",
			zap.Error(err),
			zap.Int64("threshold", threshold))
		utils.HandleServiceError(c, "Failed to retrieve low stock batches", err)
		return
	}

	h.logger.Info("Low stock batches retrieved successfully via handler",
		zap.Int64("threshold", threshold),
		zap.Int("count", len(response)))

	utils.OKResponse(c, "Low stock batches retrieved successfully", response)
}

// GetAllProductsAvailability handles GET /api/v1/products/availability
// @Summary Get Products Availability
// @Description Retrieve availability information for all products across warehouses
// @Tags Inventory
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.ProductAvailabilityResponse} "Products availability retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/availability [get]
func (h *InventoryHandler) GetAllProductsAvailability(c *gin.Context) {
	h.logger.Info("Handling get all products availability request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		h.logger.Error("Missing authentication token for products availability")
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	h.logger.Debug("Calling inventory service to get all products availability")

	// Get all products availability across warehouses
	response, err := h.inventoryService.GetAllProductsAvailability(c.Request.Context(), jwtToken)
	if err != nil {
		h.logger.Error("Failed to retrieve products availability via service",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve products availability", err)
		return
	}

	h.logger.Info("Products availability retrieved successfully via handler",
		zap.Int("count", len(response)))

	utils.OKResponse(c, "Products availability retrieved successfully", response)
}

// RegisterRoutes registers inventory routes
func (h *InventoryHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Batch routes
	batches := router.Group("/batches")
	{
		// Apply authentication middleware
		batches.Use(h.aaaMiddleware.Authenticate())

		// Create routes - CEO=CRUD, Store_Manager=CRUD, Tech_Support=R/W (temp)
		batches.POST("", h.aaaMiddleware.RequireOrgPermission("inventory_batch", "create"), h.CreateBatch)
		batches.POST("/:id/transactions", h.aaaMiddleware.RequireOrgPermission("inventory_transaction", "create"), h.CreateTransaction)

		// Read routes - Director=R, CEO=CRUD, Auditor=R, Accountant=–, Tech_Support=R/W (temp), Store_Manager=CRUD, Store_Staff=R
		batches.GET("/expiring", h.aaaMiddleware.RequireOrgPermission("inventory_batch", "read"), h.GetExpiringBatches)
		batches.GET("/low-stock", h.aaaMiddleware.RequireOrgPermission("inventory_batch", "read"), h.GetLowStockBatches)
		batches.GET("/:id", h.aaaMiddleware.RequireOrgPermission("inventory_batch", "read"), h.GetBatch)
		batches.GET("/:id/transactions", h.aaaMiddleware.RequireOrgPermission("inventory_transaction", "read"), h.GetTransactionsByBatch)
	}

	// Warehouse batch routes
	warehouses := router.Group("/warehouses")
	{
		warehouses.Use(h.aaaMiddleware.Authenticate())
		warehouses.GET("/:id/batches", h.aaaMiddleware.RequireOrgPermission("inventory_batch", "read"), h.GetBatchesByWarehouse)
	}

	// Product batch routes
	// Variant-specific batch routes
	variants := router.Group("/variants")
	{
		variants.Use(h.aaaMiddleware.Authenticate())
		variants.GET("/:id/batches", h.aaaMiddleware.RequireOrgPermission("inventory_batch", "read"), h.GetBatchesByVariant)
	}

	// Protected product availability route
	// AAA HTTP service will validate addresses permissions internally
	products := router.Group("/products")
	products.Use(h.aaaMiddleware.Authenticate())
	{
		products.GET("/availability", h.aaaMiddleware.RequireOrgPermission("inventory_batch", "read"), h.GetAllProductsAvailability)
	}
}
