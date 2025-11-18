package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
)

// WebhookHistoryService handles webhook event tracking for idempotency and audit
type WebhookHistoryService struct {
	repo *repositories.WebhookRepository
}

// NewWebhookHistoryService creates a new webhook history service
func NewWebhookHistoryService(repo *repositories.WebhookRepository) *WebhookHistoryService {
	return &WebhookHistoryService{
		repo: repo,
	}
}

// CheckIdempotency checks if a webhook event has already been processed
// Returns (alreadyProcessed bool, existingEvent *WebhookEvent, error)
func (s *WebhookHistoryService) CheckIdempotency(ctx context.Context, eventID string) (bool, *models.WebhookEvent, error) {
	event, err := s.repo.FindByEventID(ctx, eventID)
	if err != nil {
		return false, nil, errors.NewInternalServerError(fmt.Sprintf("Failed to check idempotency: %v", err))
	}

	if event == nil {
		// Event not found - first time processing
		return false, nil, nil
	}

	// Event already exists
	if event.Status == "success" || event.Status == "processing" {
		// Already processed or currently processing - idempotent response
		return true, event, nil
	}

	// Event failed previously - allow retry
	return false, event, nil
}

// RecordWebhook creates a new webhook event record in "processing" status
func (s *WebhookHistoryService) RecordWebhook(ctx context.Context, event *models.WebhookEvent) error {
	// Set default status if not set
	if event.Status == "" {
		event.Status = "processing"
	}

	return s.repo.Create(ctx, event)
}

// MarkProcessed updates a webhook event to "success" status
func (s *WebhookHistoryService) MarkProcessed(ctx context.Context, eventID, purchaseOrderID string) error {
	event, err := s.repo.FindByEventID(ctx, eventID)
	if err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("Failed to find webhook event: %v", err))
	}

	if event == nil {
		return errors.NewNotFoundError(fmt.Sprintf("Webhook event not found: %s", eventID))
	}

	now := time.Now()
	event.Status = "success"
	event.ProcessedAt = &now
	event.ErrorMessage = nil
	if purchaseOrderID != "" {
		event.PurchaseOrderID = &purchaseOrderID
	}

	return s.repo.Update(ctx, event)
}

// MarkFailed updates a webhook event to "failed" status with error message
func (s *WebhookHistoryService) MarkFailed(ctx context.Context, eventID string, err error) error {
	event, errFind := s.repo.FindByEventID(ctx, eventID)
	if errFind != nil {
		return errors.NewInternalServerError(fmt.Sprintf("Failed to find webhook event: %v", errFind))
	}

	if event == nil {
		return errors.NewNotFoundError(fmt.Sprintf("Webhook event not found: %s", eventID))
	}

	now := time.Now()
	errorMsg := err.Error()
	event.Status = "failed"
	event.ProcessedAt = &now
	event.ErrorMessage = &errorMsg

	return s.repo.Update(ctx, event)
}

// CreateWebhookEvent creates a new WebhookEvent from webhook data
func (s *WebhookHistoryService) CreateWebhookEvent(
	eventID, eventType string,
	requestBody []byte,
	externalOrderID *string,
	sourceIP, userAgent string,
	signatureValid bool,
) *models.WebhookEvent {
	// Compute payload hash (SHA256)
	hash := sha256.Sum256(requestBody)
	payloadHash := hex.EncodeToString(hash[:])

	event := models.NewWebhookEvent(eventID, eventType, payloadHash, string(requestBody))
	event.ExternalOrderID = externalOrderID
	event.SourceIP = &sourceIP
	event.UserAgent = &userAgent
	event.SignatureValid = signatureValid

	return event
}

// RecordDeliveryAttempt records a delivery attempt for a webhook event
func (s *WebhookHistoryService) RecordDeliveryAttempt(
	ctx context.Context,
	webhookEventID string,
	attemptNumber int,
	responseCode int,
	errorMessage *string,
) error {
	attempt := models.NewWebhookDeliveryAttempt(webhookEventID, attemptNumber)
	attempt.ResponseCode = responseCode
	attempt.ErrorMessage = errorMessage

	return s.repo.CreateDeliveryAttempt(ctx, attempt)
}

// GetWebhookEventsByOrderID retrieves all webhook events for a specific external order ID
func (s *WebhookHistoryService) GetWebhookEventsByOrderID(ctx context.Context, externalOrderID string) ([]models.WebhookEvent, error) {
	return s.repo.FindByExternalOrderID(ctx, externalOrderID)
}

// GetFailedWebhooks retrieves all failed webhook events (for retry processing)
func (s *WebhookHistoryService) GetFailedWebhooks(ctx context.Context, limit int) ([]models.WebhookEvent, error) {
	return s.repo.FindByStatus(ctx, "failed", limit)
}

// GetWebhookStats returns statistics for webhook events
func (s *WebhookHistoryService) GetWebhookStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	eventTypes := []string{
		"order.created",
		"order.confirmed",
		"order.shipped",
		"order.delivered",
		"order.payment",
	}

	for _, eventType := range eventTypes {
		total, err := s.repo.CountByEventType(ctx, eventType)
		if err != nil {
			return nil, errors.NewInternalServerError(fmt.Sprintf("Failed to count events for %s: %v", eventType, err))
		}

		success, err := s.repo.CountByStatusAndEventType(ctx, "success", eventType)
		if err != nil {
			return nil, errors.NewInternalServerError(fmt.Sprintf("Failed to count success for %s: %v", eventType, err))
		}

		failed, err := s.repo.CountByStatusAndEventType(ctx, "failed", eventType)
		if err != nil {
			return nil, errors.NewInternalServerError(fmt.Sprintf("Failed to count failed for %s: %v", eventType, err))
		}

		stats[eventType] = map[string]interface{}{
			"total":   total,
			"success": success,
			"failed":  failed,
		}
	}

	return stats, nil
}
