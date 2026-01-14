// backend/internal/middleware/request_id_test.go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/djryanj/k8s-rbactory-backend/internal/logging"
)

func TestRequestID(t *testing.T) {
	tests := []struct {
		name              string
		existingRequestID string
		expectGenerated   bool
	}{
		{
			name:              "generates new request ID",
			existingRequestID: "",
			expectGenerated:   true,
		},
		{
			name:              "preserves existing request ID",
			existingRequestID: "existing-id-123",
			expectGenerated:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check that request ID is in context
				requestID := logging.GetRequestID(r.Context())
				if requestID == "" {
					t.Error("expected request ID in context, got empty string")
				}

				if !tt.expectGenerated && requestID != tt.existingRequestID {
					t.Errorf("expected request ID %s, got %s", tt.existingRequestID, requestID)
				}

				w.WriteHeader(http.StatusOK)
			})

			middleware := RequestID(handler)

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.existingRequestID != "" {
				req.Header.Set("X-Request-ID", tt.existingRequestID)
			}

			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			// Check response header
			responseID := rr.Header().Get("X-Request-ID")
			if responseID == "" {
				t.Error("expected X-Request-ID in response header")
			}

			if !tt.expectGenerated && responseID != tt.existingRequestID {
				t.Errorf("expected response ID %s, got %s", tt.existingRequestID, responseID)
			}
		})
	}
}

func TestRequestID_ContextPropagation(t *testing.T) {
	var capturedRequestID string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRequestID = logging.GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestID(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "test-request-123")

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	// Verify the request ID was captured in context
	if capturedRequestID != "test-request-123" {
		t.Errorf("expected request ID 'test-request-123' in context, got %s", capturedRequestID)
	}

	// Verify the request ID is in the response header
	responseID := rr.Header().Get("X-Request-ID")
	if responseID != "test-request-123" {
		t.Errorf("expected request ID 'test-request-123' in response header, got %s", responseID)
	}
}

func TestRequestID_GeneratesUniqueIDs(t *testing.T) {
	var requestIDs []string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := logging.GetRequestID(r.Context())
		requestIDs = append(requestIDs, requestID)
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestID(handler)

	// Make multiple requests without existing request IDs
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)
	}

	// Verify all request IDs are unique
	seen := make(map[string]bool)
	for _, id := range requestIDs {
		if seen[id] {
			t.Errorf("duplicate request ID generated: %s", id)
		}
		seen[id] = true
	}

	// Verify we got 5 unique IDs
	if len(seen) != 5 {
		t.Errorf("expected 5 unique request IDs, got %d", len(seen))
	}
}

func TestRequestID_HeaderFormat(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestID(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	// Check that the generated request ID looks like a UUID
	requestID := rr.Header().Get("X-Request-ID")
	if len(requestID) != 36 {
		t.Errorf("expected UUID format (length 36), got length %d: %s", len(requestID), requestID)
	}

	// Basic UUID format check (8-4-4-4-12)
	if requestID[8] != '-' || requestID[13] != '-' || requestID[18] != '-' || requestID[23] != '-' {
		t.Errorf("request ID doesn't match UUID format: %s", requestID)
	}
}

func TestRequestID_EmptyExistingID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := logging.GetRequestID(r.Context())
		if requestID == "" {
			t.Error("expected non-empty request ID in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestID(handler)

	// Send request with empty X-Request-ID header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "")

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	// Should generate a new ID when existing is empty
	responseID := rr.Header().Get("X-Request-ID")
	if responseID == "" {
		t.Error("expected generated request ID when existing is empty")
	}
}

func TestRequestID_XForwardedFor(t *testing.T) {
	// This test ensures the middleware doesn't interfere with X-Forwarded-For
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify X-Forwarded-For is preserved
		if xff := r.Header.Get("X-Forwarded-For"); xff != "192.168.1.1" {
			t.Errorf("expected X-Forwarded-For to be preserved, got %s", xff)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestID(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	// Verify request ID was still added
	if rr.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID to be set")
	}
}
