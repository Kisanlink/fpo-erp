package database

import (
	"kisanlink-erp/internal/database/models"
	"log"

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

	// List all models for auto-migration
	models := []interface{}{
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

	// Perform auto-migration
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("Failed to migrate model %T: %v", model, err)
			return err
		}
		log.Printf("Successfully migrated model %T", model)
	}

	log.Println("Database auto-migration completed successfully")

	// Phase 3: Create application performance indexes
	// Indexes are optimization - don't fail startup if some fail
	if err := CreateApplicationIndexes(db); err != nil {
		log.Printf("⚠️  Warning: Index creation had issues: %v", err)
		// Continue - indexes are performance optimization, not critical
	}

	return nil
}
