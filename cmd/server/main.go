// @title Kisanlink ERP API
// @version 1.0
// @description Comprehensive ERP system for agricultural cooperatives with multi-tenant architecture
// @termsOfService http://swagger.io/terms/
// @contact.name Kisanlink Support
// @contact.url https://github.com/Kisanlink/fpo-erp
// @contact.email info@kisanlink.in
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:3000
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "kisanlink-erp/docs" // Import docs for Swagger
	"kisanlink-erp/internal/aaa"
	api_server "kisanlink-erp/internal/api/server"
	"kisanlink-erp/internal/config"
	"kisanlink-erp/internal/constants"
	"kisanlink-erp/internal/database"
	"kisanlink-erp/internal/utils"

	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	kdb "github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	// Initialize structured logger
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if err := zapLogger.Sync(); err != nil {
			log.Printf("Failed to sync logger: %v", err)
		}
	}()

	// Create logger adapter
	logger := utils.NewLoggerAdapter(zapLogger)

	// Set global logger for utils functions
	utils.SetGlobalLogger(zapLogger)

	// Load configuration from environment variables
	cfg := config.Load()

	// Initialize database with auto-migration
	ctx := context.Background()

	manager := kdb.NewDatabaseManager()
	if err := manager.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect using Kisanlink-DB: %v", err)
	}
	pg, err := manager.GetPostgresManager().GetDB(ctx, false)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Configure GORM logger
	utils.ConfigureGormLogger(pg)

	// Run auto-migration (if enabled)
	if cfg.Database.AutoMigrate {
		if err := database.AutoMigrate(pg); err != nil {
			log.Fatalf("Failed to run auto-migration: %v", err)
		}
	} else {
		log.Println("⚠️  WARNING: Database auto-migration is DISABLED")
		log.Println("   Ensure database schema matches application models manually")
	}

	// Initialize hash counters from database
	initializeHashCounters(pg)

	// Seed AAA roles and permissions for ERP module (non-fatal)
	if cfg.AAA.Enabled && cfg.AAA.GRPCAddress != "" {
		log.Println("Attempting to seed AAA roles and permissions for ERP module...")

		catalogClient, err := aaa.NewCatalogGRPCClient(cfg.AAA.GRPCAddress, cfg.AAA.APIKey, cfg.AAA.UseTLS)
		if err != nil {
			log.Printf("Warning: AAA catalog client initialization failed: %v", err)
		} else {
			defer func() {
				if err := catalogClient.Close(); err != nil {
					log.Printf("Warning: failed to close AAA catalog client: %v", err)
				}
			}()

			seedCtx, seedCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer seedCancel()

			// Pass "erp-module" as service_id to use ERP-specific seed provider
			if err := catalogClient.SeedRolesAndPermissions(seedCtx, "erp-module", false); err != nil {
				log.Printf("Warning: AAA role/permission seeding failed: %v", err)
				log.Println("ERP will continue to start, but ensure AAA has required roles.")
			} else {
				log.Println("AAA ERP roles and permissions seeded successfully.")
			}
		}
	} else {
		log.Println("Skipping AAA role/permission seeding (AAA disabled or gRPC address missing).")
	}

	// Initialize HTTP server with AAA middleware
	httpServer := api_server.NewServer(pg, cfg, logger)

	// Start HTTP server
	go func() {
		log.Printf("Starting HTTP server on port %s", cfg.Server.HTTPPort)
		if err := http.ListenAndServe(":"+cfg.Server.HTTPPort, httpServer.Router); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Close server resources gracefully
	if httpServer != nil {
		if err := httpServer.Close(); err != nil {
			log.Printf("Error closing server resources: %v", err)
		}
	}

	log.Println("Server stopped")
}

// initializeHashCounters initializes hash counters for all models from database
func initializeHashCounters(db *gorm.DB) {
	utils.Info("Initializing hash counters from database...")

	// Define all table identifiers and their sizes using centralized constants
	tableConfigs := map[string]hash.TableSize{
		constants.TableProduct:       hash.Medium, // Products
		constants.TableWarehouse:     hash.Medium, // Warehouses
		constants.TableSale:          hash.Medium, // Sales
		constants.TableSaleItem:      hash.Medium, // Sale Items
		constants.TableSaleSummary:   hash.Medium, // Sale Summaries
		constants.TableBatch:         hash.Medium, // Inventory Batches
		constants.TableTransaction:   hash.Medium, // Inventory Transactions
		constants.TablePrice:         hash.Medium, // Product Prices
		constants.TableDiscount:      hash.Medium, // Discounts
		constants.TableDiscountUse:   hash.Medium, // Discount Usage
		constants.TableTax:           hash.Medium, // Tax
		constants.TableTaxTier:       hash.Medium, // Tax Tiers
		constants.TableTaxApp:        hash.Medium, // Tax Applications
		constants.TableTaxSummary:    hash.Medium, // Tax Summaries
		constants.TableReturn:        hash.Medium, // Returns
		constants.TableReturnItem:    hash.Medium, // Return Items
		constants.TableReturnSummary: hash.Medium, // Return Summaries
		constants.TableRefundPolicy:  hash.Medium, // Refund Policies
		constants.TableBankPayment:   hash.Medium, // Bank Payments
		constants.TableAttachment:    hash.Medium, // Attachments
		// Procurement Module
		constants.TableCollaborator:        hash.Medium, // Collaborators/Vendors
		constants.TableCollaboratorProduct: hash.Medium, // Collaborator-Product Junction
		constants.TableProductVariant:      hash.Medium, // Product Variants
		constants.TablePurchaseOrder:       hash.Medium, // Purchase Orders
		constants.TablePurchaseOrderItem:   hash.Medium, // PO Items
		constants.TableGRN:                 hash.Medium, // Goods Receipt Notes
		constants.TableGRNItem:             hash.Medium, // GRN Items
		// Webhook Integration
		constants.TableWebhookEvent:           hash.Medium, // Webhook Events
		constants.TableWebhookDeliveryAttempt: hash.Medium, // Webhook Delivery Attempts
	}

	// Initialize counters for each table
	for tableID, size := range tableConfigs {
		// Get existing IDs from database
		existingIDs := getExistingIDs(db, tableID)

		// Initialize global counter with existing IDs
		hash.InitializeGlobalCountersFromDatabase(tableID, existingIDs, size)

		utils.Info("Initialized counter for table:", tableID, "with", len(existingIDs), "existing records")
	}

	utils.Info("Hash counters initialization completed")
}

// getExistingIDs retrieves existing IDs for a table identifier from the database
func getExistingIDs(db *gorm.DB, tableID string) []string {
	// Map table identifiers to actual database table names
	tableNameMap := map[string]string{
		constants.TableProduct:       "products",
		constants.TableWarehouse:     "warehouses",
		constants.TableSale:          "sales",
		constants.TableSaleItem:      "sale_items",
		constants.TableSaleSummary:   "sale_summaries",
		constants.TableBatch:         "inventory_batches",
		constants.TableTransaction:   "inventory_transactions",
		constants.TablePrice:         "product_prices",
		constants.TableDiscount:      "discounts",
		constants.TableDiscountUse:   "discount_usages",
		constants.TableTax:           "taxes",
		constants.TableTaxTier:       "tax_tiers",
		constants.TableTaxApp:        "tax_applications",
		constants.TableTaxSummary:    "tax_summaries",
		constants.TableReturn:        "returns",
		constants.TableReturnItem:    "return_items",
		constants.TableReturnSummary: "return_summaries",
		constants.TableRefundPolicy:  "refund_policies",
		constants.TableBankPayment:   "bank_payments",
		constants.TableAttachment:    "attachments",
		// Procurement Module
		constants.TableCollaborator:        "collaborators",
		constants.TableCollaboratorProduct: "collaborator_products",
		constants.TableProductVariant:      "product_variants",
		constants.TablePurchaseOrder:       "purchase_orders",
		constants.TablePurchaseOrderItem:   "purchase_order_items",
		constants.TableGRN:                 "goods_receipt_notes",
		constants.TableGRNItem:             "grn_items",
		// Webhook Integration
		constants.TableWebhookEvent:           "webhook_events",
		constants.TableWebhookDeliveryAttempt: "webhook_delivery_attempts",
	}

	tableName, exists := tableNameMap[tableID]
	if !exists {
		utils.Error("Unknown table identifier:", tableID)
		return []string{}
	}

	// Query database for existing IDs
	var ids []string
	query := "SELECT id FROM " + tableName + " WHERE id IS NOT NULL"

	if err := db.Raw(query).Scan(&ids).Error; err != nil {
		utils.Error("Failed to query existing IDs for table:", tableName, "error:", err)
		return []string{}
	}

	utils.Info("Found", len(ids), "existing records in table:", tableName)
	return ids
}
