package handlers

import (
	"net/http"

	logger "kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/services/interfaces"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InvoiceHandler handles invoice PDF generation requests
type InvoiceHandler struct {
	invoiceService interfaces.InvoiceServiceInterface
	logger         logger.Logger
}

// NewInvoiceHandler creates a new invoice handler
func NewInvoiceHandler(invoiceService interfaces.InvoiceServiceInterface, logger logger.Logger) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceService: invoiceService,
		logger:         logger,
	}
}

// DownloadInvoice handles GET /api/v1/sales/:id/invoice
// @Summary Download Invoice PDF
// @Description Generate and download a PDF invoice for a specific sale
// @Tags Sales
// @Produce application/pdf
// @Param id path string true "Sale ID" example(SALE00000001)
// @Success 200 {file} file "PDF invoice file"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request - missing required settings"
// @Failure 404 {object} utils.ErrorResponseModel "Sale not found"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/{id}/invoice [get]
func (h *InvoiceHandler) DownloadInvoice(c *gin.Context) {
	h.logger.Info("Handling download invoice request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get sale ID from URL
	saleID := c.Param("id")
	if saleID == "" {
		h.logger.Error("Sale ID is required but not provided")
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	h.logger.Debug("Generating invoice PDF",
		zap.String("sale_id", saleID))

	// Generate PDF
	pdfBytes, contentType, filename, err := h.invoiceService.GenerateInvoicePDF(c.Request.Context(), saleID)
	if err != nil {
		h.logger.Error("Service error generating invoice PDF",
			zap.Error(err),
			zap.String("sale_id", saleID))
		utils.HandleServiceError(c, "Failed to generate invoice", err)
		return
	}

	h.logger.Info("Invoice PDF generated successfully",
		zap.String("sale_id", saleID),
		zap.String("filename", filename),
		zap.Int("size_bytes", len(pdfBytes)))

	// Set headers for file download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", string(rune(len(pdfBytes))))
	c.Header("Cache-Control", "private, max-age=0")

	// Send the PDF
	c.Data(http.StatusOK, contentType, pdfBytes)
}

// CheckInvoiceRequirements handles GET /api/v1/sales/invoice-requirements
// @Summary Check Invoice Requirements
// @Description Check if all required settings for invoice generation exist
// @Tags Sales
// @Produce json
// @Success 200 {object} utils.Response "Invoice requirements check complete"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sales/invoice-requirements [get]
func (h *InvoiceHandler) CheckInvoiceRequirements(c *gin.Context) {
	h.logger.Info("Handling check invoice requirements request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Check invoice requirements
	ready, missing, err := h.invoiceService.CheckInvoiceRequirements(c.Request.Context())
	if err != nil {
		h.logger.Error("Service error checking invoice requirements",
			zap.Error(err))
		utils.HandleServiceError(c, "Failed to check invoice requirements", err)
		return
	}

	response := map[string]interface{}{
		"ready":            ready,
		"missing_settings": missing,
	}

	if ready {
		h.logger.Info("All invoice requirements satisfied")
		utils.OKResponse(c, "All invoice requirements satisfied", response)
	} else {
		h.logger.Warn("Missing required settings for invoice",
			zap.Strings("missing", missing))
		utils.OKResponse(c, "Missing required settings for invoice generation", response)
	}
}

// RegisterRoutes registers invoice routes on the sales group
func (h *InvoiceHandler) RegisterRoutes(v1 *gin.RouterGroup) {
	sales := v1.Group("/sales")
	{
		// Invoice requirements check
		sales.GET("/invoice-requirements", h.CheckInvoiceRequirements)

		// Invoice download for specific sale
		sales.GET("/:id/invoice", h.DownloadInvoice)
	}
}
