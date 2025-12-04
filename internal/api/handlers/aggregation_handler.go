package handlers

import (
	"net/http"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	logger "kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AggregationHandler handles aggregated API requests for frontend optimization
type AggregationHandler struct {
	service       *services.AggregationService
	aaaMiddleware *aaa.AAAMiddleware
	logger        logger.Logger
}

// NewAggregationHandler creates a new aggregation handler
func NewAggregationHandler(service *services.AggregationService, aaaMiddleware *aaa.AAAMiddleware, logger logger.Logger) *AggregationHandler {
	return &AggregationHandler{
		service:       service,
		aaaMiddleware: aaaMiddleware,
		logger:        logger,
	}
}

// GetProductDetail handles GET /api/v1/products/:id/detail
// @Summary Get Aggregated Product Detail
// @Description Retrieves complete product information including variants, prices, inventory availability, and collaborator details in a single call. Reduces API calls by 75%.
// @Tags Aggregation
// @Produce json
// @Param id path string true "Product ID" example(PROD_12345678)
// @Param include query string false "Comma-separated list of data to include: variants,prices,inventory,collaborators,taxes. Default: all" example(variants,prices,inventory)
// @Param warehouse_id query string false "Filter inventory by specific warehouse" example(WH_001)
// @Param price_type query string false "Filter prices: retail, wholesale, bulk, or all. Default: all" example(retail)
// @Param active_only query bool false "Show only active variants. Default: true" example(true)
// @Param in_stock_only query bool false "Show only variants with available stock. Default: false" example(false)
// @Success 200 {object} utils.Response{data=models.ProductDetailResponse} "Product details retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request - invalid parameters"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Product not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/{id}/detail [get]
func (h *AggregationHandler) GetProductDetail(c *gin.Context) {
	h.logger.Info("Handling get product detail request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	productID := c.Param("id")
	if productID == "" {
		h.logger.Error("Product ID is required")
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	// Parse query parameters
	req := &models.ProductDetailRequest{
		Include:     c.DefaultQuery("include", "all"),
		WarehouseID: c.Query("warehouse_id"),
		PriceType:   c.DefaultQuery("price_type", "all"),
		ActiveOnly:  c.DefaultQuery("active_only", "true") == "true",
		InStockOnly: c.DefaultQuery("in_stock_only", "false") == "true",
	}

	h.logger.Debug("Product detail request parameters",
		zap.String("product_id", productID),
		zap.String("include", req.Include),
		zap.String("warehouse_id", req.WarehouseID),
		zap.String("price_type", req.PriceType),
		zap.Bool("active_only", req.ActiveOnly),
		zap.Bool("in_stock_only", req.InStockOnly))

	response, err := h.service.GetProductDetail(productID, req)
	if err != nil {
		h.logger.Error("Failed to get product detail",
			zap.String("product_id", productID),
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to get product detail", err)
		return
	}

	h.logger.Info("Product detail retrieved successfully",
		zap.String("product_id", productID),
		zap.Int("variants_count", len(response.Variants)))

	utils.OKResponse(c, "Product details retrieved successfully", response)
}

// GetVariantDetail handles GET /api/v1/products/variants/:id/detail
// @Summary Get Aggregated Variant Detail
// @Description Retrieves complete information for a specific product variant including prices, inventory, and parent product/collaborator info.
// @Tags Aggregation
// @Produce json
// @Param id path string true "Variant ID" example(PVAR_12345678)
// @Param include query string false "Comma-separated list: prices,inventory,product,collaborator,taxes. Default: all" example(prices,inventory)
// @Param warehouse_id query string false "Filter inventory by specific warehouse" example(WH_001)
// @Success 200 {object} utils.Response{data=models.VariantDetailResponse} "Variant details retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request - invalid parameters"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Variant not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/variants/{id}/detail [get]
func (h *AggregationHandler) GetVariantDetail(c *gin.Context) {
	h.logger.Info("Handling get variant detail request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	variantID := c.Param("id")
	if variantID == "" {
		h.logger.Error("Variant ID is required")
		utils.BadRequestResponse(c, "Variant ID is required", nil)
		return
	}

	include := c.DefaultQuery("include", "all")
	warehouseID := c.Query("warehouse_id")

	h.logger.Debug("Variant detail request parameters",
		zap.String("variant_id", variantID),
		zap.String("include", include),
		zap.String("warehouse_id", warehouseID))

	response, err := h.service.GetVariantDetail(variantID, include, warehouseID)
	if err != nil {
		h.logger.Error("Failed to get variant detail",
			zap.String("variant_id", variantID),
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to get variant detail", err)
		return
	}

	h.logger.Info("Variant detail retrieved successfully",
		zap.String("variant_id", variantID))

	utils.OKResponse(c, "Variant details retrieved successfully", response)
}

// GetSalesContext handles GET /api/v1/sales/context
// @Summary Get Sales Context
// @Description Retrieves all data needed for sale/checkout operations in a single call including available inventory, prices, taxes, discounts, and payment methods. Reduces API calls by 80-83%.
// @Tags Aggregation
// @Produce json
// @Param warehouse_id query string false "Filter inventory to specific warehouse" example(WH_001)
// @Param include_zero_stock query bool false "Include products with zero stock. Default: false" example(false)
// @Param price_type query string false "Price type to use: retail, wholesale, bulk. Default: retail" example(retail)
// @Param effective_date query string false "ISO date for price effective date. Default: now" example(2024-11-21)
// @Success 200 {object} utils.Response{data=models.SalesContextResponse} "Sales context retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request - invalid parameters"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Warehouse not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/context [get]
func (h *AggregationHandler) GetSalesContext(c *gin.Context) {
	h.logger.Info("Handling get sales context request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Parse query parameters
	req := &models.SalesContextRequest{
		WarehouseID:      c.Query("warehouse_id"),
		IncludeZeroStock: c.DefaultQuery("include_zero_stock", "false") == "true",
		PriceType:        c.DefaultQuery("price_type", "retail"),
		EffectiveDate:    c.Query("effective_date"),
	}

	h.logger.Debug("Sales context request parameters",
		zap.String("warehouse_id", req.WarehouseID),
		zap.Bool("include_zero_stock", req.IncludeZeroStock),
		zap.String("price_type", req.PriceType),
		zap.String("effective_date", req.EffectiveDate))

	response, err := h.service.GetSalesContext(req)
	if err != nil {
		h.logger.Error("Failed to get sales context",
			zap.String("warehouse_id", req.WarehouseID),
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to get sales context", err)
		return
	}

	h.logger.Info("Sales context retrieved successfully",
		zap.Int("inventory_count", len(response.AvailableInventory)),
		zap.Int("discount_count", len(response.DiscountPolicies)))

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Sales context retrieved successfully",
		"data":    response,
	})
}

// RegisterRoutes registers the aggregation routes
func (h *AggregationHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Product detail aggregation - adds to products group
	products := router.Group("/products")
	{
		products.Use(h.aaaMiddleware.Authenticate())

		// Aggregated product detail endpoint
		products.GET("/:id/detail", h.aaaMiddleware.RequireOrgPermission("product", "read"), h.GetProductDetail)

		// Aggregated variant detail endpoint
		products.GET("/variants/:id/detail", h.aaaMiddleware.RequireOrgPermission("product", "read"), h.GetVariantDetail)
	}

	// Sales context aggregation - adds to sales group
	sales := router.Group("/sales")
	{
		sales.Use(h.aaaMiddleware.Authenticate())

		// Sales context endpoint - requires sale:create permission
		sales.GET("/context", h.aaaMiddleware.RequireOrgPermission("sale", "create"), h.GetSalesContext)
	}
}
