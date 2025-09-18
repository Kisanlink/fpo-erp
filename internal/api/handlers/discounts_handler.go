// internal/api/handlers/discounts_handler.go
package handlers

import (
	"strconv"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

type DiscountsHandler struct {
	discountsService *services.DiscountsService
	aaaMiddleware    *aaa.AAAMiddleware
}

func NewDiscountsHandler(discountsService *services.DiscountsService, aaaMiddleware *aaa.AAAMiddleware) *DiscountsHandler {
	return &DiscountsHandler{
		discountsService: discountsService,
		aaaMiddleware:    aaaMiddleware,
	}
}

// CreateDiscount handles POST /api/v1/discounts
// @Summary Create Discount
// @Description Create a new discount (requires authentication)
// @Tags Discounts
// @Accept json
// @Produce json
// @Param request body models.CreateDiscountRequest true "Discount data"
// @Success 201 {object} utils.Response{data=models.DiscountResponse} "Discount created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts [post]
func (h *DiscountsHandler) CreateDiscount(c *gin.Context) {
	var req models.CreateDiscountRequest

	// Validate request
	if err := utils.ValidateRequest(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	discount, err := h.discountsService.CreateDiscount(&req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create discount", err)
		return
	}

	utils.CreatedResponse(c, "Discount created successfully", discount)
}

// GetDiscount handles GET /api/v1/discounts/:id
// @Summary Get Discount
// @Description Retrieve a specific discount by ID
// @Tags Discounts
// @Produce json
// @Param id path string true "Discount ID" example(DISC_12345678)
// @Success 200 {object} utils.Response{data=models.DiscountResponse} "Discount retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Discount not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/{id} [get]
func (h *DiscountsHandler) GetDiscount(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Discount ID is required", nil)
		return
	}

	discount, err := h.discountsService.GetDiscount(id)
	if err != nil {
		utils.NotFoundResponse(c, "Discount not found")
		return
	}

	utils.OKResponse(c, "Discount retrieved successfully", discount)
}

// GetAllDiscounts handles GET /api/v1/discounts
// @Summary Get All Discounts
// @Description Retrieve all discounts with pagination
// @Tags Discounts
// @Produce json
// @Param limit query integer false "Number of records to return (default: 10)" example(10)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.Response{data=[]models.DiscountResponse} "Discounts retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts [get]
func (h *DiscountsHandler) GetAllDiscounts(c *gin.Context) {
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

	discounts, err := h.discountsService.GetAllDiscounts(limit, offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve discounts", err)
		return
	}

	utils.OKResponse(c, "Discounts retrieved successfully", discounts)
}

// GetActiveDiscounts handles GET /api/v1/discounts/active
// @Summary Get Active Discounts
// @Description Retrieve all currently active discounts
// @Tags Discounts
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.DiscountResponse} "Active discounts retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/active [get]
func (h *DiscountsHandler) GetActiveDiscounts(c *gin.Context) {
	discounts, err := h.discountsService.GetActiveDiscounts()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve active discounts", err)
		return
	}

	utils.OKResponse(c, "Active discounts retrieved successfully", discounts)
}

// UpdateDiscount handles PUT /api/v1/discounts/:id
// @Summary Update Discount
// @Description Update an existing discount by ID
// @Tags Discounts
// @Accept json
// @Produce json
// @Param id path string true "Discount ID" example(DISC_12345678)
// @Param request body models.UpdateDiscountRequest true "Updated discount data"
// @Success 200 {object} utils.Response{data=models.DiscountResponse} "Discount updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Discount not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/{id} [put]
func (h *DiscountsHandler) UpdateDiscount(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Discount ID is required", nil)
		return
	}

	var req models.UpdateDiscountRequest
	if err := utils.ValidatePartialRequest(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	discount, err := h.discountsService.UpdateDiscount(id, &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update discount", err)
		return
	}

	utils.OKResponse(c, "Discount updated successfully", discount)
}

// DeleteDiscount handles DELETE /api/v1/discounts/:id
// @Summary Delete Discount
// @Description Delete a discount by ID
// @Tags Discounts
// @Produce json
// @Param id path string true "Discount ID" example(DISC_12345678)
// @Success 200 {object} utils.Response "Discount deleted successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Discount not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/{id} [delete]
func (h *DiscountsHandler) DeleteDiscount(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Discount ID is required", nil)
		return
	}

	err := h.discountsService.DeleteDiscount(id)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete discount", err)
		return
	}

	utils.OKResponse(c, "Discount deleted successfully", nil)
}

// GetDiscountsByType handles GET /api/v1/discounts/type/:type
// @Summary Get Discounts by Type
// @Description Retrieve all discounts of a specific type
// @Tags Discounts
// @Produce json
// @Param type path string true "Discount type" example(percentage)
// @Success 200 {object} utils.Response{data=[]models.DiscountResponse} "Discounts retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/type/{type} [get]
func (h *DiscountsHandler) GetDiscountsByType(c *gin.Context) {
	discountType := c.Param("type")
	if discountType == "" {
		utils.BadRequestResponse(c, "Discount type is required", nil)
		return
	}

	discounts, err := h.discountsService.GetDiscountsByType(models.DiscountType(discountType))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve discounts by type", err)
		return
	}

	utils.OKResponse(c, "Discounts retrieved successfully", discounts)
}

// GetDiscountsByStatus handles GET /api/v1/discounts/status/:status
// @Summary Get Discounts by Status
// @Description Retrieve all discounts with a specific status
// @Tags Discounts
// @Produce json
// @Param status path string true "Discount status" example(active)
// @Success 200 {object} utils.Response{data=[]models.DiscountResponse} "Discounts retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/status/{status} [get]
func (h *DiscountsHandler) GetDiscountsByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		utils.BadRequestResponse(c, "Status is required", nil)
		return
	}

	discounts, err := h.discountsService.GetDiscountsByStatus(status)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve discounts by status", err)
		return
	}

	utils.OKResponse(c, "Discounts retrieved successfully", discounts)
}

// ValidateDiscount handles POST /api/v1/discounts/validate
// @Summary Validate Discount
// @Description Validate if a discount can be applied to a transaction
// @Tags Discounts
// @Accept json
// @Produce json
// @Param request body models.ValidateDiscountRequest true "Discount validation data"
// @Success 200 {object} utils.Response{data=models.DiscountValidationResponse} "Discount validation completed"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/validate [post]
func (h *DiscountsHandler) ValidateDiscount(c *gin.Context) {
	var req models.ValidateDiscountRequest

	// Validate request
	if err := utils.ValidateRequest(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	validation, err := h.discountsService.ValidateDiscount(&req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to validate discount", err)
		return
	}

	utils.OKResponse(c, "Discount validation completed", validation)
}

// GetDiscountUsageBySale handles GET /api/v1/discounts/usage/sale/:saleID
// @Summary Get Discount Usage by Sale
// @Description Retrieve all discount usage for a specific sale
// @Tags Discounts
// @Produce json
// @Param saleID path string true "Sale ID" example(SALE_12345678)
// @Success 200 {object} utils.Response{data=[]models.DiscountUsageResponse} "Discount usage retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/usage/sale/{saleID} [get]
func (h *DiscountsHandler) GetDiscountUsageBySale(c *gin.Context) {
	saleID := c.Param("saleID")
	if saleID == "" {
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	usages, err := h.discountsService.GetDiscountUsageBySale(saleID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve discount usage", err)
		return
	}

	utils.OKResponse(c, "Discount usage retrieved successfully", usages)
}

// GetApplicableDiscounts handles GET /api/v1/discounts/applicable
// @Summary Get Applicable Discounts
// @Description Retrieve all discounts applicable for an order with given criteria
// @Tags Discounts
// @Produce json
// @Param order_value query number true "Order value" example(1000.00)
// @Param product_ids query string false "Comma-separated product IDs" example("PROD_1,PROD_2")
// @Param category_ids query string false "Comma-separated category IDs" example("CAT_1,CAT_2")
// @Param warehouse_id query string true "Warehouse ID" example("WH_123")
// @Success 200 {object} utils.Response{data=[]models.DiscountResponse} "Applicable discounts retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/applicable [get]
func (h *DiscountsHandler) GetApplicableDiscounts(c *gin.Context) {
	orderValueStr := c.Query("order_value")
	if orderValueStr == "" {
		utils.BadRequestResponse(c, "Order value is required", nil)
		return
	}

	orderValue, err := strconv.ParseFloat(orderValueStr, 64)
	if err != nil || orderValue <= 0 {
		utils.BadRequestResponse(c, "Invalid order value", err)
		return
	}

	warehouseID := c.Query("warehouse_id")
	if warehouseID == "" {
		utils.BadRequestResponse(c, "Warehouse ID is required", nil)
		return
	}

	// Parse optional parameters
	var productIDs []string
	if productIDsStr := c.Query("product_ids"); productIDsStr != "" {
		productIDs = utils.ParseCommaSeparatedString(productIDsStr)
	}

	var categoryIDs []string
	if categoryIDsStr := c.Query("category_ids"); categoryIDsStr != "" {
		categoryIDs = utils.ParseCommaSeparatedString(categoryIDsStr)
	}

	discounts, err := h.discountsService.GetApplicableDiscountsForOrder(orderValue, productIDs, categoryIDs, warehouseID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve applicable discounts", err)
		return
	}

	utils.OKResponse(c, "Applicable discounts retrieved successfully", discounts)
}

// CalculateOptimalDiscounts handles POST /api/v1/discounts/calculate-optimal
// @Summary Calculate Optimal Discounts
// @Description Calculate the optimal combination of discounts for an order
// @Tags Discounts
// @Accept json
// @Produce json
// @Param request body models.ValidateDiscountRequest true "Order details for optimization"
// @Success 200 {object} utils.Response{data=[]models.DiscountResponse} "Optimal discounts calculated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/calculate-optimal [post]
func (h *DiscountsHandler) CalculateOptimalDiscounts(c *gin.Context) {
	var req models.ValidateDiscountRequest

	// Validate request
	if err := utils.ValidateRequest(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	optimalDiscounts, totalDiscount, err := h.discountsService.CalculateOptimalDiscounts(req.OrderValue, req.ProductIDs, req.CategoryIDs, req.WarehouseID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to calculate optimal discounts", err)
		return
	}

	utils.OKResponse(c, "Optimal discounts calculated successfully", gin.H{
		"discounts":      optimalDiscounts,
		"total_discount": totalDiscount,
	})
}

// GetDiscountUsageStats handles GET /api/v1/discounts/usage/stats/:discountID
// @Summary Get Discount Usage Statistics
// @Description Retrieve usage statistics for a specific discount
// @Tags Discounts
// @Produce json
// @Param discountID path string true "Discount ID" example(DISC_12345678)
// @Success 200 {object} utils.Response{data=map[string]interface{}} "Discount usage statistics retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Discount not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/usage/stats/{discountID} [get]
func (h *DiscountsHandler) GetDiscountUsageStats(c *gin.Context) {
	discountID := c.Param("discountID")
	if discountID == "" {
		utils.BadRequestResponse(c, "Discount ID is required", nil)
		return
	}

	stats, err := h.discountsService.GetDiscountUsageStats(discountID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve discount usage statistics", err)
		return
	}

	utils.OKResponse(c, "Discount usage statistics retrieved successfully", stats)
}

// RegisterRoutes registers all discount routes
func (h *DiscountsHandler) RegisterRoutes(router *gin.RouterGroup) {
	discounts := router.Group("/discounts")
	{
		// Apply authentication middleware
		discounts.Use(h.aaaMiddleware.Authenticate())

		// Create/Update/Delete routes - CEO=CRUD, Store_Staff=CRUD, Tech_Support=R/W (temp)
		discounts.POST("", h.aaaMiddleware.RequirePermission("aaa/discount", "*", "create"), h.CreateDiscount)
		discounts.PUT("/:id", h.aaaMiddleware.RequirePermission("aaa/discount", "*", "update"), h.UpdateDiscount)
		discounts.DELETE("/:id", h.aaaMiddleware.RequirePermission("aaa/discount", "*", "delete"), h.DeleteDiscount)

		// Read routes - Director=R, CEO=CRUD, Auditor=R, Accountant=R, Tech_Support=R/W (temp), Store_Manager=R, Store_Staff=CRUD
		discounts.GET("", h.aaaMiddleware.RequirePermission("aaa/discount", "*", "read"), h.GetAllDiscounts)
		discounts.GET("/active", h.aaaMiddleware.RequirePermission("aaa/discount", "*", "read"), h.GetActiveDiscounts)
		discounts.GET("/applicable", h.aaaMiddleware.RequirePermission("aaa/discount", "*", "read"), h.GetApplicableDiscounts)
		discounts.GET("/:id", h.aaaMiddleware.RequirePermission("aaa/discount", "*", "read"), h.GetDiscount)
		discounts.GET("/type/:type", h.aaaMiddleware.RequirePermission("aaa/discount", "*", "read"), h.GetDiscountsByType)
		discounts.GET("/status/:status", h.aaaMiddleware.RequirePermission("aaa/discount", "*", "read"), h.GetDiscountsByStatus)
		discounts.GET("/usage/sale/:saleID", h.aaaMiddleware.RequirePermission("aaa/discount", "*", "read"), h.GetDiscountUsageBySale)
		discounts.GET("/usage/stats/:discountID", h.aaaMiddleware.RequirePermission("aaa/discount", "*", "read"), h.GetDiscountUsageStats)

		// Validation and calculation routes - accessible to all authenticated users
		discounts.POST("/validate", h.aaaMiddleware.RequirePermission("aaa/discount", "*", "read"), h.ValidateDiscount)
		discounts.POST("/calculate-optimal", h.aaaMiddleware.RequirePermission("aaa/discount", "*", "read"), h.CalculateOptimalDiscounts)
	}
}
