package repositories

import (
	"log"
	"time"

	"kisanlink-erp/internal/database/models"

	"gorm.io/gorm"
)

type SalesRepository struct {
	db *gorm.DB
}

func NewSalesRepository(db *gorm.DB) *SalesRepository {
	return &SalesRepository{db: db}
}

// WithTransaction executes a function within a database transaction
func (r *SalesRepository) WithTransaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// CreateSaleWithTx creates a sale within a transaction
func (r *SalesRepository) CreateSaleWithTx(tx *gorm.DB, sale *models.Sale) error {
	return tx.Create(sale).Error
}

// CreateSaleItemWithTx creates a sale item within a transaction
func (r *SalesRepository) CreateSaleItemWithTx(tx *gorm.DB, item *models.SaleItem) error {
	return tx.Create(item).Error
}

// UpdateSaleWithTx updates a sale within a transaction
func (r *SalesRepository) UpdateSaleWithTx(tx *gorm.DB, sale *models.Sale) error {
	return tx.Save(sale).Error
}

// GetSaleForUpdateWithTx gets a sale with pessimistic lock within a transaction
func (r *SalesRepository) GetSaleForUpdateWithTx(tx *gorm.DB, id string) (*models.Sale, error) {
	var sale models.Sale
	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Preload("Items").
		Where("id = ?", id).
		First(&sale).Error
	return &sale, err
}

// Sale operations
func (r *SalesRepository) CreateSale(sale *models.Sale) error {
	return r.db.Create(sale).Error
}

func (r *SalesRepository) GetSaleByID(id string) (*models.Sale, error) {
	var sale models.Sale
	err := r.db.Preload("Items").First(&sale, "id = ?", id).Error
	return &sale, err
}

func (r *SalesRepository) GetAllSales(limit, offset int) ([]models.Sale, int64, error) {
	var sales []models.Sale
	var total int64

	// Get total count
	if err := r.db.Model(&models.Sale{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated records
	err := r.db.Preload("Items").Order("created_at DESC").Limit(limit).Offset(offset).Find(&sales).Error
	return sales, total, err
}

func (r *SalesRepository) UpdateSale(sale *models.Sale) error {
	return r.db.Save(sale).Error
}

func (r *SalesRepository) DeleteSale(id string) error {
	return r.db.Delete(&models.Sale{}, "id = ?", id).Error
}

func (r *SalesRepository) GetSalesByDateRange(startDate, endDate time.Time, limit, offset int) ([]models.Sale, int64, error) {
	var sales []models.Sale
	var total int64

	query := r.db.Model(&models.Sale{}).Where("sale_date BETWEEN ? AND ?", startDate, endDate)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated records
	err := r.db.Preload("Items").Where("sale_date BETWEEN ? AND ?", startDate, endDate).
		Order("sale_date DESC").Limit(limit).Offset(offset).Find(&sales).Error
	return sales, total, err
}

func (r *SalesRepository) GetSalesByStatus(status string, limit, offset int) ([]models.Sale, int64, error) {
	var sales []models.Sale
	var total int64

	query := r.db.Model(&models.Sale{}).Where("status = ?", status)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated records
	err := r.db.Preload("Items").Where("status = ?", status).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&sales).Error
	return sales, total, err
}

// SaleItem operations
func (r *SalesRepository) CreateSaleItem(item *models.SaleItem) error {
	if err := r.db.Create(item).Error; err != nil {
		log.Printf("[ERROR] Database error creating sale item: %v", err)
		return err
	}
	return nil
}

func (r *SalesRepository) GetSaleItemsBySaleID(saleID string) ([]models.SaleItem, error) {
	var items []models.SaleItem
	err := r.db.Where("sale_id = ?", saleID).Find(&items).Error
	return items, err
}

func (r *SalesRepository) UpdateSaleItem(item *models.SaleItem) error {
	return r.db.Save(item).Error
}

func (r *SalesRepository) DeleteSaleItem(id string) error {
	return r.db.Delete(&models.SaleItem{}, "id = ?", id).Error
}

// SaleSummary operations
func (r *SalesRepository) CreateSaleSummary(summary *models.SaleSummary) error {
	return r.db.Create(summary).Error
}

func (r *SalesRepository) GetSaleSummaryBySaleID(saleID string) (*models.SaleSummary, error) {
	var summary models.SaleSummary
	err := r.db.Where("sale_id = ?", saleID).First(&summary).Error
	return &summary, err
}

func (r *SalesRepository) UpdateSaleSummary(summary *models.SaleSummary) error {
	return r.db.Save(summary).Error
}

func (r *SalesRepository) DeleteSaleSummary(id string) error {
	return r.db.Delete(&models.SaleSummary{}, "id = ?", id).Error
}

// Invoice Number operations

// GetLastInvoiceSequence returns the last sequence number used in invoice numbers
// Invoice format: MMYYNNNN (e.g., 12250001)
// Returns 0 if no invoices exist yet
func (r *SalesRepository) GetLastInvoiceSequence() (int, error) {
	var maxInvoiceNumber string

	// Get the last invoice number by ordering desc (it's a string but sequential)
	err := r.db.Model(&models.Sale{}).
		Where("invoice_number IS NOT NULL AND invoice_number != ''").
		Order("invoice_number DESC").
		Limit(1).
		Pluck("invoice_number", &maxInvoiceNumber).Error

	if err != nil {
		return 0, err
	}

	// If no invoice numbers found, return 0
	if maxInvoiceNumber == "" {
		return 0, nil
	}

	// Extract sequence from MMYYNNNN format (last 4 digits)
	if len(maxInvoiceNumber) < 4 {
		return 0, nil
	}

	// Parse the last 4 characters as the sequence number
	var sequence int
	_, err = time.Parse("0102", maxInvoiceNumber[:4]) // Just to validate format
	if err == nil && len(maxInvoiceNumber) >= 8 {
		// Extract last 4 digits as sequence
		sequenceStr := maxInvoiceNumber[4:]
		for i := 0; i < len(sequenceStr); i++ {
			sequence = sequence*10 + int(sequenceStr[i]-'0')
		}
	}

	return sequence, nil
}

// GetLastInvoiceSequenceWithTx returns the last sequence number within a transaction (for locking)
func (r *SalesRepository) GetLastInvoiceSequenceWithTx(tx *gorm.DB) (int, error) {
	var maxInvoiceNumber string

	// Get the last invoice number with lock
	err := tx.Model(&models.Sale{}).
		Where("invoice_number IS NOT NULL AND invoice_number != ''").
		Order("invoice_number DESC").
		Limit(1).
		Pluck("invoice_number", &maxInvoiceNumber).Error

	if err != nil {
		return 0, err
	}

	// If no invoice numbers found, return 0
	if maxInvoiceNumber == "" {
		return 0, nil
	}

	// Extract sequence from MMYYNNNN format (last 4 digits)
	if len(maxInvoiceNumber) < 4 {
		return 0, nil
	}

	// Parse the last 4 characters as the sequence number
	var sequence int
	if len(maxInvoiceNumber) >= 8 {
		// Extract last 4 digits as sequence
		sequenceStr := maxInvoiceNumber[4:]
		for i := 0; i < len(sequenceStr); i++ {
			sequence = sequence*10 + int(sequenceStr[i]-'0')
		}
	}

	return sequence, nil
}

// Analytics
func (r *SalesRepository) GetTotalSalesAmount(startDate, endDate time.Time) (float64, error) {
	var total float64
	err := r.db.Model(&models.Sale{}).Where("sale_date BETWEEN ? AND ?", startDate, endDate).Select("COALESCE(SUM(total_amount), 0)").Scan(&total).Error
	return total, err
}

func (r *SalesRepository) GetTopSellingProducts(limit int) ([]struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	TotalSold   int     `json:"total_sold"`
	TotalAmount float64 `json:"total_amount"`
}, error) {
	var results []struct {
		ProductID   string  `json:"product_id"`
		ProductName string  `json:"product_name"`
		TotalSold   int     `json:"total_sold"`
		TotalAmount float64 `json:"total_amount"`
	}

	err := r.db.Table("sale_items").
		Select("sale_items.product_id, products.name as product_name, SUM(sale_items.quantity) as total_sold, SUM(sale_items.total_price) as total_amount").
		Joins("JOIN products ON sale_items.product_id = products.id").
		Group("sale_items.product_id, products.name").
		Order("total_sold DESC").
		Limit(limit).
		Find(&results).Error

	return results, err
}
