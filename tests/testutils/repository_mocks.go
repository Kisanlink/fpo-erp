package testutils

import (
	"time"

	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ========================================
// Mock Repository Interfaces (Following E-commerce Pattern)
// ========================================

// MockInventoryRepository mocks the InventoryRepository
type MockInventoryRepository struct {
	mock.Mock
}

func (m *MockInventoryRepository) CreateBatchWithTransaction(batch *models.InventoryBatch, transaction *models.InventoryTransaction) error {
	args := m.Called(batch, transaction)
	return args.Error(0)
}

func (m *MockInventoryRepository) CreateBatch(batch *models.InventoryBatch) error {
	args := m.Called(batch)
	return args.Error(0)
}

func (m *MockInventoryRepository) GetBatchByID(id string) (*models.InventoryBatch, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InventoryBatch), args.Error(1)
}

func (m *MockInventoryRepository) GetBatchesByWarehouse(warehouseID string) ([]models.InventoryBatch, error) {
	args := m.Called(warehouseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryBatch), args.Error(1)
}

func (m *MockInventoryRepository) GetBatchesByVariant(variantID string) ([]models.InventoryBatch, error) {
	args := m.Called(variantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryBatch), args.Error(1)
}

func (m *MockInventoryRepository) GetBatchesByVariantOrderedByExpiry(variantID string) ([]models.InventoryBatch, error) {
	args := m.Called(variantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryBatch), args.Error(1)
}

func (m *MockInventoryRepository) GetBatchesByVariantAndWarehouseOrderedByExpiry(variantID, warehouseID string) ([]models.InventoryBatch, error) {
	args := m.Called(variantID, warehouseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryBatch), args.Error(1)
}

func (m *MockInventoryRepository) GetAllBatches() ([]models.InventoryBatch, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryBatch), args.Error(1)
}

func (m *MockInventoryRepository) UpdateBatch(batch *models.InventoryBatch) error {
	args := m.Called(batch)
	return args.Error(0)
}

func (m *MockInventoryRepository) DeleteBatch(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockInventoryRepository) GetExpiringBatches(days int) ([]models.InventoryBatch, error) {
	args := m.Called(days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryBatch), args.Error(1)
}

func (m *MockInventoryRepository) GetLowStockBatches(threshold int64) ([]models.InventoryBatch, error) {
	args := m.Called(threshold)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryBatch), args.Error(1)
}

func (m *MockInventoryRepository) CreateTransaction(transaction *models.InventoryTransaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockInventoryRepository) GetTransactionsByBatch(batchID string) ([]models.InventoryTransaction, error) {
	args := m.Called(batchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryTransaction), args.Error(1)
}

func (m *MockInventoryRepository) GetTransactionsByType(transactionType string) ([]models.InventoryTransaction, error) {
	args := m.Called(transactionType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryTransaction), args.Error(1)
}

func (m *MockInventoryRepository) GetTransactionsByDateRange(startDate, endDate time.Time) ([]models.InventoryTransaction, error) {
	args := m.Called(startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InventoryTransaction), args.Error(1)
}

func (m *MockInventoryRepository) GetBatchStockLevel(batchID string) (int64, error) {
	args := m.Called(batchID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockInventoryRepository) UpdateBatchStock(batchID string, quantityChange int64) error {
	args := m.Called(batchID, quantityChange)
	return args.Error(0)
}

func (m *MockInventoryRepository) UpdateBatchStockWithTx(tx *gorm.DB, batchID string, quantityChange int64) error {
	args := m.Called(tx, batchID, quantityChange)
	return args.Error(0)
}

func (m *MockInventoryRepository) CreateTransactionWithTx(tx *gorm.DB, transaction *models.InventoryTransaction) error {
	args := m.Called(tx, transaction)
	return args.Error(0)
}

// ========================================
// MockPurchaseOrderRepository
// ========================================

type MockPurchaseOrderRepository struct {
	mock.Mock
}

func (m *MockPurchaseOrderRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockPurchaseOrderRepository) Create(po *models.PurchaseOrder) error {
	args := m.Called(po)
	return args.Error(0)
}

func (m *MockPurchaseOrderRepository) CreateWithTx(tx *gorm.DB, po *models.PurchaseOrder) error {
	args := m.Called(tx, po)
	return args.Error(0)
}

func (m *MockPurchaseOrderRepository) CreateItem(item *models.PurchaseOrderItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockPurchaseOrderRepository) CreateItemWithTx(tx *gorm.DB, item *models.PurchaseOrderItem) error {
	args := m.Called(tx, item)
	return args.Error(0)
}

func (m *MockPurchaseOrderRepository) GetByID(id string) (*models.PurchaseOrder, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchaseOrder), args.Error(1)
}

func (m *MockPurchaseOrderRepository) GetByIDWithItems(id string) (*models.PurchaseOrder, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchaseOrder), args.Error(1)
}

func (m *MockPurchaseOrderRepository) GetByPONumber(poNumber string) (*models.PurchaseOrder, error) {
	args := m.Called(poNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchaseOrder), args.Error(1)
}

func (m *MockPurchaseOrderRepository) GetAll() ([]models.PurchaseOrder, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PurchaseOrder), args.Error(1)
}

func (m *MockPurchaseOrderRepository) GetByCollaborator(collaboratorID string) ([]models.PurchaseOrder, error) {
	args := m.Called(collaboratorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PurchaseOrder), args.Error(1)
}

func (m *MockPurchaseOrderRepository) GetByStatus(status string) ([]models.PurchaseOrder, error) {
	args := m.Called(status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PurchaseOrder), args.Error(1)
}

func (m *MockPurchaseOrderRepository) GetPendingDeliveries() ([]models.PurchaseOrder, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PurchaseOrder), args.Error(1)
}

func (m *MockPurchaseOrderRepository) Update(po *models.PurchaseOrder) error {
	args := m.Called(po)
	return args.Error(0)
}

func (m *MockPurchaseOrderRepository) UpdateWithTx(tx *gorm.DB, po *models.PurchaseOrder) error {
	args := m.Called(tx, po)
	return args.Error(0)
}

func (m *MockPurchaseOrderRepository) UpdateStatus(poID, status string) error {
	args := m.Called(poID, status)
	return args.Error(0)
}

func (m *MockPurchaseOrderRepository) GetItemByID(itemID string) (*models.PurchaseOrderItem, error) {
	args := m.Called(itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchaseOrderItem), args.Error(1)
}

func (m *MockPurchaseOrderRepository) UpdateItemReceivedQuantity(itemID string, receivedQty int64) error {
	args := m.Called(itemID, receivedQty)
	return args.Error(0)
}

func (m *MockPurchaseOrderRepository) PONumberExists(poNumber string) (bool, error) {
	args := m.Called(poNumber)
	return args.Bool(0), args.Error(1)
}

func (m *MockPurchaseOrderRepository) FindByExternalOrderID(externalOrderID string) (*models.PurchaseOrder, error) {
	args := m.Called(externalOrderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchaseOrder), args.Error(1)
}

// ========================================
// MockGRNRepository
// ========================================

type MockGRNRepository struct {
	mock.Mock
}

func (m *MockGRNRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockGRNRepository) Create(grn *models.GRN) error {
	args := m.Called(grn)
	return args.Error(0)
}

func (m *MockGRNRepository) CreateWithTx(tx *gorm.DB, grn *models.GRN) error {
	args := m.Called(tx, grn)
	return args.Error(0)
}

func (m *MockGRNRepository) CreateItem(item *models.GRNItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockGRNRepository) CreateItemWithTx(tx *gorm.DB, item *models.GRNItem) error {
	args := m.Called(tx, item)
	return args.Error(0)
}

func (m *MockGRNRepository) UpdateItemBatchIDWithTx(tx *gorm.DB, itemID, batchID string) error {
	args := m.Called(tx, itemID, batchID)
	return args.Error(0)
}

func (m *MockGRNRepository) Update(id string, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockGRNRepository) GetByID(id string) (*models.GRN, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GRN), args.Error(1)
}

func (m *MockGRNRepository) GetByIDWithItems(id string) (*models.GRN, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GRN), args.Error(1)
}

func (m *MockGRNRepository) GetByGRNNumber(grnNumber string) (*models.GRN, error) {
	args := m.Called(grnNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GRN), args.Error(1)
}

func (m *MockGRNRepository) GetByPurchaseOrder(poID string) (*models.GRN, error) {
	args := m.Called(poID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GRN), args.Error(1)
}

func (m *MockGRNRepository) GetAll() ([]models.GRN, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GRN), args.Error(1)
}

func (m *MockGRNRepository) GetByWarehouse(warehouseID string) ([]models.GRN, error) {
	args := m.Called(warehouseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GRN), args.Error(1)
}

func (m *MockGRNRepository) GRNNumberExists(grnNumber string) (bool, error) {
	args := m.Called(grnNumber)
	return args.Bool(0), args.Error(1)
}

func (m *MockGRNRepository) GRNExistsForPO(poID string) (bool, error) {
	args := m.Called(poID)
	return args.Bool(0), args.Error(1)
}

// ========================================
// MockSalesRepository
// ========================================

type MockSalesRepository struct {
	mock.Mock
}

func (m *MockSalesRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockSalesRepository) CreateSaleWithTx(tx *gorm.DB, sale *models.Sale) error {
	args := m.Called(tx, sale)
	return args.Error(0)
}

func (m *MockSalesRepository) CreateSaleItemWithTx(tx *gorm.DB, item *models.SaleItem) error {
	args := m.Called(tx, item)
	return args.Error(0)
}

func (m *MockSalesRepository) UpdateSaleWithTx(tx *gorm.DB, sale *models.Sale) error {
	args := m.Called(tx, sale)
	return args.Error(0)
}

func (m *MockSalesRepository) CreateSale(sale *models.Sale) error {
	args := m.Called(sale)
	return args.Error(0)
}

func (m *MockSalesRepository) GetSaleByID(id string) (*models.Sale, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Sale), args.Error(1)
}

func (m *MockSalesRepository) GetAllSales(limit, offset int) ([]models.Sale, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Sale), args.Error(1)
}

func (m *MockSalesRepository) UpdateSale(sale *models.Sale) error {
	args := m.Called(sale)
	return args.Error(0)
}

func (m *MockSalesRepository) DeleteSale(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockSalesRepository) GetSalesByDateRange(startDate, endDate time.Time) ([]models.Sale, error) {
	args := m.Called(startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Sale), args.Error(1)
}

func (m *MockSalesRepository) GetSalesByStatus(status string) ([]models.Sale, error) {
	args := m.Called(status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Sale), args.Error(1)
}

func (m *MockSalesRepository) CreateSaleItem(item *models.SaleItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockSalesRepository) GetSaleItemsBySaleID(saleID string) ([]models.SaleItem, error) {
	args := m.Called(saleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SaleItem), args.Error(1)
}

// ========================================
// MockDiscountsRepository
// ========================================

type MockDiscountsRepository struct {
	mock.Mock
}

func (m *MockDiscountsRepository) CreateDiscount(discount *models.Discount) error {
	args := m.Called(discount)
	return args.Error(0)
}

func (m *MockDiscountsRepository) GetDiscountByID(id string) (*models.Discount, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Discount), args.Error(1)
}

func (m *MockDiscountsRepository) GetDiscountByCode(code string) (*models.Discount, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Discount), args.Error(1)
}

func (m *MockDiscountsRepository) GetAllDiscounts(limit, offset int) ([]models.Discount, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Discount), args.Error(1)
}

func (m *MockDiscountsRepository) GetActiveDiscounts() ([]models.Discount, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Discount), args.Error(1)
}

func (m *MockDiscountsRepository) UpdateDiscount(discount *models.Discount) error {
	args := m.Called(discount)
	return args.Error(0)
}

func (m *MockDiscountsRepository) DeleteDiscount(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockDiscountsRepository) IncrementUsage(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockDiscountsRepository) CreateDiscountUsage(usage *models.DiscountUsage) error {
	args := m.Called(usage)
	return args.Error(0)
}

func (m *MockDiscountsRepository) CreateDiscountUsageWithTx(tx *gorm.DB, usage *models.DiscountUsage) error {
	args := m.Called(tx, usage)
	return args.Error(0)
}

func (m *MockDiscountsRepository) IncrementUsageWithTx(tx *gorm.DB, discountID string) error {
	args := m.Called(tx, discountID)
	return args.Error(0)
}

func (m *MockDiscountsRepository) ValidateDiscount(code string, orderValue float64, productIDs []string, warehouseID string) (*models.Discount, error) {
	args := m.Called(code, orderValue, productIDs, warehouseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Discount), args.Error(1)
}

func (m *MockDiscountsRepository) CalculateDiscount(discount *models.Discount, orderValue float64) float64 {
	args := m.Called(discount, orderValue)
	return args.Get(0).(float64)
}

// ========================================
// MockWarehouseRepository
// ========================================

type MockWarehouseRepository struct {
	mock.Mock
}

func (m *MockWarehouseRepository) Create(warehouse *models.Warehouse) error {
	args := m.Called(warehouse)
	return args.Error(0)
}

func (m *MockWarehouseRepository) GetByID(id string) (*models.Warehouse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Warehouse), args.Error(1)
}

func (m *MockWarehouseRepository) GetAll() ([]models.Warehouse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Warehouse), args.Error(1)
}

func (m *MockWarehouseRepository) Update(warehouse *models.Warehouse) error {
	args := m.Called(warehouse)
	return args.Error(0)
}

func (m *MockWarehouseRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockWarehouseRepository) Exists(id string) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

// ========================================
// MockProductRepository
// ========================================

type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(product *models.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(id string) (*models.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) GetAll(limit, offset int) ([]models.Product, int64, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) Update(product *models.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductRepository) Exists(id string) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

// ========================================
// MockProductVariantRepository
// ========================================

type MockProductVariantRepository struct {
	mock.Mock
}

func (m *MockProductVariantRepository) Create(variant *models.ProductVariant) error {
	args := m.Called(variant)
	return args.Error(0)
}

func (m *MockProductVariantRepository) GetByID(id string) (*models.ProductVariant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariant), args.Error(1)
}

func (m *MockProductVariantRepository) GetByProductID(productID string) ([]models.ProductVariant, error) {
	args := m.Called(productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ProductVariant), args.Error(1)
}

func (m *MockProductVariantRepository) Update(variant *models.ProductVariant) error {
	args := m.Called(variant)
	return args.Error(0)
}

func (m *MockProductVariantRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// ========================================
// MockCollaboratorRepository
// ========================================

type MockCollaboratorRepository struct {
	mock.Mock
}

func (m *MockCollaboratorRepository) Create(collaborator *models.Collaborator) error {
	args := m.Called(collaborator)
	return args.Error(0)
}

func (m *MockCollaboratorRepository) GetByID(id string) (*models.Collaborator, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Collaborator), args.Error(1)
}

func (m *MockCollaboratorRepository) GetAll() ([]models.Collaborator, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Collaborator), args.Error(1)
}

func (m *MockCollaboratorRepository) Update(collaborator *models.Collaborator) error {
	args := m.Called(collaborator)
	return args.Error(0)
}

func (m *MockCollaboratorRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
