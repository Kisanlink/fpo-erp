package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/utils"
)

// OutboundWebhookService handles sending webhooks to e-commerce platforms
type OutboundWebhookService struct {
	webhookRepo         *repositories.WebhookRepository
	historyService      *WebhookHistoryService
	securityService     *WebhookSecurityService
	httpClient          *http.Client
	warehouseRepo       *repositories.WarehouseRepository
	productRepo         *repositories.ProductRepository
}

// NewOutboundWebhookService creates a new outbound webhook service
func NewOutboundWebhookService(
	webhookRepo *repositories.WebhookRepository,
	historyService *WebhookHistoryService,
	securityService *WebhookSecurityService,
	warehouseRepo *repositories.WarehouseRepository,
	productRepo *repositories.ProductRepository,
) *OutboundWebhookService {
	return &OutboundWebhookService{
		webhookRepo:     webhookRepo,
		historyService:  historyService,
		securityService: securityService,
		warehouseRepo:   warehouseRepo,
		productRepo:     productRepo,
		httpClient: &http.Client{
			Timeout: time.Second * 30, // Default timeout, will be overridden per config
		},
	}
}

// FPOSaleCompletedPayload represents the payload for completed FPO sales
type FPOSaleCompletedPayload struct {
	EventType    string                   `json:"event_type"`
	FPOID        string                   `json:"fpo_id"`
	SaleDate     string                   `json:"sale_date"`
	TotalAmount  float64                  `json:"total_amount"`
	Currency     string                   `json:"currency"`
	Warehouse    WarehouseInfo            `json:"warehouse"`
	Items        []FPOSaleCompletedItem   `json:"items"`
}

type WarehouseInfo struct {
	Name    string      `json:"name"`
	Address AddressInfo `json:"address"`
}

type AddressInfo struct {
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
}

type FPOSaleCompletedItem struct {
	ProductDetails ProductInfo `json:"product_details"`
	Quantity       int64       `json:"quantity"`
	UnitPrice      float64     `json:"unit_price"`
	TotalPrice     float64     `json:"total_price"`
	BatchInfo      BatchInfo   `json:"batch_info"`
}

type ProductInfo struct {
	Name     string `json:"name"`
	SKU      string `json:"sku"`
	Category string `json:"category"`
	Brand    string `json:"brand"`
}

type BatchInfo struct {
	ExpiryDate   string  `json:"expiry_date"`
	BatchNumber  string  `json:"batch_number"`
	CostPrice    float64 `json:"cost_price"`
}

// SendFPOSaleNotification sends a notification when a sale is completed at FPO warehouse
func (s *OutboundWebhookService) SendFPOSaleNotification(saleID, fpoID string, source string) error {
	// Prevent circular webhooks - don't send notifications for sales that originated from e-commerce
	if source == "e-commerce" {
		utils.Info("Skipping outbound webhook for e-commerce originated sale:", saleID)
		return nil
	}

	// Get webhook configuration for this FPO
	config, err := s.webhookRepo.GetConfigByFPO(fpoID)
	if err != nil {
		utils.Info("No webhook configuration found for FPO:", fpoID)
		return nil // Not an error - just no webhook configured
	}

	if !config.Enabled {
		utils.Info("Webhook disabled for FPO:", fpoID)
		return nil
	}

	// Build the webhook payload
	payload, err := s.buildSaleNotificationPayload(saleID)
	if err != nil {
		return fmt.Errorf("failed to build webhook payload: %w", err)
	}

	// Create webhook history record
	eventID := fmt.Sprintf("fpo_sale_%s_%d", saleID, time.Now().Unix())
	history, err := s.historyService.LogWebhookDelivery(config.ID, "fpo_sale_completed", eventID, payload)
	if err != nil {
		return fmt.Errorf("failed to log webhook delivery: %w", err)
	}

	// Send the webhook
	success, responseCode, responseBody, deliveryErr := s.deliverWebhook(config, payload)

	// Update delivery status
	if success {
		s.historyService.UpdateDeliverySuccess(history.ID, responseCode, responseBody)
		utils.Info("Successfully sent FPO sale notification webhook:", eventID)
	} else {
		errorMsg := "Webhook delivery failed"
		if deliveryErr != nil {
			errorMsg = deliveryErr.Error()
		}
		s.historyService.UpdateDeliveryFailure(history.ID, &responseCode, responseBody, errorMsg)

		// Queue for retry if configured
		if config.RetryAttempts > 0 {
			s.queueForRetry(config.ID, "fpo_sale_completed", payload, config.RetryAttempts)
		}
	}

	return nil
}

// buildSaleNotificationPayload builds the webhook payload for a sale notification
func (s *OutboundWebhookService) buildSaleNotificationPayload(saleID string) (*FPOSaleCompletedPayload, error) {
	// This is a simplified implementation - in a real scenario you'd fetch the sale data
	// from the sales repository and build the complete payload

	// TODO: Implement proper sale data retrieval and payload building
	// For now, return a minimal payload structure
	payload := &FPOSaleCompletedPayload{
		EventType:   "fpo_sale_completed",
		FPOID:       "placeholder_fpo_id",
		SaleDate:    time.Now().Format(time.RFC3339),
		TotalAmount: 0.0,
		Currency:    "INR",
		Warehouse: WarehouseInfo{
			Name: "Main Warehouse",
			Address: AddressInfo{
				AddressLine1: "123 Market Street",
				City:         "Mumbai",
				State:        "Maharashtra",
				PostalCode:   "400001",
				Country:      "India",
			},
		},
		Items: []FPOSaleCompletedItem{},
	}

	return payload, nil
}

// deliverWebhook sends the webhook to the configured endpoint
func (s *OutboundWebhookService) deliverWebhook(config *models.WebhookConfiguration, payload interface{}) (success bool, responseCode int, responseBody string, err error) {
	// Serialize payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return false, 0, "", fmt.Errorf("failed to serialize payload: %w", err)
	}

	// Generate timestamp and signature
	timestamp := time.Now().Unix()
	signature := s.securityService.GenerateHMACSignature(string(payloadBytes), config.SecretKey)

	// Create HTTP request
	req, err := http.NewRequest("POST", config.WebhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return false, 0, "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Kisanlink-Signature", signature)
	req.Header.Set("X-Kisanlink-Timestamp", fmt.Sprintf("%d", timestamp))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TimeoutSecs)*time.Second)
	defer cancel()

	req = req.WithContext(ctx)

	// Send the request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, 0, "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, resp.StatusCode, "", fmt.Errorf("failed to read response body: %w", err)
	}

	responseBody = string(body)
	responseCode = resp.StatusCode

	// Consider 2xx status codes as success
	success = responseCode >= 200 && responseCode < 300

	if !success {
		err = fmt.Errorf("webhook returned non-2xx status code: %d", responseCode)
	}

	return success, responseCode, responseBody, err
}

// queueForRetry adds a failed webhook to the retry queue
func (s *OutboundWebhookService) queueForRetry(configID, eventType string, payload interface{}, maxRetries int) error {
	// Serialize payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create queue item with exponential backoff for retry
	queueItem := models.NewWebhookQueue(configID, eventType, string(payloadBytes), maxRetries)
	queueItem.NextRetryAt = time.Now().Add(time.Minute * 1) // Retry after 1 minute

	return s.webhookRepo.CreateQueueItem(queueItem)
}

// ProcessRetryQueue processes pending webhook retries
func (s *OutboundWebhookService) ProcessRetryQueue(batchSize int) error {
	// Get pending items from queue
	items, err := s.webhookRepo.GetPendingQueueItems(batchSize)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := s.processQueueItem(&item); err != nil {
			utils.Error("Failed to process queue item:", item.ID, "Error:", err)
			continue
		}
	}

	return nil
}

// processQueueItem processes a single queue item
func (s *OutboundWebhookService) processQueueItem(item *models.WebhookQueue) error {
	// Deserialize payload
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(item.Payload), &payload); err != nil {
		return s.markQueueItemFailed(item, "Invalid payload JSON")
	}

	// Send webhook
	success, responseCode, responseBody, err := s.deliverWebhook(&item.Config, payload)

	if success {
		// Mark as completed and remove from queue
		item.Status = "completed"
		s.webhookRepo.UpdateQueueItem(item)
		s.webhookRepo.DeleteQueueItem(item.ID)
		return nil
	}

	// Update attempt count
	item.AttemptCount++

	if item.AttemptCount >= item.MaxRetries {
		// Max retries reached - mark as failed
		errorMsg := "Max retries reached"
		if err != nil {
			errorMsg = err.Error()
		}
		return s.markQueueItemFailed(item, errorMsg)
	}

	// Schedule next retry with exponential backoff
	backoffMinutes := 1 << item.AttemptCount // 2, 4, 8, 16 minutes...
	item.NextRetryAt = time.Now().Add(time.Duration(backoffMinutes) * time.Minute)

	if err != nil {
		errorMsg := err.Error()
		item.LastError = &errorMsg
	} else {
		errorMsg := fmt.Sprintf("HTTP %d: %s", responseCode, responseBody)
		item.LastError = &errorMsg
	}

	return s.webhookRepo.UpdateQueueItem(item)
}

// markQueueItemFailed marks a queue item as permanently failed
func (s *OutboundWebhookService) markQueueItemFailed(item *models.WebhookQueue, errorMsg string) error {
	item.Status = "failed"
	item.LastError = &errorMsg

	if err := s.webhookRepo.UpdateQueueItem(item); err != nil {
		return err
	}

	utils.Error("Webhook permanently failed:", item.ID, "Error:", errorMsg)
	return nil
}

// StartRetryProcessor starts a background process to handle webhook retries
func (s *OutboundWebhookService) StartRetryProcessor(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 5) // Process retries every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			utils.Info("Webhook retry processor stopped")
			return
		case <-ticker.C:
			if err := s.ProcessRetryQueue(10); err != nil {
				utils.Error("Error processing webhook retry queue:", err)
			}
		}
	}
}

// GetQueueStats returns statistics about the webhook queue
func (s *OutboundWebhookService) GetQueueStats() (map[string]int64, error) {
	return s.webhookRepo.GetQueueStats()
}

// CleanupOldQueueItems removes old completed/failed items from the queue
func (s *OutboundWebhookService) CleanupOldQueueItems(retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	return s.webhookRepo.CleanupCompletedQueue(cutoffTime)
}