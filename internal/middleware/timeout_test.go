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

func TestTimeout_ContextCancellation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if context is cancelled
		select {
		case <-r.Context().Done():
			t.Log("Context was cancelled as expected")
			return
		case <-time.After(200 * time.Millisecond):
			t.Error("Context should have been cancelled")
		}
	})

	middleware := Timeout(logger, 50*time.Millisecond)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusGatewayTimeout {
		t.Errorf("expected status 504, got %d", rr.Code)
	}
}

func TestTimeout_VeryShortTimeout(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	middleware := Timeout(logger, 1*time.Millisecond)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusGatewayTimeout {
		t.Errorf("expected timeout status 504, got %d", rr.Code)
	}
}

func TestTimeout_ZeroTimeout(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Zero timeout should timeout immediately
	middleware := Timeout(logger, 0)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	// With zero timeout, it should timeout
	if rr.Code != http.StatusGatewayTimeout {
		t.Errorf("expected timeout with zero duration, got %d", rr.Code)
	}
}
