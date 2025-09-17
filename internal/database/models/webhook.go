package models

import (
	"kisanlink-erp/internal/constants"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// WebhookConfiguration represents webhook settings for an FPO
type WebhookConfiguration struct {
	base.BaseModel
	FPOID         string `gorm:"type:varchar(100);not null;index" json:"fpo_id"`
	WebhookURL    string `gorm:"type:text;not null" json:"webhook_url"`
	SecretKey     string `gorm:"type:varchar(255);not null" json:"secret_key"`
	Enabled       bool   `gorm:"default:true" json:"enabled"`
	RetryAttempts int    `gorm:"default:3" json:"retry_attempts"`
	TimeoutSecs   int    `gorm:"default:30" json:"timeout_seconds"`

	// Associations
	History []WebhookHistory `gorm:"foreignKey:ConfigID" json:"history,omitempty"`
}

func (WebhookConfiguration) TableName() string {
	return "webhook_configurations"
}

// WebhookHistory tracks webhook delivery attempts
type WebhookHistory struct {
	base.BaseModel
	ConfigID        string    `gorm:"type:varchar(100);not null;index" json:"config_id"`
	EventType       string    `gorm:"type:varchar(50);not null" json:"event_type"`
	EventID         string    `gorm:"type:varchar(100);not null;index" json:"event_id"`
	Payload         string    `gorm:"type:jsonb;not null" json:"payload"`
	Status          string    `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	AttemptCount    int       `gorm:"default:0" json:"attempt_count"`
	LastAttemptAt   *time.Time `gorm:"type:timestamptz" json:"last_attempt_at"`
	ResponseCode    *int       `gorm:"type:integer" json:"response_code"`
	ResponseBody    *string    `gorm:"type:text" json:"response_body"`
	ErrorMessage    *string    `gorm:"type:text" json:"error_message"`
	ScheduledAt     time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"scheduled_at"`
	CompletedAt     *time.Time `gorm:"type:timestamptz" json:"completed_at"`

	// Associations
	Config WebhookConfiguration `gorm:"foreignKey:ConfigID" json:"config,omitempty"`
}

func (WebhookHistory) TableName() string {
	return "webhook_history"
}

// WebhookEvent represents incoming webhook events for processing
type WebhookEvent struct {
	base.BaseModel
	EventType       string    `gorm:"type:varchar(50);not null" json:"event_type"`
	EventID         string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"event_id"`
	FPOID           string    `gorm:"type:varchar(100);not null;index" json:"fpo_id"`
	Payload         string    `gorm:"type:jsonb;not null" json:"payload"`
	ProcessedStatus string    `gorm:"type:varchar(20);not null;default:'pending'" json:"processed_status"`
	ProcessedAt     *time.Time `gorm:"type:timestamptz" json:"processed_at"`
	ErrorMessage    *string    `gorm:"type:text" json:"error_message"`
	Source          string     `gorm:"type:varchar(100)" json:"source"`
	ReceivedAt      time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"received_at"`
}

func (WebhookEvent) TableName() string {
	return "webhook_events"
}

// WebhookQueue represents outbound webhook delivery queue
type WebhookQueue struct {
	base.BaseModel
	ConfigID      string    `gorm:"type:varchar(100);not null;index" json:"config_id"`
	EventType     string    `gorm:"type:varchar(50);not null" json:"event_type"`
	Payload       string    `gorm:"type:jsonb;not null" json:"payload"`
	Status        string    `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	AttemptCount  int       `gorm:"default:0" json:"attempt_count"`
	NextRetryAt   time.Time `gorm:"type:timestamptz;not null;default:now()" json:"next_retry_at"`
	MaxRetries    int       `gorm:"default:3" json:"max_retries"`
	LastError     *string   `gorm:"type:text" json:"last_error"`

	// Associations
	Config WebhookConfiguration `gorm:"foreignKey:ConfigID" json:"config,omitempty"`
}

func (WebhookQueue) TableName() string {
	return "webhook_queue"
}

// Response DTOs
type WebhookConfigurationResponse struct {
	ID            string `json:"id"`
	FPOID         string `json:"fpo_id"`
	WebhookURL    string `json:"webhook_url"`
	Enabled       bool   `json:"enabled"`
	RetryAttempts int    `json:"retry_attempts"`
	TimeoutSecs   int    `json:"timeout_seconds"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type WebhookHistoryResponse struct {
	ID            string  `json:"id"`
	ConfigID      string  `json:"config_id"`
	EventType     string  `json:"event_type"`
	EventID       string  `json:"event_id"`
	Status        string  `json:"status"`
	AttemptCount  int     `json:"attempt_count"`
	LastAttemptAt *string `json:"last_attempt_at"`
	ResponseCode  *int    `json:"response_code"`
	ErrorMessage  *string `json:"error_message"`
	ScheduledAt   string  `json:"scheduled_at"`
	CompletedAt   *string `json:"completed_at"`
	CreatedAt     string  `json:"created_at"`
}

type WebhookEventResponse struct {
	ID              string  `json:"id"`
	EventType       string  `json:"event_type"`
	EventID         string  `json:"event_id"`
	FPOID           string  `json:"fpo_id"`
	ProcessedStatus string  `json:"processed_status"`
	ProcessedAt     *string `json:"processed_at"`
	ErrorMessage    *string `json:"error_message"`
	Source          string  `json:"source"`
	ReceivedAt      string  `json:"received_at"`
	CreatedAt       string  `json:"created_at"`
}

// Request DTOs
type CreateWebhookConfigRequest struct {
	FPOID         string `json:"fpo_id" binding:"required"`
	WebhookURL    string `json:"webhook_url" binding:"required,url"`
	SecretKey     string `json:"secret_key" binding:"required,min=32"`
	Enabled       *bool  `json:"enabled"`
	RetryAttempts *int   `json:"retry_attempts"`
	TimeoutSecs   *int   `json:"timeout_seconds"`
}

type UpdateWebhookConfigRequest struct {
	WebhookURL    *string `json:"webhook_url" binding:"omitempty,url"`
	SecretKey     *string `json:"secret_key" binding:"omitempty,min=32"`
	Enabled       *bool   `json:"enabled"`
	RetryAttempts *int    `json:"retry_attempts"`
	TimeoutSecs   *int    `json:"timeout_seconds"`
}

type InboundWebhookRequest struct {
	EventID       string                 `json:"event_id" binding:"required"`
	EventType     string                 `json:"event_type" binding:"required"`
	Timestamp     int64                  `json:"timestamp" binding:"required"`
	FPOID         string                 `json:"fpo_id" binding:"required"`
	Data          map[string]interface{} `json:"data" binding:"required"`
}

// Constructor functions
func NewWebhookConfiguration(fpoID, webhookURL, secretKey string) *WebhookConfiguration {
	baseModel := base.NewBaseModel(constants.TableWebhookConfig, hash.Medium)
	return &WebhookConfiguration{
		BaseModel:     *baseModel,
		FPOID:         fpoID,
		WebhookURL:    webhookURL,
		SecretKey:     secretKey,
		Enabled:       true,
		RetryAttempts: 3,
		TimeoutSecs:   30,
	}
}

func NewWebhookHistory(configID, eventType, eventID, payload string) *WebhookHistory {
	baseModel := base.NewBaseModel(constants.TableWebhookHistory, hash.Medium)
	return &WebhookHistory{
		BaseModel:   *baseModel,
		ConfigID:    configID,
		EventType:   eventType,
		EventID:     eventID,
		Payload:     payload,
		Status:      "pending",
		AttemptCount: 0,
		ScheduledAt: time.Now(),
	}
}

func NewWebhookEvent(eventType, eventID, fpoID, payload, source string) *WebhookEvent {
	baseModel := base.NewBaseModel(constants.TableWebhookEvent, hash.Medium)
	return &WebhookEvent{
		BaseModel:       *baseModel,
		EventType:       eventType,
		EventID:         eventID,
		FPOID:           fpoID,
		Payload:         payload,
		ProcessedStatus: "pending",
		Source:          source,
		ReceivedAt:      time.Now(),
	}
}

func NewWebhookQueue(configID, eventType, payload string, maxRetries int) *WebhookQueue {
	baseModel := base.NewBaseModel(constants.TableWebhookQueue, hash.Medium)
	return &WebhookQueue{
		BaseModel:    *baseModel,
		ConfigID:     configID,
		EventType:    eventType,
		Payload:      payload,
		Status:       "pending",
		AttemptCount: 0,
		NextRetryAt:  time.Now(),
		MaxRetries:   maxRetries,
	}
}