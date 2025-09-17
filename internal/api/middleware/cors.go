package middleware

import (
	"kisanlink-erp/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware provides CORS handling
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	// Default CORS configuration
	corsConfig := cors.DefaultConfig()

	// Allow all origins in development, specific origins in production
	if cfg.Server.Mode == "debug" {
		corsConfig.AllowAllOrigins = true
	} else {
		// Use configured allowed origins
		allowedOrigins := cfg.CORS.GetAllowedOrigins()
		if len(allowedOrigins) > 0 {
			corsConfig.AllowOrigins = allowedOrigins
		} else {
			// Fallback to default origins if none configured
			corsConfig.AllowOrigins = []string{
				"https://yourdomain.com",
				"https://www.yourdomain.com",
			}
		}
	}

	// Allow credentials (cookies, authorization headers)
	corsConfig.AllowCredentials = true

	// Use configured allowed headers
	allowedHeaders := cfg.CORS.GetAllowedHeaders()
	if len(allowedHeaders) > 0 {
		corsConfig.AllowHeaders = allowedHeaders
	} else {
		// Fallback to default headers if none configured
		corsConfig.AllowHeaders = []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
		}
	}

	// Allow specific methods
	corsConfig.AllowMethods = []string{
		"GET",
		"POST",
		"PUT",
		"PATCH",
		"DELETE",
		"OPTIONS",
	}

	// Expose headers to client
	corsConfig.ExposeHeaders = []string{
		"Content-Length",
		"X-Request-ID",
	}

	return cors.New(corsConfig)
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Relaxed CSP for docs route to allow Scalar documentation
		if c.Request.URL.Path == "/docs" {
			c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline'")
		} else {
			c.Header("Content-Security-Policy", "default-src 'self'")
		}

		c.Next()
	}
}

// RateLimitMiddleware provides basic rate limiting
func RateLimitMiddleware() gin.HandlerFunc {
	// Simple in-memory rate limiter
	// In production, you would use Redis or a proper rate limiting library
	requestCounts := make(map[string]int)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Simple rate limiting: max 100 requests per minute per IP
		if requestCounts[clientIP] > 100 {
			c.JSON(429, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		requestCounts[clientIP]++

		// Reset counter after 1 minute (in a real implementation)
		// This is a simplified version

		c.Next()
	}
}
