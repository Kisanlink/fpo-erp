package aaa

import (
	"context"
	"fmt"
	"time"

	pbv2 "github.com/Kisanlink/aaa-service/v2/pkg/proto/v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// AddressGRPCClient wraps the gRPC address client
type AddressGRPCClient struct {
	conn   *grpc.ClientConn
	client pbv2.AddressServiceClient
}

// NewAddressGRPCClient creates a new address gRPC client
func NewAddressGRPCClient(aaaServiceAddr string) (*AddressGRPCClient, error) {
	// Create connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, aaaServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AAA address service: %w", err)
	}

	client := pbv2.NewAddressServiceClient(conn)

	return &AddressGRPCClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the gRPC connection
func (c *AddressGRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CreateAddress creates a new address via AAA gRPC
func (c *AddressGRPCClient) CreateAddress(ctx context.Context, req *CreateAddressRequest, jwtToken string) (*Address, error) {
	// Add JWT token to metadata
	md := metadata.Pairs("authorization", "Bearer "+jwtToken)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Convert request to proto format
	protoReq := &pbv2.CreateAddressRequest{
		UserId:    req.UserID,
		Type:      req.Type,
		IsPrimary: req.IsPrimary,
	}

	// Set optional fields
	if req.House != nil {
		protoReq.House = *req.House
	}
	if req.Street != nil {
		protoReq.Street = *req.Street
	}
	if req.Landmark != nil {
		protoReq.Landmark = *req.Landmark
	}
	if req.PostOffice != nil {
		protoReq.PostOffice = *req.PostOffice
	}
	if req.Subdistrict != nil {
		protoReq.Subdistrict = *req.Subdistrict
	}
	if req.District != nil {
		protoReq.District = *req.District
	}
	if req.VTC != nil {
		protoReq.Vtc = *req.VTC
	}
	if req.State != nil {
		protoReq.State = *req.State
	}
	if req.Country != nil {
		protoReq.Country = *req.Country
	}
	if req.Pincode != nil {
		protoReq.Pincode = *req.Pincode
	}
	if len(req.Metadata) > 0 {
		protoReq.Metadata = make(map[string]string, len(req.Metadata))
		for k, v := range req.Metadata {
			if strVal, ok := v.(string); ok {
				protoReq.Metadata[k] = strVal
			}
		}
	}

	// Call gRPC service
	resp, err := c.client.CreateAddress(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("gRPC CreateAddress failed: %w", err)
	}

	// Check status code (2xx is success)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("AAA service error (status %d): %s", resp.StatusCode, resp.Message)
	}

	if resp.Address == nil {
		return nil, fmt.Errorf("AAA service returned success but no address data")
	}

	// Convert proto response to Address
	return protoAddressToAddress(resp.Address), nil
}

// GetAddress retrieves an address by ID
func (c *AddressGRPCClient) GetAddress(ctx context.Context, addressID string, jwtToken string) (*Address, error) {
	// Add JWT token to metadata
	md := metadata.Pairs("authorization", "Bearer "+jwtToken)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Call gRPC service
	resp, err := c.client.GetAddress(ctx, &pbv2.GetAddressRequest{
		Id: addressID,
	})
	if err != nil {
		return nil, fmt.Errorf("gRPC GetAddress failed: %w", err)
	}

	// Check status code (2xx is success)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("AAA service error (status %d): %s", resp.StatusCode, resp.Message)
	}

	if resp.Address == nil {
		return nil, fmt.Errorf("AAA service returned success but no address data")
	}

	return protoAddressToAddress(resp.Address), nil
}

// UpdateAddress updates an existing address
func (c *AddressGRPCClient) UpdateAddress(ctx context.Context, req *UpdateAddressRequest, jwtToken string) (*Address, error) {
	// Add JWT token to metadata
	md := metadata.Pairs("authorization", "Bearer "+jwtToken)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Convert request to proto format
	protoReq := &pbv2.UpdateAddressRequest{
		Id:        req.ID,
		Type:      req.Type,
		IsPrimary: req.IsPrimary,
		IsActive:  req.IsActive,
	}

	// Set optional fields
	if req.House != nil {
		protoReq.House = *req.House
	}
	if req.Street != nil {
		protoReq.Street = *req.Street
	}
	if req.Landmark != nil {
		protoReq.Landmark = *req.Landmark
	}
	if req.PostOffice != nil {
		protoReq.PostOffice = *req.PostOffice
	}
	if req.Subdistrict != nil {
		protoReq.Subdistrict = *req.Subdistrict
	}
	if req.District != nil {
		protoReq.District = *req.District
	}
	if req.VTC != nil {
		protoReq.Vtc = *req.VTC
	}
	if req.State != nil {
		protoReq.State = *req.State
	}
	if req.Country != nil {
		protoReq.Country = *req.Country
	}
	if req.Pincode != nil {
		protoReq.Pincode = *req.Pincode
	}
	if len(req.Metadata) > 0 {
		protoReq.Metadata = make(map[string]string, len(req.Metadata))
		for k, v := range req.Metadata {
			if strVal, ok := v.(string); ok {
				protoReq.Metadata[k] = strVal
			}
		}
	}

	// Call gRPC service
	resp, err := c.client.UpdateAddress(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("gRPC UpdateAddress failed: %w", err)
	}

	// Check status code (2xx is success)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("AAA service error (status %d): %s", resp.StatusCode, resp.Message)
	}

	if resp.Address == nil {
		return nil, fmt.Errorf("AAA service returned success but no address data")
	}

	return protoAddressToAddress(resp.Address), nil
}

// DeleteAddress deletes an address
func (c *AddressGRPCClient) DeleteAddress(ctx context.Context, addressID string, softDelete bool, jwtToken string) error {
	// Add JWT token to metadata
	md := metadata.Pairs("authorization", "Bearer "+jwtToken)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Call gRPC service
	resp, err := c.client.DeleteAddress(ctx, &pbv2.DeleteAddressRequest{
		Id:         addressID,
		SoftDelete: softDelete,
	})
	if err != nil {
		return fmt.Errorf("gRPC DeleteAddress failed: %w", err)
	}

	// Check status code (2xx is success)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("AAA service error (status %d): %s", resp.StatusCode, resp.Message)
	}

	return nil
}

// SearchAddresses searches for addresses
func (c *AddressGRPCClient) SearchAddresses(ctx context.Context, query string, limit, offset int, jwtToken string) ([]*Address, error) {
	// Add JWT token to metadata
	md := metadata.Pairs("authorization", "Bearer "+jwtToken)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Determine pagination parameters for AAA list API
	pageSize := int32(limit)
	if pageSize <= 0 {
		pageSize = 20
	}
	page := int32(1)
	if offset > 0 && limit > 0 {
		page = int32(offset/int(limit)) + 1
	}

	// Call gRPC service
	resp, err := c.client.ListAddresses(ctx, &pbv2.ListAddressesRequest{
		Search:   query,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("gRPC SearchAddresses failed: %w", err)
	}

	// Check status code (2xx is success)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("AAA service error (status %d): %s", resp.StatusCode, resp.Message)
	}

	// Return empty slice if no addresses
	if resp.Addresses == nil {
		return []*Address{}, nil
	}

	// Convert proto addresses to Address slice
	addresses := make([]*Address, len(resp.Addresses))
	for i, protoAddr := range resp.Addresses {
		addresses[i] = protoAddressToAddress(protoAddr)
	}

	return addresses, nil
}

// ========================================
// Helper Functions
// ========================================

// protoAddressToAddress converts proto.Address to aaa.Address
func protoAddressToAddress(protoAddr *pbv2.Address) *Address {
	if protoAddr == nil {
		return nil
	}

	addr := &Address{
		ID:        protoAddr.GetId(),
		UserID:    protoAddr.GetUserId(),
		Type:      protoAddr.GetType(),
		IsPrimary: protoAddr.GetIsPrimary(),
		IsActive:  protoAddr.GetIsActive(),
	}

	// Convert optional fields
	addr.House = stringPtrIfNotEmpty(protoAddr.GetHouse())
	addr.Street = stringPtrIfNotEmpty(protoAddr.GetStreet())
	addr.Landmark = stringPtrIfNotEmpty(protoAddr.GetLandmark())
	addr.PostOffice = stringPtrIfNotEmpty(protoAddr.GetPostOffice())
	addr.Subdistrict = stringPtrIfNotEmpty(protoAddr.GetSubdistrict())
	addr.District = stringPtrIfNotEmpty(protoAddr.GetDistrict())
	addr.VTC = stringPtrIfNotEmpty(protoAddr.GetVtc())
	addr.State = stringPtrIfNotEmpty(protoAddr.GetState())
	addr.Country = stringPtrIfNotEmpty(protoAddr.GetCountry())
	addr.Pincode = stringPtrIfNotEmpty(protoAddr.GetPincode())
	addr.FullAddress = stringPtrIfNotEmpty(protoAddr.GetFullAddress())

	// Convert timestamps
	if protoAddr.GetCreatedAt() != nil {
		addr.CreatedAt = protoAddr.GetCreatedAt().AsTime()
	}
	if protoAddr.GetUpdatedAt() != nil {
		addr.UpdatedAt = protoAddr.GetUpdatedAt().AsTime()
	}

	// Build full address if not provided
	if addr.FullAddress == nil || *addr.FullAddress == "" {
		fullAddr := addr.BuildFullAddress()
		addr.FullAddress = &fullAddr
	}

	// Convert metadata
	if meta := protoAddr.GetMetadata(); len(meta) > 0 {
		addr.Metadata = make(map[string]interface{}, len(meta))
		for k, v := range meta {
			addr.Metadata[k] = v
		}
	}

	return addr
}

// stringPtrIfNotEmpty returns a pointer to the string if it is not empty, otherwise nil
func stringPtrIfNotEmpty(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}
