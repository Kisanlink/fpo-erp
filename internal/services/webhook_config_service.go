package services

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
)

// WebhookConfigService handles webhook configuration business logic
type WebhookConfigService struct {
	webhookRepo     *repositories.WebhookRepository
	historyService  *WebhookHistoryService
}

// NewWebhookConfigService creates a new webhook configuration service
func NewWebhookConfigService(
	webhookRepo *repositories.WebhookRepository,
	historyService *WebhookHistoryService,
) *WebhookConfigService {
	return &WebhookConfigService{
		webhookRepo:    webhookRepo,
		historyService: historyService,
	}
}

// CreateConfig creates a new webhook configuration
func (s *WebhookConfigService) CreateConfig(req *models.CreateWebhookConfigRequest) (*models.WebhookConfigurationResponse, error) {
	// Check if configuration already exists for this FPO
	existing, err := s.webhookRepo.GetConfigByFPO(req.FPOID)
	if err == nil && existing != nil {
		return nil, errors.NewConflictError("Configuration already exists for this FPO")
	}

	// Generate secure random secret if not provided
	secretKey := req.SecretKey
	if len(secretKey) < 32 {
		secretKey = s.generateSecureSecret()
	}

	// Set default values
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	retryAttempts := 3
	if req.RetryAttempts != nil {
		retryAttempts = *req.RetryAttempts
	}

	timeoutSecs := 30
	if req.TimeoutSecs != nil {
		timeoutSecs = *req.TimeoutSecs
	}

	// Create configuration
	config := models.NewWebhookConfiguration(req.FPOID, req.WebhookURL, secretKey)
	config.Enabled = enabled
	config.RetryAttempts = retryAttempts
	config.TimeoutSecs = timeoutSecs

	if err := s.webhookRepo.CreateConfig(config); err != nil {
		return nil, err
	}

	return s.convertToResponse(config), nil
}

// GetConfig retrieves a webhook configuration by ID
func (s *WebhookConfigService) GetConfig(id string) (*models.WebhookConfigurationResponse, error) {
	config, err := s.webhookRepo.GetConfigByID(id)
	if err != nil {
		return nil, err
	}

	return s.convertToResponse(config), nil
}

// GetConfigByFPO retrieves a webhook configuration by FPO ID
func (s *WebhookConfigService) GetConfigByFPO(fpoID string) (*models.WebhookConfigurationResponse, error) {
	config, err := s.webhookRepo.GetConfigByFPO(fpoID)
	if err != nil {
		return nil, err
	}

	return s.convertToResponse(config), nil
}

// UpdateConfig updates an existing webhook configuration
func (s *WebhookConfigService) UpdateConfig(id string, req *models.UpdateWebhookConfigRequest) (*models.WebhookConfigurationResponse, error) {
	config, err := s.webhookRepo.GetConfigByID(id)
	if err != nil {
		return nil, errors.NewNotFoundError("Webhook configuration not found")
	}

	// Update fields if provided
	if req.WebhookURL != nil {
		config.WebhookURL = *req.WebhookURL
	}

	if req.SecretKey != nil {
		if len(*req.SecretKey) < 32 {
			return nil, errors.NewBadRequestError("Secret key must be at least 32 characters long")
		}
		config.SecretKey = *req.SecretKey
	}

	if req.Enabled != nil {
		config.Enabled = *req.Enabled
	}

	if req.RetryAttempts != nil {
		if *req.RetryAttempts < 0 || *req.RetryAttempts > 10 {
			return nil, errors.NewBadRequestError("Retry attempts must be between 0 and 10")
		}
		config.RetryAttempts = *req.RetryAttempts
	}

	if req.TimeoutSecs != nil {
		if *req.TimeoutSecs < 5 || *req.TimeoutSecs > 300 {
			return nil, errors.NewBadRequestError("Timeout must be between 5 and 300 seconds")
		}
		config.TimeoutSecs = *req.TimeoutSecs
	}

	if err := s.webhookRepo.UpdateConfig(config); err != nil {
		return nil, err
	}

	return s.convertToResponse(config), nil
}

// DeleteConfig deletes a webhook configuration
func (s *WebhookConfigService) DeleteConfig(id string) error {
	// Check if configuration exists
	_, err := s.webhookRepo.GetConfigByID(id)
	if err != nil {
		return errors.NewNotFoundError("Webhook configuration not found")
	}

	return s.webhookRepo.DeleteConfig(id)
}

// GetAllConfigs retrieves all webhook configurations with pagination
func (s *WebhookConfigService) GetAllConfigs(limit, offset int) ([]models.WebhookConfigurationResponse, error) {
	configs, err := s.webhookRepo.GetAllConfigs(limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]models.WebhookConfigurationResponse, len(configs))
	for i, config := range configs {
		responses[i] = *s.convertToResponse(&config)
	}

	return responses, nil
}

// ToggleConfig enables or disables a webhook configuration
func (s *WebhookConfigService) ToggleConfig(id string) (*models.WebhookConfigurationResponse, error) {
	config, err := s.webhookRepo.GetConfigByID(id)
	if err != nil {
		return nil, errors.NewNotFoundError("Webhook configuration not found")
	}

	config.Enabled = !config.Enabled

	if err := s.webhookRepo.UpdateConfig(config); err != nil {
		return nil, err
	}

	return s.convertToResponse(config), nil
}

// GetConfigHistory retrieves delivery history for a configuration
func (s *WebhookConfigService) GetConfigHistory(configID string, limit, offset int) ([]models.WebhookHistoryResponse, error) {
	// Verify configuration exists
	_, err := s.webhookRepo.GetConfigByID(configID)
	if err != nil {
		return nil, errors.NewNotFoundError("Webhook configuration not found")
	}

	return s.historyService.GetConfigHistory(configID, limit, offset)
}

// TestConfig sends a test webhook to verify configuration
func (s *WebhookConfigService) TestConfig(configID string) error {
	config, err := s.webhookRepo.GetConfigByID(configID)
	if err != nil {
		return errors.NewNotFoundError("Webhook configuration not found")
	}

	if !config.Enabled {
		return errors.NewBadRequestError("Webhook configuration is disabled")
	}

	// Create a test payload
	testPayload := map[string]interface{}{
		"event_type": "test_connection",
		"fpo_id":     config.FPOID,
		"timestamp":  time.Now().Unix(),
		"message":    "This is a test webhook from Kisanlink ERP",
	}

	// Log the test attempt
	_, err = s.historyService.LogWebhookDelivery(config.ID, "test_connection", "test_"+config.ID, testPayload)
	if err != nil {
		return err
	}

	// TODO: Implement actual HTTP delivery in the outbound webhook service
	return nil
}

// RegenerateSecret generates a new secret key for a configuration
func (s *WebhookConfigService) RegenerateSecret(configID string) (*models.WebhookConfigurationResponse, error) {
	config, err := s.webhookRepo.GetConfigByID(configID)
	if err != nil {
		return nil, errors.NewNotFoundError("Webhook configuration not found")
	}

	// Generate new secret
	config.SecretKey = s.generateSecureSecret()

	if err := s.webhookRepo.UpdateConfig(config); err != nil {
		return nil, err
	}

	return s.convertToResponse(config), nil
}

// GetDeliveryStats retrieves delivery statistics for a configuration
func (s *WebhookConfigService) GetDeliveryStats(configID string, hours int) (map[string]interface{}, error) {
	// Verify configuration exists
	_, err := s.webhookRepo.GetConfigByID(configID)
	if err != nil {
		return nil, errors.NewNotFoundError("Webhook configuration not found")
	}

	stats := make(map[string]interface{})

	// Get success rate
	successRate, err := s.webhookRepo.GetDeliverySuccessRate(configID, hours)
	if err != nil {
		return nil, err
	}

	stats["success_rate"] = successRate
	stats["period_hours"] = hours

	return stats, nil
}

// convertToResponse converts a webhook configuration model to response DTO
func (s *WebhookConfigService) convertToResponse(config *models.WebhookConfiguration) *models.WebhookConfigurationResponse {
	return &models.WebhookConfigurationResponse{
		ID:            config.ID,
		FPOID:         config.FPOID,
		WebhookURL:    config.WebhookURL,
		Enabled:       config.Enabled,
		RetryAttempts: config.RetryAttempts,
		TimeoutSecs:   config.TimeoutSecs,
		CreatedAt:     config.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     config.UpdatedAt.Format(time.RFC3339),
		// Note: We don't expose the secret key in responses for security
	}
}

// generateSecureSecret generates a cryptographically secure random secret
func (s *WebhookConfigService) generateSecureSecret() string {
	// Generate 32 random bytes (256 bits)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based secret if crypto/rand fails
		return hex.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano) + "_kisanlink_webhook_secret"))
	}
	return hex.EncodeToString(bytes)
}