// backend/internal/testutil/testutil.go
package testutil

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"
	"os"

	"github.com/djryanj/k8s-rbactory-backend/internal/logging"
)

// NewTestLogger creates a logger for tests
func NewTestLogger(tb testing.TB) *slog.Logger {
	tb.Helper()
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

// NewTestContext creates a context with logger for tests
func NewTestContext(tb testing.TB) context.Context {
	tb.Helper()
	logger := NewTestLogger(tb)
	ctx := context.Background()
	ctx = logging.WithLogger(ctx, logger)
	ctx = logging.WithRequestID(ctx, "test-request-id")
	return ctx
}

// AssertStatusCode checks HTTP status code
func AssertStatusCode(tb testing.TB, expected, actual int) {
	tb.Helper()
	if expected != actual {
		tb.Errorf("expected status code %d, got %d", expected, actual)
	}
}

// AssertJSONResponse decodes JSON response and checks status
func AssertJSONResponse(tb testing.TB, resp *httptest.ResponseRecorder, expectedStatus int, target interface{}) {
	tb.Helper()

	AssertStatusCode(tb, expectedStatus, resp.Code)

	if target != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			tb.Fatalf("failed to read response body: %v", err)
		}

		if err := json.Unmarshal(body, target); err != nil {
			tb.Fatalf("failed to unmarshal JSON response: %v\nBody: %s", err, string(body))
		}
	}
}

// AssertContains checks if a string contains a substring
func AssertContains(tb testing.TB, haystack, needle string) {
	tb.Helper()
	if !contains(haystack, needle) {
		tb.Errorf("expected string to contain %q, got %q", needle, haystack)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// NewTestRequest creates a test HTTP request with context
func NewTestRequest(tb testing.TB, method, path string, body io.Reader) *http.Request {
	tb.Helper()
	req := httptest.NewRequest(method, path, body)
	req = req.WithContext(NewTestContext(tb))
	return req
}

// AssertNoError fails the test if err is not nil
func AssertNoError(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
}

// AssertError fails the test if err is nil
func AssertError(tb testing.TB, err error) {
	tb.Helper()
	if err == nil {
		tb.Fatal("expected error, got nil")
	}
}

// AssertEqual checks if two values are equal
func AssertEqual(tb testing.TB, expected, actual interface{}) {
	tb.Helper()
	if expected != actual {
		tb.Errorf("expected %v, got %v", expected, actual)
	}
}

// AssertNotEqual checks if two values are not equal
func AssertNotEqual(tb testing.TB, expected, actual interface{}) {
	tb.Helper()
	if expected == actual {
		tb.Errorf("expected values to be different, both are %v", expected)
	}
}

// AssertTrue checks if a condition is true
func AssertTrue(tb testing.TB, condition bool, message string) {
	tb.Helper()
	if !condition {
		tb.Errorf("assertion failed: %s", message)
	}
}

// AssertFalse checks if a condition is false
func AssertFalse(tb testing.TB, condition bool, message string) {
	tb.Helper()
	if condition {
		tb.Errorf("assertion failed: %s", message)
	}
}

// AssertNil checks if a value is nil
func AssertNil(tb testing.TB, value interface{}) {
	tb.Helper()
	if value != nil {
		tb.Errorf("expected nil, got %v", value)
	}
}

// AssertNotNil checks if a value is not nil
func AssertNotNil(tb testing.TB, value interface{}) {
	tb.Helper()
	if value == nil {
		tb.Error("expected non-nil value, got nil")
	}
}

// AssertLen checks if a slice/array/map has the expected length
func AssertLen(tb testing.TB, collection interface{}, expectedLen int) {
	tb.Helper()
	var actualLen int

	switch v := collection.(type) {
	case []interface{}:
		actualLen = len(v)
	case []string:
		actualLen = len(v)
	case []int:
		actualLen = len(v)
	case map[string]interface{}:
		actualLen = len(v)
	case string:
		actualLen = len(v)
	default:
		tb.Fatalf("AssertLen: unsupported type %T", collection)
	}

	if actualLen != expectedLen {
		tb.Errorf("expected length %d, got %d", expectedLen, actualLen)
	}
}
