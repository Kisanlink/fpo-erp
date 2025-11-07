package testutils

import (
	"gorm.io/gorm"
)

// CreateSQLiteCompatibleTables creates tables with SQLite-compatible types
// This function manually creates tables for models that use PostgreSQL-specific
// types that SQLite doesn't support (timestamptz, numeric(precision,scale), etc.)
func CreateSQLiteCompatibleTables(db *gorm.DB) error {
	// ProductPrice table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS product_prices (
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
		return err
	}

	// Create index for variant_id
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_variant_price ON product_prices(variant_id)
	`).Error; err != nil {
		return err
	}

	// Sale table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sales (
			id TEXT PRIMARY KEY,
			warehouse_id TEXT NOT NULL,
			sale_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			total_amount REAL NOT NULL,
			status TEXT NOT NULL,
			payment_method TEXT,
			sale_type TEXT,
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
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sale_items (
			id TEXT PRIMARY KEY,
			sale_id TEXT NOT NULL,
			variant_id TEXT NOT NULL,
			batch_id TEXT,
			quantity INTEGER NOT NULL,
			selling_price REAL NOT NULL,
			line_total REAL NOT NULL,
			discount_amount REAL DEFAULT 0,
			tax_amount REAL DEFAULT 0,
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
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS inventory_transactions (
			id TEXT PRIMARY KEY,
			batch_id TEXT NOT NULL,
			transaction_type TEXT NOT NULL,
			quantity_change INTEGER NOT NULL,
			previous_quantity INTEGER,
			new_quantity INTEGER,
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
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS discount_usages (
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
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS attachments (
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

	// InventoryBatch table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS inventory_batches (
			id TEXT PRIMARY KEY,
			warehouse_id TEXT NOT NULL,
			variant_id TEXT NOT NULL,
			cost_price REAL NOT NULL,
			expiry_date TEXT NOT NULL,
			total_quantity INTEGER NOT NULL,
			cgst_rate REAL DEFAULT 0,
			sgst_rate REAL DEFAULT 0,
			custom_tax_ids TEXT,
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

	return nil
}
