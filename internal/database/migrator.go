package database

import (
	"kisanlink-erp/internal/database/models"
	"log"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	log.Println("Starting database auto-migration...")

	// List all models for auto-migration
	models := []interface{}{
		// Core entities
		&models.Warehouse{},
		&models.Product{},
		&models.ProductPrice{},

		// Procurement entities
		&models.Collaborator{},
		&models.CollaboratorProduct{},
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
	return nil
}
