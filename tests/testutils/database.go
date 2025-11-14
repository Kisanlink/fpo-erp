package testutils

import (
	"io"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	originalLogWriter io.Writer
	originalLogFlags  int
	logSuppressed     bool
)

// SQLiteJSONNamingStrategy wraps the default naming strategy to handle JSON serialization
type SQLiteJSONNamingStrategy struct {
	schema.NamingStrategy
}

// ColumnName returns the column name for a field
// This is used to map fields with _date suffix in JSON to match production queries
func (s SQLiteJSONNamingStrategy) ColumnName(table, column string) string {
	// First get the default snake_case column name
	defaultName := s.NamingStrategy.ColumnName(table, column)

	// Map ExpectedDelivery field to expected_delivery_date column to match production queries
	// Check both the original column name and the converted name
	if column == "ExpectedDelivery" || defaultName == "expected_delivery" {
		return "expected_delivery_date"
	}

	// Map ActualDelivery field to actual_delivery_date column to match production queries
	if column == "ActualDelivery" || defaultName == "actual_delivery" {
		return "actual_delivery_date"
	}

	return defaultName
}

// ColumnDataType returns the data type for SQLite columns
func (s SQLiteJSONNamingStrategy) ColumnDataType(field *schema.Field) string {
	// For []string fields with type:json tag, use TEXT
	// Also handle serialization automatically
	if field.Tag.Get("gorm") == "type:json" ||
		field.Tag.Get("serializer") == "json" ||
		(field.FieldType.Kind() == reflect.Slice && field.FieldType.Elem().Kind() == reflect.String) {
		// Register serializer if not already registered
		schema.RegisterSerializer("json", JSONStringSliceSerializer{})
		return "TEXT"
	}
	return ""
}

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

	// Register SQLite JSON serializer for []string fields BEFORE opening database
	// This is needed because SQLite doesn't have native JSON type like PostgreSQL
	// Registering globally before DB connection prevents schema re-parsing issues
	RegisterSQLiteJSONSerializer(nil)

	// Create silent logger
	silentLogger := logger.New(
		log.New(io.Discard, "", 0),
		logger.Config{
			SlowThreshold:             0,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Use :memory: database with shared cache mode for transaction visibility
	// This allows transactions to see data committed outside the transaction
	// Enable WAL mode for better concurrent access and read_uncommitted for transaction visibility
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_journal_mode=WAL&_read_uncommitted=true"), &gorm.Config{
		Logger:                                   silentLogger,
		NamingStrategy:                           SQLiteJSONNamingStrategy{},
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Execute PRAGMA to enable read uncommitted mode
	if err := db.Exec("PRAGMA read_uncommitted = ON").Error; err != nil {
		t.Fatalf("failed to set read_uncommitted pragma: %v", err)
	}

	// Create SQLite-compatible tables for all models
	// We use manual table creation instead of AutoMigrate to ensure correct schema
	// including all fields like deleted_by, and to avoid GORM overwriting our tables
	if err := CreateSQLiteCompatibleTables(db); err != nil {
		t.Fatalf("Failed to create SQLite-compatible tables: %v\nThis error occurred during table creation. Check sqlite_migrations.go for SQL syntax errors.", err)
	}

	// Register callbacks to handle JSON serialization for []string fields in SQLite
	RegisterJSONCallbacks(db)

	return db
}

// RegisterJSONCallbacks patches GORM schemas to automatically apply JSON serializer
// for []string fields with type:json tag
func RegisterJSONCallbacks(db *gorm.DB) {
	// Hook into schema initialization to patch field tags
	db.Callback().Create().Before("gorm:before_create").Register("sqlite_json_patch_schema", func(db *gorm.DB) {
		patchSchemaForJSON(db)
	})

	db.Callback().Query().Before("gorm:query").Register("sqlite_json_patch_schema", func(db *gorm.DB) {
		patchSchemaForJSON(db)
	})

	db.Callback().Update().Before("gorm:before_update").Register("sqlite_json_patch_schema", func(db *gorm.DB) {
		patchSchemaForJSON(db)
	})
}

// patchSchemaForJSON modifies schema fields to use JSON serializer for []string fields with type:json
func patchSchemaForJSON(db *gorm.DB) {
	if db.Statement.Schema == nil {
		return
	}

	for _, field := range db.Statement.Schema.Fields {
		// Check if field is []string with type:json tag
		if field.FieldType.Kind() == reflect.Slice &&
			field.FieldType.Elem().Kind() == reflect.String &&
			strings.Contains(field.Tag.Get("gorm"), "type:json") {

			// Set the serializer for this field
			field.Serializer = JSONStringSliceSerializer{}
		}
	}
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
