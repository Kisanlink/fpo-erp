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

type BankPaymentsHandler struct {
	bankPaymentsService interfaces.BankPaymentsServiceInterface
	aaaMiddleware       *aaa.AAAMiddleware
	logger              logger.Logger
}

func NewBankPaymentsHandler(bankPaymentsService interfaces.BankPaymentsServiceInterface, aaaMiddleware *aaa.AAAMiddleware, logger logger.Logger) *BankPaymentsHandler {
	return &BankPaymentsHandler{
		bankPaymentsService: bankPaymentsService,
		aaaMiddleware:       aaaMiddleware,
		logger:              logger,
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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/bank-payments [post]
func (h *BankPaymentsHandler) CreateBankPayment(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling create bank payment request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	var req models.CreateBankPaymentRequest

	// 2. Validation Error
	if err := utils.ValidateRequest(c, &req); err != nil {
		h.logger.Error("Invalid request body for create bank payment",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 3. Service Call
	h.logger.Debug("Calling service to create bank payment",
		zap.Float64("amount", req.Amount))

	payment, err := h.bankPaymentsService.CreateBankPayment(&req)

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to create bank payment via service",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to create bank payment", err)
		return
	}

	// 5. Success
	h.logger.Info("Bank payment created successfully via handler",
		zap.String("payment_id", payment.ID))
	utils.CreatedResponse(c, "Bank payment created successfully", payment)
}

// GetAllBankPayments handles GET /api/v1/bank-payments
// @Summary Get All Bank Payments
// @Description Retrieve all bank payments with pagination
// @Tags Bank Payments
// @Produce json
// @Param limit query integer false "Number of records to return (default: 50, max: 200)" example(50)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.PaginatedResponseModel{data=[]models.BankPaymentResponse} "Bank payments retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/bank-payments [get]
func (h *BankPaymentsHandler) GetAllBankPayments(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get all bank payments request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// 3. Service Call
	h.logger.Debug("Calling service to get all bank payments",
		zap.Int("limit", params.Limit),
		zap.Int("offset", params.Offset))

	payments, total, err := h.bankPaymentsService.GetAllBankPayments(params.Limit, params.Offset)

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to retrieve bank payments via service",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve bank payments", err)
		return
	}

	// 5. Success
	h.logger.Info("Bank payments retrieved successfully via handler",
		zap.Int("count", len(payments)),
		zap.Int64("total", total))
	utils.PaginatedOKResponse(c, payments, total, params.Limit, params.Offset)
}

// GetBankPayment handles GET /api/v1/bank-payments/:id
// @Summary Get Bank Payment
// @Description Retrieve a specific bank payment by ID
// @Tags Bank Payments
// @Produce json
// @Param id path string true "Payment ID" example(PAY_12345678)
// @Success 200 {object} utils.Response{data=models.BankPaymentResponse} "Bank payment retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Bank payment not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/bank-payments/{id} [get]
func (h *BankPaymentsHandler) GetBankPayment(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get bank payment request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	id := c.Param("id")

	// 2. Validation Error
	if id == "" {
		h.logger.Error("Invalid request: payment ID is required")
		utils.BadRequestResponse(c, "Payment ID is required", nil)
		return
	}

	// 3. Service Call
	h.logger.Debug("Calling service to get bank payment",
		zap.String("id", id))

	payment, err := h.bankPaymentsService.GetBankPayment(id)

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to retrieve bank payment via service",
			zap.Error(err),
			zap.String("id", id))
		utils.NotFoundResponse(c, "Bank payment not found")
		return
	}

	// 5. Success
	h.logger.Info("Bank payment retrieved successfully via handler",
		zap.String("id", id))
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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/bank-payments/sale/{saleID} [get]
func (h *BankPaymentsHandler) GetBankPaymentsBySale(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get bank payments by sale request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	saleID := c.Param("saleID")

	// 2. Validation Error
	if saleID == "" {
		h.logger.Error("Invalid request: sale ID is required")
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	// 3. Service Call
	h.logger.Debug("Calling service to get bank payments by sale",
		zap.String("sale_id", saleID))

	payments, err := h.bankPaymentsService.GetBankPaymentsBySaleID(saleID)

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to retrieve bank payments for sale via service",
			zap.Error(err),
			zap.String("sale_id", saleID))
		utils.HandleServiceError(c, "Failed to retrieve bank payments for sale", err)
		return
	}

	// 5. Success
	h.logger.Info("Bank payments for sale retrieved successfully via handler",
		zap.String("sale_id", saleID),
		zap.Int("count", len(payments)))
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
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/bank-payments/return/{returnID} [get]
func (h *BankPaymentsHandler) GetBankPaymentsByReturn(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get bank payments by return request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	returnID := c.Param("returnID")

	// 2. Validation Error
	if returnID == "" {
		h.logger.Error("Invalid request: return ID is required")
		utils.BadRequestResponse(c, "Return ID is required", nil)
		return
	}

	// 3. Service Call
	h.logger.Debug("Calling service to get bank payments by return",
		zap.String("return_id", returnID))

	payments, err := h.bankPaymentsService.GetBankPaymentsByReturnID(returnID)

	// 4. Service Error
	if err != nil {
		h.logger.Error("Failed to retrieve bank payments for return via service",
			zap.Error(err),
			zap.String("return_id", returnID))
		utils.HandleServiceError(c, "Failed to retrieve bank payments for return", err)
		return
	}

	// 5. Success
	h.logger.Info("Bank payments for return retrieved successfully via handler",
		zap.String("return_id", returnID),
		zap.Int("count", len(payments)))
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
