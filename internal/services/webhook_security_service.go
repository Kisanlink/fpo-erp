package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"kisanlink-erp/internal/errors"
)

const (
	// Maximum timestamp difference allowed (5 minutes as per documentation)
	MaxTimestampDifference = 5 * time.Minute
)

// WebhookSecurityService handles webhook security operations
type WebhookSecurityService struct{}

// NewWebhookSecurityService creates a new webhook security service
func NewWebhookSecurityService() *WebhookSecurityService {
	return &WebhookSecurityService{}
}

// ValidateTimestamp validates that the timestamp is within the allowed window
func (s *WebhookSecurityService) ValidateTimestamp(timestamp int64) error {
	requestTime := time.Unix(timestamp, 0)
	currentTime := time.Now()

	// Check if timestamp is too old or too far in the future
	timeDiff := currentTime.Sub(requestTime)
	if timeDiff < -MaxTimestampDifference || timeDiff > MaxTimestampDifference {
		return errors.NewBadRequestError(fmt.Sprintf("Timestamp is outside allowed window. Current time: %d, Request time: %d", currentTime.Unix(), timestamp))
	}

	return nil
}

// GenerateHMACSignature generates SHA-256 HMAC signature for payload
func (s *WebhookSecurityService) GenerateHMACSignature(payload string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	signature := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("sha256=%s", signature)
}

// ValidateHMACSignature validates the HMAC signature against the payload
func (s *WebhookSecurityService) ValidateHMACSignature(payload string, signature string, secret string) error {
	expectedSignature := s.GenerateHMACSignature(payload, secret)

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return errors.NewUnauthorizedError("HMAC signature validation failed")
	}

	return nil
}

// ValidateWebhookHeaders validates all required webhook headers
func (s *WebhookSecurityService) ValidateWebhookHeaders(signature, timestampHeader string) (int64, error) {
	// Validate signature header presence
	if signature == "" {
		return 0, errors.NewBadRequestError("X-Kisanlink-Signature header is required")
	}

	// Validate timestamp header presence
	if timestampHeader == "" {
		return 0, errors.NewBadRequestError("X-Kisanlink-Timestamp header is required")
	}

	// Parse timestamp
	timestamp, err := strconv.ParseInt(timestampHeader, 10, 64)
	if err != nil {
		return 0, errors.NewBadRequestError("Invalid timestamp format")
	}

	// Validate timestamp
	if err := s.ValidateTimestamp(timestamp); err != nil {
		return 0, err
	}

	return timestamp, nil
}

// SecureWebhookValidation performs complete webhook security validation
func (s *WebhookSecurityService) SecureWebhookValidation(payload, signature, timestampHeader, secret string) error {
	// Validate headers and get parsed timestamp
	_, err := s.ValidateWebhookHeaders(signature, timestampHeader)
	if err != nil {
		return err
	}

	// Validate HMAC signature
	if err := s.ValidateHMACSignature(payload, signature, secret); err != nil {
		return err
	}

	return nil
}

// WebhookSecurityContext contains validated webhook security data
type WebhookSecurityContext struct {
	Timestamp         int64
	SignatureValid    bool
	TimestampValid    bool
	ValidationErrors  []error
}

// ValidateWebhookSecurity performs comprehensive webhook security validation
func (s *WebhookSecurityService) ValidateWebhookSecurity(payload, signature, timestampHeader, secret string) *WebhookSecurityContext {
	context := &WebhookSecurityContext{
		ValidationErrors: make([]error, 0),
	}

	// Parse and validate timestamp
	if timestampHeader != "" {
		if timestamp, err := strconv.ParseInt(timestampHeader, 10, 64); err != nil {
			context.ValidationErrors = append(context.ValidationErrors, errors.NewBadRequestError("Invalid timestamp format"))
		} else {
			context.Timestamp = timestamp
			if err := s.ValidateTimestamp(timestamp); err != nil {
				context.ValidationErrors = append(context.ValidationErrors, err)
			} else {
				context.TimestampValid = true
			}
		}
	} else {
		context.ValidationErrors = append(context.ValidationErrors, errors.NewBadRequestError("X-Kisanlink-Timestamp header is required"))
	}

	// Validate HMAC signature
	if signature != "" {
		if err := s.ValidateHMACSignature(payload, signature, secret); err != nil {
			context.ValidationErrors = append(context.ValidationErrors, err)
		} else {
			context.SignatureValid = true
		}
	} else {
		context.ValidationErrors = append(context.ValidationErrors, errors.NewBadRequestError("X-Kisanlink-Signature header is required"))
	}

	return context
}

// IsValid returns true if all validations passed
func (ctx *WebhookSecurityContext) IsValid() bool {
	return ctx.SignatureValid && ctx.TimestampValid && len(ctx.ValidationErrors) == 0
}

// GetFirstError returns the first validation error, if any
func (ctx *WebhookSecurityContext) GetFirstError() error {
	if len(ctx.ValidationErrors) > 0 {
		return ctx.ValidationErrors[0]
	}
	return nil
}

// GetAllErrors returns all validation errors
func (ctx *WebhookSecurityContext) GetAllErrors() []error {
	return ctx.ValidationErrors
}