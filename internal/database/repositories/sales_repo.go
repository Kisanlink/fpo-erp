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

// Sale operations
func (r *SalesRepository) CreateSale(sale *models.Sale) error {
	return r.db.Create(sale).Error
}

func (r *SalesRepository) GetSaleByID(id string) (*models.Sale, error) {
	var sale models.Sale
	err := r.db.Preload("Items").First(&sale, "id = ?", id).Error
	return &sale, err
}

func (r *SalesRepository) GetAllSales(limit, offset int) ([]models.Sale, error) {
	var sales []models.Sale
	err := r.db.Preload("Items").Limit(limit).Offset(offset).Find(&sales).Error
	return sales, err
}

func (r *SalesRepository) UpdateSale(sale *models.Sale) error {
	return r.db.Save(sale).Error
}

func (r *SalesRepository) DeleteSale(id string) error {
	return r.db.Delete(&models.Sale{}, "id = ?", id).Error
}

func (r *SalesRepository) GetSalesByDateRange(startDate, endDate time.Time) ([]models.Sale, error) {
	var sales []models.Sale
	err := r.db.Preload("Items").Where("sale_date BETWEEN ? AND ?", startDate, endDate).Find(&sales).Error
	return sales, err
}

func (r *SalesRepository) GetSalesByStatus(status string) ([]models.Sale, error) {
	var sales []models.Sale
	err := r.db.Preload("Items").Where("status = ?", status).Find(&sales).Error
	return sales, err
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
