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

	// Extract all field names and create headers
	headers := []string{}
	fieldIndexes := []int{}
	fieldTypes := []string{} // Track field types for alignment
	recordType := firstRecord.Type()

	for i := 0; i < recordType.NumField(); i++ {
		field := recordType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			fieldName := strings.Split(jsonTag, ",")[0]
			headers = append(headers, strings.Title(strings.ReplaceAll(fieldName, "_", " ")))
			fieldIndexes = append(fieldIndexes, i)
			// Track if this is a numeric field for right-alignment
			fieldTypes = append(fieldTypes, field.Type.Kind().String())
		}
	}

	numCols := len(headers)
	pageWidth := 277.0 // A4 landscape width in mm minus margins

	// Determine font size and column widths based on number of columns
	var fontSize, headerFontSize float64
	var colWidths []float64

	if numCols <= 6 {
		// Standard layout for <= 6 columns
		headerFontSize = 9
		fontSize = 8
		colWidths = make([]float64, numCols)
		for i := range colWidths {
			colWidths[i] = pageWidth / float64(numCols)
		}
	} else if numCols <= 10 {
		// Compact layout for 7-10 columns
		headerFontSize = 8
		fontSize = 7
		colWidths = e.calculateDynamicColumnWidths(headers, fieldTypes, pageWidth)
	} else {
		// Dense layout for > 10 columns - use smaller font and smart widths
		headerFontSize = 7
		fontSize = 6
		colWidths = e.calculateDynamicColumnWidths(headers, fieldTypes, pageWidth)
	}

	// Write headers with calculated widths
	pdf.SetFont("Arial", "B", headerFontSize)
	pdf.SetFillColor(220, 220, 220)
	for i, header := range headers {
		truncatedHeader := e.truncateToWidth(header, colWidths[i], headerFontSize)
		pdf.CellFormat(colWidths[i], 7, truncatedHeader, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(7)

	// Write data rows
	pdf.SetFont("Arial", "", fontSize)
	for i := 0; i < recordsValue.Len(); i++ {
		record := recordsValue.Index(i)
		if record.Kind() == reflect.Ptr {
			record = record.Elem()
		}

		// Check if new page is needed
		if pdf.GetY() > 180 {
			pdf.AddPage()
			// Re-print headers
			pdf.SetFont("Arial", "B", headerFontSize)
			pdf.SetFillColor(220, 220, 220)
			for j, header := range headers {
				truncatedHeader := e.truncateToWidth(header, colWidths[j], headerFontSize)
				pdf.CellFormat(colWidths[j], 7, truncatedHeader, "1", 0, "C", true, 0, "")
			}
			pdf.Ln(7)
			pdf.SetFont("Arial", "", fontSize)
		}

		for j, fieldIdx := range fieldIndexes {
			fieldValue := record.Field(fieldIdx)
			value := e.formatPDFValueCompact(fieldValue, colWidths[j], fontSize)
			// Right-align numeric fields
			align := "L"
			if e.isNumericKind(fieldTypes[j]) {
				align = "R"
			}
			pdf.CellFormat(colWidths[j], 6, value, "1", 0, align, false, 0, "")
		}
		pdf.Ln(6)
	}

	return nil
}

// calculateDynamicColumnWidths calculates column widths based on content type
func (e *PDFExporter) calculateDynamicColumnWidths(headers []string, fieldTypes []string, totalWidth float64) []float64 {
	numCols := len(headers)
	widths := make([]float64, numCols)

	// Assign width weights based on field type and header name
	totalWeight := 0.0
	weights := make([]float64, numCols)

	for i, header := range headers {
		headerLower := strings.ToLower(header)
		fieldType := fieldTypes[i]

		// Assign weight based on content type
		var weight float64
		switch {
		case strings.Contains(headerLower, "id") && len(header) <= 5:
			weight = 1.0 // Short IDs
		case strings.Contains(headerLower, "id"):
			weight = 2.0 // Longer IDs
		case strings.Contains(headerLower, "name") || strings.Contains(headerLower, "company"):
			weight = 3.0 // Names need more space
		case strings.Contains(headerLower, "date"):
			weight = 1.5 // Dates are fixed width
		case strings.Contains(headerLower, "status"):
			weight = 1.2 // Status values are short
		case e.isNumericKind(fieldType):
			weight = 1.5 // Numbers are relatively narrow
		case strings.Contains(headerLower, "email"):
			weight = 2.5 // Emails can be long
		default:
			weight = 1.5 // Default weight
		}
		weights[i] = weight
		totalWeight += weight
	}

	// Calculate actual widths
	minWidth := 12.0 // Minimum column width in mm
	for i := range widths {
		widths[i] = (weights[i] / totalWeight) * totalWidth
		if widths[i] < minWidth {
			widths[i] = minWidth
		}
	}

	// Adjust to fit total width
	actualTotal := 0.0
	for _, w := range widths {
		actualTotal += w
	}
	scale := totalWidth / actualTotal
	for i := range widths {
		widths[i] *= scale
	}

	return widths
}

// isNumericKind checks if the field type is numeric
func (e *PDFExporter) isNumericKind(kindStr string) bool {
	switch kindStr {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		return true
	}
	return false
}

// truncateToWidth truncates text to fit within a given width
func (e *PDFExporter) truncateToWidth(text string, width float64, fontSize float64) string {
	// Approximate characters that fit (based on average character width)
	// Arial at font size N has roughly N * 0.5 mm per character average
	avgCharWidth := fontSize * 0.45
	maxChars := int(width / avgCharWidth)

	if maxChars < 3 {
		maxChars = 3
	}

	if len(text) <= maxChars {
		return text
	}
	if maxChars <= 3 {
		return text[:maxChars]
	}
	return text[:maxChars-2] + ".."
}

// formatPDFValueCompact formats field values for compact PDF display
func (e *PDFExporter) formatPDFValueCompact(fieldValue reflect.Value, colWidth float64, fontSize float64) string {
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

	var val string

	// Format based on type
	switch fieldValue.Kind() {
	case reflect.String:
		val = fieldValue.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val = fmt.Sprintf("%d", fieldValue.Int())
	case reflect.Float32, reflect.Float64:
		// Use compact number format for large values
		f := fieldValue.Float()
		if f >= 100000 {
			val = fmt.Sprintf("%.0f", f) // No decimals for large numbers
		} else if f >= 1000 {
			val = fmt.Sprintf("%.1f", f) // 1 decimal for medium numbers
		} else {
			val = fmt.Sprintf("%.2f", f)
		}
	case reflect.Bool:
		if fieldValue.Bool() {
			val = "Y"
		} else {
			val = "N"
		}
	default:
		val = fmt.Sprintf("%v", fieldValue.Interface())
	}

	// Truncate to fit column width
	return e.truncateToWidth(val, colWidth, fontSize)
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
