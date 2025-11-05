package aaa

import (
	"context"
	"fmt"
	"time"

	"kisanlink-erp/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// AddressGRPCClient wraps the gRPC address client
type AddressGRPCClient struct {
	conn   *grpc.ClientConn
	client proto.AddressServiceClient
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

	client := proto.NewAddressServiceClient(conn)

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
	protoReq := &proto.CreateAddressRequest{
		UserId:    req.UserID,
		Type:      req.Type,
		IsPrimary: req.IsPrimary,
	}

	// Set optional fields
	if req.House != nil {
		protoReq.House = req.House
	}
	if req.Street != nil {
		protoReq.Street = req.Street
	}
	if req.Landmark != nil {
		protoReq.Landmark = req.Landmark
	}
	if req.PostOffice != nil {
		protoReq.PostOffice = req.PostOffice
	}
	if req.Subdistrict != nil {
		protoReq.Subdistrict = req.Subdistrict
	}
	if req.District != nil {
		protoReq.District = req.District
	}
	if req.VTC != nil {
		protoReq.Vtc = req.VTC
	}
	if req.State != nil {
		protoReq.State = req.State
	}
	if req.Country != nil {
		protoReq.Country = req.Country
	}
	if req.Pincode != nil {
		protoReq.Pincode = req.Pincode
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
	resp, err := c.client.GetAddress(ctx, &proto.GetAddressRequest{
		AddressId: addressID,
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
	protoReq := &proto.UpdateAddressRequest{
		Id:        req.ID,
		Type:      req.Type,
		IsPrimary: req.IsPrimary,
		IsActive:  req.IsActive,
	}

	// Set optional fields
	if req.House != nil {
		protoReq.House = req.House
	}
	if req.Street != nil {
		protoReq.Street = req.Street
	}
	if req.Landmark != nil {
		protoReq.Landmark = req.Landmark
	}
	if req.PostOffice != nil {
		protoReq.PostOffice = req.PostOffice
	}
	if req.Subdistrict != nil {
		protoReq.Subdistrict = req.Subdistrict
	}
	if req.District != nil {
		protoReq.District = req.District
	}
	if req.VTC != nil {
		protoReq.Vtc = req.VTC
	}
	if req.State != nil {
		protoReq.State = req.State
	}
	if req.Country != nil {
		protoReq.Country = req.Country
	}
	if req.Pincode != nil {
		protoReq.Pincode = req.Pincode
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
	resp, err := c.client.DeleteAddress(ctx, &proto.DeleteAddressRequest{
		AddressId:  addressID,
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

	// Call gRPC service
	resp, err := c.client.SearchAddresses(ctx, &proto.SearchAddressesRequest{
		Query:  query,
		Limit:  int32(limit),
		Offset: int32(offset),
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
func protoAddressToAddress(protoAddr *proto.Address) *Address {
	if protoAddr == nil {
		return nil
	}

	addr := &Address{
		ID:        protoAddr.Id,
		UserID:    protoAddr.UserId,
		Type:      protoAddr.Type,
		IsPrimary: protoAddr.IsPrimary,
		IsActive:  protoAddr.IsActive,
	}

	// Convert optional fields
	if protoAddr.House != nil {
		addr.House = protoAddr.House
	}
	if protoAddr.Street != nil {
		addr.Street = protoAddr.Street
	}
	if protoAddr.Landmark != nil {
		addr.Landmark = protoAddr.Landmark
	}
	if protoAddr.PostOffice != nil {
		addr.PostOffice = protoAddr.PostOffice
	}
	if protoAddr.Subdistrict != nil {
		addr.Subdistrict = protoAddr.Subdistrict
	}
	if protoAddr.District != nil {
		addr.District = protoAddr.District
	}
	if protoAddr.Vtc != nil {
		addr.VTC = protoAddr.Vtc
	}
	if protoAddr.State != nil {
		addr.State = protoAddr.State
	}
	if protoAddr.Country != nil {
		addr.Country = protoAddr.Country
	}
	if protoAddr.Pincode != nil {
		addr.Pincode = protoAddr.Pincode
	}
	if protoAddr.FullAddress != nil {
		addr.FullAddress = protoAddr.FullAddress
	}

	// Convert timestamps
	if protoAddr.CreatedAt != nil {
		addr.CreatedAt = protoAddr.CreatedAt.AsTime()
	}
	if protoAddr.UpdatedAt != nil {
		addr.UpdatedAt = protoAddr.UpdatedAt.AsTime()
	}

	// Build full address if not provided
	if addr.FullAddress == nil || *addr.FullAddress == "" {
		fullAddr := addr.BuildFullAddress()
		addr.FullAddress = &fullAddr
	}

	return addr
}
