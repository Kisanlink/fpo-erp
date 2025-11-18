package handlers

import (
	"strconv"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

type BankPaymentsHandler struct {
	bankPaymentsService interfaces.BankPaymentsServiceInterface
	aaaMiddleware       *aaa.AAAMiddleware
}

func NewBankPaymentsHandler(bankPaymentsService interfaces.BankPaymentsServiceInterface, aaaMiddleware *aaa.AAAMiddleware) *BankPaymentsHandler {
	return &BankPaymentsHandler{
		bankPaymentsService: bankPaymentsService,
		aaaMiddleware:       aaaMiddleware,
	}
}

// CreateBankPayment handles POST /api/v1/bank-payments
// @Summary Create Bank Payment
// @Description Create a new bank payment record (requires authentication)
// @Tags Bank Payments
// @Accept json
// @Produce json
// @Param request body models.CreateBankPaymentRequest true "Bank payment data"
// @Success 201 {object} utils.Response{data=models.BankPaymentResponse} "Bank payment created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/bank-payments [post]
func (h *BankPaymentsHandler) CreateBankPayment(c *gin.Context) {
	var req models.CreateBankPaymentRequest

	// Validate request
	if err := utils.ValidateRequest(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	payment, err := h.bankPaymentsService.CreateBankPayment(&req)
	if err != nil {
		utils.HandleServiceError(c, "Failed to create bank payment", err)
		return
	}

	utils.CreatedResponse(c, "Bank payment created successfully", payment)
}

// GetAllBankPayments handles GET /api/v1/bank-payments
// @Summary Get All Bank Payments
// @Description Retrieve all bank payments with pagination
// @Tags Bank Payments
// @Produce json
// @Param limit query integer false "Number of records to return (default: 10)" example(10)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.Response{data=[]models.BankPaymentResponse} "Bank payments retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/bank-payments [get]
func (h *BankPaymentsHandler) GetAllBankPayments(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid limit parameter", err)
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid offset parameter", err)
		return
	}

	payments, err := h.bankPaymentsService.GetAllBankPayments(limit, offset)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve bank payments", err)
		return
	}

	utils.OKResponse(c, "Bank payments retrieved successfully", payments)
}

// GetBankPayment handles GET /api/v1/bank-payments/:id
// @Summary Get Bank Payment
// @Description Retrieve a specific bank payment by ID
// @Tags Bank Payments
// @Produce json
// @Param id path string true "Payment ID" example(PAY_12345678)
// @Success 200 {object} utils.Response{data=models.BankPaymentResponse} "Bank payment retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Bank payment not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/bank-payments/{id} [get]
func (h *BankPaymentsHandler) GetBankPayment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Payment ID is required", nil)
		return
	}

	payment, err := h.bankPaymentsService.GetBankPayment(id)
	if err != nil {
		utils.NotFoundResponse(c, "Bank payment not found")
		return
	}

	utils.OKResponse(c, "Bank payment retrieved successfully", payment)
}

// GetBankPaymentsBySale handles GET /api/v1/bank-payments/sale/:saleID
// @Summary Get Bank Payments by Sale
// @Description Retrieve all bank payments for a specific sale
// @Tags Bank Payments
// @Produce json
// @Param saleID path string true "Sale ID" example(SALE_12345678)
// @Success 200 {object} utils.Response{data=[]models.BankPaymentResponse} "Bank payments for sale retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/bank-payments/sale/{saleID} [get]
func (h *BankPaymentsHandler) GetBankPaymentsBySale(c *gin.Context) {
	saleID := c.Param("saleID")
	if saleID == "" {
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	payments, err := h.bankPaymentsService.GetBankPaymentsBySaleID(saleID)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve bank payments for sale", err)
		return
	}

	utils.OKResponse(c, "Bank payments for sale retrieved successfully", payments)
}

// GetBankPaymentsByReturn handles GET /api/v1/bank-payments/return/:returnID
// @Summary Get Bank Payments by Return
// @Description Retrieve all bank payments for a specific return
// @Tags Bank Payments
// @Produce json
// @Param returnID path string true "Return ID" example(RET_12345678)
// @Success 200 {object} utils.Response{data=[]models.BankPaymentResponse} "Bank payments for return retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/bank-payments/return/{returnID} [get]
func (h *BankPaymentsHandler) GetBankPaymentsByReturn(c *gin.Context) {
	returnID := c.Param("returnID")
	if returnID == "" {
		utils.BadRequestResponse(c, "Return ID is required", nil)
		return
	}

	payments, err := h.bankPaymentsService.GetBankPaymentsByReturnID(returnID)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve bank payments for return", err)
		return
	}

	utils.OKResponse(c, "Bank payments for return retrieved successfully", payments)
}

// RegisterRoutes registers all bank payments routes
func (h *BankPaymentsHandler) RegisterRoutes(router *gin.RouterGroup) {
	payments := router.Group("/bank-payments")
	{
		// Apply authentication middleware
		payments.Use(h.aaaMiddleware.Authenticate())

		// Read routes - Director=R, CEO=CRUD, Auditor=R, Accountant=CRUD, Tech_Support=R/W (temp), Store_Manager=–, Store_Staff=– (organization-scoped)
		payments.GET("", h.aaaMiddleware.RequireOrgPermission("bank_payment", "read"), h.GetAllBankPayments)
		payments.GET("/:id", h.aaaMiddleware.RequireOrgPermission("bank_payment", "read"), h.GetBankPayment)
		payments.GET("/sale/:saleID", h.aaaMiddleware.RequireOrgPermission("bank_payment", "read"), h.GetBankPaymentsBySale)
		payments.GET("/return/:returnID", h.aaaMiddleware.RequireOrgPermission("bank_payment", "read"), h.GetBankPaymentsByReturn)

		// Protected routes - CEO=CRUD, Accountant=CRUD, Tech_Support=R/W (temp) (organization-scoped)
		payments.POST("", h.aaaMiddleware.RequireOrgPermission("bank_payment", "create"), h.CreateBankPayment)
	}
}
