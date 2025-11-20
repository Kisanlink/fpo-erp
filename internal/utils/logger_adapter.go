package utils

import (
	"kisanlink-erp/internal/interfaces"

	"go.uber.org/zap"
)

// LoggerAdapter adapts *zap.Logger to interfaces.Logger
type LoggerAdapter struct {
	logger *zap.Logger
}

// NewLoggerAdapter creates a new LoggerAdapter
func NewLoggerAdapter(logger *zap.Logger) interfaces.Logger {
	return &LoggerAdapter{
		logger: logger,
	}
}

// Debug logs a debug message
func (l *LoggerAdapter) Debug(msg string, fields ...interface{}) {
	zapFields := l.convertToZapFields(fields...)
	l.logger.Debug(msg, zapFields...)
}

// Info logs an info message
func (l *LoggerAdapter) Info(msg string, fields ...interface{}) {
	zapFields := l.convertToZapFields(fields...)
	l.logger.Info(msg, zapFields...)
}

// Warn logs a warning message
func (l *LoggerAdapter) Warn(msg string, fields ...interface{}) {
	zapFields := l.convertToZapFields(fields...)
	l.logger.Warn(msg, zapFields...)
}

// Error logs an error message
func (l *LoggerAdapter) Error(msg string, fields ...interface{}) {
	zapFields := l.convertToZapFields(fields...)
	l.logger.Error(msg, zapFields...)
}

// Fatal logs a fatal message
func (l *LoggerAdapter) Fatal(msg string, fields ...interface{}) {
	zapFields := l.convertToZapFields(fields...)
	l.logger.Fatal(msg, zapFields...)
}

// convertToZapFields converts interface{} fields to zap.Field
func (l *LoggerAdapter) convertToZapFields(fields ...interface{}) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if ok {
				zapFields = append(zapFields, zap.Any(key, fields[i+1]))
			}
		}
	}
	return zapFields
}

// With returns a new logger with the given fields
func (l *LoggerAdapter) With(fields ...zap.Field) interfaces.Logger {
	return &LoggerAdapter{
		logger: l.logger.With(fields...),
	}
}

// Named returns a new logger with the given name
func (l *LoggerAdapter) Named(name string) interfaces.Logger {
	return &LoggerAdapter{
		logger: l.logger.Named(name),
	}
}

// Sync flushes any buffered log entries
func (l *LoggerAdapter) Sync() error {
	return l.logger.Sync()
}

// GetZapLogger returns the underlying zap logger
func (l *LoggerAdapter) GetZapLogger() *zap.Logger {
	return l.logger
}
