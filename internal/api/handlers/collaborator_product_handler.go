package handlers

import (
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

// CollaboratorProductHandler handles collaborator-product association HTTP requests
type CollaboratorProductHandler struct {
	collabProductService *services.CollaboratorProductService
	aaaMiddleware        *aaa.AAAMiddleware
}

// NewCollaboratorProductHandler creates a new collaborator product handler
func NewCollaboratorProductHandler(collabProductService *services.CollaboratorProductService, aaaMiddleware *aaa.AAAMiddleware) *CollaboratorProductHandler {
	return &CollaboratorProductHandler{
		collabProductService: collabProductService,
		aaaMiddleware:        aaaMiddleware,
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
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators/{id}/products [post]
func (h *CollaboratorProductHandler) AddProductToCollaborator(c *gin.Context) {
	// Get collaborator ID from URL
	collaboratorID := c.Param("id")
	if collaboratorID == "" {
		utils.BadRequestResponse(c, "Collaborator ID is required", nil)
		return
	}

	var request models.CreateCollaboratorProductRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Add product to collaborator
	response, err := h.collabProductService.AddProductToCollaborator(c.Request.Context(), collaboratorID, &request)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to add product to collaborator", err)
		return
	}

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
	// Get collaborator ID from URL
	collaboratorID := c.Param("id")
	if collaboratorID == "" {
		utils.BadRequestResponse(c, "Collaborator ID is required", nil)
		return
	}

	// Get products
	response, err := h.collabProductService.GetProductsByCollaborator(c.Request.Context(), collaboratorID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve products", err)
		return
	}

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
	// Get product ID from URL
	productID := c.Param("id")
	if productID == "" {
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	// Get collaborators
	response, err := h.collabProductService.GetCollaboratorsByProduct(c.Request.Context(), productID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve collaborators", err)
		return
	}

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
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Collaborator Product ID is required", nil)
		return
	}

	// Get association
	response, err := h.collabProductService.GetCollaboratorProduct(c.Request.Context(), id)
	if err != nil {
		utils.NotFoundResponse(c, "Association not found")
		return
	}

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
// @Failure 404 {object} utils.ErrorResponseModel "Association not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborator-products/{id} [put]
func (h *CollaboratorProductHandler) UpdateCollaboratorProduct(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Collaborator Product ID is required", nil)
		return
	}

	var request models.UpdateCollaboratorProductRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Update association
	response, err := h.collabProductService.UpdateCollaboratorProduct(c.Request.Context(), id, &request)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update association", err)
		return
	}

	utils.OKResponse(c, "Association updated successfully", response)
}

// RemoveProductFromCollaborator handles DELETE /api/v1/collaborators/:collaborator_id/products/:product_id
// @Summary Remove Product from Collaborator
// @Description Remove product association from collaborator (soft delete, requires authentication)
// @Tags Collaborator Products
// @Produce json
// @Param collaborator_id path string true "Collaborator ID (format: CLAB_xxxxxxxx)" example(CLAB_12345678)
// @Param product_id path string true "Product ID (format: PROD_xxxxxxxx)" example(PROD_12345678)
// @Success 200 {object} utils.Response "Product removed successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Association not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators/{collaborator_id}/products/{product_id} [delete]
func (h *CollaboratorProductHandler) RemoveProductFromCollaborator(c *gin.Context) {
	// Get IDs from URL
	collaboratorID := c.Param("collaborator_id")
	productID := c.Param("product_id")

	if collaboratorID == "" || productID == "" {
		utils.BadRequestResponse(c, "Collaborator ID and Product ID are required", nil)
		return
	}

	// Remove association
	if err := h.collabProductService.RemoveProductFromCollaborator(c.Request.Context(), collaboratorID, productID); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to remove product from collaborator", err)
		return
	}

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
// @Failure 404 {object} utils.ErrorResponseModel "Association not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborator-products/{id} [delete]
func (h *CollaboratorProductHandler) DeleteCollaboratorProduct(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Collaborator Product ID is required", nil)
		return
	}

	// Delete association
	if err := h.collabProductService.DeleteCollaboratorProduct(c.Request.Context(), id); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete association", err)
		return
	}

	utils.OKResponse(c, "Association deleted successfully", nil)
}

// RegisterRoutes registers all collaborator-product routes
func (h *CollaboratorProductHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Collaborator product association routes
	router.GET("/collaborator-products/:id", h.aaaMiddleware.Authenticate(), h.aaaMiddleware.RequirePermission("collaborator_product", "*", "read"), h.GetCollaboratorProduct)
	router.PUT("/collaborator-products/:id", h.aaaMiddleware.Authenticate(), h.aaaMiddleware.RequirePermission("collaborator_product", "*", "update"), h.UpdateCollaboratorProduct)
	router.DELETE("/collaborator-products/:id", h.aaaMiddleware.Authenticate(), h.aaaMiddleware.RequirePermission("collaborator_product", "*", "delete"), h.DeleteCollaboratorProduct)

	// Nested routes under collaborators
	collaborators := router.Group("/collaborators")
	collaborators.Use(h.aaaMiddleware.Authenticate())
	{
		collaborators.POST("/:id/products", h.aaaMiddleware.RequirePermission("collaborator_product", "*", "create"), h.AddProductToCollaborator)
		collaborators.GET("/:id/products", h.aaaMiddleware.RequirePermission("collaborator_product", "*", "read"), h.GetProductsByCollaborator)
		collaborators.DELETE("/:collaborator_id/products/:product_id", h.aaaMiddleware.RequirePermission("collaborator_product", "*", "delete"), h.RemoveProductFromCollaborator)
	}

	// Nested routes under products
	products := router.Group("/products")
	products.Use(h.aaaMiddleware.Authenticate())
	{
		products.GET("/:id/collaborators", h.aaaMiddleware.RequirePermission("collaborator_product", "*", "read"), h.GetCollaboratorsByProduct)
	}
}
