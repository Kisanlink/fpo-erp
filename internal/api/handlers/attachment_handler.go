package handlers

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

// AttachmentHandler handles attachment HTTP requests
type AttachmentHandler struct {
	attachmentService interfaces.AttachmentServiceInterface
	aaaMiddleware     *aaa.AAAMiddleware
}

// NewAttachmentHandler creates a new attachment handler
func NewAttachmentHandler(attachmentService interfaces.AttachmentServiceInterface, aaaMiddleware *aaa.AAAMiddleware) *AttachmentHandler {
	return &AttachmentHandler{
		attachmentService: attachmentService,
		aaaMiddleware:     aaaMiddleware,
	}
}

// RegisterRoutes registers attachment routes
func (h *AttachmentHandler) RegisterRoutes(router *gin.RouterGroup) {
	attachments := router.Group("/attachments")
	{
		// Apply authentication middleware
		attachments.Use(h.aaaMiddleware.Authenticate())

		// Create/Delete routes (organization-scoped)
		attachments.POST("", h.aaaMiddleware.RequireOrgPermission("attachment", "create"), h.UploadAttachment)
		attachments.DELETE("/:id", h.aaaMiddleware.RequireOrgPermission("attachment", "delete"), h.DeleteAttachment)

		// Read routes (organization-scoped)
		attachments.GET("/:id/download", h.aaaMiddleware.RequireOrgPermission("attachment", "read"), h.DownloadAttachment)
		attachments.GET("/:id/url", h.aaaMiddleware.RequireOrgPermission("attachment", "read"), h.GenerateDownloadURL)
		attachments.GET("", h.aaaMiddleware.RequireOrgPermission("attachment", "read"), h.GetAttachments)
		attachments.GET("/:id", h.aaaMiddleware.RequireOrgPermission("attachment", "read"), h.GetAttachment)
		attachments.GET("/:id/info", h.aaaMiddleware.RequireOrgPermission("attachment", "read"), h.GetAttachmentInfo)

		// Entity-based route (for logos, POs, GRNs, etc.) - organization-scoped
		attachments.GET("/entity/:entity_type/:entity_id", h.aaaMiddleware.RequireOrgPermission("attachment", "read"), h.GetAttachmentsByEntity)
	}
}

// UploadAttachment handles file upload
// @Summary Upload Attachment
// @Description Upload a file attachment for any entity (logo, PO, GRN, etc.) (requires authentication)
// @Tags Attachments
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param entity_type formData string true "Entity type: logo, po, grn, etc." example(logo)
// @Param entity_id formData string true "Entity ID (e.g., CLAB_xxx, PO_xxx, GRN_xxx)" example(CLAB_12345678)
// @Success 201 {object} utils.Response{data=models.AttachmentResponse} "Attachment uploaded successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/attachments [post]
func (h *AttachmentHandler) UploadAttachment(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User ID not found in context")
		return
	}

	// Get form data
	entityType := c.PostForm("entity_type")
	entityID := c.PostForm("entity_id")

	// Validate required fields
	if entityType == "" {
		utils.BadRequestResponse(c, "entity_type is required", nil)
		return
	}
	if entityID == "" {
		utils.BadRequestResponse(c, "entity_id is required", nil)
		return
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequestResponse(c, "No file uploaded", err)
		return
	}

	// Upload attachment
	attachment, err := h.attachmentService.UploadAttachment(c.Request.Context(), file, entityType, entityID, userIDStr.(string))
	if err != nil {
		utils.HandleServiceError(c, "Failed to upload attachment", err)
		return
	}

	utils.CreatedResponse(c, "Attachment uploaded successfully", attachment)
}

// GetAttachment retrieves an attachment by ID
// @Summary Get Attachment
// @Description Retrieve attachment metadata by ID
// @Tags Attachments
// @Produce json
// @Param id path string true "Attachment ID" example(ATT_12345678)
// @Success 200 {object} utils.Response{data=models.AttachmentResponse} "Attachment retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Attachment not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/attachments/{id} [get]
func (h *AttachmentHandler) GetAttachment(c *gin.Context) {
	id := c.Param("id")

	attachment, err := h.attachmentService.GetAttachment(id)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve attachment", err)
		return
	}

	utils.OKResponse(c, "Attachment retrieved successfully", attachment)
}

// GetAttachments retrieves attachments with optional filters
// @Summary Get Attachments
// @Description Retrieve attachments with optional filters and pagination
// @Tags Attachments
// @Produce json
// @Param entity_type query string false "Filter by entity type (logo, po, grn, etc.)" example(logo)
// @Param entity_id query string false "Filter by entity ID" example(CLAB_12345678)
// @Param limit query integer false "Number of records to return (default: 10, max: 100)" example(10)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.Response{data=[]models.AttachmentResponse} "Attachments retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/attachments [get]
func (h *AttachmentHandler) GetAttachments(c *gin.Context) {
	// Get query parameters
	entityType := c.Query("entity_type")
	entityID := c.Query("entity_id")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	// Parse pagination parameters
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Convert empty strings to nil
	var entityTypePtr, entityIDPtr *string
	if entityType != "" {
		entityTypePtr = &entityType
	}
	if entityID != "" {
		entityIDPtr = &entityID
	}

	// Get attachments
	attachments, err := h.attachmentService.GetAttachments(entityTypePtr, entityIDPtr, limit, offset)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve attachments", err)
		return
	}

	utils.OKResponse(c, "Attachments retrieved successfully", attachments)
}

// DownloadAttachment downloads an attachment file
// @Summary Download Attachment
// @Description Download the actual file content of an attachment
// @Tags Attachments
// @Produce application/octet-stream
// @Param id path string true "Attachment ID" example(ATT_12345678)
// @Success 200 {file} file "File download"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Attachment not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/attachments/{id}/download [get]
func (h *AttachmentHandler) DownloadAttachment(c *gin.Context) {
	id := c.Param("id")

	// Download file
	fileReader, contentType, err := h.attachmentService.DownloadAttachment(c.Request.Context(), id)
	if err != nil {
		utils.HandleServiceError(c, "Failed to download attachment", err)
		return
	}
	defer func() {
		if closer, ok := fileReader.(io.Closer); ok {
			closer.Close()
		}
	}()

	// Set response headers
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename=attachment")

	// Stream the file
	r, ok := fileReader.(io.Reader)
	if !ok {
		utils.HandleServiceError(c, "Attachment stream is not readable", nil)
		return
	}
	c.DataFromReader(http.StatusOK, -1, contentType, r, nil)
}

// GenerateDownloadURL generates a presigned URL for file download
// @Summary Generate Download URL
// @Description Generate a presigned URL for downloading an attachment
// @Tags Attachments
// @Produce json
// @Param id path string true "Attachment ID" example(ATT_12345678)
// @Param expiration query integer false "URL expiration in seconds (default: 3600, max: 86400)" example(3600)
// @Success 200 {object} utils.Response{data=object{download_url=string,expires_in=integer}} "Download URL generated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Attachment not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/attachments/{id}/url [get]
func (h *AttachmentHandler) GenerateDownloadURL(c *gin.Context) {
	id := c.Param("id")
	expirationStr := c.DefaultQuery("expiration", "3600") // Default 1 hour

	// Parse expiration
	expirationSeconds, err := strconv.Atoi(expirationStr)
	if err != nil || expirationSeconds <= 0 {
		expirationSeconds = 3600
	}
	if expirationSeconds > 86400 { // Max 24 hours
		expirationSeconds = 86400
	}

	expiration := time.Duration(expirationSeconds) * time.Second

	// Generate URL
	url, err := h.attachmentService.GenerateDownloadURL(c.Request.Context(), id, expiration)
	if err != nil {
		utils.HandleServiceError(c, "Failed to generate download URL", err)
		return
	}

	utils.OKResponse(c, "Download URL generated successfully", gin.H{
		"download_url": url,
		"expires_in":   expirationSeconds,
	})
}

// GetAttachmentInfo gets detailed information about an attachment
// @Summary Get Attachment Info
// @Description Get detailed information about an attachment including metadata
// @Tags Attachments
// @Produce json
// @Param id path string true "Attachment ID" example(ATT_12345678)
// @Success 200 {object} utils.Response{data=models.AttachmentInfoResponse} "Attachment info retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Attachment not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/attachments/{id}/info [get]
func (h *AttachmentHandler) GetAttachmentInfo(c *gin.Context) {
	id := c.Param("id")

	info, err := h.attachmentService.GetAttachmentInfo(c.Request.Context(), id)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve attachment info", err)
		return
	}

	utils.OKResponse(c, "Attachment info retrieved successfully", info)
}

// GetAttachmentsByEntity retrieves all attachments for a specific entity
// @Summary Get Attachments by Entity
// @Description Retrieve all attachments associated with a specific entity (logo, PO, GRN, etc.)
// @Tags Attachments
// @Produce json
// @Param entity_type path string true "Entity type (logo, po, grn, etc.)" example(logo)
// @Param entity_id path string true "Entity ID (CLAB_xxx, PO_xxx, GRN_xxx, etc.)" example(CLAB_12345678)
// @Success 200 {object} utils.Response{data=[]models.AttachmentResponse} "Entity attachments retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/attachments/entity/{entity_type}/{entity_id} [get]
func (h *AttachmentHandler) GetAttachmentsByEntity(c *gin.Context) {
	entityType := c.Param("entity_type")
	entityID := c.Param("entity_id")

	if entityType == "" || entityID == "" {
		utils.BadRequestResponse(c, "entity_type and entity_id are required", nil)
		return
	}

	attachments, err := h.attachmentService.GetAttachmentsByEntity(entityType, entityID)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve entity attachments", err)
		return
	}

	utils.OKResponse(c, "Entity attachments retrieved successfully", attachments)
}

// DeleteAttachment deletes an attachment
// @Summary Delete Attachment
// @Description Delete an attachment and its associated file
// @Tags Attachments
// @Produce json
// @Param id path string true "Attachment ID" example(ATT_12345678)
// @Success 200 {object} utils.Response "Attachment deleted successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Attachment not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/attachments/{id} [delete]
func (h *AttachmentHandler) DeleteAttachment(c *gin.Context) {
	id := c.Param("id")

	err := h.attachmentService.DeleteAttachment(c.Request.Context(), id)
	if err != nil {
		utils.HandleServiceError(c, "Failed to delete attachment", err)
		return
	}

	utils.OKResponse(c, "Attachment deleted successfully", nil)
}
