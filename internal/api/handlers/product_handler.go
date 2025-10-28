package handlers

import (
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

// ProductHandler handles product HTTP requests
type ProductHandler struct {
	productService *services.ProductService
	aaaMiddleware  *aaa.AAAMiddleware
}

// NewProductHandler creates a new product handler
func NewProductHandler(productService *services.ProductService, aaaMiddleware *aaa.AAAMiddleware) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		aaaMiddleware:  aaaMiddleware,
	}
}

// CreateProduct handles POST /api/v1/products
// @Summary Create Product
// @Description Create a new product (requires authentication)
// @Tags Products
// @Accept json
// @Produce json
// @Param request body models.CreateProductRequest true "Product data"
// @Success 201 {object} utils.Response{data=models.ProductResponse} "Product created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var request models.CreateProductRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Create product
	response, err := h.productService.CreateProduct(&request)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create product", err)
		return
	}

	utils.CreatedResponse(c, "Product created successfully", response)
}

// GetProduct handles GET /api/v1/products/:id
// @Summary Get Product
// @Description Retrieve a specific product by ID
// @Tags Products
// @Produce json
// @Param id path string true "Product ID" example(PROD_12345678)
// @Success 200 {object} utils.Response{data=models.ProductResponse} "Product details"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Product not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	// Get product
	response, err := h.productService.GetProduct(id)
	if err != nil {
		utils.NotFoundResponse(c, "Product not found")
		return
	}

	utils.OKResponse(c, "Product retrieved successfully", response)
}

// GetProductBySKU handles GET /api/v1/products/sku/:sku
// @Summary Get Product by SKU
// @Description Retrieve a specific product by SKU
// @Tags Products
// @Produce json
// @Param sku path string true "Product SKU" example(SKU123)
// @Success 200 {object} utils.Response{data=models.ProductResponse} "Product retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Product not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/sku/{sku} [get]
func (h *ProductHandler) GetProductBySKU(c *gin.Context) {
	// Get SKU from URL
	sku := c.Param("sku")
	if sku == "" {
		utils.BadRequestResponse(c, "SKU is required", nil)
		return
	}

	// Get product by SKU
	response, err := h.productService.GetProductBySKU(sku)
	if err != nil {
		utils.NotFoundResponse(c, "Product not found")
		return
	}

	utils.OKResponse(c, "Product retrieved successfully", response)
}

// GetAllProducts handles GET /api/v1/products
// @Summary Get All Products
// @Description Retrieve all products (requires authentication)
// @Tags Products
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.ProductResponse} "Products retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products [get]
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	// Get all products
	response, err := h.productService.GetAllProducts()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve products", err)
		return
	}

	utils.OKResponse(c, "Products retrieved successfully", response)
}

// UpdateProduct handles PATCH /api/v1/products/:id
// @Summary Update Product
// @Description Update an existing product by ID (requires authentication)
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID" example(PROD_12345678)
// @Param request body models.UpdateProductRequest true "Updated product data"
// @Success 200 {object} utils.Response{data=models.ProductResponse} "Product updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Product not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/{id} [patch]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	var request models.UpdateProductRequest

	// Validate request
	if err := utils.ValidatePartialRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Update product
	response, err := h.productService.UpdateProduct(id, &request)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update product", err)
		return
	}

	utils.OKResponse(c, "Product updated successfully", response)
}

// DeleteProduct handles DELETE /api/v1/products/:id
// @Summary Delete Product
// @Description Delete a product by ID (requires authentication)
// @Tags Products
// @Produce json
// @Param id path string true "Product ID" example(PROD_12345678)
// @Success 200 {object} utils.Response "Product deleted successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Product not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	// Delete product
	if err := h.productService.DeleteProduct(id); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete product", err)
		return
	}

	utils.OKResponse(c, "Product deleted successfully", nil)
}

// SearchProducts handles GET /api/v1/products/search
// @Summary Search Products
// @Description Search products by query string (requires authentication)
// @Tags Products
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {object} utils.Response{data=[]models.ProductResponse} "Products found"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/search [get]
func (h *ProductHandler) SearchProducts(c *gin.Context) {
	// Get query parameter
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	// Search products
	response, err := h.productService.SearchProducts(query)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to search products", err)
		return
	}

	utils.OKResponse(c, "Products found", response)
}

// GetProductWithPrices handles GET /api/v1/products/:id/with-prices
// @Summary Get Product with Prices
// @Description Retrieve a product with all associated pricing information
// @Tags Products
// @Produce json
// @Param id path string true "Product ID" example(PROD_12345678)
// @Success 200 {object} utils.Response{data=models.ProductWithPricesResponse} "Product with prices retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Product not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/{id}/with-prices [get]
func (h *ProductHandler) GetProductWithPrices(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	// Get product with prices
	response, err := h.productService.GetProductWithPrices(id)
	if err != nil {
		utils.NotFoundResponse(c, "Product not found")
		return
	}

	utils.OKResponse(c, "Product with prices retrieved successfully", response)
}

// RegisterRoutes registers product routes
func (h *ProductHandler) RegisterRoutes(router *gin.RouterGroup) {
	products := router.Group("/products")
	{
		// Apply authentication middleware
		products.Use(h.aaaMiddleware.Authenticate())

		// Create/Update/Delete routes - CEO=CRUD, Store_Manager=CRUD, Tech_Support=R/W (temp)
		products.POST("", h.aaaMiddleware.RequireOrgPermission("product", "create"), h.CreateProduct)
		products.PATCH("/:id", h.aaaMiddleware.RequireOrgPermission("product", "update"), h.UpdateProduct)
		products.DELETE("/:id", h.aaaMiddleware.RequireOrgPermission("product", "delete"), h.DeleteProduct)

		// Read routes - Director=R, CEO=CRUD, Auditor=R, Accountant=–, Tech_Support=R/W (temp), Store_Manager=CRUD, Store_Staff=R
		products.GET("", h.aaaMiddleware.RequireOrgPermission("product", "read"), h.GetAllProducts)
		products.GET("/search", h.aaaMiddleware.RequireOrgPermission("product", "read"), h.SearchProducts)
		products.GET("/sku/:sku", h.aaaMiddleware.RequireOrgPermission("product", "read"), h.GetProductBySKU)
		products.GET("/:id", h.aaaMiddleware.RequireOrgPermission("product", "read"), h.GetProduct)
		products.GET("/:id/with-prices", h.aaaMiddleware.RequireOrgPermission("product", "read"), h.GetProductWithPrices)
	}
}
