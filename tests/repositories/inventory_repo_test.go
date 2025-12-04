package repositories

import (
	"testing"
	"time"

	"gorm.io/gorm"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/tests/testutils"
)

// =============================================================================
// Test Setup
// =============================================================================

func setupInventoryRepo(t *testing.T) (*repositories.InventoryRepository, *gorm.DB, func()) {
	t.Helper()

	db := testutils.SetupTestDB(t)
	repo := repositories.NewInventoryRepository(db)

	cleanup := func() {
		testutils.CleanupTestDB(db)
	}

	return repo, db, cleanup
}

// createTestBatch creates an inventory batch with specific quantities
func createTestBatch(t *testing.T, repo *repositories.InventoryRepository, warehouseID, variantID string, totalQty, reservedQty int64) *models.InventoryBatch {
	t.Helper()

	expiryDate := time.Now().UTC().Add(30 * 24 * time.Hour)
	batch := models.NewInventoryBatch(warehouseID, variantID, 100.0, expiryDate, totalQty, 9.0, 9.0, []string{}, false)
	batch.ReservedQuantity = reservedQty

	if err := repo.CreateBatch(batch); err != nil {
		t.Fatalf("Failed to create test batch: %v", err)
	}

	return batch
}

// =============================================================================
// ReserveBatchStockWithTx Tests
// =============================================================================

func TestInventoryRepo_ReserveBatchStock_Success(t *testing.T) {
	repo, db, cleanup := setupInventoryRepo(t)
	defer cleanup()

	// ARRANGE: Create batch with 100 total, 0 reserved
	batch := createTestBatch(t, repo, "WH-001", "VAR-001", 100, 0)

	// ACT: Reserve 30 units using the SAME database connection
	tx := db.Begin()
	err := repo.ReserveBatchStockWithTx(tx, batch.ID, 30)
	tx.Commit()

	// ASSERT
	testutils.AssertNoError(t, err, "ReserveBatchStockWithTx should succeed")

	// Verify batch was updated
	updatedBatch, err := repo.GetBatchByID(batch.ID)
	testutils.AssertNoError(t, err, "GetBatchByID should succeed")
	testutils.AssertEqual(t, updatedBatch.ReservedQuantity, int64(30), "Reserved should be 30")
	testutils.AssertEqual(t, updatedBatch.TotalQuantity, int64(100), "Total should be unchanged")
}

func TestInventoryRepo_ReserveBatchStock_InsufficientQuantity(t *testing.T) {
	repo, db, cleanup := setupInventoryRepo(t)
	defer cleanup()

	// ARRANGE: Create batch with 50 total, 30 reserved (20 available)
	batch := createTestBatch(t, repo, "WH-001", "VAR-001", 50, 30)

	// ACT: Try to reserve 30 (more than available 20)
	tx := db.Begin()
	err := repo.ReserveBatchStockWithTx(tx, batch.ID, 30)
	tx.Rollback()

	// ASSERT
	testutils.AssertError(t, err, "Should return error for insufficient quantity")
	testutils.AssertContains(t, err.Error(), "Insufficient available stock", "Error should mention insufficient stock")
}

// =============================================================================
// ReleaseBatchReservationWithTx Tests
// =============================================================================

func TestInventoryRepo_ReleaseBatchReservation_Success(t *testing.T) {
	repo, db, cleanup := setupInventoryRepo(t)
	defer cleanup()

	// ARRANGE: Create batch with 100 total, 50 reserved
	batch := createTestBatch(t, repo, "WH-001", "VAR-001", 100, 50)

	// ACT: Release 30 reserved using the SAME database connection
	tx := db.Begin()
	err := repo.ReleaseBatchReservationWithTx(tx, batch.ID, 30)
	tx.Commit()

	// ASSERT
	testutils.AssertNoError(t, err, "ReleaseBatchReservationWithTx should succeed")

	// Verify batch was updated
	updatedBatch, err := repo.GetBatchByID(batch.ID)
	testutils.AssertNoError(t, err, "GetBatchByID should succeed")
	testutils.AssertEqual(t, updatedBatch.ReservedQuantity, int64(20), "Reserved should be 20")
	testutils.AssertEqual(t, updatedBatch.TotalQuantity, int64(100), "Total should be unchanged")
}

func TestInventoryRepo_ReleaseBatchReservation_InsufficientReserved(t *testing.T) {
	repo, db, cleanup := setupInventoryRepo(t)
	defer cleanup()

	// ARRANGE: Create batch with 100 total, 20 reserved
	batch := createTestBatch(t, repo, "WH-001", "VAR-001", 100, 20)

	// ACT: Try to release 30 (more than reserved 20)
	tx := db.Begin()
	err := repo.ReleaseBatchReservationWithTx(tx, batch.ID, 30)
	tx.Rollback()

	// ASSERT
	testutils.AssertError(t, err, "Should return error for insufficient reserved")
	testutils.AssertContains(t, err.Error(), "Cannot release more than reserved", "Error should mention cannot release")
}

// =============================================================================
// ConvertReservationToDeductionWithTx Tests
// =============================================================================

func TestInventoryRepo_ConvertReservationToDeduction_Success(t *testing.T) {
	repo, db, cleanup := setupInventoryRepo(t)
	defer cleanup()

	// ARRANGE: Create batch with 100 total, 50 reserved
	batch := createTestBatch(t, repo, "WH-001", "VAR-001", 100, 50)

	// ACT: Convert 30 reservation to deduction using the SAME database connection
	tx := db.Begin()
	err := repo.ConvertReservationToDeductionWithTx(tx, batch.ID, 30)
	tx.Commit()

	// ASSERT
	testutils.AssertNoError(t, err, "ConvertReservationToDeductionWithTx should succeed")

	// Verify batch was updated - both reserved and total should decrease
	updatedBatch, err := repo.GetBatchByID(batch.ID)
	testutils.AssertNoError(t, err, "GetBatchByID should succeed")
	testutils.AssertEqual(t, updatedBatch.TotalQuantity, int64(70), "Total should be 70 (100 - 30)")
	testutils.AssertEqual(t, updatedBatch.ReservedQuantity, int64(20), "Reserved should be 20 (50 - 30)")
}

func TestInventoryRepo_ConvertReservationToDeduction_InsufficientReserved(t *testing.T) {
	repo, db, cleanup := setupInventoryRepo(t)
	defer cleanup()

	// ARRANGE: Create batch with 100 total, 20 reserved
	batch := createTestBatch(t, repo, "WH-001", "VAR-001", 100, 20)

	// ACT: Try to convert 30 (more than reserved 20)
	tx := db.Begin()
	err := repo.ConvertReservationToDeductionWithTx(tx, batch.ID, 30)
	tx.Rollback()

	// ASSERT
	testutils.AssertError(t, err, "Should return error for insufficient reserved")
	testutils.AssertContains(t, err.Error(), "Cannot convert more than reserved", "Error should mention cannot convert")
}
