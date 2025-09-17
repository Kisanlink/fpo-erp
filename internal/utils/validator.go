package utils

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// InitValidator initializes the validator
func InitValidator() {
	validate = validator.New()
}

// ValidateStruct validates a struct using tags
func ValidateStruct(s interface{}) error {
	if validate == nil {
		InitValidator()
	}
	return validate.Struct(s)
}

// ValidateVar validates a single variable
func ValidateVar(field interface{}, tag string) error {
	if validate == nil {
		InitValidator()
	}
	return validate.Var(field, tag)
}

// GetValidationErrors returns formatted validation errors
func GetValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			switch e.Tag() {
			case "required":
				errors[field] = field + " is required"
			case "email":
				errors[field] = field + " must be a valid email"
			case "min":
				errors[field] = field + " must be at least " + e.Param() + " characters"
			case "max":
				errors[field] = field + " must be at most " + e.Param() + " characters"
			case "gt":
				errors[field] = field + " must be greater than " + e.Param()
			case "gte":
				errors[field] = field + " must be greater than or equal to " + e.Param()
			case "lt":
				errors[field] = field + " must be less than " + e.Param()
			case "lte":
				errors[field] = field + " must be less than or equal to " + e.Param()
			default:
				errors[field] = field + " failed validation: " + e.Tag()
			}
		}
	}

	return errors
}

// ValidateRequest validates the request body and returns errors
func ValidateRequest(c *gin.Context, request interface{}) error {
	if err := c.ShouldBindJSON(request); err != nil {
		return err
	}

	if err := ValidateStruct(request); err != nil {
		return err
	}

	return nil
}

// ValidatePartialRequest validates partial request updates
func ValidatePartialRequest(c *gin.Context, request interface{}) error {
	if err := c.ShouldBindJSON(request); err != nil {
		return err
	}

	// For partial updates, only validate non-zero fields
	v := reflect.ValueOf(request).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.IsZero() {
			// Validate individual field
			tag := t.Field(i).Tag.Get("binding")
			if tag != "" {
				if err := ValidateVar(field.Interface(), tag); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// IsValidUUID checks if a string is a valid UUID
func IsValidUUID(uuid string) bool {
	return ValidateVar(uuid, "uuid") == nil
}

// IsValidEmail checks if a string is a valid email
func IsValidEmail(email string) bool {
	return ValidateVar(email, "email") == nil
}

// ParseIntParam parses a string parameter to int with error handling
func ParseIntParam(param string) (int, error) {
	return strconv.Atoi(param)
}

// ParseInt64Param parses a string parameter to int64 with error handling
func ParseInt64Param(param string) (int64, error) {
	return strconv.ParseInt(param, 10, 64)
}

// ParseFloatParam parses a string parameter to float64 with error handling
func ParseFloatParam(param string) (float64, error) {
	return strconv.ParseFloat(param, 64)
}
