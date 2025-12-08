package database

import (
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/services"
	"log"
	"strings"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	log.Println("Starting database auto-migration...")

	// Phase 1: Run pre-migration fixes for legacy schema issues
	if err := RunPreMigrationFixes(db); err != nil {
		log.Printf("Failed to run pre-migration fixes: %v", err)
		return err
	}

	// Phase 2: Run standard GORM auto-migration
	log.Println("Running GORM auto-migration...")

	// Phase 2a: Migrate category tables first (required for FK constraints)
	log.Println("Migrating category tables...")
	categoryModels := []interface{}{
		&models.Category{},
		&models.Subcategory{},
	}
	for _, model := range categoryModels {
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("Failed to migrate model %T: %v", model, err)
			return err
		}
		log.Printf("Successfully migrated model %T", model)
	}

	// Phase 2b: Seed default categories before Product migration
	// This ensures FK constraints can be satisfied for existing products
	if err := seedDefaultCategories(db); err != nil {
		log.Printf("Failed to seed default categories: %v", err)
		return err
	}

	// List remaining models for auto-migration (categories already migrated above)
	remainingModels := []interface{}{
		// Core entities
		&models.Warehouse{},
		&models.Product{},

		// Procurement entities
		&models.Collaborator{},
		// Note: CollaboratorProduct removed - deprecated in favor of ProductVariant with collaborator_id
		&models.ProductVariant{},
		&models.PurchaseOrder{},
		&models.PurchaseOrderItem{},
		&models.GRN{},
		&models.GRNItem{},

		// Inventory entities
		&models.InventoryBatch{},
		&models.InventoryTransaction{},

		// Sales entities
		&models.Sale{},
		&models.SaleItem{},
		&models.SaleSummary{},
		&models.SaleCancellation{},
		&models.SaleCancellationItem{},

		// Returns entities
		&models.Return{},
		&models.ReturnItem{},
		&models.ReturnSummary{},

		// Supporting entities
		&models.RefundPolicy{},
		&models.BankPayment{},
		&models.Attachment{},
		&models.Discount{},
		&models.DiscountUsage{},
		&models.Tax{},
		&models.TaxTier{},
		&models.TaxApplication{},
		&models.TaxSummary{},

		// Webhook integration
		&models.WebhookEvent{},
		&models.WebhookDeliveryAttempt{},
	}

	// Perform auto-migration for remaining models
	for _, model := range remainingModels {
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("Failed to migrate model %T: %v", model, err)
			return err
		}
		log.Printf("Successfully migrated model %T", model)
	}

	log.Println("Database auto-migration completed successfully")

	// Phase 3: Run post-migration data migrations
	// These run AFTER AutoMigrate because new columns need to exist first
	if err := RunPostMigrationDataMigrations(db); err != nil {
		log.Printf("Failed to run post-migration data migrations: %v", err)
		return err
	}

	// Phase 4: Create application performance indexes
	// Indexes are optimization - don't fail startup if some fail
	if err := CreateApplicationIndexes(db); err != nil {
		log.Printf("⚠️  Warning: Index creation had issues: %v", err)
		// Continue - indexes are performance optimization, not critical
	}

	return nil
}

// seedDefaultCategories seeds the predefined categories and subcategories.
// This is called during migration BEFORE Product migration to ensure FK constraints
// can be satisfied for existing products.
// This function is idempotent - checks if category exists (case-insensitive) before creating.
// Uses services.PredefinedCategories as single source of truth.
func seedDefaultCategories(db *gorm.DB) error {
	log.Println("Seeding default categories...")

	categoryCount := 0
	subcategoryCount := 0

	for _, cat := range services.PredefinedCategories {
		// Check if category already exists (case-insensitive)
		var existingCategory models.Category
		result := db.Where("LOWER(name) = LOWER(?)", cat.Name).First(&existingCategory)

		var categoryID string
		if result.Error != nil && !strings.Contains(result.Error.Error(), "record not found") {
			// Real database error
			log.Printf("Failed to check category %s: %v", cat.Name, result.Error)
			return result.Error
		}

		if result.RowsAffected == 0 {
			// Category doesn't exist, create it using proper constructor
			description := cat.Description
			newCategory := models.NewCategory(cat.Name, &description)
			if err := db.Create(newCategory).Error; err != nil {
				log.Printf("Failed to create category %s: %v", cat.Name, err)
				return err
			}
			categoryID = newCategory.ID
			categoryCount++
		} else {
			// Category exists, use its ID for subcategories
			categoryID = existingCategory.ID
		}

		// Create subcategories for this category
		for _, sub := range cat.Subcategories {
			// Check if subcategory already exists (case-insensitive name + category_id)
			var existingSub models.Subcategory
			result := db.Where("LOWER(name) = LOWER(?) AND category_id = ?", sub.Name, categoryID).First(&existingSub)

			if result.Error != nil && !strings.Contains(result.Error.Error(), "record not found") {
				log.Printf("Failed to check subcategory %s: %v", sub.Name, result.Error)
				return result.Error
			}

			if result.RowsAffected == 0 {
				// Subcategory doesn't exist, create it using proper constructor
				description := sub.Description
				newSubcategory := models.NewSubcategory(sub.Name, categoryID, &description)
				if err := db.Create(newSubcategory).Error; err != nil {
					log.Printf("Failed to create subcategory %s: %v", sub.Name, err)
					return err
				}
				subcategoryCount++
			}
		}
	}

	log.Printf("✅ Seeded %d categories and %d subcategories", categoryCount, subcategoryCount)
	return nil
}
