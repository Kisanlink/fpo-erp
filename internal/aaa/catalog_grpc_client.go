package aaa

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CatalogGRPCClient wraps the gRPC catalog client used for role/permission seeding.
type CatalogGRPCClient struct {
	conn   *grpc.ClientConn
	client pb.CatalogServiceClient
}

// NewCatalogGRPCClient creates a new catalog gRPC client targeting the provided address.
func NewCatalogGRPCClient(address string) (*CatalogGRPCClient, error) {
	if address == "" {
		return nil, fmt.Errorf("catalog gRPC address is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AAA catalog service: %w", err)
	}

	return &CatalogGRPCClient{
		conn:   conn,
		client: pb.NewCatalogServiceClient(conn),
	}, nil
}

// Close closes the underlying gRPC connection.
func (c *CatalogGRPCClient) Close() error {
	if c != nil && c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SeedRolesAndPermissions triggers the AAA catalog seeding routine.
func (c *CatalogGRPCClient) SeedRolesAndPermissions(ctx context.Context) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("catalog client is not initialized")
	}

	resp, err := c.client.SeedRolesAndPermissions(ctx, &pb.SeedRolesAndPermissionsRequest{
		Force: false,
	})
	if err != nil {
		return fmt.Errorf("AAA catalog seed call failed: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("AAA catalog seed returned status %d: %s", resp.StatusCode, resp.Message)
	}

	log.Printf("AAA catalog seed: roles=%d permissions=%d resources=%d actions=%d",
		resp.RolesCreated, resp.PermissionsCreated, resp.ResourcesCreated, resp.ActionsCreated)
	return nil
}
