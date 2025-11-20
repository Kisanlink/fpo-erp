package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"kisanlink-erp/internal/interfaces"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Logger logs request details with structured logging
func Logger(logger interfaces.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("client_ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.String("request_id", getRequestID(param.Keys)),
		)
		return ""
	})
}

// getRequestID extracts request ID from gin context keys
func getRequestID(keys map[string]interface{}) string {
	if requestID, exists := keys["request_id"]; exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return "unknown"
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based generation
		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		b := make([]byte, length)
		for i := range b {
			b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		}
		return string(b)
	}
	return hex.EncodeToString(bytes)
}
