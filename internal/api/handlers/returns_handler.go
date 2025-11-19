package handlers

import (
	"strconv"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	logger "kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ReturnsHandler struct {
	returnsService interfaces.ReturnsServiceInterface
	aaaMiddleware  *aaa.AAAMiddleware
	logger         logger.Logger
}

func NewReturnsHandler(returnsService interfaces.ReturnsServiceInterface, aaaMiddleware *aaa.AAAMiddleware, logger logger.Logger) *ReturnsHandler {
	return &ReturnsHandler{
		returnsService: returnsService,
		aaaMiddleware:  aaaMiddleware,
		logger:         logger,
	}
}

// CreateReturn handles POST /api/v1/returns
// @Summary Create Return
// @Description Create a new return record (requires authentication)
// @Tags Returns
// @Accept json
// @Produce json
// @Param request body models.CreateReturnRequest true "Return data"
// @Success 201 {object} utils.Response{data=models.ReturnResponse} "Return created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/returns [post]
func (h *ReturnsHandler) CreateReturn(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling create return request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	var req models.CreateReturnRequest

	// Validate request
	if err := utils.ValidateRequest(c, &req); err != nil {
		// 2. Validation Error Log
		h.logger.Error("Invalid request body for create return",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to create return",
		zap.String("sale_id", req.SaleID))

	ret, err := h.returnsService.CreateReturn(&req)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error creating return",
			zap.Error(err),
			zap.String("sale_id", req.SaleID))
		utils.HandleServiceError(c, "Failed to create return", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Return created successfully",
		zap.String("return_id", ret.ID),
		zap.String("sale_id", ret.SaleID))

	utils.CreatedResponse(c, "Return created successfully", ret)
}

// GetReturn handles GET /api/v1/returns/:id
// @Summary Get Return
// @Description Retrieve a specific return by ID
// @Tags Returns
// @Produce json
// @Param id path string true "Return ID" example(RET_12345678)
// @Success 200 {object} utils.Response{data=models.ReturnResponse} "Return retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Return not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/returns/{id} [get]
func (h *ReturnsHandler) GetReturn(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get return request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Return ID is required", nil)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to get return",
		zap.String("return_id", id))

	ret, err := h.returnsService.GetReturn(id)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error getting return",
			zap.Error(err),
			zap.String("return_id", id))
		utils.NotFoundResponse(c, "Return not found")
		return
	}

	// 5. Success Log
	h.logger.Info("Return retrieved successfully",
		zap.String("return_id", ret.ID),
		zap.String("sale_id", ret.SaleID))

	utils.OKResponse(c, "Return retrieved successfully", ret)
}

// GetAllReturns handles GET /api/v1/returns
// @Summary Get All Returns
// @Description Retrieve all returns with pagination
// @Tags Returns
// @Produce json
// @Param limit query integer false "Number of records to return (default: 10)" example(10)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.Response{data=[]models.ReturnResponse} "Returns retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/returns [get]
func (h *ReturnsHandler) GetAllReturns(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get all returns request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

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

	// 3. Service Call Log
	h.logger.Debug("Calling service to get all returns",
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	returns, err := h.returnsService.GetAllReturns(limit, offset)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error getting all returns",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve returns", err)
		return
	}

	// 5. Success Log
	h.logger.Info("All returns retrieved successfully",
		zap.Int("count", len(returns)))

	utils.OKResponse(c, "Returns retrieved successfully", returns)
}

// UpdateReturn handles PUT /api/v1/returns/:id
// @Summary Update Return
// @Description Update an existing return by ID
// @Tags Returns
// @Accept json
// @Produce json
// @Param id path string true "Return ID" example(RET_12345678)
// @Param request body models.UpdateReturnRequest true "Updated return data"
// @Success 200 {object} utils.Response{data=models.ReturnResponse} "Return updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Return not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/returns/{id} [put]
func (h *ReturnsHandler) UpdateReturn(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling update return request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Return ID is required", nil)
		return
	}

	var req models.UpdateReturnRequest
	if err := utils.ValidatePartialRequest(c, &req); err != nil {
		// 2. Validation Error Log
		h.logger.Error("Invalid request body for update return",
			zap.Error(err),
			zap.String("return_id", id))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to update return",
		zap.String("return_id", id))

	ret, err := h.returnsService.UpdateReturn(id, &req)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error updating return",
			zap.Error(err),
			zap.String("return_id", id))
		utils.HandleServiceError(c, "Failed to update return", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Return updated successfully",
		zap.String("return_id", ret.ID),
		zap.String("sale_id", ret.SaleID))

	utils.OKResponse(c, "Return updated successfully", ret)
}

// DeleteReturn handles DELETE /api/v1/returns/:id
// @Summary Delete Return
// @Description Delete a return by ID
// @Tags Returns
// @Produce json
// @Param id path string true "Return ID" example(RET_12345678)
// @Success 200 {object} utils.Response "Return deleted successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Return not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/returns/{id} [delete]
func (h *ReturnsHandler) DeleteReturn(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling delete return request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Return ID is required", nil)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to delete return",
		zap.String("return_id", id))

	err := h.returnsService.DeleteReturn(id)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error deleting return",
			zap.Error(err),
			zap.String("return_id", id))
		utils.HandleServiceError(c, "Failed to delete return", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Return deleted successfully",
		zap.String("return_id", id))

	utils.OKResponse(c, "Return deleted successfully", nil)
}

// GetReturnsByCustomer handles GET /api/v1/returns/customer/:customerID
// @Summary Get Returns by Customer
// @Description Retrieve all returns for a specific customer
// @Tags Returns
// @Produce json
// @Param customerID path string true "Customer ID" example(CUST_12345678)
// @Success 200 {object} utils.Response{data=[]models.ReturnResponse} "Returns retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/returns/customer/{customerID} [get]
func (h *ReturnsHandler) GetReturnsByCustomer(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get returns by customer request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	customerID := c.Param("customerID")
	if customerID == "" {
		utils.BadRequestResponse(c, "Customer ID is required", nil)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to get returns by customer",
		zap.String("customer_id", customerID))

	returns, err := h.returnsService.GetReturnsByCustomer(customerID)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error getting returns by customer",
			zap.Error(err),
			zap.String("customer_id", customerID))
		utils.HandleServiceError(c, "Failed to retrieve returns", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Returns by customer retrieved successfully",
		zap.String("customer_id", customerID),
		zap.Int("count", len(returns)))

	utils.OKResponse(c, "Returns retrieved successfully", returns)
}

// GetReturnsBySaleID handles GET /api/v1/returns/sale/:saleID
// @Summary Get Returns by Sale ID
// @Description Retrieve all returns for a specific sale
// @Tags Returns
// @Produce json
// @Param saleID path string true "Sale ID" example(SALE_12345678)
// @Success 200 {object} utils.Response{data=[]models.ReturnResponse} "Returns retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/returns/sale/{saleID} [get]
func (h *ReturnsHandler) GetReturnsBySaleID(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get returns by sale ID request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	saleID := c.Param("saleID")
	if saleID == "" {
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to get returns by sale",
		zap.String("sale_id", saleID))

	returns, err := h.returnsService.GetReturnsBySaleID(saleID)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error getting returns by sale",
			zap.Error(err),
			zap.String("sale_id", saleID))
		utils.HandleServiceError(c, "Failed to retrieve returns", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Returns by sale retrieved successfully",
		zap.String("sale_id", saleID),
		zap.Int("count", len(returns)))

	utils.OKResponse(c, "Returns retrieved successfully", returns)
}

// GetReturnsByDateRange handles GET /api/v1/returns/date-range
// @Summary Get Returns by Date Range
// @Description Retrieve returns within a specific date range
// @Tags Returns
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)" example(2024-01-01)
// @Param end_date query string true "End date (YYYY-MM-DD)" example(2024-12-31)
// @Success 200 {object} utils.Response{data=[]models.ReturnResponse} "Returns retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/returns/date-range [get]
func (h *ReturnsHandler) GetReturnsByDateRange(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get returns by date range request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		utils.BadRequestResponse(c, "Start date and end date are required", nil)
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid start date format", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid end date format", err)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to get returns by date range",
		zap.String("start_date", startDateStr),
		zap.String("end_date", endDateStr))

	returns, err := h.returnsService.GetReturnsByDateRange(startDate, endDate)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error getting returns by date range",
			zap.Error(err),
			zap.String("start_date", startDateStr),
			zap.String("end_date", endDateStr))
		utils.HandleServiceError(c, "Failed to retrieve returns", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Returns by date range retrieved successfully",
		zap.String("start_date", startDateStr),
		zap.String("end_date", endDateStr),
		zap.Int("count", len(returns)))

	utils.OKResponse(c, "Returns retrieved successfully", returns)
}

// GetReturnsByStatus handles GET /api/v1/returns/status/:status
// @Summary Get Returns by Status
// @Description Retrieve all returns with a specific status
// @Tags Returns
// @Produce json
// @Param status path string true "Return status" example(processed)
// @Success 200 {object} utils.Response{data=[]models.ReturnResponse} "Returns retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/returns/status/{status} [get]
func (h *ReturnsHandler) GetReturnsByStatus(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get returns by status request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	status := c.Param("status")
	if status == "" {
		utils.BadRequestResponse(c, "Status is required", nil)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to get returns by status",
		zap.String("status", status))

	returns, err := h.returnsService.GetReturnsByStatus(status)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error getting returns by status",
			zap.Error(err),
			zap.String("status", status))
		utils.HandleServiceError(c, "Failed to retrieve returns", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Returns by status retrieved successfully",
		zap.String("status", status),
		zap.Int("count", len(returns)))

	utils.OKResponse(c, "Returns retrieved successfully", returns)
}

// GetTotalReturnsAmount handles GET /api/v1/returns/total-amount
// @Summary Get Total Returns Amount
// @Description Calculate total returns amount within a date range
// @Tags Returns
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)" example(2024-01-01)
// @Param end_date query string true "End date (YYYY-MM-DD)" example(2024-12-31)
// @Success 200 {object} utils.Response{data=object{total_amount=number,start_date=string,end_date=string}} "Total returns amount calculated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/returns/total-amount [get]
func (h *ReturnsHandler) GetTotalReturnsAmount(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get total returns amount request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		utils.BadRequestResponse(c, "Start date and end date are required", nil)
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid start date format", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid end date format", err)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to get total returns amount",
		zap.String("start_date", startDateStr),
		zap.String("end_date", endDateStr))

	totalAmount, err := h.returnsService.GetTotalReturnsAmount(startDate, endDate)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error calculating total returns amount",
			zap.Error(err),
			zap.String("start_date", startDateStr),
			zap.String("end_date", endDateStr))
		utils.HandleServiceError(c, "Failed to calculate total amount", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Total returns amount calculated successfully",
		zap.String("start_date", startDateStr),
		zap.String("end_date", endDateStr),
		zap.Float64("total_amount", totalAmount))

	utils.OKResponse(c, "Total returns amount calculated successfully", gin.H{
		"total_amount": totalAmount,
		"start_date":   startDateStr,
		"end_date":     endDateStr,
	})
}

// GetReturnRateByProduct handles GET /api/v1/returns/return-rate/:productID
// @Summary Get Return Rate by Product
// @Description Calculate return rate for a specific product within a date range
// @Tags Returns
// @Produce json
// @Param productID path string true "Product ID" example(PROD_12345678)
// @Param start_date query string true "Start date (YYYY-MM-DD)" example(2024-01-01)
// @Param end_date query string true "End date (YYYY-MM-DD)" example(2024-12-31)
// @Success 200 {object} utils.Response{data=object{product_id=string,return_rate=number,start_date=string,end_date=string}} "Return rate calculated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/returns/return-rate/{productID} [get]
func (h *ReturnsHandler) GetReturnRateByProduct(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get return rate by product request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	productID := c.Param("productID")
	if productID == "" {
		utils.BadRequestResponse(c, "Product ID is required", nil)
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		utils.BadRequestResponse(c, "Start date and end date are required", nil)
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid start date format", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid end date format", err)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to get return rate by product",
		zap.String("product_id", productID),
		zap.String("start_date", startDateStr),
		zap.String("end_date", endDateStr))

	returnRate, err := h.returnsService.GetReturnRateByProduct(productID, startDate, endDate)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error calculating return rate by product",
			zap.Error(err),
			zap.String("product_id", productID))
		utils.HandleServiceError(c, "Failed to calculate return rate", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Return rate by product calculated successfully",
		zap.String("product_id", productID),
		zap.Float64("return_rate", returnRate))

	utils.OKResponse(c, "Return rate calculated successfully", gin.H{
		"product_id":  productID,
		"return_rate": returnRate,
		"start_date":  startDateStr,
		"end_date":    endDateStr,
	})
}

// GetMostReturnedProducts handles GET /api/v1/returns/most-returned
// @Summary Get Most Returned Products
// @Description Retrieve the most frequently returned products
// @Tags Returns
// @Produce json
// @Param limit query integer false "Number of products to return (default: 10)" example(10)
// @Success 200 {object} utils.Response{data=[]models.MostReturnedProductResponse} "Most returned products retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/returns/most-returned [get]
func (h *ReturnsHandler) GetMostReturnedProducts(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling get most returned products request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid limit parameter", err)
		return
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to get most returned products",
		zap.Int("limit", limit))

	products, err := h.returnsService.GetMostReturnedProducts(limit)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error getting most returned products",
			zap.Error(err),
			zap.Int("limit", limit))
		utils.HandleServiceError(c, "Failed to retrieve most returned products", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Most returned products retrieved successfully",
		zap.Int("count", len(products)))

	utils.OKResponse(c, "Most returned products retrieved successfully", products)
}

// UpdateReturnStatus handles PATCH /api/v1/returns/:id/status
// @Summary Update Return Status
// @Description Update the status of a specific return
// @Tags Returns
// @Accept json
// @Produce json
// @Param id path string true "Return ID" example(RET_12345678)
// @Param request body models.UpdateReturnStatusRequest true "New status"
// @Success 200 {object} utils.Response{data=models.ReturnResponse} "Return status updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Return not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/returns/{id}/status [patch]
func (h *ReturnsHandler) UpdateReturnStatus(c *gin.Context) {
	// 1. Entry Log
	h.logger.Info("Handling update return status request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Return ID is required", nil)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := utils.ValidateRequest(c, &req); err != nil {
		// 2. Validation Error Log
		h.logger.Error("Invalid request body for update return status",
			zap.Error(err),
			zap.String("return_id", id))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	updateReq := models.UpdateReturnRequest{
		Status: &req.Status,
	}

	// 3. Service Call Log
	h.logger.Debug("Calling service to update return status",
		zap.String("return_id", id),
		zap.String("new_status", req.Status))

	ret, err := h.returnsService.UpdateReturn(id, &updateReq)
	if err != nil {
		// 4. Service Error Log
		h.logger.Error("Service error updating return status",
			zap.Error(err),
			zap.String("return_id", id),
			zap.String("new_status", req.Status))
		utils.HandleServiceError(c, "Failed to update return status", err)
		return
	}

	// 5. Success Log
	h.logger.Info("Return status updated successfully",
		zap.String("return_id", ret.ID),
		zap.String("new_status", ret.Status))

	utils.OKResponse(c, "Return status updated successfully", ret)
}

// RegisterRoutes registers all returns routes
func (h *ReturnsHandler) RegisterRoutes(router *gin.RouterGroup) {
	returns := router.Group("/returns")
	{
		// Apply authentication middleware
		returns.Use(h.aaaMiddleware.Authenticate())

		// Create/Update/Delete routes - CEO=CRUD, Store_Staff=CRUD, Tech_Support=R/W (temp)
		returns.POST("", h.aaaMiddleware.RequireOrgPermission("return", "create"), h.CreateReturn)
		returns.PUT("/:id", h.aaaMiddleware.RequireOrgPermission("return", "update"), h.UpdateReturn)
		returns.PATCH("/:id/status", h.aaaMiddleware.RequireOrgPermission("return", "update"), h.UpdateReturnStatus)
		returns.DELETE("/:id", h.aaaMiddleware.RequireOrgPermission("return", "delete"), h.DeleteReturn)

		// Read routes - Director=R, CEO=CRUD, Auditor=R, Accountant=R, Tech_Support=R/W (temp), Store_Manager=R, Store_Staff=CRUD
		returns.GET("", h.aaaMiddleware.RequireOrgPermission("return", "read"), h.GetAllReturns)
		returns.GET("/:id", h.aaaMiddleware.RequireOrgPermission("return", "read"), h.GetReturn)
		returns.GET("/customer/:customerID", h.aaaMiddleware.RequireOrgPermission("return", "read"), h.GetReturnsByCustomer)
		returns.GET("/sale/:saleID", h.aaaMiddleware.RequireOrgPermission("return", "read"), h.GetReturnsBySaleID)
		returns.GET("/date-range", h.aaaMiddleware.RequireOrgPermission("return", "read"), h.GetReturnsByDateRange)
		returns.GET("/status/:status", h.aaaMiddleware.RequireOrgPermission("return", "read"), h.GetReturnsByStatus)
		returns.GET("/total-amount", h.aaaMiddleware.RequireOrgPermission("return", "read"), h.GetTotalReturnsAmount)
		returns.GET("/return-rate/:productID", h.aaaMiddleware.RequireOrgPermission("return", "read"), h.GetReturnRateByProduct)
		returns.GET("/most-returned", h.aaaMiddleware.RequireOrgPermission("return", "read"), h.GetMostReturnedProducts)
	}
}
