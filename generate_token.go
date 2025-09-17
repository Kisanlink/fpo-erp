package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// StoreManagerClaims represents the JWT claims for a Store Manager
type StoreManagerClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	IsValidated bool     `json:"is_validated"`
	Roles       []Role   `json:"roles"`
	Permissions []string `json:"permissions"`
	TokenType   string   `json:"token_type"`
	jwt.RegisteredClaims
}

// Role represents a user role
type Role struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	RoleID    string    `json:"role_id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func main() {
	// TODO: Replace this with your actual JWT secret from your ERP service
	// You can find this in your .env file or config
	secret := "f8d4e39c2b1a749ee07134f8792ac3b51c9eaf1e6d86d7817a9c5e11d75b3c42"

	// Current time
	now := time.Now()

	// Token expires in 7 days
	expiresAt := now.AddDate(0, 0, 7)

	// Store Manager permissions based on the matrix
	permissions := []string{
		"sale_summaries:read",
		"warehouses:read",
		"warehouses:create",
		"warehouses:update",
		"warehouses:delete",
		"inventory_batches:read",
		"inventory_batches:create",
		"inventory_batches:update",
		"inventory_batches:delete",
		"inventory_transactions:read",
		"inventory_transactions:create",
		"sku:read",
		"sku:create",
		"sku:update",
		"sku:delete",
		"sales:read",
		"returns:read",
		"attachments:read",
	}

	// Create claims
	claims := StoreManagerClaims{
		UserID:      "USER_store_manager_001",
		Username:    "store.manager@kisanlink.com",
		IsValidated: true,
		Roles: []Role{
			{
				ID:        "ROL_store_manager_001",
				UserID:    "USER_store_manager_001",
				RoleID:    "ROL_store_manager",
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		Permissions: permissions,
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "aaa-service",
			Audience:  []string{"aaa-clients"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        "jwt_store_manager_123456789abcdef",
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Printf("Error signing token: %v\n", err)
		return
	}

	fmt.Println("=== Store Manager JWT Token (Valid for 7 days) ===")
	fmt.Println()
	fmt.Println("Token:")
	fmt.Println(tokenString)
	fmt.Println()
	fmt.Println("=== Token Details ===")
	fmt.Printf("User ID: %s\n", claims.UserID)
	fmt.Printf("Username: %s\n", claims.Username)
	fmt.Printf("Role: %s\n", claims.Roles[0].RoleID)
	fmt.Printf("Issued At: %s\n", now.Format("2006-01-02 15:04:05 UTC"))
	fmt.Printf("Expires At: %s\n", expiresAt.Format("2006-01-02 15:04:05 UTC"))
	fmt.Printf("Total Permissions: %d\n", len(permissions))
	fmt.Println()
	fmt.Println("=== Permissions ===")
	for i, perm := range permissions {
		fmt.Printf("%2d. %s\n", i+1, perm)
	}
	fmt.Println()
	fmt.Println("=== Usage Example ===")
	fmt.Println("curl -X GET http://localhost:8080/api/v1/warehouses \\")
	fmt.Printf("  -H \"Authorization: Bearer %s\" \\\n", tokenString)
	fmt.Println("  -H \"Content-Type: application/json\"")
	fmt.Println()
	fmt.Println("=== Instructions ===")
	fmt.Println("1. Replace 'REPLACE_WITH_YOUR_ACTUAL_JWT_SECRET' with your actual JWT secret")
	fmt.Println("2. Run: go run generate_token.go")
	fmt.Println("3. Copy the generated token and use it in your API requests")
}
