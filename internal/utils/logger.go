package utils

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var zapLogger *zap.Logger

// Init initializes the logger (deprecated - kept for backward compatibility)
func Init() {
	// This function is deprecated but kept for backward compatibility
	// The actual logger initialization happens in main.go
	// This function now does nothing as we use the logger from main.go
}

// SetGlobalLogger sets the global zap logger instance
func SetGlobalLogger(logger *zap.Logger) {
	zapLogger = logger
}

// GetZapLogger returns the zap logger instance
func GetZapLogger() *zap.Logger {
	if zapLogger == nil {
		// Fallback: create a default logger if not set
		var err error
		zapLogger, err = zap.NewProduction()
		if err != nil {
			panic("Failed to create fallback logger: " + err.Error())
		}
	}
	return zapLogger
}

// ConfigureGormLogger configures GORM logger based on LOG_LEVEL environment variable
func ConfigureGormLogger(db *gorm.DB) {
	// Get log level from environment
	logLevel := "info" // default
	if os.Getenv("LOG_LEVEL") != "" {
		logLevel = os.Getenv("LOG_LEVEL")
	}

	// Map to GORM logger levels - NOW SHOWING SQL QUERIES
	var gormLogLevel logger.LogLevel
	switch logLevel {
	case "error", "fatal", "panic":
		gormLogLevel = logger.Error
	case "warn":
		gormLogLevel = logger.Warn
	case "debug":
		gormLogLevel = logger.Info // Show all SQL queries in debug mode
	case "info":
		gormLogLevel = logger.Warn // Show warnings and errors in info mode
	default:
		gormLogLevel = logger.Info // Default: show SQL queries
	}

	db.Logger = db.Logger.LogMode(gormLogLevel)
}

// formatMessage concatenates all arguments into a single message string
func formatMessage(args ...interface{}) string {
	if len(args) == 0 {
		return ""
	}

	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = fmt.Sprint(arg)
	}
	return strings.Join(parts, " ")
}

// Info logs info level message by concatenating all arguments
func Info(args ...interface{}) {
	logger := GetZapLogger()
	msg := formatMessage(args...)
	logger.Info(msg)
}

// Error logs error level message by concatenating all arguments
func Error(args ...interface{}) {
	logger := GetZapLogger()
	msg := formatMessage(args...)
	logger.Error(msg)
}

// Debug logs debug level message by concatenating all arguments
func Debug(args ...interface{}) {
	logger := GetZapLogger()
	msg := formatMessage(args...)
	logger.Debug(msg)
}

// Warn logs warning level message by concatenating all arguments
func Warn(args ...interface{}) {
	logger := GetZapLogger()
	msg := formatMessage(args...)
	logger.Warn(msg)
}

// Fatal logs a fatal message and exits by concatenating all arguments
func Fatal(args ...interface{}) {
	logger := GetZapLogger()
	msg := formatMessage(args...)
	logger.Fatal(msg)
}

// initZapLogger initializes a production zap logger with the specified log level
func initZapLogger(logLevel string) (*zap.Logger, error) {
	// Parse log level
	var level zapcore.Level
	switch logLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Create production config
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)

	return config.Build()
}
