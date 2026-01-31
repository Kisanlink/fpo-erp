package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"
	svcInterfaces "kisanlink-erp/internal/services/interfaces"

	"github.com/jung-kurt/gofpdf"
)

// InvoiceService handles invoice PDF generation
type InvoiceService struct {
	salesRepo      *repositories.SalesRepository
	settingsRepo   *repositories.SettingsRepository
	inventoryRepo  *repositories.InventoryRepository
	variantRepo    *repositories.ProductVariantRepository
	productRepo    *repositories.ProductRepository
	warehouseRepo  *repositories.WarehouseRepository
	attachmentRepo *repositories.AttachmentRepository
	s3Service      *S3Service
	logger         interfaces.Logger
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(
	salesRepo *repositories.SalesRepository,
	settingsRepo *repositories.SettingsRepository,
	inventoryRepo *repositories.InventoryRepository,
	variantRepo *repositories.ProductVariantRepository,
	productRepo *repositories.ProductRepository,
	warehouseRepo *repositories.WarehouseRepository,
	attachmentRepo *repositories.AttachmentRepository,
	s3Service *S3Service,
	logger interfaces.Logger,
) *InvoiceService {
	return &InvoiceService{
		salesRepo:      salesRepo,
		settingsRepo:   settingsRepo,
		inventoryRepo:  inventoryRepo,
		variantRepo:    variantRepo,
		productRepo:    productRepo,
		warehouseRepo:  warehouseRepo,
		attachmentRepo: attachmentRepo,
		s3Service:      s3Service,
		logger:         logger,
	}
}

// Ensure InvoiceService implements InvoiceServiceInterface
var _ svcInterfaces.InvoiceServiceInterface = (*InvoiceService)(nil)

// CheckInvoiceRequirements checks if all required settings exist for invoice generation
func (s *InvoiceService) CheckInvoiceRequirements(ctx context.Context) (bool, []string, error) {
	s.logger.Info("Checking invoice requirements")

	requiredKeys := models.RequiredSettingsForInvoice()
	missing, err := s.settingsRepo.CheckRequiredSettings(requiredKeys)
	if err != nil {
		s.logger.Error("Failed to check invoice requirements", "error", err)
		return false, nil, err
	}

	if len(missing) > 0 {
		s.logger.Warn("Missing required settings for invoice", "missing", missing)
		return false, missing, nil
	}

	s.logger.Info("All invoice requirements satisfied")
	return true, nil, nil
}

// AddressJSON represents the JSON structure for addresses stored in settings
type AddressJSON struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	Pincode string `json:"pincode"`
	Country string `json:"country"`
}

// BankDetailsJSON represents the JSON structure for bank details stored in settings
type BankDetailsJSON struct {
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	IFSCCode      string `json:"ifsc_code"`
	BankName      string `json:"bank_name"`
	Branch        string `json:"branch"`
}

// InvoiceLineItem represents a line item for invoice rendering
type InvoiceLineItem struct {
	SNo        int
	ItemName   string
	HSNCode    string
	Units      string
	Quantity   int64
	Rate       float64
	NetValue   float64
	GSTRate    float64
	CGSTAmount float64
	SGSTAmount float64
	TotalValue float64
}

// GenerateInvoicePDF generates a PDF invoice for a sale
func (s *InvoiceService) GenerateInvoicePDF(ctx context.Context, saleID string) ([]byte, string, string, error) {
	s.logger.Info("Generating invoice PDF", "sale_id", saleID)

	// 1. Check requirements first
	ready, missing, err := s.CheckInvoiceRequirements(ctx)
	if err != nil {
		return nil, "", "", err
	}
	if !ready {
		return nil, "", "", errors.NewBadRequestError(fmt.Sprintf("Missing required settings: %v", missing))
	}

	// 2. Get the sale with items
	sale, err := s.salesRepo.GetSaleByID(saleID)
	if err != nil {
		s.logger.Error("Failed to get sale", "sale_id", saleID, "error", err)
		return nil, "", "", errors.NewNotFoundError("Sale")
	}

	// 3. Get all settings as a map
	settingsMap, err := s.settingsRepo.GetSettingsMap()
	if err != nil {
		s.logger.Error("Failed to get settings", "error", err)
		return nil, "", "", errors.NewInternalServerError("Failed to get settings")
	}

	// 4. Get header fields for dynamic rendering
	headerFields, err := s.settingsRepo.GetHeaderFields()
	if err != nil {
		s.logger.Error("Failed to get header fields", "error", err)
		return nil, "", "", errors.NewInternalServerError("Failed to get header fields")
	}

	// 5. Get warehouse info
	warehouse, err := s.warehouseRepo.GetByID(sale.WarehouseID)
	if err != nil {
		s.logger.Error("Failed to get warehouse", "warehouse_id", sale.WarehouseID, "error", err)
		return nil, "", "", errors.NewInternalServerError("Failed to get warehouse info")
	}

	// 6. Build line items with variant/product info
	lineItems, err := s.buildLineItems(sale)
	if err != nil {
		s.logger.Error("Failed to build line items", "error", err)
		return nil, "", "", err
	}

	// 7. Fetch logo if configured (don't fail if not available)
	var logoBytes []byte
	var logoContentType string
	if logoURL, ok := settingsMap[models.SettingKeyFPOLogoURL]; ok && logoURL != "" {
		logoBytes, logoContentType, _ = s.getLogoFromURL(ctx, logoURL)
		if logoBytes != nil {
			s.logger.Info("Logo loaded for invoice", "size_bytes", len(logoBytes))
		}
	}

	// 8. Generate the PDF
	pdfBytes, err := s.renderPDF(sale, settingsMap, headerFields, warehouse, lineItems, logoBytes, logoContentType)
	if err != nil {
		s.logger.Error("Failed to render PDF", "error", err)
		return nil, "", "", errors.NewInternalServerError("Failed to generate PDF")
	}

	filename := fmt.Sprintf("Invoice_%s.pdf", sale.InvoiceNumber)
	return pdfBytes, "application/pdf", filename, nil
}

// buildLineItems builds the line items for the invoice from sale items
func (s *InvoiceService) buildLineItems(sale *models.Sale) ([]InvoiceLineItem, error) {
	var items []InvoiceLineItem

	for i, saleItem := range sale.Items {
		// Get batch to get variant info
		batch, err := s.inventoryRepo.GetBatchByID(saleItem.BatchID)
		if err != nil {
			s.logger.Error("Failed to get batch", "batch_id", saleItem.BatchID, "error", err)
			return nil, errors.NewInternalServerError("Failed to get batch info")
		}

		// Get variant for HSN code and name
		variant, err := s.variantRepo.GetByID(batch.VariantID)
		if err != nil {
			s.logger.Error("Failed to get variant", "variant_id", batch.VariantID, "error", err)
			return nil, errors.NewInternalServerError("Failed to get variant info")
		}

		// Get product for full name
		product, err := s.productRepo.GetByID(variant.ProductID)
		if err != nil {
			s.logger.Error("Failed to get product", "product_id", variant.ProductID, "error", err)
			return nil, errors.NewInternalServerError("Failed to get product info")
		}

		// Build item name: Product Name + Variant quantity (e.g., "NFL KISAN HEXA 500 ML")
		itemName := product.Name
		if variant.Quantity != "" {
			itemName = fmt.Sprintf("%s %s", product.Name, variant.Quantity)
		}

		// Calculate net value (before tax) = Excl. Rate × Quantity
		netValue := saleItem.SellingPrice * float64(saleItem.Quantity)

		// Total value including tax (for this line)
		totalValue := netValue + saleItem.CGSTAmount + saleItem.SGSTAmount + saleItem.IGSTAmount

		items = append(items, InvoiceLineItem{
			SNo:        i + 1,
			ItemName:   itemName,
			HSNCode:    variant.HSNCode,
			Units:      variant.PackSize,
			Quantity:   saleItem.Quantity,
			Rate:       saleItem.SellingPrice, // Excl. Rate
			NetValue:   netValue,
			GSTRate:    variant.GSTRate,
			CGSTAmount: saleItem.CGSTAmount,
			SGSTAmount: saleItem.SGSTAmount,
			TotalValue: totalValue,
		})
	}

	return items, nil
}

// getLogoFromURL fetches logo bytes from attachment ID or direct URL
func (s *InvoiceService) getLogoFromURL(ctx context.Context, logoValue string) ([]byte, string, error) {
	// Check if logoValue is an attachment ID (starts with "ATCH_")
	if strings.HasPrefix(logoValue, "ATCH_") {
		// Look up attachment in database
		attachment, err := s.attachmentRepo.GetByID(logoValue)
		if err != nil {
			s.logger.Warn("Failed to get attachment", "id", logoValue, "error", err)
			return nil, "", nil // Silent fail, invoice still generates without logo
		}

		// Generate presigned URL from S3
		presignedURL, err := s.s3Service.GeneratePresignedURL(ctx, attachment.FilePath, 1*time.Hour)
		if err != nil {
			s.logger.Warn("Failed to generate presigned URL", "path", attachment.FilePath, "error", err)
			return nil, "", nil
		}

		// Fetch logo from presigned URL
		return s.fetchLogoFromHTTP(ctx, presignedURL)
	}

	// Fallback: treat as direct URL (backward compatibility)
	return s.fetchLogoFromHTTP(ctx, logoValue)
}

// fetchLogoFromHTTP fetches logo bytes from an HTTP URL
func (s *InvoiceService) fetchLogoFromHTTP(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		s.logger.Warn("Failed to create logo request", "url", url, "error", err)
		return nil, "", nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Warn("Failed to fetch logo from URL", "url", url, "error", err)
		return nil, "", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Warn("Logo URL returned non-200 status", "url", url, "status", resp.StatusCode)
		return nil, "", nil
	}

	logoBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Warn("Failed to read logo bytes", "error", err)
		return nil, "", nil
	}

	return logoBytes, resp.Header.Get("Content-Type"), nil
}

// renderPDF creates the actual PDF document
func (s *InvoiceService) renderPDF(
	sale *models.Sale,
	settings map[string]string,
	headerFields []models.Setting,
	warehouse *models.Warehouse,
	lineItems []InvoiceLineItem,
	logoBytes []byte,
	logoContentType string,
) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	// Parse addresses from settings
	var branchAddress AddressJSON
	if addr, ok := settings[models.SettingKeyFPOBranchAddress]; ok {
		json.Unmarshal([]byte(addr), &branchAddress)
	}

	var registeredAddress AddressJSON
	if addr, ok := settings[models.SettingKeyFPORegisteredAddress]; ok {
		json.Unmarshal([]byte(addr), &registeredAddress)
	}

	var bankDetails BankDetailsJSON
	if bank, ok := settings[models.SettingKeyFPOBankAccount]; ok {
		json.Unmarshal([]byte(bank), &bankDetails)
	}

	fpoName := settings[models.SettingKeyFPOName]

	// =============== HEADER SECTION ===============
	s.renderHeader(pdf, fpoName, branchAddress, registeredAddress, logoBytes, logoContentType)

	// =============== INVOICE DETAILS ROW ===============
	s.renderInvoiceDetails(pdf, sale, headerFields, warehouse)

	// =============== RECEIVER SECTION ===============
	s.renderReceiverSection(pdf, sale, warehouse)

	// =============== LINE ITEMS TABLE ===============
	totals := s.renderLineItemsTable(pdf, lineItems)

	// =============== SUMMARY SECTION ===============
	s.renderSummary(pdf, sale, totals)

	// =============== BANK DETAILS & FOOTER ===============
	s.renderFooter(pdf, fpoName, bankDetails)

	// Output to bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// renderHeader renders the company header section with optional logo
func (s *InvoiceService) renderHeader(pdf *gofpdf.Fpdf, fpoName string, branchAddr, registeredAddr AddressJSON, logoBytes []byte, logoContentType string) {
	startY := pdf.GetY()
	logoWidth := 0.0 // Track logo width for text offset

	// Render logo if available (left side)
	if len(logoBytes) > 0 {
		// Detect image type from magic bytes (more reliable than content type)
		imgType := detectImageType(logoBytes)
		if imgType == "" {
			// Fallback to content type if magic bytes don't match
			imgType = "PNG" // Default
			if strings.Contains(strings.ToLower(logoContentType), "jpeg") || strings.Contains(strings.ToLower(logoContentType), "jpg") {
				imgType = "JPG"
			} else if strings.Contains(strings.ToLower(logoContentType), "gif") {
				imgType = "GIF"
			}
		}

		// Register image from bytes
		reader := bytes.NewReader(logoBytes)
		pdf.RegisterImageOptionsReader("logo", gofpdf.ImageOptions{ImageType: imgType}, reader)

		// Place logo (25x25mm in top-left, auto-height to maintain aspect ratio)
		logoWidth = 25.0
		pdf.ImageOptions("logo", 10, startY, logoWidth, 0, false, gofpdf.ImageOptions{ImageType: imgType}, 0, "")
	}

	// Company name (centered or offset if logo exists, bold)
	pdf.SetFont("Arial", "B", 16)
	if logoWidth > 0 {
		// Offset text to accommodate logo
		textX := 10 + logoWidth + 5 // logo position + logo width + margin
		textWidth := 190 - logoWidth - 5
		pdf.SetXY(textX, startY)
		pdf.CellFormat(textWidth, 8, strings.ToUpper(fpoName), "", 1, "C", false, 0, "")
	} else {
		pdf.CellFormat(190, 8, strings.ToUpper(fpoName), "", 1, "C", false, 0, "")
	}

	// Branch Office
	pdf.SetFont("Arial", "", 9)
	branchLine := fmt.Sprintf("Branch Office: %s, %s, %s - %s",
		branchAddr.Street, branchAddr.City, branchAddr.State, branchAddr.Pincode)
	if logoWidth > 0 {
		textX := 10 + logoWidth + 5
		textWidth := 190 - logoWidth - 5
		pdf.SetX(textX)
		pdf.CellFormat(textWidth, 5, branchLine, "", 1, "C", false, 0, "")
	} else {
		pdf.CellFormat(190, 5, branchLine, "", 1, "C", false, 0, "")
	}

	// Registered Office (if available)
	if registeredAddr.Street != "" {
		regLine := fmt.Sprintf("Registered Office: %s, %s, %s - %s",
			registeredAddr.Street, registeredAddr.City, registeredAddr.State, registeredAddr.Pincode)
		if logoWidth > 0 {
			textX := 10 + logoWidth + 5
			textWidth := 190 - logoWidth - 5
			pdf.SetX(textX)
			pdf.CellFormat(textWidth, 5, regLine, "", 1, "C", false, 0, "")
		} else {
			pdf.CellFormat(190, 5, regLine, "", 1, "C", false, 0, "")
		}
	}

	// Ensure we move past the logo if it's taller than the text
	if logoWidth > 0 {
		// Logo height is approximately 25mm, ensure Y position is past it
		logoBottomY := startY + 25
		currentY := pdf.GetY()
		if currentY < logoBottomY {
			pdf.SetY(logoBottomY)
		}
	}

	pdf.Ln(3)
}

// detectImageType detects image format from magic bytes (more reliable than content type)
func detectImageType(data []byte) string {
	if len(data) < 8 {
		return ""
	}
	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "PNG"
	}
	// JPEG: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "JPG"
	}
	// GIF: GIF87a or GIF89a
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "GIF"
	}
	return ""
}

// renderInvoiceDetails renders the invoice number, date, and header fields
func (s *InvoiceService) renderInvoiceDetails(
	pdf *gofpdf.Fpdf,
	sale *models.Sale,
	headerFields []models.Setting,
	warehouse *models.Warehouse,
) {
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.3)
	pdf.SetFont("Arial", "", 10)

	leftX := 10.0
	rightX := 105.0
	rowH := 6.0
	y := pdf.GetY()

	// Safe header accessor
	getHeader := func(idx int) string {
		if idx >= 0 && idx < len(headerFields) {
			f := headerFields[idx]
			if f.DisplayLabel != nil && *f.DisplayLabel != "" {
				return fmt.Sprintf("%s: %s", *f.DisplayLabel, f.Value)
			}
			return f.Value
		}
		return ""
	}

	// -------- Row 1: Invoice No | GSTIN --------
	pdf.SetXY(leftX, y)
	pdf.CellFormat(95, rowH, fmt.Sprintf("Invoice No: %s", sale.InvoiceNumber), "LT", 0, "L", false, 0, "")
	pdf.SetXY(rightX, y)
	pdf.CellFormat(95, rowH, getHeader(0), "RT", 0, "L", false, 0, "")
	y += rowH

	// -------- Row 2: Invoice Date | Fert. Lic. --------
	pdf.SetXY(leftX, y)
	pdf.CellFormat(
		95,
		rowH,
		fmt.Sprintf("Invoice Date: %s", sale.SaleDate.Format("02/01/2006")),
		"L",
		0,
		"L",
		false,
		0,
		"",
	)
	pdf.SetXY(rightX, y)
	pdf.CellFormat(95, rowH, getHeader(1), "R", 0, "L", false, 0, "")
	y += rowH

	// -------- Row 3+: State | Pest & Seeds Lic (NO GAP, Pest on same row as State) --------
	state := "Telangana"
	if warehouse.State != nil && *warehouse.State != "" {
		state = *warehouse.State
	}

	var pestLic, seedsLic string
	if len(headerFields) > 2 {
		val := headerFields[2].Value

		if strings.Contains(val, "Pest. Lic.:") {
			p := strings.Split(val, "Pest. Lic.:")
			if len(p) > 1 {
				// take up to Seeds Lic if present
				pestLic = strings.TrimSpace(strings.Split(p[1], "Seeds Lic.:")[0])
			}
		}
		if strings.Contains(val, "Seeds Lic.:") {
			spl := strings.Split(val, "Seeds Lic.:")
			if len(spl) > 1 {
				seedsLic = strings.TrimSpace(spl[1])
			}
		}
	}

	// We always want State cell to cover exactly 2 rows visually (Row3 + Row4),
	// because Pest is Row3 and Seeds is Row4 (even if Seeds is empty we still close borders cleanly).
	rows := 2
	pdf.SetXY(leftX, y)
	pdf.CellFormat(95, float64(rows)*rowH, fmt.Sprintf("State: %s", state), "LB", 0, "L", false, 0, "")
	pdf.SetXY(rightX, y)
	rightBorderRow3 := "R"
	pdf.CellFormat(95, rowH, "Pest. Lic.: "+pestLic, rightBorderRow3, 0, "L", false, 0, "")
	pdf.SetXY(rightX, y+rowH)
	pdf.CellFormat(95, rowH, "Seeds Lic.: "+seedsLic, "RB", 0, "L", false, 0, "")
	y += float64(rows) * rowH
	pdf.SetY(y)
	pdf.Ln(5)
}

// renderReceiverSection renders the billed-to and place-of-supply sections side by side
func (s *InvoiceService) renderReceiverSection(pdf *gofpdf.Fpdf, sale *models.Sale, warehouse *models.Warehouse) {
	y := pdf.GetY()
	leftX := 10.0
	rightX := 105.0
	width := 95.0

	// Header row for both columns
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(leftX, y)
	pdf.CellFormat(width, 6, "Details of Receiver (Billed To)", "LTR", 0, "L", false, 0, "")
	pdf.SetXY(rightX, y)
	pdf.CellFormat(width, 6, "Place of Supply / Shipped To", "LTR", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 9)
	pdf.SetY(y + 6)

	// ----- Left column: Billed To -----
	leftY := pdf.GetY()

	billedName := "Walk-in Customer"
	if sale.CustomerName != nil && *sale.CustomerName != "" {
		billedName = *sale.CustomerName
	}

	phoneText := "N/A"
	if sale.CustomerPhone != nil && *sale.CustomerPhone != "" {
		phoneText = *sale.CustomerPhone
	}

	memberStatus := "Non-Member"
	if sale.IsOrgMember {
		memberStatus = "Member"
	}

	leftLines := fmt.Sprintf("%s\nPhone: %s\nCustomer Type: %s",
		billedName,
		phoneText,
		memberStatus,
	)

	pdf.SetXY(leftX, leftY)
	pdf.MultiCell(width, 4.5, leftLines, "L", "L", false)
	leftBottomY := pdf.GetY()

	// ----- Right column: Place of Supply / Shipped To -----
	// Get place of supply from warehouse state or default
	placeOfSupply := "Telangana" // Default fallback
	if warehouse.State != nil && *warehouse.State != "" {
		placeOfSupply = *warehouse.State
	}
	shipToName := billedName
	shipToAddress := "" // Customer address - can be extended when available in sale model
	shipToGSTIN := ""   // Customer GSTIN - can be extended when available in sale model

	rightLines := fmt.Sprintf("Place of Supply: %s\n%s", placeOfSupply, shipToName)
	if shipToAddress != "" {
		rightLines = fmt.Sprintf("%s\n%s", rightLines, shipToAddress)
	}
	if shipToGSTIN != "" {
		rightLines = fmt.Sprintf("%s\nGSTIN: %s", rightLines, shipToGSTIN)
	}

	pdf.SetXY(rightX, leftY)
	pdf.MultiCell(width, 4.5, rightLines, "R", "L", false)
	rightBottomY := pdf.GetY()

	// Determine bottom of the two-column block
	bottomY := leftBottomY
	if rightBottomY > bottomY {
		bottomY = rightBottomY
	}

	// Draw rectangles around both columns to mimic two boxes
	pdf.Rect(leftX, y+6, width, bottomY-(y+6), "")
	pdf.Rect(rightX, y+6, width, bottomY-(y+6), "")

	pdf.SetY(bottomY)
	pdf.Ln(4)
}

// InvoiceTotals holds calculated totals for the invoice
type InvoiceTotals struct {
	TotalNetValue float64
	TotalCGST     float64
	TotalSGST     float64
	TotalAmount   float64
}

// renderLineItemsTable renders the items table
func (s *InvoiceService) renderLineItemsTable(pdf *gofpdf.Fpdf, items []InvoiceLineItem) InvoiceTotals {
	// Table header
	pdf.SetFont("Arial", "B", 8)
	pdf.SetFillColor(240, 240, 240)

	// Column widths (sum should fit within ~190mm)
	colWidths := []float64{
		8,  // S.No
		45, // Item Name
		18, // HSN Code
		12, // Units
		8,  // Qty
		18, // Excl. Rate
		18, // Net Value
		10, // GST%
		13, // CGST
		13, // SGST
		18, // Total Value
	}

	headers := []string{
		"S.No",
		"Item Name",
		"HSN Code",
		"Units",
		"Qty",
		"Excl. Rate",
		"Net Value",
		"GST%",
		"CGST",
		"SGST",
		"Total Value",
	}

	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 7, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Arial", "", 8)
	pdf.SetFillColor(255, 255, 255)

	var totals InvoiceTotals

	for _, item := range items {
		pdf.CellFormat(colWidths[0], 6, fmt.Sprintf("%d", item.SNo), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[1], 6, truncateString(item.ItemName, 30), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[2], 6, item.HSNCode, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[3], 6, truncateString(item.Units, 10), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[4], 6, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[5], 6, fmt.Sprintf("%.2f", item.Rate), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths[6], 6, fmt.Sprintf("%.2f", item.NetValue), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths[7], 6, fmt.Sprintf("%.2f", item.GSTRate), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[8], 6, fmt.Sprintf("%.2f", item.CGSTAmount), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths[9], 6, fmt.Sprintf("%.2f", item.SGSTAmount), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths[10], 6, fmt.Sprintf("%.2f", item.TotalValue), "1", 0, "R", false, 0, "")
		pdf.Ln(-1)

		totals.TotalNetValue += item.NetValue
		totals.TotalCGST += item.CGSTAmount
		totals.TotalSGST += item.SGSTAmount
		totals.TotalAmount += item.TotalValue
	}

	pdf.Ln(4)
	return totals
}

// renderSummary renders the totals summary
func (s *InvoiceService) renderSummary(pdf *gofpdf.Fpdf, sale *models.Sale, totals InvoiceTotals) {
	y := pdf.GetY()

	// Left side - Amount in words
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(10, y)
	amountInWords := numberToWords(int64(sale.TotalAmount))
	pdf.MultiCell(100, 5, fmt.Sprintf("Grand Total: %s Only", amountInWords), "", "L", false)

	// Right side - Summary table
	pdf.SetXY(120, y)
	pdf.SetFont("Arial", "", 10)

	// Total Amount Before Tax
	pdf.CellFormat(50, 6, "Total Amount Before Tax", "1", 0, "L", false, 0, "")
	pdf.CellFormat(30, 6, fmt.Sprintf("%.2f", totals.TotalNetValue), "1", 1, "R", false, 0, "")
	pdf.SetX(120)

	// Add CGST
	pdf.CellFormat(50, 6, "Add: CGST", "1", 0, "L", false, 0, "")
	pdf.CellFormat(30, 6, fmt.Sprintf("%.2f", totals.TotalCGST), "1", 1, "R", false, 0, "")
	pdf.SetX(120)

	// Add SGST
	pdf.CellFormat(50, 6, "Add: SGST", "1", 0, "L", false, 0, "")
	pdf.CellFormat(30, 6, fmt.Sprintf("%.2f", totals.TotalSGST), "1", 1, "R", false, 0, "")
	pdf.SetX(120)

	// Total Amount
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(50, 6, "Total Amount", "1", 0, "L", false, 0, "")
	pdf.CellFormat(30, 6, fmt.Sprintf("%.2f", sale.TotalAmount), "1", 1, "R", false, 0, "")

	pdf.Ln(5)
}

// renderFooter renders bank details and terms
func (s *InvoiceService) renderFooter(pdf *gofpdf.Fpdf, fpoName string, bank BankDetailsJSON) {
	y := pdf.GetY()

	// Left side - Bank Details
	pdf.SetFont("Arial", "B", 9)
	pdf.SetXY(10, y)
	pdf.CellFormat(100, 5, "Virtual Payment Details:", "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(100, 5, fpoName, "", 1, "L", false, 0, "")
	if bank.AccountNumber != "" {
		pdf.CellFormat(100, 5, fmt.Sprintf("Account Number: %s", bank.AccountNumber), "", 1, "L", false, 0, "")
		pdf.CellFormat(100, 5, fmt.Sprintf("IFSC: %s", bank.IFSCCode), "", 1, "L", false, 0, "")
		pdf.CellFormat(100, 5, fmt.Sprintf("Branch: %s", bank.Branch), "", 1, "L", false, 0, "")
	}

	// Right side - Authorized Signatory
	pdf.SetXY(130, y)
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(70, 5, fmt.Sprintf("For %s", fpoName), "", 1, "R", false, 0, "")
	pdf.SetXY(130, y+25)
	pdf.CellFormat(70, 5, "Authorised Signatory", "", 1, "R", false, 0, "")

	// Terms and Conditions
	pdf.Ln(15)
	pdf.SetFont("Arial", "", 7)
	terms := []string{
		"1. We will not be liable for any demurrage, losses etc., for delay in clearing of goods.",
		"2. Goods once dispatched will not be accepted back or exchanged.",
		"3. Every Care is taken in the packing and forwarding of goods we are not responsible for any breakages and losses in transit.",
		"4. Claims should be filed with carriers. Interest at 24% p.a. will be charged on all outstanding of more than one month.",
		"5. Any dispute is subject only to a Court in local Jurisdiction.",
		"6. Any claim or report after a period of 15 days will not be entertained.",
	}

	for _, term := range terms {
		pdf.CellFormat(190, 4, term, "", 1, "L", false, 0, "")
	}
}

// Helper functions

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// numberToWords converts a number to words (simplified version for Indian currency)
func numberToWords(n int64) string {
	if n == 0 {
		return "Zero"
	}

	ones := []string{"", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine",
		"Ten", "Eleven", "Twelve", "Thirteen", "Fourteen", "Fifteen", "Sixteen", "Seventeen", "Eighteen", "Nineteen"}
	tens := []string{"", "", "Twenty", "Thirty", "Forty", "Fifty", "Sixty", "Seventy", "Eighty", "Ninety"}

	var result string

	// Crores (10,000,000)
	if n >= 10000000 {
		crores := n / 10000000
		result += numberToWordsHelper(crores, ones, tens) + " Crore "
		n %= 10000000
	}

	// Lakhs (100,000)
	if n >= 100000 {
		lakhs := n / 100000
		result += numberToWordsHelper(lakhs, ones, tens) + " Lakh "
		n %= 100000
	}

	// Thousands
	if n >= 1000 {
		thousands := n / 1000
		result += numberToWordsHelper(thousands, ones, tens) + " Thousand "
		n %= 1000
	}

	// Hundreds
	if n >= 100 {
		hundreds := n / 100
		result += ones[hundreds] + " Hundred "
		n %= 100
	}

	// Remaining
	if n > 0 {
		result += numberToWordsHelper(n, ones, tens)
	}

	return strings.TrimSpace(result)
}

func numberToWordsHelper(n int64, ones, tens []string) string {
	if n < 20 {
		return ones[n]
	}
	result := tens[n/10]
	if n%10 > 0 {
		result += " " + ones[n%10]
	}
	return result
}
