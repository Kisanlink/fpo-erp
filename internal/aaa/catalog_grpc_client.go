package aaa

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// CatalogGRPCClient wraps the gRPC catalog client used for role/permission seeding.
type CatalogGRPCClient struct {
	conn   *grpc.ClientConn
	client pb.CatalogServiceClient
}

// apiKeyInterceptor creates a unary interceptor that adds x-api-key to all requests
func apiKeyInterceptor(apiKey string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Only add x-api-key if it's configured
		if apiKey != "" {
			// Get existing metadata or create new
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.New(nil)
			} else {
				// Clone to avoid modifying the original
				md = md.Copy()
			}

			// Add x-api-key header
			md.Set("x-api-key", apiKey)

			// Create new context with updated metadata
			ctx = metadata.NewOutgoingContext(ctx, md)
			log.Printf("AAA Catalog Client: Added x-api-key to request for method: %s", method)
		}

		// Call the actual RPC
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// NewCatalogGRPCClient creates a new catalog gRPC client targeting the provided address.
// apiKey is optional - if provided, it will be used for service-to-service authentication.
func NewCatalogGRPCClient(address string, apiKey string, useTLS bool) (*CatalogGRPCClient, error) {
	if address == "" {
		return nil, fmt.Errorf("catalog gRPC address is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create dial options
	var opts []grpc.DialOption
	if useTLS {
		// Use system CA pool for TLS
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	opts = append(opts, grpc.WithBlock())

	// Add unary interceptor for x-api-key if configured
	if apiKey != "" {
		log.Printf("AAA Catalog Client: x-api-key configured for service authentication")
		opts = append(opts, grpc.WithUnaryInterceptor(apiKeyInterceptor(apiKey)))
	} else {
		log.Printf("AAA Catalog Client: Warning - no x-api-key configured, service-to-service auth may fail")
	}

	conn, err := grpc.DialContext(ctx, address, opts...)
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
// serviceID specifies which seed provider to use (e.g., "erp-module", "farmers-module").
// If empty, defaults to farmers-module for backward compatibility.
// force determines whether to re-seed even if data already exists.
func (c *CatalogGRPCClient) SeedRolesAndPermissions(ctx context.Context, serviceID string, force bool) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("catalog client is not initialized")
	}

	// Note: Authentication is handled via x-api-key interceptor set up in NewCatalogGRPCClient
	// No need to add metadata here - the API key is automatically added to all requests

	resp, err := c.client.SeedRolesAndPermissions(ctx, &pb.SeedRolesAndPermissionsRequest{
		ServiceId: serviceID,
		Force:     force,
	})
	if err != nil {
		return fmt.Errorf("AAA catalog seed call failed: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("AAA catalog seed returned status %d: %s", resp.StatusCode, resp.Message)
	}

	log.Printf("AAA catalog seed (service=%s): roles=%d permissions=%d resources=%d actions=%d",
		serviceID, resp.RolesCreated, resp.PermissionsCreated, resp.ResourcesCreated, resp.ActionsCreated)
	return nil
}
