package handlers

import (
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

// ProductPriceHandler handles product price HTTP requests
type ProductPriceHandler struct {
	priceService  *services.ProductPriceService
	aaaMiddleware *aaa.AAAMiddleware
}

// NewProductPriceHandler creates a new product price handler
func NewProductPriceHandler(priceService *services.ProductPriceService, aaaMiddleware *aaa.AAAMiddleware) *ProductPriceHandler {
	return &ProductPriceHandler{
		priceService:  priceService,
		aaaMiddleware: aaaMiddleware,
	}
}

// CreateProductPrice handles POST /api/v1/prices
// @Summary Create Product Price
// @Description Create a new product price configuration (requires authentication)
// @Tags Product Prices
// @Accept json
// @Produce json
// @Param request body models.CreateProductPriceRequest true "Product price data"
// @Success 201 {object} utils.Response{data=models.ProductPriceResponse} "Product price created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/prices [post]
func (h *ProductPriceHandler) CreateProductPrice(c *gin.Context) {
	var request models.CreateProductPriceRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Create product price
	response, err := h.priceService.CreateProductPrice(&request)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create product price", err)
		return
	}

	utils.CreatedResponse(c, "Product price created successfully", response)
}

// GetProductPrice handles GET /api/v1/prices/:id
// @Summary Get Product Price
// @Description Retrieve a specific product price by ID
// @Tags Product Prices
// @Produce json
// @Param id path string true "Price ID" example(PRICE_12345678)
// @Success 200 {object} utils.Response{data=models.ProductPriceResponse} "Product price retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Product price not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/prices/{id} [get]
func (h *ProductPriceHandler) GetProductPrice(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Price ID is required", nil)
		return
	}

	// Get product price
	response, err := h.priceService.GetProductPrice(id)
	if err != nil {
		utils.NotFoundResponse(c, "Product price not found")
		return
	}

	utils.OKResponse(c, "Product price retrieved successfully", response)
}

// GetProductPrices handles GET /api/v1/products/:id/prices
// @Summary Get Product Prices
// @Description Retrieve all prices for a specific product
// @Tags Product Prices
// @Produce json
// @Param id path string true "Product ID" example(PROD_12345678)
// @Success 200 {object} utils.Response{data=[]models.ProductPriceResponse} "Product prices retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/{id}/prices [get]
func (h *ProductPriceHandler) GetProductPrices(c *gin.Context) {
	// Get product ID from URL
	productID := c.Param("id")
	if productID == "" {
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	// Get product prices
	response, err := h.priceService.GetProductPrices(productID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve product prices", err)
		return
	}

	utils.OKResponse(c, "Product prices retrieved successfully", response)
}

// GetCurrentPrice handles GET /api/v1/products/:id/prices/current
// @Summary Get Current Product Price
// @Description Retrieve the current active price for a specific product
// @Tags Product Prices
// @Produce json
// @Param id path string true "Product ID" example(PROD_12345678)
// @Param type query string false "Price type (default: retail)" example(retail)
// @Success 200 {object} utils.Response{data=models.ProductPriceResponse} "Current price retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Current price not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/{id}/prices/current [get]
func (h *ProductPriceHandler) GetCurrentPrice(c *gin.Context) {
	// Get product ID from URL
	productID := c.Param("id")
	if productID == "" {
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	// Get price type from query parameter
	priceType := c.Query("type")
	if priceType == "" {
		priceType = "retail" // Default to retail price
	}

	// Get current price
	response, err := h.priceService.GetCurrentPrice(productID, priceType)
	if err != nil {
		utils.NotFoundResponse(c, "Current price not found")
		return
	}

	utils.OKResponse(c, "Current price retrieved successfully", response)
}

// UpdateProductPrice handles PATCH /api/v1/prices/:id
// @Summary Update Product Price
// @Description Update an existing product price by ID
// @Tags Product Prices
// @Accept json
// @Produce json
// @Param id path string true "Price ID" example(PRICE_12345678)
// @Param request body models.UpdateProductPriceRequest true "Updated price data"
// @Success 200 {object} utils.Response{data=models.ProductPriceResponse} "Product price updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Product price not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/prices/{id} [patch]
func (h *ProductPriceHandler) UpdateProductPrice(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Price ID is required", nil)
		return
	}

	var request models.UpdateProductPriceRequest

	// Validate request
	if err := utils.ValidatePartialRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Update product price
	response, err := h.priceService.UpdateProductPrice(id, &request)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update product price", err)
		return
	}

	utils.OKResponse(c, "Product price updated successfully", response)
}

// DeleteProductPrice handles DELETE /api/v1/prices/:id
// @Summary Delete Product Price
// @Description Delete a product price by ID
// @Tags Product Prices
// @Produce json
// @Param id path string true "Price ID" example(PRICE_12345678)
// @Success 200 {object} utils.Response "Product price deleted successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Product price not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/prices/{id} [delete]
func (h *ProductPriceHandler) DeleteProductPrice(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Price ID is required", nil)
		return
	}

	// Delete product price
	if err := h.priceService.DeleteProductPrice(id); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete product price", err)
		return
	}

	utils.OKResponse(c, "Product price deleted successfully", nil)
}

// GetExpiredPrices handles GET /api/v1/prices/expired
// @Summary Get Expired Prices
// @Description Retrieve all expired product prices
// @Tags Product Prices
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.ProductPriceResponse} "Expired prices retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/prices/expired [get]
func (h *ProductPriceHandler) GetExpiredPrices(c *gin.Context) {
	// Get expired prices
	response, err := h.priceService.GetExpiredPrices()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve expired prices", err)
		return
	}

	utils.OKResponse(c, "Expired prices retrieved successfully", response)
}

// CreateProductPriceForProduct handles POST /api/v1/products/:id/prices
// @Summary Create Product Price for Product
// @Description Create a new price for a specific product
// @Tags Product Prices
// @Accept json
// @Produce json
// @Param id path string true "Product ID" example(PROD_12345678)
// @Param request body models.CreateProductPriceRequest true "Product price data"
// @Success 201 {object} utils.Response{data=models.ProductPriceResponse} "Product price created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/{id}/prices [post]
func (h *ProductPriceHandler) CreateProductPriceForProduct(c *gin.Context) {
	// Get product ID from URL
	productID := c.Param("id")
	if productID == "" {
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	var request models.CreateProductPriceRequest

	// Set the product ID from URL parameter first
	request.ProductID = productID

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Create product price
	response, err := h.priceService.CreateProductPrice(&request)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create product price", err)
		return
	}

	utils.CreatedResponse(c, "Product price created successfully", response)
}

// RegisterRoutes registers product price routes
func (h *ProductPriceHandler) RegisterRoutes(router *gin.RouterGroup) {
	prices := router.Group("/prices")
	{
		// Apply authentication middleware
		prices.Use(h.aaaMiddleware.Authenticate())

		// Create/Update/Delete routes - CEO=CRUD, Store_Manager=CRUD, Tech_Support=R/W (temp)
		prices.POST("", h.aaaMiddleware.RequirePermission("aaa/product_price", "*", "create"), h.CreateProductPrice)
		prices.PATCH("/:id", h.aaaMiddleware.RequirePermission("aaa/product_price", "*", "update"), h.UpdateProductPrice)
		prices.DELETE("/:id", h.aaaMiddleware.RequirePermission("aaa/product_price", "*", "delete"), h.DeleteProductPrice)

		// Read routes - Director=R, CEO=CRUD, Auditor=R, Accountant=–, Tech_Support=R/W (temp), Store_Manager=CRUD, Store_Staff=R
		prices.GET("/:id", h.aaaMiddleware.RequirePermission("aaa/product_price", "*", "read"), h.GetProductPrice)
		prices.GET("/expired", h.aaaMiddleware.RequirePermission("aaa/product_price", "*", "read"), h.GetExpiredPrices)
	}

	// Product-specific price routes
	products := router.Group("/products")
	{
		products.Use(h.aaaMiddleware.Authenticate())
		products.GET("/:id/prices", h.aaaMiddleware.RequirePermission("aaa/product_price", "*", "read"), h.GetProductPrices)
		products.GET("/:id/prices/current", h.aaaMiddleware.RequirePermission("aaa/product_price", "*", "read"), h.GetCurrentPrice)
		products.POST("/:id/prices", h.aaaMiddleware.RequirePermission("aaa/product_price", "*", "create"), h.CreateProductPriceForProduct)
	}
}
