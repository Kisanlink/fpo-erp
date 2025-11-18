package handlers

import (
	"strconv"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

// InventoryHandler handles inventory HTTP requests
type InventoryHandler struct {
	inventoryService interfaces.InventoryServiceInterface
	aaaMiddleware    *aaa.AAAMiddleware
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(inventoryService interfaces.InventoryServiceInterface, aaaMiddleware *aaa.AAAMiddleware) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
		aaaMiddleware:    aaaMiddleware,
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
	var request models.CreateInventoryBatchRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Parse expiry date
	expiryDate, err := time.Parse("2006-01-02", request.ExpiryDate)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid expiry date format (YYYY-MM-DD)", err)
		return
	}

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
		utils.HandleServiceError(c, "Failed to create batch", err)
		return
	}

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
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Batch ID is required", nil)
		return
	}

	// Get batch
	response, err := h.inventoryService.GetBatch(id)
	if err != nil {
		utils.NotFoundResponse(c, "Batch not found")
		return
	}

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
	// Get warehouse ID from URL
	warehouseID := c.Param("id")
	if warehouseID == "" {
		utils.BadRequestResponse(c, "Warehouse ID is required", nil)
		return
	}

	// Get batches by warehouse
	response, err := h.inventoryService.GetBatchesByWarehouse(warehouseID)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve batches", err)
		return
	}

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
	// Get variant ID from URL
	variantID := c.Param("id")
	if variantID == "" {
		utils.BadRequestResponse(c, "Variant ID is required", nil)
		return
	}

	// Get batches by variant
	response, err := h.inventoryService.GetBatchesByVariant(variantID)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve batches", err)
		return
	}

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
	// Get batch ID from URL
	batchID := c.Param("id")
	if batchID == "" {
		utils.BadRequestResponse(c, "Batch ID is required", nil)
		return
	}

	var request models.CreateInventoryTransactionRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Create transaction
	response, err := h.inventoryService.CreateTransaction(batchID, &request)
	if err != nil {
		utils.HandleServiceError(c, "Failed to create transaction", err)
		return
	}

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
	// Get batch ID from URL
	batchID := c.Param("id")
	if batchID == "" {
		utils.BadRequestResponse(c, "Batch ID is required", nil)
		return
	}

	// Get transactions by batch
	response, err := h.inventoryService.GetTransactionsByBatch(batchID)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve transactions", err)
		return
	}

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
	// Get days parameter
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid days parameter", err)
		return
	}

	// Get expiring batches
	response, err := h.inventoryService.GetExpiringBatches(days)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve expiring batches", err)
		return
	}

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
	// Get threshold parameter
	thresholdStr := c.DefaultQuery("threshold", "10")
	threshold, err := strconv.ParseInt(thresholdStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid threshold parameter", err)
		return
	}

	// Get low stock batches
	response, err := h.inventoryService.GetLowStockBatches(threshold)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve low stock batches", err)
		return
	}

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
	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	// Get all products availability across warehouses
	response, err := h.inventoryService.GetAllProductsAvailability(c.Request.Context(), jwtToken)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve products availability", err)
		return
	}

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
