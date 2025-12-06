package models

import "time"

// IncludeOptions specifies which related data to include in aggregated responses
type IncludeOptions struct {
	Variants      bool
	Prices        bool
	Inventory     bool
	Collaborators bool
	Taxes         bool
}

// ProductDetailRequest represents query parameters for product detail aggregation
type ProductDetailRequest struct {
	Include      string `form:"include"`      // Comma-separated: variants,prices,inventory,collaborators,taxes
	WarehouseID  string `form:"warehouse_id"` // Filter inventory by warehouse
	PriceType    string `form:"price_type"`   // Filter prices: retail, wholesale, bulk, all
	ActiveOnly   bool   `form:"active_only"`  // Show only active variants (default: true)
	InStockOnly  bool   `form:"in_stock_only"` // Show only variants with stock
}

// SalesContextRequest represents query parameters for sales context
type SalesContextRequest struct {
	WarehouseID      string `form:"warehouse_id"`
	IncludeZeroStock bool   `form:"include_zero_stock"`
	PriceType        string `form:"price_type"`
	EffectiveDate    string `form:"effective_date"` // ISO date for price effective date
}

// === Product Detail Response Types ===

// ProductDetailResponse represents the aggregated product detail response
type ProductDetailResponse struct {
	Product      ProductInfo       `json:"product"`
	Collaborator *CollaboratorInfo `json:"collaborator,omitempty"`
	Variants     []VariantDetail   `json:"variants,omitempty"`
	Metadata     ProductMetadata   `json:"metadata"`
}

// ProductInfo represents basic product information
type ProductInfo struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Description    *string `json:"description,omitempty"`
	Category       *string `json:"category,omitempty"`
	OrganizationID string  `json:"organization_id,omitempty"`
	IsActive       bool    `json:"is_active"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// CollaboratorInfo represents collaborator/vendor information
type CollaboratorInfo struct {
	ID            string       `json:"id"`
	CompanyName   string       `json:"company_name"`
	ContactPerson *string      `json:"contact_person,omitempty"`
	Phone         *string      `json:"phone,omitempty"`
	Email         *string      `json:"email,omitempty"`
	Address       *AddressInfo `json:"address,omitempty"`
	IsActive      bool         `json:"is_active"`
}

// VariantDetail represents variant information with prices and inventory
type VariantDetail struct {
	ID                 string              `json:"id"`
	ProductID          string              `json:"product_id"`
	VariantName        string              `json:"variant_name"`
	Description        *string             `json:"description,omitempty"`
	SKU                string              `json:"sku"`
	Barcode            *string             `json:"barcode,omitempty"`
	ExternalID         *string             `json:"external_id,omitempty"`
	BrandName          *string             `json:"brand_name,omitempty"`
	Quantity           string              `json:"quantity"`
	PackSize           *string             `json:"pack_size,omitempty"`
	Images             []string            `json:"images,omitempty"`
	HSNCode            *string             `json:"hsn_code,omitempty"`
	GSTRate            float64             `json:"gst_rate"`
	DosageInstructions *string             `json:"dosage_instructions,omitempty"`
	UsageDetails       *string             `json:"usage_details,omitempty"`
	IsActive           bool                `json:"is_active"`
	CreatedAt          string              `json:"created_at"`
	UpdatedAt          string              `json:"updated_at"`
	Prices             *VariantPrices      `json:"prices,omitempty"`
	StockSummary       *StockSummary       `json:"stock_summary,omitempty"`
	WarehouseStock     []WarehouseStock    `json:"warehouse_stock,omitempty"`
	TaxConfiguration   *TaxConfiguration   `json:"tax_configuration,omitempty"`
}

// VariantPrices represents pricing information for a variant
type VariantPrices struct {
	Currency       string      `json:"currency"`
	HasActivePrice bool        `json:"has_active_price"`
	RetailPrice    *PriceInfo  `json:"retail_price,omitempty"`
	WholesalePrice *PriceInfo  `json:"wholesale_price,omitempty"`
	BulkPrice      *PriceInfo  `json:"bulk_price,omitempty"`
}

// PriceInfo represents individual price details
type PriceInfo struct {
	Price         float64 `json:"price"`
	EffectiveFrom string  `json:"effective_from"`
	EffectiveTo   *string `json:"effective_to,omitempty"`
}

// StockSummary represents aggregated stock information
type StockSummary struct {
	TotalQuantity     int64   `json:"total_quantity"`
	AvailableQuantity int64   `json:"available_quantity"`
	ReservedQuantity  int64   `json:"reserved_quantity"`
	InStock           bool    `json:"in_stock"`
	WarehouseCount    int     `json:"warehouse_count"`
	MinCostPrice      float64 `json:"min_cost_price"`
	MaxCostPrice      float64 `json:"max_cost_price"`
	EarliestExpiry    *string `json:"earliest_expiry,omitempty"`
}

// WarehouseStock represents stock in a specific warehouse
type WarehouseStock struct {
	WarehouseID   string  `json:"warehouse_id"`
	WarehouseName string  `json:"warehouse_name"`
	Quantity      int64   `json:"quantity"`
	CostPrice     float64 `json:"cost_price"`
	ExpiryDate    string  `json:"expiry_date"`
	BatchCount    int     `json:"batch_count"`
}

// TaxConfiguration represents tax settings for a variant
type TaxConfiguration struct {
	CGSTRate     float64  `json:"cgst_rate"`
	SGSTRate     float64  `json:"sgst_rate"`
	IsTaxExempt  bool     `json:"is_tax_exempt"`
	CustomTaxIDs []string `json:"custom_tax_ids,omitempty"`
}

// ProductMetadata represents metadata for the response
type ProductMetadata struct {
	TotalVariants    int     `json:"total_variants"`
	ActiveVariants   int     `json:"active_variants"`
	TotalStockValue  float64 `json:"total_stock_value"`
	ReadTimestamp    string  `json:"read_timestamp"`
	ConsistencyToken string  `json:"consistency_token,omitempty"`
}

// === Variant Detail Response (Single Variant) ===

// VariantDetailResponse represents the single variant detail response
type VariantDetailResponse struct {
	Variant  VariantDetailWithProduct `json:"variant"`
	Metadata ResponseMetadata         `json:"metadata"`
}

// VariantDetailWithProduct includes parent product info
type VariantDetailWithProduct struct {
	VariantDetail
	Product      *ProductBasicInfo    `json:"product,omitempty"`
	Collaborator *CollaboratorBasicInfo `json:"collaborator,omitempty"`
}

// ProductBasicInfo represents minimal product information
type ProductBasicInfo struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Category *string `json:"category,omitempty"`
}

// CollaboratorBasicInfo represents minimal collaborator information
type CollaboratorBasicInfo struct {
	ID          string `json:"id"`
	CompanyName string `json:"company_name"`
}

// ResponseMetadata represents common metadata
type ResponseMetadata struct {
	ReadTimestamp    string `json:"read_timestamp"`
	ConsistencyToken string `json:"consistency_token,omitempty"`
}

// === Sales Context Response Types ===

// SalesContextResponse represents the aggregated sales context response
type SalesContextResponse struct {
	Warehouse              WarehouseInfo            `json:"warehouse"`
	AvailableInventory     []InventoryWithPricing   `json:"available_inventory"`
	GlobalTaxConfiguration GlobalTaxConfig          `json:"global_tax_configuration"`
	DiscountPolicies       []DiscountPolicy         `json:"discount_policies"`
	RefundPolicies         []RefundPolicyInfo       `json:"refund_policies"`
	PaymentMethods         []PaymentMethodInfo      `json:"payment_methods"`
	Metadata               SalesContextMetadata     `json:"metadata"`
}

// WarehouseInfo represents warehouse details for sales context
type WarehouseInfo struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Address        *AddressInfo `json:"address,omitempty"`
	ContactPhone   *string      `json:"contact_phone,omitempty"`
	IsActive       bool         `json:"is_active"`
	OrganizationID string       `json:"organization_id,omitempty"`
}

// InventoryWithPricing represents inventory batch with pricing and product info
type InventoryWithPricing struct {
	BatchID          string               `json:"batch_id"`
	VariantID        string               `json:"variant_id"`
	Variant          VariantInfoForSales  `json:"variant"`
	Product          ProductInfoForSales  `json:"product"`
	QuantityTotal    int64                `json:"quantity_total"`    // Total inventory (renamed from QuantityAvailable for clarity)
	QuantityReserved int64                `json:"quantity_reserved"` // Reserved by pending sales
	QuantitySellable int64                `json:"quantity_sellable"` // Available for new sales (total - reserved)
	CostPrice         float64              `json:"cost_price"`
	ExpiryDate        string               `json:"expiry_date"`
	ManufacturingDate *string              `json:"manufacturing_date,omitempty"`
	BatchNumber       *string              `json:"batch_number,omitempty"`
	SellingPrice      *SellingPriceInfo    `json:"selling_price,omitempty"`
	AlternatePrices   []AlternatePriceInfo `json:"alternate_prices,omitempty"`
	TaxConfig         BatchTaxConfig       `json:"tax_config"`
	Margin            *MarginInfo          `json:"margin,omitempty"`
}

// VariantInfoForSales represents variant info for sales context
type VariantInfoForSales struct {
	ID          string   `json:"id"`
	VariantName string   `json:"variant_name"`
	SKU         string   `json:"sku"`
	Barcode     *string  `json:"barcode,omitempty"`
	BrandName   *string  `json:"brand_name,omitempty"`
	Quantity    string   `json:"quantity"`
	PackSize    *string  `json:"pack_size,omitempty"`
	Images      []string `json:"images,omitempty"`
	HSNCode     *string  `json:"hsn_code,omitempty"`
	IsActive    bool     `json:"is_active"`
}

// ProductInfoForSales represents product info for sales context
type ProductInfoForSales struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Category    *string `json:"category,omitempty"`
	Description *string `json:"description,omitempty"`
}

// SellingPriceInfo represents the primary selling price
type SellingPriceInfo struct {
	PriceID       string  `json:"price_id"`
	Price         float64 `json:"price"`
	PriceType     string  `json:"price_type"`
	Currency      string  `json:"currency"`
	EffectiveFrom string  `json:"effective_from"`
	EffectiveTo   *string `json:"effective_to,omitempty"`
	IsActive      bool    `json:"is_active"`
}

// AlternatePriceInfo represents alternative pricing tiers
type AlternatePriceInfo struct {
	Price       float64 `json:"price"`
	PriceType   string  `json:"price_type"`
	MinQuantity *int64  `json:"min_quantity,omitempty"`
}

// BatchTaxConfig represents tax configuration for a batch
type BatchTaxConfig struct {
	CGSTRate     float64  `json:"cgst_rate"`
	SGSTRate     float64  `json:"sgst_rate"`
	TotalGSTRate float64  `json:"total_gst_rate"`
	IsTaxExempt  bool     `json:"is_tax_exempt"`
	CustomTaxes  []string `json:"custom_taxes,omitempty"`
	HSNCode      *string  `json:"hsn_code,omitempty"`
}

// MarginInfo represents margin calculation
type MarginInfo struct {
	CostPrice        float64 `json:"cost_price"`
	SellingPrice     float64 `json:"selling_price"`
	MarginAmount     float64 `json:"margin_amount"`
	MarginPercentage float64 `json:"margin_percentage"`
}

// GlobalTaxConfig represents global tax configuration
type GlobalTaxConfig struct {
	DefaultCGSTRate        float64       `json:"default_cgst_rate"`
	DefaultSGSTRate        float64       `json:"default_sgst_rate"`
	TaxCalculationMethod   string        `json:"tax_calculation_method"`
	ActiveTaxes            []ActiveTaxInfo `json:"active_taxes"`
}

// ActiveTaxInfo represents an active tax configuration
type ActiveTaxInfo struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	TaxType  string  `json:"tax_type"`
	CGSTRate float64 `json:"cgst_rate"`
	SGSTRate float64 `json:"sgst_rate"`
	IsActive bool    `json:"is_active"`
}

// DiscountPolicy represents a discount policy for sales
type DiscountPolicy struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	DiscountType          string   `json:"discount_type"`
	DiscountValue         float64  `json:"discount_value"`
	MinQuantity           *int64   `json:"min_quantity,omitempty"`
	MinAmount             *float64 `json:"min_amount,omitempty"`
	ApplicableCategories  []string `json:"applicable_categories,omitempty"`
	StartDate             string   `json:"start_date"`
	EndDate               string   `json:"end_date"`
	IsActive              bool     `json:"is_active"`
}

// RefundPolicyInfo represents a refund policy
type RefundPolicyInfo struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Description      *string  `json:"description,omitempty"`
	RefundPercentage float64  `json:"refund_percentage"`
	ValidDays        int      `json:"valid_days"`
	Conditions       []string `json:"conditions,omitempty"`
	IsActive         bool     `json:"is_active"`
}

// PaymentMethodInfo represents a payment method
type PaymentMethodInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"` // cash, card, upi, credit
	CreditDays *int   `json:"credit_days,omitempty"`
	IsActive   bool   `json:"is_active"`
}

// SalesContextMetadata represents metadata for sales context
type SalesContextMetadata struct {
	TotalProducts                int       `json:"total_products"`
	TotalVariants                int       `json:"total_variants"`
	TotalBatches                 int       `json:"total_batches"`
	TotalStockValue              float64   `json:"total_stock_value"`
	WarehouseCapacityUsedPercent float64   `json:"warehouse_capacity_used_percent"`
	ReadTimestamp                string    `json:"read_timestamp"`
	ConsistencyToken             string    `json:"consistency_token"`
	ExpiresAt                    time.Time `json:"expires_at"`
}

// === Purchase Order Detail Response Types ===

// PODetailRequest represents query parameters for PO detail aggregation
type PODetailRequest struct {
	Include string `form:"include"` // Comma-separated: collaborator,warehouse,items,grns,inventory,payments
}

// PODetailResponse represents the aggregated purchase order detail response
type PODetailResponse struct {
	PurchaseOrder POInfo            `json:"purchase_order"`
	Collaborator  *CollaboratorInfo `json:"collaborator,omitempty"`
	Warehouse     *WarehouseInfo    `json:"warehouse,omitempty"`
	Items         []POItemDetail    `json:"items,omitempty"`
	GRNs          []GRNDetail       `json:"grns,omitempty"`
	Payments      []POPaymentDetail `json:"payments,omitempty"`
	Summary       POSummary         `json:"summary"`
	Timeline      []POTimelineEvent `json:"timeline,omitempty"`
	Metadata      ResponseMetadata  `json:"metadata"`
}

// POInfo represents purchase order information
type POInfo struct {
	ID                   string  `json:"id"`
	PONumber             string  `json:"po_number"`
	CollaboratorID       string  `json:"collaborator_id"`
	WarehouseID          string  `json:"warehouse_id"`
	Status               string  `json:"status"`
	OrderDate            string  `json:"order_date"`
	ExpectedDeliveryDate string  `json:"expected_delivery_date"`
	ActualDeliveryDate   *string `json:"actual_delivery_date,omitempty"`
	TotalAmount          float64 `json:"total_amount"`
	PaidAmount           float64 `json:"paid_amount"`
	PendingAmount        float64 `json:"pending_amount"`
	PaymentStatus        string  `json:"payment_status"`
	Currency             string  `json:"currency"`
	ExternalOrderID      *string `json:"external_order_id,omitempty"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
}

// POItemDetail represents a purchase order item with variant and product info
type POItemDetail struct {
	ID               string              `json:"id"`
	VariantID        string              `json:"variant_id"`
	Variant          *VariantInfoForPO   `json:"variant,omitempty"`
	Product          *ProductBasicInfo   `json:"product,omitempty"`
	OrderedQuantity  int64               `json:"ordered_quantity"`
	ReceivedQuantity int64               `json:"received_quantity"`
	PendingQuantity  int64               `json:"pending_quantity"`
	UnitCost         float64             `json:"unit_cost"`
	TotalCost        float64             `json:"total_cost"`
	ReceivedStatus   string              `json:"received_status"` // pending, partially_received, fully_received
}

// VariantInfoForPO represents variant info for purchase order context
type VariantInfoForPO struct {
	ID          string   `json:"id"`
	VariantName string   `json:"variant_name"`
	SKU         string   `json:"sku"`
	BrandName   *string  `json:"brand_name,omitempty"`
	Quantity    string   `json:"quantity"`
	PackSize    *string  `json:"pack_size,omitempty"`
	Images      []string `json:"images,omitempty"`
}

// GRNDetail represents a goods receipt note with items and inventory
type GRNDetail struct {
	ID               string               `json:"id"`
	GRNNumber        string               `json:"grn_number"`
	POID             string               `json:"purchase_order_id"`
	ReceivedDate     string               `json:"received_date"`
	Status           string               `json:"status"`
	QualityStatus    string               `json:"quality_status"`
	ReceivedBy       string               `json:"received_by"`
	Remarks          *string              `json:"remarks,omitempty"`
	Items            []GRNItemDetail      `json:"items,omitempty"`
	InventoryCreated []InventoryCreated   `json:"inventory_created,omitempty"`
}

// GRNItemDetail represents a GRN item with quantities
type GRNItemDetail struct {
	POItemID         string  `json:"po_item_id"`
	VariantID        string  `json:"variant_id"`
	OrderedQuantity  int64   `json:"ordered_quantity"`
	ReceivedQuantity int64   `json:"received_quantity"`
	AcceptedQuantity int64   `json:"accepted_quantity"`
	RejectedQuantity int64   `json:"rejected_quantity"`
	UnitCost         float64 `json:"unit_cost"`
	TotalCost        float64 `json:"total_cost"`
	ExpiryDate       string  `json:"expiry_date"`
	BatchNumber      *string `json:"batch_number,omitempty"`
}

// InventoryCreated represents inventory batch created from GRN
type InventoryCreated struct {
	BatchID           string  `json:"batch_id"`
	VariantID         string  `json:"variant_id"`
	WarehouseID       string  `json:"warehouse_id"`
	Quantity          int64   `json:"quantity"`
	CostPrice         float64 `json:"cost_price"`
	ExpiryDate        string  `json:"expiry_date"`
	ManufacturingDate *string `json:"manufacturing_date,omitempty"`
	BatchNumber       *string `json:"batch_number,omitempty"`
}

// POPaymentDetail represents a payment for a purchase order
type POPaymentDetail struct {
	ID              string  `json:"id"`
	POID            string  `json:"purchase_order_id"`
	PaymentDate     string  `json:"payment_date"`
	Amount          float64 `json:"amount"`
	PaymentMethod   string  `json:"payment_method"`
	ReferenceNumber *string `json:"reference_number,omitempty"`
	Notes           *string `json:"notes,omitempty"`
	Status          string  `json:"status"`
	CreatedBy       string  `json:"created_by"`
}

// POSummary represents purchase order summary calculations
type POSummary struct {
	TotalOrderValue      float64 `json:"total_order_value"`
	TotalReceivedValue   float64 `json:"total_received_value"`
	TotalPendingValue    float64 `json:"total_pending_value"`
	TotalRejectedValue   float64 `json:"total_rejected_value"`
	CompletionPercentage float64 `json:"completion_percentage"`
	TotalItemsOrdered    int64   `json:"total_items_ordered"`
	TotalItemsReceived   int64   `json:"total_items_received"`
	TotalItemsPending    int64   `json:"total_items_pending"`
	PaymentStatus        string  `json:"payment_status"`
	FulfillmentStatus    string  `json:"fulfillment_status"` // pending, partially_received, fully_received
}

// POTimelineEvent represents an event in the PO timeline
type POTimelineEvent struct {
	Timestamp   string  `json:"timestamp"`
	Event       string  `json:"event"` // purchase_order_created, payment_received, grn_created, status_changed
	Description string  `json:"description"`
	Actor       *string `json:"actor,omitempty"`
}

// === Inventory List Response Types ===

// InventoryListRequest represents query parameters for inventory list
type InventoryListRequest struct {
	WarehouseID       string `form:"warehouse_id"`
	VariantID         string `form:"variant_id"`
	ProductID         string `form:"product_id"`
	Category          string `form:"category"`
	InStockOnly       bool   `form:"in_stock_only"`
	ExpiringSoon      bool   `form:"expiring_soon"`
	LowStockThreshold *int64 `form:"low_stock_threshold"`
	Include           string `form:"include"` // variant,product,warehouse,prices,taxes
	SortBy            string `form:"sort_by"` // expiry_date, quantity, cost_price
	SortOrder         string `form:"sort_order"` // asc, desc
	Limit             int    `form:"limit"`
	Offset            int    `form:"offset"`
}

// InventoryListResponse represents the aggregated inventory list response
type InventoryListResponse struct {
	Batches    []BatchWithContext   `json:"batches"`
	Pagination InventoryPagination  `json:"pagination"`
	Summary    InventorySummary     `json:"summary"`
	Metadata   InventoryListMetadata `json:"metadata"`
}

// BatchWithContext represents an inventory batch with full context
type BatchWithContext struct {
	ID              string               `json:"id"`
	Warehouse       *WarehouseBasicInfo  `json:"warehouse,omitempty"`
	Variant         *VariantInfoForSales `json:"variant,omitempty"`
	Product         *ProductInfoForSales `json:"product,omitempty"`
	QuantityDetails QuantityDetails      `json:"quantity_details"`
	Pricing         *BatchPricing        `json:"pricing,omitempty"`
	BatchInfo       BatchDetails         `json:"batch_info"`
	TaxConfig       *BatchTaxConfig      `json:"tax_config,omitempty"`
	Metadata        BatchMetadata        `json:"metadata"`
}

// WarehouseBasicInfo represents minimal warehouse information
type WarehouseBasicInfo struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Location *WarehouseLocationInfo `json:"location,omitempty"`
}

// WarehouseLocationInfo represents warehouse location
type WarehouseLocationInfo struct {
	City  *string `json:"city,omitempty"`
	State *string `json:"state,omitempty"`
}

// QuantityDetails represents quantity breakdown for a batch
type QuantityDetails struct {
	TotalQuantity     int64 `json:"total_quantity"`
	AvailableQuantity int64 `json:"available_quantity"`
	ReservedQuantity  int64 `json:"reserved_quantity"`
	SoldQuantity      int64 `json:"sold_quantity"`
	InStock           bool  `json:"in_stock"`
}

// BatchPricing represents pricing information for a batch
type BatchPricing struct {
	CostPrice     float64                   `json:"cost_price"`
	SellingPrices *BatchSellingPrices       `json:"selling_prices,omitempty"`
	Margin        *BatchMargin              `json:"margin,omitempty"`
	Currency      string                    `json:"currency"`
}

// BatchSellingPrices represents selling prices by type
type BatchSellingPrices struct {
	Retail    *float64 `json:"retail,omitempty"`
	Wholesale *float64 `json:"wholesale,omitempty"`
	Bulk      *float64 `json:"bulk,omitempty"`
}

// BatchMargin represents margin calculation
type BatchMargin struct {
	RetailMargin           float64 `json:"retail_margin"`
	RetailMarginPercentage float64 `json:"retail_margin_percentage"`
}

// BatchDetails represents batch-specific information
type BatchDetails struct {
	BatchNumber       *string `json:"batch_number,omitempty"`
	ManufacturingDate *string `json:"manufacturing_date,omitempty"`
	ExpiryDate        string  `json:"expiry_date"`
	DaysUntilExpiry   int     `json:"days_until_expiry"`
	ExpiryStatus      string  `json:"expiry_status"` // good, warning, critical, expired
}

// BatchMetadata represents batch metadata
type BatchMetadata struct {
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	CreatedBy *string `json:"created_by,omitempty"`
}

// InventoryPagination represents pagination info
type InventoryPagination struct {
	Total      int  `json:"total"`
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
	HasMore    bool `json:"has_more"`
	NextOffset *int `json:"next_offset,omitempty"`
}

// InventorySummary represents summary statistics
type InventorySummary struct {
	TotalBatches       int     `json:"total_batches"`
	TotalProducts      int     `json:"total_products"`
	TotalVariants      int     `json:"total_variants"`
	TotalStockQuantity int64   `json:"total_stock_quantity"`
	TotalStockValue    float64 `json:"total_stock_value"`
	ExpiringSoonCount  int     `json:"expiring_soon_count"`
	LowStockCount      int     `json:"low_stock_count"`
	ZeroStockCount     int     `json:"zero_stock_count"`
}

// InventoryListMetadata represents metadata for inventory list
type InventoryListMetadata struct {
	ReadTimestamp  string                 `json:"read_timestamp"`
	FiltersApplied map[string]interface{} `json:"filters_applied,omitempty"`
}
