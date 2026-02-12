package middleware

import (
	"strings"

	"kisanlink-erp/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware provides CORS handling.
// Supports exact origins and wildcard subdomain patterns (e.g., "https://*.kisanlink.in").
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()

	corsConfig.AllowCredentials = true

	// Use configured allowed headers
	allowedHeaders := cfg.CORS.GetAllowedHeaders()
	if len(allowedHeaders) > 0 {
		corsConfig.AllowHeaders = allowedHeaders
	} else {
		corsConfig.AllowHeaders = []string{
			"Origin", "Content-Type", "Accept", "Authorization",
			"X-Requested-With", "X-Request-ID", "X-Organization-ID",
		}
	}

	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.ExposeHeaders = []string{"Content-Length", "X-Request-ID"}

	// Allow all origins in development
	if cfg.Server.Mode == "debug" {
		corsConfig.AllowAllOrigins = true
		return cors.New(corsConfig)
	}

	// Production: parse configured origins
	allowedOrigins := cfg.CORS.GetAllowedOrigins()
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"https://yourdomain.com", "https://www.yourdomain.com"}
	}

	// Check if any origin contains a wildcard pattern
	hasWildcard := false
	for _, o := range allowedOrigins {
		if strings.Contains(o, "*.") {
			hasWildcard = true
			break
		}
	}

	if !hasWildcard {
		corsConfig.AllowOrigins = allowedOrigins
		return cors.New(corsConfig)
	}

	// Separate exact origins from wildcard patterns
	exactOrigins := make(map[string]bool)
	var wildcardPatterns []string
	for _, o := range allowedOrigins {
		if strings.Contains(o, "*.") {
			wildcardPatterns = append(wildcardPatterns, o)
		} else {
			exactOrigins[o] = true
		}
	}

	corsConfig.AllowOriginFunc = func(origin string) bool {
		if exactOrigins[origin] {
			return true
		}
		for _, pattern := range wildcardPatterns {
			if isOriginAllowed(origin, pattern) {
				return true
			}
		}
		return false
	}

	return cors.New(corsConfig)
}

// isOriginAllowed checks if an origin matches a wildcard subdomain pattern.
// Pattern format: "https://*.kisanlink.in" matches "https://admin.kisanlink.in"
// but not "https://a.b.kisanlink.in" (single subdomain level only).
func isOriginAllowed(origin, pattern string) bool {
	wildcardIdx := strings.Index(pattern, "*.")
	if wildcardIdx < 0 {
		return origin == pattern
	}

	scheme := pattern[:wildcardIdx]   // e.g., "https://"
	suffix := pattern[wildcardIdx+1:] // e.g., ".kisanlink.in"

	if !strings.HasPrefix(origin, scheme) {
		return false
	}
	if !strings.HasSuffix(origin, suffix) {
		return false
	}

	subdomain := origin[len(scheme) : len(origin)-len(suffix)]
	return len(subdomain) > 0 && !strings.Contains(subdomain, ".") && !strings.Contains(subdomain, "/")
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
		if c.Request.URL.Path == "/docs" || c.Request.URL.Path == "/api-docs" {
			c.Header("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline' 'wasm-unsafe-eval' https://cdn.jsdelivr.net; "+
					"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; "+
					"font-src 'self' https://cdn.jsdelivr.net data:; "+
					"connect-src 'self' https://cdn.jsdelivr.net; "+
					"img-src 'self' data: https:; "+
					"worker-src 'self' blob:")
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
