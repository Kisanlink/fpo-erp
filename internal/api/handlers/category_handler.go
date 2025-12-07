package handlers

import (
	"kisanlink-erp/internal/database/models"
	logger "kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CategoryHandler handles category HTTP requests
type CategoryHandler struct {
	categoryService interfaces.CategoryServiceInterface
	logger          logger.Logger
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(categoryService interfaces.CategoryServiceInterface, logger logger.Logger) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
		logger:          logger,
	}
}

// SeedCategories handles POST /api/v1/categories/seed
// @Summary Seed Categories
// @Description Seed all predefined categories and subcategories (idempotent, admin-only)
// @Tags Categories
// @Produce json
// @Success 200 {object} utils.Response{data=models.SeedCategoriesResponse} "Categories seeded successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/categories/seed [post]
func (h *CategoryHandler) SeedCategories(c *gin.Context) {
	h.logger.Info("Handling seed categories request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Seed categories
	response, err := h.categoryService.SeedCategories(c.Request.Context())
	if err != nil {
		h.logger.Error("Service error seeding categories",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to seed categories", err)
		return
	}

	h.logger.Info("Categories seeded successfully",
		zap.Int("categories_created", response.CategoriesCreated),
		zap.Int("subcategories_created", response.SubcategoriesCreated))

	utils.OKResponse(c, "Categories seeded successfully", response)
}

// CreateCategory handles POST /api/v1/categories
// @Summary Create Category
// @Description Create a new category
// @Tags Categories
// @Accept json
// @Produce json
// @Param request body models.CreateCategoryRequest true "Category data"
// @Success 201 {object} utils.Response{data=models.CategoryResponse} "Category created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - category already exists"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	h.logger.Info("Handling create category request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	var request models.CreateCategoryRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		h.logger.Error("Invalid request body for create category",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	h.logger.Debug("Calling service to create category",
		zap.String("name", request.Name))

	// Create category
	response, err := h.categoryService.CreateCategory(c.Request.Context(), &request)
	if err != nil {
		h.logger.Error("Service error creating category",
			zap.Error(err),
			zap.String("name", request.Name))
		utils.HandleServiceError(c, "Failed to create category", err)
		return
	}

	h.logger.Info("Category created successfully",
		zap.String("category_id", response.ID),
		zap.String("name", response.Name))

	utils.CreatedResponse(c, "Category created successfully", response)
}

// GetCategory handles GET /api/v1/categories/:id
// @Summary Get Category
// @Description Retrieve a specific category by ID
// @Tags Categories
// @Produce json
// @Param id path string true "Category ID (format: CATG_xxxxxxxx)" example(CATG_12345678)
// @Success 200 {object} utils.Response{data=models.CategoryResponse} "Category details"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Category not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/categories/{id} [get]
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	h.logger.Info("Handling get category request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Category ID is required but not provided")
		utils.BadRequestResponse(c, "Category ID is required", nil)
		return
	}

	h.logger.Debug("Fetching category by ID",
		zap.String("category_id", id))

	// Get category
	response, err := h.categoryService.GetCategory(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Category not found",
			zap.Error(err),
			zap.String("category_id", id))
		utils.HandleServiceError(c, "Failed to retrieve category", err)
		return
	}

	h.logger.Info("Category retrieved successfully",
		zap.String("category_id", response.ID),
		zap.String("name", response.Name))

	utils.OKResponse(c, "Category retrieved successfully", response)
}

// GetCategoryByName handles GET /api/v1/categories/name/:name
// @Summary Get Category by Name
// @Description Retrieve a specific category by name
// @Tags Categories
// @Produce json
// @Param name path string true "Category name" example(Fertilizers)
// @Success 200 {object} utils.Response{data=models.CategoryResponse} "Category details"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Category not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/categories/name/{name} [get]
func (h *CategoryHandler) GetCategoryByName(c *gin.Context) {
	h.logger.Info("Handling get category by name request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get name from URL
	name := c.Param("name")
	if name == "" {
		h.logger.Error("Category name is required but not provided")
		utils.BadRequestResponse(c, "Category name is required", nil)
		return
	}

	h.logger.Debug("Fetching category by name",
		zap.String("name", name))

	// Get category by name
	response, err := h.categoryService.GetCategoryByName(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Category not found",
			zap.Error(err),
			zap.String("name", name))
		utils.HandleServiceError(c, "Failed to retrieve category", err)
		return
	}

	h.logger.Info("Category retrieved successfully",
		zap.String("category_id", response.ID),
		zap.String("name", response.Name))

	utils.OKResponse(c, "Category retrieved successfully", response)
}

// GetAllCategories handles GET /api/v1/categories
// @Summary Get All Categories
// @Description Retrieve all categories
// @Tags Categories
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.CategoryResponse} "Categories retrieved successfully"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/categories [get]
func (h *CategoryHandler) GetAllCategories(c *gin.Context) {
	h.logger.Info("Handling get all categories request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get all categories
	response, err := h.categoryService.GetAllCategories(c.Request.Context())
	if err != nil {
		h.logger.Error("Service error retrieving all categories",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve categories", err)
		return
	}

	h.logger.Info("All categories retrieved successfully",
		zap.Int("category_count", len(response)))

	utils.OKResponse(c, "Categories retrieved successfully", response)
}

// GetAllCategoriesWithSubcategories handles GET /api/v1/categories/with-subcategories
// @Summary Get All Categories with Subcategories
// @Description Retrieve all categories with their subcategories
// @Tags Categories
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.CategoryResponse} "Categories with subcategories retrieved successfully"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/categories/with-subcategories [get]
func (h *CategoryHandler) GetAllCategoriesWithSubcategories(c *gin.Context) {
	h.logger.Info("Handling get all categories with subcategories request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get all categories with subcategories
	response, err := h.categoryService.GetAllCategoriesWithSubcategories(c.Request.Context())
	if err != nil {
		h.logger.Error("Service error retrieving categories with subcategories",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve categories", err)
		return
	}

	h.logger.Info("All categories with subcategories retrieved successfully",
		zap.Int("category_count", len(response)))

	utils.OKResponse(c, "Categories retrieved successfully", response)
}

// UpdateCategory handles PATCH /api/v1/categories/:id
// @Summary Update Category
// @Description Update an existing category by ID
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID (format: CATG_xxxxxxxx)" example(CATG_12345678)
// @Param request body models.UpdateCategoryRequest true "Updated category data"
// @Success 200 {object} utils.Response{data=models.CategoryResponse} "Category updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Category not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - name already exists"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/categories/{id} [patch]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	h.logger.Info("Handling update category request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Category ID is required but not provided")
		utils.BadRequestResponse(c, "Category ID is required", nil)
		return
	}

	var request models.UpdateCategoryRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		h.logger.Error("Invalid request body for update category",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	h.logger.Debug("Calling service to update category",
		zap.String("category_id", id))

	// Update category
	response, err := h.categoryService.UpdateCategory(c.Request.Context(), id, &request)
	if err != nil {
		h.logger.Error("Service error updating category",
			zap.Error(err),
			zap.String("category_id", id))
		utils.HandleServiceError(c, "Failed to update category", err)
		return
	}

	h.logger.Info("Category updated successfully",
		zap.String("category_id", response.ID),
		zap.String("name", response.Name))

	utils.OKResponse(c, "Category updated successfully", response)
}

// DeleteCategory handles DELETE /api/v1/categories/:id
// @Summary Delete Category
// @Description Delete a category by ID
// @Tags Categories
// @Produce json
// @Param id path string true "Category ID (format: CATG_xxxxxxxx)" example(CATG_12345678)
// @Success 200 {object} utils.Response "Category deleted successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Category not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	h.logger.Info("Handling delete category request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Category ID is required but not provided")
		utils.BadRequestResponse(c, "Category ID is required", nil)
		return
	}

	h.logger.Debug("Calling service to delete category",
		zap.String("category_id", id))

	// Delete category
	if err := h.categoryService.DeleteCategory(c.Request.Context(), id); err != nil {
		h.logger.Error("Service error deleting category",
			zap.Error(err),
			zap.String("category_id", id))
		utils.HandleServiceError(c, "Failed to delete category", err)
		return
	}

	h.logger.Info("Category deleted successfully",
		zap.String("category_id", id))

	utils.OKResponse(c, "Category deleted successfully", nil)
}

// RegisterRoutes registers all category routes
func (h *CategoryHandler) RegisterRoutes(v1 *gin.RouterGroup) {
	categories := v1.Group("/categories")
	{
		// Public endpoints (read operations)
		categories.GET("", h.GetAllCategories)
		categories.GET("/with-subcategories", h.GetAllCategoriesWithSubcategories)
		categories.GET("/name/:name", h.GetCategoryByName)
		categories.GET("/:id", h.GetCategory)

		// Admin-only endpoint (requires authentication)
		categories.POST("/seed", h.SeedCategories)

		// Authenticated endpoints (write operations)
		categories.POST("", h.CreateCategory)
		categories.PATCH("/:id", h.UpdateCategory)
		categories.DELETE("/:id", h.DeleteCategory)
	}
}
