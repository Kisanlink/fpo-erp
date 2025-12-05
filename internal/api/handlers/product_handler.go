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

// ProductHandler handles product HTTP requests
type ProductHandler struct {
	productService interfaces.ProductServiceInterface
	aaaMiddleware  *aaa.AAAMiddleware
	logger         logger.Logger
}

// NewProductHandler creates a new product handler
func NewProductHandler(productService interfaces.ProductServiceInterface, aaaMiddleware *aaa.AAAMiddleware, logger logger.Logger) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		aaaMiddleware:  aaaMiddleware,
		logger:         logger,
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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	h.logger.Info("Handling create product request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	var request models.CreateProductRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		h.logger.Error("Invalid request body for create product",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	h.logger.Debug("Calling service to create product",
		zap.String("product_name", request.Name))

	// Create product
	response, err := h.productService.CreateProduct(&request)
	if err != nil {
		h.logger.Error("Service error creating product",
			zap.Error(err),
			zap.String("product_name", request.Name))
		utils.HandleServiceError(c, "Failed to create product", err)
		return
	}

	h.logger.Info("Product created successfully",
		zap.String("product_id", response.ID),
		zap.String("product_name", response.Name))

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
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Product not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	h.logger.Info("Handling get product request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Product ID is required but not provided")
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	h.logger.Debug("Fetching product by ID",
		zap.String("product_id", id))

	// Get product
	response, err := h.productService.GetProduct(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Product not found",
			zap.Error(err),
			zap.String("product_id", id))
		utils.NotFoundResponse(c, "Product not found")
		return
	}

	h.logger.Info("Product retrieved successfully",
		zap.String("product_id", response.ID),
		zap.String("product_name", response.Name))

	utils.OKResponse(c, "Product retrieved successfully", response)
}

// GetAllProducts handles GET /api/v1/products
// @Summary Get All Products
// @Description Retrieve all products (requires authentication)
// @Tags Products
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.ProductResponse} "Products retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products [get]
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	h.logger.Info("Handling get all products request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get all products
	response, err := h.productService.GetAllProducts(c.Request.Context())
	if err != nil {
		h.logger.Error("Service error retrieving all products",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve products", err)
		return
	}

	h.logger.Info("All products retrieved successfully",
		zap.Int("product_count", len(response)))

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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Product not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/{id} [patch]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	h.logger.Info("Handling update product request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Product ID is required but not provided for update")
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	var request models.UpdateProductRequest

	// Validate request
	if err := utils.ValidatePartialRequest(c, &request); err != nil {
		h.logger.Error("Invalid request body for update product",
			zap.Error(err),
			zap.String("product_id", id))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	h.logger.Debug("Calling service to update product",
		zap.String("product_id", id))

	// Update product
	response, err := h.productService.UpdateProduct(id, &request)
	if err != nil {
		h.logger.Error("Service error updating product",
			zap.Error(err),
			zap.String("product_id", id))
		utils.HandleServiceError(c, "Failed to update product", err)
		return
	}

	h.logger.Info("Product updated successfully",
		zap.String("product_id", response.ID),
		zap.String("product_name", response.Name))

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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Product not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	h.logger.Info("Handling delete product request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Product ID is required but not provided for delete")
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	h.logger.Debug("Calling service to delete product",
		zap.String("product_id", id))

	// Delete product
	if err := h.productService.DeleteProduct(id); err != nil {
		h.logger.Error("Service error deleting product",
			zap.Error(err),
			zap.String("product_id", id))
		utils.HandleServiceError(c, "Failed to delete product", err)
		return
	}

	h.logger.Info("Product deleted successfully",
		zap.String("product_id", id))

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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/search [get]
func (h *ProductHandler) SearchProducts(c *gin.Context) {
	h.logger.Info("Handling search products request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get query parameter
	query := c.Query("q")
	if query == "" {
		h.logger.Error("Search query is required but not provided")
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	h.logger.Debug("Searching products",
		zap.String("query", query))

	// Search products
	response, err := h.productService.SearchProducts(query)
	if err != nil {
		h.logger.Error("Service error searching products",
			zap.Error(err),
			zap.String("query", query))
		utils.HandleServiceError(c, "Failed to search products", err)
		return
	}

	h.logger.Info("Products search completed",
		zap.String("query", query),
		zap.Int("results_count", len(response)))

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
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Product not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/products/{id}/with-prices [get]
func (h *ProductHandler) GetProductWithPrices(c *gin.Context) {
	h.logger.Info("Handling get product with prices request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Product ID is required but not provided for get with prices")
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	h.logger.Debug("Fetching product with prices",
		zap.String("product_id", id))

	// Get product with prices
	response, err := h.productService.GetProductWithPrices(id)
	if err != nil {
		h.logger.Error("Product not found for get with prices",
			zap.Error(err),
			zap.String("product_id", id))
		utils.NotFoundResponse(c, "Product not found")
		return
	}

	h.logger.Info("Product with prices retrieved successfully",
		zap.String("product_id", id))

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
		products.GET("/:id", h.aaaMiddleware.RequireOrgPermission("product", "read"), h.GetProduct)
		products.GET("/:id/with-prices", h.aaaMiddleware.RequireOrgPermission("product", "read"), h.GetProductWithPrices)
	}
}
