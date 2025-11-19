// internal/api/handlers/discounts_handler.go
package handlers

import (
	"strconv"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	logger "kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DiscountsHandler struct {
	discountsService interfaces.DiscountsServiceInterface
	aaaMiddleware    *aaa.AAAMiddleware
	logger           logger.Logger
}

func NewDiscountsHandler(discountsService interfaces.DiscountsServiceInterface, aaaMiddleware *aaa.AAAMiddleware, logger logger.Logger) *DiscountsHandler {
	return &DiscountsHandler{
		discountsService: discountsService,
		aaaMiddleware:    aaaMiddleware,
		logger:           logger,
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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts [post]
func (h *DiscountsHandler) CreateDiscount(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Creating discount",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	var req models.CreateDiscountRequest

	// Validate request
	if err := utils.ValidateRequest(c, &req); err != nil {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for create discount",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling CreateDiscount service",
		zap.String("discount_type", string(req.DiscountType)))

	discount, err := h.discountsService.CreateDiscount(&req)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to create discount",
			zap.Error(err),
			zap.String("discount_type", string(req.DiscountType)))
		utils.HandleServiceError(c, "Failed to create discount", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Discount created successfully",
		zap.String("discount_id", discount.ID),
		zap.String("discount_type", string(discount.DiscountType)))

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
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Discount not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/{id} [get]
func (h *DiscountsHandler) GetDiscount(c *gin.Context) {
	// 1. Entry Log
	id := c.Param("id")
	h.logger.Info("Getting discount",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("discount_id", id))

	if id == "" {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for get discount",
			zap.String("error", "discount ID is required"))
		utils.BadRequestResponse(c, "Discount ID is required", nil)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling GetDiscount service",
		zap.String("discount_id", id))

	discount, err := h.discountsService.GetDiscount(id)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to get discount",
			zap.Error(err),
			zap.String("discount_id", id))
		utils.NotFoundResponse(c, "Discount not found")
		return
	}

	// 5. Success Log
	h.logger.Info("Discount retrieved successfully",
		zap.String("discount_id", discount.ID),
		zap.String("discount_type", string(discount.DiscountType)))

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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts [get]
func (h *DiscountsHandler) GetAllDiscounts(c *gin.Context) {
	// 1. Entry Log
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	h.logger.Info("Getting all discounts",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("limit", limitStr),
		zap.String("offset", offsetStr))

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for get all discounts",
			zap.Error(err),
			zap.String("limit", limitStr))
		utils.BadRequestResponse(c, "Invalid limit parameter", err)
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for get all discounts",
			zap.Error(err),
			zap.String("offset", offsetStr))
		utils.BadRequestResponse(c, "Invalid offset parameter", err)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling GetAllDiscounts service",
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	discounts, err := h.discountsService.GetAllDiscounts(limit, offset)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to retrieve discounts",
			zap.Error(err),
			zap.Int("limit", limit),
			zap.Int("offset", offset))
		utils.HandleServiceError(c, "Failed to retrieve discounts", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Discounts retrieved successfully",
		zap.Int("count", len(discounts)),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	utils.OKResponse(c, "Discounts retrieved successfully", discounts)
}

// GetActiveDiscounts handles GET /api/v1/discounts/active
// @Summary Get Active Discounts
// @Description Retrieve all currently active discounts
// @Tags Discounts
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.DiscountResponse} "Active discounts retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/active [get]
func (h *DiscountsHandler) GetActiveDiscounts(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Getting active discounts",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// 3. Service Call Log
	h.logger.Debug("Calling GetActiveDiscounts service")

	discounts, err := h.discountsService.GetActiveDiscounts()
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to retrieve active discounts",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve active discounts", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Active discounts retrieved successfully",
		zap.Int("count", len(discounts)))

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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Discount not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/{id} [put]
func (h *DiscountsHandler) UpdateDiscount(c *gin.Context) {
	// 1. Entry Log
	id := c.Param("id")
	h.logger.Info("Updating discount",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("discount_id", id))

	if id == "" {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for update discount",
			zap.String("error", "discount ID is required"))
		utils.BadRequestResponse(c, "Discount ID is required", nil)
		return
	}

	var req models.UpdateDiscountRequest
	if err := utils.ValidatePartialRequest(c, &req); err != nil {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for update discount",
			zap.Error(err),
			zap.String("discount_id", id))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling UpdateDiscount service",
		zap.String("discount_id", id))

	discount, err := h.discountsService.UpdateDiscount(id, &req)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to update discount",
			zap.Error(err),
			zap.String("discount_id", id))
		utils.HandleServiceError(c, "Failed to update discount", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Discount updated successfully",
		zap.String("discount_id", discount.ID))

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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Discount not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/{id} [delete]
func (h *DiscountsHandler) DeleteDiscount(c *gin.Context) {
	// 1. Entry Log
	id := c.Param("id")
	h.logger.Info("Deleting discount",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("discount_id", id))

	if id == "" {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for delete discount",
			zap.String("error", "discount ID is required"))
		utils.BadRequestResponse(c, "Discount ID is required", nil)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling DeleteDiscount service",
		zap.String("discount_id", id))

	err := h.discountsService.DeleteDiscount(id)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to delete discount",
			zap.Error(err),
			zap.String("discount_id", id))
		utils.HandleServiceError(c, "Failed to delete discount", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Discount deleted successfully",
		zap.String("discount_id", id))

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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/type/{type} [get]
func (h *DiscountsHandler) GetDiscountsByType(c *gin.Context) {
	// 1. Entry Log
	discountType := c.Param("type")
	h.logger.Info("Getting discounts by type",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("discount_type", discountType))

	if discountType == "" {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for get discounts by type",
			zap.String("error", "discount type is required"))
		utils.BadRequestResponse(c, "Discount type is required", nil)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling GetDiscountsByType service",
		zap.String("discount_type", discountType))

	discounts, err := h.discountsService.GetDiscountsByType(models.DiscountType(discountType))
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to retrieve discounts by type",
			zap.Error(err),
			zap.String("discount_type", discountType))
		utils.HandleServiceError(c, "Failed to retrieve discounts by type", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Discounts retrieved successfully by type",
		zap.String("discount_type", discountType),
		zap.Int("count", len(discounts)))

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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/status/{status} [get]
func (h *DiscountsHandler) GetDiscountsByStatus(c *gin.Context) {
	// 1. Entry Log
	status := c.Param("status")
	h.logger.Info("Getting discounts by status",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("status", status))

	if status == "" {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for get discounts by status",
			zap.String("error", "status is required"))
		utils.BadRequestResponse(c, "Status is required", nil)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling GetDiscountsByStatus service",
		zap.String("status", status))

	discounts, err := h.discountsService.GetDiscountsByStatus(status)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to retrieve discounts by status",
			zap.Error(err),
			zap.String("status", status))
		utils.HandleServiceError(c, "Failed to retrieve discounts by status", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Discounts retrieved successfully by status",
		zap.String("status", status),
		zap.Int("count", len(discounts)))

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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/validate [post]
func (h *DiscountsHandler) ValidateDiscount(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Validating discount",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	var req models.ValidateDiscountRequest

	// Validate request
	if err := utils.ValidateRequest(c, &req); err != nil {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for validate discount",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling ValidateDiscount service",
		zap.Float64("order_value", req.OrderValue),
		zap.String("warehouse_id", req.WarehouseID))

	validation, err := h.discountsService.ValidateDiscount(&req)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to validate discount",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to validate discount", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Discount validation completed",
		zap.Bool("is_valid", validation.IsValid))

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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/usage/sale/{saleID} [get]
func (h *DiscountsHandler) GetDiscountUsageBySale(c *gin.Context) {
	// 1. Entry Log
	saleID := c.Param("saleID")
	h.logger.Info("Getting discount usage by sale",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("sale_id", saleID))

	if saleID == "" {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for get discount usage by sale",
			zap.String("error", "sale ID is required"))
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling GetDiscountUsageBySale service",
		zap.String("sale_id", saleID))

	usages, err := h.discountsService.GetDiscountUsageBySale(saleID)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to retrieve discount usage",
			zap.Error(err),
			zap.String("sale_id", saleID))
		utils.HandleServiceError(c, "Failed to retrieve discount usage", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Discount usage retrieved successfully",
		zap.String("sale_id", saleID),
		zap.Int("count", len(usages)))

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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/applicable [get]
func (h *DiscountsHandler) GetApplicableDiscounts(c *gin.Context) {
	// 1. Entry Log
	orderValueStr := c.Query("order_value")
	warehouseID := c.Query("warehouse_id")
	h.logger.Info("Getting applicable discounts",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("order_value", orderValueStr),
		zap.String("warehouse_id", warehouseID))

	if orderValueStr == "" {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for get applicable discounts",
			zap.String("error", "order value is required"))
		utils.BadRequestResponse(c, "Order value is required", nil)
		return
	}

	orderValue, err := strconv.ParseFloat(orderValueStr, 64)
	if err != nil || orderValue <= 0 {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for get applicable discounts",
			zap.Error(err),
			zap.String("order_value", orderValueStr))
		utils.BadRequestResponse(c, "Invalid order value", err)
		return
	}

	if warehouseID == "" {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for get applicable discounts",
			zap.String("error", "warehouse ID is required"))
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

	// 3. Service Call Log
	h.logger.Debug("Calling GetApplicableDiscountsForOrder service",
		zap.Float64("order_value", orderValue),
		zap.String("warehouse_id", warehouseID),
		zap.Int("product_count", len(productIDs)),
		zap.Int("category_count", len(categoryIDs)))

	discounts, err := h.discountsService.GetApplicableDiscountsForOrder(orderValue, productIDs, categoryIDs, warehouseID)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to retrieve applicable discounts",
			zap.Error(err),
			zap.Float64("order_value", orderValue),
			zap.String("warehouse_id", warehouseID))
		utils.HandleServiceError(c, "Failed to retrieve applicable discounts", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Applicable discounts retrieved successfully",
		zap.Int("count", len(discounts)),
		zap.Float64("order_value", orderValue))

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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/calculate-optimal [post]
func (h *DiscountsHandler) CalculateOptimalDiscounts(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Calculating optimal discounts",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	var req models.ValidateDiscountRequest

	// Validate request
	if err := utils.ValidateRequest(c, &req); err != nil {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for calculate optimal discounts",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling CalculateOptimalDiscounts service",
		zap.Float64("order_value", req.OrderValue),
		zap.String("warehouse_id", req.WarehouseID))

	optimalDiscounts, totalDiscount, err := h.discountsService.CalculateOptimalDiscounts(req.OrderValue, req.ProductIDs, req.CategoryIDs, req.WarehouseID)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to calculate optimal discounts",
			zap.Error(err),
			zap.Float64("order_value", req.OrderValue))
		utils.HandleServiceError(c, "Failed to calculate optimal discounts", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Optimal discounts calculated successfully",
		zap.Int("discount_count", len(optimalDiscounts)),
		zap.Float64("total_discount", totalDiscount))

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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Discount not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/discounts/usage/stats/{discountID} [get]
func (h *DiscountsHandler) GetDiscountUsageStats(c *gin.Context) {
	// 1. Entry Log
	discountID := c.Param("discountID")
	h.logger.Info("Getting discount usage statistics",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("discount_id", discountID))

	if discountID == "" {
		// 2. Validation Error Log
		h.logger.Error("Validation failed for get discount usage stats",
			zap.String("error", "discount ID is required"))
		utils.BadRequestResponse(c, "Discount ID is required", nil)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling GetDiscountUsageStats service",
		zap.String("discount_id", discountID))

	stats, err := h.discountsService.GetDiscountUsageStats(discountID)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Failed to retrieve discount usage statistics",
			zap.Error(err),
			zap.String("discount_id", discountID))
		utils.HandleServiceError(c, "Failed to retrieve discount usage statistics", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Discount usage statistics retrieved successfully",
		zap.String("discount_id", discountID))

	utils.OKResponse(c, "Discount usage statistics retrieved successfully", stats)
}

// RegisterRoutes registers all discount routes
func (h *DiscountsHandler) RegisterRoutes(router *gin.RouterGroup) {
	discounts := router.Group("/discounts")
	{
		// Apply authentication middleware
		discounts.Use(h.aaaMiddleware.Authenticate())

		// Create/Update/Delete routes - CEO=CRUD, Store_Staff=CRUD, Tech_Support=R/W (temp)
		discounts.POST("", h.aaaMiddleware.RequireOrgPermission("discount", "create"), h.CreateDiscount)
		discounts.PUT("/:id", h.aaaMiddleware.RequireOrgPermission("discount", "update"), h.UpdateDiscount)
		discounts.DELETE("/:id", h.aaaMiddleware.RequireOrgPermission("discount", "delete"), h.DeleteDiscount)

		// Read routes - Director=R, CEO=CRUD, Auditor=R, Accountant=R, Tech_Support=R/W (temp), Store_Manager=R, Store_Staff=CRUD
		discounts.GET("", h.aaaMiddleware.RequireOrgPermission("discount", "read"), h.GetAllDiscounts)
		discounts.GET("/active", h.aaaMiddleware.RequireOrgPermission("discount", "read"), h.GetActiveDiscounts)
		discounts.GET("/applicable", h.aaaMiddleware.RequireOrgPermission("discount", "read"), h.GetApplicableDiscounts)
		discounts.GET("/:id", h.aaaMiddleware.RequireOrgPermission("discount", "read"), h.GetDiscount)
		discounts.GET("/type/:type", h.aaaMiddleware.RequireOrgPermission("discount", "read"), h.GetDiscountsByType)
		discounts.GET("/status/:status", h.aaaMiddleware.RequireOrgPermission("discount", "read"), h.GetDiscountsByStatus)
		discounts.GET("/usage/sale/:saleID", h.aaaMiddleware.RequireOrgPermission("discount", "read"), h.GetDiscountUsageBySale)
		discounts.GET("/usage/stats/:discountID", h.aaaMiddleware.RequireOrgPermission("discount", "read"), h.GetDiscountUsageStats)

		// Validation and calculation routes - accessible to all authenticated users
		discounts.POST("/validate", h.aaaMiddleware.RequireOrgPermission("discount", "read"), h.ValidateDiscount)
		discounts.POST("/calculate-optimal", h.aaaMiddleware.RequireOrgPermission("discount", "read"), h.CalculateOptimalDiscounts)
	}
}
