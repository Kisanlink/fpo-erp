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

type RefundPoliciesHandler struct {
	refundPoliciesService interfaces.RefundPoliciesServiceInterface
	aaaMiddleware         *aaa.AAAMiddleware
	logger                logger.Logger
}

func NewRefundPoliciesHandler(refundPoliciesService interfaces.RefundPoliciesServiceInterface, aaaMiddleware *aaa.AAAMiddleware, logger logger.Logger) *RefundPoliciesHandler {
	return &RefundPoliciesHandler{
		refundPoliciesService: refundPoliciesService,
		aaaMiddleware:         aaaMiddleware,
		logger:                logger,
	}
}

// CreateRefundPolicy handles POST /api/v1/refund-policies
// @Summary Create Refund Policy
// @Description Create a new refund policy (requires authentication)
// @Tags Refund Policies
// @Accept json
// @Produce json
// @Param request body models.CreateRefundPolicyRequest true "Refund policy data"
// @Success 201 {object} utils.Response{data=models.RefundPolicyResponse} "Refund policy created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/refund-policies [post]
func (h *RefundPoliciesHandler) CreateRefundPolicy(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling create refund policy request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	var req models.CreateRefundPolicyRequest

	// 2. Validation Error
	if err := utils.ValidateRequest(c, &req); err != nil {
		h.logger.Error("Invalid request body for create refund policy",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 3. Service Call
	h.logger.Debug("Calling service to create refund policy",
		zap.String("policy_name", req.PolicyName))

	policy, err := h.refundPoliciesService.CreateRefundPolicy(&req)

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to create refund policy via service",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to create refund policy", err)
		return
	}

	// 5. Success
	h.logger.Info("Refund policy created successfully via handler",
		zap.String("policy_id", policy.ID))
	utils.CreatedResponse(c, "Refund policy created successfully", policy)
}

// GetAllRefundPolicies handles GET /api/v1/refund-policies
// @Summary Get All Refund Policies
// @Description Retrieve all refund policies with pagination
// @Tags Refund Policies
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.RefundPolicyResponse} "Refund policies retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/refund-policies [get]
func (h *RefundPoliciesHandler) GetAllRefundPolicies(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get all refund policies request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get query parameters for pagination
	limit := 10 // default limit
	offset := 0 // default offset

	// 3. Service Call
	h.logger.Debug("Calling service to get all refund policies",
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	policies, err := h.refundPoliciesService.GetAllRefundPolicies(limit, offset)

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to retrieve refund policies via service",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve refund policies", err)
		return
	}

	// 5. Success
	h.logger.Info("Refund policies retrieved successfully via handler",
		zap.Int("count", len(policies)))
	utils.OKResponse(c, "Refund policies retrieved successfully", policies)
}

// GetRefundPolicy handles GET /api/v1/refund-policies/:id
// @Summary Get Refund Policy
// @Description Retrieve a specific refund policy by ID
// @Tags Refund Policies
// @Produce json
// @Param id path string true "Policy ID" example(RFPOL_12345678)
// @Success 200 {object} utils.Response{data=models.RefundPolicyResponse} "Refund policy retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Refund policy not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/refund-policies/{id} [get]
func (h *RefundPoliciesHandler) GetRefundPolicy(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get refund policy request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	id := c.Param("id")

	// 2. Validation Error
	if id == "" {
		h.logger.Error("Invalid request: policy ID is required")
		utils.BadRequestResponse(c, "Policy ID is required", nil)
		return
	}

	// 3. Service Call
	h.logger.Debug("Calling service to get refund policy",
		zap.String("id", id))

	policy, err := h.refundPoliciesService.GetRefundPolicy(id)

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to retrieve refund policy via service",
			zap.Error(err),
			zap.String("id", id))
		utils.NotFoundResponse(c, "Refund policy not found")
		return
	}

	// 5. Success
	h.logger.Info("Refund policy retrieved successfully via handler",
		zap.String("id", id))
	utils.OKResponse(c, "Refund policy retrieved successfully", policy)
}

// UpdateRefundPolicy handles PATCH /api/v1/refund-policies/:id
// @Summary Update Refund Policy
// @Description Update an existing refund policy by ID
// @Tags Refund Policies
// @Accept json
// @Produce json
// @Param id path string true "Policy ID" example(RFPOL_12345678)
// @Param request body models.UpdateRefundPolicyRequest true "Updated refund policy data"
// @Success 200 {object} utils.Response{data=models.RefundPolicyResponse} "Refund policy updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Refund policy not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/refund-policies/{id} [patch]
func (h *RefundPoliciesHandler) UpdateRefundPolicy(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling update refund policy request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	id := c.Param("id")

	// 2. Validation Error
	if id == "" {
		h.logger.Error("Invalid request: policy ID is required")
		utils.BadRequestResponse(c, "Policy ID is required", nil)
		return
	}

	var req models.UpdateRefundPolicyRequest

	if err := utils.ValidateRequest(c, &req); err != nil {
		h.logger.Error("Invalid request body for update refund policy",
			zap.Error(err),
			zap.String("id", id))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 3. Service Call
	h.logger.Debug("Calling service to update refund policy",
		zap.String("id", id))

	policy, err := h.refundPoliciesService.UpdateRefundPolicy(id, &req)

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to update refund policy via service",
			zap.Error(err),
			zap.String("id", id))
		utils.HandleServiceError(c, "Failed to update refund policy", err)
		return
	}

	// 5. Success
	h.logger.Info("Refund policy updated successfully via handler",
		zap.String("id", id))
	utils.OKResponse(c, "Refund policy updated successfully", policy)
}

// RegisterRoutes registers all refund policies routes
func (h *RefundPoliciesHandler) RegisterRoutes(router *gin.RouterGroup) {
	policies := router.Group("/refund-policies")
	{
		// Apply authentication middleware
		policies.Use(h.aaaMiddleware.Authenticate())

		// Read routes - Director=R, CEO=CRUD, Auditor=R, Accountant=CRUD, Tech_Support=R/W (temp), Store_Manager=–, Store_Staff=–
		policies.GET("", h.aaaMiddleware.RequireOrgPermission("refund_policy", "read"), h.GetAllRefundPolicies)
		policies.GET("/:id", h.aaaMiddleware.RequireOrgPermission("refund_policy", "read"), h.GetRefundPolicy)

		// Protected routes - CEO=CRUD, Accountant=CRUD, Tech_Support=R/W (temp)
		policies.POST("", h.aaaMiddleware.RequireOrgPermission("refund_policy", "create"), h.CreateRefundPolicy)
		policies.PATCH("/:id", h.aaaMiddleware.RequireOrgPermission("refund_policy", "update"), h.UpdateRefundPolicy)
	}
}
