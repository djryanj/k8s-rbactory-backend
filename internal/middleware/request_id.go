// backend/internal/middleware/request_id.go
package middleware

import (
	"net/http"

	"github.com/djryanj/k8s-rbactory-backend/internal/logging"
	"github.com/google/uuid"
)

// RequestID middleware adds a unique request ID to each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID already exists (from upstream proxy)
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to response headers for client tracking
		w.Header().Set("X-Request-ID", requestID)

		// Add to context
		ctx := logging.WithRequestID(r.Context(), requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
