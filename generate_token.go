package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// StoreManagerClaims represents the JWT claims for a Store Manager (compatible with new AAA format)
type StoreManagerClaims struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	IsValidated bool   `json:"isvalidate"` // Updated field name
	RoleIDs     []Role `json:"roleIds"`    // Updated field name
	jwt.RegisteredClaims
}

// SimpleRole represents the role definition with permissions
type SimpleRole struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Scope       string      `json:"scope"`
	Permissions interface{} `json:"permissions"`
}

// Role represents a user role (matching AAA structure)
type Role struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy string     `json:"created_by"`
	UpdatedBy string     `json:"updated_by"`
	UserID    string     `json:"UserID"`
	RoleID    string     `json:"RoleID"`
	IsActive  bool       `json:"is_active"`
	Role      SimpleRole `json:"role"`
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
		RoleIDs: []Role{
			{
				ID:        "USERROLE_store_manager_001",
				CreatedAt: now,
				UpdatedAt: now,
				CreatedBy: "",
				UpdatedBy: "",
				UserID:    "USER_store_manager_001",
				RoleID:    "ROL_store_manager",
				IsActive:  true,
				Role: SimpleRole{
					Name:        "Store Manager",
					Description: "Full access store manager role",
					Scope:       "store",
					Permissions: permissions, // Include the permissions array
				},
			},
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "aaa-service",
			Audience:  []string{"aaa-frontend"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   "USER_store_manager_001",
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
	fmt.Printf("Role: %s\n", claims.RoleIDs[0].RoleID)
	fmt.Printf("IsValidated: %t\n", claims.IsValidated)
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
