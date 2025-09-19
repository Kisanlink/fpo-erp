package handlers

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

// AttachmentHandler handles attachment HTTP requests
type AttachmentHandler struct {
	attachmentService *services.AttachmentService
	aaaMiddleware     *aaa.AAAMiddleware
}

// NewAttachmentHandler creates a new attachment handler
func NewAttachmentHandler(attachmentService *services.AttachmentService, aaaMiddleware *aaa.AAAMiddleware) *AttachmentHandler {
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

		// Create/Delete routes - CEO=CRUD, Tech_Support=R/W (temp)
		attachments.POST("", h.aaaMiddleware.RequirePermission("attachment", "*", "create"), h.UploadAttachment)
		attachments.DELETE("/:id", h.aaaMiddleware.RequirePermission("attachment", "*", "delete"), h.DeleteAttachment)

		// Read routes - Director=R, CEO=CRUD, Auditor=R, Accountant=R, Tech_Support=R/W (temp), Store_Manager=R, Store_Staff=R
		attachments.GET("/:id/download", h.aaaMiddleware.RequirePermission("attachment", "*", "read"), h.DownloadAttachment)
		attachments.GET("/:id/url", h.aaaMiddleware.RequirePermission("attachment", "*", "read"), h.GenerateDownloadURL)
		attachments.GET("", h.aaaMiddleware.RequirePermission("attachment", "*", "read"), h.GetAttachments)
		attachments.GET("/:id", h.aaaMiddleware.RequirePermission("attachment", "*", "read"), h.GetAttachment)
		attachments.GET("/:id/info", h.aaaMiddleware.RequirePermission("attachment", "*", "read"), h.GetAttachmentInfo)
		attachments.GET("/sale/:sale_id", h.aaaMiddleware.RequirePermission("attachment", "*", "read"), h.GetAttachmentsBySale)
		attachments.GET("/return/:return_id", h.aaaMiddleware.RequirePermission("attachment", "*", "read"), h.GetAttachmentsByReturn)
	}
}

// UploadAttachment handles file upload
// @Summary Upload Attachment
// @Description Upload a file attachment for sales or returns (requires authentication)
// @Tags Attachments
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param sale_id formData string false "Sale ID (optional)" example(SALE_12345678)
// @Param return_id formData string false "Return ID (optional)" example(RET_12345678)
// @Success 201 {object} utils.Response{data=models.AttachmentResponse} "Attachment uploaded successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
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

	// No need to parse userID, just use userIDStr.(string) directly
	// No need to check for err here since userIDStr is already obtained from context

	// Get form data
	saleID := c.PostForm("sale_id")
	returnID := c.PostForm("return_id")

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequestResponse(c, "No file uploaded", err)
		return
	}

	// Convert empty strings to nil
	var saleIDPtr, returnIDPtr *string
	if saleID != "" {
		saleIDPtr = &saleID
	}
	if returnID != "" {
		returnIDPtr = &returnID
	}

	// Upload attachment
	attachment, err := h.attachmentService.UploadAttachment(c.Request.Context(), file, saleIDPtr, returnIDPtr, userIDStr.(string))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to upload attachment", err)
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
// @Failure 404 {object} utils.ErrorResponseModel "Attachment not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/attachments/{id} [get]
func (h *AttachmentHandler) GetAttachment(c *gin.Context) {
	id := c.Param("id")

	attachment, err := h.attachmentService.GetAttachment(id)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve attachment", err)
		return
	}

	utils.OKResponse(c, "Attachment retrieved successfully", attachment)
}

// GetAttachments retrieves attachments with optional filters
// @Summary Get Attachments
// @Description Retrieve attachments with optional filters and pagination
// @Tags Attachments
// @Produce json
// @Param sale_id query string false "Filter by Sale ID" example(SALE_12345678)
// @Param return_id query string false "Filter by Return ID" example(RET_12345678)
// @Param limit query integer false "Number of records to return (default: 10, max: 100)" example(10)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.Response{data=[]models.AttachmentResponse} "Attachments retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/attachments [get]
func (h *AttachmentHandler) GetAttachments(c *gin.Context) {
	// Get query parameters
	saleID := c.Query("sale_id")
	returnID := c.Query("return_id")
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
	var saleIDPtr, returnIDPtr *string
	if saleID != "" {
		saleIDPtr = &saleID
	}
	if returnID != "" {
		returnIDPtr = &returnID
	}

	// Get attachments
	attachments, err := h.attachmentService.GetAttachments(saleIDPtr, returnIDPtr, limit, offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve attachments", err)
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
// @Failure 404 {object} utils.ErrorResponseModel "Attachment not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/attachments/{id}/download [get]
func (h *AttachmentHandler) DownloadAttachment(c *gin.Context) {
	id := c.Param("id")

	// Download file
	fileReader, contentType, err := h.attachmentService.DownloadAttachment(c.Request.Context(), id)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to download attachment", err)
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
	c.DataFromReader(http.StatusOK, -1, contentType, fileReader.(io.Reader), nil)
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
// @Failure 404 {object} utils.ErrorResponseModel "Attachment not found"
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
		utils.InternalServerErrorResponse(c, "Failed to generate download URL", err)
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
// @Failure 404 {object} utils.ErrorResponseModel "Attachment not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/attachments/{id}/info [get]
func (h *AttachmentHandler) GetAttachmentInfo(c *gin.Context) {
	id := c.Param("id")

	info, err := h.attachmentService.GetAttachmentInfo(c.Request.Context(), id)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve attachment info", err)
		return
	}

	utils.OKResponse(c, "Attachment info retrieved successfully", info)
}

// GetAttachmentsBySale retrieves all attachments for a sale
// @Summary Get Attachments by Sale
// @Description Retrieve all attachments associated with a specific sale
// @Tags Attachments
// @Produce json
// @Param sale_id path string true "Sale ID" example(SALE_12345678)
// @Success 200 {object} utils.Response{data=[]models.AttachmentResponse} "Sale attachments retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/attachments/sale/{sale_id} [get]
func (h *AttachmentHandler) GetAttachmentsBySale(c *gin.Context) {
	saleID := c.Param("sale_id")

	attachments, err := h.attachmentService.GetAttachmentsBySale(saleID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve sale attachments", err)
		return
	}

	utils.OKResponse(c, "Sale attachments retrieved successfully", attachments)
}

// GetAttachmentsByReturn retrieves all attachments for a return
// @Summary Get Attachments by Return
// @Description Retrieve all attachments associated with a specific return
// @Tags Attachments
// @Produce json
// @Param return_id path string true "Return ID" example(RET_12345678)
// @Success 200 {object} utils.Response{data=[]models.AttachmentResponse} "Return attachments retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/attachments/return/{return_id} [get]
func (h *AttachmentHandler) GetAttachmentsByReturn(c *gin.Context) {
	returnID := c.Param("return_id")

	attachments, err := h.attachmentService.GetAttachmentsByReturn(returnID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve return attachments", err)
		return
	}

	utils.OKResponse(c, "Return attachments retrieved successfully", attachments)
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
// @Failure 404 {object} utils.ErrorResponseModel "Attachment not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/attachments/{id} [delete]
func (h *AttachmentHandler) DeleteAttachment(c *gin.Context) {
	id := c.Param("id")

	err := h.attachmentService.DeleteAttachment(c.Request.Context(), id)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete attachment", err)
		return
	}

	utils.OKResponse(c, "Attachment deleted successfully", nil)
}
