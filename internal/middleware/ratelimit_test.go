// backend/internal/middleware/ratelimit_test.go
package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create rate limiter: 2 requests per second, burst of 2
	middleware := RateLimit(logger, 2, 2)
	wrappedHandler := middleware(handler)

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i+1, rr.Code)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", rr.Code)
	}

	// Wait for rate limit to reset
	time.Sleep(time.Second)

	// Should succeed again
	req = httptest.NewRequest("GET", "/test", nil)
	rr = httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 after reset, got %d", rr.Code)
	}
}

func TestRateLimit_DifferentIPs(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(logger, 1, 1)
	wrappedHandler := middleware(handler)

	// Request from IP 1
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	rr1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("IP1 first request: expected 200, got %d", rr1.Code)
	}

	// Request from IP 2 (different IP, should not be rate limited)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:12345"
	rr2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("IP2 first request: expected 200, got %d", rr2.Code)
	}

	// Second request from IP 1 (should be rate limited)
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.1:12345"
	rr3 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusTooManyRequests {
		t.Errorf("IP1 second request: expected 429, got %d", rr3.Code)
	}
}

func TestRateLimit_Cleanup(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(logger, 10, 10)
	wrappedHandler := middleware(handler)

	// Make requests from many different IPs
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = fmt.Sprintf("192.168.1.%d:12345", i)
		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)
	}

	// The cleanup should happen automatically
	// This test mainly ensures no panics occur
}

func TestRateLimit_BurstAllowance(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Rate: 1 request per second, Burst: 3
	middleware := RateLimit(logger, 1, 3)
	wrappedHandler := middleware(handler)

	// First 3 requests should succeed (burst)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("burst request %d: expected 200, got %d", i+1, rr.Code)
		}
	}

	// 4th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 after burst, got %d", rr.Code)
	}
}

func TestRateLimit_XForwardedFor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(logger, 1, 1)
	wrappedHandler := middleware(handler)

	// First request with X-Forwarded-For
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.100:12345"
	req1.Header.Set("X-Forwarded-For", "10.0.0.1")
	rr1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("first request: expected 200, got %d", rr1.Code)
	}

	// Second request with same X-Forwarded-For should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.200:12345"        // Different RemoteAddr
	req2.Header.Set("X-Forwarded-For", "10.0.0.1") // Same X-Forwarded-For
	rr2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", rr2.Code)
	}
}
