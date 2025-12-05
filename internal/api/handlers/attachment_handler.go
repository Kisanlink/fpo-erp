package handlers

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"kisanlink-erp/internal/aaa"
	logger "kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AttachmentHandler handles attachment HTTP requests
type AttachmentHandler struct {
	attachmentService interfaces.AttachmentServiceInterface
	aaaMiddleware     *aaa.AAAMiddleware
	logger            logger.Logger
}

// NewAttachmentHandler creates a new attachment handler
func NewAttachmentHandler(attachmentService interfaces.AttachmentServiceInterface, aaaMiddleware *aaa.AAAMiddleware, logger logger.Logger) *AttachmentHandler {
	return &AttachmentHandler{
		attachmentService: attachmentService,
		logger:            logger,
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
	// 1. Entry Log
	h.logger.Info("Handling upload attachment request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get user ID from context (set by auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for attachment upload")
		utils.UnauthorizedResponse(c, "User ID not found in context")
		return
	}

	// Get form data
	entityType := c.PostForm("entity_type")
	entityID := c.PostForm("entity_id")

	// 2. Validation Error
	if entityType == "" {
		h.logger.Error("Invalid request: entity_type is required")
		utils.BadRequestResponse(c, "entity_type is required", nil)
		return
	}
	if entityID == "" {
		h.logger.Error("Invalid request: entity_id is required")
		utils.BadRequestResponse(c, "entity_id is required", nil)
		return
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Error("Invalid request: no file uploaded",
			zap.Error(err))
		utils.BadRequestResponse(c, "No file uploaded", err)
		return
	}

	// 3. Service Call
	h.logger.Debug("Calling service to upload attachment",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
		zap.String("filename", file.Filename))

	attachment, err := h.attachmentService.UploadAttachment(c.Request.Context(), file, entityType, entityID, userIDStr.(string))

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to upload attachment via service",
			zap.Error(err),
			zap.String("entity_type", entityType),
			zap.String("entity_id", entityID))
		utils.HandleServiceError(c, "Failed to upload attachment", err)
		return
	}

	// 5. Success
	h.logger.Info("Attachment upload completed successfully via handler",
		zap.String("attachment_id", attachment.ID))
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
	// 1. Entry Log
	h.logger.Info("Handling get attachment request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	id := c.Param("id")

	// 3. Service Call
	h.logger.Debug("Calling service to get attachment",
		zap.String("id", id))

	attachment, err := h.attachmentService.GetAttachment(id)

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to retrieve attachment via service",
			zap.Error(err),
			zap.String("id", id))
		utils.HandleServiceError(c, "Failed to retrieve attachment", err)
		return
	}

	// 5. Success
	h.logger.Info("Attachment retrieved successfully via handler",
		zap.String("id", id))
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
	// 1. Entry Log
	h.logger.Info("Handling get attachments request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

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

	// 3. Service Call
	h.logger.Debug("Calling service to get attachments",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	attachments, err := h.attachmentService.GetAttachments(c.Request.Context(), entityTypePtr, entityIDPtr, limit, offset)

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to retrieve attachments via service",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve attachments", err)
		return
	}

	// 5. Success
	h.logger.Info("Attachments retrieved successfully via handler",
		zap.Int("count", len(attachments)))
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
	// 1. Entry Log
	h.logger.Info("Handling download attachment request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	id := c.Param("id")

	// 3. Service Call
	h.logger.Debug("Calling service to download attachment",
		zap.String("id", id))

	fileReader, contentType, err := h.attachmentService.DownloadAttachment(c.Request.Context(), id)

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to download attachment via service",
			zap.Error(err),
			zap.String("id", id))
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
		h.logger.Error("Attachment stream is not readable",
			zap.String("id", id))
		utils.HandleServiceError(c, "Attachment stream is not readable", nil)
		return
	}

	// 5. Success
	h.logger.Info("Attachment download completed successfully via handler",
		zap.String("id", id),
		zap.String("content_type", contentType))
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
	// 1. Entry Log
	h.logger.Info("Handling generate download URL request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

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

	// 3. Service Call
	h.logger.Debug("Calling service to generate download URL",
		zap.String("id", id),
		zap.Int("expiration_seconds", expirationSeconds))

	url, err := h.attachmentService.GenerateDownloadURL(c.Request.Context(), id, expiration)

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to generate download URL via service",
			zap.Error(err),
			zap.String("id", id))
		utils.HandleServiceError(c, "Failed to generate download URL", err)
		return
	}

	// 5. Success
	h.logger.Info("Download URL generated successfully via handler",
		zap.String("id", id),
		zap.Int("expires_in", expirationSeconds))
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
	// 1. Entry Log
	h.logger.Info("Handling get attachment info request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	id := c.Param("id")

	// 3. Service Call
	h.logger.Debug("Calling service to get attachment info",
		zap.String("id", id))

	info, err := h.attachmentService.GetAttachmentInfo(c.Request.Context(), id)

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to retrieve attachment info via service",
			zap.Error(err),
			zap.String("id", id))
		utils.HandleServiceError(c, "Failed to retrieve attachment info", err)
		return
	}

	// 5. Success
	h.logger.Info("Attachment info retrieved successfully via handler",
		zap.String("id", id))
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
	// 1. Entry Log
	h.logger.Info("Handling get attachments by entity request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	entityType := c.Param("entity_type")
	entityID := c.Param("entity_id")

	// 2. Validation Error
	if entityType == "" || entityID == "" {
		h.logger.Error("Invalid request: entity_type and entity_id are required")
		utils.BadRequestResponse(c, "entity_type and entity_id are required", nil)
		return
	}

	// 3. Service Call
	h.logger.Debug("Calling service to get attachments by entity",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID))

	attachments, err := h.attachmentService.GetAttachmentsByEntity(c.Request.Context(), entityType, entityID)

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to retrieve entity attachments via service",
			zap.Error(err),
			zap.String("entity_type", entityType),
			zap.String("entity_id", entityID))
		utils.HandleServiceError(c, "Failed to retrieve entity attachments", err)
		return
	}

	// 5. Success
	h.logger.Info("Entity attachments retrieved successfully via handler",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
		zap.Int("count", len(attachments)))
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
	// 1. Entry Log
	h.logger.Info("Handling delete attachment request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	id := c.Param("id")

	// 3. Service Call
	h.logger.Debug("Calling service to delete attachment",
		zap.String("id", id))

	err := h.attachmentService.DeleteAttachment(c.Request.Context(), id)

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to delete attachment via service",
			zap.Error(err),
			zap.String("id", id))
		utils.HandleServiceError(c, "Failed to delete attachment", err)
		return
	}

	// 5. Success
	h.logger.Info("Attachment deleted successfully via handler",
		zap.String("id", id))
	utils.OKResponse(c, "Attachment deleted successfully", nil)
}
