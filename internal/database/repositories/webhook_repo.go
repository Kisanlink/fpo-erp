package repositories

import (
	"kisanlink-erp/internal/database/models"
	"time"

	"gorm.io/gorm"
)

// WebhookRepository handles webhook-related database operations
type WebhookRepository struct {
	db *gorm.DB
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *gorm.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// Configuration operations
func (r *WebhookRepository) CreateConfig(config *models.WebhookConfiguration) error {
	return r.db.Create(config).Error
}

func (r *WebhookRepository) GetConfigByID(id string) (*models.WebhookConfiguration, error) {
	var config models.WebhookConfiguration
	err := r.db.Where("id = ?", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *WebhookRepository) GetConfigByFPO(fpoID string) (*models.WebhookConfiguration, error) {
	var config models.WebhookConfiguration
	err := r.db.Where("fpo_id = ?", fpoID).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *WebhookRepository) UpdateConfig(config *models.WebhookConfiguration) error {
	return r.db.Save(config).Error
}

func (r *WebhookRepository) DeleteConfig(id string) error {
	return r.db.Delete(&models.WebhookConfiguration{}, "id = ?", id).Error
}

func (r *WebhookRepository) GetAllConfigs(limit, offset int) ([]models.WebhookConfiguration, error) {
	var configs []models.WebhookConfiguration
	err := r.db.Limit(limit).Offset(offset).Find(&configs).Error
	return configs, err
}

// Event operations
func (r *WebhookRepository) CreateEvent(event *models.WebhookEvent) error {
	return r.db.Create(event).Error
}

func (r *WebhookRepository) GetEventByID(id string) (*models.WebhookEvent, error) {
	var event models.WebhookEvent
	err := r.db.Where("id = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *WebhookRepository) GetEventByEventID(eventID string) (*models.WebhookEvent, error) {
	var event models.WebhookEvent
	err := r.db.Where("event_id = ?", eventID).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *WebhookRepository) UpdateEvent(event *models.WebhookEvent) error {
	return r.db.Save(event).Error
}

func (r *WebhookRepository) GetEventsByFPO(fpoID string, limit, offset int) ([]models.WebhookEvent, error) {
	var events []models.WebhookEvent
	err := r.db.Where("fpo_id = ?", fpoID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&events).Error
	return events, err
}

func (r *WebhookRepository) GetEventsByStatus(status string, limit, offset int) ([]models.WebhookEvent, error) {
	var events []models.WebhookEvent
	err := r.db.Where("processed_status = ?", status).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&events).Error
	return events, err
}

// History operations
func (r *WebhookRepository) CreateHistory(history *models.WebhookHistory) error {
	return r.db.Create(history).Error
}

func (r *WebhookRepository) GetHistoryByID(id string) (*models.WebhookHistory, error) {
	var history models.WebhookHistory
	err := r.db.Preload("Config").Where("id = ?", id).First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

func (r *WebhookRepository) UpdateHistory(history *models.WebhookHistory) error {
	return r.db.Save(history).Error
}

func (r *WebhookRepository) GetHistoryByConfig(configID string, limit, offset int) ([]models.WebhookHistory, error) {
	var history []models.WebhookHistory
	err := r.db.Where("config_id = ?", configID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&history).Error
	return history, err
}

func (r *WebhookRepository) GetHistoryByEventID(eventID string) ([]models.WebhookHistory, error) {
	var history []models.WebhookHistory
	err := r.db.Where("event_id = ?", eventID).
		Order("created_at DESC").
		Find(&history).Error
	return history, err
}

func (r *WebhookRepository) GetFailedDeliveries(limit, offset int) ([]models.WebhookHistory, error) {
	var history []models.WebhookHistory
	err := r.db.Where("status = ? OR status = ?", "failed", "error").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&history).Error
	return history, err
}

// Queue operations
func (r *WebhookRepository) CreateQueueItem(item *models.WebhookQueue) error {
	return r.db.Create(item).Error
}

func (r *WebhookRepository) GetQueueByID(id string) (*models.WebhookQueue, error) {
	var item models.WebhookQueue
	err := r.db.Preload("Config").Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *WebhookRepository) UpdateQueueItem(item *models.WebhookQueue) error {
	return r.db.Save(item).Error
}

func (r *WebhookRepository) DeleteQueueItem(id string) error {
	return r.db.Delete(&models.WebhookQueue{}, "id = ?", id).Error
}

func (r *WebhookRepository) GetPendingQueueItems(limit int) ([]models.WebhookQueue, error) {
	var items []models.WebhookQueue
	err := r.db.Preload("Config").
		Where("status = ? AND next_retry_at <= ? AND attempt_count < max_retries", "pending", time.Now()).
		Order("next_retry_at ASC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *WebhookRepository) GetQueueByConfig(configID string, limit, offset int) ([]models.WebhookQueue, error) {
	var items []models.WebhookQueue
	err := r.db.Where("config_id = ?", configID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&items).Error
	return items, err
}

func (r *WebhookRepository) GetQueueStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count by status
	var pendingCount, failedCount, completedCount int64

	r.db.Model(&models.WebhookQueue{}).Where("status = ?", "pending").Count(&pendingCount)
	r.db.Model(&models.WebhookQueue{}).Where("status = ?", "failed").Count(&failedCount)
	r.db.Model(&models.WebhookQueue{}).Where("status = ?", "completed").Count(&completedCount)

	stats["pending"] = pendingCount
	stats["failed"] = failedCount
	stats["completed"] = completedCount

	return stats, nil
}

// Utility functions
func (r *WebhookRepository) CleanupOldHistory(olderThan time.Time) error {
	return r.db.Where("created_at < ?", olderThan).Delete(&models.WebhookHistory{}).Error
}

func (r *WebhookRepository) CleanupProcessedEvents(olderThan time.Time) error {
	return r.db.Where("processed_status = ? AND processed_at < ?", "completed", olderThan).
		Delete(&models.WebhookEvent{}).Error
}

func (r *WebhookRepository) CleanupCompletedQueue(olderThan time.Time) error {
	return r.db.Where("status = ? AND updated_at < ?", "completed", olderThan).
		Delete(&models.WebhookQueue{}).Error
}

// Advanced queries for monitoring
func (r *WebhookRepository) GetEventStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	var pendingCount, processedCount, failedCount int64

	r.db.Model(&models.WebhookEvent{}).Where("processed_status = ?", "pending").Count(&pendingCount)
	r.db.Model(&models.WebhookEvent{}).Where("processed_status = ?", "completed").Count(&processedCount)
	r.db.Model(&models.WebhookEvent{}).Where("processed_status = ?", "failed").Count(&failedCount)

	stats["pending"] = pendingCount
	stats["completed"] = processedCount
	stats["failed"] = failedCount

	return stats, nil
}

func (r *WebhookRepository) GetRecentEventsByType(eventType string, hours int, limit int) ([]models.WebhookEvent, error) {
	var events []models.WebhookEvent
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	err := r.db.Where("event_type = ? AND created_at >= ?", eventType, since).
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error

	return events, err
}

func (r *WebhookRepository) GetDeliverySuccessRate(configID string, hours int) (float64, error) {
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	var totalCount, successCount int64

	r.db.Model(&models.WebhookHistory{}).
		Where("config_id = ? AND created_at >= ?", configID, since).
		Count(&totalCount)

	r.db.Model(&models.WebhookHistory{}).
		Where("config_id = ? AND created_at >= ? AND status = ?", configID, since, "success").
		Count(&successCount)

	if totalCount == 0 {
		return 0.0, nil
	}

	return float64(successCount) / float64(totalCount), nil
}