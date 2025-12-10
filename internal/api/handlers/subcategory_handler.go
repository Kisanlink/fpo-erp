package handlers

import (
	"kisanlink-erp/internal/database/models"
	logger "kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SubcategoryHandler handles subcategory HTTP requests
type SubcategoryHandler struct {
	subcategoryService interfaces.SubcategoryServiceInterface
	logger             logger.Logger
}

// NewSubcategoryHandler creates a new subcategory handler
func NewSubcategoryHandler(subcategoryService interfaces.SubcategoryServiceInterface, logger logger.Logger) *SubcategoryHandler {
	return &SubcategoryHandler{
		subcategoryService: subcategoryService,
		logger:             logger,
	}
}

// CreateSubcategory handles POST /api/v1/subcategories
// @Summary Create Subcategory
// @Description Create a new subcategory
// @Tags Subcategories
// @Accept json
// @Produce json
// @Param request body models.CreateSubcategoryRequest true "Subcategory data"
// @Success 201 {object} utils.Response{data=models.SubcategoryResponse} "Subcategory created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request - category does not exist"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - subcategory already exists"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/subcategories [post]
func (h *SubcategoryHandler) CreateSubcategory(c *gin.Context) {
	h.logger.Info("Handling create subcategory request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	var request models.CreateSubcategoryRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		h.logger.Error("Invalid request body for create subcategory",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	h.logger.Debug("Calling service to create subcategory",
		zap.String("name", request.Name),
		zap.String("category_id", request.CategoryID))

	// Create subcategory
	response, err := h.subcategoryService.CreateSubcategory(c.Request.Context(), &request)
	if err != nil {
		h.logger.Error("Service error creating subcategory",
			zap.Error(err),
			zap.String("name", request.Name))
		utils.HandleServiceError(c, "Failed to create subcategory", err)
		return
	}

	h.logger.Info("Subcategory created successfully",
		zap.String("subcategory_id", response.ID),
		zap.String("name", response.Name))

	utils.CreatedResponse(c, "Subcategory created successfully", response)
}

// GetSubcategory handles GET /api/v1/subcategories/:id
// @Summary Get Subcategory
// @Description Retrieve a specific subcategory by ID
// @Tags Subcategories
// @Produce json
// @Param id path string true "Subcategory ID (format: SCAT_xxxxxxxx)" example(SCAT_12345678)
// @Success 200 {object} utils.Response{data=models.SubcategoryResponse} "Subcategory details"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Subcategory not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/subcategories/{id} [get]
func (h *SubcategoryHandler) GetSubcategory(c *gin.Context) {
	h.logger.Info("Handling get subcategory request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Subcategory ID is required but not provided")
		utils.BadRequestResponse(c, "Subcategory ID is required", nil)
		return
	}

	h.logger.Debug("Fetching subcategory by ID",
		zap.String("subcategory_id", id))

	// Get subcategory
	response, err := h.subcategoryService.GetSubcategory(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Subcategory not found",
			zap.Error(err),
			zap.String("subcategory_id", id))
		utils.HandleServiceError(c, "Failed to retrieve subcategory", err)
		return
	}

	h.logger.Info("Subcategory retrieved successfully",
		zap.String("subcategory_id", response.ID),
		zap.String("name", response.Name))

	utils.OKResponse(c, "Subcategory retrieved successfully", response)
}

// GetSubcategoryByName handles GET /api/v1/subcategories/name/:name
// @Summary Get Subcategory by Name
// @Description Retrieve a specific subcategory by name
// @Tags Subcategories
// @Produce json
// @Param name path string true "Subcategory name" example(Water Soluble)
// @Success 200 {object} utils.Response{data=models.SubcategoryResponse} "Subcategory details"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Subcategory not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/subcategories/name/{name} [get]
func (h *SubcategoryHandler) GetSubcategoryByName(c *gin.Context) {
	h.logger.Info("Handling get subcategory by name request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get name from URL
	name := c.Param("name")
	if name == "" {
		h.logger.Error("Subcategory name is required but not provided")
		utils.BadRequestResponse(c, "Subcategory name is required", nil)
		return
	}

	h.logger.Debug("Fetching subcategory by name",
		zap.String("name", name))

	// Get subcategory by name
	response, err := h.subcategoryService.GetSubcategoryByName(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Subcategory not found",
			zap.Error(err),
			zap.String("name", name))
		utils.HandleServiceError(c, "Failed to retrieve subcategory", err)
		return
	}

	h.logger.Info("Subcategory retrieved successfully",
		zap.String("subcategory_id", response.ID),
		zap.String("name", response.Name))

	utils.OKResponse(c, "Subcategory retrieved successfully", response)
}

// GetSubcategoriesByCategory handles GET /api/v1/subcategories/category/:categoryId
// @Summary Get Subcategories by Category
// @Description Retrieve all subcategories for a specific category by ID with pagination
// @Tags Subcategories
// @Produce json
// @Param categoryId path string true "Category ID (format: CATG_xxxxxxxx)" example(CATG_12345678)
// @Param limit query integer false "Number of records to return (default: 50, max: 200)" example(50)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.PaginatedResponseModel{data=[]models.SubcategoryResponse} "Subcategories retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/subcategories/category/{categoryId} [get]
func (h *SubcategoryHandler) GetSubcategoriesByCategory(c *gin.Context) {
	h.logger.Info("Handling get subcategories by category request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get category ID from URL
	categoryID := c.Param("categoryId")
	if categoryID == "" {
		h.logger.Error("Category ID is required but not provided")
		utils.BadRequestResponse(c, "Category ID is required", nil)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	h.logger.Debug("Fetching subcategories by category",
		zap.String("category_id", categoryID),
		zap.Int("limit", params.Limit),
		zap.Int("offset", params.Offset))

	// Get subcategories by category ID with pagination
	response, total, err := h.subcategoryService.GetSubcategoriesByCategoryPaginated(c.Request.Context(), categoryID, params.Limit, params.Offset)
	if err != nil {
		h.logger.Error("Service error retrieving subcategories",
			zap.Error(err),
			zap.String("category_id", categoryID))
		utils.HandleServiceError(c, "Failed to retrieve subcategories", err)
		return
	}

	h.logger.Info("Subcategories retrieved successfully",
		zap.String("category_id", categoryID),
		zap.Int("count", len(response)),
		zap.Int64("total", total))

	utils.PaginatedOKResponse(c, response, total, params.Limit, params.Offset)
}

// GetAllSubcategories handles GET /api/v1/subcategories
// @Summary Get All Subcategories
// @Description Retrieve all subcategories with pagination
// @Tags Subcategories
// @Produce json
// @Param limit query integer false "Number of records to return (default: 50, max: 200)" example(50)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.PaginatedResponseModel{data=[]models.SubcategoryResponse} "Subcategories retrieved successfully"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/subcategories [get]
func (h *SubcategoryHandler) GetAllSubcategories(c *gin.Context) {
	h.logger.Info("Handling get all subcategories request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	h.logger.Debug("Calling subcategory service to get all subcategories",
		zap.Int("limit", params.Limit),
		zap.Int("offset", params.Offset))

	// Get all subcategories with pagination
	response, total, err := h.subcategoryService.GetAllSubcategoriesPaginated(c.Request.Context(), params.Limit, params.Offset)
	if err != nil {
		h.logger.Error("Service error retrieving all subcategories",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve subcategories", err)
		return
	}

	h.logger.Info("All subcategories retrieved successfully",
		zap.Int("subcategory_count", len(response)),
		zap.Int64("total", total))

	utils.PaginatedOKResponse(c, response, total, params.Limit, params.Offset)
}

// SearchSubcategories handles GET /api/v1/subcategories/search
// @Summary Search Subcategories
// @Description Search subcategories by name with pagination
// @Tags Subcategories
// @Produce json
// @Param q query string true "Search query" example(BULK)
// @Param limit query integer false "Number of records to return (default: 50, max: 200)" example(50)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.PaginatedResponseModel{data=[]models.SubcategoryResponse} "Subcategories found successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request - search query required"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/subcategories/search [get]
func (h *SubcategoryHandler) SearchSubcategories(c *gin.Context) {
	h.logger.Info("Handling search subcategories request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get search query
	query := c.Query("q")
	if query == "" {
		h.logger.Error("Search query is required but not provided")
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	h.logger.Debug("Calling subcategory service to search subcategories",
		zap.String("query", query),
		zap.Int("limit", params.Limit),
		zap.Int("offset", params.Offset))

	// Search subcategories
	response, total, err := h.subcategoryService.SearchSubcategories(c.Request.Context(), query, params.Limit, params.Offset)
	if err != nil {
		h.logger.Error("Service error searching subcategories",
			zap.Error(err),
			zap.String("query", query))
		utils.HandleServiceError(c, "Failed to search subcategories", err)
		return
	}

	h.logger.Info("Subcategories search completed successfully",
		zap.String("query", query),
		zap.Int("subcategory_count", len(response)),
		zap.Int64("total", total))

	utils.PaginatedOKResponse(c, response, total, params.Limit, params.Offset)
}

// UpdateSubcategory handles PATCH /api/v1/subcategories/:id
// @Summary Update Subcategory
// @Description Update an existing subcategory by ID
// @Tags Subcategories
// @Accept json
// @Produce json
// @Param id path string true "Subcategory ID (format: SCAT_xxxxxxxx)" example(SCAT_12345678)
// @Param request body models.UpdateSubcategoryRequest true "Updated subcategory data"
// @Success 200 {object} utils.Response{data=models.SubcategoryResponse} "Subcategory updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Subcategory not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - name already exists"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/subcategories/{id} [patch]
func (h *SubcategoryHandler) UpdateSubcategory(c *gin.Context) {
	h.logger.Info("Handling update subcategory request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Subcategory ID is required but not provided")
		utils.BadRequestResponse(c, "Subcategory ID is required", nil)
		return
	}

	var request models.UpdateSubcategoryRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		h.logger.Error("Invalid request body for update subcategory",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	h.logger.Debug("Calling service to update subcategory",
		zap.String("subcategory_id", id))

	// Update subcategory
	response, err := h.subcategoryService.UpdateSubcategory(c.Request.Context(), id, &request)
	if err != nil {
		h.logger.Error("Service error updating subcategory",
			zap.Error(err),
			zap.String("subcategory_id", id))
		utils.HandleServiceError(c, "Failed to update subcategory", err)
		return
	}

	h.logger.Info("Subcategory updated successfully",
		zap.String("subcategory_id", response.ID),
		zap.String("name", response.Name))

	utils.OKResponse(c, "Subcategory updated successfully", response)
}

// DeleteSubcategory handles DELETE /api/v1/subcategories/:id
// @Summary Delete Subcategory
// @Description Delete a subcategory by ID
// @Tags Subcategories
// @Produce json
// @Param id path string true "Subcategory ID (format: SCAT_xxxxxxxx)" example(SCAT_12345678)
// @Success 200 {object} utils.Response "Subcategory deleted successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Subcategory not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/subcategories/{id} [delete]
func (h *SubcategoryHandler) DeleteSubcategory(c *gin.Context) {
	h.logger.Info("Handling delete subcategory request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		h.logger.Error("Subcategory ID is required but not provided")
		utils.BadRequestResponse(c, "Subcategory ID is required", nil)
		return
	}

	h.logger.Debug("Calling service to delete subcategory",
		zap.String("subcategory_id", id))

	// Delete subcategory
	if err := h.subcategoryService.DeleteSubcategory(c.Request.Context(), id); err != nil {
		h.logger.Error("Service error deleting subcategory",
			zap.Error(err),
			zap.String("subcategory_id", id))
		utils.HandleServiceError(c, "Failed to delete subcategory", err)
		return
	}

	h.logger.Info("Subcategory deleted successfully",
		zap.String("subcategory_id", id))

	utils.OKResponse(c, "Subcategory deleted successfully", nil)
}

// RegisterRoutes registers all subcategory routes
func (h *SubcategoryHandler) RegisterRoutes(v1 *gin.RouterGroup) {
	subcategories := v1.Group("/subcategories")
	{
		// Public endpoints (read operations)
		subcategories.GET("", h.GetAllSubcategories)
		subcategories.GET("/search", h.SearchSubcategories)
		subcategories.GET("/name/:name", h.GetSubcategoryByName)
		subcategories.GET("/category/:categoryId", h.GetSubcategoriesByCategory)
		subcategories.GET("/:id", h.GetSubcategory)

		// Authenticated endpoints (write operations)
		subcategories.POST("", h.CreateSubcategory)
		subcategories.PATCH("/:id", h.UpdateSubcategory)
		subcategories.DELETE("/:id", h.DeleteSubcategory)
	}
}
