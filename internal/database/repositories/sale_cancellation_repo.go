package repositories

import (
	"kisanlink-erp/internal/database/models"

	"gorm.io/gorm"
)

type SaleCancellationRepository struct {
	db *gorm.DB
}

func NewSaleCancellationRepository(db *gorm.DB) *SaleCancellationRepository {
	return &SaleCancellationRepository{db: db}
}

// CreateCancellationWithTx creates a sale cancellation within a transaction
func (r *SaleCancellationRepository) CreateCancellationWithTx(tx *gorm.DB, cancellation *models.SaleCancellation) error {
	return tx.Create(cancellation).Error
}

// CreateCancellationItemWithTx creates a sale cancellation item within a transaction
func (r *SaleCancellationRepository) CreateCancellationItemWithTx(tx *gorm.DB, item *models.SaleCancellationItem) error {
	return tx.Create(item).Error
}

// UpdateCancellationWithTx updates a sale cancellation within a transaction
func (r *SaleCancellationRepository) UpdateCancellationWithTx(tx *gorm.DB, cancellation *models.SaleCancellation) error {
	return tx.Save(cancellation).Error
}

// GetCancellationByID retrieves a sale cancellation by ID
func (r *SaleCancellationRepository) GetCancellationByID(id string) (*models.SaleCancellation, error) {
	var cancellation models.SaleCancellation
	err := r.db.Preload("Items").First(&cancellation, "id = ?", id).Error
	return &cancellation, err
}

// GetCancellationsBySaleID retrieves all cancellations for a specific sale
func (r *SaleCancellationRepository) GetCancellationsBySaleID(saleID string) ([]models.SaleCancellation, error) {
	var cancellations []models.SaleCancellation
	err := r.db.Preload("Items").Where("sale_id = ?", saleID).Find(&cancellations).Error
	return cancellations, err
}

// CreateCancellation creates a sale cancellation
func (r *SaleCancellationRepository) CreateCancellation(cancellation *models.SaleCancellation) error {
	return r.db.Create(cancellation).Error
}

// CreateCancellationItem creates a sale cancellation item
func (r *SaleCancellationRepository) CreateCancellationItem(item *models.SaleCancellationItem) error {
	return r.db.Create(item).Error
}

// UpdateCancellation updates a sale cancellation
func (r *SaleCancellationRepository) UpdateCancellation(cancellation *models.SaleCancellation) error {
	return r.db.Save(cancellation).Error
}
