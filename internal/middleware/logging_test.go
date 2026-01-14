// backend/internal/middleware/logging_test.go
package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/djryanj/k8s-rbactory-backend/internal/logging"
)

func TestLogging(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedStatus int
	}{
		{
			name: "successful request",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "error request",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "implicit 200",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("OK"))
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := Logging(logger)
			handler := middleware(tt.handler)

			req := httptest.NewRequest("GET", "/test", nil)
			ctx := logging.WithRequestID(req.Context(), "test-123")
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check that request ID is in response header
			if reqID := rr.Header().Get("X-Request-ID"); reqID == "" {
				t.Error("expected X-Request-ID header to be set")
			}
		})
	}
}

func TestLogging_PanicRecovery(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}

	middleware := Logging(logger)
	wrappedHandler := middleware(http.HandlerFunc(handler))

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := logging.WithRequestID(req.Context(), "test-123")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	// Should not panic
	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 after panic, got %d", rr.Code)
	}
}

func TestResponseWriter_StatusTracking(t *testing.T) {
	tests := []struct {
		name           string
		writeHeader    bool
		statusCode     int
		expectedStatus int
	}{
		{
			name:           "explicit 200",
			writeHeader:    true,
			statusCode:     http.StatusOK,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "explicit 404",
			writeHeader:    true,
			statusCode:     http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "implicit 200",
			writeHeader:    false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			wrapped := wrapResponseWriter(rr)

			if tt.writeHeader {
				wrapped.WriteHeader(tt.statusCode)
			}
			wrapped.Write([]byte("test"))

			if wrapped.Status() != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, wrapped.Status())
			}

			if wrapped.Size() != 4 {
				t.Errorf("expected size 4, got %d", wrapped.Size())
			}
		})
	}
}
