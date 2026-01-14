// backend/internal/middleware/cors_test.go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCORS(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		origin        string
		envOrigins    string
		expectAllowed bool
		expectHeaders map[string]string
	}{
		{
			name:          "preflight request",
			method:        "OPTIONS",
			origin:        "http://localhost:3000",
			envOrigins:    "http://localhost:3000",
			expectAllowed: true,
			expectHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "http://localhost:3000",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
			},
		},
		{
			name:          "GET request with allowed origin",
			method:        "GET",
			origin:        "http://localhost:3000",
			envOrigins:    "http://localhost:3000",
			expectAllowed: true,
			expectHeaders: map[string]string{
				"Access-Control-Allow-Origin": "http://localhost:3000",
			},
		},
		{
			name:          "multiple allowed origins",
			method:        "GET",
			origin:        "https://example.com",
			envOrigins:    "http://localhost:3000,https://example.com",
			expectAllowed: true,
			expectHeaders: map[string]string{
				"Access-Control-Allow-Origin": "https://example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envOrigins != "" {
				os.Setenv("ALLOWED_ORIGINS", tt.envOrigins)
				defer os.Unsetenv("ALLOWED_ORIGINS")
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			corsMiddleware := CORS()
			wrappedHandler := corsMiddleware.Handler(handler)

			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			// Check expected headers
			for header, expectedValue := range tt.expectHeaders {
				actualValue := rr.Header().Get(header)
				if actualValue != expectedValue {
					t.Errorf("expected header %s=%s, got %s", header, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestCORS_DefaultOrigins(t *testing.T) {
	// Ensure no environment variable is set
	os.Unsetenv("ALLOWED_ORIGINS")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsMiddleware := CORS()
	wrappedHandler := corsMiddleware.Handler(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	// Should allow localhost:3000 by default
	origin := rr.Header().Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:3000" {
		t.Errorf("expected default origin http://localhost:3000, got %s", origin)
	}
}
