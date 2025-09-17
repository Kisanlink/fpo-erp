package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter represents a simple rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// CreateRateLimitMiddleware creates a rate limiting middleware
func CreateRateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Check if client has exceeded rate limit
		if !limiter.Allow(clientIP) {
			TooManyRequestsResponse(c, "Rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

// Allow checks if a request is allowed for the given key
func (r *RateLimiter) Allow(key string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-r.window)

	// Get existing requests for this key
	requests, exists := r.requests[key]
	if !exists {
		requests = []time.Time{}
	}

	// Filter out old requests outside the window
	var validRequests []time.Time
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if we're under the limit
	if len(validRequests) < r.limit {
		// Add current request
		validRequests = append(validRequests, now)
		r.requests[key] = validRequests
		return true
	}

	// Update requests for this key (even if rejected)
	r.requests[key] = validRequests
	return false
}

// Cleanup removes old entries to prevent memory leaks
func (r *RateLimiter) Cleanup() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-r.window)

	for key, requests := range r.requests {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(windowStart) {
				validRequests = append(validRequests, reqTime)
			}
		}

		if len(validRequests) == 0 {
			delete(r.requests, key)
		} else {
			r.requests[key] = validRequests
		}
	}
}

// StartCleanup starts a background goroutine to clean up old entries
func (r *RateLimiter) StartCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			r.Cleanup()
		}
	}()
}

// TooManyRequestsResponse sends a 429 response
func TooManyRequestsResponse(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, gin.H{
		"success": false,
		"message": message,
		"error":   "Rate limit exceeded",
	})
}
