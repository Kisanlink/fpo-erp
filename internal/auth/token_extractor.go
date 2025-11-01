package auth

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// BearerTokenPrefix is the expected prefix for JWT tokens
	BearerTokenPrefix = "Bearer "
)

var (
	// ErrMissingAuthorizationHeader is returned when Authorization header is missing
	ErrMissingAuthorizationHeader = errors.New("missing authorization header")
	// ErrInvalidAuthorizationFormat is returned when Authorization header format is invalid
	ErrInvalidAuthorizationFormat = errors.New("invalid authorization header format")
	// ErrMissingToken is returned when token is missing
	ErrMissingToken = errors.New("missing token")
)

// ExtractTokenFromHeader extracts the JWT token from the Authorization header
func ExtractTokenFromHeader(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", ErrMissingAuthorizationHeader
	}

	// Case-insensitive check for Bearer prefix
	if len(authHeader) < len(BearerTokenPrefix) ||
		!strings.EqualFold(authHeader[:len(BearerTokenPrefix)], BearerTokenPrefix) {
		return "", ErrInvalidAuthorizationFormat
	}

	token := authHeader[len(BearerTokenPrefix):]
	// Return the token even if empty - let the middleware handle empty tokens
	return token, nil
}

// ExtractTokenFromQuery extracts the JWT token from query parameters (fallback)
func ExtractTokenFromQuery(c *gin.Context) (string, error) {
	token := c.Query("token")
	if token == "" {
		return "", ErrMissingToken
	}
	return token, nil
}

// ExtractToken tries to extract token from header first, then query as fallback
func ExtractToken(c *gin.Context) (string, error) {
	// Try header first
	if token, err := ExtractTokenFromHeader(c); err == nil {
		return token, nil
	}

	// Fallback to query parameter
	return ExtractTokenFromQuery(c)
}
