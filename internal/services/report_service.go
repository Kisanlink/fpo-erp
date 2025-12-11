package services

import (
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ReportService implements report generation functionality
type ReportService struct {
	db               *gorm.DB
	productRepo      *repositories.ProductRepository
	variantRepo      *repositories.ProductVariantRepository
	collaboratorRepo *repositories.CollaboratorRepository
	inventoryRepo    *repositories.InventoryRepository
	purchaseRepo     *repositories.PurchaseOrderRepository
	salesRepo        *repositories.SalesRepository
	returnsRepo      *repositories.ReturnsRepository
	warehouseRepo    *repositories.WarehouseRepository
	logger           interfaces.Logger
}

// NewReportService creates a new report service
func NewReportService(
	db *gorm.DB,
	productRepo *repositories.ProductRepository,
	variantRepo *repositories.ProductVariantRepository,
	collaboratorRepo *repositories.CollaboratorRepository,
	inventoryRepo *repositories.InventoryRepository,
	purchaseRepo *repositories.PurchaseOrderRepository,
	salesRepo *repositories.SalesRepository,
	returnsRepo *repositories.ReturnsRepository,
	warehouseRepo *repositories.WarehouseRepository,
	logger interfaces.Logger,
) *ReportService {
	return &ReportService{
		db:               db,
		productRepo:      productRepo,
		variantRepo:      variantRepo,
		collaboratorRepo: collaboratorRepo,
		inventoryRepo:    inventoryRepo,
		purchaseRepo:     purchaseRepo,
		salesRepo:        salesRepo,
		returnsRepo:      returnsRepo,
		warehouseRepo:    warehouseRepo,
		logger:           logger,
	}
}

// GenerateProductReport generates product master report
func (s *ReportService) GenerateProductReport(filter *models.ProductReportFilter) (*models.ReportResponse, error) {
	s.logger.Info("Generating product report", zap.Any("filter", filter))

	// Apply defaults
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.SortBy == "" {
		filter.SortBy = "created_at"
	}

	// Build query
	query := s.db.Model(&models.Product{}).Select(`
		products.id,
		products.name,
		products.description,
		products.external_id,
		COUNT(DISTINCT product_variants.id) as variant_count,
		COALESCE(SUM(inventory_batches.total_quantity), 0) as total_stock,
		products.created_at,
		products.updated_at
	`).
		Joins("LEFT JOIN product_variants ON product_variants.product_id = products.id").
		Joins("LEFT JOIN inventory_batches ON inventory_batches.variant_id = product_variants.id").
		Group("products.id")

	// Apply filters
	builder := utils.NewReportQueryBuilder(query)
	if filter.Search != "" {
		builder.ApplySearch([]string{"products.name", "products.description"}, filter.Search)
	}
	if filter.HasVariants != nil {
		if *filter.HasVariants {
			builder.Build().Having("COUNT(DISTINCT product_variants.id) > 0")
		} else {
			builder.Build().Having("COUNT(DISTINCT product_variants.id) = 0")
		}
	}

	// Get total count
	var total int64
	if err := s.db.Model(&models.Product{}).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count products", zap.Error(err))
		return nil, err
	}

	// Apply pagination and sorting
	builder.ApplySorting(filter.SortBy, filter.SortOrder, "created_at DESC")
	builder.ApplyPagination(filter.Limit, filter.Offset)

	// Execute query - initialize as empty slice to return [] instead of null in JSON
	records := make([]models.ProductReportRecord, 0)
	if err := builder.Build().Scan(&records).Error; err != nil {
		s.logger.Error("Failed to fetch product records", zap.Error(err))
		return nil, err
	}

	// Calculate summary
	var summary models.ProductReportSummary
	s.db.Model(&models.Product{}).Count(&summary.TotalProducts)
	s.db.Model(&models.ProductVariant{}).Count(&summary.TotalVariants)
	s.db.Model(&models.Product{}).
		Joins("JOIN product_variants ON product_variants.product_id = products.id").
		Joins("JOIN inventory_batches ON inventory_batches.variant_id = product_variants.id").
		Where("inventory_batches.total_quantity > 0").
		Distinct("products.id").
		Count(&summary.ProductsWithStock)

	// Format timestamps
	for i := range records {
		if t, err := time.Parse(time.RFC3339, records[i].CreatedAt); err == nil {
			records[i].CreatedAt = t.Format("2006-01-02 15:04:05")
		}
		if t, err := time.Parse(time.RFC3339, records[i].UpdatedAt); err == nil {
			records[i].UpdatedAt = t.Format("2006-01-02 15:04:05")
		}
	}

	return &models.ReportResponse{
		ReportType:  "product_master",
		GeneratedAt: time.Now().Format(time.RFC3339),
		Summary:     summary,
		Records:     records,
		Pagination: &models.PaginationInfo{
			Total:   total,
			Limit:   filter.Limit,
			Offset:  filter.Offset,
			HasMore: filter.Offset+filter.Limit < int(total),
		},
	}, nil
}

// GenerateVendorReport generates vendor master report
func (s *ReportService) GenerateVendorReport(filter *models.VendorReportFilter) (*models.ReportResponse, error) {
	s.logger.Info("Generating vendor report", zap.Any("filter", filter))

	// Apply defaults
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	// Build query - product_count uses product_variants.collaborator_ids JSON array
	// Use subquery for total_po_value to avoid JOIN multiplication issues
	query := s.db.Model(&models.Collaborator{}).Select(`
		collaborators.id,
		collaborators.company_name,
		collaborators.contact_person,
		collaborators.contact_number,
		collaborators.email,
		collaborators.gst_number,
		collaborators.pan_number,
		collaborators.bank_account_no,
		collaborators.bank_ifsc,
		collaborators.bank_name,
		collaborators.is_active,
		(SELECT COUNT(DISTINCT pv.id) FROM product_variants pv WHERE pv.collaborator_ids::jsonb @> to_jsonb(collaborators.id::text)) as product_count,
		COALESCE((SELECT SUM(po.total_amount) FROM purchase_orders po WHERE po.collaborator_id = collaborators.id), 0) as total_po_value,
		collaborators.created_at,
		collaborators.updated_at
	`)

	// Apply filters
	builder := utils.NewReportQueryBuilder(query)
	if filter.Search != "" {
		builder.ApplySearch([]string{"collaborators.company_name", "collaborators.contact_person", "collaborators.gst_number"}, filter.Search)
	}
	builder.ApplyBoolFilter("collaborators.is_active", filter.IsActive)
	if filter.HasGST != nil && *filter.HasGST {
		builder.Build().Where("collaborators.gst_number != ''")
	}

	// Get total count
	var total int64
	if err := s.db.Model(&models.Collaborator{}).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count vendors", zap.Error(err))
		return nil, err
	}

	// Apply pagination
	builder.ApplySorting(filter.SortBy, filter.SortOrder, "company_name ASC")
	builder.ApplyPagination(filter.Limit, filter.Offset)

	// Execute query - initialize as empty slice to return [] instead of null in JSON
	records := make([]models.VendorReportRecord, 0)
	if err := builder.Build().Scan(&records).Error; err != nil {
		s.logger.Error("Failed to fetch vendor records", zap.Error(err))
		return nil, err
	}

	// Calculate summary
	var summary models.VendorReportSummary
	s.db.Model(&models.Collaborator{}).Count(&summary.TotalVendors)
	s.db.Model(&models.Collaborator{}).Where("is_active = ?", true).Count(&summary.ActiveVendors)
	s.db.Model(&models.Collaborator{}).Where("gst_number != ''").Count(&summary.VendorsWithGST)
	s.db.Model(&models.PurchaseOrder{}).Select("COALESCE(SUM(total_amount), 0)").Scan(&summary.TotalPOValue)

	return &models.ReportResponse{
		ReportType:  "vendor_master",
		GeneratedAt: time.Now().Format(time.RFC3339),
		Summary:     summary,
		Records:     records,
		Pagination: &models.PaginationInfo{
			Total:   total,
			Limit:   filter.Limit,
			Offset:  filter.Offset,
			HasMore: filter.Offset+filter.Limit < int(total),
		},
	}, nil
}

// GenerateCustomerReport generates customer report
func (s *ReportService) GenerateCustomerReport(filter *models.CustomerReportFilter) (*models.ReportResponse, error) {
	s.logger.Info("Generating customer report", zap.Any("filter", filter))

	// Apply defaults
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	// Build query - uses customer_phone as the customer identifier
	query := s.db.Model(&models.Sale{}).Select(`
		sales.customer_phone,
		sales.customer_name,
		COUNT(*) as total_purchases,
		SUM(sales.total_amount) as total_amount,
		MAX(sales.sale_date) as last_purchase_date,
		sales.warehouse_id,
		warehouses.name as warehouse_name,
		COUNT(*) as purchase_count
	`).
		Joins("LEFT JOIN warehouses ON warehouses.id = sales.warehouse_id").
		Where("sales.customer_phone IS NOT NULL AND sales.customer_phone != ''").
		Where("sales.status != 'cancelled'").
		Group("sales.customer_phone, sales.customer_name, sales.warehouse_id, warehouses.name")

	// Apply filters
	builder := utils.NewReportQueryBuilder(query)
	if filter.Search != "" {
		builder.Build().Where("(sales.customer_phone ILIKE ? OR sales.customer_name ILIKE ?)", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}
	builder.ApplyStringFilter("sales.warehouse_id", filter.WarehouseID)
	builder.ApplyDateFilter("sales.sale_date", filter.StartDate, filter.EndDate)

	// Get total count
	var total int64
	countQuery := s.db.Model(&models.Sale{}).
		Where("customer_phone IS NOT NULL AND customer_phone != ''").
		Where("status != 'cancelled'").
		Distinct("customer_phone")
	countQuery.Count(&total)

	// Apply pagination
	builder.ApplySorting(filter.SortBy, filter.SortOrder, "total_amount DESC")
	builder.ApplyPagination(filter.Limit, filter.Offset)

	// Execute query - initialize as empty slice to return [] instead of null in JSON
	records := make([]models.CustomerReportRecord, 0)
	if err := builder.Build().Scan(&records).Error; err != nil {
		s.logger.Error("Failed to fetch customer records", zap.Error(err))
		return nil, err
	}

	// Calculate summary
	var summary models.CustomerReportSummary
	s.db.Model(&models.Sale{}).
		Where("customer_phone IS NOT NULL AND customer_phone != ''").
		Where("status != 'cancelled'").
		Distinct("customer_phone").
		Count(&summary.TotalCustomers)
	s.db.Model(&models.Sale{}).
		Where("customer_phone IS NOT NULL AND customer_phone != ''").
		Where("status != 'cancelled'").
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&summary.TotalRevenue)
	if summary.TotalCustomers > 0 {
		summary.AveragePurchaseValue = summary.TotalRevenue / float64(summary.TotalCustomers)
	}

	return &models.ReportResponse{
		ReportType:  "customers",
		GeneratedAt: time.Now().Format(time.RFC3339),
		Summary:     summary,
		Records:     records,
		Pagination: &models.PaginationInfo{
			Total:   total,
			Limit:   filter.Limit,
			Offset:  filter.Offset,
			HasMore: filter.Offset+filter.Limit < int(total),
		},
	}, nil
}

// GenerateInventoryReport generates inventory report
func (s *ReportService) GenerateInventoryReport(filter *models.InventoryReportFilter) (*models.ReportResponse, error) {
	s.logger.Info("Generating inventory report", zap.Any("filter", filter))

	// Apply defaults
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	// Build query - tax rates are on product_variants, not inventory_batches
	query := s.db.Model(&models.InventoryBatch{}).Select(`
		inventory_batches.id as batch_id,
		inventory_batches.warehouse_id,
		warehouses.name as warehouse_name,
		inventory_batches.variant_id,
		products.name as product_name,
		product_variants.sku as variant_sku,
		inventory_batches.total_quantity,
		inventory_batches.reserved_quantity,
		(inventory_batches.total_quantity - COALESCE(inventory_batches.reserved_quantity, 0)) as available_quantity,
		inventory_batches.cost_price,
		(inventory_batches.total_quantity * inventory_batches.cost_price) as total_value,
		inventory_batches.expiry_date,
		(inventory_batches.expiry_date - CURRENT_DATE) as days_to_expiry,
		(CURRENT_DATE - inventory_batches.created_at::date) as days_on_shelf,
		grn_items.batch_number,
		goods_receipt_notes.received_date,
		goods_receipt_notes.quality_status,
		goods_receipt_notes.id as source_grn_id,
		CASE WHEN inventory_batches.expiry_date <= CURRENT_DATE + INTERVAL '30 days' THEN true ELSE false END as is_expiring_soon,
		CASE WHEN inventory_batches.expiry_date < CURRENT_DATE THEN true ELSE false END as is_expired,
		CASE
			WHEN inventory_batches.expiry_date < CURRENT_DATE THEN 'expired'
			WHEN (inventory_batches.expiry_date - CURRENT_DATE) < 7 THEN 'critical'
			WHEN (inventory_batches.expiry_date - CURRENT_DATE) <= 30 THEN 'warning'
			ELSE 'safe'
		END as expiry_category,
		inventory_batches.created_at,
		inventory_batches.updated_at
	`).
		Joins("JOIN warehouses ON warehouses.id = inventory_batches.warehouse_id").
		Joins("JOIN product_variants ON product_variants.id = inventory_batches.variant_id").
		Joins("JOIN products ON products.id = product_variants.product_id").
		Joins("LEFT JOIN grn_items ON grn_items.inventory_batch_id = inventory_batches.id").
		Joins("LEFT JOIN goods_receipt_notes ON goods_receipt_notes.id = grn_items.grn_id").
		Where("inventory_batches.total_quantity > 0")

	// Apply filters
	builder := utils.NewReportQueryBuilder(query)
	builder.ApplyStringFilter("inventory_batches.warehouse_id", filter.WarehouseID)
	builder.ApplyStringFilter("product_variants.product_id", filter.ProductID)
	builder.ApplyStringFilter("inventory_batches.variant_id", filter.VariantID)
	builder.ApplyIntRangeFilter("inventory_batches.total_quantity", filter.MinQuantity, filter.MaxQuantity)

	if filter.ExpiringSoon != nil && *filter.ExpiringSoon {
		builder.Build().Where("inventory_batches.expiry_date <= ?", time.Now().AddDate(0, 0, 30))
	}
	if filter.Expired != nil && *filter.Expired {
		builder.Build().Where("inventory_batches.expiry_date < ?", time.Now())
	}

	// Get total count
	var total int64
	if err := s.db.Model(&models.InventoryBatch{}).Where("total_quantity > 0").Count(&total).Error; err != nil {
		s.logger.Error("Failed to count inventory batches", zap.Error(err))
		return nil, err
	}

	// Apply pagination
	builder.ApplySorting(filter.SortBy, filter.SortOrder, "expiry_date ASC")
	builder.ApplyPagination(filter.Limit, filter.Offset)

	// Execute query - initialize as empty slice to return [] instead of null in JSON
	records := make([]models.InventoryReportRecord, 0)
	if err := builder.Build().Scan(&records).Error; err != nil {
		s.logger.Error("Failed to fetch inventory records", zap.Error(err))
		return nil, err
	}

	// Calculate summary
	var summary models.InventoryReportSummary
	s.db.Model(&models.InventoryBatch{}).Where("total_quantity > 0").Count(&summary.TotalBatches)
	s.db.Model(&models.InventoryBatch{}).Where("total_quantity > 0").Select("COALESCE(SUM(total_quantity), 0)").Scan(&summary.TotalQuantity)
	s.db.Model(&models.InventoryBatch{}).Where("total_quantity > 0").Select("COALESCE(SUM(total_quantity * cost_price), 0)").Scan(&summary.TotalValue)
	s.db.Model(&models.InventoryBatch{}).Where("expiry_date <= ?", time.Now().AddDate(0, 0, 30)).Where("expiry_date >= ?", time.Now()).Count(&summary.ExpiringSoonCount)

	return &models.ReportResponse{
		ReportType:  "inventory",
		GeneratedAt: time.Now().Format(time.RFC3339),
		Summary:     summary,
		Records:     records,
		Pagination: &models.PaginationInfo{
			Total:   total,
			Limit:   filter.Limit,
			Offset:  filter.Offset,
			HasMore: filter.Offset+filter.Limit < int(total),
		},
	}, nil
}

// GeneratePurchaseReport generates purchase orders report
func (s *ReportService) GeneratePurchaseReport(filter *models.PurchaseReportFilter) (*models.ReportResponse, error) {
	s.logger.Info("Generating purchase report", zap.Any("filter", filter))

	// Apply defaults
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	// Build query
	// Note: GORM converts ExpectedDelivery -> expected_delivery, ActualDelivery -> actual_delivery
	query := s.db.Model(&models.PurchaseOrder{}).Select(`
		purchase_orders.id,
		purchase_orders.po_number,
		purchase_orders.external_order_id,
		purchase_orders.collaborator_id,
		collaborators.company_name as collaborator_name,
		purchase_orders.warehouse_id,
		warehouses.name as warehouse_name,
		purchase_orders.order_date,
		purchase_orders.expected_delivery,
		purchase_orders.actual_delivery,
		purchase_orders.status,
		purchase_orders.payment_status,
		purchase_orders.total_amount,
		purchase_orders.paid_amount,
		(purchase_orders.total_amount - purchase_orders.paid_amount) as outstanding_amount,
		COUNT(DISTINCT purchase_order_items.id) as item_count,
		goods_receipt_notes.id as grn_id,
		COALESCE(SUM(grn_items.received_quantity), 0) as received_quantity,
		COALESCE(SUM(grn_items.accepted_quantity), 0) as accepted_quantity,
		COALESCE(SUM(grn_items.rejected_quantity), 0) as rejected_quantity,
		CASE
			WHEN SUM(grn_items.received_quantity) > 0 THEN
				(SUM(grn_items.rejected_quantity)::float / SUM(grn_items.received_quantity)::float * 100)
			ELSE 0
		END as rejection_rate,
		COALESCE(SUM(grn_items.rejected_quantity * purchase_order_items.unit_price), 0) as total_rejected_value,
		CASE
			WHEN SUM(grn_items.received_quantity) > 0 THEN
				(SUM(grn_items.accepted_quantity)::float / SUM(grn_items.received_quantity)::float * 100)
			ELSE 0
		END as acceptance_rate_percent,
		purchase_orders.created_at,
		purchase_orders.updated_at
	`).
		Joins("JOIN collaborators ON collaborators.id = purchase_orders.collaborator_id").
		Joins("JOIN warehouses ON warehouses.id = purchase_orders.warehouse_id").
		Joins("LEFT JOIN purchase_order_items ON purchase_order_items.po_id = purchase_orders.id").
		Joins("LEFT JOIN goods_receipt_notes ON goods_receipt_notes.po_id = purchase_orders.id").
		Joins("LEFT JOIN grn_items ON grn_items.grn_id = goods_receipt_notes.id AND grn_items.po_item_id = purchase_order_items.id").
		Group("purchase_orders.id, collaborators.company_name, warehouses.name, goods_receipt_notes.id")

	// Apply filters
	builder := utils.NewReportQueryBuilder(query)
	builder.ApplyStringFilter("purchase_orders.collaborator_id", filter.CollaboratorID)
	builder.ApplyStringFilter("purchase_orders.warehouse_id", filter.WarehouseID)
	builder.ApplyStatusFilter("purchase_orders.status", filter.Status)
	builder.ApplyStatusFilter("purchase_orders.payment_status", filter.PaymentStatus)
	if filter.PONumber != "" {
		builder.Build().Where("purchase_orders.po_number ILIKE ?", "%"+filter.PONumber+"%")
	}
	builder.ApplyDateFilter("purchase_orders.order_date", filter.StartDate, filter.EndDate)

	// Get total count
	var total int64
	if err := s.db.Model(&models.PurchaseOrder{}).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count purchase orders", zap.Error(err))
		return nil, err
	}

	// Apply pagination
	builder.ApplySorting(filter.SortBy, filter.SortOrder, "order_date DESC")
	builder.ApplyPagination(filter.Limit, filter.Offset)

	// Execute query - initialize as empty slice to return [] instead of null in JSON
	records := make([]models.PurchaseReportRecord, 0)
	if err := builder.Build().Scan(&records).Error; err != nil {
		s.logger.Error("Failed to fetch purchase records", zap.Error(err))
		return nil, err
	}

	// Calculate summary
	var summary models.PurchaseReportSummary
	s.db.Model(&models.PurchaseOrder{}).Count(&summary.TotalOrders)
	s.db.Model(&models.PurchaseOrder{}).Select("COALESCE(SUM(total_amount), 0)").Scan(&summary.TotalAmount)
	s.db.Model(&models.PurchaseOrder{}).Select("COALESCE(SUM(paid_amount), 0)").Scan(&summary.PaidAmount)
	summary.OutstandingAmount = summary.TotalAmount - summary.PaidAmount

	// Status breakdown
	summary.ByStatus = make(map[string]int64)
	var statusCounts []struct {
		Status string
		Count  int64
	}
	s.db.Model(&models.PurchaseOrder{}).Select("status, COUNT(*) as count").Group("status").Scan(&statusCounts)
	for _, sc := range statusCounts {
		summary.ByStatus[sc.Status] = sc.Count
	}

	return &models.ReportResponse{
		ReportType:  "purchases",
		GeneratedAt: time.Now().Format(time.RFC3339),
		Summary:     summary,
		Records:     records,
		Pagination: &models.PaginationInfo{
			Total:   total,
			Limit:   filter.Limit,
			Offset:  filter.Offset,
			HasMore: filter.Offset+filter.Limit < int(total),
		},
	}, nil
}

// GenerateSalesReport generates sales report
func (s *ReportService) GenerateSalesReport(filter *models.SalesReportFilter) (*models.ReportResponse, error) {
	s.logger.Info("Generating sales report", zap.Any("filter", filter))

	// Apply defaults
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	// Build query - sale_items has cgst_amount, sgst_amount, igst_amount (not rates)
	query := s.db.Model(&models.Sale{}).Select(`
		sales.id,
		warehouses.name as warehouse_name,
		sales.sale_date,
		sales.status,
		sales.customer_phone,
		sales.customer_name,
		sales.payment_mode,
		sales.sale_type,
		sales.total_amount,
		sales.total_amount as landing_price,
		COALESCE(SUM(sale_items.cost_price * sale_items.quantity), 0) as purchase_value,
		sales.apply_taxes,
		COALESCE(SUM(sale_items.total_tax_amount), 0) as total_tax,
		COALESCE(SUM(sale_items.cgst_amount), 0) as cgst_amount,
		COALESCE(SUM(sale_items.sgst_amount), 0) as sgst_amount,
		COALESCE(SUM(sale_items.igst_amount), 0) as igst_amount,
		COALESCE(SUM(CASE WHEN sale_items.total_tax_amount = 0 THEN sale_items.line_total ELSE 0 END), 0) as tax_exempt_amount,
		CASE
			WHEN sales.total_amount > 0 THEN (COALESCE(SUM(sale_items.total_tax_amount), 0) / sales.total_amount * 100)
			ELSE 0
		END as effective_tax_rate,
		COALESCE(SUM(sale_items.margin * sale_items.quantity), 0) as total_margin,
		CASE
			WHEN sales.total_amount > 0 THEN (COALESCE(SUM(sale_items.margin * sale_items.quantity), 0) / sales.total_amount * 100)
			ELSE 0
		END as gross_margin_percent,
		CASE
			WHEN SUM(sale_items.quantity) > 0 THEN (COALESCE(SUM(sale_items.margin * sale_items.quantity), 0) / SUM(sale_items.quantity))
			ELSE 0
		END as per_unit_margin,
		COUNT(sale_items.id) as item_count,
		sales.cancelled_at,
		sales.cancellation_reason,
		sales.created_at,
		sales.updated_at
	`).
		Joins("JOIN warehouses ON warehouses.id = sales.warehouse_id").
		Joins("LEFT JOIN sale_items ON sale_items.sale_id = sales.id").
		Group("sales.id, warehouses.name")

	// Apply filters
	builder := utils.NewReportQueryBuilder(query)
	builder.ApplyStringFilter("sales.warehouse_id", filter.WarehouseID)
	builder.ApplyStatusFilter("sales.status", filter.Status)
	builder.ApplyStatusFilter("sales.payment_mode", filter.PaymentMode)
	builder.ApplyStatusFilter("sales.sale_type", filter.SaleType)
	builder.ApplyRangeFilter("sales.total_amount", filter.MinAmount, filter.MaxAmount)
	builder.ApplyDateFilter("sales.sale_date", filter.StartDate, filter.EndDate)

	// Get total count
	var total int64
	if err := s.db.Model(&models.Sale{}).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count sales", zap.Error(err))
		return nil, err
	}

	// Apply pagination
	builder.ApplySorting(filter.SortBy, filter.SortOrder, "sale_date DESC")
	builder.ApplyPagination(filter.Limit, filter.Offset)

	// Execute query - initialize as empty slice to return [] instead of null in JSON
	records := make([]models.SalesReportRecord, 0)
	if err := builder.Build().Scan(&records).Error; err != nil {
		s.logger.Error("Failed to fetch sales records", zap.Error(err))
		return nil, err
	}

	// Calculate summary
	var summary models.SalesReportSummary
	s.db.Model(&models.Sale{}).Count(&summary.TotalSales)
	s.db.Model(&models.Sale{}).Select("COALESCE(SUM(total_amount), 0)").Scan(&summary.TotalRevenue)
	s.db.Model(&models.SaleItem{}).Select("COALESCE(SUM(cost_price * quantity), 0)").Scan(&summary.TotalPurchaseValue)
	s.db.Model(&models.SaleItem{}).Select("COALESCE(SUM(total_tax_amount), 0)").Scan(&summary.TotalTax)
	s.db.Model(&models.SaleItem{}).Select("COALESCE(SUM(margin * quantity), 0)").Scan(&summary.TotalMargin)
	if summary.TotalSales > 0 {
		summary.AverageSaleValue = summary.TotalRevenue / float64(summary.TotalSales)
	}

	// Payment mode breakdown
	summary.ByPaymentMode = make(map[string]int64)
	var pmCounts []struct {
		PaymentMode string
		Count       int64
	}
	s.db.Model(&models.Sale{}).Select("payment_mode, COUNT(*) as count").Group("payment_mode").Scan(&pmCounts)
	for _, pm := range pmCounts {
		summary.ByPaymentMode[pm.PaymentMode] = pm.Count
	}

	// Sale type breakdown
	summary.BySaleType = make(map[string]int64)
	var stCounts []struct {
		SaleType string
		Count    int64
	}
	s.db.Model(&models.Sale{}).Select("sale_type, COUNT(*) as count").Group("sale_type").Scan(&stCounts)
	for _, st := range stCounts {
		summary.BySaleType[st.SaleType] = st.Count
	}

	return &models.ReportResponse{
		ReportType:  "sales",
		GeneratedAt: time.Now().Format(time.RFC3339),
		Summary:     summary,
		Records:     records,
		Pagination: &models.PaginationInfo{
			Total:   total,
			Limit:   filter.Limit,
			Offset:  filter.Offset,
			HasMore: filter.Offset+filter.Limit < int(total),
		},
	}, nil
}

// GenerateReturnsReport generates returns report
func (s *ReportService) GenerateReturnsReport(filter *models.ReturnsReportFilter) (*models.ReportResponse, error) {
	s.logger.Info("Generating returns report", zap.Any("filter", filter))

	// Apply defaults
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	// Build query
	query := s.db.Model(&models.Return{}).Select(`
		returns.id,
		returns.sale_id,
		returns.return_date,
		returns.status,
		returns.total_refund,
		sales.warehouse_id,
		warehouses.name as warehouse_name,
		COUNT(return_items.id) as item_count,
		returns.created_at,
		returns.updated_at
	`).
		Joins("JOIN sales ON sales.id = returns.sale_id").
		Joins("JOIN warehouses ON warehouses.id = sales.warehouse_id").
		Joins("LEFT JOIN return_items ON return_items.return_id = returns.id").
		Group("returns.id, sales.warehouse_id, warehouses.name")

	// Apply filters
	builder := utils.NewReportQueryBuilder(query)
	builder.ApplyStringFilter("returns.sale_id", filter.SaleID)
	builder.ApplyStringFilter("sales.warehouse_id", filter.WarehouseID)
	builder.ApplyStatusFilter("returns.status", filter.Status)
	builder.ApplyRangeFilter("returns.total_refund", filter.MinRefund, filter.MaxRefund)
	builder.ApplyDateFilter("returns.return_date", filter.StartDate, filter.EndDate)

	// Get total count
	var total int64
	if err := s.db.Model(&models.Return{}).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count returns", zap.Error(err))
		return nil, err
	}

	// Apply pagination
	builder.ApplySorting(filter.SortBy, filter.SortOrder, "return_date DESC")
	builder.ApplyPagination(filter.Limit, filter.Offset)

	// Execute query - initialize as empty slice to return [] instead of null in JSON
	records := make([]models.ReturnsReportRecord, 0)
	if err := builder.Build().Scan(&records).Error; err != nil {
		s.logger.Error("Failed to fetch returns records", zap.Error(err))
		return nil, err
	}

	// Calculate summary
	var summary models.ReturnsReportSummary
	s.db.Model(&models.Return{}).Count(&summary.TotalReturns)
	s.db.Model(&models.Return{}).Select("COALESCE(SUM(total_refund), 0)").Scan(&summary.TotalRefundAmount)
	if summary.TotalReturns > 0 {
		summary.AverageRefundValue = summary.TotalRefundAmount / float64(summary.TotalReturns)
	}

	// Status breakdown
	summary.ByStatus = make(map[string]int64)
	var statusCounts []struct {
		Status string
		Count  int64
	}
	s.db.Model(&models.Return{}).Select("status, COUNT(*) as count").Group("status").Scan(&statusCounts)
	for _, sc := range statusCounts {
		summary.ByStatus[sc.Status] = sc.Count
	}

	return &models.ReportResponse{
		ReportType:  "returns",
		GeneratedAt: time.Now().Format(time.RFC3339),
		Summary:     summary,
		Records:     records,
		Pagination: &models.PaginationInfo{
			Total:   total,
			Limit:   filter.Limit,
			Offset:  filter.Offset,
			HasMore: filter.Offset+filter.Limit < int(total),
		},
	}, nil
}

// GenerateGRNReport generates goods receipt note report
func (s *ReportService) GenerateGRNReport(filter *models.GRNReportFilter) (*models.ReportResponse, error) {
	s.logger.Info("Generating GRN report", zap.Any("filter", filter))

	// Apply defaults
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	// Build query
	query := s.db.Model(&models.GRN{}).Select(`
		goods_receipt_notes.id,
		goods_receipt_notes.grn_number,
		purchase_orders.po_number,
		goods_receipt_notes.po_id,
		purchase_orders.collaborator_id as vendor_id,
		collaborators.company_name as vendor_name,
		goods_receipt_notes.warehouse_id,
		warehouses.name as warehouse_name,
		goods_receipt_notes.received_date,
		goods_receipt_notes.quality_status,
		COUNT(DISTINCT grn_items.id) as item_count,
		COALESCE(SUM(grn_items.received_quantity), 0) as total_received,
		COALESCE(SUM(grn_items.accepted_quantity), 0) as total_accepted,
		COALESCE(SUM(grn_items.rejected_quantity), 0) as total_rejected,
		COALESCE(SUM(grn_items.received_quantity * purchase_order_items.unit_price), 0) as received_value,
		COALESCE(SUM(grn_items.rejected_quantity * purchase_order_items.unit_price), 0) as rejected_value,
		CASE
			WHEN SUM(grn_items.received_quantity) > 0 THEN
				(SUM(grn_items.accepted_quantity)::float / SUM(grn_items.received_quantity)::float * 100)
			ELSE 0
		END as acceptance_rate,
		goods_receipt_notes.created_at,
		goods_receipt_notes.updated_at
	`).
		Joins("JOIN purchase_orders ON purchase_orders.id = goods_receipt_notes.po_id").
		Joins("JOIN collaborators ON collaborators.id = purchase_orders.collaborator_id").
		Joins("JOIN warehouses ON warehouses.id = goods_receipt_notes.warehouse_id").
		Joins("LEFT JOIN grn_items ON grn_items.grn_id = goods_receipt_notes.id").
		Joins("LEFT JOIN purchase_order_items ON purchase_order_items.id = grn_items.po_item_id").
		Group("goods_receipt_notes.id, purchase_orders.po_number, purchase_orders.collaborator_id, collaborators.company_name, warehouses.name")

	// Apply filters
	builder := utils.NewReportQueryBuilder(query)
	builder.ApplyStringFilter("goods_receipt_notes.po_id", filter.POID)
	builder.ApplyStringFilter("goods_receipt_notes.warehouse_id", filter.WarehouseID)
	builder.ApplyStringFilter("purchase_orders.collaborator_id", filter.VendorID)
	builder.ApplyStatusFilter("goods_receipt_notes.quality_status", filter.QualityStatus)
	if filter.PONumber != "" {
		builder.Build().Where("purchase_orders.po_number ILIKE ?", "%"+filter.PONumber+"%")
	}
	builder.ApplyDateFilter("goods_receipt_notes.received_date", filter.StartDate, filter.EndDate)

	// Get total count
	var total int64
	if err := s.db.Model(&models.GRN{}).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count GRNs", zap.Error(err))
		return nil, err
	}

	// Apply pagination
	builder.ApplySorting(filter.SortBy, filter.SortOrder, "received_date DESC")
	builder.ApplyPagination(filter.Limit, filter.Offset)

	// Execute query
	var records []models.GRNReportRecord
	if err := builder.Build().Scan(&records).Error; err != nil {
		s.logger.Error("Failed to fetch GRN records", zap.Error(err))
		return nil, err
	}

	// Calculate summary
	var summary models.GRNReportSummary
	s.db.Model(&models.GRN{}).Count(&summary.TotalGRNs)

	// Aggregate metrics
	var aggregates struct {
		TotalItems         int64
		TotalReceivedValue float64
		TotalRejectedValue float64
		TotalAcceptedValue float64
		TotalReceived      int64
		TotalAccepted      int64
	}
	s.db.Model(&models.GRN{}).Select(`
		COUNT(DISTINCT grn_items.id) as total_items,
		COALESCE(SUM(grn_items.received_quantity * purchase_order_items.unit_price), 0) as total_received_value,
		COALESCE(SUM(grn_items.rejected_quantity * purchase_order_items.unit_price), 0) as total_rejected_value,
		COALESCE(SUM(grn_items.accepted_quantity * purchase_order_items.unit_price), 0) as total_accepted_value,
		COALESCE(SUM(grn_items.received_quantity), 0) as total_received,
		COALESCE(SUM(grn_items.accepted_quantity), 0) as total_accepted
	`).
		Joins("LEFT JOIN grn_items ON grn_items.grn_id = goods_receipt_notes.id").
		Joins("LEFT JOIN purchase_order_items ON purchase_order_items.id = grn_items.po_item_id").
		Scan(&aggregates)

	summary.TotalItemCount = aggregates.TotalItems
	summary.TotalReceivedValue = aggregates.TotalReceivedValue
	summary.TotalRejectedValue = aggregates.TotalRejectedValue
	summary.TotalAcceptedValue = aggregates.TotalAcceptedValue
	if aggregates.TotalReceived > 0 {
		summary.AverageAcceptanceRate = (float64(aggregates.TotalAccepted) / float64(aggregates.TotalReceived)) * 100
	}

	// Quality status breakdown
	summary.ByQualityStatus = make(map[string]int64)
	var statusCounts []struct {
		QualityStatus string
		Count         int64
	}
	s.db.Model(&models.GRN{}).Select("quality_status, COUNT(*) as count").Group("quality_status").Scan(&statusCounts)
	for _, sc := range statusCounts {
		summary.ByQualityStatus[sc.QualityStatus] = sc.Count
	}

	return &models.ReportResponse{
		ReportType:  "grn",
		GeneratedAt: time.Now().Format(time.RFC3339),
		Summary:     summary,
		Records:     records,
		Pagination: &models.PaginationInfo{
			Total:   total,
			Limit:   filter.Limit,
			Offset:  filter.Offset,
			HasMore: filter.Offset+filter.Limit < int(total),
		},
	}, nil
}
