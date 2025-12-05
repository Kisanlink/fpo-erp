package models

import "time"

// BaseReportFilter contains common filter parameters for all reports
type BaseReportFilter struct {
	StartDate *time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate   *time.Time `form:"end_date" time_format:"2006-01-02"`
	Limit     int        `form:"limit" binding:"min=0,max=500"`
	Offset    int        `form:"offset" binding:"min=0"`
	SortBy    string     `form:"sort_by"`
	SortOrder string     `form:"sort_order" binding:"omitempty,oneof=asc desc"`
	Columns   []string   `form:"columns"`
	Format    string     `form:"format" binding:"omitempty,oneof=json xlsx pdf"`
}

// ProductReportFilter for product master report
type ProductReportFilter struct {
	BaseReportFilter
	Search      string `form:"search"`
	HasVariants *bool  `form:"has_variants"`
	IsActive    *bool  `form:"is_active"`
}

// VendorReportFilter for vendor master report
type VendorReportFilter struct {
	BaseReportFilter
	Search   string `form:"search"`
	IsActive *bool  `form:"is_active"`
	HasGST   *bool  `form:"has_gst"`
}

// CustomerReportFilter for customer report
type CustomerReportFilter struct {
	BaseReportFilter
	Search           string   `form:"search"`
	WarehouseID      string   `form:"warehouse_id"`
	MinPurchaseValue *float64 `form:"min_purchase_value"`
	MaxPurchaseValue *float64 `form:"max_purchase_value"`
}

// InventoryReportFilter for inventory report
type InventoryReportFilter struct {
	BaseReportFilter
	WarehouseID  string `form:"warehouse_id"`
	ProductID    string `form:"product_id"`
	VariantID    string `form:"variant_id"`
	LowStock     *bool  `form:"low_stock"`
	ExpiringSoon *bool  `form:"expiring_soon"`
	Expired      *bool  `form:"expired"`
	MinQuantity  *int64 `form:"min_quantity"`
	MaxQuantity  *int64 `form:"max_quantity"`
}

// PurchaseReportFilter for purchases report
type PurchaseReportFilter struct {
	BaseReportFilter
	CollaboratorID string   `form:"collaborator_id"`
	WarehouseID    string   `form:"warehouse_id"`
	Status         []string `form:"status"`
	PaymentStatus  []string `form:"payment_status"`
	PONumber       string   `form:"po_number"`
}

// SalesReportFilter for sales report
type SalesReportFilter struct {
	BaseReportFilter
	WarehouseID string   `form:"warehouse_id"`
	CustomerID  string   `form:"customer_id"`
	Status      []string `form:"status"`
	PaymentMode []string `form:"payment_mode"`
	SaleType    []string `form:"sale_type"`
	MinAmount   *float64 `form:"min_amount"`
	MaxAmount   *float64 `form:"max_amount"`
}

// ReturnsReportFilter for returns report
type ReturnsReportFilter struct {
	BaseReportFilter
	SaleID      string   `form:"sale_id"`
	WarehouseID string   `form:"warehouse_id"`
	Status      []string `form:"status"`
	MinRefund   *float64 `form:"min_refund"`
	MaxRefund   *float64 `form:"max_refund"`
}

// ReportResponse is the standard report response wrapper
type ReportResponse struct {
	ReportType     string                 `json:"report_type"`
	GeneratedAt    string                 `json:"generated_at"`
	FiltersApplied map[string]interface{} `json:"filters_applied,omitempty"`
	Summary        interface{}            `json:"summary"`
	Records        interface{}            `json:"records"`
	Pagination     *PaginationInfo        `json:"pagination,omitempty"`
}

// PaginationInfo for paginated responses
type PaginationInfo struct {
	Total   int64 `json:"total"`
	Limit   int   `json:"limit"`
	Offset  int   `json:"offset"`
	HasMore bool  `json:"has_more"`
}

// ProductReportSummary aggregates for product report
type ProductReportSummary struct {
	TotalProducts     int64 `json:"total_products"`
	TotalVariants     int64 `json:"total_variants"`
	ProductsWithStock int64 `json:"products_with_stock"`
}

// VendorReportSummary aggregates for vendor report
type VendorReportSummary struct {
	TotalVendors   int64   `json:"total_vendors"`
	ActiveVendors  int64   `json:"active_vendors"`
	VendorsWithGST int64   `json:"vendors_with_gst"`
	TotalPOValue   float64 `json:"total_po_value"`
}

// CustomerReportSummary aggregates for customer report
type CustomerReportSummary struct {
	TotalCustomers       int64   `json:"total_customers"`
	TotalRevenue         float64 `json:"total_revenue"`
	AveragePurchaseValue float64 `json:"average_purchase_value"`
}

// InventoryReportSummary aggregates for inventory report
type InventoryReportSummary struct {
	TotalBatches      int64   `json:"total_batches"`
	TotalQuantity     int64   `json:"total_quantity"`
	TotalValue        float64 `json:"total_value"`
	ExpiringSoonCount int64   `json:"expiring_soon_count"`
	LowStockCount     int64   `json:"low_stock_count"`
}

// PurchaseReportSummary aggregates for purchase report
type PurchaseReportSummary struct {
	TotalOrders       int64            `json:"total_orders"`
	TotalAmount       float64          `json:"total_amount"`
	PaidAmount        float64          `json:"paid_amount"`
	OutstandingAmount float64          `json:"outstanding_amount"`
	ByStatus          map[string]int64 `json:"by_status"`
}

// SalesReportSummary aggregates for sales report
type SalesReportSummary struct {
	TotalSales       int64            `json:"total_sales"`
	TotalRevenue     float64          `json:"total_revenue"`
	TotalTax         float64          `json:"total_tax"`
	TotalMargin      float64          `json:"total_margin"`
	AverageSaleValue float64          `json:"average_sale_value"`
	ByPaymentMode    map[string]int64 `json:"by_payment_mode"`
	BySaleType       map[string]int64 `json:"by_sale_type"`
}

// ReturnsReportSummary aggregates for returns report
type ReturnsReportSummary struct {
	TotalReturns       int64            `json:"total_returns"`
	TotalRefundAmount  float64          `json:"total_refund_amount"`
	AverageRefundValue float64          `json:"average_refund_value"`
	ByStatus           map[string]int64 `json:"by_status"`
}

// ProductReportRecord for product report rows
type ProductReportRecord struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	ExternalID   *string `json:"external_id,omitempty"`
	VariantCount int64   `json:"variant_count"`
	TotalStock   int64   `json:"total_stock"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// VendorReportRecord for vendor report rows
type VendorReportRecord struct {
	ID            string  `json:"id"`
	CompanyName   string  `json:"company_name"`
	ContactPerson string  `json:"contact_person"`
	ContactNumber string  `json:"contact_number"`
	Email         *string `json:"email,omitempty"`
	GSTNumber     string  `json:"gst_number"`
	PANNumber     *string `json:"pan_number,omitempty"`
	BankAccountNo string  `json:"bank_account_no"`
	BankIFSC      string  `json:"bank_ifsc"`
	BankName      *string `json:"bank_name,omitempty"`
	IsActive      bool    `json:"is_active"`
	ProductCount  int64   `json:"product_count"`
	TotalPOValue  float64 `json:"total_po_value"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// CustomerReportRecord for customer report rows
type CustomerReportRecord struct {
	CustomerID       string  `json:"customer_id"`
	TotalPurchases   int64   `json:"total_purchases"`
	TotalAmount      float64 `json:"total_amount"`
	LastPurchaseDate string  `json:"last_purchase_date"`
	WarehouseID      string  `json:"warehouse_id,omitempty"`
	WarehouseName    string  `json:"warehouse_name,omitempty"`
	PurchaseCount    int64   `json:"purchase_count"`
}

// InventoryReportRecord for inventory report rows
type InventoryReportRecord struct {
	BatchID       string  `json:"batch_id"`
	WarehouseID   string  `json:"warehouse_id"`
	WarehouseName string  `json:"warehouse_name"`
	VariantID     string  `json:"variant_id"`
	ProductName   string  `json:"product_name"`
	VariantSKU    string  `json:"variant_sku"`
	TotalQuantity int64   `json:"total_quantity"`
	CostPrice     float64 `json:"cost_price"`
	TotalValue    float64 `json:"total_value"`
	ExpiryDate    string  `json:"expiry_date"`
	DaysToExpiry  int     `json:"days_to_expiry"`
	CGSTRate      float64 `json:"cgst_rate"`
	SGSTRate      float64 `json:"sgst_rate"`
	IsTaxExempt   bool    `json:"is_tax_exempt"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// PurchaseReportRecord for purchase report rows
type PurchaseReportRecord struct {
	ID                   string  `json:"id"`
	PONumber             string  `json:"po_number"`
	ExternalOrderID      *string `json:"external_order_id,omitempty"`
	CollaboratorID       string  `json:"collaborator_id"`
	CollaboratorName     string  `json:"collaborator_name"`
	WarehouseID          string  `json:"warehouse_id"`
	WarehouseName        string  `json:"warehouse_name"`
	OrderDate            string  `json:"order_date"`
	ExpectedDeliveryDate string  `json:"expected_delivery_date"`
	ActualDeliveryDate   *string `json:"actual_delivery_date,omitempty"`
	Status               string  `json:"status"`
	PaymentStatus        string  `json:"payment_status"`
	TotalAmount          float64 `json:"total_amount"`
	PaidAmount           float64 `json:"paid_amount"`
	OutstandingAmount    float64 `json:"outstanding_amount"`
	ItemCount            int     `json:"item_count"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
}

// SalesReportRecord for sales report rows
type SalesReportRecord struct {
	ID                 string  `json:"id"`
	WarehouseID        string  `json:"warehouse_id"`
	WarehouseName      string  `json:"warehouse_name"`
	SaleDate           string  `json:"sale_date"`
	Status             string  `json:"status"`
	CustomerID         *string `json:"customer_id,omitempty"`
	PaymentMode        string  `json:"payment_mode"`
	SaleType           string  `json:"sale_type"`
	TotalAmount        float64 `json:"total_amount"`
	ApplyTaxes         bool    `json:"apply_taxes"`
	TotalTax           float64 `json:"total_tax"`
	TotalMargin        float64 `json:"total_margin"`
	ItemCount          int     `json:"item_count"`
	CancelledAt        *string `json:"cancelled_at,omitempty"`
	CancellationReason *string `json:"cancellation_reason,omitempty"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

// ReturnsReportRecord for returns report rows
type ReturnsReportRecord struct {
	ID            string  `json:"id"`
	SaleID        string  `json:"sale_id"`
	ReturnDate    string  `json:"return_date"`
	Status        string  `json:"status"`
	TotalRefund   float64 `json:"total_refund"`
	WarehouseID   string  `json:"warehouse_id"`
	WarehouseName string  `json:"warehouse_name"`
	ItemCount     int     `json:"item_count"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}
