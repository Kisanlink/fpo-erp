package handlers

import (
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

// ProductVariantHandler handles product variant HTTP requests
type ProductVariantHandler struct {
	variantService *services.ProductVariantService
	aaaMiddleware  *aaa.AAAMiddleware
}

// NewProductVariantHandler creates a new product variant handler
func NewProductVariantHandler(variantService *services.ProductVariantService, aaaMiddleware *aaa.AAAMiddleware) *ProductVariantHandler {
	return &ProductVariantHandler{
		variantService: variantService,
		aaaMiddleware:  aaaMiddleware,
	}
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
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/{id}/variants [post]
func (h *ProductVariantHandler) CreateProductVariant(c *gin.Context) {
	// Get product ID from URL
	productID := c.Param("id")
	if productID == "" {
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	var request models.CreateProductVariantRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Create variant
	response, err := h.variantService.CreateProductVariant(c.Request.Context(), productID, &request)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create variant", err)
		return
	}

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
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Variant ID is required", nil)
		return
	}

	// Get variant
	response, err := h.variantService.GetProductVariant(c.Request.Context(), id)
	if err != nil {
		utils.NotFoundResponse(c, "Variant not found")
		return
	}

	utils.OKResponse(c, "Variant retrieved successfully", response)
}

// GetVariantsByProduct handles GET /api/v1/products/:id/variants
// @Summary Get Variants by Product
// @Description Retrieve all variants for a specific product
// @Tags Product Variants
// @Produce json
// @Param id path string true "Product ID (format: PROD_xxxxxxxx)" example(PROD_12345678)
// @Success 200 {object} utils.Response{data=[]models.ProductVariantResponse} "Variants retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Product not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/products/{id}/variants [get]
func (h *ProductVariantHandler) GetVariantsByProduct(c *gin.Context) {
	// Get product ID from URL
	productID := c.Param("id")
	if productID == "" {
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	// Get variants
	response, err := h.variantService.GetVariantsByProduct(c.Request.Context(), productID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve variants", err)
		return
	}

	utils.OKResponse(c, "Variants retrieved successfully", response)
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
	// Get SKU from URL
	sku := c.Param("sku")
	if sku == "" {
		utils.BadRequestResponse(c, "SKU is required", nil)
		return
	}

	// Get variant
	response, err := h.variantService.GetVariantBySKU(c.Request.Context(), sku)
	if err != nil {
		utils.NotFoundResponse(c, "Variant not found")
		return
	}

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
	// Get barcode from URL
	barcode := c.Param("barcode")
	if barcode == "" {
		utils.BadRequestResponse(c, "Barcode is required", nil)
		return
	}

	// Get variant
	response, err := h.variantService.GetVariantByBarcode(c.Request.Context(), barcode)
	if err != nil {
		utils.NotFoundResponse(c, "Variant not found")
		return
	}

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
// @Failure 404 {object} utils.ErrorResponseModel "Variant not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/variants/{id} [put]
func (h *ProductVariantHandler) UpdateProductVariant(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Variant ID is required", nil)
		return
	}

	var request models.UpdateProductVariantRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Update variant
	response, err := h.variantService.UpdateProductVariant(c.Request.Context(), id, &request)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update variant", err)
		return
	}

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
// @Failure 404 {object} utils.ErrorResponseModel "Variant not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/variants/{id} [delete]
func (h *ProductVariantHandler) DeleteProductVariant(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Variant ID is required", nil)
		return
	}

	// Delete variant
	if err := h.variantService.DeleteProductVariant(c.Request.Context(), id); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete variant", err)
		return
	}

	utils.OKResponse(c, "Variant deleted successfully", nil)
}

// RegisterRoutes registers all product variant routes
func (h *ProductVariantHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Variant routes
	variants := router.Group("/variants")
	variants.Use(h.aaaMiddleware.Authenticate())
	{
		variants.GET("/:id", h.aaaMiddleware.RequirePermission("variant", "*", "read"), h.GetProductVariant)
		variants.PUT("/:id", h.aaaMiddleware.RequirePermission("variant", "*", "update"), h.UpdateProductVariant)
		variants.DELETE("/:id", h.aaaMiddleware.RequirePermission("variant", "*", "delete"), h.DeleteProductVariant)
		variants.GET("/sku/:sku", h.aaaMiddleware.RequirePermission("variant", "*", "read"), h.GetVariantBySKU)
		variants.GET("/barcode/:barcode", h.aaaMiddleware.RequirePermission("variant", "*", "read"), h.GetVariantByBarcode)
	}

	// Nested routes under products
	products := router.Group("/products")
	products.Use(h.aaaMiddleware.Authenticate())
	{
		products.POST("/:id/variants", h.aaaMiddleware.RequirePermission("variant", "*", "create"), h.CreateProductVariant)
		products.GET("/:id/variants", h.aaaMiddleware.RequirePermission("variant", "*", "read"), h.GetVariantsByProduct)
	}
}
