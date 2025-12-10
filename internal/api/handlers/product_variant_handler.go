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

// ProductVariantHandler handles product variant HTTP requests
type ProductVariantHandler struct {
	variantService interfaces.ProductVariantServiceInterface
	aaaMiddleware  *aaa.AAAMiddleware
	logger         logger.Logger
}

// NewProductVariantHandler creates a new product variant handler
func NewProductVariantHandler(variantService interfaces.ProductVariantServiceInterface, aaaMiddleware *aaa.AAAMiddleware, logger logger.Logger) *ProductVariantHandler {
	return &ProductVariantHandler{
		variantService: variantService,
		aaaMiddleware:  aaaMiddleware,
		logger:         logger,
	}
}

// safeStringDeref safely dereferences a string pointer, returning a default value if nil
func safeStringDeref(ptr *string, defaultVal string) string {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}

// CreateProductVariant handles POST /api/v1/products/:id/variants
// @Summary Create Product Variant
// @Description Create a new variant for a product (requires authentication)
// @Tags Product Variants
// @Accept json
// @Produce json
// @Param id path string true "Product ID (format: PROD_xxxxxxxx)" example(PROD_12345678)
// @Param request body models.CreateProductVariantRequest true "Variant data"
// @Success 201 {object} utils.Response{data=models.ProductVariantResponse} "Variant created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/{id}/variants [post]
func (h *ProductVariantHandler) CreateProductVariant(c *gin.Context) {
	h.logger.Info("Handling create product variant request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get product ID from URL
	productID := c.Param("id")
	if productID == "" {
		h.logger.Error("Product ID is required but not provided")
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	var request models.CreateProductVariantRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		h.logger.Error("Invalid request body for create product variant",
			zap.Error(err),
			zap.String("product_id", productID))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	h.logger.Debug("Calling variant service to create product variant",
		zap.String("product_id", productID),
		zap.String("sku", safeStringDeref(request.SKU, "<none>")),
		zap.String("variant_name", request.VariantName))

	// Create variant
	response, err := h.variantService.CreateProductVariant(c.Request.Context(), productID, &request)
	if err != nil {
		h.logger.Error("Failed to create product variant via service",
			zap.Error(err),
			zap.String("product_id", productID),
			zap.String("sku", safeStringDeref(request.SKU, "<none>")))
		utils.HandleServiceError(c, "Failed to create variant", err)
		return
	}

	h.logger.Info("Product variant created successfully via handler",
		zap.String("variant_id", response.ID),
		zap.String("product_id", productID),
		zap.String("sku", safeStringDeref(response.SKU, "<none>")))

	utils.CreatedResponse(c, "Variant created successfully", response)
}

// GetProductVariant handles GET /api/v1/variants/:id
// @Summary Get Product Variant
// @Description Retrieve a specific product variant by ID
// @Tags Product Variants
// @Produce json
// @Param id path string true "Variant ID (format: PVAR_xxxxxxxx)" example(PVAR_12345678)
// @Success 200 {object} utils.Response{data=models.ProductVariantResponse} "Variant details"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Variant not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/variants/{id} [get]
func (h *ProductVariantHandler) GetProductVariant(c *gin.Context) {
	h.logger.Info("Handling get product variant request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Variant ID is required but not provided")
		utils.BadRequestResponse(c, "Variant ID is required", nil)
		return
	}

	h.logger.Debug("Calling variant service to get product variant",
		zap.String("variant_id", id))

	// Get variant
	response, err := h.variantService.GetProductVariant(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Product variant not found",
			zap.Error(err),
			zap.String("variant_id", id))
		utils.NotFoundResponse(c, "Variant not found")
		return
	}

	h.logger.Info("Product variant retrieved successfully via handler",
		zap.String("variant_id", response.ID),
		zap.String("sku", safeStringDeref(response.SKU, "<none>")))

	utils.OKResponse(c, "Variant retrieved successfully", response)
}

// GetVariantsByProduct handles GET /api/v1/products/:id/variants
// @Summary Get Variants by Product
// @Description Retrieve all variants for a specific product with pagination
// @Tags Product Variants
// @Produce json
// @Param id path string true "Product ID (format: PROD_xxxxxxxx)" example(PROD_12345678)
// @Param limit query integer false "Number of records to return (default: 50, max: 200)" example(50)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.PaginatedResponseModel{data=[]models.ProductVariantResponse} "Variants retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Product not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/products/{id}/variants [get]
func (h *ProductVariantHandler) GetVariantsByProduct(c *gin.Context) {
	h.logger.Info("Handling get variants by product request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get product ID from URL
	productID := c.Param("id")
	if productID == "" {
		h.logger.Error("Product ID is required but not provided")
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	h.logger.Debug("Calling variant service to get variants by product",
		zap.String("product_id", productID),
		zap.Int("limit", params.Limit),
		zap.Int("offset", params.Offset))

	// Get variants with pagination
	response, total, err := h.variantService.GetVariantsByProductPaginated(c.Request.Context(), productID, params.Limit, params.Offset)
	if err != nil {
		h.logger.Error("Failed to retrieve variants for product via service",
			zap.Error(err),
			zap.String("product_id", productID))
		utils.HandleServiceError(c, "Failed to retrieve variants", err)
		return
	}

	h.logger.Info("Product variants retrieved successfully via handler",
		zap.String("product_id", productID),
		zap.Int("count", len(response)),
		zap.Int64("total", total))

	utils.PaginatedOKResponse(c, response, total, params.Limit, params.Offset)
}

// GetVariantBySKU handles GET /api/v1/variants/sku/:sku
// @Summary Get Variant by SKU
// @Description Retrieve a product variant by SKU
// @Tags Product Variants
// @Produce json
// @Param sku path string true "Variant SKU"
// @Success 200 {object} utils.Response{data=models.ProductVariantResponse} "Variant details"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Variant not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/variants/sku/{sku} [get]
func (h *ProductVariantHandler) GetVariantBySKU(c *gin.Context) {
	h.logger.Info("Handling get variant by SKU request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get SKU from URL
	sku := c.Param("sku")
	if sku == "" {
		h.logger.Error("SKU is required but not provided")
		utils.BadRequestResponse(c, "SKU is required", nil)
		return
	}

	h.logger.Debug("Calling variant service to get variant by SKU",
		zap.String("sku", sku))

	// Get variant
	response, err := h.variantService.GetVariantBySKU(c.Request.Context(), sku)
	if err != nil {
		h.logger.Error("Variant not found by SKU",
			zap.Error(err),
			zap.String("sku", sku))
		utils.NotFoundResponse(c, "Variant not found")
		return
	}

	h.logger.Info("Variant retrieved successfully by SKU via handler",
		zap.String("variant_id", response.ID),
		zap.String("sku", sku))

	utils.OKResponse(c, "Variant retrieved successfully", response)
}

// GetVariantByBarcode handles GET /api/v1/variants/barcode/:barcode
// @Summary Get Variant by Barcode
// @Description Retrieve a product variant by barcode
// @Tags Product Variants
// @Produce json
// @Param barcode path string true "Variant Barcode"
// @Success 200 {object} utils.Response{data=models.ProductVariantResponse} "Variant details"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Variant not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/variants/barcode/{barcode} [get]
func (h *ProductVariantHandler) GetVariantByBarcode(c *gin.Context) {
	h.logger.Info("Handling get variant by barcode request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get barcode from URL
	barcode := c.Param("barcode")
	if barcode == "" {
		h.logger.Error("Barcode is required but not provided")
		utils.BadRequestResponse(c, "Barcode is required", nil)
		return
	}

	h.logger.Debug("Calling variant service to get variant by barcode",
		zap.String("barcode", barcode))

	// Get variant
	response, err := h.variantService.GetVariantByBarcode(c.Request.Context(), barcode)
	if err != nil {
		h.logger.Error("Variant not found by barcode",
			zap.Error(err),
			zap.String("barcode", barcode))
		utils.NotFoundResponse(c, "Variant not found")
		return
	}

	h.logger.Info("Variant retrieved successfully by barcode via handler",
		zap.String("variant_id", response.ID),
		zap.String("barcode", barcode))

	utils.OKResponse(c, "Variant retrieved successfully", response)
}

// UpdateProductVariant handles PUT /api/v1/variants/:id
// @Summary Update Product Variant
// @Description Update an existing product variant (requires authentication)
// @Tags Product Variants
// @Accept json
// @Produce json
// @Param id path string true "Variant ID (format: PVAR_xxxxxxxx)" example(PVAR_12345678)
// @Param request body models.UpdateProductVariantRequest true "Updated variant data"
// @Success 200 {object} utils.Response{data=models.ProductVariantResponse} "Variant updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Variant not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/variants/{id} [put]
func (h *ProductVariantHandler) UpdateProductVariant(c *gin.Context) {
	h.logger.Info("Handling update product variant request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Variant ID is required but not provided")
		utils.BadRequestResponse(c, "Variant ID is required", nil)
		return
	}

	var request models.UpdateProductVariantRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		h.logger.Error("Invalid request body for update product variant",
			zap.Error(err),
			zap.String("variant_id", id))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	h.logger.Debug("Calling variant service to update product variant",
		zap.String("variant_id", id))

	// Update variant
	response, err := h.variantService.UpdateProductVariant(c.Request.Context(), id, &request)
	if err != nil {
		h.logger.Error("Failed to update product variant via service",
			zap.Error(err),
			zap.String("variant_id", id))
		utils.HandleServiceError(c, "Failed to update variant", err)
		return
	}

	h.logger.Info("Product variant updated successfully via handler",
		zap.String("variant_id", response.ID),
		zap.String("sku", safeStringDeref(response.SKU, "<none>")))

	utils.OKResponse(c, "Variant updated successfully", response)
}

// DeleteProductVariant handles DELETE /api/v1/variants/:id
// @Summary Delete Product Variant
// @Description Delete a product variant (soft delete, requires authentication)
// @Tags Product Variants
// @Produce json
// @Param id path string true "Variant ID (format: PVAR_xxxxxxxx)" example(PVAR_12345678)
// @Success 200 {object} utils.Response "Variant deleted successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Variant not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/variants/{id} [delete]
func (h *ProductVariantHandler) DeleteProductVariant(c *gin.Context) {
	h.logger.Info("Handling delete product variant request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Variant ID is required but not provided")
		utils.BadRequestResponse(c, "Variant ID is required", nil)
		return
	}

	h.logger.Debug("Calling variant service to delete product variant",
		zap.String("variant_id", id))

	// Delete variant
	if err := h.variantService.DeleteProductVariant(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete product variant via service",
			zap.Error(err),
			zap.String("variant_id", id))
		utils.HandleServiceError(c, "Failed to delete variant", err)
		return
	}

	h.logger.Info("Product variant deleted successfully via handler",
		zap.String("variant_id", id))

	utils.OKResponse(c, "Variant deleted successfully", nil)
}

// RegisterRoutes registers all product variant routes
func (h *ProductVariantHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Variant routes
	variants := router.Group("/variants")
	variants.Use(h.aaaMiddleware.Authenticate())
	{
		variants.GET("/:id", h.aaaMiddleware.RequireOrgPermission("variant", "read"), h.GetProductVariant)
		variants.PUT("/:id", h.aaaMiddleware.RequireOrgPermission("variant", "update"), h.UpdateProductVariant)
		variants.DELETE("/:id", h.aaaMiddleware.RequireOrgPermission("variant", "delete"), h.DeleteProductVariant)
		variants.GET("/sku/:sku", h.aaaMiddleware.RequireOrgPermission("variant", "read"), h.GetVariantBySKU)
		variants.GET("/barcode/:barcode", h.aaaMiddleware.RequireOrgPermission("variant", "read"), h.GetVariantByBarcode)
	}

	// Nested routes under products
	products := router.Group("/products")
	products.Use(h.aaaMiddleware.Authenticate())
	{
		products.POST("/:id/variants", h.aaaMiddleware.RequireOrgPermission("variant", "create"), h.CreateProductVariant)
		products.GET("/:id/variants", h.aaaMiddleware.RequireOrgPermission("variant", "read"), h.GetVariantsByProduct)
	}
}
