package middleware

import (
	"time"

	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware provides request/response logging
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log request details
		utils.Info("HTTP Request",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"client_ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
		)

		return ""
	})
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID (you could use UUID here)
		requestID := time.Now().Format("20060102150405") + "-" + c.ClientIP()

		// Set request ID in context
		c.Set("request_id", requestID)

		// Add request ID to response headers
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// ErrorLoggingMiddleware logs errors that occur during request processing
func ErrorLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Check if there were any errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				utils.Error("Request Error",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"status", c.Writer.Status(),
					"error", err.Error(),
					"client_ip", c.ClientIP(),
				)
			}
		}
	}
}

// PerformanceMiddleware logs slow requests
func PerformanceMiddleware(threshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log if request took longer than threshold
		if duration > threshold {
			utils.Warn("Slow Request",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"duration", duration,
				"threshold", threshold,
				"client_ip", c.ClientIP(),
			)
		}
	}
}



