// backend/internal/middleware/integration_test.go
package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// TestMiddlewareChain tests multiple middleware working together
func TestMiddlewareChain(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Apply middleware in order - use http.Handler interface type
	var wrapped http.Handler = handler
	wrapped = Logging(logger)(wrapped)
	wrapped = Recovery(logger)(wrapped)
	wrapped = Timeout(logger, 5*time.Second)(wrapped)
	wrapped = RequestID(wrapped)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	// Check that request ID was set
	if reqID := rr.Header().Get("X-Request-ID"); reqID == "" {
		t.Error("expected X-Request-ID header")
	}

	if body := rr.Body.String(); body != "success" {
		t.Errorf("expected body 'success', got %s", body)
	}
}

// TestMiddlewareChain_WithPanic tests panic recovery in middleware chain
func TestMiddlewareChain_WithPanic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Apply middleware in order
	var wrapped http.Handler = handler
	wrapped = Logging(logger)(wrapped)
	wrapped = Recovery(logger)(wrapped)
	wrapped = RequestID(wrapped)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// Should not panic
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 after panic, got %d", rr.Code)
	}

	// Request ID should still be set
	if reqID := rr.Header().Get("X-Request-ID"); reqID == "" {
		t.Error("expected X-Request-ID header even after panic")
	}
}

// TestMiddlewareChain_WithTimeout tests timeout in middleware chain
func TestMiddlewareChain_WithTimeout(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	// Apply middleware in order
	var wrapped http.Handler = handler
	wrapped = Logging(logger)(wrapped)
	wrapped = Recovery(logger)(wrapped)
	wrapped = Timeout(logger, 50*time.Millisecond)(wrapped)
	wrapped = RequestID(wrapped)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusGatewayTimeout {
		t.Errorf("expected status 504, got %d", rr.Code)
	}
}

// TestMiddlewareChain_Order tests that middleware order matters
func TestMiddlewareChain_Order(t *testing.T) {
	var executionOrder []string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		executionOrder = append(executionOrder, "handler")
		w.WriteHeader(http.StatusOK)
	})

	// Custom middleware to track execution order
	trackingMiddleware := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				executionOrder = append(executionOrder, name+"-before")
				next.ServeHTTP(w, r)
				executionOrder = append(executionOrder, name+"-after")
			})
		}
	}

	// Apply middleware
	var wrapped http.Handler = handler
	wrapped = trackingMiddleware("third")(wrapped)
	wrapped = trackingMiddleware("second")(wrapped)
	wrapped = trackingMiddleware("first")(wrapped)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	expected := []string{
		"first-before",
		"second-before",
		"third-before",
		"handler",
		"third-after",
		"second-after",
		"first-after",
	}

	if len(executionOrder) != len(expected) {
		t.Errorf("expected %d execution steps, got %d", len(expected), len(executionOrder))
		t.Logf("Got: %v", executionOrder)
		t.Logf("Expected: %v", expected)
	}

	for i, step := range expected {
		if i >= len(executionOrder) {
			t.Errorf("step %d: expected %s, but execution stopped early", i, step)
			break
		}
		if executionOrder[i] != step {
			t.Errorf("step %d: expected %s, got %s", i, step, executionOrder[i])
		}
	}
}

// TestMiddlewareChain_RealWorldScenario tests a realistic middleware stack
func TestMiddlewareChain_RealWorldScenario(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some work
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Build a realistic middleware stack
	// Order matters: RequestID -> Logging -> Recovery -> RateLimit -> Timeout -> Handler
	var wrapped http.Handler = handler
	wrapped = Timeout(logger, 1*time.Second)(wrapped)
	wrapped = RateLimit(logger, 100, 10)(wrapped)
	wrapped = Recovery(logger)(wrapped)
	wrapped = Logging(logger)(wrapped)
	wrapped = RequestID(wrapped)

	req := httptest.NewRequest("GET", "/api/status", nil)
	req.Header.Set("User-Agent", "TestClient/1.0")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	// Verify response
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	// Verify request ID is set
	if reqID := rr.Header().Get("X-Request-ID"); reqID == "" {
		t.Error("expected X-Request-ID header")
	}

	// Verify response body
	if body := rr.Body.String(); body != `{"status":"ok"}` {
		t.Errorf("expected JSON response, got %s", body)
	}
}

// TestMiddlewareChain_CORSIntegration tests CORS with other middleware
func TestMiddlewareChain_CORSIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Build middleware stack with CORS
	corsConfig := CORSConfig{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	}

	var wrapped http.Handler = handler
	wrapped = Logging(logger)(wrapped)
	wrapped = Recovery(logger)(wrapped)
	wrapped = RequestID(wrapped)
	wrapped = CORSWithConfig(corsConfig).Handler(wrapped)

	// Test actual request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	// Verify CORS headers
	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "http://localhost:3000" {
		t.Errorf("expected CORS origin header, got %s", origin)
	}

	// Verify request ID
	if reqID := rr.Header().Get("X-Request-ID"); reqID == "" {
		t.Error("expected X-Request-ID header")
	}

	// Verify response
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

// TestMiddlewareChain_CompressionIntegration tests compression with other middleware
func TestMiddlewareChain_CompressionIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	largeBody := []byte(strings.Repeat("test data ", 200))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(largeBody)
	})

	// Build middleware stack with compression
	var wrapped http.Handler = handler
	wrapped = Compression(logger)(wrapped)
	wrapped = Logging(logger)(wrapped)
	wrapped = Recovery(logger)(wrapped)
	wrapped = RequestID(wrapped)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	// Verify compression
	if encoding := rr.Header().Get("Content-Encoding"); encoding != "gzip" {
		t.Errorf("expected gzip encoding, got %s", encoding)
	}

	// Verify request ID
	if reqID := rr.Header().Get("X-Request-ID"); reqID == "" {
		t.Error("expected X-Request-ID header")
	}

	// Verify response
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}
