package interfaces

import (
	"context"
	"kisanlink-erp/internal/database/models"
)

// SettingsServiceInterface defines the contract for settings operations
type SettingsServiceInterface interface {
	// CRUD operations
	GetSetting(ctx context.Context, key string) (*models.SettingResponse, error)
	GetAllSettings(ctx context.Context) ([]models.SettingResponse, error)
	UpsertSetting(ctx context.Context, key string, request *models.CreateSettingRequest) (*models.SettingResponse, error)
	DeleteSetting(ctx context.Context, key string) error

	// Header field operations (for invoice)
	GetHeaderFields(ctx context.Context) ([]models.HeaderFieldResponse, error)

	// Invoice validation
	CheckInvoiceRequirements(ctx context.Context) (bool, []string, error)
	GetSettingsMap(ctx context.Context) (map[string]string, error)
}
