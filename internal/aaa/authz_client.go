package aaa

import (
	"context"
	"fmt"
	"time"

	"kisanlink-erp/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AuthzClient wraps the gRPC authorization client
type AuthzClient struct {
	conn   *grpc.ClientConn
	client proto.AuthorizationServiceClient
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

	client := proto.NewAuthorizationServiceClient(conn)

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
	// Create the permission check request
	req := &proto.CheckRequest{
		PrincipalId:  userID,       // The user asking for permission
		ResourceType: resourceType, // What type of resource (e.g., "aaa/user")
		ResourceId:   resourceID,   // Which specific resource (e.g., "USER_123" or "*")
		Action:       action,       // What action (e.g., "read", "create", "delete")
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
	var items []*proto.CheckItem

	// Convert our permissions to gRPC format
	for i, perm := range permissions {
		item := &proto.CheckItem{
			RequestId:    fmt.Sprintf("req_%d", i),
			PrincipalId:  userID,
			ResourceType: perm.ResourceType,
			ResourceId:   perm.ResourceID,
			Action:       perm.Action,
		}
		items = append(items, item)
	}

	req := &proto.BatchCheckRequest{
		Items: items,
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

// GetUserPermissions gets all permissions for a user
func (c *AuthzClient) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	req := &proto.GetUserPermissionsRequest{
		UserId: userID,
	}

	resp, err := c.client.GetUserPermissions(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return resp.Permissions, nil
}

// GetUserRoles gets all roles for a user
func (c *AuthzClient) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	req := &proto.GetUserRolesRequest{
		UserId: userID,
	}

	resp, err := c.client.GetUserRoles(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return resp.Roles, nil
}

// Permission represents a permission to check
type Permission struct {
	ResourceType string // e.g., "aaa/user"
	ResourceID   string // e.g., "USER_123" or "*"
	Action       string // e.g., "read", "create", "delete"
}
