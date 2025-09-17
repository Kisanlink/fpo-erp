package repositories

import (
	"time"

	"kisanlink-erp/internal/database/models"

	"gorm.io/gorm"
)

type ReturnsRepository struct {
	db *gorm.DB
}

func NewReturnsRepository(db *gorm.DB) *ReturnsRepository {
	return &ReturnsRepository{db: db}
}

// Return operations
func (r *ReturnsRepository) CreateReturn(ret *models.Return) error {
	return r.db.Create(ret).Error
}

func (r *ReturnsRepository) GetReturnByID(id string) (*models.Return, error) {
	var ret models.Return
	err := r.db.Preload("Items").First(&ret, "id = ?", id).Error
	return &ret, err
}

func (r *ReturnsRepository) GetAllReturns(limit, offset int) ([]models.Return, error) {
	var returns []models.Return
	err := r.db.Preload("Items").Limit(limit).Offset(offset).Find(&returns).Error
	return returns, err
}

func (r *ReturnsRepository) UpdateReturn(ret *models.Return) error {
	return r.db.Save(ret).Error
}

func (r *ReturnsRepository) DeleteReturn(id string) error {
	return r.db.Delete(&models.Return{}, "id = ?", id).Error
}

func (r *ReturnsRepository) GetReturnsByCustomer(customerID string) ([]models.Return, error) {
	var returns []models.Return
	err := r.db.Preload("Items").Where("customer_id = ?", customerID).Find(&returns).Error
	return returns, err
}

func (r *ReturnsRepository) GetReturnsBySaleID(saleID string) ([]models.Return, error) {
	var returns []models.Return
	err := r.db.Preload("Items").Where("sale_id = ?", saleID).Find(&returns).Error
	return returns, err
}

func (r *ReturnsRepository) GetReturnsByDateRange(startDate, endDate time.Time) ([]models.Return, error) {
	var returns []models.Return
	err := r.db.Preload("Items").Where("return_date BETWEEN ? AND ?", startDate, endDate).Find(&returns).Error
	return returns, err
}

func (r *ReturnsRepository) GetReturnsByStatus(status string) ([]models.Return, error) {
	var returns []models.Return
	err := r.db.Preload("Items").Where("status = ?", status).Find(&returns).Error
	return returns, err
}

// ReturnItem operations
func (r *ReturnsRepository) CreateReturnItem(item *models.ReturnItem) error {
	return r.db.Create(item).Error
}

func (r *ReturnsRepository) GetReturnItemsByReturnID(returnID string) ([]models.ReturnItem, error) {
	var items []models.ReturnItem
	err := r.db.Where("return_id = ?", returnID).Find(&items).Error
	return items, err
}

func (r *ReturnsRepository) UpdateReturnItem(item *models.ReturnItem) error {
	return r.db.Save(item).Error
}

func (r *ReturnsRepository) DeleteReturnItem(id string) error {
	return r.db.Delete(&models.ReturnItem{}, "id = ?", id).Error
}

// ReturnSummary operations
func (r *ReturnsRepository) CreateReturnSummary(summary *models.ReturnSummary) error {
	return r.db.Create(summary).Error
}

func (r *ReturnsRepository) GetReturnSummaryByReturnID(returnID string) (*models.ReturnSummary, error) {
	var summary models.ReturnSummary
	err := r.db.Where("return_id = ?", returnID).First(&summary).Error
	return &summary, err
}

func (r *ReturnsRepository) UpdateReturnSummary(summary *models.ReturnSummary) error {
	return r.db.Save(summary).Error
}

func (r *ReturnsRepository) DeleteReturnSummary(id string) error {
	return r.db.Delete(&models.ReturnSummary{}, "id = ?", id).Error
}

// Analytics
func (r *ReturnsRepository) GetTotalReturnsAmount(startDate, endDate time.Time) (float64, error) {
	var total float64
	err := r.db.Model(&models.Return{}).Where("return_date BETWEEN ? AND ?", startDate, endDate).Select("COALESCE(SUM(total_amount), 0)").Scan(&total).Error
	return total, err
}

func (r *ReturnsRepository) GetReturnRateByProduct(productID string, startDate, endDate time.Time) (float64, error) {
	var returnRate float64
	err := r.db.Raw(`
		SELECT 
			COALESCE(
				(SUM(return_items.quantity) * 100.0 / NULLIF(SUM(sale_items.quantity), 0)
			, 0) as return_rate
		FROM return_items 
		JOIN returns ON return_items.return_id = returns.id
		LEFT JOIN sale_items ON return_items.sale_item_id = sale_items.id
		WHERE return_items.product_id = ? 
		AND returns.return_date BETWEEN ? AND ?
	`, productID, startDate, endDate).Scan(&returnRate).Error
	return returnRate, err
}

func (r *ReturnsRepository) GetMostReturnedProducts(limit int) ([]struct {
	ProductID     string  `json:"product_id"`
	ProductName   string  `json:"product_name"`
	TotalReturned int     `json:"total_returned"`
	ReturnAmount  float64 `json:"return_amount"`
}, error) {
	var results []struct {
		ProductID     string  `json:"product_id"`
		ProductName   string  `json:"product_name"`
		TotalReturned int     `json:"total_returned"`
		ReturnAmount  float64 `json:"return_amount"`
	}

	err := r.db.Table("return_items").
		Select("return_items.product_id, sku.name as product_name, SUM(return_items.quantity) as total_returned, SUM(return_items.total_price) as return_amount").
		Joins("JOIN sku ON return_items.product_id = sku.id").
		Group("return_items.product_id, sku.name").
		Order("total_returned DESC").
		Limit(limit).
		Find(&results).Error

	return results, err
}
