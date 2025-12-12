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

// setupCollaboratorService creates a CollaboratorService with in-memory database
// Note: This creates a service WITHOUT e-commerce client to avoid external dependencies
func setupCollaboratorService(t *testing.T) (*services.CollaboratorService, *gorm.DB, func()) {
	t.Helper()

	db := testutils.SetupTestDB(t)

	// Create repository
	collaboratorRepo := repositories.NewCollaboratorRepository(db)

	// Create service with nil clients (for testing legacy path without external dependencies)
	mockLogger := utils.NewLoggerAdapter(utils.GetZapLogger())
	service := services.NewCollaboratorService(
		collaboratorRepo,
		nil, // addressClient (not needed for validation tests)
		nil, // s3Service (not needed for validation tests)
		nil, // attachmentService (not needed for validation tests)
		nil, // ecomClient (nil = use legacy path)
		0,   // ecomTimeout
		"",  // ecomAuthToken
		mockLogger,
	)

	cleanup := func() {
		testutils.CleanupTestDB(db)
	}

	return service, db, cleanup
}

// ========================================
// Bank Details Validation Tests
// ========================================
// Tests for the validateBankDetails function through CreateCollaborator
// validateBankDetails ensures that if one bank field is provided, both must be provided

func TestCollaboratorService_CreateCollaborator_BankAccountWithoutIFSC_Fails(t *testing.T) {
	service, _, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContextWithUserID("USER_12345")
	accountNo := "123456789012"

	request := &models.CreateCollaboratorRequest{
		CompanyName:   "Test Company",
		ContactPerson: "John Doe",
		ContactNumber: "9876543210",
		BankAccountNo: &accountNo,
		BankIFSC:      nil, // Missing IFSC - should fail
	}

	_, err := service.CreateCollaborator(ctx, request, "ORG_001", "USER_12345", "fake-jwt-token")

	testutils.AssertError(t, err, "Should fail when bank account provided without IFSC")
	testutils.AssertContains(t, err.Error(), "bank_ifsc is required", "Error should mention missing IFSC")
}

func TestCollaboratorService_CreateCollaborator_IFSCWithoutBankAccount_Fails(t *testing.T) {
	service, _, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContextWithUserID("USER_12345")
	ifsc := "IFSC0001234"

	request := &models.CreateCollaboratorRequest{
		CompanyName:   "Test Company",
		ContactPerson: "John Doe",
		ContactNumber: "9876543210",
		BankAccountNo: nil, // Missing account number - should fail
		BankIFSC:      &ifsc,
	}

	_, err := service.CreateCollaborator(ctx, request, "ORG_001", "USER_12345", "fake-jwt-token")

	testutils.AssertError(t, err, "Should fail when IFSC provided without bank account")
	testutils.AssertContains(t, err.Error(), "bank_account_no is required", "Error should mention missing account number")
}

func TestCollaboratorService_CreateCollaborator_BothBankFieldsNil_Succeeds(t *testing.T) {
	service, _, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContextWithUserID("USER_12345")

	request := &models.CreateCollaboratorRequest{
		CompanyName:   "Test Company",
		ContactPerson: "John Doe",
		ContactNumber: "9876543210",
		GSTNumber:     "29GGGGG1314R9Z6",
		BankAccountNo: nil, // Both nil is OK
		BankIFSC:      nil,
	}

	response, err := service.CreateCollaborator(ctx, request, "ORG_001", "USER_12345", "fake-jwt-token")

	testutils.AssertNoError(t, err, "Should succeed when both bank fields are nil")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.CompanyName, "Test Company", "Company name should match")
}

func TestCollaboratorService_CreateCollaborator_BothBankFieldsProvided_Succeeds(t *testing.T) {
	service, _, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContextWithUserID("USER_12345")
	accountNo := "123456789012"
	ifsc := "IFSC0001234"

	request := &models.CreateCollaboratorRequest{
		CompanyName:   "Test Company",
		ContactPerson: "John Doe",
		ContactNumber: "9876543210",
		GSTNumber:     "29GGGGG1314R9Z7",
		BankAccountNo: &accountNo,
		BankIFSC:      &ifsc,
	}

	response, err := service.CreateCollaborator(ctx, request, "ORG_001", "USER_12345", "fake-jwt-token")

	testutils.AssertNoError(t, err, "Should succeed when both bank fields are provided")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, *response.BankAccountNo, accountNo, "Bank account should match")
	testutils.AssertEqual(t, *response.BankIFSC, ifsc, "IFSC should match")
}

func TestCollaboratorService_CreateCollaborator_BankAccountWhitespaceOnly_TreatedAsEmpty(t *testing.T) {
	service, _, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContextWithUserID("USER_12345")
	accountNo := "   " // Whitespace only - treated as empty

	request := &models.CreateCollaboratorRequest{
		CompanyName:   "Test Company",
		ContactPerson: "John Doe",
		ContactNumber: "9876543210",
		GSTNumber:     "29GGGGG1314R9Z8",
		BankAccountNo: &accountNo,
		BankIFSC:      nil,
	}

	response, err := service.CreateCollaborator(ctx, request, "ORG_001", "USER_12345", "fake-jwt-token")

	testutils.AssertNoError(t, err, "Should succeed when bank account is whitespace only (treated as empty)")
	testutils.AssertNotNil(t, response, "Response should not be nil")
}

func TestCollaboratorService_CreateCollaborator_IFSCWhitespaceOnly_TreatedAsEmpty(t *testing.T) {
	service, _, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContextWithUserID("USER_12345")
	ifsc := "   " // Whitespace only - treated as empty

	request := &models.CreateCollaboratorRequest{
		CompanyName:   "Test Company",
		ContactPerson: "John Doe",
		ContactNumber: "9876543210",
		GSTNumber:     "29GGGGG1314R9Z9",
		BankAccountNo: nil,
		BankIFSC:      &ifsc,
	}

	response, err := service.CreateCollaborator(ctx, request, "ORG_001", "USER_12345", "fake-jwt-token")

	testutils.AssertNoError(t, err, "Should succeed when IFSC is whitespace only (treated as empty)")
	testutils.AssertNotNil(t, response, "Response should not be nil")
}

func TestCollaboratorService_CreateCollaborator_BothBankFieldsWhitespace_Succeeds(t *testing.T) {
	service, _, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContextWithUserID("USER_12345")
	accountNo := "  " // Whitespace only
	ifsc := "  "      // Whitespace only

	request := &models.CreateCollaboratorRequest{
		CompanyName:   "Test Company",
		ContactPerson: "John Doe",
		ContactNumber: "9876543210",
		GSTNumber:     "29GGGGG1314R9ZA",
		BankAccountNo: &accountNo,
		BankIFSC:      &ifsc,
	}

	response, err := service.CreateCollaborator(ctx, request, "ORG_001", "USER_12345", "fake-jwt-token")

	testutils.AssertNoError(t, err, "Should succeed when both fields are whitespace (treated as empty)")
	testutils.AssertNotNil(t, response, "Response should not be nil")
}

// ========================================
// Bank Details Validation Tests - UpdateCollaborator
// ========================================
// Same validation rules apply to updates

func TestCollaboratorService_UpdateCollaborator_BankAccountWithoutIFSC_Fails(t *testing.T) {
	service, db, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContextWithUserID("USER_12345")

	// Create initial collaborator with complete bank details
	existingAccount := "999999999999"
	existingIFSC := "IFSC9999999"
	collaborator := testutils.FixtureCollaborator("Existing Company")
	collaborator.BankAccountNo = &existingAccount
	collaborator.BankIFSC = &existingIFSC
	err := db.Create(collaborator).Error
	testutils.AssertNoError(t, err, "Should create initial collaborator")

	// Try to update with only bank account (removing IFSC)
	newAccount := "123456789012"
	request := &models.UpdateCollaboratorRequest{
		BankAccountNo: &newAccount,
		BankIFSC:      nil, // Trying to remove IFSC while keeping account - should fail
	}

	_, err = service.UpdateCollaborator(ctx, collaborator.ID, request, "ORG_001", "USER_12345", "fake-jwt-token")

	testutils.AssertError(t, err, "Should fail when updating bank account without IFSC")
	testutils.AssertContains(t, err.Error(), "bank_ifsc is required", "Error should mention missing IFSC")
}

func TestCollaboratorService_UpdateCollaborator_IFSCWithoutBankAccount_Fails(t *testing.T) {
	service, db, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContextWithUserID("USER_12345")

	// Create initial collaborator with complete bank details
	existingAccount := "999999999999"
	existingIFSC := "IFSC9999999"
	collaborator := testutils.FixtureCollaborator("Existing Company")
	collaborator.BankAccountNo = &existingAccount
	collaborator.BankIFSC = &existingIFSC
	err := db.Create(collaborator).Error
	testutils.AssertNoError(t, err, "Should create initial collaborator")

	// Try to update with only IFSC (removing account)
	newIFSC := "IFSC0001234"
	request := &models.UpdateCollaboratorRequest{
		BankAccountNo: nil, // Trying to remove account while keeping IFSC - should fail
		BankIFSC:      &newIFSC,
	}

	_, err = service.UpdateCollaborator(ctx, collaborator.ID, request, "ORG_001", "USER_12345", "fake-jwt-token")

	testutils.AssertError(t, err, "Should fail when updating IFSC without bank account")
	testutils.AssertContains(t, err.Error(), "bank_account_no is required", "Error should mention missing account")
}

func TestCollaboratorService_UpdateCollaborator_BothBankFieldsProvided_Succeeds(t *testing.T) {
	service, db, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContextWithUserID("USER_12345")

	// Create initial collaborator
	collaborator := testutils.FixtureCollaborator("Existing Company")
	err := db.Create(collaborator).Error
	testutils.AssertNoError(t, err, "Should create initial collaborator")

	// Update with both bank fields
	newAccount := "123456789012"
	newIFSC := "IFSC0001234"
	request := &models.UpdateCollaboratorRequest{
		BankAccountNo: &newAccount,
		BankIFSC:      &newIFSC,
	}

	response, err := service.UpdateCollaborator(ctx, collaborator.ID, request, "ORG_001", "USER_12345", "fake-jwt-token")

	testutils.AssertNoError(t, err, "Should succeed when both bank fields are provided")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, *response.BankAccountNo, newAccount, "Bank account should be updated")
	testutils.AssertEqual(t, *response.BankIFSC, newIFSC, "IFSC should be updated")
}

// ========================================
// Additional CreateCollaborator Tests
// ========================================
// Basic functionality tests to ensure service works correctly

func TestCollaboratorService_CreateCollaborator_Success_MinimalFields(t *testing.T) {
	service, _, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContextWithUserID("USER_12345")

	request := &models.CreateCollaboratorRequest{
		CompanyName:   "Minimal Company",
		ContactPerson: "Jane Smith",
		ContactNumber: "9876543210",
		GSTNumber:     "29GGGGG1314R9ZB",
	}

	response, err := service.CreateCollaborator(ctx, request, "ORG_001", "USER_12345", "fake-jwt-token")

	testutils.AssertNoError(t, err, "Should succeed with minimal required fields")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.CompanyName, "Minimal Company", "Company name should match")
	testutils.AssertEqual(t, response.ContactPerson, "Jane Smith", "Contact person should match")
	testutils.AssertEqual(t, response.ContactNumber, "9876543210", "Contact number should match")
	testutils.AssertEqual(t, response.IsActive, true, "Should be active by default")
}

func TestCollaboratorService_CreateCollaborator_Success_AllFields(t *testing.T) {
	service, _, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContextWithUserID("USER_12345")

	email := "contact@example.com"
	gstNumber := "29GGGGG1314R9Z6"
	panNumber := "ABCDE1234F"
	accountNo := "123456789012"
	ifsc := "IFSC0001234"
	bankName := "State Bank of India"
	experience := "5 years"

	request := &models.CreateCollaboratorRequest{
		CompanyName:   "Complete Company",
		ContactPerson: "John Doe",
		ContactNumber: "9876543210",
		Email:         &email,
		GSTNumber:     gstNumber,
		PANNumber:     &panNumber,
		BankAccountNo: &accountNo,
		BankIFSC:      &ifsc,
		BankName:      &bankName,
		Experience:    &experience,
	}

	response, err := service.CreateCollaborator(ctx, request, "ORG_001", "USER_12345", "fake-jwt-token")

	testutils.AssertNoError(t, err, "Should succeed with all fields")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.CompanyName, "Complete Company", "Company name should match")
	testutils.AssertEqual(t, *response.Email, email, "Email should match")
	testutils.AssertEqual(t, response.GSTNumber, gstNumber, "GST number should match")
	testutils.AssertEqual(t, *response.PANNumber, panNumber, "PAN number should match")
	testutils.AssertEqual(t, *response.BankAccountNo, accountNo, "Bank account should match")
	testutils.AssertEqual(t, *response.BankIFSC, ifsc, "IFSC should match")
	testutils.AssertEqual(t, *response.BankName, bankName, "Bank name should match")
	testutils.AssertEqual(t, *response.Experience, experience, "Experience should match")
}

func TestCollaboratorService_GetCollaborator_Success(t *testing.T) {
	service, db, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContext()

	// Create collaborator directly in database
	collaborator := testutils.FixtureCollaborator("Get Test Company")
	err := db.Create(collaborator).Error
	testutils.AssertNoError(t, err, "Should create collaborator in database")

	// Retrieve collaborator
	response, err := service.GetCollaborator(ctx, collaborator.ID, "fake-jwt-token")

	testutils.AssertNoError(t, err, "Should retrieve collaborator successfully")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.ID, collaborator.ID, "ID should match")
	testutils.AssertEqual(t, response.CompanyName, "Get Test Company", "Company name should match")
}

func TestCollaboratorService_GetCollaborator_NotFound(t *testing.T) {
	service, _, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContext()

	// Try to get non-existent collaborator
	_, err := service.GetCollaborator(ctx, "CLAB_NONEXISTENT", "fake-jwt-token")

	testutils.AssertError(t, err, "Should fail for non-existent collaborator")
	testutils.AssertContains(t, err.Error(), "not found", "Error should mention not found")
}

func TestCollaboratorService_GetAllCollaborators_Success(t *testing.T) {
	service, db, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContext()

	// Create multiple collaborators with unique IDs
	collab1 := testutils.FixtureCollaboratorWithID("COLLAB-001", "Company A")
	collab2 := testutils.FixtureCollaboratorWithID("COLLAB-002", "Company B")
	collab3 := testutils.FixtureCollaboratorWithID("COLLAB-003", "Company C")

	err := db.Create(collab1).Error
	testutils.AssertNoError(t, err, "Should create collaborator 1")
	err = db.Create(collab2).Error
	testutils.AssertNoError(t, err, "Should create collaborator 2")
	err = db.Create(collab3).Error
	testutils.AssertNoError(t, err, "Should create collaborator 3")

	// Get all collaborators
	collabList, total, err := service.GetAllCollaborators(ctx, "fake-jwt-token", 10, 0)

	testutils.AssertNoError(t, err, "Should retrieve all collaborators")
	testutils.AssertTrue(t, len(collabList) >= 3, "Should have at least 3 collaborators")
	testutils.AssertTrue(t, total >= 3, "Total count should be at least 3")
}

func TestCollaboratorService_UpdateCollaborator_Success(t *testing.T) {
	service, db, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContextWithUserID("USER_12345")

	// Create initial collaborator
	collaborator := testutils.FixtureCollaborator("Original Company")
	err := db.Create(collaborator).Error
	testutils.AssertNoError(t, err, "Should create initial collaborator")

	// Update collaborator
	newName := "Updated Company"
	newEmail := "updated@example.com"
	request := &models.UpdateCollaboratorRequest{
		CompanyName: &newName,
		Email:       &newEmail,
	}

	response, err := service.UpdateCollaborator(ctx, collaborator.ID, request, "ORG_001", "USER_12345", "fake-jwt-token")

	testutils.AssertNoError(t, err, "Should update collaborator successfully")
	testutils.AssertNotNil(t, response, "Response should not be nil")
	testutils.AssertEqual(t, response.CompanyName, newName, "Company name should be updated")
	testutils.AssertEqual(t, *response.Email, newEmail, "Email should be updated")
}

func TestCollaboratorService_UpdateCollaborator_NotFound(t *testing.T) {
	service, _, cleanup := setupCollaboratorService(t)
	defer cleanup()

	ctx := testutils.CreateTestContextWithUserID("USER_12345")

	// Try to update non-existent collaborator
	newName := "Updated Company"
	request := &models.UpdateCollaboratorRequest{
		CompanyName: &newName,
	}

	_, err := service.UpdateCollaborator(ctx, "CLAB_NONEXISTENT", request, "ORG_001", "USER_12345", "fake-jwt-token")

	testutils.AssertError(t, err, "Should fail for non-existent collaborator")
	testutils.AssertContains(t, err.Error(), "not found", "Error should mention not found")
}
