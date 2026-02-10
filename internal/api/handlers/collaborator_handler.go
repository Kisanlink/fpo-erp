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

// CollaboratorHandler handles collaborator HTTP requests
type CollaboratorHandler struct {
	collaboratorService interfaces.CollaboratorServiceInterface
	aaaMiddleware       *aaa.AAAMiddleware
	logger              logger.Logger
}

// NewCollaboratorHandler creates a new collaborator handler
func NewCollaboratorHandler(collaboratorService interfaces.CollaboratorServiceInterface, aaaMiddleware *aaa.AAAMiddleware, logger logger.Logger) *CollaboratorHandler {
	return &CollaboratorHandler{
		collaboratorService: collaboratorService,
		aaaMiddleware:       aaaMiddleware,
		logger:              logger,
	}
}

// CreateCollaborator handles POST /api/v1/collaborators
// @Summary Create Collaborator
// @Description Create a new collaborator/vendor (requires authentication)
// @Tags Collaborators
// @Accept json
// @Produce json
// @Param request body models.CreateCollaboratorRequest true "Collaborator data"
// @Success 201 {object} utils.Response{data=models.CollaboratorResponse} "Collaborator created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators [post]
func (h *CollaboratorHandler) CreateCollaborator(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling create collaborator request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	var request models.CreateCollaboratorRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		// 2. Validation Error Log
		h.logger.Error("Invalid request body for create collaborator",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Get authenticated user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system" // Fallback for unauthenticated contexts
	}

	// ✅ Extract organization ID from context (set by auth middleware)
	organizationID := c.GetString("organization_id")
	if organizationID == "" {
		utils.BadRequestResponse(c, "Organization context not found. Ensure you're authenticated.", nil)
		return
	}

	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to create collaborator",
		zap.String("company_name", request.CompanyName),
		zap.String("organization_id", organizationID))

	// Create collaborator
	response, err := h.collaboratorService.CreateCollaborator(c.Request.Context(), &request, organizationID, userID, jwtToken)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error creating collaborator",
			zap.Error(err),
			zap.String("company_name", request.CompanyName))
		utils.HandleServiceError(c, "Failed to create collaborator", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Collaborator created successfully",
		zap.String("collaborator_id", response.ID),
		zap.String("company_name", response.CompanyName))

	// TODO: Sync collaborator to master table service via gRPC
	// This should be done asynchronously to avoid blocking the response
	// Example implementation:
	//
	// go func() {
	//     err := h.syncCollaboratorToMasterTable(c.Request.Context(), response, organizationID)
	//     if err != nil {
	//         utils.Error("Failed to sync collaborator to master table:",
	//             "collaborator_id", response.ID,
	//             "organization_id", organizationID,
	//             "error", err)
	//         // Add to retry queue for later processing
	//     }
	// }()

	utils.CreatedResponse(c, "Collaborator created successfully", response)
}

// GetCollaborator handles GET /api/v1/collaborators/:id
// @Summary Get Collaborator
// @Description Retrieve a specific collaborator by ID
// @Tags Collaborators
// @Produce json
// @Param id path string true "Collaborator ID (format: CLAB_xxxxxxxx)" example(CLAB_12345678)
// @Success 200 {object} utils.Response{data=models.CollaboratorResponse} "Collaborator details"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Collaborator not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /api/v1/collaborators/{id} [get]
func (h *CollaboratorHandler) GetCollaborator(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get collaborator request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Collaborator ID is required", nil)
		return
	}

	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to get collaborator",
		zap.String("collaborator_id", id))

	// Get collaborator
	response, err := h.collaboratorService.GetCollaborator(c.Request.Context(), id, jwtToken)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error getting collaborator",
			zap.Error(err),
			zap.String("collaborator_id", id))
		utils.NotFoundResponse(c, "Collaborator not found")
		return
	}

	// 5. Success Log
	h.logger.Info("Collaborator retrieved successfully",
		zap.String("collaborator_id", response.ID),
		zap.String("company_name", response.CompanyName))

	utils.OKResponse(c, "Collaborator retrieved successfully", response)
}

// GetAllCollaborators handles GET /api/v1/collaborators
// @Summary Get All Collaborators
// @Description Retrieve all collaborators with pagination (requires authentication)
// @Tags Collaborators
// @Produce json
// @Param limit query int false "Number of results per page (default: 50, max: 200)" default(50)
// @Param offset query int false "Number of results to skip (default: 0)" default(0)
// @Success 200 {object} utils.PaginatedResponseModel{data=[]models.CollaboratorResponse} "Collaborators retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators [get]
func (h *CollaboratorHandler) GetAllCollaborators(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get all collaborators request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// 3. Service Call Log
	h.logger.Debug("Calling service to get all collaborators",
		zap.Int("limit", params.Limit),
		zap.Int("offset", params.Offset))

	// Get all collaborators
	response, total, err := h.collaboratorService.GetAllCollaborators(c.Request.Context(), jwtToken, params.Limit, params.Offset)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error getting all collaborators",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve collaborators", err)
		return
	}

	// 5. Success Log
	h.logger.Info("All collaborators retrieved successfully",
		zap.Int("count", len(response)),
		zap.Int64("total", total))

	utils.PaginatedOKResponse(c, response, total, params.Limit, params.Offset)
}

// GetActiveCollaborators handles GET /api/v1/collaborators/active
// @Summary Get Active Collaborators
// @Description Retrieve all active collaborators with pagination (requires authentication)
// @Tags Collaborators
// @Produce json
// @Param limit query int false "Number of results per page (default: 50, max: 200)" default(50)
// @Param offset query int false "Number of results to skip (default: 0)" default(0)
// @Success 200 {object} utils.PaginatedResponseModel{data=[]models.CollaboratorResponse} "Active collaborators retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators/active [get]
func (h *CollaboratorHandler) GetActiveCollaborators(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get active collaborators request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// 3. Service Call Log
	h.logger.Debug("Calling service to get active collaborators",
		zap.Int("limit", params.Limit),
		zap.Int("offset", params.Offset))

	// Get active collaborators
	response, total, err := h.collaboratorService.GetActiveCollaborators(c.Request.Context(), jwtToken, params.Limit, params.Offset)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error getting active collaborators",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve active collaborators", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Active collaborators retrieved successfully",
		zap.Int("count", len(response)),
		zap.Int64("total", total))

	utils.PaginatedOKResponse(c, response, total, params.Limit, params.Offset)
}

// UpdateCollaborator handles PUT /api/v1/collaborators/:id
// @Summary Update Collaborator
// @Description Update an existing collaborator (requires authentication)
// @Tags Collaborators
// @Accept json
// @Produce json
// @Param id path string true "Collaborator ID (format: CLAB_xxxxxxxx)" example(CLAB_12345678)
// @Param request body models.UpdateCollaboratorRequest true "Updated collaborator data"
// @Success 200 {object} utils.Response{data=models.CollaboratorResponse} "Collaborator updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Collaborator not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators/{id} [put]
func (h *CollaboratorHandler) UpdateCollaborator(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling update collaborator request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Collaborator ID is required", nil)
		return
	}

	var request models.UpdateCollaboratorRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		// 2. Validation Error Log
		h.logger.Error("Invalid request body for update collaborator",
			zap.Error(err),
			zap.String("collaborator_id", id))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// ✅ Extract organization ID from context (set by auth middleware)
	organizationID := c.GetString("organization_id")
	if organizationID == "" {
		utils.BadRequestResponse(c, "Organization context not found. Ensure you're authenticated.", nil)
		return
	}

	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	// Update collaborator
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system"
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to update collaborator",
		zap.String("collaborator_id", id),
		zap.String("organization_id", organizationID))

	response, err := h.collaboratorService.UpdateCollaborator(c.Request.Context(), id, &request, organizationID, userID, jwtToken)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error updating collaborator",
			zap.Error(err),
			zap.String("collaborator_id", id))
		utils.HandleServiceError(c, "Failed to update collaborator", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Collaborator updated successfully",
		zap.String("collaborator_id", response.ID),
		zap.String("company_name", response.CompanyName))

	// TODO: Sync updated collaborator to master table service via gRPC
	// This should be done asynchronously to avoid blocking the response
	// Example implementation:
	//
	// go func() {
	//     err := h.syncCollaboratorToMasterTable(c.Request.Context(), response, organizationID)
	//     if err != nil {
	//         utils.Error("Failed to sync collaborator update to master table:",
	//             "collaborator_id", response.ID,
	//             "organization_id", organizationID,
	//             "error", err)
	//         // Add to retry queue for later processing
	//     }
	// }()

	utils.OKResponse(c, "Collaborator updated successfully", response)
}

// DeleteCollaborator handles DELETE /api/v1/collaborators/:id
// @Summary Delete Collaborator
// @Description Delete a collaborator (soft delete, requires authentication)
// @Tags Collaborators
// @Produce json
// @Param id path string true "Collaborator ID (format: CLAB_xxxxxxxx)" example(CLAB_12345678)
// @Success 200 {object} utils.Response "Collaborator deleted successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Collaborator not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators/{id} [delete]
func (h *CollaboratorHandler) DeleteCollaborator(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling delete collaborator request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Collaborator ID is required", nil)
		return
	}

	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	// Delete collaborator
	organizationID := c.GetString("organization_id")
	if organizationID == "" {
		utils.BadRequestResponse(c, "Organization context not found. Ensure you're authenticated.", nil)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to delete collaborator",
		zap.String("collaborator_id", id),
		zap.String("organization_id", organizationID))

	if err := h.collaboratorService.DeleteCollaborator(c.Request.Context(), id, organizationID, jwtToken); err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error deleting collaborator",
			zap.Error(err),
			zap.String("collaborator_id", id))
		utils.HandleServiceError(c, "Failed to delete collaborator", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Collaborator deleted successfully",
		zap.String("collaborator_id", id))

	utils.OKResponse(c, "Collaborator deleted successfully", nil)
}

// SearchCollaborators handles GET /api/v1/collaborators/search
// @Summary Search Collaborators
// @Description Search collaborators by company name with pagination (requires authentication)
// @Tags Collaborators
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Number of results per page (default: 50, max: 200)" default(50)
// @Param offset query int false "Number of results to skip (default: 0)" default(0)
// @Success 200 {object} utils.PaginatedResponseModel{data=[]models.CollaboratorResponse} "Search results"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators/search [get]
func (h *CollaboratorHandler) SearchCollaborators(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling search collaborators request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get search query
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// 3. Service Call Log
	h.logger.Debug("Calling service to search collaborators",
		zap.String("query", query),
		zap.Int("limit", params.Limit),
		zap.Int("offset", params.Offset))

	// Search collaborators
	response, total, err := h.collaboratorService.SearchCollaborators(c.Request.Context(), query, jwtToken, params.Limit, params.Offset)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error searching collaborators",
			zap.Error(err),
			zap.String("query", query))
		utils.HandleServiceError(c, "Failed to search collaborators", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Collaborators search completed successfully",
		zap.String("query", query),
		zap.Int("count", len(response)),
		zap.Int64("total", total))

	utils.PaginatedOKResponse(c, response, total, params.Limit, params.Offset)
}

// GetCollaboratorStats handles GET /api/v1/collaborators/:id/stats
// @Summary Get Collaborator Transaction Statistics
// @Description Retrieve transaction statistics for a specific collaborator
// @Tags Collaborators
// @Produce json
// @Param id path string true "Collaborator ID (format: CLAB_xxxxxxxx)" example(CLAB_12345678)
// @Success 200 {object} utils.Response{data=models.CollaboratorStats} "Collaborator statistics"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Collaborator not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators/{id}/stats [get]
func (h *CollaboratorHandler) GetCollaboratorStats(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get collaborator stats request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Collaborator ID is required", nil)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to get collaborator stats",
		zap.String("collaborator_id", id))

	// Get collaborator stats
	stats, err := h.collaboratorService.GetCollaboratorStats(c.Request.Context(), id)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error getting collaborator stats",
			zap.Error(err),
			zap.String("collaborator_id", id))
		utils.HandleServiceError(c, "Failed to retrieve collaborator stats", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Collaborator stats retrieved successfully",
		zap.String("collaborator_id", id),
		zap.String("company_name", stats.CompanyName),
		zap.Int64("po_count", stats.POCount),
		zap.Int64("grn_count", stats.GRNCount))

	utils.OKResponse(c, "Collaborator statistics retrieved successfully", stats)
}

// GetAllCollaboratorsStats handles GET /api/v1/collaborators/stats
// @Summary Get All Collaborators Stats
// @Description Retrieve transaction statistics for all collaborators (PO counts) and total PO count
// @Tags Collaborators
// @Produce json
// @Success 200 {object} utils.Response{data=models.AllCollaboratorsStatsResponse} "All collaborators statistics"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators/stats [get]
func (h *CollaboratorHandler) GetAllCollaboratorsStats(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get all collaborators stats request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// 3. Service Call Log
	h.logger.Debug("Calling service to get all collaborators stats")

	// Get all collaborators stats
	stats, err := h.collaboratorService.GetAllCollaboratorsStats(c.Request.Context())
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error getting all collaborators stats",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve all collaborators stats", err)
		return
	}

	// 5. Success Log
	h.logger.Info("All collaborators stats retrieved successfully",
		zap.Int("collaborators_count", len(stats.Collaborators)),
		zap.Int64("total_po_count", stats.TotalPOCount))

	utils.OKResponse(c, "All collaborators statistics retrieved successfully", stats)
}

// RegisterRoutes registers all collaborator routes
func (h *CollaboratorHandler) RegisterRoutes(router *gin.RouterGroup, variantHandler *ProductVariantHandler) {
	collaborators := router.Group("/collaborators")
	{
		// Apply authentication middleware to all routes
		collaborators.Use(h.aaaMiddleware.Authenticate())

		// Create: AAA HTTP service will validate addresses permissions internally
		collaborators.POST("", h.aaaMiddleware.RequireOrgPermission("collaborator", "create"), h.CreateCollaborator)

		// Read operations: AAA HTTP service will validate addresses permissions internally
		collaborators.GET("", h.aaaMiddleware.RequireOrgPermission("collaborator", "read"), h.GetAllCollaborators)
		collaborators.GET("/active", h.aaaMiddleware.RequireOrgPermission("collaborator", "read"), h.GetActiveCollaborators)
		collaborators.GET("/search", h.aaaMiddleware.RequireOrgPermission("collaborator", "read"), h.SearchCollaborators)
		collaborators.GET("/stats", h.aaaMiddleware.RequireOrgPermission("collaborator", "read"), h.GetAllCollaboratorsStats)
		collaborators.GET("/:id", h.aaaMiddleware.RequireOrgPermission("collaborator", "read"), h.GetCollaborator)
		collaborators.GET("/:id/stats", h.aaaMiddleware.RequireOrgPermission("collaborator", "read"), h.GetCollaboratorStats)

		// Nested route: Get variants by collaborator
		collaborators.GET("/:id/variants", h.aaaMiddleware.RequireOrgPermission("variant", "read"), variantHandler.GetVariantsByCollaborator)

		// Update: AAA HTTP service will validate addresses permissions internally
		collaborators.PUT("/:id", h.aaaMiddleware.RequireOrgPermission("collaborator", "update"), h.UpdateCollaborator)

		// Delete: AAA HTTP service will validate addresses permissions internally
		collaborators.DELETE("/:id", h.aaaMiddleware.RequireOrgPermission("collaborator", "delete"), h.DeleteCollaborator)
	}
}
