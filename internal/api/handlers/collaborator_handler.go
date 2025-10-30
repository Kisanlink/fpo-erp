package handlers

import (
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

// CollaboratorHandler handles collaborator HTTP requests
type CollaboratorHandler struct {
	collaboratorService *services.CollaboratorService
	aaaMiddleware       *aaa.AAAMiddleware
}

// NewCollaboratorHandler creates a new collaborator handler
func NewCollaboratorHandler(collaboratorService *services.CollaboratorService, aaaMiddleware *aaa.AAAMiddleware) *CollaboratorHandler {
	return &CollaboratorHandler{
		collaboratorService: collaboratorService,
		aaaMiddleware:       aaaMiddleware,
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
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators [post]
func (h *CollaboratorHandler) CreateCollaborator(c *gin.Context) {
	var request models.CreateCollaboratorRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
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

	// Log organization context for debugging
	utils.Debug("Creating collaborator for organization:", organizationID)
	utils.Debug("Organization Name:", c.GetString("organization_name"))

	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	// Create collaborator
	response, err := h.collaboratorService.CreateCollaborator(c.Request.Context(), &request, userID, jwtToken)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create collaborator", err)
		return
	}

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

	// Get collaborator
	response, err := h.collaboratorService.GetCollaborator(c.Request.Context(), id, jwtToken)
	if err != nil {
		utils.NotFoundResponse(c, "Collaborator not found")
		return
	}

	utils.OKResponse(c, "Collaborator retrieved successfully", response)
}

// GetAllCollaborators handles GET /api/v1/collaborators
// @Summary Get All Collaborators
// @Description Retrieve all collaborators (requires authentication)
// @Tags Collaborators
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.CollaboratorResponse} "Collaborators retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators [get]
func (h *CollaboratorHandler) GetAllCollaborators(c *gin.Context) {
	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	// Get all collaborators
	response, err := h.collaboratorService.GetAllCollaborators(c.Request.Context(), jwtToken)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve collaborators", err)
		return
	}

	utils.OKResponse(c, "Collaborators retrieved successfully", response)
}

// GetActiveCollaborators handles GET /api/v1/collaborators/active
// @Summary Get Active Collaborators
// @Description Retrieve all active collaborators (requires authentication)
// @Tags Collaborators
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.CollaboratorResponse} "Active collaborators retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators/active [get]
func (h *CollaboratorHandler) GetActiveCollaborators(c *gin.Context) {
	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	// Get active collaborators
	response, err := h.collaboratorService.GetActiveCollaborators(c.Request.Context(), jwtToken)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve active collaborators", err)
		return
	}

	utils.OKResponse(c, "Active collaborators retrieved successfully", response)
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
// @Failure 404 {object} utils.ErrorResponseModel "Collaborator not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators/{id} [put]
func (h *CollaboratorHandler) UpdateCollaborator(c *gin.Context) {
	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Collaborator ID is required", nil)
		return
	}

	var request models.UpdateCollaboratorRequest

	// Validate request
	if err := utils.ValidateRequest(c, &request); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// ✅ Extract organization ID from context (set by auth middleware)
	organizationID := c.GetString("organization_id")
	if organizationID == "" {
		utils.BadRequestResponse(c, "Organization context not found. Ensure you're authenticated.", nil)
		return
	}

	// Log organization context for debugging
	utils.Debug("Updating collaborator for organization:", organizationID)

	// Extract JWT token for AAA service calls
	jwtToken := c.GetString("jwt_token")
	if jwtToken == "" {
		utils.UnauthorizedResponse(c, "Missing authentication token")
		return
	}

	// Update collaborator
	response, err := h.collaboratorService.UpdateCollaborator(c.Request.Context(), id, &request, jwtToken)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update collaborator", err)
		return
	}

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
// @Failure 404 {object} utils.ErrorResponseModel "Collaborator not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators/{id} [delete]
func (h *CollaboratorHandler) DeleteCollaborator(c *gin.Context) {
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
	if err := h.collaboratorService.DeleteCollaborator(c.Request.Context(), id, jwtToken); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete collaborator", err)
		return
	}

	utils.OKResponse(c, "Collaborator deleted successfully", nil)
}

// SearchCollaborators handles GET /api/v1/collaborators/search
// @Summary Search Collaborators
// @Description Search collaborators by company name (requires authentication)
// @Tags Collaborators
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {object} utils.Response{data=[]models.CollaboratorResponse} "Search results"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/collaborators/search [get]
func (h *CollaboratorHandler) SearchCollaborators(c *gin.Context) {
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

	// Search collaborators
	response, err := h.collaboratorService.SearchCollaborators(c.Request.Context(), query, jwtToken)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to search collaborators", err)
		return
	}

	utils.OKResponse(c, "Search completed successfully", response)
}

// RegisterRoutes registers all collaborator routes
func (h *CollaboratorHandler) RegisterRoutes(router *gin.RouterGroup) {
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
		collaborators.GET("/:id", h.aaaMiddleware.RequireOrgPermission("collaborator", "read"), h.GetCollaborator)

		// Update: AAA HTTP service will validate addresses permissions internally
		collaborators.PUT("/:id", h.aaaMiddleware.RequireOrgPermission("collaborator", "update"), h.UpdateCollaborator)

		// Delete: AAA HTTP service will validate addresses permissions internally
		collaborators.DELETE("/:id", h.aaaMiddleware.RequireOrgPermission("collaborator", "delete"), h.DeleteCollaborator)
	}
}
