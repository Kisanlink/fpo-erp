package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"kisanlink-erp/internal/errors"

	"github.com/gin-gonic/gin"
)

// WebhookSecurityService handles webhook security validation
type WebhookSecurityService struct {
	secret string
}

// NewWebhookSecurityService creates a new webhook security service
func NewWebhookSecurityService(secret string) *WebhookSecurityService {
	return &WebhookSecurityService{
		secret: secret,
	}
}

// WebhookHeaders contains extracted webhook headers
type WebhookHeaders struct {
	Signature string
	EventID   string
	Timestamp string
}

// VerifyHMACSignature verifies the HMAC-SHA256 signature of the webhook payload
func (s *WebhookSecurityService) VerifyHMACSignature(payload []byte, signature string) bool {
	// Remove "sha256=" prefix if present
	signature = strings.TrimPrefix(signature, "sha256=")

	// Compute HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(s.secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures (constant time comparison)
	return hmac.Equal([]byte(expectedMAC), []byte(signature))
}

// ValidateTimestamp prevents replay attacks by ensuring webhook is not too old
// Rejects webhooks older than 5 minutes
func (s *WebhookSecurityService) ValidateTimestamp(timestamp string) error {
	// Parse timestamp as Unix timestamp (seconds)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return errors.NewValidationError(fmt.Sprintf("Invalid timestamp format: %v", err))
	}

	// Convert to time.Time
	webhookTime := time.Unix(ts, 0)
	now := time.Now()

	// Check if webhook is too old (> 5 minutes)
	maxAge := 5 * time.Minute
	age := now.Sub(webhookTime)

	if age > maxAge {
		return errors.NewBadRequestError(fmt.Sprintf("Webhook timestamp too old: %v (max age: %v)", age, maxAge))
	}

	// Check if webhook is from the future (clock skew tolerance: 1 minute)
	if webhookTime.After(now.Add(1 * time.Minute)) {
		return errors.NewBadRequestError("Webhook timestamp is in the future")
	}

	return nil
}

// ExtractHeaders extracts and validates required webhook headers from Gin context
func (s *WebhookSecurityService) ExtractHeaders(c *gin.Context) (*WebhookHeaders, error) {
	signature := c.GetHeader("X-Webhook-Signature")
	if signature == "" {
		return nil, errors.NewBadRequestError("Missing X-Webhook-Signature header")
	}

	eventID := c.GetHeader("X-Event-ID")
	if eventID == "" {
		return nil, errors.NewBadRequestError("Missing X-Event-ID header")
	}

	timestamp := c.GetHeader("X-Timestamp")
	if timestamp == "" {
		return nil, errors.NewBadRequestError("Missing X-Timestamp header")
	}

	return &WebhookHeaders{
		Signature: signature,
		EventID:   eventID,
		Timestamp: timestamp,
	}, nil
}

// ValidateWebhook performs complete webhook validation (headers + signature + timestamp)
func (s *WebhookSecurityService) ValidateWebhook(c *gin.Context, payload []byte) (*WebhookHeaders, error) {
	// Extract headers
	headers, err := s.ExtractHeaders(c)
	if err != nil {
		return nil, errors.NewBadRequestError(fmt.Sprintf("Header validation failed: %v", err))
	}

	// Validate timestamp
	if err := s.ValidateTimestamp(headers.Timestamp); err != nil {
		return nil, errors.NewBadRequestError(fmt.Sprintf("Timestamp validation failed: %v", err))
	}

	// Verify HMAC signature
	if !s.VerifyHMACSignature(payload, headers.Signature) {
		return nil, errors.NewUnauthorizedError("HMAC signature verification failed")
	}

	return headers, nil
}

// GetSourceIP extracts the source IP from the request
func (s *WebhookSecurityService) GetSourceIP(c *gin.Context) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return c.ClientIP()
}

// GetUserAgent extracts the User-Agent from the request
func (s *WebhookSecurityService) GetUserAgent(c *gin.Context) string {
	return c.GetHeader("User-Agent")
}
