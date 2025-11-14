package aaa

import (
	"context"
	"fmt"
	"time"

	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// AuthzClient wraps the gRPC authorization client
type AuthzClient struct {
	conn   *grpc.ClientConn
	client pb.AuthorizationServiceClient
}

// NewAuthzClient creates a new authorization client
func NewAuthzClient(aaaServiceAddr string) (*AuthzClient, error) {
	// Create connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, aaaServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AAA service: %w", err)
	}

	client := pb.NewAuthorizationServiceClient(conn)

	return &AuthzClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the gRPC connection
func (c *AuthzClient) Close() error {
	return c.conn.Close()
}

// CheckPermission checks if a user has permission to perform an action on a resource
func (c *AuthzClient) CheckPermission(ctx context.Context, userID, resourceType, resourceID, action string) (bool, error) {
	return c.CheckPermissionWithToken(ctx, userID, resourceType, resourceID, action, "")
}

// CheckPermissionWithToken checks if a user has permission to perform an action on a resource with JWT token
func (c *AuthzClient) CheckPermissionWithToken(ctx context.Context, userID, resourceType, resourceID, action, jwtToken string) (bool, error) {
	// Create the permission check request
	req := &pb.CheckRequest{
		PrincipalId:  userID,       // The user asking for permission
		ResourceType: resourceType, // What type of resource (e.g., "user")
		ResourceId:   resourceID,   // Which specific resource (e.g., "USER_123" or "*")
		Action:       action,       // What action (e.g., "read", "create", "delete")
	}

	// Add authorization token to context if provided
	if jwtToken != "" {
		ctx = c.addAuthTokenToContext(ctx, jwtToken)
	}

	// Send request to AAA service
	resp, err := c.client.Check(ctx, req)
	if err != nil {
		return false, fmt.Errorf("permission check failed: %w", err)
	}

	// Return whether permission is granted
	return resp.Allowed, nil
}

// CheckMultiplePermissions checks multiple permissions in a single request
func (c *AuthzClient) CheckMultiplePermissions(ctx context.Context, userID string, permissions []Permission) (map[string]bool, error) {
	return c.CheckMultiplePermissionsWithToken(ctx, userID, permissions, "")
}

// CheckMultiplePermissionsWithToken checks multiple permissions in a single request with JWT token
func (c *AuthzClient) CheckMultiplePermissionsWithToken(ctx context.Context, userID string, permissions []Permission, jwtToken string) (map[string]bool, error) {
	var items []*pb.CheckItem

	// Convert our permissions to gRPC format
	for i, perm := range permissions {
		item := &pb.CheckItem{
			RequestId:    fmt.Sprintf("req_%d", i),
			PrincipalId:  userID,
			ResourceType: perm.ResourceType,
			ResourceId:   perm.ResourceID,
			Action:       perm.Action,
		}
		items = append(items, item)
	}

	req := &pb.BatchCheckRequest{
		Items: items,
	}

	// Add authorization token to context if provided
	if jwtToken != "" {
		ctx = c.addAuthTokenToContext(ctx, jwtToken)
	}

	resp, err := c.client.BatchCheck(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("batch permission check failed: %w", err)
	}

	// Convert results back to our format
	results := make(map[string]bool)
	for i, result := range resp.Results {
		key := fmt.Sprintf("%s:%s:%s", permissions[i].ResourceType, permissions[i].ResourceID, permissions[i].Action)
		results[key] = result.Allowed
	}

	return results, nil
}

// addAuthTokenToContext adds the JWT token to gRPC metadata
func (c *AuthzClient) addAuthTokenToContext(ctx context.Context, jwtToken string) context.Context {
	// Create metadata with authorization header
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + jwtToken,
	})

	// Add metadata to context
	return metadata.NewOutgoingContext(ctx, md)
}

// Permission represents a permission to check
type Permission struct {
	ResourceType string // e.g., "user"
	ResourceID   string // e.g., "USER_123" or "*"
	Action       string // e.g., "read", "create", "delete"
}

// ResourceAction represents a resource type and action pair for permission checking
// Used for building organization-scoped permissions dynamically
type ResourceAction struct {
	ResourceType string // e.g., "collaborator", "address"
	Action       string // e.g., "create", "read", "update", "delete"
}
