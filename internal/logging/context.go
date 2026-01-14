// backend/internal/logging/context.go
package logging

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	loggerKey    contextKey = "logger"
)

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// WithLogger adds a logger to the context
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves a logger from context with request ID already attached
// If no logger is found in context, returns the default logger
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	// Fallback to default logger
	return slog.Default()
}

// NewRequestLogger creates a logger with request-specific fields
func NewRequestLogger(baseLogger *slog.Logger, requestID, method, path string) *slog.Logger {
	return baseLogger.With(
		slog.String("request_id", requestID),
		slog.String("method", method),
		slog.String("path", path),
	)
}

// GenerateRequestID generates a new UUID for request tracking
func GenerateRequestID() string {
	return uuid.New().String()
}
