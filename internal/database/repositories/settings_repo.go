package repositories

import (
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"

	"gorm.io/gorm"
)

// SettingsRepository handles settings data access
type SettingsRepository struct {
	db *gorm.DB
}

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// Create creates a new setting (or updates if exists - upsert behavior)
func (r *SettingsRepository) Create(setting *models.Setting) error {
	if err := r.db.Create(setting).Error; err != nil {
		return errors.NewInternalServerError("Failed to create setting")
	}
	return nil
}

// Upsert creates or updates a setting by key
func (r *SettingsRepository) Upsert(setting *models.Setting) error {
	if err := r.db.Save(setting).Error; err != nil {
		return errors.NewInternalServerError("Failed to save setting")
	}
	return nil
}

// GetByKey retrieves a setting by key
func (r *SettingsRepository) GetByKey(key string) (*models.Setting, error) {
	var setting models.Setting
	if err := r.db.Where("key = ?", key).First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Setting")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve setting")
	}
	return &setting, nil
}

// GetByKeyOrNil retrieves a setting by key, returns nil if not found (no error)
func (r *SettingsRepository) GetByKeyOrNil(key string) (*models.Setting, error) {
	var setting models.Setting
	if err := r.db.Where("key = ?", key).First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.NewInternalServerError("Failed to retrieve setting")
	}
	return &setting, nil
}

// GetAll retrieves all settings
func (r *SettingsRepository) GetAll() ([]models.Setting, error) {
	var settings []models.Setting
	if err := r.db.Order("key ASC").Find(&settings).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve settings")
	}
	return settings, nil
}

// GetHeaderFields retrieves all settings that are header fields, ordered by display_order
func (r *SettingsRepository) GetHeaderFields() ([]models.Setting, error) {
	var settings []models.Setting
	if err := r.db.Where("is_header_field = ?", true).
		Order("display_order ASC").
		Find(&settings).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve header fields")
	}
	return settings, nil
}

// GetByKeys retrieves multiple settings by their keys
func (r *SettingsRepository) GetByKeys(keys []string) ([]models.Setting, error) {
	var settings []models.Setting
	if err := r.db.Where("key IN ?", keys).Find(&settings).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve settings by keys")
	}
	return settings, nil
}

// Update updates a setting
func (r *SettingsRepository) Update(setting *models.Setting) error {
	if err := r.db.Save(setting).Error; err != nil {
		return errors.NewInternalServerError("Failed to update setting")
	}
	return nil
}

// Delete deletes a setting by key
func (r *SettingsRepository) Delete(key string) error {
	result := r.db.Where("key = ?", key).Delete(&models.Setting{})
	if result.Error != nil {
		return errors.NewInternalServerError("Failed to delete setting")
	}
	if result.RowsAffected == 0 {
		return errors.NewNotFoundError("Setting")
	}
	return nil
}

// Exists checks if a setting exists by key
func (r *SettingsRepository) Exists(key string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Setting{}).Where("key = ?", key).Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check setting existence")
	}
	return count > 0, nil
}

// CheckRequiredSettings checks if all required keys exist and returns missing ones
func (r *SettingsRepository) CheckRequiredSettings(requiredKeys []string) ([]string, error) {
	settings, err := r.GetByKeys(requiredKeys)
	if err != nil {
		return nil, err
	}

	existingKeys := make(map[string]bool)
	for _, s := range settings {
		existingKeys[s.Key] = true
	}

	var missing []string
	for _, key := range requiredKeys {
		if !existingKeys[key] {
			missing = append(missing, key)
		}
	}
	return missing, nil
}

// GetSettingsMap returns all settings as a map[key]value for easy lookup
func (r *SettingsRepository) GetSettingsMap() (map[string]string, error) {
	settings, err := r.GetAll()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}
