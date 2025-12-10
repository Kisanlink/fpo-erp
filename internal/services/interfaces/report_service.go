package interfaces

import "kisanlink-erp/internal/database/models"

// ReportServiceInterface defines the contract for report generation operations
type ReportServiceInterface interface {
	// GenerateProductReport generates product master report
	GenerateProductReport(filter *models.ProductReportFilter) (*models.ReportResponse, error)

	// GenerateVendorReport generates vendor master report
	GenerateVendorReport(filter *models.VendorReportFilter) (*models.ReportResponse, error)

	// GenerateCustomerReport generates customer report
	GenerateCustomerReport(filter *models.CustomerReportFilter) (*models.ReportResponse, error)

	// GenerateInventoryReport generates inventory report
	GenerateInventoryReport(filter *models.InventoryReportFilter) (*models.ReportResponse, error)

	// GeneratePurchaseReport generates purchase orders report
	GeneratePurchaseReport(filter *models.PurchaseReportFilter) (*models.ReportResponse, error)

	// GenerateSalesReport generates sales report
	GenerateSalesReport(filter *models.SalesReportFilter) (*models.ReportResponse, error)

	// GenerateReturnsReport generates returns report
	GenerateReturnsReport(filter *models.ReturnsReportFilter) (*models.ReportResponse, error)

	// GenerateGRNReport generates goods receipt note report
	GenerateGRNReport(filter *models.GRNReportFilter) (*models.ReportResponse, error)
}
