package database

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// RunPreMigrationFixes handles legacy schema issues before GORM auto-migration
// This function ensures that old schema artifacts (like the 'sku' table) are properly
// migrated to the current schema ('products' table) before GORM AutoMigrate runs.
func RunPreMigrationFixes(db *gorm.DB) error {
	log.Println("Running pre-migration fixes...")

	if err := fixProductTableName(db); err != nil {
		return fmt.Errorf("failed to fix product table name: %w", err)
	}

	if err := fixAttachmentFileTypeColumn(db); err != nil {
		return fmt.Errorf("failed to fix attachment file_type column: %w", err)
	}

	if err := renameFarmerIDToCustomerID(db); err != nil {
		return fmt.Errorf("failed to rename farmer_id to customer_id: %w", err)
	}

	if err := fixSubcategoryUniqueIndex(db); err != nil {
		return fmt.Errorf("failed to fix subcategory unique index: %w", err)
	}

	if err := migrateProductCategoryColumns(db); err != nil {
		return fmt.Errorf("failed to migrate product category columns: %w", err)
	}

	log.Println("Pre-migration fixes completed successfully")
	return nil
}

// fixAttachmentFileTypeColumn increases the file_type column size to accommodate long MIME types
// Example: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet (71 chars)
func fixAttachmentFileTypeColumn(db *gorm.DB) error {
	if !db.Migrator().HasTable("attachments") {
		log.Println("Attachments table does not exist yet - will be created by AutoMigrate")
		return nil
	}

	log.Println("Checking attachments.file_type column size...")

	// Check current column size
	var columnType string
	query := `
		SELECT character_maximum_length::text
		FROM information_schema.columns
		WHERE table_name = 'attachments' AND column_name = 'file_type'
	`
	if err := db.Raw(query).Scan(&columnType).Error; err != nil {
		log.Printf("Could not check column size: %v - skipping", err)
		return nil
	}

	if columnType == "150" {
		log.Println("attachments.file_type already has correct size (150)")
		return nil
	}

	log.Printf("Altering attachments.file_type from varchar(%s) to varchar(150)...", columnType)
	if err := db.Exec("ALTER TABLE attachments ALTER COLUMN file_type TYPE varchar(150)").Error; err != nil {
		return fmt.Errorf("failed to alter file_type column: %w", err)
	}

	log.Println("Successfully updated attachments.file_type column size")
	return nil
}

// renameFarmerIDToCustomerID renames the farmer_id column to customer_id in the sales table
// This is a more generic naming convention for customer identifier
func renameFarmerIDToCustomerID(db *gorm.DB) error {
	if !db.Migrator().HasTable("sales") {
		log.Println("Sales table does not exist yet - will be created by AutoMigrate")
		return nil
	}

	// Check if farmer_id column exists
	var columnExists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'sales' AND column_name = 'farmer_id'
		)
	`
	if err := db.Raw(query).Scan(&columnExists).Error; err != nil {
		log.Printf("Could not check for farmer_id column: %v - skipping", err)
		return nil
	}

	if !columnExists {
		log.Println("sales.farmer_id column does not exist - already renamed or never existed")
		return nil
	}

	log.Println("Renaming sales.farmer_id to sales.customer_id...")
	if err := db.Exec("ALTER TABLE sales RENAME COLUMN farmer_id TO customer_id").Error; err != nil {
		return fmt.Errorf("failed to rename farmer_id column: %w", err)
	}

	log.Println("Successfully renamed sales.farmer_id to sales.customer_id")
	return nil
}

// fixProductTableName handles the sku → products table rename
// This addresses a legacy issue where Product.TableName() was "sku" instead of "products"
//
// Handles 4 scenarios:
//  1. Only 'sku' exists - rename to 'products' and recreate FK
//  2. Both exist - merge data from sku → products, drop sku, recreate FK
//  3. Only 'products' exists - ideal state, no action needed
//  4. Neither exists - will be created by AutoMigrate
func fixProductTableName(db *gorm.DB) error {
	// Check if 'sku' table exists
	skuExists := db.Migrator().HasTable("sku")
	productsExists := db.Migrator().HasTable("products")

	log.Printf("Table check: sku=%v, products=%v", skuExists, productsExists)

	// Scenario 1: Only sku exists - rename it
	if skuExists && !productsExists {
		log.Println("Renaming 'sku' table to 'products'...")
		if err := db.Exec("ALTER TABLE sku RENAME TO products").Error; err != nil {
			return fmt.Errorf("failed to rename sku table: %w", err)
		}
		log.Println("Successfully renamed 'sku' to 'products'")
		return recreateProductVariantFK(db)
	}

	// Scenario 2: Both exist - merge and drop sku
	if skuExists && productsExists {
		log.Println("Both 'sku' and 'products' exist - merging data...")

		// Drop FK constraints first
		if err := dropProductVariantFKs(db); err != nil {
			return err
		}

		// Merge data (avoiding duplicates)
		mergeSQL := `
			INSERT INTO products (id, external_id, name, description, created_at, updated_at, deleted_at)
			SELECT s.id, s.external_id, s.name, s.description, s.created_at, s.updated_at, s.deleted_at
			FROM sku s
			WHERE NOT EXISTS (
				SELECT 1 FROM products p WHERE p.id = s.id
			)
		`
		if err := db.Exec(mergeSQL).Error; err != nil {
			return fmt.Errorf("failed to merge sku data: %w", err)
		}

		// Drop sku table
		if err := db.Exec("DROP TABLE sku CASCADE").Error; err != nil {
			return fmt.Errorf("failed to drop sku table: %w", err)
		}

		log.Println("Successfully merged and removed 'sku' table")
		return recreateProductVariantFK(db)
	}

	// Scenario 3: Only products exists - ideal state
	if !skuExists && productsExists {
		log.Println("'products' table already exists - no rename needed")
		return nil
	}

	// Scenario 4: Neither exists - let GORM create it
	log.Println("Neither 'sku' nor 'products' exists - will be created by AutoMigrate")
	return nil
}

// dropProductVariantFKs drops all foreign key constraints on product_variants.product_id
// This is necessary before renaming/merging the product table to avoid constraint violations
func dropProductVariantFKs(db *gorm.DB) error {
	log.Println("Dropping product_variants foreign key constraints...")

	// Query PostgreSQL information_schema to find FK constraints
	var constraints []string
	query := `
		SELECT tc.constraint_name
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
		  ON tc.constraint_name = kcu.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
		  AND tc.table_name = 'product_variants'
		  AND kcu.column_name = 'product_id'
	`

	if err := db.Raw(query).Scan(&constraints).Error; err != nil {
		return fmt.Errorf("failed to query FK constraints: %w", err)
	}

	// Drop each constraint
	for _, constraint := range constraints {
		dropSQL := fmt.Sprintf("ALTER TABLE product_variants DROP CONSTRAINT IF EXISTS %s CASCADE", constraint)
		if err := db.Exec(dropSQL).Error; err != nil {
			return fmt.Errorf("failed to drop constraint %s: %w", constraint, err)
		}
		log.Printf("Dropped constraint: %s", constraint)
	}

	return nil
}

// recreateProductVariantFK recreates the foreign key constraint
// This ensures the FK points to the correct 'products' table after rename/merge
func recreateProductVariantFK(db *gorm.DB) error {
	log.Println("Recreating product_variants foreign key constraint...")

	// Drop existing constraint first (idempotent)
	db.Exec("ALTER TABLE product_variants DROP CONSTRAINT IF EXISTS fk_products_variants")

	// Recreate with correct reference
	createFK := `
		ALTER TABLE product_variants
		ADD CONSTRAINT fk_products_variants
		FOREIGN KEY (product_id)
		REFERENCES products(id)
		ON DELETE CASCADE
	`

	if err := db.Exec(createFK).Error; err != nil {
		return fmt.Errorf("failed to create FK constraint: %w", err)
	}

	log.Println("Successfully recreated foreign key constraint")
	return nil
}

// fixSubcategoryUniqueIndex fixes the incorrect unique index on subcategories.name
// The original index was on 'name' only, but subcategory names can be duplicated across categories
// (e.g., "Others" can exist in both "Bio Products" and "Irrigation" categories).
// This drops the incorrect single-column unique index so GORM can create the correct composite index.
func fixSubcategoryUniqueIndex(db *gorm.DB) error {
	if !db.Migrator().HasTable("subcategories") {
		log.Println("Subcategories table does not exist yet - will be created by AutoMigrate")
		return nil
	}

	log.Println("Checking subcategories unique index...")

	// Check if the incorrect single-column unique index exists
	var indexExists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes
			WHERE tablename = 'subcategories'
			AND indexname = 'idx_subcategories_name'
			AND schemaname = CURRENT_SCHEMA()
		)
	`
	if err := db.Raw(query).Scan(&indexExists).Error; err != nil {
		log.Printf("Could not check for subcategories name index: %v - skipping", err)
		return nil
	}

	if !indexExists {
		log.Println("idx_subcategories_name index does not exist - nothing to fix")
		return nil
	}

	// Check if it's a single-column index (incorrect) or multi-column (correct)
	// A single-column index on 'name' alone is wrong because subcategory names
	// are only unique WITHIN a category, not globally
	var columnCount int
	countQuery := `
		SELECT count(*)
		FROM pg_index i
		JOIN pg_class c ON c.oid = i.indexrelid
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE c.relname = 'idx_subcategories_name'
	`
	if err := db.Raw(countQuery).Scan(&columnCount).Error; err != nil {
		log.Printf("Could not check index column count: %v - skipping", err)
		return nil
	}

	if columnCount > 1 {
		log.Println("idx_subcategories_name is already a composite index - nothing to fix")
		return nil
	}

	// Drop the incorrect single-column unique index
	log.Println("Dropping incorrect single-column unique index idx_subcategories_name...")
	if err := db.Exec("DROP INDEX IF EXISTS idx_subcategories_name").Error; err != nil {
		return fmt.Errorf("failed to drop incorrect index: %w", err)
	}

	log.Println("Successfully dropped incorrect idx_subcategories_name index")
	log.Println("GORM will create the correct composite index during migration")
	return nil
}

// migrateProductCategoryColumns migrates products table from category_name/subcategory_name
// (string-based) to category_id/subcategory_id (ID-based foreign keys)
// This is idempotent - checks if old columns exist before migrating
func migrateProductCategoryColumns(db *gorm.DB) error {
	if !db.Migrator().HasTable("products") {
		log.Println("Products table does not exist yet - will be created by AutoMigrate")
		return nil
	}

	// Check if old category_name column exists
	var hasOldColumn bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'products' AND column_name = 'category_name'
		)
	`
	if err := db.Raw(query).Scan(&hasOldColumn).Error; err != nil {
		log.Printf("Could not check for category_name column: %v - skipping", err)
		return nil
	}

	if !hasOldColumn {
		log.Println("products.category_name column does not exist - already migrated to ID-based")
		return nil
	}

	log.Println("Migrating products from category_name to category_id...")

	// Step 1: Add new columns if they don't exist
	log.Println("Adding category_id and subcategory_id columns...")
	db.Exec("ALTER TABLE products ADD COLUMN IF NOT EXISTS category_id VARCHAR(50)")
	db.Exec("ALTER TABLE products ADD COLUMN IF NOT EXISTS subcategory_id VARCHAR(50)")

	// Step 2: Migrate data from name to ID (only if categories table exists)
	if db.Migrator().HasTable("categories") {
		log.Println("Migrating category_name to category_id...")
		migrateCategory := `
			UPDATE products p
			SET category_id = c.id
			FROM categories c
			WHERE p.category_name = c.name AND p.category_id IS NULL
		`
		if err := db.Exec(migrateCategory).Error; err != nil {
			log.Printf("Warning: Could not migrate category_name to category_id: %v", err)
		}
	}

	if db.Migrator().HasTable("subcategories") {
		log.Println("Migrating subcategory_name to subcategory_id...")
		migrateSubcategory := `
			UPDATE products p
			SET subcategory_id = s.id
			FROM subcategories s
			WHERE p.subcategory_name = s.name
			AND p.category_name = s.category_name
			AND p.subcategory_id IS NULL
		`
		if err := db.Exec(migrateSubcategory).Error; err != nil {
			log.Printf("Warning: Could not migrate subcategory_name to subcategory_id: %v", err)
		}
	}

	// Step 3: Drop old FK constraints
	log.Println("Dropping old FK constraints...")
	db.Exec("ALTER TABLE products DROP CONSTRAINT IF EXISTS fk_products_category")
	db.Exec("ALTER TABLE products DROP CONSTRAINT IF EXISTS fk_products_subcategory")

	// Step 4: Drop old columns
	log.Println("Dropping old category_name and subcategory_name columns...")
	if err := db.Exec("ALTER TABLE products DROP COLUMN IF EXISTS category_name").Error; err != nil {
		log.Printf("Warning: Could not drop category_name column: %v", err)
	}
	if err := db.Exec("ALTER TABLE products DROP COLUMN IF EXISTS subcategory_name").Error; err != nil {
		log.Printf("Warning: Could not drop subcategory_name column: %v", err)
	}

	log.Println("Successfully migrated products to ID-based category references")
	return nil
}

// RunPostMigrationDataMigrations runs data migrations AFTER GORM AutoMigrate
// These migrations copy data from old columns to new columns and clean up old columns
// This must be called AFTER AutoMigrate because new columns need to exist first
func RunPostMigrationDataMigrations(db *gorm.DB) error {
	log.Println("Running post-migration data migrations...")

	if err := migrateCustomerIDToPhone(db); err != nil {
		return fmt.Errorf("failed to migrate customer_id to customer_phone: %w", err)
	}

	log.Println("Post-migration data migrations completed successfully")
	return nil
}

// migrateCustomerIDToPhone migrates existing customer_id values to customer_phone
// customer_id may contain either AAA user IDs or phone numbers
// We migrate phone-like values (10-15 digits) to customer_phone, then drop customer_id
// This is idempotent - checks if old column exists before migrating
func migrateCustomerIDToPhone(db *gorm.DB) error {
	if !db.Migrator().HasTable("sales") {
		log.Println("Sales table does not exist yet - skipping customer_id migration")
		return nil
	}

	// Check if customer_id column exists (old column to migrate from)
	var customerIDExists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'sales' AND column_name = 'customer_id'
		)
	`
	if err := db.Raw(query).Scan(&customerIDExists).Error; err != nil {
		log.Printf("Could not check for customer_id column: %v - skipping", err)
		return nil
	}

	if !customerIDExists {
		log.Println("sales.customer_id column does not exist - already migrated")
		return nil
	}

	// Check if customer_phone column exists (new column to migrate to)
	var customerPhoneExists bool
	query2 := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'sales' AND column_name = 'customer_phone'
		)
	`
	if err := db.Raw(query2).Scan(&customerPhoneExists).Error; err != nil {
		log.Printf("Could not check for customer_phone column: %v - skipping", err)
		return nil
	}

	if !customerPhoneExists {
		log.Println("sales.customer_phone column does not exist yet - AutoMigrate may not have run")
		return nil
	}

	log.Println("Migrating sales.customer_id to sales.customer_phone...")

	// Step 1: Copy phone-like customer_id values to customer_phone
	// Phone-like: 10-15 digits (Indian and international phone numbers)
	// Uses PostgreSQL regex: ^ matches start, $ matches end, [0-9]{10,15} matches 10-15 digits
	migrateSQL := `
		UPDATE sales
		SET customer_phone = customer_id
		WHERE customer_id ~ '^[0-9]{10,15}$'
		AND customer_phone IS NULL
	`
	result := db.Exec(migrateSQL)
	if result.Error != nil {
		log.Printf("Warning: Could not migrate customer_id to customer_phone: %v", result.Error)
		// Don't fail - continue to drop the column
	} else {
		log.Printf("Migrated %d phone numbers from customer_id to customer_phone", result.RowsAffected)
	}

	// Step 2: Drop the old customer_id column
	log.Println("Dropping old sales.customer_id column...")
	if err := db.Exec("ALTER TABLE sales DROP COLUMN IF EXISTS customer_id").Error; err != nil {
		log.Printf("Warning: Could not drop customer_id column: %v", err)
		// Don't fail - column will be orphaned but system will work
	}

	log.Println("Successfully migrated customer_id to customer_phone")
	return nil
}
