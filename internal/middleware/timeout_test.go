// backend/internal/middleware/timeout_test.go
package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name           string
		timeout        time.Duration
		handlerDelay   time.Duration
		expectedStatus int
	}{
		{
			name:           "completes before timeout",
			timeout:        100 * time.Millisecond,
			handlerDelay:   10 * time.Millisecond,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "times out",
			timeout:        50 * time.Millisecond,
			handlerDelay:   200 * time.Millisecond,
			expectedStatus: http.StatusGatewayTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(tt.handlerDelay)
				w.WriteHeader(http.StatusOK)
			})

			middleware := Timeout(logger, tt.timeout)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
