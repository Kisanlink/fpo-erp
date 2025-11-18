package handlers

import (
	"strconv"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

type SalesHandler struct {
	salesService  interfaces.SalesServiceInterface
	aaaMiddleware *aaa.AAAMiddleware
}

func NewSalesHandler(salesService interfaces.SalesServiceInterface, aaaMiddleware *aaa.AAAMiddleware) *SalesHandler {
	return &SalesHandler{
		salesService:  salesService,
		aaaMiddleware: aaaMiddleware,
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
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales [post]
func (h *SalesHandler) CreateSale(c *gin.Context) {
	var req models.CreateSaleRequest

	// Validate request
	if err := utils.ValidateRequest(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	sale, err := h.salesService.CreateSale(&req)
	if err != nil {
		utils.HandleServiceError(c, "Failed to create sale", err)
		return
	}

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
// @Failure 404 {object} utils.ErrorResponseModel "Sale not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/{id} [get]
func (h *SalesHandler) GetSale(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	sale, err := h.salesService.GetSale(id)
	if err != nil {
		utils.NotFoundResponse(c, "Sale not found")
		return
	}

	utils.OKResponse(c, "Sale retrieved successfully", sale)
}

// GetAllSales handles GET /api/v1/sales
// @Summary Get All Sales
// @Description Retrieve all sales with pagination (requires authentication)
// @Tags Sales
// @Produce json
// @Param limit query integer false "Number of records to return (default: 10)" example(10)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.Response{data=[]models.SaleResponse} "Sales retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales [get]
func (h *SalesHandler) GetAllSales(c *gin.Context) {
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

	sales, err := h.salesService.GetAllSales(limit, offset)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve sales", err)
		return
	}

	utils.OKResponse(c, "Sales retrieved successfully", sales)
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
// @Failure 404 {object} utils.ErrorResponseModel "Sale not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/{id} [put]
func (h *SalesHandler) UpdateSale(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	var req models.UpdateSaleRequest
	if err := utils.ValidatePartialRequest(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	sale, err := h.salesService.UpdateSale(id, &req)
	if err != nil {
		utils.HandleServiceError(c, "Failed to update sale", err)
		return
	}

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
// @Failure 404 {object} utils.ErrorResponseModel "Sale not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/{id} [delete]
func (h *SalesHandler) DeleteSale(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	err := h.salesService.DeleteSale(id)
	if err != nil {
		utils.HandleServiceError(c, "Failed to delete sale", err)
		return
	}

	utils.OKResponse(c, "Sale deleted successfully", nil)
}

// GetSalesByDateRange handles GET /api/v1/sales/date-range
// @Summary Get Sales by Date Range
// @Description Retrieve sales within a specific date range
// @Tags Sales
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)" example(2024-01-01)
// @Param end_date query string true "End date (YYYY-MM-DD)" example(2024-12-31)
// @Success 200 {object} utils.Response{data=[]models.SaleResponse} "Sales retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/date-range [get]
func (h *SalesHandler) GetSalesByDateRange(c *gin.Context) {
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

	sales, err := h.salesService.GetSalesByDateRange(startDate, endDate)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve sales", err)
		return
	}

	utils.OKResponse(c, "Sales retrieved successfully", sales)
}

// GetSalesByStatus handles GET /api/v1/sales/status/:status
// @Summary Get Sales by Status
// @Description Retrieve all sales with a specific status
// @Tags Sales
// @Produce json
// @Param status path string true "Sale status" example(completed)
// @Success 200 {object} utils.Response{data=[]models.SaleResponse} "Sales retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/status/{status} [get]
func (h *SalesHandler) GetSalesByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		utils.BadRequestResponse(c, "Status is required", nil)
		return
	}

	sales, err := h.salesService.GetSalesByStatus(status)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve sales", err)
		return
	}

	utils.OKResponse(c, "Sales retrieved successfully", sales)
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
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/total-amount [get]
func (h *SalesHandler) GetTotalSalesAmount(c *gin.Context) {
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

	totalAmount, err := h.salesService.GetTotalSalesAmount(startDate, endDate)
	if err != nil {
		utils.HandleServiceError(c, "Failed to calculate total amount", err)
		return
	}

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
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/top-selling [get]
func (h *SalesHandler) GetTopSellingProducts(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid limit parameter", err)
		return
	}

	products, err := h.salesService.GetTopSellingProducts(limit)
	if err != nil {
		utils.HandleServiceError(c, "Failed to retrieve top selling products", err)
		return
	}

	utils.OKResponse(c, "Top selling products retrieved successfully", products)
}

// GetSalesSummary handles GET /api/v1/sales/summary
// @Summary Get Sales Summary
// @Description Get overall sales summary statistics
// @Tags Sales
// @Produce json
// @Success 200 {object} utils.Response{data=object{total_sales=integer,total_amount=number,period=string}} "Sales summary retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/summary [get]
func (h *SalesHandler) GetSalesSummary(c *gin.Context) {
	// For now, return a simple summary - this can be enhanced later
	utils.OKResponse(c, "Sales summary retrieved successfully", gin.H{
		"total_sales":  0,
		"total_amount": 0.0,
		"period":       "all_time",
	})
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
// @Failure 404 {object} utils.ErrorResponseModel "Sale not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/{id}/status [patch]
func (h *SalesHandler) UpdateSaleStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := utils.ValidateRequest(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	updateReq := models.UpdateSaleRequest{
		Status: &req.Status,
	}

	sale, err := h.salesService.UpdateSale(id, &updateReq)
	if err != nil {
		utils.HandleServiceError(c, "Failed to update sale status", err)
		return
	}

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
// @Failure 404 {object} utils.ErrorResponseModel "Sale not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/{id}/returns [get]
func (h *SalesHandler) GetReturnsForSale(c *gin.Context) {
	saleID := c.Param("id")
	if saleID == "" {
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	// For now, return empty array - this can be enhanced later
	utils.OKResponse(c, "Returns for sale retrieved successfully", []interface{}{})
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
