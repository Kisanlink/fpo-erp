package testutils

import (
	"fmt"

	"gorm.io/gorm"
)

// CreateSQLiteCompatibleTables creates tables with SQLite-compatible types
// This function manually creates tables for models that use PostgreSQL-specific
// types that SQLite doesn't support (timestamptz, numeric(precision,scale), etc.)
func CreateSQLiteCompatibleTables(db *gorm.DB) error {
	// CRITICAL: Use GORM's Exec to stay within GORM's transaction management context
	// Using raw sqlDB bypasses GORM and causes connection issues with :memory: databases

	// ProductPrice table
	if err := db.Exec(`DROP TABLE IF EXISTS product_prices`).Error; err != nil {
		return fmt.Errorf("failed to drop product_prices table: %w", err)
	}

	if err := db.Exec(`
		CREATE TABLE product_prices (
			id TEXT PRIMARY KEY,
			variant_id TEXT NOT NULL,
			price_type TEXT NOT NULL,
			price REAL NOT NULL,
			currency TEXT NOT NULL DEFAULT 'INR',
			effective_from DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			effective_to DATETIME,
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create product_prices table: %w", err)
	}

	// Create index for variant_id
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_variant_price ON product_prices(variant_id)
	`).Error; err != nil {
		return fmt.Errorf("failed to create index idx_variant_price: %w", err)
	}

	// Sale table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS sales`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE sales (
			id TEXT PRIMARY KEY,
			warehouse_id TEXT NOT NULL,
			sale_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			total_amount REAL NOT NULL,
			status TEXT NOT NULL,
			farmer_id TEXT,
			payment_mode TEXT NOT NULL,
			sale_type TEXT NOT NULL,
			is_returned INTEGER DEFAULT 0,
			apply_taxes INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// SaleItem table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS sale_items`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE sale_items (
			id TEXT PRIMARY KEY,
			sale_id TEXT NOT NULL,
			batch_id TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			selling_price REAL NOT NULL,
			line_total REAL NOT NULL,
			cost_price REAL NOT NULL,
			margin REAL NOT NULL,
			cgst_amount REAL DEFAULT 0,
			sgst_amount REAL DEFAULT 0,
			custom_tax_amount REAL DEFAULT 0,
			total_tax_amount REAL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// InventoryTransaction table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS inventory_transactions`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE inventory_transactions (
			id TEXT PRIMARY KEY,
			batch_id TEXT NOT NULL,
			transaction_type TEXT NOT NULL,
			quantity_change INTEGER NOT NULL,
			previous_quantity INTEGER,
			new_quantity INTEGER,
			related_entity_id TEXT,
			performed_by TEXT,
			note TEXT,
			occurred_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// DiscountUsage table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS discount_usages`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE discount_usages (
			id TEXT PRIMARY KEY,
			discount_id TEXT NOT NULL,
			sale_id TEXT NOT NULL,
			get_discount_value REAL,
			used_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			amount REAL NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Attachment table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS attachments`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE attachments (
			id TEXT PRIMARY KEY,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			file_name TEXT NOT NULL,
			file_path TEXT NOT NULL,
			file_size INTEGER NOT NULL,
			content_type TEXT,
			uploaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create index for entity lookup
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_attachments_entity ON attachments(entity_type, entity_id)
	`).Error; err != nil {
		return err
	}

	// PurchaseOrder table
	// Note: Using expected_delivery_date to match the custom naming strategy (SQLiteJSONNamingStrategy)
	// which maps expected_delivery → expected_delivery_date
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS purchase_orders`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE purchase_orders (
			id TEXT PRIMARY KEY,
			po_number TEXT NOT NULL UNIQUE,
			external_order_id TEXT UNIQUE,
			collaborator_id TEXT NOT NULL,
			warehouse_id TEXT NOT NULL,
			order_date DATETIME NOT NULL,
			expected_delivery_date DATETIME NOT NULL,
			actual_delivery_date DATETIME,
			status TEXT NOT NULL,
			total_amount REAL NOT NULL,
			payment_status TEXT NOT NULL,
			paid_amount REAL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create indexes for purchase_orders
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_po_collaborator ON purchase_orders(collaborator_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_po_warehouse ON purchase_orders(warehouse_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_po_status ON purchase_orders(status)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_po_external_order ON purchase_orders(external_order_id)
	`).Error; err != nil {
		return err
	}

	// PurchaseOrderItem table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS purchase_order_items`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE purchase_order_items (
			id TEXT PRIMARY KEY,
			po_id TEXT NOT NULL,
			variant_id TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			unit_price REAL NOT NULL,
			line_total REAL NOT NULL,
			product_name TEXT,
			product_sku TEXT,
			received_quantity INTEGER,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create indexes for purchase_order_items
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_po_item_po ON purchase_order_items(po_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_po_item_variant ON purchase_order_items(variant_id)
	`).Error; err != nil {
		return err
	}

	// Warehouse table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS warehouses`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE warehouses (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			address_id TEXT,
			address_type TEXT,
			house TEXT,
			street TEXT,
			landmark TEXT,
			post_office TEXT,
			subdistrict TEXT,
			district TEXT,
			vtc TEXT,
			state TEXT,
			country TEXT,
			pincode TEXT,
			is_primary_address INTEGER DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Product table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS products`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE products (
			id TEXT PRIMARY KEY,
			external_id TEXT UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create index for external_id
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_product_external_id ON products(external_id)
	`).Error; err != nil {
		return err
	}

	// ProductVariant table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS product_variants`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE product_variants (
			id TEXT PRIMARY KEY,
			product_id TEXT NOT NULL,
			external_id TEXT UNIQUE,
			variant_name TEXT NOT NULL,
			description TEXT,
			quantity TEXT NOT NULL,
			pack_size TEXT NOT NULL,
			sku TEXT UNIQUE,
			barcode TEXT,
			collaborator_id TEXT,
			brand_name TEXT,
			hsn_code TEXT,
			gst_rate REAL,
			images TEXT,
			dosage_instructions TEXT,
			usage_details TEXT,
			is_active INTEGER DEFAULT 1,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create indexes for product_variants
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_product_variant_product ON product_variants(product_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_product_variant_collaborator ON product_variants(collaborator_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_product_collaborator ON product_variants(product_id, collaborator_id)
	`).Error; err != nil {
		return err
	}

	// InventoryBatch table
	// Drop table first to ensure schema is correct
	// IMPORTANT: Created after warehouses and product_variants since it references them
	if err := db.Exec(`DROP TABLE IF EXISTS inventory_batches`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE inventory_batches (
			id TEXT PRIMARY KEY,
			warehouse_id TEXT NOT NULL,
			variant_id TEXT NOT NULL,
			cost_price REAL NOT NULL,
			expiry_date DATETIME NOT NULL,
			total_quantity INTEGER NOT NULL,
			cgst_rate REAL DEFAULT 0,
			sgst_rate REAL DEFAULT 0,
			custom_tax_ids TEXT DEFAULT '[]',
			is_tax_exempt INTEGER DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create composite index for warehouse and variant lookup
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_inventory_batch_warehouse_variant ON inventory_batches(warehouse_id, variant_id)
	`).Error; err != nil {
		return err
	}

	// Collaborator table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS collaborators`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE collaborators (
			id TEXT PRIMARY KEY,
			external_id TEXT UNIQUE,
			company_name TEXT NOT NULL,
			logo TEXT,
			contact_person TEXT NOT NULL,
			contact_number TEXT NOT NULL,
			email TEXT,
			gst_number TEXT,
			pan_number TEXT,
			bank_account_no TEXT NOT NULL,
			bank_ifsc TEXT NOT NULL,
			bank_name TEXT,
			address_id TEXT,
			experience TEXT,
			is_active INTEGER,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create indexes for collaborators
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_collaborator_external_id ON collaborators(external_id)
	`).Error; err != nil {
		return err
	}

	// CollaboratorProduct table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS collaborator_products`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE collaborator_products (
			id TEXT PRIMARY KEY,
			collaborator_id TEXT NOT NULL,
			product_id TEXT NOT NULL,
			brand_name TEXT NOT NULL,
			hsn_code TEXT NOT NULL,
			gst_rate REAL NOT NULL,
			images TEXT,
			dosage_instructions TEXT,
			usage_details TEXT,
			is_active INTEGER DEFAULT 1,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create indexes for collaborator_products
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_collab_product ON collaborator_products(collaborator_id, product_id)
	`).Error; err != nil {
		return err
	}

	// GRN table (goods_receipt_notes)
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS goods_receipt_notes`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE goods_receipt_notes (
			id TEXT PRIMARY KEY,
			grn_number TEXT NOT NULL UNIQUE,
			grn_document TEXT,
			po_id TEXT NOT NULL,
			warehouse_id TEXT NOT NULL,
			received_date DATETIME NOT NULL,
			received_by TEXT NOT NULL,
			quality_status TEXT NOT NULL,
			remarks TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create indexes for goods_receipt_notes
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_grn_po ON goods_receipt_notes(po_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_grn_warehouse ON goods_receipt_notes(warehouse_id)
	`).Error; err != nil {
		return err
	}

	// GRNItem table (grn_items)
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS grn_items`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE grn_items (
			id TEXT PRIMARY KEY,
			grn_id TEXT NOT NULL,
			po_item_id TEXT NOT NULL,
			variant_id TEXT NOT NULL,
			ordered_quantity INTEGER NOT NULL,
			received_quantity INTEGER NOT NULL,
			accepted_quantity INTEGER NOT NULL,
			rejected_quantity INTEGER DEFAULT 0,
			expiry_date DATE NOT NULL,
			batch_number TEXT,
			inventory_batch_id TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create indexes for grn_items
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_grn_item_grn ON grn_items(grn_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_grn_item_po_item ON grn_items(po_item_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_grn_item_variant ON grn_items(variant_id)
	`).Error; err != nil {
		return err
	}

	// Discount table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS discounts`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE discounts (
			id TEXT PRIMARY KEY,
			code TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			discount_type TEXT NOT NULL,
			value REAL NOT NULL,
			max_discount_amount REAL,
			min_order_value REAL,
			max_order_value REAL,
			applicable_products TEXT,
			excluded_products TEXT,
			applicable_categories TEXT,
			excluded_categories TEXT,
			applicable_warehouses TEXT,
			usage_limit INTEGER,
			current_usage INTEGER DEFAULT 0,
			valid_from DATETIME NOT NULL,
			valid_until DATETIME NOT NULL,
			is_active INTEGER DEFAULT 1,
			is_stackable INTEGER DEFAULT 0,
			priority INTEGER DEFAULT 0,
			terms TEXT,
			buy_quantity INTEGER,
			get_quantity INTEGER,
			get_discount_type TEXT,
			get_discount_value REAL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create index for code
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_discount_code ON discounts(code)
	`).Error; err != nil {
		return err
	}

	// Tax table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS taxes`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE taxes (
			id TEXT PRIMARY KEY,
			code TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			tax_type TEXT NOT NULL,
			calculation_type TEXT NOT NULL,
			rate REAL NOT NULL,
			min_amount REAL,
			max_amount REAL,
			min_order_value REAL,
			max_order_value REAL,
			applicable_products TEXT,
			excluded_products TEXT,
			applicable_categories TEXT,
			excluded_categories TEXT,
			applicable_warehouses TEXT,
			excluded_warehouses TEXT,
			applicable_states TEXT,
			excluded_states TEXT,
			applicable_customer_groups TEXT,
			excluded_customer_groups TEXT,
			valid_from DATETIME NOT NULL,
			valid_until DATETIME,
			is_active INTEGER DEFAULT 1,
			priority INTEGER DEFAULT 0,
			is_stackable INTEGER DEFAULT 1,
			stacking_order INTEGER DEFAULT 0,
			requires_gstin INTEGER DEFAULT 0,
			requires_pan INTEGER DEFAULT 0,
			is_inter_state INTEGER DEFAULT 0,
			hsn_code TEXT,
			sac_code TEXT,
			tax_category TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create index for code
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tax_code ON taxes(code)
	`).Error; err != nil {
		return err
	}

	// WebhookEvent table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS webhook_events`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE webhook_events (
			id TEXT PRIMARY KEY,
			event_id TEXT NOT NULL UNIQUE,
			event_type TEXT NOT NULL,
			payload_hash TEXT NOT NULL,
			request_body TEXT NOT NULL,
			status TEXT NOT NULL,
			error_message TEXT,
			processed_at DATETIME,
			external_order_id TEXT,
			purchase_order_id TEXT,
			source_ip TEXT,
			user_agent TEXT,
			signature_valid INTEGER DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create indexes for webhook_events
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_webhook_event_id ON webhook_events(event_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_webhook_event_type ON webhook_events(event_type)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_webhook_status ON webhook_events(status)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_webhook_external_order ON webhook_events(external_order_id)
	`).Error; err != nil {
		return err
	}

	// Returns table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS returns`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE returns (
			id TEXT PRIMARY KEY,
			sale_id TEXT NOT NULL,
			return_date DATETIME NOT NULL,
			total_refund REAL NOT NULL,
			status TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create indexes for returns
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_return_sale_id ON returns(sale_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_return_status ON returns(status)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_return_date ON returns(return_date)
	`).Error; err != nil {
		return err
	}

	// ReturnItem table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS return_items`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE return_items (
			id TEXT PRIMARY KEY,
			return_id TEXT NOT NULL,
			batch_id TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			refund_amount REAL NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create indexes for return_items
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_return_item_return_id ON return_items(return_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_return_item_batch_id ON return_items(batch_id)
	`).Error; err != nil {
		return err
	}

	// BankPayment table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS bank_payments`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE bank_payments (
			id TEXT PRIMARY KEY,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			payment_method TEXT NOT NULL,
			amount REAL NOT NULL,
			transaction_reference TEXT,
			payment_date DATETIME NOT NULL,
			notes TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create indexes for bank_payments
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_bank_payment_entity ON bank_payments(entity_type, entity_id)
	`).Error; err != nil {
		return err
	}

	// RefundPolicy table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS refund_policies`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE refund_policies (
			id TEXT PRIMARY KEY,
			policy_name TEXT NOT NULL,
			description TEXT,
			refund_window_days INTEGER NOT NULL,
			refund_percentage REAL NOT NULL,
			conditions TEXT,
			is_active INTEGER DEFAULT 1,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// SaleSummary table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS sale_summaries`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE sale_summaries (
			id TEXT PRIMARY KEY,
			summary_date DATE NOT NULL UNIQUE,
			total_sales INTEGER NOT NULL,
			total_amount REAL NOT NULL,
			total_discount REAL DEFAULT 0,
			total_tax REAL DEFAULT 0,
			net_amount REAL NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create index for summary_date
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_sale_summary_date ON sale_summaries(summary_date)
	`).Error; err != nil {
		return err
	}

	// ReturnSummary table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS return_summaries`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE return_summaries (
			id TEXT PRIMARY KEY,
			summary_date DATE NOT NULL UNIQUE,
			total_returns INTEGER NOT NULL,
			total_refund REAL NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create index for summary_date
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_return_summary_date ON return_summaries(summary_date)
	`).Error; err != nil {
		return err
	}

	// TaxTier table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS tax_tiers`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE tax_tiers (
			id TEXT PRIMARY KEY,
			tax_id TEXT NOT NULL,
			min_amount REAL NOT NULL,
			max_amount REAL,
			rate REAL NOT NULL,
			tier_order INTEGER NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create indexes for tax_tiers
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tax_tier_tax_id ON tax_tiers(tax_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tax_tier_order ON tax_tiers(tax_id, tier_order)
	`).Error; err != nil {
		return err
	}

	// TaxApplication table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS tax_applications`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE tax_applications (
			id TEXT PRIMARY KEY,
			tax_id TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			tax_amount REAL NOT NULL,
			taxable_amount REAL NOT NULL,
			applied_at DATETIME NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create indexes for tax_applications
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tax_application_entity ON tax_applications(entity_type, entity_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tax_application_tax_id ON tax_applications(tax_id)
	`).Error; err != nil {
		return err
	}

	// TaxSummary table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS tax_summaries`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE tax_summaries (
			id TEXT PRIMARY KEY,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			total_tax REAL NOT NULL,
			cgst_amount REAL DEFAULT 0,
			sgst_amount REAL DEFAULT 0,
			igst_amount REAL DEFAULT 0,
			custom_tax_amount REAL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create index for tax_summaries
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tax_summary_entity ON tax_summaries(entity_type, entity_id)
	`).Error; err != nil {
		return err
	}

	// WebhookDeliveryAttempt table
	// Drop table first to ensure schema is correct
	if err := db.Exec(`DROP TABLE IF EXISTS webhook_delivery_attempts`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE webhook_delivery_attempts (
			id TEXT PRIMARY KEY,
			webhook_event_id TEXT NOT NULL,
			attempt_number INTEGER NOT NULL,
			response_code INTEGER,
			error_message TEXT,
			attempted_at DATETIME NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		)
	`).Error; err != nil {
		return err
	}

	// Create indexes for webhook_delivery_attempts
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_webhook_attempt_event ON webhook_delivery_attempts(webhook_event_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_webhook_attempt_number ON webhook_delivery_attempts(webhook_event_id, attempt_number)
	`).Error; err != nil {
		return err
	}

	return nil
}
