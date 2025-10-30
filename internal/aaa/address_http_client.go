package aaa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AddressHTTPClient is an HTTP client for AAA address service
type AddressHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAddressHTTPClient creates a new HTTP address client
func NewAddressHTTPClient(baseURL string) *AddressHTTPClient {
	return &AddressHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateAddress creates a new address via AAA REST API
// POST /api/v2/addresses
func (c *AddressHTTPClient) CreateAddress(ctx context.Context, req *CreateAddressRequest, jwtToken string) (*Address, error) {
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
	c.httpClient.CloseIdleConnections()
	return nil
}
