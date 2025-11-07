package aaa

// ========================================
// DEPRECATED: This HTTP client is deprecated in favor of AddressGRPCClient
// for server-to-server communication. Use address_grpc_client.go instead.
// This file is kept for backward compatibility only.
// ========================================

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// AddressHTTPClient is an HTTP client for AAA address service
type AddressHTTPClient struct {
	baseURL    string
	httpClient *http.Client
	disabled   bool // If true, returns mock responses without making HTTP calls
}

// NewAddressHTTPClient creates a new HTTP address client
func NewAddressHTTPClient(baseURL string) *AddressHTTPClient {
	return &AddressHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		disabled: false,
	}
}

// NewMockAddressHTTPClient creates a mock address client for testing (no HTTP calls)
func NewMockAddressHTTPClient() *AddressHTTPClient {
	return &AddressHTTPClient{
		baseURL:    "mock://localhost",
		httpClient: nil,
		disabled:   true,
	}
}

// CreateAddress creates a new address via AAA REST API
// POST /api/v2/addresses
func (c *AddressHTTPClient) CreateAddress(ctx context.Context, req *CreateAddressRequest, jwtToken string) (*Address, error) {
	// Early return if disabled - return mock address
	if c.disabled {
		mockAddr := &Address{
			ID:          "mock-addr-" + uuid.New().String()[:8],
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
			FullAddress: stringPtr(fmt.Sprintf("%s, %s, %s",
				stringValue(req.Street), stringValue(req.District), stringValue(req.State))),
			IsPrimary: req.IsPrimary,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		return mockAddr, nil
	}

	url := fmt.Sprintf("%s/api/v2/addresses", c.baseURL)

	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+jwtToken)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("AAA service error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var aaaResp struct {
		Success bool     `json:"success"`
		Message string   `json:"message"`
		Data    *Address `json:"data"`
	}
	if err := json.Unmarshal(respBody, &aaaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !aaaResp.Success {
		return nil, fmt.Errorf("AAA service error: %s", aaaResp.Message)
	}

	return aaaResp.Data, nil
}

// GetAddress retrieves an address by ID
// GET /api/v2/addresses/{id}
func (c *AddressHTTPClient) GetAddress(ctx context.Context, addressID string, jwtToken string) (*Address, error) {
	// Early return if disabled - return mock address
	if c.disabled {
		return &Address{
			ID:          addressID,
			UserID:      "test-user-123",
			Type:        "business",
			Street:      stringPtr("Mock Street"),
			District:    stringPtr("Mock District"),
			State:       stringPtr("Mock State"),
			Country:     stringPtr("India"),
			Pincode:     stringPtr("000000"),
			FullAddress: stringPtr("Mock Street, Mock District, Mock State, India - 000000"),
			IsPrimary:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}, nil
	}

	url := fmt.Sprintf("%s/api/v2/addresses/%s", c.baseURL, addressID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+jwtToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AAA service error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var aaaResp struct {
		Success bool     `json:"success"`
		Message string   `json:"message"`
		Data    *Address `json:"data"`
	}
	if err := json.Unmarshal(respBody, &aaaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !aaaResp.Success {
		return nil, fmt.Errorf("AAA service error: %s", aaaResp.Message)
	}

	return aaaResp.Data, nil
}

// UpdateAddress updates an existing address
// PUT /api/v2/addresses/{id}
func (c *AddressHTTPClient) UpdateAddress(ctx context.Context, req *UpdateAddressRequest, jwtToken string) (*Address, error) {
	// Early return if disabled - return mock updated address
	if c.disabled {
		return &Address{
			ID:          req.ID,
			UserID:      "test-user-123",
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
			FullAddress: stringPtr(fmt.Sprintf("%s, %s, %s",
				stringValue(req.Street), stringValue(req.District), stringValue(req.State))),
			IsPrimary: req.IsPrimary,
			CreatedAt: time.Now().Add(-24 * time.Hour), // Mock created time
			UpdatedAt: time.Now(),
		}, nil
	}

	url := fmt.Sprintf("%s/api/v2/addresses/%s", c.baseURL, req.ID)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+jwtToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AAA service error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var aaaResp struct {
		Success bool     `json:"success"`
		Message string   `json:"message"`
		Data    *Address `json:"data"`
	}
	if err := json.Unmarshal(respBody, &aaaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !aaaResp.Success {
		return nil, fmt.Errorf("AAA service error: %s", aaaResp.Message)
	}

	return aaaResp.Data, nil
}

// DeleteAddress deletes an address
// DELETE /api/v2/addresses/{id}
func (c *AddressHTTPClient) DeleteAddress(ctx context.Context, addressID string, softDelete bool, jwtToken string) error {
	// Early return if disabled - mock success
	if c.disabled {
		return nil
	}

	url := fmt.Sprintf("%s/api/v2/addresses/%s?soft_delete=%t", c.baseURL, addressID, softDelete)

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+jwtToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AAA service error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var aaaResp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(respBody, &aaaResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !aaaResp.Success {
		return fmt.Errorf("AAA service error: %s", aaaResp.Message)
	}

	return nil
}

// SearchAddresses searches addresses
// GET /api/v2/addresses/search?q={query}&limit={limit}&offset={offset}
func (c *AddressHTTPClient) SearchAddresses(ctx context.Context, query string, limit, offset int, jwtToken string) ([]*Address, error) {
	// Early return if disabled - return empty list
	if c.disabled {
		return []*Address{}, nil
	}

	url := fmt.Sprintf("%s/api/v2/addresses/search?q=%s&limit=%d&offset=%d",
		c.baseURL, query, limit, offset)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+jwtToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AAA service error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var aaaResp struct {
		Success bool       `json:"success"`
		Message string     `json:"message"`
		Data    []*Address `json:"data"`
	}
	if err := json.Unmarshal(respBody, &aaaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !aaaResp.Success {
		return nil, fmt.Errorf("AAA service error: %s", aaaResp.Message)
	}

	return aaaResp.Data, nil
}

// Close is a no-op for HTTP client (implements same interface as gRPC client)
func (c *AddressHTTPClient) Close() error {
	if c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
	}
	return nil
}

// Helper functions for mock address creation
func stringPtr(s string) *string {
	return &s
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
