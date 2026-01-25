// backend/internal/middleware/recovery_test.go
package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRecovery(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedStatus int
		shouldPanic    bool
	}{
		{
			name: "normal request",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			},
			expectedStatus: http.StatusOK,
			shouldPanic:    false,
		},
		{
			name: "panic with string",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("test panic")
			},
			expectedStatus: http.StatusInternalServerError,
			shouldPanic:    true,
		},
		{
			name: "panic with error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic(http.ErrAbortHandler)
			},
			expectedStatus: http.StatusInternalServerError,
			shouldPanic:    true,
		},
		{
			name: "panic with nil",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var ptr *string
				_ = *ptr // nil pointer dereference
			},
			expectedStatus: http.StatusInternalServerError,
			shouldPanic:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := Recovery(logger)
			wrappedHandler := middleware(tt.handler)

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			// Should not panic - middleware should catch it
			wrappedHandler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.shouldPanic {
				body := rr.Body.String()
				if body != "Internal Server Error\n" {
					t.Errorf("expected 'Internal Server Error', got %s", body)
				}
			}
		})
	}
}

func TestRecovery_PreservesResponseBeforePanic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("partial response"))
		panic("panic after partial write")
	})

	middleware := Recovery(logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	// The status should be OK since it was written before panic
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	// The partial response should be preserved
	// Note: Additional error text may be appended since headers were already sent
	body := rr.Body.String()
	if !strings.HasPrefix(body, "partial response") {
		t.Errorf("expected body to start with 'partial response', got %s", body)
	}

	// Verify the partial response is there
	if !strings.Contains(body, "partial response") {
		t.Errorf("expected body to contain 'partial response', got %s", body)
	}
}

func TestRecovery_PanicBeforeWrite(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Panic before writing anything
		panic("panic before write")
	})

	middleware := Recovery(logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	// Should get 500 error
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}

	// Should have error message
	body := rr.Body.String()
	if body != "Internal Server Error\n" {
		t.Errorf("expected 'Internal Server Error', got %s", body)
	}
}

func TestRecovery_PanicAfterWriteHeader(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated) // 201
		panic("panic after WriteHeader")
	})

	middleware := Recovery(logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	// Status should be 201 since it was set before panic
	// (headers are already sent, can't be changed)
	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201 (set before panic), got %d", rr.Code)
	}

	// Body will have the error message appended
	body := rr.Body.String()
	if !strings.Contains(body, "Internal Server Error") {
		t.Errorf("expected body to contain error message, got %s", body)
	}
}

func TestRecovery_MultiplePanics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	panicCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panicCount++
		panic("test panic")
	})

	middleware := Recovery(logger)
	wrappedHandler := middleware(handler)

	// Make multiple requests that panic
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("request %d: expected status 500, got %d", i+1, rr.Code)
		}
	}

	if panicCount != 3 {
		t.Errorf("expected 3 panics, got %d", panicCount)
	}
}

func TestRecovery_DoesNotAffectNormalRequests(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := Recovery(logger)
	wrappedHandler := middleware(handler)

	// Make multiple normal requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i+1, rr.Code)
		}

		if rr.Body.String() != "success" {
			t.Errorf("request %d: expected 'success', got %s", i+1, rr.Body.String())
		}
	}

	if callCount != 5 {
		t.Errorf("expected 5 calls, got %d", callCount)
	}
}

func TestRecovery_PanicTypes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name      string
		panicWith interface{}
	}{
		{"string", "panic string"},
		{"int", 42},
		{"error", http.ErrAbortHandler},
		{"nil", nil},
		{"struct", struct{ msg string }{"panic"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(tt.panicWith)
			})

			middleware := Recovery(logger)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			// Should not panic regardless of panic type
			wrappedHandler.ServeHTTP(rr, req)

			if rr.Code != http.StatusInternalServerError {
				t.Errorf("expected status 500, got %d", rr.Code)
			}
		})
	}
}

func BenchmarkRecovery(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := Recovery(logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)
	}
}

func BenchmarkRecovery_WithPanic(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("benchmark panic")
	})

	middleware := Recovery(logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)
	}
}

func TestRecovery_PanicWithDifferentTypes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name      string
		panicWith interface{}
	}{
		{"panic with string", "panic message"},
		{"panic with int", 42},
		{"panic with error", errors.New("error message")},
		{"panic with nil", nil},
		{"panic with struct", struct{ msg string }{"panic"}},
		{"panic with slice", []string{"a", "b"}},
		{"panic with map", map[string]string{"key": "value"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(tt.panicWith)
			})

			middleware := Recovery(logger)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			// Should not panic
			wrappedHandler.ServeHTTP(rr, req)

			if rr.Code != http.StatusInternalServerError {
				t.Errorf("expected status 500, got %d", rr.Code)
			}
		})
	}
}

func TestRecovery_NoPanic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := Recovery(logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if callCount != 1 {
		t.Errorf("expected handler to be called once, got %d", callCount)
	}

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	if rr.Body.String() != "success" {
		t.Errorf("expected body 'success', got %s", rr.Body.String())
	}
}
