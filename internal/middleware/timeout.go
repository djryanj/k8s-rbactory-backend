// k8s-rbactory-backend/internal/middleware/timeout.go
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// Timeout middleware to prevent hanging requests
func Timeout(logger *slog.Logger, timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				logger.Error("Request timeout",
					slog.String("path", r.URL.Path),
					slog.String("method", r.Method),
				)
				http.Error(w, "Request timeout", http.StatusGatewayTimeout)
			}
		})
	}
}
