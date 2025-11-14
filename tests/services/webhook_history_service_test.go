package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/tests/testutils"

	"gorm.io/gorm"
)

// setupWebhookHistoryService creates a WebhookHistoryService with test database
func setupWebhookHistoryService(t *testing.T) (*services.WebhookHistoryService, *gorm.DB, func()) {
	db := testutils.SetupTestDB(t)

	// Create repository
	webhookRepo := repositories.NewWebhookRepository(db)

	// Create service
	service := services.NewWebhookHistoryService(webhookRepo)

	// Cleanup function
	cleanup := func() {
		testutils.CleanupTestDB(db)
	}

	return service, db, cleanup
}

// ============================================================================
// CheckIdempotency Tests
// ============================================================================

func TestWebhookHistory_CheckIdempotency_FirstTime(t *testing.T) {
	service, _, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	// Test: Check idempotency for new event ID
	alreadyProcessed, existingEvent, err := service.CheckIdempotency(ctx, "event-new-12345")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if alreadyProcessed {
		t.Error("Expected alreadyProcessed to be false for new event, got true")
	}
	if existingEvent != nil {
		t.Errorf("Expected existingEvent to be nil for new event, got: %v", existingEvent)
	}
}

func TestWebhookHistory_CheckIdempotency_AlreadySuccess(t *testing.T) {
	service, db, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	// Create existing successful webhook event
	existingWebhook := models.NewWebhookEvent("event-success-123", "order.created", "hash123", "payload")
	existingWebhook.Status = "success"
	if err := db.Create(existingWebhook).Error; err != nil {
		t.Fatalf("Failed to create test webhook: %v", err)
	}

	// Test: Check idempotency
	alreadyProcessed, existingEvent, err := service.CheckIdempotency(ctx, "event-success-123")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !alreadyProcessed {
		t.Error("Expected alreadyProcessed to be true for successful event, got false")
	}
	if existingEvent == nil {
		t.Fatal("Expected existingEvent to be returned, got nil")
	}
	if existingEvent.EventID != "event-success-123" {
		t.Errorf("Expected event ID 'event-success-123', got: %s", existingEvent.EventID)
	}
	if existingEvent.Status != "success" {
		t.Errorf("Expected status 'success', got: %s", existingEvent.Status)
	}
}

func TestWebhookHistory_CheckIdempotency_AlreadyProcessing(t *testing.T) {
	service, db, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	// Create existing processing webhook event
	existingWebhook := models.NewWebhookEvent("event-processing-456", "order.created", "hash456", "payload")
	existingWebhook.Status = "processing"
	if err := db.Create(existingWebhook).Error; err != nil {
		t.Fatalf("Failed to create test webhook: %v", err)
	}

	// Test: Check idempotency
	alreadyProcessed, existingEvent, err := service.CheckIdempotency(ctx, "event-processing-456")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !alreadyProcessed {
		t.Error("Expected alreadyProcessed to be true for processing event, got false")
	}
	if existingEvent == nil {
		t.Fatal("Expected existingEvent to be returned, got nil")
	}
	if existingEvent.Status != "processing" {
		t.Errorf("Expected status 'processing', got: %s", existingEvent.Status)
	}
}

func TestWebhookHistory_CheckIdempotency_PreviouslyFailed_AllowRetry(t *testing.T) {
	service, db, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	// Create existing failed webhook event
	existingWebhook := models.NewWebhookEvent("event-failed-789", "order.created", "hash789", "payload")
	existingWebhook.Status = "failed"
	errorMsg := "Previous error"
	existingWebhook.ErrorMessage = &errorMsg
	if err := db.Create(existingWebhook).Error; err != nil {
		t.Fatalf("Failed to create test webhook: %v", err)
	}

	// Test: Check idempotency
	alreadyProcessed, existingEvent, err := service.CheckIdempotency(ctx, "event-failed-789")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if alreadyProcessed {
		t.Error("Expected alreadyProcessed to be false for failed event (allow retry), got true")
	}
	if existingEvent == nil {
		t.Fatal("Expected existingEvent to be returned (to track retry), got nil")
	}
	if existingEvent.Status != "failed" {
		t.Errorf("Expected status 'failed', got: %s", existingEvent.Status)
	}
}

// ============================================================================
// RecordWebhook Tests
// ============================================================================

func TestWebhookHistory_RecordWebhook_Success(t *testing.T) {
	service, db, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	// Create new webhook event
	newEvent := models.NewWebhookEvent("event-record-001", "order.created", "hash001", `{"data":"test"}`)

	// Test: Record webhook
	err := service.RecordWebhook(ctx, newEvent)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify event was saved in database
	var savedEvent models.WebhookEvent
	result := db.Where("event_id = ?", "event-record-001").First(&savedEvent)
	if result.Error != nil {
		t.Errorf("Expected to find saved event, got error: %v", result.Error)
	}
	if savedEvent.EventID != "event-record-001" {
		t.Errorf("Expected event ID 'event-record-001', got: %s", savedEvent.EventID)
	}
	if savedEvent.Status != "processing" {
		t.Errorf("Expected default status 'processing', got: %s", savedEvent.Status)
	}
}

func TestWebhookHistory_RecordWebhook_SetsDefaultStatus(t *testing.T) {
	service, db, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	// Create new webhook event without status
	newEvent := models.NewWebhookEvent("event-record-002", "order.confirmed", "hash002", `{"data":"test"}`)
	newEvent.Status = "" // Explicitly empty

	// Test: Record webhook
	err := service.RecordWebhook(ctx, newEvent)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify default status was set
	var savedEvent models.WebhookEvent
	result := db.Where("event_id = ?", "event-record-002").First(&savedEvent)
	if result.Error != nil {
		t.Fatalf("Expected to find saved event, got error: %v", result.Error)
	}
	if savedEvent.Status != "processing" {
		t.Errorf("Expected default status 'processing', got: %s", savedEvent.Status)
	}
}

// ============================================================================
// MarkProcessed Tests
// ============================================================================

func TestWebhookHistory_MarkProcessed_Success(t *testing.T) {
	service, db, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	// Create webhook in processing status
	webhook := models.NewWebhookEvent("event-mark-success-001", "order.created", "hash001", "payload")
	webhook.Status = "processing"
	if err := db.Create(webhook).Error; err != nil {
		t.Fatalf("Failed to create test webhook: %v", err)
	}

	// Test: Mark as processed
	err := service.MarkProcessed(ctx, "event-mark-success-001", "PO123")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify event was updated
	var updatedEvent models.WebhookEvent
	result := db.Where("event_id = ?", "event-mark-success-001").First(&updatedEvent)
	if result.Error != nil {
		t.Fatalf("Expected to find updated event, got error: %v", result.Error)
	}
	if updatedEvent.Status != "success" {
		t.Errorf("Expected status 'success', got: %s", updatedEvent.Status)
	}
	if updatedEvent.ProcessedAt == nil {
		t.Error("Expected ProcessedAt to be set, got nil")
	}
	if updatedEvent.PurchaseOrderID == nil || *updatedEvent.PurchaseOrderID != "PO123" {
		t.Errorf("Expected PurchaseOrderID 'PO123', got: %v", updatedEvent.PurchaseOrderID)
	}
	if updatedEvent.ErrorMessage != nil {
		t.Errorf("Expected ErrorMessage to be nil, got: %v", *updatedEvent.ErrorMessage)
	}
}

func TestWebhookHistory_MarkProcessed_NotFound(t *testing.T) {
	service, _, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	// Test: Mark non-existent event as processed
	err := service.MarkProcessed(ctx, "event-nonexistent", "PO123")

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent event, got nil")
	}
}

// ============================================================================
// MarkFailed Tests
// ============================================================================

func TestWebhookHistory_MarkFailed_Success(t *testing.T) {
	service, db, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	// Create webhook in processing status
	webhook := models.NewWebhookEvent("event-mark-failed-001", "order.created", "hash001", "payload")
	webhook.Status = "processing"
	if err := db.Create(webhook).Error; err != nil {
		t.Fatalf("Failed to create test webhook: %v", err)
	}

	// Test: Mark as failed
	testError := errors.New("Test error: something went wrong")
	err := service.MarkFailed(ctx, "event-mark-failed-001", testError)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify event was updated
	var updatedEvent models.WebhookEvent
	result := db.Where("event_id = ?", "event-mark-failed-001").First(&updatedEvent)
	if result.Error != nil {
		t.Fatalf("Expected to find updated event, got error: %v", result.Error)
	}
	if updatedEvent.Status != "failed" {
		t.Errorf("Expected status 'failed', got: %s", updatedEvent.Status)
	}
	if updatedEvent.ProcessedAt == nil {
		t.Error("Expected ProcessedAt to be set, got nil")
	}
	if updatedEvent.ErrorMessage == nil {
		t.Fatal("Expected ErrorMessage to be set, got nil")
	}
	if *updatedEvent.ErrorMessage != testError.Error() {
		t.Errorf("Expected error message '%s', got: %s", testError.Error(), *updatedEvent.ErrorMessage)
	}
}

func TestWebhookHistory_MarkFailed_NotFound(t *testing.T) {
	service, _, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	// Test: Mark non-existent event as failed
	testError := errors.New("Test error")
	err := service.MarkFailed(ctx, "event-nonexistent", testError)

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent event, got nil")
	}
}

// ============================================================================
// CreateWebhookEvent Tests
// ============================================================================

func TestWebhookHistory_CreateWebhookEvent_SHA256Hash(t *testing.T) {
	service, _, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	// Test payload
	payload := []byte(`{"event":"order.created","orderId":"12345"}`)

	// Calculate expected hash
	hash := sha256.Sum256(payload)
	expectedHash := hex.EncodeToString(hash[:])

	// Test metadata
	externalOrderID := "ORDER-12345"
	sourceIP := "192.168.1.1"
	userAgent := "EcommerceWebhook/1.0"

	// Test: Create webhook event
	event := service.CreateWebhookEvent(
		"event-create-001",
		"order.created",
		payload,
		&externalOrderID,
		sourceIP,
		userAgent,
		true,
	)

	// Assert
	if event == nil {
		t.Fatal("Expected event to be created, got nil")
	}
	if event.EventID != "event-create-001" {
		t.Errorf("Expected event ID 'event-create-001', got: %s", event.EventID)
	}
	if event.EventType != "order.created" {
		t.Errorf("Expected event type 'order.created', got: %s", event.EventType)
	}
	if event.PayloadHash != expectedHash {
		t.Errorf("Expected payload hash '%s', got: %s", expectedHash, event.PayloadHash)
	}
	if event.RequestBody != string(payload) {
		t.Errorf("Expected request body '%s', got: %s", string(payload), event.RequestBody)
	}
	if event.ExternalOrderID == nil || *event.ExternalOrderID != externalOrderID {
		t.Errorf("Expected external order ID '%s', got: %v", externalOrderID, event.ExternalOrderID)
	}
	if event.SourceIP == nil || *event.SourceIP != sourceIP {
		t.Errorf("Expected source IP '%s', got: %v", sourceIP, event.SourceIP)
	}
	if event.UserAgent == nil || *event.UserAgent != userAgent {
		t.Errorf("Expected user agent '%s', got: %v", userAgent, event.UserAgent)
	}
	if !event.SignatureValid {
		t.Error("Expected signature valid to be true, got false")
	}
}

func TestWebhookHistory_CreateWebhookEvent_EmptyPayload(t *testing.T) {
	service, _, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	// Test with empty payload
	emptyPayload := []byte("")

	// Calculate expected hash
	hash := sha256.Sum256(emptyPayload)
	expectedHash := hex.EncodeToString(hash[:])

	// Test: Create webhook event
	event := service.CreateWebhookEvent(
		"event-empty-001",
		"order.created",
		emptyPayload,
		nil,
		"",
		"",
		false,
	)

	// Assert
	if event == nil {
		t.Fatal("Expected event to be created, got nil")
	}
	if event.PayloadHash != expectedHash {
		t.Errorf("Expected payload hash '%s', got: %s", expectedHash, event.PayloadHash)
	}
}

// ============================================================================
// RecordDeliveryAttempt Tests
// ============================================================================

func TestWebhookHistory_RecordDeliveryAttempt_Success(t *testing.T) {
	t.Skip("Skipping: webhook_delivery_attempts table not in SQLite test schema")

	service, db, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	// Create webhook event first
	webhook := models.NewWebhookEvent("event-delivery-001", "order.created", "hash001", "payload")
	if err := db.Create(webhook).Error; err != nil {
		t.Fatalf("Failed to create test webhook: %v", err)
	}

	// Test: Record delivery attempt
	errorMsg := "Connection timeout"
	err := service.RecordDeliveryAttempt(ctx, webhook.ID, 1, 500, &errorMsg)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify delivery attempt was saved
	var attempt models.WebhookDeliveryAttempt
	result := db.Where("webhook_event_id = ? AND attempt_number = ?", webhook.ID, 1).First(&attempt)
	if result.Error != nil {
		t.Errorf("Expected to find delivery attempt, got error: %v", result.Error)
	}
	if attempt.ResponseCode != 500 {
		t.Errorf("Expected response code 500, got: %d", attempt.ResponseCode)
	}
	if attempt.ErrorMessage == nil || *attempt.ErrorMessage != errorMsg {
		t.Errorf("Expected error message '%s', got: %v", errorMsg, attempt.ErrorMessage)
	}
}

// ============================================================================
// GetWebhookEventsByOrderID Tests
// ============================================================================

func TestWebhookHistory_GetWebhookEventsByOrderID_Success(t *testing.T) {
	service, db, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	externalOrderID := "ORDER-12345"

	// Create multiple webhooks for same order
	webhook1 := models.NewWebhookEvent("event-order-001", "order.created", "hash001", "payload1")
	webhook1.ExternalOrderID = &externalOrderID
	if err := db.Create(webhook1).Error; err != nil {
		t.Fatalf("Failed to create webhook 1: %v", err)
	}

	webhook2 := models.NewWebhookEvent("event-order-002", "order.confirmed", "hash002", "payload2")
	webhook2.ExternalOrderID = &externalOrderID
	if err := db.Create(webhook2).Error; err != nil {
		t.Fatalf("Failed to create webhook 2: %v", err)
	}

	// Test: Get webhooks by order ID
	events, err := service.GetWebhookEventsByOrderID(ctx, externalOrderID)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got: %d", len(events))
	}
}

func TestWebhookHistory_GetWebhookEventsByOrderID_Empty(t *testing.T) {
	service, _, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	// Test: Get webhooks for non-existent order
	events, err := service.GetWebhookEventsByOrderID(ctx, "ORDER-NONEXISTENT")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("Expected 0 events, got: %d", len(events))
	}
}

// ============================================================================
// GetFailedWebhooks Tests
// ============================================================================

func TestWebhookHistory_GetFailedWebhooks_Success(t *testing.T) {
	service, db, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	// Create failed webhooks
	for i := 1; i <= 3; i++ {
		webhook := models.NewWebhookEvent("event-failed-"+string(rune('0'+i)), "order.created", "hash", "payload")
		webhook.Status = "failed"
		if err := db.Create(webhook).Error; err != nil {
			t.Fatalf("Failed to create webhook %d: %v", i, err)
		}
	}

	// Create successful webhook (should not be returned)
	webhookSuccess := models.NewWebhookEvent("event-success-999", "order.created", "hash", "payload")
	webhookSuccess.Status = "success"
	if err := db.Create(webhookSuccess).Error; err != nil {
		t.Fatalf("Failed to create success webhook: %v", err)
	}

	// Test: Get failed webhooks
	events, err := service.GetFailedWebhooks(ctx, 10)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if len(events) != 3 {
		t.Errorf("Expected 3 failed events, got: %d", len(events))
	}
	// Verify all returned events are failed
	for _, event := range events {
		if event.Status != "failed" {
			t.Errorf("Expected all events to have status 'failed', got: %s", event.Status)
		}
	}
}

func TestWebhookHistory_GetFailedWebhooks_RespectsLimit(t *testing.T) {
	service, db, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	// Create 5 failed webhooks
	for i := 1; i <= 5; i++ {
		webhook := models.NewWebhookEvent("event-failed-limit-"+string(rune('0'+i)), "order.created", "hash", "payload")
		webhook.Status = "failed"
		if err := db.Create(webhook).Error; err != nil {
			t.Fatalf("Failed to create webhook %d: %v", i, err)
		}
	}

	// Test: Get failed webhooks with limit of 2
	events, err := service.GetFailedWebhooks(ctx, 2)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("Expected 2 events (limit), got: %d", len(events))
	}
}

// ============================================================================
// GetWebhookStats Tests
// ============================================================================

func TestWebhookHistory_GetWebhookStats_Success(t *testing.T) {
	service, db, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	// Create webhook events with different types and statuses
	// order.created: 2 success, 1 failed
	webhook1 := models.NewWebhookEvent("event-stats-001", "order.created", "hash1", "payload")
	webhook1.Status = "success"
	if err := db.Create(webhook1).Error; err != nil {
		t.Fatalf("Failed to create webhook 1: %v", err)
	}

	webhook2 := models.NewWebhookEvent("event-stats-002", "order.created", "hash2", "payload")
	webhook2.Status = "success"
	if err := db.Create(webhook2).Error; err != nil {
		t.Fatalf("Failed to create webhook 2: %v", err)
	}

	webhook3 := models.NewWebhookEvent("event-stats-003", "order.created", "hash3", "payload")
	webhook3.Status = "failed"
	if err := db.Create(webhook3).Error; err != nil {
		t.Fatalf("Failed to create webhook 3: %v", err)
	}

	// order.confirmed: 1 success
	webhook4 := models.NewWebhookEvent("event-stats-004", "order.confirmed", "hash4", "payload")
	webhook4.Status = "success"
	if err := db.Create(webhook4).Error; err != nil {
		t.Fatalf("Failed to create webhook 4: %v", err)
	}

	// Test: Get webhook stats
	stats, err := service.GetWebhookStats(ctx)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if stats == nil {
		t.Fatal("Expected stats to be returned, got nil")
	}

	// Check order.created stats
	if createdStats, ok := stats["order.created"].(map[string]interface{}); ok {
		if total, ok := createdStats["total"].(int64); ok && total != 3 {
			t.Errorf("Expected total 3 for order.created, got: %d", total)
		}
		if success, ok := createdStats["success"].(int64); ok && success != 2 {
			t.Errorf("Expected success 2 for order.created, got: %d", success)
		}
		if failed, ok := createdStats["failed"].(int64); ok && failed != 1 {
			t.Errorf("Expected failed 1 for order.created, got: %d", failed)
		}
	} else {
		t.Error("Expected stats for order.created, got none")
	}

	// Check order.confirmed stats
	if confirmedStats, ok := stats["order.confirmed"].(map[string]interface{}); ok {
		if total, ok := confirmedStats["total"].(int64); ok && total != 1 {
			t.Errorf("Expected total 1 for order.confirmed, got: %d", total)
		}
		if success, ok := confirmedStats["success"].(int64); ok && success != 1 {
			t.Errorf("Expected success 1 for order.confirmed, got: %d", success)
		}
	} else {
		t.Error("Expected stats for order.confirmed, got none")
	}
}

func TestWebhookHistory_GetWebhookStats_EmptyDatabase(t *testing.T) {
	service, _, cleanup := setupWebhookHistoryService(t)
	defer cleanup()

	ctx := context.Background()

	// Test: Get stats with no webhooks
	stats, err := service.GetWebhookStats(ctx)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if stats == nil {
		t.Fatal("Expected stats to be returned, got nil")
	}

	// All event types should have 0 counts
	eventTypes := []string{"order.created", "order.confirmed", "order.shipped", "order.delivered", "order.payment"}
	for _, eventType := range eventTypes {
		if typeStats, ok := stats[eventType].(map[string]interface{}); ok {
			if total, ok := typeStats["total"].(int64); ok && total != 0 {
				t.Errorf("Expected total 0 for %s, got: %d", eventType, total)
			}
		} else {
			t.Errorf("Expected stats for %s, got none", eventType)
		}
	}
}
