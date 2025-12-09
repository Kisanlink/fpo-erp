package services

import (
	"context"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"
)

// SettingsService handles settings business logic
type SettingsService struct {
	settingsRepo *repositories.SettingsRepository
	logger       interfaces.Logger
}

// NewSettingsService creates a new settings service
func NewSettingsService(settingsRepo *repositories.SettingsRepository, logger interfaces.Logger) *SettingsService {
	return &SettingsService{
		settingsRepo: settingsRepo,
		logger:       logger,
	}
}

// GetSetting retrieves a setting by key
func (s *SettingsService) GetSetting(ctx context.Context, key string) (*models.SettingResponse, error) {
	s.logger.Info("Getting setting", "key", key)

	setting, err := s.settingsRepo.GetByKey(key)
	if err != nil {
		s.logger.Error("Failed to get setting", "key", key, "error", err)
		return nil, err
	}

	return setting.ToResponse(), nil
}

// GetAllSettings retrieves all settings
func (s *SettingsService) GetAllSettings(ctx context.Context) ([]models.SettingResponse, error) {
	s.logger.Info("Getting all settings")

	settings, err := s.settingsRepo.GetAll()
	if err != nil {
		s.logger.Error("Failed to get all settings", "error", err)
		return nil, err
	}

	responses := make([]models.SettingResponse, 0, len(settings))
	for _, setting := range settings {
		responses = append(responses, *setting.ToResponse())
	}

	return responses, nil
}

// UpsertSetting creates or updates a setting
func (s *SettingsService) UpsertSetting(ctx context.Context, key string, request *models.CreateSettingRequest) (*models.SettingResponse, error) {
	s.logger.Info("Upserting setting", "key", key)

	// Validate key is not empty
	if key == "" {
		return nil, errors.NewBadRequestError("Setting key cannot be empty")
	}

	// Check if setting exists
	existing, err := s.settingsRepo.GetByKeyOrNil(key)
	if err != nil {
		s.logger.Error("Failed to check existing setting", "key", key, "error", err)
		return nil, err
	}

	var setting *models.Setting

	if existing != nil {
		// Update existing setting
		setting = existing
		setting.Value = request.Value

		if request.DisplayLabel != nil {
			setting.DisplayLabel = request.DisplayLabel
		}
		if request.DisplayOrder != nil {
			setting.DisplayOrder = *request.DisplayOrder
		}
		if request.IsHeaderField != nil {
			setting.IsHeaderField = *request.IsHeaderField
		}

		if err := s.settingsRepo.Update(setting); err != nil {
			s.logger.Error("Failed to update setting", "key", key, "error", err)
			return nil, err
		}

		s.logger.Info("Setting updated successfully", "key", key)
	} else {
		// Create new setting
		setting = models.NewSetting(key, request.Value)

		if request.DisplayLabel != nil {
			setting.DisplayLabel = request.DisplayLabel
		}
		if request.DisplayOrder != nil {
			setting.DisplayOrder = *request.DisplayOrder
		}
		if request.IsHeaderField != nil {
			setting.IsHeaderField = *request.IsHeaderField
		}

		if err := s.settingsRepo.Create(setting); err != nil {
			s.logger.Error("Failed to create setting", "key", key, "error", err)
			return nil, err
		}

		s.logger.Info("Setting created successfully", "key", key)
	}

	// Refetch to get updated timestamps
	updated, err := s.settingsRepo.GetByKey(key)
	if err != nil {
		s.logger.Error("Failed to refetch setting", "key", key, "error", err)
		return nil, err
	}

	return updated.ToResponse(), nil
}

// DeleteSetting deletes a setting by key
func (s *SettingsService) DeleteSetting(ctx context.Context, key string) error {
	s.logger.Info("Deleting setting", "key", key)

	if err := s.settingsRepo.Delete(key); err != nil {
		s.logger.Error("Failed to delete setting", "key", key, "error", err)
		return err
	}

	s.logger.Info("Setting deleted successfully", "key", key)
	return nil
}

// GetHeaderFields retrieves all settings configured as header fields, ordered by display_order
func (s *SettingsService) GetHeaderFields(ctx context.Context) ([]models.HeaderFieldResponse, error) {
	s.logger.Info("Getting header fields")

	settings, err := s.settingsRepo.GetHeaderFields()
	if err != nil {
		s.logger.Error("Failed to get header fields", "error", err)
		return nil, err
	}

	responses := make([]models.HeaderFieldResponse, 0)
	for _, setting := range settings {
		// Only include if it has a display label
		if headerField := setting.ToHeaderFieldResponse(); headerField != nil {
			responses = append(responses, *headerField)
		}
	}

	return responses, nil
}

// CheckInvoiceRequirements checks if all required settings for invoice generation exist
// Returns (ready bool, missing []string, error)
func (s *SettingsService) CheckInvoiceRequirements(ctx context.Context) (bool, []string, error) {
	s.logger.Info("Checking invoice requirements")

	requiredKeys := models.RequiredSettingsForInvoice()
	missing, err := s.settingsRepo.CheckRequiredSettings(requiredKeys)
	if err != nil {
		s.logger.Error("Failed to check invoice requirements", "error", err)
		return false, nil, err
	}

	if len(missing) > 0 {
		s.logger.Warn("Missing required settings for invoice", "missing", missing)
		return false, missing, nil
	}

	s.logger.Info("All invoice requirements satisfied")
	return true, nil, nil
}

// GetSettingsMap returns all settings as a key-value map
func (s *SettingsService) GetSettingsMap(ctx context.Context) (map[string]string, error) {
	s.logger.Info("Getting settings map")

	settingsMap, err := s.settingsRepo.GetSettingsMap()
	if err != nil {
		s.logger.Error("Failed to get settings map", "error", err)
		return nil, err
	}

	return settingsMap, nil
}
