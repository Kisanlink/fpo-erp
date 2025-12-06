package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Response represents a standard API response
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// ErrorResponseModel represents an error response structure for Swagger documentation
type ErrorResponseModel struct {
	Success   bool   `json:"success" example:"false"`
	Message   string `json:"message,omitempty" example:"Error occurred"`
	Error     string `json:"error,omitempty" example:"Detailed error message"`
	Timestamp string `json:"timestamp" example:"2024-01-15T10:30:00Z"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
	RequestID  string         `json:"request_id"`
	Success    bool           `json:"success"`
	Timestamp  string         `json:"timestamp"`
}

// PaginatedResponseModel represents a paginated response structure for Swagger documentation
type PaginatedResponseModel struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
	RequestID  string         `json:"request_id" example:"34bc4e39-a3d7-4c18-99d9-3645139e3bd8"`
	Success    bool           `json:"success" example:"true"`
	Timestamp  string         `json:"timestamp" example:"2024-01-15T10:30:00Z"`
}

// SuccessResponse sends a success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	response := Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	c.JSON(statusCode, response)
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	response := Response{
		Success:   false,
		Message:   message,
		Error:     errorMsg,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	c.JSON(statusCode, response)
}

// CreatedResponse sends a 201 Created response
func CreatedResponse(c *gin.Context, message string, data interface{}) {
	SuccessResponse(c, http.StatusCreated, message, data)
}

// OKResponse sends a 200 OK response
func OKResponse(c *gin.Context, message string, data interface{}) {
	SuccessResponse(c, http.StatusOK, message, data)
}

// BadRequestResponse sends a 400 Bad Request response
func BadRequestResponse(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusBadRequest, message, err)
}

// NotFoundResponse sends a 404 Not Found response
func NotFoundResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, message, nil)
}

// InternalServerErrorResponse sends a 500 Internal Server Error response
func InternalServerErrorResponse(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusInternalServerError, message, err)
}

// UnauthorizedResponse sends a 401 Unauthorized response
func UnauthorizedResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, message, nil)
}

// ForbiddenResponse sends a 403 Forbidden response
func ForbiddenResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, message, nil)
}

// PaginatedOKResponse sends a 200 OK response with pagination metadata
func PaginatedOKResponse(c *gin.Context, data interface{}, total int64, limit, offset int) {
	response := PaginatedResponse{
		Data:       data,
		Pagination: NewPaginationMeta(limit, offset, total),
		RequestID:  uuid.New().String(),
		Success:    true,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
	c.JSON(http.StatusOK, response)
}

// PaginatedSuccessResponse sends a paginated success response with custom status code
func PaginatedSuccessResponse(c *gin.Context, statusCode int, data interface{}, total int64, limit, offset int) {
	response := PaginatedResponse{
		Data:       data,
		Pagination: NewPaginationMeta(limit, offset, total),
		RequestID:  uuid.New().String(),
		Success:    true,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
	c.JSON(statusCode, response)
}
