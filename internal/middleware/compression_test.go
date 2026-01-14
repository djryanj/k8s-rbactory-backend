// backend/internal/middleware/compression_test.go
package middleware

import (
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCompression(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	largeBody := strings.Repeat("test data ", 1000)

	tests := []struct {
		name           string
		acceptEncoding string
		body           string
		expectGzip     bool
	}{
		{
			name:           "compresses when client accepts gzip",
			acceptEncoding: "gzip",
			body:           largeBody,
			expectGzip:     true,
		},
		{
			name:           "no compression when client doesn't accept gzip",
			acceptEncoding: "",
			body:           largeBody,
			expectGzip:     false,
		},
		{
			name:           "no compression for small body",
			acceptEncoding: "gzip",
			body:           "small",
			expectGzip:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.body))
			})

			middleware := Compression(logger)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}

			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			contentEncoding := rr.Header().Get("Content-Encoding")

			if tt.expectGzip {
				if contentEncoding != "gzip" {
					t.Errorf("expected Content-Encoding: gzip, got %s", contentEncoding)
				}

				// Verify the body is actually gzipped
				reader, err := gzip.NewReader(rr.Body)
				if err != nil {
					t.Fatalf("failed to create gzip reader: %v", err)
				}
				defer reader.Close()

				decompressed, err := io.ReadAll(reader)
				if err != nil {
					t.Fatalf("failed to decompress: %v", err)
				}

				if string(decompressed) != tt.body {
					t.Errorf("decompressed body doesn't match original")
				}
			} else {
				if contentEncoding == "gzip" {
					t.Error("unexpected gzip compression")
				}

				if rr.Body.String() != tt.body {
					t.Errorf("body doesn't match expected")
				}
			}
		})
	}
}
