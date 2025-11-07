package testutils

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"
)

// ========================================
// Context Helpers
// ========================================

// CreateTestContext creates a context with optional JWT values
func CreateTestContext() context.Context {
	return context.Background()
}

// CreateTestContextWithUserID creates a context with user_id in metadata
func CreateTestContextWithUserID(userID string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "user_id", userID)
	return ctx
}

// ========================================
// Assertion Helpers
// ========================================

// AssertNoError asserts that an error is nil
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: got error %v, want nil", msg, err)
	}
}

// AssertError asserts that an error is not nil
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: got nil error, want error", msg)
	}
}

// AssertEqual asserts that two values are equal
func AssertEqual(t *testing.T, got, want interface{}, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", msg, got, want)
	}
}

// AssertNotEqual asserts that two values are not equal
func AssertNotEqual(t *testing.T, got, notWant interface{}, msg string) {
	t.Helper()
	if got == notWant {
		t.Errorf("%s: got %v, want different value", msg, got)
	}
}

// AssertTrue asserts that a condition is true
func AssertTrue(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Errorf("%s: got false, want true", msg)
	}
}

// AssertFalse asserts that a condition is false
func AssertFalse(t *testing.T, condition bool, msg string) {
	t.Helper()
	if condition {
		t.Errorf("%s: got true, want false", msg)
	}
}

// AssertNil asserts that a value is nil
func AssertNil(t *testing.T, value interface{}, msg string) {
	t.Helper()
	if value == nil {
		return
	}
	// Use reflection to handle typed nil pointers (e.g., (*string)(nil))
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return
	}
	t.Errorf("%s: got %v, want nil", msg, value)
}

// AssertNotNil asserts that a value is not nil
func AssertNotNil(t *testing.T, value interface{}, msg string) {
	t.Helper()
	if value == nil {
		t.Errorf("%s: got nil, want non-nil", msg)
	}
}

// AssertGreaterThan asserts that a > b
func AssertGreaterThan(t *testing.T, a, b int64, msg string) {
	t.Helper()
	if a <= b {
		t.Errorf("%s: got %d, want > %d", msg, a, b)
	}
}

// AssertLessThan asserts that a < b
func AssertLessThan(t *testing.T, a, b int64, msg string) {
	t.Helper()
	if a >= b {
		t.Errorf("%s: got %d, want < %d", msg, a, b)
	}
}

// AssertContains asserts that a string contains a substring
func AssertContains(t *testing.T, str, substr, msg string) {
	t.Helper()
	if !strings.Contains(str, substr) {
		t.Errorf("%s: got %q, want to contain %q", msg, str, substr)
	}
}

// ========================================
// Time Helpers
// ========================================

// FutureDate returns a date in the future by adding days
func FutureDate(days int) time.Time {
	return time.Now().AddDate(0, 0, days)
}

// PastDate returns a date in the past by subtracting days
func PastDate(days int) time.Time {
	return time.Now().AddDate(0, 0, -days)
}

// TodayDate returns today's date at midnight
func TodayDate() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// RandomString generates a random string of given length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
