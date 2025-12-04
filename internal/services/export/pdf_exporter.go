package export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/interfaces"

	"github.com/jung-kurt/gofpdf"
	"go.uber.org/zap"
)

// PDFExporter handles PDF file export
type PDFExporter struct {
	logger interfaces.Logger
}

// NewPDFExporter creates a new PDF exporter
func NewPDFExporter(logger interfaces.Logger) *PDFExporter {
	return &PDFExporter{logger: logger}
}

// ExportToPDF exports report data to PDF format
func (e *PDFExporter) ExportToPDF(reportType string, data *models.ReportResponse) ([]byte, error) {
	e.logger.Info("Starting PDF export",
		zap.String("report_type", reportType))

	// Create PDF in landscape mode
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 10)
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, fmt.Sprintf("%s Report", strings.Title(reportType)))
	pdf.Ln(12)

	// Timestamp
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Generated: %s", data.GeneratedAt))
	pdf.Ln(10)

	// Summary section
	e.writeSummaryPDF(pdf, data.Summary)

	// Data table
	if err := e.writeDataTablePDF(pdf, data.Records); err != nil {
		e.logger.Error("Failed to write data table", zap.Error(err))
		return nil, err
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		e.logger.Error("Failed to write PDF to buffer", zap.Error(err))
		return nil, err
	}

	e.logger.Info("PDF export completed",
		zap.String("report_type", reportType),
		zap.Int("size_bytes", buf.Len()))

	return buf.Bytes(), nil
}

// writeSummaryPDF writes the summary section to PDF
func (e *PDFExporter) writeSummaryPDF(pdf *gofpdf.Fpdf, summary interface{}) {
	if summary == nil {
		return
	}

	// Summary header
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Summary")
	pdf.Ln(10)

	// Convert summary to map
	summaryMap := make(map[string]interface{})
	jsonBytes, _ := json.Marshal(summary)
	json.Unmarshal(jsonBytes, &summaryMap)

	// Write summary items
	pdf.SetFont("Arial", "", 10)
	for key, value := range summaryMap {
		displayKey := strings.Title(strings.ReplaceAll(key, "_", " "))
		pdf.Cell(80, 6, displayKey+":")
		pdf.Cell(0, 6, fmt.Sprintf("%v", value))
		pdf.Ln(7)
	}
	pdf.Ln(5)
}

// writeDataTablePDF writes the data table to PDF
func (e *PDFExporter) writeDataTablePDF(pdf *gofpdf.Fpdf, records interface{}) error {
	if records == nil {
		return nil
	}

	// Convert records to slice using reflection
	recordsValue := reflect.ValueOf(records)
	if recordsValue.Kind() != reflect.Slice {
		return fmt.Errorf("records must be a slice")
	}

	if recordsValue.Len() == 0 {
		pdf.SetFont("Arial", "I", 10)
		pdf.Cell(0, 6, "No records found")
		return nil
	}

	// Get first record to determine columns
	firstRecord := recordsValue.Index(0)
	if firstRecord.Kind() == reflect.Ptr {
		firstRecord = firstRecord.Elem()
	}

	// Extract field names and create headers (limit to first 6 fields for PDF width)
	headers := []string{}
	fieldIndexes := []int{}
	recordType := firstRecord.Type()

	maxFields := 6 // Limit columns to fit PDF width
	fieldCount := 0

	for i := 0; i < recordType.NumField() && fieldCount < maxFields; i++ {
		field := recordType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			fieldName := strings.Split(jsonTag, ",")[0]
			headers = append(headers, e.truncateHeader(strings.Title(strings.ReplaceAll(fieldName, "_", " "))))
			fieldIndexes = append(fieldIndexes, i)
			fieldCount++
		}
	}

	// Calculate column width
	pageWidth := 277.0 // A4 landscape width in mm minus margins
	colWidth := pageWidth / float64(len(headers))

	// Write headers
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(220, 220, 220)
	for _, header := range headers {
		pdf.CellFormat(colWidth, 8, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(8)

	// Write data rows
	pdf.SetFont("Arial", "", 8)
	for i := 0; i < recordsValue.Len(); i++ {
		record := recordsValue.Index(i)
		if record.Kind() == reflect.Ptr {
			record = record.Elem()
		}

		// Check if new page is needed
		if pdf.GetY() > 180 {
			pdf.AddPage()
			// Re-print headers
			pdf.SetFont("Arial", "B", 9)
			pdf.SetFillColor(220, 220, 220)
			for _, header := range headers {
				pdf.CellFormat(colWidth, 8, header, "1", 0, "C", true, 0, "")
			}
			pdf.Ln(8)
			pdf.SetFont("Arial", "", 8)
		}

		for _, fieldIdx := range fieldIndexes {
			fieldValue := record.Field(fieldIdx)
			value := e.formatPDFValue(fieldValue)
			pdf.CellFormat(colWidth, 7, value, "1", 0, "L", false, 0, "")
		}
		pdf.Ln(7)
	}

	return nil
}

// truncateHeader truncates header text to fit PDF column
func (e *PDFExporter) truncateHeader(text string) string {
	if len(text) > 15 {
		return text[:12] + "..."
	}
	return text
}

// formatPDFValue formats field values for PDF display
func (e *PDFExporter) formatPDFValue(fieldValue reflect.Value) string {
	if !fieldValue.IsValid() {
		return ""
	}

	// Handle pointer types
	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			return ""
		}
		fieldValue = fieldValue.Elem()
	}

	// Format based on type
	switch fieldValue.Kind() {
	case reflect.String:
		val := fieldValue.String()
		if len(val) > 30 {
			return val[:27] + "..."
		}
		return val
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", fieldValue.Int())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", fieldValue.Float())
	case reflect.Bool:
		if fieldValue.Bool() {
			return "Yes"
		}
		return "No"
	}

	val := fmt.Sprintf("%v", fieldValue.Interface())
	if len(val) > 30 {
		return val[:27] + "..."
	}
	return val
}
