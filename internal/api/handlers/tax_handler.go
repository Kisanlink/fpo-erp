package handlers

import (
	"strconv"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

type TaxHandler struct {
	taxService    *services.TaxService
	aaaMiddleware *aaa.AAAMiddleware
}

func NewTaxHandler(taxService *services.TaxService, aaaMiddleware *aaa.AAAMiddleware) *TaxHandler {
	return &TaxHandler{
		taxService:    taxService,
		aaaMiddleware: aaaMiddleware,
	}
}

// RegisterRoutes registers all tax routes
func (h *TaxHandler) RegisterRoutes(router *gin.RouterGroup) {
	taxes := router.Group("/taxes")
	{
		// CRUD operations
		taxes.POST("", h.aaaMiddleware.RequireOrgPermission("tax", "create"), h.CreateTax)
		taxes.GET("", h.aaaMiddleware.RequireOrgPermission("tax", "read"), h.GetAllTaxes)
		taxes.GET("/:id", h.aaaMiddleware.RequireOrgPermission("tax", "read"), h.GetTax)
		taxes.PUT("/:id", h.aaaMiddleware.RequireOrgPermission("tax", "update"), h.UpdateTax)
		taxes.DELETE("/:id", h.aaaMiddleware.RequireOrgPermission("tax", "delete"), h.DeleteTax)

		// Specialized endpoints
		taxes.GET("/active", h.aaaMiddleware.RequireOrgPermission("tax", "read"), h.GetActiveTaxes)
		taxes.GET("/type/:type", h.aaaMiddleware.RequireOrgPermission("tax", "read"), h.GetTaxesByType)
		taxes.GET("/status/:status", h.aaaMiddleware.RequireOrgPermission("tax", "read"), h.GetTaxesByStatus)

		// Tax calculation
		taxes.POST("/calculate", h.aaaMiddleware.RequireOrgPermission("tax", "calculate"), h.CalculateTax)

		// Tax applications and summaries
		taxes.GET("/applications/sale/:saleID", h.aaaMiddleware.RequireOrgPermission("tax", "read"), h.GetTaxApplicationsBySale)
		taxes.GET("/applications/return/:returnID", h.aaaMiddleware.RequireOrgPermission("tax", "read"), h.GetTaxApplicationsByReturn)
		taxes.GET("/summary/sale/:saleID", h.aaaMiddleware.RequireOrgPermission("tax", "read"), h.GetTaxSummaryBySale)
		taxes.GET("/summary/return/:returnID", h.aaaMiddleware.RequireOrgPermission("tax", "read"), h.GetTaxSummaryByReturn)
	}
}

// CreateTax creates a new tax
// @Summary Create Tax
// @Description Create a new tax configuration (requires authentication)
// @Tags Taxes
// @Accept json
// @Produce json
// @Param request body models.CreateTaxRequest true "Tax data"
// @Success 201 {object} utils.Response{data=models.TaxResponse} "Tax created successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/taxes [post]
func (h *TaxHandler) CreateTax(c *gin.Context) {
	var req models.CreateTaxRequest

	// Validate request
	if err := utils.ValidateRequest(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Get user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	tax, err := h.taxService.CreateTax(&req, userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create tax", err)
		return
	}

	utils.CreatedResponse(c, "Tax created successfully", tax)
}

// GetTax retrieves a tax by ID
// @Summary Get Tax
// @Description Retrieve a specific tax configuration by ID
// @Tags Taxes
// @Produce json
// @Param id path string true "Tax ID" example(TAX_12345678)
// @Success 200 {object} utils.Response{data=models.TaxResponse} "Tax retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Tax not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/taxes/{id} [get]
func (h *TaxHandler) GetTax(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Tax ID is required", nil)
		return
	}

	tax, err := h.taxService.GetTax(id)
	if err != nil {
		utils.NotFoundResponse(c, "Tax not found")
		return
	}

	utils.OKResponse(c, "Tax retrieved successfully", tax)
}

// GetAllTaxes retrieves all taxes with pagination
// @Summary Get All Taxes
// @Description Retrieve all tax configurations with pagination
// @Tags Taxes
// @Produce json
// @Param limit query integer false "Number of records to return (default: 10)" example(10)
// @Param offset query integer false "Number of records to skip (default: 0)" example(0)
// @Success 200 {object} utils.Response{data=[]models.TaxResponse} "Taxes retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/taxes [get]
func (h *TaxHandler) GetAllTaxes(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	taxes, err := h.taxService.GetAllTaxes(limit, offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve taxes", err)
		return
	}

	utils.OKResponse(c, "Taxes retrieved successfully", taxes)
}

// GetActiveTaxes retrieves all currently active taxes
// @Summary Get Active Taxes
// @Description Retrieve all currently active tax configurations
// @Tags Taxes
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.TaxResponse} "Active taxes retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/taxes/active [get]
func (h *TaxHandler) GetActiveTaxes(c *gin.Context) {
	taxes, err := h.taxService.GetActiveTaxes()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve active taxes", err)
		return
	}

	utils.OKResponse(c, "Active taxes retrieved successfully", taxes)
}

// GetTaxesByType retrieves taxes by type
// @Summary Get Taxes by Type
// @Description Retrieve all taxes of a specific type
// @Tags Taxes
// @Produce json
// @Param type path string true "Tax type" Enums(cgst,sgst,igst,vat,stt,tds,tcs,excise,customs,item_specific,category,flat)
// @Success 200 {object} utils.Response{data=[]models.TaxResponse} "Taxes retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/taxes/type/{type} [get]
func (h *TaxHandler) GetTaxesByType(c *gin.Context) {
	taxType := models.TaxType(c.Param("type"))
	if taxType == "" {
		utils.BadRequestResponse(c, "Tax type is required", nil)
		return
	}

	// Validate tax type
	validTypes := []models.TaxType{
		models.TaxTypeCGST, models.TaxTypeSGST, models.TaxTypeIGST,
		models.TaxTypeVAT, models.TaxTypeSTT, models.TaxTypeTDS,
		models.TaxTypeTCS, models.TaxTypeExcise, models.TaxTypeCustoms,
		models.TaxTypeItemSpecific, models.TaxTypeCategory, models.TaxTypeFlat,
	}

	isValid := false
	for _, validType := range validTypes {
		if taxType == validType {
			isValid = true
			break
		}
	}

	if !isValid {
		utils.BadRequestResponse(c, "Invalid tax type", nil)
		return
	}

	taxes, err := h.taxService.GetTaxesByType(taxType)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve taxes by type", err)
		return
	}

	utils.OKResponse(c, "Taxes retrieved successfully", taxes)
}

// GetTaxesByStatus retrieves taxes by status
// @Summary Get Taxes by Status
// @Description Retrieve all taxes with a specific status
// @Tags Taxes
// @Produce json
// @Param status path string true "Tax status" Enums(active,inactive,expired,scheduled)
// @Success 200 {object} utils.Response{data=[]models.TaxResponse} "Taxes retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/taxes/status/{status} [get]
func (h *TaxHandler) GetTaxesByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		utils.BadRequestResponse(c, "Tax status is required", nil)
		return
	}

	// Validate status
	validStatuses := []string{"active", "inactive", "expired", "scheduled"}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}

	if !isValid {
		utils.BadRequestResponse(c, "Invalid tax status", nil)
		return
	}

	taxes, err := h.taxService.GetTaxesByStatus(status)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve taxes by status", err)
		return
	}

	utils.OKResponse(c, "Taxes retrieved successfully", taxes)
}

// UpdateTax updates an existing tax
// @Summary Update Tax
// @Description Update an existing tax configuration by ID
// @Tags Taxes
// @Accept json
// @Produce json
// @Param id path string true "Tax ID" example(TAX_12345678)
// @Param request body models.UpdateTaxRequest true "Updated tax data"
// @Success 200 {object} utils.Response{data=models.TaxResponse} "Tax updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Tax not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/taxes/{id} [put]
func (h *TaxHandler) UpdateTax(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Tax ID is required", nil)
		return
	}

	var req models.UpdateTaxRequest
	if err := utils.ValidatePartialRequest(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Get user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	tax, err := h.taxService.UpdateTax(id, &req, userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update tax", err)
		return
	}

	utils.OKResponse(c, "Tax updated successfully", tax)
}

// DeleteTax deletes a tax
// @Summary Delete Tax
// @Description Delete a tax configuration by ID
// @Tags Taxes
// @Produce json
// @Param id path string true "Tax ID" example(TAX_12345678)
// @Success 200 {object} utils.Response "Tax deleted successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 404 {object} utils.ErrorResponseModel "Tax not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/taxes/{id} [delete]
func (h *TaxHandler) DeleteTax(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Tax ID is required", nil)
		return
	}

	err := h.taxService.DeleteTax(id)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete tax", err)
		return
	}

	utils.OKResponse(c, "Tax deleted successfully", nil)
}

// CalculateTax calculates taxes for a given transaction
// @Summary Calculate Tax
// @Description Calculate taxes for a transaction with given items
// @Tags Taxes
// @Accept json
// @Produce json
// @Param request body models.TaxCalculationRequest true "Tax calculation data"
// @Success 200 {object} utils.Response{data=models.TaxCalculationResponse} "Tax calculation completed successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/taxes/calculate [post]
func (h *TaxHandler) CalculateTax(c *gin.Context) {
	var req models.TaxCalculationRequest
	if err := utils.ValidateRequest(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Validate request
	if len(req.Items) == 0 {
		utils.BadRequestResponse(c, "At least one item is required", nil)
		return
	}

	// Validate items
	for i, item := range req.Items {
		if item.ProductID == "" {
			utils.BadRequestResponse(c, "Product ID is required for all items", nil)
			return
		}
		if item.Quantity <= 0 {
			utils.BadRequestResponse(c, "Quantity must be greater than 0", nil)
			return
		}
		if item.UnitPrice < 0 {
			utils.BadRequestResponse(c, "Unit price cannot be negative", nil)
			return
		}
		if item.LineTotal < 0 {
			utils.BadRequestResponse(c, "Line total cannot be negative", nil)
			return
		}

		// Recalculate line total if it doesn't match
		calculatedTotal := item.UnitPrice * float64(item.Quantity)
		if item.LineTotal != calculatedTotal {
			req.Items[i].LineTotal = calculatedTotal
		}
	}

	taxCalculation, err := h.taxService.CalculateTax(&req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to calculate taxes", err)
		return
	}

	utils.OKResponse(c, "Tax calculation completed successfully", taxCalculation)
}

// GetTaxApplicationsBySale retrieves tax applications for a sale
// @Summary Get Tax Applications by Sale
// @Description Retrieve all tax applications for a specific sale
// @Tags Taxes
// @Produce json
// @Param saleID path string true "Sale ID" example(SALE_12345678)
// @Success 200 {object} utils.Response{data=[]models.TaxApplicationResponse} "Tax applications retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/taxes/applications/sale/{saleID} [get]
func (h *TaxHandler) GetTaxApplicationsBySale(c *gin.Context) {
	saleID := c.Param("saleID")
	if saleID == "" {
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	taxApps, err := h.taxService.GetTaxApplicationsBySale(saleID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve tax applications", err)
		return
	}

	utils.OKResponse(c, "Tax applications retrieved successfully", taxApps)
}

// GetTaxApplicationsByReturn retrieves tax applications for a return
// @Summary Get Tax Applications by Return
// @Description Retrieve all tax applications for a specific return
// @Tags Taxes
// @Produce json
// @Param returnID path string true "Return ID" example(RET_12345678)
// @Success 200 {object} utils.Response{data=[]models.TaxApplicationResponse} "Tax applications retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 401 {object} utils.ErrorResponseModel "Unauthorized"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/taxes/applications/return/{returnID} [get]
func (h *TaxHandler) GetTaxApplicationsByReturn(c *gin.Context) {
	returnID := c.Param("returnID")
	if returnID == "" {
		utils.BadRequestResponse(c, "Return ID is required", nil)
		return
	}

	taxApps, err := h.taxService.GetTaxApplicationsByReturn(returnID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve tax applications", err)
		return
	}

	utils.OKResponse(c, "Tax applications retrieved successfully", taxApps)
}

// GetTaxSummaryBySale retrieves tax summary for a sale
// @Summary Get Tax Summary by Sale
// @Description Retrieve tax summary for a specific sale
// @Tags Taxes
// @Produce json
// @Param saleID path string true "Sale ID" example(SALE_12345678)
// @Success 200 {object} utils.Response{data=models.TaxSummaryResponse} "Tax summary retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Tax summary not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/taxes/summary/sale/{saleID} [get]
func (h *TaxHandler) GetTaxSummaryBySale(c *gin.Context) {
	saleID := c.Param("saleID")
	if saleID == "" {
		utils.BadRequestResponse(c, "Sale ID is required", nil)
		return
	}

	taxSummary, err := h.taxService.GetTaxSummaryBySale(saleID)
	if err != nil {
		utils.NotFoundResponse(c, "Tax summary not found")
		return
	}

	utils.OKResponse(c, "Tax summary retrieved successfully", taxSummary)
}

// GetTaxSummaryByReturn retrieves tax summary for a return
// @Summary Get Tax Summary by Return
// @Description Retrieve tax summary for a specific return
// @Tags Taxes
// @Produce json
// @Param returnID path string true "Return ID" example(RET_12345678)
// @Success 200 {object} utils.Response{data=models.TaxSummaryResponse} "Tax summary retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Bad request"
// @Failure 404 {object} utils.ErrorResponseModel "Tax summary not found"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Security BearerAuth
// @Router /api/v1/taxes/summary/return/{returnID} [get]
func (h *TaxHandler) GetTaxSummaryByReturn(c *gin.Context) {
	returnID := c.Param("returnID")
	if returnID == "" {
		utils.BadRequestResponse(c, "Return ID is required", nil)
		return
	}

	taxSummary, err := h.taxService.GetTaxSummaryByReturn(returnID)
	if err != nil {
		utils.NotFoundResponse(c, "Tax summary not found")
		return
	}

	utils.OKResponse(c, "Tax summary retrieved successfully", taxSummary)
}
