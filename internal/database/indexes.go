package database

import (
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
)

// IndexDefinition represents a database index configuration
type IndexDefinition struct {
	Table   string // Table name
	Columns string // Column(s) - comma-separated for composite
	Name    string // Index name
	Unique  bool   // Is unique index
	Where   string // Partial index WHERE clause (optional)
}

// CreateApplicationIndexes creates all application performance indexes
// This runs as Phase 3 of the migration process, after GORM AutoMigrate
// Errors are logged but don't block startup - indexes are optimization, not critical
func CreateApplicationIndexes(db *gorm.DB) error {
	log.Println("📊 Creating application performance indexes...")

	indexes := getIndexDefinitions()
	created := 0
	skipped := 0
	failed := 0

	for _, idx := range indexes {
		sql := buildIndexSQL(idx)
		if err := db.Exec(sql).Error; err != nil {
			// Check if it's a "already exists" error - that's OK
			if strings.Contains(err.Error(), "already exists") {
				skipped++
				continue
			}
			log.Printf("⚠️  Warning: Failed to create index %s: %v", idx.Name, err)
			failed++
			continue
		}
		created++
	}

	log.Printf("✓ Index creation complete: %d created, %d already existed, %d failed", created, skipped, failed)

	if failed > 0 {
		return fmt.Errorf("%d indexes failed to create", failed)
	}
	return nil
}

// buildIndexSQL generates the CREATE INDEX SQL statement
func buildIndexSQL(idx IndexDefinition) string {
	var sb strings.Builder

	if idx.Unique {
		sb.WriteString("CREATE UNIQUE INDEX IF NOT EXISTS ")
	} else {
		sb.WriteString("CREATE INDEX IF NOT EXISTS ")
	}

	sb.WriteString(idx.Name)
	sb.WriteString(" ON ")
	sb.WriteString(idx.Table)
	sb.WriteString("(")
	sb.WriteString(idx.Columns)
	sb.WriteString(")")

	if idx.Where != "" {
		sb.WriteString(" WHERE ")
		sb.WriteString(idx.Where)
	}

	return sb.String()
}

// getIndexDefinitions returns all 75 index definitions
func getIndexDefinitions() []IndexDefinition {
	return []IndexDefinition{
		// ============================================================================
		// CATEGORIES TABLE INDEXES
		// ============================================================================
		{Table: "categories", Columns: "name", Name: "idx_categories_name", Unique: true},
		{Table: "categories", Columns: "LOWER(name)", Name: "idx_categories_name_lower"},

		// ============================================================================
		// SUBCATEGORIES TABLE INDEXES
		// Note: idx_subcategory_name_category is created by GORM from model tags
		// ============================================================================
		{Table: "subcategories", Columns: "category_id", Name: "idx_subcategories_category_id"},
		{Table: "subcategories", Columns: "LOWER(name)", Name: "idx_subcategories_name_lower"},

		// ============================================================================
		// PRODUCTS TABLE INDEXES
		// ============================================================================
		{Table: "products", Columns: "external_id", Name: "idx_products_external_id", Where: "external_id IS NOT NULL"},
		{Table: "products", Columns: "LOWER(name)", Name: "idx_products_name_lower"},
		{Table: "products", Columns: "created_at DESC", Name: "idx_products_created_at"},
		{Table: "products", Columns: "category_id", Name: "idx_products_category_id", Where: "category_id IS NOT NULL"},
		{Table: "products", Columns: "subcategory_id", Name: "idx_products_subcategory_id", Where: "subcategory_id IS NOT NULL"},

		// ============================================================================
		// PRODUCT VARIANTS TABLE INDEXES
		// ============================================================================
		{Table: "product_variants", Columns: "product_id", Name: "idx_product_variants_product_id"},
		{Table: "product_variants", Columns: "sku", Name: "idx_product_variants_sku", Where: "sku IS NOT NULL"},
		{Table: "product_variants", Columns: "barcode", Name: "idx_product_variants_barcode", Where: "barcode IS NOT NULL"},
		{Table: "product_variants", Columns: "external_id", Name: "idx_product_variants_external_id", Where: "external_id IS NOT NULL"},
		{Table: "product_variants", Columns: "product_id, is_active", Name: "idx_product_variants_product_active", Where: "is_active = true"},
		{Table: "product_variants", Columns: "brand_name", Name: "idx_product_variants_brand", Where: "brand_name IS NOT NULL"},
		{Table: "product_variants", Columns: "collaborator_id", Name: "idx_product_variants_collaborator", Where: "collaborator_id IS NOT NULL"},
		{Table: "product_variants", Columns: "variant_name", Name: "idx_product_variants_name"},
		{Table: "product_variants", Columns: "created_at DESC", Name: "idx_product_variants_created_at"},

		// ============================================================================
		// PRODUCT PRICES TABLE INDEXES
		// ============================================================================
		{Table: "product_prices", Columns: "variant_id, price_type", Name: "idx_product_prices_variant_type"},
		{Table: "product_prices", Columns: "variant_id, is_active", Name: "idx_product_prices_active", Where: "is_active = true"},
		{Table: "product_prices", Columns: "variant_id, effective_from DESC, effective_to", Name: "idx_product_prices_effective"},
		{Table: "product_prices", Columns: "price_type", Name: "idx_product_prices_type"},

		// ============================================================================
		// INVENTORY BATCHES TABLE INDEXES
		// ============================================================================
		{Table: "inventory_batches", Columns: "variant_id, expiry_date ASC, total_quantity DESC", Name: "idx_inventory_batches_fefo", Where: "total_quantity > 0"},
		{Table: "inventory_batches", Columns: "warehouse_id, variant_id, total_quantity DESC", Name: "idx_inventory_batches_warehouse", Where: "total_quantity > 0"},
		{Table: "inventory_batches", Columns: "variant_id, total_quantity DESC", Name: "idx_inventory_batches_stock", Where: "total_quantity > 0"},
		{Table: "inventory_batches", Columns: "expiry_date ASC", Name: "idx_inventory_batches_expiry", Where: "total_quantity > 0 AND expiry_date IS NOT NULL"},
		{Table: "inventory_batches", Columns: "warehouse_id, total_quantity ASC", Name: "idx_inventory_batches_low_stock", Where: "total_quantity > 0 AND total_quantity < 100"},
		{Table: "inventory_batches", Columns: "batch_number", Name: "idx_inventory_batches_batch_number", Where: "batch_number IS NOT NULL"},
		{Table: "inventory_batches", Columns: "grn_item_id", Name: "idx_inventory_batches_grn", Where: "grn_item_id IS NOT NULL"},

		// ============================================================================
		// INVENTORY TRANSACTIONS TABLE INDEXES
		// ============================================================================
		{Table: "inventory_transactions", Columns: "batch_id", Name: "idx_inventory_transactions_batch"},
		{Table: "inventory_transactions", Columns: "transaction_type", Name: "idx_inventory_transactions_type"},
		{Table: "inventory_transactions", Columns: "related_entity_id", Name: "idx_inventory_transactions_related", Where: "related_entity_id IS NOT NULL"},
		{Table: "inventory_transactions", Columns: "occurred_at DESC", Name: "idx_inventory_transactions_date"},

		// ============================================================================
		// WAREHOUSES TABLE INDEXES
		// ============================================================================
		{Table: "warehouses", Columns: "name", Name: "idx_warehouses_name"},
		{Table: "warehouses", Columns: "address_id", Name: "idx_warehouses_address", Where: "address_id IS NOT NULL"},

		// ============================================================================
		// COLLABORATORS TABLE INDEXES
		// ============================================================================
		{Table: "collaborators", Columns: "is_active", Name: "idx_collaborators_active", Where: "is_active = true"},
		{Table: "collaborators", Columns: "company_name", Name: "idx_collaborators_company_name"},
		{Table: "collaborators", Columns: "contact_person", Name: "idx_collaborators_contact_person", Where: "contact_person IS NOT NULL"},
		{Table: "collaborators", Columns: "email", Name: "idx_collaborators_email", Where: "email IS NOT NULL"},
		{Table: "collaborators", Columns: "gst_number", Name: "idx_collaborators_gst", Where: "gst_number IS NOT NULL"},
		{Table: "collaborators", Columns: "external_id", Name: "idx_collaborators_external", Where: "external_id IS NOT NULL"},

		// ============================================================================
		// SALES TABLE INDEXES
		// ============================================================================
		{Table: "sales", Columns: "warehouse_id", Name: "idx_sales_warehouse"},
		{Table: "sales", Columns: "sale_date DESC", Name: "idx_sales_date"},
		{Table: "sales", Columns: "status", Name: "idx_sales_status"},
		{Table: "sales", Columns: "status, sale_date DESC", Name: "idx_sales_status_date"},
		{Table: "sales", Columns: "payment_mode", Name: "idx_sales_payment_mode"},
		{Table: "sales", Columns: "sale_type", Name: "idx_sales_type"},
		{Table: "sales", Columns: "farmer_id", Name: "idx_sales_farmer", Where: "farmer_id IS NOT NULL"},
		{Table: "sales", Columns: "cancelled_at DESC", Name: "idx_sales_cancelled", Where: "cancelled_at IS NOT NULL"},
		{Table: "sales", Columns: "created_at DESC", Name: "idx_sales_created_at"},

		// ============================================================================
		// SALE ITEMS TABLE INDEXES
		// ============================================================================
		{Table: "sale_items", Columns: "sale_id", Name: "idx_sale_items_sale"},
		{Table: "sale_items", Columns: "batch_id", Name: "idx_sale_items_batch"},

		// ============================================================================
		// SALE CANCELLATIONS TABLE INDEXES
		// ============================================================================
		{Table: "sale_cancellations", Columns: "sale_id", Name: "idx_sale_cancellations_sale"},
		{Table: "sale_cancellations", Columns: "cancelled_at DESC", Name: "idx_sale_cancellations_date"},
		{Table: "sale_cancellations", Columns: "cancelled_by", Name: "idx_sale_cancellations_user"},

		// ============================================================================
		// SALE CANCELLATION ITEMS TABLE INDEXES
		// ============================================================================
		{Table: "sale_cancellation_items", Columns: "cancellation_id", Name: "idx_sale_cancellation_items_cancellation"},
		{Table: "sale_cancellation_items", Columns: "sale_item_id", Name: "idx_sale_cancellation_items_sale_item"},
		{Table: "sale_cancellation_items", Columns: "batch_id", Name: "idx_sale_cancellation_items_batch"},

		// ============================================================================
		// PURCHASE ORDERS TABLE INDEXES
		// ============================================================================
		{Table: "purchase_orders", Columns: "po_number", Name: "idx_purchase_orders_po_number"},
		{Table: "purchase_orders", Columns: "collaborator_id", Name: "idx_purchase_orders_collaborator"},
		{Table: "purchase_orders", Columns: "warehouse_id", Name: "idx_purchase_orders_warehouse"},
		{Table: "purchase_orders", Columns: "status", Name: "idx_purchase_orders_status"},
		{Table: "purchase_orders", Columns: "status, created_at DESC", Name: "idx_purchase_orders_status_date"},
		{Table: "purchase_orders", Columns: "expected_delivery_date ASC", Name: "idx_purchase_orders_delivery", Where: "status IN ('placed', 'confirmed', 'out_for_delivery')"},
		{Table: "purchase_orders", Columns: "external_order_id", Name: "idx_purchase_orders_external", Where: "external_order_id IS NOT NULL"},
		{Table: "purchase_orders", Columns: "payment_status", Name: "idx_purchase_orders_payment"},

		// ============================================================================
		// PURCHASE ORDER ITEMS TABLE INDEXES
		// ============================================================================
		{Table: "purchase_order_items", Columns: "purchase_order_id", Name: "idx_purchase_order_items_po"},
		{Table: "purchase_order_items", Columns: "variant_id", Name: "idx_purchase_order_items_variant"},
		{Table: "purchase_order_items", Columns: "external_variant_id", Name: "idx_purchase_order_items_external", Where: "external_variant_id IS NOT NULL"},

		// ============================================================================
		// GRN (GOODS RECEIPT NOTES) TABLE INDEXES
		// ============================================================================
		{Table: "goods_receipt_notes", Columns: "grn_number", Name: "idx_grns_grn_number"},
		{Table: "goods_receipt_notes", Columns: "purchase_order_id", Name: "idx_grns_purchase_order"},
		{Table: "goods_receipt_notes", Columns: "quality_status", Name: "idx_grns_quality_status"},
		{Table: "goods_receipt_notes", Columns: "warehouse_id", Name: "idx_grns_warehouse"},
		{Table: "goods_receipt_notes", Columns: "received_date DESC", Name: "idx_grns_received_date"},

		// ============================================================================
		// GRN ITEMS TABLE INDEXES
		// ============================================================================
		{Table: "grn_items", Columns: "grn_id", Name: "idx_grn_items_grn"},
		{Table: "grn_items", Columns: "po_item_id", Name: "idx_grn_items_po_item"},
		{Table: "grn_items", Columns: "inventory_batch_id", Name: "idx_grn_items_batch", Where: "inventory_batch_id IS NOT NULL"},

		// ============================================================================
		// TAXES TABLE INDEXES
		// ============================================================================
		{Table: "taxes", Columns: "is_active", Name: "idx_taxes_active", Where: "is_active = true"},
		{Table: "taxes", Columns: "tax_type", Name: "idx_taxes_type"},
		{Table: "taxes", Columns: "hsn_code", Name: "idx_taxes_hsn", Where: "hsn_code IS NOT NULL"},

		// ============================================================================
		// DISCOUNTS TABLE INDEXES
		// ============================================================================
		{Table: "discounts", Columns: "is_active", Name: "idx_discounts_active", Where: "is_active = true"},
		{Table: "discounts", Columns: "discount_type", Name: "idx_discounts_type"},
		{Table: "discounts", Columns: "valid_from, valid_until", Name: "idx_discounts_validity", Where: "is_active = true"},
		{Table: "discounts", Columns: "priority DESC, created_at DESC", Name: "idx_discounts_priority", Where: "is_active = true"},
		{Table: "discounts", Columns: "code", Name: "idx_discounts_code", Where: "code IS NOT NULL"},

		// ============================================================================
		// DISCOUNT USAGES TABLE INDEXES
		// ============================================================================
		{Table: "discount_usages", Columns: "discount_id", Name: "idx_discount_usages_discount"},
		{Table: "discount_usages", Columns: "sale_id", Name: "idx_discount_usages_sale"},

		// ============================================================================
		// RETURNS TABLE INDEXES
		// ============================================================================
		{Table: "returns", Columns: "sale_id", Name: "idx_returns_sale"},
		{Table: "returns", Columns: "status", Name: "idx_returns_status"},
		{Table: "returns", Columns: "created_at DESC", Name: "idx_returns_date"},

		// ============================================================================
		// RETURN ITEMS TABLE INDEXES
		// ============================================================================
		{Table: "return_items", Columns: "return_id", Name: "idx_return_items_return"},
		{Table: "return_items", Columns: "sale_item_id", Name: "idx_return_items_sale_item"},

		// ============================================================================
		// BANK PAYMENTS TABLE INDEXES
		// ============================================================================
		{Table: "bank_payments", Columns: "sale_id", Name: "idx_bank_payments_sale", Where: "sale_id IS NOT NULL"},
		{Table: "bank_payments", Columns: "return_id", Name: "idx_bank_payments_return", Where: "return_id IS NOT NULL"},
		{Table: "bank_payments", Columns: "payment_date DESC", Name: "idx_bank_payments_date"},
		{Table: "bank_payments", Columns: "transaction_reference", Name: "idx_bank_payments_reference", Where: "transaction_reference IS NOT NULL"},

		// ============================================================================
		// ATTACHMENTS TABLE INDEXES
		// ============================================================================
		{Table: "attachments", Columns: "entity_type, entity_id", Name: "idx_attachments_entity"},

		// ============================================================================
		// WEBHOOK EVENTS TABLE INDEXES
		// ============================================================================
		{Table: "webhook_events", Columns: "event_id", Name: "idx_webhook_events_event_id", Unique: true},
		{Table: "webhook_events", Columns: "event_type", Name: "idx_webhook_events_type"},
		{Table: "webhook_events", Columns: "status", Name: "idx_webhook_events_status"},
		{Table: "webhook_events", Columns: "external_order_id", Name: "idx_webhook_events_external_order", Where: "external_order_id IS NOT NULL"},
		{Table: "webhook_events", Columns: "purchase_order_id", Name: "idx_webhook_events_po", Where: "purchase_order_id IS NOT NULL"},
		{Table: "webhook_events", Columns: "payload_hash", Name: "idx_webhook_events_hash"},

		// ============================================================================
		// WEBHOOK DELIVERY ATTEMPTS TABLE INDEXES
		// ============================================================================
		{Table: "webhook_delivery_attempts", Columns: "webhook_event_id", Name: "idx_webhook_delivery_attempts_event"},

		// ============================================================================
		// COMPOSITE/COVERING INDEXES FOR COMMON QUERY PATTERNS
		// ============================================================================
		{Table: "product_variants", Columns: "product_id, is_active, created_at DESC", Name: "idx_product_variants_list"},
		{Table: "inventory_batches", Columns: "warehouse_id, variant_id, expiry_date ASC, total_quantity DESC", Name: "idx_inventory_batches_sales_context", Where: "total_quantity > 0"},
		{Table: "sales", Columns: "warehouse_id, sale_date DESC, status", Name: "idx_sales_list"},
		{Table: "purchase_orders", Columns: "collaborator_id, warehouse_id, created_at DESC", Name: "idx_purchase_orders_list"},
	}
}
