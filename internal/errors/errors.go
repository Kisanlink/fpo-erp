package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	StatusCode int    `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Details)
}

// NewAppError creates a new application error
func NewAppError(code int, message, details string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: http.StatusInternalServerError,
	}
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:       400,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       404,
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return &AppError{
		Code:       422,
		Message:    message,
		StatusCode: http.StatusUnprocessableEntity,
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:       401,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:       403,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:       409,
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

// NewInternalServerError creates an internal server error
func NewInternalServerError(message string) *AppError {
	return &AppError{
		Code:       500,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
	}
}

// Common error messages
const (
	ErrInvalidID           = "Invalid ID format"
	ErrInvalidRequest      = "Invalid request data"
	ErrDatabaseConnection  = "Database connection failed"
	ErrDatabaseQuery       = "Database query failed"
	ErrDatabaseTransaction = "Database transaction failed"
	ErrValidationFailed    = "Validation failed"
	ErrUnauthorized        = "Unauthorized access"
	ErrForbidden           = "Access forbidden"
	ErrNotFound            = "Resource not found"
	ErrConflict            = "Resource conflict"
	ErrInternalServer      = "Internal server error"
)
