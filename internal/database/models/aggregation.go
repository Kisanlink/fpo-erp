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
	Include     string `form:"include"`       // Comma-separated: variants,prices,inventory,collaborators,taxes
	WarehouseID string `form:"warehouse_id"`  // Filter inventory by warehouse
	PriceType   string `form:"price_type"`    // Filter prices: retail, wholesale, bulk, all
	ActiveOnly  bool   `form:"active_only"`   // Show only active variants (default: true)
	InStockOnly bool   `form:"in_stock_only"` // Show only variants with stock
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
// Uses existing ProductResponse instead of custom ProductInfo
type ProductDetailResponse struct {
	Product      ProductResponse        `json:"product"`
	Collaborator *CollaboratorResponse  `json:"collaborator,omitempty"`
	Variants     []VariantWithAggData   `json:"variants,omitempty"`
	Metadata     ProductMetadata        `json:"metadata"`
}

// VariantWithAggData extends ProductVariantResponse with aggregation-specific computed data
// This embeds the existing ProductVariantResponse and adds stock/pricing summaries
type VariantWithAggData struct {
	ProductVariantResponse
	StockSummary     *StockSummary     `json:"stock_summary,omitempty"`
	WarehouseStock   []WarehouseStock  `json:"warehouse_stock,omitempty"`
	TaxConfiguration *TaxConfiguration `json:"tax_configuration,omitempty"`
	PricesSummary    *VariantPrices    `json:"prices_summary,omitempty"` // Aggregated price summary
}

// VariantPrices represents pricing information for a variant
type VariantPrices struct {
	Currency       string     `json:"currency"`
	HasActivePrice bool       `json:"has_active_price"`
	RetailPrice    *PriceInfo `json:"retail_price,omitempty"`
	WholesalePrice *PriceInfo `json:"wholesale_price,omitempty"`
	BulkPrice      *PriceInfo `json:"bulk_price,omitempty"`
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

// TaxConfiguration represents GST tax settings for a variant
// Simplified for GST-only system - tax rate from ProductVariant.GSTRate
type TaxConfiguration struct {
	GSTRate  float64 `json:"gst_rate"`  // Total GST rate from variant
	CGSTRate float64 `json:"cgst_rate"` // CGST (50% of GSTRate) for intra-state
	SGSTRate float64 `json:"sgst_rate"` // SGST (50% of GSTRate) for intra-state
	HSNCode  string  `json:"hsn_code"`  // HSN code from variant
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

// VariantDetailWithProduct extends ProductVariantResponse with parent product info
type VariantDetailWithProduct struct {
	VariantWithAggData                           // Embed variant with aggregation data
	Product            *ProductResponse          `json:"product,omitempty"`
	Collaborator       *CollaboratorResponse     `json:"collaborator,omitempty"`
}

// ResponseMetadata represents common metadata
type ResponseMetadata struct {
	ReadTimestamp    string `json:"read_timestamp"`
	ConsistencyToken string `json:"consistency_token,omitempty"`
}

// === Sales Context Response Types ===

// SalesContextResponse represents the aggregated sales context response
// Uses existing WarehouseResponse instead of custom WarehouseInfo
type SalesContextResponse struct {
	Warehouse              WarehouseResponse      `json:"warehouse"`
	AvailableInventory     []InventoryWithPricing `json:"available_inventory"`
	GlobalTaxConfiguration GlobalTaxConfig        `json:"global_tax_configuration"`
	DiscountPolicies       []DiscountPolicy       `json:"discount_policies"`
	RefundPolicies         []RefundPolicyInfo     `json:"refund_policies"`
	PaymentMethods         []PaymentMethodInfo    `json:"payment_methods"`
	Metadata               SalesContextMetadata   `json:"metadata"`
}

// InventoryWithPricing represents inventory batch with pricing and product info
// Uses existing ProductVariantResponse and ProductResponse instead of custom types
type InventoryWithPricing struct {
	BatchID           string                  `json:"batch_id"`
	VariantID         string                  `json:"variant_id"`
	Variant           ProductVariantResponse  `json:"variant"`
	Product           ProductResponse         `json:"product"`
	QuantityTotal     int64                   `json:"quantity_total"`    // Total inventory
	QuantityReserved  int64                   `json:"quantity_reserved"` // Reserved by pending sales
	QuantitySellable  int64                   `json:"quantity_sellable"` // Available for new sales
	CostPrice         float64                 `json:"cost_price"`
	ExpiryDate        string                  `json:"expiry_date"`
	ManufacturingDate *string                 `json:"manufacturing_date,omitempty"`
	BatchNumber       *string                 `json:"batch_number,omitempty"`
	SellingPrice      *SellingPriceInfo       `json:"selling_price,omitempty"`
	AlternatePrices   []AlternatePriceInfo    `json:"alternate_prices,omitempty"`
	TaxConfig         BatchTaxConfig          `json:"tax_config"`
	Margin            *MarginInfo             `json:"margin,omitempty"`
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

// BatchTaxConfig represents GST tax configuration for a batch
// Tax rate is from ProductVariant.GSTRate, not batch level
type BatchTaxConfig struct {
	GSTRate      float64 `json:"gst_rate"`       // Total GST rate from variant
	CGSTRate     float64 `json:"cgst_rate"`      // CGST (50% of GSTRate) for intra-state
	SGSTRate     float64 `json:"sgst_rate"`      // SGST (50% of GSTRate) for intra-state
	TotalGSTRate float64 `json:"total_gst_rate"` // Same as GSTRate for compatibility
	HSNCode      string  `json:"hsn_code"`       // HSN code from variant
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
	DefaultCGSTRate      float64         `json:"default_cgst_rate"`
	DefaultSGSTRate      float64         `json:"default_sgst_rate"`
	TaxCalculationMethod string          `json:"tax_calculation_method"`
	ActiveTaxes          []ActiveTaxInfo `json:"active_taxes"`
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
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	DiscountType         string   `json:"discount_type"`
	DiscountValue        float64  `json:"discount_value"`
	MinQuantity          *int64   `json:"min_quantity,omitempty"`
	MinAmount            *float64 `json:"min_amount,omitempty"`
	ApplicableCategories []string `json:"applicable_categories,omitempty"`
	StartDate            string   `json:"start_date"`
	EndDate              string   `json:"end_date"`
	IsActive             bool     `json:"is_active"`
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
// Uses existing PurchaseOrderResponse, CollaboratorResponse, WarehouseResponse
type PODetailResponse struct {
	PurchaseOrder PurchaseOrderResponse          `json:"purchase_order"`
	Collaborator  *CollaboratorResponse          `json:"collaborator,omitempty"`
	Warehouse     *WarehouseResponse             `json:"warehouse,omitempty"`
	Items         []PurchaseOrderItemResponse    `json:"items,omitempty"`
	GRNs          []GRNDetail                    `json:"grns,omitempty"`
	Payments      []POPaymentDetail              `json:"payments,omitempty"`
	Summary       POSummary                      `json:"summary"`
	Timeline      []POTimelineEvent              `json:"timeline,omitempty"`
	Metadata      ResponseMetadata               `json:"metadata"`
}

// GRNDetail represents a goods receipt note with items and inventory
type GRNDetail struct {
	ID               string             `json:"id"`
	GRNNumber        string             `json:"grn_number"`
	POID             string             `json:"purchase_order_id"`
	ReceivedDate     string             `json:"received_date"`
	Status           string             `json:"status"`
	QualityStatus    string             `json:"quality_status"`
	ReceivedBy       string             `json:"received_by"`
	Remarks          *string            `json:"remarks,omitempty"`
	Items            []GRNItemDetail    `json:"items,omitempty"`
	InventoryCreated []InventoryCreated `json:"inventory_created,omitempty"`
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
	Include           string `form:"include"`    // variant,product,warehouse,prices,taxes
	SortBy            string `form:"sort_by"`    // expiry_date, quantity, cost_price
	SortOrder         string `form:"sort_order"` // asc, desc
	Limit             int    `form:"limit"`
	Offset            int    `form:"offset"`
}

// InventoryListResponse represents the aggregated inventory list response
type InventoryListResponse struct {
	Batches    []BatchWithContext    `json:"batches"`
	Pagination InventoryPagination   `json:"pagination"`
	Summary    InventorySummary      `json:"summary"`
	Metadata   InventoryListMetadata `json:"metadata"`
}

// BatchWithContext represents an inventory batch with full context
// Uses existing WarehouseResponse, ProductVariantResponse, ProductResponse
type BatchWithContext struct {
	ID              string                  `json:"id"`
	Warehouse       *WarehouseResponse      `json:"warehouse,omitempty"`
	Variant         *ProductVariantResponse `json:"variant,omitempty"`
	Product         *ProductResponse        `json:"product,omitempty"`
	QuantityDetails QuantityDetails         `json:"quantity_details"`
	Pricing         *BatchPricing           `json:"pricing,omitempty"`
	BatchInfo       BatchDetails            `json:"batch_info"`
	TaxConfig       *BatchTaxConfig         `json:"tax_config,omitempty"`
	Metadata        BatchMetadata           `json:"metadata"`
}

// QuantityDetails represents quantity breakdown for a batch
type QuantityDetails struct {
	TotalQuantity     int64 `json:"total_quantity"`
	AvailableQuantity int64 `json:"available_quantity"`
	ReservedQuantity  int64 `json:"reserved_quantity"`
	InStock           bool  `json:"in_stock"`
}

// BatchPricing represents pricing information for a batch
type BatchPricing struct {
	CostPrice     float64             `json:"cost_price"`
	SellingPrices *BatchSellingPrices `json:"selling_prices,omitempty"`
	Margin        *BatchMargin        `json:"margin,omitempty"`
	Currency      string              `json:"currency"`
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
