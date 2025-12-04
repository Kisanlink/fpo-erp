package export

import (
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/interfaces"
)

// ExportService provides export functionality for reports
type ExportService struct {
	xlsxExporter *XLSXExporter
	pdfExporter  *PDFExporter
	logger       interfaces.Logger
}

// NewExportService creates a new export service
func NewExportService(xlsxExporter *XLSXExporter, pdfExporter *PDFExporter, logger interfaces.Logger) *ExportService {
	return &ExportService{
		xlsxExporter: xlsxExporter,
		pdfExporter:  pdfExporter,
		logger:       logger,
	}
}

// ExportToXLSX exports report data to Excel format
func (s *ExportService) ExportToXLSX(reportType string, data *models.ReportResponse) ([]byte, error) {
	return s.xlsxExporter.ExportToXLSX(reportType, data)
}

// ExportToPDF exports report data to PDF format
func (s *ExportService) ExportToPDF(reportType string, data *models.ReportResponse) ([]byte, error) {
	return s.pdfExporter.ExportToPDF(reportType, data)
}
