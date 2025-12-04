package interfaces

import "kisanlink-erp/internal/database/models"

// ExportServiceInterface defines the contract for report export operations
type ExportServiceInterface interface {
	// ExportToXLSX exports report data to Excel format
	ExportToXLSX(reportType string, data *models.ReportResponse) ([]byte, error)

	// ExportToPDF exports report data to PDF format
	ExportToPDF(reportType string, data *models.ReportResponse) ([]byte, error)
}
