package export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/interfaces"

	excelize "github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

// XLSXExporter handles Excel file export
type XLSXExporter struct {
	logger interfaces.Logger
}

// NewXLSXExporter creates a new XLSX exporter
func NewXLSXExporter(logger interfaces.Logger) *XLSXExporter {
	return &XLSXExporter{logger: logger}
}

// ExportToXLSX exports report data to Excel format
func (e *XLSXExporter) ExportToXLSX(reportType string, data *models.ReportResponse) ([]byte, error) {
	e.logger.Info("Starting XLSX export",
		zap.String("report_type", reportType))

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			e.logger.Error("Failed to close Excel file", zap.Error(err))
		}
	}()

	// Create sheet with report type name
	sheetName := strings.Title(reportType)
	index, err := f.NewSheet(sheetName)
	if err != nil {
		e.logger.Error("Failed to create sheet", zap.Error(err))
		return nil, err
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// Define styles
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	summaryStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"E7E6E6"}, Pattern: 1},
	})

	// Write report title
	currentRow := 1
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("%s Report", strings.Title(reportType)))
	f.MergeCell(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("F%d", currentRow))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("F%d", currentRow), headerStyle)
	currentRow++

	// Write generated timestamp
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("Generated: %s", data.GeneratedAt))
	currentRow++
	currentRow++ // Empty row

	// Write summary section
	currentRow = e.writeSummary(f, sheetName, data.Summary, summaryStyle, currentRow)
	currentRow++ // Empty row

	// Write data headers and records
	if err := e.writeDataSection(f, sheetName, data.Records, headerStyle, currentRow); err != nil {
		e.logger.Error("Failed to write data section", zap.Error(err))
		return nil, err
	}

	// Auto-fit columns
	cols, _ := f.GetCols(sheetName)
	for idx := range cols {
		colName, _ := excelize.ColumnNumberToName(idx + 1)
		f.SetColWidth(sheetName, colName, colName, 15)
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		e.logger.Error("Failed to write Excel file to buffer", zap.Error(err))
		return nil, err
	}

	e.logger.Info("XLSX export completed",
		zap.String("report_type", reportType),
		zap.Int("size_bytes", buf.Len()))

	return buf.Bytes(), nil
}

// writeSummary writes the summary section to Excel
func (e *XLSXExporter) writeSummary(f *excelize.File, sheetName string, summary interface{}, style int, startRow int) int {
	if summary == nil {
		return startRow
	}

	// Write "Summary" header
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", startRow), "Summary")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", startRow), fmt.Sprintf("B%d", startRow), style)
	startRow++

	// Convert summary to map for iteration
	summaryMap := make(map[string]interface{})
	jsonBytes, _ := json.Marshal(summary)
	json.Unmarshal(jsonBytes, &summaryMap)

	// Write summary key-value pairs
	for key, value := range summaryMap {
		// Convert snake_case to Title Case
		displayKey := strings.Title(strings.ReplaceAll(key, "_", " "))
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", startRow), displayKey)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", startRow), value)
		startRow++
	}

	return startRow
}

// writeDataSection writes the data table to Excel
func (e *XLSXExporter) writeDataSection(f *excelize.File, sheetName string, records interface{}, headerStyle int, startRow int) error {
	if records == nil {
		return nil
	}

	// Convert records to slice using reflection
	recordsValue := reflect.ValueOf(records)
	if recordsValue.Kind() != reflect.Slice {
		return fmt.Errorf("records must be a slice")
	}

	if recordsValue.Len() == 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", startRow), "No records found")
		return nil
	}

	// Get first record to determine columns
	firstRecord := recordsValue.Index(0)
	if firstRecord.Kind() == reflect.Ptr {
		firstRecord = firstRecord.Elem()
	}

	// Extract field names and create headers
	headers := []string{}
	fieldIndexes := []int{}
	recordType := firstRecord.Type()

	for i := 0; i < recordType.NumField(); i++ {
		field := recordType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			// Remove ",omitempty" suffix if present
			fieldName := strings.Split(jsonTag, ",")[0]
			headers = append(headers, strings.Title(strings.ReplaceAll(fieldName, "_", " ")))
			fieldIndexes = append(fieldIndexes, i)
		}
	}

	// Write headers
	for idx, header := range headers {
		col, _ := excelize.ColumnNumberToName(idx + 1)
		cell := fmt.Sprintf("%s%d", col, startRow)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	startRow++

	// Write data rows
	for i := 0; i < recordsValue.Len(); i++ {
		record := recordsValue.Index(i)
		if record.Kind() == reflect.Ptr {
			record = record.Elem()
		}

		for colIdx, fieldIdx := range fieldIndexes {
			col, _ := excelize.ColumnNumberToName(colIdx + 1)
			cell := fmt.Sprintf("%s%d", col, startRow)

			fieldValue := record.Field(fieldIdx)
			value := e.formatFieldValue(fieldValue)
			f.SetCellValue(sheetName, cell, value)
		}
		startRow++
	}

	return nil
}

// formatFieldValue formats field values for Excel display
func (e *XLSXExporter) formatFieldValue(fieldValue reflect.Value) interface{} {
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

	// Handle different types
	switch fieldValue.Kind() {
	case reflect.String:
		return fieldValue.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fieldValue.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fieldValue.Uint()
	case reflect.Float32, reflect.Float64:
		return fieldValue.Float()
	case reflect.Bool:
		return fieldValue.Bool()
	case reflect.Struct:
		// Check if it's a time.Time
		if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
			t := fieldValue.Interface().(time.Time)
			return t.Format("2006-01-02 15:04:05")
		}
	}

	return fmt.Sprintf("%v", fieldValue.Interface())
}
