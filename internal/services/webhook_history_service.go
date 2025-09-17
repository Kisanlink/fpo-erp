package services

import (
	"encoding/json"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/utils"
)

// WebhookHistoryService handles webhook history operations and idempotency
type WebhookHistoryService struct {
	webhookRepo *repositories.WebhookRepository
}

// NewWebhookHistoryService creates a new webhook history service
func NewWebhookHistoryService(webhookRepo *repositories.WebhookRepository) *WebhookHistoryService {
	return &WebhookHistoryService{
		webhookRepo: webhookRepo,
	}
}

// CheckIdempotency checks if an event has already been processed
func (s *WebhookHistoryService) CheckIdempotency(eventID string) (*models.WebhookEvent, bool, error) {
	existingEvent, err := s.webhookRepo.GetEventByEventID(eventID)
	if err != nil {
		// If event not found, it's new (not an error in this context)
		// Check if it's a "not found" error by checking if it's an AppError with 404 code
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == 404 {
			return nil, false, nil
		}
		return nil, false, err
	}

	// Event exists - check if it was already processed
	isProcessed := existingEvent.ProcessedStatus == "completed"
	return existingEvent, isProcessed, nil
}

// CreateEventRecord creates a new webhook event record for idempotency tracking
func (s *WebhookHistoryService) CreateEventRecord(eventType, eventID, fpoID string, payload interface{}, source string) (*models.WebhookEvent, error) {
	// Convert payload to JSON string
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to serialize payload")
	}

	// Create new event
	event := models.NewWebhookEvent(eventType, eventID, fpoID, string(payloadBytes), source)

	if err := s.webhookRepo.CreateEvent(event); err != nil {
		return nil, err
	}

	return event, nil
}

// MarkEventProcessed marks an event as successfully processed
func (s *WebhookHistoryService) MarkEventProcessed(eventID string) error {
	event, err := s.webhookRepo.GetEventByEventID(eventID)
	if err != nil {
		return err
	}

	now := time.Now()
	event.ProcessedStatus = "completed"
	event.ProcessedAt = &now

	return s.webhookRepo.UpdateEvent(event)
}

// MarkEventFailed marks an event as failed with error message
func (s *WebhookHistoryService) MarkEventFailed(eventID string, errorMsg string) error {
	event, err := s.webhookRepo.GetEventByEventID(eventID)
	if err != nil {
		return err
	}

	now := time.Now()
	event.ProcessedStatus = "failed"
	event.ProcessedAt = &now
	event.ErrorMessage = &errorMsg

	return s.webhookRepo.UpdateEvent(event)
}

// LogWebhookDelivery logs outbound webhook delivery attempt
func (s *WebhookHistoryService) LogWebhookDelivery(configID, eventType, eventID string, payload interface{}) (*models.WebhookHistory, error) {
	// Convert payload to JSON string
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to serialize payload")
	}

	// Create history record
	history := models.NewWebhookHistory(configID, eventType, eventID, string(payloadBytes))

	if err := s.webhookRepo.CreateHistory(history); err != nil {
		return nil, err
	}

	return history, nil
}

// UpdateDeliverySuccess updates webhook delivery as successful
func (s *WebhookHistoryService) UpdateDeliverySuccess(historyID string, responseCode int, responseBody string) error {
	history, err := s.webhookRepo.GetHistoryByID(historyID)
	if err != nil {
		return err
	}

	now := time.Now()
	history.Status = "success"
	history.AttemptCount += 1
	history.LastAttemptAt = &now
	history.CompletedAt = &now
	history.ResponseCode = &responseCode
	history.ResponseBody = &responseBody

	return s.webhookRepo.UpdateHistory(history)
}

// UpdateDeliveryFailure updates webhook delivery as failed
func (s *WebhookHistoryService) UpdateDeliveryFailure(historyID string, responseCode *int, responseBody, errorMsg string) error {
	history, err := s.webhookRepo.GetHistoryByID(historyID)
	if err != nil {
		return err
	}

	now := time.Now()
	history.Status = "failed"
	history.AttemptCount += 1
	history.LastAttemptAt = &now
	history.ResponseCode = responseCode
	history.ResponseBody = &responseBody
	history.ErrorMessage = &errorMsg

	return s.webhookRepo.UpdateHistory(history)
}

// GetEventHistory retrieves history for a specific event
func (s *WebhookHistoryService) GetEventHistory(eventID string) ([]models.WebhookHistoryResponse, error) {
	history, err := s.webhookRepo.GetHistoryByEventID(eventID)
	if err != nil {
		return nil, err
	}

	responses := make([]models.WebhookHistoryResponse, len(history))
	for i, h := range history {
		responses[i] = models.WebhookHistoryResponse{
			ID:           h.ID,
			ConfigID:     h.ConfigID,
			EventType:    h.EventType,
			EventID:      h.EventID,
			Status:       h.Status,
			AttemptCount: h.AttemptCount,
			ResponseCode: h.ResponseCode,
			ErrorMessage: h.ErrorMessage,
			ScheduledAt:  h.ScheduledAt.Format(time.RFC3339),
			CreatedAt:    h.CreatedAt.Format(time.RFC3339),
		}

		if h.LastAttemptAt != nil {
			lastAttempt := h.LastAttemptAt.Format(time.RFC3339)
			responses[i].LastAttemptAt = &lastAttempt
		}

		if h.CompletedAt != nil {
			completed := h.CompletedAt.Format(time.RFC3339)
			responses[i].CompletedAt = &completed
		}
	}

	return responses, nil
}

// GetConfigHistory retrieves delivery history for a webhook configuration
func (s *WebhookHistoryService) GetConfigHistory(configID string, limit, offset int) ([]models.WebhookHistoryResponse, error) {
	history, err := s.webhookRepo.GetHistoryByConfig(configID, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]models.WebhookHistoryResponse, len(history))
	for i, h := range history {
		responses[i] = models.WebhookHistoryResponse{
			ID:           h.ID,
			ConfigID:     h.ConfigID,
			EventType:    h.EventType,
			EventID:      h.EventID,
			Status:       h.Status,
			AttemptCount: h.AttemptCount,
			ResponseCode: h.ResponseCode,
			ErrorMessage: h.ErrorMessage,
			ScheduledAt:  h.ScheduledAt.Format(time.RFC3339),
			CreatedAt:    h.CreatedAt.Format(time.RFC3339),
		}

		if h.LastAttemptAt != nil {
			lastAttempt := h.LastAttemptAt.Format(time.RFC3339)
			responses[i].LastAttemptAt = &lastAttempt
		}

		if h.CompletedAt != nil {
			completed := h.CompletedAt.Format(time.RFC3339)
			responses[i].CompletedAt = &completed
		}
	}

	return responses, nil
}

// GetFailedDeliveries retrieves failed webhook deliveries for monitoring
func (s *WebhookHistoryService) GetFailedDeliveries(limit, offset int) ([]models.WebhookHistoryResponse, error) {
	history, err := s.webhookRepo.GetFailedDeliveries(limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]models.WebhookHistoryResponse, len(history))
	for i, h := range history {
		responses[i] = models.WebhookHistoryResponse{
			ID:           h.ID,
			ConfigID:     h.ConfigID,
			EventType:    h.EventType,
			EventID:      h.EventID,
			Status:       h.Status,
			AttemptCount: h.AttemptCount,
			ResponseCode: h.ResponseCode,
			ErrorMessage: h.ErrorMessage,
			ScheduledAt:  h.ScheduledAt.Format(time.RFC3339),
			CreatedAt:    h.CreatedAt.Format(time.RFC3339),
		}

		if h.LastAttemptAt != nil {
			lastAttempt := h.LastAttemptAt.Format(time.RFC3339)
			responses[i].LastAttemptAt = &lastAttempt
		}

		if h.CompletedAt != nil {
			completed := h.CompletedAt.Format(time.RFC3339)
			responses[i].CompletedAt = &completed
		}
	}

	return responses, nil
}

// GetEventsByStatus retrieves events by processing status
func (s *WebhookHistoryService) GetEventsByStatus(status string, limit, offset int) ([]models.WebhookEventResponse, error) {
	events, err := s.webhookRepo.GetEventsByStatus(status, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]models.WebhookEventResponse, len(events))
	for i, e := range events {
		responses[i] = models.WebhookEventResponse{
			ID:              e.ID,
			EventType:       e.EventType,
			EventID:         e.EventID,
			FPOID:           e.FPOID,
			ProcessedStatus: e.ProcessedStatus,
			ErrorMessage:    e.ErrorMessage,
			Source:          e.Source,
			ReceivedAt:      e.ReceivedAt.Format(time.RFC3339),
			CreatedAt:       e.CreatedAt.Format(time.RFC3339),
		}

		if e.ProcessedAt != nil {
			processed := e.ProcessedAt.Format(time.RFC3339)
			responses[i].ProcessedAt = &processed
		}
	}

	return responses, nil
}

// CleanupOldRecords removes old webhook records for maintenance
func (s *WebhookHistoryService) CleanupOldRecords(retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	utils.Info("Cleaning up webhook records older than", cutoffTime.Format("2006-01-02"))

	// Clean up history
	if err := s.webhookRepo.CleanupOldHistory(cutoffTime); err != nil {
		utils.Error("Failed to cleanup old webhook history:", err)
		return err
	}

	// Clean up processed events
	if err := s.webhookRepo.CleanupProcessedEvents(cutoffTime); err != nil {
		utils.Error("Failed to cleanup old webhook events:", err)
		return err
	}

	// Clean up completed queue items
	if err := s.webhookRepo.CleanupCompletedQueue(cutoffTime); err != nil {
		utils.Error("Failed to cleanup old webhook queue items:", err)
		return err
	}

	utils.Info("Webhook cleanup completed successfully")
	return nil
}

// GetWebhookStats returns statistics about webhook processing
func (s *WebhookHistoryService) GetWebhookStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Event stats
	eventStats, err := s.webhookRepo.GetEventStats()
	if err != nil {
		return nil, err
	}
	stats["events"] = eventStats

	// Queue stats
	queueStats, err := s.webhookRepo.GetQueueStats()
	if err != nil {
		return nil, err
	}
	stats["queue"] = queueStats

	return stats, nil
}