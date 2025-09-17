package utils

import (
	"os"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var log *logrus.Logger

// Init initializes the logger
func Init() {
	log = logrus.New()

	// Set output to stdout
	log.SetOutput(os.Stdout)

	// Set log level based on environment
	if os.Getenv("LOG_LEVEL") != "" {
		level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
		if err == nil {
			log.SetLevel(level)
		}
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	// Set formatter
	log.SetFormatter(&logrus.JSONFormatter{})
}

// GetLogger returns the logger instance
func GetLogger() *logrus.Logger {
	if log == nil {
		Init()
	}
	return log
}

// ConfigureGormLogger configures GORM logger based on LOG_LEVEL environment variable
func ConfigureGormLogger(db *gorm.DB) {
	// Get the same log level as your application logger
	logLevel := logrus.InfoLevel // default
	if os.Getenv("LOG_LEVEL") != "" {
		if level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL")); err == nil {
			logLevel = level
		}
	}

	// Map logrus levels to GORM logger levels
	var gormLogLevel logger.LogLevel
	switch logLevel {
	case logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel:
		gormLogLevel = logger.Error
	case logrus.WarnLevel:
		gormLogLevel = logger.Silent // Change to Silent to suppress all warnings
	case logrus.InfoLevel, logrus.DebugLevel, logrus.TraceLevel:
		gormLogLevel = logger.Silent // Change to Silent to suppress all warnings
	default:
		gormLogLevel = logger.Silent // Change to Silent to suppress all warnings
	}

	db.Logger = db.Logger.LogMode(gormLogLevel)
}

// Info logs info level message
func Info(args ...interface{}) {
	if log == nil {
		Init()
	}
	log.Info(args...)
}

// Error logs error level message
func Error(args ...interface{}) {
	if log == nil {
		Init()
	}
	log.Error(args...)
}

// Debug logs debug level message
func Debug(args ...interface{}) {
	if log == nil {
		Init()
	}
	log.Debug(args...)
}

// Warn logs warning level message
func Warn(args ...interface{}) {
	if log == nil {
		Init()
	}
	log.Warn(args...)
}
