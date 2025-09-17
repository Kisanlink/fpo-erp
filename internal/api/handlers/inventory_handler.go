package handlers

import (
	"strconv"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

// InventoryHandler handles inventory HTTP requests
type InventoryHandler struct {
	inventoryService *services.InventoryService
	aaaMiddleware    *aaa.AAAMiddleware
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(inventoryService *services.InventoryService, aaaMiddleware *aaa.AAAMiddleware) *InventoryHandler {
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
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/batches [post]
func (h *InventoryHandler) CreateBatch(c *gin.Context) {
	var request struct {
		WarehouseID string  `json:"warehouse_id" binding:"required"`
		ProductID   string  `json:"product_id" binding:"required"`
		CostPrice   float64 `json:"cost_price" binding:"required,gt=0"`
		ExpiryDate  string  `json:"expiry_date" binding:"required"`
		Quantity    int64   `json:"quantity" binding:"required,gt=0"`
	}

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Parse UUIDs
	warehouseID := request.WarehouseID
	if warehouseID == "" {
		utils.BadRequestResponse(c, "Warehouse ID is required", nil)
		return
	}

	productID := request.ProductID
	if productID == "" {
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	// Parse expiry date
	expiryDate, err := time.Parse("2006-01-02", request.ExpiryDate)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid expiry date format (YYYY-MM-DD)", err)
		return
	}

	// Create batch
	response, err := h.inventoryService.CreateBatch(warehouseID, productID, request.CostPrice, expiryDate, request.Quantity)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create batch", err)
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
// @Failure 404 {object} utils.ErrorResponseModel "Batch not found"
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
		utils.InternalServerErrorResponse(c, "Failed to retrieve batches", err)
		return
	}

	utils.OKResponse(c, "Batches retrieved successfully", response)
}

// GetBatchesByProduct handles GET /api/v1/products/:id/batches
// @Summary Get Batches by Product
// @Description Retrieve all inventory batches for a specific product
// @Tags Inventory
// @Produce json
// @Param id path string true "Product ID" example(PROD_12345678)
// @Success 200 {object} utils.Response{data=[]models.InventoryBatchResponse} "Batches retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/{id}/batches [get]
func (h *InventoryHandler) GetBatchesByProduct(c *gin.Context) {
	// Get product ID from URL
	productID := c.Param("id")
	if productID == "" {
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	// Get batches by product
	response, err := h.inventoryService.GetBatchesByProduct(productID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve batches", err)
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
		utils.InternalServerErrorResponse(c, "Failed to create transaction", err)
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
		utils.InternalServerErrorResponse(c, "Failed to retrieve transactions", err)
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
		utils.InternalServerErrorResponse(c, "Failed to retrieve expiring batches", err)
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
		utils.InternalServerErrorResponse(c, "Failed to retrieve low stock batches", err)
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
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/availability [get]
func (h *InventoryHandler) GetAllProductsAvailability(c *gin.Context) {
	// Get all products availability across warehouses
	response, err := h.inventoryService.GetAllProductsAvailability(c.Request.Context())
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve products availability", err)
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
		batches.POST("", h.aaaMiddleware.RequirePermission("aaa/inventory_batch", "*", "create"), h.CreateBatch)
		batches.POST("/:id/transactions", h.aaaMiddleware.RequirePermission("aaa/inventory_transaction", "*", "create"), h.CreateTransaction)

		// Read routes - Director=R, CEO=CRUD, Auditor=R, Accountant=–, Tech_Support=R/W (temp), Store_Manager=CRUD, Store_Staff=R
		batches.GET("/expiring", h.aaaMiddleware.RequirePermission("aaa/inventory_batch", "*", "read"), h.GetExpiringBatches)
		batches.GET("/low-stock", h.aaaMiddleware.RequirePermission("aaa/inventory_batch", "*", "read"), h.GetLowStockBatches)
		batches.GET("/:id", h.aaaMiddleware.RequirePermission("aaa/inventory_batch", "*", "read"), h.GetBatch)
		batches.GET("/:id/transactions", h.aaaMiddleware.RequirePermission("aaa/inventory_transaction", "*", "read"), h.GetTransactionsByBatch)
	}

	// Warehouse batch routes
	warehouses := router.Group("/warehouses")
	{
		warehouses.Use(h.aaaMiddleware.Authenticate())
		warehouses.GET("/:id/batches", h.aaaMiddleware.RequirePermission("aaa/inventory_batch", "*", "read"), h.GetBatchesByWarehouse)
	}

	// Product batch routes
	products := router.Group("/products")
	{
		products.Use(h.aaaMiddleware.Authenticate())
		products.GET("/:id/batches", h.aaaMiddleware.RequirePermission("aaa/inventory_batch", "*", "read"), h.GetBatchesByProduct)
	}
	// Protected product availability route
	protected := products.Group("")
	{
		protected.GET("/availability", h.aaaMiddleware.RequirePermission("aaa/inventory_batch", "*", "read"), h.GetAllProductsAvailability)
	}
}
