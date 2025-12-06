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

type SalesHandler struct {
	salesService  interfaces.SalesServiceInterface
	aaaMiddleware *aaa.AAAMiddleware
	logger        logger.Logger
}

func NewSalesHandler(salesService interfaces.SalesServiceInterface, aaaMiddleware *aaa.AAAMiddleware, logger logger.Logger) *SalesHandler {
	return &SalesHandler{
		salesService:  salesService,
		aaaMiddleware: aaaMiddleware,
		logger:        logger,
	}
}

// CreateSale handles POST /api/v1/sales
// @Summary Create Sale
// @Description Create a new sale (requires authentication)
// @Tags Sales
// @Accept json
// @Produce json
// @Param request body models.CreateSaleRequest true "Sale data"
// @Success 201 {object} utils.Response{data=models.SaleResponse} "Sale created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales [post]
func (h *SalesHandler) CreateSale(c *gin.Context) {
	h.logger.Info("Handling create sale request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	var req models.CreateSaleRequest

	// Validate request
	if err := utils.ValidateRequest(c, &req); err != nil {
		h.logger.Error("Invalid request body for create sale",
			zap.Error(err))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Log BRD-specific fields
	h.logger.Debug("Calling sales service to create sale",
		zap.String("warehouse_id", req.WarehouseID),
		zap.String("payment_mode", req.PaymentMode),
		zap.String("sale_type", req.SaleType),
		zap.Bool("apply_taxes", req.ApplyTaxes != nil && *req.ApplyTaxes),
		zap.Int("items_count", len(req.Items)))

	// Log customer_id if present
	if req.CustomerID != nil {
		h.logger.Debug("Sale includes customer tracking",
			zap.String("customer_id", *req.CustomerID))
	}

	sale, err := h.salesService.CreateSale(&req)
	if err != nil {
		h.logger.Error("Failed to create sale via service",
			zap.Error(err),
			zap.String("warehouse_id", req.WarehouseID),
			zap.String("payment_mode", req.PaymentMode))
		utils.HandleServiceError(c, "Failed to create sale", err)
		return
	}

	h.logger.Info("Sale created successfully via handler",
		zap.String("sale_id", sale.ID),
		zap.String("warehouse_id", req.WarehouseID),
		zap.Float64("total_amount", sale.TotalAmount),
		zap.Int("items_count", len(sale.Items)))

	utils.CreatedResponse(c, "Sale created successfully", sale)
}

// GetSale handles GET /api/v1/sales/:id
// @Summary Get Sale
// @Description Retrieve a specific sale by ID
// @Tags Sales
// @Produce json
// @Param id path string true "Sale ID" example(SALE_12345678)
// @Success 200 {object} utils.Response{data=models.SaleResponse} "Sale retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Sale not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/{id} [get]
func (h *SalesHandler) GetSale(c *gin.Context) {
	h.logger.Info("Handling get sale request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	id := c.Param("id")
	if id == "" {
		h.logger.Error("Sale ID is required but not provided")
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	h.logger.Debug("Calling sales service to get sale",
		zap.String("sale_id", id))

	sale, err := h.salesService.GetSale(id)
	if err != nil {
		h.logger.Error("Sale not found",
			zap.Error(err),
			zap.String("sale_id", id))
		utils.NotFoundResponse(c, "Sale not found")
		return
	}

	h.logger.Info("Sale retrieved successfully via handler",
		zap.String("sale_id", sale.ID),
		zap.Float64("total_amount", sale.TotalAmount))

	utils.OKResponse(c, "Sale retrieved successfully", sale)
}

// GetAllSales handles GET /api/v1/sales
// @Summary Get All Sales
// @Description Retrieve all sales with pagination (requires authentication)
// @Tags Sales
// @Produce json
// @Param limit query integer false "Number of records to return (default: 50, max: 200)" example(50)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.PaginatedResponseModel{data=[]models.SaleResponse} "Sales retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales [get]
func (h *SalesHandler) GetAllSales(c *gin.Context) {
	h.logger.Info("Handling get all sales request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	h.logger.Debug("Calling sales service to get all sales",
		zap.Int("limit", params.Limit),
		zap.Int("offset", params.Offset))

	sales, total, err := h.salesService.GetAllSales(params.Limit, params.Offset)
	if err != nil {
		h.logger.Error("Failed to retrieve sales via service",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to retrieve sales", err)
		return
	}

	h.logger.Info("Sales retrieved successfully via handler",
		zap.Int("count", len(sales)),
		zap.Int64("total", total))

	utils.PaginatedOKResponse(c, sales, total, params.Limit, params.Offset)
}

// UpdateSale handles PUT /api/v1/sales/:id
// @Summary Update Sale
// @Description Update an existing sale by ID (requires authentication)
// @Tags Sales
// @Accept json
// @Produce json
// @Param id path string true "Sale ID" example(SALE_12345678)
// @Param request body models.UpdateSaleRequest true "Updated sale data"
// @Success 200 {object} utils.Response{data=models.SaleResponse} "Sale updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Sale not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/{id} [put]
func (h *SalesHandler) UpdateSale(c *gin.Context) {
	h.logger.Info("Handling update sale request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	id := c.Param("id")
	if id == "" {
		h.logger.Error("Sale ID is required but not provided")
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	var req models.UpdateSaleRequest
	if err := utils.ValidatePartialRequest(c, &req); err != nil {
		h.logger.Error("Invalid request body for update sale",
			zap.Error(err),
			zap.String("sale_id", id))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	h.logger.Debug("Calling sales service to update sale",
		zap.String("sale_id", id))

	sale, err := h.salesService.UpdateSale(id, &req)
	if err != nil {
		h.logger.Error("Failed to update sale via service",
			zap.Error(err),
			zap.String("sale_id", id))
		utils.HandleServiceError(c, "Failed to update sale", err)
		return
	}

	h.logger.Info("Sale updated successfully via handler",
		zap.String("sale_id", sale.ID))

	utils.OKResponse(c, "Sale updated successfully", sale)
}

// DeleteSale handles DELETE /api/v1/sales/:id
// @Summary Delete Sale
// @Description Delete a sale by ID (requires authentication)
// @Tags Sales
// @Produce json
// @Param id path string true "Sale ID" example(SALE_12345678)
// @Success 200 {object} utils.Response "Sale deleted successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Sale not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/{id} [delete]
func (h *SalesHandler) DeleteSale(c *gin.Context) {
	h.logger.Info("Handling delete sale request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	id := c.Param("id")
	if id == "" {
		h.logger.Error("Sale ID is required but not provided")
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	h.logger.Debug("Calling sales service to delete sale",
		zap.String("sale_id", id))

	err := h.salesService.DeleteSale(id)
	if err != nil {
		h.logger.Error("Failed to delete sale via service",
			zap.Error(err),
			zap.String("sale_id", id))
		utils.HandleServiceError(c, "Failed to delete sale", err)
		return
	}

	h.logger.Info("Sale deleted successfully via handler",
		zap.String("sale_id", id))

	utils.OKResponse(c, "Sale deleted successfully", nil)
}

// GetSalesByDateRange handles GET /api/v1/sales/date-range
// @Summary Get Sales by Date Range
// @Description Retrieve sales within a specific date range with pagination
// @Tags Sales
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)" example(2024-01-01)
// @Param end_date query string true "End date (YYYY-MM-DD)" example(2024-12-31)
// @Param limit query integer false "Number of records to return (default: 50, max: 200)" example(50)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.PaginatedResponseModel{data=[]models.SaleResponse} "Sales retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/date-range [get]
func (h *SalesHandler) GetSalesByDateRange(c *gin.Context) {
	h.logger.Info("Handling get sales by date range request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		h.logger.Error("Start date and end date are required but not provided")
		utils.BadRequestResponse(c, "Start date and end date are required", nil)
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		h.logger.Error("Invalid start date format",
			zap.Error(err),
			zap.String("start_date", startDateStr))
		utils.BadRequestResponse(c, "Invalid start date format", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		h.logger.Error("Invalid end date format",
			zap.Error(err),
			zap.String("end_date", endDateStr))
		utils.BadRequestResponse(c, "Invalid end date format", err)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	h.logger.Debug("Calling sales service to get sales by date range",
		zap.String("start_date", startDateStr),
		zap.String("end_date", endDateStr),
		zap.Int("limit", params.Limit),
		zap.Int("offset", params.Offset))

	sales, total, err := h.salesService.GetSalesByDateRange(startDate, endDate, params.Limit, params.Offset)
	if err != nil {
		h.logger.Error("Failed to retrieve sales by date range via service",
			zap.Error(err),
			zap.String("start_date", startDateStr),
			zap.String("end_date", endDateStr))
		utils.HandleServiceError(c, "Failed to retrieve sales", err)
		return
	}

	h.logger.Info("Sales by date range retrieved successfully via handler",
		zap.Int("count", len(sales)),
		zap.Int64("total", total),
		zap.String("start_date", startDateStr),
		zap.String("end_date", endDateStr))

	utils.PaginatedOKResponse(c, sales, total, params.Limit, params.Offset)
}

// GetSalesByStatus handles GET /api/v1/sales/status/:status
// @Summary Get Sales by Status
// @Description Retrieve all sales with a specific status with pagination
// @Tags Sales
// @Produce json
// @Param status path string true "Sale status" example(completed)
// @Param limit query integer false "Number of records to return (default: 50, max: 200)" example(50)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.PaginatedResponseModel{data=[]models.SaleResponse} "Sales retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/status/{status} [get]
func (h *SalesHandler) GetSalesByStatus(c *gin.Context) {
	h.logger.Info("Handling get sales by status request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	status := c.Param("status")
	if status == "" {
		h.logger.Error("Status is required but not provided")
		utils.BadRequestResponse(c, "Status is required", nil)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	h.logger.Debug("Calling sales service to get sales by status",
		zap.String("status", status),
		zap.Int("limit", params.Limit),
		zap.Int("offset", params.Offset))

	sales, total, err := h.salesService.GetSalesByStatus(status, params.Limit, params.Offset)
	if err != nil {
		h.logger.Error("Failed to retrieve sales by status via service",
			zap.Error(err),
			zap.String("status", status))
		utils.HandleServiceError(c, "Failed to retrieve sales", err)
		return
	}

	h.logger.Info("Sales by status retrieved successfully via handler",
		zap.Int("count", len(sales)),
		zap.Int64("total", total),
		zap.String("status", status))

	utils.PaginatedOKResponse(c, sales, total, params.Limit, params.Offset)
}

// GetTotalSalesAmount handles GET /api/v1/sales/total-amount
// @Summary Get Total Sales Amount
// @Description Calculate total sales amount within a date range
// @Tags Sales
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)" example(2024-01-01)
// @Param end_date query string true "End date (YYYY-MM-DD)" example(2024-12-31)
// @Success 200 {object} utils.Response{data=object{total_amount=number,start_date=string,end_date=string}} "Total sales amount calculated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/total-amount [get]
func (h *SalesHandler) GetTotalSalesAmount(c *gin.Context) {
	h.logger.Info("Handling get total sales amount request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		h.logger.Error("Start date and end date are required but not provided")
		utils.BadRequestResponse(c, "Start date and end date are required", nil)
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		h.logger.Error("Invalid start date format",
			zap.Error(err),
			zap.String("start_date", startDateStr))
		utils.BadRequestResponse(c, "Invalid start date format", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		h.logger.Error("Invalid end date format",
			zap.Error(err),
			zap.String("end_date", endDateStr))
		utils.BadRequestResponse(c, "Invalid end date format", err)
		return
	}

	h.logger.Debug("Calling sales service to calculate total sales amount",
		zap.String("start_date", startDateStr),
		zap.String("end_date", endDateStr))

	totalAmount, err := h.salesService.GetTotalSalesAmount(startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to calculate total sales amount via service",
			zap.Error(err),
			zap.String("start_date", startDateStr),
			zap.String("end_date", endDateStr))
		utils.HandleServiceError(c, "Failed to calculate total amount", err)
		return
	}

	h.logger.Info("Total sales amount calculated successfully via handler",
		zap.Float64("total_amount", totalAmount),
		zap.String("start_date", startDateStr),
		zap.String("end_date", endDateStr))

	utils.OKResponse(c, "Total sales amount calculated successfully", gin.H{
		"total_amount": totalAmount,
		"start_date":   startDateStr,
		"end_date":     endDateStr,
	})
}

// GetTopSellingProducts handles GET /api/v1/sales/top-selling
// @Summary Get Top Selling Products
// @Description Retrieve the top selling products
// @Tags Sales
// @Produce json
// @Param limit query integer false "Number of products to return (default: 10)" example(10)
// @Success 200 {object} utils.Response{data=[]models.TopSellingProductResponse} "Top selling products retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/top-selling [get]
func (h *SalesHandler) GetTopSellingProducts(c *gin.Context) {
	h.logger.Info("Handling get top selling products request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		h.logger.Error("Invalid limit parameter",
			zap.Error(err),
			zap.String("limit", limitStr))
		utils.BadRequestResponse(c, "Invalid limit parameter", err)
		return
	}

	h.logger.Debug("Calling sales service to get top selling products",
		zap.Int("limit", limit))

	products, err := h.salesService.GetTopSellingProducts(limit)
	if err != nil {
		h.logger.Error("Failed to retrieve top selling products via service",
			zap.Error(err),
			zap.Int("limit", limit))
		utils.HandleServiceError(c, "Failed to retrieve top selling products", err)
		return
	}

	h.logger.Info("Top selling products retrieved successfully via handler",
		zap.Int("count", len(products)),
		zap.Int("limit", limit))

	utils.OKResponse(c, "Top selling products retrieved successfully", products)
}

// GetSalesSummary handles GET /api/v1/sales/summary
// @Summary Get Sales Summary
// @Description Get overall sales summary statistics
// @Tags Sales
// @Produce json
// @Success 200 {object} utils.Response{data=object{total_sales=integer,total_amount=number,period=string}} "Sales summary retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/summary [get]
func (h *SalesHandler) GetSalesSummary(c *gin.Context) {
	h.logger.Info("Handling get sales summary request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	h.logger.Debug("Calling sales service to get sales summary")

	// For now, return a simple summary - this can be enhanced later
	summary := gin.H{
		"total_sales":  0,
		"total_amount": 0.0,
		"period":       "all_time",
	}

	h.logger.Info("Sales summary retrieved successfully via handler")

	utils.OKResponse(c, "Sales summary retrieved successfully", summary)
}

// UpdateSaleStatus handles PATCH /api/v1/sales/:id/status
// @Summary Update Sale Status
// @Description Update the status of a specific sale
// @Tags Sales
// @Accept json
// @Produce json
// @Param id path string true "Sale ID" example(SALE_12345678)
// @Param request body models.UpdateSaleStatusRequest true "New status"
// @Success 200 {object} utils.Response{data=models.SaleResponse} "Sale status updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Sale not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/{id}/status [patch]
func (h *SalesHandler) UpdateSaleStatus(c *gin.Context) {
	h.logger.Info("Handling update sale status request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	id := c.Param("id")
	if id == "" {
		h.logger.Error("Sale ID is required but not provided")
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := utils.ValidateRequest(c, &req); err != nil {
		h.logger.Error("Invalid request body for update sale status",
			zap.Error(err),
			zap.String("sale_id", id))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	h.logger.Debug("Calling sales service to update sale status",
		zap.String("sale_id", id),
		zap.String("new_status", req.Status))

	updateReq := models.UpdateSaleRequest{
		Status: &req.Status,
	}

	sale, err := h.salesService.UpdateSale(id, &updateReq)
	if err != nil {
		h.logger.Error("Failed to update sale status via service",
			zap.Error(err),
			zap.String("sale_id", id),
			zap.String("new_status", req.Status))
		utils.HandleServiceError(c, "Failed to update sale status", err)
		return
	}

	h.logger.Info("Sale status updated successfully via handler",
		zap.String("sale_id", sale.ID),
		zap.String("new_status", req.Status))

	utils.OKResponse(c, "Sale status updated successfully", sale)
}

// GetReturnsForSale handles GET /api/v1/sales/:id/returns
// @Summary Get Returns for Sale
// @Description Retrieve all returns associated with a specific sale
// @Tags Sales
// @Produce json
// @Param id path string true "Sale ID" example(SALE_12345678)
// @Success 200 {object} utils.Response{data=[]models.ReturnResponse} "Returns for sale retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Sale not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - resource already exists"
// @Failure 422 {object} utils.ErrorResponseModel "Unprocessable Entity - validation failed"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/{id}/returns [get]
func (h *SalesHandler) GetReturnsForSale(c *gin.Context) {
	h.logger.Info("Handling get returns for sale request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	saleID := c.Param("id")
	if saleID == "" {
		h.logger.Error("Sale ID is required but not provided")
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	h.logger.Debug("Calling sales service to get returns for sale",
		zap.String("sale_id", saleID))

	// For now, return empty array - this can be enhanced later
	returns := []interface{}{}

	h.logger.Info("Returns for sale retrieved successfully via handler",
		zap.String("sale_id", saleID),
		zap.Int("count", 0))

	utils.OKResponse(c, "Returns for sale retrieved successfully", returns)
}

// CancelSale handles POST /api/v1/sales/:id/cancel
// @Summary Cancel Sale
// @Description Cancel a sale and return inventory to original batches
// @Tags Sales
// @Accept json
// @Produce json
// @Param id path string true "Sale ID" example(SALE_12345678)
// @Param request body models.CancelSaleRequest true "Cancellation data"
// @Success 200 {object} utils.Response{data=models.CancelSaleResponse} "Sale cancelled successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 403 {object} utils.ErrorResponseModel "Forbidden - insufficient permissions"
// @Failure 404 {object} utils.ErrorResponseModel "Sale not found"
// @Failure 409 {object} utils.ErrorResponseModel "Conflict - sale cannot be cancelled"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/{id}/cancel [post]
func (h *SalesHandler) CancelSale(c *gin.Context) {
	saleID := c.Param("id")
	h.logger.Info("Handling cancel sale request",
		zap.String("sale_id", saleID),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	var req models.CancelSaleRequest

	// Validate request
	if err := utils.ValidateRequest(c, &req); err != nil {
		h.logger.Error("Invalid request body for cancel sale",
			zap.Error(err),
			zap.String("sale_id", saleID))
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	h.logger.Debug("Calling sales service to cancel sale",
		zap.String("sale_id", saleID),
		zap.String("reason", req.Reason),
		zap.String("performed_by", req.PerformedBy))

	response, err := h.salesService.CancelSale(saleID, &req)
	if err != nil {
		h.logger.Error("Failed to cancel sale via service",
			zap.Error(err),
			zap.String("sale_id", saleID),
			zap.String("reason", req.Reason))
		utils.HandleServiceError(c, "Failed to cancel sale", err)
		return
	}

	h.logger.Info("Sale cancelled successfully via handler",
		zap.String("sale_id", saleID),
		zap.String("cancellation_id", response.CancellationID),
		zap.Int("items_restored", len(response.InventoryRestored)))

	utils.OKResponse(c, "Sale cancelled successfully", response)
}

// RegisterRoutes registers all sales routes
func (h *SalesHandler) RegisterRoutes(router *gin.RouterGroup) {
	sales := router.Group("/sales")
	{
		// Apply authentication middleware
		sales.Use(h.aaaMiddleware.Authenticate())

		// Create/Update/Delete routes - CEO=CRUD, Store_Staff=CRUD, Tech_Support=R/W (temp)
		sales.POST("", h.aaaMiddleware.RequireOrgPermission("sale", "create"), h.CreateSale)
		sales.PUT("/:id", h.aaaMiddleware.RequireOrgPermission("sale", "update"), h.UpdateSale)
		sales.PATCH("/:id/status", h.aaaMiddleware.RequireOrgPermission("sale", "update"), h.UpdateSaleStatus)
		sales.DELETE("/:id", h.aaaMiddleware.RequireOrgPermission("sale", "delete"), h.DeleteSale)

		// Cancellation route - requires sale:cancel permission
		sales.POST("/:id/cancel", h.aaaMiddleware.RequireOrgPermission("sale", "cancel"), h.CancelSale)

		// Read routes - Director=R, CEO=CRUD, Auditor=R, Accountant=R, Tech_Support=R/W (temp), Store_Manager=R, Store_Staff=CRUD
		sales.GET("", h.aaaMiddleware.RequireOrgPermission("sale", "read"), h.GetAllSales)
		sales.GET("/summary", h.aaaMiddleware.RequireOrgPermission("sale_summary", "read"), h.GetSalesSummary)
		sales.GET("/:id", h.aaaMiddleware.RequireOrgPermission("sale", "read"), h.GetSale)
		sales.GET("/:id/returns", h.aaaMiddleware.RequireOrgPermission("sale", "read"), h.GetReturnsForSale)
		sales.GET("/date-range", h.aaaMiddleware.RequireOrgPermission("sale", "read"), h.GetSalesByDateRange)
		sales.GET("/status/:status", h.aaaMiddleware.RequireOrgPermission("sale", "read"), h.GetSalesByStatus)
		sales.GET("/total-amount", h.aaaMiddleware.RequireOrgPermission("sale", "read"), h.GetTotalSalesAmount)
		sales.GET("/top-selling", h.aaaMiddleware.RequireOrgPermission("sale", "read"), h.GetTopSellingProducts)
	}
}
