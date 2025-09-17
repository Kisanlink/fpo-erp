package aaa

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"kisanlink-erp/pkg/proto"
)

type AddressClient struct {
	conn   *grpc.ClientConn
	client proto.AddressServiceClient
}

func NewAddressClient(aaaServiceURL string) (*AddressClient, error) {
	conn, err := grpc.Dial(aaaServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AAA service: %w", err)
	}

	client := proto.NewAddressServiceClient(conn)

	return &AddressClient{
		conn:   conn,
		client: client,
	}, nil
}

func (ac *AddressClient) Close() error {
	return ac.conn.Close()
}

// GetAddress retrieves an address by ID
func (ac *AddressClient) GetAddress(ctx context.Context, addressID string) (*Address, error) {
	req := &proto.GetAddressRequest{
		Id: addressID,
	}

	resp, err := ac.client.GetAddress(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("AAA service error: %s (code: %d)", resp.Message, resp.StatusCode)
	}

	return convertProtoAddressToAddress(resp.Address), nil
}

// CreateAddress creates a new address
func (ac *AddressClient) CreateAddress(ctx context.Context, req *CreateAddressRequest) (*Address, error) {
	protoReq := &proto.CreateAddressRequest{
		UserId:        req.UserID,
		Type:          req.Type,
		AddressLine_1: req.AddressLine1,
		AddressLine_2: req.AddressLine2,
		City:          req.City,
		State:         req.State,
		PostalCode:    req.PostalCode,
		Country:       req.Country,
		IsPrimary:     req.IsPrimary,
		Metadata:      req.Metadata,
	}

	resp, err := ac.client.CreateAddress(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("AAA service error: %s (code: %d)", resp.Message, resp.StatusCode)
	}

	return convertProtoAddressToAddress(resp.Address), nil
}

// UpdateAddress updates an existing address
func (ac *AddressClient) UpdateAddress(ctx context.Context, req *UpdateAddressRequest) (*Address, error) {
	protoReq := &proto.UpdateAddressRequest{
		Id:            req.ID,
		Type:          req.Type,
		AddressLine_1: req.AddressLine1,
		AddressLine_2: req.AddressLine2,
		City:          req.City,
		State:         req.State,
		PostalCode:    req.PostalCode,
		Country:       req.Country,
		IsPrimary:     req.IsPrimary,
		IsActive:      req.IsActive,
		Metadata:      req.Metadata,
	}

	resp, err := ac.client.UpdateAddress(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update address: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("AAA service error: %s (code: %d)", resp.Message, resp.StatusCode)
	}

	return convertProtoAddressToAddress(resp.Address), nil
}

// DeleteAddress deletes an address
func (ac *AddressClient) DeleteAddress(ctx context.Context, addressID string, softDelete bool) error {
	req := &proto.DeleteAddressRequest{
		Id:         addressID,
		SoftDelete: softDelete,
	}

	resp, err := ac.client.DeleteAddress(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete address: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("AAA service error: %s (code: %d)", resp.Message, resp.StatusCode)
	}

	return nil
}

// convertProtoAddressToAddress converts proto.Address to internal Address
func convertProtoAddressToAddress(protoAddr *proto.Address) *Address {
	if protoAddr == nil {
		return nil
	}

	return &Address{
		ID:           protoAddr.Id,
		UserID:       protoAddr.UserId,
		Type:         protoAddr.Type,
		AddressLine1: protoAddr.AddressLine_1,
		AddressLine2: protoAddr.AddressLine_2,
		City:         protoAddr.City,
		State:        protoAddr.State,
		PostalCode:   protoAddr.PostalCode,
		Country:      protoAddr.Country,
		IsPrimary:    protoAddr.IsPrimary,
		IsActive:     protoAddr.IsActive,
		Metadata:     protoAddr.Metadata,
	}
}
