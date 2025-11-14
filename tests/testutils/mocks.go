package testutils

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/services"
)

// ========================================
// AAA Address Client Mock
// ========================================

// MockAAAClient is a mock implementation of AddressGRPCClient for testing
type MockAAAClient struct {
	// Mock storage for addresses
	Addresses map[string]*aaa.Address
	// Control mock behavior
	ShouldFail     bool
	FailureMessage string
}

// NewMockAAAClient creates a new mock AAA client
func NewMockAAAClient() *MockAAAClient {
	return &MockAAAClient{
		Addresses:  make(map[string]*aaa.Address),
		ShouldFail: false,
	}
}

// CreateAddress mocks address creation
func (m *MockAAAClient) CreateAddress(ctx context.Context, req *aaa.CreateAddressRequest, jwtToken string) (*aaa.Address, error) {
	if m.ShouldFail {
		return nil, fmt.Errorf("%s", m.FailureMessage)
	}

	// Generate mock address ID
	addressID := fmt.Sprintf("ADDR_%d", len(m.Addresses)+1)

	address := &aaa.Address{
		ID:          addressID,
		UserID:      req.UserID,
		Type:        req.Type,
		House:       req.House,
		Street:      req.Street,
		Landmark:    req.Landmark,
		PostOffice:  req.PostOffice,
		Subdistrict: req.Subdistrict,
		District:    req.District,
		VTC:         req.VTC,
		State:       req.State,
		Country:     req.Country,
		Pincode:     req.Pincode,
		IsPrimary:   req.IsPrimary,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Store in mock storage
	m.Addresses[addressID] = address

	return address, nil
}

// GetAddress mocks address retrieval
func (m *MockAAAClient) GetAddress(ctx context.Context, addressID string, jwtToken string) (*aaa.Address, error) {
	if m.ShouldFail {
		return nil, fmt.Errorf("%s", m.FailureMessage)
	}

	address, exists := m.Addresses[addressID]
	if !exists {
		return nil, fmt.Errorf("address not found: %s", addressID)
	}

	return address, nil
}

// UpdateAddress mocks address update
func (m *MockAAAClient) UpdateAddress(ctx context.Context, req *aaa.UpdateAddressRequest, jwtToken string) (*aaa.Address, error) {
	if m.ShouldFail {
		return nil, fmt.Errorf("%s", m.FailureMessage)
	}

	address, exists := m.Addresses[req.ID]
	if !exists {
		return nil, fmt.Errorf("address not found: %s", req.ID)
	}

	// Update fields
	address.Type = req.Type
	address.House = req.House
	address.Street = req.Street
	address.Landmark = req.Landmark
	address.PostOffice = req.PostOffice
	address.Subdistrict = req.Subdistrict
	address.District = req.District
	address.VTC = req.VTC
	address.State = req.State
	address.Country = req.Country
	address.Pincode = req.Pincode
	address.IsPrimary = req.IsPrimary
	address.IsActive = req.IsActive
	address.UpdatedAt = time.Now()

	return address, nil
}

// DeleteAddress mocks address deletion
func (m *MockAAAClient) DeleteAddress(ctx context.Context, addressID string, softDelete bool, jwtToken string) error {
	if m.ShouldFail {
		return fmt.Errorf("%s", m.FailureMessage)
	}

	if softDelete {
		// Soft delete - mark as inactive
		address, exists := m.Addresses[addressID]
		if !exists {
			return fmt.Errorf("address not found: %s", addressID)
		}
		address.IsActive = false
	} else {
		// Hard delete - remove from map
		delete(m.Addresses, addressID)
	}

	return nil
}

// SearchAddresses mocks address search
func (m *MockAAAClient) SearchAddresses(ctx context.Context, query string, limit, offset int, jwtToken string) ([]*aaa.Address, error) {
	if m.ShouldFail {
		return nil, fmt.Errorf("%s", m.FailureMessage)
	}

	var results []*aaa.Address
	for _, addr := range m.Addresses {
		// Simple search: check if query matches any field
		if strings.Contains(strings.ToLower(*addr.VTC), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(*addr.Pincode), strings.ToLower(query)) {
			results = append(results, addr)
		}
	}

	// Apply pagination
	if offset >= len(results) {
		return []*aaa.Address{}, nil
	}

	end := offset + limit
	if end > len(results) {
		end = len(results)
	}

	return results[offset:end], nil
}

// Close mocks connection closure
func (m *MockAAAClient) Close() error {
	return nil
}

// SetShouldFail configures the mock to fail
func (m *MockAAAClient) SetShouldFail(shouldFail bool, message string) {
	m.ShouldFail = shouldFail
	m.FailureMessage = message
}

// ========================================
// S3 Service Mock
// ========================================

// MockS3Service is a mock implementation of S3Service for testing
type MockS3Service struct {
	// Mock storage for files
	Files map[string][]byte
	// File metadata
	Metadata map[string]*services.FileInfo
	// Control mock behavior
	ShouldFail     bool
	FailureMessage string
}

// NewMockS3Service creates a new mock S3 service
func NewMockS3Service() *MockS3Service {
	return &MockS3Service{
		Files:      make(map[string][]byte),
		Metadata:   make(map[string]*services.FileInfo),
		ShouldFail: false,
	}
}

// UploadFile mocks file upload to S3
func (m *MockS3Service) UploadFile(ctx context.Context, file *multipart.FileHeader, entityType, entityID string) (string, error) {
	if m.ShouldFail {
		return "", fmt.Errorf("%s", m.FailureMessage)
	}

	// Generate mock S3 key
	key := fmt.Sprintf("%s/%s/%s", entityType, entityID, file.Filename)

	// Read file content (in tests, this would be mock data)
	m.Files[key] = []byte("mock file content")

	// Store metadata
	m.Metadata[key] = &services.FileInfo{
		Key:          key,
		Size:         file.Size,
		ContentType:  file.Header.Get("Content-Type"),
		LastModified: time.Now(),
		Metadata: map[string]string{
			"original-filename": file.Filename,
			"uploaded-at":       time.Now().UTC().Format(time.RFC3339),
			"entity-type":       entityType,
			"entity-id":         entityID,
		},
	}

	return key, nil
}

// DownloadFile mocks file download from S3
func (m *MockS3Service) DownloadFile(ctx context.Context, s3URL string) (io.ReadCloser, string, error) {
	if m.ShouldFail {
		return nil, "", fmt.Errorf("%s", m.FailureMessage)
	}

	// Extract key from S3 URL
	key := strings.TrimPrefix(s3URL, "s3://mock-bucket/")

	content, exists := m.Files[key]
	if !exists {
		return nil, "", fmt.Errorf("file not found: %s", key)
	}

	metadata := m.Metadata[key]
	contentType := "application/octet-stream"
	if metadata != nil {
		contentType = metadata.ContentType
	}

	return io.NopCloser(strings.NewReader(string(content))), contentType, nil
}

// DeleteFile mocks file deletion from S3
func (m *MockS3Service) DeleteFile(ctx context.Context, s3URL string) error {
	if m.ShouldFail {
		return fmt.Errorf("%s", m.FailureMessage)
	}

	// Extract key from S3 URL
	key := strings.TrimPrefix(s3URL, "s3://mock-bucket/")

	delete(m.Files, key)
	delete(m.Metadata, key)

	return nil
}

// GeneratePresignedURL mocks presigned URL generation
func (m *MockS3Service) GeneratePresignedURL(ctx context.Context, s3URL string, expiration time.Duration) (string, error) {
	if m.ShouldFail {
		return "", fmt.Errorf("%s", m.FailureMessage)
	}

	// Return mock presigned URL
	return fmt.Sprintf("https://mock-s3.amazonaws.com/%s?expires=%d", s3URL, time.Now().Add(expiration).Unix()), nil
}

// FileExists mocks file existence check
func (m *MockS3Service) FileExists(ctx context.Context, s3URL string) (bool, error) {
	if m.ShouldFail {
		return false, fmt.Errorf("%s", m.FailureMessage)
	}

	key := strings.TrimPrefix(s3URL, "s3://mock-bucket/")
	_, exists := m.Files[key]
	return exists, nil
}

// GetFileInfo mocks file info retrieval
func (m *MockS3Service) GetFileInfo(ctx context.Context, s3URL string) (*services.FileInfo, error) {
	if m.ShouldFail {
		return nil, fmt.Errorf("%s", m.FailureMessage)
	}

	key := strings.TrimPrefix(s3URL, "s3://mock-bucket/")
	info, exists := m.Metadata[key]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", key)
	}

	return info, nil
}

// ValidateFileType mocks file type validation
func (m *MockS3Service) ValidateFileType(filename string) error {
	if m.ShouldFail {
		return fmt.Errorf("%s", m.FailureMessage)
	}

	allowedExtensions := []string{".pdf", ".jpg", ".jpeg", ".png", ".gif", ".doc", ".docx", ".xls", ".xlsx"}
	for _, ext := range allowedExtensions {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			return nil
		}
	}

	return fmt.Errorf("file type not allowed: %s", filename)
}

// GetFileSize mocks file size retrieval
func (m *MockS3Service) GetFileSize(file *multipart.FileHeader) int64 {
	return file.Size
}

// GetContentType mocks content type retrieval
func (m *MockS3Service) GetContentType(file *multipart.FileHeader) string {
	return file.Header.Get("Content-Type")
}

// SetShouldFail configures the mock to fail
func (m *MockS3Service) SetShouldFail(shouldFail bool, message string) {
	m.ShouldFail = shouldFail
	m.FailureMessage = message
}

// ========================================
// Mock Database Transaction
// ========================================

// MockTransaction is a mock implementation for testing transaction scenarios
type MockTransaction struct {
	ShouldFail     bool
	FailureMessage string
	Committed      bool
	RolledBack     bool
}

// NewMockTransaction creates a new mock transaction
func NewMockTransaction() *MockTransaction {
	return &MockTransaction{
		ShouldFail: false,
		Committed:  false,
		RolledBack: false,
	}
}

// Commit mocks transaction commit
func (m *MockTransaction) Commit() error {
	if m.ShouldFail {
		return fmt.Errorf("%s", m.FailureMessage)
	}
	m.Committed = true
	return nil
}

// Rollback mocks transaction rollback
func (m *MockTransaction) Rollback() error {
	m.RolledBack = true
	return nil
}

// SetShouldFail configures the mock to fail
func (m *MockTransaction) SetShouldFail(shouldFail bool, message string) {
	m.ShouldFail = shouldFail
	m.FailureMessage = message
}

// ========================================
// Mock Repository Interfaces (Following E-commerce Pattern)
// ========================================

// These mocks use testify/mock for easy test expectations
// Import required: "github.com/stretchr/testify/mock"

// Note: Repository mocks would go here when testify/mock is added
// For now, tests can use the real database with SetupTestDB() from database.go
// This follows the e-commerce pattern of simple SQLite for integration tests
