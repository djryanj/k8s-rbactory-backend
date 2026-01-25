// backend/internal/middleware/cors_test.go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestCORS tests the main CORS functionality
func TestCORS(t *testing.T) {
	tests := []struct {
		name           string
		config         CORSConfig
		method         string
		origin         string
		requestMethod  string // For preflight (OPTIONS) requests
		requestHeaders string // For preflight requests
		expectAllowed  bool
		checkHeaders   map[string]string
		checkContains  map[string]string
	}{
		{
			name: "simple GET request with allowed origin",
			config: CORSConfig{
				AllowedOrigins:   []string{"http://localhost:5173"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"*"}, // Wildcard for development
				AllowCredentials: true,
				MaxAge:           300,
			},
			method:        "GET",
			origin:        "http://localhost:5173",
			expectAllowed: true,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://localhost:5173",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name: "simple POST request with allowed origin",
			config: CORSConfig{
				AllowedOrigins:   []string{"http://localhost:5173"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"*"},
				AllowCredentials: true,
				MaxAge:           300,
			},
			method:        "POST",
			origin:        "http://localhost:5173",
			expectAllowed: true,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://localhost:5173",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name: "preflight request for POST with wildcard headers",
			config: CORSConfig{
				AllowedOrigins:   []string{"http://localhost:5173"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"*"},
				ExposedHeaders:   []string{"X-Request-ID"},
				AllowCredentials: true,
				MaxAge:           300,
			},
			method:         "OPTIONS",
			origin:         "http://localhost:5173",
			requestMethod:  "POST",
			requestHeaders: "content-type",
			expectAllowed:  true,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://localhost:5173",
				"Access-Control-Allow-Credentials": "true",
			},
			checkContains: map[string]string{
				"Access-Control-Allow-Methods": "POST",
			},
		},
		{
			name: "preflight request for PUT with multiple headers",
			config: CORSConfig{
				AllowedOrigins:   []string{"http://localhost:5173"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"*"}, // Wildcard handles multiple headers
				AllowCredentials: true,
				MaxAge:           300,
			},
			method:         "OPTIONS",
			origin:         "http://localhost:5173",
			requestMethod:  "PUT",
			requestHeaders: "content-type, authorization",
			expectAllowed:  true,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://localhost:5173",
				"Access-Control-Allow-Credentials": "true",
			},
			checkContains: map[string]string{
				"Access-Control-Allow-Methods": "PUT",
			},
		},
		{
			name: "multiple allowed origins - first origin",
			config: CORSConfig{
				AllowedOrigins:   []string{"http://localhost:5173", "https://example.com"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"*"},
				AllowCredentials: true,
				MaxAge:           300,
			},
			method:        "GET",
			origin:        "http://localhost:5173",
			expectAllowed: true,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin": "http://localhost:5173",
			},
		},
		{
			name: "multiple allowed origins - second origin",
			config: CORSConfig{
				AllowedOrigins:   []string{"http://localhost:5173", "https://example.com"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"*"},
				AllowCredentials: true,
				MaxAge:           300,
			},
			method:        "GET",
			origin:        "https://example.com",
			expectAllowed: true,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin": "https://example.com",
			},
		},
		{
			name: "disallowed origin",
			config: CORSConfig{
				AllowedOrigins:   []string{"http://localhost:5173"},
				AllowedMethods:   []string{"GET", "POST"},
				AllowedHeaders:   []string{"*"},
				AllowCredentials: true,
				MaxAge:           300,
			},
			method:        "GET",
			origin:        "https://evil.com",
			expectAllowed: false,
		},
		{
			name: "preflight with disallowed origin",
			config: CORSConfig{
				AllowedOrigins:   []string{"http://localhost:5173"},
				AllowedMethods:   []string{"GET", "POST"},
				AllowedHeaders:   []string{"*"},
				AllowCredentials: true,
				MaxAge:           300,
			},
			method:         "OPTIONS",
			origin:         "https://evil.com",
			requestMethod:  "POST",
			requestHeaders: "content-type",
			expectAllowed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			corsMiddleware := CORSWithConfig(tt.config)
			wrappedHandler := corsMiddleware.Handler(handler)

			req := httptest.NewRequest(tt.method, "/test", nil)
			req.Header.Set("Origin", tt.origin)

			// For preflight requests, add required headers
			if tt.method == "OPTIONS" && tt.requestMethod != "" {
				req.Header.Set("Access-Control-Request-Method", tt.requestMethod)
				if tt.requestHeaders != "" {
					req.Header.Set("Access-Control-Request-Headers", tt.requestHeaders)
				}
			}

			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			if tt.expectAllowed {
				// Check expected headers (exact match)
				for header, expectedValue := range tt.checkHeaders {
					actualValue := rr.Header().Get(header)
					if actualValue != expectedValue {
						t.Errorf("expected header %s=%s, got %s", header, expectedValue, actualValue)
					}
				}

				// Check headers that should contain certain values
				for header, expectedSubstring := range tt.checkContains {
					actualValue := rr.Header().Get(header)
					if actualValue == "" {
						t.Errorf("expected header %s to be set, but it's empty", header)
					} else if !strings.Contains(actualValue, expectedSubstring) {
						t.Errorf("expected header %s to contain %s, got %s", header, expectedSubstring, actualValue)
					}
				}

				// Verify origin is set
				if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin == "" {
					t.Error("expected Access-Control-Allow-Origin header to be set for allowed origin")
				}
			} else {
				// For disallowed origins, verify CORS headers are NOT present
				if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "" {
					t.Errorf("expected no CORS headers for disallowed origin, but got Access-Control-Allow-Origin: %s", origin)
				}
			}

			// Debug output on failure
			if t.Failed() {
				t.Logf("Request: %s %s, Origin: %s", tt.method, req.URL.Path, tt.origin)
				if tt.requestMethod != "" {
					t.Logf("Preflight for method: %s, headers: %s", tt.requestMethod, tt.requestHeaders)
				}
				t.Logf("Response status: %d", rr.Code)
				t.Logf("Response headers: %v", rr.Header())
				t.Logf("Config: %+v", tt.config)
			}
		})
	}
}

func TestCORS_DefaultConfiguration(t *testing.T) {
	// Ensure no environment variables are set
	os.Unsetenv("ALLOWED_ORIGINS")
	os.Unsetenv("ALLOWED_HEADERS")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsMiddleware := CORS()
	wrappedHandler := corsMiddleware.Handler(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	origin := rr.Header().Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:5173" {
		t.Errorf("expected default origin http://localhost:5173, got %s", origin)
		t.Logf("Response headers: %v", rr.Header())
	}
}

func TestCORS_EnvironmentVariables(t *testing.T) {
	tests := []struct {
		name            string
		originsEnv      string
		headersEnv      string
		testOrigin      string
		shouldBeAllowed bool
	}{
		{
			name:            "custom origins from env",
			originsEnv:      "http://localhost:3000,http://localhost:5173,https://example.com",
			headersEnv:      "",
			testOrigin:      "http://localhost:3000",
			shouldBeAllowed: true,
		},
		{
			name:            "origin not in env list",
			originsEnv:      "http://localhost:3000,https://example.com",
			headersEnv:      "",
			testOrigin:      "https://evil.com",
			shouldBeAllowed: false,
		},
		{
			name:            "custom headers from env",
			originsEnv:      "http://localhost:5173",
			headersEnv:      "Accept,Authorization,Content-Type",
			testOrigin:      "http://localhost:5173",
			shouldBeAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			if tt.originsEnv != "" {
				os.Setenv("ALLOWED_ORIGINS", tt.originsEnv)
			}
			if tt.headersEnv != "" {
				os.Setenv("ALLOWED_HEADERS", tt.headersEnv)
			}
			defer func() {
				os.Unsetenv("ALLOWED_ORIGINS")
				os.Unsetenv("ALLOWED_HEADERS")
			}()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			corsMiddleware := CORS()
			wrappedHandler := corsMiddleware.Handler(handler)

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", tt.testOrigin)

			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			allowedOrigin := rr.Header().Get("Access-Control-Allow-Origin")

			if tt.shouldBeAllowed {
				if allowedOrigin != tt.testOrigin {
					t.Errorf("expected origin %s to be allowed, got %s", tt.testOrigin, allowedOrigin)
				}
			} else {
				if allowedOrigin == tt.testOrigin {
					t.Errorf("expected origin %s to be blocked, but it was allowed", tt.testOrigin)
				}
			}
		})
	}
}

func TestCORS_VaryHeader(t *testing.T) {
	config := CORSConfig{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsMiddleware := CORSWithConfig(config)
	wrappedHandler := corsMiddleware.Handler(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	vary := rr.Header().Get("Vary")
	if !strings.Contains(vary, "Origin") {
		t.Errorf("expected Vary header to contain Origin, got %s", vary)
	}
}

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()

	tests := []struct {
		name  string
		check func(t *testing.T, config CORSConfig)
	}{
		{
			name: "default origin is Vite dev server",
			check: func(t *testing.T, config CORSConfig) {
				if len(config.AllowedOrigins) != 1 || config.AllowedOrigins[0] != "http://localhost:5173" {
					t.Errorf("expected default origin [http://localhost:5173], got %v", config.AllowedOrigins)
				}
			},
		},
		{
			name: "includes standard HTTP methods",
			check: func(t *testing.T, config CORSConfig) {
				expectedMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
				if len(config.AllowedMethods) != len(expectedMethods) {
					t.Errorf("expected %d methods, got %d", len(expectedMethods), len(config.AllowedMethods))
				}
			},
		},
		{
			name: "uses wildcard for headers in development",
			check: func(t *testing.T, config CORSConfig) {
				if len(config.AllowedHeaders) != 1 || config.AllowedHeaders[0] != "*" {
					t.Errorf("expected wildcard headers for development, got %v", config.AllowedHeaders)
				}
			},
		},
		{
			name: "credentials enabled",
			check: func(t *testing.T, config CORSConfig) {
				if !config.AllowCredentials {
					t.Error("expected AllowCredentials to be true")
				}
			},
		},
		{
			name: "max age set to 5 minutes",
			check: func(t *testing.T, config CORSConfig) {
				if config.MaxAge != 300 {
					t.Errorf("expected MaxAge 300, got %d", config.MaxAge)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, config)
		})
	}
}

func TestProductionCORSConfig(t *testing.T) {
	config := ProductionCORSConfig()

	tests := []struct {
		name  string
		check func(t *testing.T, config CORSConfig)
	}{
		{
			name: "no default origins (must be explicitly set)",
			check: func(t *testing.T, config CORSConfig) {
				if len(config.AllowedOrigins) != 0 {
					t.Errorf("expected no default origins in production config, got %v", config.AllowedOrigins)
				}
			},
		},
		{
			name: "uses explicit headers (not wildcard)",
			check: func(t *testing.T, config CORSConfig) {
				if len(config.AllowedHeaders) == 0 {
					t.Error("expected explicit headers in production config")
				}
				for _, h := range config.AllowedHeaders {
					if h == "*" {
						t.Error("production config should not use wildcard headers")
					}
				}
			},
		},
		{
			name: "includes common secure headers",
			check: func(t *testing.T, config CORSConfig) {
				requiredHeaders := []string{"Authorization", "Content-Type"}
				for _, required := range requiredHeaders {
					found := false
					for _, h := range config.AllowedHeaders {
						if h == required {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected production config to include %s header", required)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, config)
		})
	}
}

func TestGetAllowedOriginsFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected []string
	}{
		{
			name:     "single origin",
			envValue: "http://localhost:5173",
			expected: []string{"http://localhost:5173"},
		},
		{
			name:     "multiple origins",
			envValue: "http://localhost:5173,https://example.com",
			expected: []string{"http://localhost:5173", "https://example.com"},
		},
		{
			name:     "origins with whitespace",
			envValue: "http://localhost:5173, https://example.com , http://test.com",
			expected: []string{"http://localhost:5173", "https://example.com", "http://test.com"},
		},
		{
			name:     "empty value",
			envValue: "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("ALLOWED_ORIGINS", tt.envValue)
			} else {
				os.Unsetenv("ALLOWED_ORIGINS")
			}
			defer os.Unsetenv("ALLOWED_ORIGINS")

			result := getAllowedOriginsFromEnv()

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d origins, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("expected origin[%d]=%s, got %s", i, expected, result[i])
				}
			}
		})
	}
}

func TestGetAllowedHeadersFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected []string
	}{
		{
			name:     "single header",
			envValue: "Authorization",
			expected: []string{"Authorization"},
		},
		{
			name:     "multiple headers",
			envValue: "Accept,Authorization,Content-Type",
			expected: []string{"Accept", "Authorization", "Content-Type"},
		},
		{
			name:     "headers with whitespace",
			envValue: "Accept, Authorization , Content-Type",
			expected: []string{"Accept", "Authorization", "Content-Type"},
		},
		{
			name:     "empty value",
			envValue: "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("ALLOWED_HEADERS", tt.envValue)
			} else {
				os.Unsetenv("ALLOWED_HEADERS")
			}
			defer os.Unsetenv("ALLOWED_HEADERS")

			result := getAllowedHeadersFromEnv()

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d headers, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("expected header[%d]=%s, got %s", i, expected, result[i])
				}
			}
		})
	}
}

func TestCORS_PreflightWithMultipleHeaders(t *testing.T) {
	// This test specifically validates that wildcard headers work with multiple headers
	config := CORSConfig{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for preflight requests")
		w.WriteHeader(http.StatusOK)
	})

	corsMiddleware := CORSWithConfig(config)
	wrappedHandler := corsMiddleware.Handler(handler)

	// Test preflight with multiple headers
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "content-type, authorization, x-custom-header")

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	// Verify CORS headers are present
	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "http://localhost:5173" {
		t.Errorf("expected Access-Control-Allow-Origin, got %s", origin)
	}

	if methods := rr.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(methods, "POST") {
		t.Errorf("expected Access-Control-Allow-Methods to contain POST, got %s", methods)
	}

	// The allowed headers should reflect what was requested
	allowedHeaders := rr.Header().Get("Access-Control-Allow-Headers")
	if allowedHeaders == "" {
		t.Error("expected Access-Control-Allow-Headers to be set")
	}

	t.Logf("Preflight response headers: %v", rr.Header())
}
