// backend/internal/middleware/logging_test.go
package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

func TestLogging_LargeResponse(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Test with a large response body
	largeBody := strings.Repeat("x", 10000)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeBody))
	})

	middleware := Logging(logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := logging.WithRequestID(req.Context(), "test-large")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	if rr.Body.Len() != len(largeBody) {
		t.Errorf("expected body size %d, got %d", len(largeBody), rr.Body.Len())
	}
}

func TestLogging_MultipleWrites(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("first"))
		w.Write([]byte("second"))
		w.Write([]byte("third"))
	})

	middleware := Logging(logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := logging.WithRequestID(req.Context(), "test-multi")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	expectedBody := "firstsecondthird"
	if rr.Body.String() != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, rr.Body.String())
	}
}

func TestLogging_DifferentStatusCodes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name       string
		statusCode int
		expectLog  string // Expected log level
	}{
		{"1xx Informational", http.StatusContinue, "INFO"},
		{"2xx Success", http.StatusCreated, "INFO"},
		{"3xx Redirect", http.StatusMovedPermanently, "INFO"},
		{"4xx Client Error", http.StatusBadRequest, "WARN"},
		{"404 Not Found", http.StatusNotFound, "WARN"},
		{"5xx Server Error", http.StatusInternalServerError, "ERROR"},
		{"503 Service Unavailable", http.StatusServiceUnavailable, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			middleware := Logging(logger)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest("GET", "/test", nil)
			ctx := logging.WithRequestID(req.Context(), "test-status")
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			if rr.Code != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, rr.Code)
			}
		})
	}
}

func TestLogging_WithoutRequestID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := Logging(logger)
	wrappedHandler := middleware(handler)

	// Request without request ID in context
	req := httptest.NewRequest("GET", "/test", nil)

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	// Should still work, just with empty request ID
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestResponseWriter_MultipleWriteHeaderCalls(t *testing.T) {
	rr := httptest.NewRecorder()
	wrapped := wrapResponseWriter(rr)

	// First call should work
	wrapped.WriteHeader(http.StatusOK)

	// Second call should be ignored
	wrapped.WriteHeader(http.StatusInternalServerError)

	if wrapped.Status() != http.StatusOK {
		t.Errorf("expected status 200, got %d", wrapped.Status())
	}
}

func TestResponseWriter_WriteBeforeWriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	wrapped := wrapResponseWriter(rr)

	// Write without calling WriteHeader first
	n, err := wrapped.Write([]byte("test"))

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if n != 4 {
		t.Errorf("expected 4 bytes written, got %d", n)
	}

	// Should default to 200
	if wrapped.Status() != http.StatusOK {
		t.Errorf("expected default status 200, got %d", wrapped.Status())
	}
}

func TestResponseWriter_Unwrap(t *testing.T) {
	rr := httptest.NewRecorder()
	wrapped := wrapResponseWriter(rr)

	unwrapped := wrapped.Unwrap()
	if unwrapped != rr {
		t.Error("Unwrap() should return the original ResponseWriter")
	}
}
