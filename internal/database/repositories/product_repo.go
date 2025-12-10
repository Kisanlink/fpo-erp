package repositories

import (
	"fmt"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/errors"

	"gorm.io/gorm"
)

// ProductRepository handles product data access
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create creates a new product
func (r *ProductRepository) Create(product *models.Product) error {
	if err := r.db.Create(product).Error; err != nil {
		fmt.Printf("DEBUG: Database error in ProductRepository.Create: %v\n", err)
		return errors.NewInternalServerError("Failed to create product")
	}
	return nil
}

// GetByID retrieves a product by ID
func (r *ProductRepository) GetByID(id string) (*models.Product, error) {
	var product models.Product
	if err := r.db.Where("id = ?", id).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Product")
		}
		return nil, errors.NewInternalServerError("Failed to retrieve product")
	}
	return &product, nil
}

// GetAllPaginated retrieves all products with pagination
func (r *ProductRepository) GetAllPaginated(limit, offset int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	// Get total count
	if err := r.db.Model(&models.Product{}).Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to count products")
	}

	// Get paginated records
	if err := r.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve products")
	}

	return products, total, nil
}

// GetAll retrieves all products without pagination
func (r *ProductRepository) GetAll() ([]models.Product, error) {
	var products []models.Product
	if err := r.db.Order("created_at DESC").Find(&products).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve products")
	}
	return products, nil
}

// Update updates a product
func (r *ProductRepository) Update(product *models.Product) error {
	if err := r.db.Save(product).Error; err != nil {
		return errors.NewInternalServerError("Failed to update product")
	}
	return nil
}

// Delete deletes a product
func (r *ProductRepository) Delete(id string) error {
	if err := r.db.Where("id = ?", id).Delete(&models.Product{}).Error; err != nil {
		return errors.NewInternalServerError("Failed to delete product")
	}
	return nil
}

// Exists checks if a product exists by ID
func (r *ProductRepository) Exists(id string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Product{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, errors.NewInternalServerError("Failed to check product existence")
	}
	return count > 0, nil
}

// GetByName retrieves products by name (for search functionality)
func (r *ProductRepository) GetByName(name string) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.Where("name ILIKE ?", "%"+name+"%").Find(&products).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to search products")
	}
	return products, nil
}

// FindByExternalID finds a product by external_id (for webhook integration)
func (r *ProductRepository) FindByExternalID(externalID string) (*models.Product, error) {
	var product models.Product
	if err := r.db.Where("external_id = ?", externalID).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found, but not an error for smart matching
		}
		return nil, errors.NewInternalServerError("Failed to find product by external_id")
	}
	return &product, nil
}

// GetByCategory retrieves all products in a category by category ID
func (r *ProductRepository) GetByCategory(categoryID string) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.Where("category_id = ?", categoryID).Find(&products).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve products by category")
	}
	return products, nil
}

// GetBySubcategory retrieves all products in a subcategory by subcategory ID
func (r *ProductRepository) GetBySubcategory(subcategoryID string) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.Where("subcategory_id = ?", subcategoryID).Find(&products).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve products by subcategory")
	}
	return products, nil
}

// GetByCategoryAndSubcategory retrieves all products in a specific category and subcategory by IDs
func (r *ProductRepository) GetByCategoryAndSubcategory(categoryID string, subcategoryID *string) ([]models.Product, error) {
	var products []models.Product
	query := r.db.Where("category_id = ?", categoryID)
	if subcategoryID != nil {
		query = query.Where("subcategory_id = ?", *subcategoryID)
	}
	if err := query.Find(&products).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve products by category and subcategory")
	}
	return products, nil
}

// GetWithFilters retrieves products with optional category and subcategory ID filters
func (r *ProductRepository) GetWithFilters(categoryID *string, subcategoryID *string) ([]models.Product, error) {
	var products []models.Product
	query := r.db.Model(&models.Product{})

	if categoryID != nil && *categoryID != "" {
		query = query.Where("category_id = ?", *categoryID)
	}
	if subcategoryID != nil && *subcategoryID != "" {
		query = query.Where("subcategory_id = ?", *subcategoryID)
	}

	if err := query.Find(&products).Error; err != nil {
		return nil, errors.NewInternalServerError("Failed to retrieve products with filters")
	}
	return products, nil
}

// GetProductsByQuantityRange retrieves products within a quantity range across all warehouses
// Only counts non-expired batches (expiry_date > NOW()) with available stock
func (r *ProductRepository) GetProductsByQuantityRange(minQty, maxQty int64, limit, offset int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	// Subquery to calculate total available quantity per product
	// SUM(total_quantity) aggregates quantity across all warehouses and variants
	subQuery := r.db.Table("inventory_batches AS ib").
		Select("pv.product_id, SUM(ib.total_quantity - ib.reserved_quantity) as available_qty").
		Joins("JOIN product_variants AS pv ON ib.variant_id = pv.id").
		Where("ib.expiry_date > NOW()").                   // Only non-expired batches
		Where("ib.total_quantity > ib.reserved_quantity"). // Only batches with available stock
		Group("pv.product_id").
		Having("SUM(ib.total_quantity - ib.reserved_quantity) BETWEEN ? AND ?", minQty, maxQty)

	// Get count of matching products
	countQuery := r.db.Table("(?) as filtered", subQuery).Count(&total)
	if countQuery.Error != nil {
		fmt.Printf("DEBUG: Database error counting products by quantity range: %v\n", countQuery.Error)
		return nil, 0, errors.NewInternalServerError("Failed to count products by quantity range")
	}

	// Get paginated products
	var productIDs []string
	if err := r.db.Table("(?) as filtered", subQuery).
		Select("product_id").
		Limit(limit).
		Offset(offset).
		Pluck("product_id", &productIDs).Error; err != nil {
		fmt.Printf("DEBUG: Database error retrieving products by quantity range: %v\n", err)
		return nil, 0, errors.NewInternalServerError("Failed to retrieve products by quantity range")
	}

	// If no products found, return empty slice
	if len(productIDs) == 0 {
		return []models.Product{}, total, nil
	}

	// Fetch full product details
	if err := r.db.Where("id IN ?", productIDs).Order("created_at DESC").Find(&products).Error; err != nil {
		fmt.Printf("DEBUG: Database error fetching product details: %v\n", err)
		return nil, 0, errors.NewInternalServerError("Failed to fetch product details")
	}

	return products, total, nil
}
