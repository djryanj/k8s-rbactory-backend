// backend/internal/logging/helpers_test.go
package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestLogError(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	ctx := context.Background()
	ctx = WithLogger(ctx, logger)

	testErr := errors.New("test error")

	// Log error
	returnedErr := LogError(ctx, "test message", testErr,
		slog.String("key", "value"),
	)

	// Should return the same error
	if returnedErr != testErr {
		t.Error("expected LogError to return the same error")
	}

	// Parse log output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	// Check log level
	if level := logEntry["level"]; level != "ERROR" {
		t.Errorf("expected level ERROR, got %v", level)
	}

	// Check message
	if msg := logEntry["msg"]; msg != "test message" {
		t.Errorf("expected message 'test message', got %v", msg)
	}

	// Check error is present
	if _, ok := logEntry["error"]; !ok {
		t.Error("expected error field in log output")
	}

	// Check custom attribute
	if val := logEntry["key"]; val != "value" {
		t.Errorf("expected key='value', got %v", val)
	}
}

func TestLogInfo(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	ctx := context.Background()
	ctx = WithLogger(ctx, logger)

	LogInfo(ctx, "info message", slog.String("key", "value"))

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if level := logEntry["level"]; level != "INFO" {
		t.Errorf("expected level INFO, got %v", level)
	}

	if msg := logEntry["msg"]; msg != "info message" {
		t.Errorf("expected message 'info message', got %v", msg)
	}
}

func TestLogWarn(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	ctx := context.Background()
	ctx = WithLogger(ctx, logger)

	LogWarn(ctx, "warning message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if level := logEntry["level"]; level != "WARN" {
		t.Errorf("expected level WARN, got %v", level)
	}
}

func TestLogDebug(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	ctx := context.Background()
	ctx = WithLogger(ctx, logger)

	LogDebug(ctx, "debug message")

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Error("expected debug message in output")
	}
}
