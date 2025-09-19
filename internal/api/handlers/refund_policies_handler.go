package handlers

import (
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

type RefundPoliciesHandler struct {
	refundPoliciesService *services.RefundPoliciesService
	aaaMiddleware         *aaa.AAAMiddleware
}

func NewRefundPoliciesHandler(refundPoliciesService *services.RefundPoliciesService, aaaMiddleware *aaa.AAAMiddleware) *RefundPoliciesHandler {
	return &RefundPoliciesHandler{
		refundPoliciesService: refundPoliciesService,
		aaaMiddleware:         aaaMiddleware,
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
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/refund-policies [post]
func (h *RefundPoliciesHandler) CreateRefundPolicy(c *gin.Context) {
	var req models.CreateRefundPolicyRequest

	// Validate request
	if err := utils.ValidateRequest(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	policy, err := h.refundPoliciesService.CreateRefundPolicy(&req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create refund policy", err)
		return
	}

	utils.CreatedResponse(c, "Refund policy created successfully", policy)
}

// GetAllRefundPolicies handles GET /api/v1/refund-policies
// @Summary Get All Refund Policies
// @Description Retrieve all refund policies with pagination
// @Tags Refund Policies
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.RefundPolicyResponse} "Refund policies retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/refund-policies [get]
func (h *RefundPoliciesHandler) GetAllRefundPolicies(c *gin.Context) {
	// Get query parameters for pagination
	limit := 10 // default limit
	offset := 0 // default offset

	policies, err := h.refundPoliciesService.GetAllRefundPolicies(limit, offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve refund policies", err)
		return
	}

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
// @Failure 404 {object} utils.ErrorResponseModel "Refund policy not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/refund-policies/{id} [get]
func (h *RefundPoliciesHandler) GetRefundPolicy(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Policy ID is required", nil)
		return
	}

	policy, err := h.refundPoliciesService.GetRefundPolicy(id)
	if err != nil {
		utils.NotFoundResponse(c, "Refund policy not found")
		return
	}

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
// @Failure 404 {object} utils.ErrorResponseModel "Refund policy not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/refund-policies/{id} [patch]
func (h *RefundPoliciesHandler) UpdateRefundPolicy(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Policy ID is required", nil)
		return
	}

	var req models.UpdateRefundPolicyRequest

	// Validate request
	if err := utils.ValidateRequest(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	policy, err := h.refundPoliciesService.UpdateRefundPolicy(id, &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update refund policy", err)
		return
	}

	utils.OKResponse(c, "Refund policy updated successfully", policy)
}

// RegisterRoutes registers all refund policies routes
func (h *RefundPoliciesHandler) RegisterRoutes(router *gin.RouterGroup) {
	policies := router.Group("/refund-policies")
	{
		// Apply authentication middleware
		policies.Use(h.aaaMiddleware.Authenticate())

		// Read routes - Director=R, CEO=CRUD, Auditor=R, Accountant=CRUD, Tech_Support=R/W (temp), Store_Manager=–, Store_Staff=–
		policies.GET("", h.aaaMiddleware.RequirePermission("refund_policy", "*", "read"), h.GetAllRefundPolicies)
		policies.GET("/:id", h.aaaMiddleware.RequirePermission("refund_policy", "*", "read"), h.GetRefundPolicy)

		// Protected routes - CEO=CRUD, Accountant=CRUD, Tech_Support=R/W (temp)
		policies.POST("", h.aaaMiddleware.RequirePermission("refund_policy", "*", "create"), h.CreateRefundPolicy)
		policies.PATCH("/:id", h.aaaMiddleware.RequirePermission("refund_policy", "*", "update"), h.UpdateRefundPolicy)
	}
}

