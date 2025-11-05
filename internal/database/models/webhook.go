package models

import (
	"kisanlink-erp/internal/constants"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// WebhookEvent represents a received webhook event for audit and idempotency tracking
type WebhookEvent struct {
	base.BaseModel

	// Event identification
	EventID   string `gorm:"type:varchar(100);unique;not null;index" json:"event_id"` // Unique event ID from e-commerce
	EventType string `gorm:"type:varchar(50);not null;index" json:"event_type"`       // order.created, order.confirmed, etc.

	// Payload tracking
	PayloadHash string `gorm:"type:varchar(64);not null" json:"payload_hash"` // SHA256 of payload for deduplication
	RequestBody string `gorm:"type:text;not null" json:"request_body"`        // Full JSON payload for debugging

	// Processing status
	Status       string     `gorm:"type:varchar(20);not null;index" json:"status"` // processing, success, failed
	ErrorMessage *string    `gorm:"type:text" json:"error_message"`                // Error details if failed
	ProcessedAt  *time.Time `gorm:"type:timestamptz" json:"processed_at"`

	// Reference tracking
	ExternalOrderID *string `gorm:"type:varchar(100);index" json:"external_order_id"` // For quick lookup
	PurchaseOrderID *string `gorm:"type:varchar(100)" json:"purchase_order_id"`      // Created/Updated PO ID

	// Webhook metadata
	SourceIP       *string `gorm:"type:varchar(50)" json:"source_ip"`        // Request source IP
	UserAgent      *string `gorm:"type:varchar(255)" json:"user_agent"`     // Request user agent
	SignatureValid bool    `gorm:"default:false" json:"signature_valid"`   // HMAC signature verification result
}

// NewWebhookEvent creates a new WebhookEvent with initialized fields
func NewWebhookEvent(eventID, eventType, payloadHash, requestBody string) *WebhookEvent {
	baseModel := base.NewBaseModel(constants.TableWebhookEvent, hash.Medium)
	return &WebhookEvent{
		BaseModel:      *baseModel,
		EventID:        eventID,
		EventType:      eventType,
		PayloadHash:    payloadHash,
		RequestBody:    requestBody,
		Status:         "processing",
		SignatureValid: false,
	}
}

func (WebhookEvent) TableName() string {
	return "webhook_events"
}

// WebhookDeliveryAttempt represents retry attempts for webhook processing
type WebhookDeliveryAttempt struct {
	base.BaseModel

	WebhookEventID string `gorm:"type:varchar(100);not null;index" json:"webhook_event_id"`
	AttemptNumber  int    `gorm:"type:int;not null" json:"attempt_number"`
	ResponseCode   int    `gorm:"type:int" json:"response_code"`
	ErrorMessage   *string `gorm:"type:text" json:"error_message"`
	AttemptedAt    time.Time `gorm:"type:timestamptz;not null;default:now()" json:"attempted_at"`

	// Associations
	WebhookEvent WebhookEvent `gorm:"foreignKey:WebhookEventID" json:"webhook_event,omitempty"`
}

// NewWebhookDeliveryAttempt creates a new WebhookDeliveryAttempt with initialized fields
func NewWebhookDeliveryAttempt(webhookEventID string, attemptNumber int) *WebhookDeliveryAttempt {
	baseModel := base.NewBaseModel(constants.TableWebhookDeliveryAttempt, hash.Medium)
	return &WebhookDeliveryAttempt{
		BaseModel:      *baseModel,
		WebhookEventID: webhookEventID,
		AttemptNumber:  attemptNumber,
		AttemptedAt:    time.Now(),
	}
}

func (WebhookDeliveryAttempt) TableName() string {
	return "webhook_delivery_attempts"
}

// WebhookEventResponse represents the API response for webhook event
type WebhookEventResponse struct {
	ID              string  `json:"id"`
	EventID         string  `json:"event_id"`
	EventType       string  `json:"event_type"`
	Status          string  `json:"status"`
	ErrorMessage    *string `json:"error_message"`
	ProcessedAt     *string `json:"processed_at"`
	ExternalOrderID *string `json:"external_order_id"`
	PurchaseOrderID *string `json:"purchase_order_id"`
	SignatureValid  bool    `json:"signature_valid"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}
