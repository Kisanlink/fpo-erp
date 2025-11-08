package testutils

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// StringSlice wraps []string for SQLite JSON serialization compatibility.
// SQLite stores JSON as TEXT, so we need custom serializer to convert
// []string <-> JSON string without GORM errors.
type StringSlice []string

// Value implements driver.Valuer interface for writing to SQLite.
// Converts []string to JSON string representation for TEXT column.
func (s StringSlice) Value() (driver.Value, error) {
	// Handle nil slice - SQLite TEXT column needs valid JSON
	if s == nil {
		return "[]", nil
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal string slice: %w", err)
	}

	// Return as string for SQLite TEXT column
	return string(jsonBytes), nil
}

// Scan implements sql.Scanner interface for reading from SQLite.
// Converts SQLite TEXT (JSON string) back to []string.
func (s *StringSlice) Scan(value interface{}) error {
	// Handle NULL values
	if value == nil {
		*s = []string{}
		return nil
	}

	// Convert to string based on value type
	var source string
	switch v := value.(type) {
	case string:
		source = v
	case []byte:
		source = string(v)
	default:
		return errors.New("incompatible type for StringSlice - expected string or []byte")
	}

	// Handle empty string
	if source == "" {
		*s = []string{}
		return nil
	}

	// Unmarshal JSON string to []string
	var result []string
	if err := json.Unmarshal([]byte(source), &result); err != nil {
		return fmt.Errorf("failed to unmarshal string slice: %w", err)
	}

	*s = StringSlice(result)
	return nil
}

// JSONStringSliceSerializer implements GORM serializer interface for []string in SQLite
type JSONStringSliceSerializer struct{}

// Scan implements serializer interface for reading from database
func (JSONStringSliceSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	fieldValue := reflect.New(field.FieldType)

	if dbValue != nil {
		var bytes []byte
		switch v := dbValue.(type) {
		case []byte:
			bytes = v
		case string:
			bytes = []byte(v)
		default:
			return fmt.Errorf("failed to unmarshal JSONB value: %#v", dbValue)
		}

		if len(bytes) > 0 {
			err = json.Unmarshal(bytes, fieldValue.Interface())
		}
	}

	field.ReflectValueOf(ctx, dst).Set(fieldValue.Elem())
	return
}

// Value implements serializer interface for writing to database
func (JSONStringSliceSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	result, err := json.Marshal(fieldValue)
	if err != nil {
		return nil, err
	}
	return string(result), nil
}

// RegisterSQLiteJSONSerializer registers a custom JSON serializer for []string fields
// to ensure SQLite compatibility. This is only needed for tests.
func RegisterSQLiteJSONSerializer(db *gorm.DB) {
	// Register the JSON serializer for fields with serializer:json tag
	schema.RegisterSerializer("json", JSONStringSliceSerializer{})
}
