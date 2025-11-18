package utils

import (
	"net/http"
	"strings"
	"time"

	"kisanlink-erp/internal/errors"

	"github.com/gin-gonic/gin"
)

// HandleServiceError intelligently maps service errors to appropriate HTTP status codes
// This function checks if the error is a custom AppError type and returns its status code,
// or uses pattern matching to determine the appropriate status code for plain errors.
//
// Usage:
//
//	response, err := h.service.CreateProduct(&request)
//	if err != nil {
//	    utils.HandleServiceError(c, "Failed to create product", err)
//	    return
//	}
func HandleServiceError(c *gin.Context, defaultMessage string, err error) {
	// Check if error is a custom AppError type
	if appErr, ok := err.(*errors.AppError); ok {
		c.JSON(appErr.StatusCode, Response{
			Success:   false,
			Message:   appErr.Message,
			Error:     appErr.Details,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	// Pattern match error message to determine appropriate status code
	errMsg := strings.ToLower(err.Error())
	statusCode := determineStatusCode(errMsg)

	// Return response with determined status code
	c.JSON(statusCode, Response{
		Success:   false,
		Message:   defaultMessage,
		Error:     err.Error(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// determineStatusCode uses pattern matching to determine the appropriate HTTP status code
// based on common error message patterns
func determineStatusCode(errMsg string) int {
	// 404 Not Found patterns
	if strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "does not exist") ||
		strings.Contains(errMsg, "no such") {
		return http.StatusNotFound
	}

	// 409 Conflict patterns
	if strings.Contains(errMsg, "already exists") ||
		strings.Contains(errMsg, "duplicate") ||
		strings.Contains(errMsg, "conflict") {
		return http.StatusConflict
	}

	// 400 Bad Request patterns
	if strings.Contains(errMsg, "required") ||
		strings.Contains(errMsg, "invalid") ||
		strings.Contains(errMsg, "must be") ||
		strings.Contains(errMsg, "cannot be") ||
		strings.Contains(errMsg, "should be") ||
		strings.Contains(errMsg, "expected") ||
		strings.Contains(errMsg, "missing") ||
		strings.Contains(errMsg, "empty") ||
		strings.Contains(errMsg, "exceeds") ||
		strings.Contains(errMsg, "out of range") ||
		strings.Contains(errMsg, "validation failed") {
		return http.StatusBadRequest
	}

	// 422 Unprocessable Entity patterns (business logic violations)
	if strings.Contains(errMsg, "insufficient") ||
		strings.Contains(errMsg, "unavailable") ||
		strings.Contains(errMsg, "status transition") ||
		strings.Contains(errMsg, "not allowed") ||
		strings.Contains(errMsg, "not permitted") {
		return http.StatusUnprocessableEntity
	}

	// 401 Unauthorized patterns
	if strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "authentication") ||
		strings.Contains(errMsg, "token") {
		return http.StatusUnauthorized
	}

	// 403 Forbidden patterns
	if strings.Contains(errMsg, "forbidden") ||
		strings.Contains(errMsg, "access denied") ||
		strings.Contains(errMsg, "permission denied") {
		return http.StatusForbidden
	}

	// Default to 500 Internal Server Error for unknown errors
	return http.StatusInternalServerError
}
