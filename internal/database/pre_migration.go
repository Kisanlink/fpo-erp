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
