package repositories

import (
	"context"

	"kisanlink-erp/internal/database/models"

	"gorm.io/gorm"
)

// WebhookRepository handles database operations for webhook events
type WebhookRepository struct {
	db *gorm.DB
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *gorm.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// FindByEventID finds a webhook event by event_id
func (r *WebhookRepository) FindByEventID(ctx context.Context, eventID string) (*models.WebhookEvent, error) {
	var event models.WebhookEvent
	if err := r.db.WithContext(ctx).Where("event_id = ?", eventID).First(&event).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found, but not an error
		}
		return nil, err
	}
	return &event, nil
}

// Create creates a new webhook event
func (r *WebhookRepository) Create(ctx context.Context, event *models.WebhookEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// Update updates an existing webhook event
// Update updates an existing webhook event
func (r *WebhookRepository) Update(ctx context.Context, event *models.WebhookEvent) error {
	// Use explicit UPDATE with WHERE clause to avoid Save() issues with custom ID fields
	return r.db.WithContext(ctx).Model(&models.WebhookEvent{}).
		Where("id = ?", event.ID).
		Updates(map[string]interface{}{
			"status":            event.Status,
			"processed_at":      event.ProcessedAt,
			"error_message":     event.ErrorMessage,
			"purchase_order_id": event.PurchaseOrderID,
			"external_order_id": event.ExternalOrderID,
			"updated_at":        event.UpdatedAt,
		}).Error
}

// FindByExternalOrderID finds all webhook events for a specific external order ID
func (r *WebhookRepository) FindByExternalOrderID(ctx context.Context, externalOrderID string) ([]models.WebhookEvent, error) {
	var events []models.WebhookEvent
	if err := r.db.WithContext(ctx).
		Where("external_order_id = ?", externalOrderID).
		Order("created_at DESC").
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// FindByStatus finds all webhook events with a specific status
func (r *WebhookRepository) FindByStatus(ctx context.Context, status string, limit int) ([]models.WebhookEvent, error) {
	var events []models.WebhookEvent
	query := r.db.WithContext(ctx).Where("status = ?", status).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// CreateDeliveryAttempt creates a new delivery attempt record
func (r *WebhookRepository) CreateDeliveryAttempt(ctx context.Context, attempt *models.WebhookDeliveryAttempt) error {
	return r.db.WithContext(ctx).Create(attempt).Error
}

// GetDeliveryAttempts retrieves all delivery attempts for a webhook event
func (r *WebhookRepository) GetDeliveryAttempts(ctx context.Context, webhookEventID string) ([]models.WebhookDeliveryAttempt, error) {
	var attempts []models.WebhookDeliveryAttempt
	if err := r.db.WithContext(ctx).
		Where("webhook_event_id = ?", webhookEventID).
		Order("attempt_number ASC").
		Find(&attempts).Error; err != nil {
		return nil, err
	}
	return attempts, nil
}

// CountByEventType counts webhook events by event type
func (r *WebhookRepository) CountByEventType(ctx context.Context, eventType string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.WebhookEvent{}).
		Where("event_type = ?", eventType).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountByStatusAndEventType counts webhook events by status and event type
func (r *WebhookRepository) CountByStatusAndEventType(ctx context.Context, status, eventType string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.WebhookEvent{}).
		Where("status = ? AND event_type = ?", status, eventType).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// FindRecent finds recent webhook events (most recent first)
func (r *WebhookRepository) FindRecent(ctx context.Context, limit int) ([]models.WebhookEvent, error) {
	var events []models.WebhookEvent
	query := r.db.WithContext(ctx).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}
