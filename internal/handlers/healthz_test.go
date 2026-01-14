// backend/internal/handlers/healthz_test.go
package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/djryanj/k8s-rbactory-backend/internal/testutil"
)

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "GET request returns healthy",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "POST request (should still work)",
			method:         "POST",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "HEAD request",
			method:         "HEAD",
			expectedStatus: http.StatusOK,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/healthz", nil)
			rr := httptest.NewRecorder()

			SimpleHealthHandler(rr, req)

			testutil.AssertStatusCode(t, tt.expectedStatus, rr.Code)

			if tt.checkResponse {
				var response HealthResponse
				testutil.AssertJSONResponse(t, rr, http.StatusOK, &response)

				if response.Status != "healthy" {
					t.Errorf("expected status 'healthy', got %s", response.Status)
				}

				if response.Version == "" {
					t.Error("expected version to be set")
				}
			}

			// Check Content-Type header
			contentType := rr.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got %s", contentType)
			}
		})
	}
}

func TestHealthHandler_ResponseFormat(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()

	SimpleHealthHandler(rr, req)

	// Verify JSON structure
	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Check required fields
	requiredFields := []string{"status", "version"}
	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("expected field %s in response", field)
		}
	}
}

func BenchmarkHealthHandler(b *testing.B) {
	req := httptest.NewRequest("GET", "/healthz", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		SimpleHealthHandler(rr, req)
	}
}
