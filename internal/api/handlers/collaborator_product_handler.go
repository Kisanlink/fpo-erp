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

// CollaboratorProductHandler handles collaborator-product association HTTP requests
type CollaboratorProductHandler struct {
	collabProductService interfaces.CollaboratorProductServiceInterface
	aaaMiddleware        *aaa.AAAMiddleware
	logger               logger.Logger
}

// NewCollaboratorProductHandler creates a new collaborator product handler
func NewCollaboratorProductHandler(collabProductService interfaces.CollaboratorProductServiceInterface, aaaMiddleware *aaa.AAAMiddleware, logger logger.Logger) *CollaboratorProductHandler {
	return &CollaboratorProductHandler{
		collabProductService: collabProductService,
		aaaMiddleware:        aaaMiddleware,
		logger:               logger,
	}
}

// AddProductToCollaborator handles POST /api/v1/collaborators/:id/products
// @Summary Add Product to Collaborator
// @Description Associate a product with a collaborator (requires authentication)
// @Tags Collaborator Products
// @Accept json
// @Produce json
// @Param id path string true "Collaborator ID (format: CLAB_xxxxxxxx)" example(CLAB_12345678)
// @Param request body models.CreateCollaboratorProductRequest true "Product association data"
// @Success 201 {object} utils.Response{data=models.CollaboratorProductResponse} "Product added successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators/{id}/products [post]
func (h *CollaboratorProductHandler) AddProductToCollaborator(c *gin.Context) {
	h.logger.Info("Handling add product to collaborator request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get collaborator ID from URL
	collaboratorID := c.Param("id")
	if collaboratorID == "" {
		h.logger.Error("Collaborator ID is required but not provided")
		utils.BadRequestResponse(c, "Collaborator ID is required", nil)
		return
	}

	var request models.CreateCollaboratorProductRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		h.logger.Error("Invalid request body for add product to collaborator",
			zap.Error(err),
			zap.String("collaborator_id", collaboratorID))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	h.logger.Debug("Calling collaborator product service to add product",
		zap.String("collaborator_id", collaboratorID),
		zap.String("product_id", request.ProductID))

	// Add product to collaborator
	response, err := h.collabProductService.AddProductToCollaborator(c.Request.Context(), collaboratorID, &request)
	if err != nil {
		h.logger.Error("Failed to add product to collaborator via service",
			zap.Error(err),
			zap.String("collaborator_id", collaboratorID),
			zap.String("product_id", request.ProductID))
		utils.HandleServiceError(c, "Failed to add product to collaborator", err)
		return
	}

	h.logger.Info("Product added to collaborator successfully via handler",
		zap.String("collaborator_product_id", response.ID),
		zap.String("collaborator_id", collaboratorID),
		zap.String("product_id", request.ProductID))

	utils.CreatedResponse(c, "Product added to collaborator successfully", response)
}

// GetProductsByCollaborator handles GET /api/v1/collaborators/:id/products
// @Summary Get Products by Collaborator
// @Description Retrieve all products for a specific collaborator
// @Tags Collaborator Products
// @Produce json
// @Param id path string true "Collaborator ID (format: CLAB_xxxxxxxx)" example(CLAB_12345678)
// @Success 200 {object} utils.Response{data=[]models.CollaboratorProductResponse} "Products retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Collaborator not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/collaborators/{id}/products [get]
func (h *CollaboratorProductHandler) GetProductsByCollaborator(c *gin.Context) {
	h.logger.Info("Handling get products by collaborator request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get collaborator ID from URL
	collaboratorID := c.Param("id")
	if collaboratorID == "" {
		h.logger.Error("Collaborator ID is required but not provided")
		utils.BadRequestResponse(c, "Collaborator ID is required", nil)
		return
	}

	h.logger.Debug("Calling collaborator product service to get products by collaborator",
		zap.String("collaborator_id", collaboratorID))

	// Get products
	response, err := h.collabProductService.GetProductsByCollaborator(c.Request.Context(), collaboratorID)
	if err != nil {
		h.logger.Error("Failed to retrieve products by collaborator via service",
			zap.Error(err),
			zap.String("collaborator_id", collaboratorID))
		utils.HandleServiceError(c, "Failed to retrieve products", err)
		return
	}

	h.logger.Info("Products by collaborator retrieved successfully via handler",
		zap.String("collaborator_id", collaboratorID),
		zap.Int("count", len(response)))

	utils.OKResponse(c, "Products retrieved successfully", response)
}

// GetCollaboratorsByProduct handles GET /api/v1/products/:id/collaborators
// @Summary Get Collaborators by Product
// @Description Retrieve all collaborators for a specific product
// @Tags Collaborator Products
// @Produce json
// @Param id path string true "Product ID (format: PROD_xxxxxxxx)" example(PROD_12345678)
// @Success 200 {object} utils.Response{data=[]models.CollaboratorProductResponse} "Collaborators retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Product not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/products/{id}/collaborators [get]
func (h *CollaboratorProductHandler) GetCollaboratorsByProduct(c *gin.Context) {
	h.logger.Info("Handling get collaborators by product request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get product ID from URL
	productID := c.Param("id")
	if productID == "" {
		h.logger.Error("Product ID is required but not provided")
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	h.logger.Debug("Calling collaborator product service to get collaborators by product",
		zap.String("product_id", productID))

	// Get collaborators
	response, err := h.collabProductService.GetCollaboratorsByProduct(c.Request.Context(), productID)
	if err != nil {
		h.logger.Error("Failed to retrieve collaborators by product via service",
			zap.Error(err),
			zap.String("product_id", productID))
		utils.HandleServiceError(c, "Failed to retrieve collaborators", err)
		return
	}

	h.logger.Info("Collaborators by product retrieved successfully via handler",
		zap.String("product_id", productID),
		zap.Int("count", len(response)))

	utils.OKResponse(c, "Collaborators retrieved successfully", response)
}

// GetCollaboratorProduct handles GET /api/v1/collaborator-products/:id
// @Summary Get Collaborator Product
// @Description Retrieve a specific collaborator-product association
// @Tags Collaborator Products
// @Produce json
// @Param id path string true "Collaborator Product ID (format: CPRD_xxxxxxxx)" example(CPRD_12345678)
// @Success 200 {object} utils.Response{data=models.CollaboratorProductResponse} "Association retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Association not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/collaborator-products/{id} [get]
func (h *CollaboratorProductHandler) GetCollaboratorProduct(c *gin.Context) {
	h.logger.Info("Handling get collaborator product request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Collaborator Product ID is required but not provided")
		utils.BadRequestResponse(c, "Collaborator Product ID is required", nil)
		return
	}

	h.logger.Debug("Calling collaborator product service to get association",
		zap.String("collaborator_product_id", id))

	// Get association
	response, err := h.collabProductService.GetCollaboratorProduct(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Association not found",
			zap.Error(err),
			zap.String("collaborator_product_id", id))
		utils.NotFoundResponse(c, "Association not found")
		return
	}

	h.logger.Info("Association retrieved successfully via handler",
		zap.String("collaborator_product_id", response.ID))

	utils.OKResponse(c, "Association retrieved successfully", response)
}

// UpdateCollaboratorProduct handles PUT /api/v1/collaborator-products/:id
// @Summary Update Collaborator Product
// @Description Update collaborator-product metadata (requires authentication)
// @Tags Collaborator Products
// @Accept json
// @Produce json
// @Param id path string true "Collaborator Product ID (format: CPRD_xxxxxxxx)" example(CPRD_12345678)
// @Param request body models.UpdateCollaboratorProductRequest true "Updated metadata"
// @Success 200 {object} utils.Response{data=models.CollaboratorProductResponse} "Association updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Association not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborator-products/{id} [put]
func (h *CollaboratorProductHandler) UpdateCollaboratorProduct(c *gin.Context) {
	h.logger.Info("Handling update collaborator product request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Collaborator Product ID is required but not provided")
		utils.BadRequestResponse(c, "Collaborator Product ID is required", nil)
		return
	}

	var request models.UpdateCollaboratorProductRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		h.logger.Error("Invalid request body for update collaborator product",
			zap.Error(err),
			zap.String("collaborator_product_id", id))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	h.logger.Debug("Calling collaborator product service to update association",
		zap.String("collaborator_product_id", id))

	// Update association
	response, err := h.collabProductService.UpdateCollaboratorProduct(c.Request.Context(), id, &request)
	if err != nil {
		h.logger.Error("Failed to update association via service",
			zap.Error(err),
			zap.String("collaborator_product_id", id))
		utils.HandleServiceError(c, "Failed to update association", err)
		return
	}

	h.logger.Info("Association updated successfully via handler",
		zap.String("collaborator_product_id", response.ID))

	utils.OKResponse(c, "Association updated successfully", response)
}

// RemoveProductFromCollaborator handles DELETE /api/v1/collaborators/:id/products/:product_id
// @Summary Remove Product from Collaborator
// @Description Remove product association from collaborator (soft delete, requires authentication)
// @Tags Collaborator Products
// @Produce json
// @Param id path string true "Collaborator ID (format: CLAB_xxxxxxxx)" example(CLAB_12345678)
// @Param product_id path string true "Product ID (format: PROD_xxxxxxxx)" example(PROD_12345678)
// @Success 200 {object} utils.Response "Product removed successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Association not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators/{id}/products/{product_id} [delete]
func (h *CollaboratorProductHandler) RemoveProductFromCollaborator(c *gin.Context) {
	h.logger.Info("Handling remove product from collaborator request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get IDs from URL
	collaboratorID := c.Param("id")
	productID := c.Param("product_id")

	if collaboratorID == "" || productID == "" {
		h.logger.Error("Collaborator ID and Product ID are required but not provided")
		utils.BadRequestResponse(c, "Collaborator ID and Product ID are required", nil)
		return
	}

	h.logger.Debug("Calling collaborator product service to remove product",
		zap.String("collaborator_id", collaboratorID),
		zap.String("product_id", productID))

	// Remove association
	if err := h.collabProductService.RemoveProductFromCollaborator(c.Request.Context(), collaboratorID, productID); err != nil {
		h.logger.Error("Failed to remove product from collaborator via service",
			zap.Error(err),
			zap.String("collaborator_id", collaboratorID),
			zap.String("product_id", productID))
		utils.HandleServiceError(c, "Failed to remove product from collaborator", err)
		return
	}

	h.logger.Info("Product removed from collaborator successfully via handler",
		zap.String("collaborator_id", collaboratorID),
		zap.String("product_id", productID))

	utils.OKResponse(c, "Product removed from collaborator successfully", nil)
}

// DeleteCollaboratorProduct handles DELETE /api/v1/collaborator-products/:id
// @Summary Delete Collaborator Product
// @Description Delete collaborator-product association by ID (soft delete, requires authentication)
// @Tags Collaborator Products
// @Produce json
// @Param id path string true "Collaborator Product ID (format: CPRD_xxxxxxxx)" example(CPRD_12345678)
// @Success 200 {object} utils.Response "Association deleted successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Association not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborator-products/{id} [delete]
func (h *CollaboratorProductHandler) DeleteCollaboratorProduct(c *gin.Context) {
	h.logger.Info("Handling delete collaborator product request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Collaborator Product ID is required but not provided")
		utils.BadRequestResponse(c, "Collaborator Product ID is required", nil)
		return
	}

	h.logger.Debug("Calling collaborator product service to delete association",
		zap.String("collaborator_product_id", id))

	// Delete association
	if err := h.collabProductService.DeleteCollaboratorProduct(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete association via service",
			zap.Error(err),
			zap.String("collaborator_product_id", id))
		utils.HandleServiceError(c, "Failed to delete association", err)
		return
	}

	h.logger.Info("Association deleted successfully via handler",
		zap.String("collaborator_product_id", id))

	utils.OKResponse(c, "Association deleted successfully", nil)
}

// RegisterRoutes registers all collaborator-product routes
func (h *CollaboratorProductHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Collaborator product association routes (organization-scoped)
	router.GET("/collaborator-products/:id", h.aaaMiddleware.Authenticate(), h.aaaMiddleware.RequireOrgPermission("collaborator_product", "read"), h.GetCollaboratorProduct)
	router.PUT("/collaborator-products/:id", h.aaaMiddleware.Authenticate(), h.aaaMiddleware.RequireOrgPermission("collaborator_product", "update"), h.UpdateCollaboratorProduct)
	router.DELETE("/collaborator-products/:id", h.aaaMiddleware.Authenticate(), h.aaaMiddleware.RequireOrgPermission("collaborator_product", "delete"), h.DeleteCollaboratorProduct)

	// Nested routes under collaborators (organization-scoped)
	collaborators := router.Group("/collaborators")
	collaborators.Use(h.aaaMiddleware.Authenticate())
	{
		collaborators.POST("/:id/products", h.aaaMiddleware.RequireOrgPermission("collaborator_product", "create"), h.AddProductToCollaborator)
		collaborators.GET("/:id/products", h.aaaMiddleware.RequireOrgPermission("collaborator_product", "read"), h.GetProductsByCollaborator)
		collaborators.DELETE("/:id/products/:product_id", h.aaaMiddleware.RequireOrgPermission("collaborator_product", "delete"), h.RemoveProductFromCollaborator)
	}

	// Nested routes under products (organization-scoped)
	products := router.Group("/products")
	products.Use(h.aaaMiddleware.Authenticate())
	{
		products.GET("/:id/collaborators", h.aaaMiddleware.RequireOrgPermission("collaborator_product", "read"), h.GetCollaboratorsByProduct)
	}
}
