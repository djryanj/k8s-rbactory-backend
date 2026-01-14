// backend/internal/logging/helpers.go
package logging

import (
	"context"
	"log/slog"
)

// LogError logs an error with context and returns it (for chaining)
func LogError(ctx context.Context, msg string, err error, attrs ...any) error {
	logger := FromContext(ctx)

	// Prepend the error to the attributes
	allAttrs := make([]any, 0, len(attrs)+1)
	allAttrs = append(allAttrs, slog.Any("error", err))
	allAttrs = append(allAttrs, attrs...)

	logger.Error(msg, allAttrs...)
	return err
}

// LogInfo logs info with context
func LogInfo(ctx context.Context, msg string, attrs ...any) {
	logger := FromContext(ctx)
	logger.Info(msg, attrs...)
}

// LogWarn logs warning with context
func LogWarn(ctx context.Context, msg string, attrs ...any) {
	logger := FromContext(ctx)
	logger.Warn(msg, attrs...)
}

// LogDebug logs debug with context
func LogDebug(ctx context.Context, msg string, attrs ...any) {
	logger := FromContext(ctx)
	logger.Debug(msg, attrs...)
}

// LogWithLevel logs at the specified level with context
func LogWithLevel(ctx context.Context, level slog.Level, msg string, attrs ...any) {
	logger := FromContext(ctx)
	logger.Log(ctx, level, msg, attrs...)
}
