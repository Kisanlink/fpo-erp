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

var (
	originalLogWriter io.Writer
	originalLogFlags  int
	logSuppressed     bool
)

func init() {
	// Save original log writer and flags
	originalLogWriter = log.Writer()
	originalLogFlags = log.Flags()

	// Suppress log output immediately at package init
	// This prevents any log.Printf calls from outputting during tests
	suppressLogOutput()
}

// suppressLogOutput suppresses standard log package output
func suppressLogOutput() {
	if !logSuppressed {
		log.SetOutput(io.Discard)
		log.SetFlags(0) // Remove all flags including timestamps
		logSuppressed = true
	}
}

// restoreLogOutput restores standard log package output
func restoreLogOutput() {
	if logSuppressed {
		log.SetOutput(originalLogWriter)
		log.SetFlags(originalLogFlags)
		logSuppressed = false
	}
}

// SetupTestDB creates an in-memory SQLite database for testing
// This matches the e-commerce pattern for simple, fast test setup
func SetupTestDB(t *testing.T) *gorm.DB {
	// Ensure log output is suppressed (already done in init, but ensure it's set)
	suppressLogOutput()
	// Create a custom logger using log.New with io.Discard and no flags
	// This completely suppresses all output including formatting and timestamps
	// Using flags=0 prevents any formatting (timestamps, prefixes) that could leave blank lines
	customLogger := logger.New(
		log.New(io.Discard, "", 0), // Use io.Discard with no flags to suppress all output
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

	// Create SQLite-compatible tables for models with PostgreSQL-specific types
	if err := CreateSQLiteCompatibleTables(db); err != nil {
		t.Logf("WARNING: Failed to create SQLite-compatible tables: %v", err)
		// Continue with AutoMigrate as fallback
	}

	// Auto-migrate models one by one to identify which fails
	// Skip models already created manually: ProductPrice, InventoryTransaction, Sale, SaleItem, DiscountUsage, Attachment
	modelsToMigrate := []interface{}{
		&models.Warehouse{},
		&models.Product{},
		&models.ProductVariant{},
		// Skip ProductPrice - already created manually
		&models.InventoryBatch{},
		// Skip InventoryTransaction - already created manually
		&models.Collaborator{},
		&models.CollaboratorProduct{},
		&models.PurchaseOrder{},
		&models.PurchaseOrderItem{},
		&models.GRN{},
		&models.GRNItem{},
		// Skip Sale - already created manually
		// Skip SaleItem - already created manually
		&models.Discount{},
		// Skip DiscountUsage - already created manually
		&models.Tax{},
		&models.WebhookEvent{},
		// Skip Attachment - already created manually
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

// CleanupTestDB closes the database connection and restores log output
func CleanupTestDB(db *gorm.DB) {
	// Restore original log writer and flags
	restoreLogOutput()

	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}
