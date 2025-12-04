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

	// Execute query
	var records []models.ProductReportRecord
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

	// Build query
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
		COUNT(DISTINCT collaborator_products.id) as product_count,
		COALESCE(SUM(purchase_orders.total_amount), 0) as total_po_value,
		collaborators.created_at,
		collaborators.updated_at
	`).
		Joins("LEFT JOIN collaborator_products ON collaborator_products.collaborator_id = collaborators.id").
		Joins("LEFT JOIN purchase_orders ON purchase_orders.collaborator_id = collaborators.id").
		Group("collaborators.id")

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

	// Execute query
	var records []models.VendorReportRecord
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

	// Build query
	query := s.db.Model(&models.Sale{}).Select(`
		sales.farmer_id,
		COUNT(*) as total_purchases,
		SUM(sales.total_amount) as total_amount,
		MAX(sales.sale_date) as last_purchase_date,
		sales.warehouse_id,
		warehouses.name as warehouse_name,
		COUNT(*) as purchase_count
	`).
		Joins("JOIN warehouses ON warehouses.id = sales.warehouse_id").
		Where("sales.farmer_id IS NOT NULL AND sales.farmer_id != ''").
		Where("sales.status != 'cancelled'").
		Group("sales.farmer_id, sales.warehouse_id, warehouses.name")

	// Apply filters
	builder := utils.NewReportQueryBuilder(query)
	if filter.Search != "" {
		builder.Build().Where("sales.farmer_id ILIKE ?", "%"+filter.Search+"%")
	}
	builder.ApplyStringFilter("sales.warehouse_id", filter.WarehouseID)
	builder.ApplyRangeFilter("SUM(sales.total_amount)", filter.MinPurchaseValue, filter.MaxPurchaseValue)
	builder.ApplyDateFilter("sales.sale_date", filter.StartDate, filter.EndDate)

	// Get total count
	var total int64
	countQuery := s.db.Model(&models.Sale{}).
		Where("farmer_id IS NOT NULL AND farmer_id != ''").
		Where("status != 'cancelled'").
		Distinct("farmer_id")
	countQuery.Count(&total)

	// Apply pagination
	builder.ApplySorting(filter.SortBy, filter.SortOrder, "total_amount DESC")
	builder.ApplyPagination(filter.Limit, filter.Offset)

	// Execute query
	var records []models.CustomerReportRecord
	if err := builder.Build().Scan(&records).Error; err != nil {
		s.logger.Error("Failed to fetch customer records", zap.Error(err))
		return nil, err
	}

	// Calculate summary
	var summary models.CustomerReportSummary
	s.db.Model(&models.Sale{}).
		Where("farmer_id IS NOT NULL AND farmer_id != ''").
		Where("status != 'cancelled'").
		Distinct("farmer_id").
		Count(&summary.TotalCustomers)
	s.db.Model(&models.Sale{}).
		Where("farmer_id IS NOT NULL").
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

	// Build query
	query := s.db.Model(&models.InventoryBatch{}).Select(`
		inventory_batches.id as batch_id,
		inventory_batches.warehouse_id,
		warehouses.name as warehouse_name,
		inventory_batches.variant_id,
		products.name as product_name,
		product_variants.sku as variant_sku,
		inventory_batches.total_quantity,
		inventory_batches.cost_price,
		(inventory_batches.total_quantity * inventory_batches.cost_price) as total_value,
		inventory_batches.expiry_date,
		EXTRACT(DAY FROM (inventory_batches.expiry_date - CURRENT_DATE)) as days_to_expiry,
		inventory_batches.cgst_rate,
		inventory_batches.sgst_rate,
		inventory_batches.is_tax_exempt,
		inventory_batches.created_at,
		inventory_batches.updated_at
	`).
		Joins("JOIN warehouses ON warehouses.id = inventory_batches.warehouse_id").
		Joins("JOIN product_variants ON product_variants.id = inventory_batches.variant_id").
		Joins("JOIN products ON products.id = product_variants.product_id").
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

	// Execute query
	var records []models.InventoryReportRecord
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
	query := s.db.Model(&models.PurchaseOrder{}).Select(`
		purchase_orders.id,
		purchase_orders.po_number,
		purchase_orders.external_order_id,
		purchase_orders.collaborator_id,
		collaborators.company_name as collaborator_name,
		purchase_orders.warehouse_id,
		warehouses.name as warehouse_name,
		purchase_orders.order_date,
		purchase_orders.expected_delivery_date,
		purchase_orders.actual_delivery_date,
		purchase_orders.status,
		purchase_orders.payment_status,
		purchase_orders.total_amount,
		purchase_orders.paid_amount,
		(purchase_orders.total_amount - purchase_orders.paid_amount) as outstanding_amount,
		0 as item_count,
		purchase_orders.created_at,
		purchase_orders.updated_at
	`).
		Joins("JOIN collaborators ON collaborators.id = purchase_orders.collaborator_id").
		Joins("JOIN warehouses ON warehouses.id = purchase_orders.warehouse_id")

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

	// Execute query
	var records []models.PurchaseReportRecord
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

	// Build query
	query := s.db.Model(&models.Sale{}).Select(`
		sales.id,
		sales.warehouse_id,
		warehouses.name as warehouse_name,
		sales.sale_date,
		sales.status,
		sales.farmer_id,
		sales.payment_mode,
		sales.sale_type,
		sales.total_amount,
		sales.apply_taxes,
		COALESCE(SUM(sale_items.total_tax_amount), 0) as total_tax,
		COALESCE(SUM(sale_items.margin * sale_items.quantity), 0) as total_margin,
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
	builder.ApplyStringFilter("sales.farmer_id", filter.FarmerID)
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

	// Execute query
	var records []models.SalesReportRecord
	if err := builder.Build().Scan(&records).Error; err != nil {
		s.logger.Error("Failed to fetch sales records", zap.Error(err))
		return nil, err
	}

	// Calculate summary
	var summary models.SalesReportSummary
	s.db.Model(&models.Sale{}).Count(&summary.TotalSales)
	s.db.Model(&models.Sale{}).Select("COALESCE(SUM(total_amount), 0)").Scan(&summary.TotalRevenue)
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

	// Execute query
	var records []models.ReturnsReportRecord
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
