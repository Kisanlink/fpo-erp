package constants

// Table identifiers for hash-based ID generation
// All identifiers must be exactly 4 characters as required by kisanlink-db
const (
	// Core entities
	TableProduct   = "PROD" // Products
	TableWarehouse = "WHSE" // Warehouses

	// Sales related
	TableSale        = "SALE" // Sales
	TableSaleItem    = "SITM" // Sale Items
	TableSaleSummary = "SSUM" // Sale Summaries

	// Inventory related
	TableBatch       = "BATC" // Inventory Batches
	TableTransaction = "TRAN" // Inventory Transactions

	// Pricing and discounts
	TablePrice       = "PRIC" // Product Prices
	TableDiscount    = "DISC" // Discounts
	TableDiscountUse = "DUSE" // Discount Usage

	// Tax related
	TableTax        = "TAXX" // Tax
	TableTaxTier    = "TIER" // Tax Tiers
	TableTaxApp     = "TAPP" // Tax Applications
	TableTaxSummary = "TSUM" // Tax Summaries

	// Returns related
	TableReturn        = "RETN" // Returns
	TableReturnItem    = "RITM" // Return Items
	TableReturnSummary = "RSUM" // Return Summaries
	TableRefundPolicy  = "RPOL" // Refund Policies

	// Payment and attachments
	TableBankPayment = "BPAY" // Bank Payments
	TableAttachment  = "ATCH" // Attachments

	// Procurement Module
	TableCollaborator        = "CLAB" // Collaborators/Vendors
	TableCollaboratorProduct = "CPRD" // Collaborator-Product junction
	TableProductVariant      = "PVAR" // Product Variants
	TablePurchaseOrder       = "PORD" // Purchase Orders
	TablePurchaseOrderItem   = "POIM" // PO Items
	TableGRN                 = "GRNX" // Goods Receipt Notes
	TableGRNItem             = "GRIT" // GRN Items
)
