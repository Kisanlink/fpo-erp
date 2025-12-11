package handlers

import (
	"fmt"
	"net/http"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	logger "kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ReportHandler handles report generation endpoints
type ReportHandler struct {
	reportService interfaces.ReportServiceInterface
	exportService interfaces.ExportServiceInterface
	aaaMiddleware *aaa.AAAMiddleware
	logger        logger.Logger
}

// NewReportHandler creates a new report handler
func NewReportHandler(
	reportService interfaces.ReportServiceInterface,
	exportService interfaces.ExportServiceInterface,
	aaaMiddleware *aaa.AAAMiddleware,
	logger logger.Logger,
) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		exportService: exportService,
		aaaMiddleware: aaaMiddleware,
		logger:        logger,
	}
}

// GetProductReport handles GET /api/v1/reports/products
// @Summary Generate Product Master Report
// @Description Generate a product master report with filtering, pagination, and export options
// @Tags Reports
// @Produce json,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/pdf
// @Param format query string false "Output format: json, xlsx, pdf" Enums(json, xlsx, pdf) default(json)
// @Param limit query integer false "Records per page (max: 500)" default(50)
// @Param offset query integer false "Records to skip" default(0)
// @Param search query string false "Search by product name or description"
// @Param has_variants query boolean false "Filter products with/without variants"
// @Param is_active query boolean false "Filter by active status"
// @Param sort_by query string false "Sort field" default(created_at)
// @Param sort_order query string false "Sort order: asc, desc" Enums(asc, desc) default(desc)
// @Success 200 {object} utils.Response{data=models.ReportResponse} "Report generated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Invalid parameters"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/reports/products [get]
func (h *ReportHandler) GetProductReport(c *gin.Context) {
	h.logger.Info("Handling product report request", zap.String("path", c.Request.URL.Path))

	var filter models.ProductReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Invalid filter parameters", zap.Error(err))
		utils.BadRequestResponse(c, "Invalid filter parameters", err)
		return
	}

	// Apply defaults
	h.applyDefaults(&filter.BaseReportFilter)

	report, err := h.reportService.GenerateProductReport(&filter)
	if err != nil {
		h.logger.Error("Failed to generate product report", zap.Error(err))
		utils.HandleServiceError(c, "Failed to generate report", err)
		return
	}

	// Handle export formats
	if filter.Format == "xlsx" || filter.Format == "pdf" {
		h.handleExport(c, "products", report, filter.Format)
		return
	}

	utils.OKResponse(c, "Report generated successfully", report)
}

// GetVendorReport handles GET /api/v1/reports/vendors
// @Summary Generate Vendor Master Report
// @Description Generate a vendor master report with filtering, pagination, and export options
// @Tags Reports
// @Produce json,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/pdf
// @Param format query string false "Output format: json, xlsx, pdf" Enums(json, xlsx, pdf) default(json)
// @Param limit query integer false "Records per page (max: 500)" default(50)
// @Param offset query integer false "Records to skip" default(0)
// @Param search query string false "Search by company name, contact person, or GST"
// @Param is_active query boolean false "Filter by active status"
// @Param has_gst query boolean false "Filter vendors with/without GST"
// @Param sort_by query string false "Sort field"
// @Param sort_order query string false "Sort order: asc, desc" Enums(asc, desc)
// @Success 200 {object} utils.Response{data=models.ReportResponse} "Report generated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Invalid parameters"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/reports/vendors [get]
func (h *ReportHandler) GetVendorReport(c *gin.Context) {
	h.logger.Info("Handling vendor report request", zap.String("path", c.Request.URL.Path))

	var filter models.VendorReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Invalid filter parameters", zap.Error(err))
		utils.BadRequestResponse(c, "Invalid filter parameters", err)
		return
	}

	h.applyDefaults(&filter.BaseReportFilter)

	report, err := h.reportService.GenerateVendorReport(&filter)
	if err != nil {
		h.logger.Error("Failed to generate vendor report", zap.Error(err))
		utils.HandleServiceError(c, "Failed to generate report", err)
		return
	}

	if filter.Format == "xlsx" || filter.Format == "pdf" {
		h.handleExport(c, "vendors", report, filter.Format)
		return
	}

	utils.OKResponse(c, "Report generated successfully", report)
}

// GetCustomerReport handles GET /api/v1/reports/customers
// @Summary Generate Customer Report
// @Description Generate a customer report with filtering, pagination, and export options
// @Tags Reports
// @Produce json,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/pdf
// @Param format query string false "Output format: json, xlsx, pdf" Enums(json, xlsx, pdf) default(json)
// @Param limit query integer false "Records per page (max: 500)" default(50)
// @Param offset query integer false "Records to skip" default(0)
// @Param search query string false "Search by farmer ID"
// @Param warehouse_id query string false "Filter by warehouse ID"
// @Param min_purchase_value query number false "Minimum total purchase value"
// @Param max_purchase_value query number false "Maximum total purchase value"
// @Param start_date query string false "Filter start date (YYYY-MM-DD)"
// @Param end_date query string false "Filter end date (YYYY-MM-DD)"
// @Success 200 {object} utils.Response{data=models.ReportResponse} "Report generated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Invalid parameters"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/reports/customers [get]
func (h *ReportHandler) GetCustomerReport(c *gin.Context) {
	h.logger.Info("Handling customer report request", zap.String("path", c.Request.URL.Path))

	var filter models.CustomerReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Invalid filter parameters", zap.Error(err))
		utils.BadRequestResponse(c, "Invalid filter parameters", err)
		return
	}

	h.applyDefaults(&filter.BaseReportFilter)

	report, err := h.reportService.GenerateCustomerReport(&filter)
	if err != nil {
		h.logger.Error("Failed to generate customer report", zap.Error(err))
		utils.HandleServiceError(c, "Failed to generate report", err)
		return
	}

	if filter.Format == "xlsx" || filter.Format == "pdf" {
		h.handleExport(c, "customers", report, filter.Format)
		return
	}

	utils.OKResponse(c, "Report generated successfully", report)
}

// GetInventoryReport handles GET /api/v1/reports/inventory
// @Summary Generate Inventory Report
// @Description Generate an inventory report with filtering, pagination, and export options
// @Tags Reports
// @Produce json,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/pdf
// @Param format query string false "Output format: json, xlsx, pdf" Enums(json, xlsx, pdf) default(json)
// @Param limit query integer false "Records per page (max: 500)" default(50)
// @Param offset query integer false "Records to skip" default(0)
// @Param warehouse_id query string false "Filter by warehouse ID"
// @Param product_id query string false "Filter by product ID"
// @Param variant_id query string false "Filter by variant ID"
// @Param low_stock query boolean false "Show only low stock items"
// @Param expiring_soon query boolean false "Show items expiring within 30 days"
// @Param expired query boolean false "Include expired items"
// @Param min_quantity query integer false "Minimum quantity filter"
// @Param max_quantity query integer false "Maximum quantity filter"
// @Success 200 {object} utils.Response{data=models.ReportResponse} "Report generated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Invalid parameters"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/reports/inventory [get]
func (h *ReportHandler) GetInventoryReport(c *gin.Context) {
	h.logger.Info("Handling inventory report request", zap.String("path", c.Request.URL.Path))

	var filter models.InventoryReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Invalid filter parameters", zap.Error(err))
		utils.BadRequestResponse(c, "Invalid filter parameters", err)
		return
	}

	h.applyDefaults(&filter.BaseReportFilter)

	report, err := h.reportService.GenerateInventoryReport(&filter)
	if err != nil {
		h.logger.Error("Failed to generate inventory report", zap.Error(err))
		utils.HandleServiceError(c, "Failed to generate report", err)
		return
	}

	if filter.Format == "xlsx" || filter.Format == "pdf" {
		h.handleExport(c, "inventory", report, filter.Format)
		return
	}

	utils.OKResponse(c, "Report generated successfully", report)
}

// GetPurchaseReport handles GET /api/v1/reports/purchases
// @Summary Generate Purchase Report
// @Description Generate a purchase orders report with filtering, pagination, and export options
// @Tags Reports
// @Produce json,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/pdf
// @Param format query string false "Output format: json, xlsx, pdf" Enums(json, xlsx, pdf) default(json)
// @Param limit query integer false "Records per page (max: 500)" default(50)
// @Param offset query integer false "Records to skip" default(0)
// @Param collaborator_id query string false "Filter by vendor/collaborator ID"
// @Param warehouse_id query string false "Filter by warehouse ID"
// @Param status query []string false "Filter by status (comma-separated)"
// @Param payment_status query []string false "Filter by payment status"
// @Param po_number query string false "Search by PO number"
// @Param start_date query string false "Filter start date (YYYY-MM-DD)"
// @Param end_date query string false "Filter end date (YYYY-MM-DD)"
// @Success 200 {object} utils.Response{data=models.ReportResponse} "Report generated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Invalid parameters"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/reports/purchases [get]
func (h *ReportHandler) GetPurchaseReport(c *gin.Context) {
	h.logger.Info("Handling purchase report request", zap.String("path", c.Request.URL.Path))

	var filter models.PurchaseReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Invalid filter parameters", zap.Error(err))
		utils.BadRequestResponse(c, "Invalid filter parameters", err)
		return
	}

	h.applyDefaults(&filter.BaseReportFilter)

	report, err := h.reportService.GeneratePurchaseReport(&filter)
	if err != nil {
		h.logger.Error("Failed to generate purchase report", zap.Error(err))
		utils.HandleServiceError(c, "Failed to generate report", err)
		return
	}

	if filter.Format == "xlsx" || filter.Format == "pdf" {
		h.handleExport(c, "purchases", report, filter.Format)
		return
	}

	utils.OKResponse(c, "Report generated successfully", report)
}

// GetSalesReport handles GET /api/v1/reports/sales
// @Summary Generate Sales Report
// @Description Generate a sales report with filtering, pagination, and export options
// @Tags Reports
// @Produce json,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/pdf
// @Param format query string false "Output format: json, xlsx, pdf" Enums(json, xlsx, pdf) default(json)
// @Param limit query integer false "Records per page (max: 500)" default(50)
// @Param offset query integer false "Records to skip" default(0)
// @Param warehouse_id query string false "Filter by warehouse ID"
// @Param customer_id query string false "Filter by customer ID"
// @Param status query []string false "Filter by status (comma-separated)"
// @Param payment_mode query []string false "Filter by payment mode"
// @Param sale_type query []string false "Filter by sale type"
// @Param min_amount query number false "Minimum sale amount"
// @Param max_amount query number false "Maximum sale amount"
// @Param start_date query string false "Filter start date (YYYY-MM-DD)"
// @Param end_date query string false "Filter end date (YYYY-MM-DD)"
// @Success 200 {object} utils.Response{data=models.ReportResponse} "Report generated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Invalid parameters"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/reports/sales [get]
func (h *ReportHandler) GetSalesReport(c *gin.Context) {
	h.logger.Info("Handling sales report request", zap.String("path", c.Request.URL.Path))

	var filter models.SalesReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Invalid filter parameters", zap.Error(err))
		utils.BadRequestResponse(c, "Invalid filter parameters", err)
		return
	}

	h.applyDefaults(&filter.BaseReportFilter)

	report, err := h.reportService.GenerateSalesReport(&filter)
	if err != nil {
		h.logger.Error("Failed to generate sales report", zap.Error(err))
		utils.HandleServiceError(c, "Failed to generate report", err)
		return
	}

	if filter.Format == "xlsx" || filter.Format == "pdf" {
		h.handleExport(c, "sales", report, filter.Format)
		return
	}

	utils.OKResponse(c, "Report generated successfully", report)
}

// GetReturnsReport handles GET /api/v1/reports/returns
// @Summary Generate Returns Report
// @Description Generate a returns report with filtering, pagination, and export options
// @Tags Reports
// @Produce json,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/pdf
// @Param format query string false "Output format: json, xlsx, pdf" Enums(json, xlsx, pdf) default(json)
// @Param limit query integer false "Records per page (max: 500)" default(50)
// @Param offset query integer false "Records to skip" default(0)
// @Param sale_id query string false "Filter by original sale ID"
// @Param warehouse_id query string false "Filter by warehouse ID"
// @Param status query []string false "Filter by status (comma-separated)"
// @Param min_refund query number false "Minimum refund amount"
// @Param max_refund query number false "Maximum refund amount"
// @Param start_date query string false "Filter start date (YYYY-MM-DD)"
// @Param end_date query string false "Filter end date (YYYY-MM-DD)"
// @Success 200 {object} utils.Response{data=models.ReportResponse} "Report generated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Invalid parameters"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/reports/returns [get]
func (h *ReportHandler) GetReturnsReport(c *gin.Context) {
	h.logger.Info("Handling returns report request", zap.String("path", c.Request.URL.Path))

	var filter models.ReturnsReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Invalid filter parameters", zap.Error(err))
		utils.BadRequestResponse(c, "Invalid filter parameters", err)
		return
	}

	h.applyDefaults(&filter.BaseReportFilter)

	report, err := h.reportService.GenerateReturnsReport(&filter)
	if err != nil {
		h.logger.Error("Failed to generate returns report", zap.Error(err))
		utils.HandleServiceError(c, "Failed to generate report", err)
		return
	}

	if filter.Format == "xlsx" || filter.Format == "pdf" {
		h.handleExport(c, "returns", report, filter.Format)
		return
	}

	utils.OKResponse(c, "Report generated successfully", report)
}

// GetGRNReport handles GET /api/v1/reports/grn
// @Summary Generate GRN Report
// @Description Generate a goods receipt note report with filtering, pagination, and export options
// @Tags Reports
// @Produce json,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/pdf
// @Param format query string false "Output format: json, xlsx, pdf" Enums(json, xlsx, pdf) default(json)
// @Param limit query integer false "Records per page (max: 500)" default(50)
// @Param offset query integer false "Records to skip" default(0)
// @Param po_id query string false "Filter by purchase order ID"
// @Param po_number query string false "Search by PO number"
// @Param vendor_id query string false "Filter by vendor/collaborator ID"
// @Param warehouse_id query string false "Filter by warehouse ID"
// @Param quality_status query []string false "Filter by quality status (comma-separated)"
// @Param min_value query number false "Minimum received value"
// @Param max_value query number false "Maximum received value"
// @Param start_date query string false "Filter start date (YYYY-MM-DD)"
// @Param end_date query string false "Filter end date (YYYY-MM-DD)"
// @Success 200 {object} utils.Response{data=models.ReportResponse} "Report generated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Invalid parameters"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/reports/grn [get]
func (h *ReportHandler) GetGRNReport(c *gin.Context) {
	h.logger.Info("Handling GRN report request", zap.String("path", c.Request.URL.Path))

	var filter models.GRNReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Invalid filter parameters", zap.Error(err))
		utils.BadRequestResponse(c, "Invalid filter parameters", err)
		return
	}

	h.applyDefaults(&filter.BaseReportFilter)

	report, err := h.reportService.GenerateGRNReport(&filter)
	if err != nil {
		h.logger.Error("Failed to generate GRN report", zap.Error(err))
		utils.HandleServiceError(c, "Failed to generate report", err)
		return
	}

	if filter.Format == "xlsx" || filter.Format == "pdf" {
		h.handleExport(c, "grn", report, filter.Format)
		return
	}

	utils.OKResponse(c, "Report generated successfully", report)
}

// applyDefaults applies default values to base report filter
func (h *ReportHandler) applyDefaults(filter *models.BaseReportFilter) {
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.Limit > 500 {
		filter.Limit = 500
	}
	if filter.Format == "" {
		filter.Format = "json"
	}
	if filter.SortOrder == "" {
		filter.SortOrder = "desc"
	}
}

// handleExport handles report export in XLSX or PDF format
func (h *ReportHandler) handleExport(c *gin.Context, reportType string, data *models.ReportResponse, format string) {
	h.logger.Info("Exporting report",
		zap.String("report_type", reportType),
		zap.String("format", format))

	var content []byte
	var err error
	var contentType, fileExt string

	switch format {
	case "xlsx":
		content, err = h.exportService.ExportToXLSX(reportType, data)
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		fileExt = "xlsx"
	case "pdf":
		content, err = h.exportService.ExportToPDF(reportType, data)
		contentType = "application/pdf"
		fileExt = "pdf"
	default:
		utils.BadRequestResponse(c, "Invalid export format", nil)
		return
	}

	if err != nil {
		h.logger.Error("Failed to export report",
			zap.Error(err),
			zap.String("format", format),
			zap.String("report_type", reportType))
		utils.InternalServerErrorResponse(c, "Failed to export report", err)
		return
	}

	filename := fmt.Sprintf("%s_report_%s.%s", reportType, time.Now().Format("2006-01-02"), fileExt)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, contentType, content)

	h.logger.Info("Report exported successfully",
		zap.String("filename", filename),
		zap.Int("size_bytes", len(content)))
}

// RegisterRoutes registers report routes
func (h *ReportHandler) RegisterRoutes(router *gin.RouterGroup) {
	reports := router.Group("/reports")
	{
		// Apply authentication middleware
		reports.Use(h.aaaMiddleware.Authenticate())

		// Report routes - require read permissions
		reports.GET("/products", h.aaaMiddleware.RequireOrgPermission("report", "read"), h.GetProductReport)
		reports.GET("/vendors", h.aaaMiddleware.RequireOrgPermission("report", "read"), h.GetVendorReport)
		reports.GET("/customers", h.aaaMiddleware.RequireOrgPermission("report", "read"), h.GetCustomerReport)
		reports.GET("/inventory", h.aaaMiddleware.RequireOrgPermission("report", "read"), h.GetInventoryReport)
		reports.GET("/purchases", h.aaaMiddleware.RequireOrgPermission("report", "read"), h.GetPurchaseReport)
		reports.GET("/sales", h.aaaMiddleware.RequireOrgPermission("report", "read"), h.GetSalesReport)
		reports.GET("/returns", h.aaaMiddleware.RequireOrgPermission("report", "read"), h.GetReturnsReport)
		reports.GET("/grn", h.aaaMiddleware.RequireOrgPermission("report", "read"), h.GetGRNReport)
	}
}
