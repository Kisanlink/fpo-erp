package services

import (
	"testing"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"
	"kisanlink-erp/tests/testutils"

	"gorm.io/gorm"
)

// ========================================
// Setup and Helper Functions
// ========================================

// setupTaxService creates a TaxService with in-memory database
func setupTaxService(t *testing.T) (*services.TaxService, *gorm.DB, func()) {
	t.Helper()

	db := testutils.SetupTestDB(t)

	// Create repository
	taxRepo := repositories.NewTaxRepository(db)

	// Create service
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewTaxService(taxRepo, mockLogger)

	cleanup := func() {
		testutils.CleanupTestDB(db)
	}

	return service, db, cleanup
}

// assertGSTCalculation verifies all fields of a GSTCalculation result
func assertGSTCalculation(t *testing.T, result *models.GSTCalculation, expected *models.GSTCalculation, testName string) {
	t.Helper()

	testutils.AssertNotNil(t, result, testName+": Result should not be nil")
	testutils.AssertEqual(t, result.CGSTAmount, expected.CGSTAmount, testName+": CGST amount mismatch")
	testutils.AssertEqual(t, result.SGSTAmount, expected.SGSTAmount, testName+": SGST amount mismatch")
	testutils.AssertEqual(t, result.IGSTAmount, expected.IGSTAmount, testName+": IGST amount mismatch")
	testutils.AssertEqual(t, result.TotalTaxAmount, expected.TotalTaxAmount, testName+": Total tax amount mismatch")
	testutils.AssertEqual(t, result.IsInterState, expected.IsInterState, testName+": IsInterState flag mismatch")
}

// ========================================
// Intra-State (CGST + SGST) Tests
// ========================================

func TestCalculateGST_IntraState_5Percent(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// Test case: 100 INR at 5% GST (intra-state)
	// Expected: CGST = 2.50, SGST = 2.50, Total = 5.00
	result := service.CalculateGST(100.00, 5.0, false)

	expected := &models.GSTCalculation{
		CGSTAmount:     2.50,
		SGSTAmount:     2.50,
		IGSTAmount:     0.00,
		TotalTaxAmount: 5.00,
		IsInterState:   false,
	}

	assertGSTCalculation(t, result, expected, "Intra-state 5% GST")
}

func TestCalculateGST_IntraState_18Percent(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// Test case: 1000 INR at 18% GST (intra-state)
	// Expected: CGST = 90.00, SGST = 90.00, Total = 180.00
	result := service.CalculateGST(1000.00, 18.0, false)

	expected := &models.GSTCalculation{
		CGSTAmount:     90.00,
		SGSTAmount:     90.00,
		IGSTAmount:     0.00,
		TotalTaxAmount: 180.00,
		IsInterState:   false,
	}

	assertGSTCalculation(t, result, expected, "Intra-state 18% GST")
}

func TestCalculateGST_IntraState_12Percent(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// Test case: 500 INR at 12% GST (intra-state)
	// Expected: CGST = 30.00, SGST = 30.00, Total = 60.00
	result := service.CalculateGST(500.00, 12.0, false)

	expected := &models.GSTCalculation{
		CGSTAmount:     30.00,
		SGSTAmount:     30.00,
		IGSTAmount:     0.00,
		TotalTaxAmount: 60.00,
		IsInterState:   false,
	}

	assertGSTCalculation(t, result, expected, "Intra-state 12% GST")
}

func TestCalculateGST_IntraState_28Percent(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// Test case: 200 INR at 28% GST (intra-state)
	// Expected: CGST = 28.00, SGST = 28.00, Total = 56.00
	result := service.CalculateGST(200.00, 28.0, false)

	expected := &models.GSTCalculation{
		CGSTAmount:     28.00,
		SGSTAmount:     28.00,
		IGSTAmount:     0.00,
		TotalTaxAmount: 56.00,
		IsInterState:   false,
	}

	assertGSTCalculation(t, result, expected, "Intra-state 28% GST")
}

func TestCalculateGST_IntraState_OddAmount(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// Test case: 100.01 INR at 5% GST (intra-state)
	// Expected: CGST = 2.50, SGST = 2.50
	// This verifies that CGST == SGST invariant holds even with odd amounts
	result := service.CalculateGST(100.01, 5.0, false)

	testutils.AssertNotNil(t, result, "Result should not be nil")
	testutils.AssertEqual(t, result.IsInterState, false, "Should be intra-state")
	testutils.AssertEqual(t, result.IGSTAmount, 0.00, "IGST should be zero for intra-state")

	// Critical invariant: CGST must equal SGST
	testutils.AssertEqual(t, result.CGSTAmount, result.SGSTAmount, "CGST must equal SGST (50-50 split)")

	// Verify rounding behavior
	expectedCGST := 2.50 // (100.01 * 2.5 / 100) = 2.50025 → 2.50 (rounded)
	testutils.AssertEqual(t, result.CGSTAmount, expectedCGST, "CGST should be rounded to nearest paisa")
}

// ========================================
// Inter-State (IGST) Tests
// ========================================

func TestCalculateGST_InterState_5Percent(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// Test case: 100 INR at 5% GST (inter-state)
	// Expected: IGST = 5.00
	result := service.CalculateGST(100.00, 5.0, true)

	expected := &models.GSTCalculation{
		CGSTAmount:     0.00,
		SGSTAmount:     0.00,
		IGSTAmount:     5.00,
		TotalTaxAmount: 5.00,
		IsInterState:   true,
	}

	assertGSTCalculation(t, result, expected, "Inter-state 5% GST")
}

func TestCalculateGST_InterState_18Percent(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// Test case: 1000 INR at 18% GST (inter-state)
	// Expected: IGST = 180.00
	result := service.CalculateGST(1000.00, 18.0, true)

	expected := &models.GSTCalculation{
		CGSTAmount:     0.00,
		SGSTAmount:     0.00,
		IGSTAmount:     180.00,
		TotalTaxAmount: 180.00,
		IsInterState:   true,
	}

	assertGSTCalculation(t, result, expected, "Inter-state 18% GST")
}

func TestCalculateGST_InterState_12Percent(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// Test case: 750 INR at 12% GST (inter-state)
	// Expected: IGST = 90.00
	result := service.CalculateGST(750.00, 12.0, true)

	expected := &models.GSTCalculation{
		CGSTAmount:     0.00,
		SGSTAmount:     0.00,
		IGSTAmount:     90.00,
		TotalTaxAmount: 90.00,
		IsInterState:   true,
	}

	assertGSTCalculation(t, result, expected, "Inter-state 12% GST")
}

func TestCalculateGST_InterState_28Percent(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// Test case: 500 INR at 28% GST (inter-state)
	// Expected: IGST = 140.00
	result := service.CalculateGST(500.00, 28.0, true)

	expected := &models.GSTCalculation{
		CGSTAmount:     0.00,
		SGSTAmount:     0.00,
		IGSTAmount:     140.00,
		TotalTaxAmount: 140.00,
		IsInterState:   true,
	}

	assertGSTCalculation(t, result, expected, "Inter-state 28% GST")
}

// ========================================
// Zero Rate Tests
// ========================================

func TestCalculateGST_ZeroRate_IntraState(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// Test case: 100 INR at 0% GST (intra-state)
	// Expected: All amounts should be zero
	result := service.CalculateGST(100.00, 0.0, false)

	expected := &models.GSTCalculation{
		CGSTAmount:     0.00,
		SGSTAmount:     0.00,
		IGSTAmount:     0.00,
		TotalTaxAmount: 0.00,
		IsInterState:   false,
	}

	assertGSTCalculation(t, result, expected, "Zero rate intra-state")
}

func TestCalculateGST_ZeroRate_InterState(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// Test case: 100 INR at 0% GST (inter-state)
	// Expected: All amounts should be zero
	result := service.CalculateGST(100.00, 0.0, true)

	expected := &models.GSTCalculation{
		CGSTAmount:     0.00,
		SGSTAmount:     0.00,
		IGSTAmount:     0.00,
		TotalTaxAmount: 0.00,
		IsInterState:   true,
	}

	assertGSTCalculation(t, result, expected, "Zero rate inter-state")
}

// ========================================
// Invariant Tests (Critical GST Rules)
// ========================================

func TestCalculateGST_Invariant_CGSTEqualsSGST(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// GST Law invariant: For intra-state sales, CGST must ALWAYS equal SGST (50-50 split)
	testCases := []struct {
		lineTotal float64
		gstRate   float64
		name      string
	}{
		{100.00, 5.0, "100 INR at 5%"},
		{1000.00, 18.0, "1000 INR at 18%"},
		{99.99, 12.0, "99.99 INR at 12%"},
		{123.45, 28.0, "123.45 INR at 28%"},
		{0.01, 5.0, "0.01 INR at 5%"},
		{10000.00, 18.0, "10000 INR at 18%"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.CalculateGST(tc.lineTotal, tc.gstRate, false)

			testutils.AssertNotNil(t, result, "Result should not be nil")
			testutils.AssertEqual(t, result.CGSTAmount, result.SGSTAmount, "CGST must equal SGST for intra-state")
			testutils.AssertEqual(t, result.IGSTAmount, 0.00, "IGST must be zero for intra-state")
		})
	}
}

func TestCalculateGST_Invariant_TotalTaxCorrect(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// Invariant: TotalTaxAmount = CGST + SGST + IGST
	testCases := []struct {
		lineTotal    float64
		gstRate      float64
		isInterState bool
		name         string
	}{
		{100.00, 5.0, false, "Intra-state 5%"},
		{1000.00, 18.0, false, "Intra-state 18%"},
		{100.00, 5.0, true, "Inter-state 5%"},
		{1000.00, 18.0, true, "Inter-state 18%"},
		{99.99, 12.0, false, "Intra-state 12% odd amount"},
		{750.50, 28.0, true, "Inter-state 28% odd amount"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.CalculateGST(tc.lineTotal, tc.gstRate, tc.isInterState)

			testutils.AssertNotNil(t, result, "Result should not be nil")

			calculatedTotal := result.CGSTAmount + result.SGSTAmount + result.IGSTAmount
			testutils.AssertEqual(t, result.TotalTaxAmount, calculatedTotal, "TotalTaxAmount must equal sum of all tax components")
		})
	}
}

func TestCalculateGST_Invariant_MutuallyExclusive(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// GST Law invariant: Either (CGST+SGST) OR IGST, never both
	testCases := []struct {
		lineTotal    float64
		gstRate      float64
		isInterState bool
		name         string
	}{
		{100.00, 5.0, false, "Intra-state should have CGST+SGST only"},
		{100.00, 5.0, true, "Inter-state should have IGST only"},
		{1000.00, 18.0, false, "Intra-state 18% should have CGST+SGST only"},
		{1000.00, 18.0, true, "Inter-state 18% should have IGST only"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.CalculateGST(tc.lineTotal, tc.gstRate, tc.isInterState)

			testutils.AssertNotNil(t, result, "Result should not be nil")

			if tc.isInterState {
				// Inter-state: IGST only, CGST+SGST must be zero
				testutils.AssertEqual(t, result.CGSTAmount, 0.00, "CGST must be zero for inter-state")
				testutils.AssertEqual(t, result.SGSTAmount, 0.00, "SGST must be zero for inter-state")
				testutils.AssertTrue(t, result.IGSTAmount > 0, "IGST must be positive for inter-state")
			} else {
				// Intra-state: CGST+SGST only, IGST must be zero
				testutils.AssertEqual(t, result.IGSTAmount, 0.00, "IGST must be zero for intra-state")
				testutils.AssertTrue(t, result.CGSTAmount > 0, "CGST must be positive for intra-state")
				testutils.AssertTrue(t, result.SGSTAmount > 0, "SGST must be positive for intra-state")
			}
		})
	}
}

// ========================================
// Rounding Behavior Tests
// ========================================

func TestCalculateGST_RoundingBehavior(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// Test various amounts that exercise rounding to nearest paisa (0.01 INR)
	testCases := []struct {
		lineTotal      float64
		gstRate        float64
		isInterState   bool
		expectedCGST   float64
		expectedSGST   float64
		expectedIGST   float64
		expectedTotal  float64
		name           string
	}{
		// Intra-state cases requiring rounding
		{100.33, 5.0, false, 2.51, 2.51, 0.00, 5.02, "100.33 at 5% intra-state"},
		{99.99, 18.0, false, 9.00, 9.00, 0.00, 18.00, "99.99 at 18% intra-state"},
		{123.45, 12.0, false, 7.41, 7.41, 0.00, 14.82, "123.45 at 12% intra-state"},

		// Inter-state cases requiring rounding
		{100.33, 5.0, true, 0.00, 0.00, 5.02, 5.02, "100.33 at 5% inter-state"},
		{99.99, 18.0, true, 0.00, 0.00, 18.00, 18.00, "99.99 at 18% inter-state"},
		{123.45, 12.0, true, 0.00, 0.00, 14.81, 14.81, "123.45 at 12% inter-state"},

		// Edge case: Very small amounts
		{0.01, 18.0, false, 0.00, 0.00, 0.00, 0.00, "0.01 INR at 18% intra-state (rounds to 0)"},
		{0.10, 18.0, false, 0.01, 0.01, 0.00, 0.02, "0.10 INR at 18% intra-state"},

		// Edge case: Large amounts
		{10000.00, 28.0, false, 1400.00, 1400.00, 0.00, 2800.00, "10000 INR at 28% intra-state"},
		{10000.00, 28.0, true, 0.00, 0.00, 2800.00, 2800.00, "10000 INR at 28% inter-state"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.CalculateGST(tc.lineTotal, tc.gstRate, tc.isInterState)

			expected := &models.GSTCalculation{
				CGSTAmount:     tc.expectedCGST,
				SGSTAmount:     tc.expectedSGST,
				IGSTAmount:     tc.expectedIGST,
				TotalTaxAmount: tc.expectedTotal,
				IsInterState:   tc.isInterState,
			}

			assertGSTCalculation(t, result, expected, tc.name)
		})
	}
}

// ========================================
// Table-Driven Comprehensive Test
// ========================================

func TestCalculateGST_TableDriven(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	testCases := []struct {
		lineTotal      float64
		gstRate        float64
		isInterState   bool
		expectedCGST   float64
		expectedSGST   float64
		expectedIGST   float64
		expectedTotal  float64
		name           string
	}{
		// Standard GST rates - Intra-state
		{100.00, 5.0, false, 2.50, 2.50, 0.00, 5.00, "Standard: 100 INR at 5% intra-state"},
		{100.00, 12.0, false, 6.00, 6.00, 0.00, 12.00, "Standard: 100 INR at 12% intra-state"},
		{100.00, 18.0, false, 9.00, 9.00, 0.00, 18.00, "Standard: 100 INR at 18% intra-state"},
		{100.00, 28.0, false, 14.00, 14.00, 0.00, 28.00, "Standard: 100 INR at 28% intra-state"},

		// Standard GST rates - Inter-state
		{100.00, 5.0, true, 0.00, 0.00, 5.00, 5.00, "Standard: 100 INR at 5% inter-state"},
		{100.00, 12.0, true, 0.00, 0.00, 12.00, 12.00, "Standard: 100 INR at 12% inter-state"},
		{100.00, 18.0, true, 0.00, 0.00, 18.00, 18.00, "Standard: 100 INR at 18% inter-state"},
		{100.00, 28.0, true, 0.00, 0.00, 28.00, 28.00, "Standard: 100 INR at 28% inter-state"},

		// Zero rate (tax-exempt goods)
		{100.00, 0.0, false, 0.00, 0.00, 0.00, 0.00, "Zero rate: 100 INR at 0% intra-state"},
		{100.00, 0.0, true, 0.00, 0.00, 0.00, 0.00, "Zero rate: 100 INR at 0% inter-state"},

		// Large amounts
		{10000.00, 18.0, false, 900.00, 900.00, 0.00, 1800.00, "Large: 10000 INR at 18% intra-state"},
		{10000.00, 18.0, true, 0.00, 0.00, 1800.00, 1800.00, "Large: 10000 INR at 18% inter-state"},

		// Odd amounts with rounding
		{99.99, 5.0, false, 2.50, 2.50, 0.00, 5.00, "Odd: 99.99 INR at 5% intra-state"},
		{123.45, 18.0, false, 11.11, 11.11, 0.00, 22.22, "Odd: 123.45 INR at 18% intra-state"},
		{99.99, 12.0, true, 0.00, 0.00, 12.00, 12.00, "Odd: 99.99 INR at 12% inter-state"},

		// Edge case: Very small amounts
		{0.10, 18.0, false, 0.01, 0.01, 0.00, 0.02, "Edge: 0.10 INR at 18% intra-state"},
		{0.10, 18.0, true, 0.00, 0.00, 0.02, 0.02, "Edge: 0.10 INR at 18% inter-state"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.CalculateGST(tc.lineTotal, tc.gstRate, tc.isInterState)

			expected := &models.GSTCalculation{
				CGSTAmount:     tc.expectedCGST,
				SGSTAmount:     tc.expectedSGST,
				IGSTAmount:     tc.expectedIGST,
				TotalTaxAmount: tc.expectedTotal,
				IsInterState:   tc.isInterState,
			}

			assertGSTCalculation(t, result, expected, tc.name)
		})
	}
}

// ========================================
// Additional Edge Cases
// ========================================

func TestCalculateGST_EdgeCase_VerySmallAmount(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// Test case: 0.01 INR at 18% GST
	// (0.01 * 9 / 100) = 0.0009 → rounds to 0.00
	result := service.CalculateGST(0.01, 18.0, false)

	testutils.AssertNotNil(t, result, "Result should not be nil")
	testutils.AssertEqual(t, result.CGSTAmount, 0.00, "CGST should round to 0")
	testutils.AssertEqual(t, result.SGSTAmount, 0.00, "SGST should round to 0")
	testutils.AssertEqual(t, result.TotalTaxAmount, 0.00, "Total should be 0")
}

func TestCalculateGST_EdgeCase_VeryLargeAmount(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// Test case: 1,00,000 INR at 18% GST
	result := service.CalculateGST(100000.00, 18.0, false)

	expected := &models.GSTCalculation{
		CGSTAmount:     9000.00,
		SGSTAmount:     9000.00,
		IGSTAmount:     0.00,
		TotalTaxAmount: 18000.00,
		IsInterState:   false,
	}

	assertGSTCalculation(t, result, expected, "Very large amount")
}

func TestCalculateGST_EdgeCase_PrecisionBoundary(t *testing.T) {
	service, _, cleanup := setupTaxService(t)
	defer cleanup()

	// Test amounts that test precision boundaries
	// 33.33 * 18% = 5.9994 → should round correctly
	result := service.CalculateGST(33.33, 18.0, false)

	testutils.AssertNotNil(t, result, "Result should not be nil")

	// Verify CGST == SGST (critical invariant)
	testutils.AssertEqual(t, result.CGSTAmount, result.SGSTAmount, "CGST must equal SGST")

	// Verify total is sum of components
	calculatedTotal := result.CGSTAmount + result.SGSTAmount
	testutils.AssertEqual(t, result.TotalTaxAmount, calculatedTotal, "Total must equal sum of components")
}
