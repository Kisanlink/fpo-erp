package testutils

import (
	"io"
	"log"
	"testing"

	"kisanlink-erp/internal/database/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates an in-memory SQLite database for testing
// This matches the e-commerce pattern for simple, fast test setup
func SetupTestDB(t *testing.T) *gorm.DB {
	// Create a custom logger that discards all output (including warnings)
	customLogger := logger.New(
		log.New(io.Discard, "", log.LstdFlags), // Use standard log.New with io.Discard
		logger.Config{
			SlowThreshold:             0,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: customLogger,
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Auto-migrate models one by one to identify which fails
	modelsToMigrate := []interface{}{
		&models.Warehouse{},
		&models.Product{},
		&models.ProductVariant{},
		&models.ProductPrice{},
		&models.InventoryBatch{},
		&models.InventoryTransaction{},
		&models.Collaborator{},
		&models.CollaboratorProduct{},
		&models.PurchaseOrder{},
		&models.PurchaseOrderItem{},
		&models.GRN{},
		&models.GRNItem{},
		&models.Sale{},
		&models.SaleItem{},
		&models.Discount{},
		&models.DiscountUsage{},
		&models.Tax{},
		&models.WebhookEvent{},
		&models.Attachment{},
	}

	// Suppress GORM warnings by disabling colored output and using a silent session
	// Create a completely silent database session for migration only
	silentDB := db.Session(&gorm.Session{
		Logger:                   customLogger,
		SkipDefaultTransaction:   true,
		DisableNestedTransaction: true,
		AllowGlobalUpdate:        false,
		QueryFields:              false,
		CreateBatchSize:          0,
		PrepareStmt:              false,
	})

	for _, model := range modelsToMigrate {
		// Migrate using the silent session - errors will still be logged to test output
		if err := silentDB.AutoMigrate(model); err != nil {
			t.Logf("WARNING: Failed to migrate %T: %v", model, err)
		}
	}

	return db
}

// CleanupTestDB closes the database connection
func CleanupTestDB(db *gorm.DB) {
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}
