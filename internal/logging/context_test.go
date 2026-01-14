// backend/internal/logging/context_test.go
package logging

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

func TestRequestIDContext(t *testing.T) {
	ctx := context.Background()

	// Test empty context
	if id := GetRequestID(ctx); id != "" {
		t.Errorf("expected empty request ID, got %s", id)
	}

	// Test with request ID
	testID := "test-123"
	ctx = WithRequestID(ctx, testID)

	if id := GetRequestID(ctx); id != testID {
		t.Errorf("expected request ID %s, got %s", testID, id)
	}
}

func TestLoggerContext(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Test with logger
	ctx = WithLogger(ctx, logger)

	retrievedLogger := FromContext(ctx)
	if retrievedLogger == nil {
		t.Error("expected logger from context, got nil")
	}

	// Test without logger (should return default)
	emptyCtx := context.Background()
	defaultLogger := FromContext(emptyCtx)
	if defaultLogger == nil {
		t.Error("expected default logger, got nil")
	}
}

func TestNewRequestLogger(t *testing.T) {
	baseLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	requestLogger := NewRequestLogger(baseLogger, "req-123", "GET", "/api/v1/roles")

	if requestLogger == nil {
		t.Error("expected request logger, got nil")
	}

	// The logger should have the request metadata attached
	// This is hard to test directly, but we can verify it doesn't panic
	requestLogger.Info("test message")
}

func TestGenerateRequestID(t *testing.T) {
	id1 := GenerateRequestID()
	id2 := GenerateRequestID()

	if id1 == "" {
		t.Error("expected non-empty request ID")
	}

	if id1 == id2 {
		t.Error("expected unique request IDs")
	}

	// UUID format check (basic)
	if len(id1) != 36 {
		t.Errorf("expected UUID length 36, got %d", len(id1))
	}
}
